//go:build linux

package systemmetrics

import (
	"bufio"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

type Snapshot struct {
	ProcessCPUPercent        float64   `json:"processCpuPercent"`
	ProcessMemoryBytes       uint64    `json:"processMemoryBytes"`
	ProcessMemoryPercent     float64   `json:"processMemoryPercent"`
	SystemCPUPercent         float64   `json:"systemCpuPercent"`
	SystemMemoryBytes        uint64    `json:"systemMemoryBytes"`
	SystemMemoryUsedBytes    uint64    `json:"systemMemoryUsedBytes"`
	SystemMemoryPercent      float64   `json:"systemMemoryPercent"`
	SystemDiskBytes          uint64    `json:"systemDiskBytes"`
	SystemDiskUsedBytes      uint64    `json:"systemDiskUsedBytes"`
	SystemDiskAvailableBytes uint64    `json:"systemDiskAvailableBytes"`
	SystemDiskPercent        float64   `json:"systemDiskPercent"`
	Goroutines               int       `json:"goroutines"`
	SampledAt                time.Time `json:"sampledAt"`
}

func readDisk() (uint64, uint64, uint64) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs("/", &stat); err != nil {
		return 0, 0, 0
	}
	blockSize := uint64(stat.Bsize)
	total := stat.Blocks * blockSize
	used := (stat.Blocks - min(stat.Blocks, stat.Bfree)) * blockSize
	available := stat.Bavail * blockSize
	return total, used, available
}

var sampleMu sync.Mutex
var previous struct {
	at                time.Time
	proc, total, idle uint64
}

func readProcTicks() (uint64, uint64, uint64) {
	b, err := os.ReadFile("/proc/self/stat")
	if err != nil {
		return 0, 0, 0
	}
	end := strings.LastIndexByte(string(b), ')')
	if end < 0 {
		return 0, 0, 0
	}
	f := strings.Fields(string(b)[end+1:])
	if len(f) < 22 {
		return 0, 0, 0
	}
	user, _ := strconv.ParseUint(f[11], 10, 64)
	sys, _ := strconv.ParseUint(f[12], 10, 64)
	rssPages, _ := strconv.ParseUint(f[21], 10, 64)
	return user + sys, rssPages * uint64(os.Getpagesize()), 0
}

func readSystemTicks() (uint64, uint64) {
	f, err := os.Open("/proc/stat")
	if err != nil {
		return 0, 0
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	if !s.Scan() {
		return 0, 0
	}
	fields := strings.Fields(s.Text())
	var total uint64
	values := make([]uint64, 0, len(fields)-1)
	for _, field := range fields[1:] {
		n, _ := strconv.ParseUint(field, 10, 64)
		values = append(values, n)
		total += n
	}
	var idle uint64
	if len(values) > 3 {
		idle = values[3]
		if len(values) > 4 {
			idle += values[4]
		}
	}
	return total, idle
}

func readMemory() (uint64, uint64) {
	// Prefer the service's cgroup memory limit when running in a container; in
	// that case /proc/meminfo describes the host, not the VPS allocation.
	if limit, limitErr := readCgroupUint("/sys/fs/cgroup/memory.max"); limitErr == nil && limit > 0 {
		if current, currentErr := readCgroupUint("/sys/fs/cgroup/memory.current"); currentErr == nil {
			return limit, limit - min(limit, current)
		}
	}
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0, 0
	}
	defer f.Close()
	var total, available uint64
	s := bufio.NewScanner(f)
	for s.Scan() {
		fields := strings.Fields(s.Text())
		if len(fields) < 2 {
			continue
		}
		n, _ := strconv.ParseUint(fields[1], 10, 64)
		switch fields[0] {
		case "MemTotal:":
			total = n * 1024
		case "MemAvailable:":
			available = n * 1024
		}
	}
	return total, available
}

func readCgroupUint(path string) (uint64, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	value := strings.TrimSpace(string(b))
	if value == "max" {
		return 0, os.ErrNotExist
	}
	return strconv.ParseUint(value, 10, 64)
}

func Read() Snapshot {
	now := time.Now()
	proc, rss, _ := readProcTicks()
	total, idle := readSystemTicks()
	memTotal, memAvailable := readMemory()
	diskTotal, diskUsed, diskAvailable := readDisk()
	result := Snapshot{
		ProcessMemoryBytes: rss, SystemMemoryBytes: memTotal,
		SystemDiskBytes: diskTotal, SystemDiskUsedBytes: diskUsed,
		SystemDiskAvailableBytes: diskAvailable,
		Goroutines:               runtime.NumGoroutine(), SampledAt: now,
	}
	if diskTotal > 0 {
		result.SystemDiskPercent = float64(diskUsed) * 100 / float64(diskTotal)
	}
	if memTotal > 0 {
		result.SystemMemoryUsedBytes = memTotal - memAvailable
		result.SystemMemoryPercent = float64(result.SystemMemoryUsedBytes) * 100 / float64(memTotal)
	}
	sampleMu.Lock()
	if !previous.at.IsZero() {
		if total > previous.total && idle >= previous.idle {
			result.SystemCPUPercent = float64(total-previous.total-idle+previous.idle) * 100 / float64(total-previous.total)
		}
		if proc >= previous.proc {
			result.ProcessCPUPercent = float64(proc-previous.proc) * 100 / 100 / now.Sub(previous.at).Seconds()
		}
	}
	previous.at, previous.proc, previous.total, previous.idle = now, proc, total, idle
	sampleMu.Unlock()
	if memTotal > 0 {
		result.ProcessMemoryPercent = float64(rss) * 100 / float64(memTotal)
	}
	return result
}
