---
application:      { pattern: ^magento-1|magento-2|wordpress|drupal$, description: what type of php application will you use }
php-version:      { value: 8.0, pattern: <php-version>, description: what php version do you want to use }
with-varnish:     { value: yes, pattern: <yes-no>, description: do you want to use varnish }
varnish-endpoint: { value: varnish:80, pattern: <endpoint>, description: if you use varnish, you can pass the host:port to it }
---

server {
    listen 80;
    listen [::]:80;
    server_name .{{{ domain }}};
    return 301 https://$server_name$request_uri;
}

server {
    listen 8080;
    listen [::]:8080;
    server_name .{{{ domain }}};
    set $DOMAIN_NAME {{{ domain }}};
    set $PHP_VERSION {{{ php-version }}};
    include /etc/nginx/conf.d/applications/{{{ application }}}.conf;
    access_log /home/jetrails/{{{ domain }}}/logs/{{{ domain }}}-access_log;
    error_log /home/jetrails/{{{ domain }}}/logs/{{{ domain }}}-error_log;
}

server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    ssl_certificate      /home/jetrails/{{{ domain }}}/conf/ssl/{{{ domain }}}.crt;
    ssl_certificate_key  /home/jetrails/{{{ domain }}}/conf/ssl/{{{ domain }}}.key;
    include /etc/nginx/conf.d/mixins/ssl-options.conf;
    server_name .{{{ domain }}};
    root /home/jetrails/{{{ domain }}}/html;
    access_log /home/jetrails/{{{ domain }}}/logs/{{{ domain }}}-ssl-access_log;
    error_log /home/jetrails/{{{ domain }}}/logs/{{{ domain }}}-ssl-error_log;

    set $WITH_VARNISH {{{ with-varnish }}};
    set $VARNISH_ENDPOINT {{{ varnish-endpoint }}};

    if ( $WITH_VARNISH = "no" ) {
        set $VARNISH_ENDPOINT 127.0.0.1:8080;
    }

    location /jetrails/varnish-config/ {
        deny all;
    }
    location / {
        include /etc/nginx/conf.d/mixins/cors.conf;
        limit_except GET POST PUT DELETE HEAD OPTIONS { deny  all; }
        resolver 127.0.0.11;
        proxy_pass http://$VARNISH_ENDPOINT;
        proxy_set_header X-Real-IP  $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto https;
        proxy_set_header X-Forwarded-Port 443;
        proxy_set_header Host $host;
        proxy_buffers 1024 8k;
        proxy_buffer_size 256k;
        proxy_busy_buffers_size 256k;
    }
}