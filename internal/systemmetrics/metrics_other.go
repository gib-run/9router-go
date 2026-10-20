//go:build !linux

package systemmetrics

import (
	"runtime"
	"time"
)

type Snapshot struct {
	ProcessCPUPercent     float64   `json:"processCpuPercent"`
	ProcessMemoryBytes    uint64    `json:"processMemoryBytes"`
	ProcessMemoryPercent  float64   `json:"processMemoryPercent"`
	SystemCPUPercent      float64   `json:"systemCpuPercent"`
	SystemMemoryBytes     uint64    `json:"systemMemoryBytes"`
	SystemMemoryUsedBytes uint64    `json:"systemMemoryUsedBytes"`
	SystemMemoryPercent   float64   `json:"systemMemoryPercent"`
	Goroutines            int       `json:"goroutines"`
	SampledAt             time.Time `json:"sampledAt"`
}

func Read() Snapshot {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return Snapshot{ProcessMemoryBytes: m.Sys, Goroutines: runtime.NumGoroutine(), SampledAt: time.Now()}
}
