# Custom container vs Symfony Service Container

## What this project has

- Manual service definitions in `config/services.php` (bind + singleton).
- Lazy resolution on first `get()`.
- Singleton lifecycle via internal `$instances` and a bound closure (`Closure::bind`) so the factory can access private state.
- Tags and a boot phase (`bootServices()`, `TAG_BOOTABLE`).
- No autowiring, no compiled container, no PSR-11 by default.

## What Symfony offers (and we don’t)

- Autowiring and autoconfiguration.
- Compiled container (dependency graph built at build time).
- Public/private services, decorators, compiler passes, service subscribers.
- Rich tooling (e.g. debug commands, circular dependency detection).
- PSR-11 compliance out of the box.

## Verdict for this codebase

- **Appropriate**: For a small FlightPHP app with a limited number of services, the custom container is understandable, maintainable, and consistent with the project’s principles (minimal surface, no framework bloat).
- **Trade-off**: If the app grows (many services, multiple teams, need for stricter dependency analysis), introducing Symfony’s container (or another mature DI solution) could become justified. The current design does not preclude a later migration: services are already resolved from a single container and defined in one place.

## Summary

The custom container is a **deliberate, professional choice for the current scope**, not a missing dependency. It manages **service instances in memory** for the request/CLI run; it does not implement application data caching — that is handled by `CacheStore` and the file cache in `cache/`.
