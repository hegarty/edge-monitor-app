package main

import (
    "fmt"
    "net"
    "time"
)

var tcpPorts = []int{443, 80}

func tcpProbe(host string, timeout time.Duration) (bool, time.Duration, error) {
    for _, port := range tcpPorts {
        addr := fmt.Sprintf("%s:%d", host, port)
        start := time.Now()
        conn, err := net.DialTimeout("tcp", addr, timeout)
        latency := time.Since(start)

        if err == nil {
            conn.Close()
            return true, latency, nil
        }
    }
    return false, 0, fmt.Errorf("no tcp ports reachable")
}
