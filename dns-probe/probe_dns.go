package main

import (
	"context"
	"net"
	"time"
)

// dnsProbe resolves the given domain using net.Resolver with a context deadline.
// Returns success, latency, and any error encountered.
func dnsProbe(domain string, timeout time.Duration) (bool, time.Duration, error) {
	resolver := &net.Resolver{}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	start := time.Now()
	addrs, err := resolver.LookupHost(ctx, domain)
	latency := time.Since(start)

	if err != nil {
		return false, latency, err
	}

	if len(addrs) == 0 {
		return false, latency, nil
	}

	return true, latency, nil
}
