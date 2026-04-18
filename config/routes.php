<?php

declare(strict_types = 1);

use App\Localization\SiteLocaleManager;
use PhpFasty\Core\ContainerInterface;
use PhpFasty\Core\Data\DataProviderInterface;
use PhpFasty\Core\Middleware\SecurityHeaders;
use App\Service\PageRenderer;

$container = Flight::get('appContainer');
if (!$container instanceof ContainerInterface) {
    throw new RuntimeException('Application container is not available.');
}

/** @var SecurityHeaders $securityHeaders */
$securityHeaders = $container->get(SecurityHeaders::class);
/** @var DataProviderInterface $dataProvider */
$dataProvider = $container->get(DataProviderInterface::class);
/** @var PageRenderer $pageRenderer */
$pageRenderer = $container->get(PageRenderer::class);
/** @var SiteLocaleManager $localeManager */
$localeManager = $container->get(SiteLocaleManager::class);
/** @var array<string, array<string, mixed>> $pages */
$pages = $container->get('pages_config');

$extractRouteParameters = static function (string $routePath, array $arguments): array {
    $parameterNames = [];
    if (preg_match_all('/@([a-zA-Z_][a-zA-Z0-9_]*)/', $routePath, $matches) === 1) {
        $parameterNames = $matches[1];
    }

    $parameters = [];

    foreach ($parameterNames as $index => $parameterName) {
        $argument = $arguments[$index] ?? '';
        $parameters[$parameterName] = is_scalar($argument) ? (string) $argument : '';
    }

    return $parameters;
};

$resolveLocale = static function () use ($localeManager): string {
    // ?voice=<id> takes precedence over ?lang and persisted cookie.
    // Voices are registered as pseudo-locales (voice/<id>); unknown voices
    // gracefully fall back to the standard lang resolution.
    $voiceParam = trim((string) ($_GET['voice'] ?? ''));
    if ($voiceParam !== '') {
        $voiceLocale = 'voice/' . strtolower($voiceParam);
        $normalized = $localeManager->normalize($voiceLocale);
        if ($normalized === $voiceLocale) {
            return $normalized;
        }
    }

    $resolvedLocale = $localeManager->resolveRequestLocale(
        (string) ($_GET['lang'] ?? ''),
        (string) ($_COOKIE['site-lang'] ?? ''),
        (string) ($_SERVER['HTTP_ACCEPT_LANGUAGE'] ?? '')
    );

    // Persist only real locales; voice pseudo-locales are stateless.
    if (!$localeManager->isVoiceLocale($resolvedLocale)) {
        setcookie('site-lang', $resolvedLocale, [
            'expires' => time() + 60 * 60 * 24 * 365,
            'path' => '/',
            'samesite' => 'Lax',
        ]);
    }

    return $resolvedLocale;
};

$buildLanguageSwitchUrl = static function (string $locale): string {
    $requestUri = (string) ($_SERVER['REQUEST_URI'] ?? '/');
    $path = (string) (parse_url($requestUri, PHP_URL_PATH) ?: '/');
    if ($path === '') {
        $path = '/';
    }

    $query = $_GET;
    // Switching language always exits voice mode.
    unset($query['lang'], $query['voice']);
    $query['lang'] = $locale;

    return $path . '?' . http_build_query($query);
};

/**
 * Build the "Read as" voice-switcher widget data: SRE (canonical RU),
 * DevOps and IT-Pop (?voice=<id>). Visible only in the Russian context;
 * the canonical option always strips ?voice= and forces ?lang=ru.
 *
 * @return array{visible:bool, current:string, items:array<int, array{id:string,label:string,url:string,active:bool}>}
 */
$buildVoiceWidget = static function (string $locale) use ($localeManager): array {
    $htmlLang = $localeManager->getHtmlLang($locale);
    if ($htmlLang !== 'ru') {
        return ['visible' => false, 'current' => '', 'items' => []];
    }

    $requestUri = (string) ($_SERVER['REQUEST_URI'] ?? '/');
    $path = (string) (parse_url($requestUri, PHP_URL_PATH) ?: '/');
    if ($path === '') {
        $path = '/';
    }

    $current = $localeManager->isVoiceLocale($locale)
        ? substr($locale, strlen('voice/'))
        : 'sre';

    $voices = [
        ['id' => 'sre', 'label' => 'SRE', 'voice' => null, 'lang' => 'ru'],
        ['id' => 'vika', 'label' => 'Vika', 'voice' => 'vika', 'lang' => null],
        ['id' => 'devops', 'label' => 'DevOps', 'voice' => 'devops', 'lang' => null],
        ['id' => 'itpop', 'label' => 'Айтишник', 'voice' => 'itpop', 'lang' => null],
        ['id' => 'burnout', 'label' => 'Синьор', 'voice' => 'burnout', 'lang' => null],
    ];

    $items = [];
    foreach ($voices as $v) {
        $query = $_GET;
        unset($query['voice'], $query['lang']);
        if ($v['voice'] !== null) {
            $query['voice'] = $v['voice'];
        }
        if ($v['lang'] !== null) {
            $query['lang'] = $v['lang'];
        }

        $url = $query !== [] ? $path . '?' . http_build_query($query) : $path;

        $items[] = [
            'id' => $v['id'],
            'label' => $v['label'],
            'url' => $url,
            'active' => $v['id'] === $current,
        ];
    }

    return ['visible' => true, 'current' => $current, 'items' => $items];
};

Flight::route('GET /api/health', static function () use ($securityHeaders): void {
    $securityHeaders->applyApiHeaders();

    Flight::json([
        'status' => 'ok',
    ]);
});

Flight::route('GET /api/landing', static function () use ($dataProvider, $securityHeaders): void {
    $securityHeaders->applyApiHeaders();
    Flight::json($dataProvider->get('landing'));
});

foreach ($pages as $routePath => $pageConfig) {
    Flight::route('GET ' . $routePath, static function (...$arguments) use (
        $extractRouteParameters,
        $pageConfig,
        $pageRenderer,
        $routePath,
        $securityHeaders,
        $localeManager,
        $resolveLocale,
        $buildLanguageSwitchUrl,
        $buildVoiceWidget
    ): void {
        $routeParameters = $extractRouteParameters($routePath, $arguments);
        $locale = $resolveLocale();
        $nextLocale = $localeManager->getNextLocale($locale);
        $pageRenderer->setLocale($locale);
        $languageSwitchUrl = $buildLanguageSwitchUrl($nextLocale);
        $voiceWidget = $buildVoiceWidget($locale);
        $extraStylesheets = is_array($pageConfig['stylesheets'] ?? null) ? $pageConfig['stylesheets'] : [];
        $extraData = [
            'lang_switch_url' => $languageSwitchUrl,
            'lang_toggle_label' => $localeManager->getLabel($nextLocale),
            'html_lang' => $localeManager->getHtmlLang($locale),
            'og_locale' => $localeManager->getOpenGraphLocale($locale),
            'show_language_switch' => $localeManager->hasMultipleLocales(),
            'voice_widget' => $voiceWidget,
            'extra_stylesheets' => $extraStylesheets,
            'hide_layout' => $pageConfig['hide_layout'] ?? false,
        ];

        try {
            $html = $pageRenderer->renderPage(
                $routePath,
                $pageConfig,
                $routeParameters,
                false,
                $locale,
                $extraData
            );
        } catch (RuntimeException) {
            Flight::notFound();

            return;
        }

        $securityHeaders->applyStaticHeaders();
        Flight::response()->header('Content-Type', 'text/html; charset=UTF-8');
        echo $html;
    });
}
