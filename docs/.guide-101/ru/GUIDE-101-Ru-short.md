# Deployment 101 — короткий cheat sheet (RU)

## 1) Что важно помнить сразу

- Код проекта хранится в GitHub, релизы — в Docker Registry, а работающий сайт — в контейнере на сервере.
- `dashboard` хранит compose-файлы в `/opt/stacks/<APP_ID>`, но **не** хранит исходники приложения.
- Приложение хранится внутри image, а не в `/opt/stacks/<APP_ID>`.
- В проде используется `docker-compose.yml` с образом, локально — `docker-compose.local.yml`.
- Всегда делайте `APP_HEALTHCHECK_URL` пустым, пока домен не живой.

## 2) Где что и за что отвечает

- `Dockerfile` — собирает production image (`php:8.4-fpm-alpine` + `nginx` + app code).
- `docker-compose.yml` — шаблон для deploy через dashboard (`APP_IMAGE`).
- `.github/workflows/deploy.yml` — CI/CD (build → push → update dashboard → deploy).
- `docs/deployment/README.md` — краткая рабочая инструкция.
- `.project/GUIDE-101.md` — подробный разбор.
- `.project/GUIDE-101-Ru.md` — подробный русский вариант.

## 3) Как устроен CI/CD

1. Push в `main`.
2. GitHub Actions собирает image по `Dockerfile`.
3. Пушит теги:
   - `buildy-apps.registry.twcstorage.ru/phpfasty/starter:sha-<12char>`
   - `buildy-apps.registry.twcstorage.ru/phpfasty/starter:main`
4. Рендерит `docker-compose.yml` с подставленным `APP_IMAGE`.
5. `PUT /api/apps/${APP_ID}` в dashboard.
6. `POST /api/apps/${APP_ID}/deploy`.
7. Проверяет внешний health-check (если задан `APP_HEALTHCHECK_URL`).

## 4) Ключевые переменные GitHub

Variables:

- `REGISTRY_HOST=buildy-apps.registry.twcstorage.ru`
- `IMAGE_REPOSITORY=phpfasty/starter`
- `APP_NAME=starter`
- `APP_HEALTHCHECK_URL=https://authfly.ru/api/health` (можно временно пусто)

Secrets:

- `REGISTRY_USERNAME`
- `REGISTRY_PASSWORD`
- `DASHBOARD_URL`
- `DASHBOARD_USER`
- `DASHBOARD_PASS`
- `DASHBOARD_APP_ID`

## 5) Один раз: создание app в dashboard

```bash
git archive --format=tar HEAD | ssh root@your-server "mkdir -p /opt/bootstrap/starter && tar -xf - -C /opt/bootstrap/starter"
ssh root@your-server
cd /opt/bootstrap/starter

export DASHBOARD_URL="http://127.0.0.1:3000"
export DASHBOARD_USER="admin"
export DASHBOARD_PASS="admin@123"
export APP_NAME="starter"
export APP_IMAGE="buildy-apps.registry.twcstorage.ru/phpfasty/starter:sha-bootstrap"
export APP_PUBLIC_DOMAIN="authfly.ru"
export APP_PROXY_TARGET_PORT="80"

APP_IMAGE="${APP_IMAGE}" docker compose -f docker-compose.yml config > rendered-compose.yml
```

Сделать payload для создания приложения и отправить его через `POST /api/apps` (см. `docs/deployment/README.md`).

## 6) Критично: login в registry внутри dashboard

`docker login` на хосте НЕ хватает, если dashboard тянет image.

```bash
docker exec -it dashboard sh
docker login buildy-apps.registry.twcstorage.ru
exit
```

Проверка:

```bash
docker exec dashboard sh -lc 'docker pull buildy-apps.registry.twcstorage.ru/phpfasty/starter:sha-bootstrap'
```

## 7) Пошаговый быстрый deploy (после настройки)

1. `git add .`
2. `git commit ...`
3. `git push origin main`
4. Ждать Actions
5. Проверить в dashboard статус app

## 8) Экстренные проверки при сбоях

- В GitHub Actions fails early:
  - переменные/секреты не выставлены
