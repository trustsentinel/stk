<p align="center">
  <img src="docs/images/stk-logo.png" width="180" alt="stk">
</p>

<h1 align="center">stk</h1>

<p align="center">Secure lightweight broker for browser-based remote shell access.</p>

<p align="center">
  <code>Status: Active</code> · <code>Go · TypeScript</code> · <code>Noise Protocol</code> · part of <a href="https://trustsentinel.eu">TrustSentinel</a>
</p>

## Overview

stk gives a client a shell on a host without exposing SSH — or any inbound port —
to the network. A small Go **hub** brokers the connection; the client and the
host-side **agent** run an end-to-end, mutually-authenticated **Noise** session on
top, so the hub only ever relays ciphertext. The agent *dials out* to the hub, so
there is no shell port to attack.

A runnable MVP (`hub` + `agent` + `client`) builds, is unit-tested, and ships with
a Docker Compose end-to-end demo — see [Quick start](#quick-start). The original
2019 browser/xterm.js front end is kept under [`_legacy/`](_legacy/) as reference
while it is modernized onto the new core.

<p align="center">
  <img src="docs/images/stk_demo1.gif" width="700" alt="stk demo">
</p>

## Architecture

```
┌──────────────┐   Noise    ┌────────────┐   Noise    ┌──────────────┐
│ client       │ ─────────> │  stk hub   │ <───────── │  agent (host)│
│ CLI / browser│   (WS)     │  (broker)  │   (WS)     │  PTY shell   │
└──────────────┘            └────────────┘            └──────────────┘
   end-to-end encrypted (XX, mutual auth) — hub relays ciphertext, no open shell port
```

## Quick start

Build and unit-test:
```bash
go build ./...
go test ./...
```

Run the full end-to-end demo (client gets a brokered, encrypted shell on the agent):
```bash
cd deploy/compose
docker compose build
docker compose run --rm e2e      # exits 0 on PASS
docker compose down -v
```

Or without Docker: `./smoke.sh` starts a hub, agent, and client locally and
asserts a command round-trips through the encrypted broker.

## Layout
- `cmd/` — `stk-hub` (broker), `stk-agent` (PTY shell), `stk-client` (CLI), `stk-keygen`
- `internal/transport/` — message-framed connection (websocket / in-memory pipe)
- `internal/secure/` — Noise XX handshake, mutual-auth session (over `flynn/noise`)
- `internal/shell/` — PTY-backed shell (`creack/pty`)
- `crypto/` — key handling + secure randomness, a separately-tested module
- `deploy/` — container scenarios: `compose/` (full e2e) and `test/` (crypto tests)
- `_legacy/` — the original 2019 GOPATH prototype + browser front end, kept for reference (not built)

## Status & roadmap
Working Go MVP (hub + agent + client) with mutual-auth Noise, unit tests, and a
runnable Compose e2e. Next:
- restore the **browser / xterm.js** client on the new core (from `_legacy/auth/web`)
- **per-device identity** + enrollment (attestation) instead of static demo keys
- earlier initiator auth (**IK/KK**) so unauthorized clients are refused up front
- GitHub Actions CI (`go test` + `govulncheck` + `gosec`); Kubernetes manifests

## License
MIT

<sub>Originally prototyped at bluebycode/stk (2019); continued under TrustSentinel.</sub>
