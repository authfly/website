Теперь у меня достаточно контекста. Вот детальный анализ.

---

## WorkOS vs `@SSO` — структурное сравнение

WorkOS — это **платформа** из нескольких независимых продуктов, каждый из которых — отдельный API:

| WorkOS | Ваш стек сейчас | Готовность |
|--------|-----------------|------------|
| **SSO API** — middleware для SAML/OIDC, возвращает нормализованный `Profile` | `@SSO` (`/sso`, `/authorize`, `/token`, `/userinfo`, `/metadata`, `/jwks`) | MVP есть, но «захардкожен» на один deployment |
| **User Management** — CRUD пользователей, пароли, MFA, сессии, организации, инвайты | `@OIDC` (Hanko) + `@SSO/internal/admin` (обёртка над Hanko Admin API) | Частично: Hanko даёт credentials, а API user management — примитивное (list + create) |
| **AuthKit** — hosted UI + SDK для встраивания (login box, session management) | `@SSO` login/logout views + `auth-flow.js` + `hanko-frontend-sdk` | Монолитно внутри IdP, нет SDK для внешнего потребления |
| **Directory Sync (SCIM)** | нет | — |
| **Audit Logs** | нет | — |
| **Organizations** — multi-tenant группировка | нет | — |
| **Admin Portal** — dashboard для управления connections, users, orgs | `@SSO /console` (в том же бинарнике) | Минимальный, встроен в IdP |
| **Go SDK** (`workos-go/v6/pkg/sso`, `/usermanagement`, ...) | нет отдельного SDK | — |

---

## Ключевое архитектурное расхождение

Сейчас `@SSO` — это **монолит**, который одновременно:

1. IdP-ядро (SAML/OIDC протоколы)
2. Login UI (templ + JS)
3. Admin Console (user CRUD, mailer)
4. Конфиг-файл с захардкоженными SP/клиентами

WorkOS разделяет это на **3 слоя**:

```
┌─────────────────────────────────────────────┐
│  Admin Dashboard (отдельный веб-проект)      │  ← управление connections, users, orgs
├─────────────────────────────────────────────┤
│  API Layer (stateless REST/gRPC)             │  ← SSO, User Management, Directory Sync, Audit
│  - Organizations                             │
│  - Connections (SAML/OIDC per org)            │
│  - Users + Sessions                          │
│  - Invitations                               │
│  - CORS origins, Redirect URIs               │
├─────────────────────────────────────────────┤
│  AuthKit (embeddable UI + client SDK)        │  ← виджеты авторизации для сайтов-потребителей
└─────────────────────────────────────────────┘
```

---

## Предлагаемая целевая архитектура (self-hosted WorkOS)

```
Репозитории:

@OIDC          — Hanko (credential backend, без изменений)
@SSO           — IdP Core: протоколы + API + auth widgets
@Console       — Admin Dashboard (новый проект)
@UI8Kit        — UI-компоненты (библиотека, без изменений)
```

### `@SSO` — IdP Core (тонкий, без админки)

Остаётся:
- SAML endpoints (`/sso`, `/sso/complete`, `/metadata`, `/logout`)
- OIDC endpoints (`/authorize`, `/token`, `/userinfo`, `/jwks`, `/.well-known/...`)
- Login/Logout UI (`/auth/login`, `/auth/logout`) — «AuthKit» виджеты
- **Новое: Management API** (`/api/v1/...`) — REST-интерфейс для программного доступа:
  - `/api/v1/users` — CRUD (прокси к Hanko Admin + собственные метаданные)
  - `/api/v1/organizations` — мульти-тенант группировка
  - `/api/v1/connections` — SAML/OIDC connections per organization (сейчас — `config.yaml`)
  - `/api/v1/invitations` — инвайт-ссылки
  - `/api/v1/sessions` — список активных сессий для GDPR-отзыва
  - `/api/v1/audit` — лог событий аутентификации

SDK-подход или фреймворк:
- **SDK** (рекомендуется, как у WorkOS): `@SSO` выставляет API, потребители вызывают его через Go/JS SDK или напрямую через REST. Виджет авторизации — это `auth-flow.js` как **встраиваемый скрипт** (аналог AuthKit), а не серверные templ-страницы.
- **Фреймворк** — это если вы хотите, чтобы потребители копировали templ-компоненты к себе. Для self-hosted IdP это **менее удобно**: лучше один hosted UI на домене IdP, а сайты-потребители просто редиректят туда.

