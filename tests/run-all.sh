#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TEST_DIR="$ROOT_DIR/tests"

TEST_SCRIPTS=(
  "$TEST_DIR/01_context_integrity.sh"
  "$TEST_DIR/02_docs_consistency.sh"
  "$TEST_DIR/03_makefile_surface.sh"
  "$TEST_DIR/04_chart_contract.sh"
  "$TEST_DIR/05_remote_write_contract.sh"
  "$TEST_DIR/10_cluster_smoke.sh"
  "$TEST_DIR/11_wifi_probe_metrics.sh"
  "$TEST_DIR/12_dns_probe_metrics.sh"
  "$TEST_DIR/13_jitter_probe_metrics.sh"
  "$TEST_DIR/14_gateway_monitor_metrics.sh"
  "$TEST_DIR/15_alert_receiver_metrics.sh"
)

printf "Running repository verification tests...\n"

for test_script in "${TEST_SCRIPTS[@]}"; do
  if [[ ! -x "$test_script" ]]; then
    printf "ERROR: missing or non-executable test script: %s\n" "$test_script" >&2
    exit 1
  fi

  printf "\n==> %s\n" "$(basename "$test_script")"
  "$test_script"
done

printf "\nAll verification tests completed successfully.\n"
