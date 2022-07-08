---
---

server {
    listen 80 default_server;
    listen 443 ssl default_server;
    ssl_certificate  /home/jetrails/{{{ domain }}}/conf/ssl/{{{ domain }}}.crt;
    ssl_certificate_key  /home/jetrails/{{{ domain }}}/conf/ssl/{{{ domain }}}.key;
    include  /etc/nginx/conf.d/mixins/ssl-options.conf;
    include  /etc/nginx/conf.d/mixins/cors.conf;
    server_name  _;
    location / {
        root  /home/jetrails/{{{ domain }}}/public;
        index  index.html;
    }
}
