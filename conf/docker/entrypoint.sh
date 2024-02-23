#!/bin/sh

set -e

# Install vhost
desired_arch=arm64
if [ `arch | xargs` != "aarch64" ]; then desired_arch="amd64"; fi
apt-get -y -qq install /usr/local/dist/vhost_*$desired_arch.deb

# Start php-fpm in background
php-fpm8.2 -D

exit 0