#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
# shellcheck source=tests/lib/cluster_common.sh
source "$ROOT_DIR/tests/lib/cluster_common.sh"

skip_unless_cluster_tests "jitter-probe metrics test"
init_kubectl

wait_for_deployment "jitter-probe" "jitter-probe"
svc="$(resolve_service_name "jitter-probe" "jitter-probe")"
payload="$(fetch_metrics_payload "jitter-probe" "$svc" "9092")"

assert_metric_present "$payload" "network_latency_ms"
assert_metric_present "$payload" "network_jitter_ms"

printf "jitter-probe metrics test passed.\n"
