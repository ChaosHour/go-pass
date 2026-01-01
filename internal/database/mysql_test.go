package database

import (
	"context"
	"os"
	"testing"

	"github.com/ChaosHour/go-pass/internal/config"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestDumpUserAccounts_Raw(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	cfg := &config.Config{
		OnlyUser: "",
		Format:   "raw",
		DumpFile: "/tmp/test_raw.sql",
	}

	// Mock user query
	mock.ExpectQuery("SELECT user, host FROM mysql.user WHERE user NOT IN").
		WillReturnRows(sqlmock.NewRows([]string{"user", "host"}).
			AddRow("testuser", "%"))

	// Mock SHOW CREATE USER
	mock.ExpectQuery("SHOW CREATE USER `testuser`@`%`").
		WillReturnRows(sqlmock.NewRows([]string{"Create User"}).
			AddRow("CREATE USER `testuser`@`%` IDENTIFIED WITH 'caching_sha2_password' AS '$A$005$hash' REQUIRE NONE"))

	// Mock SHOW GRANTS
	mock.ExpectQuery("SHOW GRANTS FOR `testuser`@`%`").
		WillReturnRows(sqlmock.NewRows([]string{"Grants"}).
			AddRow("GRANT USAGE ON *.* TO `testuser`@`%`"))

	ctx := context.Background()
	err = DumpUserAccounts(ctx, db, cfg)
	assert.NoError(t, err)

	// Check file content
	data, err := os.ReadFile(cfg.DumpFile)
	assert.NoError(t, err)
	expected := "SHOW CREATE USER `testuser`@`%`; SHOW GRANTS FOR `testuser`@`%`;\n"
	assert.Equal(t, expected, string(data))

	os.Remove(cfg.DumpFile)
}

func TestDumpUserAccounts_Import(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	cfg := &config.Config{
		OnlyUser: "",
		Format:   "import",
		DumpFile: "/tmp/test_import.sql",
	}

	// Mock user query
	mock.ExpectQuery("SELECT user, host FROM mysql.user WHERE user NOT IN").
		WillReturnRows(sqlmock.NewRows([]string{"user", "host"}).
			AddRow("flyway", "%"))

	// Mock SET
	mock.ExpectExec("SET print_identified_with_as_hex = 1").
		WillReturnResult(sqlmock.NewResult(0, 0))

	// Mock SHOW CREATE USER (with hex)
	mock.ExpectQuery("SHOW CREATE USER `flyway`@`%`").
		WillReturnRows(sqlmock.NewRows([]string{"Create User"}).
			AddRow("CREATE USER `flyway`@`%` IDENTIFIED WITH 'caching_sha2_password' AS 0x244124303035240A2B5D1718083E295E5D03126644062C6829654E793531634B6C6C55355452656246575576492F55703576633058307A5856595A4B4B4F51774B6C52556438 REQUIRE NONE PASSWORD EXPIRE DEFAULT ACCOUNT UNLOCK PASSWORD HISTORY DEFAULT PASSWORD REUSE INTERVAL DEFAULT PASSWORD REQUIRE CURRENT DEFAULT"))

	// Mock SHOW GRANTS
	mock.ExpectQuery("SHOW GRANTS FOR `flyway`@`%`").
		WillReturnRows(sqlmock.NewRows([]string{"Grants"}).
			AddRow("GRANT SELECT, INSERT, UPDATE, DELETE, CREATE, DROP, RELOAD, SHUTDOWN, PROCESS, FILE, REFERENCES, INDEX, ALTER, SHOW DATABASES, SUPER, CREATE TEMPORARY TABLES, LOCK TABLES, EXECUTE, REPLICATION SLAVE, REPLICATION CLIENT, CREATE VIEW, SHOW VIEW, CREATE ROUTINE, ALTER ROUTINE, CREATE USER, EVENT, TRIGGER, CREATE TABLESPACE, CREATE ROLE, DROP ROLE ON *.* TO `flyway`@`%`"))

	mock.ExpectQuery("SHOW GRANTS FOR `flyway`@`%`").
		WillReturnRows(sqlmock.NewRows([]string{"Grants"}).
			AddRow("GRANT APPLICATION_PASSWORD_ADMIN,AUDIT_ABORT_EXEMPT,AUDIT_ADMIN,AUTHENTICATION_POLICY_ADMIN,BACKUP_ADMIN,BINLOG_ADMIN,BINLOG_ENCRYPTION_ADMIN,CLONE_ADMIN,CONNECTION_ADMIN,ENCRYPTION_KEY_ADMIN,FIREWALL_EXEMPT,FLUSH_OPTIMIZER_COSTS,FLUSH_STATUS,FLUSH_TABLES,FLUSH_USER_RESOURCES,GROUP_REPLICATION_ADMIN,GROUP_REPLICATION_STREAM,INNODB_REDO_LOG_ARCHIVE,INNODB_REDO_LOG_ENABLE,PASSWORDLESS_USER_ADMIN,PERSIST_RO_VARIABLES_ADMIN,REPLICATION_APPLIER,REPLICATION_SLAVE_ADMIN,RESOURCE_GROUP_ADMIN,RESOURCE_GROUP_USER,ROLE_ADMIN,SENSITIVE_VARIABLES_OBSERVER,SERVICE_CONNECTION_ADMIN,SESSION_VARIABLES_ADMIN,SET_USER_ID,SHOW_ROUTINE,SYSTEM_USER,SYSTEM_VARIABLES_ADMIN,TABLE_ENCRYPTION_ADMIN,TELEMETRY_LOG_ADMIN,XA_RECOVER_ADMIN ON *.* TO `flyway`@`%`"))

	ctx := context.Background()
	err = DumpUserAccounts(ctx, db, cfg)
	assert.NoError(t, err)

	// Check file content
	data, err := os.ReadFile(cfg.DumpFile)
	assert.NoError(t, err)
	content := string(data)
	assert.Contains(t, content, "-- CREATE USER IF NOT EXISTS for flyway@%:")
	assert.Contains(t, content, "CREATE USER IF NOT EXISTS `flyway`@`%` IDENTIFIED WITH 'caching_sha2_password' AS 0x244124303035240A2B5D1718083E295E5D03126644062C6829654E793531634B6C6C55355452656246575576492F55703576633058307A5856595A4B4B4F51774B6C52556438")
	assert.Contains(t, content, "GRANT SELECT, INSERT, UPDATE, DELETE, CREATE, DROP, RELOAD, SHUTDOWN, PROCESS, FILE, REFERENCES, INDEX, ALTER, SHOW DATABASES, SUPER, CREATE TEMPORARY TABLES, LOCK TABLES, EXECUTE, REPLICATION SLAVE, REPLICATION CLIENT, CREATE VIEW, SHOW VIEW, CREATE ROUTINE, ALTER ROUTINE, CREATE USER, EVENT, TRIGGER, CREATE TABLESPACE, CREATE ROLE, DROP ROLE ON *.* TO `flyway`@`%`;")

	os.Remove(cfg.DumpFile)
}

func TestDumpUserAccounts_PtLike(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	cfg := &config.Config{
		OnlyUser: "",
		Format:   "pt-like",
		DumpFile: "/tmp/test_ptlike.sql",
	}

	// Mock user query
	mock.ExpectQuery("SELECT user, host FROM mysql.user WHERE user NOT IN").
		WillReturnRows(sqlmock.NewRows([]string{"user", "host"}).
			AddRow("flyway", "%"))

	// Mock SET
	mock.ExpectExec("SET print_identified_with_as_hex = 1").
		WillReturnResult(sqlmock.NewResult(0, 0))

	// Mock SHOW CREATE USER (with hex)
	mock.ExpectQuery("SHOW CREATE USER `flyway`@`%`").
		WillReturnRows(sqlmock.NewRows([]string{"Create User"}).
			AddRow("CREATE USER `flyway`@`%` IDENTIFIED WITH 'caching_sha2_password' AS 0x244124303035240A2B5D1718083E295E5D03126644062C6829654E793531634B6C6C55355452656246575576492F55703576633058307A5856595A4B4B4F51774B6C52556438 REQUIRE NONE PASSWORD EXPIRE DEFAULT ACCOUNT UNLOCK PASSWORD HISTORY DEFAULT PASSWORD REUSE INTERVAL DEFAULT PASSWORD REQUIRE CURRENT DEFAULT"))

	// Mock SHOW GRANTS
	mock.ExpectQuery("SHOW GRANTS FOR `flyway`@`%`").
		WillReturnRows(sqlmock.NewRows([]string{"Grants"}).
			AddRow("GRANT SELECT, INSERT, UPDATE, DELETE, CREATE, DROP, RELOAD, SHUTDOWN, PROCESS, FILE, REFERENCES, INDEX, ALTER, SHOW DATABASES, SUPER, CREATE TEMPORARY TABLES, LOCK TABLES, EXECUTE, REPLICATION SLAVE, REPLICATION CLIENT, CREATE VIEW, SHOW VIEW, CREATE ROUTINE, ALTER ROUTINE, CREATE USER, EVENT, TRIGGER, CREATE TABLESPACE, CREATE ROLE, DROP ROLE ON *.* TO `flyway`@`%`"))

	mock.ExpectQuery("SHOW GRANTS FOR `flyway`@`%`").
		WillReturnRows(sqlmock.NewRows([]string{"Grants"}).
			AddRow("GRANT APPLICATION_PASSWORD_ADMIN,AUDIT_ABORT_EXEMPT,AUDIT_ADMIN,AUTHENTICATION_POLICY_ADMIN,BACKUP_ADMIN,BINLOG_ADMIN,BINLOG_ENCRYPTION_ADMIN,CLONE_ADMIN,CONNECTION_ADMIN,ENCRYPTION_KEY_ADMIN,FIREWALL_EXEMPT,FLUSH_OPTIMIZER_COSTS,FLUSH_STATUS,FLUSH_TABLES,FLUSH_USER_RESOURCES,GROUP_REPLICATION_ADMIN,GROUP_REPLICATION_STREAM,INNODB_REDO_LOG_ARCHIVE,INNODB_REDO_LOG_ENABLE,PASSWORDLESS_USER_ADMIN,PERSIST_RO_VARIABLES_ADMIN,REPLICATION_APPLIER,REPLICATION_SLAVE_ADMIN,RESOURCE_GROUP_ADMIN,RESOURCE_GROUP_USER,ROLE_ADMIN,SENSITIVE_VARIABLES_OBSERVER,SERVICE_CONNECTION_ADMIN,SESSION_VARIABLES_ADMIN,SET_USER_ID,SHOW_ROUTINE,SYSTEM_USER,SYSTEM_VARIABLES_ADMIN,TABLE_ENCRYPTION_ADMIN,TELEMETRY_LOG_ADMIN,XA_RECOVER_ADMIN ON *.* TO `flyway`@`%`"))

	ctx := context.Background()
	err = DumpUserAccounts(ctx, db, cfg)
	assert.NoError(t, err)

	// Check file content
	data, err := os.ReadFile(cfg.DumpFile)
	assert.NoError(t, err)
	content := string(data)
	assert.Contains(t, content, "-- Grants dumped by go-pass")
	assert.Contains(t, content, "-- Grants for 'flyway'@'%'")
	assert.Contains(t, content, "CREATE USER IF NOT EXISTS `flyway`@`%`;")
	assert.Contains(t, content, "ALTER USER `flyway`@`%` IDENTIFIED WITH 'caching_sha2_password' AS 0x244124303035240A2B5D1718083E295E5D03126644062C6829654E793531634B6C6C55355452656246575576492F55703576633058307A5856595A4B4B4F51774B6C52556438")
	assert.Contains(t, content, "GRANT SELECT, INSERT, UPDATE, DELETE, CREATE, DROP, RELOAD, SHUTDOWN, PROCESS, FILE, REFERENCES, INDEX, ALTER, SHOW DATABASES, SUPER, CREATE TEMPORARY TABLES, LOCK TABLES, EXECUTE, REPLICATION SLAVE, REPLICATION CLIENT, CREATE VIEW, SHOW VIEW, CREATE ROUTINE, ALTER ROUTINE, CREATE USER, EVENT, TRIGGER, CREATE TABLESPACE, CREATE ROLE, DROP ROLE ON *.* TO `flyway`@`%`;")

	os.Remove(cfg.DumpFile)
}

func TestDumpUserAccounts_OnlyUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	cfg := &config.Config{
		OnlyUser: "specificuser",
		Format:   "raw",
		DumpFile: "/tmp/test_only.sql",
	}

	// Mock user query with only user
	mock.ExpectQuery("SELECT user, host FROM mysql.user WHERE user = ?").
		WithArgs("specificuser").
		WillReturnRows(sqlmock.NewRows([]string{"user", "host"}).
			AddRow("specificuser", "localhost"))

	// Mock SHOW CREATE USER
	mock.ExpectQuery("SHOW CREATE USER `specificuser`@`localhost`").
		WillReturnRows(sqlmock.NewRows([]string{"Create User"}).
			AddRow("CREATE USER `specificuser`@`localhost` IDENTIFIED WITH 'mysql_native_password' AS '*HASH'"))

	// Mock SHOW GRANTS
	mock.ExpectQuery("SHOW GRANTS FOR `specificuser`@`localhost`").
		WillReturnRows(sqlmock.NewRows([]string{"Grants"}).
			AddRow("GRANT ALL PRIVILEGES ON *.* TO `specificuser`@`localhost` WITH GRANT OPTION"))

	ctx := context.Background()
	err = DumpUserAccounts(ctx, db, cfg)
	assert.NoError(t, err)

	// Check file content
	data, err := os.ReadFile(cfg.DumpFile)
	assert.NoError(t, err)
	expected := "SHOW CREATE USER `specificuser`@`localhost`; SHOW GRANTS FOR `specificuser`@`localhost`;\n"
	assert.Equal(t, expected, string(data))

	os.Remove(cfg.DumpFile)
}
