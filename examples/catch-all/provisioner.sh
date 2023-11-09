#!/usr/bin/env sh

set -e

timestamp=`date '+%Y-%m-%d-001'`

mkdir -p /var/www/$SITE_NAME/logs
mkdir -p /var/www/$SITE_NAME/releases/$timestamp
mkdir -p /var/www/$SITE_NAME/conf/ssl

openssl req \
    -x509 \
    -nodes \
    -newkey rsa:4096 \
    -keyout /var/www/$SITE_NAME/conf/ssl/$SITE_NAME.key \
    -out /var/www/$SITE_NAME/conf/ssl/$SITE_NAME.crt \
    -days 365 \
    -subj "/C=US/ST=Illinois/L=Chicago/O=JetRails/OU=IT/CN=$SITE_NAME" \
    2> /dev/null

cp assets/index.html /var/www/$SITE_NAME/releases/$timestamp
ln -s ./releases/$timestamp /var/www/$SITE_NAME/live
