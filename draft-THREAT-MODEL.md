<!--
  Licensed to the Apache Software Foundation (ASF) under one
  or more contributor license agreements.  See the NOTICE file
  distributed with this work for additional information
  regarding copyright ownership.  The ASF licenses this file
  to you under the Apache License, Version 2.0 (the
  "License"); you may not use this file except in compliance
  with the License.  You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

  Unless required by applicable law or agreed to in writing,
  software distributed under the License is distributed on an
  "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
  KIND, either express or implied.  See the License for the
  specific language governing permissions and limitations
  under the License.
-->

# Apache CloudStack CloudMonkey Security Threat Model — delta (draft)

> **Delta document.** Inherits §3, §4 B1, and §7 from
> `cloudstack-threat-model-draft.md`. Read the main model first.

## §1 Header

- **Project:** Apache CloudStack CloudMonkey (`apache/cloudstack-cloudmonkey`)
  — interactive shell and CLI for the CloudStack management server.
- **Commit:** `76df16c` (HEAD of `main` at draft time).
- **Date:** 2026-05-29.
- **Authors:** ASF Security team draft.
- **Status:** Draft delta over `cloudstack-threat-model-draft.md`.
- **Version binding:** as of the commit above.
- **Reporting:** as in the main model.
- **Provenance legend:** as in the main model.
- **Draft confidence:** 7 documented / 0 maintainer / 9 inferred.

**About the project.** CloudMonkey is a Go-language interactive shell
and command-line tool that simplifies CloudStack management
*(documented: `README.md`)*. It is the modern rewrite of the legacy
Python `cloudmonkey`. It is compatible with Apache CloudStack 4.9+. Its
moving parts: the `cli/` interactive REPL with autocompletion, the
`cmd/` command dispatch, the `config/` package that owns credential
storage and TLS-verify policy, and a bundled `apis.json` describing the
CloudStack API surface *(documented: `config/apis.json`)*.

## §2 Scope and intended use

**Primary intended use.** Run interactively from an operator's
workstation or non-interactively from a CI / scripting context to drive
the CloudStack JSON API. Used by admins and end users.

**Deployment shape.** A single user-space binary, run on the operator's
machine. No daemon, no listener.

**Caller expectations.** The caller (the operator running `cmk`) is
trusted to:

- supply a `~/.cmk/config` (INI-style) file (or env-equivalent) with
  the management-server URL, API key, secret, and `verifycert`
  preference,
- not run untrusted CloudMonkey extension scripts.

**Component-family table.**

| Family | Representative entry | Touches outside the process? | In this delta? |
| --- | --- | --- | --- |
| Config loader and signer (`config/config.go`) | `cfg.NewRequest` | **yes — network + creds** | yes |
| REPL / autocompletion (`cli/`) | `cmk` interactive prompt | reads `~/.cmk/` config and history | yes |
| Command dispatch (`cmd/`) | one-shot CLI invocation | as above | yes |
| Bundled `apis.json` | static description of API surface | none | yes |
| `snap/`, `Dockerfile` packaging | builds the binary for distribution | n/a at runtime | **out of model** *(§3)* |
| `vendor/` | vendored Go dependencies | n/a | **out of model** *(§3)* — vendored upstream code is in-model only at the wrapper boundary |

## §3 Out of scope (explicit non-goals)

The main model's §3 applies. **Additional** out-of-scope items
specific to CloudMonkey:

1. **Where the operator stores `~/.cmk/config`.** File permissions on
   the config file are the operator's responsibility. *(inferred — Q1)*
2. **Server-side correctness of CloudStack responses.** Same as the
   Go SDK delta.
3. **TLS verification when `verifycert = false`.** When the operator
   sets `verifycert = false` in the config, the HTTP client uses
   `InsecureSkipVerify: true` *(documented: `config/config.go` line
   ~226)*. That is operator choice.
4. **Operator typing wrong commands.** A wrong destructive command
   correctly executed is not a vulnerability.
5. **`vendor/`, `snap/`, `Dockerfile`, `performrelease.sh`.**
6. **The four sibling repos.**

## §4 Trust boundaries and data flow

The boundary is **the operator's invocation of `cmk` and its access to
`~/.cmk/config`**. The config file is operator-supplied credentials
+ endpoint + TLS-verify policy. Per-call API parameters typed at the
REPL are likewise operator input.

Bytes that come back from the management server (JSON responses) are
trusted control-plane content per the main model B1 transition.

**Per-call request shape** *(inferred — Q2)*: CloudMonkey uses the same
HMAC-SHA1 signing flow as the Go SDK (likely via the SDK or a near-copy).
See main model §8 P1 and §11a "SHA1 / constant-time" notes.

## §5 Assumptions about the environment

- **Host OS**: Linux, macOS, Windows; binaries published on GitHub
  Releases and as Docker / snap packages *(documented: `README.md`)*.
- **Filesystem**: writes a config file under `~/.cmk/`; writes a
  shell-history file *(inferred — Q1)*.
- **Network**: outbound HTTPS to the configured `apiurl`.
- **Auth**: stdlib TLS, opt-out via `verifycert = false`.
- **What CloudMonkey does not do**: no listener, no env-var
  consumption beyond shell defaults, no background process *(inferred
  — Q3)*.

## §5a Build-time and configuration variants

