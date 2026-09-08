# stk — porting status

The 2019 prototype is GOPATH-era and does **not** build under Go modules. Rather
than force-port it in place, we rebuilt the core as a clean, tested Go module and
kept the original tree under [`_legacy/`](../_legacy/) as reference (the same
approach that worked for stuk).

## Done — clean MVP (2026-09-07)
A runnable module now lives at the repo root (`module github.com/trustsentinel/stk`):

- **`cmd/stk-hub`** — the broker. Pairs a client and an agent by room and relays
  opaque frames; holds no keys, so it only ever sees ciphertext.
- **`cmd/stk-agent`** — dials **out** to the hub (no inbound port), authenticates
  the client, and brokers a PTY `/bin/sh` over the encrypted session.
- **`cmd/stk-client`** — connects, authenticates the agent, drives the shell
  (interactive, or `-exec` for one command).
- **`internal/secure`** — Noise **XX** (mutual static-key auth) over
  `github.com/flynn/noise`, with a peer allow-list. This **replaces the dead
  `gopkg.in/noisesocket.v0`** transport that blocked the original build.
- **`internal/transport`** — message-framed conn (websocket / in-memory pipe).
- **`internal/shell`** — PTY-backed shell (`github.com/creack/pty`).
- Unit tests for the handshake (incl. **unauthorized-peer rejection**) and
  transport; a runnable Compose e2e in [`deploy/compose/`](../deploy/compose/):
  *blocked → brokered → encrypted → executed; no open shell port.*
- **`cmd/stk-wasm` + `web/`** — the **browser client**: the same `secure`/`transport`
  Go code compiled to WebAssembly (`GOOS=js GOARCH=wasm`) over a browser-WebSocket
  `MsgConn`, driving an xterm.js terminal. One audited Noise implementation, shared
  by the CLI and the browser — no JS reimplementation. Verified in a real browser
  (handshake → PTY shell → command execution → multiple sessions).
- **CI** — `.github/workflows/ci.yml`: build+test (`-race`), gofmt, wasm build,
  crypto-module tests, `govulncheck`/`gosec`, and the Compose e2e.
- **Noise IK + early auth** — switched the handshake from XX to **IK**: the client
  pins the agent up front (defeats a malicious hub/MITM) and the agent authenticates
  the client on the FIRST message, dropping an unauthorized client before any
  session or shell exists (tested).
- **Per-device identity + enrollment** — persistent identities (`-identity`,
  created 0600 on first use) and an SSH-`authorized_keys`-style client registry
  (`-authorized-clients`, re-read each session, so enrolling needs no restart).
- **`crypto/`** — the stdlib-only crypto/random helpers as a separate tested
  module. Fixes a real bug: the original `DecodeKey` copied only 4 of 32 key bytes.

## How the original blockers were resolved
| 2019 blocker | Resolution |
|---|---|
| `gopkg.in/noisesocket.v0` (dead module) | rewritten as `internal/secure` on `flynn/noise` (IK) |
| relative imports (`"./protocol"`, `"websocket"`, `"common"`) | proper module paths in the new code |
| symlinked shared `common.go` across two `package main`s | real packages under `internal/` |
| `externals/` experimental tree | left in `_legacy/`, excluded from the build |
| proto2 generated code | not needed by the MVP; a proto3/connect-go schema is future work if wire-compat with the 2019 protocol is wanted |

## Not yet ported / next
1. **Hardware attestation** — bind the device identity to a TPM/secure element so a
   key can't be exfiltrated; add a provisioning/enrollment service.
2. **Kubernetes manifests** — see [`deploy/k8s/`](../deploy/k8s/).
3. **Richer browser UX** (from `_legacy/auth/web`) — window resize/SIGWINCH,
   reconnect, session list, TOTP/WebAuthn login.
