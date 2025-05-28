// ABOUTME: Comprehensive pool performance monitoring and optimization system
// ABOUTME: Provides detailed analytics, benchmarking, and optimization recommendations for object pools

package monitoring

import (
	"context"
	"fmt"
	"sync"
	"time"

	"tamarou.com/pvm/internal/core"
)

// PoolPerformanceMonitor tracks and analyzes pool performance across the system
type PoolPerformanceMonitor struct {
	mu              sync.RWMutex
	pools           map[string]*PoolMetrics
	baselines       map[string]PerformanceBaseline
	optimizations   []OptimizationRecommendation
	benchmarks      []BenchmarkResult
	regressionTests []RegressionTest
	alertThresholds AlertThresholds
	config          MonitoringConfig
	ticker          *time.Ticker
	done            chan struct{}
	analysisHistory []AnalysisSnapshot
	maxHistory      int
}

// PoolMetrics contains detailed metrics for a single pool
type PoolMetrics struct {
	Name             string
	RegistrationTime time.Time
	LastUpdateTime   time.Time

	// Basic statistics
	TotalAllocations int64
	TotalGrows       int64
	TotalResets      int64
	CurrentSize      int64
	CurrentCapacity  int64
	PeakSize         int64
	PeakCapacity     int64

	// Performance metrics
	AllocationRate   float64 // Allocations per second
	GrowthRate       float64 // Grows per second
	UtilizationRatio float64 // CurrentSize / CurrentCapacity
	EfficiencyScore  float64 // Calculated efficiency metric

	// Memory metrics
	EstimatedMemoryKB int64
	MemoryEfficiency  float64
	WastedMemoryKB    int64

	// Historical data
	HourlyStats    []HourlyPoolStats
	PeakUsageTimes []time.Time
	ResetFrequency time.Duration

	// Optimization status
	OptimizationLevel  string
	LastOptimizedTime  time.Time
	RecommendedActions []string
}

// HourlyPoolStats contains aggregated hourly statistics
type HourlyPoolStats struct {
	Hour             time.Time
	Allocations      int64
	Grows            int64
	AverageSize      float64
	PeakSize         int64
	UtilizationRatio float64
}

// PerformanceBaseline stores baseline metrics for comparison
type PerformanceBaseline struct {
	Name                 string
	BaselineDate         time.Time
	AllocationsPerSecond float64
	MemoryEfficiency     float64
	UtilizationRatio     float64
	GrowthFrequency      time.Duration
	TargetImprovement    float64
}

// OptimizationRecommendation suggests specific optimizations
type OptimizationRecommendation struct {
	PoolName             string
	Type                 string
	Priority             string
	Description          string
	EstimatedImprovement float64
	ImplementationSteps  []string
	CreatedAt            time.Time
	Status               string
}

// BenchmarkResult stores performance benchmark data
type BenchmarkResult struct {
	Name               string
	Timestamp          time.Time
	PooledAllocations  BenchmarkMetrics
	DirectAllocations  BenchmarkMetrics
	ImprovementPercent float64
	MemorySavings      int64
	TestConfiguration  BenchmarkConfig
}

// BenchmarkMetrics contains timing and memory metrics
type BenchmarkMetrics struct {
	TotalTime        time.Duration
	OperationsPerSec float64
	AllocationsCount int64
	MemoryAllocated  int64
	GCPauses         int
	CPUUsage         float64
}

// BenchmarkConfig defines benchmark parameters
type BenchmarkConfig struct {
	Operations        int
	Concurrency       int
	ObjectSize        int
	AllocationPattern string
	TestDuration      time.Duration
}

// RegressionTest defines performance regression testing
type RegressionTest struct {
	Name                string
	TestFunction        func() BenchmarkResult
	Baseline            BenchmarkResult
	Tolerance           float64
	LastRun             time.Time
	Status              string
	FailureThreshold    int
	ConsecutiveFailures int
}

