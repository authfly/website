# Deployment 101 for `starter`

This guide is written for a complete beginner version of future me.

Assumption:

- I have forgotten how GitHub Actions, Docker, Docker Registry, Dashboard API, and deployment work
- I want to deploy the project from scratch again
- I want one document that explains both the **what** and the **why**

This guide uses the real names from this project:

- GitHub repository: `phpfasty/starter`
- Registry host: `buildy-apps.registry.twcstorage.ru`
- Registry image path: `phpfasty/starter`
- Public domain: `authfly.ru`
- App name in dashboard: `starter`

---

## 1. Big picture

The deployment flow for this project is:

1. Code lives in GitHub
2. GitHub Actions builds a Docker image from `Dockerfile`
3. GitHub Actions pushes the image to the private registry
4. GitHub Actions renders `docker-compose.yml` with a concrete image tag
5. GitHub Actions updates the application in the dashboard through its API
6. The dashboard runs `docker compose up -d`
7. The site becomes available on `authfly.ru`

Important:

- The site code is **not** stored in `/opt/stacks/<APP_ID>`
- The site code is stored **inside the Docker image**
- `/opt/stacks/<APP_ID>` stores the compose files managed by the dashboard

---

## 2. What each file does

### `Dockerfile`

This file describes how the production image is built.

What it does in this project:

- installs PHP dependencies with Composer in a build stage
- builds a runtime image on `php:8.4-fpm-alpine`
- installs `nginx`, `APCu`, and PHP runtime config
- copies the application source into `/app`
- exposes internal port `80`
- defines an internal health check on `http://127.0.0.1/api/health`

In simple words:

- `Dockerfile` tells Docker how to turn the repository into a runnable container image

### `docker-compose.yml`

This is the **production compose template**.

It is used for:

- GitHub Actions render step
- dashboard deployment

It is **not** the old local development stack.

What it says:

- run one service called `web`
- use image `${APP_IMAGE}`
- expose internal port `80`
- pass environment variables into the container
- mount two Docker volumes for cache directories

In simple words:

- `docker-compose.yml` tells the dashboard how to run the already built image

### `docker-compose.local.yml`

This is the **old local development stack**.

It exists for:

- local development
- bind mounts
- manual Docker development with source files from the host

It is **not** used in the registry-based production flow.

### `.github/workflows/deploy.yml`

This is the GitHub Actions workflow.

It does the automated deployment when code is pushed to `main`.

In simple words:

- it is the robot that builds, pushes, updates, and deploys

### `docs/deployment/README.md`

This is the short operational deployment guide.

### This file: `.project/GUIDE-101.md`

This is the long explanation and recovery guide for a beginner.

---

## 3. The most important distinction

There are **three different places** involved in deployment.

### 1. GitHub repository

This is the source code.

Examples:

- `Dockerfile`
- `docker-compose.yml`
- `src/`
- `public/`
- `templates/`

### 2. Docker registry

This stores built container images.

Example image:

- `buildy-apps.registry.twcstorage.ru/phpfasty/starter:sha-b1db41199ff1`

### 3. Dashboard server

This stores runtime compose definitions and starts containers.

Example path:

- `/opt/stacks/starter-6e62b32b`

That path contains stack files, not the project source code.

---

## 4. How the site is physically stored

### On GitHub

The source code is in the Git repository.

### In Docker Registry

The production artifact is stored as an image.

### On the server after deploy

There are three relevant kinds of data:

1. Dashboard stack files
   Example:
   `/opt/stacks/starter-6e62b32b/docker-compose.yml`

2. Container filesystem
   The app code is inside the running container under:
   `/app`

3. Docker volumes
   Cache is stored in Docker volumes, not in the repository directory

Useful server commands:

```bash
docker exec starter-6e62b32b-web-1 sh -lc 'ls -la /app && ls -la /app/public'
docker volume inspect starter-6e62b32b_app-cache starter-6e62b32b_template-cache
docker inspect starter-6e62b32b-web-1 --format '{{json .GraphDriver.Data}}'
```

---

## 5. Naming used in this project

These are the real values used here.

### GitHub repository

- owner: `phpfasty`
- repo: `starter`

### Registry

- registry host: `buildy-apps.registry.twcstorage.ru`
- image repository path: `phpfasty/starter`

Final image references look like:

- `buildy-apps.registry.twcstorage.ru/phpfasty/starter:main`
- `buildy-apps.registry.twcstorage.ru/phpfasty/starter:sha-b1db41199ff1`

