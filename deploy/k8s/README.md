# stk on Kubernetes

Manifests to run stk on a cluster, mapping the architecture onto Kubernetes:

| Resource | Role |
|---|---|
| `stk-hub` **Deployment** + **Service** | the broker; the only component that accepts inbound traffic (relays ciphertext, holds no keys) |
| `stk-agent` **Deployment** | dials **out** to the hub Service — no inbound port, no Service (production: a `DaemonSet` or sidecar) |
| **Secrets** `stk-agent-identity`, `stk-client-identity` | persistent Noise identities (never committed; created by `gen-secrets.sh`) |
| **ConfigMap** `stk-authorized-clients` | the enrollment registry the agent reads |
| **ConfigMap** `stk-agent-pub` | the agent key the client pins |
| **NetworkPolicies** | default-deny ingress; hub reachable only from stk pods on 8443; agent egress limited to the hub + DNS |
| `stk-e2e` **Job** (`test-job.yaml`) | on-cluster end-to-end test: an enrolled client runs a command through the brokered, encrypted channel |

## One-command test (throwaway kind cluster)

```bash
deploy/k8s/kind-test.sh
```
Builds the image, creates a kind cluster, loads the image, provisions identities,
applies the manifests, and runs the e2e Job. Expected tail:
```
K8S RESULT: PASS
```

## Manual (any cluster)

```bash
# 1. build + make the image available to the cluster
docker build -f deploy/compose/Dockerfile -t stk:local .
kind load docker-image stk:local            # or push to a registry the cluster can pull

# 2. provision identities + the enrollment registry (keys never touch git)
deploy/k8s/gen-secrets.sh                    # prints the agent public key

# 3. deploy
kubectl apply -k deploy/k8s
kubectl -n stk rollout status deploy/stk-hub deploy/stk-agent

# 4. use it — run the e2e Job, or a one-off client
kubectl -n stk apply -f deploy/k8s/test-job.yaml
kubectl -n stk logs job/stk-e2e
```

Enrolling another client: append its public key to the `stk-authorized-clients`
ConfigMap (the agent re-reads it each session — no restart).

## Exposing the hub / real deployment notes
- **Ingress + TLS** — put the hub behind an Ingress with `cert-manager` to terminate
  the browser WebSocket over TLS. The hub is the only thing you expose; sessions
  stay end-to-end encrypted regardless.
- **Agent as DaemonSet** — to broker a shell into each node, run the agent as a
  `DaemonSet` (or a sidecar next to a workload); the pod spec is the same, it just
  dials out to the hub.
- **NetworkPolicy enforcement** — needs a CNI that enforces it (Calico/Cilium).
  The default kind CNI does not, so on plain kind the policies are inert but valid.
- **Identities** — back them with a secrets manager / CSI and, ideally, bind them
  to hardware (TPM/secure element) so a key can't be exfiltrated.
