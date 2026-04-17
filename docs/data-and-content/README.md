# Data and content

How content is sourced, bound to pages, and how much you can change without touching `src/`.

## DataProviderInterface

- **Contract**: `get(string $key): array`, `has(string $key): bool`, `getMany(array $keys): array`. Keys are logical names (e.g. `site`, `navigation`, `landing`, `blog`).
- **Current implementation**: `FixtureDataProvider` reads JSON from `FIXTURES_PATH`. Key `landing` for locale `en` → `{basePath}/en/landing.json`, falling back to `{basePath}/landing.json` if the localized file is absent. Returns decoded array or empty array on missing/invalid file.
- **Switching source**: To use an API or another backend, implement `DataProviderInterface` and register it in `config/services.php` (e.g. via `DATA_SOURCE` or a new binding). PageRenderer and the rest of the app depend only on the interface.

## How pages get their data

- **`config/pages.php`** defines per-route: `template`, `title` (pattern with placeholders like `{site.name}`), `data` (array of keys, e.g. `['site', 'navigation', 'landing']`), and optionally `dynamic` (param, dataset, collection, lookup, item).
- **PageRenderer** loads data with `$dataProvider->getMany($pageConfig['data'])`, merges in dynamic resolution (e.g. current blog post), resolves the title pattern, then passes the merged array to Latte. So templates receive a single data array; keys match the fixture keys and any extra keys from the dynamic block (e.g. `post`).

## Content flexibility

- **Changing content (same shape)**: Edit JSON under `fixtures/en/` (or whatever locales you define in `config/localization.php`). The stock project uses **one** locale folder (`en`); adding more locales requires updating localization config and adding matching `fixtures/<locale>/` trees. As long as structure and keys stay compatible with templates, no code changes.
- **Changing theme / layout**: Edit `templates/` (layout and page templates) and static assets. No change in `src/` required.
- **Adding/removing pages or routes**: Edit `config/pages.php` (and add a template if needed). Routes and warmup both follow `pages_config`.
- **Changing data shape**: If you rename or restructure fixture keys (e.g. `hero` → `banner`), templates must be updated to use the new keys. The app does not enforce a schema; it only passes through arrays. So “change types of data” is flexible at the config/template level as long as templates and page config stay in sync with the actual structure.

## Limits

- **Templates are coupled to structure**: They reference concrete keys (e.g. `$landing['hero']['title']`). Changing structure implies template changes.
- **API routes**: Some endpoints (e.g. `GET /api/landing`) are still hardcoded in `config/routes.php` with a fixed path and data key. Moving to a config-driven API map would reduce that coupling further.
- **Single provider at a time**: The container binds one `DataProviderInterface` implementation. Multi-source or per-key providers would require an extra abstraction or configuration layer.

## Summary

Architecture is **flexible for content, theme, and page set** when changes stay within config, fixtures, and templates. Changing the **data source type** (fixtures → API) is supported by design via a new provider. Changing **data shape** requires coordinated updates in fixtures and templates; there is no separate schema layer.
