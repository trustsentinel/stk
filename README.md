# stk

Secure lightweight broker for browser-based remote shell access.

`Status: Active` · `Go · TypeScript` · `Noise Protocol` · part of [TrustSentinel](https://trustsentinel.eu)

## Overview
stk puts a shell in the browser without exposing SSH to the network. A Go
agent/hub brokers end-to-end encrypted sessions over the Noise Protocol, with an
xterm.js terminal front end and TOTP/U2F authentication.

## Layout
- `agent/`, `auth/` — Go backend (hub, peers, Noise transport, protobuf)
- `auth/web/` — TypeScript/React terminal UI
- `reference/marshmallows-agent/` — sibling Noise agent kept for reference

## License
MIT

<sub>Originally prototyped at bluebycode/stk (2019); continued under TrustSentinel.</sub>
