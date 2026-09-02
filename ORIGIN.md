# Origin & Provenance

## Original project

`stk` was originally prototyped as **[bluebycode/stk](https://github.com/bluebycode/stk)**
(2019), by [@bluebycode](https://github.com/bluebycode) (vrandkode) — a secure
lightweight broker for browser-based remote shell access, built on the Noise
Protocol framework with a Go backend and a TypeScript/React (xterm.js) terminal
frontend, U2F/2FA, and a protobuf transport.

- Original repository: https://github.com/bluebycode/stk
- Original default branch: `develop` (49 commits)
- Companion prototypes (reference): `bluebycode/stk.s` (SSH bridge),
  `bluebycode/marshmallows` (IoT secure-comms, shares the Noise agent design)

## This repository

TrustSentinel's continuation. Imported **clean-start** (no upstream git history)
on 2026-09-02 from the original `develop` tree, then sanitized:

- Removed thesis/report PDFs (personal material).
- Replaced an example TOTP secret in `auth/auth.go` with a placeholder.
- Full `gitleaks` scan: clean (0 findings).

## Not to be confused with `stuk`

[`trustsentinel/stuk`](https://github.com/trustsentinel/stuk) is a **different
product**: a port-knocking SSH access manager (originally
[bluebycode/stuk](https://github.com/bluebycode/stuk), Ruby/Python). `stk` is the
browser broker; `stuk` is the port-knocking manager.
