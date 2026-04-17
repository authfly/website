# Twelve-Factor FlightPHP Prototype

This repository is a lightweight FlightPHP application built as a live PHP site with Latte templates, adapter-based data access, and a PHP page cache.

## What is included

- **FlightPHP** for routing and request handling
- **Template rendering** with a pluggable renderer contract in core (`TemplateRendererInterface`) and app-level Latte adapter
- **DataProviderInterface** so content can come from JSON fixtures now and API/blob sources later
- **PHP page cache** for WP-style cached responses stored as PHP files
- **Docker compose skeleton** for PHP-FPM + Nginx deployment (`webdevops/php:8.4-alpine`)
- **APCu support** via dedicated config (`apcu.ini`)
- Local **12-factor style** environment configuration using `.env` and `.env.example`

## Directory layout

- `public/` — Web root and front controller
- `src/` — Application PHP source code
- `config/` — Container, page map, and route configuration
- `templates/` — Latte layout and page templates
- `fixtures/en/` — Default locale (English) JSON content; `fixtures/<locale>/` for other languages
- `cache/` — Generated PHP page cache
- `scripts/` — Utility scripts such as cache warmup
- `docker-compose.local.yml` — PHP + Nginx local development stack
- `docker-compose.yml` — Single-image production stack for dashboard/CD
- `nginx/default.conf` — Nginx front-controller config

## Quick setup (local)

