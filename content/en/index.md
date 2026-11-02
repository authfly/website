---
title: Authfly
nav: Overview
description: What Authfly is, why it runs on Stalwart, and how the pieces fit: one contract, nine platforms.
---
# What Authfly is

Authfly is a self-hosted identity layer for your own apps. It runs next to **Stalwart**, the open-source mail server with a built-in directory and OpenID Connect provider, and adds everything an app needs around accounts:

- registration (by invite by default), sign-in with password, username or magic link;
- email verification, password reset, sign out everywhere;
- an account area: profile, privacy export, account deletion, consents and an audit trail;
- an OpenID Connect provider for single sign-on into your apps (authorization code with PKCE);
- policy packs that set technical defaults for GDPR, US privacy law (CCPA/CPRA) or Russian 152-FZ.

Your users' accounts are mailboxes in Stalwart. Authfly never exposes Stalwart's management API: the internet sees Authfly and the mail ports, nothing else.

## One system, 9 platforms

The same six widgets (sign-in, sign-up, request, notice, consent, account) render the same HTML on nine runtimes, built on UI8Kit 3 primitives:

| Family | Runtimes |
|---|---|
| JavaScript | React, Vue 3, Svelte 5, Solid, Web Components |
| Go | templ |
| PHP | Twig, Latte |
| Python | Jinja2 |

Every widget reads only the capability flags of one contract, `@authfly/core`, so swapping the front end never changes the back end. See [Targets](/en/targets).

## How it fits together

| Piece | Repository | Role |
|---|---|---|
| `authfly` CLI and server | authfly/cli | `npx authfly init` / `deploy`: generates config, secrets and containers; serves hosted pages, the JSON API and the OIDC endpoints |
| Contract and SDKs | authfly/authfly | `@authfly/core` (one JSON schema), server SDKs for TypeScript, Go, Python and PHP |
| Widgets | authfly/authkit | the six widgets for nine runtimes, installed with `npx ui8kit@3 add` from the Authfly registry |
| Stalwart | stalwartlabs/stalwart | accounts, mail, OIDC signing; kept private behind Authfly |

## Where to go next

1. [Quickstart](/en/quickstart): a local install in about ten minutes.
2. [Concepts](/en/concepts): the contract, adapters, widgets, the registry and the signed release channel.
3. [Install Authfly + Stalwart](/en/guides/install-authfly-stalwart): the production path, with DNS, TLS and operations.
4. [React + Docker on Authfly](/en/guides/react-docker): a modern app with SSO, tested end to end.

> **For AI agents:** every page is available as Markdown (add `.md` to the URL), the index is [/llms.txt](/llms.txt) and everything in one file is [/llms-full.txt](/llms-full.txt). Facts come from the Authfly source at release 0.4.2; anything not checked against code is marked UNVERIFIED.
