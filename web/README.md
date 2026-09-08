# stk browser client

An xterm.js terminal plus the Go client compiled to **WebAssembly**. It runs the
*same* Noise XX handshake as the native CLI (`internal/secure`) over a browser
WebSocket (`internal/transport` JSConn), so keystrokes are end-to-end encrypted to
the agent — the hub only relays ciphertext. No Noise is reimplemented in JS.

## Build
```bash
GOOS=js GOARCH=wasm go build -o web/stk.wasm ./cmd/stk-wasm   # or: make web
```
`wasm_exec.js` (Go's wasm runtime shim, copied from GOROOT) and `vendor/xterm.*`
are checked in so the page is self-contained and needs no CDN (works offline).
`stk.wasm` is a build artifact and is git-ignored — build it before serving.

## Serve
```bash
stk-hub -addr :8443 -webdir web
```
Open `http://localhost:8443/?room=demo&agent=<agent-public-key>`, then start an
agent that dials the same hub and room:
```bash
stk-agent -hub ws://localhost:8443/ws -room demo -key <priv> -pubkey <pub>
```
One-command local hub+agent for manual testing: [`../browser-demo.sh`](../browser-demo.sh).

## Files
| File | Purpose |
|---|---|
| `index.html` / `main.js` | UI + glue: wires xterm.js to the wasm session |
| `wasm_exec.js` | Go wasm runtime shim (from GOROOT) |
| `vendor/xterm.*` | terminal emulator, vendored (no CDN) |
| `stk.wasm` | compiled Go client (build artifact, git-ignored) |
