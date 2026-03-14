package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port               int
	PrometheusURL      string
	PrometheusLookback time.Duration
	PrometheusTimeout  time.Duration
	LLMTimeout         time.Duration
	JobQueueSize       int
	WorkerCount        int
	MaxStoredAnalyses  int
	Backends           []BackendConfig
	MetricQueries      []MetricQuery
}

type BackendConfig struct {
	Name         string  `json:"name"`
	Type         string  `json:"type"`
	Model        string  `json:"model"`
	BaseURL      string  `json:"base_url,omitempty"`
	APIKeyEnv    string  `json:"api_key_env,omitempty"`
	Region       string  `json:"region,omitempty"`
	SystemPrompt string  `json:"system_prompt,omitempty"`
	MaxTokens    int     `json:"max_tokens,omitempty"`
	Temperature  float64 `json:"temperature,omitempty"`
}

type MetricQuery struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Query       string `json:"query"`
}

func loadConfig() (Config, error) {
	cfg := Config{
		Port:               envInt("PORT", 9094),
		PrometheusURL:      envString("PROMETHEUS_URL", "http://host.k3d.internal:9090"),
		PrometheusLookback: envDuration("PROMETHEUS_LOOKBACK", 30*time.Minute),
		PrometheusTimeout:  envDuration("PROMETHEUS_TIMEOUT", 10*time.Second),
		LLMTimeout:         envDuration("LLM_TIMEOUT", 30*time.Second),
		JobQueueSize:       envInt("JOB_QUEUE_SIZE", 32),
		WorkerCount:        envInt("WORKER_CONCURRENCY", 2),
		MaxStoredAnalyses:  envInt("MAX_STORED_ANALYSES", 25),
	}

	var err error
	cfg.Backends, err = parseBackends(envString("LLM_BACKENDS_JSON", "[]"))
	if err != nil {
		return Config{}, err
	}

	metricQueryJSON := strings.TrimSpace(os.Getenv("METRIC_QUERIES_JSON"))
	if metricQueryJSON != "" {
		cfg.MetricQueries, err = parseMetricQueries(metricQueryJSON)
		if err != nil {
			return Config{}, err
		}
	} else {
		cfg.MetricQueries = defaultMetricQueries(cfg.PrometheusLookback)
	}

	return cfg, nil
}

func parseBackends(raw string) ([]BackendConfig, error) {
	var backends []BackendConfig
	if err := json.Unmarshal([]byte(raw), &backends); err != nil {
		return nil, fmt.Errorf("parse LLM_BACKENDS_JSON: %w", err)
	}
	for i := range backends {
		backends[i].Type = strings.ToLower(strings.TrimSpace(backends[i].Type))
		if backends[i].Type == "" {
			backends[i].Type = "openai"
		}
		if backends[i].Name == "" {
			backends[i].Name = backends[i].Type
		}
		if backends[i].MaxTokens == 0 {
			backends[i].MaxTokens = 900
		}
		if backends[i].Temperature == 0 {
			backends[i].Temperature = 0.2
		}
	}
	return backends, nil
}

func parseMetricQueries(raw string) ([]MetricQuery, error) {
	var queries []MetricQuery
	if err := json.Unmarshal([]byte(raw), &queries); err != nil {
		return nil, fmt.Errorf("parse METRIC_QUERIES_JSON: %w", err)
	}
	return queries, nil
}

