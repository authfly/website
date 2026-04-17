# Service container

The application uses the **DI container from `phpfasty/core`** (`PhpFasty\Core\Container`), wired in `config/services.php`. There is no Symfony DependencyInjection bundle in this project.

## Capabilities

- **Bindings**: Register a string key to a value or a factory closure. `Container::bind($abstract, $concrete)`.
- **Singletons**: Register a factory; on first `get($abstract)` the factory is invoked and the result stored in `$instances`; subsequent `get()` return the same instance. `Container::singleton($abstract, $concrete)`.
- **Lazy resolution**: Services are created only when `get()` is called (or when another service’s factory calls `$container->get(...)`).
- **Tags**: Services can be tagged (e.g. `bootable`); `findTaggedServiceIds()` and `bootServices()` run `boot()` (or `register()`) on them after the container is built.
- **Extend**: Wrap an existing binding with a closure that receives the resolved instance and the container (`extend()`).
- **Factory**: `factory($abstract)` returns a closure that calls `make($abstract)` (new instance, bypassing singleton cache).

## Where definitions live

- **`config/services.php`**: One callable that receives the `Container` and registers all bindings and singletons. Paths and options come from environment variables and `$rootDir`.

## Memory and Closure::bind

- The container holds: `$bindings`, `$instances`, `$factories`, `$classMap`, `$tags`. All are in-memory for the lifetime of the process (one request in FPM, or one CLI run).
- For singletons, the container stores a **bound closure** created with `Closure::bind($factory, $this, self::class)` so that when the closure is invoked it runs in the scope of the container and can read/write private `$instances` without exposing the container or using reflection. The closure is stored in `$bindings[$abstract]`; the same key may have the original factory in `$factories` for `make()` / `getFactory()`.
- So: **yes, memory is involved** — the container holds closures and resolved singleton instances. This is normal for a DI container; it does not implement application-level “data in memory” (that would be a cache or session layer).

## Relation to FlightPHP

Flight is given the container via `Flight::set('appContainer', $container)` in `public/index.php`. Routes in `config/routes.php` obtain the container with `Flight::get('appContainer')` and resolve services (e.g. `PageRenderer`, `CacheStore`, `pages_config`) from it. The app does not use Flight’s built-in container for application services; it uses the PhpFasty container.

## When to consider Symfony DI

For a small, single-team app this container is sufficient and aligns with “zero redundancy” and “FlightPHP first”. For a larger codebase with many services, autowiring, and need for compile-time checks and tooling, migrating to Symfony’s container (or another PSR-11 implementation) may become worthwhile. See [Container vs Symfony](container-vs-symfony.md).
