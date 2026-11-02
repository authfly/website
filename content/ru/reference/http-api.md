---
title: HTTP API
description: JSON API Authfly (/api/v1), готовые страницы и публичные OIDC-эндпоинты — по маршрутам сервера 0.4.2.
---
# HTTP API

Все пути — на `public_url`. Тела в JSON, `Content-Type: application/json`. Аутентифицированные вызовы передают токен сессии как `Authorization: Bearer <token>`. Ошибки — `{"error": <code>, "message": …}` с кодами из [контракта](/ru/concepts#контракт-authfly-core).

## JSON API

| Метод и путь | Тело | Результат |
|---|---|---|
| `POST /api/v1/sign-up` | `{name, username, password, invite, kind, terms}` | 201, пользователь; `invite` обязателен, если публичная регистрация выключена |
| `POST /api/v1/sign-in` | `{identifier, password, mfa_token}` | 200 `{token, user}`; 401 `invalid_credentials` или `unverified` |
| `POST /api/v1/magic-link` | `{identifier}` | 202 `{sent_to}` |
| `POST /api/v1/magic-link/verify` | `{token}` | 200 `{token, user}` |
| `POST /api/v1/verify-email` | `{token}` | 200 |
| `POST /api/v1/forgot` | `{email}` | 202 `{sent_to}` |
| `POST /api/v1/reset` | `{token, password}` | 200 |
| `GET /api/v1/me` | | 200 `{user}` |
| `GET /api/v1/me/export` | | 200, формат выгрузки `authfly-export/1` |
| `GET /api/v1/me/audit` | | журнал действий пользователя |
| `GET /api/v1/me/consents` | | записанные согласия |
| `DELETE /api/v1/me` | пароль ещё раз | удаление |
| `POST /api/v1/me/sign-out-everywhere` | | отзывает все сессии |
| `GET /api/v1/config` | | публичная конфигурация (включённые способы входа) |

> **Не проверено:** Go SDK и старая документация вызывают `GET /api/v1/capabilities`; сервер 0.4.2 такой маршрут не обслуживает. Читайте `GET /api/v1/config`.

## Готовые страницы

| Путь | Страница |
|---|---|
| `/auth/sign-in`, `/auth/sign-up`, `/auth/sign-out`, `/auth/signed-out` | вход, регистрация, выход |
| `/auth/magic`, `/auth/magic/verify` | запрос magic-ссылки и переход по ней |
| `/auth/forgot`, `/auth/reset` | восстановление пароля |
| `/auth/verify-email` | подтверждение почты |
| `/auth/2fa` | шаг второго фактора |
| `/account`, `/account/personal`, `/account/security`, `/account/privacy`, `/account/apps` | личный кабинет |
| `/account/privacy/export` | скачать выгрузку |
| `/oauth/authorize`, `/oauth/consent` | авторизация и согласие OIDC |
| `/legal/{doc}` | документы пакета политик (`privacy`, `dpa`, …) |

## OIDC-эндпоинты

В режиме прокси Authfly публикует только этот белый список OIDC-поверхности Stalwart:

| Путь | Заметки |
|---|---|
| `GET /.well-known/openid-configuration`, `GET /.well-known/oauth-authorization-server` | переписаны: `authorization_endpoint` — это `/oauth/authorize` Authfly; device, registration и introspection убраны |
| `GET /auth/jwks.json` | ключи подписи |
| `POST /auth/token` | обмен кода |
| `GET`/`POST /auth/userinfo` | claims |

Все остальные пути Stalwart (`/jmap`, `/api/*`, `/admin`, `/auth/register`, `/auth/device`, `/auth/introspect`) отвечают 404 или 405. Читайте discovery-документ, а не зашивайте эндпоинты: в режиме IdP издатель остаётся `public_url`, а эндпоинты могут отличаться.

> **Примечание:** в 0.4.2 поле `protected_resources` discovery-документа содержит внутреннее имя хоста контейнера. Клиентам это не мешает, но раскрывает внутреннее имя.
