package main

import "github.com/prometheus/client_golang/prometheus"

var (
	alertsReceivedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "alert_receiver_alerts_received_total",
			Help: "Total number of Grafana webhook payloads received",
		},
		[]string{"status"},
	)

	queueDepthGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "alert_receiver_queue_depth",
			Help: "Current number of queued alert analysis jobs",
		},
	)

	jobResultsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "alert_receiver_jobs_total",
			Help: "Total number of alert analysis jobs by result",
		},
		[]string{"result"},
	)

	jobDurationSeconds = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "alert_receiver_job_duration_seconds",
			Help:    "Time spent enriching and dispatching an alert analysis job",
			Buckets: prometheus.DefBuckets,
		},
	)

	providerRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "alert_receiver_provider_requests_total",
			Help: "Total LLM provider requests by provider and result",
		},
		[]string{"provider", "result"},
	)

	prometheusQueriesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "alert_receiver_prometheus_queries_total",
			Help: "Total Prometheus enrichment queries by query name and result",
		},
		[]string{"query", "result"},
	)
)

func registerMetrics() {
	prometheus.MustRegister(
		alertsReceivedTotal,
		queueDepthGauge,
		jobResultsTotal,
		jobDurationSeconds,
		providerRequestsTotal,
		prometheusQueriesTotal,
	)
}
