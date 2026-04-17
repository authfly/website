# Deployment 101 — ultra short runbook (RU)

```bash
# 1) Commit and push
git add .
git commit -m "chore: update"
git push origin main

# 2) CI variables are set (one-time check in GitHub > Actions > Variables/Secrets)
echo "REGISTRY_HOST=$REGISTRY_HOST"
echo "IMAGE_REPOSITORY=$IMAGE_REPOSITORY"
echo "APP_NAME=$APP_NAME"
echo "APP_HEALTHCHECK_URL=$APP_HEALTHCHECK_URL"

# 3) Check last deploy run (GitHub Actions UI)
gh run list --workflow deploy.yml --limit 1

# 4) On server: ensure dashboard can see registry
docker exec -it dashboard sh -lc "docker login buildy-apps.registry.twcstorage.ru"
docker exec dashboard sh -lc "docker pull buildy-apps.registry.twcstorage.ru/phpfasty/starter:main"

# 5) Manual redeploy (if needed)
export DASHBOARD_URL="http://127.0.0.1:3000"
export DASHBOARD_USER="admin"
export DASHBOARD_PASS="admin@123"
export APP_ID="starter-6e62b32b"
curl -u "${DASHBOARD_USER}:${DASHBOARD_PASS}" -X POST "${DASHBOARD_URL}/api/apps/${APP_ID}/deploy"

# 6) Verify
curl "${DASHBOARD_URL}/api/apps/${APP_ID}"
curl "${DASHBOARD_URL}/api/apps/${APP_ID}/logs?lines=120"
docker ps --format 'table {{.Names}}\t{{.Image}}\t{{.Status}}'
curl -I https://authfly.ru/
curl https://authfly.ru/api/health
```

## Emergency fallback

If deploy fails with 500:

- check `docker logs --tail 120 dashboard`
- check dashboard registry login above
- confirm image exists in registry for pushed tag

If DNS not ready yet, keep `APP_HEALTHCHECK_URL` empty in repo variables.
