#!/bin/sh
# Starts a hub (serving the browser client) + an agent for manual/browser testing.
# Prints the agent public key and the URL to open, then stays running.
set -e
cd "$(dirname "$0")"
PORT=18443
BIN=$(mktemp -d)
KEYS=$(mktemp -d)

cleanup() { kill "$HUB_PID" "$AGENT_PID" 2>/dev/null || true; rm -rf "$BIN" "$KEYS"; }
trap cleanup EXIT INT TERM

go build -o "$BIN/hub" ./cmd/stk-hub
go build -o "$BIN/agent" ./cmd/stk-agent
go build -o "$BIN/keygen" ./cmd/stk-keygen
# build the browser client + copy the matching wasm runtime shim
GOROOT="$(go env GOROOT)"
if [ -f "$GOROOT/lib/wasm/wasm_exec.js" ]; then cp "$GOROOT/lib/wasm/wasm_exec.js" web/wasm_exec.js
elif [ -f "$GOROOT/misc/wasm/wasm_exec.js" ]; then cp "$GOROOT/misc/wasm/wasm_exec.js" web/wasm_exec.js; fi
GOOS=js GOARCH=wasm go build -o web/stk.wasm ./cmd/stk-wasm

"$BIN/keygen" -dir "$KEYS" agent >/dev/null 2>&1
AGENT_PUB=$(cat "$KEYS/agent.pub"); AGENT_KEY=$(cat "$KEYS/agent.key")

"$BIN/hub" -addr ":$PORT" -webdir web >"$BIN/hub.log" 2>&1 &
HUB_PID=$!
i=0; until curl -fsS "http://localhost:$PORT/healthz" >/dev/null 2>&1; do i=$((i+1)); [ "$i" -gt 50 ] && { cat "$BIN/hub.log"; exit 1; }; sleep 0.1; done

# agent accepts any authenticated client (browser uses an ephemeral key)
"$BIN/agent" -hub "ws://localhost:$PORT/ws" -room demo \
  -key "$AGENT_KEY" -pubkey "$AGENT_PUB" >"$BIN/agent.log" 2>&1 &
AGENT_PID=$!
i=0; until grep -q "agent registered" "$BIN/hub.log" 2>/dev/null; do i=$((i+1)); [ "$i" -gt 50 ] && { cat "$BIN/agent.log"; exit 1; }; sleep 0.1; done

echo "READY"
echo "AGENT_PUB=$AGENT_PUB"
echo "URL=http://localhost:$PORT/?room=demo&agent=$AGENT_PUB"
echo "(logs: $BIN/hub.log $BIN/agent.log)"
# keep running until killed
while kill -0 "$HUB_PID" 2>/dev/null; do sleep 1; done
