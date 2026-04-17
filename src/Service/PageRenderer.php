<?php

declare(strict_types = 1);

namespace App\Service;

use PhpFasty\Core\Cache\CacheStore;
use PhpFasty\Core\Data\DataProviderInterface;
use PhpFasty\Core\Locale\LocaleAwareInterface;
use PhpFasty\Core\Locale\LocaleResolverInterface;
use PhpFasty\Core\View\TemplateRendererInterface;

final class PageRenderer
{
    private string $locale;

    public function __construct(
        private readonly DataProviderInterface $dataProvider,
        private readonly TemplateRendererInterface $renderer,
        private readonly CacheStore $cacheStore,
        private readonly LocaleResolverInterface $localeResolver
    ) {
        $this->locale = $this->localeResolver->getDefaultLocale();
    }

    public function setLocale(string $locale): void
    {
        $this->locale = $this->localeResolver->normalize($locale);

        if ($this->dataProvider instanceof LocaleAwareInterface) {
            $this->dataProvider->setLocale($this->locale);
        }
    }

    /**
     * @param array<int, string> $dataKeys
     * @param array<string, mixed> $extra
     */
    public function render(string $template, array $dataKeys, array $extra = [], ?string $locale = null): string
    {
        if ($locale !== null) {
            $this->setLocale($locale);
        }

        $templateData = array_replace($this->dataProvider->getMany($dataKeys), $extra);
        $pageHtml = $this->renderer->render('pages/' . $template, $templateData);

        $fontsCssPath = dirname(__DIR__, 2) . '/public/static/fonts/fonts.css';
        $inlineFontsCss = is_file($fontsCssPath) ? (string) file_get_contents($fontsCssPath) : '';

        return $this->renderer->render('layout.latte', array_merge($templateData, [
            'content' => $pageHtml,
            'title' => is_string($templateData['title'] ?? null) ? $templateData['title'] : 'Landing page',
            'description' => is_string($templateData['description'] ?? null) ? $templateData['description'] : null,
            'keywords' => is_string($templateData['keywords'] ?? null) ? $templateData['keywords'] : null,
            'robots' => is_string($templateData['robots'] ?? null) ? $templateData['robots'] : null,
            'canonical_url' => is_string($templateData['canonical_url'] ?? null)
                ? $templateData['canonical_url']
                : null,
            'lang' => $this->locale,
            'inline_fonts_css' => $inlineFontsCss,
        ]));
    }

    /**
     * @param array<int, string> $dataKeys
     * @param array<string, mixed> $extra
     */
    public function renderAndCache(string $cacheKey, string $template, array $dataKeys, array $extra = []): string
    {
        $cachedHtml = $this->cacheStore->get($cacheKey);
        if ($cachedHtml !== null) {
            return $cachedHtml;
        }

        $html = $this->render($template, $dataKeys, $extra);
        $this->cacheStore->set($cacheKey, $html);

        return $html;
    }

    /**
     * @param array<string, mixed> $pageConfig
     * @param array<string, string> $routeParameters
     */
    public function renderPage(
        string $routePath,
        array $pageConfig,
        array $routeParameters = [],
        bool $forceRefresh = false,
        ?string $locale = null,
        array $extraData = []
    ): string {
        if ($locale !== null) {
            $this->setLocale($locale);
        }

        $requestPath = $this->buildRequestPath($routePath, $routeParameters);
        $cacheKey = 'page:' . $requestPath . ':lang=' . $this->locale;
        $cacheKeySuffix = $extraData['cache_key_suffix'] ?? '';
        if (is_string($cacheKeySuffix) && $cacheKeySuffix !== '') {
            $cacheKey .= $cacheKeySuffix;
        }

        if ($forceRefresh) {
            $this->cacheStore->invalidate($cacheKey);
        }

        $pageData = $this->buildPageData($routePath, $pageConfig, $routeParameters, $extraData);

        return $this->renderAndCache(
            $cacheKey,
            (string) ($pageConfig['template'] ?? 'home.latte'),
            (array) ($pageConfig['data'] ?? []),
            $pageData
        );
    }

    /**
     * @param array<string, mixed> $pageConfig
     * @return array<int, array<string, string>>
     */
    public function getRouteParameterSets(array $pageConfig): array
    {
        $dynamicConfig = $pageConfig['dynamic'] ?? null;
        if (!is_array($dynamicConfig)) {
            return [[]];
        }

        $datasetKey = (string) ($dynamicConfig['dataset'] ?? '');
        $collectionPath = (string) ($dynamicConfig['collection'] ?? '');
        $lookupKey = (string) ($dynamicConfig['lookup'] ?? 'slug');
        $parameterName = (string) ($dynamicConfig['param'] ?? 'slug');

        $dataset = $this->dataProvider->get($datasetKey);
        $collection = $this->extractValue($dataset, $collectionPath);
        if (!is_array($collection)) {
            return [];
        }

        $parameterSets = [];

        foreach ($collection as $item) {
            if (!is_array($item)) {
                continue;
            }

            $parameterValue = $item[$lookupKey] ?? null;
            if (!is_scalar($parameterValue)) {
                continue;
            }

            $parameterSets[] = [
                $parameterName => trim((string) $parameterValue),
            ];
        }

        return $parameterSets;
    }

