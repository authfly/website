---
title: Установка Authfly + Stalwart
nav: Установка Authfly + Stalwart
description: Установка на свой сервер для продакшена: топология, DNS для почты и TLS, authfly.yaml, развёртывание, проверки закрытости, бэкапы и эксплуатация.
---
# Установка Authfly + Stalwart

Гайд ведёт от чистого Linux-сервера до публичного Authfly со Stalwart за ним. Он объединяет исходный код и документацию Authfly (релиз 0.4.2) с полевым отчётом оператора, который держит Authfly на Incus; пункты из отчёта помечены, а всё, что не подтверждено кодом, помечено «НЕ ПРОВЕРЕНО».

Плейсхолдеры: `auth.example.com` (Authfly, OIDC-издатель), `example.com` (почтовый домен), `mail.example.com` (почтовый хост), `app.example.com` (ваше приложение), `203.0.113.10` (адрес сервера).

## 1. Топология

| Вариант | TLS | Когда |
|---|---|---|
| A. `target: docker`, `tls: acme` | Authfly сам получает сертификаты Let's Encrypt (HTTP-01 на порту 80) | один сервер, на 80/443 больше ничего нет |
| B. `target: docker` или `incus`, `tls: off`, `public_url: https://…` | TLS завершает ваш обратный прокси (Caddy, nginx) и проксирует на Authfly :4000 | прокси для других сайтов уже есть |
| C. `target: incus` | как A или B | вы предпочитаете системные контейнеры; `authfly deploy` создаёт сеть `authfly0` и инстансы |

В любом варианте публичны только Authfly (и почтовые порты, если нужна почта). HTTP-листенер Stalwart (управление, JMAP, `/api/*`) остаётся на `127.0.0.1` и во внутренней сети.

## 2. Подготовка сервера

- Поддерживаемый Linux с Docker Engine + плагином Compose (варианты A/B) или Incus (вариант C).
- Откройте 80 и 443; 25, 465 и 993 — только если нужна входящая и исходящая почта.
- Запускайте `deploy` от root: он пишет секреты с владельцем — пользователем контейнера (uid 10001); без root этот шаг не проходит.

## 3. DNS

Записи для входа:

| Имя | Тип | Значение |
|---|---|---|
| `auth.example.com` | A / AAAA | `203.0.113.10` |
| `app.example.com` | A / AAAA | хост вашего приложения |

Записи для почты (по документации Stalwart о DNS):

| Имя | Тип | Значение |
|---|---|---|
| `example.com` | MX | `10 mail.example.com` |
| `mail.example.com` | A / AAAA | `203.0.113.10` |
| `example.com` | TXT (SPF) | `v=spf1 mx -all` |
| `mail.example.com` | TXT (SPF) | `v=spf1 a -all` |
| `<selector>._domainkey.example.com` | TXT (DKIM) | публичные ключи, которые генерирует Stalwart (селекторы Ed25519 и RSA) |
| `_dmarc.example.com` | TXT | начните с `v=DMARC1; p=none; rua=mailto:dmarc@example.com`, переходите на `quarantine` или `reject`, когда отчёты чистые |
| `203.0.113.10` | PTR | `mail.example.com`, настраивается у хостера |

> **Не проверено:** точную зону (с селекторами DKIM) Stalwart показывает в веб-админке или через `stalwart-cli`. Authfly выключает веб-админку, поэтому читайте её через CLI изнутри приватной сети. Эта команда CLI с Authfly 0.4.2 не проверялась. PTR — общая практика для почты, Authfly его не настраивает.

> **Примечание:** для входа внешняя почта не нужна, если адреса пользователей на домене Stalwart: письма в локальные ящики доставляются внутри. Доставка на другие домены требует записей выше; для этого гайда она не проверялась.

## 4. authfly.yaml

Вариант A, TLS завершает Authfly:

```sh
authfly init -y --name prod --target docker \
  --runtime react --adapter node \
  --domain example.com --public-url https://auth.example.com \
  --tls acme --acme-email ops@example.com \
  --mail-hostname mail.example.com \
  --signin magic,password \
  --client "my-app:My app=https://app.example.com/callback" \
  --policy gdpr --operator "Example Ltd" --contact privacy@example.com
```

Вариант B, TLS завершает ваш прокси: `--tls off --public-url https://auth.example.com`. `http.tls: off` с `https://` в public URL разрешён именно для этого случая; cookies тогда `Secure`. Authfly слушает порт 4000.

Ключевые решения (все ключи: [authfly.yaml](/ru/reference/authfly-yaml)):

- `public_url` — это OIDC-издатель. Выберите его один раз: от него зависят токены и регистрации клиентов.
- `app.clients` перечисляет ваши приложения и их точные redirect URI. Authfly также регистрирует собственного клиента `authfly` с redirect `<public_url>/auth/callback`; вручную создавать OIDC-клиентов в Stalwart не нужно.
- `sign_in.social`, `two_factor` и `webauthn` в 0.4.2 должны оставаться `false`.

## 5. Развёртывание

```sh
sudo authfly deploy
sudo authfly status
sudo authfly doctor
```

Что делает `deploy`: проверяет подписанный канал, генерирует секреты в `.authfly/` (0600), рендерит документы пакета политик, запускает Stalwart, выполняет `stalwart-init` (создаёт домен, сервисный аккаунт с минимальными правами `authfly-svc@<domain>`, выключает веб-админку и приложение аккаунта, закрывает регистрацию OIDC-клиентов и регистрирует клиентов), затем запускает Authfly.

