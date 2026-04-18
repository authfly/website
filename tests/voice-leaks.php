<?php

declare(strict_types = 1);

/**
 * Voice leak auditor.
 *
 * For a given ?voice=<id> (e.g. devops, itpop, vika) crawls every public
 * route, parses the rendered HTML and reports internal links that DO NOT
 * carry the expected ?voice=<id> query parameter — i.e. links through which
 * navigation would silently "fall out" of the voice context.
 *
 * Excluded by design (these are SUPPOSED to be voice-free):
 *   - <link rel="canonical">                  — canonical must point to base URL
 *   - <link rel="alternate" hreflang=...>     — language alternates
 *   - <meta property="og:url">                — OG points to canonical
 *   - .mk-lang-toggle                         — language switch exits voice
 *   - aside.mk-voice (voice switcher widget)  — its links are the switcher itself
 *   - external URLs (http(s) to other hosts)
 *   - assets (/static/*, /favicon.ico)
 *   - anchors, mailto:, tel:, data:, javascript:
 *
 * Usage:
 *   php -S 127.0.0.1:8765 -t public public/index.php &
 *   php tests/voice-leaks.php                         # all voices
 *   php tests/voice-leaks.php devops                  # one voice
 *   php tests/voice-leaks.php devops itpop --json     # multiple + JSON
 *   php tests/voice-leaks.php --base=http://localhost:8000 devops
 *
 * Exit codes: 0 — clean, 1 — leaks found, 2 — fatal (server unreachable etc.)
 */

// ---------------------------------------------------------------------------
// CLI args
// ---------------------------------------------------------------------------

$base = 'http://127.0.0.1:8765';
$asJson = false;
$voices = [];
$sampleSlug = null;

foreach (array_slice($argv, 1) as $arg) {
    if ($arg === '--json') {
        $asJson = true;
    } elseif (str_starts_with($arg, '--base=')) {
        $base = rtrim(substr($arg, 7), '/');
    } elseif (str_starts_with($arg, '--slug=')) {
        $sampleSlug = substr($arg, 7);
    } elseif ($arg[0] !== '-') {
        $voices[] = $arg;
    }
}

if ($voices === []) {
    $voices = ['vika', 'devops', 'itpop'];
}

// ---------------------------------------------------------------------------
// Build the route list from config/pages.php
// ---------------------------------------------------------------------------

$pagesConfig = require __DIR__ . '/../config/pages.php';

// Pick a real blog slug to substitute into /blog/@slug routes.
if ($sampleSlug === null) {
    $blogFixture = json_decode(
        (string) file_get_contents(__DIR__ . '/../fixtures/ru/blog.json'),
        true,
    );
    $sampleSlug = $blogFixture['posts'][0]['slug'] ?? 'why-auth-fly-own-your-idp';
}

$routes = [];
foreach (array_keys($pagesConfig) as $route) {
    // Deduplicate /blog and /blog/ — they render the same content.
    // We keep only the trailing-slash variant for the report.
    if (!str_ends_with($route, '/') && isset($pagesConfig[$route . '/'])) {
        continue;
    }
    $resolved = str_replace('@slug', $sampleSlug, $route);
    $routes[] = $resolved;
}
$routes = array_values(array_unique($routes));

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

/**
 * Safe substring for terminal output. Uses mbstring when available,
 * falls back to byte-based substr (with a small heuristic to avoid
 * breaking UTF-8 sequences in the output).
 */
function safe_cut(string $s, int $len): string
{
    if (function_exists('mb_substr')) {
        return mb_substr($s, 0, $len, 'UTF-8');
    }
    if (strlen($s) <= $len) {
        return $s;
    }
    $cut = substr($s, 0, $len);
    // Trim trailing partial UTF-8 byte to avoid garbled tail.
    while ($cut !== '' && (ord($cut[strlen($cut) - 1]) & 0xC0) === 0x80) {
        $cut = substr($cut, 0, -1);
    }
    return $cut;
}

// ---------------------------------------------------------------------------
// Leak detection
// ---------------------------------------------------------------------------

/**
 * @return array{
 *   leaks: list<array{page:string,href:string,tag:string,text:string,reason:string}>,
 *   ok_count: int,
 *   skipped_count: int,
 *   pages_scanned: int
 * }
 */
