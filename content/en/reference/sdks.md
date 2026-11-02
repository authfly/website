---
title: SDKs
description: Server SDKs for TypeScript, Go, Python and PHP from authfly/authfly, what each exposes, and known gaps.
---
# SDKs

All SDKs live in authfly/authfly under `sdk/`. They wrap two things: the JSON API (custom forms) and an OIDC relying party (hosted sign-in). Any standards-compliant OIDC library works too, see [React + Docker](/en/guides/react-docker).

| Language | Package | Status |
|---|---|---|
| TypeScript (Node, Bun, Deno) | `@authfly/node` 1.0.0 | on npm |
| Go | `github.com/authfly/authfly/sdk/go` | Go module |
| Python | `authfly` 0.1.0 | not published to PyPI: `pip install ./sdk/python` or copy `authfly.py` |
| PHP | `authfly/sdk-php` | read-only mirror authfly/sdk-php; `composer require authfly/sdk-php` |

## TypeScript: @authfly/node

```ts
import { Authfly, AuthflyApi } from "@authfly/node";

const oidc = new Authfly({
  authflyUrl: "http://localhost:4000",   // browser-facing Authfly URL
  stalwartUrl: "http://stalwart:8080",   // server-side URL of Stalwart (token, JWKS)
  clientId: "my-app",
  redirectUri: "http://localhost:5173/callback",
  sessionSecret: process.env.SESSION_SECRET!, // 32+ characters
});
const { url, pending } = await oidc.authorizeUrl();
// ...after the redirect: const { user, tokens } = await oidc.callback(searchParams, pending);

const api = new AuthflyApi("http://localhost:4000");
await api.signIn("alice@example.com", "a long password"); // {token, user}
```

`AuthflyApi` methods: `signUp`, `signIn`, `requestMagicLink`, `verifyMagicLink`, `verifyEmail`, `forgot`, `reset`, `me`. `Authfly` also has `discovery`, `verifyIdToken`, `seal` / `unseal` (HMAC-signed cookies).

> **Warning:** in 1.0.0 `verifyIdToken` rejects an array `aud` ("id_token audience mismatch"), which IdP mode issues. The relying party also needs `stalwartUrl`, a server-side route to Stalwart's HTTP port; keep that on a private network.

## Go

```go
c := authfly.New("http://localhost:4000", apiKey)
s, err := c.SignIn(ctx, contract.SignInInput{Identifier: "alice@example.com", Password: "a long password"})
```

`Client` implements the contract: `Capabilities`, `SignIn`, `SignUp`, `VerifyEmail`, `SendMagicLink`, `VerifyMagicLink`, `ResetPassword`, `Session`, `SignOut`, `ExportUser`, `DeleteUser`.

> **Unverified:** `Capabilities` calls `GET /api/v1/capabilities`, which the 0.4.2 server does not route.

## Python

```python
from authfly import Authfly, AuthflyApi

api = AuthflyApi("http://localhost:4000")
session = api.sign_in("alice@example.com", "a long password")
oidc = Authfly("http://localhost:4000", "http://localhost:8080", "my-app", "http://localhost:8000/callback")
url, pending = oidc.authorize_url()
```

Standard library only. Methods: `sign_up`, `sign_in`, `request_magic_link`, `verify_magic_link`, `verify_email`, `forgot`, `reset`, `me`, `config`.

## PHP

```php
use Authfly\{Authfly, AuthflyApi};

$api = new AuthflyApi('http://localhost:4000');
$session = $api->signIn('alice@example.com', 'a long password');
$oidc = new Authfly('http://localhost:4000', 'http://localhost:8080', 'my-app', 'http://localhost:8000/callback');
[$url, $pending] = $oidc->authorizeUrl();
```

No dependencies.

## Known gap: invites

With public sign-up off (the default) `POST /api/v1/sign-up` needs `invite`, but none of the SDK `signUp` helpers takes it. Call the endpoint directly or use the hosted sign-up page.
