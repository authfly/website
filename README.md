# AuthFly marketing site

Go + templ SSR site, patterned after [YMAIL_SITE](../_@Go/@YMAIL_SITE). Content is the former PHP **burnout** voice (RU) and the English fixture pack (EN). Voice switcher is gone.

## Locales / deploy

| Target | Locale | Fixtures |
| --- | --- | --- |
| Docker / VPS | `ru` | `internal/adapters/fixture/data` (from `.old-php/fixtures/voice/burnout`) |
| Vercel | `en` | `internal/adapters/fixture/data-en` (from `.old-php/fixtures/en`) |

```bash
# Russian (default)
export APP_DEFAULT_LOCALE=ru
export FIXTURE_DIR=internal/adapters/fixture/data

# English
export APP_DEFAULT_LOCALE=en
export FIXTURE_DIR=internal/adapters/fixture/data-en
```

## Develop

Needs `github.com/fastygo/framework` (pinned in `go.mod`). For a local Framework checkout:

```bash
cp go.work.example go.work
# uncomment the Framework `use` line, then:
go work sync
```

```bash
bun install
bun run build:css
cp .env.example .env
bun run dev   # templ generate + go run on :3011
```

## Vercel (English)

Project: import `authfly/website`, root = repo root, Framework Preset = **Go**.

`vercel.json` already sets EN fixtures and builds CSS + templ:

| Env | Value |
| --- | --- |
| `APP_DEFAULT_LOCALE` | `en` |
| `FIXTURE_DIR` | `internal/adapters/fixture/data-en` |
| `APP_STATIC_DIR` | `web/static` |

CLI:

```bash
npx vercel          # preview
npx vercel --prod   # production
```

Server listens on Vercel `PORT` via `cmd/server/main.go`. Static assets: `/assets/...`.

## Docker (RU)

```bash
docker compose up --build
# http://127.0.0.1:8080
```

## Routes

`/`, `/projects`, `/projects/{slug}`, `/blog`, `/blog/{slug}`, `/about`, `/contacts` (`/contact` alias), `/legal/{privacy|terms}`

## Regenerating fixtures from PHP archive

```bash
python3 scripts/convert-php-fixtures.py
```

`.old-php/` is the archived FlightPHP site; delete when the Go site is verified.
