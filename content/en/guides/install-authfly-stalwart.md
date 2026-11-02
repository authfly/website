---
title: Install Authfly + Stalwart
nav: Install Authfly + Stalwart
description: Production install on your own server: topology, DNS for mail and TLS, authfly.yaml, deploy, lock-down checks, backups and day-two operations.
---
# Install Authfly + Stalwart

This guide takes a fresh Linux server to a public Authfly with Stalwart behind it. It combines the Authfly source and docs (release 0.4.2) with a field report from an operator who runs Authfly on Incus; field-report items are marked as such, and anything not confirmed in code is marked UNVERIFIED.

Placeholders: `auth.example.com` (Authfly, the OIDC issuer), `example.com` (mail domain), `mail.example.com` (mail host), `app.example.com` (your app), `203.0.113.10` (server address).

## 1. Choose a topology

| Option | TLS | When |
|---|---|---|
| A. `target: docker`, `tls: acme` | Authfly gets Let's Encrypt certificates itself (HTTP-01 on port 80) | one server, nothing else on 80/443 |
| B. `target: docker` or `incus`, `tls: off`, `public_url: https://…` | your reverse proxy (Caddy, nginx) terminates TLS and forwards to Authfly :4000 | you already run a proxy for other sites |
| C. `target: incus` | as A or B | you prefer system containers; `authfly deploy` creates the network `authfly0` and the instances |

In every option only Authfly (and the mail ports, if you send mail) are public. Stalwart's HTTP listener (management, JMAP, `/api/*`) stays on `127.0.0.1` and the private network.

## 2. Prepare the server

- A supported Linux with Docker Engine + Compose plugin (option A/B) or Incus (option C).
- Open 80 and 443; open 25, 465 and 993 only if you want mail in and out.
- Run `deploy` as root: it writes secrets owned by the container user (uid 10001); without root that step fails.

## 3. DNS

Records for sign-in:

| Name | Type | Value |
|---|---|---|
| `auth.example.com` | A / AAAA | `203.0.113.10` |
| `app.example.com` | A / AAAA | your app host |

Records for mail (from the Stalwart DNS documentation):

| Name | Type | Value |
|---|---|---|
| `example.com` | MX | `10 mail.example.com` |
| `mail.example.com` | A / AAAA | `203.0.113.10` |
| `example.com` | TXT (SPF) | `v=spf1 mx -all` |
| `mail.example.com` | TXT (SPF) | `v=spf1 a -all` |
| `<selector>._domainkey.example.com` | TXT (DKIM) | the public keys Stalwart generates (an Ed25519 and an RSA selector) |
| `_dmarc.example.com` | TXT | start with `v=DMARC1; p=none; rua=mailto:dmarc@example.com`, move to `quarantine` or `reject` once reports are clean |
| `203.0.113.10` | PTR | `mail.example.com`, set at your hosting provider |

> **Unverified:** Stalwart prints the exact zone (DKIM selectors included) in its web admin or with `stalwart-cli`. Authfly disables the web admin, so read it with the CLI from inside the private network. The CLI command for this was not tested with Authfly 0.4.2. PTR is general mail practice and is not set by Authfly.

> **Note:** sign-in does not need external mail if your users' addresses are on the Stalwart domain: mail to local mailboxes is delivered internally. Delivery to other domains needs the records above; it was not tested for this guide.

## 4. Write authfly.yaml

Option A, Authfly terminates TLS:

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

Option B, your proxy terminates TLS: use `--tls off --public-url https://auth.example.com`. `http.tls: off` with an `https://` public URL is allowed exactly for this case; cookies are then `Secure`. Authfly listens on port 4000.

Key choices (all keys: [authfly.yaml](/en/reference/authfly-yaml)):

- `public_url` is the OIDC issuer. Pick it once; tokens and client registrations depend on it.
- `app.clients` lists your apps and their exact redirect URIs. Authfly also registers its own client `authfly` with redirect `<public_url>/auth/callback`; you do not create OIDC clients by hand in Stalwart.
- `sign_in.social`, `two_factor` and `webauthn` must stay `false` in 0.4.2.

## 5. Deploy

```sh
sudo authfly deploy
sudo authfly status
sudo authfly doctor
```

What `deploy` does: verifies the signed channel, generates secrets into `.authfly/` (0600), renders the policy-pack documents, starts Stalwart, runs `stalwart-init` (creates the domain, a least-privilege service account `authfly-svc@<domain>`, disables the web admin and account apps, locks OIDC client registration and registers the clients), then starts Authfly.