### Dashboard app

- app name: `starter`
- example app id: `starter-6e62b32b`

### Public website

- `authfly.ru`

### Health check

Once the domain is live, the external health check URL is:

- `https://authfly.ru/api/health`

Important:

- the dashboard URL is **not** the same as the app health check URL

---

## 6. Dashboard URL vs app URL vs health check

These three things are different.

### Dashboard URL

This is the administrative panel API.

Examples:

- `http://127.0.0.1:3000` when commands are run on the server itself
- `https://gui.example.com` when GitHub Actions must reach the dashboard from outside

### Public app domain

This is the real website:

- `https://authfly.ru`

### Health check URL

This is the health endpoint of the **site**, not the dashboard:

- `https://authfly.ru/api/health`

The health endpoint comes from the app itself.

---

## 7. Understanding the health check

This project has two different health checks.

### Internal container health check

Defined in `Dockerfile`.

It checks:

- `http://127.0.0.1/api/health` inside the container

This tells Docker whether the container itself looks healthy.

### External GitHub Actions smoke test

Defined in `.github/workflows/deploy.yml`.

It checks:

- `${APP_HEALTHCHECK_URL}`

This is meant to verify the **publicly reachable site** after deploy.

Important beginner rule:

- if the public domain is not ready yet, leave `APP_HEALTHCHECK_URL` empty
- otherwise the workflow may fail even though the deploy itself succeeded

---

## 8. Why `docker-compose.yml` looks so small

Production compose is intentionally small because:

- the image already contains the code
- the image already contains Nginx and PHP-FPM
- the dashboard only needs the run definition

Current production template:

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

What this means:

- use one container called `web`
- use the image tag that CI provides
- keep runtime cache in Docker volumes

---

## 9. Why `docker-compose.local.yml` still exists

Because local development and production are different concerns.

### `docker-compose.local.yml`

Used for:

- local development
- source code bind mounts
- editing files on the host and seeing changes immediately

### `docker-compose.yml`

Used for:

- production
- dashboard deployment
- registry-based image deploy

Beginner rule:

- local work -> `docker-compose.local.yml`
- production deploy -> `docker-compose.yml`

---

## 10. GitHub Actions workflow explained line by line

The workflow file is `.github/workflows/deploy.yml`.

### Trigger

```yaml
on:
  push:
    branches:
      - main
  workflow_dispatch:
```

Meaning:

- every push to `main` triggers deployment
- manual запуск from GitHub UI is also possible

### Node 24 flag

```yaml
FORCE_JAVASCRIPT_ACTIONS_TO_NODE24: "true"
```

Meaning:

- GitHub warned that Node.js 20 actions are deprecated
- this workflow opts into Node.js 24 now

### Variables and secrets

The workflow reads:

- repository variables via `vars.*`
- repository secrets via `secrets.*`

If they are empty, the workflow fails early.

### Build image tag

The workflow takes the first 12 characters of the commit SHA.

Example:

- full commit SHA: `b1db41199ff1d7c8...`
- image tag: `sha-b1db41199ff1`

### Build and push

The workflow pushes:

- `buildy-apps.registry.twcstorage.ru/phpfasty/starter:sha-<commit>`
- `buildy-apps.registry.twcstorage.ru/phpfasty/starter:main`

### Render compose

This step does not deploy.

It only replaces `${APP_IMAGE}` with the concrete tag and writes the final compose file.

### Update dashboard app

The workflow sends:

- `PUT /api/apps/${APP_ID}`

with new `compose_yaml`.

### Deploy app

The workflow sends:

- `POST /api/apps/${APP_ID}/deploy`

The dashboard then runs `docker compose up -d`.

---

## 11. Repository variables and secrets

### Repository variables

Configured in:

- GitHub repo -> Settings -> Secrets and variables -> Actions -> Variables

Current values:

- `REGISTRY_HOST=buildy-apps.registry.twcstorage.ru`
- `IMAGE_REPOSITORY=phpfasty/starter`
- `APP_NAME=starter`
- `APP_HEALTHCHECK_URL=https://authfly.ru/api/health`

Important:

- if the domain is not live yet, leave `APP_HEALTHCHECK_URL` empty

### Repository secrets

Configured in:

- GitHub repo -> Settings -> Secrets and variables -> Actions -> Secrets

Current set:

- `REGISTRY_USERNAME`
- `REGISTRY_PASSWORD`
- `DASHBOARD_URL`
- `DASHBOARD_USER`
- `DASHBOARD_PASS`
- `DASHBOARD_APP_ID`

