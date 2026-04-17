# Deployment 101 — мини чек-лист (RU)

## Инициализация (один раз)

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
python3 -c 'import json, os; from pathlib import Path; payload={"name": os.environ["APP_NAME"], "compose_yaml": Path("rendered-compose.yml").read_text(encoding="utf-8")}; Path("create-app.json").write_text(json.dumps(payload, ensure_ascii=True), encoding="utf-8")'
curl -u "${DASHBOARD_USER}:${DASHBOARD_PASS}" -X POST "${DASHBOARD_URL}/api/apps" -H 'Content-Type: application/json' --data @create-app.json

# в ответе взять id:
# {"id":"starter-6e62b32b",...}
export APP_ID="starter-6e62b32b"

cat > app-config.json <<'EOF'
{
  "public_domain": "authfly.ru",
  "proxy_target_port": 80,
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
EOF

curl -u "${DASHBOARD_USER}:${DASHBOARD_PASS}" -X PUT "${DASHBOARD_URL}/api/apps/${APP_ID}/config" -H 'Content-Type: application/json' --data @app-config.json

docker exec -it dashboard sh
docker login buildy-apps.registry.twcstorage.ru
exit

curl -u "${DASHBOARD_USER}:${DASHBOARD_PASS}" -X POST "${DASHBOARD_URL}/api/apps/${APP_ID}/deploy"
```

## Ежедневный деплой

```bash
git add .
git commit -m "..."
git push origin main
# дождаться GitHub Actions, затем проверить
```

## Проверки

```bash
curl -u "${DASHBOARD_USER}:${DASHBOARD_PASS}" "${DASHBOARD_URL}/api/apps/${APP_ID}"
curl -u "${DASHBOARD_USER}:${DASHBOARD_PASS}" "${DASHBOARD_URL}/api/apps/${APP_ID}/logs?lines=120"
docker ps --format 'table {{.Names}}\t{{.Image}}\t{{.Status}}'
docker exec starter-6e62b32b-web-1 sh -lc 'ls -la /app && ls -la /app/public'
curl -I https://authfly.ru/
curl https://authfly.ru/api/health
```

## При сбоях

```bash
docker logs --tail 120 dashboard
docker exec dashboard sh -lc 'docker login buildy-apps.registry.twcstorage.ru'
docker exec dashboard sh -lc 'docker pull buildy-apps.registry.twcstorage.ru/phpfasty/starter:main'
```

## Быстрый deploy без нового push

```bash
curl -u "${DASHBOARD_USER}:${DASHBOARD_PASS}" -X POST "${DASHBOARD_URL}/api/apps/${APP_ID}/deploy"
```