// AlertThresholds defines when to alert on performance issues
type AlertThresholds struct {
	UtilizationBelow    float64
	MemoryWasteAbove    int64
	AllocationRateAbove float64
	EfficiencyBelow     float64
	RegressionPercent   float64
}

// MonitoringConfig contains monitoring configuration
type MonitoringConfig struct {
	UpdateInterval   time.Duration
	EnableBenchmarks bool
	EnableRegression bool
	AnalysisDepth    string
	ReportingEnabled bool
	OptimizationAuto bool
	AlertsEnabled    bool
}

// AnalysisSnapshot captures system-wide analysis at a point in time
type AnalysisSnapshot struct {
	Timestamp             time.Time
	TotalPools            int
	TotalAllocations      int64
	SystemEfficiency      float64
	MemoryUtilization     float64
	ActiveRecommendations int
	CriticalIssues        int
	PerformanceGrade      string
}

// NewPoolPerformanceMonitor creates a new performance monitor
func NewPoolPerformanceMonitor(config MonitoringConfig) *PoolPerformanceMonitor {
	return &PoolPerformanceMonitor{
		pools:           make(map[string]*PoolMetrics),
		baselines:       make(map[string]PerformanceBaseline),
		alertThresholds: DefaultAlertThresholds(),
		config:          config,
		done:            make(chan struct{}),
		maxHistory:      100,
	}
}

// DefaultAlertThresholds returns sensible default alert thresholds
func DefaultAlertThresholds() AlertThresholds {
	return AlertThresholds{
		UtilizationBelow:    0.3,  // Alert if pool utilization below 30%
		MemoryWasteAbove:    1024, // Alert if wasted memory above 1MB
		AllocationRateAbove: 1000, // Alert if allocation rate above 1000/sec
		EfficiencyBelow:     0.7,  // Alert if efficiency below 70%
		RegressionPercent:   5.0,  // Alert if performance regression > 5%
	}
}

// RegisterPool adds a pool for monitoring
func (ppm *PoolPerformanceMonitor) RegisterPool(name string, pool core.PoolStatsProvider) {
	ppm.mu.Lock()
	defer ppm.mu.Unlock()

	metrics := &PoolMetrics{
		Name:               name,
		RegistrationTime:   time.Now(),
		LastUpdateTime:     time.Now(),
		HourlyStats:        make([]HourlyPoolStats, 0, 24),
		PeakUsageTimes:     make([]time.Time, 0),
		RecommendedActions: make([]string, 0),
		OptimizationLevel:  "none",
	}

	ppm.pools[name] = metrics
	ppm.updatePoolMetrics(name, pool)
}

// Start begins performance monitoring
func (ppm *PoolPerformanceMonitor) Start(ctx context.Context) {
	ppm.ticker = time.NewTicker(ppm.config.UpdateInterval)

	go func() {
		defer ppm.ticker.Stop()

		for {
			select {
			case <-ppm.ticker.C:
				ppm.performAnalysis()
				ppm.generateRecommendations()
				if ppm.config.EnableBenchmarks {
					ppm.runBenchmarks()
				}
				if ppm.config.EnableRegression {
					ppm.runRegressionTests()
				}
				ppm.checkAlerts()

			case <-ctx.Done():
				return
			case <-ppm.done:
				return
			}
		}
	}()
}

// Stop stops the performance monitor
func (ppm *PoolPerformanceMonitor) Stop() {
	close(ppm.done)
	if ppm.ticker != nil {
		ppm.ticker.Stop()
	}
}

// UpdatePool updates metrics for a specific pool
func (ppm *PoolPerformanceMonitor) UpdatePool(name string, pool core.PoolStatsProvider) {
	ppm.mu.Lock()
	defer ppm.mu.Unlock()
	ppm.updatePoolMetrics(name, pool)
}

