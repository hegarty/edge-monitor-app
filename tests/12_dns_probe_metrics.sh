#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
# shellcheck source=tests/lib/cluster_common.sh
source "$ROOT_DIR/tests/lib/cluster_common.sh"

skip_unless_cluster_tests "dns-probe metrics test"
init_kubectl

wait_for_deployment "dns-probe" "dns-probe"
svc="$(resolve_service_name "dns-probe" "dns-probe")"
payload="$(fetch_metrics_payload "dns-probe" "$svc" "9091")"

assert_metric_present "$payload" "dns_probe_up"
assert_metric_present "$payload" "dns_probe_latency_seconds"

printf "dns-probe metrics test passed.\n"
