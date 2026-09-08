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

A runnable MVP builds, is unit-tested, and ships with a Docker Compose end-to-end
demo — see [Quick start](#quick-start). There are two clients: a native CLI
(`stk-client`) and a **browser client** — the same Go Noise code compiled to
WebAssembly, driving an xterm.js terminal (see [`web/`](web/)). The original 2019
browser front end is kept under [`_legacy/`](_legacy/) for reference.

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

Try the **browser client** (xterm.js + wasm):
```bash
make web                         # build web/stk.wasm
./browser-demo.sh                # starts a hub (serving web/) + an agent, prints a URL
```
Open the printed URL and press Connect. See [`web/`](web/).

Without Docker: `./smoke.sh` (or `make smoke`) starts a hub, agent, and client
locally and asserts a command round-trips through the encrypted broker.

## Layout
- `cmd/` — `stk-hub` (broker), `stk-agent` (PTY shell), `stk-client` (CLI), `stk-keygen`, `stk-wasm` (browser client)
- `internal/transport/` — message-framed connection (websocket / browser websocket / in-memory pipe)
- `internal/secure/` — Noise XX handshake, mutual-auth session (over `flynn/noise`)
- `internal/shell/` — PTY-backed shell (`creack/pty`)
- `web/` — browser client: xterm.js UI + the wasm build glue
- `crypto/` — key handling + secure randomness, a separately-tested module
- `deploy/` — container scenarios: `compose/` (full e2e) and `test/` (crypto tests)
- `_legacy/` — the original 2019 GOPATH prototype + browser front end, kept for reference (not built)

## Status & roadmap
Working Go MVP (hub + agent + CLI **and** browser clients) with mutual-auth Noise,
unit tests, a runnable Compose e2e, and GitHub Actions CI. Next:
- **per-device identity** + enrollment (attestation) instead of generated demo keys
- earlier initiator auth (**IK/KK**) so unauthorized clients are refused up front
- Kubernetes manifests (see [`deploy/README.md`](deploy/README.md))

## License
MIT

<sub>Originally prototyped at bluebycode/stk (2019); continued under TrustSentinel.</sub>
