<?php

declare(strict_types = 1);

$requestUri = (string) ($_SERVER['REQUEST_URI'] ?? '/');
$requestPath = parse_url($requestUri, PHP_URL_PATH);

if (is_string($requestPath) && $requestPath !== '/') {
    $requestedFile = __DIR__ . $requestPath;

    if (is_file($requestedFile)) {
        return false;
    }
}

$_SERVER['SCRIPT_NAME'] = '/index.php';
$_SERVER['PHP_SELF'] = '/index.php';

require __DIR__ . '/index.php';
