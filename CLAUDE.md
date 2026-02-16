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
- Makefiles (one per service)
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

This repository contains multiple independent monitoring services, each compiled as a separate binary.

Repository structure:

```
/wifi-probe       — TCP and HTTP reachability prober (:9090)
/dns-probe        — DNS resolution prober (:9091)
/jitter-probe     — High-frequency latency and jitter sampler (:9092)
/gateway-monitor  — LAN vs WAN failure domain isolator (:9093)
```

Each service:

- Is an independent Go module and binary.
- Has its own Makefile, Dockerfile, go.mod, and go.sum.
- Exposes Prometheus metrics on a dedicated port.
- Uses structured JSON logging via `log/slog` (Go stdlib).
- Logs to stdout only.
- Is deployable independently in Kubernetes.
- Supports multi-arch builds (linux/amd64 and linux/arm64).

Do not merge services into a monolithic application.

---

# Implemented Services

All services are implemented and verified.

---

## 1. wifi-probe (port 9090)

Purpose:
Detect basic TCP and HTTP reachability failures.

Behavior:
- Probe TCP targets (tries ports 443, 80).
- Probe HTTP targets.
- Measure latency.
- Detect connection failures.

Metrics:
- wifi_probe_up
- wifi_probe_latency_seconds
- wifi_probe_runs_total
- wifi_probe_errors_total

---

## 2. dns-probe (port 9091)

Purpose:
Detect DNS resolution failures and latency spikes.

Behavior:
- Resolve configurable domains repeatedly using Go stdlib net.Resolver.
- Measure lookup latency.
- Track timeout events.

Metrics:
- dns_probe_up
- dns_probe_latency_seconds
- dns_probe_timeouts_total

This helps identify DNS-related micro-outages.

---

## 3. jitter-probe (port 9092)

Purpose:
Detect short (1–3 second) latency spikes and packet loss bursts.

Behavior:
- High-frequency TCP sampling (default 500ms interval).
- Track rolling latency via bounded ring buffer sliding window.
- Track packet loss bursts (2+ consecutive failures).
- Calculate jitter (std deviation over sliding window).
- Calculate p95 and p99 latency percentiles.

Metrics:
- network_latency_ms
- network_jitter_ms
- packet_loss_total
- packet_loss_burst_total
- latency_p95
- latency_p99

This is critical for detecting WiFi RF instability and bufferbloat.

---

## 4. gateway-monitor (port 9093)

Purpose:
Isolate failure domain.

Behavior:
Continuously probe via TCP:
- Router IP (e.g., 192.168.1.1)
- External IP (e.g., 1.1.1.1)

Compare reachability to determine failure domain on state transitions:

- LAN instability (gateway down, WAN up)
- WAN instability (gateway up, WAN down)
- Full network interruption (both down)

Metrics:
- gateway_reachable
- wan_reachable
- failure_domain_events_total (labels: domain=lan|wan|full)

---

# Sampling Requirements

To detect 1–3 second drops:

- Default INTERVAL_SECONDS is 2 for all probes (5 is insufficient).
- jitter-probe uses SAMPLE_INTERVAL_MS (default 500) for sub-second sampling.
- Sampling interval must be configurable via environment variable.
- jitter-probe uses a bounded ring buffer sliding window for calculations.

Changes that reduce short-drop detection sensitivity are not acceptable without explicit instruction.

---

# Makefile as Deployment Guide

Each service has its own Makefile in its directory. The Makefile is the canonical interface for:

- Running locally
- Building binaries (host OS/arch and cross-compile)
- Building Docker images (multi-arch)
- Importing images into k3d

Claude must:

- Keep Makefiles clean and declarative.
- Avoid embedding business logic in Makefiles.
- Use Makefile targets as the standard deployment interface.

Standard targets (run from within each service directory):

```bash
make run                # Run locally with default env vars
make build-bin          # Build binary for host OS/arch
make build-linux-amd64  # Cross-compile for linux/amd64
make build-linux-arm64  # Cross-compile for linux/arm64
make build-all          # Build both linux/amd64 and linux/arm64
make build-image        # Build Docker image for host arch
make build-image-amd64  # Build Docker image for linux/amd64
make build-image-arm64  # Build Docker image for linux/arm64
make build-image-all    # Build Docker images for both architectures
make push-k3d           # Import image into k3d cluster
make clean              # Remove built binaries
```

The Makefile should guide deployment, not replace Kubernetes configuration.

---

# Environment Variables

All configuration must be environment-driven.

| Variable | Used by | Description | Default |
|----------|---------|-------------|---------|
| PING_TARGETS | wifi-probe, jitter-probe | TCP targets (comma-separated) | 192.168.1.1,1.1.1.1 |
| HTTP_TARGETS | wifi-probe | HTTP targets (comma-separated) | https://ifconfig.me/ip |
| DNS_TARGETS | dns-probe | Domains to resolve (comma-separated) | google.com,cloudflare.com |
| GATEWAY_IP | gateway-monitor | Router IP | 192.168.1.1 |
| WAN_TARGET | gateway-monitor | External IP | 1.1.1.1 |
| INTERVAL_SECONDS | wifi-probe, dns-probe, gateway-monitor | Probe interval in seconds | 2 |
| SAMPLE_INTERVAL_MS | jitter-probe | High-frequency sampling interval in ms | 500 |
| WINDOW_SIZE | jitter-probe | Sliding window size for jitter/percentile | 60 |

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
- Only external dependency: github.com/prometheus/client_golang.

## 3. Resource Constraints

- Avoid unbounded goroutines.
- Avoid memory-heavy buffers.
- Use bounded data structures (ring buffers, fixed-size windows).
- Use bounded worker pools if concurrency is introduced.
- Assume Raspberry Pi resource limits.

## 4. Logging

- Log to stdout only.
- Use `log/slog` with `slog.NewJSONHandler` for structured JSON output.
- Do not use `log.Println` or `fmt.Printf` for application logging.
- Do not write to files.
- Do not embed S3 upload logic.

## 5. Metrics

- Avoid high-cardinality labels.
- Do not dynamically create unlimited label values.
- Keep metric names stable.
- Each service exposes metrics on its own port (9090–9093).

## 6. Error Handling

- Never panic in production paths.
- Return explicit errors.
- Log probe failures clearly.
- Do not suppress repeated short drop events.

## 7. Probing

- Use TCP dial (net.DialTimeout) for network probing, not ICMP.
- ICMP requires root/elevated privileges and is not suitable for unprivileged containers.

---

# Observability Model

Prometheus:

- Scrapes each service's /metrics endpoint.
- Configuration lives in edge-monitor-infra.

Service ports:

| Service | Metrics Port |
|---------|-------------|
| wifi-probe | 9090 |
| dns-probe | 9091 |
| jitter-probe | 9092 |
| gateway-monitor | 9093 |

Logging:

- Logs emitted to stdout as structured JSON.
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
