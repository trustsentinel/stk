#!/bin/sh
# Generates agent + client Noise identities and creates the Kubernetes
# Secrets/ConfigMaps stk needs. Keys never touch git. Run from the repo root.
#
#   deploy/k8s/gen-secrets.sh [namespace]
set -e
NS="${1:-stk}"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

# Generate <name>.key (base64 private) and <name>.pub (base64 public).
go run ./cmd/stk-keygen -dir "$TMP" agent client >/dev/null 2>&1

# Identity files are two base64 lines: private, then public.
printf '%s\n%s\n' "$(cat "$TMP/agent.key")"  "$(cat "$TMP/agent.pub")"  > "$TMP/agent.identity"
printf '%s\n%s\n' "$(cat "$TMP/client.key")" "$(cat "$TMP/client.pub")" > "$TMP/client.identity"
# Enroll the client: its public key goes into the agent's authorized-clients file.
cp "$TMP/client.pub" "$TMP/authorized_clients"

kubectl create namespace "$NS" --dry-run=client -o yaml | kubectl apply -f -
kubectl -n "$NS" create secret generic stk-agent-identity  --from-file=identity="$TMP/agent.identity"   --dry-run=client -o yaml | kubectl apply -f -
kubectl -n "$NS" create secret generic stk-client-identity --from-file=identity="$TMP/client.identity"  --dry-run=client -o yaml | kubectl apply -f -
kubectl -n "$NS" create configmap stk-authorized-clients   --from-file=authorized_clients="$TMP/authorized_clients" --dry-run=client -o yaml | kubectl apply -f -
kubectl -n "$NS" create configmap stk-agent-pub            --from-file=agent.pub="$TMP/agent.pub"        --dry-run=client -o yaml | kubectl apply -f -

echo "created identities + enrollment registry in namespace $NS"
echo "agent pubkey: $(cat "$TMP/agent.pub")"
