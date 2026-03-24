#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
# shellcheck source=tests/lib/cluster_common.sh
source "$ROOT_DIR/tests/lib/cluster_common.sh"

skip_unless_cluster_tests "gateway-monitor metrics test"
init_kubectl

wait_for_deployment "gateway-monitor" "gateway-monitor"
svc="$(resolve_service_name "gateway-monitor" "gateway-monitor")"
payload="$(fetch_metrics_payload "gateway-monitor" "$svc" "9093")"

assert_metric_present "$payload" "gateway_reachable"
assert_metric_present "$payload" "wan_reachable"

printf "gateway-monitor metrics test passed.\n"
