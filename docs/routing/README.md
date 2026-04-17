# Routing and entrypoint

## Two entrypoints

- **`public/index.php`**: Main front controller. Loads env, boots application and container, registers routes, runs Flight. Used in production (Nginx/Apache point here).
- **`public/router.php`**: Used by PHP’s built-in server (`php -S ... -t public public/router.php`). Serves static files from `public/` when the path matches a file; otherwise sets `SCRIPT_NAME`/`PHP_SELF` and includes `index.php`.

## The `PHP_SAPI === 'cli-server'` block in index.php

When the request is handled by the **built-in PHP web server** (`php -S`), `PHP_SAPI` is `'cli-server'`. In that case only:

1. **Static files**: If `REQUEST_URI` path corresponds to an existing file under `public/`, the script returns `false`. The built-in server then serves that file itself. So CSS/JS/images are served without going through Flight.
2. **URL normalization**: If the path starts with `/index.php/` (e.g. `/index.php/blog`), `REQUEST_URI` is rewritten to `/blog` (or `/`) so the app sees a clean path. Flight then matches routes correctly.
3. **SCRIPT_NAME / PHP_SELF**: Set to `'/index.php'` so the app behaves as if every request goes through the front controller.

When using Nginx or Apache, `PHP_SAPI` is not `cli-server`, so this block is skipped. Static files and rewriting are handled by the web server config (e.g. `nginx/default.conf`).

## How routes are registered

- **`config/routes.php`** is included after the container is set on Flight. It gets the container, then:
  - Registers a few **fixed routes** (e.g. `GET /api/health`, `GET /api/landing`).
  - Iterates over `$pages = $container->get('pages_config')` and for each key (e.g. `'/'`, `'/blog/@slug'`) registers `Flight::route('GET ' . $routePath, ...)`. The handler extracts route parameters from the URL, calls `PageRenderer::renderPage()`, applies security headers, and outputs HTML.

So the **page map is the single source of truth**: the same `pages_config` drives both web routes and the cache warmup script. Adding or changing a page is done in `config/pages.php`; no need to touch route registration code for standard pages.

## Dynamic routes

- A page entry in `config/pages.php` can include a `dynamic` block (e.g. `param`, `dataset`, `collection`, `lookup`, `item`). That tells `PageRenderer` how to resolve one “item” from a dataset (e.g. a blog post by slug) and inject it into the template. The route path uses a placeholder like `@slug`, which Flight fills from the URL. So `/blog/my-post` yields `$routeParameters = ['slug' => 'my-post']`, and the renderer finds the matching post in the `blog` dataset and passes it as `post` to the template.