    /**
     * @param array<string, string> $routeParameters
     */
    public function buildRequestPath(string $routePath, array $routeParameters = []): string
    {
        $resolvedPath = $routePath;

        foreach ($routeParameters as $name => $value) {
            $resolvedPath = str_replace('@' . $name, trim($value), $resolvedPath);
        }

        if ($resolvedPath === '') {
            return '/';
        }

        return $resolvedPath === '/' ? '/' : '/' . trim($resolvedPath, '/');
    }

    /**
     * @param array<string, mixed> $pageConfig
     * @param array<string, string> $routeParameters
     * @param array<string, mixed> $extraData
     * @return array<string, mixed>
     */
    private function buildPageData(
        string $routePath,
        array $pageConfig,
        array $routeParameters,
        array $extraData
    ): array {
        $dataKeys = (array) ($pageConfig['data'] ?? []);
        $datasets = $this->dataProvider->getMany($dataKeys);
        $dynamicData = $this->resolveDynamicData($pageConfig, $datasets, $routeParameters);
        $pageMeta = [
            'hide_layout' => (bool) ($pageConfig['hide_layout'] ?? false),
            'extra_stylesheets' => is_array($pageConfig['stylesheets'] ?? null)
                ? $pageConfig['stylesheets']
                : [],
        ];
        $pageData = array_replace($datasets, $dynamicData, $pageMeta, $routeParameters, $extraData);
        $pageData['lang'] = $this->locale;
        $pageData['title'] = $this->resolveTemplate((string) ($pageConfig['title'] ?? 'Landing page'), $pageData);
        $pageData['description'] = $this->resolveDescription(
            (string) ($pageConfig['description'] ?? ''),
            $pageData
        );
        $pageData['keywords'] = $this->resolveTemplate((string) ($pageConfig['keywords'] ?? ''), $pageData);
        $pageData['robots'] = (string) ($pageConfig['robots'] ?? '');

        $canonicalBase = $this->extractValue($pageData, 'site.seo.canonical_base');
        $canonicalBase = is_scalar($canonicalBase) ? trim((string) $canonicalBase) : '';
        $requestPath = $this->buildRequestPath($routePath, $routeParameters);
        $normalizedPath = ($requestPath !== '/' && str_ends_with($requestPath, '/'))
            ? rtrim($requestPath, '/')
            : $requestPath;
        if ($canonicalBase !== '') {
            $pageData['canonical_url'] = rtrim($canonicalBase, '/') . $normalizedPath;
        } else {
            $pageData['canonical_url'] = $normalizedPath;
        }

        return $pageData;
    }

    /**
     * @param array<string, mixed> $pageConfig
     * @param array<string, array> $datasets
     * @param array<string, string> $routeParameters
     * @return array<string, mixed>
     */
    private function resolveDynamicData(array $pageConfig, array $datasets, array $routeParameters): array
    {
        $dynamicConfig = $pageConfig['dynamic'] ?? null;
        if (!is_array($dynamicConfig)) {
            return [];
        }

        $parameterName = (string) ($dynamicConfig['param'] ?? 'slug');
        $parameterValue = $routeParameters[$parameterName] ?? '';
        if ($parameterValue === '') {
            throw new \RuntimeException('Missing dynamic route parameter: ' . $parameterName);
        }

        $datasetKey = (string) ($dynamicConfig['dataset'] ?? '');
        $collectionPath = (string) ($dynamicConfig['collection'] ?? '');
        $lookupKey = (string) ($dynamicConfig['lookup'] ?? 'slug');
        $itemKey = (string) ($dynamicConfig['item'] ?? 'item');

        $dataset = $datasets[$datasetKey] ?? [];
        $collection = $this->extractValue($dataset, $collectionPath);
        if (!is_array($collection)) {
            throw new \RuntimeException('Dynamic collection not found for route.');
        }

        foreach ($collection as $item) {
            if (!is_array($item)) {
                continue;
            }

            $lookupValue = $item[$lookupKey] ?? null;
            if (!is_scalar($lookupValue)) {
                continue;
            }

            if ((string) $lookupValue === $parameterValue) {
                return [$itemKey => $item];
            }
        }

        throw new \RuntimeException('Dynamic page item was not found.');
    }

    /**
     * @param array<string, mixed> $context
     */
    private function resolveTitle(string $pattern, array $context): string
    {
        if ($pattern === '') {
            return 'Landing page';
        }

        $resolved = $this->resolveTemplate($pattern, $context);

        return $resolved === '' ? 'Landing page' : $resolved;
    }

    private function resolveDescription(string $pattern, array $context): string
    {
        if ($pattern !== '') {
            $resolved = $this->resolveTemplate($pattern, $context);
            if ($resolved !== '') {
                return $resolved;
            }
        }

        $fallback = $this->extractValue($context, 'site.description');
        return is_scalar($fallback) ? (string) $fallback : '';
    }

    private function resolveTemplate(string $pattern, array $context): string
    {
        if ($pattern === '') {
            return '';
        }

        $resolved = preg_replace_callback('/\{([a-zA-Z0-9_.]+)\}/', function (array $matches) use ($context): string {
            $value = $this->extractValue($context, $matches[1]);

            return is_scalar($value) ? (string) $value : '';
        }, $pattern);

        return is_string($resolved) ? $resolved : '';
    }

    /**
     * @param array<string, mixed> $source
     */
    private function extractValue(array $source, string $path): mixed
    {
        if ($path === '') {
            return $source;
        }

        $segments = explode('.', $path);
        $value = $source;

        foreach ($segments as $segment) {
            if (!is_array($value) || !array_key_exists($segment, $value)) {
                return null;
            }

            $value = $value[$segment];
        }

        return $value;
    }
}
