package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/ChaosHour/go-pass/internal/config"
	"github.com/fatih/color"
	_ "github.com/go-sql-driver/mysql"
)

var green = color.New(color.FgGreen).SprintFunc()
var red = color.New(color.FgRed).SprintFunc()

// Connect establishes a connection to the MySQL database
func Connect(ctx context.Context, cfg *config.Config) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:3306)/", cfg.MySQLUser, cfg.MySQLPass, cfg.SourceHost)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Printf(red("[!]"), "Failed to open database: %v", err)
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	if err := db.PingContext(ctx); err != nil {
		log.Printf(red("[!]"), "Failed to ping database: %v", err)
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	log.Println(green("[+]"), "Connected to database:", cfg.MySQLUser+"@tcp("+cfg.SourceHost+":3306)/mysql")
	return db, nil
}

// DumpUserAccounts dumps user accounts to a file
func DumpUserAccounts(ctx context.Context, db *sql.DB, cfg *config.Config) error {
	var query string
	if cfg.OnlyUser != "" {
		query = "SELECT user, host FROM mysql.user WHERE user = ?"
	} else {
		query = "SELECT user, host FROM mysql.user WHERE user NOT IN ('mysql.infoschema', 'mysql.session', 'mysql.sys')"
	}

	rows, err := db.QueryContext(ctx, query, cfg.OnlyUser)
	if err != nil {
		return fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []struct{ user, host string }
	for rows.Next() {
		var user, host string
		if err := rows.Scan(&user, &host); err != nil {
			return fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, struct{ user, host string }{user, host})
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("rows error: %w", err)
	}

	if cfg.Format == "pt-like" || cfg.Format == "import" {
		_, err = db.ExecContext(ctx, "SET print_identified_with_as_hex = 1;")
		if err != nil {
			return fmt.Errorf("failed to set print_identified_with_as_hex: %w", err)
		}
		defer func() {
			db.ExecContext(ctx, "SET print_identified_with_as_hex = 0;")
		}()
	}

	var outputLines []string
	if cfg.Format == "pt-like" {
		outputLines = append(outputLines, "-- Grants dumped by go-pass")
		outputLines = append(outputLines, fmt.Sprintf("-- Dumped from server %s via TCP/IP, MySQL at %s", cfg.SourceHost, time.Now().Format("2006-01-02 15:04:05")))
	}

	for _, u := range users {
		switch cfg.Format {
		case "raw":
			outputLines = append(outputLines, fmt.Sprintf("SHOW CREATE USER `%s`@`%s`; SHOW GRANTS FOR `%s`@`%s`;", u.user, u.host, u.user, u.host))
		case "pt-like", "import":
			// Execute SHOW CREATE USER
			var createStmt string
			err = db.QueryRowContext(ctx, fmt.Sprintf("SHOW CREATE USER `%s`@`%s`", u.user, u.host)).Scan(&createStmt)
			if err != nil {
				return fmt.Errorf("failed to show create user for %s@%s: %w", u.user, u.host, err)
			}

			switch cfg.Format {
			case "pt-like":
				outputLines = append(outputLines, fmt.Sprintf("-- Grants for '%s'@'%s'", u.user, u.host))
				// Split IDENTIFIED for ALTER
				if strings.HasPrefix(createStmt, "CREATE USER ") {
					afterCreate := createStmt[12:] // remove "CREATE USER "
					idx := strings.Index(afterCreate, " IDENTIFIED ")
					if idx > 0 {
						userHost := strings.TrimSpace(afterCreate[:idx])
						identified := strings.TrimSpace(afterCreate[idx+12:])
						outputLines = append(outputLines, fmt.Sprintf("CREATE USER IF NOT EXISTS %s;", userHost))
						outputLines = append(outputLines, fmt.Sprintf("ALTER USER %s IDENTIFIED %s;", userHost, identified))
					} else {
						// no IDENTIFIED, just CREATE USER
						userHost := strings.TrimSpace(afterCreate)
						outputLines = append(outputLines, fmt.Sprintf("CREATE USER IF NOT EXISTS %s;", userHost))
					}
				} else {
					outputLines = append(outputLines, createStmt+";")
				}
			case "import":
				createStmt = strings.Replace(createStmt, "CREATE USER", "CREATE USER IF NOT EXISTS", 1)
				outputLines = append(outputLines, fmt.Sprintf("-- CREATE USER IF NOT EXISTS for %s@%s: ", u.user, u.host))
				outputLines = append(outputLines, createStmt+";")
			}

			// Execute SHOW GRANTS
			grantRows, err := db.QueryContext(ctx, fmt.Sprintf("SHOW GRANTS FOR `%s`@`%s`", u.user, u.host))
			if err != nil {
				return fmt.Errorf("failed to show grants for %s@%s: %w", u.user, u.host, err)
			}
			defer grantRows.Close()

			for grantRows.Next() {
				var grant string
				if err := grantRows.Scan(&grant); err != nil {
					return fmt.Errorf("failed to scan grant: %w", err)
				}
				if cfg.Format == "import" {
					grant = strings.Replace(grant, "CREATE USER IF NOT EXISTS", "CREATE USER", -1)
				}
				outputLines = append(outputLines, grant+";")
			}
			if err := grantRows.Err(); err != nil {
				return fmt.Errorf("grant rows error: %w", err)
			}
		}
	}

	file, err := os.Create(cfg.DumpFile)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	for _, line := range outputLines {
		if _, err := file.WriteString(line + "\n"); err != nil {
			return fmt.Errorf("failed to write to file: %w", err)
		}
	}

	if err := file.Sync(); err != nil {
		return fmt.Errorf("failed to sync file: %w", err)
	}

	return nil
}

// RunQuery executes SQL statements from the dump file and prints results
func RunQuery(ctx context.Context, db *sql.DB, cfg *config.Config) error {
	data, err := os.ReadFile(cfg.DumpFile)
	if err != nil {
		return fmt.Errorf("failed to read dump file: %w", err)
	}

	statements := strings.Split(string(data), ";")
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" || strings.HasPrefix(stmt, "--") {
			continue
		}
		rows, err := db.QueryContext(ctx, stmt)
		if err != nil {
			return fmt.Errorf("failed to execute statement: %w", err)
		}
		defer rows.Close()

		columns, err := rows.Columns()
		if err != nil {
			return fmt.Errorf("failed to get columns: %w", err)
		}

		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		for rows.Next() {
			if err := rows.Scan(valuePtrs...); err != nil {
				return fmt.Errorf("failed to scan row: %w", err)
			}
			for i, col := range values {
				if col == nil {
					fmt.Printf("-- %s: NULL\n", columns[i])
				} else {
					fmt.Printf("-- %s: %s\n", columns[i], col)
				}
			}
			fmt.Println()
		}
	}
	return nil
}
