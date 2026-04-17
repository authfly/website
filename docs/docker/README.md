# Docker setup and project preparation

How to prepare the project for Docker, which images and configs are used, and why.

## Prerequisites

- Docker and Docker Compose
- PHP 8.4+ and Composer for local development (optional if you only run in Docker)

## Quick start

```bash
cp .env.example .env
# Edit .env if needed (PHP_INI_FILE, APCU_INI_FILE are optional)
docker compose -f docker-compose.local.yml up --build
```

Then open **http://localhost:8080**. The app is served by Nginx; PHP runs as PHP-FPM in a separate container.

## Why `webdevops/php:8.4-alpine`

- **PHP 8.4**: Matches the project’s required runtime; allows use of current PHP features and strict types.
- **Alpine**: Small image size and minimal attack surface. Fits “zero redundancy” and fast pulls.
- **webdevops/php**: Image is built for PHP-FPM in container setups: includes FPM, common extensions, and a known layout. Config is overridden by mounting `php.ini` and `apcu.ini` into `/opt/docker/etc/php/` so we don’t need a custom Dockerfile for standard tuning.
- **Alternative**: You could switch to official `php:8.4-fpm-alpine` and add extensions (e.g. APCu) in a custom Dockerfile if you prefer to avoid third-party images; the current choice avoids maintaining that Dockerfile.

## docker-compose.local.yml overview

### Services

| Service | Role |
|---------|------|
| **cache-init** | One-shot job: creates cache directories on named volumes and sets ownership to UID/GID `1000` so PHP-FPM can write page cache and Latte compile cache reliably (bind-mounted project tree may be root-owned on the host). |
| **php** | PHP-FPM; project at `/var/www/html`; APCu healthcheck before Nginx is started. |
| **nginx** | Reverse proxy and static files; host **8080** → container **80**. |

### Startup order

1. **cache-init** completes successfully (`depends_on` with `service_completed_successfully`).
2. **php** starts and becomes **healthy** (APCu loaded).
3. **nginx** starts (`depends_on` PHP with `condition: service_healthy`).

### Volumes

- **Project bind mount**: `.` → `/var/www/html` on both `php` and `nginx` (code, `vendor/`, `.env`, fixtures, templates source).
- **Named volume `defense-cache`** → `/var/www/html/cache` on **php** only.
  Holds **file-based page HTML cache** (`CacheStore`). The volume name is historical; content is ordinary page cache, not tied to a specific feature.
- **Named volume `template-cache`** → `/var/www/html/templates/cache` on **php** only.
  Latte’s compiled templates live here so they are **not** written into the bind-mounted `./templates/cache` (avoids permission issues and keeps compile output off the host tree).

**Important**: Warmup or cache inspection **inside** the stack uses these volumes (`docker compose exec php php scripts/build-static.php`). Running `php scripts/build-static.php` **on the host** writes to host `./cache`, which is **not** the same storage as the PHP container’s `/var/www/html/cache` when using Compose.

### Environment and `.env`

- Compose may read a root **`.env`** file for **variable substitution** in `docker-compose.local.yml` (e.g. `${PHP_INI_FILE}`, `${APCU_INI_FILE}`). That is separate from injecting variables into a container.
- Application settings (`APP_ENV`, `CACHE_DIR`, etc.) are loaded in **`public/index.php`** via **Dotenv** from `.env` on disk. Because the project root is bind-mounted, the same `.env` file is visible inside the PHP container; there is no `env_file:` directive on services for that.

### php service details

- Volumes: bind mount; `defense-cache` and `template-cache` as above; `php.ini` and `apcu.ini` mounted from the host (paths from `PHP_INI_FILE` / `APCU_INI_FILE`, defaults `php.ini`, `apcu.ini`).
- **Healthcheck**: `php -m | grep -q apcu` ensures APCu is loaded before Nginx starts.

### nginx service details

- **`ports: "8080:80"`**: Local development URL http://localhost:8080.
- **`restart: unless-stopped`**: Container restarts after host reboot unless stopped explicitly.
- Mounts project root and `nginx/default.conf` read-only. See [Nginx](../nginx/README.md).

In production behind another reverse proxy, you may remove host `ports` and attach only to an external network; document that in your own deployment notes.

## php.ini

Mounted as the main PHP config (path controlled by `PHP_INI_FILE`, default `php.ini`).

| Directive | Value | Purpose |
|-----------|--------|---------|
| `memory_limit` | 128M | Cap per-request memory. |
| `display_errors` | Off | No PHP errors in output (production-safe). |
| `upload_max_filesize` | 10M | Max upload size. |
| `post_max_size` | 10M | Max POST body; should be ≥ upload limit. |
| `date.timezone` | UTC | Consistent timestamps (Twelve-Factor: explicit config). |

Override by changing `php.ini` or setting `PHP_INI_FILE` in `.env` to another file and keeping the same mount pattern in `docker-compose.local.yml`.

## apcu.ini

Mounted into PHP’s `conf.d` (path controlled by `APCU_INI_FILE`, default `apcu.ini`).

| Directive | Value | Purpose |
|-----------|--------|---------|
| `apc.enabled` | 1 | Enable APCu. |
| `apc.shm_size` | 64M | Shared memory size for APCu cache. |
| `apc.ttl` | 3600 | Default TTL in seconds for cache entries. |
| `apc.enable_cli` | 0 | Disable APCu in CLI (e.g. `build-static.php`); avoid wasting memory in short-lived scripts. |

**Why APCu in Docker**: Principles call for “APCu (Cache-Aside strategy)” and “maximal performance.” The app’s **page HTML** cache is file-based (in the named volume at `/var/www/html/cache` in Compose); APCu is available for future in-process caching. The healthcheck ensures the extension is loaded. If you never use APCu in code, you can still leave it enabled for consistency or disable it and adjust the healthcheck.

## Environment variables used by Docker / Compose

- **PHP_INI_FILE**: Filename for main PHP config (default `php.ini`). Used by Compose for the bind mount path.
- **APCU_INI_FILE**: Filename for APCu config (default `apcu.ini`). Same idea.
- Other vars in `.env` are consumed by the PHP app via Dotenv when `index.php` or CLI scripts load the project root `.env`.

## Project preparation checklist

1. Copy `.env.example` to `.env` and set at least `APP_ENV`, `APP_DEBUG`, `CACHE_DIR`, `DATA_SOURCE`, `FIXTURES_PATH`. Optionally set `PHP_INI_FILE` and `APCU_INI_FILE` if you use custom filenames.
2. Run `composer install` (locally or in a one-off container) so `vendor/` exists; the app needs it at `/var/www/html/vendor`.
3. Ensure `php.ini` and `apcu.ini` exist in the project root (or wherever `PHP_INI_FILE` / `APCU_INI_FILE` point) so the compose mounts are valid.
4. Run `docker compose -f docker-compose.local.yml up --build`. After **cache-init** and a healthy **php** service, **nginx** serves the app on **http://localhost:8080**.
5. Optional: warm the **container** file cache: `docker compose -f docker-compose.local.yml exec php php scripts/build-static.php`.
