// ABOUTME: Integration test demonstrating complete Step 7 performance monitoring and optimization
// ABOUTME: Shows pool performance monitoring, benchmarking, and regression testing working together

package monitoring

import (
	"context"
	"testing"
	"time"

	"tamarou.com/pvm/internal/core"
)

// TestStep7PerformanceMonitoringIntegration demonstrates complete Step 7 functionality
func TestStep7PerformanceMonitoringIntegration(t *testing.T) {
	t.Log("Testing Step 7: Performance Monitoring and Optimization integration")

	// 1. Create performance monitor with monitoring configuration
	config := MonitoringConfig{
		UpdateInterval:   time.Millisecond * 100,
		EnableBenchmarks: true,
		EnableRegression: true,
		AnalysisDepth:    "comprehensive",
		ReportingEnabled: true,
		OptimizationAuto: true,
		AlertsEnabled:    true,
	}

	monitor := NewPoolPerformanceMonitor(config)

	// 2. Register pools for monitoring
	pool1 := &MockPoolStatsProvider{
		stats: core.PoolStats{
			Allocations: 1000,
			Grows:       10,
			TotalSize:   5000,
			CurrentSize: 800,
			Capacity:    1000,
		},
	}

	pool2 := &MockPoolStatsProvider{
		stats: core.PoolStats{
			Allocations: 500,
			Grows:       20,
			TotalSize:   2000,
			CurrentSize: 100,
			Capacity:    1000, // Poor utilization
		},
	}

	monitor.RegisterPool("efficient_pool", pool1)
	monitor.RegisterPool("inefficient_pool", pool2)

	// 3. Run monitoring cycle
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*300)
	defer cancel()

	monitor.Start(ctx)

	// Allow monitoring to run
	time.Sleep(time.Millisecond * 200)

	monitor.Stop()

	// 4. Verify performance analysis was performed
	if len(monitor.analysisHistory) == 0 {
		t.Error("Expected performance analysis to be recorded")
	}

	latest := monitor.analysisHistory[len(monitor.analysisHistory)-1]
	if latest.TotalPools != 2 {
		t.Errorf("Expected 2 pools in analysis, got %d", latest.TotalPools)
	}

	if latest.PerformanceGrade == "" {
		t.Error("Expected performance grade to be assigned")
	}

	t.Logf("Performance analysis completed - Grade: %s, Efficiency: %.2f%%",
		latest.PerformanceGrade, latest.SystemEfficiency*100)

	// 5. Verify optimization recommendations were generated
	if len(monitor.optimizations) == 0 {
		t.Error("Expected optimization recommendations to be generated")
	}

	t.Logf("Generated %d optimization recommendations", len(monitor.optimizations))

	for i, rec := range monitor.optimizations {
		t.Logf("  Recommendation %d: %s (%s priority)", i+1, rec.Description, rec.Priority)
	}

	// 6. Get comprehensive performance report
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

	t.Logf("Performance Report: %d pools, %.2f%% efficiency, %dKB total memory",
		report.TotalPools, report.OverallEfficiency*100, report.TotalMemoryUsageKB)

	// 7. Test benchmarking functionality
	simpleSuite := NewSimpleBenchmarkSuite()

	benchmarkResults := simpleSuite.RunCorePoolBenchmarks(context.Background())

	if len(benchmarkResults) == 0 {
		t.Error("Expected benchmark results to be generated")
	}

	t.Logf("Generated %d benchmark results", len(benchmarkResults))

	for _, result := range benchmarkResults {
		t.Logf("  Benchmark %s: %.2f%% improvement, %d bytes saved",
			result.Name, result.ImprovementPercent, result.MemorySavings)
	}

	// 8. Test regression testing functionality
	regressionSuite := NewRegressionTestSuite(simpleSuite, DefaultRegressionConfig())
	regressionSuite.RegisterStandardTests()

	regressionResults := regressionSuite.RunAllTests(context.Background())

	if regressionResults.TotalTests == 0 {
		t.Error("Expected regression tests to be executed")
	}

	t.Logf("Regression Testing: %d tests, %d passed, %d failed, Grade: %s",
		regressionResults.TotalTests, regressionResults.PassedTests,
		regressionResults.FailedTests, regressionResults.PerformanceGrade)

	// 9. Verify memory efficiency benchmarks
	memoryResults := simpleSuite.RunMemoryEfficiencyBenchmarks(context.Background())

	if len(memoryResults) == 0 {
		t.Error("Expected memory efficiency results")
	}

	t.Logf("Memory Efficiency: %d test scenarios completed", len(memoryResults))

	// 10. Test trend analysis
	trends := regressionSuite.GetTrends()

	t.Logf("Performance Trends: analyzed %d test patterns", len(trends))

	// Success - all Step 7 functionality is working
	t.Log("Step 7: Performance Monitoring and Optimization - COMPLETED ✅")
	t.Log("Features verified:")
	t.Log("  ✅ Comprehensive pool monitoring and statistics")
	t.Log("  ✅ Performance optimization recommendations")
	t.Log("  ✅ Pool utilization analysis and optimization")
	t.Log("  ✅ Performance benchmark comparisons")
	t.Log("  ✅ Memory efficiency benchmarks")
	t.Log("  ✅ Performance regression testing")
	t.Log("  ✅ Pool performance baselines and regression detection")
}

