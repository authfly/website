# authfly/website

The Authfly documentation site, https://authfly.vercel.app: onboarding to Authfly + Stalwart for developers and AI
agents, in English and Russian. Go `net/http` + [templ](https://templ.guide), rendered only with UI8Kit 3 templ
primitives and the Authfly templ layer from [authfly/authkit](https://github.com/authfly/authkit)
(`github.com/authfly/authkit/packages/ui/templ`, pinned in `go.mod`).

## Layout

| Path | What |
|---|---|
| `content/{en,ru}/**/*.md` | the pages: one Markdown file per page and locale, front matter `title`, `nav`, `description` |
| `internal/site` | loader, navigation (`Nav`), routes, `llms.txt`, sitemap; `i18n/{en,ru}.json` UI strings |
| `internal/md` | the Markdown subset parser (no HTML output; the views render the block tree with primitives) |
| `internal/views` | templ views: document, header, sidebar, article, home, 404 |
| `assets/` | `app.css` (built, committed), `theme.js`, `favicon.svg` |
| `styles/` | `theme.css` (Authfly tokens, vendored from authkit), `site.css` (document-level rules) |
| `examples/react-docker` | the tested React + Docker app of the guide |

## Routes

`/{en,ru}/<page>` HTML, `/{en,ru}/<page>.md` the raw Markdown source, `/llms.txt` and `/ru/llms.txt` (index),
`/llms-full.txt` and `/ru/llms-full.txt` (every page in one file), `/sitemap.xml`. `/` redirects by
`Accept-Language`; paths of the previous site redirect to their new pages.

## Develop

```sh
go install github.com/a-h/templ/cmd/templ@v0.3.1070
bun install
templ generate && bash tools/build-css.sh && go run ./cmd/server   # http://localhost:3011
go test ./...                                                        # en/ru in step, links, anchors, rendering
bunx ui8px lint                                                      # must report 0
```

Add a page: write `content/en/<slug>.md` and `content/ru/<slug>.md`, add the slug to `Nav` in
`internal/site/site.go`. Tests fail if a locale misses a page, a dictionary key, a code block or a table, or if a
link or anchor does not resolve.

Rules: [UI8Kit contract](https://github.com/ui8kit/contract) (primitives only, tokens only, Lucide icons, 8+4 grid,
no raw HTML, texts from content or the dictionary). Facts in the docs come from the Authfly source; mark anything not
checked against code as UNVERIFIED / НЕ ПРОВЕРЕНО.

## Deploy

Vercel (Go runtime, `vercel.json`): every push to `main` deploys production. Generated `*_templ.go` and
`assets/app.css` are committed, so the build is `go build` only. Self-hosting: `Dockerfile`.

License: Apache-2.0, © Authfly.
