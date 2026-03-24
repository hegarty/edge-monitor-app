# k3s Remote Write Forwarding (Optional)

## Objective

Forward `edge-monitor-app` metrics from `pi-1.local` to an external Prometheus-compatible backend without running a full Prometheus and Grafana stack on the Pi.

## Recommended Shape

Use one lightweight scrape-forwarder in a dedicated namespace:

- namespace: `edge-metrics`
- forwarder: Grafana Alloy
- scrape source: in-cluster service DNS for each application
- sink: Prometheus remote write endpoint in AWS

This keeps scrape traffic local to the cluster and avoids using ingress as the hop between the Pi and the forwarder. The per-service ingress endpoints remain useful for manual checks and external scrapers.

## Why This Path

- lower footprint than a full Prometheus plus Grafana deployment on Raspberry Pi
- bounded local disk usage through a small Alloy WAL
- Prometheus-compatible egress via remote write
- deterministic, single-purpose deployment owned by this repository

Do not send raw metrics directly to S3. S3 is object storage, not a Prometheus ingestion endpoint. If you want AWS-hosted metrics, send Prometheus remote write to a metrics backend such as Amazon Managed Service for Prometheus.

## Inputs

```bash
kubectl config get-contexts -o name

export KUBE_CONTEXT="<pi-k3s-context-from-kubeconfig>"
export AWS_REGION="us-east-1"
export AMP_WORKSPACE_ID="ws-REPLACE_ME"
export REMOTE_WRITE_URL="https://aps-workspaces.${AWS_REGION}.amazonaws.com/workspaces/${AMP_WORKSPACE_ID}/api/v1/remote_write"
export AWS_ACCESS_KEY_ID="REPLACE_ME"
export AWS_SECRET_ACCESS_KEY="REPLACE_ME"
```

The remote write URL path format is:

```text
/workspaces/<workspace-id>/api/v1/remote_write
```

## Example Artifacts

This plan ships two concrete example files:

- [`plans/examples/edge-metrics-forwarder.alloy`](/Users/terencehegarty/projects/hegarty/edge-monitor-app/plans/examples/edge-metrics-forwarder.alloy)
- [`plans/examples/edge-metrics-forwarder.yaml`](/Users/terencehegarty/projects/hegarty/edge-monitor-app/plans/examples/edge-metrics-forwarder.yaml)

The YAML example deploys:

- namespace `edge-metrics`
- secret `edge-metrics-remote-write`
- configmap `edge-metrics-forwarder-config`
- deployment `edge-metrics-forwarder`
- cluster-internal service `edge-metrics-forwarder`

## Deploy

Review and edit the example YAML first:

```bash
sed -n '1,240p' plans/examples/edge-metrics-forwarder.yaml
```

Replace the placeholder secret values in the manifest, then apply it:

```bash
kubectl --context "$KUBE_CONTEXT" apply -f plans/examples/edge-metrics-forwarder.yaml
kubectl --context "$KUBE_CONTEXT" -n edge-metrics rollout status deployment/edge-metrics-forwarder --timeout=180s
```

## Verify

Check the deployment:

```bash
kubectl --context "$KUBE_CONTEXT" -n edge-metrics get all
kubectl --context "$KUBE_CONTEXT" -n edge-metrics logs deployment/edge-metrics-forwarder --tail=100
```

Port-forward the Alloy UI when you need to inspect targets and remote write state:

```bash
kubectl --context "$KUBE_CONTEXT" -n edge-metrics port-forward svc/edge-metrics-forwarder 12345:12345
```

Then open:

```text
http://127.0.0.1:12345
```

Expected behavior:

- five scrape targets present
- successful scrapes against in-cluster app services
- remote write queue draining normally

## Operational Notes

- the example uses an `emptyDir` volume with a 512MiB limit for the Alloy WAL
- the example WAL keeps a short buffer window to handle transient WAN outages without turning the Pi into a long-retention store
- if you need stronger durability across pod restarts, replace `emptyDir` with a small PVC such as `1Gi`

## Security Notes

- the example uses static AWS credentials in a Kubernetes secret because this cluster is running on `k3s` outside AWS
- prefer a narrow IAM policy scoped to remote write for a single AMP workspace
- rotate the credentials and re-apply the secret when needed

## Rollback

```bash
kubectl --context "$KUBE_CONTEXT" delete -f plans/examples/edge-metrics-forwarder.yaml
```
