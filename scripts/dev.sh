#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."

if [[ -f .env ]]; then
  set -a
  # shellcheck disable=SC1091
  source .env
  set +a
fi

export APP_BIND="${APP_BIND:-127.0.0.1:3011}"
export APP_STATIC_DIR="${APP_STATIC_DIR:-web/static}"
export FIXTURE_DIR="${FIXTURE_DIR:-internal/adapters/fixture/data}"

echo "templ generate..."
if go tool templ version >/dev/null 2>&1; then
  go tool templ generate ./...
else
  go run github.com/a-h/templ/cmd/templ@v0.3.1001 generate ./...
fi

echo "go run ./cmd/server (bind=${APP_BIND})"
go run ./cmd/server