| Knob | Default | Stance | Effect |
| --- | --- | --- | --- |
| `verifycert` | `true` *(inferred — Q4)* | dev / self-signed | flips `InsecureSkipVerify` *(documented: `config/config.go` line ~226)* |
| `apikey` / `secretkey` | none | operator config | persisted in `~/.cmk/config` |
| `output` (`json`/`text`/`table`/`csv`) | per-config | display | not security-relevant |
| `timeout` | per-config | bounds async polling | not security-relevant by itself |
| history-file path | `~/.cmk/history` *(inferred — Q5)* | operator | may contain sensitive params (passwords, API keys typed inline) |
| `theme`, profiles | per-config | operator | not security-relevant |

## §6 Assumptions about inputs

| Entry point | Parameter | Attacker-controllable in the model? | Caller must enforce |
| --- | --- | --- | --- |
| `~/.cmk/config` | URL, key, secret, verifycert | **no** — operator config | filesystem permissions; do not check into VCS |
| Interactive REPL | typed command | **no** — operator input | nothing |
| `-p` profile selection from command line | name | **no** | nothing |
| Bundled `apis.json` | static description | **no** — ships with binary | n/a |
| HTTP response | JSON body | trusted (B1) | typed-decoded |

## §7 Adversary model

Main-model §7 applies. **Adjustments specific to CloudMonkey**:

- "Unauthenticated network peer reaching `:8080`" actor is upstream of
  CloudMonkey.
- An additional adversary worth naming: **a local attacker on the
  operator's host with non-operator UID who can read `~/.cmk/config`
  or `~/.cmk/history`**. Whether CloudMonkey enforces strict file
  perms on these files is Q1.

## §8 Security properties the CLI provides

### C1 — HMAC-SHA1 signature per the main model (via shared signer logic)

- **Property.** Each API call carries a valid HMAC-SHA1 signature, with
  the canonical parameter string. *(inferred — Q2; needs maintainer
  confirmation that CloudMonkey uses the same signer as the Go SDK or
  a vetted equivalent.)*
- **Conditions / violation / severity.** As main model §8 P1.

### C2 — TLS verification on by default

- **Property.** *(inferred — Q4)* default `verifycert = true`.
- **Conditions / violation / severity.** As Go-SDK delta §8 S2.

### C3 — Strict permissions on `~/.cmk/config`

- **Property.** *(inferred — Q1)* CloudMonkey writes its config with
  mode 0600 (or similar) on POSIX so that other users on the same host
  cannot read the API secret.
- **Violation symptom.** Config file is world-readable after `cmk`
  writes it.
- **Severity.** Security-critical, `VALID` per main-model §13.

## §9 Security properties the CLI does *not* provide

- **No protection against a local attacker on the same host with
  shell access.** If they can read `~/.cmk/config`, they have the API
  secret.
- **No history-file scrubbing.** Sensitive parameters typed inline
  (e.g. `createUser password=...`) end up in `~/.cmk/history`
  *(inferred — Q5)*.
- **No defence when `verifycert = false`.**
- **No re-validation of management-server response correctness.**

### False-friend properties

- **`verifycert = false` is not a "test mode."** It silently disables
  TLS verification across all subsequent calls.
- **HMAC-SHA1 — see main model §11a.**

## §10 Downstream responsibilities

The operator running `cmk` MUST:

1. Keep `~/.cmk/config` mode 0600 and outside any shared/synced
   location.
2. Set `verifycert = true` unless connecting to a known dev cluster.
3. Avoid typing API secrets / passwords on the interactive prompt
   (they land in history). Prefer scripted invocations that read from
   a secret store.
4. Rotate `apikey` / `secretkey` on suspicion of compromise.
5. Match `cmk` version to the management-server version where API
   compatibility matters *(inferred — Q6)*.

## §11 Known misuse patterns

- Checking `~/.cmk/config` into a Git repo.
- Setting `verifycert = false` in production.
- Sharing `~/.cmk/` across users on a multi-user host.
- Running `cmk` over an unencrypted SSH session that gets logged on
  the jump host.

## §11a Known non-findings (recurring false positives)

- **"HMAC-SHA1 — SHA1 is deprecated."** → `KNOWN-NON-FINDING` per main
  model §11a.
- **"`InsecureSkipVerify` is reachable from a config setting."** Gated
  by `verifycert`. → `OUT-OF-MODEL: trusted-input`.
- **"Config file contains plaintext credentials."** By design — see
  §10 item 1. → `BY-DESIGN: property-disclaimed`.
- **"`vendor/` has CVE-X."** Vendored upstream. → `OUT-OF-MODEL:
  unsupported-component` (upstream pointer).

## §12 Conditions that would change this delta

- A credential-store integration (keychain / agent) replacing the
  plaintext config file.
- A history-scrubbing feature.
- Change in signing algorithm at the main-model layer.

## §13 Triage dispositions

Use the same table as the main model.

## §14 Open questions for the maintainers

**Q1.** Does CloudMonkey set strict permissions (mode 0600) on
`~/.cmk/config` when it writes it? Does it warn the operator if the
file is found with looser perms?

**Q2.** Does CloudMonkey use `apache/cloudstack-go` as its signer, or
does it have its own near-copy? If a copy, does it stay in sync? Does
it set `signatureversion=3`?

**Q3.** Confirm CloudMonkey has no background daemon, no
env-var-credential channel, no listener.

**Q4.** Confirm `verifycert` default — `true`?

**Q5.** Does `~/.cmk/history` redact sensitive parameter values
(passwords, secret keys typed inline)? Is history per-profile or
shared?

**Q6.** Compatibility matrix between `cmk` version and CloudStack
version — where is it documented?

**Q7.** Meta — should this delta live at `docs/threat-model.md` in
`apache/cloudstack-cloudmonkey`?
