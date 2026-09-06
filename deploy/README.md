# stk — deployment & test scenarios

stk is mid-port from a 2019 GOPATH prototype, so its container story comes in
**two honest tiers**. There is no single "run the whole thing" demo yet, because
the broker/agent code does not build — see [`../docs/PORTING.md`](../docs/PORTING.md).

| Tier | File | Runs today? | What it proves |
|---|---|:---:|---|
| **1 · Crypto tests** | [`test/`](test/) | ✅ yes | The tested core (key handling + secure randomness) builds and passes in a container |
| **2 · Full e2e** | [`e2e.blueprint.yml`](e2e.blueprint.yml) | ⛔ blueprint | The target browser→hub→agent scenario, ready to wire up once the transport is ported |

## Tier 1 — crypto tests (runnable now)

```bash
docker compose -f deploy/test/compose.yml run --rm crypto-test
```

Builds the `crypto/` module in a `golang:1.21-alpine` container and runs its unit
tests (including the regression test for the original key-truncation bug, where
`DecodeKey` copied only 4 of 32 key bytes). `go vet` runs at build time as a
compile gate; the tests run at container start so you see the output live and the
exit code reflects pass/fail. No network needed — the module is stdlib-only.

Expected tail:
```
ok  	github.com/trustsentinel/stk/crypto
```

## Tier 2 — full end-to-end (blueprint, not yet runnable)

[`e2e.blueprint.yml`](e2e.blueprint.yml) describes the real test topology that
mirrors the README architecture:

```
web (xterm.js)  ──WS/Noise──>  hub (broker)  <──WS/Noise──  agent (host shell)
```

It **does not run today**: the `hub` and `agent` Go services don't compile
(`gopkg.in/noisesocket.v0` is no longer a resolvable module). To keep the file
safe in-tree, every app service is gated behind the `e2e` compose profile, so a
plain `docker compose up` starts nothing. It becomes a real, runnable demo — the
way [stuk's `deploy/compose/`](https://github.com/trustsentinel/stuk/tree/main/deploy/compose)
already works — once the port checklist in the file header is done:

1. Root `go.mod` + module-path imports (drop the relative `./protocol` imports).
2. Replace `noisesocket.v0` with a `flynn/noise` transport in `agent/` + `auth/`.
3. Add `deploy/hub.Dockerfile` and `deploy/agent.Dockerfile`.
4. Author `deploy/test-e2e.sh` asserting a command round-trips through the broker
   and that no direct shell port is ever exposed.

## Kubernetes — later, on purpose

Compose is the right first target; Kubernetes is not harder conceptually but adds
nothing until the images build. When the port lands, the natural mapping is:

- **hub** → `Deployment` + `Service` (ClusterIP), fronted by an `Ingress` with TLS
  (the only externally reachable component; it terminates the browser WebSocket).
- **agent** → `DaemonSet` on the hosts you want a brokered shell into (or a
  sidecar next to a workload); it dials **out** to the hub, so it needs no inbound
  Service — which is the whole security point.
- **web** → static assets served by the hub or a tiny `nginx` Deployment.
- **secrets** → Noise static keys + TOTP secrets via `Secret`/CSI, never baked
  into images.

A single-node `kind`/`k3d` cluster is enough to test it; a Helm chart is a
nice-to-have once the compose demo is green. Until then, k8s manifests would just
be YAML for images that don't exist yet.