### `@Console` — Admin Dashboard (новый проект)

- Отдельный Go-бинарник + templ + UI8Kit
- Полноценный дашборд: управление users, organizations, connections, invitations, mail, audit log
- Вызывает `@SSO` Management API (тот же `api/v1/...`), **не** Hanko Admin API напрямую
- Развёртывается на отдельном домене (например, `console.cybeross.ru`)
- Разделение означает: `@SSO` можно отдавать как продукт, а Console — отдельный deployment

---

## Дорожная карта

### Phase 0: Текущий MVP (сейчас)
- [x] SAML IdP + OIDC Provider
- [x] Hanko как credential backend
- [x] Custom login UI (templ + auth-flow.js)
- [x] Базовая админка (create user, list users, mailer)
- [x] Feature flags (registration modes)
- [x] i18n (en/ru)
- [x] Dark theme

### Phase 1: Изоляция и API Layer
- Вынести `/console` из `@SSO` в отдельный `@Console`
- Создать Management API в `@SSO` (`/api/v1/users`, `/api/v1/sessions`)
- API Key auth для Management API (аналог `WORKOS_API_KEY`)
- Перевести `@Console` на вызов Management API вместо Hanko Admin API напрямую
- Убрать `build:ui8kit-js` из package.json, заменить на Go generate
- Добавить dev-CORS в `@OIDC` для локальной разработки

### Phase 2: Organizations + Connections
- Модель `Organization` — группировка пользователей (multi-tenant)
- Модель `Connection` — SAML/OIDC connection per organization (сейчас: `config.yaml` -> БД/API)
- Dynamic SP registration (SP больше не захардкожены в `config.yaml`)
- Dynamic OIDC client registration
- Console UI для управления connections

### Phase 3: Invitations + Sessions
- Invitation API: создание инвайт-ссылок, отправка email, подтверждение
- Session Management API: список активных сессий, отзыв (GDPR)
- User Cabinet (кабинет пользователя на `@SSO`): «мои сессии», «мои данные»
- Magic Link auth (если включить `email_delivery` в Hanko)

### Phase 4: SDK + AuthKit
- `@SSO` Go SDK (аналог `workos-go`): пакеты `sso`, `usermanagement`, `organizations`
- `@SSO` JS SDK: встраиваемый auth widget для сторонних сайтов
- Документация API (OpenAPI spec)
- Примеры интеграции

### Phase 5: Advanced
- Directory Sync (SCIM) — синхронизация пользователей из корпоративных IdP
- Audit Logs API — структурированные события для compliance
- MFA/Passkey (включение в Hanko + UI)
- RBAC (роли и разрешения)
- Webhooks (уведомления о событиях)
- Rate limiting, API quotas

---

## SDK vs Framework — решение сейчас

Для вашей модели (self-hosted IdP, аналог WorkOS) — **SDK**:

- `@SSO` — **сервис** с API, а не библиотека
- Сайты-потребители **не импортируют** код `@SSO`, а вызывают его API или редиректят на его hosted UI
- Admin Dashboard — **отдельный потребитель** того же API
- JS-виджет (`auth-flow.js`) может в будущем стать `<script src="https://idp.cybeross.ru/sdk/auth.js">` — embed-скрипт для любого сайта

---

## Что нужно решить прямо сейчас (Phase 1)

1. **Создавать `@Console` как отдельный репозиторий** — да, это соответствует модели WorkOS и позволяет `@SSO` оставаться тонким IdP-ядром
2. **Management API** — это центральное изменение: все операции (users, sessions, connections) идут через API `@SSO`, а не через прямые вызовы Hanko Admin
3. **API Key** — минимум один ключ для авторизации вызовов к Management API (как `WORKOS_API_KEY`)
4. **Persistent storage для connections/organizations** — на MVP можно начать с того же PostgreSQL (через Hanko или отдельную БД), уйдя от `config.yaml` для SP/OIDC clients

Если готовы двигаться — в Agent mode можно начать с выноса `/console` и создания скелета Management API.