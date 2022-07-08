#!/usr/bin/env sh

set -e

mkdir -p /home/jetrails/$DOMAIN/logs
mkdir -p /home/jetrails/$DOMAIN/public
mkdir -p /home/jetrails/$DOMAIN/conf/ssl

openssl req \
    -x509 \
    -nodes \
    -newkey rsa:4096 \
    -keyout /home/jetrails/$DOMAIN/conf/ssl/$DOMAIN.key \
    -out /home/jetrails/$DOMAIN/conf/ssl/$DOMAIN.crt \
    -days 365 \
    -subj "/C=US/ST=Illinois/L=Chicago/O=JetRails/OU=IT/CN=$DOMAIN"

curl -Ls https://maintenance.jetrails.cloud > /home/jetrails/$DOMAIN/public/index.html