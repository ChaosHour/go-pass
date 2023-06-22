# go-pass
- caching_sha2_password
- variable print_identified_with_as_hex 


## The why?
In older versions of MySQL, you used to be able to use the password() funtion and use that hash for scripts, Ansible and what not.
You can't anymore and I wanted to see what I could do or what has been done.

Why not just use `pt-show-grants` 
- https://www.percona.com/doc/percona-toolkit/LATEST/pt-show-grants.html

I just wanted to learn more about it and how to use it, plus I wanted to see how to do it with `Golang`.


## Testing Environment
- Docker run `docker run -d --name ps -d -p 3306:3306/tcp  -e MYSQL_ROOT_PASSWORD=root percona/percona-server:8.0.32-24`
- Percona-Server `8.0.32-24`

## reference
- https://dev.mysql.com/doc/refman/8.0/en/caching-sha2-pluggable-authentication.html
- https://dev.mysql.com/doc/refman/8.0/en/system-variables.html
- https://dev.mysql.com/doc/refman/8.0/en/caching-sha2-pluggable-authentication.html#caching-sha2-pluggable-authentication-password-hashing
- https://dev.mysql.com/doc/refman/8.0/en/system-variables.html#sysvar_print_identified_with_as_hex


Author of the `Bug`: [Simon Mudd](https://github.com/sjmudd)  https://bugs.mysql.com/bug.php?id=98732


## Usage
```Go
Usage: ./go-pass -s < source host> -f <dump file>"
  -f string
        Dump file
  -s string
        Source host
```


```Go
go-pass -s 10.8.0.15 -f show_users.sql 
2023/06/21 17:12:20 [+] Connecting to database: root:root@tcp(10.8.0.15:3306)/mysql
[+] Dumping user accounts to file: show_users.sql
-- CREATE USER IF NOT EXISTS for root@%: 
 CREATE USER `root`@`%` IDENTIFIED WITH 'caching_sha2_password' AS 0x24412430303524542E705C456F693A4E034D541F791E5E3264236E6E61724A71316A6654594667564661444F4777506862534A7A6653342E307677446A6E526F55656F685A36 REQUIRE NONE PASSWORD EXPIRE DEFAULT ACCOUNT UNLOCK PASSWORD HISTORY DEFAULT PASSWORD REUSE INTERVAL DEFAULT PASSWORD REQUIRE CURRENT DEFAULT;
-- CREATE USER IF NOT EXISTS for root@localhost: 
 CREATE USER `root`@`localhost` IDENTIFIED WITH 'caching_sha2_password' AS 0x244124303035240566230F3279056A495A7870484E424E62780318336A62674D71524F4F5A482E7255497738324874337953795268676878666345494556586B633471416530 REQUIRE NONE PASSWORD EXPIRE DEFAULT ACCOUNT UNLOCK PASSWORD HISTORY DEFAULT PASSWORD REUSE INTERVAL DEFAULT PASSWORD REQUIRE CURRENT DEFAULT;
```
