# Twelve-Factor compliance

This section diagnoses how well the project currently adheres to the [Twelve-Factor App](https://12factor.net/) methodology. It is a snapshot of compliance, not a troubleshooting guide or a product roadmap. Use it to see gaps and decide what to improve next.

## Summary

The project is oriented toward Twelve-Factor but **does not fully satisfy** every factor. It is best described as an MVP skeleton with the right direction: single codebase, config from env, stateless web process, Docker stack, and a clear separation of build-time vs runtime. To claim “fully Twelve-Factor compliant,” the main areas to address are: **unified config/env handling and validation**, **centralized logging (stdout/stderr)**, and a **formal build → release → run pipeline** (including admin/CLI as first-class steps).

---

## 1. Codebase (One codebase, one app)

- **OK**: Single repo; clear layout: `src/`, `config/`, `public/`, `templates/`, `scripts/`. Described in `README.md`.
- **OK**: `.gitignore` excludes generated artifacts and caches (e.g. `cache/`, `templates/cache/`), keeping the repo clean.

---

## 2. Dependencies (Explicit and isolated)

- **OK**: Dependencies are declared in `composer.json`; `composer.lock` is present for reproducible installs.
- **Partial**: Versions use caret (`^`) for flexibility; for strict reproducibility in production you may want to pin versions in your pipeline or use `composer install --no-dev` and lock validation in CI.

---

## 3. Config (Store config in the environment)

- **OK**: `.env` and `.env.example` exist. Docker Compose can read `.env` at the project root for **compose-file interpolation** (e.g. `PHP_INI_FILE`, `APCU_INI_FILE` in volume paths). Application vars (`APP_ENV`, `APP_DEBUG`, `CACHE_TTL`, `CACHE_DIR`, `FIXTURES_PATH`, and the rest) are loaded in `public/index.php` (and CLI scripts) via **Dotenv** from the mounted project tree; services do not need `env_file:` for that.
- **Partial**: `scripts/build-static.php` loads `.env` the same way but does not share a single validated config layer with the web app (no unified env loader or schema). No formal validation or default types for required vars.

---

## 4. Backing services (Treat backing services as attached resources)

- **OK**: HTTP is provided by Nginx + PHP-FPM; APCu is configured. Services are attached via config (env and Docker), not hardcoded.
- **Partial**: The API (`/api/health`, `/api/landing`) is minimal and does not yet represent a full service layer over external stores; it is adequate for the current scope but not a complete “backing service” story.

---

## 5. Build, release, run (Strict separation of stages)

- **Partial**: There is a build-time step (`scripts/build-static.php` warms the page cache into `cache/`), but no formal **release** stage (versioned artifact, checks, migrations, or pre-deploy verification). No single pipeline that does “build → release artifact → run.”
- **Note**: `build-static.php` writes to `cache/` (via `CacheStore`), not to `public/`; the app serves pages from that cache at runtime. With **Docker Compose**, that directory is often the **`defense-cache` volume** inside the PHP container, not the host’s `./cache` — see [Docker](../docker/README.md).

---

## 6. Processes (Execute the app as one or more stateless processes)

- **OK**: The web app runs as a stateless front process via Flight in `public/index.php`. No in-process sessions or global file-based state; each request is independent.

---

## 7. Port binding (Export services via port binding)

- **OK**: The app is served via Nginx listening on a port; in Docker, host port 8080 maps to container port 80. The app is reachable via that bound port.

---

## 8. Concurrency (Scale out via the process model)

- **OK**: Architecture allows scaling (e.g. more Nginx or PHP-FPM workers/containers). No in-memory state that would prevent horizontal scaling.
- **Partial**: There is no example of running multiple PHP replicas with a **shared** file cache. In Docker Compose, page cache and Latte compile cache live on **named volumes** attached to the PHP container; scaling out would require a shared filesystem or switching cache storage — document your chosen approach for full compliance.

---

## 9. Disposability (Fast startup and graceful shutdown)

- **OK**: Containers start and stop quickly; minimal services.
- **Partial**: Health checks exist for the PHP container (e.g. APCu check in `docker-compose.local.yml`); Nginx depends on it. No application-level readiness endpoint or health checks for higher-level containers beyond that.

---

## 10. Dev/prod parity (Keep development, staging, and production as similar as possible)

- **OK**: Same stack in Docker (PHP 8.4, Nginx, APCu), same config files (`php.ini`, `apcu.ini`), env-driven config. Parity is possible when running via Docker.
- **Partial**: If local development uses the built-in PHP server and no Docker, there is an intentional gap between “local” and “prod”; parity is then a choice (e.g. “run Docker locally too”).

---

## 11. Logs (Treat logs as event streams)

- **Gap**: Logs are not explicitly normalized to stdout/stderr. There is no central logging layer in the app (`public/index.php` does not wire a logger to stdout/stderr), no structured log format, and no documented rotation or aggregation. `php.ini` has `display_errors = Off`; errors are not necessarily streamed in a 12-factor way. This is the main missing piece for “logs as stream.”

---

## 12. Admin processes (Run admin/management tasks as one-off processes)

- **OK**: Admin-style work is done via a one-off CLI process (`scripts/build-static.php`), separate from the web process.
- **Partial**: It is not yet formalized as a runbook or a defined step in CI (with env checks, artifact checks, and clear success/failure). For full compliance, one-off admin tasks should be documented and integrated into the release/run pipeline.

---

## What to improve next (priority)

To move toward full Twelve-Factor compliance, focus on:

1. **Config**: Unified env loading and validation (shared between web and CLI); optional schema or required-vars check.
2. **Logs**: Route application logs to stdout/stderr; define a simple format (e.g. JSON or key=value); document rotation/aggregation if needed.
3. **Build → release → run**: Define a release step (version, checks, optional migrations) and a single pipeline that produces a runnable release; include cache warmup and other admin tasks as first-class steps.
4. **API**: If the app is to be “API-first” with backing services, evolve the API from placeholders to a clear service layer; the current state is acceptable for a content-oriented MVP.

This document reflects the current state; update it as the project closes these gaps.
