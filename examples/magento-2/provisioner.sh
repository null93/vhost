#!/usr/bin/env bash

set -e

echo "creating directories under /var/www/$SITE_NAME"
timestamp=`date '+%Y-%m-%d-001'`
mkdir -p /var/www/$SITE_NAME/shared
mkdir -p /var/www/$SITE_NAME/releases/$timestamp
mkdir -p /var/www/$SITE_NAME/repo
mkdir -p /var/www/$SITE_NAME/logs
mkdir -p /var/www/$SITE_NAME/conf/ssl
ln -s ./releases/$timestamp /var/www/$SITE_NAME/live

echo "generating self-signed ssl certificate"
openssl req \
    -x509 \
    -nodes \
    -newkey rsa:4096 \
    -keyout /var/www/$SITE_NAME/conf/ssl/$SITE_NAME.key \
    -out /var/www/$SITE_NAME/conf/ssl/$SITE_NAME.crt \
    -days 365 \
    -subj "/C=US/ST=Illinois/L=Chicago/O=JetRails/OU=IT/CN=$SITE_NAME" \
    2> /dev/null

if [[ "$DOWNLOAD" == "yes" ]]; then
    echo "downloading release"
    curl -Ls -o magento.tar.gz https://api.magerepo.com/download/release/community/$MAGENTO_VERSION

    echo "extracting release tar"
    tar -xf magento.tar.gz -C /var/www/$SITE_NAME/releases/$timestamp --strip-components=1

    echo "deleting tar file"
    rm magento.tar.gz

    echo "moving shared directories out of release"
    mv /var/www/$SITE_NAME/releases/$timestamp/var /var/www/$SITE_NAME/shared/var
    mv /var/www/$SITE_NAME/releases/$timestamp/pub/media /var/www/$SITE_NAME/shared/media

    echo "creating symlink for shared directories"
    ln -s /var/www/$SITE_NAME/shared/var /var/www/$SITE_NAME/releases/$timestamp/var
    ln -s /var/www/$SITE_NAME/shared/media /var/www/$SITE_NAME/releases/$timestamp/pub/media

    echo "changing file ownership"
    chown :www-data -R /var/www/$SITE_NAME/releases/$timestamp

    echo "copying nginx configuration"
    mkdir -p /etc/nginx/conf.d/$SITE_NAME
    cp /var/www/$SITE_NAME/releases/$timestamp/nginx.conf.sample /etc/nginx/conf.d/$SITE_NAME/magento.conf
    sed -i "s/fastcgi_backend/fastcgi_backend_${SITE_NAME}/g" /etc/nginx/conf.d/$SITE_NAME/magento.conf
fi

echo "/var/www/$SITE_NAME is done provisioning"