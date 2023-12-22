# Proposal — NGINX

> NGINX config structure for provisioning virtual hosts with a CLI tool

### About

This is a proposal for a new NGINX config structure that allows for provisioning virtual hosts with a CLI tool.
Please keep in mind that this is a proof of concept and the actual config files in the example nginx config directory are not complete.
It is the minimum viable product to demonstrate the CLI tool rather than a complete NGINX config.

### Requirements

All you need is docker installed with the docker compose plugin.

### Try It Out

On your host machine, run the following commands:

```
docker-compose up -d
docker-compose exec nginx bash
```

Once you are inside the docker container, you can create some virtual hosts:

```
vhost create catch-all default-backend
vhost enable default-backend

vhost create wordpress my-blog domain_names=wordpress-127-0-0-1.nip.io
vhost enable my-blog

vhost create magento-2 my-store magento_version=2.4.6.3 domain_names=magento-127-0-0-1.nip.io
vhost enable my-store
```

Finally you can reload nginx to apply the changes:

```
nginx -s reload
```

That's it! You can now visit the following URLs:

```
http://localhost
http://wordpress-127-0-0-1.nip.io
http://magento-127-0-0-1.nip.io
```
