# Architecture

High-level structure, request flow, and main components.

## Request flow

1. **HTTP request** hits the web root. With the built-in PHP server, `public/router.php` is used; in production, Nginx/Apache forwards to `public/index.php`.
2. **Entrypoint** (`index.php`) loads `.env`, boots `PhpFasty\Core\Application::getInstance()`, gets the container, sets Flight options, then includes `config/routes.php` and runs `Flight::start()`.
3. **Routes** are registered in `config/routes.php`: fixed API routes (e.g. `GET /api/health`, `GET /api/landing`) and a loop over `pages_config` that registers one `GET <routePath>` per page.
4. **Page request**: For a path like `/` or `/blog/my-post`, the matching handler calls `PageRenderer::renderPage($routePath, $pageConfig, $routeParameters)`. No fourth argument → `forceRefresh = false`.
5. **PageRenderer** builds a cache key `page:<requestPath>`, checks `CacheStore::get()`; on miss, builds page data (via `DataProvider` and optional dynamic resolution), renders Latte (page + layout), then `CacheStore::set()`. Response is the HTML.
6. **Static assets**: Served directly by the web server (or by the `index.php` cli-server block when the path is a file under `public/`).

## Directory layout

| Path | Purpose |
|------|--------|
| `public/` | Web root; `index.php` (front controller), `router.php` (built-in server), static assets |
| `src/` | Application PHP: `Service` (e.g. `PageRenderer`), `View` (Latte), `Defense`, `Localization` |
| `config/` | `services.php` (container wiring), `pages.php` (page map), `routes.php` (Flight routes) |
| `templates/` | Latte layout and page templates |
| `fixtures/<locale>/` | JSON content per locale (default: `en/`; keys match `data` in `config/pages.php`) |
| `cache/` | Generated page cache (PHP files); path from `CACHE_DIR` |
| `scripts/` | CLI scripts (e.g. `build-static.php` for cache warmup) |

## Main components

- **Application** (`PhpFasty\Core\Application`, `vendor/phpfasty/core`): Singleton that builds the container from `config/services.php` and runs `bootServices()` for tagged services.
- **Container** (`PhpFasty\Core\Container`): Registers bindings and singletons; resolves services on `get()`; holds singleton instances in memory for the request (or CLI run).
- **DataProviderInterface** (`vendor/phpfasty/core`): Abstraction for content; `FixtureDataProvider` reads JSON from `fixtures/<locale>/` with fallback to `fixtures/<key>.json` if missing.
- **PageRenderer** (`src/Service/PageRenderer.php`): Orchestrates data loading, Latte rendering, and cache get/set; supports dynamic routes via `config/pages.php` `dynamic` section.
- **CacheStore** (`PhpFasty\Core\Cache\CacheStore`): File-based key-value store for HTML; key → `md5(key).php` in `CACHE_DIR`.
- **LatteRenderer** (`src/View/LatteRenderer.php`): Wraps Latte Engine; renders template names under `templates/` with given params.
- **RequestDefenseService** (`src/Defense/RequestDefenseService.php`): Optional request filtering; behavior is driven by `config/defense.php` (disabled by default in this repo).

## Separation of concerns

- **Content** is defined by fixture files and `config/pages.php` (data keys, template, title pattern, dynamic config).
- **Routing** is defined in `config/routes.php` (API) and derived from `pages_config` (pages).
- **Rendering** is in `PageRenderer` + Latte; **caching** is in `CacheStore` and used only inside `PageRenderer`. No duplicate caching logic elsewhere.
- **Warmup** reuses the same container, `PageRenderer`, and `CacheStore`; it only runs in CLI and passes `forceRefresh = true` when filling the cache.
