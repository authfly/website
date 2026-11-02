---
title: React + Docker on Authfly
nav: React + Docker on Authfly
description: A typical modern stack tested end to end: Authfly and Stalwart in Docker, a React 19 app with a Bun backend-for-frontend, OIDC with PKCE, registration, verification, magic link, password recovery and the account area.
---
# React + Docker on Authfly

The stack: Authfly + Stalwart deployed with `authfly deploy` (Docker), and a React 19 + Vite app with a small Bun server (backend-for-frontend, BFF) that signs users in with OpenID Connect. The browser never sees a token; the session is a signed HttpOnly cookie. Every command here was run on a clean Linux machine in November 2026 with Authfly 0.4.2. The full app is in [examples/react-docker](https://github.com/authfly/website/tree/main/examples/react-docker).

## What you get

| Flow | Where it happens |
|---|---|
| Registration (invite), email verification | Authfly hosted pages `/auth/sign-up`, mail link `/auth/verify-email` |
| Password and magic-link sign-in | `/auth/sign-in`, mail link `/auth/magic/verify` |
| Forgot / reset password | `/auth/forgot`, mail link `/auth/reset` |
| Account: profile, export, deletion | `/account` |
| Single sign-on into your app | `/oauth/authorize` → your `/callback` |

## 1. Authfly in Docker

Register the app as an OIDC client with both redirect URIs (Vite dev server and the container):

```sh
./authfly init -y --name demo --target docker --runtime react --adapter node \
  --domain example.test --public-url http://localhost:4000 --tls off \
  --signin magic,password \
  --client "my-app:My app=http://localhost:5173/callback,http://localhost:3000/callback" \
  --policy gdpr --operator "Example Ltd" --contact privacy@example.test
sudo ./authfly deploy
sudo ./authfly doctor
```

### Two settings 0.4.2 needs

> **Known issue:** the 0.4.2 channel pins Stalwart v0.16.27, but the server needs Stalwart 1.0 (`stalwart-init` fails with `invalidPatch enabledPermissions`). Tested workaround: in `.authfly/compose.yml` set the stalwart image to `stalwartlabs/stalwart:v1.0.0`, then `docker compose -p authfly-demo -f .authfly/compose.yml up -d`. `authfly doctor` then reports a FAIL on the Stalwart image digest, which is expected for this override.

> **Known issue:** in the default proxy mode (`AUTHFLY_IDP=off`) Stalwart answers `invalid_scope` to `scope=openid email profile` from app clients. Tested fix: add to the `authfly` service environment in `.authfly/compose.yml`

```yaml
      AUTHFLY_IDP: "on"
      AUTHFLY_IDP_KEYGEN: "true"
```

and recreate it: `docker compose -p authfly-demo -f .authfly/compose.yml up -d authfly`. Both edits are overwritten by the next `authfly deploy`; re-apply them until a release fixes this.

Create an invite for the first user (printed once):

```sh
docker compose -p authfly-demo -f .authfly/compose.yml exec -T authfly app invite create -note "first user"
```

## 2. The React app

```sh
mkdir my-app && cd my-app
# copy package.json, server.ts, vite.config.ts, index.html, tsconfig.json, src/ from examples/react-docker
npx ui8kit@3.2.1 init -y --framework react --dir src --registry-version 3.3.0
npx ui8kit@3.2.1 add authfly-notice authfly-account --registry-url https://registry.buildy.tw/authfly/react
bun install
```

Dependencies: `react` 19, `openid-client` 6, `@authfly/core` 1.0.0, `clsx`, `tailwind-merge`; dev: `vite` 7, `@vitejs/plugin-react` 5, `tailwindcss` 4 with `@tailwindcss/vite`, `typescript`.

> **Warning:** pin `ui8kit@3.2.1`. `ui8kit@3` (3.4.0) mixed registry utils 0.1.0 with primitives 3.3.0 and the build failed. After `init`, `src/styles/index.css` contained a stray `@source "../../../kit/react/src"` line; delete it.

### The BFF (server.ts)

The server does four things:

1. `GET /login`: creates PKCE verifier, `state` and `nonce`, keeps them in a signed `app_pending` cookie (15 min) and redirects to Authfly's `/oauth/authorize` with `scope=openid email profile`.
2. `GET /callback`: exchanges the code (`authorizationCodeGrant` checks state, nonce and the ID token), then sets `app_session` (12 h) with `sub`, `email`, `name`, `username`.
3. `POST /logout`: clears the session cookie.
4. `GET /api/me`: returns the user or 401. Everything else serves the built SPA.

The client is public (no client secret) and uses PKCE. Discovery runs against `OIDC_ISSUER`; plain `http://` is allowed only when the issuer itself is `http://` (local development).

```ts
const config = await oidc.discovery(new URL(ISSUER), CLIENT_ID, undefined, oidc.None(),
  ISSUER.startsWith("http://") ? { execute: [oidc.allowInsecureRequests] } : undefined);
```

> **Warning:** `@authfly/node` 1.0.0 `verifyIdToken` compares `aud` as a string and rejects ID tokens whose `aud` is an array ("id_token audience mismatch"), which is what IdP mode issues. Use a standard OIDC client such as `openid-client` until that is fixed.

### The UI (src/App.tsx)

Signed out, the app renders `AuthflyNotice` with a "Sign in" link to `/login`; signed in, `AuthflyAccount` with the user from `/api/me` and `signoutUrl="/logout"`. Texts are in one `t` object, ready for your i18n.

### Run in development

```sh
export SESSION_SECRET="$(openssl rand -hex 32)"
bun server.ts &          # BFF on :3000
bunx vite                # app on :5173, proxies /login /callback /logout /api to :3000
```

Open `http://localhost:5173`, choose Sign in, register with the invite, confirm the email, and you land back in the app signed in.

## 3. The app in Docker

The Dockerfile builds the SPA in `oven/bun:1` and runs the BFF in `oven/bun:1-slim` as the unprivileged `bun` user. No secrets are baked into the image.

```sh
docker build -t my-app .
docker run --rm --network host \
  -e SESSION_SECRET -e APP_URL=http://localhost:3000 -e OIDC_ISSUER=http://localhost:4000 my-app
```

`--network host` lets the container reach Authfly on `localhost:4000` in this local setup. In production give the app its own hostname, set `APP_URL=https://app.example.com` (cookies become `Secure`), `OIDC_ISSUER=https://auth.example.com`, and register `https://app.example.com/callback` in `app.clients`.

## 4. Mail without external DNS

With the domain on Stalwart, verification, magic-link and reset mails to local addresses are delivered into the users' Stalwart mailboxes; nothing leaves the machine. For mail to other domains, set up DNS as in [Install Authfly + Stalwart](/en/guides/install-authfly-stalwart#h-3-dns).

## 5. Custom forms with the JSON API

If you build your own forms instead of using the hosted pages, call the Authfly JSON API from your server ([HTTP API](/en/reference/http-api)). Tested sequence:

| Step | Request | Result |
|---|---|---|
| sign up | `POST /api/v1/sign-up {name, username, password, invite}` | 201 |
| sign in before verification | `POST /api/v1/sign-in` | 401 `unverified` |
| verify | `POST /api/v1/verify-email {token}` | 200 |
| sign in | `POST /api/v1/sign-in {identifier, password}` | 200, session token |
| who am I | `GET /api/v1/me` with `Authorization: Bearer` | 200 |
| magic link | `POST /api/v1/magic-link`, then `/magic-link/verify {token}` | 202, then 200 |
| forgot / reset | `POST /api/v1/forgot {email}`, then `/reset {token, password}` | 202, then 200 |

> **Note:** the SDK `signUp` helpers have no `invite` parameter in 1.0.0 / 0.1.0; with public sign-up off, send the `invite` field yourself.

## 6. Clean up

```sh
sudo ./authfly down --purge      # containers and data
docker image rm my-app
```
