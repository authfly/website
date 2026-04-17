# Architecture Manifesto: Twelve-Factor Application on FlightPHP

## Key Principles

### 1. Twelve-Factor: Configuration in Environment

- Strict separation of code and configuration
- All settings via environment variables (`APP_ENV`, `CACHE_TTL`, `FIXTURES_PATH`)
- `.env` is not committed; template in `.env.example`
- The same code deploys across all environments

### 2. Stateless and Scalable

- API without sessions (or JWT/tokens only)
- No local state storage
- Processes without shared state — horizontal scaling without code changes

### 3. Unified Code Style and DI

- PSR-12, PSR-4, `declare(strict_types=1)`
- Dependencies via constructor; interfaces for abstractions
- Service container as the single point for creating services

### 4. API-First Pattern

- JSON API as the primary interface
- Static assets (HTML, CSS, JS) separate from dynamic content
- Hybrid: first N posts are static; the rest via API with cache

### 5. Unified Data Context

- Services receive data through interfaces and DI
- Controllers are thin: they call services and return JSON
- Cache (APCu) as a layer between API and data source

## Architectural Organization

### Directory Structure

```
project/
├── docker-compose.local.yml
├── php.ini
├── apcu.ini
├── .env.example
├── public/
│   ├── index.php
│   └── static/
├── src/
│   ├── Controller/
│   ├── Service/
│   └── ...
├── config/
├── fixtures/
└── nginx/
```

### Controller — Service — Data Source

- Controllers receive requests and call services
- Services handle business logic and data access
- Cache (APCu) as a layer between service and fixtures/DB

## Quality Principles

### Code Economy

- Minimal dependencies: FlightPHP, Container, phpdotenv
- No unnecessary libraries
- No dead code at any level

### Performance and Security

- APCu for in-memory caching
- Cache-Aside: on miss — load and write to cache
- Stateless — no state leakage between requests

### SOLID and DRY

- One class — one responsibility
- Dependency inversion via interfaces
- Reuse through composition and DI

### Extensibility

- New routes added without changing existing code
- Services registered in the container
- Documented extension points

## Technology Stack

- **Backend**: FlightPHP, PHP 8.4
- **Containerization**: Docker, webdevops/php:8.4-alpine, nginx:alpine
- **Cache**: APCu
- **Configuration**: vlucas/phpdotenv, environment variables

Following this manifesto ensures creating scalable, maintainable, and high-performance Twelve-Factor applications on FlightPHP.
