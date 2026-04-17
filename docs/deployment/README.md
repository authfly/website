# Registry Deployment via Dashboard

This guide documents both supported deployment flows for this repository:

1. create the app once in the dashboard
2. choose either GitHub Actions or the local `paas` Go runner
3. build, push, and deploy using immutable image tags

It does **not** use repository import. The dashboard stores the application as a normal compose-managed app, while either GitHub Actions or the local Go runner updates its `compose_yaml` to a new immutable image tag like `sha-<commit>`.

## Current recommended flows

The registry-based path is still documented below, but it is now treated as a secondary option. The current preferred `paas` flows are direct **local PC -> server** deploys without a registry hop:

- `.paas/extensions/bootstrap-direct.yml`
  First deployment for a brand new app. It:
  1. uploads the repository snapshot over SSH
  2. builds the image directly on the server
  3. creates the dashboard app
  4. configures routing
  5. triggers the first deploy

- `.paas/extensions/deploy-direct.yml`
  Regular update flow for an already existing app. It:
  1. uploads the repository snapshot over SSH
  2. builds the image directly on the server
  3. renders `docker-compose.yml` on the server with a local image tag
  4. updates the existing dashboard app
  5. triggers redeploy

- `.paas/extensions/deploy.yml`
  Older experiment that still pushes through the registry. Keep it only if you specifically want immutable images stored outside the server. For normal direct deployments, prefer the two direct variants above.

### Which file to run

Use:

```bash
./paas.exe run bootstrap-direct
```

when the dashboard app does not exist yet, and:

```bash
./paas.exe run deploy-direct
```

when the app already exists and only needs an update.

### Shared assumptions for both direct flows

- the server is reachable over SSH
- the local shell can run `ssh` for the upload step
- the server has Docker, `docker compose`, `curl`, and `jq`
- the dashboard API is reachable from the server itself
- the dashboard and the built image use the same Docker host, so a registry push is not required

## Step 1. Create the app once in the dashboard

Set these values first:

```bash
export DASHBOARD_URL="http://127.0.0.1:3000"
export DASHBOARD_USER="admin"
export DASHBOARD_PASS="admin@123"

export APP_NAME="starter"
export APP_IMAGE="buildy-apps.registry.twcstorage.ru/phpfasty/starter:sha-bootstrap"
export APP_PUBLIC_DOMAIN="starter.example.com"
export APP_PROXY_TARGET_PORT="80"
```

If Docker is only available on the server, upload the tracked project files to a temporary bootstrap directory first:

```bash
git archive --format=tar HEAD | ssh root@your-server "mkdir -p /opt/bootstrap/starter && tar -xf - -C /opt/bootstrap/starter"
```

Then connect to the server and switch to that directory:

```bash
ssh root@your-server
cd /opt/bootstrap/starter
```

Render the production compose from the root template in that repository directory:

```bash
APP_IMAGE="${APP_IMAGE}" docker compose -f docker-compose.yml config > rendered-compose.yml
```

Create the application record in the dashboard:

```bash
python3 - <<'PY'
import json
from pathlib import Path

payload = {
    "name": "starter",
    "compose_yaml": Path("rendered-compose.yml").read_text(encoding="utf-8"),
}

Path("create-app.json").write_text(json.dumps(payload, ensure_ascii=True), encoding="utf-8")
PY

curl -u "${DASHBOARD_USER}:${DASHBOARD_PASS}" \
  -X POST "${DASHBOARD_URL}/api/apps" \
  -H 'Content-Type: application/json' \
  --data @create-app.json
```

Copy the returned `id`, then export it:

```bash
export APP_ID="starter-a1b2c3d4"
```

Configure public access and managed environment:

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

curl -u "${DASHBOARD_USER}:${DASHBOARD_PASS}" \
  -X PUT "${DASHBOARD_URL}/api/apps/${APP_ID}/config" \
  -H 'Content-Type: application/json' \
  --data @app-config.json
```

If the registry is private, the dashboard runtime must also be able to pull the image. The login must be available **inside** the `dashboard` container, not only on the host:

```bash
docker exec -it dashboard sh
docker login buildy-apps.registry.twcstorage.ru
exit
```

Optional verification from inside the dashboard container:

```bash
docker exec dashboard sh -lc 'docker pull buildy-apps.registry.twcstorage.ru/phpfasty/starter:sha-bootstrap'
```

Run the first deploy:

```bash
curl -u "${DASHBOARD_USER}:${DASHBOARD_PASS}" \
  -X POST "${DASHBOARD_URL}/api/apps/${APP_ID}/deploy"
