package benchmark

import (
	"fmt"
	"runtime"
	"sync"
	"time"
	
	tea "github.com/charmbracelet/bubbletea"
)

// PerformanceMonitor tracks TUI performance metrics
type PerformanceMonitor struct {
	metrics   map[string]*Metric
	mutex     sync.RWMutex
	startTime time.Time
}

// Metric represents a performance metric
type Metric struct {
	Name         string
	Count        int64
	TotalTime    time.Duration
	MinTime      time.Duration
	MaxTime      time.Duration
	LastRecorded time.Time
}

// NewPerformanceMonitor creates a new performance monitor
func NewPerformanceMonitor() *PerformanceMonitor {
	return &PerformanceMonitor{
		metrics:   make(map[string]*Metric),
		startTime: time.Now(),
	}
}

// StartTimer starts timing an operation
func (pm *PerformanceMonitor) StartTimer(name string) *Timer {
	return &Timer{
		name:    name,
		start:   time.Now(),
		monitor: pm,
	}
}

// RecordDuration records a duration for a metric
func (pm *PerformanceMonitor) RecordDuration(name string, duration time.Duration) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	
	metric, exists := pm.metrics[name]
	if !exists {
		metric = &Metric{
			Name:    name,
			MinTime: duration,
			MaxTime: duration,
		}
		pm.metrics[name] = metric
	}
	
	metric.Count++
	metric.TotalTime += duration
	metric.LastRecorded = time.Now()
	
	if duration < metric.MinTime {
		metric.MinTime = duration
	}
	if duration > metric.MaxTime {
		metric.MaxTime = duration
	}
}

// GetStats returns performance statistics
func (pm *PerformanceMonitor) GetStats() PerformanceStats {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()
	
	stats := PerformanceStats{
		Uptime:     time.Since(pm.startTime),
		Metrics:    make([]MetricStats, 0, len(pm.metrics)),
		MemoryInfo: pm.getMemoryInfo(),
	}
	
	for _, metric := range pm.metrics {
		avgTime := time.Duration(0)
		if metric.Count > 0 {
			avgTime = metric.TotalTime / time.Duration(metric.Count)
		}
		
		stats.Metrics = append(stats.Metrics, MetricStats{
			Name:      metric.Name,
			Count:     metric.Count,
			AvgTime:   avgTime,
			MinTime:   metric.MinTime,
			MaxTime:   metric.MaxTime,
			TotalTime: metric.TotalTime,
		})
	}
	
	return stats
}

// getMemoryInfo returns current memory usage information
func (pm *PerformanceMonitor) getMemoryInfo() MemoryInfo {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	return MemoryInfo{
		Alloc:        m.Alloc,
		TotalAlloc:   m.TotalAlloc,
		Sys:          m.Sys,
		NumGC:        m.NumGC,
		HeapObjects:  m.HeapObjects,
		StackInuse:   m.StackInuse,
	}
}

// Reset clears all metrics
func (pm *PerformanceMonitor) Reset() {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	
	pm.metrics = make(map[string]*Metric)
	pm.startTime = time.Now()
}

// Timer helps measure operation duration
type Timer struct {
	name    string
	start   time.Time
	monitor *PerformanceMonitor
}

// Stop stops the timer and records the duration
func (t *Timer) Stop() time.Duration {
	duration := time.Since(t.start)
	t.monitor.RecordDuration(t.name, duration)
	return duration
}

// PerformanceStats provides performance statistics
type PerformanceStats struct {
	Uptime     time.Duration `json:"uptime"`
	Metrics    []MetricStats `json:"metrics"`
	MemoryInfo MemoryInfo    `json:"memory"`
}

// MetricStats provides statistics for a single metric
type MetricStats struct {
	Name      string        `json:"name"`
	Count     int64         `json:"count"`
	AvgTime   time.Duration `json:"avg_time"`
	MinTime   time.Duration `json:"min_time"`
	MaxTime   time.Duration `json:"max_time"`
	TotalTime time.Duration `json:"total_time"`
}

