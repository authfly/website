# syntax=docker/dockerfile:1.7

FROM composer:2 AS vendor

WORKDIR /app

COPY composer.json composer.lock ./

RUN composer install \
    --no-dev \
    --no-interaction \
    --no-progress \
    --no-scripts \
    --prefer-dist \
    --no-autoloader

COPY config ./config
COPY public ./public
COPY scripts ./scripts
COPY src ./src
COPY templates ./templates

RUN composer dump-autoload \
    --no-dev \
    --classmap-authoritative \
    --optimize


FROM php:8.4-fpm-alpine AS runtime

RUN apk add --no-cache nginx su-exec \
    && apk add --no-cache --virtual .build-deps $PHPIZE_DEPS linux-headers \
    && pecl install apcu \
    && docker-php-ext-enable apcu opcache \
    && apk del .build-deps \
    && rm -rf /tmp/pear

ENV APP_ENV=production \
    APP_DEBUG=false \
    DATA_SOURCE=fixtures \
    CACHE_TTL=3600 \
    CACHE_DIR=/app/cache \
    FIXTURES_PATH=/app/fixtures \
    DEFENSE_CONFIG_PATH=/app/config/defense.php

WORKDIR /app

COPY composer.json composer.lock ./
COPY config ./config
COPY fixtures ./fixtures
COPY public ./public
COPY scripts ./scripts
COPY src ./src
COPY templates ./templates
COPY --from=vendor /app/vendor ./vendor

COPY php.ini /usr/local/etc/php/conf.d/zz-app.ini
COPY apcu.ini /usr/local/etc/php/conf.d/zz-apcu.ini
COPY docker/production/nginx/default.conf /etc/nginx/http.d/default.conf
COPY docker/production/entrypoint.sh /usr/local/bin/docker-entrypoint

RUN mkdir -p /app/cache /app/templates/cache /run/nginx /var/lib/nginx /var/log/nginx \
    && chown -R www-data:www-data /app/cache /app/templates/cache \
    && sed -i 's/\r$//' /usr/local/bin/docker-entrypoint \
    && chmod +x /usr/local/bin/docker-entrypoint

EXPOSE 80

HEALTHCHECK --interval=30s --timeout=5s --start-period=15s --retries=3 \
    CMD wget -q -O- http://127.0.0.1/api/health | grep -q '"status":"ok"' || exit 1

ENTRYPOINT ["docker-entrypoint"]
