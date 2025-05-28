// ABOUTME: Comprehensive tests for pool performance monitoring system
// ABOUTME: Tests monitoring capabilities, benchmarking, and optimization recommendations

package monitoring

import (
	"context"
	"fmt"
	"testing"
	"time"

	"tamarou.com/pvm/internal/core"
)

// MockPoolStatsProvider implements core.PoolStatsProvider for testing
type MockPoolStatsProvider struct {
	stats core.PoolStats
}

func (m *MockPoolStatsProvider) Stats() core.PoolStats {
	return m.stats
}

func TestPoolPerformanceMonitor_RegisterPool(t *testing.T) {
	config := MonitoringConfig{
		UpdateInterval:   time.Second,
		EnableBenchmarks: false,
		EnableRegression: false,
		AnalysisDepth:    "basic",
	}
	monitor := NewPoolPerformanceMonitor(config)

	mockPool := &MockPoolStatsProvider{
		stats: core.PoolStats{
			Allocations: 100,
			Grows:       5,
			TotalSize:   1000,
			CurrentSize: 800,
			Capacity:    1024,
		},
	}

	monitor.RegisterPool("test_pool", mockPool)

	// Verify pool was registered
	monitor.mu.RLock()
	metrics, exists := monitor.pools["test_pool"]
	monitor.mu.RUnlock()

	if !exists {
		t.Fatal("Pool was not registered")
	}

	if metrics.Name != "test_pool" {
		t.Errorf("Expected pool name 'test_pool', got '%s'", metrics.Name)
	}

	if metrics.TotalAllocations != 100 {
		t.Errorf("Expected 100 allocations, got %d", metrics.TotalAllocations)
	}
}

func TestPoolPerformanceMonitor_EfficiencyCalculation(t *testing.T) {
	config := MonitoringConfig{
		UpdateInterval: time.Second,
	}
	monitor := NewPoolPerformanceMonitor(config)

	// Test high efficiency scenario
	highEfficiencyMetrics := &PoolMetrics{
		UtilizationRatio:  0.9, // 90% utilization
		EstimatedMemoryKB: 1000,
		WastedMemoryKB:    100, // 10% waste
		TotalAllocations:  1000,
		TotalGrows:        50, // 5% growth rate
	}

	efficiency := monitor.calculateEfficiencyScore(highEfficiencyMetrics)
	if efficiency < 0.8 {
		t.Errorf("Expected high efficiency (>0.8), got %.3f", efficiency)
	}

	// Test low efficiency scenario
	lowEfficiencyMetrics := &PoolMetrics{
		UtilizationRatio:  0.2, // 20% utilization
		EstimatedMemoryKB: 1000,
		WastedMemoryKB:    800, // 80% waste
		TotalAllocations:  1000,
		TotalGrows:        200, // 20% growth rate
	}

	efficiency = monitor.calculateEfficiencyScore(lowEfficiencyMetrics)
	if efficiency > 0.4 {
		t.Errorf("Expected low efficiency (<0.4), got %.3f", efficiency)
	}
}

func TestPoolPerformanceMonitor_RecommendationGeneration(t *testing.T) {
	config := MonitoringConfig{
		UpdateInterval: time.Second,
	}
	monitor := NewPoolPerformanceMonitor(config)
	monitor.alertThresholds = AlertThresholds{
		UtilizationBelow:    0.5,
		MemoryWasteAbove:    50, // Lower threshold for test
		AllocationRateAbove: 100,
		EfficiencyBelow:     0.6,
	}

	// Create pools with various issues
	lowUtilizationPool := &MockPoolStatsProvider{
		stats: core.PoolStats{
			Allocations: 100,
			CurrentSize: 200,
			Capacity:    1000, // Very low utilization
		},
	}

	highWastePool := &MockPoolStatsProvider{
		stats: core.PoolStats{
			Allocations: 500,
			CurrentSize: 100,
			Capacity:    1000, // High waste
		},
	}

	monitor.RegisterPool("low_util_pool", lowUtilizationPool)
	monitor.RegisterPool("high_waste_pool", highWastePool)

	// Update pools to set utilization and waste metrics
	monitor.UpdatePool("low_util_pool", lowUtilizationPool)
	monitor.UpdatePool("high_waste_pool", highWastePool)

	// Generate recommendations
	monitor.generateRecommendations()

	if len(monitor.optimizations) == 0 {
		t.Error("Expected optimization recommendations to be generated")
	}

	// Check for specific recommendation types
	hasUtilizationRec := false
	hasMemoryRec := false

	for _, rec := range monitor.optimizations {
		switch rec.Type {
		case "pool_sizing":
			hasUtilizationRec = true
		case "memory_optimization":
			hasMemoryRec = true
		}
	}

	if !hasUtilizationRec {
		t.Error("Expected pool sizing recommendation")
	}
	if !hasMemoryRec {
		t.Error("Expected memory optimization recommendation")
	}
}

