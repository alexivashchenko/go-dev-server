
#user  nobody;
worker_processes  1;

error_log {root_folder}logs/nginx/error.log;
#error_log {root_folder}logs/nginx/error.log  notice;
#error_log {root_folder}logs/nginx/error.log  info;

pid        {root_folder}logs/nginx/nginx.pid;


events {
    worker_connections  1024;
}


http {
    include       mime.types;
    default_type  application/octet-stream;

    #log_format  main  '$remote_addr - $remote_user [$time_local] "$request" '
    #                  '$status $body_bytes_sent "$http_referer" '
    #                  '"$http_user_agent" "$http_x_forwarded_for"';

	access_log {root_folder}logs/nginx/access.log;


    sendfile        on;
    #tcp_nopush     on;

    #keepalive_timeout  0;
    keepalive_timeout  65;

    #gzip  on;

    include "{root_folder}etc/nginx/php_upstream.conf";
    include "{root_folder}etc/nginx/sites-enabled/*.conf";
    client_max_body_size 2000M;
	server_names_hash_bucket_size 64;
}
