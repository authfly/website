# Self-hosted fonts

WOFF2 files and `fonts.css` are inlined into the `<style>` tag in `templates/layout.latte`,
keeping the brand promise: 0 third-party requests, no telemetry.

## Bundled families (Auth Fly stack)

- **Nunito** — display / accents (700, 800).
- **Nunito Sans** — UI / body (400, 600).
- **JetBrains Mono** — code (400, 600).

Latin and Cyrillic subsets are bundled for both Nunito families
(filenames: `nunito-700-{lat,lat-ext,cyr,cyr-ext}.woff2`,
`nunitosans-{400,600}-{lat,lat-ext,cyr,cyr-ext}.woff2`).

JetBrains Mono ships with the original Google Fonts subset filenames (`tDbv2o-...woff2`).

## Updating

1. Request the same Google Fonts CSS2 URL with a **browser User-Agent** (so the response lists `woff2`, not TTF).
2. Extract unique `https://fonts.gstatic.com/...woff2` URLs and download into this folder.
3. Rewrite `src: url(...)` to `url(/static/fonts/<filename>.woff2)` (see existing `fonts.css`).
4. Adjust `<link rel="preload">` in `layout.latte` if primary subsets change.

All fonts are licensed under the **SIL Open Font License**.
