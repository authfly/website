---
title: Security model
nav: Security
description: What is reachable from where, the OIDC allowlist proxy, Stalwart hardening by stalwart-init, the private Console, and a production checklist.
---
# Security model

Authfly runs Stalwart as a private back end. The internet sees Authfly and the mail ports, nothing else.

## What is reachable from where

| Surface | Where | Why |
|---|---|---|
| Authfly `:4000` (or 443) | public | hosted pages, JSON API, the OIDC endpoints it proxies |
| SMTP 25, submissions 465, IMAPS 993 | public, if you use mail | mail; no plain IMAP listener |
| Stalwart HTTP `:8080` (management, JMAP, `/api/*`) | `127.0.0.1` and the private network | management must never be public |
| Authfly Console `:4500` | `127.0.0.1` | private configurator, reached through an SSH tunnel |
| Authfly admin API `:4001` | private network only, not published | used by the Console |

## OIDC through an allowlist

Stalwart serves OIDC, JMAP management, `/api/*`, device flow, dynamic client registration and introspection on one listener. Publishing it, even behind a path filter in a generic proxy, exposes all of that. Authfly publishes discovery (rewritten), JWKS, token and userinfo only; every other path answers 404 or 405. See [HTTP API](/en/reference/http-api#oidc-endpoints).

## Hardening by stalwart-init

Before Authfly starts, a one-shot `stalwart-init` step (idempotent) runs with the bootstrap admin:

1. ensures the mail domain exists;
2. creates the least-privilege role `authfly-service` (User role plus account management, domain read, impersonation and blob fetch for the GDPR export, store maintenance for purge after erasure);
3. creates the service account `authfly-svc@<domain>` with that role; Authfly uses it, never the admin;
4. disables Stalwart's web admin and account apps;
5. disables client metadata documents and registers the clients (`app.clients` plus Authfly itself).

## Sessions and abuse controls

- Public sign-up is off by default; registration needs an invite.
- Built-in reserved names (`admin`, `postmaster`, `abuse`, `root`, …) cannot be registered; extend them with `AUTHFLY_RESERVED_NAMES_FILE`.
- The client address is taken from `X-Forwarded-For` only when the request comes from `AUTHFLY_TRUSTED_PROXIES`.
- Policy packs keep IP addresses out of the Authfly audit log; `doctor` checks it.

## Console

The Console is an admin UI for GDPR actions and Authfly state: signed session cookie (8 h), lock-out after five failed logins, CSRF tokens, CSP without inline scripts, a hash-chained audit log, typed confirmation before destructive actions. Open it only through an SSH tunnel:

```sh
ssh -N -L 4500:127.0.0.1:4500 you@203.0.113.10
```

## Release channel

Releases are trusted only through the signed channel: SSHSIG over `channel.json` by a FIDO2 hardware key, trusted keys built into the CLI, binaries pinned by sha256 and images by digest. See [Concepts](/en/concepts#the-signed-release-channel).

## Production checklist

- `public_url` is `https://`; cookies are then `Secure`.
- Ports 8080, 4500 and 4001 are not reachable from outside; `/jmap`, `/api/*` and `/admin/` return 404/405 on the public host.
- `authfly secrets` shows every secret as 0600; no secret is in git, argv or shell history.
- `authfly doctor` passes (an expected FAIL only for a deliberate image override).
- A backup was restored on a spare host at least once.
- Mail DNS (SPF, DKIM, DMARC, PTR) is in place before you rely on email to other domains.

## Reporting a vulnerability

Report privately through GitHub security advisories on the affected repository of the authfly organisation, not in a public issue.

> **Unverified:** private vulnerability reporting must be enabled per repository; check the repository's Security tab.
