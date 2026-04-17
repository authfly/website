# Documentation — Twelve-Factor FlightPHP

Documentation for developers: architecture, design decisions, and operational details. **New to the project?** Start with the [Getting started](getting-started/README.md) guide (101).

---

## Overview

This repository is a **lightweight PHP application** built on [FlightPHP](https://flightphp.com/) and aligned with the [Twelve-Factor App](https://12factor.net/) methodology. It delivers a content site with:

- **FlightPHP** for routing and request handling  
- **Latte** for server-side rendering  
- **Config-driven page map** (`config/pages.php`) for static and dynamic routes  
- **Adapter-based data access** via `DataProviderInterface` (fixtures today; API/blob later)  
- **File-based page cache** (PHP files in `cache/`) with optional CLI warmup  
- **DI container** from **`phpfasty/core`** (`PhpFasty\Core\Container`), configured in `config/services.php` — no Symfony dependency
- **Docker** stack: init job for cache volumes, PHP-FPM + Nginx, named volumes for page cache and Latte compile cache

Content and presentation are intended to be changeable via **config**, **templates**, and **fixtures** without touching core application code.

---

## Why (Motivation)

### Why this stack?

- **Minimal surface area**: One framework (Flight), one template engine (Latte), one data abstraction (DataProvider). No ORM, no full-stack framework, no frontend build by default.
- **Performance**: Latte compiles to plain PHP; page output is cached as PHP files (include-based). Cache can be warmed by a separate CLI process so the first request is already a hit.
- **Twelve-Factor alignment**: Config from env, stateless processes, disposable workers, same stack in Docker for dev and prod.

### Why PhpFasty’s container instead of Symfony DI?

- **Principles**: Zero redundancy and FlightPHP-first mean avoiding a large framework dependency when the app has few services and simple wiring.
- **Scope**: The app needs bindings, singletons, lazy resolution, and a boot phase (tagged services). **`phpfasty/core`** ships a small container that matches that model; no autowiring or compiled Symfony container is required.
- **Trade-off**: For a small, single-team codebase this is a reasonable and maintainable choice. For a growing, multi-service product, migrating to Symfony (or another mature container) may become worthwhile — see [Container vs Symfony](container/container-vs-symfony.md).

### Why file-based page cache instead of only APCu?

- **Persistence**: File cache in `cache/` survives process restarts and is shared across PHP-FPM workers. APCu is process-local (or per-node) and not used for page HTML in this app.
- **Warmup**: `scripts/build-static.php` runs as a separate CLI process: it flushes the file cache, then renders every configured page (including dynamic variants) and writes HTML into the same `cache/` directory. Web and CLI share the same `CacheStore` and `PageRenderer`; no duplicate caching logic.
- **APCu**: Referenced in principles for future use (e.g. data caching); currently the main cache is file-based page cache.

### Why config-driven pages and data keys?

- **Flexibility**: Adding or changing pages, routes, and data keys is done in `config/pages.php` and fixtures. Templates consume whatever keys are passed; the core does not hardcode page structure.
- **Single source of truth**: Routes for the page map are generated from `pages_config` in `config/routes.php`; the same config drives `build-static.php` for cache warmup. No duplicated route lists.

### Why the `PHP_SAPI === 'cli-server'` block in `public/index.php`?

- **Built-in PHP server**: When using `php -S ... -t public public/router.php`, that server does not rewrite URLs. The block serves static files directly (return false), normalizes `/index.php/...` to `...`, and sets `SCRIPT_NAME`/`PHP_SELF` so the app sees a single front controller. In production behind Nginx/Apache this block is not executed (`PHP_SAPI` is not `cli-server`).

---

## Documentation map

| Section | Description |
|--------|-------------|
| [Getting started](getting-started/README.md) | **101 guide** — from zero to a running app, step by step |
| [Architecture](architecture/README.md) | Request flow, directory layout, runtime components |
| [Principles](principles/README.md) | Three Whales, FlightPHP First, Twelve-Factor |
| [Container](container/README.md) | DI container, bindings, singletons, memory, Closure::bind |
| [Cache](cache/README.md) | Page cache, warmup script, cache directory |
| [Routing & entrypoint](routing/README.md) | index.php, router.php, cli-server, routes and pages |
| [Data & content](data-and-content/README.md) | DataProvider, fixtures, config-driven content, flexibility |
| [Development](development/README.md) | Templates (Latte), Tailwind build, design tokens, preparing data for views |
| [API](api/README.md) | JSON API (health, landing), basic by design, extending via config and handler classes |
| [Configuration](configuration/README.md) | Environment variables, services.php, key settings |
| [Compliance](compliance/README.md) | Twelve-Factor compliance: current state and gaps |
| [Deployment](deployment/README.md) | Registry build, dashboard API, GitHub Actions release flow |
| [Docker](docker/README.md) | Project preparation, docker-compose, php.ini, apcu.ini, why webdevops/php:8.4-alpine |
| [Private GitHub](private-repository/README.md) | This guide explains how to import a private GitHub repository from the terminal |
| [Nginx](nginx/README.md) | Server block, locations, static/API, security headers |

---

## Quick reference

- **Web entry**: `public/index.php` (or `public/router.php` for PHP built-in server).  
- **Container config**: `config/services.php`.  
- **Page map**: `config/pages.php`.  
- **Route registration**: `config/routes.php`.  
- **API**: `GET /api/health`, `GET /api/landing` (see [API](api/README.md)).  
- **Content source**: `fixtures/<locale>/` (JSON), default `fixtures/en/`, via `DataProviderInterface`.  
- **Page cache**: `cache/` (path from `CACHE_DIR`).  
- **Warmup**: `php scripts/build-static.php`.  
- **Docker**: `docker compose -f docker-compose.local.yml up --build`; app at **http://localhost:8080**; PHP config: `php.ini`, `apcu.ini` (see [Docker](docker/README.md)).
- **Nginx**: `nginx/default.conf` (see [Nginx](nginx/README.md)).
