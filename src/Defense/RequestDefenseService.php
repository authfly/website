<?php

declare(strict_types = 1);

namespace App\Defense;

final class RequestDefenseService
{
    private const DATA_FILE = 'request-defenses.json';

    private const DEFAULT_LIMIT_RULE = [
        'max_requests' => 3,
        'time_window_seconds' => 10,
        'block_seconds' => 10,
    ];

    /** @var array<string, mixed> */
    private array $config;

    public function __construct(
        private readonly string $basePath,
        array $config = []
    ) {
        $this->config = $this->normalizeConfig($config);
    }

    public function getClientIp(array $server = []): string
    {
        $serverData = $server !== [] ? $server : $_SERVER;
        $candidateHeaders = [
            $serverData['HTTP_CF_CONNECTING_IP'] ?? null,
            $serverData['HTTP_X_FORWARDED_FOR'] ?? null,
            $serverData['HTTP_X_REAL_IP'] ?? null,
            $serverData['REMOTE_ADDR'] ?? null,
        ];

        foreach ($candidateHeaders as $candidateHeader) {
            if (!is_string($candidateHeader)) {
                continue;
            }

            $candidate = trim($candidateHeader);
            if ($candidate === '') {
                continue;
            }

            $parts = explode(',', $candidate);
            foreach ($parts as $part) {
                $ipCandidate = trim($part);
                if (filter_var($ipCandidate, FILTER_VALIDATE_IP) !== false) {
                    return $ipCandidate;
                }
            }
        }

        return 'unknown';
    }

    public function isLimitExceeded(string $ip, string $routePath, string $method = 'GET', array $server = []): bool
    {
        if (!($this->config['enabled'] ?? false)) {
            return false;
        }

        $this->ensureDirectory();

        $normalizedRoute = $this->normalizeRoutePath($routePath);
        $requestMethod = strtoupper(trim($method));
        if ($requestMethod === '') {
            $requestMethod = 'GET';
        }
        $serverData = $server !== [] ? $server : $_SERVER;
        $isSuspicious = $this->isSuspiciousRequest($serverData);

        $now = time();

        if ($this->isLimitExceededForScope($ip, 'scan', $this->resolveScanPolicy($isSuspicious), $now)) {
            return true;
        }

        $policy = $this->resolveRequestPolicy($normalizedRoute, $requestMethod, $isSuspicious);

        return $this->isLimitExceededForScope(
            $ip,
            $this->resolveRouteScope($normalizedRoute),
            $policy,
            $now
        );
    }

    private function isLimitExceededForScope(string $ip, string $scope, array $policy, int $now): bool
    {
        if (($policy['enabled'] ?? true) === false) {
            return false;
        }

        $stateFile = $this->buildStateFile($ip, $scope);
        $handle = fopen($stateFile, 'c+');
        if ($handle === false) {
            return false;
        }

        try {
            if (!flock($handle, LOCK_EX)) {
                return false;
            }

            $content = stream_get_contents($handle);
            $state = $this->decodeState($content);

            if ($state['blocked_until'] > $now) {
                return true;
            }

            $maxRequests = (int) ($policy['max_requests'] ?? 3);
            $timeWindowSeconds = (int) ($policy['time_window_seconds'] ?? 10);
            $blockSeconds = (int) ($policy['block_seconds'] ?? 10);

            $history = $this->normalizeHistory($state['history'] ?? [], $now, $timeWindowSeconds);
            $history[] = $now;

            if (count($history) > $maxRequests) {
                $this->writeState(
                    $handle,
                    [
                        'blocked_until' => $now + $blockSeconds,
                        'history' => [],
                    ]
                );

                return true;
            }

            $this->writeState(
                $handle,
                [
                    'blocked_until' => 0,
                    'history' => $history,
                ]
            );

            return false;
        } finally {
            fclose($handle);
        }
    }

    private function buildStateFile(string $ip, string $scope): string
    {
        return rtrim($this->basePath, DIRECTORY_SEPARATOR)
            . DIRECTORY_SEPARATOR
            . self::DATA_FILE
            . '.'
            . hash('sha256', $ip . '|' . trim($scope));
    }

    private function resolveRouteScope(string $normalizedRoute): string
    {
        return 'route:' . $normalizedRoute;
    }

    private function resolveRequestPolicy(string $normalizedRoute, string $requestMethod, bool $isSuspicious): array
    {
        $policy = $this->config['default'];

        $methodRules = $this->config['methods'];
        if (isset($methodRules[$requestMethod])) {
            $policy = $this->mergePolicy($policy, $methodRules[$requestMethod]);
        }

        $policy = $this->mergePolicy($policy, $this->resolveRoutePolicy($normalizedRoute));

        if ($isSuspicious && ($this->config['suspicious']['enabled'] ?? false) === true) {
            $policy = $this->mergePolicy($policy, $this->config['suspicious']);
        }

        return $this->normalizePolicy($policy);
    }

