---
title: authfly.yaml
description: Every key of authfly.yaml (version 1) with its default and allowed values, taken from the Authfly 0.4.2 config loader.
---
# authfly.yaml

`authfly.yaml` is the single source of truth of a deployment. `authfly init` writes it, `authfly deploy` converges to it. The loader is strict: an unknown key is an error. Secrets never live here; they are generated into `.authfly/` (mode 0600).

## Top level

| Key | Default | Values |
|---|---|---|
| `version` | `1` | only `1` |
| `name` | `authfly` | 2–31 chars, lowercase letters, digits, dash; the compose project is `authfly-<name>` |
| `target` | `docker` | `docker`, `incus` |
| `provider` | `stalwart` | `stalwart` (`supabase` and `authjs` are refused: contract stubs) |
| `domain` | `example.com` | mail domain of the accounts (`user@domain`) |
| `public_url` | `https://auth.<domain>` | Authfly public URL, which is the OIDC issuer |

## app

| Key | Default | Values |
|---|---|---|
| `app.runtime` | `react` | `react`, `vue`, `svelte`, `solid`, `templ`, `twig`, `latte`, `jinja2` |
| `app.adapter` | `node` | `node`, `bun`, `php`, `python`, `go` |
| `app.clients[]` | none | `{id, name, redirect_uris[]}`: the OIDC clients registered in Stalwart |

## http

| Key | Default | Values |
|---|---|---|
| `http.tls` | `acme` | `acme` (Let's Encrypt), `manual`, `off` |
| `http.port` | `443`, or `4000` when `tls: off` | public port of Authfly |
| `http.http_port` | `80` with `acme`, else `0` | ACME HTTP-01 and redirect; `0` = off |
| `http.cert_file`, `http.key_file` | none | required with `manual` |
| `http.acme_email` | none | Let's Encrypt account email |

`public_url` may be `https://` with `tls: off` (TLS terminated by your own proxy); `http://` with TLS on is refused.

## mail

| Key | Default | Values |
|---|---|---|
| `mail.hostname` | `mail.<domain>` | MX host name |
| `mail.tls` | same as `http.tls` | `acme`, `manual`, `off` (with `off` the mail ports keep TLS with Stalwart's self-signed certificate) |
| `mail.storage` | `rocksdb` | only `rocksdb` |
| `mail.encryption_at_rest` | `false` | Stalwart encryption at rest |
| `mail.ports.smtp` / `submissions` / `imaps` | `25` / `465` / `993` | public mail ports |

## sign_in

| Key | Default | Meaning |
|---|---|---|
| `sign_in.magic_link` | `true` | magic link by email |
| `sign_in.password` | `true` | email + password |
| `sign_in.username` | `false` | username (the local part) + password |
| `sign_in.require_verify` | from `init` | email verification before the first sign-in |
| `sign_in.social`, `two_factor`, `webauthn` | `false` | must stay `false` in 0.4.2 |

If none of `magic_link`, `password`, `username` is set, magic link and password are enabled.

## policy

| Key | Default | Meaning |
|---|---|---|
| `policy.pack` | `none` | `none`, `gdpr`, `us-pii`, `ru-152fz` ([Policy packs](/en/policy-packs)) |
| `policy.operator` | none | legal name of the controller / operator |
| `policy.contact` | none | privacy contact email |
| `policy.us_states` | `[CA]` for `us-pii` | state laws in scope (only California is verified) |
| `policy.residency` | `RU` for `ru-152fz` | ISO country the data must stay in |
| `policy.host_country` | none | where the server is, declared by you, checked by `doctor` |

## ports, channel, incus, backup

| Key | Default | Meaning |
|---|---|---|
| `ports.management` | `8080` | Stalwart HTTP on `127.0.0.1` only |
| `ports.console` | `4500` | Authfly Console on `127.0.0.1` only |
| `channel.name` | `stable` | channel name |
| `channel.manifest` | `https://github.com/authfly/cli/releases/latest/download/channel.json` | URL or file |
| `channel.version` | empty (newest) | release of the first deploy; later moves are `authfly upgrade` |
| `channel.allow_unsigned` | `false` | dev channels only, never for `https://` |
| `incus.network` / `subnet` | `authfly0` / `10.118.42.1/24` | private bridge |
| `incus.pool` / `image` | `default` / `images:debian/13` | storage pool and base image |
| `incus.firewall` | `false` | write and load `/etc/nftables.d/authfly-<name>.nft` |
| `incus.no_limits` | `false` | skip memory/process limits |
| `backup.dir` / `keep` | `backups` / `7` | encrypted backups and how many to keep |

## Example

```yaml
version: 1
name: demo
target: docker
provider: stalwart
app:
  runtime: react
  adapter: node
  clients:
    - id: my-app
      name: My app
      redirect_uris:
        - http://localhost:5173/callback
domain: example.test
public_url: http://localhost:4000
http:
  port: 4000
  tls: "off"
sign_in:
  magic_link: true
  password: true
  require_verify: true
policy:
  pack: gdpr
  operator: Example Ltd
  contact: privacy@example.test
```

## Server settings outside authfly.yaml

These are environment variables of the `authfly` service (set in the generated `.authfly/compose.yml`; a later `deploy` regenerates it):

| Variable | Default | Meaning |
|---|---|---|
| `AUTHFLY_PUBLIC_SIGNUP` | `false` | `true` allows sign-up without an invite |
| `AUTHFLY_IDP` | `off` | `on`: Authfly signs ID tokens itself ([Concepts](/en/concepts#oidc-two-modes)) |
| `AUTHFLY_IDP_KEYGEN` | unset | `true` generates the IdP signing key |
| `AUTHFLY_TRUSTED_PROXIES` | `127.0.0.1,::1` | proxies whose `X-Forwarded-For` is trusted |
| `AUTHFLY_RESERVED_NAMES_FILE` | unset | extra reserved names, one per line |
| `SMTP_ADDR`, `SMTP_FROM` | set by deploy | outgoing mail relay and sender |
