---
title: Policy packs
description: Technical defaults and document templates for GDPR, CCPA/CPRA and Russian 152-FZ, what each pack sets and what it does not do. Not legal advice.
---
# Policy packs

> **Warning:** not legal advice. Packs set technical defaults and render document templates with `[BRACKETS]` to fill. Every rendered document starts with a not-legal-advice banner. Have counsel review them.

`authfly init --policy gdpr|us-pii|ru-152fz|none` stores `policy.pack`. On every `deploy` the pack:

1. renders documents into `.authfly/legal/*.md`, served at `/legal/{name}`;
2. sets server behaviour: the consent label on sign-up, extra consent kinds, the consent version, audit retention;
3. adds `doctor` checks: documents served, no IP addresses in the Authfly audit log, and with `host_country` set, the residency declaration.

## Rights in every pack

| Right | How |
|---|---|
| access / portability | `GET /api/v1/me/export` and the account widget "Download" (format `authfly-export/1`) |
| erasure | `DELETE /api/v1/me` with the password; Stalwart account and mailbox removed, sessions revoked, address removed from store and audit log, pseudonym kept |
| operator side | Console export and erasure with typed confirmation, logged in both audit logs |

## gdpr: Regulation (EU) 2016/679

| Setting | Value | Article |
|---|---|---|
| Documents | `privacy` (Art. 13 notice), `dpa` (Art. 28(3) template), `records-of-processing` (Art. 30) | 13, 28, 30 |
| Sign-up consent | terms and privacy notice, with a version record | 7(1) |
| IP addresses | none in the Authfly audit log | 5(1)(c) |
| Audit retention | 365 days, then the subject is pseudonymised | 5(1)(e) |
| Response time (docs) | 1 month, extendable by 2 | 12(3) |
| Breach notice (reminder) | 72 h | 33 |

## us-pii: CCPA as amended by CPRA

| Setting | Value | Section |
|---|---|---|
| Documents | `privacy` (policy plus notice at collection), `dpa` (service-provider contract) | §1798.100, §1798.140(ag) |
| Rights listed | know, delete, correct, opt-out, limit, non-discrimination | §1798.105–.125 |
| Response time | 45 days, plus one 45-day extension | §1798.130(a)(2) |
| `policy.us_states` | only California is verified; other state laws are UNVERIFIED | |
| Revenue threshold | CPI-adjusted; current amount UNVERIFIED | §1798.140(d) |

## ru-152fz: Federal Law No. 152-FZ "On Personal Data"

| Setting | Value | Article |
|---|---|---|
| Documents | `privacy` (processing policy, in Russian), `pd-consent` (separate consent), `dpa` (operator's instruction) | 18.1(2), 9(1), 6(3) |
| Sign-up consent | two consents recorded: `terms` and `pd_processing` | 9 |
| Residency | `policy.host_country: RU` required by the wizard; `doctor` checks the declaration and local stores | 18(5) |
| IP addresses | none in the Authfly audit log | 5(7) |
| Response time | 10 working days, +5 with a reasoned notice | 20 |
| Breach (reminder) | Roskomnadzor 24 h / 72 h | 21(3.1) |
| Reminders | notification to Roskomnadzor before processing; cross-border notice | 22, 12 |

> **Unverified:** the amendment dates quoted in the pack reminders (separate consent from 2025-09-01, localisation wording from 2025-07-01) are taken from the pack text and were not re-checked.

## What a pack does not do

- It files nothing: no notification, no signed DPA, no records kept for you.
- It does not decide where you host; residency is a declaration plus a check.
- It does not keep IP addresses out of Stalwart's own logs; keep Stalwart log retention short.
