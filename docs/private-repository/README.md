# Private GitHub Repo Guide

## Configurator
Fill these variables first, then run the commands below with the same values.

```bash
# Dashboard access
export DASHBOARD_URL="http://127.0.0.1:3000"
export DASHBOARD_USER="admin"
export DASHBOARD_PASS="admin@123"

# GitHub repo access
export GITHUB_USERNAME="YOUR_GITHUB_USERNAME"
export GITHUB_TOKEN="github_pat_xxxxx"
export GITHUB_OWNER="owner"
export GITHUB_REPO="private-repo"
export GITHUB_BRANCH="main"

# App settings in dashboard
export APP_NAME="private-app"
export COMPOSE_PATH="docker-compose.local.yml"
export APP_PORT="80"

# Optional: fill after app import when dashboard returns app id
export APP_ID=""
export APP_DIR=""
```

## Goal
This guide explains how to import a private GitHub repository from the terminal so that:

- the application is created through the dashboard API
- a record appears in the dashboard database
- the application appears in the GUI
- the repository is cloned from GitHub using a temporary token
- the token can then be removed from `.git/config`

This guide does not cover automatic repo updates in the GUI. Updates are terminal-only in this MVP flow.

## Before You Start
Make sure all of the following are already working:

- the dashboard container is running
- you can open the dashboard in the browser
- dashboard API auth works with `admin` and `admin@123` or your own credentials
- the private GitHub repository exists
- you created a GitHub token with read access to the target repository
- the repository contains either `docker-compose.local.yml` or a valid `compose.yml`

## Step 1. Check that the dashboard API is reachable

```bash
curl -u "${DASHBOARD_USER}:${DASHBOARD_PASS}" \
  "${DASHBOARD_URL}/api/apps"
```

Expected result:

- you get JSON back
- no `401 Unauthorized`

## Step 2. Import the private repository through the dashboard API

This request creates the application in the dashboard exactly like the GUI import flow does.

```bash
curl -u "${DASHBOARD_USER}:${DASHBOARD_PASS}" \
  -X POST "${DASHBOARD_URL}/api/apps/import" \
  -H 'Content-Type: application/json' \
  -d "{
    \"name\": \"${APP_NAME}\",
    \"repo_url\": \"https://${GITHUB_USERNAME}:${GITHUB_TOKEN}@github.com/${GITHUB_OWNER}/${GITHUB_REPO}.git\",
    \"branch\": \"${GITHUB_BRANCH}\",
    \"compose_path\": \"${COMPOSE_PATH}\",
    \"app_port\": ${APP_PORT},
    \"auto_deploy\": true
  }"
```

What this does:

- clones the private repository
- resolves the compose file
- creates the app record in BoltDB
- deploys the app immediately because `auto_deploy` is `true`

## Step 3. Save the returned app id

The API response contains the created application object. Copy its `id`.

Example response shape:

```json
{
  "id": "private-app-a1b2c3d4",
  "name": "private-app",
  "repo_url": "https://YOUR_GITHUB_USERNAME:github_pat_xxxxx@github.com/owner/private-repo.git",
  "repo_branch": "main",
  "compose_path": "docker-compose.local.yml"
}
```

Then export it:

```bash
export APP_ID="private-app-a1b2c3d4"
export APP_DIR="/opt/stacks/${APP_ID}"
```

## Step 4. Verify that the app appeared in the GUI

Open the dashboard in the browser and check the Apps page.

You can also confirm by API:

```bash
curl -u "${DASHBOARD_USER}:${DASHBOARD_PASS}" \
  "${DASHBOARD_URL}/api/apps"
```

Expected result:

- your new app is listed
- the app has an `id`
- the app status eventually becomes running

## Step 5. Verify that the repo was cloned to disk

```bash
ls -la "${APP_DIR}"
ls -la "${APP_DIR}/.git"
git -C "${APP_DIR}" remote -v
```

Expected result:

- the directory exists
- `.git` exists
- `origin` points to your GitHub repository

At this point, the token is usually still embedded in the remote URL.

## Step 6. Remove the token from `.git/config`

This is strongly recommended after the first successful import.

```bash
git -C "${APP_DIR}" remote set-url origin \
  "https://github.com/${GITHUB_OWNER}/${GITHUB_REPO}.git"
```

## Quick Cheat Sheet

