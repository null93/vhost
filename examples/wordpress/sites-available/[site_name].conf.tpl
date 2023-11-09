upstream fastcgi_backend_{{ .site_name }} {
   server unix:/var/run/php/php8.2-fpm.sock;
}

server {
    listen 80;
    listen 443 ssl;

    server_name {{ .domain_names }};
    root /var/www/{{ .site_name }}/live;
    index index.php;

    location = /favicon.ico {
        log_not_found off;
        access_log off;
    }

    location = /robots.txt {
        allow all;
        log_not_found off;
        access_log off;
    }

    location / {
        try_files $uri $uri/ /index.php?$args;
    }

    location ~ \.php$ {
        include fastcgi_params;
        fastcgi_intercept_errors on;
        fastcgi_pass fastcgi_backend_{{ .site_name }};
        fastcgi_param  SCRIPT_FILENAME $document_root$fastcgi_script_name;
    }

    location ~* \.(js|css|png|jpg|jpeg|gif|ico)$ {
        expires max;
        log_not_found off;
    }
}