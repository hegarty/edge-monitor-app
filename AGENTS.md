# AGENTS.md

## Purpose
This file is the primary entrypoint for coding agents working in this repository.
Use it for operational guidance and constraints.

Detailed context lives in:
- `CLAUDE.md` (canonical architecture and deployment contract)
- `plans/` (incremental deterministic deployment execution plans)
- `tests/TESTS.md` (verification suite and execution policy)

## Repository Scope
This repository owns edge-monitor application code and application-level deployment artifacts.

In scope:
- Go application code
- Dockerfiles and per-service Makefiles
- per-service Helm charts
- deterministic app deployment procedures to k3d and k3s

Out of scope:
- cluster provisioning and node lifecycle operations
- base Prometheus/Grafana platform installation
- cluster networking and ingress controller installation

## Current Operating Status
Deterministic deployment process is defined in plans and must be followed in sequence.

Current state:
- release contract and target-specific steps are documented in `plans/`
- verification suite exists in `tests/`
- deployment is expected to be repeatable for `k3d` and `k3s` with one shared `RELEASE_ID`
- optional k3s metrics forwarding path is documented for remote write to an external Prometheus-compatible backend

## Deterministic Deployment Rules
- use one immutable `RELEASE_ID` for all services in a release run
- deploy services in fixed order: `wifi-probe`, `dns-probe`, `jitter-probe`, `gateway-monitor`, `alert-receiver`
- explicitly set target context in every `kubectl` and `helm` invocation
- use target-specific Helm values profiles (`values.yaml` for k3d, `values-k3s.yaml` for k3s)
- never use mutable tags (`latest`) for shared environments
- use `hello-world` only for local experiments, not production deployment workflows

## Verification Workflow
Run verification before handing work back:

1. `./tests/run-all.sh`
2. `RUN_CLUSTER_TESTS=1 ./tests/run-all.sh` for cluster-affecting changes

If a check fails, update docs/code and rerun until green.

## Constraints
- preserve repository boundary: no infrastructure provisioning logic in this repo
- keep workloads lightweight for Raspberry Pi resource constraints
- keep metric names stable and low-cardinality
- preserve short-drop detection sensitivity (1-3 second events)

## Update Rules
When making meaningful deployment workflow changes:
1. update `AGENTS.md` if operating instructions changed
2. update `CLAUDE.md` if architecture or deployment contract changed
3. update relevant files in `plans/` when command sequences change
4. update `tests/TESTS.md` and scripts when verification policy changes

## Canonical References
- Architecture and engineering constraints: `CLAUDE.md`
- Deterministic execution order: `plans/00-START.md`
- k3d deployment path: `plans/02-K3D-DEPLOYMENT.md`
- k3s deployment path: `plans/03-K3S-DEPLOYMENT.md`
- optional k3s remote-write path: `plans/04-K3S-REMOTE-WRITE.md`
- Verification suite and policy: `tests/TESTS.md`
