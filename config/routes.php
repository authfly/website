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
        ['id' => 'vika', 'label' => 'Вика', 'voice' => 'vika', 'lang' => null],
        ['id' => 'devops', 'label' => 'DevOps', 'voice' => 'devops', 'lang' => null],
        ['id' => 'itpop', 'label' => 'Айтишник', 'voice' => 'itpop', 'lang' => null],
        ['id' => 'burnout', 'label' => 'Senior', 'voice' => 'burnout', 'lang' => null],
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

/**
 * Catalogue of public voices (excluding the canonical SRE).
 * Single source of truth for the floating discovery FAB and the
 * /about/ voice cards section. Slogans are intentionally short
 * and lifted from each voice's own hero subtitle.
 *
 * @return list<array{id:string,label:string,avatar:string,slogan:string,url:string}>
 */
$voiceCatalogue = static function (): array {
    return [
        [
            'id' => 'vika',
            'label' => 'Вика',
            'avatar' => '/static/img/vika.png',
            'slogan' => 'Один пропуск — и человек у вас «свой» на каждой площадке.',
            'url' => '/about/?voice=vika',
        ],
        [
            'id' => 'devops',
            'label' => 'DevOps',
            'avatar' => '/static/img/devops.png',
            'slogan' => 'Один бинарь. Локальные ключи. Никаких звонков домой.',
            'url' => '/about/?voice=devops',
        ],
        [
            'id' => 'itpop',
            'label' => 'Айтишник',
            'avatar' => '/static/img/itpop.png',
            'slogan' => 'Один логин — много сервисов. Можно собрать самому.',
            'url' => '/about/?voice=itpop',
        ],
        [
            'id' => 'burnout',
            'label' => 'Senior',
            'avatar' => '/static/img/senior.png',
            'slogan' => 'Войти быстрее, чем заполнить форму согласия на обработку.',
            'url' => '/about/?voice=burnout',
        ],
    ];
};

/**
 * Floating circular avatar in the bottom-left corner that invites
 * the visitor to discover the alternative voices. Visible only on
 * the canonical RU pages: not shown on /about/ (cards live there),
 * not shown when ?voice= is already active, not shown for English.
 *
 * Avatar selection is deterministic per request path — same path
 * yields the same avatar across reloads, friendly to page caches,
 * but different paths show different avatars: visitor browsing the
 * site sees the rotation organically.
 *
 * @return array{visible:bool, avatar:string, label:string, url:string}
 */
$buildVoiceFab = static function (string $locale, string $routePath) use ($localeManager, $voiceCatalogue): array {
    $hidden = ['visible' => false, 'avatar' => '', 'label' => '', 'url' => ''];

    if ($locale !== 'ru' || $localeManager->isVoiceLocale($locale)) {
        return $hidden;
    }
    if (isset($_GET['voice']) && trim((string) $_GET['voice']) !== '') {
        return $hidden;
    }

    $requestUri = (string) ($_SERVER['REQUEST_URI'] ?? '/');
    $path = (string) (parse_url($requestUri, PHP_URL_PATH) ?: '/');
    if ($path === '' || $path === '/about/' || $path === '/about') {
        if ($path === '/about/' || $path === '/about') {
            return $hidden;
        }
        $path = '/';
    }

    $catalogue = $voiceCatalogue();
    $index = abs(crc32($path)) % count($catalogue);
    $voice = $catalogue[$index];

    return [
        'visible' => true,
        'avatar' => $voice['avatar'],
        'label' => 'Читать как: открыть карточки голосов',
        // Anchor jumps directly to the discovery section, skipping the /about/
        // hero — the FAB exists precisely to surface the voices.
        'url' => '/about/#voices',
    ];
};

/**
 * Voice discovery cards rendered on /about/ right after the hero.
 * Visible only on the canonical RU /about/ page; on ?voice= the
 * visitor has already chosen a tone — no point promoting siblings.
 *
 * @return array{visible:bool, items:list<array{id:string,label:string,avatar:string,slogan:string,url:string}>}
 */
$buildVoiceCards = static function (string $locale, string $routePath) use ($localeManager, $voiceCatalogue): array {
    $hidden = ['visible' => false, 'items' => []];

    if ($routePath !== '/about/' && $routePath !== '/about') {
        return $hidden;
    }
    if ($locale !== 'ru' || $localeManager->isVoiceLocale($locale)) {
        return $hidden;
    }
    if (isset($_GET['voice']) && trim((string) $_GET['voice']) !== '') {
        return $hidden;
    }

    return ['visible' => true, 'items' => $voiceCatalogue()];
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
        $buildVoiceWidget,
        $buildVoiceFab,
        $buildVoiceCards
    ): void {
        $routeParameters = $extractRouteParameters($routePath, $arguments);
        $locale = $resolveLocale();
        $nextLocale = $localeManager->getNextLocale($locale);
        $pageRenderer->setLocale($locale);
        $languageSwitchUrl = $buildLanguageSwitchUrl($nextLocale);
        $voiceWidget = $buildVoiceWidget($locale);
        $voiceFab = $buildVoiceFab($locale, $routePath);
        $voiceCards = $buildVoiceCards($locale, $routePath);
        $extraStylesheets = is_array($pageConfig['stylesheets'] ?? null) ? $pageConfig['stylesheets'] : [];
        $extraData = [
            'lang_switch_url' => $languageSwitchUrl,
            'lang_toggle_label' => $localeManager->getLabel($nextLocale),
            'html_lang' => $localeManager->getHtmlLang($locale),
            'og_locale' => $localeManager->getOpenGraphLocale($locale),
            'show_language_switch' => $localeManager->hasMultipleLocales(),
            'voice_widget' => $voiceWidget,
            'voice_fab' => $voiceFab,
            'voice_cards' => $voiceCards,
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
