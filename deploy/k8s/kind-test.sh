#!/bin/sh
# End-to-end test of the k8s manifests on a throwaway kind cluster:
# build image -> load -> provision identities -> apply -> run the e2e Job.
# Run from anywhere; resolves the repo root itself.
set -e
CLUSTER="${STK_KIND_CLUSTER:-stk-test}"
IMAGE=stk:local
REPO_ROOT="$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)"
cd "$REPO_ROOT"

cleanup() { kind delete cluster --name "$CLUSTER" >/dev/null 2>&1 || true; }
trap cleanup EXIT

echo "[1/6] build image $IMAGE"
docker build -f deploy/compose/Dockerfile -t "$IMAGE" . >/dev/null

echo "[2/6] create kind cluster $CLUSTER"
kind create cluster --name "$CLUSTER" >/dev/null

echo "[3/6] load image into the cluster"
kind load docker-image "$IMAGE" --name "$CLUSTER" >/dev/null

echo "[4/6] provision identities + apply manifests"
deploy/k8s/gen-secrets.sh stk
kubectl apply -k deploy/k8s >/dev/null

echo "[5/6] wait for rollouts"
kubectl -n stk rollout status deploy/stk-hub  --timeout=120s
kubectl -n stk rollout status deploy/stk-agent --timeout=120s

echo "[6/6] run the e2e Job"
kubectl -n stk delete job stk-e2e --ignore-not-found >/dev/null 2>&1 || true
kubectl -n stk apply -f deploy/k8s/test-job.yaml >/dev/null
if kubectl -n stk wait --for=condition=complete job/stk-e2e --timeout=150s 2>/dev/null; then
  echo "----- e2e job logs -----"; kubectl -n stk logs job/stk-e2e
  echo "K8S RESULT: PASS"
else
  echo "----- e2e job logs -----"; kubectl -n stk logs job/stk-e2e 2>/dev/null || true
  kubectl -n stk get pods; kubectl -n stk describe job/stk-e2e 2>/dev/null | tail -20 || true
  echo "K8S RESULT: FAIL"; exit 1
fi