Example:

- `DASHBOARD_APP_ID=starter-6e62b32b`

---

## 12. First deployment from zero

This section assumes:

- the project is not yet deployed
- Docker is available on the server
- the dashboard is already installed and working

### Step 1. Upload project files to the server

If Docker is only available on the server:

```bash
git archive --format=tar HEAD | ssh root@your-server "mkdir -p /opt/bootstrap/starter && tar -xf - -C /opt/bootstrap/starter"
```

Why:

- this sends tracked repository files only
- it avoids `.git/`
- it avoids random local junk

### Step 2. SSH to the server

```bash
ssh root@your-server
cd /opt/bootstrap/starter
```

### Step 3. Export variables

Example:

```bash
export DASHBOARD_URL="http://127.0.0.1:3000"
export DASHBOARD_USER="admin"
export DASHBOARD_PASS="admin@123"

export APP_NAME="starter"
export APP_IMAGE="buildy-apps.registry.twcstorage.ru/phpfasty/starter:sha-bootstrap"
export APP_PUBLIC_DOMAIN="authfly.ru"
export APP_PROXY_TARGET_PORT="80"
```

### Step 4. Render production compose

```bash
APP_IMAGE="${APP_IMAGE}" docker compose -f docker-compose.yml config > rendered-compose.yml
```

This does **not** start containers.

It only generates final YAML with the image tag inserted.

### Step 5. Create dashboard payload

```bash
python3 - <<'PY'
import json
import os
from pathlib import Path

payload = {
    "name": os.environ["APP_NAME"],
    "compose_yaml": Path("rendered-compose.yml").read_text(encoding="utf-8"),
}

Path("create-app.json").write_text(
    json.dumps(payload, ensure_ascii=True),
    encoding="utf-8",
)
PY
```

### Step 6. Create the app in dashboard

```bash
curl -u "${DASHBOARD_USER}:${DASHBOARD_PASS}" \
  -X POST "${DASHBOARD_URL}/api/apps" \
  -H 'Content-Type: application/json' \
  --data @create-app.json
```

The response returns an `id`.

Example:

```bash
export APP_ID="starter-6e62b32b"
```

### Step 7. Configure public domain and env

```bash
cat > app-config.json <<EOF
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
EOF
```

Send it:

```bash
curl -u "${DASHBOARD_USER}:${DASHBOARD_PASS}" \
  -X PUT "${DASHBOARD_URL}/api/apps/${APP_ID}/config" \
  -H 'Content-Type: application/json' \
  --data @app-config.json
```

### Step 8. Login to private registry inside dashboard container

This is very important.

The login must happen **inside the `dashboard` container**, because the dashboard runs the Docker commands.

```bash
docker exec -it dashboard sh
docker login buildy-apps.registry.twcstorage.ru
exit
```

What `exit` means here:

- it exits only from the shell **inside the container**
- it does **not** stop the container
- it does **not** close your original SSH session

Why beginners get confused:

- `docker exec -it dashboard sh` opens a shell inside the container
- inside that shell, `exit` means “leave this shell”
- after `exit`, you go back to the normal host shell

Optional verification:

```bash
docker exec dashboard sh -lc 'docker pull buildy-apps.registry.twcstorage.ru/phpfasty/starter:sha-bootstrap'
```

### Step 9. First deploy

```bash
curl -u "${DASHBOARD_USER}:${DASHBOARD_PASS}" \
  -X POST "${DASHBOARD_URL}/api/apps/${APP_ID}/deploy"
```

### Step 10. Enable TLS later

Once DNS and public HTTP routing work, enable HTTPS in the dashboard and verify:

```bash
curl -I https://authfly.ru/
curl https://authfly.ru/api/health
```

---

## 13. What to do for normal day-to-day deploys

Once everything is configured:

1. change code locally
2. commit
3. push to `main`
4. GitHub Actions builds and pushes the image
5. GitHub Actions updates dashboard compose
6. GitHub Actions triggers deploy

That is the normal path.

You do **not** need to manually upload the repository again for every release.

---

## 14. How to manually redeploy

If needed, redeploy the current image without a new push:

```bash
export DASHBOARD_URL="http://127.0.0.1:3000"
export DASHBOARD_USER="admin"
export DASHBOARD_PASS="admin@123"
export APP_ID="starter-6e62b32b"

curl -u "${DASHBOARD_USER}:${DASHBOARD_PASS}" \
  -X POST "${DASHBOARD_URL}/api/apps/${APP_ID}/deploy"
```

