---
title: HTTP API
description: The Authfly JSON API (/api/v1), the hosted pages and the public OIDC endpoints, as routed by the 0.4.2 server.
---
# HTTP API

All paths are on `public_url`. JSON bodies, `Content-Type: application/json`. Authenticated calls send the session token as `Authorization: Bearer <token>`. Errors are `{"error": <code>, "message": …}` with the codes from the [contract](/en/concepts#the-contract-authfly-core).

## JSON API

| Method and path | Body | Result |
|---|---|---|
| `POST /api/v1/sign-up` | `{name, username, password, invite, kind, terms}` | 201 user; `invite` is required unless public sign-up is on |
| `POST /api/v1/sign-in` | `{identifier, password, mfa_token}` | 200 `{token, user}`; 401 `invalid_credentials` or `unverified` |
| `POST /api/v1/magic-link` | `{identifier}` | 202 `{sent_to}` |
| `POST /api/v1/magic-link/verify` | `{token}` | 200 `{token, user}` |
| `POST /api/v1/verify-email` | `{token}` | 200 |
| `POST /api/v1/forgot` | `{email}` | 202 `{sent_to}` |
| `POST /api/v1/reset` | `{token, password}` | 200 |
| `GET /api/v1/me` | | 200 `{user}` |
| `GET /api/v1/me/export` | | 200, export format `authfly-export/1` |
| `GET /api/v1/me/audit` | | the user's audit trail |
| `GET /api/v1/me/consents` | | recorded consents |
| `DELETE /api/v1/me` | password again | erasure |
| `POST /api/v1/me/sign-out-everywhere` | | revokes every session |
| `GET /api/v1/config` | | public configuration (enabled sign-in methods) |

> **Unverified:** the Go SDK and older docs call `GET /api/v1/capabilities`; the 0.4.2 server does not route it. Read `GET /api/v1/config` instead.

## Hosted pages

| Path | Page |
|---|---|
| `/auth/sign-in`, `/auth/sign-up`, `/auth/sign-out`, `/auth/signed-out` | sign-in, registration, sign-out |
| `/auth/magic`, `/auth/magic/verify` | magic link request and landing |
| `/auth/forgot`, `/auth/reset` | password recovery |
| `/auth/verify-email` | email verification landing |
| `/auth/2fa` | second factor step |
| `/account`, `/account/personal`, `/account/security`, `/account/privacy`, `/account/apps` | account area |
| `/account/privacy/export` | download the export |
| `/oauth/authorize`, `/oauth/consent` | OIDC authorization and consent |
| `/legal/{doc}` | policy-pack documents (`privacy`, `dpa`, …) |

## OIDC endpoints

In proxy mode Authfly publishes only this allowlist of Stalwart's OIDC surface:

| Path | Notes |
|---|---|
| `GET /.well-known/openid-configuration`, `GET /.well-known/oauth-authorization-server` | rewritten: `authorization_endpoint` is Authfly's `/oauth/authorize`; device, registration and introspection removed |
| `GET /auth/jwks.json` | signing keys |
| `POST /auth/token` | code exchange |
| `GET`/`POST /auth/userinfo` | claims |

Every other Stalwart path (`/jmap`, `/api/*`, `/admin`, `/auth/register`, `/auth/device`, `/auth/introspect`) returns 404 or 405. Read the discovery document rather than hard-coding endpoints: in IdP mode the issuer is still `public_url`, the endpoints may differ.

> **Note:** in 0.4.2 the discovery document's `protected_resources` field contains the internal container host name. It is harmless for clients but discloses an internal name.
