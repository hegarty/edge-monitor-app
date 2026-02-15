package main

import (
	"math"
	"sort"
)

// Window is a fixed-size ring buffer for latency samples in milliseconds.
type Window struct {
	data  []float64
	pos   int
	count int
	cap   int
}

// NewWindow creates a ring buffer with the given capacity.
func NewWindow(capacity int) *Window {
	return &Window{
		data: make([]float64, capacity),
		cap:  capacity,
	}
}

// Add inserts a latency sample (in milliseconds) into the ring buffer.
func (w *Window) Add(latencyMs float64) {
	w.data[w.pos] = latencyMs
	w.pos = (w.pos + 1) % w.cap
	if w.count < w.cap {
		w.count++
	}
}

// Len returns the number of samples currently in the window.
func (w *Window) Len() int {
	return w.count
}

// values returns a copy of the current samples.
func (w *Window) values() []float64 {
	if w.count == 0 {
		return nil
	}
	out := make([]float64, w.count)
	if w.count < w.cap {
		copy(out, w.data[:w.count])
	} else {
		// Ring buffer is full; read from pos (oldest) to end, then start to pos.
		n := copy(out, w.data[w.pos:])
		copy(out[n:], w.data[:w.pos])
	}
	return out
}

// StdDev calculates the population standard deviation of the samples.
func (w *Window) StdDev() float64 {
	if w.count < 2 {
		return 0
	}
	vals := w.values()
	mean := 0.0
	for _, v := range vals {
		mean += v
	}
	mean /= float64(len(vals))

	variance := 0.0
	for _, v := range vals {
		d := v - mean
		variance += d * d
	}
	variance /= float64(len(vals))
	return math.Sqrt(variance)
}

// Percentile calculates the p-th percentile (0-100) using nearest-rank method.
func (w *Window) Percentile(p float64) float64 {
	if w.count == 0 {
		return 0
	}
	vals := w.values()
	sort.Float64s(vals)

	rank := (p / 100.0) * float64(len(vals))
	idx := int(math.Ceil(rank)) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(vals) {
		idx = len(vals) - 1
	}
	return vals[idx]
}
