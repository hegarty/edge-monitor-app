package main

import (
	"log/slog"
	"net/http"
	"os"
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

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	registerMetrics()

	interval := 5 * time.Second
	if v := os.Getenv("INTERVAL_SECONDS"); v != "" {
		if d, err := time.ParseDuration(v + "s"); err == nil {
			interval = d
		}
	}

	tcpTargets := envList("PING_TARGETS")
	httpTargets := envList("HTTP_TARGETS")

	slog.Info("starting wifi-probe",
		"tcp_targets", tcpTargets,
		"http_targets", httpTargets,
		"interval", interval.String(),
	)

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			<-ticker.C

			for _, t := range tcpTargets {
				probeRuns.WithLabelValues("tcp", t).Inc()

				ok, latency, err := tcpProbe(t, 2*time.Second)
				probeUp.WithLabelValues("tcp", t).Set(boolToFloat(ok))

				if ok {
					probeLatency.WithLabelValues("tcp", t).Set(latency.Seconds())
				} else {
					probeErrors.WithLabelValues("tcp", t).Inc()
					if err != nil {
						slog.Warn("tcp probe failed", "target", t, "error", err)
					}
				}
			}

			for _, u := range httpTargets {
				probeRuns.WithLabelValues("http", u).Inc()

				ok, latency, err := httpProbe(u, 3*time.Second)
				probeUp.WithLabelValues("http", u).Set(boolToFloat(ok))

				if ok {
					probeLatency.WithLabelValues("http", u).Set(latency.Seconds())
				} else {
					probeErrors.WithLabelValues("http", u).Inc()
					if err != nil {
						slog.Warn("http probe failed", "target", u, "error", err)
					}
				}
			}
		}
	}()

	http.Handle("/metrics", promhttp.Handler())
	slog.Info("metrics server listening", "addr", ":9090", "path", "/metrics")
	if err := http.ListenAndServe(":9090", nil); err != nil {
		slog.Error("metrics server failed", "error", err)
		os.Exit(1)
	}
}

func boolToFloat(b bool) float64 {
	if b {
		return 1
	}
	return 0
}
