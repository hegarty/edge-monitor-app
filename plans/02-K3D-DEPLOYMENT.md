# Deterministic Deployment to k3d (macOS)

## Objective

Deploy the full service set to local `k3d` with deterministic image tags and deterministic rollout verification.

## Required Outcome

- every service image imported into `k3d` with the same `RELEASE_ID`
- Helm releases updated with matching `image.repository` and `image.tag`
- all deployments report successful rollout
- all metrics endpoints return HTTP 200 from inside the cluster

## Inputs

```bash
export TARGET="k3d"
export K3D_CLUSTER="k3d-local"
export KUBE_CONTEXT="k3d-${K3D_CLUSTER}"
export RELEASE_ID="$(date +%Y%m%d-%H%M)-$(git rev-parse --short HEAD)"
```

Set the kube context explicitly if your cluster naming differs.

## Preconditions

```bash
k3d cluster list
kubectl --context "$KUBE_CONTEXT" get nodes
```

If `alert-receiver` secret does not exist, scaffold one before deployment:

```bash
kubectl --context "$KUBE_CONTEXT" -n alert-receiver create secret generic alert-receiver-secrets \
  --from-literal=OPENAI_API_KEY=dummy \
  --dry-run=client -o yaml | kubectl --context "$KUBE_CONTEXT" apply -f -
```

## Deploy

```bash
services=(wifi-probe dns-probe jitter-probe gateway-monitor alert-receiver)

for svc in "${services[@]}"; do
  make -C "$svc" push-k3d IMAGE_TAG="$RELEASE_ID" K3D_CLUSTER="$K3D_CLUSTER"

  kubectl --context "$KUBE_CONTEXT" create namespace "$svc" --dry-run=client -o yaml | kubectl --context "$KUBE_CONTEXT" apply -f -

  helm upgrade --install "$svc" "$svc/charts/$svc" \
    --kube-context "$KUBE_CONTEXT" \
    --namespace "$svc" \
    -f "$svc/charts/$svc/values.yaml" \
    --set image.tag="$RELEASE_ID"

  kubectl --context "$KUBE_CONTEXT" rollout status "deployment/$svc" -n "$svc" --timeout=120s
done
```

## Verification

```bash
kubectl --context "$KUBE_CONTEXT" get deploy -A
kubectl --context "$KUBE_CONTEXT" get pods -A
```

In-cluster metric smoke checks:

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

## Completion Criteria

- all rollouts succeeded
- no pod in `CrashLoopBackOff`, `ImagePullBackOff`, or `ErrImagePull`
- all five metric checks pass
