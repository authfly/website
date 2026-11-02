#!/usr/bin/env bash
# Build assets/app.css: Tailwind v4 over styles/app.css plus the templ primitives of the authkit module pinned in go.mod.
# Usage: tools/build-css.sh [--check]   (--check fails when the committed assets/app.css is stale)
set -euo pipefail
cd "$(dirname "$0")/.."
mod=$(go list -m -f '{{.Dir}}' github.com/authfly/authkit/packages/ui/templ)
tmp=.cache/css; rm -rf "$tmp"; mkdir -p "$tmp"; trap 'rm -rf .cache' EXIT # inside the repo so node_modules resolves
cp styles/theme.css styles/site.css "$tmp/"
{
  echo '@import "./theme.css";'; echo '@import "./site.css";'
  echo "@source \"$PWD/internal/views\";"
  echo "@source \"$mod/ui\";"
  echo "@source \"$mod/authfly\";"
} > "$tmp/app.css"
out=assets/app.css
if [ "${1:-}" = --check ]; then out="$tmp/out.css"; fi
node_modules/.bin/tailwindcss -i "$tmp/app.css" -o "$out" --minify 2>/dev/null
if [ "${1:-}" = --check ]; then
  cmp -s "$out" assets/app.css || { echo "assets/app.css is stale: run tools/build-css.sh" >&2; exit 1; }
  echo "assets/app.css up to date"
fi
