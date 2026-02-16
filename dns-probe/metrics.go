package main

import "github.com/prometheus/client_golang/prometheus"

var (
	probeUp = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "dns_probe_up",
			Help: "DNS probe success (1) or failure (0)",
		},
		[]string{"target"},
	)

	probeLatency = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "dns_probe_latency_seconds",
			Help: "DNS probe latency in seconds",
		},
		[]string{"target"},
	)

	probeTimeouts = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "dns_probe_timeouts_total",
			Help: "Total number of DNS probe timeouts",
		},
		[]string{"target"},
	)
)

func registerMetrics() {
	prometheus.MustRegister(
		probeUp,
		probeLatency,
		probeTimeouts,
	)
}
