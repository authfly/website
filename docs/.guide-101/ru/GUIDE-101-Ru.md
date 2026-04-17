# Deployment 101 для `starter`

Это руководство написано для «забытого» будущего меня — полностью с нуля.

Предположения:

- Я забыл, как работают GitHub Actions, Docker, Docker Registry, Dashboard API и деплой
- Я хочу развернуть проект заново
- Мне нужен один документ, который объясняет и **что**, и **почему**

В этом документе используются реальные значения этого проекта:

- Репозиторий GitHub: `phpfasty/starter`
- Хост Docker Registry: `buildy-apps.registry.twcstorage.ru`
- Путь образа в registry: `phpfasty/starter`
- Публичный домен: `authfly.ru`
- Имя приложения в dashboard: `starter`

---

## 1. Ключевая картина

Поток деплоя для этого проекта такой:

1. Код лежит в GitHub
2. GitHub Actions собирает Docker-образ из `Dockerfile`
3. GitHub Actions пушит образ в приватный registry
4. GitHub Actions рендерит `docker-compose.yml` с конкретным тегом образа
5. GitHub Actions обновляет приложение в dashboard через API
6. Dashboard выполняет `docker compose up -d`
7. Сайт становится доступен на `authfly.ru`

Важно:

- Код сайта **не** хранится в `/opt/stacks/<APP_ID>`
- Код сайта хранится **внутри Docker-образа**
- `/opt/stacks/<APP_ID>` хранит compose-файлы, которыми управляет dashboard

---

## 2. Что делает каждый файл

### `Dockerfile`

Этот файл описывает, как собирается production-образ.

Что он делает в этом проекте:

- ставит зависимости PHP через Composer на стадии сборки
- собирает runtime-образ на основе `php:8.4-fpm-alpine`
- ставит `nginx`, `APCu`, настройки PHP
- копирует исходники приложения в `/app`
- открывает внутренний порт `80`
- задает внутренний health-check `http://127.0.0.1/api/health`

Проще:

- `Dockerfile` говорит Docker, как превратить репозиторий в готовый запускаемый контейнерный образ

### `docker-compose.yml`

Это **production шаблон compose**.

Его используют:

- шаг рендеринга в GitHub Actions
- деплой через dashboard

Это **не** локальный dev stack.

Что в нем указано:

- один сервис `web`
- образ `${APP_IMAGE}`
- expose внутреннего порта `80`
- переменные окружения передаются в контейнер
- монтируются два Docker volume для кэшей

Проще:

- `docker-compose.yml` описывает dashboard, как запускать уже собранный образ

### `docker-compose.local.yml`

Это **старый локальный stack для разработки**.

Зачем нужен:

- локальная разработка
- bind mounts
- ручная работа с Docker локально и просмотр изменений сразу на хосте

В registry-based проде **не используется**.

### `.github/workflows/deploy.yml`

Это workflow для GitHub Actions.

Он автоматизирует деплой при push в `main`.

Проще:

- это «робот», который собирает, пушит, обновляет и запускает деплой

### `docs/deployment/README.md`

Короткий рабочий мануал по деплою.

### Файл `.project/GUIDE-101.md`

Длинное объяснение и recovery-гайд для новичка.

---

## 3. Самое важное разделение

Есть **три разных места**, участвующих в деплое.

### 1. Репозиторий GitHub

Здесь исходный код.

Примеры:

- `Dockerfile`
- `docker-compose.yml`
- `src/`
- `public/`
- `templates/`

### 2. Docker Registry

Здесь лежат собранные контейнерные образы.

Пример:

- `buildy-apps.registry.twcstorage.ru/phpfasty/starter:sha-b1db41199ff1`

### 3. Сервер с Dashboard

Здесь хранятся runtime compose-конфиги и запускаются контейнеры.

Пример пути:

- `/opt/stacks/starter-6e62b32b`

В этом пути лежат только файлы стека, не исходники приложения.

---

## 4. Где физически живет сайт

### На GitHub

Исходный код хранится в git-репозитории.

### В Docker Registry

Артефакт production хранится в виде Docker-образа.

### На сервере после деплоя

Три важных типа данных:

1. Файлы dashboard-стека
   Пример:
   `/opt/stacks/starter-6e62b32b/docker-compose.yml`

