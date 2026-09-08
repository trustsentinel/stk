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

## Kubernetes

Manifests in [`k8s/`](k8s/) run stk on a cluster — hub `Deployment`+`Service`,
agent `Deployment` (dials out, no inbound port), identities as `Secret`s, an
enrollment `ConfigMap`, and `NetworkPolicies`. One-command test on a throwaway
kind cluster:

```bash
deploy/k8s/kind-test.sh      # build -> load -> deploy -> run the e2e Job -> PASS
```

Details, the manual flow, and production notes (Ingress+TLS, agent as a
`DaemonSet`): [`k8s/README.md`](k8s/README.md).