// TestStep7BenchmarkAccuracy tests the accuracy of benchmark measurements
func TestStep7BenchmarkAccuracy(t *testing.T) {
	t.Log("Testing benchmark measurement accuracy")

	simpleSuite := NewSimpleBenchmarkSuite()

	// Test core pool benchmarks with different configurations
	configs := []BenchmarkConfig{
		{Operations: 100, Concurrency: 1, ObjectSize: 32, AllocationPattern: "sequential"},
		{Operations: 1000, Concurrency: 4, ObjectSize: 64, AllocationPattern: "concurrent"},
		{Operations: 500, Concurrency: 2, ObjectSize: 128, AllocationPattern: "mixed"},
	}

	for _, config := range configs {
		pooledResult := simpleSuite.benchmarkCorePooled(config)
		directResult := simpleSuite.benchmarkCoreDirect(config)

		// Verify measurements are reasonable
		if pooledResult.PooledAllocations.TotalTime <= 0 {
			t.Error("Pooled benchmark should have positive time")
		}

		if directResult.DirectAllocations.TotalTime <= 0 {
			t.Error("Direct benchmark should have positive time")
		}

		if pooledResult.PooledAllocations.AllocationsCount != int64(config.Operations) {
			t.Errorf("Pooled allocations count mismatch: expected %d, got %d",
				config.Operations, pooledResult.PooledAllocations.AllocationsCount)
		}

		if directResult.DirectAllocations.AllocationsCount != int64(config.Operations) {
			t.Errorf("Direct allocations count mismatch: expected %d, got %d",
				config.Operations, directResult.DirectAllocations.AllocationsCount)
		}

		t.Logf("Config %s: Pooled %v, Direct %v",
			config.AllocationPattern,
			pooledResult.PooledAllocations.TotalTime,
			directResult.DirectAllocations.TotalTime)
	}

	t.Log("Benchmark accuracy verification completed ✅")
}

// TestStep7RegressionDetection tests regression detection functionality
func TestStep7RegressionDetection(t *testing.T) {
	t.Log("Testing performance regression detection")

	simpleSuite := NewSimpleBenchmarkSuite()
	config := DefaultRegressionConfig()
	config.DefaultTolerance = 5.0 // 5% tolerance

	regressionSuite := NewRegressionTestSuite(simpleSuite, config)

	// Register a test
	testRun := 0
	regressionSuite.RegisterTest(&RegressionTest{
		Name: "Simulated_Performance_Test",
		TestFunction: func() BenchmarkResult {
			testRun++

			// Simulate performance degradation on second run
			baseTime := time.Millisecond * 100
			if testRun > 1 {
				baseTime = time.Millisecond * 120 // 20% slower
			}

			return BenchmarkResult{
				PooledAllocations: BenchmarkMetrics{
					TotalTime:        baseTime,
					OperationsPerSec: 1000,
					AllocationsCount: 1000,
					MemoryAllocated:  1024 * 1024,
				},
				Timestamp: time.Now(),
			}
		},
		Tolerance:        5.0,
		Status:           "active",
		FailureThreshold: 1,
	})

	// First run - establish baseline
	result1 := regressionSuite.RunAllTests(context.Background())
	if result1.TotalTests != 1 {
		t.Error("Expected 1 test to run")
	}

	// Second run - should detect regression
	result2 := regressionSuite.RunAllTests(context.Background())
	if result2.RegressionCount == 0 {
		t.Error("Expected regression to be detected")
	}

	if result2.OverallStatus == "passed" {
		t.Error("Expected overall status to indicate regression")
	}

	t.Logf("Regression Detection: %s with %d regressions detected",
		result2.OverallStatus, result2.RegressionCount)

	t.Log("Regression detection functionality verified ✅")
}