// updatePoolMetrics updates internal pool metrics (caller must hold lock)
func (ppm *PoolPerformanceMonitor) updatePoolMetrics(name string, pool core.PoolStatsProvider) {
	stats := pool.Stats()
	metrics := ppm.pools[name]

	now := time.Now()
	timeDelta := now.Sub(metrics.LastUpdateTime).Seconds()

	// Update basic statistics
	prevAllocations := metrics.TotalAllocations
	metrics.TotalAllocations = stats.Allocations
	metrics.TotalGrows = stats.Grows
	metrics.CurrentSize = stats.CurrentSize
	metrics.CurrentCapacity = stats.Capacity

	// Track peaks
	if stats.CurrentSize > metrics.PeakSize {
		metrics.PeakSize = stats.CurrentSize
		metrics.PeakUsageTimes = append(metrics.PeakUsageTimes, now)
	}
	if stats.Capacity > metrics.PeakCapacity {
		metrics.PeakCapacity = stats.Capacity
	}

	// Calculate rates
	if timeDelta > 0 {
		allocDelta := float64(stats.Allocations - prevAllocations)
		metrics.AllocationRate = allocDelta / timeDelta
	}

	// Calculate efficiency metrics
	if stats.Capacity > 0 {
		metrics.UtilizationRatio = float64(stats.CurrentSize) / float64(stats.Capacity)
	}

	// Estimate memory usage (assuming 64 bytes average object size)
	metrics.EstimatedMemoryKB = stats.Capacity * 64 / 1024

	// Calculate waste
	if stats.Capacity > stats.CurrentSize {
		metrics.WastedMemoryKB = (stats.Capacity - stats.CurrentSize) * 64 / 1024
	}

	// Calculate efficiency score (0-1, higher is better)
	metrics.EfficiencyScore = ppm.calculateEfficiencyScore(metrics)

	metrics.LastUpdateTime = now
}

// calculateEfficiencyScore computes an overall efficiency score
func (ppm *PoolPerformanceMonitor) calculateEfficiencyScore(metrics *PoolMetrics) float64 {
	// Weighted score based on utilization, memory efficiency, and allocation patterns
	utilizationScore := metrics.UtilizationRatio

	memoryScore := 1.0
	if metrics.EstimatedMemoryKB > 0 {
		memoryScore = 1.0 - (float64(metrics.WastedMemoryKB) / float64(metrics.EstimatedMemoryKB))
	}

	// Penalty for excessive growth
	growthPenalty := 1.0
	if metrics.TotalGrows > 0 && metrics.TotalAllocations > 0 {
		growthRatio := float64(metrics.TotalGrows) / float64(metrics.TotalAllocations)
		if growthRatio > 0.1 { // More than 10% grows vs allocations is concerning
			growthPenalty = 1.0 - (growthRatio - 0.1)
		}
	}

	// Weighted average
	return (0.4*utilizationScore + 0.4*memoryScore + 0.2*growthPenalty)
}

// performAnalysis runs comprehensive analysis of all pools
func (ppm *PoolPerformanceMonitor) performAnalysis() {
	ppm.mu.Lock()
	defer ppm.mu.Unlock()

	var totalAllocations int64
	var totalEfficiency float64
	var totalMemory int64
	var totalUtilization float64
	var criticalIssues int

	for _, metrics := range ppm.pools {
		totalAllocations += metrics.TotalAllocations
		totalEfficiency += metrics.EfficiencyScore
		totalMemory += metrics.EstimatedMemoryKB
		totalUtilization += metrics.UtilizationRatio

		// Count critical issues
		if metrics.EfficiencyScore < 0.5 {
			criticalIssues++
		}
	}

	poolCount := len(ppm.pools)
	if poolCount == 0 {
		return
	}

	avgEfficiency := totalEfficiency / float64(poolCount)
	avgUtilization := totalUtilization / float64(poolCount)

	// Determine performance grade
	grade := "A"
	if avgEfficiency < 0.9 {
		grade = "B"
	}
	if avgEfficiency < 0.8 {
		grade = "C"
	}
	if avgEfficiency < 0.7 {
		grade = "D"
	}
	if avgEfficiency < 0.6 {
		grade = "F"
	}

	snapshot := AnalysisSnapshot{
		Timestamp:             time.Now(),
		TotalPools:            poolCount,
		TotalAllocations:      totalAllocations,
		SystemEfficiency:      avgEfficiency,
		MemoryUtilization:     avgUtilization,
		ActiveRecommendations: len(ppm.optimizations),
		CriticalIssues:        criticalIssues,
		PerformanceGrade:      grade,
	}

	ppm.analysisHistory = append(ppm.analysisHistory, snapshot)

	// Trim history
	if len(ppm.analysisHistory) > ppm.maxHistory {
		ppm.analysisHistory = ppm.analysisHistory[len(ppm.analysisHistory)-ppm.maxHistory:]
	}
}

