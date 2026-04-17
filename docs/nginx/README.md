# Nginx

How the HTTP server is configured and why it is structured this way.

## Config file

- **Path in repo**: `nginx/default.conf`
- **Mounted in Docker**: `./nginx/default.conf` → `/etc/nginx/conf.d/default.conf` (read-only). Nginx loads all `.conf` from `conf.d`; this file is the only server block for this app.

## Server block overview

- **Listen**: Port 80 inside the container; host maps `8080:80` in Docker Compose.
- **Server name**: `_` (default server).
- **Root**: `/var/www/html/public` — the web root is `public/`; no direct access to project root or `src/`, `config/`, etc.

## Locations

### `location /`

- **Behavior**: `try_files /index.php?$query_string =404;` — every request that does not match a more specific location is passed to the front controller with the original query string.
- **Effect**: Clean URLs (e.g. `/blog`, `/about`) are handled by `public/index.php`; Flight routing then matches the path. No rewrite to `/index.php/...` is needed because `try_files` already sends everything to `index.php`.

### `location /static/`

- **Behavior**: `try_files $uri =404;` — serve the file from disk if it exists; otherwise 404.
- **Headers**: `Cache-Control: public, max-age=31536000, immutable` and `expires 1y` — long-term caching for static assets (CSS, JS, images under `public/static/`). No PHP involved.
- **Purpose**: Static assets are served directly by Nginx with aggressive caching; the PHP container is not used for these requests.

### `location /api/`

- **Behavior**: Proxies to PHP-FPM with `SCRIPT_FILENAME` set to `/var/www/html/public/index.php` and `SCRIPT_NAME` to `/index.php`. So all `/api/*` requests are handled by the same front controller; the app sees `REQUEST_URI` like `/api/health` or `/api/landing` and routes them in `config/routes.php`.
- **Purpose**: Explicit block for API routes so they always go through the front controller with a consistent script name. Optional: you could rely on `location /` only; this block makes API routing explicit and allows different FastCGI params or headers later if needed.

### `location ~ \.php$`

- **Behavior**: Only matches requests that end in `.php`. Proxies to PHP-FPM with `SCRIPT_FILENAME` set to the requested file under `/var/www/html/public`. So a request to `/index.php` or `/index.php/foo` is passed to `public/index.php`.
- **Purpose**: Ensures any direct request to a `.php` file in `public/` is executed via FPM. In practice, the app uses `location /` and does not expose other PHP files; this is a safety net and keeps Nginx’s PHP handling consistent.

## Security headers

- **X-Frame-Options**: `SAMEORIGIN` — reduce clickjacking risk.
- **X-Content-Type-Options**: `nosniff` — discourage MIME sniffing.

These are applied to all responses from this server block. The application can add more headers (e.g. CSP) via `SecurityHeaders` middleware for HTML or API responses.

## Relation to PHP built-in server

Locally you can run `php -S localhost:8080 -t public public/router.php`. That setup uses `public/router.php` to serve static files and forward everything else to `index.php`; it does not use Nginx. The Nginx config here is the **production-style** setup: single front controller, static files from `public/static/`, clean URLs via `try_files`, and PHP-FPM in a separate container. Behavior is aligned so the same app works in both environments.
