# edge-monitor-app

A lightweight, modular edge network observability suite designed to detect and quantify short-lived WiFi instability.

## Problem

Home WiFi intermittently drops for 1–5 seconds — long enough to disrupt video playback, stock chart updates, and real-time services, but too brief to trigger router alarms. These drops are irregular and hard to prove without continuous monitoring.

## Approach

Four independent Go services run continuously, probing network reachability at high frequency and exposing Prometheus metrics. Together they answer:

- **Is the network up?** (wifi-probe)
- **Is DNS working?** (dns-probe)
- **Is latency stable or spiking?** (jitter-probe)
- **Is it the LAN or the WAN?** (gateway-monitor)

## Services

| Service | Port | Purpose |
|---------|------|---------|
| [wifi-probe](wifi-probe/) | 9090 | TCP and HTTP reachability with latency tracking |
| [dns-probe](dns-probe/) | 9091 | DNS resolution monitoring with timeout detection |
| [jitter-probe](jitter-probe/) | 9092 | High-frequency latency sampling with jitter, p95/p99, and burst detection |
| [gateway-monitor](gateway-monitor/) | 9093 | LAN vs WAN failure domain isolation |

Each service is an independent Go binary with its own module, Dockerfile, and Makefile.

## Quick Start

Run any service locally (requires Go 1.22+):

```bash
cd wifi-probe
make run
```

In another terminal:

```bash
curl http://localhost:9090/metrics
```

### Running all services

```bash
# Terminal 1
cd wifi-probe && make run

# Terminal 2
cd dns-probe && make run

# Terminal 3
cd jitter-probe && make run

# Terminal 4
cd gateway-monitor && make run
```

## Building

Each service supports the same Makefile targets:

```bash
make build-bin          # Build for host OS/arch (local testing)
make build-linux-amd64  # Cross-compile for linux/amd64
make build-linux-arm64  # Cross-compile for linux/arm64 (Raspberry Pi)
make build-all          # Build both linux architectures
make build-image        # Build Docker image
make push-k3d           # Build and import into k3d cluster
make clean              # Remove built binaries
```

## Configuration

All services are configured via environment variables. No hardcoded values.

| Variable | Service(s) | Description | Default |
|----------|-----------|-------------|---------|
| `PING_TARGETS` | wifi-probe, jitter-probe | TCP targets (comma-separated) | `192.168.1.1,1.1.1.1` |
| `HTTP_TARGETS` | wifi-probe | HTTP URLs to probe | `https://ifconfig.me/ip` |
| `DNS_TARGETS` | dns-probe | Domains to resolve | `google.com,cloudflare.com` |
| `GATEWAY_IP` | gateway-monitor | Router IP address | `192.168.1.1` |
| `WAN_TARGET` | gateway-monitor | External IP to test WAN | `1.1.1.1` |
| `INTERVAL_SECONDS` | wifi-probe, dns-probe, gateway-monitor | Probe interval | `2` |
| `SAMPLE_INTERVAL_MS` | jitter-probe | Sampling interval in ms | `500` |
| `WINDOW_SIZE` | jitter-probe | Sliding window size | `60` |

## Metrics

### wifi-probe

| Metric | Type | Description |
|--------|------|-------------|
| `wifi_probe_up` | Gauge | 1 if target reachable, 0 if not |
| `wifi_probe_latency_seconds` | Gauge | Probe latency |
| `wifi_probe_runs_total` | Counter | Total probe executions |
| `wifi_probe_errors_total` | Counter | Total probe failures |

### dns-probe

| Metric | Type | Description |
|--------|------|-------------|
| `dns_probe_up` | Gauge | 1 if DNS resolution succeeded |
| `dns_probe_latency_seconds` | Gauge | Resolution latency |
| `dns_probe_timeouts_total` | Counter | DNS timeout count |

### jitter-probe

| Metric | Type | Description |
|--------|------|-------------|
| `network_latency_ms` | Gauge | Latest sample latency in ms |
| `network_jitter_ms` | Gauge | Std deviation of latencies in sliding window |
| `packet_loss_total` | Counter | Total failed probes |
| `packet_loss_burst_total` | Counter | Burst events (2+ consecutive failures) |
| `latency_p95` | Gauge | 95th percentile latency in ms |
| `latency_p99` | Gauge | 99th percentile latency in ms |

### gateway-monitor

| Metric | Type | Description |
|--------|------|-------------|
| `gateway_reachable` | Gauge | 1 if router is reachable |
| `wan_reachable` | Gauge | 1 if external target is reachable |
| `failure_domain_events_total` | Counter | Failure transitions (labels: `lan`, `wan`, `full`) |

## Architecture

- **Language:** Go 1.22, standard library preferred
- **Logging:** Structured JSON via `log/slog` to stdout
- **Metrics:** Prometheus client library (`/metrics` endpoint per service)
- **Probing:** TCP dial (no ICMP — runs unprivileged)
- **Containers:** Multi-stage Docker builds, distroless runtime images
- **Target platform:** Raspberry Pi (arm64) and x86_64

Infrastructure configuration (Kubernetes manifests, Prometheus scrape configs, Helm charts) lives in the separate [edge-monitor-infra](https://github.com/hegarty/edge-monitor-infra) repository.