```bash
# 1. Fill variables
export DASHBOARD_URL="http://127.0.0.1:3000"
export DASHBOARD_USER="admin"
export DASHBOARD_PASS="admin@123"
export GITHUB_USERNAME="YOUR_GITHUB_USERNAME"
export GITHUB_TOKEN="github_pat_xxxxx"
export GITHUB_OWNER="owner"
export GITHUB_REPO="private-repo"
export GITHUB_BRANCH="main"
export APP_NAME="private-app"
export COMPOSE_PATH="docker-compose.local.yml"
export APP_PORT="80"

# 2. Import private repo through dashboard API
curl -u "${DASHBOARD_USER}:${DASHBOARD_PASS}" \
  -X POST "${DASHBOARD_URL}/api/apps/import" \
  -H 'Content-Type: application/json' \
  -d "{
    \"name\": \"${APP_NAME}\",
    \"repo_url\": \"https://${GITHUB_USERNAME}:${GITHUB_TOKEN}@github.com/${GITHUB_OWNER}/${GITHUB_REPO}.git\",
    \"branch\": \"${GITHUB_BRANCH}\",
    \"compose_path\": \"${COMPOSE_PATH}\",
    \"app_port\": ${APP_PORT},
    \"auto_deploy\": true
  }"

# 3. Copy id from JSON response, then set:
export APP_ID="private-app-a1b2c3d4"
export APP_DIR="/opt/stacks/${APP_ID}"

# 4. Verify app exists
curl -u "${DASHBOARD_USER}:${DASHBOARD_PASS}" "${DASHBOARD_URL}/api/apps"
ls -la "${APP_DIR}"
git -C "${APP_DIR}" remote -v

# 5. Remove token from .git/config
git -C "${APP_DIR}" remote set-url origin \
  "https://github.com/${GITHUB_OWNER}/${GITHUB_REPO}.git"

# 6. Check remote again
git -C "${APP_DIR}" remote -v

# 7. Later: temporarily add a new token for manual pull
git -C "${APP_DIR}" remote set-url origin \
  "https://${GITHUB_USERNAME}:${GITHUB_TOKEN}@github.com/${GITHUB_OWNER}/${GITHUB_REPO}.git"

git -C "${APP_DIR}" pull origin "${GITHUB_BRANCH}"

# 8. Remove token again after pull
git -C "${APP_DIR}" remote set-url origin \
  "https://github.com/${GITHUB_OWNER}/${GITHUB_REPO}.git"
```

Check again:

```bash
git -C "${APP_DIR}" remote -v
```

Expected result:

- `origin` no longer contains the token

## Step 7. Verify the deployed app

Check logs:

```bash
docker compose -p "${APP_ID}" -f "${APP_DIR}/${COMPOSE_PATH}" logs --tail=80
```

If the dashboard generated or uses an override file, this is also useful:

```bash
ls -la "${APP_DIR}"
```

If the app is attached to a domain, test the domain:

```bash
curl -I "http://your-domain.example.com"
curl -k -I "https://your-domain.example.com"
```

## Terminal-Only Token Refresh for Future Manual Pulls

This MVP does not add repo update support to the GUI. If later you want to run `git pull` manually from the terminal, use this temporary token workflow.

### 1. Add a new token to `origin`

```bash
git -C "${APP_DIR}" remote set-url origin \
  "https://${GITHUB_USERNAME}:${GITHUB_TOKEN}@github.com/${GITHUB_OWNER}/${GITHUB_REPO}.git"
```

### 2. Pull changes manually

```bash
git -C "${APP_DIR}" pull origin "${GITHUB_BRANCH}"
```

### 3. Remove the token from `.git/config` again

```bash
git -C "${APP_DIR}" remote set-url origin \
  "https://github.com/${GITHUB_OWNER}/${GITHUB_REPO}.git"
```

### 4. Confirm cleanup

```bash
git -C "${APP_DIR}" remote -v
```

## Notes for Beginners

- `repo_url` must be a full HTTPS GitHub URL.
- `app_port` must be the internal container port used by the app, not the public host port.
- if your repo uses `nginx:alpine` with default config, that port is usually `80`
- if you remove the token from `.git/config`, the deployed app keeps working normally
- removing the token only affects future Git operations
- if you revoke the token on GitHub, existing files on disk stay in place
- if you want to pull again later, add a fresh token to `origin`, pull, then remove it again

## Common Mistakes

### Wrong `repo_url`
Bad:

```bash
https://github.com/owner/private-repo
```

Preferred:

```bash
https://github.com/owner/private-repo.git
```

### Wrong `app_port`
If the container listens on `80`, do not use `8888` unless the app inside the container really listens on `8888`.

### Forgetting to save the returned `id`
You need `APP_ID` to inspect files, logs, and later run manual Git commands.

### Leaving the token in `.git/config`
The app will work, but the token stays on disk. Remove it after clone unless you intentionally need it for immediate manual pulls.

## Minimal Example with Real Flow

```bash
export DASHBOARD_URL="http://127.0.0.1:3000"
export DASHBOARD_USER="admin"
export DASHBOARD_PASS="admin@123"
export GITHUB_USERNAME="myuser"
export GITHUB_TOKEN="github_pat_xxxxx"
export GITHUB_OWNER="myorg"
export GITHUB_REPO="private-site"
export GITHUB_BRANCH="main"
export APP_NAME="private-site"
export COMPOSE_PATH="docker-compose.local.yml"
export APP_PORT="80"

curl -u "${DASHBOARD_USER}:${DASHBOARD_PASS}" \
  -X POST "${DASHBOARD_URL}/api/apps/import" \
  -H 'Content-Type: application/json' \
  -d "{
    \"name\": \"${APP_NAME}\",
    \"repo_url\": \"https://${GITHUB_USERNAME}:${GITHUB_TOKEN}@github.com/${GITHUB_OWNER}/${GITHUB_REPO}.git\",
    \"branch\": \"${GITHUB_BRANCH}\",
    \"compose_path\": \"${COMPOSE_PATH}\",
    \"app_port\": ${APP_PORT},
    \"auto_deploy\": true
  }"

export APP_ID="private-site-a1b2c3d4"
export APP_DIR="/opt/stacks/${APP_ID}"

git -C "${APP_DIR}" remote set-url origin \
  "https://github.com/${GITHUB_OWNER}/${GITHUB_REPO}.git"
```
