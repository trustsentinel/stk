# stk — Docker Compose test scenario

A self-contained, runnable demo of the whole stk flow:

> **A client gets a shell on the agent, brokered through the hub over an
> end-to-end encrypted, mutually-authenticated channel — and the agent exposes
> no inbound port.**

```
client ──WS──> hub (broker) <──WS── agent (PTY /bin/sh)
       \___________ end-to-end Noise (XX) ___________/
```

The hub only relays opaque frames; it holds no keys and never sees plaintext. The
agent **dials out** to the hub, so there is no shell port to attack. Identities
are generated at startup by a one-shot `keygen` (not hardcoded).

## Services

| Service | Role |
|---|---|
| **keygen** | One-shot. Generates `agent` + `client` Noise keypairs into a shared volume. |
| **hub** | `stk-hub` broker on `:8443`. The only component that listens. Healthchecked. |
| **agent** | `stk-agent` dials the hub, authenticates the client, brokers a PTY `/bin/sh`. No inbound port. |
| **e2e** | Test driver: runs the assertions and exits 0 on PASS. |

## Quick start — automated test
```bash
cd deploy/compose
docker compose build
docker compose run --rm e2e      # exits 0 on PASS
docker compose down -v           # cleanup
```
`run` starts the dependencies via `depends_on` (keygen → completed, hub → healthy,
agent → started) and returns the test's exit code. (Avoid `up --exit-code-from e2e`
here: it aborts the moment the one-shot `keygen` exits, before the test runs.)

It builds the images, provisions keys, starts the hub and agent, and asserts:
1. a command **round-trips through the encrypted broker and actually executes** (`echo E2E:$((6*7))` → `E2E:42`)
2. the **agent exposes no inbound shell port** (tcp/22, /2222, /8443 all refused)
3. only the **hub** listens (contrast)
4. an **unauthorized client gets no shell** (the agent rejects an unknown key)

Expected final line:
```
RESULT: PASS  (brokered -> encrypted -> executed; no open shell port; unauthorized rejected)
```

## Manual walk-through
```bash
cd deploy/compose
docker compose up -d --build hub agent   # keygen runs first via depends_on

# run a one-shot command as the authorized client
docker compose run --rm --no-deps -v stk-compose_keys:/keys e2e sh -c '
  stk-client -hub ws://hub:8443/ws -room demo \
    -key "$(cat /keys/client.key)" -pubkey "$(cat /keys/client.pub)" \
    -authorized-agent "$(cat /keys/agent.pub)" -exec "id; uname -a"'

docker compose logs hub     # shows: pairing client<->agent ... hub sees only ciphertext
docker compose logs agent   # shows: secure session established with client <pubkey>
```

For an **interactive** shell, drop `-exec` and run `stk-client` with a TTY
(`docker compose run` allocates one by default).

## Configuration
- **room** — rendezvous id the client and agent share (`-room`, default `demo`).
- **keys** — `stk-keygen -dir /keys <name>...` writes `<name>.key` / `<name>.pub`.
  The agent authorizes one client key (`-authorized-client`); the client requires
  a specific agent key (`-authorized-agent`). Empty = accept any authenticated
  peer (dev only).
- **hub port** — internal by default; uncomment `ports: ["8443:8443"]` in
  `compose.yml` to reach it from the host.

## Cleanup
```bash
docker compose down -v
```

## Demo simplifications (vs production)
- The **hub is unauthenticated at the transport edge** — anyone who can reach it
  can attempt to pair. Security comes from the end-to-end Noise layer (the hub
  can't read sessions) and mutual key authorization. Production should still put
  the hub behind TLS + authn and rate-limit pairing.
- Keys live in a **shared volume** for demo convenience; production should use a
  secrets store and per-device provisioning/attestation.
- The agent brokers a plain **`/bin/sh`**; production should scope the shell,
  drop privileges, and audit sessions.
- XX authenticates the **initiator on the third message**, so an unauthorized
  client is rejected by the agent (no shell) but only learns so as the session
  drops. A pattern like IK/KK can reject earlier if that matters.