2. Файловая система контейнера
   Код приложения внутри работающего контейнера: `/app`

3. Docker volumes
   Кэш хранится в Docker volumes, а не в директории репозитория

Полезные команды на сервере:

```bash
docker exec starter-6e62b32b-web-1 sh -lc 'ls -la /app && ls -la /app/public'
docker volume inspect starter-6e62b32b_app-cache starter-6e62b32b_template-cache
docker inspect starter-6e62b32b-web-1 --format '{{json .GraphDriver.Data}}'
```

---

## 5. Имена и значения, которые используются здесь

### GitHub репозиторий

- owner: `phpfasty`
- repo: `starter`

### Registry

- хост: `buildy-apps.registry.twcstorage.ru`
- путь образа: `phpfasty/starter`

Итоговый вид ссылок на образы:

- `buildy-apps.registry.twcstorage.ru/phpfasty/starter:main`
- `buildy-apps.registry.twcstorage.ru/phpfasty/starter:sha-b1db41199ff1`

### Dashboard app

- имя: `starter`
- пример app id: `starter-6e62b32b`

### Публичный сайт

- `authfly.ru`

### Health check

После того как домен включен, внешний health check:

- `https://authfly.ru/api/health`

Важно:

- dashboard URL не равен app health check URL

---

## 6. Dashboard URL vs app URL vs health check

Это три разных понятия.

### Dashboard URL

Это административная панель API.

Примеры:

- `http://127.0.0.1:3000` если команды выполняются на сервере
- `https://gui.example.com` если GitHub Actions должен обращаться к dashboard извне

### Публичный домен приложения

Это реальный сайт:

- `https://authfly.ru`

### Health check URL

Это endpoint здоровья **сайта**, а не dashboard:

- `https://authfly.ru/api/health`

Health endpoint приходит из самого приложения.

---

## 7. Что такое health check и зачем их две

У проекта два разных health check.

### Внутренний контейнерный health check

Определен в `Dockerfile`.

Проверяет:

- `http://127.0.0.1/api/health` внутри контейнера

Это показывает Docker, здоров ли сам контейнер.

### Внешний smoke test в GitHub Actions

Определен в `.github/workflows/deploy.yml`.

Проверяет:

- `${APP_HEALTHCHECK_URL}`

Это проверка, что сайт реально доступен извне после деплоя.

Важно для новичка:

- если публичный домен еще не готов, оставьте `APP_HEALTHCHECK_URL` пустым
- иначе workflow может падать, хотя деплой уже успешен

---

## 8. Почему `docker-compose.yml` такой маленький

Production compose намеренно маленький, потому что:

- код уже включен в образ
- в образ уже встроен Nginx и PHP-FPM
- dashboardу нужен только рецепт запуска

Текущий production шаблон:

```yaml
services:
  web:
    image: ${APP_IMAGE:?APP_IMAGE is required}
    expose:
      - "80"
    restart: unless-stopped
    environment:
      APP_ENV: ${APP_ENV:-production}
      APP_DEBUG: ${APP_DEBUG:-false}
      DATA_SOURCE: ${DATA_SOURCE:-fixtures}
      CACHE_TTL: ${CACHE_TTL:-3600}
      CACHE_DIR: ${CACHE_DIR:-/app/cache}
      FIXTURES_PATH: ${FIXTURES_PATH:-/app/fixtures}
      DEFENSE_CONFIG_PATH: ${DEFENSE_CONFIG_PATH:-/app/config/defense.php}
      APP_WARM_CACHE: ${APP_WARM_CACHE:-0}
    volumes:
      - app-cache:/app/cache
      - template-cache:/app/templates/cache
```

Это означает:

- запуск ровно одного контейнера `web`
- контейнер берется по тегу, который сформировал CI
- runtime-кэш идет в Docker volumes

---

## 9. Почему `docker-compose.local.yml` всё еще существует

Потому что локальная разработка и production — разные контексты.

### `docker-compose.local.yml`

Используется для:

- локальной разработки
- bind mounts
- немедленного отображения изменений с хоста

### `docker-compose.yml`

Используется для:

- production
- деплоя через dashboard
- deploy по готовому image из registry

Правило новичка:

- локально: `docker-compose.local.yml`
- прод: `docker-compose.yml`

---

## 10. GitHub Actions workflow, разбор по шагам

