# Retag — composition without raw HTML

**retag** = replace raw HTML tags in composition layers with kit atoms / registry components, choosing atoms by **HTML5 role**.

Normative Cursor rule: `.cursor/rules/templ-retag.mdc`.

**Allowlist registry (SoT):** [`fastygo.config.mjs`](../fastygo.config.mjs) → `retag.allow`  
Tooling reads it via `bun scripts/config-get.mjs retag.allowHiddenFiles`.

Tag allowlists live in codegen: `internal/kit/utils/tags.go` (`TagGroup*`, `ResolveTag`).

## Zones

| Zone | Paths | Raw HTML |
| --- | --- | --- |
| **U — atoms** | `internal/kit/ui/**` | Required. Do not retag / do not hand-edit codegen. |
| **L — document** | `internal/ui/layout/**` | Allowlist only: `html`, `head`, `body`, `meta`, `title`, `link`, `script`, `noscript`, `main`. |
| **C — composition** | `internal/ui/{blocks,components,widgets}/**`, `internal/views/**` | **0 raw**, except registry / marker below. |

## Composition allowlist

1. **Hidden inputs (registry)** — `<input type="hidden">` only in files listed under `retag.allow` in `fastygo.config.mjs`.  
   Comment `// retag:allow: hidden input — …` stays as local documentation; **it is not enough** without a config entry.
2. **Other escapes** — rare non-hidden raw tags need `// retag:allow: <reason>` on the previous line (prefer extending kit instead).

## Tag → atom map

| Role | Prefer (`internal/kit/ui`) |
| --- | --- |
| Landmark / page region | `block` — **always** set `Tag` (`section`, `article`, `header`, `footer`, `nav`, `aside`, …) |
| Inner surface / card shell | `box` (`div` only) |
| Vertical / horizontal / grid | `stack` / `group` / `grid` |
| Width shell | `container` (`div` \| `main` \| `section`) |
| `h1`–`h6` | `title` (`As`) |
| Paragraph | `text` |
| Inline fragment | `inline` |
| Lists | `list` + `ListItem` |
| `a` (prose) | `link` |
| `a` (CTA) | `button` + `Href` |
| controls / `img` | matching kit atoms |
| repeated clusters | `internal/ui/components` |

### Block once, Box inside

- **One root `@block.Block` per** block / widget / large component; `Tag` is mandatory.
- `@box.Box` for non-landmark internals (cards, drawers, overlays).
- Nested `block` only for HTML5-allowed landmarks (`nav` in `header`, `header`/`section` in `article`).
- Never use `block` without `Tag` (emits `div`).

## Tooling

```bash
./scripts/retag-check.sh          # ratchet check (default)
./scripts/retag-check.sh report
./scripts/retag-check.sh write-baseline
bun scripts/config-get.mjs retag.allowHiddenFiles
```

Config also documents CSS/templ paths (`css`, `templ`) for Bun scripts — Go server does not load this file.
