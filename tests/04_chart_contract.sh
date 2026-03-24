#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
services=(wifi-probe dns-probe jitter-probe gateway-monitor alert-receiver)

for svc in "${services[@]}"; do
  values="$ROOT_DIR/$svc/charts/$svc/values.yaml"
  values_k3s="$ROOT_DIR/$svc/charts/$svc/values-k3s.yaml"
  deployment_tpl="$ROOT_DIR/$svc/charts/$svc/templates/deployment.yaml"
  ingress_tpl="$ROOT_DIR/$svc/charts/$svc/templates/ingress.yaml"
  servicemonitor_tpl="$ROOT_DIR/$svc/charts/$svc/templates/servicemonitor.yaml"

  grep -qE '^image:' "$values" || { printf "%s values.yaml missing image section\n" "$svc" >&2; exit 1; }
  grep -qE '^  repository:' "$values" || { printf "%s values.yaml missing image.repository\n" "$svc" >&2; exit 1; }
  grep -qE '^  tag:' "$values" || { printf "%s values.yaml missing image.tag\n" "$svc" >&2; exit 1; }

  grep -qE '^image:' "$values_k3s" || { printf "%s values-k3s.yaml missing image section\n" "$svc" >&2; exit 1; }
  grep -qE "^  repository: pi-1.local:5000/$svc$" "$values_k3s" || { printf "%s values-k3s.yaml has unexpected image.repository\n" "$svc" >&2; exit 1; }
  grep -qE '^  tag:' "$values_k3s" || { printf "%s values-k3s.yaml missing image.tag\n" "$svc" >&2; exit 1; }
  grep -qE '^ingress:' "$values_k3s" || { printf "%s values-k3s.yaml missing ingress section\n" "$svc" >&2; exit 1; }
  grep -qE '^  enabled: true$' "$values_k3s" || { printf "%s values-k3s.yaml must enable ingress\n" "$svc" >&2; exit 1; }
  grep -qE '^  path: /metrics$' "$values_k3s" || { printf "%s values-k3s.yaml must expose /metrics via ingress\n" "$svc" >&2; exit 1; }
  grep -qE '^serviceMonitor:' "$values_k3s" || { printf "%s values-k3s.yaml missing serviceMonitor section\n" "$svc" >&2; exit 1; }
  grep -A6 '^serviceMonitor:' "$values_k3s" | grep -qE '^  enabled: true$' || { printf "%s values-k3s.yaml must enable serviceMonitor\n" "$svc" >&2; exit 1; }
  grep -A6 '^serviceMonitor:' "$values_k3s" | grep -qE '^    release: prometheus$' || { printf "%s values-k3s.yaml must label serviceMonitor for Prometheus Operator\n" "$svc" >&2; exit 1; }

  grep -qF '{{ .Values.image.repository }}:{{ .Values.image.tag }}' "$deployment_tpl" || {
    printf "%s deployment template missing image repository/tag wiring\n" "$svc" >&2
    exit 1
  }

  grep -q 'Values.ingress.enabled' "$ingress_tpl" || {
    printf "%s ingress template missing ingress.enabled gating\n" "$svc" >&2
    exit 1
  }

  grep -q 'kind: ServiceMonitor' "$servicemonitor_tpl" || {
    printf "%s serviceMonitor template missing ServiceMonitor kind\n" "$svc" >&2
    exit 1
  }

  grep -q 'jobLabel: app' "$servicemonitor_tpl" || {
    printf "%s serviceMonitor template missing jobLabel mapping\n" "$svc" >&2
    exit 1
  }
done

printf "Chart contract checks passed.\n"
