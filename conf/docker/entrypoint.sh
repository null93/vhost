#!/bin/sh

set -e

# Install vhost
apt-get -y -qq install /usr/local/dist/vhost*.deb

# Start php-fpm in background
php-fpm8.2 -D

exit 0