Файл workflow: `.github/workflows/deploy.yml`.

### Триггеры

```yaml
on:
  push:
    branches:
      - main
  workflow_dispatch:
```

Значение:

- каждый push в `main` триггерит деплой
- доступен ручной запуск из UI GitHub

### Node 24 флаг

```yaml
FORCE_JAVASCRIPT_ACTIONS_TO_NODE24: "true"
```

Значение:

- GitHub предупреждает о deprecation Node.js 20
- workflow уже принудительно включает Node.js 24

### Переменные и секреты

Workflow читает:

- переменные репозитория через `vars.*`
- секреты репозитория через `secrets.*`

Если переменные пустые, workflow падает на ранней проверке.

### Подготовка тега образа

Берутся первые 12 символов SHA коммита.

Пример:

- полный SHA: `b1db41199ff1d7c8...`
- тег: `sha-b1db41199ff1`

### Build и push

Workflow пушит два тега:

- `buildy-apps.registry.twcstorage.ru/phpfasty/starter:sha-<commit>`
- `buildy-apps.registry.twcstorage.ru/phpfasty/starter:main`

### Render compose

Эта стадия **не** деплоит.

Она только подставляет `${APP_IMAGE}` и пишет финальный compose.

### Обновление dashboard app

Workflow отправляет:

- `PUT /api/apps/${APP_ID}`

с новым `compose_yaml`.

### Deploy app

Workflow отправляет:

- `POST /api/apps/${APP_ID}/deploy`

Dashboard затем выполняет `docker compose up -d`.

---

## 11. Репозиторные переменные и secrets

### Variables (Actions → Variables)

Настраиваются в:

- GitHub repo -> Settings -> Secrets and variables -> Actions -> Variables

Текущие значения:

- `REGISTRY_HOST=buildy-apps.registry.twcstorage.ru`
- `IMAGE_REPOSITORY=phpfasty/starter`
- `APP_NAME=starter`
- `APP_HEALTHCHECK_URL=https://authfly.ru/api/health`

Важно:

- если домен еще не опубликован, `APP_HEALTHCHECK_URL` можно оставить пустым

### Secrets (Actions → Secrets)

Настраиваются в:

- GitHub repo -> Settings -> Secrets and variables -> Actions -> Secrets

Что должно быть выставлено:

- `REGISTRY_USERNAME`
- `REGISTRY_PASSWORD`
- `DASHBOARD_URL`
- `DASHBOARD_USER`
- `DASHBOARD_PASS`
- `DASHBOARD_APP_ID`

Пример:

- `DASHBOARD_APP_ID=starter-6e62b32b`

---

## 12. Первый деплой с нуля

Предположения:

- приложение еще не развернуто
- на сервере доступен Docker
- dashboard уже установлен и работает

### Шаг 1. Залить файлы проекта на сервер

Если Docker доступен только на сервере:

```bash
git archive --format=tar HEAD | ssh root@your-server "mkdir -p /opt/bootstrap/starter && tar -xf - -C /opt/bootstrap/starter"
```

Почему так:

- передаются только tracked файлы
- не попадет `.git/`
- не попадает локальный мусор

### Шаг 2. Подключиться по SSH

```bash
ssh root@your-server
cd /opt/bootstrap/starter
```

### Шаг 3. Экспорт переменных

Пример:

```bash
export DASHBOARD_URL="http://127.0.0.1:3000"
export DASHBOARD_USER="admin"
export DASHBOARD_PASS="admin@123"

export APP_NAME="starter"
export APP_IMAGE="buildy-apps.registry.twcstorage.ru/phpfasty/starter:sha-bootstrap"
export APP_PUBLIC_DOMAIN="authfly.ru"
export APP_PROXY_TARGET_PORT="80"
```

### Шаг 4. Рендер production compose

```bash
APP_IMAGE="${APP_IMAGE}" docker compose -f docker-compose.yml config > rendered-compose.yml
```

Эта команда **не** запускает контейнеры.

Она только генерирует итоговый YAML с подставленным тегом образа.

### Шаг 5. Сборка payload для dashboard

```bash
python3 -c 'import json, os; from pathlib import Path; payload={"name": os.environ["APP_NAME"], "compose_yaml": Path("rendered-compose.yml").read_text(encoding="utf-8")}; Path("create-app.json").write_text(json.dumps(payload, ensure_ascii=True), encoding="utf-8")'
```

