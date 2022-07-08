# Proposal — NGINX
> NGINX config structure for provisioning virtual hosts with a CLI tool

### Try It Out

```
docker-compose up -d
rm ./conf/nginx/sites-*/*.conf
docker-compose exec nginx bash
vhost create catch-all default.localhost
vhost create lemp m2.howtospeedupmagento.com application=magento-2
vhost create lemp m1.howtospeedupmagento.com application=magento-1 with-varnish=no
vhost create lemp wp.howtospeedupmagento.com application=wordpress with-varnish=no php-version=8.0
nginx -s reload
```

### Possible Future Commands

```
vhost create TEMPLATE DOMAIN application=magento-1 php-version=7.2
vhost delete DOMAIN
vhost upgrade DOMAIN application=magento-1 php-version=8.0
vhost modify DOMAIN php-version=8.0
vhost enable DOMAIN
vhost disable DOMAIN
vhost list
vhost info DOMAIN

vhost template list
vhost template info TEMPLATE
```