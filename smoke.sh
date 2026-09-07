#!/bin/sh
# Local (no-Docker) end-to-end smoke test: hub + agent + client, generated keys,
# mutual Noise auth, one brokered command. Asserts E2E:42 round-trips.
set -e
cd "$(dirname "$0")"
BIN=$(mktemp -d)
KEYS=$(mktemp -d)
PORT=18443
HUB="ws://localhost:${PORT}/ws"

cleanup() {
  [ -n "$HUB_PID" ] && kill "$HUB_PID" 2>/dev/null || true
  [ -n "$AGENT_PID" ] && kill "$AGENT_PID" 2>/dev/null || true
  rm -rf "$BIN" "$KEYS"
}
trap cleanup EXIT

echo "building binaries..."
go build -o "$BIN/hub" ./cmd/stk-hub
go build -o "$BIN/agent" ./cmd/stk-agent
go build -o "$BIN/client" ./cmd/stk-client
go build -o "$BIN/keygen" ./cmd/stk-keygen

echo "generating keys..."
"$BIN/keygen" -dir "$KEYS" agent client >/dev/null 2>&1

AGENT_PUB=$(cat "$KEYS/agent.pub"); AGENT_KEY=$(cat "$KEYS/agent.key")
CLIENT_PUB=$(cat "$KEYS/client.pub"); CLIENT_KEY=$(cat "$KEYS/client.key")

echo "starting hub..."
"$BIN/hub" -addr ":${PORT}" >"$BIN/hub.log" 2>&1 &
HUB_PID=$!

# wait for hub health
i=0; until curl -fsS "http://localhost:${PORT}/healthz" >/dev/null 2>&1; do
  i=$((i+1)); [ "$i" -gt 50 ] && { echo "hub did not start"; cat "$BIN/hub.log"; exit 1; }
  sleep 0.1
done

echo "starting agent..."
"$BIN/agent" -hub "$HUB" -room demo -once \
  -key "$AGENT_KEY" -pubkey "$AGENT_PUB" -authorized-client "$CLIENT_PUB" \
  >"$BIN/agent.log" 2>&1 &
AGENT_PID=$!

# wait for the agent to register with the hub
i=0; until grep -q "agent registered" "$BIN/hub.log" 2>/dev/null; do
  i=$((i+1)); [ "$i" -gt 50 ] && { echo "agent did not register"; cat "$BIN/agent.log"; exit 1; }
  sleep 0.1
done

echo "running client exec..."
OUT=$("$BIN/client" -hub "$HUB" -room demo \
  -key "$CLIENT_KEY" -pubkey "$CLIENT_PUB" -authorized-agent "$AGENT_PUB" \
  -exec 'echo E2E:$((6*7))' 2>"$BIN/client.log" || true)

echo "----- client output -----"; printf '%s\n' "$OUT"; echo "-------------------------"
if printf '%s' "$OUT" | grep -q 'E2E:42'; then
  echo "SMOKE: PASS (brokered, mutually-authenticated, encrypted command round-trip)"
else
  echo "SMOKE: FAIL"; echo "--- agent.log ---"; cat "$BIN/agent.log"; echo "--- client.log ---"; cat "$BIN/client.log"; exit 1
fi
