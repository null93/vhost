
upstream fastcgi_backend_{{ .site_name }} {
   server unix:/var/run/php/php8.2-fpm.sock;
}

server {
    listen 80;
    listen 443 ssl;
    include /etc/nginx/conf.d/{{ .site_name }}/*.conf;
    server_name  {{ .domain_names }};
    set $MAGE_ROOT /var/www/{{ .site_name }}/live;
}