1. Install dependencies (requires [Composer](https://getcomposer.org/) on your `PATH`):

```bash
composer install
```

2. Copy environment defaults:

```bash
cp .env.example .env
```

3. Start the live PHP application:

```bash
php -S localhost:8080 -t public public/router.php
```

4. Open `http://localhost:8080`

## Cache warmup

Warm the PHP page cache for all configured routes:

```bash
php scripts/build-static.php
```

This script no longer generates static HTML files in `public/`. It renders all configured pages through the same application services and writes cached PHP responses into `cache/`.

## Runtime architecture

- `public/index.php` boots FlightPHP and the application container
- `public/router.php` is the local PHP built-in server router for clean URLs
- `config/pages.php` defines the public page map
- `config/routes.php` registers API and page routes
- `PhpFasty\Core\\Data\\DataProviderInterface` abstracts the content source
- `App\\Service\\PageRenderer` loads data, renders templates through the renderer contract, and stores/retrieves cached pages

Runtime classes such as `Application`, `Container`, `CacheStore`, and `SecurityHeaders` are now extracted into `phpfasty/core` and consumed as a Composer library from this app.  
App-level template rendering is bound in `config/services.php` as `TemplateRendererInterface` → `App\View\LatteRenderer`, so future sites can swap in Twig (or another engine) without touching core.

## Compose file policy

This repository intentionally keeps two separate Compose files with different roles:

- `docker-compose.local.yml` is the local development and legacy repo-import stack.
- `docker-compose.yml` is the registry-based production deploy template.

Use `docker-compose.local.yml` when:

- you want live local development with bind mounts
- you want to run the old two-container PHP + Nginx stack manually
- you want to keep compatibility with the dashboard GUI or API repo-import flow by setting `compose_path` to `docker-compose.local.yml`

Use `docker-compose.yml` when:

- CI builds a Docker image and pushes it to the registry
- the dashboard stores `compose_yaml` directly instead of cloning the repo
- deployment is driven by immutable image tags such as `sha-<commit>`

Do not use `docker-compose.local.yml` for the new production pipeline. It is kept for development convenience and backward compatibility only.

## Docker setup (optional)

The local development stack includes:

- `webdevops/php:8.4-alpine` (PHP-FPM + extensions)
- `nginx:alpine` as the HTTP front

Run:

```bash
docker compose -f docker-compose.local.yml up --build
```

Then visit the configured host and port for the application.

## Production image

For registry-based deployment, the repository now includes a dedicated production build:

- `Dockerfile` — multi-stage build that installs Composer dependencies and packages the app into a single `nginx + php-fpm` runtime image
- `docker/production/nginx/default.conf` — production front-controller config for the bundled Nginx
- `docker/production/entrypoint.sh` — starts PHP-FPM and Nginx; optional cache warmup via `APP_WARM_CACHE=1`
- `docker-compose.yml` — dashboard/CD-oriented compose template that expects `APP_IMAGE` and keeps runtime caches in named volumes
- `docker-compose.local.yml` remains available for local development and the legacy repo-import deployment path

Example local production-style run:

```bash
docker build -t twelve-factor:local .
APP_IMAGE=twelve-factor:local docker compose up -d
```

The production container exposes only internal port `80`, which matches the dashboard proxy model. Runtime cache directories stay writable through named volumes instead of host bind mounts.

If you still want to deploy the old way through the dashboard GUI or a `curl` import request, use `docker-compose.local.yml` as the repository compose path. For the new recommended flow, CI should render `docker-compose.yml` with the target image tag and send that YAML to the dashboard API.

## Deployment modes

This repository now supports two production deployment styles:

- **GitHub Actions** via `.github/workflows/deploy.yml`:
  GitHub builds the image, pushes it to the registry, renders `docker-compose.yml`, updates the dashboard app, and triggers deploy.
- **Local `paas` Go runner** via `.paas/extensions/deploy.yml`:
  your local machine uploads the repository snapshot over SSH, the server builds the image, pushes it to the registry, renders `docker-compose.yml` on the server, updates the dashboard app, and triggers deploy.

The second flow is useful when:

- you want to deploy manually without waiting for CI
- Docker is available on the server but not on the local workstation
- you want the image to be built on the target host and still keep immutable registry tags

The current project-level `deploy` extension is a **project override** for `.gorunner` and is intentionally different from the built-in runner `deploy`: it builds on the remote Linux server instead of locally.

### Current `paas` deploy command

For the current Windows + Git Bash workflow, the proven command is:

```bash
env -i \
  HOME="$HOME" \
  USERPROFILE="$USERPROFILE" \
  HOMEDRIVE="$HOMEDRIVE" \
  HOMEPATH="$HOMEPATH" \
  PATH="$PATH" \
  TERM="${TERM:-xterm-256color}" \
  LANG="${LANG:-en_EN.UTF-8}" \
  SSH_AUTH_SOCK="$SSH_AUTH_SOCK" \
  SSH_AGENT_PID="$SSH_AGENT_PID" \
  INPUT_REGISTRY_USERNAME="$INPUT_REGISTRY_USERNAME" \
  INPUT_REGISTRY_PASSWORD="$INPUT_REGISTRY_PASSWORD" \
  ./paas.exe run deploy
```

Why the wrapped environment is needed:

- the current runner propagates process environment variables into remote steps
- Windows shells expose variables like `ProgramFiles(x86)` which are not valid `export` identifiers in Bash
- a raw Windows `PATH` can also break remote `bash -lc` execution on Linux

Using `env -i` with a small allowlist keeps the local command predictable and prevents Windows-specific environment variables from breaking remote steps.

### Hardening already applied

The production image now protects against accidental Windows line endings in shell scripts:

- `.gitattributes` enforces LF for `*.sh` and `Dockerfile`
- `Dockerfile` runs `sed -i 's/\r$//' /usr/local/bin/docker-entrypoint` before `chmod +x`

That protection was added after a real deployment failure where the container restarted with:

```text
exec /usr/local/bin/docker-entrypoint: no such file or directory
```

even though the file existed in the image. The real cause was a `CRLF` shebang (`#!/bin/sh^M`).

## PHP and Composer on PATH

Install **PHP 8.4+** and **Composer** and ensure both are available as `php` and `composer` on your `PATH`.

Optional: if you keep a Windows PHP build under `.php\php`, run `setphp.bat` (Command Prompt) or `.\setphp.ps1` (PowerShell) to prepend that folder to `PATH` for the current session so `php` resolves there.

## APCu configuration

`apcu.ini`:

```ini
apc.enabled = 1
apc.shm_size = 64M
apc.ttl = 3600
apc.enable_cli = 0
```

## Environment variables

See `.env.example` for baseline settings.

Common values:

- `APP_ENV`
- `APP_DEBUG`
- `CACHE_TTL`
- `CACHE_DIR`
- `DATA_SOURCE`
- `FIXTURES_PATH`
- `PHP_INI_FILE`
- `APCU_INI_FILE`

## Notes

- Page content is rendered through the current app-level renderer (default `App\View\LatteRenderer`) on demand and cached as PHP files in `cache/`
- Current content comes from JSON fixtures through the adapter layer, but the application is prepared for API and blob-backed providers
- After upgrading PHP, verify:

```bash
php -v
composer -V
```