func defaultMetricQueries(lookback time.Duration) []MetricQuery {
	lb := promDuration(lookback)
	return []MetricQuery{
		{Name: "gateway_reachable_avg", Description: "Average gateway reachability over the lookback window", Query: fmt.Sprintf("avg_over_time(gateway_reachable{job=\"gateway-monitor\"}[%s])", lb)},
		{Name: "wan_reachable_avg", Description: "Average WAN reachability over the lookback window", Query: fmt.Sprintf("avg_over_time(wan_reachable{job=\"gateway-monitor\"}[%s])", lb)},
		{Name: "wifi_probe_up_avg", Description: "Average WiFi probe success over the lookback window", Query: fmt.Sprintf("avg_over_time(wifi_probe_up{job=\"wifi-probe\"}[%s])", lb)},
		{Name: "wifi_probe_errors", Description: "WiFi probe errors accumulated over the lookback window", Query: fmt.Sprintf("increase(wifi_probe_errors_total{job=\"wifi-probe\"}[%s])", lb)},
		{Name: "jitter_avg_ms", Description: "Average jitter in milliseconds over the lookback window", Query: fmt.Sprintf("avg_over_time(network_jitter_ms{job=\"jitter-probe\"}[%s])", lb)},
		{Name: "jitter_max_ms", Description: "Worst jitter in milliseconds over the lookback window", Query: fmt.Sprintf("max_over_time(network_jitter_ms{job=\"jitter-probe\"}[%s])", lb)},
		{Name: "latency_p99_avg_ms", Description: "Average p99 latency over the lookback window", Query: fmt.Sprintf("avg_over_time(latency_p99{job=\"jitter-probe\"}[%s])", lb)},
		{Name: "latency_p99_max_ms", Description: "Worst p99 latency over the lookback window", Query: fmt.Sprintf("max_over_time(latency_p99{job=\"jitter-probe\"}[%s])", lb)},
		{Name: "packet_loss_total", Description: "Packet loss accumulated over the lookback window", Query: fmt.Sprintf("increase(packet_loss_total{job=\"jitter-probe\"}[%s])", lb)},
		{Name: "packet_loss_bursts", Description: "Packet loss bursts accumulated over the lookback window", Query: fmt.Sprintf("increase(packet_loss_burst_total{job=\"jitter-probe\"}[%s])", lb)},
		{Name: "dns_timeouts", Description: "DNS timeouts accumulated over the lookback window", Query: fmt.Sprintf("increase(dns_probe_timeouts_total{job=\"dns-probe\"}[%s])", lb)},
		{Name: "dns_latency_avg_seconds", Description: "Average DNS latency over the lookback window", Query: fmt.Sprintf("avg_over_time(dns_probe_latency_seconds{job=\"dns-probe\"}[%s])", lb)},
		{Name: "failure_domain_events", Description: "Gateway monitor domain transitions over the lookback window", Query: fmt.Sprintf("increase(failure_domain_events_total{job=\"gateway-monitor\"}[%s])", lb)},
		{Name: "carrier_changes", Description: "Host carrier changes on likely uplink devices", Query: fmt.Sprintf("increase(node_network_carrier_changes_total{job=\"node-exporter\",device=~\"eth0|wlan0|en0\"}[%s])", lb)},
		{Name: "link_drops", Description: "Receive and transmit drops on likely uplink devices", Query: fmt.Sprintf("rate(node_network_receive_drop_total{job=\"node-exporter\",device=~\"eth0|wlan0|en0\"}[%s]) + rate(node_network_transmit_drop_total{job=\"node-exporter\",device=~\"eth0|wlan0|en0\"}[%s])", lb, lb)},
		{Name: "link_errors", Description: "Receive and transmit errors on likely uplink devices", Query: fmt.Sprintf("rate(node_network_receive_errs_total{job=\"node-exporter\",device=~\"eth0|wlan0|en0\"}[%s]) + rate(node_network_transmit_errs_total{job=\"node-exporter\",device=~\"eth0|wlan0|en0\"}[%s])", lb, lb)},
		{Name: "tcp_retransmits", Description: "TCP retransmit rate from node-exporter", Query: fmt.Sprintf("rate(node_netstat_Tcp_RetransSegs{job=\"node-exporter\"}[%s])", lb)},
		{Name: "softnet_squeezed", Description: "Softnet times squeezed rate", Query: fmt.Sprintf("sum(rate(node_softnet_times_squeezed_total{job=\"node-exporter\"}[%s]))", lb)},
		{Name: "softnet_dropped", Description: "Softnet drop rate", Query: fmt.Sprintf("sum(rate(node_softnet_dropped_total{job=\"node-exporter\"}[%s]))", lb)},
		{Name: "uplink_rx_bps", Description: "Receive throughput on likely uplink devices", Query: fmt.Sprintf("rate(node_network_receive_bytes_total{job=\"node-exporter\",device=~\"eth0|wlan0|en0\"}[%s])", lb)},
		{Name: "uplink_tx_bps", Description: "Transmit throughput on likely uplink devices", Query: fmt.Sprintf("rate(node_network_transmit_bytes_total{job=\"node-exporter\",device=~\"eth0|wlan0|en0\"}[%s])", lb)},
	}
}

func envString(key, defaultVal string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return defaultVal
}

func envInt(key string, defaultVal int) int {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return defaultVal
}

func envDuration(key string, defaultVal time.Duration) time.Duration {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return defaultVal
}

func promDuration(d time.Duration) string {
	switch {
	case d%time.Hour == 0:
		return fmt.Sprintf("%dh", int(d/time.Hour))
	case d%time.Minute == 0:
		return fmt.Sprintf("%dm", int(d/time.Minute))
	default:
		return fmt.Sprintf("%ds", int(d/time.Second))
	}
}
