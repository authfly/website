#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."

./scripts/retag-check.sh check

if go tool templ version >/dev/null 2>&1; then
  go tool templ generate ./...
else
  go run github.com/a-h/templ/cmd/templ@v0.3.1001 generate ./...
fi
go test ./...
