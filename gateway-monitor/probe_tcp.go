package main

import (
	"fmt"
	"net"
	"time"
)

func tcpProbe(host string, ports []int, timeout time.Duration) (bool, time.Duration, error) {
	for _, port := range ports {
		addr := fmt.Sprintf("%s:%d", host, port)
		start := time.Now()
		conn, err := net.DialTimeout("tcp", addr, timeout)
		latency := time.Since(start)

		if err == nil {
			conn.Close()
			return true, latency, nil
		}
	}
	return false, 0, fmt.Errorf("no tcp ports reachable on %s", host)
}
