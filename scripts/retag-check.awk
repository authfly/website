# Retag violation scanner for one .templ file.
# Input vars:
#   file          — path printed in findings (./internal/...)
#   allow_hidden  — pipe-separated normalized paths allowed for <input type="hidden">
# Output lines: path|line|tag|snippet

BEGIN {
  n = split(allow_hidden, ah, "|")
  for (i = 1; i <= n; i++) {
    p = ah[i]
    if (p == "") continue
    gsub(/^\.\//, "", p)
    hidden_ok[p] = 1
  }
}

function zone_of(path) {
  if (path ~ /^\.\/internal\/ui\/layout\//) return "L"
  return "C"
}

function norm_path(path,    p) {
  p = path
  sub(/^\.\//, "", p)
  return p
}

function tag_name(line,    s, i, c, out) {
  s = line
  sub(/^[[:space:]]*</, "", s)
  out = ""
  for (i = 1; i <= length(s); i++) {
    c = substr(s, i, 1)
    if (c ~ /[A-Za-z0-9]/) out = out c
    else break
  }
  return tolower(out)
}

function is_hidden_input(line, tag) {
  if (tag != "input") return 0
  if (line ~ /type[[:space:]]*=[[:space:]]*"hidden"/) return 1
  if (line ~ /type[[:space:]]*=[[:space:]]*'hidden'/) return 1
  return 0
}

function is_layout_allowed(tag) {
  return tag == "html" || tag == "head" || tag == "body" || tag == "meta" || \
    tag == "title" || tag == "link" || tag == "script" || tag == "noscript" || \
    tag == "main"
}

function is_opening_tag_line(line) {
  if (line !~ /^[[:space:]]*</) return 0
  if (line ~ /^[[:space:]]*<\//) return 0
  if (line ~ /^[[:space:]]*<!/) return 0
  if (line ~ /^[[:space:]]*<\?/) return 0
  return 1
}

function allowed(zone, tag, line, prev, path) {
  # Hidden inputs: only files listed in fastygo.config.mjs retag.allow
  if (is_hidden_input(line, tag)) {
    return (norm_path(path) in hidden_ok)
  }
  # Other rare escapes: explicit marker on previous non-empty line
  if (prev ~ /retag:allow:/) return 1
  if (zone == "L" && is_layout_allowed(tag)) return 1
  return 0
}

{
  line = $0
  if (is_opening_tag_line(line)) {
    tag = tag_name(line)
    if (tag != "") {
      z = zone_of(file)
      if (!allowed(z, tag, line, prev_nonempty, file)) {
        printf "%s|%d|%s|%s\n", file, NR, tag, line
      }
    }
  }
  if (line ~ /[^[:space:]]/) prev_nonempty = line
}
