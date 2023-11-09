#!/usr/bin/env bash

set -e

echo "creating directories under /var/www/$SITE_NAME"
timestamp=`date '+%Y-%m-%d-001'`
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

echo "downloading release"
curl -Ls -o wordpress.tar.gz https://wordpress.org/latest.tar.gz

echo "extracting release tar"
tar -xf wordpress.tar.gz -C /var/www/$SITE_NAME/releases/$timestamp --strip-components=1

echo "deleting tar file"
rm wordpress.tar.gz

echo "changing file ownership"
chown :www-data -R /var/www/$SITE_NAME/releases/$timestamp

echo "/var/www/$SITE_NAME is done provisioning"
