#!/usr/bin/env bash
set -euo pipefail

if [[ "${RUN_CLUSTER_TESTS:-0}" != "1" ]]; then
  printf "Cluster smoke tests skipped (set RUN_CLUSTER_TESTS=1 to enable).\n"
  exit 0
fi

if ! command -v kubectl >/dev/null 2>&1; then
  printf "kubectl is not installed; cannot run cluster smoke tests.\n" >&2
  exit 1
fi

KUBE_CONTEXT="${KUBE_CONTEXT:-}"
KUBECTL=(kubectl)
if [[ -n "$KUBE_CONTEXT" ]]; then
  KUBECTL+=(--context "$KUBE_CONTEXT")
fi

printf "Running cluster smoke tests...\n"

"${KUBECTL[@]}" get nodes -o wide >/dev/null
"${KUBECTL[@]}" get pods -A >/dev/null
"${KUBECTL[@]}" get deploy -A >/dev/null

printf "Cluster smoke tests passed.\n"
