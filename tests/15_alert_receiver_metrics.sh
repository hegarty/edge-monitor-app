#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
# shellcheck source=tests/lib/cluster_common.sh
source "$ROOT_DIR/tests/lib/cluster_common.sh"

skip_unless_cluster_tests "alert-receiver metrics test"
init_kubectl

wait_for_deployment "alert-receiver" "alert-receiver"
svc="$(resolve_service_name "alert-receiver" "alert-receiver")"
payload="$(fetch_metrics_payload "alert-receiver" "$svc" "9094")"

assert_metric_present "$payload" "alert_receiver_alerts_received_total"
assert_metric_present "$payload" "alert_receiver_queue_depth"

printf "alert-receiver metrics test passed.\n"
