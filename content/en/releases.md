---
title: Releases
description: Authfly release history, how to verify a release, and the known issues of the current release.
---
# Releases

Releases are published on GitHub in authfly/cli with the binaries, `SHA256SUMS`, the signed `channel.json` and container images on GHCR pinned by digest.

| Version | Date | Highlights |
|---|---|---|
| 0.4.2 | 2026-11-02 | widgets and the UI8Kit middle layer generated for nine runtimes by authfly/codegen; release tooling packs npm from the release binaries; Twig and Latte mirrors tagged |
| 0.4.1 | 2026-10-24 | signed release channel (`channel.json.sig`, SSHSIG namespace `authfly-channel`) |
| 0.4.0 | 2026-10-24 | first release with the channel, `ghcr.io/authfly/authfly` and `ghcr.io/authfly/console` images |

Images of 0.4.2: `ghcr.io/authfly/authfly:0.4.2` and `ghcr.io/authfly/console:0.4.2` (pull by the digest from `channel.json`).

## Verify a release

```sh
curl -fsSLO https://github.com/authfly/cli/releases/download/v0.4.2/SHA256SUMS
sha256sum --ignore-missing -c SHA256SUMS
```

`authfly deploy` and `doctor` verify the channel signature themselves; `doctor` prints the signing key on PASS.

## Known issues

Release 0.4.2, found while testing the guides in November 2026:

| Issue | Effect | Workaround |
|---|---|---|
| The channel pins `stalwartlabs/stalwart:v0.16.27`, the server needs Stalwart 1.0 | `stalwart-init` fails: `role: notCreated invalidPatch enabledPermissions` | set the stalwart image to `stalwartlabs/stalwart:v1.0.0` in `.authfly/compose.yml` and `docker compose … up -d`; `doctor` then FAILs the digest check |
| Proxy mode (`AUTHFLY_IDP=off`) with app clients | `scope=openid email profile` is answered with `invalid_scope` | `AUTHFLY_IDP: "on"` and `AUTHFLY_IDP_KEYGEN: "true"` on the authfly service |
| IdP mode has no consent step | users are not asked to approve an app | none; register only first-party apps |
| `@authfly/node` 1.0.0 `verifyIdToken` | rejects an array `aud` | use `openid-client` |
| SDK `signUp` has no `invite` | sign-up fails with public sign-up off | call `POST /api/v1/sign-up` with `invite` |
| `GET /api/v1/capabilities` not routed | Go SDK `Capabilities` fails | read `GET /api/v1/config` |
| Discovery `protected_resources` | shows the internal container host name | none needed for clients |
| npm `authfly` | `latest` is 0.4.1 | download the 0.4.2 binary from GitHub |

Both compose edits are overwritten by the next `authfly deploy`.
