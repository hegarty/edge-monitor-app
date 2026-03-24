#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

required_files=(
  "$ROOT_DIR/plans/04-K3S-REMOTE-WRITE.md"
  "$ROOT_DIR/plans/examples/edge-metrics-forwarder.alloy"
  "$ROOT_DIR/plans/examples/edge-metrics-forwarder.yaml"
)

for file in "${required_files[@]}"; do
  [[ -f "$file" ]] || {
    printf "Missing remote write artifact: %s\n" "$file" >&2
    exit 1
  }
done

grep -qF 'edge-metrics' "$ROOT_DIR/plans/04-K3S-REMOTE-WRITE.md" || {
  printf "Remote write plan missing edge-metrics namespace reference\n" >&2
  exit 1
}

grep -qF 'Amazon Managed Service for Prometheus' "$ROOT_DIR/plans/04-K3S-REMOTE-WRITE.md" || {
  printf "Remote write plan missing AMP reference\n" >&2
  exit 1
}

grep -qF 'prometheus.remote_write "amp"' "$ROOT_DIR/plans/examples/edge-metrics-forwarder.alloy" || {
  printf "Alloy example missing remote_write block\n" >&2
  exit 1
}

grep -qF 'sigv4 {' "$ROOT_DIR/plans/examples/edge-metrics-forwarder.alloy" || {
  printf "Alloy example missing sigv4 auth block\n" >&2
  exit 1
}

for svc in wifi-probe dns-probe jitter-probe gateway-monitor alert-receiver; do
  grep -qF "$svc.$svc.svc.cluster.local" "$ROOT_DIR/plans/examples/edge-metrics-forwarder.alloy" || {
    printf "Alloy example missing scrape target for %s\n" "$svc" >&2
    exit 1
  }
done

grep -qF 'grafana/alloy:v1.12.1' "$ROOT_DIR/plans/examples/edge-metrics-forwarder.yaml" || {
  printf "Remote write manifest missing pinned Alloy image\n" >&2
  exit 1
}

printf "Remote write contract checks passed.\n"
