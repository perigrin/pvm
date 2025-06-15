// ABOUTME: Performance analysis and optimization utilities for PVM
// ABOUTME: Provides profiling, benchmarking, and bottleneck identification

package performance

import (
	"runtime"
	"runtime/debug"
	"sync"
	"time"
)

// Analyzer tracks performance metrics across the system
type Analyzer struct {
	mu      sync.RWMutex
	metrics map[string]*Metric
	enabled bool
}

// Metric represents a performance measurement
type Metric struct {
	Name           string
	Count          int64
	TotalTime      time.Duration
	MinTime        time.Duration
	MaxTime        time.Duration
	AvgTime        time.Duration
	LastTime       time.Duration
	StartTime      time.Time
	MemUsage       int64
	GoroutineCount int
}

// NewAnalyzer creates a new performance analyzer
func NewAnalyzer() *Analyzer {
	return &Analyzer{
		metrics: make(map[string]*Metric),
		enabled: true,
	}
}

// StartMeasurement begins timing a specific operation
func (a *Analyzer) StartMeasurement(name string) *MeasurementContext {
	if !a.enabled {
		return &MeasurementContext{analyzer: a, name: name}
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	if _, exists := a.metrics[name]; !exists {
		a.metrics[name] = &Metric{
			Name:    name,
			MinTime: time.Hour, // Large initial value
		}
	}

	return &MeasurementContext{
		analyzer:        a,
		name:            name,
		startTime:       time.Now(),
		startMem:        getCurrentMemoryUsage(),
		startGoroutines: runtime.NumGoroutine(),
	}
}

// MeasurementContext tracks a single measurement
type MeasurementContext struct {
	analyzer        *Analyzer
	name            string
	startTime       time.Time
	startMem        int64
	startGoroutines int
}

// Finish completes the measurement and records results
func (mc *MeasurementContext) Finish() {
	if !mc.analyzer.enabled {
		return
	}

	duration := time.Since(mc.startTime)
	memUsage := getCurrentMemoryUsage() - mc.startMem
	goroutineCount := runtime.NumGoroutine() - mc.startGoroutines

	mc.analyzer.mu.Lock()
	defer mc.analyzer.mu.Unlock()

	metric := mc.analyzer.metrics[mc.name]
	metric.Count++
	metric.TotalTime += duration
	metric.LastTime = duration

	if duration < metric.MinTime {
		metric.MinTime = duration
	}
	if duration > metric.MaxTime {
		metric.MaxTime = duration
	}

	metric.AvgTime = metric.TotalTime / time.Duration(metric.Count)
	metric.MemUsage += memUsage
	metric.GoroutineCount += goroutineCount
}

// GetMetrics returns all collected metrics
func (a *Analyzer) GetMetrics() map[string]*Metric {
	a.mu.RLock()
	defer a.mu.RUnlock()

	result := make(map[string]*Metric)
	for name, metric := range a.metrics {
		// Create a copy to avoid data races
		result[name] = &Metric{
			Name:           metric.Name,
			Count:          metric.Count,
			TotalTime:      metric.TotalTime,
			MinTime:        metric.MinTime,
			MaxTime:        metric.MaxTime,
			AvgTime:        metric.AvgTime,
			LastTime:       metric.LastTime,
			MemUsage:       metric.MemUsage,
			GoroutineCount: metric.GoroutineCount,
		}
	}
	return result
}

// Reset clears all metrics
func (a *Analyzer) Reset() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.metrics = make(map[string]*Metric)
}

// Enable/Disable the analyzer
func (a *Analyzer) SetEnabled(enabled bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.enabled = enabled
}

// IsEnabled returns whether the analyzer is enabled
func (a *Analyzer) IsEnabled() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.enabled
}

// GetSlowestOperations returns the N slowest operations by average time
func (a *Analyzer) GetSlowestOperations(n int) []*Metric {
	metrics := a.GetMetrics()

	var sortedMetrics []*Metric
	for _, metric := range metrics {
		sortedMetrics = append(sortedMetrics, metric)
	}

	// Simple bubble sort by average time (descending)
	for i := 0; i < len(sortedMetrics)-1; i++ {
		for j := 0; j < len(sortedMetrics)-1-i; j++ {
			if sortedMetrics[j].AvgTime < sortedMetrics[j+1].AvgTime {
				sortedMetrics[j], sortedMetrics[j+1] = sortedMetrics[j+1], sortedMetrics[j]
			}
		}
	}

	if n > len(sortedMetrics) {
		n = len(sortedMetrics)
	}
	return sortedMetrics[:n]
}

// GetMemoryUsage returns current memory usage statistics
func (a *Analyzer) GetMemoryUsage() runtime.MemStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m
}

// ForceGC forces garbage collection and returns memory stats
func (a *Analyzer) ForceGC() runtime.MemStats {
	runtime.GC()
	debug.FreeOSMemory()
	return a.GetMemoryUsage()
}

// getCurrentMemoryUsage returns current allocated memory in bytes
func getCurrentMemoryUsage() int64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return int64(m.Alloc)
}

// Global analyzer instance
var globalAnalyzer = NewAnalyzer()

// Measure is a convenience function for measuring operations
func Measure(name string) *MeasurementContext {
	return globalAnalyzer.StartMeasurement(name)
}

// GetGlobalMetrics returns metrics from the global analyzer
func GetGlobalMetrics() map[string]*Metric {
	return globalAnalyzer.GetMetrics()
}

// ResetGlobalMetrics clears the global analyzer
func ResetGlobalMetrics() {
	globalAnalyzer.Reset()
}

// SetGlobalAnalyzerEnabled enables/disables the global analyzer
func SetGlobalAnalyzerEnabled(enabled bool) {
	globalAnalyzer.SetEnabled(enabled)
}
