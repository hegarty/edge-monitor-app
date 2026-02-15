package main

import (
	"fmt"
	"net"
	"time"
)

func tcpProbe(host string, timeout time.Duration) (bool, time.Duration, error) {
	addr := fmt.Sprintf("%s:%d", host, 443)
	start := time.Now()
	conn, err := net.DialTimeout("tcp", addr, timeout)
	latency := time.Since(start)

	if err != nil {
		return false, 0, fmt.Errorf("tcp dial %s: %w", addr, err)
	}
	conn.Close()
	return true, latency, nil
}
