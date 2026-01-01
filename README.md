# go-pass

- caching_sha2_password
- variable print_identified_with_as_hex

## Why

In older versions of MySQL, you used to be able to use the password() funtion and use that hash for scripts, Ansible and what not.
You can't anymore and I wanted to see what I could do or what has been done.

Why not just use `pt-show-grants`

- <https://www.percona.com/doc/percona-toolkit/LATEST/pt-show-grants.html>

I just wanted to learn more about it and how to use it, plus I wanted to see how to do it with `Golang`.

## Project Structure

This project follows Go best practices:

- `cmd/pass/main.go`: Main application entry point
- `internal/config/`: Configuration handling (flags, MySQL credentials)
- `internal/database/`: Database operations (connection, dumping, querying)
- `examples/`: Example SQL output files for different formats
- `Makefile`: Build and development tasks

## Testing Environment

- Docker run `docker run -d --name ps -d -p 3306:3306/tcp  -e MYSQL_ROOT_PASSWORD=root percona/percona-server:8.0.32-24`
- Percona-Server `8.0.32-24`

## reference

- <https://dev.mysql.com/doc/refman/8.0/en/caching-sha2-pluggable-authentication.html>
- <https://dev.mysql.com/doc/refman/8.0/en/system-variables.html>
- <https://dev.mysql.com/doc/refman/8.0/en/caching-sha2-pluggable-authentication.html#caching-sha2-pluggable-authentication-password-hashing>
- <https://dev.mysql.com/doc/refman/8.0/en/system-variables.html#sysvar_print_identified_with_as_hex>

