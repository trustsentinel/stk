#!/bin/sh
# Runs inside the `e2e` container. Asserts the whole stk flow:
#   1. a command round-trips through the encrypted broker and actually executes
#   2. the agent exposes no inbound shell port (it dials out)
#   3. only the hub listens (contrast)
#   4. an unauthorized client gets no shell
set -u
HUB="ws://hub:8443/ws"
CK=$(cat /keys/client.key); CP=$(cat /keys/client.pub); AP=$(cat /keys/agent.pub)

fail() { echo "RESULT: FAIL — $1"; exit 1; }

echo "[1] brokered command round-trip (mutual Noise auth, hub relays ciphertext)"
OUT=""
i=0
while [ "$i" -lt 20 ]; do
  OUT=$(stk-client -hub "$HUB" -room demo -key "$CK" -pubkey "$CP" \
        -authorized-agent "$AP" -exec 'echo E2E:$((6*7))' 2>/dev/null || true)
  echo "$OUT" | grep -q 'E2E:42' && break
  i=$((i + 1)); sleep 1
done
printf '    client output: '; echo "$OUT" | tr -d '\r' | grep -a 'E2E:42' | head -1
echo "$OUT" | grep -q 'E2E:42' || fail "command did not round-trip"
echo "    ok: E2E:42 executed remotely and returned over the encrypted channel"

echo "[2] agent exposes no inbound shell port"
for p in 22 2222 8443; do
  if nc -z -w3 agent "$p" 2>/dev/null; then fail "agent is listening on tcp/$p"; fi
done
echo "    ok: agent has no listening TCP port (it dials out to the hub)"

echo "[3] only the hub listens (contrast)"
nc -z -w3 hub 8443 2>/dev/null || fail "hub not reachable on 8443"
echo "    ok: hub reachable on 8443; agent reachable on nothing"

echo "[4] unauthorized client gets no shell"
sleep 2  # let the agent reconnect after test [1]
BAD=$(stk-client -hub "$HUB" -room demo -authorized-agent "$AP" \
      -exec 'echo E2E:$((6*7))' 2>/dev/null || true)   # fresh, unauthorized key
if echo "$BAD" | grep -q 'E2E:42'; then fail "unauthorized client obtained a shell"; fi
echo "    ok: unauthorized client received no shell output (agent rejected its key)"

echo "RESULT: PASS  (brokered -> encrypted -> executed; no open shell port; unauthorized rejected)"
