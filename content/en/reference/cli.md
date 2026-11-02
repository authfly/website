---
title: CLI
description: The authfly commands and init flags in release 0.4.2.
---
# CLI

`npx authfly <command> [flags]` or the `authfly` binary. Common flag: `-f authfly.yaml` (default `./authfly.yaml`).

## Commands

| Command | What it does |
|---|---|
| `init` | create `authfly.yaml` (interactive wizard; `-y` with flags for non-interactive) |
| `deploy` | deploy or converge to `authfly.yaml`; idempotent; run as root |
| `status` | what runs: release, services, ports |
| `doctor` | audit the install against `authfly.yaml` and its policy pack |
| `upgrade [--to <version>]` | snapshot, upgrade to the newest (or given) release, roll back automatically on failure |
| `rollback` | restore the snapshot taken by the last upgrade |
| `backup` | encrypted backup (AES-256-GCM with the `backup_key`) |
| `restore <file.afbk>` | restore an encrypted backup |
| `secrets` | list secrets and their permissions; values are never printed |
| `down [--purge]` | stop and remove the containers; `--purge` also deletes the data |
| `serve` | run the Authfly server (what the containers run) |
| `invite create\|list\|revoke` | sign-up invites (see below) |
| `version` | print the version |

## init flags

| Flag | Sets |
|---|---|
| `-y` | non-interactive |
| `-config <file>`, `-force` | output file, overwrite |
| `-name`, `-target`, `-provider` | `name`, `target`, `provider` |
| `-runtime`, `-adapter` | `app.runtime`, `app.adapter` |
| `-client "id[:Name]=uri[,uri];…"` | `app.clients` |
| `-domain`, `-public-url` | `domain`, `public_url` |
| `-tls`, `-acme-email`, `-cert`, `-key`, `-http-port` | `http.*` |
| `-mail-hostname`, `-smtp-port`, `-submissions-port`, `-imaps-port`, `-encryption-at-rest` | `mail.*` |
| `-signin magic,password,username`, `-require-verify` | `sign_in.*` |
| `-policy`, `-operator`, `-contact`, `-host-country`, `-us-states` | `policy.*` |
| `-management-port`, `-console-port` | `ports.*` |
| `-channel`, `-allow-unsigned-channel` | `channel.*` |
| `-incus-network`, `-incus-subnet`, `-firewall`, `-incus-no-limits` | `incus.*` |

Go's flag parser accepts both `-flag` and `--flag`.

## invite

Run where the Authfly store is, usually inside the container:

```sh
docker compose -p authfly-<name> -f .authfly/compose.yml exec -T authfly app invite create -note "who it is for" -ttl 72h
docker compose -p authfly-<name> -f .authfly/compose.yml exec -T authfly app invite list
docker compose -p authfly-<name> -f .authfly/compose.yml exec -T authfly app invite revoke <id>
```

| Flag | Meaning |
|---|---|
| `-store <file>` | invite store (default from `AUTHFLY_STORE`) |
| `-kind user\|org` | invite kind |
| `-note <text>` | your note |
| `-ttl <duration>` | lifetime, default `72h` |

The token is printed once; keep it out of shell history and logs.
