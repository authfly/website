#!/bin/sh
set -eu

mkdir -p /app/cache /app/templates/cache /run/nginx /var/lib/nginx /var/log/nginx
chown -R www-data:www-data /app/cache /app/templates/cache

case "${APP_WARM_CACHE:-0}" in
    1|true|TRUE|yes|YES)
        su-exec www-data php /app/scripts/build-static.php
        ;;
esac

php-fpm -D
exec nginx -g 'daemon off;'
