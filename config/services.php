<?php

declare(strict_types = 1);

use PhpFasty\Core\Cache\CacheStore;
use PhpFasty\Core\Container;
use PhpFasty\Core\Data\DataProviderInterface;
use PhpFasty\Core\Data\FixtureDataProvider;
use PhpFasty\Core\Locale\LocaleResolverInterface;
use App\Defense\RequestDefenseService;
use App\Localization\SiteLocaleManager;
use App\View\LatteRenderer;
use PhpFasty\Core\Middleware\SecurityHeaders;
use App\Service\PageRenderer;
use PhpFasty\Core\View\TemplateRendererInterface;

return function (Container $container): void {
    $rootDir = dirname(__DIR__);

    $resolvePath = static function (string $path) use ($rootDir): string {
        $normalizedPath = trim($path);
        if ($normalizedPath === '') {
            return $rootDir;
        }

        if (preg_match('/^(?:[A-Za-z]:[\\\\\\/]|[\\\\\\/]{2}|\\/)/', $normalizedPath) === 1) {
            return $normalizedPath;
        }

        $relativePath = preg_replace('/^[.][\\\\\\/]/', '', $normalizedPath);
        $relativePath = $relativePath === null ? $normalizedPath : $relativePath;
        $relativePath = str_replace(['/', '\\'], DIRECTORY_SEPARATOR, $relativePath);

        return $rootDir . DIRECTORY_SEPARATOR . ltrim($relativePath, DIRECTORY_SEPARATOR);
    };

    $fixturesPath = $resolvePath((string) (getenv('FIXTURES_PATH') ?: './fixtures'));
    $cacheDir = $resolvePath((string) (getenv('CACHE_DIR') ?: './cache'));
    $templatesDir = $rootDir . DIRECTORY_SEPARATOR . 'templates';
    $templateCacheDir = $templatesDir . DIRECTORY_SEPARATOR . 'cache';
    $defenseConfigPath = $resolvePath((string) (getenv('DEFENSE_CONFIG_PATH') ?: './config/defense.php'));
    $cacheTtl = (int) (getenv('CACHE_TTL') ?: 3600);
    $dataSource = (string) (getenv('DATA_SOURCE') ?: 'fixtures');

    $container->bind('app.root_dir', $rootDir);
    $container->bind('app.templates_dir', $templatesDir);
    $container->bind('app.template_cache_dir', $templateCacheDir);
    $container->bind('app.cache_dir', $cacheDir);
    $container->bind('fixtures_path', $fixturesPath);
    $container->bind('localization_config', require $rootDir . '/config/localization.php');
    $container->bind('pages_config', require $rootDir . '/config/pages.php');
    $container->bind('defense_config', require $defenseConfigPath);

    $container->singleton(SecurityHeaders::class, static fn (): SecurityHeaders => new SecurityHeaders());

    $container->singleton(
        SiteLocaleManager::class,
        static fn (Container $container): SiteLocaleManager => new SiteLocaleManager(
            (array) $container->get('localization_config')
        )
    );

    $container->singleton(
        LocaleResolverInterface::class,
        static fn (Container $container): LocaleResolverInterface => $container->get(SiteLocaleManager::class)
    );

    $container->singleton(
        TemplateRendererInterface::class,
        static fn (Container $container): LatteRenderer => new LatteRenderer(
            (string) $container->get('app.templates_dir'),
            (string) $container->get('app.template_cache_dir')
        )
    );

    $container->singleton(
        CacheStore::class,
        static fn (Container $container): CacheStore => new CacheStore(
            (string) $container->get('app.cache_dir'),
            $cacheTtl
        )
    );

    $container->singleton(
        DataProviderInterface::class,
        static fn (Container $container): DataProviderInterface => match ($dataSource) {
            'fixtures' => new FixtureDataProvider(
                $fixturesPath,
                $container->get(LocaleResolverInterface::class)
            ),
            default => throw new RuntimeException('Unsupported DATA_SOURCE: ' . $dataSource),
        }
    );

    $container->singleton(
        PageRenderer::class,
        static fn (Container $container): PageRenderer => new PageRenderer(
            $container->get(DataProviderInterface::class),
            $container->get(TemplateRendererInterface::class),
            $container->get(CacheStore::class),
            $container->get(LocaleResolverInterface::class)
        )
    );

    $container->singleton(
        RequestDefenseService::class,
        static fn (Container $container): RequestDefenseService => new RequestDefenseService(
            (string) $container->get('app.cache_dir'),
            (array) $container->get('defense_config')
        )
    );
};
