server {
    listen 80 default_server;
    listen 443 ssl default_server;
    server_name localhost;
    root "{root_folder}www/";

    index index.html index.htm index.php;

    # Access Restrictions
    allow       127.0.0.1;
    deny        all;

    include "{root_folder}etc/nginx/alias/*.conf";

    location / {
        try_files $uri $uri/ =404;
		autoindex on;
    }

    location ~ \.php$ {
        include snippets/fastcgi-php.conf;
        fastcgi_pass php_upstream;
        #fastcgi_pass unix:/run/php/php7.0-fpm.sock;
    }

    ssl_certificate "{root_folder}etc/ssl/certificate.crt";
    ssl_certificate_key "{root_folder}etc/ssl/private.key";
    ssl_session_timeout 5m;
    ssl_protocols TLSv1 TLSv1.1 TLSv1.2;
    ssl_ciphers ALL:!ADH:!EXPORT56:RC4+RSA:+HIGH:+MEDIUM:+LOW:+SSLv3:+EXP;
    ssl_prefer_server_ciphers on;


    charset utf-8;

    location = /favicon.ico { access_log off; log_not_found off; }
    location = /robots.txt  { access_log off; log_not_found off; }
    location ~ /\.ht {
        deny all;
    }

}
