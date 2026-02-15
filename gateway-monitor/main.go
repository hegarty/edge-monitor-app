package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func boolToFloat(b bool) float64 {
	if b {
		return 1
	}
	return 0
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	registerMetrics()

	gatewayIP := envOrDefault("GATEWAY_IP", "192.168.1.1")
	wanTarget := envOrDefault("WAN_TARGET", "1.1.1.1")

	interval := 2 * time.Second
	if v := os.Getenv("INTERVAL_SECONDS"); v != "" {
		if d, err := time.ParseDuration(v + "s"); err == nil {
			interval = d
		}
	}

	probePorts := []int{443, 80}
	probeTimeout := 2 * time.Second

	slog.Info("starting gateway-monitor",
		"gateway_ip", gatewayIP,
		"wan_target", wanTarget,
		"interval", interval.String(),
	)

	go func() {
		prevGatewayUp := true
		prevWanUp := true

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			<-ticker.C

			gwUp, gwLatency, gwErr := tcpProbe(gatewayIP, probePorts, probeTimeout)
			gatewayReachable.Set(boolToFloat(gwUp))

			if gwUp {
				slog.Debug("gateway probe succeeded", "target", gatewayIP, "latency", gwLatency.String())
			} else {
				slog.Warn("gateway probe failed", "target", gatewayIP, "error", gwErr)
			}

			wUp, wLatency, wErr := tcpProbe(wanTarget, probePorts, probeTimeout)
			wanReachable.Set(boolToFloat(wUp))

			if wUp {
				slog.Debug("wan probe succeeded", "target", wanTarget, "latency", wLatency.String())
			} else {
				slog.Warn("wan probe failed", "target", wanTarget, "error", wErr)
			}

			// Detect state transitions into failure
			gwTransitionDown := prevGatewayUp && !gwUp
			wanTransitionDown := prevWanUp && !wUp

			if gwTransitionDown && wanTransitionDown {
				failureDomainEventsTotal.WithLabelValues("full").Inc()
				slog.Error("failure domain: full network interruption",
					"gateway", gatewayIP, "wan", wanTarget)
			} else if gwTransitionDown && !wanTransitionDown {
				// Gateway just went down, WAN was already down or is still up
				if wUp {
					failureDomainEventsTotal.WithLabelValues("lan").Inc()
					slog.Error("failure domain: LAN instability",
						"gateway", gatewayIP)
				} else {
					// Both are now down but WAN went down earlier
					failureDomainEventsTotal.WithLabelValues("full").Inc()
					slog.Error("failure domain: full network interruption (gateway joined)",
						"gateway", gatewayIP, "wan", wanTarget)
				}
			} else if wanTransitionDown && !gwTransitionDown {
				// WAN just went down, gateway was already down or is still up
				if gwUp {
					failureDomainEventsTotal.WithLabelValues("wan").Inc()
					slog.Error("failure domain: WAN instability",
						"wan", wanTarget)
				} else {
					// Both are now down but gateway went down earlier
					failureDomainEventsTotal.WithLabelValues("full").Inc()
					slog.Error("failure domain: full network interruption (wan joined)",
						"gateway", gatewayIP, "wan", wanTarget)
				}
			}

			prevGatewayUp = gwUp
			prevWanUp = wUp
		}
	}()

	http.Handle("/metrics", promhttp.Handler())
	slog.Info("metrics server listening", "addr", ":9093", "path", "/metrics")
	if err := http.ListenAndServe(":9093", nil); err != nil {
		slog.Error("metrics server failed", "error", err)
		os.Exit(1)
	}
}