// generateRecommendations creates optimization recommendations
func (ppm *PoolPerformanceMonitor) generateRecommendations() {
	ppm.mu.Lock()
	defer ppm.mu.Unlock()

	// Clear old recommendations
	ppm.optimizations = ppm.optimizations[:0]

	for name, metrics := range ppm.pools {
		// Low utilization recommendation
		if metrics.UtilizationRatio < ppm.alertThresholds.UtilizationBelow {
			rec := OptimizationRecommendation{
				PoolName:             name,
				Type:                 "pool_sizing",
				Priority:             "medium",
				Description:          fmt.Sprintf("Pool utilization is low (%.1f%%). Consider reducing initial pool size.", metrics.UtilizationRatio*100),
				EstimatedImprovement: 15.0,
				ImplementationSteps: []string{
					"Reduce initial pool size by 30-50%",
					"Monitor allocation patterns",
					"Adjust growth factor if needed",
				},
				CreatedAt: time.Now(),
				Status:    "pending",
			}
			ppm.optimizations = append(ppm.optimizations, rec)
		}

		// High memory waste recommendation
		if metrics.WastedMemoryKB > ppm.alertThresholds.MemoryWasteAbove {
			rec := OptimizationRecommendation{
				PoolName:             name,
				Type:                 "memory_optimization",
				Priority:             "high",
				Description:          fmt.Sprintf("High memory waste detected (%dKB). Pool may be oversized.", metrics.WastedMemoryKB),
				EstimatedImprovement: 25.0,
				ImplementationSteps: []string{
					"Implement periodic pool reset",
					"Tune growth algorithm",
					"Consider object-specific pools",
				},
				CreatedAt: time.Now(),
				Status:    "pending",
			}
			ppm.optimizations = append(ppm.optimizations, rec)
		}

		// High allocation rate recommendation
		if metrics.AllocationRate > ppm.alertThresholds.AllocationRateAbove {
			rec := OptimizationRecommendation{
				PoolName:             name,
				Type:                 "allocation_optimization",
				Priority:             "high",
				Description:          fmt.Sprintf("High allocation rate (%.0f/sec). Consider pool warming.", metrics.AllocationRate),
				EstimatedImprovement: 20.0,
				ImplementationSteps: []string{
					"Implement pool warming at startup",
					"Increase initial pool size",
					"Consider allocation batching",
				},
				CreatedAt: time.Now(),
				Status:    "pending",
			}
			ppm.optimizations = append(ppm.optimizations, rec)
		}

		// Low efficiency recommendation
		if metrics.EfficiencyScore < ppm.alertThresholds.EfficiencyBelow {
			rec := OptimizationRecommendation{
				PoolName:             name,
				Type:                 "efficiency_improvement",
				Priority:             "high",
				Description:          fmt.Sprintf("Pool efficiency is low (%.1f%%). Comprehensive optimization needed.", metrics.EfficiencyScore*100),
				EstimatedImprovement: 30.0,
				ImplementationSteps: []string{
					"Analyze allocation patterns",
					"Optimize growth strategy",
					"Implement usage monitoring",
					"Consider pool specialization",
				},
				CreatedAt: time.Now(),
				Status:    "pending",
			}
			ppm.optimizations = append(ppm.optimizations, rec)
		}
	}
}