- 500 из dashboard:
  - смотреть `docker logs --tail 120 dashboard`
  - проверить auth внутри dashboard
  - проверить, что тег image существует
- health check падает:
  - домен/SLL/routing/проксирование

Быстрый локальный чек:

```bash
docker ps --format 'table {{.Names}}\t{{.Image}}\t{{.Status}}'
docker exec starter-6e62b32b-web-1 sh -lc 'ls -la /app && ls -la /app/public'
```

## 9) Типовые ошибки новичка

- `/root/docker-compose.yml` — забыли перейти в директорию проекта перед `docker compose config`.
- Смешали dashboard URL и site URL.
- Ищете код в `/opt/stacks/<APP_ID>`.
- Думали, что `exit` закроет SSH.
- Уставивали `APP_HEALTHCHECK_URL`, пока домен ещё не поднят.

## 10) Финальная формула

GitHub = исходники  
Registry = образы  
Dashboard = compose-конфиг  
Docker = работающий сайт

## 11) Экстренный runbook на память (10 команд)

```bash
# 1) Запушить изменения
git add .
git commit -m "chore: update"
git push origin main

# 2) Проверить переменные/секреты в GitHub Actions (один раз)
echo "REGISTRY_HOST=$REGISTRY_HOST"
echo "IMAGE_REPOSITORY=$IMAGE_REPOSITORY"
echo "APP_NAME=$APP_NAME"
echo "APP_HEALTHCHECK_URL=$APP_HEALTHCHECK_URL"

# 3) Проверить последний запуск deploy workflow
gh run list --workflow deploy.yml --limit 1

# 4) Проверить доступ dashboard к registry
docker exec -it dashboard sh -lc "docker login buildy-apps.registry.twcstorage.ru"
docker exec dashboard sh -lc "docker pull buildy-apps.registry.twcstorage.ru/phpfasty/starter:main"

# 5) Manual redeploy (если нужно)
export DASHBOARD_URL="http://127.0.0.1:3000"
export DASHBOARD_USER="admin"
export DASHBOARD_PASS="admin@123"
export APP_ID="starter-6e62b32b"
curl -u "${DASHBOARD_USER}:${DASHBOARD_PASS}" -X POST "${DASHBOARD_URL}/api/apps/${APP_ID}/deploy"

# 6) Быстрая проверка статуса
curl "${DASHBOARD_URL}/api/apps/${APP_ID}"
curl "${DASHBOARD_URL}/api/apps/${APP_ID}/logs?lines=120"
docker ps --format 'table {{.Names}}\t{{.Image}}\t{{.Status}}'
curl -I https://authfly.ru/
curl https://authfly.ru/api/health

# 7) Если упало на dashboard 500
docker logs --tail 120 dashboard
docker exec dashboard sh -lc "docker pull buildy-apps.registry.twcstorage.ru/phpfasty/starter:main"

# 8) Если домен еще не живой — временно без smoke check
# APP_HEALTHCHECK_URL должно быть пустым в GitHub Variables
```

---

## Быстрые переходы в подробный Russian guide

- [1. Ключевая картина](.project/GUIDE-101-Ru.md#1-ключевая-картина)
- [2. Что делает каждый файл](.project/GUIDE-101-Ru.md#2-что-делает-каждый-файл)
- [7. Что такое health check и зачем их две](.project/GUIDE-101-Ru.md#7-что-такое-health-check-и-зачем-их-две)
- [10. GitHub Actions workflow, разбор по шагам](.project/GUIDE-101-Ru.md#10-github-actions-workflow-разбор-по-шагам)
- [11. Репозиторные переменные и secrets](.project/GUIDE-101-Ru.md#11-репозиторные-переменные-и-secrets)
- [12. Первый деплой с нуля](.project/GUIDE-101-Ru.md#12-первый-деплой-с-нуля)
- [17. Ошибки, которые уже случались](.project/GUIDE-101-Ru.md#17-ошибки-которые-уже-случались)
- [19. Что делать, если деплой снова упадет](.project/GUIDE-101-Ru.md#19-что-делать-если-деплой-снова-упадет)