    private function resolveScanPolicy(bool $isSuspicious): array
    {
        if (!($this->config['scan']['enabled'] ?? false)) {
            return ['enabled' => false];
        }

        $policy = $this->mergePolicy($this->config['default'], $this->config['scan']);

        if ($isSuspicious && ($this->config['suspicious']['enabled'] ?? false) === true) {
            $policy = $this->mergePolicy($policy, $this->config['suspicious']);
        }

        return $policy;
    }

    private function isSuspiciousRequest(array $serverData): bool
    {
        if (($this->config['headers']['enabled'] ?? false) === false) {
            return false;
        }

        $userAgent = trim((string) ($serverData['HTTP_USER_AGENT'] ?? ''));
        if (($this->config['headers']['require_user_agent'] ?? true) && $userAgent === '') {
            return true;
        }

        if (
            $this->config['headers']['min_user_agent_length'] > 0
            && strlen($userAgent) < (int) $this->config['headers']['min_user_agent_length']
        ) {
            return true;
        }

        $userAgentLower = strtolower($userAgent);
        $blockedUserAgents = $this->normalizeHeaderList($this->config['headers']['blocked_user_agents'] ?? []);
        foreach ($blockedUserAgents as $pattern) {
            if (str_contains($userAgentLower, $pattern)) {
                return true;
            }
        }

        if (($this->config['headers']['require_accept_language'] ?? false)) {
            $acceptLanguage = trim((string) ($serverData['HTTP_ACCEPT_LANGUAGE'] ?? ''));
            if ($acceptLanguage === '') {
                return true;
            }
        }

        return false;
    }

    private function normalizeHeaderList(array $patterns): array
    {
        $normalized = [];
        foreach ($patterns as $pattern) {
            if (!is_string($pattern)) {
                continue;
            }

            $trimmed = strtolower(trim($pattern));
            if ($trimmed === '') {
                continue;
            }

            $normalized[] = $trimmed;
        }

        return $normalized;
    }

    private function resolveRoutePolicy(string $normalizedRoute): array
    {
        $routeConfig = $this->config['routes'];

        $exact = $routeConfig['exact'] ?? [];
        if (is_array($exact) && isset($exact[$normalizedRoute]) && is_array($exact[$normalizedRoute])) {
            return $exact[$normalizedRoute];
        }

        $prefixRules = $routeConfig['prefix'] ?? [];
        if (!is_array($prefixRules)) {
            return [];
        }

        $matchedRule = [];
        $matchedLength = -1;

        foreach ($prefixRules as $prefix => $rule) {
            if (!is_string($prefix) || !is_array($rule)) {
                continue;
            }

            $normalizedPrefix = $this->normalizePrefix($prefix);
            if ($normalizedPrefix === '') {
                continue;
            }

            if (!str_starts_with($normalizedRoute, $normalizedPrefix)) {
                continue;
            }

            $prefixLength = strlen($normalizedPrefix);
            if ($prefixLength <= $matchedLength) {
                continue;
            }

            $matchedRule = $rule;
            $matchedLength = $prefixLength;
        }

        return $matchedRule;
    }

    private function normalizePrefix(string $prefix): string
    {
        $normalizedPrefix = trim($prefix);
        if ($normalizedPrefix === '') {
            return '';
        }

        if (!str_starts_with($normalizedPrefix, '/')) {
            $normalizedPrefix = '/' . $normalizedPrefix;
        }

        return rtrim($normalizedPrefix, '/');
    }

    private function normalizeRoutePath(string $routePath): string
    {
        $route = trim($routePath);
        if ($route === '') {
            return 'unknown';
        }

        if ($route !== '/' && str_ends_with($route, '/')) {
            $route = rtrim($route, '/');
        }

        if ($route === '') {
            return '/';
        }

        return $route;
    }

    private function decodeState(?string $payload): array
    {
        if ($payload === null || $payload === '') {
            return ['blocked_until' => 0, 'history' => []];
        }

        $decoded = json_decode($payload, true);
        if (!is_array($decoded)) {
            return ['blocked_until' => 0, 'history' => []];
        }

        $blockedUntil = $decoded['blocked_until'] ?? 0;
        if (!is_int($blockedUntil)) {
            $blockedUntil = (int) $blockedUntil;
            if ($blockedUntil < 0) {
                $blockedUntil = 0;
            }
        }

        $history = $decoded['history'] ?? [];
        if (!is_array($history)) {
            $history = [];
        }

        return [
            'blocked_until' => $blockedUntil,
            'history' => $history,
        ];
    }

