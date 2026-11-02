# Agents working on authfly/website

- Content lives in `content/{en,ru}`; change both locales in the same commit (tests enforce matching pages, code
  blocks and tables). Internal links stay inside their locale (`/en/...` or `/ru/...`).
- Do not invent CLI flags, `authfly.yaml` keys, endpoints or SDK methods: check authfly/cli and authfly/authfly
  first; write UNVERIFIED (en) / НЕ ПРОВЕРЕНО (ru) in a note for what you could not check.
- No real server names, domains or IP addresses: use `example.com`, `example.test` and `203.0.113.10`.
- Views: UI8Kit 3 primitives and the Authfly layer only, classes from the `.ui8px` policy (`bunx ui8px lint` = 0),
  no inline style or script, no Cyrillic in code (UI strings in `internal/site/i18n`).
- Before a commit: `templ generate`, `bash tools/build-css.sh`, `gofmt -l .`, `go test ./...`, `bunx ui8px lint`.
