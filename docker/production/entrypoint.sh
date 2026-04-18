#!/bin/sh
set -eu

mkdir -p /app/cache /app/templates/cache /run/nginx /var/lib/nginx /var/log/nginx

# Critical: /app/cache and /app/templates/cache are mounted as named
# docker volumes (see docker-compose.yml), so they survive container
# replacement on deploy. Without an explicit purge here, a new image
# would still serve stale rendered pages and stale compiled Latte
# templates from the previous version. Wipe both on every container
# start; rebuilding the cache costs <100ms per page on first hit.
find /app/cache -mindepth 1 -delete 2>/dev/null || true
find /app/templates/cache -mindepth 1 -delete 2>/dev/null || true

chown -R www-data:www-data /app/cache /app/templates/cache

case "${APP_WARM_CACHE:-0}" in
    1|true|TRUE|yes|YES)
        su-exec www-data php /app/scripts/build-static.php
        ;;
esac

php-fpm -D
exec nginx -g 'daemon off;'
