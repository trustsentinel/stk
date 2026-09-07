# Reference: marshmallows Noise agent

A sibling Go implementation of a Noise-Protocol agent, lifted from the
**marshmallows** project (2019, IoT secure-comms, CyberCamp) as reference to
mine for `stk`.

Worth borrowing into stk's own `agent/`:
- `beacon.go` — periodic beacon / keep-alive + reconnect logic
- `comm.go`, `conn.go`, `io.go` — connection lifecycle and framed I/O
- `noise.go` — Noise handshake wiring
- `protocol/requests.proto` — a cleaner protobuf message set

This is reference code (not wired into the build). Integration & modernization
happen in Phase 5. gitleaks: clean.