---

## 15. How to know the current image tag

The workflow creates tags like:

- `sha-b1db41199ff1`

It comes from the first 12 characters of `GITHUB_SHA`.

To see the current tag:

- open GitHub Actions logs
- inspect the dashboard app compose
- inspect the running container image

Useful commands:

```bash
docker inspect starter-6e62b32b-web-1 --format '{{.Config.Image}}'
sed -n '1,160p' /opt/stacks/starter-6e62b32b/docker-compose.yml
```

---

## 16. The difference between tag and digest

Two valid image references exist:

### By tag

```bash
buildy-apps.registry.twcstorage.ru/phpfasty/starter:sha-b1db41199ff1
```

### By digest

```bash
buildy-apps.registry.twcstorage.ru/phpfasty/starter@sha256:xxxxxxxx
```

Important:

- `:sha256:...` is wrong
- digest uses `@sha256:...`
- the workflow currently deploys by **tag**

---

## 17. Common mistakes I already made once

### Mistake 1. Running `docker compose -f docker-compose.yml config` in `/root`

Problem:

- `open /root/docker-compose.yml: no such file or directory`

Fix:

- `cd /opt/bootstrap/starter` first

### Mistake 2. Confusing dashboard URL with site URL

Wrong:

- using dashboard URL as app health check

Correct:

- dashboard URL is for dashboard API
- app health check URL is `https://authfly.ru/api/health`

### Mistake 3. Thinking `/opt/stacks/<APP_ID>` contains the site source code

Wrong assumption.

Reality:

- it stores compose files
- the code is in the image and inside the container

### Mistake 4. Logging into registry on the host only

Problem:

- deploy still fails with `401 Unauthorized`

Reason:

- the dashboard container itself needs credentials

Fix:

- run `docker login` inside `dashboard`

### Mistake 5. Thinking `exit` would leave SSH

When inside:

```bash
docker exec -it dashboard sh
```

`exit` leaves only the container shell and returns to the host shell.

### Mistake 6. Forgetting that `APP_HEALTHCHECK_URL` is optional

If the public domain is not live yet:

- leave it empty

Otherwise:

- the workflow can fail at the smoke test stage even if deploy itself worked

---

## 18. How to verify everything

### Check dashboard app state

```bash
curl -u "${DASHBOARD_USER}:${DASHBOARD_PASS}" \
  "${DASHBOARD_URL}/api/apps/${APP_ID}"
```

### Check dashboard app logs

```bash
curl -u "${DASHBOARD_USER}:${DASHBOARD_PASS}" \
  "${DASHBOARD_URL}/api/apps/${APP_ID}/logs?lines=120"
```

### Check running containers

```bash
docker ps --format 'table {{.Names}}\t{{.Image}}\t{{.Status}}'
```

### Check site from the server

```bash
curl -I https://authfly.ru/
curl https://authfly.ru/api/health
```

### Check app code in the running container

```bash
docker exec starter-6e62b32b-web-1 sh -lc 'ls -la /app && ls -la /app/public'
```

---

## 19. What to do if deploy fails again

### If GitHub Actions fails before build

Check:

- repository variables
- repository secrets

### If push to registry fails

Check:

- `REGISTRY_USERNAME`
- `REGISTRY_PASSWORD`
- registry host

### If dashboard deploy returns `500`

Check:

```bash
docker logs --tail 120 dashboard
```

Then usually one of these is true:

- registry auth is missing inside `dashboard`
- image tag does not exist
- compose runtime startup failed

### If public HTTPS fails

Check:

- dashboard TLS settings
- DNS records for `authfly.ru`
- Nginx routing created by dashboard
- certificate issuance logs

---

## 20. Minimal emergency checklist

If I forget everything, do this:

1. confirm app id
   `starter-6e62b32b`
2. confirm image exists in registry
3. confirm dashboard container is running
4. confirm `docker login` exists inside `dashboard`
5. confirm GitHub secrets and variables still exist
6. push to `main`
7. if deploy fails, read:
   `docker logs --tail 120 dashboard`
8. if deploy succeeds, verify:
   `curl https://authfly.ru/api/health`

---

## 21. Final mental model

Remember this sentence:

> GitHub stores source code, the registry stores release images, the dashboard stores compose runtime config, and Docker runs the actual site.

If I remember only that sentence, the rest can be reconstructed.
