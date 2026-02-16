package main

import "github.com/prometheus/client_golang/prometheus"

var (
	gatewayReachable = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "gateway_reachable",
			Help: "Gateway (router) reachability: 1 = up, 0 = down",
		},
	)

	wanReachable = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "wan_reachable",
			Help: "WAN target reachability: 1 = up, 0 = down",
		},
	)

	failureDomainEventsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "failure_domain_events_total",
			Help: "Total failure domain transition events",
		},
		[]string{"domain"},
	)
)

func registerMetrics() {
	prometheus.MustRegister(
		gatewayReachable,
		wanReachable,
		failureDomainEventsTotal,
	)
}
