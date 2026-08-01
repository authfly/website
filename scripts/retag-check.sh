#!/usr/bin/env bash
# Zoned retag check: composition layers must not own raw HTML (except registry).
# Allowlist SoT: fastygo.config.mjs → retag.allow
# Modes: check (default) | report | write-baseline
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

MODE="${1:-check}"
AWK_PROG="$ROOT/scripts/retag-check.awk"

if [[ -f "$ROOT/fastygo.config.mjs" ]] && command -v bun >/dev/null 2>&1; then
  BASELINE="${RETAG_BASELINE:-$ROOT/$(bun "$ROOT/scripts/config-get.mjs" retag.baseline)}"
  ALLOW_HIDDEN="$(bun "$ROOT/scripts/config-get.mjs" retag.allowHiddenFiles | awk 'NF{printf sep $0; sep="|"}')"
else
  BASELINE="${RETAG_BASELINE:-$ROOT/.project/retag-baseline.txt}"
  ALLOW_HIDDEN=""
  echo "retag: warn — bun/config missing; hidden-input registry empty (strict)" >&2
fi

# Collect violations: path|line|tag|snippet
collect_violations() {
  find . -name '*.templ' -type f -not -path './internal/kit/ui/*' -print0 |
  while IFS= read -r -d '' f; do
    awk -v file="$f" -v allow_hidden="$ALLOW_HIDDEN" -f "$AWK_PROG" "$f"
  done
}

counts_from_violations() {
  awk -F'|' 'NF{ c[$1]++ } END { for (f in c) printf "%s %d\n", f, c[f] }' | sort
}

print_report() {
  local tmp="$1"
  local verbose="${2:-0}"
  if [[ ! -s "$tmp" ]]; then
    echo "retag: 0 violations"
    return 0
  fi
  if [[ "$verbose" == "1" ]]; then
    echo "retag violations:"
    awk -F'|' '{ printf "  %s:%s <%s>\n", $1, $2, $3 }' "$tmp"
    echo
  fi
  echo "retag counts:"
  counts_from_violations <"$tmp" | awk '{ printf "  %s %s\n", $1, $2 }'
}

write_baseline() {
  local tmp="$1"
  mkdir -p "$(dirname "$BASELINE")"
  if [[ ! -s "$tmp" ]]; then
    : >"$BASELINE"
    echo "retag: wrote empty baseline → $BASELINE"
    return 0
  fi
  counts_from_violations <"$tmp" >"$BASELINE"
  echo "retag: wrote baseline → $BASELINE"
  sed 's/^/  /' "$BASELINE"
}

ratchet_check() {
  local tmp="$1"
  local counts_file
  counts_file="$(mktemp)"
  counts_from_violations <"$tmp" >"$counts_file" || true

  if [[ ! -f "$BASELINE" ]]; then
    echo "retag: missing baseline $BASELINE — run: $0 write-baseline" >&2
    rm -f "$counts_file"
    return 1
  fi

  local failed=0
  local path count base
  while read -r path count; do
    [[ -z "${path:-}" ]] && continue
    base="$(awk -v p="$path" '$1==p {print $2; exit}' "$BASELINE")"
    base="${base:-0}"
    if (( count > base )); then
      echo "retag: debt grew: $path $count > baseline $base" >&2
      failed=1
    fi
  done <"$counts_file"

  while read -r path base; do
    [[ -z "${path:-}" ]] && continue
    count="$(awk -v p="$path" '$1==p {print $2; exit}' "$counts_file")"
    count="${count:-0}"
    if (( count < base )); then
      echo "retag: progress $path $count < baseline $base (run write-baseline to lock in)" >&2
    fi
  done <"$BASELINE"

  rm -f "$counts_file"
  return "$failed"
}

TMP="$(mktemp)"
trap 'rm -f "$TMP"' EXIT
collect_violations >"$TMP" || true

case "$MODE" in
  report)
    print_report "$TMP" 1
    ;;
  write-baseline)
    write_baseline "$TMP"
    ;;
  check)
    print_report "$TMP" 0
    if ! ratchet_check "$TMP"; then
      echo "retag: tip — ./scripts/retag-check.sh report" >&2
      echo "retag: hidden allowlist — fastygo.config.mjs → retag.allow" >&2
      exit 1
    fi
    echo "retag: ratchet OK"
    ;;
  *)
    echo "usage: $0 [check|report|write-baseline]" >&2
    exit 2
    ;;
esac
