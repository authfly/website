---
title: SDK
description: Серверные SDK для TypeScript, Go, Python и PHP из authfly/authfly, что каждый умеет, и известные пробелы.
---
# SDK

Все SDK лежат в authfly/authfly в `sdk/`. Они оборачивают две вещи: JSON API (свои формы) и OIDC relying party (готовые страницы входа). Подойдёт и любая OIDC-библиотека по стандарту, см. [React + Docker](/ru/guides/react-docker).

| Язык | Пакет | Статус |
|---|---|---|
| TypeScript (Node, Bun, Deno) | `@authfly/node` 1.0.0 | в npm |
| Go | `github.com/authfly/authfly/sdk/go` | Go-модуль |
| Python | `authfly` 0.1.0 | не опубликован в PyPI: `pip install ./sdk/python` или скопируйте `authfly.py` |
| PHP | `authfly/sdk-php` | read-only зеркало authfly/sdk-php; `composer require authfly/sdk-php` |

## TypeScript: @authfly/node

```ts
import { Authfly, AuthflyApi } from "@authfly/node";

const oidc = new Authfly({
  authflyUrl: "http://localhost:4000",   // URL Authfly для браузера
  stalwartUrl: "http://stalwart:8080",   // серверный URL Stalwart (token, JWKS)
  clientId: "my-app",
  redirectUri: "http://localhost:5173/callback",
  sessionSecret: process.env.SESSION_SECRET!, // 32+ символа
});
const { url, pending } = await oidc.authorizeUrl();
// ...после редиректа: const { user, tokens } = await oidc.callback(searchParams, pending);

const api = new AuthflyApi("http://localhost:4000");
await api.signIn("alice@example.com", "a long password"); // {token, user}
```

Методы `AuthflyApi`: `signUp`, `signIn`, `requestMagicLink`, `verifyMagicLink`, `verifyEmail`, `forgot`, `reset`, `me`. У `Authfly` есть также `discovery`, `verifyIdToken`, `seal` / `unseal` (cookies с HMAC-подписью).

> **Внимание:** в 1.0.0 `verifyIdToken` отклоняет `aud`-массив («id_token audience mismatch»), а именно его выдаёт режим IdP. Кроме того, relying party нужен `stalwartUrl` — серверный маршрут к HTTP-порту Stalwart; держите его во внутренней сети.

## Go

```go
c := authfly.New("http://localhost:4000", apiKey)
s, err := c.SignIn(ctx, contract.SignInInput{Identifier: "alice@example.com", Password: "a long password"})
```

`Client` реализует контракт: `Capabilities`, `SignIn`, `SignUp`, `VerifyEmail`, `SendMagicLink`, `VerifyMagicLink`, `ResetPassword`, `Session`, `SignOut`, `ExportUser`, `DeleteUser`.

> **Не проверено:** `Capabilities` вызывает `GET /api/v1/capabilities`, которого сервер 0.4.2 не обслуживает.

## Python

```python
from authfly import Authfly, AuthflyApi

api = AuthflyApi("http://localhost:4000")
session = api.sign_in("alice@example.com", "a long password")
oidc = Authfly("http://localhost:4000", "http://localhost:8080", "my-app", "http://localhost:8000/callback")
url, pending = oidc.authorize_url()
```

Только стандартная библиотека. Методы: `sign_up`, `sign_in`, `request_magic_link`, `verify_magic_link`, `verify_email`, `forgot`, `reset`, `me`, `config`.

## PHP

```php
use Authfly\{Authfly, AuthflyApi};

$api = new AuthflyApi('http://localhost:4000');
$session = $api->signIn('alice@example.com', 'a long password');
$oidc = new Authfly('http://localhost:4000', 'http://localhost:8080', 'my-app', 'http://localhost:8000/callback');
[$url, $pending] = $oidc->authorizeUrl();
```

Без зависимостей.

## Известный пробел: инвайты

При выключенной публичной регистрации (по умолчанию) `POST /api/v1/sign-up` требует `invite`, но ни один помощник `signUp` в SDK его не принимает. Вызывайте эндпоинт напрямую или используйте готовую страницу регистрации.
