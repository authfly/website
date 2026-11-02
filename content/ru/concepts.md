---
title: Концепции
description: Контракт @authfly/core, адаптеры и провайдеры, шесть виджетов, реестр UI8Kit, режимы OIDC и подписанный канал релизов.
---
# Концепции

## Контракт: @authfly/core

`@authfly/core` — одна JSON-схема (`schema/contract.json`) со сгенерированными типами для TypeScript, Go, Python и PHP. Операции:

| Операция | Смысл |
|---|---|
| `capabilities` | какие возможности поддерживает провайдер |
| `signIn`, `signUp` | вход по паролю и регистрация |
| `verifyEmail` | подтверждение адреса по токену из письма |
| `sendMagicLink`, `verifyMagicLink` | вход без пароля |
| `resetPassword` | запрос `{email}`, подтверждение `{email, token, password}` |
| `session`, `signOut` | чтение сессии, выход из всех сессий |
| `exportUser`, `deleteUser` | переносимость данных и удаление (удаление требует повторного ввода пароля) |

Ошибки — `AuthError{code, message}`, где `code` — одно из `invalid_credentials`, `unverified`, `account_exists`, `invalid_token`, `invalid_input`, `mfa_required`, `unauthenticated`, `not_implemented`, `error`.

Флаги возможностей: `magicLink`, `password`, `social`, `twoFactor`, `webauthn`. Виджеты читают только их (`widgetProps(caps)`) и ничего специфичного для провайдера.

## Провайдеры

| Провайдер | Статус |
|---|---|
| `@authfly/provider-stalwart` | реализован: контракт поверх JSON API Authfly (`/api/v1`), который и работает со Stalwart |
| `@authfly/provider-supabase` | заглушка: каждая операция отвечает `not_implemented` |
| `@authfly/provider-authjs` | заглушка: то же |

`authfly.yaml` принимает только `provider: stalwart`.

## Адаптеры

`app.adapter` — серверный рантайм вашего приложения: `node` (также Bun и Deno), `bun`, `php`, `python` или `go`. У каждого семейства есть серверный SDK, см. [SDK](/ru/reference/sdks). Фронтенд никогда не хранит секреты: сессия живёт в подписанной cookie, которую ставит ваш сервер.

## Виджеты

Шесть виджетов с одинаковыми props на всех рантаймах:

| Виджет | Назначение |
|---|---|
| `AuthflySignIn` | вход по паролю / имени пользователя / magic-ссылке |
| `AuthflySignUp` | регистрация (поле инвайта, если публичная регистрация выключена) |
| `AuthflyRequest` | забытый пароль, запрос magic-ссылки |
| `AuthflyNotice` | «проверьте почту», успех и ошибки |
| `AuthflyConsent` | экран согласия OAuth |
| `AuthflyAccount` | профиль, выход, выгрузка данных, удаление |

Это обычные HTML-формы на примитивах UI8Kit 3 с токенами shadcn, иконками Lucide и сеткой отступов 8+4. Направьте `action` формы на любой бэкенд.

## Реестр

Виджеты копируются в ваш проект, а не импортируются из чёрного ящика. `npx ui8kit@3 init` ставит примитивы UI8Kit, затем `npx ui8kit@3 add` копирует элементы Authfly из брендового реестра для вашего рантайма:

```sh
npx ui8kit@3.2.1 init -y --framework react --dir src --registry-version 3.3.0
npx ui8kit@3.2.1 add authfly-sign-in authfly-account --registry-url https://registry.buildy.tw/authfly/react
```

Элементы: `authfly-lib` (скин, иконки, тема), `authfly-components` (Card, Field, Alert, Badge, LucideIcon, раскладка), по блоку на виджет и `authfly-widgets` для webcomponent и templ (все шесть в одном файле).

> **Внимание:** закрепляйте CLI на `ui8kit@3.2.1`. Когда `ui8kit@3` указывал на 3.4.0, `init` смешал utils 0.1.0 из реестра с примитивами 3.3.0, и сборка падала (`safeRest` не экспортирован). Проверено в ноябре 2026.

## OIDC: два режима

Для ваших приложений издатель (issuer) — Authfly. Серверная переменная `AUTHFLY_IDP` выбирает, кто выпускает токены:

| Режим | Как работает | Заметки |
|---|---|---|
| `off` (по умолчанию) | Authfly проксирует белый список OIDC-эндпоинтов Stalwart (discovery, JWKS, token, userinfo) и сам показывает страницу входа | в 0.4.2 Stalwart отклоняет `scope=openid email profile` у клиентов приложений (`invalid_scope`), см. [Релизы](/ru/releases#известные-проблемы) |
| `on` | Authfly сам подписывает ID-токены (`AUTHFLY_IDP_KEYGEN=true` генерирует ключ) | проверено от начала до конца с React-приложением; шага согласия в этом режиме нет |

`AUTHFLY_IDP` — переменная окружения сервера, а не ключ `authfly.yaml`: задаётся в `.authfly/compose.yml` (следующий `deploy` перезапишет этот файл).

## Подписанный канал релизов

`deploy`, `upgrade`, `rollback` и `doctor` читают `channel.json` (по умолчанию из последнего релиза authfly/cli на GitHub) и ничему в нём не доверяют, пока не сойдётся `channel.json.sig`:

- OpenSSH SSHSIG, namespace `authfly-channel`, sha512, по точным байтам манифеста;
- ключ релизов — аппаратный FIDO2-ключ (`sk-ssh-ed25519`); CLI требует флаг присутствия пользователя (касание);
- доверенные публичные ключи встроены в CLI;
- бинарники закреплены по sha256, образы — по digest их multi-arch индекса.

`channel.allow_unsigned: true` существует только для локальных dev-каналов; для `https://` он отклоняется.
