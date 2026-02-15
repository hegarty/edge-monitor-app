# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

Repository: edge-monitor-app

This repository contains application logic only. It does NOT contain Kubernetes infrastructure definitions.

All cluster and infrastructure configuration lives in a separate repository:

edge-monitor-infra

This project is intentionally minimal, edge-focused, and observability-driven. All AI-assisted edits must preserve simplicity, low resource usage, and Kubernetes-native design.

---

# Real-World Problem Statement

Members of this household believe that the home WiFi intermittently cuts out for brief periods.

Observed characteristics:

- Drops are irregular.
- Typically last only 1–5 seconds.
- Long enough to:
  - Disrupt video playback.
  - Interrupt stock chart updates.
  - Affect real-time services.
- Not long enough to trigger obvious router alarms.

The purpose of this repository is to build lightweight observability applications capable of detecting and quantifying short-lived WiFi instability.

The system must be capable of detecting outages lasting as little as 1–3 seconds.

---

# Repository Scope

edge-monitor-app contains:

- Golang applications
- Application-level probing logic
- Dockerfiles
- Makefile or build.sh scripts
- Container build logic
- Prometheus metric exposure

edge-monitor-app does NOT contain:

- k3d cluster definitions
- Kubernetes manifests (unless strictly application-specific)
- Helm charts
- Terraform
- Prometheus configuration
- Alertmanager configuration

All infrastructure belongs in edge-monitor-infra.

Do not introduce infrastructure code into this repository.

---

# Application Strategy

This repository will contain multiple independent monitoring services, each compiled as a separate binary.

Recommended structure:

```
/cmd
  /wifi-probe
  /dns-probe
  /jitter-probe
  /gateway-monitor
```

Each service:

- Is an independent Go binary.
- Exposes Prometheus metrics.
- Logs to stdout only.
- Has its own Dockerfile (or shared multi-target build).
- Is deployable independently in Kubernetes.

Do not merge services into a monolithic application.

---

# Applications to Build

Claude should implement the following services.

---

## 1. wifi-probe

Purpose:
Detect basic TCP and HTTP reachability failures.

Behavior:
- Probe TCP targets.
- Probe HTTP targets.
- Measure latency.
- Detect connection failures.

Metrics:
- wifi_probe_up
- wifi_probe_latency_seconds
- wifi_probe_runs_total
- wifi_probe_errors_total

---

## 2. dns-probe

Purpose:
Detect DNS resolution failures and latency spikes.

Behavior:
- Resolve configurable domains repeatedly.
- Measure lookup latency.
- Track timeout events.

Metrics:
- dns_probe_up
- dns_probe_latency_seconds
- dns_probe_timeouts_total

This helps identify DNS-related micro-outages.

---

## 3. jitter-probe

Purpose:
Detect short (1–3 second) latency spikes and packet loss bursts.

Behavior:
- High-frequency TCP or ICMP sampling (250–500ms recommended).
- Track rolling latency.
- Track packet loss bursts.
- Calculate jitter (std deviation over sliding window).

Metrics:
- network_latency_ms
- network_jitter_ms
- packet_loss_total
- packet_loss_burst_total
- latency_p95
- latency_p99

This is critical for detecting WiFi RF instability and bufferbloat.

---

## 4. gateway-monitor

Purpose:
Isolate failure domain.

Behavior:
Continuously probe:
- Router IP (e.g., 192.168.1.1)
- External IP (e.g., 1.1.1.1)

Compare reachability to determine:

- LAN instability
- WAN instability
- Full network interruption

Metrics:
- gateway_reachable
- wan_reachable
- failure_domain_events_total

---

# Sampling Requirements

To detect 1–3 second drops:

- Default INTERVAL_SECONDS of 5 is insufficient.
- jitter-probe should support sub-second sampling.
- Sampling interval must be configurable via environment variable.
- Sliding window calculations preferred over single-sample logic.

Changes that reduce short-drop detection sensitivity are not acceptable without explicit instruction.

---

# Makefile as Deployment Guide

The Makefile (or build.sh) is the canonical interface for:

- Running locally
- Building binaries
- Building Docker images
- Importing images into k3d

Claude must:

- Keep Makefile clean and declarative.
- Avoid embedding business logic in Makefile.
- Use Makefile targets as the standard deployment interface.

Typical targets:

```bash
make run
make build-bin
make build-image
make push-k3d
make clean
```

If multiple apps exist, support:

```bash
make build APP=wifi-probe
make build APP=dns-probe
make build APP=jitter-probe
make build APP=gateway-monitor
```

The Makefile should guide deployment, not replace Kubernetes configuration.

---

# Environment Variables

All configuration must be environment-driven.

Examples:

| Variable | Description |
|----------|------------|
| PING_TARGETS | TCP targets |
| HTTP_TARGETS | HTTP targets |
| DNS_TARGETS | Domains to resolve |
| GATEWAY_IP | Router IP |
| WAN_TARGET | External IP |
| INTERVAL_SECONDS | Probe interval |
| SAMPLE_INTERVAL_MS | High-frequency sampling interval |

Do not hardcode configuration values.

---

# Development Conventions (AI-Enforced)

## 1. Repository Boundaries

- No Terraform.
- No Helm.
- No Kubernetes manifests beyond minimal examples.
- No cluster provisioning logic.

Infrastructure lives in edge-monitor-infra.

## 2. Simplicity

- Prefer Go standard library.
- Avoid frameworks.
- Avoid heavy external dependencies.

## 3. Resource Constraints

- Avoid unbounded goroutines.
- Avoid memory-heavy buffers.
- Use bounded worker pools if concurrency is introduced.
- Assume Raspberry Pi resource limits.

## 4. Logging

- Log to stdout only.
- Prefer structured JSON.
- Do not write to files.
- Do not embed S3 upload logic.

## 5. Metrics

- Avoid high-cardinality labels.
- Do not dynamically create unlimited label values.
- Keep metric names stable.

## 6. Error Handling

- Never panic in production paths.
- Return explicit errors.
- Log probe failures clearly.
- Do not suppress repeated short drop events.

---

# Observability Model

Prometheus:

- Scrapes each service’s /metrics endpoint.
- Configuration lives in edge-monitor-infra.

Logging:

- Logs emitted to stdout.
- Collected externally (Promtail, Fluent Bit, etc.).
- Possibly shipped to S3.
- Application does not directly manage log shipping.

---

# Current Priority

Primary goal:

Run all monitoring apps locally in k3d and confirm whether short WiFi drops can be objectively measured.

Development should optimize for:

- Accurate detection of 1–3 second instability.
- Low CPU and memory usage.
- Clear separation from infrastructure code.
- Clean container builds.

---

# Non-Goals

- Not a full APM platform.
- Not packet capture.
- Not deep packet inspection.
- Not a router replacement.

This is a lightweight, modular edge network observability suite.

---

End of CLAUDE.md
