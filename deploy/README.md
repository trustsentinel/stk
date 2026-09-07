# stk — deployment & test scenarios

Two containerized scenarios, both runnable today.

| Scenario | Path | What it proves |
|---|---|---|
| **Full end-to-end** | [`compose/`](compose/) | A client gets a shell on the agent, brokered over an end-to-end encrypted, mutually-authenticated Noise channel — and the agent exposes no inbound port |
| **Crypto unit tests** | [`test/`](test/) | The stdlib-only `crypto/` module builds and passes in a container |

## Full end-to-end (recommended)

```bash
cd compose
docker compose build
docker compose run --rm e2e      # exits 0 on PASS
docker compose down -v
```

Topology (matches the README architecture):

```
client ──WS──> hub (broker) <──WS── agent (PTY /bin/sh)
       \___________ end-to-end Noise (XX) ___________/
```

The hub only relays ciphertext (it holds no keys); the agent dials out and listens
on nothing. Asserts: a command round-trips **and executes** through the encrypted
broker, the agent has no inbound shell port, and an unauthorized client gets no
shell. Details + manual walk-through: [`compose/README.md`](compose/README.md).

## Crypto unit tests

```bash
docker compose -f test/compose.yml run --rm crypto-test
```

Builds the `crypto/` module and runs its unit tests (including the key-truncation
regression test). See [`test/`](test/).

## Kubernetes — the next step

Compose is the proving ground; the images it builds map cleanly onto Kubernetes
when you need multi-node, real TLS, and network policy:

- **hub** → `Deployment` + `Service`, fronted by an `Ingress` with TLS
  (`cert-manager`) terminating the browser WebSocket — the only externally
  reachable component.
- **agent** → `DaemonSet` (or a sidecar) that **dials out** to the hub, so it
  needs no inbound `Service` — the whole security point. A scoped
  `securityContext` gates what shell it brokers.
- **keys** → Noise static keys as `Secret`s / CSI, never baked into images.
- **NetworkPolicy** → agent egress to the hub only; hub ingress only from the
  ingress controller and agents.

A `kind`/`k3d` cluster plus a small Helm chart is enough to test it. Tracked in
the repo's `TASKS.md`.
