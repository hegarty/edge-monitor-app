#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
# shellcheck source=tests/lib/cluster_common.sh
source "$ROOT_DIR/tests/lib/cluster_common.sh"

skip_unless_cluster_tests "wifi-probe metrics test"
init_kubectl

wait_for_deployment "wifi-probe" "wifi-probe"
svc="$(resolve_service_name "wifi-probe" "wifi-probe")"
payload="$(fetch_metrics_payload "wifi-probe" "$svc" "9090")"

assert_metric_present "$payload" "wifi_probe_up"
assert_metric_present "$payload" "wifi_probe_latency_seconds"

printf "wifi-probe metrics test passed.\n"
