# stk — porting status

The 2019 prototype is GOPATH-era and does **not** build under Go modules as-is.
Blockers found (2026-09-02):

- **`gopkg.in/noisesocket.v0`** — the socket-framing dependency used by
  `agent/channel.go` and `auth/noise.go` is not a resolvable Go module
  (`go get` → "no matching versions"). Needs replacing with a transport built on
  `github.com/flynn/noise` (already a dependency).
- **Relative imports** (`import requests "./protocol"`) — unsupported in modules;
  rewrite to full module paths.
- **Symlinked shared file** — `agent/common.go` / `auth/common.go` symlink to a
  shared `common/common.go`; materialize as real files.
- **`externals/`** — experimental tree with more relative imports; exclude from the build.
- Duplicate/commented `func main()` and a `package main` used as a library.

## Done
- `crypto/` — the stdlib-only crypto/random helpers, extracted into a proper
  module with tests. Fixes a real bug: the original `DecodeKey` copied only 4 of
  32 key bytes.

## Next (see repo TASKS)
1. Root `go.mod`; rewrite relative imports; materialize symlinks; exclude `externals/`.
2. Replace `noisesocket.v0` transport with a `flynn/noise` framing.
3. proto2 → proto3; tests + CI + Docker demo (mirroring stuk).
