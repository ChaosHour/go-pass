# go-pass
Testing MySQL 8 caching_sha2_password and the system variable print_identified_with_as_hex


## The why?
In older versions of MySQL, you used to be able to use the password() funtion and use that hash for scripts, Ansible and what not.
You can't anymore and I wanted to see what I could do or what has been done.


## Testing Environment
- Docker run `docker run -d --name ps -d -p 3306:3306/tcp  -e MYSQL_ROOT_PASSWORD=root percona/percona-server:8.0.32-24`
- Percona-Server `8.0.32-24`

## reference
- https://dev.mysql.com/doc/refman/8.0/en/caching-sha2-pluggable-authentication.html
- https://dev.mysql.com/doc/refman/8.0/en/system-variables.html
- https://dev.mysql.com/doc/refman/8.0/en/caching-sha2-pluggable-authentication.html#caching-sha2-pluggable-authentication-password-hashing
- https://dev.mysql.com/doc/refman/8.0/en/system-variables.html#sysvar_print_identified_with_as_hex


Author of the bug: `Simon Mudd` https://bugs.mysql.com/bug.php?id=98732

```Go
klarsen@Mac-Book-Pro2 go-pass % go run main.go -s 10.8.0.10 -f createUsers.sql
2023/05/22 19:12:12 [+] Connecting to database: root:root@tcp(10.8.0.10:3306)/mysql
[+] Dumping user accounts to file: createUsers.sql
CREATE USER for root@%: CREATE USER `root`@`%` IDENTIFIED WITH 'caching_sha2_password' AS 0x244124303035240E6201545B641F35231C1D280A6F64537B7B25504945564C30355659705643717A6E544442584A61463850354F6E692E5137694D4F5959376E535051554D32 REQUIRE NONE PASSWORD EXPIRE DEFAULT ACCOUNT UNLOCK PASSWORD HISTORY DEFAULT PASSWORD REUSE INTERVAL DEFAULT PASSWORD REQUIRE CURRENT DEFAULT
CREATE USER for root@localhost: CREATE USER `root`@`localhost` IDENTIFIED WITH 'caching_sha2_password' AS 0x244124303035241D322B704A304A472D6C39023F365E040E32011E6A51326867414341473061447250722F74504F754E44434F5376744C4238573651727073536E6F50427742 REQUIRE NONE PASSWORD EXPIRE DEFAULT ACCOUNT UNLOCK PASSWORD HISTORY DEFAULT PASSWORD REUSE INTERVAL DEFAULT PASSWORD REQUIRE CURRENT DEFAULT
```
