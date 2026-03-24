#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

required_files=(
  "$ROOT_DIR/AGENTS.md"
  "$ROOT_DIR/CLAUDE.md"
  "$ROOT_DIR/plans/00-START.md"
  "$ROOT_DIR/plans/01-DEPLOYMENT-CONTRACT.md"
  "$ROOT_DIR/plans/02-K3D-DEPLOYMENT.md"
  "$ROOT_DIR/plans/03-K3S-DEPLOYMENT.md"
  "$ROOT_DIR/tests/TESTS.md"
  "$ROOT_DIR/tests/lib/cluster_common.sh"
  "$ROOT_DIR/tests/11_wifi_probe_metrics.sh"
  "$ROOT_DIR/tests/12_dns_probe_metrics.sh"
  "$ROOT_DIR/tests/13_jitter_probe_metrics.sh"
  "$ROOT_DIR/tests/14_gateway_monitor_metrics.sh"
  "$ROOT_DIR/tests/15_alert_receiver_metrics.sh"
)

services=(
  wifi-probe
  dns-probe
  jitter-probe
  gateway-monitor
  alert-receiver
)

for f in "${required_files[@]}"; do
  [[ -f "$f" ]] || { printf "Missing required file: %s\n" "$f" >&2; exit 1; }
done

for svc in "${services[@]}"; do
  [[ -f "$ROOT_DIR/$svc/Makefile" ]] || { printf "Missing Makefile for service: %s\n" "$svc" >&2; exit 1; }
  [[ -f "$ROOT_DIR/$svc/Dockerfile" ]] || { printf "Missing Dockerfile for service: %s\n" "$svc" >&2; exit 1; }
  [[ -f "$ROOT_DIR/$svc/charts/$svc/values.yaml" ]] || { printf "Missing values.yaml for service chart: %s\n" "$svc" >&2; exit 1; }
  [[ -f "$ROOT_DIR/$svc/charts/$svc/values-k3s.yaml" ]] || { printf "Missing values-k3s.yaml for service chart: %s\n" "$svc" >&2; exit 1; }
done

printf "Context integrity checks passed.\n"
