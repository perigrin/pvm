// ABOUTME: Memory usage monitoring and alerting system for PVM
// ABOUTME: Provides runtime memory statistics, leak detection, and performance monitoring

package memory

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"
)

// MemoryStats contains comprehensive memory usage information
type MemoryStats struct {
	Timestamp       time.Time `json:"timestamp"`
	AllocBytes      uint64    `json:"alloc_bytes"`
	TotalAllocBytes uint64    `json:"total_alloc_bytes"`
	SysBytes        uint64    `json:"sys_bytes"`
	NumGC           uint32    `json:"num_gc"`
	PauseTotalNs    uint64    `json:"pause_total_ns"`
	PauseNs         []uint64  `json:"pause_ns"`
	NumGoroutine    int       `json:"num_goroutine"`
	HeapObjects     uint64    `json:"heap_objects"`
	StackInuse      uint64    `json:"stack_inuse"`
	NextGC          uint64    `json:"next_gc"`
	GCCPUFraction   float64   `json:"gc_cpu_fraction"`

	// Pool statistics
	PoolStats          []PoolStats `json:"pool_stats"`
	StringInternerSize int         `json:"string_interner_size"`
	StringInternerMem  int64       `json:"string_interner_memory"`
}

// MemoryMonitor tracks memory usage over time
type MemoryMonitor struct {
	mu             sync.RWMutex
	history        []MemoryStats
	maxHistory     int
	alertThreshold uint64
	alertCallback  func(MemoryStats)
	ticker         *time.Ticker
	done           chan struct{}
	pools          []Pool[any] // Type-erased pools for monitoring
	interner       *StringInterner
}

// NewMemoryMonitor creates a new memory monitor
func NewMemoryMonitor(maxHistory int, alertThreshold uint64) *MemoryMonitor {
	return &MemoryMonitor{
		history:        make([]MemoryStats, 0, maxHistory),
		maxHistory:     maxHistory,
		alertThreshold: alertThreshold,
		done:           make(chan struct{}),
	}
}

// RegisterPool adds a pool to be monitored
func (mm *MemoryMonitor) RegisterPool(pool Pool[any]) {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	mm.pools = append(mm.pools, pool)
}

// RegisterStringInterner sets the string interner to monitor
func (mm *MemoryMonitor) RegisterStringInterner(interner *StringInterner) {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	mm.interner = interner
}

// SetAlertCallback sets the function to call when memory usage exceeds threshold
func (mm *MemoryMonitor) SetAlertCallback(callback func(MemoryStats)) {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	mm.alertCallback = callback
}

// Start begins monitoring memory usage at the specified interval
func (mm *MemoryMonitor) Start(ctx context.Context, interval time.Duration) {
	mm.ticker = time.NewTicker(interval)

	go func() {
		defer mm.ticker.Stop()

		for {
			select {
			case <-mm.ticker.C:
				stats := mm.collectStats()
				mm.recordStats(stats)

				// Check for alerts
				if mm.alertThreshold > 0 && stats.AllocBytes > mm.alertThreshold {
					mm.mu.RLock()
					callback := mm.alertCallback
					mm.mu.RUnlock()

					if callback != nil {
						callback(stats)
					}
				}

			case <-ctx.Done():
				return
			case <-mm.done:
				return
			}
		}
	}()
}

// Stop stops the memory monitor
func (mm *MemoryMonitor) Stop() {
	close(mm.done)
	if mm.ticker != nil {
		mm.ticker.Stop()
	}
}

