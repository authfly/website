---
title: authfly.yaml
description: Все ключи authfly.yaml (версия 1) со значениями по умолчанию и допустимыми значениями, из загрузчика конфигурации Authfly 0.4.2.
---
# authfly.yaml

`authfly.yaml` — единственный источник правды развёртывания. `authfly init` его пишет, `authfly deploy` приводит систему к нему. Загрузчик строгий: неизвестный ключ — ошибка. Секретов здесь нет; они генерируются в `.authfly/` (права 0600).

## Верхний уровень

| Ключ | По умолчанию | Значения |
|---|---|---|
| `version` | `1` | только `1` |
| `name` | `authfly` | 2–31 символ: строчные буквы, цифры, дефис; проект compose — `authfly-<name>` |
| `target` | `docker` | `docker`, `incus` |
| `provider` | `stalwart` | `stalwart` (`supabase` и `authjs` отклоняются: это заглушки контракта) |
| `domain` | `example.com` | почтовый домен аккаунтов (`user@domain`) |
| `public_url` | `https://auth.<domain>` | публичный URL Authfly, он же OIDC-издатель |

## app

| Ключ | По умолчанию | Значения |
|---|---|---|
| `app.runtime` | `react` | `react`, `vue`, `svelte`, `solid`, `templ`, `twig`, `latte`, `jinja2` |
| `app.adapter` | `node` | `node`, `bun`, `php`, `python`, `go` |
| `app.clients[]` | нет | `{id, name, redirect_uris[]}`: OIDC-клиенты, регистрируемые в Stalwart |

## http

| Ключ | По умолчанию | Значения |
|---|---|---|
| `http.tls` | `acme` | `acme` (Let's Encrypt), `manual`, `off` |
| `http.port` | `443`, или `4000` при `tls: off` | публичный порт Authfly |
| `http.http_port` | `80` при `acme`, иначе `0` | ACME HTTP-01 и редирект; `0` — выключен |
| `http.cert_file`, `http.key_file` | нет | обязательны при `manual` |
| `http.acme_email` | нет | email аккаунта Let's Encrypt |

`public_url` может быть `https://` при `tls: off` (TLS завершает ваш прокси); `http://` при включённом TLS отклоняется.

## mail

| Ключ | По умолчанию | Значения |
|---|---|---|
| `mail.hostname` | `mail.<domain>` | имя MX-хоста |
| `mail.tls` | как `http.tls` | `acme`, `manual`, `off` (при `off` почтовые порты сохраняют TLS с самоподписанным сертификатом Stalwart) |
| `mail.storage` | `rocksdb` | только `rocksdb` |
| `mail.encryption_at_rest` | `false` | шифрование хранения в Stalwart |
| `mail.ports.smtp` / `submissions` / `imaps` | `25` / `465` / `993` | публичные почтовые порты |

## sign_in

| Ключ | По умолчанию | Смысл |
|---|---|---|
| `sign_in.magic_link` | `true` | magic-ссылка по почте |
| `sign_in.password` | `true` | email + пароль |
| `sign_in.username` | `false` | имя пользователя (локальная часть) + пароль |
| `sign_in.require_verify` | из `init` | подтверждение почты до первого входа |
| `sign_in.social`, `two_factor`, `webauthn` | `false` | в 0.4.2 должны оставаться `false` |

Если не задан ни один из `magic_link`, `password`, `username`, включаются magic-ссылка и пароль.

## policy

| Ключ | По умолчанию | Смысл |
|---|---|---|
| `policy.pack` | `none` | `none`, `gdpr`, `us-pii`, `ru-152fz` ([Пакеты политик](/ru/policy-packs)) |
| `policy.operator` | нет | юридическое имя контролёра / оператора |
| `policy.contact` | нет | контактный email по вопросам данных |
| `policy.us_states` | `[CA]` для `us-pii` | законы штатов в охвате (проверена только Калифорния) |
| `policy.residency` | `RU` для `ru-152fz` | ISO-код страны, где должны оставаться данные |
| `policy.host_country` | нет | где стоит сервер; заявляете вы, проверяет `doctor` |

## ports, channel, incus, backup

| Ключ | По умолчанию | Смысл |
|---|---|---|
| `ports.management` | `8080` | HTTP Stalwart, только на `127.0.0.1` |
| `ports.console` | `4500` | Authfly Console, только на `127.0.0.1` |
| `channel.name` | `stable` | имя канала |
| `channel.manifest` | `https://github.com/authfly/cli/releases/latest/download/channel.json` | URL или файл |
| `channel.version` | пусто (новейший) | релиз первого развёртывания; дальше — `authfly upgrade` |
| `channel.allow_unsigned` | `false` | только для dev-каналов, никогда для `https://` |
| `incus.network` / `subnet` | `authfly0` / `10.118.42.1/24` | приватный мост |
| `incus.pool` / `image` | `default` / `images:debian/13` | пул хранения и базовый образ |
| `incus.firewall` | `false` | записать и загрузить `/etc/nftables.d/authfly-<name>.nft` |
| `incus.no_limits` | `false` | не ставить лимиты памяти/процессов |
| `backup.dir` / `keep` | `backups` / `7` | зашифрованные бэкапы и сколько хранить |

## Пример

```yaml
version: 1
name: demo
target: docker
provider: stalwart
app:
  runtime: react
  adapter: node
  clients:
    - id: my-app
      name: My app
      redirect_uris:
        - http://localhost:5173/callback
domain: example.test
public_url: http://localhost:4000
http:
  port: 4000
  tls: "off"
sign_in:
  magic_link: true
  password: true
  require_verify: true
policy:
  pack: gdpr
  operator: Example Ltd
  contact: privacy@example.test
```

## Настройки сервера вне authfly.yaml

Это переменные окружения сервиса `authfly` (задаются в сгенерированном `.authfly/compose.yml`; следующий `deploy` его перезапишет):

| Переменная | По умолчанию | Смысл |
|---|---|---|
| `AUTHFLY_PUBLIC_SIGNUP` | `false` | `true` разрешает регистрацию без инвайта |
| `AUTHFLY_IDP` | `off` | `on`: Authfly сам подписывает ID-токены ([Концепции](/ru/concepts#oidc-два-режима)) |
| `AUTHFLY_IDP_KEYGEN` | не задана | `true` генерирует ключ подписи IdP |
| `AUTHFLY_TRUSTED_PROXIES` | `127.0.0.1,::1` | прокси, чьему `X-Forwarded-For` можно доверять |
| `AUTHFLY_RESERVED_NAMES_FILE` | не задана | дополнительные зарезервированные имена, по одному на строку |
| `SMTP_ADDR`, `SMTP_FROM` | задаёт deploy | релей исходящей почты и отправитель |
