---
title: CLI
description: Команды authfly и флаги init в релизе 0.4.2.
---
# CLI

`npx authfly <command> [flags]` или бинарник `authfly`. Общий флаг: `-f authfly.yaml` (по умолчанию `./authfly.yaml`).

## Команды

| Команда | Что делает |
|---|---|
| `init` | создаёт `authfly.yaml` (интерактивный мастер; `-y` с флагами — без вопросов) |
| `deploy` | развёртывает или приводит к `authfly.yaml`; идемпотентна; запускать от root |
| `status` | что запущено: релиз, сервисы, порты |
| `doctor` | аудит установки против `authfly.yaml` и пакета политик |
| `upgrade [--to <version>]` | снимок, обновление до новейшего (или указанного) релиза, автоматический откат при ошибке |
| `rollback` | восстанавливает снимок последнего обновления |
| `backup` | зашифрованный бэкап (AES-256-GCM с `backup_key`) |
| `restore <file.afbk>` | восстанавливает зашифрованный бэкап |
| `secrets` | список секретов и их прав; значения не печатаются никогда |
| `down [--purge]` | останавливает и удаляет контейнеры; `--purge` удаляет и данные |
| `serve` | запускает сервер Authfly (то, что работает в контейнерах) |
| `invite create\|list\|revoke` | инвайты на регистрацию (см. ниже) |
| `version` | печатает версию |

## Флаги init

| Флаг | Задаёт |
|---|---|
| `-y` | без вопросов |
| `-config <file>`, `-force` | выходной файл, перезапись |
| `-name`, `-target`, `-provider` | `name`, `target`, `provider` |
| `-runtime`, `-adapter` | `app.runtime`, `app.adapter` |
| `-client "id[:Name]=uri[,uri];…"` | `app.clients` |
| `-domain`, `-public-url` | `domain`, `public_url` |
| `-tls`, `-acme-email`, `-cert`, `-key`, `-http-port` | `http.*` |
| `-mail-hostname`, `-smtp-port`, `-submissions-port`, `-imaps-port`, `-encryption-at-rest` | `mail.*` |
| `-signin magic,password,username`, `-require-verify` | `sign_in.*` |
| `-policy`, `-operator`, `-contact`, `-host-country`, `-us-states` | `policy.*` |
| `-management-port`, `-console-port` | `ports.*` |
| `-channel`, `-allow-unsigned-channel` | `channel.*` |
| `-incus-network`, `-incus-subnet`, `-firewall`, `-incus-no-limits` | `incus.*` |

Парсер флагов Go принимает и `-flag`, и `--flag`.

## invite

Запускается там, где лежит хранилище Authfly, обычно внутри контейнера:

```sh
docker compose -p authfly-<name> -f .authfly/compose.yml exec -T authfly app invite create -note "who it is for" -ttl 72h
docker compose -p authfly-<name> -f .authfly/compose.yml exec -T authfly app invite list
docker compose -p authfly-<name> -f .authfly/compose.yml exec -T authfly app invite revoke <id>
```

| Флаг | Смысл |
|---|---|
| `-store <file>` | хранилище инвайтов (по умолчанию из `AUTHFLY_STORE`) |
| `-kind user\|org` | тип инвайта |
| `-note <text>` | ваша заметка |
| `-ttl <duration>` | срок жизни, по умолчанию `72h` |

Токен печатается один раз; не оставляйте его в истории shell и логах.
