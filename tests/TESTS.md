# Tests Guide

This directory contains verification tests for deterministic deployment documentation and command surface.

## Quick Start

Run all default verification checks:

```bash
./tests/run-all.sh
```

Run with live cluster smoke checks enabled:

```bash
RUN_CLUSTER_TESTS=1 ./tests/run-all.sh
```

Optional explicit context for cluster checks:

```bash
KUBE_CONTEXT=<pi-k3s-context-from-kubeconfig> RUN_CLUSTER_TESTS=1 ./tests/run-all.sh
```

## Test Inventory

- `01_context_integrity.sh`
  - ensures required context files exist
  - ensures each canonical service has Makefile, Dockerfile, `values.yaml`, and `values-k3s.yaml`

- `02_docs_consistency.sh`
  - ensures `AGENTS.md` references canonical docs and plans
  - ensures `plans/00-START.md` includes the full ordered plan list
  - ensures `CLAUDE.md` includes required deterministic deployment sections

- `03_makefile_surface.sh`
  - verifies each service Makefile exposes expected deterministic deployment targets
  - verifies key Makefile variables (`IMAGE_TAG`, `REGISTRY`, `K3D_CLUSTER`, `K3S_REGISTRY`, `KUBE_CONTEXT`) are present
  - verifies Helm and kubectl targets require an explicit kube context

- `04_chart_contract.sh`
  - verifies each service chart has `image.repository` and `image.tag` in both k3d and k3s values profiles
  - verifies k3s chart profile repositories are pinned to `pi-1.local:5000/<service>`
  - verifies deployment templates wire image fields from chart values
  - verifies k3s chart profiles enable ingress on `/metrics`
  - verifies ingress templates only render when ingress is enabled
  - verifies k3s chart profiles enable ServiceMonitor resources for Prometheus Operator

- `05_remote_write_contract.sh`
  - verifies the optional k3s remote-write plan and example artifacts exist
  - verifies the Alloy example scrapes all canonical services
  - verifies the example manifest pins an Alloy image tag

- `10_cluster_smoke.sh`
  - optional live cluster checks (`RUN_CLUSTER_TESTS=1`)
  - validates nodes, pods, and deployments are queryable via `kubectl`

- `11_wifi_probe_metrics.sh`
  - optional live app test (`RUN_CLUSTER_TESTS=1`)
  - verifies `wifi-probe` rollout and expected metrics in `/metrics`

- `12_dns_probe_metrics.sh`
  - optional live app test (`RUN_CLUSTER_TESTS=1`)
  - verifies `dns-probe` rollout and expected metrics in `/metrics`

- `13_jitter_probe_metrics.sh`
  - optional live app test (`RUN_CLUSTER_TESTS=1`)
  - verifies `jitter-probe` rollout and expected metrics in `/metrics`

- `14_gateway_monitor_metrics.sh`
  - optional live app test (`RUN_CLUSTER_TESTS=1`)
  - verifies `gateway-monitor` rollout and expected metrics in `/metrics`

- `15_alert_receiver_metrics.sh`
  - optional live app test (`RUN_CLUSTER_TESTS=1`)
  - verifies `alert-receiver` rollout and expected metrics in `/metrics`

## Agent Usage Pattern

For documentation or workflow updates:

```bash
./tests/run-all.sh
```

For cluster-affecting updates:

```bash
RUN_CLUSTER_TESTS=1 ./tests/run-all.sh
```

## Exit Behavior

- Any failing check exits non-zero.
- `run-all.sh` stops at first failure.
- Cluster smoke checks are opt-in by design.
