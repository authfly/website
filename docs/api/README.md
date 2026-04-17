# API

The application exposes a small JSON API. It is kept minimal on purpose and can be extended via config and dedicated classes without changing the core.

## Current endpoints

- **`GET /api/health`** — Liveness: returns `{"status":"ok"}`. Uses `SecurityHeaders::applyApiHeaders()`.
- **`GET /api/landing`** — Returns the same dataset as the home page, as JSON: `$dataProvider->get('landing')`. Same data key used in `config/pages.php` for the `/` page. Intended for SPA, mobile clients, or preview tooling.

Both are registered in `config/routes.php` and use the shared container (`SecurityHeaders`, `DataProviderInterface`). Responses are JSON; API-related security headers are applied via the middleware.

## Design: basic by default

The API is **intentionally minimal**:

- A few fixed routes are enough for health checks and one or two data endpoints.
- Paths and data keys are currently hardcoded in `config/routes.php`. For a single endpoint like `/api/landing` this is acceptable and keeps everything in one place.
- No separate API layer, versioning, or auth in the core. The app stays “zero redundancy” and Flight-first.

So “basic” here means: a small, clear set of JSON endpoints that reuse the same `DataProvider` and security headers as the rest of the app, without extra frameworks or abstraction until needed.

## Extending the API

If you add more data endpoints (e.g. `/api/blog`, `/api/about`), you can extend in two ways without cluttering `routes.php` with copy-paste.

### 1. Config-driven path → data key

- Add a config file (e.g. `config/api.php`) that returns a map: **path suffix → data key** (and optionally options), for example:

  ```php
  return [
      'landing' => ['key' => 'landing'],
      'blog'    => ['key' => 'blog'],
      'about'  => ['key' => 'about'],
  ];
  ```

- In `config/routes.php`: load this config, then in a single loop register `GET /api/{suffix}` for each entry, resolving the data key from the config and calling `$dataProvider->get($key)`. Adding or removing endpoints or changing keys becomes a config change only; no new route closures.

This mirrors the approach used for pages (`pages_config`): one source of truth, one registration loop.

### 2. Dedicated handler classes

- For endpoints that need more than “return one data key” (e.g. filtering, multiple keys, custom response shape), introduce **API handler classes** (e.g. under `src/Api/` or `src/Controller/Api/`).
- Each handler receives the container (or only the services it needs) and implements a single action (e.g. `__invoke()` or `handle()`). In `config/routes.php` you only register the route and call the handler; logic lives in the class.
- Optionally wire handlers from config (e.g. path → handler class or service id) so new endpoints are added by config + new class, not by editing route closures.

Combining both: use **config for simple “path → data key”** endpoints and **handler classes for anything more complex**. The API stays basic in the core (no framework, same container and DataProvider), but grows in a structured way via config and separate classes.

## Summary

- **Current**: Basic JSON API with `/api/health` and `/api/landing`; paths and keys are hardcoded in `config/routes.php`.
- **Extension**: Move to a config-driven map (path → data key) and/or dedicated API handler classes so new endpoints are added through config and new classes rather than more inline closures. The app does not implement this by default; it is the recommended direction when the number of API routes grows.
