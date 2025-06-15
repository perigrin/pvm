// ABOUTME: Tests for performance analyzer functionality
// ABOUTME: Validates metrics collection, measurement accuracy, and system health

package performance

import (
	"strings"
	"testing"
	"time"
)

func TestAnalyzer_BasicMeasurement(t *testing.T) {
	analyzer := NewAnalyzer()
	
	// Measure a simple operation
	ctx := analyzer.StartMeasurement("test_operation")
	time.Sleep(10 * time.Millisecond) // Simulate work
	ctx.Finish()
	
	// Check metrics
	metrics := analyzer.GetMetrics()
	if len(metrics) != 1 {
		t.Errorf("Expected 1 metric, got %d", len(metrics))
	}
	
	metric := metrics["test_operation"]
	if metric == nil {
		t.Fatal("test_operation metric not found")
	}
	
	if metric.Count != 1 {
		t.Errorf("Expected count 1, got %d", metric.Count)
	}
	
	if metric.TotalTime < 10*time.Millisecond {
		t.Errorf("Expected at least 10ms, got %v", metric.TotalTime)
	}
}

func TestAnalyzer_MultipleMeasurements(t *testing.T) {
	analyzer := NewAnalyzer()
	
	// Perform multiple measurements
	for i := 0; i < 5; i++ {
		ctx := analyzer.StartMeasurement("repeated_operation")
		time.Sleep(5 * time.Millisecond)
		ctx.Finish()
	}
	
	metrics := analyzer.GetMetrics()
	metric := metrics["repeated_operation"]
	
	if metric.Count != 5 {
		t.Errorf("Expected count 5, got %d", metric.Count)
	}
	
	expectedAvg := metric.TotalTime / time.Duration(metric.Count)
	if metric.AvgTime != expectedAvg {
		t.Errorf("Average time calculation incorrect: got %v, expected %v", metric.AvgTime, expectedAvg)
	}
}

func TestAnalyzer_MinMaxTracking(t *testing.T) {
	analyzer := NewAnalyzer()
	
	// First measurement (will be both min and max initially)
	ctx1 := analyzer.StartMeasurement("minmax_test")
	time.Sleep(20 * time.Millisecond)
	ctx1.Finish()
	
	// Second measurement (shorter)
	ctx2 := analyzer.StartMeasurement("minmax_test") 
	time.Sleep(5 * time.Millisecond)
	ctx2.Finish()
	
	// Third measurement (longer)
	ctx3 := analyzer.StartMeasurement("minmax_test")
	time.Sleep(30 * time.Millisecond)
	ctx3.Finish()
	
	metrics := analyzer.GetMetrics()
	metric := metrics["minmax_test"]
	
	if metric.MinTime >= metric.MaxTime {
		t.Errorf("MinTime (%v) should be less than MaxTime (%v)", metric.MinTime, metric.MaxTime)
	}
	
	// Min should be around 5ms, max around 30ms
	if metric.MinTime > 15*time.Millisecond {
		t.Errorf("MinTime too large: %v", metric.MinTime)
	}
	
	if metric.MaxTime < 25*time.Millisecond {
		t.Errorf("MaxTime too small: %v", metric.MaxTime)
	}
}

func TestAnalyzer_SlowestOperations(t *testing.T) {
	analyzer := NewAnalyzer()
	
	// Create operations with different speeds
	operations := []struct {
		name     string
		duration time.Duration
	}{
		{"fast_op", 1 * time.Millisecond},
		{"slow_op", 50 * time.Millisecond},
		{"medium_op", 10 * time.Millisecond},
	}
	
	for _, op := range operations {
		ctx := analyzer.StartMeasurement(op.name)
		time.Sleep(op.duration)
		ctx.Finish()
	}
	
	slowest := analyzer.GetSlowestOperations(2)
	if len(slowest) != 2 {
		t.Errorf("Expected 2 slowest operations, got %d", len(slowest))
	}
	
	// Should be ordered by average time (descending)
	if slowest[0].Name != "slow_op" {
		t.Errorf("Expected slow_op first, got %s", slowest[0].Name)
	}
	
	if slowest[1].Name != "medium_op" {
		t.Errorf("Expected medium_op second, got %s", slowest[1].Name)
	}
}

func TestAnalyzer_EnableDisable(t *testing.T) {
	analyzer := NewAnalyzer()
	
	// Disable analyzer
	analyzer.SetEnabled(false)
	
	if analyzer.IsEnabled() {
		t.Error("Analyzer should be disabled")
	}
	
	// Measurements should not be recorded when disabled
	ctx := analyzer.StartMeasurement("disabled_test")
	time.Sleep(10 * time.Millisecond)
	ctx.Finish()
	
	metrics := analyzer.GetMetrics()
	if len(metrics) != 0 {
		t.Errorf("Expected no metrics when disabled, got %d", len(metrics))
	}
	
	// Re-enable
	analyzer.SetEnabled(true)
	
	ctx2 := analyzer.StartMeasurement("enabled_test")
	time.Sleep(5 * time.Millisecond)
	ctx2.Finish()
	
	metrics = analyzer.GetMetrics()
	if len(metrics) != 1 {
		t.Errorf("Expected 1 metric when enabled, got %d", len(metrics))
	}
}

