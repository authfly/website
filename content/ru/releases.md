---
title: Релизы
description: История релизов Authfly, как проверить релиз, и известные проблемы текущего релиза.
---
# Релизы

Релизы публикуются на GitHub в authfly/cli: бинарники, `SHA256SUMS`, подписанный `channel.json` и образы контейнеров в GHCR, закреплённые по digest.

| Версия | Дата | Главное |
|---|---|---|
| 0.4.2 | 2026-11-02 | виджеты и средний слой UI8Kit генерируются для девяти рантаймов через authfly/codegen; npm собирается из бинарников релиза; зеркала Twig и Latte помечены тегом |
| 0.4.1 | 2026-10-24 | подписанный канал релизов (`channel.json.sig`, SSHSIG namespace `authfly-channel`) |
| 0.4.0 | 2026-10-24 | первый релиз с каналом, образы `ghcr.io/authfly/authfly` и `ghcr.io/authfly/console` |

Образы 0.4.2: `ghcr.io/authfly/authfly:0.4.2` и `ghcr.io/authfly/console:0.4.2` (скачивайте по digest из `channel.json`).

## Проверка релиза

```sh
curl -fsSLO https://github.com/authfly/cli/releases/download/v0.4.2/SHA256SUMS
sha256sum --ignore-missing -c SHA256SUMS
```

`authfly deploy` и `doctor` сами проверяют подпись канала; при PASS `doctor` показывает ключ подписи.

## Известные проблемы

Релиз 0.4.2, найдено при проверке гайдов в ноябре 2026:

| Проблема | Последствие | Обход |
|---|---|---|
| Канал закрепляет `stalwartlabs/stalwart:v0.16.27`, серверу нужен Stalwart 1.0 | `stalwart-init` падает: `role: notCreated invalidPatch enabledPermissions` | поставить образ `stalwartlabs/stalwart:v1.0.0` в `.authfly/compose.yml` и `docker compose … up -d`; `doctor` затем даёт FAIL по digest |
| Режим прокси (`AUTHFLY_IDP=off`) с клиентами приложений | `scope=openid email profile` получает `invalid_scope` | `AUTHFLY_IDP: "on"` и `AUTHFLY_IDP_KEYGEN: "true"` у сервиса authfly |
| В режиме IdP нет шага согласия | пользователя не спрашивают, разрешить ли приложение | нет; регистрируйте только свои приложения |
| `verifyIdToken` в `@authfly/node` 1.0.0 | отклоняет `aud`-массив | используйте `openid-client` |
| У `signUp` в SDK нет `invite` | регистрация падает при выключенной публичной регистрации | вызывайте `POST /api/v1/sign-up` с `invite` |
| `GET /api/v1/capabilities` не обслуживается | `Capabilities` в Go SDK падает | читайте `GET /api/v1/config` |
| `protected_resources` в discovery | показывает внутреннее имя хоста контейнера | клиентам ничего делать не нужно |
| npm `authfly` | `latest` — 0.4.1 | скачайте бинарник 0.4.2 с GitHub |

Обе правки compose перезаписывает следующий `authfly deploy`.