// GetPerformanceReport generates a comprehensive performance report
func (ppm *PoolPerformanceMonitor) GetPerformanceReport() PerformanceReport {
	ppm.mu.RLock()
	defer ppm.mu.RUnlock()

	report := PerformanceReport{
		GeneratedAt:      time.Now(),
		TotalPools:       len(ppm.pools),
		Recommendations:  make([]OptimizationRecommendation, len(ppm.optimizations)),
		BenchmarkResults: make([]BenchmarkResult, len(ppm.benchmarks)),
		PoolSummaries:    make([]PoolSummary, 0, len(ppm.pools)),
	}

	copy(report.Recommendations, ppm.optimizations)
	copy(report.BenchmarkResults, ppm.benchmarks)

	// Generate pool summaries
	var totalEfficiency float64
	var totalMemory int64

	for name, metrics := range ppm.pools {
		summary := PoolSummary{
			Name:                name,
			EfficiencyScore:     metrics.EfficiencyScore,
			UtilizationRatio:    metrics.UtilizationRatio,
			MemoryUsageKB:       metrics.EstimatedMemoryKB,
			AllocationRate:      metrics.AllocationRate,
			TotalAllocations:    metrics.TotalAllocations,
			RecommendationCount: 0,
		}

		// Count recommendations for this pool
		for _, rec := range ppm.optimizations {
			if rec.PoolName == name {
				summary.RecommendationCount++
			}
		}

		report.PoolSummaries = append(report.PoolSummaries, summary)
		totalEfficiency += metrics.EfficiencyScore
		totalMemory += metrics.EstimatedMemoryKB
	}

	if len(ppm.pools) > 0 {
		report.OverallEfficiency = totalEfficiency / float64(len(ppm.pools))
	}
	report.TotalMemoryUsageKB = totalMemory

	return report
}

// PerformanceReport contains comprehensive performance analysis
type PerformanceReport struct {
	GeneratedAt        time.Time                    `json:"generated_at"`
	TotalPools         int                          `json:"total_pools"`
	OverallEfficiency  float64                      `json:"overall_efficiency"`
	TotalMemoryUsageKB int64                        `json:"total_memory_usage_kb"`
	Recommendations    []OptimizationRecommendation `json:"recommendations"`
	BenchmarkResults   []BenchmarkResult            `json:"benchmark_results"`
	PoolSummaries      []PoolSummary                `json:"pool_summaries"`
}

// PoolSummary provides a high-level summary of pool performance
type PoolSummary struct {
	Name                string  `json:"name"`
	EfficiencyScore     float64 `json:"efficiency_score"`
	UtilizationRatio    float64 `json:"utilization_ratio"`
	MemoryUsageKB       int64   `json:"memory_usage_kb"`
	AllocationRate      float64 `json:"allocation_rate"`
	TotalAllocations    int64   `json:"total_allocations"`
	RecommendationCount int     `json:"recommendation_count"`
}

// runBenchmarks executes performance benchmarks
func (ppm *PoolPerformanceMonitor) runBenchmarks() {
	simpleSuite := NewSimpleBenchmarkSuite()

	// Run core pool benchmarks
	ctx := context.Background()
	coreResults := simpleSuite.RunCorePoolBenchmarks(ctx)

	// Run memory efficiency benchmarks
	memoryResults := simpleSuite.RunMemoryEfficiencyBenchmarks(ctx)

	// Store results
	ppm.mu.Lock()
	ppm.benchmarks = append(ppm.benchmarks, coreResults...)
	ppm.benchmarks = append(ppm.benchmarks, memoryResults...)
	ppm.mu.Unlock()
}

// runRegressionTests checks for performance regressions
func (ppm *PoolPerformanceMonitor) runRegressionTests() {
	// Implementation for regression testing
	// This would run stored regression tests and compare against baselines
}

// checkAlerts checks if any alert thresholds are exceeded
func (ppm *PoolPerformanceMonitor) checkAlerts() {
	// Implementation for alert checking
	// This would trigger alerts when thresholds are exceeded
}
