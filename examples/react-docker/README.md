# React + Docker on Authfly (example)

A React 19 single-page app with a small Bun backend-for-frontend: sign-in through Authfly with OpenID Connect
(authorization code + PKCE, `openid-client` 6), the session in a signed HttpOnly cookie, and the Authfly widgets
from the UI8Kit registry. Tested in November 2026 against Authfly 0.4.2 on Docker.

Walkthrough: https://authfly.vercel.app/en/guides/react-docker (Russian: /ru/guides/react-docker).

```sh
# 1. primitives + widgets (CLI-owned files, git-ignored here)
npx ui8kit@3.2.1 init -y --framework react --dir src --registry-version 3.3.0
npx ui8kit@3.2.1 add authfly-notice authfly-account --registry-url https://registry.buildy.tw/authfly/react

# 2. dev: BFF on :3000, Vite on :5173 (proxies /login /callback /logout /api)
bun install
export SESSION_SECRET="$(openssl rand -hex 32)"
bun server.ts &
bunx vite

# 3. container
docker build -t my-app .
docker run --rm --network host -e SESSION_SECRET -e APP_URL=http://localhost:3000 my-app
```

| Variable | Default | Meaning |
|---|---|---|
| `OIDC_ISSUER` | `http://localhost:4000` | Authfly `public_url` |
| `OIDC_CLIENT_ID` | `my-app` | `app.clients[].id` in authfly.yaml |
| `APP_URL` | `http://localhost:5173` | where the browser opens the app; the redirect URI is `$APP_URL/callback` |
| `SESSION_SECRET` | none, required | 32+ random bytes |
| `PORT` | `3000` | BFF port |

`authfly.yaml.example` is the Authfly config used for the test. No secrets are stored in this directory.

License: Apache-2.0, © Authfly.
