<?php

declare(strict_types=1);

require __DIR__ . '/../vendor/autoload.php';

use App\Localization\SiteLocaleManager;
use PhpFasty\Core\Cache\CacheStore;
use PhpFasty\Core\Application;
use App\Service\PageRenderer;
use Dotenv\Dotenv;

$rootDir = dirname(__DIR__);

if (is_readable($rootDir . '/.env')) {
    Dotenv::createImmutable($rootDir)->load();
}

$rootDir = dirname(__DIR__);
$application = Application::getInstance($rootDir, $rootDir . '/config/services.php');
$container = $application->getContainer();

/** @var CacheStore $cacheStore */
$cacheStore = $container->get(CacheStore::class);
/** @var PageRenderer $pageRenderer */
$pageRenderer = $container->get(PageRenderer::class);
/** @var SiteLocaleManager $localeManager */
$localeManager = $container->get(SiteLocaleManager::class);
/** @var array<string, array<string, mixed>> $pages */
$pages = $container->get('pages_config');

$cacheStore->flush();

$warmedPaths = [];

foreach ($localeManager->getSupportedLocales() as $locale) {
    $pageRenderer->setLocale($locale);
    $nextLocale = $localeManager->getNextLocale($locale);
    foreach ($pages as $routePath => $pageConfig) {
        foreach ($pageRenderer->getRouteParameterSets($pageConfig) as $routeParameters) {
            $path = $pageRenderer->buildRequestPath($routePath, $routeParameters);
            $langSwitchUrl = $path . '?lang=' . $nextLocale;
            $extraData = [
                'lang_switch_url' => $langSwitchUrl,
                'lang_toggle_label' => $localeManager->getLabel($nextLocale),
                'html_lang' => $localeManager->getHtmlLang($locale),
                'og_locale' => $localeManager->getOpenGraphLocale($locale),
                'show_language_switch' => $localeManager->hasMultipleLocales(),
            ];
            $pageRenderer->renderPage($routePath, $pageConfig, $routeParameters, true, $locale, $extraData);
            $warmedPaths[] = $path . ' [' . $locale . ']';
        }
    }
}

echo 'Page cache warmed for ' . count($warmedPaths) . ' route(s).' . PHP_EOL;

foreach ($warmedPaths as $path) {
    echo ' - ' . $path . PHP_EOL;
}
