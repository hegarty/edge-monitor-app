package main

import (
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func envList(key string) []string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return nil
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func envInt(key string, defaultVal int) int {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return defaultVal
	}
	return n
}

// targetState tracks per-target probe state for burst detection.
type targetState struct {
	window           *Window
	consecutiveFails int
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	registerMetrics()

	targets := envList("PING_TARGETS")
	sampleIntervalMs := envInt("SAMPLE_INTERVAL_MS", 500)
	windowSize := envInt("WINDOW_SIZE", 60)

	if len(targets) == 0 {
		slog.Error("PING_TARGETS is required")
		os.Exit(1)
	}

	slog.Info("starting jitter-probe",
		"targets", targets,
		"sample_interval_ms", sampleIntervalMs,
		"window_size", windowSize,
	)

	interval := time.Duration(sampleIntervalMs) * time.Millisecond
	timeout := 2 * time.Second

	// Initialize per-target state.
	states := make(map[string]*targetState, len(targets))
	for _, t := range targets {
		states[t] = &targetState{
			window: NewWindow(windowSize),
		}
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			for _, target := range targets {
				st := states[target]
				ok, latency, err := tcpProbe(target, timeout)

				if ok {
					latencyMs := float64(latency.Nanoseconds()) / 1e6

					// If we were in a burst (2+ consecutive failures), record it.
					if st.consecutiveFails >= 2 {
						packetLossBurstTotal.WithLabelValues(target).Inc()
						slog.Warn("packet loss burst ended",
							"target", target,
							"consecutive_failures", st.consecutiveFails,
						)
					}
					st.consecutiveFails = 0

					st.window.Add(latencyMs)

					networkLatency.WithLabelValues(target).Set(latencyMs)
					networkJitter.WithLabelValues(target).Set(st.window.StdDev())
					latencyP95.WithLabelValues(target).Set(st.window.Percentile(95))
					latencyP99.WithLabelValues(target).Set(st.window.Percentile(99))
				} else {
					packetLossTotal.WithLabelValues(target).Inc()
					st.consecutiveFails++

					if err != nil {
						slog.Warn("tcp probe failed",
							"target", target,
							"error", err,
							"consecutive_failures", st.consecutiveFails,
						)
					}
				}
			}
		}
	}()

	http.Handle("/metrics", promhttp.Handler())
	slog.Info("metrics server listening", "addr", ":9092", "path", "/metrics")
	if err := http.ListenAndServe(":9092", nil); err != nil {
		slog.Error("metrics server failed", "error", err)
		os.Exit(1)
	}
}