func TestAnalyzer_Reset(t *testing.T) {
	analyzer := NewAnalyzer()
	
	// Add some metrics
	ctx := analyzer.StartMeasurement("test_reset")
	ctx.Finish()
	
	metrics := analyzer.GetMetrics()
	if len(metrics) != 1 {
		t.Errorf("Expected 1 metric before reset, got %d", len(metrics))
	}
	
	// Reset
	analyzer.Reset()
	
	metrics = analyzer.GetMetrics()
	if len(metrics) != 0 {
		t.Errorf("Expected no metrics after reset, got %d", len(metrics))
	}
}

func TestAnalyzer_ConcurrentAccess(t *testing.T) {
	analyzer := NewAnalyzer()
	
	// Run multiple measurements concurrently
	done := make(chan bool, 10)
	
	for i := 0; i < 10; i++ {
		go func(id int) {
			ctx := analyzer.StartMeasurement("concurrent_test")
			time.Sleep(time.Millisecond)
			ctx.Finish()
			done <- true
		}(i)
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
	
	metrics := analyzer.GetMetrics()
	metric := metrics["concurrent_test"]
	
	if metric == nil {
		t.Fatal("concurrent_test metric not found")
	}
	
	if metric.Count != 10 {
		t.Errorf("Expected count 10, got %d", metric.Count)
	}
}

func TestGlobalAnalyzer(t *testing.T) {
	// Test global analyzer functions
	ResetGlobalMetrics()
	
	ctx := Measure("global_test")
	time.Sleep(5 * time.Millisecond)
	ctx.Finish()
	
	metrics := GetGlobalMetrics()
	if len(metrics) != 1 {
		t.Errorf("Expected 1 global metric, got %d", len(metrics))
	}
	
	metric := metrics["global_test"]
	if metric == nil {
		t.Fatal("global_test metric not found")
	}
	
	if metric.Count != 1 {
		t.Errorf("Expected global count 1, got %d", metric.Count)
	}
	
	// Test disable
	SetGlobalAnalyzerEnabled(false)
	
	ctx2 := Measure("disabled_global_test")
	ctx2.Finish()
	
	metrics = GetGlobalMetrics()
	if len(metrics) != 1 { // Should still be just the first metric
		t.Errorf("Expected 1 metric after disable, got %d", len(metrics))
	}
	
	// Re-enable for other tests
	SetGlobalAnalyzerEnabled(true)
}

func TestAnalyzer_MemoryTracking(t *testing.T) {
	analyzer := NewAnalyzer()
	
	// Get memory stats
	memStats := analyzer.GetMemoryUsage()
	if memStats.Alloc == 0 {
		t.Error("Memory allocation should be greater than 0")
	}
	
	// Force GC and check
	afterStats := analyzer.ForceGC()
	
	// After GC, allocated memory might be lower (but not always guaranteed)
	// Just check that we can read the stats
	if afterStats.NumGC <= memStats.NumGC {
		t.Error("NumGC should increase after ForceGC")
	}
}

func TestMetricCopy(t *testing.T) {
	analyzer := NewAnalyzer()
	
	ctx := analyzer.StartMeasurement("copy_test")
	time.Sleep(5 * time.Millisecond)
	ctx.Finish()
	
	metrics1 := analyzer.GetMetrics()
	metrics2 := analyzer.GetMetrics()
	
	// Modify one copy
	metrics1["copy_test"].Count = 999
	
	// Other copy should be unchanged
	if metrics2["copy_test"].Count != 1 {
		t.Error("Metric copy not isolated - changes affected other copy")
	}
}

// Benchmark tests
func BenchmarkMeasurement(b *testing.B) {
	analyzer := NewAnalyzer()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx := analyzer.StartMeasurement("benchmark_test")
		ctx.Finish()
	}
}

func BenchmarkConcurrentMeasurement(b *testing.B) {
	analyzer := NewAnalyzer()
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ctx := analyzer.StartMeasurement("concurrent_benchmark")
			ctx.Finish()
		}
	})
}

// Test helper to ensure test name compliance
func TestCompliance(t *testing.T) {
	testName := t.Name()
	if !strings.HasPrefix(testName, "Test") {
		t.Errorf("Test name should start with 'Test', got: %s", testName)
	}
}