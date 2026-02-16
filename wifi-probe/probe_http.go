package main

import (
    "net/http"
    "time"
)

func httpProbe(url string, timeout time.Duration) (bool, time.Duration, error) {
    client := http.Client{
        Timeout: timeout,
    }

    start := time.Now()
    resp, err := client.Get(url)
    latency := time.Since(start)

    if err != nil {
        return false, 0, err
    }
    defer resp.Body.Close()

    if resp.StatusCode >= 200 && resp.StatusCode < 400 {
        return true, latency, nil
    }

    return false, latency, nil
}
