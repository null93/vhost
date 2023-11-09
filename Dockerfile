FROM nginx:latest

RUN true \
    && mkdir -p /etc/nginx/sites-enabled \
    && mkdir -p /etc/nginx/sites-available \
    && apt-get -y -qq update \
    && apt-get -y -qq install tree vim php-fpm php-mysqli

COPY ./conf/nginx/nginx.conf /etc/nginx/nginx.conf
COPY ./conf/docker/entrypoint.sh /docker-entrypoint.d/setup-environment.sh