Author of the `Bug`: [Simon Mudd](https://github.com/sjmudd) <https://bugs.mysql.com/bug.php?id=98732>

## Usage

```bash
go build -o bin/go-pass ./cmd/pass
./bin/go-pass -h
```

Output:

```bash
Usage: go-pass -s <source host> -f <dump file>
Options:
  -s <source host>  Source MySQL host
  -f <dump file>    Output dump file
  -o <user>         Only dump the specified user
  --format <fmt>    Output format: raw, import, pt-like (default: raw)
  -h                Print this help
```

## Output Formats

go-pass supports three output formats controlled by the `--format` flag:

- `raw` (default): Outputs raw `SHOW CREATE USER` and `SHOW GRANTS` statements. This is not executable SQL but shows the queries that would be run to recreate users.
- `import`: Generates clean, executable SQL statements ready for import into another MySQL database. Includes `CREATE USER IF NOT EXISTS` and `GRANT` statements with comments for each user. The output can be piped directly to `mysql` for execution (e.g., `cat output.sql | mysql`). This format is ideal for migrating users between databases or creating backups that can be easily restored.
- `pt-like`: Mimics the output of Percona Toolkit's `pt-show-grants` tool. Splits user creation into separate `CREATE USER` and `ALTER USER` statements for better compatibility.

### Import Format

```bash
./bin/go-pass -s 127.0.0.1 -f import.sql -o flyway --format=import
```

Outputs clean SQL for import:

```sql
-- CREATE USER IF NOT EXISTS for flyway@%:
CREATE USER IF NOT EXISTS `flyway`@`%` IDENTIFIED WITH 'caching_sha2_password' AS 0x244124303035240A2B5D1718083E295E5D03126644062C6829654E793531634B6C6C55355452656246575576492F55703576633058307A5856595A4B4B4F51774B6C52556438 REQUIRE NONE PASSWORD EXPIRE DEFAULT ACCOUNT UNLOCK PASSWORD HISTORY DEFAULT PASSWORD REUSE INTERVAL DEFAULT PASSWORD REQUIRE CURRENT DEFAULT;
GRANT SELECT, INSERT, UPDATE, DELETE, CREATE, DROP, RELOAD, SHUTDOWN, PROCESS, FILE, REFERENCES, INDEX, ALTER, SHOW DATABASES, SUPER, CREATE TEMPORARY TABLES, LOCK TABLES, EXECUTE, REPLICATION SLAVE, REPLICATION CLIENT, CREATE VIEW, SHOW VIEW, CREATE ROUTINE, ALTER ROUTINE, CREATE USER, EVENT, TRIGGER, CREATE TABLESPACE, CREATE ROLE, DROP ROLE ON *.* TO `flyway`@`%`;
GRANT APPLICATION_PASSWORD_ADMIN,AUDIT_ABORT_EXEMPT,AUDIT_ADMIN,AUTHENTICATION_POLICY_ADMIN,BACKUP_ADMIN,BINLOG_ADMIN,BINLOG_ENCRYPTION_ADMIN,CLONE_ADMIN,CONNECTION_ADMIN,ENCRYPTION_KEY_ADMIN,FIREWALL_EXEMPT,FLUSH_OPTIMIZER_COSTS,FLUSH_STATUS,FLUSH_TABLES,FLUSH_USER_RESOURCES,GROUP_REPLICATION_ADMIN,GROUP_REPLICATION_STREAM,INNODB_REDO_LOG_ARCHIVE,INNODB_REDO_LOG_ENABLE,PASSWORDLESS_USER_ADMIN,PERSIST_RO_VARIABLES_ADMIN,REPLICATION_APPLIER,REPLICATION_SLAVE_ADMIN,RESOURCE_GROUP_ADMIN,RESOURCE_GROUP_USER,ROLE_ADMIN,SENSITIVE_VARIABLES_OBSERVER,SERVICE_CONNECTION_ADMIN,SESSION_VARIABLES_ADMIN,SET_USER_ID,SHOW_ROUTINE,SYSTEM_USER,SYSTEM_VARIABLES_ADMIN,TABLE_ENCRYPTION_ADMIN,TELEMETRY_LOG_ADMIN,XA_RECOVER_ADMIN ON *.* TO `flyway`@`%`;
```

### PT-Like Format

```bash
./bin/go-pass -s 127.0.0.1 -f pt-like.sql -o flyway --format=pt-like
```

Outputs in pt-show-grants style:

```sql
-- Grants dumped by go-pass
-- Dumped from server 127.0.0.1 via TCP/IP, MySQL at 2026-01-01 09:30:54
-- Grants for 'flyway'@'%'
CREATE USER IF NOT EXISTS `flyway`@`%`;
ALTER USER `flyway`@`%` IDENTIFIED WITH 'caching_sha2_password' AS 0x244124303035240A2B5D1718083E295E5D03126644062C6829654E793531634B6C6C55355452656246575576492F55703576633058307A5856595A4B4B4F51774B6C52556438 REQUIRE NONE PASSWORD EXPIRE DEFAULT ACCOUNT UNLOCK PASSWORD HISTORY DEFAULT PASSWORD REUSE INTERVAL DEFAULT PASSWORD REQUIRE CURRENT DEFAULT;
GRANT SELECT, INSERT, UPDATE, DELETE, CREATE, DROP, RELOAD, SHUTDOWN, PROCESS, FILE, REFERENCES, INDEX, ALTER, SHOW DATABASES, SUPER, CREATE TEMPORARY TABLES, LOCK TABLES, EXECUTE, REPLICATION SLAVE, REPLICATION CLIENT, CREATE VIEW, SHOW VIEW, CREATE ROUTINE, ALTER ROUTINE, CREATE USER, EVENT, TRIGGER, CREATE TABLESPACE, CREATE ROLE, DROP ROLE ON *.* TO `flyway`@`%`;
GRANT APPLICATION_PASSWORD_ADMIN,AUDIT_ABORT_EXEMPT,AUDIT_ADMIN,AUTHENTICATION_POLICY_ADMIN,BACKUP_ADMIN,BINLOG_ADMIN,BINLOG_ENCRYPTION_ADMIN,CLONE_ADMIN,CONNECTION_ADMIN,ENCRYPTION_KEY_ADMIN,FIREWALL_EXEMPT,FLUSH_OPTIMIZER_COSTS,FLUSH_STATUS,FLUSH_TABLES,FLUSH_USER_RESOURCES,GROUP_REPLICATION_ADMIN,GROUP_REPLICATION_STREAM,INNODB_REDO_LOG_ARCHIVE,INNODB_REDO_LOG_ENABLE,PASSWORDLESS_USER_ADMIN,PERSIST_RO_VARIABLES_ADMIN,REPLICATION_APPLIER,REPLICATION_SLAVE_ADMIN,RESOURCE_GROUP_ADMIN,RESOURCE_GROUP_USER,ROLE_ADMIN,SENSITIVE_VARIABLES_OBSERVER,SERVICE_CONNECTION_ADMIN,SESSION_VARIABLES_ADMIN,SET_USER_ID,SHOW_ROUTINE,SYSTEM_USER,SYSTEM_VARIABLES_ADMIN,TABLE_ENCRYPTION_ADMIN,TELEMETRY_LOG_ADMIN,XA_RECOVER_ADMIN ON *.* TO `flyway`@`%`;
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details
