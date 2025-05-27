// ABOUTME: Performance monitoring and metrics collection for LSP operations
// ABOUTME: Provides operation timing, memory usage tracking, and performance regression detection

package ls

import (
	"context"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// PerformanceMonitor tracks LSP operation performance metrics
type PerformanceMonitor struct {
	// Operation metrics
	hoverMetrics      *OperationMetrics
	completionMetrics *OperationMetrics
	definitionMetrics *OperationMetrics
	referencesMetrics *OperationMetrics
	parseMetrics      *OperationMetrics
	bindMetrics       *OperationMetrics
	typeCheckMetrics  *OperationMetrics

	// Global counters
	totalRequests int64
	totalErrors   int64
	cacheHits     int64
	cacheMisses   int64

	// Memory tracking
	memoryUsage     int64
	peakMemoryUsage int64

	// Configuration
	enableDetailedMetrics bool
	maxHistorySize        int

	mu sync.RWMutex
}

// OperationMetrics tracks performance for a specific operation type
type OperationMetrics struct {
	name          string
	totalCount    int64
	totalDuration int64 // nanoseconds
	minDuration   int64
	maxDuration   int64
	errorCount    int64

	// Percentile tracking
	durations      []int64 // Recent durations for percentile calculation
	maxHistorySize int

	mu sync.RWMutex
}

// PerformanceStats provides a snapshot of performance metrics
type PerformanceStats struct {
	// Operation stats
	HoverStats      OperationStats
	CompletionStats OperationStats
	DefinitionStats OperationStats
	ReferencesStats OperationStats
	ParseStats      OperationStats
	BindStats       OperationStats
	TypeCheckStats  OperationStats

	// Global stats
	TotalRequests   int64
	TotalErrors     int64
	ErrorRate       float64
	CacheHitRate    float64
	CurrentMemoryMB float64
	PeakMemoryMB    float64

	// Timing
	CollectedAt   time.Time
	UptimeSeconds int64
}

// OperationStats provides statistics for a specific operation
type OperationStats struct {
	Name          string
	Count         int64
	ErrorCount    int64
	ErrorRate     float64
	AvgDurationMs float64
	MinDurationMs float64
	MaxDurationMs float64
	P50DurationMs float64
	P95DurationMs float64
	P99DurationMs float64
}

// PerformanceTargets defines expected performance characteristics
type PerformanceTargets struct {
	HoverMaxMs      int64
	CompletionMaxMs int64
	DefinitionMaxMs int64
	ReferencesMaxMs int64
	ParseMaxMs      int64
	BindMaxMs       int64
	TypeCheckMaxMs  int64
	MaxErrorRate    float64
	MinCacheHitRate float64
}

// TimedOperation wraps an operation with performance monitoring
type TimedOperation struct {
	ctx       context.Context
	monitor   *PerformanceMonitor
	metrics   *OperationMetrics
	startTime time.Time
	completed bool
}

var (
	// Default performance targets based on LSP responsiveness requirements
	DefaultTargets = PerformanceTargets{
		HoverMaxMs:      25,   // Hover should be near-instant
		CompletionMaxMs: 100,  // Completion should be responsive
		DefinitionMaxMs: 50,   // Go-to-definition should be fast
		ReferencesMaxMs: 200,  // Find references can be slightly slower
		ParseMaxMs:      500,  // Parsing can take more time for large files
		BindMaxMs:       200,  // Symbol binding should be reasonably fast
		TypeCheckMaxMs:  1000, // Type checking can be slower for complex files
		MaxErrorRate:    0.05, // 5% max error rate
		MinCacheHitRate: 0.7,  // 70% minimum cache hit rate
	}
)

// NewPerformanceMonitor creates a new performance monitor
func NewPerformanceMonitor() *PerformanceMonitor {
	maxHistory := 1000 // Keep last 1000 operations for percentile calculation

	return &PerformanceMonitor{
		hoverMetrics:          NewOperationMetrics("hover", maxHistory),
		completionMetrics:     NewOperationMetrics("completion", maxHistory),
		definitionMetrics:     NewOperationMetrics("definition", maxHistory),
		referencesMetrics:     NewOperationMetrics("references", maxHistory),
		parseMetrics:          NewOperationMetrics("parse", maxHistory),
		bindMetrics:           NewOperationMetrics("bind", maxHistory),
		typeCheckMetrics:      NewOperationMetrics("typecheck", maxHistory),
		enableDetailedMetrics: true,
		maxHistorySize:        maxHistory,
	}
}

// NewOperationMetrics creates metrics tracking for a specific operation
func NewOperationMetrics(name string, maxHistory int) *OperationMetrics {
	return &OperationMetrics{
		name:           name,
		minDuration:    int64(^uint64(0) >> 1), // max int64
		maxDuration:    0,
		durations:      make([]int64, 0, maxHistory),
		maxHistorySize: maxHistory,
	}
}

// StartOperation begins timing an operation
func (pm *PerformanceMonitor) StartOperation(ctx context.Context, operation string) *TimedOperation {
	var metrics *OperationMetrics

	switch operation {
	case "hover":
		metrics = pm.hoverMetrics
	case "completion":
		metrics = pm.completionMetrics
	case "definition":
		metrics = pm.definitionMetrics
	case "references":
		metrics = pm.referencesMetrics
	case "parse":
		metrics = pm.parseMetrics
	case "bind":
		metrics = pm.bindMetrics
	case "typecheck":
		metrics = pm.typeCheckMetrics
	default:
		// Create a temporary metrics for unknown operations
		metrics = NewOperationMetrics(operation, 100)
	}

	atomic.AddInt64(&pm.totalRequests, 1)

	return &TimedOperation{
		ctx:       ctx,
		monitor:   pm,
		metrics:   metrics,
		startTime: time.Now(),
		completed: false,
	}
}

// Complete finishes timing an operation and records the duration
func (to *TimedOperation) Complete() {
	if to.completed {
		return
	}
	to.completed = true

	duration := time.Since(to.startTime)
	to.metrics.RecordDuration(duration)
}

// CompleteWithError finishes timing an operation that resulted in an error
func (to *TimedOperation) CompleteWithError(err error) {
	if to.completed {
		return
	}
	to.completed = true

	duration := time.Since(to.startTime)
	to.metrics.RecordDuration(duration)
	to.metrics.RecordError()

	atomic.AddInt64(&to.monitor.totalErrors, 1)
}

// RecordDuration adds a duration measurement to the metrics
func (om *OperationMetrics) RecordDuration(duration time.Duration) {
	om.mu.Lock()
	defer om.mu.Unlock()

	durationNs := duration.Nanoseconds()

	om.totalCount++
	om.totalDuration += durationNs

	if durationNs < om.minDuration {
		om.minDuration = durationNs
	}
	if durationNs > om.maxDuration {
		om.maxDuration = durationNs
	}

	// Add to durations history for percentile calculation
	om.durations = append(om.durations, durationNs)
	if len(om.durations) > om.maxHistorySize {
		// Remove oldest entries to maintain size limit
		copy(om.durations, om.durations[len(om.durations)-om.maxHistorySize:])
		om.durations = om.durations[:om.maxHistorySize]
	}
}

// RecordError increments the error count
func (om *OperationMetrics) RecordError() {
	atomic.AddInt64(&om.errorCount, 1)
}

// GetStats returns current operation statistics
func (om *OperationMetrics) GetStats() OperationStats {
	om.mu.RLock()
	defer om.mu.RUnlock()

	stats := OperationStats{
		Name:       om.name,
		Count:      om.totalCount,
		ErrorCount: om.errorCount,
	}

	if om.totalCount > 0 {
		stats.ErrorRate = float64(om.errorCount) / float64(om.totalCount)
		stats.AvgDurationMs = float64(om.totalDuration) / float64(om.totalCount) / 1e6
		stats.MinDurationMs = float64(om.minDuration) / 1e6
		stats.MaxDurationMs = float64(om.maxDuration) / 1e6

		// Calculate percentiles
		if len(om.durations) > 0 {
			sortedDurations := make([]int64, len(om.durations))
			copy(sortedDurations, om.durations)
			sort.Slice(sortedDurations, func(i, j int) bool {
				return sortedDurations[i] < sortedDurations[j]
			})

			stats.P50DurationMs = float64(om.percentile(sortedDurations, 0.5)) / 1e6
			stats.P95DurationMs = float64(om.percentile(sortedDurations, 0.95)) / 1e6
			stats.P99DurationMs = float64(om.percentile(sortedDurations, 0.99)) / 1e6
		}
	}

	return stats
}

// percentile calculates the given percentile from sorted durations
func (om *OperationMetrics) percentile(sortedDurations []int64, p float64) int64 {
	if len(sortedDurations) == 0 {
		return 0
	}

	index := int(float64(len(sortedDurations)-1) * p)
	if index < 0 {
		index = 0
	}
	if index >= len(sortedDurations) {
		index = len(sortedDurations) - 1
	}

	return sortedDurations[index]
}

// RecordCacheHit increments the cache hit counter
func (pm *PerformanceMonitor) RecordCacheHit() {
	atomic.AddInt64(&pm.cacheHits, 1)
}

// RecordCacheMiss increments the cache miss counter
func (pm *PerformanceMonitor) RecordCacheMiss() {
	atomic.AddInt64(&pm.cacheMisses, 1)
}

// UpdateMemoryUsage updates the current memory usage
func (pm *PerformanceMonitor) UpdateMemoryUsage(bytes int64) {
	atomic.StoreInt64(&pm.memoryUsage, bytes)

	// Update peak if necessary
	for {
		current := atomic.LoadInt64(&pm.peakMemoryUsage)
		if bytes <= current {
			break
		}
		if atomic.CompareAndSwapInt64(&pm.peakMemoryUsage, current, bytes) {
			break
		}
	}
}

// GetStats returns comprehensive performance statistics
func (pm *PerformanceMonitor) GetStats() PerformanceStats {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	totalRequests := atomic.LoadInt64(&pm.totalRequests)
	totalErrors := atomic.LoadInt64(&pm.totalErrors)
	cacheHits := atomic.LoadInt64(&pm.cacheHits)
	cacheMisses := atomic.LoadInt64(&pm.cacheMisses)
	memoryUsage := atomic.LoadInt64(&pm.memoryUsage)
	peakMemoryUsage := atomic.LoadInt64(&pm.peakMemoryUsage)

	var errorRate float64
	if totalRequests > 0 {
		errorRate = float64(totalErrors) / float64(totalRequests)
	}

	var cacheHitRate float64
	totalCacheRequests := cacheHits + cacheMisses
	if totalCacheRequests > 0 {
		cacheHitRate = float64(cacheHits) / float64(totalCacheRequests)
	}

	return PerformanceStats{
		HoverStats:      pm.hoverMetrics.GetStats(),
		CompletionStats: pm.completionMetrics.GetStats(),
		DefinitionStats: pm.definitionMetrics.GetStats(),
		ReferencesStats: pm.referencesMetrics.GetStats(),
		ParseStats:      pm.parseMetrics.GetStats(),
		BindStats:       pm.bindMetrics.GetStats(),
		TypeCheckStats:  pm.typeCheckMetrics.GetStats(),
		TotalRequests:   totalRequests,
		TotalErrors:     totalErrors,
		ErrorRate:       errorRate,
		CacheHitRate:    cacheHitRate,
		CurrentMemoryMB: float64(memoryUsage) / (1024 * 1024),
		PeakMemoryMB:    float64(peakMemoryUsage) / (1024 * 1024),
		CollectedAt:     time.Now(),
	}
}

// CheckPerformanceTargets validates current performance against targets
func (pm *PerformanceMonitor) CheckPerformanceTargets(targets PerformanceTargets) []string {
	stats := pm.GetStats()
	var violations []string

	// Check individual operation targets
	if stats.HoverStats.P95DurationMs > float64(targets.HoverMaxMs) {
		violations = append(violations, "Hover P95 latency exceeds target")
	}
	if stats.CompletionStats.P95DurationMs > float64(targets.CompletionMaxMs) {
		violations = append(violations, "Completion P95 latency exceeds target")
	}
	if stats.DefinitionStats.P95DurationMs > float64(targets.DefinitionMaxMs) {
		violations = append(violations, "Definition P95 latency exceeds target")
	}
	if stats.ReferencesStats.P95DurationMs > float64(targets.ReferencesMaxMs) {
		violations = append(violations, "References P95 latency exceeds target")
	}
	if stats.ParseStats.P95DurationMs > float64(targets.ParseMaxMs) {
		violations = append(violations, "Parse P95 latency exceeds target")
	}
	if stats.BindStats.P95DurationMs > float64(targets.BindMaxMs) {
		violations = append(violations, "Bind P95 latency exceeds target")
	}
	if stats.TypeCheckStats.P95DurationMs > float64(targets.TypeCheckMaxMs) {
		violations = append(violations, "TypeCheck P95 latency exceeds target")
	}

	// Check global targets
	if stats.ErrorRate > targets.MaxErrorRate {
		violations = append(violations, "Error rate exceeds target")
	}
	if stats.CacheHitRate < targets.MinCacheHitRate {
		violations = append(violations, "Cache hit rate below target")
	}

	return violations
}

// Reset clears all performance metrics
func (pm *PerformanceMonitor) Reset() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.hoverMetrics = NewOperationMetrics("hover", pm.maxHistorySize)
	pm.completionMetrics = NewOperationMetrics("completion", pm.maxHistorySize)
	pm.definitionMetrics = NewOperationMetrics("definition", pm.maxHistorySize)
	pm.referencesMetrics = NewOperationMetrics("references", pm.maxHistorySize)
	pm.parseMetrics = NewOperationMetrics("parse", pm.maxHistorySize)
	pm.bindMetrics = NewOperationMetrics("bind", pm.maxHistorySize)
	pm.typeCheckMetrics = NewOperationMetrics("typecheck", pm.maxHistorySize)

	atomic.StoreInt64(&pm.totalRequests, 0)
	atomic.StoreInt64(&pm.totalErrors, 0)
	atomic.StoreInt64(&pm.cacheHits, 0)
	atomic.StoreInt64(&pm.cacheMisses, 0)
	atomic.StoreInt64(&pm.memoryUsage, 0)
	atomic.StoreInt64(&pm.peakMemoryUsage, 0)
}
