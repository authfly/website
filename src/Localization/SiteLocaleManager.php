<?php

declare(strict_types = 1);

namespace App\Localization;

use PhpFasty\Core\Locale\LocaleResolverInterface;

final class SiteLocaleManager implements LocaleResolverInterface
{
    /** @var array<string, array{label:string, html_lang:string, og_locale:string}> */
    private array $supportedLocales = [];

    private string $defaultLocale;

    /**
     * @param array{
     *     default?: string,
     *     supported: array<string, array{label?: string, html_lang?: string, og_locale?: string}>
     * } $config
     */
    public function __construct(array $config)
    {
        $supported = $config['supported'] ?? [];
        foreach ($supported as $locale => $meta) {
            $normalizedLocale = $this->normalizeToken($locale);
            if ($normalizedLocale === '') {
                continue;
            }

            $this->supportedLocales[$normalizedLocale] = [
                'label' => (string) ($meta['label'] ?? strtoupper($normalizedLocale)),
                'html_lang' => (string) ($meta['html_lang'] ?? $normalizedLocale),
                'og_locale' => (string) ($meta['og_locale'] ?? str_replace('-', '_', $normalizedLocale)),
            ];
        }

        if ($this->supportedLocales === []) {
            throw new \RuntimeException('At least one supported locale must be configured.');
        }

        $configuredDefault = $this->normalizeToken((string) ($config['default'] ?? ''));
        $this->defaultLocale = isset($this->supportedLocales[$configuredDefault])
            ? $configuredDefault
            : array_key_first($this->supportedLocales);
    }

    public function normalize(string $locale): string
    {
        $normalized = $this->normalizeToken($locale);
        if ($normalized === '') {
            return $this->defaultLocale;
        }

        if (isset($this->supportedLocales[$normalized])) {
            return $normalized;
        }

        $primaryTag = explode('-', $normalized, 2)[0];
        if (isset($this->supportedLocales[$primaryTag])) {
            return $primaryTag;
        }

        return $this->defaultLocale;
    }

    public function getDefaultLocale(): string
    {
        return $this->defaultLocale;
    }

    /**
     * @return array<int, string>
     */
    public function getSupportedLocales(): array
    {
        return array_keys($this->supportedLocales);
    }

    /**
     * Public locales eligible for the language switcher.
     * Voice pseudo-locales (voice/*) are excluded.
     *
     * @return array<int, string>
     */
    public function getPublicLocales(): array
    {
        return array_values(array_filter(
            array_keys($this->supportedLocales),
            static fn (string $locale): bool => !str_starts_with($locale, 'voice/')
        ));
    }

    public function isVoiceLocale(string $locale): bool
    {
        $normalized = $this->normalizeToken($locale);

        return $normalized !== '' && str_starts_with($normalized, 'voice/');
    }

    public function hasMultipleLocales(): bool
    {
        return count($this->getPublicLocales()) > 1;
    }

    public function resolveRequestLocale(
        ?string $requestedLocale,
        ?string $cookieLocale,
        ?string $acceptLanguage
    ): string {
        foreach ([$requestedLocale, $cookieLocale] as $candidate) {
            if (!is_string($candidate) || trim($candidate) === '') {
                continue;
            }

            $normalized = $this->normalize($candidate);
            if ($normalized !== $this->defaultLocale || $this->isExplicitlySupported($candidate)) {
                return $normalized;
            }
        }

        foreach ($this->extractAcceptLanguageCandidates($acceptLanguage) as $candidate) {
            $normalized = $this->normalize($candidate);
            if ($normalized !== $this->defaultLocale || $this->isExplicitlySupported($candidate)) {
                return $normalized;
            }
        }

        return $this->defaultLocale;
    }

    public function getNextLocale(string $locale): string
    {
        $publicLocales = $this->getPublicLocales();
        if (count($publicLocales) < 2) {
            return $publicLocales[0] ?? $this->defaultLocale;
        }

        $current = $this->normalize($locale);
        // Voice pseudo-locales render Russian content, so the toggle should
        // behave as if the user were on the Russian page — and offer the
        // English alternative as the next step.
        if ($this->isVoiceLocale($current)) {
            $current = isset($this->supportedLocales['ru']) ? 'ru' : $this->defaultLocale;
        }

        $currentIndex = array_search($current, $publicLocales, true);
        if ($currentIndex === false) {
            return $publicLocales[0];
        }

        $nextIndex = ($currentIndex + 1) % count($publicLocales);

        return $publicLocales[$nextIndex];
    }

    public function getLabel(string $locale): string
    {
        $normalized = $this->normalize($locale);

        return $this->supportedLocales[$normalized]['label'];
    }

    public function getHtmlLang(string $locale): string
    {
        $normalized = $this->normalize($locale);

        return $this->supportedLocales[$normalized]['html_lang'];
    }

    public function getOpenGraphLocale(string $locale): string
    {
        $normalized = $this->normalize($locale);

        return $this->supportedLocales[$normalized]['og_locale'];
    }

    private function normalizeToken(string $locale): string
    {
        $normalized = strtolower(trim($locale));
        if ($normalized === '') {
            return '';
        }

        return str_replace('_', '-', $normalized);
    }

    private function isExplicitlySupported(string $locale): bool
    {
        $normalized = $this->normalizeToken($locale);
        if ($normalized === '') {
            return false;
        }

        if (isset($this->supportedLocales[$normalized])) {
            return true;
        }

        $primaryTag = explode('-', $normalized, 2)[0];

        return isset($this->supportedLocales[$primaryTag]);
    }

    /**
     * @return array<int, string>
     */
    private function extractAcceptLanguageCandidates(?string $acceptLanguage): array
    {
        if (!is_string($acceptLanguage) || trim($acceptLanguage) === '') {
            return [];
        }

        $candidates = [];
        foreach (explode(',', $acceptLanguage) as $part) {
            $candidate = trim(explode(';', $part, 2)[0]);
            if ($candidate !== '') {
                $candidates[] = $candidate;
            }
        }

        return $candidates;
    }
}
