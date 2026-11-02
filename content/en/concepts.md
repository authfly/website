---
title: Concepts
description: The @authfly/core contract, adapters and providers, the six widgets, the UI8Kit registry, OIDC modes and the signed release channel.
---
# Concepts

## The contract: @authfly/core

`@authfly/core` is one JSON schema (`schema/contract.json`) with generated types for TypeScript, Go, Python and PHP. Operations:

| Operation | Meaning |
|---|---|
| `capabilities` | which features the provider supports |
| `signIn`, `signUp` | password sign-in and registration |
| `verifyEmail` | confirm the address from the mailed token |
| `sendMagicLink`, `verifyMagicLink` | passwordless sign-in |
| `resetPassword` | request `{email}`, confirm `{email, token, password}` |
| `session`, `signOut` | read the session, sign out every session |
| `exportUser`, `deleteUser` | data portability and erasure (deletion needs the password again) |

Errors are `AuthError{code, message}` with `code` one of `invalid_credentials`, `unverified`, `account_exists`, `invalid_token`, `invalid_input`, `mfa_required`, `unauthenticated`, `not_implemented`, `error`.

Capability flags: `magicLink`, `password`, `social`, `twoFactor`, `webauthn`. Widgets read only these flags (`widgetProps(caps)`), never anything provider-specific.

## Providers

| Provider | Status |
|---|---|
| `@authfly/provider-stalwart` | implemented: the contract over the Authfly JSON API (`/api/v1`), which owns the Stalwart integration |
| `@authfly/provider-supabase` | stub: every operation answers `not_implemented` |
| `@authfly/provider-authjs` | stub: same |

`authfly.yaml` accepts only `provider: stalwart`.

## Adapters

`app.adapter` names the server runtime of your app: `node` (also Bun and Deno), `bun`, `php`, `python` or `go`. Each family has a server SDK, see [SDKs](/en/reference/sdks). Front ends never hold secrets: the session lives in a signed cookie set by your server.

## Widgets

Six widgets with identical props on every runtime:

| Widget | Use |
|---|---|
| `AuthflySignIn` | password / username / magic link sign-in |
| `AuthflySignUp` | registration (invite field when public sign-up is off) |
| `AuthflyRequest` | forgot password, magic link request |
| `AuthflyNotice` | "check your inbox", success and error states |
| `AuthflyConsent` | OAuth consent screen |
| `AuthflyAccount` | profile, sign out, privacy export, deletion |

They are plain HTML forms built from UI8Kit 3 primitives with shadcn tokens, Lucide icons and an 8+4 spacing grid. Point the form `action` at any backend.

## The registry

Widgets are installed into your project, not imported from a black box. `npx ui8kit@3 init` adds the UI8Kit primitives, then `npx ui8kit@3 add` copies the Authfly items from the brand registry for your runtime:

```sh
npx ui8kit@3.2.1 init -y --framework react --dir src --registry-version 3.3.0
npx ui8kit@3.2.1 add authfly-sign-in authfly-account --registry-url https://registry.buildy.tw/authfly/react
```

Items: `authfly-lib` (skin, icons, theme), `authfly-components` (Card, Field, Alert, Badge, LucideIcon, layout), one block per widget, and `authfly-widgets` for webcomponent and templ (all six in one file).

> **Warning:** pin the CLI to `ui8kit@3.2.1`. With `ui8kit@3` resolving to 3.4.0, `init` mixed utils 0.1.0 from the registry with primitives 3.3.0 and the build failed (`safeRest` not exported). Tested in November 2026.

## OIDC: two modes

Authfly is the issuer your apps see. The server variable `AUTHFLY_IDP` selects how tokens are made:

| Mode | How it works | Notes |
|---|---|---|
| `off` (default) | Authfly proxies an allowlist of Stalwart's OIDC endpoints (discovery, JWKS, token, userinfo) and hosts the sign-in page | in 0.4.2 Stalwart rejects `scope=openid email profile` for app clients (`invalid_scope`), see [Releases](/en/releases#known-issues) |
| `on` | Authfly signs ID tokens itself (`AUTHFLY_IDP_KEYGEN=true` generates the key) | tested end to end with a React app; there is no consent step in this mode |

`AUTHFLY_IDP` is a server environment variable, not an `authfly.yaml` key: set it in `.authfly/compose.yml` (a later `deploy` regenerates that file).

## The signed release channel

`deploy`, `upgrade`, `rollback` and `doctor` read `channel.json` (by default from the latest GitHub release of authfly/cli) and trust nothing in it until `channel.json.sig` verifies:

- OpenSSH SSHSIG, namespace `authfly-channel`, sha512, over the exact bytes of the manifest;
- the release key is a FIDO2 security key (`sk-ssh-ed25519`); the CLI requires the user-presence flag;
- the trusted public keys are built into the CLI;
- binaries are pinned by sha256 and images by the digest of their multi-arch index.

`channel.allow_unsigned: true` exists for local dev channels only; it is refused for `https://` channels.
