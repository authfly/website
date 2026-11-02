---
title: React + Docker на Authfly
nav: React + Docker на Authfly
description: Типичный современный стек, проверенный от начала до конца: Authfly и Stalwart в Docker, приложение на React 19 с BFF на Bun, OIDC с PKCE, регистрация, подтверждение почты, magic-ссылка, восстановление пароля и личный кабинет.
---
# React + Docker на Authfly

Стек: Authfly + Stalwart, развёрнутые через `authfly deploy` (Docker), и приложение React 19 + Vite с небольшим сервером на Bun (backend-for-frontend, BFF), который выполняет вход по OpenID Connect. Браузер никогда не видит токенов; сессия — подписанная HttpOnly cookie. Каждая команда здесь выполнена на чистой Linux-машине в ноябре 2026 с Authfly 0.4.2. Полное приложение — в [examples/react-docker](https://github.com/authfly/website/tree/main/examples/react-docker).

## Что получится

| Сценарий | Где происходит |
|---|---|
| Регистрация (по инвайту), подтверждение почты | страницы Authfly `/auth/sign-up`, ссылка из письма `/auth/verify-email` |
| Вход по паролю и magic-ссылке | `/auth/sign-in`, ссылка из письма `/auth/magic/verify` |
| Забыли пароль / сброс | `/auth/forgot`, ссылка из письма `/auth/reset` |
| Кабинет: профиль, выгрузка, удаление | `/account` |
| Единый вход в ваше приложение | `/oauth/authorize` → ваш `/callback` |

## 1. Authfly в Docker

Зарегистрируйте приложение как OIDC-клиента с обоими redirect URI (dev-сервер Vite и контейнер):

```sh
./authfly init -y --name demo --target docker --runtime react --adapter node \
  --domain example.test --public-url http://localhost:4000 --tls off \
  --signin magic,password \
  --client "my-app:My app=http://localhost:5173/callback,http://localhost:3000/callback" \
  --policy gdpr --operator "Example Ltd" --contact privacy@example.test
sudo ./authfly deploy
sudo ./authfly doctor
```

### Две настройки, которые нужны 0.4.2

> **Известная проблема:** канал 0.4.2 закрепляет Stalwart v0.16.27, а серверу нужен Stalwart 1.0 (`stalwart-init` падает с `invalidPatch enabledPermissions`). Проверенный обход: в `.authfly/compose.yml` поставьте образ stalwart `stalwartlabs/stalwart:v1.0.0`, затем `docker compose -p authfly-demo -f .authfly/compose.yml up -d`. После этого `authfly doctor` показывает FAIL по digest образа Stalwart — для такой подмены это ожидаемо.

> **Известная проблема:** в режиме прокси по умолчанию (`AUTHFLY_IDP=off`) Stalwart отвечает `invalid_scope` на `scope=openid email profile` от клиентов приложений. Проверенное исправление: добавьте в окружение сервиса `authfly` в `.authfly/compose.yml`

```yaml
      AUTHFLY_IDP: "on"
      AUTHFLY_IDP_KEYGEN: "true"
```

и пересоздайте его: `docker compose -p authfly-demo -f .authfly/compose.yml up -d authfly`. Обе правки перезаписывает следующий `authfly deploy`; повторяйте их, пока релиз это не исправит.

Создайте инвайт для первого пользователя (печатается один раз):

```sh
docker compose -p authfly-demo -f .authfly/compose.yml exec -T authfly app invite create -note "first user"
```

## 2. Приложение на React

```sh
mkdir my-app && cd my-app
# скопируйте package.json, server.ts, vite.config.ts, index.html, tsconfig.json, src/ из examples/react-docker
npx ui8kit@3.2.1 init -y --framework react --dir src --registry-version 3.3.0
npx ui8kit@3.2.1 add authfly-notice authfly-account --registry-url https://registry.buildy.tw/authfly/react
bun install
```

Зависимости: `react` 19, `openid-client` 6, `@authfly/core` 1.0.0, `clsx`, `tailwind-merge`; dev: `vite` 7, `@vitejs/plugin-react` 5, `tailwindcss` 4 с `@tailwindcss/vite`, `typescript`.

> **Внимание:** закрепляйте `ui8kit@3.2.1`. `ui8kit@3` (3.4.0) смешал utils 0.1.0 из реестра с примитивами 3.3.0, и сборка падала. После `init` в `src/styles/index.css` оказалась лишняя строка `@source "../../../kit/react/src"`; удалите её.

### BFF (server.ts)

Сервер делает четыре вещи:

1. `GET /login`: создаёт PKCE verifier, `state` и `nonce`, хранит их в подписанной cookie `app_pending` (15 мин) и перенаправляет на `/oauth/authorize` Authfly со `scope=openid email profile`.
2. `GET /callback`: обменивает код (`authorizationCodeGrant` проверяет state, nonce и ID-токен) и ставит `app_session` (12 ч) с `sub`, `email`, `name`, `username`.
3. `POST /logout`: очищает cookie сессии.
4. `GET /api/me`: возвращает пользователя или 401. Всё остальное — собранное SPA.

Клиент публичный (без client secret) и использует PKCE. Discovery идёт к `OIDC_ISSUER`; обычный `http://` разрешён, только если сам издатель на `http://` (локальная разработка).

```ts
const config = await oidc.discovery(new URL(ISSUER), CLIENT_ID, undefined, oidc.None(),
  ISSUER.startsWith("http://") ? { execute: [oidc.allowInsecureRequests] } : undefined);
```

> **Внимание:** в `@authfly/node` 1.0.0 `verifyIdToken` сравнивает `aud` как строку и отклоняет ID-токены, где `aud` — массив («id_token audience mismatch»), а именно такие выдаёт режим IdP. Пока это не исправлено, используйте стандартный OIDC-клиент, например `openid-client`.

### Интерфейс (src/App.tsx)

Без входа приложение показывает `AuthflyNotice` со ссылкой «Sign in» на `/login`; после входа — `AuthflyAccount` с пользователем из `/api/me` и `signoutUrl="/logout"`. Тексты собраны в одном объекте `t`, готовом для вашей i18n.

### Запуск в разработке

```sh
export SESSION_SECRET="$(openssl rand -hex 32)"
bun server.ts &          # BFF на :3000
bunx vite                # приложение на :5173, проксирует /login /callback /logout /api на :3000
```

Откройте `http://localhost:5173`, нажмите Sign in, зарегистрируйтесь по инвайту, подтвердите почту — и вы вернётесь в приложение уже вошедшим.

## 3. Приложение в Docker

Dockerfile собирает SPA в `oven/bun:1` и запускает BFF в `oven/bun:1-slim` от непривилегированного пользователя `bun`. Секреты в образ не попадают.

```sh
docker build -t my-app .
docker run --rm --network host \
  -e SESSION_SECRET -e APP_URL=http://localhost:3000 -e OIDC_ISSUER=http://localhost:4000 my-app
```

`--network host` даёт контейнеру доступ к Authfly на `localhost:4000` в этой локальной схеме. В продакшене дайте приложению своё имя хоста, задайте `APP_URL=https://app.example.com` (cookies станут `Secure`), `OIDC_ISSUER=https://auth.example.com` и зарегистрируйте `https://app.example.com/callback` в `app.clients`.

## 4. Почта без внешнего DNS

Когда домен на Stalwart, письма подтверждения, magic-ссылки и сброса на локальные адреса доставляются в ящики пользователей в Stalwart; ничего не покидает машину. Для почты на другие домены настройте DNS, как в [Установка Authfly + Stalwart](/ru/guides/install-authfly-stalwart#h-3-dns).

## 5. Свои формы через JSON API

Если вместо готовых страниц вы делаете свои формы, вызывайте JSON API Authfly со своего сервера ([HTTP API](/ru/reference/http-api)). Проверенная последовательность:

| Шаг | Запрос | Результат |
|---|---|---|
| регистрация | `POST /api/v1/sign-up {name, username, password, invite}` | 201 |
| вход до подтверждения | `POST /api/v1/sign-in` | 401 `unverified` |
| подтверждение | `POST /api/v1/verify-email {token}` | 200 |
| вход | `POST /api/v1/sign-in {identifier, password}` | 200, токен сессии |
| кто я | `GET /api/v1/me` с `Authorization: Bearer` | 200 |
| magic-ссылка | `POST /api/v1/magic-link`, затем `/magic-link/verify {token}` | 202, затем 200 |
| забыли / сброс | `POST /api/v1/forgot {email}`, затем `/reset {token, password}` | 202, затем 200 |

> **Примечание:** у помощников `signUp` в SDK 1.0.0 / 0.1.0 нет параметра `invite`; при выключенной публичной регистрации передавайте поле `invite` сами.

## 6. Уборка

```sh
sudo ./authfly down --purge      # контейнеры и данные
docker image rm my-app
```
