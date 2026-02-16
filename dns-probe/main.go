package main

import (
	"context"
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

	interval := 2 * time.Second
	if v := os.Getenv("INTERVAL_SECONDS"); v != "" {
		if d, err := time.ParseDuration(v + "s"); err == nil {
			interval = d
		}
	}

	dnsTargets := envList("DNS_TARGETS")

	slog.Info("starting dns-probe",
		"dns_targets", dnsTargets,
		"interval", interval.String(),
	)

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			<-ticker.C

			for _, domain := range dnsTargets {
				ok, latency, err := dnsProbe(domain, 2*time.Second)

				if ok {
					probeUp.WithLabelValues(domain).Set(1)
					probeLatency.WithLabelValues(domain).Set(latency.Seconds())
				} else {
					probeUp.WithLabelValues(domain).Set(0)

					if err != nil {
						// Check if the error is a timeout
						if isTimeout(err) {
							probeTimeouts.WithLabelValues(domain).Inc()
							slog.Warn("dns probe timed out", "target", domain, "error", err)
						} else {
							slog.Warn("dns probe failed", "target", domain, "error", err)
						}
					}
				}
			}
		}
	}()

	http.Handle("/metrics", promhttp.Handler())
	slog.Info("metrics server listening", "addr", ":9091", "path", "/metrics")
	if err := http.ListenAndServe(":9091", nil); err != nil {
		slog.Error("metrics server failed", "error", err)
		os.Exit(1)
	}
}

// isTimeout checks whether the error is a context deadline exceeded or timeout.
func isTimeout(err error) bool {
	if err == context.DeadlineExceeded {
		return true
	}
	// net errors may wrap a timeout
	type timeouter interface {
		Timeout() bool
	}
	if t, ok := err.(timeouter); ok {
		return t.Timeout()
	}
	return false
}