### Шаг 6. Создать приложение в dashboard

```bash
curl -u "${DASHBOARD_USER}:${DASHBOARD_PASS}" \
  -X POST "${DASHBOARD_URL}/api/apps" \
  -H 'Content-Type: application/json' \
  --data @create-app.json
```

В ответе приходит `id`.

Пример:

```bash
export APP_ID="starter-6e62b32b"
```

### Шаг 7. Настроить public domain и env

```bash
cat > app-config.json
{
  "public_domain": "${APP_PUBLIC_DOMAIN}",
  "proxy_target_port": ${APP_PROXY_TARGET_PORT},
  "use_tls": false,
  "managed_env": {
    "APP_ENV": "production",
    "APP_DEBUG": "false",
    "DATA_SOURCE": "fixtures",
    "CACHE_TTL": "3600",
    "CACHE_DIR": "/app/cache",
    "FIXTURES_PATH": "/app/fixtures",
    "DEFENSE_CONFIG_PATH": "/app/config/defense.php",
    "APP_WARM_CACHE": "1"
  }
}
```

Вставьте этот JSON в `app-config.json` и отправьте:

```bash
curl -u "${DASHBOARD_USER}:${DASHBOARD_PASS}" \
  -X PUT "${DASHBOARD_URL}/api/apps/${APP_ID}/config" \
  -H 'Content-Type: application/json' \
  --data @app-config.json
```

### Шаг 8. Логин в приватный registry внутри контейнера dashboard

Это очень важно.

Логин должен выполняться **внутри контейнера `dashboard`**, потому что именно он запускает Docker-команды.

```bash
docker exec -it dashboard sh
docker login buildy-apps.registry.twcstorage.ru
exit
```

Что означает `exit`:

- выходит только из shell внутри контейнера
- сам контейнер **не** останавливается
- SSH сессия пользователя не закрывается

Почему новички путаются:

- `docker exec -it dashboard sh` открывает shell внутри контейнера
- в этом shell `exit` означает «выйти из контейнера`
- после `exit` вы возвращаетесь в обычный host shell

Дополнительная проверка:

```bash
docker exec dashboard sh -lc 'docker pull buildy-apps.registry.twcstorage.ru/phpfasty/starter:sha-bootstrap'
```

### Шаг 9. Первый деплой

```bash
curl -u "${DASHBOARD_USER}:${DASHBOARD_PASS}" \
  -X POST "${DASHBOARD_URL}/api/apps/${APP_ID}/deploy"
```

### Шаг 10. Включить TLS позже

Когда DNS и HTTP роутинг уже работают, включите HTTPS в dashboard и проверьте:

```bash
curl -I https://authfly.ru/
curl https://authfly.ru/api/health
```

---

## 13. Как работаем ежедневно после настройки

После первичной настройки обычно:

1. меняем код локально
2. коммитим
3. пушим в `main`
4. GitHub Actions собирает и пушит образ
5. GitHub Actions обновляет compose в dashboard
6. GitHub Actions триггерит deploy

Это стандартный путь.

Нечего вручную зальвать репозиторий на каждый релиз.

---

## 14. Как вручную перезапустить деплой

Если нужно быстро перезапустить текущее развернутое состояние без нового push:

```bash
export DASHBOARD_URL="http://127.0.0.1:3000"
export DASHBOARD_USER="admin"
export DASHBOARD_PASS="admin@123"
export APP_ID="starter-6e62b32b"

curl -u "${DASHBOARD_USER}:${DASHBOARD_PASS}" \
  -X POST "${DASHBOARD_URL}/api/apps/${APP_ID}/deploy"
