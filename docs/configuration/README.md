# Configuration

Environment variables and main configuration files.

## Environment variables

All runtime configuration is driven by environment variables. No secrets in repo; copy `.env.example` to `.env` and adjust.

| Variable | Purpose | Example / default |
|----------|---------|-------------------|
| `APP_ENV` | Environment name | `production`, `development` |
| `APP_DEBUG` | Enable Flight error logging | `false`, `true` |
| `CACHE_TTL` | Page cache TTL in seconds; 0 = no expiry | `3600` |
| `CACHE_DIR` | Directory for page cache files (relative to project root or absolute) | `./cache` |
| `DATA_SOURCE` | Which data provider to use | `fixtures` (only supported value today) |
| `FIXTURES_PATH` | Base path for JSON fixtures (relative or absolute) | `./fixtures` |
| `PHP_INI_FILE` | Optional PHP ini file | `php.ini` |
| `APCU_INI_FILE` | Optional APCu ini file | `apcu.ini` |
| `DEFENSE_CONFIG_PATH` | Path to request-defense PHP config (relative or absolute) | `./config/defense.php` |

Path resolution in `config/services.php`: paths starting with `/` or `C:\` (or similar) are used as-is; relative paths are resolved from the project root (`dirname(__DIR__)` when loading from `config/`).

## config/services.php

- **Role**: Defines all container bindings and singletons. Invoked once when the application is built.
- **Input**: Reads `getenv()` for paths and options; uses `$rootDir = dirname(__DIR__)` for project root.
- **Bindings**: `app.root_dir`, `app.templates_dir`, `app.template_cache_dir`, `app.cache_dir`, `fixtures_path`, `pages_config` (from `config/pages.php`).
- **Singletons**: `SecurityHeaders`, `LatteRenderer`, `CacheStore`, `DataProviderInterface`, `PageRenderer`. Dependencies are resolved via `$container->get(...)` inside the factory closures.

No separate “dev” vs “prod” config file; behavior is controlled by env (e.g. `APP_DEBUG`, `CACHE_TTL`).

## config/pages.php

- **Role**: Page map for the site. Return an associative array: route path → config (template, title, data keys, optional dynamic block).
- **Consumers**: `config/routes.php` (registers GET routes and handlers) and `scripts/build-static.php` (enumerates all pages and dynamic variants for warmup). Single source of truth for “what pages exist.”

## config/routes.php

- **Role**: Registers Flight routes. Depends on the container being set on Flight (`appContainer`). Registers API routes and then a loop over `pages_config` for page routes. Does not define the page list itself; that stays in `config/pages.php`.

## config/localization.php

- **Role**: Declares which site locales exist and related options (e.g. default locale, URL prefix behavior). The fixture layout uses `fixtures/<locale>/`; this project is set up for a **single** default locale (`en`) unless you extend the config and add more folders under `fixtures/`.

## config/defense.php

- **Role**: Toggles and options for `RequestDefenseService` (`src/Defense/RequestDefenseService.php`). In the default repo settings, defense is **disabled**; adjust here if you enable filtering.

## Key settings summary

- **Cache**: File-based page cache in `CACHE_DIR`, TTL from `CACHE_TTL`. Latte template cache is under `templates/cache/` (fixed relative to `templates/`).
- **Data**: `DATA_SOURCE=fixtures` and `FIXTURES_PATH` point to the JSON content directory. Switching to another provider would require a new implementation and a new branch in `services.php` (or env-driven binding).
- **Debug**: `APP_DEBUG` controls Flight’s `flight.log_errors`. No separate debug toolbar or dev-only routes in the core.