```

Expected result:

- the app appears in the dashboard GUI
- the app keeps its own `id`
- future releases only update `compose_yaml` and redeploy this same app

## Step 2A. Configure GitHub repository variables and secrets

Create these **Repository variables**:

- `REGISTRY_HOST` = `buildy-apps.registry.twcstorage.ru`
- `IMAGE_REPOSITORY` = `phpfasty/starter`
- `APP_NAME` = `starter`
- `APP_HEALTHCHECK_URL` = `http://starter.example.com/api/health`

Create these **Repository secrets**:

- `REGISTRY_USERNAME`
- `REGISTRY_PASSWORD`
- `DASHBOARD_URL`
- `DASHBOARD_USER`
- `DASHBOARD_PASS`
- `DASHBOARD_APP_ID`

Notes:

- `DASHBOARD_APP_ID` is the application `id` returned in step 1
- if you already created the app manually, fill `DASHBOARD_APP_ID` now and continue with step 3
- `APP_HEALTHCHECK_URL` is optional; leave it empty until the public domain is live
- the dashboard host itself must already be able to pull from `buildy-apps.registry.twcstorage.ru`
- for a private registry, run `docker login` inside the `dashboard` container once before the first automated deploy

## Step 3A. Push to `main`

The repository workflow is `.github/workflows/deploy.yml`.

On every push to `main` it will:

1. build the production image from `Dockerfile`
2. push two tags to the registry:
   `sha-<commit>`
   `main`
3. render `docker-compose.yml` with `APP_IMAGE=<registry>/<repo>:sha-<commit>`
4. send the rendered YAML to `PUT /api/apps/{id}`
5. call `POST /api/apps/{id}/deploy`
6. optionally run a smoke test against `APP_HEALTHCHECK_URL`

Manual trigger is also available via `workflow_dispatch`.

The workflow also opts GitHub JavaScript actions into Node.js 24 now, which avoids the current Node.js 20 deprecation warning on GitHub-hosted runners.

## Step 2B. Deploy with the local Go runner (`paas.exe`)

This repository also supports a manual deploy flow with the local Go runner stored in `.gorunner/`.

The current project override at `.paas/extensions/deploy.yml` does **not** use the built-in local-image deploy. It performs a remote build over SSH:

1. get the current git SHA locally
2. upload the tracked repository snapshot with `git archive | ssh`
3. build the image on the server
4. push the image to the private registry
5. render `docker-compose.yml` on the server with `APP_IMAGE=<registry>/<repo>:sha-<commit>`
6. update the dashboard app via API
7. trigger deploy
8. optionally run a local smoke test

### One-time setup

Build the runner:

```bash
cd .gorunner
go build -o ../paas.exe ./cmd/paas
cd ..
```

Create project config and project deploy override:

```bash
./paas.exe init
```

Create user SSH config in `~/.config/paas/servers.yml`:

```yaml
servers:
  production:
    host: 5.42.114.251
    port: 22
    user: root
    key: C:/Users/alexe/.ssh/id_ed25519
    dashboard_user: admin
    dashboard_pass: "admin@123"
    host_key_check: tofu
```

Project defaults in `.paas/config.yml`:

```yaml
server: production

defaults:
  INPUT_APP_NAME: starter
  INPUT_APP_ID: starter-6e62b32b
  INPUT_REGISTRY_HOST: buildy-apps.registry.twcstorage.ru
  INPUT_IMAGE_REPOSITORY: phpfasty/starter
  INPUT_DASHBOARD_URL: http://127.0.0.1:3000
  INPUT_DASHBOARD_USER: admin
  INPUT_DASHBOARD_PASS: "admin@123"
  INPUT_HEALTHCHECK_URL: ""

extensions_dir: .paas/extensions
```

Export only the registry credentials in the shell session:

```bash
export INPUT_REGISTRY_USERNAME="buildy-apps"
export INPUT_REGISTRY_PASSWORD="..."
```

### Windows + Git Bash command

For the current runner implementation on Windows, use this exact wrapper command:

```bash
env -i \
  HOME="$HOME" \
  USERPROFILE="$USERPROFILE" \
  HOMEDRIVE="$HOMEDRIVE" \
  HOMEPATH="$HOMEPATH" \
  PATH="$PATH" \
  TERM="${TERM:-xterm-256color}" \
  LANG="${LANG:-en_EN.UTF-8}" \
  SSH_AUTH_SOCK="$SSH_AUTH_SOCK" \
  SSH_AGENT_PID="$SSH_AGENT_PID" \
  INPUT_REGISTRY_USERNAME="$INPUT_REGISTRY_USERNAME" \
  INPUT_REGISTRY_PASSWORD="$INPUT_REGISTRY_PASSWORD" \
  ./paas.exe run deploy
```

This wrapper is important because the current runner forwards process environment variables into remote Bash steps. A normal Windows shell environment contains names like `ProgramFiles(x86)` and a Windows `PATH`, both of which can break remote `export ...; bash -lc ...`.

### Required SSH preparation

The current project override uploads source with the local system `ssh` command:

```bash
git archive --format=tar HEAD | ssh root@server "..."
```

That means you must prepare SSH outside the runner:

```bash
eval "$(ssh-agent -s)"
ssh-add ~/.ssh/id_ed25519
ssh root@5.42.114.251
```

If the final `ssh` command works without asking for the server password, the upload step will work too.

### Validation and execution

Validate the extension first:

```bash
./paas.exe validate deploy
```

List the detected extension source:

```bash
./paas.exe list
./paas.exe servers
```

Then run the wrapped deploy command shown above.

## Problems discovered during the real runner deployment

These issues were hit during the first successful end-to-end deployment and are now part of the recommended operating procedure.

### 1. SSH key login is required

Password-only SSH access is not enough for this runner. The successful setup required:

- generating `~/.ssh/id_ed25519`
- adding the public key to `root@server:~/.ssh/authorized_keys`
- loading the private key into `ssh-agent`

### 2. The upload step uses system `ssh`, not the runner SSH client

The project override uses `git archive | ssh ...` for source upload, so even though remote steps use the runner's Go SSH client, the upload step still depends on:

- the local `ssh` binary
- local key availability through `ssh-agent` or explicit SSH configuration

### 3. Windows environment variables can break remote Bash

Without the `env -i` wrapper, remote steps failed before execution because Windows variables such as `ProgramFiles(x86)` are invalid Bash export names. A Windows `PATH` can also hide Linux `bash` on the remote host.

### 4. CLI flags after the extension name are not safe

For this MVP, prefer environment variables like `INPUT_REGISTRY_USERNAME` and `INPUT_REGISTRY_PASSWORD` over:

```bash
./paas.exe run deploy --input ...
```

The safer forms are:

```bash
./paas.exe run --input key=value deploy
```

or exported `INPUT_*` variables.

### 5. `CRLF` shell scripts break Linux entrypoints

The first remote-built image deployed successfully but the container immediately restarted with:

```text
exec /usr/local/bin/docker-entrypoint: no such file or directory
```

The root cause was a Windows `CRLF` shebang in `docker/production/entrypoint.sh` (`#!/bin/sh^M`).

The repository now protects against this in two layers:

- `.gitattributes` forces LF for `*.sh` and `Dockerfile`
- `Dockerfile` strips carriage returns during build:

```dockerfile
COPY docker/production/entrypoint.sh /usr/local/bin/docker-entrypoint
RUN sed -i 's/\r$//' /usr/local/bin/docker-entrypoint \
    && chmod +x /usr/local/bin/docker-entrypoint
```

## Production compose template

The root `docker-compose.yml` is now the production template used by CI and dashboard deployment.

It is intentionally different from `docker-compose.local.yml`:

- `docker-compose.yml` uses a registry image and named volumes for runtime cache
- `docker-compose.local.yml` keeps the old bind-mounted development stack

## Quick verification

After a deployment, these checks are useful:

```bash
curl -u "${DASHBOARD_USER}:${DASHBOARD_PASS}" \
  "${DASHBOARD_URL}/api/apps/${APP_ID}"

curl -u "${DASHBOARD_USER}:${DASHBOARD_PASS}" \
  "${DASHBOARD_URL}/api/apps/${APP_ID}/logs?lines=120"

curl -I "http://starter.example.com/"
curl "http://starter.example.com/api/health"
```
