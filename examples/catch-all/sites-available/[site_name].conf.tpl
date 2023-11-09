server {
    listen 80 default_server;
    listen 443 ssl default_server;
    include  /etc/nginx/conf.d/{{ .site_name }}/ssl-options.conf;
    server_name  _;
    location / {
        root  /var/www/{{ .site_name }}/live;
        index  index.html;
    }
}