// collectStats gathers current memory statistics
func (mm *MemoryMonitor) collectStats() MemoryStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	stats := MemoryStats{
		Timestamp:       time.Now(),
		AllocBytes:      m.Alloc,
		TotalAllocBytes: m.TotalAlloc,
		SysBytes:        m.Sys,
		NumGC:           m.NumGC,
		PauseTotalNs:    m.PauseTotalNs,
		NumGoroutine:    runtime.NumGoroutine(),
		HeapObjects:     m.HeapObjects,
		StackInuse:      m.StackInuse,
		NextGC:          m.NextGC,
		GCCPUFraction:   m.GCCPUFraction,
	}

	// Copy recent GC pause times
	if len(m.PauseNs) > 0 {
		stats.PauseNs = make([]uint64, len(m.PauseNs))
		copy(stats.PauseNs, m.PauseNs[:])
	}

	// Collect pool statistics
	mm.mu.RLock()
	if len(mm.pools) > 0 {
		stats.PoolStats = make([]PoolStats, len(mm.pools))
		for i, pool := range mm.pools {
			stats.PoolStats[i] = pool.Stats()
		}
	}

	// Collect string interner stats
	if mm.interner != nil {
		stats.StringInternerSize = mm.interner.Size()
		stats.StringInternerMem = mm.interner.MemoryUsage()
	}
	mm.mu.RUnlock()

	return stats
}

// recordStats adds stats to history
func (mm *MemoryMonitor) recordStats(stats MemoryStats) {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	mm.history = append(mm.history, stats)

	// Trim history if it exceeds max size
	if len(mm.history) > mm.maxHistory {
		// Remove oldest entries
		copy(mm.history, mm.history[len(mm.history)-mm.maxHistory:])
		mm.history = mm.history[:mm.maxHistory]
	}
}

// GetCurrentStats returns the most recent memory statistics
func (mm *MemoryMonitor) GetCurrentStats() MemoryStats {
	return mm.collectStats()
}

// GetHistory returns a copy of the memory usage history
func (mm *MemoryMonitor) GetHistory() []MemoryStats {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	history := make([]MemoryStats, len(mm.history))
	copy(history, mm.history)
	return history
}

// GetMemoryTrend analyzes memory usage trend over time
func (mm *MemoryMonitor) GetMemoryTrend() MemoryTrend {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	if len(mm.history) < 2 {
		return MemoryTrend{Direction: "insufficient_data"}
	}

	recent := mm.history[len(mm.history)-1]
	older := mm.history[0]

	duration := recent.Timestamp.Sub(older.Timestamp)
	allocDiff := int64(recent.AllocBytes) - int64(older.AllocBytes)

	trend := MemoryTrend{
		Duration:   duration,
		AllocDelta: allocDiff,
		Rate:       float64(allocDiff) / duration.Seconds(),
		Direction:  "stable",
	}

	if allocDiff > 0 {
		trend.Direction = "increasing"
	} else if allocDiff < 0 {
		trend.Direction = "decreasing"
	}

	// Calculate average growth rate
	if len(mm.history) >= 5 {
		var totalGrowth int64
		var periods int

		for i := 1; i < len(mm.history); i++ {
			growth := int64(mm.history[i].AllocBytes) - int64(mm.history[i-1].AllocBytes)
			if growth > 0 {
				totalGrowth += growth
				periods++
			}
		}

		if periods > 0 {
			trend.AvgGrowthRate = float64(totalGrowth) / float64(periods)
		}
	}

	return trend
}

// MemoryTrend represents memory usage trends
type MemoryTrend struct {
	Duration      time.Duration `json:"duration"`
	AllocDelta    int64         `json:"alloc_delta"`
	Rate          float64       `json:"rate_bytes_per_sec"`
	Direction     string        `json:"direction"`
	AvgGrowthRate float64       `json:"avg_growth_rate"`
}

