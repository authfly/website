# Page cache and warmup

## Where cache lives

- **Directory**: Configured by `CACHE_DIR` (default `./cache`), resolved from project root in `config/services.php` and passed to `CacheStore`.
- **Docker Compose**: The `php` service mounts a **named volume** at `/var/www/html/cache`, which overrides the bind-mounted project’s `./cache` for that container. Page cache written inside the container (including `docker compose exec php php scripts/build-static.php`) lives in that volume, not necessarily in your host `./cache` folder. Latte compile output similarly uses the **`template-cache`** volume at `/var/www/html/templates/cache`.
- **Format**: One PHP file per key: `md5('page:' . requestPath) . '.php'`. File contains a `return '<html>...';` statement plus metadata comments (timestamp, key, TTL).
- **Shared**: Both the web app and the CLI warmup script use the same container and thus the same `CacheStore` and same directory. No separate “static” vs “runtime” cache.

## How caching works

- **PageRenderer::renderPage()** builds a cache key `'page:' . $requestPath`. If `forceRefresh` is false, it calls `renderAndCache()`, which does:
  - `CacheStore::get($cacheKey)`; if hit, return HTML.
  - Else: build page data, render Latte (page + layout), then `CacheStore::set($cacheKey, $html)` and return HTML.
- **TTL**: `CacheStore::get()` checks file mtime and `CACHE_TTL`; if expired, the file is removed and treated as miss. So cache is invalidated by time or by explicit `invalidate()` / `flush()`.

## build-static.php (warmup)

- **Process**: Runs as a **separate CLI process**. Does not require the web server to be running.
- **Steps**:
  1. Load autoload and `.env`, get `Application::getInstance()` and container.
  2. Resolve `CacheStore`, `PageRenderer`, and `pages_config`.
  3. `$cacheStore->flush()` — clear all cached PHP files in `CACHE_DIR`.
  4. For each entry in `pages_config`, get all route parameter sets (e.g. for `/blog/@slug`, one set per post). For each set, call `$pageRenderer->renderPage($routePath, $pageConfig, $routeParameters, true)` and collect the resulting request path.
  5. Print how many routes were warmed and list the paths.
- **Same logic**: No duplicate caching implementation. Warmup uses the same `PageRenderer::renderPage(..., true)` so the first request after warmup is a cache hit. `forceRefresh = true` forces invalidation of that key before render so the cache is repopulated.

## When to run warmup

- After deployment or after changing content/templates, so the first user request does not trigger a full render.
- Optionally in CI/CD before switching traffic, so the cache is pre-filled.

## APCu

- APCu is mentioned in project principles for potential use (e.g. data caching). The **page HTML** cache is file-based in `cache/`, not APCu. Latte’s own compiled templates go to `templates/cache/` (separate from page cache).