func TestPoolPerformanceMonitor_PerformanceReport(t *testing.T) {
	config := MonitoringConfig{
		UpdateInterval: time.Second,
	}
	monitor := NewPoolPerformanceMonitor(config)

	// Add some pools
	pool1 := &MockPoolStatsProvider{
		stats: core.PoolStats{
			Allocations: 1000,
			CurrentSize: 800,
			Capacity:    1000,
		},
	}

	pool2 := &MockPoolStatsProvider{
		stats: core.PoolStats{
			Allocations: 500,
			CurrentSize: 200,
			Capacity:    1000,
		},
	}

	monitor.RegisterPool("efficient_pool", pool1)
	monitor.RegisterPool("inefficient_pool", pool2)

	// Update metrics
	monitor.UpdatePool("efficient_pool", pool1)
	monitor.UpdatePool("inefficient_pool", pool2)

	// Generate recommendations
	monitor.generateRecommendations()

	// Get performance report
	report := monitor.GetPerformanceReport()

	if report.TotalPools != 2 {
		t.Errorf("Expected 2 pools in report, got %d", report.TotalPools)
	}

	if len(report.PoolSummaries) != 2 {
		t.Errorf("Expected 2 pool summaries, got %d", len(report.PoolSummaries))
	}

	if report.OverallEfficiency == 0 {
		t.Error("Expected non-zero overall efficiency")
	}

	if len(report.Recommendations) == 0 {
		t.Error("Expected recommendations in report")
	}
}

func TestPoolPerformanceMonitor_Analysis(t *testing.T) {
	config := MonitoringConfig{
		UpdateInterval: time.Second,
	}
	monitor := NewPoolPerformanceMonitor(config)

	// Add pools with different characteristics
	pools := map[string]*MockPoolStatsProvider{
		"good_pool": {
			stats: core.PoolStats{
				Allocations: 1000,
				CurrentSize: 900,
				Capacity:    1000,
			},
		},
		"bad_pool": {
			stats: core.PoolStats{
				Allocations: 100,
				CurrentSize: 50,
				Capacity:    1000,
			},
		},
		"critical_pool": {
			stats: core.PoolStats{
				Allocations: 50,
				CurrentSize: 10,
				Capacity:    1000,
			},
		},
	}

	for name, pool := range pools {
		monitor.RegisterPool(name, pool)
		monitor.UpdatePool(name, pool)
	}

	// Run analysis
	monitor.performAnalysis()

	if len(monitor.analysisHistory) == 0 {
		t.Error("Expected analysis history to be recorded")
	}

	latest := monitor.analysisHistory[len(monitor.analysisHistory)-1]

	if latest.TotalPools != 3 {
		t.Errorf("Expected 3 pools in analysis, got %d", latest.TotalPools)
	}

	if latest.CriticalIssues == 0 {
		t.Error("Expected critical issues to be detected")
	}

	if latest.PerformanceGrade == "" {
		t.Error("Expected performance grade to be assigned")
	}
}

