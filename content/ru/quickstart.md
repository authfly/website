---
title: Быстрый старт
description: Локальная установка Authfly и Stalwart в Docker, первое приглашение и вход — примерно за десять минут.
---
# Быстрый старт

Здесь Authfly и Stalwart запускаются на вашей машине в Docker, по обычному HTTP на порту 4000. Для публичного сервера читайте [Установка Authfly + Stalwart](/ru/guides/install-authfly-stalwart).

## Что нужно

- Linux или macOS с Docker Engine и плагином Compose (`docker compose version`).
- Root (или `sudo`) для `deploy`: он записывает секреты, принадлежащие пользователю контейнера.
- Node.js 18+ для `npx` или бинарник релиза (ниже).

## 1. CLI

Через npm:

```sh
npx authfly version
```

Или бинарник релиза с проверкой опубликованных контрольных сумм:

```sh
curl -fsSLO https://github.com/authfly/cli/releases/download/v0.4.2/authfly-linux-amd64
curl -fsSLO https://github.com/authfly/cli/releases/download/v0.4.2/SHA256SUMS
sha256sum --ignore-missing -c SHA256SUMS
install -m 0755 authfly-linux-amd64 ./authfly
```

> **Примечание:** на момент написания в npm `latest` — `authfly` 0.4.1; релиз 0.4.2 есть только на GitHub. Ниже используется `./authfly`; `npx authfly` принимает те же команды.

## 2. authfly.yaml

`init` без флагов запускает интерактивный мастер. Без вопросов:

```sh
./authfly init -y --name demo --target docker \
  --runtime react --adapter node \
  --domain example.test --public-url http://localhost:4000 --tls off \
  --signin magic,password \
  --client "my-app:My app=http://localhost:5173/callback" \
  --policy gdpr --operator "Example Ltd" --contact privacy@example.test
```

`--client` регистрирует OIDC-клиента вашего приложения (`id[:Name]=redirect[,redirect]`). Все ключи — в [authfly.yaml](/ru/reference/authfly-yaml).

## 3. Развёртывание

```sh
sudo ./authfly deploy
sudo ./authfly status
sudo ./authfly doctor
```

`deploy` скачивает подписанный канал релизов, проверяет подпись, генерирует секреты в `.authfly/` (права 0600), пишет `.authfly/compose.yml`, запускает Stalwart, закрывает его (сервисный аккаунт, админка выключена, OIDC заблокирован) и запускает Authfly. Команда идемпотентна: после правки `authfly.yaml` просто запустите её снова.

> **Известная проблема (0.4.2):** канал 0.4.2 закрепляет Stalwart v0.16.27, а сервер 0.4.2 ожидает Stalwart 1.0. Если шаг `stalwart-init` падает с `invalidPatch enabledPermissions`, см. обход в [Релизах](/ru/releases#известные-проблемы).

## 4. Приглашение и регистрация

Публичная регистрация по умолчанию выключена, нужен инвайт:

```sh
docker compose -p authfly-demo -f .authfly/compose.yml exec -T authfly app invite create -note "first user"
```

Токен печатается один раз. Откройте `http://localhost:4000/auth/sign-up`, вставьте инвайт и создайте аккаунт. Так как почтовый домен локальный, письмо с подтверждением приходит в новый ящик в Stalwart; с настоящим DNS оно дойдёт до любого ящика.

## 5. Вход

Откройте `http://localhost:4000/auth/sign-in`, затем кабинет `http://localhost:4000/account`.

## Дальше

- Подключить приложение через SSO: [React + Docker на Authfly](/ru/guides/react-docker).
- Удалить всё: `sudo ./authfly down --purge` (данные тоже удаляются).
