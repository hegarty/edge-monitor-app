# Deterministic Deployment to k3s (Raspberry Pi)

## Objective

Deploy the full service set to Raspberry Pi `k3s` from deterministic container images in a registry reachable by the cluster.

## Required Outcome

- every service image pushed to one registry with one shared `RELEASE_ID`
- Helm releases updated in k3s with explicit repository and tag
- rollout and metrics checks succeed from the k3s context

## Inputs

```bash
export TARGET="k3s"
kubectl config get-contexts -o name
export KUBE_CONTEXT="<pi-k3s-context-from-kubeconfig>"
export K3S_REGISTRY="pi-1.local:5000"
export RELEASE_ID="$(date +%Y%m%d-%H%M)-$(git rev-parse --short HEAD)"
```

Set `KUBE_CONTEXT` to the context in your kubeconfig that points at the Pi k3s API server. On this workstation that context is currently `default` for `https://pi-1.local:6443`.

## Preconditions

```bash
kubectl --context "$KUBE_CONTEXT" get nodes
curl -fsS "http://$K3S_REGISTRY/v2/" >/dev/null
```

If `alert-receiver` secret does not exist, scaffold one before deployment:

```bash
kubectl --context "$KUBE_CONTEXT" -n alert-receiver create secret generic alert-receiver-secrets \
  --from-literal=OPENAI_API_KEY=dummy \
  --dry-run=client -o yaml | kubectl --context "$KUBE_CONTEXT" apply -f -
```

## Build, Push, Deploy

`make deploy-k3s` uses each service chart profile at `charts/<service>/values-k3s.yaml`. Each k3s profile also enables a metrics ingress endpoint at `http://<service>.pi-1.local/metrics`.

```bash
services=(wifi-probe dns-probe jitter-probe gateway-monitor alert-receiver)

for svc in "${services[@]}"; do
  kubectl --context "$KUBE_CONTEXT" create namespace "$svc" --dry-run=client -o yaml | kubectl --context "$KUBE_CONTEXT" apply -f -

  make -C "$svc" deploy-k3s \
    IMAGE_TAG="$RELEASE_ID" \
    K3S_REGISTRY="$K3S_REGISTRY" \
    KUBE_CONTEXT="$KUBE_CONTEXT" \
    NAMESPACE="$svc"

  kubectl --context "$KUBE_CONTEXT" rollout status "deployment/$svc" -n "$svc" --timeout=180s
done
```

## Verification

```bash
kubectl --context "$KUBE_CONTEXT" get deploy -A
kubectl --context "$KUBE_CONTEXT" get pods -A
```

Endpoint checks from inside cluster:

```bash
kubectl --context "$KUBE_CONTEXT" run curl-check --rm -it --restart=Never --image=curlimages/curl:8.7.1 -- \
  sh -c '
    set -e
    for p in \
      wifi-probe.wifi-probe.svc.cluster.local:9090 \
      dns-probe.dns-probe.svc.cluster.local:9091 \
      jitter-probe.jitter-probe.svc.cluster.local:9092 \
      gateway-monitor.gateway-monitor.svc.cluster.local:9093 \
      alert-receiver.alert-receiver.svc.cluster.local:9094; do
      curl -fsS "http://$p/metrics" >/dev/null
      echo "OK $p"
    done
  '
```

Ingress endpoint checks from outside cluster:

```bash
for svc in wifi-probe dns-probe jitter-probe gateway-monitor alert-receiver; do
  curl -fsS "http://$svc.pi-1.local/metrics" >/dev/null
  echo "OK ingress $svc"
done
```

## Rollback

Redeploy using a known previous `RELEASE_ID`:

```bash
export RELEASE_ID="<previous-release-id>"
# rerun the same deployment loop
```

This keeps rollback deterministic and auditable.
