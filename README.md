# Proposal — NGINX
> NGINX config structure for provisioning virtual hosts with a CLI tool

### Try It Out

```
rm ./conf/nginx/sites-*/*.conf
docker-compose up -d
docker-compose exec nginx bash
vhost create catch-all default.localhost
vhost create lemp m2.howtospeedupmagento.com application=magento-2
vhost create lemp m1.howtospeedupmagento.com application=magento-1 with-varnish=no
vhost create lemp wp.howtospeedupmagento.com application=wordpress with-varnish=no php-version=8.0
nginx -s reload
```
