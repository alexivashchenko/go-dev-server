# Local Development Server for Windows

## Description

This is a simple development server made for Windows.
It runs with NGINX, MySQL, PHP.

## How to use it:

1. Build `server.exe` file
2. Drop site folder to www directory
3. Start server

### 1. Build server.exe file

`cd src`

`go build -o ../server.exe`

### 2. Drop site folder to www directory

`cd ..`

`mkdir www/site-1`

`touch www/site-1/index.php`

### 3. Start server

`./server start`

## Commands

### Start server

`./server start`

### Stop server

`./server stop`

### Restart server

`./server restart`


## Notes

Developed and tested for `C:\server\` path.

Different versions of MySQL, NGX and PHP can be used, - just drop new version to `apps/` related folder and update the `.env` file.

### PHP versions:

[PHP-8.4](https://windows.php.net/downloads/releases/archives/php-8.4.3-nts-Win32-vs17-x64.zip)

[PHP-8.3](https://windows.php.net/downloads/releases/archives/php-8.3.16-nts-Win32-vs16-x64.zip)

[PHP-8.2](https://windows.php.net/downloads/releases/archives/php-8.2.26-nts-Win32-vs16-x64.zip)

[PHP-8.1](https://windows.php.net/downloads/releases/archives/php-8.1.30-nts-Win32-vs16-x64.zip)

### NGINX versions:

[Nginx-1.27.4](https://nginx.org/download/nginx-1.27.4.zip)

### MySQL versions

[mysql-9.1](https://dev.mysql.com/get/Downloads/MySQL-9.1/mysql-9.1.0-winx64.zip)

[mysql-8.4](https://dev.mysql.com/get/Downloads/MySQL-8.4/mysql-8.4.3-winx64.zip)

[mysql-8.0](https://dev.mysql.com/get/Downloads/MySQL-8.0/mysql-8.0.40-winx64.zip)

[mysql-5.7](https://dev.mysql.com/get/Downloads/MySQL-5.7/mysql-5.7.39-winx64.zip)

