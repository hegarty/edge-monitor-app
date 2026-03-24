#!/usr/bin/env bash

cluster_tests_enabled() {
  [[ "${RUN_CLUSTER_TESTS:-0}" == "1" ]]
}

skip_unless_cluster_tests() {
  local test_name="$1"
  if ! cluster_tests_enabled; then
    printf "%s skipped (set RUN_CLUSTER_TESTS=1 to enable).\n" "$test_name"
    exit 0
  fi
}

init_kubectl() {
  if ! command -v kubectl >/dev/null 2>&1; then
    printf "kubectl is not installed; cannot run cluster tests.\n" >&2
    exit 1
  fi

  KUBE_CONTEXT="${KUBE_CONTEXT:-}"
  KUBECTL=(kubectl)
  if [[ -n "$KUBE_CONTEXT" ]]; then
    KUBECTL+=(--context "$KUBE_CONTEXT")
  fi
}

wait_for_deployment() {
  local namespace="$1"
  local deployment="$2"
  local timeout="${3:-180s}"

  "${KUBECTL[@]}" rollout status "deployment/${deployment}" -n "$namespace" --timeout="$timeout" >/dev/null
}

resolve_service_name() {
  local namespace="$1"
  local app_label="$2"

  local service_name
  service_name="$("${KUBECTL[@]}" -n "$namespace" get svc -l "app=${app_label}" -o jsonpath='{.items[0].metadata.name}')"

  if [[ -z "$service_name" ]]; then
    printf "No service found in namespace %s with label app=%s\n" "$namespace" "$app_label" >&2
    exit 1
  fi

  printf "%s" "$service_name"
}

fetch_metrics_payload() {
  local namespace="$1"
  local service_name="$2"
  local port="$3"

  "${KUBECTL[@]}" get --raw "/api/v1/namespaces/${namespace}/services/http:${service_name}:${port}/proxy/metrics"
}

assert_metric_present() {
  local payload="$1"
  local metric_name="$2"

  if ! grep -qE "^${metric_name}(\{|\s|$)" <<<"$payload"; then
    printf "Metric not found: %s\n" "$metric_name" >&2
    exit 1
  fi
}
