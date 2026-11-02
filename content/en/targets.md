---
title: Targets
description: The nine runtimes of the Authfly widgets, how to install each from the UI8Kit registry, and which server adapter pairs with it.
---
# Targets: one system, 9 platforms

Every target renders the same DOM, classes and ARIA for the six widgets. The authkit CI renders 20 reference states on each runtime and compares the HTML with React (parity), and installs every registry item into an empty project per runtime (registry smoke).

| Target | Install dir | `app.runtime` | Typical `app.adapter` | Package in authkit |
|---|---|---|---|---|
| React (reference) | `src` | `react` | `node` / `bun` | `packages/ui/react` |
| Vue 3 | `src` | `vue` | `node` / `bun` | `packages/ui/vue` |
| Svelte 5 | `src` | `svelte` | `node` / `bun` | `packages/ui/svelte` |
| Solid | `src` | `solid` | `node` / `bun` | `packages/ui/solid` |
| Web Components | `src` | not accepted | any | `packages/ui/webcomponent` |
| templ (Go) | `.` | `templ` | `go` | `packages/ui/templ` |
| Twig (PHP) | `.` | `twig` | `php` | `packages/ui/twig`, mirror authfly/twig |
| Latte (PHP) | `.` | `latte` | `php` | `packages/ui/latte`, mirror authfly/latte |
| Jinja2 (Python) | `.` | `jinja2` | `python` | `packages/ui/jinja2` |

## Install

```sh
npx ui8kit@3.2.1 init -y --framework <target> --dir <install dir> --registry-version 3.3.0
npx ui8kit@3.2.1 add authfly-sign-in authfly-sign-up authfly-account \
  --registry-url https://registry.buildy.tw/authfly/<target>
```

For webcomponent and templ, add `authfly-widgets` (all six widgets in one file) instead of single widgets.

> **Unverified:** only the React install was run by hand for this documentation (November 2026). The other targets are covered by the authkit registry smoke in CI, which installs from a local copy of the registry with primitives from unpkg; the exact `init` flags above were not run for them.

## Per-target notes

- **Twig, Latte:** with install dir `.` the UI8Kit CLI (3.4.0) installs `authfly-lib` (`src/*.php`) into the project root; move the files back to `src/` or map the namespace (`Authfly\Twig\`, `Authfly\Latte\`). Composer autoload is not written by the CLI.
- **templ:** Authfly Go files import `github.com/authfly/authkit/packages/ui/templ/...`. The CLI rewrites import paths only for UI8Kit primitives, so either use the authkit templ module directly (`go get github.com/authfly/authkit/packages/ui/templ`) or rewrite the imports. Generated `*_templ.go` files are shipped (templ v0.3.1070).
- **Jinja2:** `authfly_jinja` is installed as files; add it to your own `requirements.txt` or `pyproject.toml`.
- **Web Components:** `ui8kit init --framework webcomponent` needs `UI8KIT_EXPERIMENTAL=webcomponent`; the Authfly registry has no webcomponent build, primitives come from `@ui8kit/registry@3.3.0` on unpkg.
- Framework dependencies (`react`, `vue`, `twig/twig`, …) are not declared by the registry items; your app brings them.

## Server side

Widgets are forms; your server talks to Authfly. Use the SDK of your adapter ([SDKs](/en/reference/sdks)) or any standards-compliant OIDC client. The [React + Docker guide](/en/guides/react-docker) shows the complete pattern for the JavaScript family.