```

---

## 15. Как узнать текущий тег образа

Workflow формирует такие теги:

- `sha-b1db41199ff1`

Это первые 12 символов `GITHUB_SHA`.

Как посмотреть:

- в логах GitHub Actions
- через compose приложения в dashboard
- через image запущенного контейнера

Полезные команды:

```bash
docker inspect starter-6e62b32b-web-1 --format '{{.Config.Image}}'
sed -n '1,160p' /opt/stacks/starter-6e62b32b/docker-compose.yml
```

## 16. Tag vs digest

Есть два корректных варианта ссылок:

### По тегу

```bash
buildy-apps.registry.twcstorage.ru/phpfasty/starter:sha-b1db41199ff1
```

### По digest

```bash
buildy-apps.registry.twcstorage.ru/phpfasty/starter@sha256:xxxxxxxx
```

Важно:

- `:sha256:...` — неверно
- digest задается через `@sha256:...`
- текущий workflow деплоит **по тегу**

---

## 17. Ошибки, которые уже случались

### Ошибка 1. Выполнение `docker compose -f docker-compose.yml config` в `/root`

Проблема:

- `open /root/docker-compose.yml: no such file or directory`

Фикс:

- сначала `cd /opt/bootstrap/starter`

### Ошибка 2. Путаница dashboard URL и сайта

Неправильно:

- использовать dashboard URL как app health check

Правильно:

- dashboard URL используется только для API dashboard
- app health check: `https://authfly.ru/api/health`

### Ошибка 3. Ожидание, что в `/opt/stacks/<APP_ID>` лежит код приложения

Неправильное предположение.

Реальность:

- там хранятся compose-файлы
- код в образе и внутри контейнера

### Ошибка 4. Логин только на хосте

Проблема:

- deploy продолжает падать с `401 Unauthorized`

Причина:

- логин нужен именно контейнеру dashboard

Фикс:

- выполнить `docker login` внутри `dashboard`

### Ошибка 5. Ожидание, что `exit` закроет SSH

Когда вы внутри:

```bash
docker exec -it dashboard sh
```

`exit` выдает только из shell контейнера и возвращает в host shell.

### Ошибка 6. `APP_HEALTHCHECK_URL` обязателен всегда

Если публичный домен еще не жив:

- оставляйте переменную пустой

Иначе:

- workflow может упасть на smoke test, хотя deploy уже успешный

---

## 18. Как проверить всё целиком

### Проверить состояние приложения в dashboard

```bash
curl -u "${DASHBOARD_USER}:${DASHBOARD_PASS}" \
  "${DASHBOARD_URL}/api/apps/${APP_ID}"
```

### Проверить логи приложения dashboard

```bash
curl -u "${DASHBOARD_USER}:${DASHBOARD_PASS}" \
  "${DASHBOARD_URL}/api/apps/${APP_ID}/logs?lines=120"
```

### Проверить запущенные контейнеры

```bash
docker ps --format 'table {{.Names}}\t{{.Image}}\t{{.Status}}'
```

### Проверить сайт с сервера

```bash
curl -I https://authfly.ru/
curl https://authfly.ru/api/health
```

### Проверить код внутри running контейнера

```bash
docker exec starter-6e62b32b-web-1 sh -lc 'ls -la /app && ls -la /app/public'
```

---

## 19. Что делать, если деплой снова упадет

### Если GitHub Actions падает до сборки

Проверить:

- repository variables
- repository secrets

### Если push в registry не проходит

Проверить:

- `REGISTRY_USERNAME`
- `REGISTRY_PASSWORD`
- host registry

### Если dashboard возвращает `500`

Проверить:

```bash
docker logs --tail 120 dashboard
```

Чаще всего причина одна из трех:

- отсутствует auth внутри `dashboard`
- такого image tag не существует
- не стартует compose на стороне контейнера

### Если не работает public HTTPS

Проверить:

- настройки TLS в dashboard
- DNS записи для `authfly.ru`
- роутинг Nginx, созданный dashboard
- логи выдачи сертификата

---

## 20. Минимальный экстренный чек-лист

Если «всё забыл», выполняй:

1. убедиться в APP ID
   `starter-6e62b32b`
2. убедиться, что образ есть в registry
3. убедиться, что контейнер `dashboard` работает
4. убедиться, что login в registry есть внутри `dashboard`
5. убедиться, что secrets и variables в GitHub всё еще есть
6. сделать push в `main`
7. если деплой не идет — посмотреть:
   `docker logs --tail 120 dashboard`
8. если деплой успешен — проверить:
   `curl https://authfly.ru/api/health`

---

## 21. Финальная ментальная модель

Запомни эту фразу:

> GitHub хранит исходники, registry хранит релизные образы, dashboard хранит runtime compose-конфиг, а Docker запускает реальный сайт.

Если запомнишь только это, остальное можно восстановить.