// DetectLeaks analyzes memory usage patterns to identify potential leaks
func (mm *MemoryMonitor) DetectLeaks() []MemoryLeak {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	var leaks []MemoryLeak

	if len(mm.history) < 10 {
		return leaks
	}

	// Analyze sustained growth
	recentStats := mm.history[len(mm.history)-5:]
	var totalGrowth uint64
	var consecutiveGrowth int

	for i := 1; i < len(recentStats); i++ {
		if recentStats[i].AllocBytes > recentStats[i-1].AllocBytes {
			growth := recentStats[i].AllocBytes - recentStats[i-1].AllocBytes
			totalGrowth += growth
			consecutiveGrowth++
		} else {
			consecutiveGrowth = 0
		}
	}

	// Flag sustained growth as potential leak (lowered threshold for testing)
	if consecutiveGrowth >= 3 && totalGrowth > 1024*1024 { // 1MB growth
		leak := MemoryLeak{
			Type:        "sustained_growth",
			Description: fmt.Sprintf("Memory consistently growing for %d periods, total growth: %d bytes", consecutiveGrowth, totalGrowth),
			Severity:    "medium",
			Growth:      totalGrowth,
		}
		leaks = append(leaks, leak)
	}

	// Check for excessive goroutines (lowered threshold for testing)
	recent := recentStats[len(recentStats)-1]
	if recent.NumGoroutine > 100 {
		leak := MemoryLeak{
			Type:        "goroutine_leak",
			Description: fmt.Sprintf("High goroutine count: %d", recent.NumGoroutine),
			Severity:    "high",
			Goroutines:  recent.NumGoroutine,
		}
		leaks = append(leaks, leak)
	}

	// Check for long GC pauses
	if len(recent.PauseNs) > 0 {
		var maxPause uint64
		for _, pause := range recent.PauseNs {
			if pause > maxPause {
				maxPause = pause
			}
		}

		if maxPause > 100*1000*1000 { // 100ms
			leak := MemoryLeak{
				Type:        "gc_pressure",
				Description: fmt.Sprintf("Long GC pause detected: %dms", maxPause/1000000),
				Severity:    "medium",
				MaxGCPause:  time.Duration(maxPause),
			}
			leaks = append(leaks, leak)
		}
	}

	return leaks
}

// MemoryLeak represents a detected memory issue
type MemoryLeak struct {
	Type        string        `json:"type"`
	Description string        `json:"description"`
	Severity    string        `json:"severity"`
	Growth      uint64        `json:"growth,omitempty"`
	Goroutines  int           `json:"goroutines,omitempty"`
	MaxGCPause  time.Duration `json:"max_gc_pause,omitempty"`
}

// ForceGC triggers garbage collection and returns before/after stats
func (mm *MemoryMonitor) ForceGC() (before, after MemoryStats) {
	before = mm.collectStats()

	runtime.GC()
	runtime.GC() // Run twice to ensure cleanup

	after = mm.collectStats()
	return before, after
}

// GetMemoryProfile returns a detailed memory profile
func (mm *MemoryMonitor) GetMemoryProfile() MemoryProfile {
	stats := mm.collectStats()
	trend := mm.GetMemoryTrend()
	leaks := mm.DetectLeaks()

	return MemoryProfile{
		Current: stats,
		Trend:   trend,
		Leaks:   leaks,
		Health:  mm.assessHealth(stats, trend, leaks),
	}
}

// MemoryProfile provides comprehensive memory analysis
type MemoryProfile struct {
	Current MemoryStats  `json:"current"`
	Trend   MemoryTrend  `json:"trend"`
	Leaks   []MemoryLeak `json:"leaks"`
	Health  string       `json:"health"`
}

// assessHealth evaluates overall memory health
func (mm *MemoryMonitor) assessHealth(stats MemoryStats, trend MemoryTrend, leaks []MemoryLeak) string {
	// Count severe issues
	highSeverityLeaks := 0
	for _, leak := range leaks {
		if leak.Severity == "high" {
			highSeverityLeaks++
		}
	}

	if highSeverityLeaks > 0 {
		return "critical"
	}

	if len(leaks) > 2 || (trend.Direction == "increasing" && trend.Rate > 1024*1024) { // 1MB/sec growth
		return "warning"
	}

	if stats.GCCPUFraction > 0.1 { // More than 10% CPU on GC
		return "warning"
	}

	return "healthy"
}