function auditVoice(string $base, string $voice, array $routes): array
{
    $expected = 'voice=' . $voice;
    $leaks = [];
    $okCount = 0;
    $skippedCount = 0;
    $pagesScanned = 0;
    $host = parse_url($base, PHP_URL_HOST);

    foreach ($routes as $route) {
        $url = $base . $route . (str_contains($route, '?') ? '&' : '?') . $expected;

        $ctx = stream_context_create(['http' => ['timeout' => 5, 'ignore_errors' => true]]);
        $html = @file_get_contents($url, false, $ctx);
        if ($html === false || $html === '') {
            $leaks[] = [
                'page' => $route,
                'href' => '',
                'tag' => '',
                'text' => '',
                'reason' => 'page-fetch-failed',
            ];
            continue;
        }
        $pagesScanned++;

        $dom = new DOMDocument();
        $prev = libxml_use_internal_errors(true);
        $dom->loadHTML('<?xml encoding="UTF-8">' . $html);
        libxml_clear_errors();
        libxml_use_internal_errors($prev);

        $xp = new DOMXPath($dom);

        // Anchors that are NOT inside excluded containers and NOT excluded by class.
        $nodes = $xp->query(
            '//a[@href]'
            . '[not(ancestor::aside[contains(concat(" ", normalize-space(@class), " "), " mk-voice ")])]'
            . '[not(contains(concat(" ", normalize-space(@class), " "), " mk-lang-toggle "))]'
        );

        foreach ($nodes as $a) {
            /** @var DOMElement $a */
            $href = trim($a->getAttribute('href'));

            if ($href === '' || $href[0] === '#') {
                $skippedCount++;
                continue;
            }
            if (preg_match('~^(mailto:|tel:|data:|javascript:)~i', $href)) {
                $skippedCount++;
                continue;
            }
            if (preg_match('~^https?://~i', $href)) {
                $linkHost = parse_url($href, PHP_URL_HOST);
                if ($linkHost !== null && $linkHost !== $host) {
                    $skippedCount++;
                    continue;
                }
                // Same-host absolute link — normalize to path+query for analysis.
                $href = (string) (parse_url($href, PHP_URL_PATH) ?? '/')
                    . (($q = parse_url($href, PHP_URL_QUERY)) !== null ? '?' . $q : '');
            }
            if (str_starts_with($href, '/static/') || $href === '/favicon.ico') {
                $skippedCount++;
                continue;
            }
            // Language switch URLs always exit voice — that's by design.
            if (preg_match('~[?&]lang=~', $href)) {
                $skippedCount++;
                continue;
            }

            // Internal link — must carry the expected voice.
            if (str_contains($href, $expected)) {
                $okCount++;
                continue;
            }

            // It might still carry a DIFFERENT voice — that's a stronger leak.
            $reason = preg_match('~[?&]voice=([a-z0-9_-]+)~i', $href, $m) === 1
                ? 'wrong-voice (' . $m[1] . ')'
                : 'missing-voice';

            $leaks[] = [
                'page' => $route,
                'href' => $href,
                'tag' => $a->nodeName,
                'text' => safe_cut(trim((string) preg_replace('/\s+/', ' ', $a->textContent)), 60),
                'reason' => $reason,
            ];
        }
    }

    return [
        'leaks' => $leaks,
        'ok_count' => $okCount,
        'skipped_count' => $skippedCount,
        'pages_scanned' => $pagesScanned,
    ];
}

// ---------------------------------------------------------------------------
// Run + report
// ---------------------------------------------------------------------------

$report = [];
$totalLeaks = 0;

foreach ($voices as $voice) {
    $result = auditVoice($base, $voice, $routes);
    $report[$voice] = $result;
    $totalLeaks += count($result['leaks']);
}

if ($asJson) {
    echo json_encode([
        'base' => $base,
        'routes' => $routes,
        'voices' => $report,
        'total_leaks' => $totalLeaks,
    ], JSON_PRETTY_PRINT | JSON_UNESCAPED_SLASHES | JSON_UNESCAPED_UNICODE), "\n";
    exit($totalLeaks === 0 ? 0 : 1);
}

echo "Voice leak audit\n";
echo "Base:   {$base}\n";
echo 'Routes: ' . count($routes) . ' (' . implode(', ', $routes) . ")\n\n";

foreach ($voices as $voice) {
    $r = $report[$voice];
    $count = count($r['leaks']);
    $marker = $count === 0 ? '[OK]   ' : '[FAIL] ';
    printf(
        "%s ?voice=%-7s pages=%d  ok_links=%d  skipped=%d  leaks=%d\n",
        $marker,
        $voice,
        $r['pages_scanned'],
        $r['ok_count'],
        $r['skipped_count'],
        $count,
    );

    if ($count === 0) {
        continue;
    }

    echo "\n";
    printf("       %-26s %-22s %-50s %s\n", 'PAGE', 'REASON', 'HREF', 'TEXT');
    printf("       %s\n", str_repeat('-', 120));
    foreach ($r['leaks'] as $l) {
        printf(
            "       %-26s %-22s %-50s %s\n",
            safe_cut($l['page'], 26),
            $l['reason'],
            safe_cut($l['href'], 50),
            $l['text'],
        );
    }
    echo "\n";
}

echo "\n";
echo $totalLeaks === 0
    ? "OK: no voice leaks found.\n"
    : "FAIL: {$totalLeaks} voice leak(s) found.\n";

exit($totalLeaks === 0 ? 0 : 1);
