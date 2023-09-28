package main

import (
	"database/sql"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	_ "github.com/go-sql-driver/mysql"
)

// Define flags
var (
	source = flag.String("s", "", "Source Host")
	file   = flag.String("f", "", "dump file")
	only   = flag.String("o", "", "Only dump the specified user")
	help   = flag.Bool("h", false, "Print help")
)

// define colors
var green = color.New(color.FgGreen).SprintFunc()
var red = color.New(color.FgRed).SprintFunc()
var yellow = color.New(color.FgYellow).SprintFunc()

//var blue = color.New(color.FgBlue).SprintFunc()

// parse flags
func init() {
	flag.Parse()
}

// global variables
var (
	db  *sql.DB
	err error
)

// read the ~/.my.cnf file to get the database credentials
func readMyCnf() {
	file, err := ioutil.ReadFile(os.Getenv("HOME") + "/.my.cnf")
	if err != nil {
		handleError(err)
	}
	lines := strings.Split(string(file), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "user") {
			os.Setenv("MYSQL_USER", strings.TrimSpace(line[5:]))
		}
		if strings.HasPrefix(line, "password") {
			os.Setenv("MYSQL_PASSWORD", strings.TrimSpace(line[9:]))
		}
	}
}

// connet to the source database and create a connection
func connectToDatabase() {

	db, err = sql.Open("mysql", os.Getenv("MYSQL_USER")+":"+os.Getenv("MYSQL_PASSWORD")+"@tcp("+*source+":3306)/")

	if err != nil {
		handleError(err)
	}
	err = db.Ping()
	if err != nil {
		handleError(err)
	}
	log.Println(green("[+]"), "Connecting to database:", os.Getenv("MYSQL_USER")+":"+os.Getenv("MYSQL_PASSWORD")+"@tcp("+*source+":3306)/mysql")
}

// Create a function to dump the user accounts to a file
func dumpUserAccounts() {
	// Get the user accounts from the source database
	var rows *sql.Rows
	var err error
	if *only != "" {
		rows, err = db.Query("SELECT CONCAT('SHOW CREATE USER ', quote(user), '@', quote(host), '; SHOW GRANTS FOR ', quote(user), '@', quote(host), ';') as user FROM mysql.user WHERE user = ?", *only)
	} else {
		rows, err = db.Query("SELECT CONCAT('SHOW CREATE USER ', quote(user), '@', quote(host), '; SHOW GRANTS FOR ', quote(user), '@', quote(host), ';') as user FROM mysql.user WHERE user NOT IN ('mysql.infoschema', 'mysql.session', 'mysql.sys')")
	}
	if err != nil {
		handleError(err)
	}
	defer rows.Close()

	var users []string
	for rows.Next() {
		var user string
		err := rows.Scan(&user)
		if err != nil {
			handleError(err)
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		handleError(err)
	}

	fileName := *file
	// Check if file exists and has write permissions
	fileInfo, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		// File doesn't exist, try to create it
		file, err := os.Create(fileName)
		if err != nil {
			handleError(err)
		}
		defer file.Close()
		fileInfo, err = file.Stat()
		if err != nil {
			handleError(err)
		}
	} else if err != nil {
		handleError(err)
	}

	// File exists and has write permissions, do something with it
	if !fileInfo.Mode().IsRegular() {
		handleError(fmt.Errorf("error: Not a regular file"))
	} else if fileInfo.Mode().Perm()&os.FileMode(0200) == 0 {
		handleError(fmt.Errorf("error: File is not writable"))
	}

	// Create the file and write the user accounts to it
	file, err := os.Create(fileName)
	if err != nil {
		handleError(err)
	}
	defer file.Close()

	// add this to the top of the file -> SET print_identified_with_as_hex = 1;
	file.Seek(0, 0)
	file.WriteString("SET print_identified_with_as_hex = 1;\n")
	defer func() {
		_, err := db.Exec("SET print_identified_with_as_hex = 0;")
		if err != nil {
			handleError(err)
		}
	}()

	for _, user := range users {
		if _, err = file.WriteString(user + "\n"); err != nil {
			handleError(err)
		}
	}

	if err = file.Sync(); err != nil {
		handleError(err)
	}

	//fmt.Println(yellow("[+]"), "Wrote to file:", fileName)
}

// Create a function to read and apply the sql query from the file back to the source database
func runQuery() {
	// Read SQL file
	file, err := ioutil.ReadFile(*file)
	if err != nil {
		handleError(err)
	}

	// Split SQL file into statements
	statements := strings.Split(string(file), ";")

	// Execute each statement one by one
	for _, statement := range statements {
		if strings.TrimSpace(statement) == "" {
			continue
		}
		rows, err := db.Query(statement)
		if err != nil {
			handleError(err)
		}
		defer rows.Close()

		// Print out each row of results
		columns, err := rows.Columns()
		if err != nil {
			handleError(err)
		}

		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		for rows.Next() {
			if err := rows.Scan(valuePtrs...); err != nil {
				handleError(err)
			}

			for i, col := range values {
				if col == nil {
					fmt.Printf("-- %s: \n NULL;", columns[i]) // append semicolon to printed string
				} else {
					fmt.Printf("-- %s: \n %s;", columns[i], col) // append semicolon to printed string
				}
			}
			fmt.Println()

		}
	}
}

// print the help message
func printHelp() {
	fmt.Println("Usage: ./go-pass -s < source host> -f <dump file>")
	fmt.Println("Options:")
	fmt.Println("Usage: ./go-pass -s < source host> -f <dump file>" + yellow(" -o <user>"))

}

// handleError is a helper function to handle errors
func handleError(err error) {
	log.Fatal(red("[!]"), err)
}

// main is the entry point of the application
func main() {
	if *help {
		printHelp()
		os.Exit(0)
	}

	flag.Parse()

	// read the ~/.my.cnf file to get the database credentials. check that the file exists
	if _, err := os.Stat(os.Getenv("HOME") + "/.my.cnf"); os.IsNotExist(err) {
		fmt.Println(red("[+]"), "Please create a ~/.my.cnf file with the database credentials.")
		os.Exit(1)
	}

	readMyCnf()
	connectToDatabase()

	// make sure the source and target flags are set
	if *source == "" || *file == "" {
		printHelp()
		os.Exit(1)
	} else if *source == *file {
		printHelp()
		os.Exit(1)
	} else {

		if *file != "" {
			fmt.Println(yellow("[+]"), "Dumping user accounts to file:", *file)
			dumpUserAccounts()
			defer db.Close()

			// sleep for 5 seconds
			time.Sleep(5 * time.Second)
			runQuery()
			defer db.Close()
		}
	}
}