    private function normalizeHistory(array $history, int $now, int $timeWindowSeconds): array
    {
        $threshold = $now - $timeWindowSeconds;
        $filtered = [];

        foreach ($history as $entry) {
            if (!is_numeric($entry)) {
                continue;
            }

            $normalizedEntry = (int) $entry;
            if ($normalizedEntry >= $threshold) {
                $filtered[] = $normalizedEntry;
            }
        }

        return $filtered;
    }

    private function normalizePolicy(array $policy): array
    {
        $normalized = array_merge(self::DEFAULT_LIMIT_RULE, $policy);
        $normalized['max_requests'] = max(1, (int) $normalized['max_requests']);
        $normalized['time_window_seconds'] = max(1, (int) $normalized['time_window_seconds']);
        $normalized['block_seconds'] = max(1, (int) $normalized['block_seconds']);

        return $normalized;
    }

    private function mergePolicy(array $base, array $overrides): array
    {
        return array_merge($base, $overrides);
    }

    private function normalizeConfig(array $config): array
    {
        $normalized = [];
        $normalized['enabled'] = (bool) ($config['enabled'] ?? false);
        $normalized['default'] = $this->normalizePolicy($config['default'] ?? []);
        $normalized['suspicious'] = $this->normalizePolicy($config['suspicious'] ?? []);
        $normalized['scan'] = is_array($config['scan'] ?? []) ? $config['scan'] : [];
        if (isset($normalized['scan']['enabled']) && !is_bool($normalized['scan']['enabled'])) {
            $normalized['scan']['enabled'] = (bool) $normalized['scan']['enabled'];
        }
        if (isset($normalized['suspicious']['enabled']) && !is_bool($normalized['suspicious']['enabled'])) {
            $normalized['suspicious']['enabled'] = (bool) $normalized['suspicious']['enabled'];
        }

        $normalizedHeaders = is_array($config['headers'] ?? []) ? $config['headers'] : [];
        if (($normalizedHeaders['enabled'] ?? true) !== false) {
            $normalizedHeaders['enabled'] = (bool) $normalizedHeaders['enabled'];
        }
        if (($normalizedHeaders['require_user_agent'] ?? true) !== false) {
            $normalizedHeaders['require_user_agent'] = (bool) $normalizedHeaders['require_user_agent'];
        }
        if (($normalizedHeaders['require_accept_language'] ?? false) !== false) {
            $normalizedHeaders['require_accept_language'] = (bool) $normalizedHeaders['require_accept_language'];
        }

        $normalizedHeaders['min_user_agent_length'] = max(0, (int) ($normalizedHeaders['min_user_agent_length'] ?? 0));
        $normalizedHeaders['blocked_user_agents'] = $this->normalizeHeaderList(
            is_array($normalizedHeaders['blocked_user_agents'] ?? []) ? $normalizedHeaders['blocked_user_agents'] : []
        );
        $normalized['headers'] = $normalizedHeaders;

        $normalized['methods'] = [];
        $methods = $config['methods'] ?? [];
        if (is_array($methods)) {
            foreach ($methods as $method => $methodConfig) {
                if (!is_string($method) || !is_array($methodConfig)) {
                    continue;
                }

                $normalized['methods'][strtoupper($method)] = $this->normalizePolicy($methodConfig);
            }
        }

        $normalized['routes'] = [
            'exact' => [],
            'prefix' => [],
        ];

        $routes = $config['routes'] ?? [];
        if (is_array($routes)) {
            $exactRoutes = is_array($routes['exact'] ?? null) ? $routes['exact'] : [];
            foreach ($exactRoutes as $route => $routeConfig) {
                if (!is_string($route) || !is_array($routeConfig)) {
                    continue;
                }

                $normalized['routes']['exact'][$route] = $this->normalizePolicy($routeConfig);
            }

            $prefixRoutes = is_array($routes['prefix'] ?? null) ? $routes['prefix'] : [];
            foreach ($prefixRoutes as $prefix => $routeConfig) {
                if (!is_string($prefix) || !is_array($routeConfig)) {
                    continue;
                }

                $normalizedPrefix = $this->normalizePrefix($prefix);
                if ($normalizedPrefix === '') {
                    continue;
                }

                $normalized['routes']['prefix'][$normalizedPrefix] = $this->normalizePolicy($routeConfig);
            }
        }

        return $normalized;
    }

    private function writeState($handle, array $state): void
    {
        rewind($handle);
        ftruncate($handle, 0);
        fwrite($handle, json_encode($state, JSON_THROW_ON_ERROR));
    }

    private function ensureDirectory(): void
    {
        if (is_dir($this->basePath)) {
            return;
        }

        if (!mkdir($this->basePath, 0777, true) && !is_dir($this->basePath)) {
            throw new \RuntimeException('Cannot create defense storage directory: ' . $this->basePath);
        }
    }
}