// MemoryInfo provides memory usage information
type MemoryInfo struct {
	Alloc        uint64 `json:"alloc"`         // Currently allocated memory
	TotalAlloc   uint64 `json:"total_alloc"`   // Total allocated memory
	Sys          uint64 `json:"sys"`           // System memory
	NumGC        uint32 `json:"num_gc"`        // Number of GC runs
	HeapObjects  uint64 `json:"heap_objects"`  // Number of heap objects
	StackInuse   uint64 `json:"stack_inuse"`   // Stack memory in use
}

// FormatBytes formats bytes in human-readable format
func FormatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// BenchmarkView benchmarks view rendering performance
func BenchmarkView(view ViewRenderer, iterations int) ViewBenchmark {
	benchmark := ViewBenchmark{
		ViewName:   "Unknown",
		Iterations: iterations,
		StartTime:  time.Now(),
	}
	
	var totalTime time.Duration
	minTime := time.Duration(^uint64(0) >> 1) // Max duration
	maxTime := time.Duration(0)
	
	for i := 0; i < iterations; i++ {
		start := time.Now()
		_ = view.Render()
		duration := time.Since(start)
		
		totalTime += duration
		if duration < minTime {
			minTime = duration
		}
		if duration > maxTime {
			maxTime = duration
		}
	}
	
	benchmark.EndTime = time.Now()
	benchmark.TotalTime = totalTime
	benchmark.AvgTime = totalTime / time.Duration(iterations)
	benchmark.MinTime = minTime
	benchmark.MaxTime = maxTime
	
	return benchmark
}

// ViewRenderer interface for benchmarking views
type ViewRenderer interface {
	Render() string
}

// ViewBenchmark provides view rendering benchmark results
type ViewBenchmark struct {
	ViewName   string        `json:"view_name"`
	Iterations int           `json:"iterations"`
	StartTime  time.Time     `json:"start_time"`
	EndTime    time.Time     `json:"end_time"`
	TotalTime  time.Duration `json:"total_time"`
	AvgTime    time.Duration `json:"avg_time"`
	MinTime    time.Duration `json:"min_time"`
	MaxTime    time.Duration `json:"max_time"`
}

// String returns a formatted string representation of benchmark results
func (vb ViewBenchmark) String() string {
	return fmt.Sprintf("View: %s, Iterations: %d, Avg: %v, Min: %v, Max: %v", 
		vb.ViewName, vb.Iterations, vb.AvgTime, vb.MinTime, vb.MaxTime)
}

// BenchmarkCommand benchmarks Bubble Tea command execution
func BenchmarkCommand(cmd tea.Cmd, iterations int) CommandBenchmark {
	benchmark := CommandBenchmark{
		Iterations: iterations,
		StartTime:  time.Now(),
	}
	
	var totalTime time.Duration
	minTime := time.Duration(^uint64(0) >> 1)
	maxTime := time.Duration(0)
	
	for i := 0; i < iterations; i++ {
		start := time.Now()
		_ = cmd()
		duration := time.Since(start)
		
		totalTime += duration
		if duration < minTime {
			minTime = duration
		}
		if duration > maxTime {
			maxTime = duration
		}
	}
	
	benchmark.EndTime = time.Now()
	benchmark.TotalTime = totalTime
	benchmark.AvgTime = totalTime / time.Duration(iterations)
	benchmark.MinTime = minTime
	benchmark.MaxTime = maxTime
	
	return benchmark
}

// CommandBenchmark provides command execution benchmark results
type CommandBenchmark struct {
	Iterations int           `json:"iterations"`
	StartTime  time.Time     `json:"start_time"`
	EndTime    time.Time     `json:"end_time"`
	TotalTime  time.Duration `json:"total_time"`
	AvgTime    time.Duration `json:"avg_time"`
	MinTime    time.Duration `json:"min_time"`
	MaxTime    time.Duration `json:"max_time"`
}

// String returns a formatted string representation of command benchmark results
func (cb CommandBenchmark) String() string {
	return fmt.Sprintf("Command Benchmark: Iterations: %d, Avg: %v, Min: %v, Max: %v", 
		cb.Iterations, cb.AvgTime, cb.MinTime, cb.MaxTime)
}