> **Known issue (0.4.2):** the channel pins Stalwart v0.16.27 but the server needs Stalwart 1.0; see [Releases](/en/releases#known-issues) before a production install.

### Reverse proxy (option B)

Proxy only Authfly. A Caddy example:

```text
auth.example.com {
	reverse_proxy 127.0.0.1:4000
}
```

Never proxy Stalwart's HTTP port, not even behind a path filter: the same listener serves management, JMAP and dynamic client registration. Authfly publishes only an allowlist of OIDC endpoints. Requests carry the client address in `X-Forwarded-For`; Authfly trusts it only from `AUTHFLY_TRUSTED_PROXIES` (default `127.0.0.1,::1`). Add your proxy's address there if it runs elsewhere.

## 6. Lock-down checks

From outside the server:

```sh
curl -sI https://auth.example.com/.well-known/openid-configuration   # 200
curl -sI https://auth.example.com/jmap                               # 404 or 405
curl -sI https://auth.example.com/api/principal                      # 404 or 405
curl -sI https://auth.example.com/admin/                             # 404
```

Also check that ports 8080 (Stalwart HTTP), 4500 (Console) and 4001 (admin API) are not reachable from outside. Open the Console through an SSH tunnel only:

```sh
ssh -N -L 4500:127.0.0.1:4500 you@203.0.113.10
```

## 7. First users

Public sign-up is off (`AUTHFLY_PUBLIC_SIGNUP=false`), so registration needs an invite:

```sh
docker compose -p authfly-prod -f .authfly/compose.yml exec -T authfly app invite create -note "first admin" -ttl 72h
```

The token is printed once; hand it over through a private channel. `invite list` and `invite revoke <id>` manage them.

Reserved names: a built-in list (`admin`, `postmaster`, `abuse`, `root`, `support` and about sixty more) cannot be registered. To add names, mount a file with one name per line (`#` comments) and set `AUTHFLY_RESERVED_NAMES_FILE` on the authfly service. It extends the built-in list. This is an environment variable, not an `authfly.yaml` key.

Smoke test: invite, sign up, verify the email, sign in, open `/account`, sign out; try a reserved name and an already taken name (both refused).

## 8. Operations

- `authfly backup` writes an encrypted backup (AES-256-GCM with the generated `backup_key`) to `backup.dir`, keeping `backup.keep` (default 7). `authfly restore <file.afbk>` restores it. A backup is only proven by a restore: rehearse one on a spare host.
- `authfly upgrade` snapshots, upgrades to the newest release (or `--to <version>`) and rolls back automatically if it fails; `authfly rollback` restores the last snapshot.
- `authfly secrets` lists secrets and their permissions; values are never printed.
- `authfly doctor` audits the install against `authfly.yaml` and the policy pack (documents served, no IP addresses in the Authfly audit log, channel signature, image digests).

## 9. Lessons from the field

From an operator running Authfly, Stalwart and Caddy as separate unprivileged Incus containers (field report, adapted):

- **Keep secrets in files** (root-only, `chmod 600`), never in argv, shell history or the repository. Special characters in secrets break shell and YAML quoting. `authfly deploy` already writes its secrets as 0600 files under `.authfly/`.
- **Never publish the Stalwart admin UI**, and test from outside that its paths return 404.
- **Pin container addresses** so proxy upstreams survive restarts.
- **If you run the binary under systemd yourself** (outside the containers `deploy` manages), start it after the network is online:

```ini
# /etc/systemd/system/authfly.service.d/10-network-online.conf
[Unit]
Wants=network-online.target
After=network-online.target

[Service]
Restart=on-failure
RestartSec=5
```

- **Snapshot before every deploy** and keep off-host backups (the operator uses `incus snapshot` with a 7-day expiry and nightly restic); harden SSH and rate-limit it (for example fail2ban).
- **Wrap invite creation** in a small root-only script so tokens do not land in shell history.
- **Reboot test:** after a reboot every service must come back on its own.

> **Unverified:** the field report was written against a pre-release Authfly build (0.5.x) with Stalwart 1.0.0 used as an OIDC IdP and mail listeners disabled; it used a client id and secret read from a file and Stalwart's ES256 signing. Authfly 0.4.2 generates its own secrets and client registrations, so those settings have no `authfly.yaml` keys. The Docker variant of that setup was not tested by the operator; the Docker steps in this guide were tested separately (see [React + Docker on Authfly](/en/guides/react-docker)).
