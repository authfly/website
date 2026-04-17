# Twelve-Factor FlightPHP Application

A lightweight, high-performance API application built on [FlightPHP](https://flightphp.com/) and the [Twelve-Factor App](https://12factor.net/) methodology. It prioritizes speed, statelessness, and minimalism.

## Core Principles

Our development philosophy is guided by two main doctrines:

### 1. The Three Whales

Every aspect of the application adheres to:

- **Zero Redundancy**: Eliminating superfluous code, libraries, and dependencies. A minimal core that does exactly what's needed, and nothing more.
- **Minimal Requests**: Optimizing interactions with caches, APIs, and the browser. Only essential communication occurs.
- **Maximal Performance**: Prioritizing speed in both perception (fast responses) and reality (efficient backend, APCu cache). When choosing engines (e.g. Latte vs Twig), the differentiator is *what* they compile to—simpler generated PHP and less runtime overhead—not “we have compilation and cache too”; both have that. Zero redundant layers.

### 2. FlightPHP First

We leverage the capabilities of the Flight PHP framework:

- **FlightPHP First**: All solutions must prioritize existing FlightPHP functionality and official extensions before introducing external libraries or custom solutions.
- **No Reinvention**: We use Flight's built-in features: Dependency Injection Container, routing, and more. We avoid redundant development.
- **Modular Services**: New features are developed as services registered in the container, integrating seamlessly with Flight's architecture.

### 3. Twelve-Factor Compliance

- **Config in env**: All configuration via environment variables. No config files with secrets.
- **Stateless**: No sessions or local state. Each request is independent.
- **Disposability**: Processes can start and stop quickly. Graceful shutdown when needed.
- **Dev/prod parity**: Same stack in development and production (Docker).

## Architecture

- **Backend**: FlightPHP, PHP 8.4
- **Containerization**: Docker (webdevops/php:8.4-alpine, nginx:alpine)
- **Cache**: APCu (Cache-Aside strategy)
- **Data**: JSON fixtures, or external APIs; no database required for minimal setup
- **Deployment**: `docker compose up` from the project root

## Features

- JSON API endpoints
- Environment-based configuration
- Service container with lazy DI
- APCu caching with configurable TTL
- Hybrid static + dynamic content (optional)
- PSR-12, PSR-4, strict types

## Requirements

- PHP 8.4+
- Docker and Docker Compose
- Composer
