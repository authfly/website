# Principles

Design and product principles that drive architecture and technology choices.

## Core doctrines

### 1. The Three Whales

- **Zero redundancy**: Minimal core; no superfluous libraries or layers. Each piece does one job.
- **Minimal requests**: Optimize interactions with caches, APIs, and the client. Only essential communication.
- **Maximal performance**: Prefer components that compile or run with minimal overhead (e.g. Latte vs Twig chosen for simpler generated PHP and less indirection).

### 2. FlightPHP First

- Prefer FlightPHP features and official extensions before adding external libraries or custom solutions.
- Use Flight’s routing and integration points; avoid reinventing them.
- New features are implemented as services registered in the application container.

### 3. Twelve-Factor compliance

- **Config in env**: All configuration via environment variables; no secrets in config files.
- **Stateless**: No sessions or local request state; each request is independent.
- **Disposability**: Processes can start and stop quickly; graceful shutdown when needed.
- **Dev/prod parity**: Same stack in development and production (Docker).

## Implications for this codebase

- Custom lightweight DI container instead of pulling in Symfony DI (zero redundancy, FlightPHP-first).
- File-based page cache and optional CLI warmup instead of only on-demand rendering (performance, minimal requests).
- Single page map in `config/pages.php` driving both web routes and cache warmup (single source of truth, no duplication).
- Data source behind `DataProviderInterface` so content can switch from fixtures to API/blob without changing rendering logic (flexibility without extra layers).
