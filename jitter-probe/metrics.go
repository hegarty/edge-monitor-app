package main

import "github.com/prometheus/client_golang/prometheus"

var (
	networkLatency = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "network_latency_ms",
			Help: "Latest TCP probe latency in milliseconds",
		},
		[]string{"target"},
	)

	networkJitter = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "network_jitter_ms",
			Help: "Standard deviation of latencies in sliding window (ms)",
		},
		[]string{"target"},
	)

	packetLossTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "packet_loss_total",
			Help: "Total number of failed TCP probes",
		},
		[]string{"target"},
	)

	packetLossBurstTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "packet_loss_burst_total",
			Help: "Total number of packet loss bursts (2+ consecutive failures)",
		},
		[]string{"target"},
	)

	latencyP95 = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "latency_p95",
			Help: "95th percentile latency in sliding window (ms)",
		},
		[]string{"target"},
	)

	latencyP99 = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "latency_p99",
			Help: "99th percentile latency in sliding window (ms)",
		},
		[]string{"target"},
	)
)

func registerMetrics() {
	prometheus.MustRegister(
		networkLatency,
		networkJitter,
		packetLossTotal,
		packetLossBurstTotal,
		latencyP95,
		latencyP99,
	)
}