func TestPoolPerformanceMonitor_StartStop(t *testing.T) {
	config := MonitoringConfig{
		UpdateInterval:   time.Millisecond * 100,
		EnableBenchmarks: false,
		EnableRegression: false,
	}
	monitor := NewPoolPerformanceMonitor(config)

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*300)
	defer cancel()

	// Start monitoring
	monitor.Start(ctx)

	// Add a pool
	pool := &MockPoolStatsProvider{
		stats: core.PoolStats{
			Allocations: 100,
			CurrentSize: 80,
			Capacity:    100,
		},
	}
	monitor.RegisterPool("test_pool", pool)

	// Wait for analysis to run
	time.Sleep(time.Millisecond * 200)

	// Stop monitoring
	monitor.Stop()

	// Verify analysis was performed
	if len(monitor.analysisHistory) == 0 {
		t.Error("Expected analysis to be performed during monitoring")
	}
}

func TestDefaultAlertThresholds(t *testing.T) {
	thresholds := DefaultAlertThresholds()

	if thresholds.UtilizationBelow <= 0 || thresholds.UtilizationBelow >= 1 {
		t.Errorf("Invalid utilization threshold: %f", thresholds.UtilizationBelow)
	}

	if thresholds.MemoryWasteAbove <= 0 {
		t.Errorf("Invalid memory waste threshold: %d", thresholds.MemoryWasteAbove)
	}

	if thresholds.AllocationRateAbove <= 0 {
		t.Errorf("Invalid allocation rate threshold: %f", thresholds.AllocationRateAbove)
	}

	if thresholds.EfficiencyBelow <= 0 || thresholds.EfficiencyBelow >= 1 {
		t.Errorf("Invalid efficiency threshold: %f", thresholds.EfficiencyBelow)
	}

	if thresholds.RegressionPercent <= 0 {
		t.Errorf("Invalid regression threshold: %f", thresholds.RegressionPercent)
	}
}

func TestMonitoringConfig_Defaults(t *testing.T) {
	config := MonitoringConfig{
		UpdateInterval: time.Second,
	}

	monitor := NewPoolPerformanceMonitor(config)

	if monitor.config.UpdateInterval != time.Second {
		t.Errorf("Expected update interval of 1s, got %v", monitor.config.UpdateInterval)
	}

	if monitor.maxHistory == 0 {
		t.Error("Expected non-zero max history")
	}
}

// Benchmark tests for performance monitoring overhead
func BenchmarkPoolPerformanceMonitor_UpdatePool(b *testing.B) {
	config := MonitoringConfig{
		UpdateInterval: time.Second,
	}
	monitor := NewPoolPerformanceMonitor(config)

	pool := &MockPoolStatsProvider{
		stats: core.PoolStats{
			Allocations: 1000,
			CurrentSize: 800,
			Capacity:    1000,
		},
	}

	monitor.RegisterPool("test_pool", pool)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Simulate changing stats
		pool.stats.Allocations = int64(i)
		pool.stats.CurrentSize = int64(i % 1000)

		monitor.UpdatePool("test_pool", pool)
	}
}

func BenchmarkPoolPerformanceMonitor_GenerateRecommendations(b *testing.B) {
	config := MonitoringConfig{
		UpdateInterval: time.Second,
	}
	monitor := NewPoolPerformanceMonitor(config)

	// Add multiple pools with various characteristics
	for i := 0; i < 10; i++ {
		pool := &MockPoolStatsProvider{
			stats: core.PoolStats{
				Allocations: int64(i * 100),
				CurrentSize: int64(i * 50),
				Capacity:    1000,
			},
		}
		monitor.RegisterPool(fmt.Sprintf("pool_%d", i), pool)
		monitor.UpdatePool(fmt.Sprintf("pool_%d", i), pool)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		monitor.generateRecommendations()
	}
}

func BenchmarkPoolPerformanceMonitor_PerformAnalysis(b *testing.B) {
	config := MonitoringConfig{
		UpdateInterval: time.Second,
	}
	monitor := NewPoolPerformanceMonitor(config)

	// Add many pools to test analysis performance
	for i := 0; i < 100; i++ {
		pool := &MockPoolStatsProvider{
			stats: core.PoolStats{
				Allocations: int64(i * 10),
				CurrentSize: int64(i * 8),
				Capacity:    int64(i * 10),
			},
		}
		monitor.RegisterPool(fmt.Sprintf("pool_%d", i), pool)
		monitor.UpdatePool(fmt.Sprintf("pool_%d", i), pool)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		monitor.performAnalysis()
	}
}