> **Известная проблема (0.4.2):** канал закрепляет Stalwart v0.16.27, а серверу нужен Stalwart 1.0; перед продакшен-установкой см. [Релизы](/ru/releases#известные-проблемы).

### Обратный прокси (вариант B)

Проксируйте только Authfly. Пример для Caddy:

```text
auth.example.com {
	reverse_proxy 127.0.0.1:4000
}
```

Никогда не проксируйте HTTP-порт Stalwart, даже с фильтром путей: тот же листенер обслуживает управление, JMAP и динамическую регистрацию клиентов. Authfly публикует только белый список OIDC-эндпоинтов. Адрес клиента приходит в `X-Forwarded-For`; Authfly доверяет ему только от `AUTHFLY_TRUSTED_PROXIES` (по умолчанию `127.0.0.1,::1`). Добавьте туда адрес прокси, если он на другой машине.

## 6. Проверки закрытости

Снаружи сервера:

```sh
curl -sI https://auth.example.com/.well-known/openid-configuration   # 200
curl -sI https://auth.example.com/jmap                               # 404 или 405
curl -sI https://auth.example.com/api/principal                      # 404 или 405
curl -sI https://auth.example.com/admin/                             # 404
```

Проверьте также, что порты 8080 (HTTP Stalwart), 4500 (Console) и 4001 (admin API) недоступны снаружи. Console открывайте только через SSH-туннель:

```sh
ssh -N -L 4500:127.0.0.1:4500 you@203.0.113.10
```

## 7. Первые пользователи

Публичная регистрация выключена (`AUTHFLY_PUBLIC_SIGNUP=false`), поэтому нужен инвайт:

```sh
docker compose -p authfly-prod -f .authfly/compose.yml exec -T authfly app invite create -note "first admin" -ttl 72h
```

Токен печатается один раз; передайте его по закрытому каналу. `invite list` и `invite revoke <id>` управляют инвайтами.

Зарезервированные имена: встроенный список (`admin`, `postmaster`, `abuse`, `root`, `support` и ещё около шестидесяти) занять нельзя. Чтобы добавить имена, смонтируйте файл — одно имя на строку, комментарии через `#` — и задайте `AUTHFLY_RESERVED_NAMES_FILE` у сервиса authfly. Файл дополняет встроенный список. Это переменная окружения, а не ключ `authfly.yaml`.

Смоук-тест: инвайт, регистрация, подтверждение почты, вход, `/account`, выход; попробуйте зарезервированное и уже занятое имя (оба отклоняются).

## 8. Эксплуатация

- `authfly backup` пишет зашифрованный бэкап (AES-256-GCM с сгенерированным `backup_key`) в `backup.dir` и хранит `backup.keep` штук (по умолчанию 7). `authfly restore <file.afbk>` восстанавливает. Бэкап доказывает только восстановление: отрепетируйте его на запасном хосте.
- `authfly upgrade` делает снимок, обновляет до новейшего релиза (или `--to <version>`) и сам откатывается при ошибке; `authfly rollback` восстанавливает последний снимок.
- `authfly secrets` показывает секреты и их права; значения не печатаются никогда.
- `authfly doctor` сверяет установку с `authfly.yaml` и пакетом политик (документы отдаются, в журнале Authfly нет IP-адресов, подпись канала, digest образов).

## 9. Уроки из практики

От оператора, который держит Authfly, Stalwart и Caddy в отдельных непривилегированных контейнерах Incus (полевой отчёт, адаптирован):

- **Храните секреты в файлах** (только root, `chmod 600`), никогда в argv, истории shell или репозитории. Спецсимволы в секретах ломают кавычки shell и YAML. `authfly deploy` уже пишет свои секреты файлами 0600 в `.authfly/`.
- **Никогда не публикуйте админку Stalwart** и проверяйте снаружи, что её пути отдают 404.
- **Закрепите адреса контейнеров**, чтобы upstream прокси переживали перезапуск.
- **Если запускаете бинарник под systemd сами** (вне контейнеров, которыми управляет `deploy`), стартуйте после поднятия сети:

```ini
# /etc/systemd/system/authfly.service.d/10-network-online.conf
[Unit]
Wants=network-online.target
After=network-online.target

[Service]
Restart=on-failure
RestartSec=5
```

- **Снимок перед каждым deploy** и бэкапы вне хоста (оператор использует `incus snapshot` со сроком 7 дней и ночной restic); закройте SSH и ограничьте попытки входа (например, fail2ban).
- **Оберните создание инвайтов** в маленький скрипт только для root, чтобы токены не попадали в историю shell.
- **Тест перезагрузки:** после reboot все сервисы должны подняться сами.

> **Не проверено:** полевой отчёт написан для предрелизной сборки Authfly (0.5.x) со Stalwart 1.0.0 в роли OIDC IdP и выключенными почтовыми листенерами; в нём client id и secret читались из файла, подпись Stalwart — ES256. Authfly 0.4.2 сам генерирует секреты и регистрации клиентов, поэтому для этих настроек ключей в `authfly.yaml` нет. Docker-вариант этой схемы оператор не проверял; шаги с Docker в этом гайде проверены отдельно (см. [React + Docker на Authfly](/ru/guides/react-docker)).
