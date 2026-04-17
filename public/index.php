<?php

declare(strict_types = 1);

if (PHP_SAPI === 'cli-server') {
    $requestUri = (string) ($_SERVER['REQUEST_URI'] ?? '/');
    $requestPath = parse_url($requestUri, PHP_URL_PATH);
    $requestedFile = is_string($requestPath) ? __DIR__ . $requestPath : null;

    if ($requestedFile !== null && is_file($requestedFile)) {
        return false;
    }

    if (is_string($requestPath) && str_starts_with($requestPath, '/index.php/')) {
        $normalizedPath = substr($requestPath, strlen('/index.php'));
        $_SERVER['REQUEST_URI'] = $normalizedPath === '' ? '/' : $normalizedPath;
    }

    $_SERVER['SCRIPT_NAME'] = '/index.php';
    $_SERVER['PHP_SELF'] = '/index.php';
}

require_once __DIR__ . '/../vendor/autoload.php';

use PhpFasty\Core\Application;
use Dotenv\Dotenv;

$dotenvPath = dirname(__DIR__);
if (is_readable($dotenvPath . '/.env')) {
    Dotenv::createImmutable($dotenvPath)->load();
}

$baseDir = dirname(__DIR__);
$application = Application::getInstance($baseDir, $baseDir . '/config/services.php');
$container = $application->getContainer();
Flight::set('appContainer', $container);

$debug = filter_var(
    getenv('APP_DEBUG') ?: 'false',
    FILTER_VALIDATE_BOOLEAN,
    FILTER_NULL_ON_FAILURE
);
$debug = $debug === null ? false : $debug;

Flight::set('flight.base_url', '/');
Flight::set('flight.log_errors', $debug);

require_once __DIR__ . '/../config/routes.php';

// Hermetic SEO guard: any URL with ?voice= must never enter search indexes,
// even if external links bypass robots.txt.
if (isset($_GET['voice']) && trim((string) $_GET['voice']) !== '') {
    header('X-Robots-Tag: noindex, nofollow', true);
}

Flight::start();
