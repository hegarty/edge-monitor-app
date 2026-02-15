package main

import "github.com/prometheus/client_golang/prometheus"

var (
    probeUp = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "wifi_probe_up",
            Help: "Probe success (1) or failure (0)",
        },
        []string{"probe", "target"},
    )

    probeLatency = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "wifi_probe_latency_seconds",
            Help: "Probe latency in seconds",
        },
        []string{"probe", "target"},
    )

    probeRuns = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "wifi_probe_runs_total",
            Help: "Total number of probe executions",
        },
        []string{"probe", "target"},
    )

    probeErrors = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "wifi_probe_errors_total",
            Help: "Total number of probe errors",
        },
        []string{"probe", "target"},
    )
)

func registerMetrics() {
    prometheus.MustRegister(
        probeUp,
        probeLatency,
        probeRuns,
        probeErrors,
    )
}
