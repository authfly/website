# Technical Instructions: Twelve-Factor Application on FlightPHP

## Development Environment Setup

### Coding Standards (PHPCS)

PHP_CodeSniffer is used with PSR-12, Symfony-style, PSR-4, and `declare(strict_types=1)` configuration.

#### PHPCS Configuration File (`phpcs.xml` in project root)

See the current `phpcs.xml` in the repository root. Install dependencies:

```bash
composer require --dev slevomat/coding-standard squizlabs/php_codesniffer
```

#### VS Code / Cursor Settings

File `.vscode/settings.json`:

```json
{
  "phpcs.enable": true,
  "phpcs.standard": "${workspaceFolder}/phpcs.xml",
  "phpcs.executablePath": "${workspaceFolder}/vendor/bin/phpcs",
  "phpcs.showSources": true,
  "phpcbf.enable": true,
  "phpcbf.executablePath": "${workspaceFolder}/vendor/bin/phpcbf",
  "phpcbf.standard": "${workspaceFolder}/phpcs.xml",
  "phpcs.composerJsonPath": "${workspaceFolder}/composer.json"
}
```

### Coding Rules

1. **Comments** — always in English
2. **Naming** — `camelCase` for methods and properties (PSR-12)
3. **Autoloading** — PSR-4, namespace `App` for `src/`
4. **Strict types** — `declare(strict_types=1)` in every PHP file
5. **DRY** — no logic duplication, extract to services
6. **SOLID** — single responsibility, DI via constructor

## Application Structure (Twelve-Factor)

```
project/
├── docker-compose.local.yml
├── php.ini
├── apcu.ini
├── .env.example
├── .gitignore
├── composer.json
├── public/              # document root
│   ├── index.php
│   └── static/          # HTML, CSS, JS
├── src/                 # App\ namespace
│   ├── Controller/
│   ├── Service/
│   └── ...
├── config/
├── fixtures/            # JSON data
└── nginx/
    └── default.conf
```

## Twelve-Factor Principles

### Config via Environment

- `APP_ENV`, `APP_DEBUG`, `CACHE_TTL`, `FIXTURES_PATH`
- Load via `vlucas/phpdotenv` (only if `.env` exists)
- `.env` in `.gitignore`, template in `.env.example`

### Stateless API

- No sessions (or JWT/tokens only)
- No local state storage
- Each request is independent

### Service Container and DI

- `Flight::registerContainerHandler()` with `flightphp/container`
- Lazy service creation on first `get()`
- Dependencies via constructor

## Entry Points

### `public/index.php`

```php
<?php

declare(strict_types=1);

require __DIR__ . '/../vendor/autoload.php';

// Load env (if .env exists)
if (file_exists(__DIR__ . '/../.env')) {
    $dotenv = Dotenv\Dotenv::createImmutable(__DIR__ . '/../');
    $dotenv->load();
}

Flight::registerContainerHandler(new \flight\Engine\Container());

Flight::route('GET /api/health', fn() => Flight::json(['status' => 'ok']));
Flight::route('GET /api/landing', [LandingController::class, 'index']);

Flight::start();
```

## Services and Controllers

### `src/Service/DataProvider.php`

```php
<?php

declare(strict_types=1);

namespace App\Service;

class DataProvider
{
    public function __construct(
        private readonly string $fixturesPath
    ) {
    }

    public function getLanding(string $lang): array
    {
        $path = $this->fixturesPath . "/{$lang}/landing.json";
        return json_decode(file_get_contents($path), true);
    }
}
```

### `src/Controller/LandingController.php`

```php
<?php

declare(strict_types=1);

namespace App\Controller;

use App\Service\DataProvider;
use App\Service\CacheInterface;

class LandingController
{
    public function __construct(
        private readonly DataProvider $dataProvider,
        private readonly CacheInterface $cache
    ) {
    }

    public function index(): void
    {
        $lang = $_GET['lang'] ?? 'en';
        $key = "landing:{$lang}";

        $data = $this->cache->get($key);
        if ($data === null) {
            $data = $this->dataProvider->getLanding($lang);
            $this->cache->set($key, $data, (int) ($_ENV['CACHE_TTL'] ?? 3600));
        }

        Flight::json($data);
    }
}
```

## Caching (APCu)

- Cache-Aside strategy: cache → on miss load from fixtures → write to cache
- Keys: `landing:{lang}`, `blog:list:{page}`, `blog:post:{id}`
- TTL from `CACHE_TTL`
- Fallback without cache when APCu is disabled

## Commands

```bash
# Check standards
composer phpcs

# Auto-fix
composer phpcbf

# Run
docker compose up
```

## Links

- [FlightPHP](https://docs.flightphp.com/)
- [Twelve-Factor App](https://12factor.net/)
- [12factor.net/config](https://12factor.net/config)
