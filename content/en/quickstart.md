---
title: Quickstart
description: Install Authfly and Stalwart locally with Docker, create the first invite and sign in, in about ten minutes.
---
# Quickstart

This runs Authfly and Stalwart on your machine with Docker and plain HTTP on port 4000. For a public server read [Install Authfly + Stalwart](/en/guides/install-authfly-stalwart).

## Requirements

- Linux or macOS with Docker Engine and the Compose plugin (`docker compose version`).
- Root (or `sudo`) for `deploy`: it writes the secrets owned by the container user.
- Node.js 18+ for `npx`, or the release binary (below).

## 1. Get the CLI

With npm:

```sh
npx authfly version
```

Or the release binary, checked against the published checksums:

```sh
curl -fsSLO https://github.com/authfly/cli/releases/download/v0.4.2/authfly-linux-amd64
curl -fsSLO https://github.com/authfly/cli/releases/download/v0.4.2/SHA256SUMS
sha256sum --ignore-missing -c SHA256SUMS
install -m 0755 authfly-linux-amd64 ./authfly
```

> **Note:** at the time of writing npm has `authfly` 0.4.1 as `latest`; release 0.4.2 is on GitHub only. The examples below use `./authfly`; `npx authfly` takes the same commands.

## 2. Create authfly.yaml

`init` without flags runs an interactive wizard. Non-interactive:

```sh
./authfly init -y --name demo --target docker \
  --runtime react --adapter node \
  --domain example.test --public-url http://localhost:4000 --tls off \
  --signin magic,password \
  --client "my-app:My app=http://localhost:5173/callback" \
  --policy gdpr --operator "Example Ltd" --contact privacy@example.test
```

`--client` registers an OIDC client for your app (`id[:Name]=redirect[,redirect]`). All keys are listed in [authfly.yaml](/en/reference/authfly-yaml).

## 3. Deploy

```sh
sudo ./authfly deploy
sudo ./authfly status
sudo ./authfly doctor
```

`deploy` downloads the signed release channel, verifies it, generates secrets into `.authfly/` (mode 0600), writes `.authfly/compose.yml`, starts Stalwart, hardens it (service account, admin UI off, OIDC locked) and starts Authfly. It is idempotent: run it again after editing `authfly.yaml`.

> **Known issue (0.4.2):** the 0.4.2 release channel pins Stalwart v0.16.27, while the 0.4.2 server expects Stalwart 1.0. If the `stalwart-init` step fails with `invalidPatch enabledPermissions`, see the workaround in [Releases](/en/releases#known-issues).

## 4. Create an invite and sign up

Public sign-up is off by default; sign-up needs an invite:

```sh
docker compose -p authfly-demo -f .authfly/compose.yml exec -T authfly app invite create -note "first user"
```

The token is printed once. Open `http://localhost:4000/auth/sign-up`, paste the invite and create the account. Because the mail domain is local, the verification mail is delivered into the new mailbox in Stalwart; with real DNS it reaches any inbox.

## 5. Sign in

Open `http://localhost:4000/auth/sign-in`, then the account area at `http://localhost:4000/account`.

## Next

- Connect an app with SSO: [React + Docker on Authfly](/en/guides/react-docker).
- Remove everything: `sudo ./authfly down --purge` (deletes the data too).
