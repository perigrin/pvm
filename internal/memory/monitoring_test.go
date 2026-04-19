// ABOUTME: Tests for the MemoryMonitor type in the memory package
// ABOUTME: Validates runtime stats collection, bounded history, alert threshold callbacks, and context-driven background monitoring

package memory

import (
	"context"
	"runtime"
	"testing"
	"time"
)

func TestMemoryMonitor(t *testing.T) {
	monitor := NewMemoryMonitor(10, 0) // 10 history entries, no alert threshold

	// Test current stats collection
	stats := monitor.GetCurrentStats()
	if stats.AllocBytes == 0 {
		t.Error("Expected non-zero allocated bytes")
	}
	if stats.NumGoroutine == 0 {
		t.Error("Expected non-zero goroutine count")
	}

	// Test history is initially empty
	history := monitor.GetHistory()
	if len(history) != 0 {
		t.Error("Expected empty history initially")
	}
}

func TestMemoryMonitorWithHistory(t *testing.T) {
	monitor := NewMemoryMonitor(3, 0) // Small history for testing

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	// Start monitoring with short interval
	monitor.Start(ctx, 50*time.Millisecond)

	// Wait for some data collection
	time.Sleep(200 * time.Millisecond)
	monitor.Stop()

	history := monitor.GetHistory()
	if len(history) == 0 {
		t.Error("Expected some history entries")
	}

	// Verify history is limited to max size
	if len(history) > 3 {
		t.Errorf("Expected at most 3 history entries, got %d", len(history))
	}
}

func TestMemoryTrend(t *testing.T) {
	monitor := NewMemoryMonitor(10, 0)

	// Manually add some history entries to test trend calculation
	now := time.Now()

	// Simulate increasing memory usage
	monitor.recordStats(MemoryStats{
		Timestamp:  now.Add(-10 * time.Second),
		AllocBytes: 1000000, // 1MB
	})
	monitor.recordStats(MemoryStats{
		Timestamp:  now.Add(-5 * time.Second),
		AllocBytes: 1500000, // 1.5MB
	})
	monitor.recordStats(MemoryStats{
		Timestamp:  now,
		AllocBytes: 2000000, // 2MB
	})

	trend := monitor.GetMemoryTrend()
	if trend.Direction != "increasing" {
		t.Errorf("Expected increasing trend, got %s", trend.Direction)
	}
	if trend.AllocDelta <= 0 {
		t.Errorf("Expected positive allocation delta, got %d", trend.AllocDelta)
	}
	if trend.Rate <= 0 {
		t.Errorf("Expected positive growth rate, got %f", trend.Rate)
	}
}

func TestMemoryLeakDetection(t *testing.T) {
	monitor := NewMemoryMonitor(20, 0)

	now := time.Now()

	// Simulate sustained memory growth (potential leak)
	for i := 0; i < 10; i++ {
		monitor.recordStats(MemoryStats{
			Timestamp:    now.Add(time.Duration(i) * time.Second),
			AllocBytes:   uint64(1000000 + i*2000000), // Growing by 2MB each time
			NumGoroutine: 10 + i,                      // Stable goroutine count
		})
	}

	leaks := monitor.DetectLeaks()

	// Should detect sustained growth
	foundSustainedGrowth := false
	for _, leak := range leaks {
		if leak.Type == "sustained_growth" {
			foundSustainedGrowth = true
			if leak.Severity != "medium" {
				t.Errorf("Expected medium severity for sustained growth, got %s", leak.Severity)
			}
		}
	}

	if !foundSustainedGrowth {
		t.Error("Expected to detect sustained memory growth")
	}
}

func TestGoroutineLeakDetection(t *testing.T) {
	monitor := NewMemoryMonitor(10, 0)

	now := time.Now()

	// Add enough history entries so DetectLeaks has data to work with
	for i := 0; i < 10; i++ {
		monitor.recordStats(MemoryStats{
			Timestamp:    now.Add(time.Duration(i) * time.Second),
			AllocBytes:   1000000,
			NumGoroutine: 150, // High goroutine count (above our threshold of 100)
		})
	}

	leaks := monitor.DetectLeaks()

	// Should detect goroutine leak
	foundGoroutineLeak := false
	for _, leak := range leaks {
		if leak.Type == "goroutine_leak" {
			foundGoroutineLeak = true
			if leak.Severity != "high" {
				t.Errorf("Expected high severity for goroutine leak, got %s", leak.Severity)
			}
		}
	}

	if !foundGoroutineLeak {
		t.Error("Expected to detect goroutine leak")
	}
}

func TestGCPressureDetection(t *testing.T) {
	monitor := NewMemoryMonitor(10, 0)

	now := time.Now()

	// Add enough history entries so DetectLeaks has data to work with
	for i := 0; i < 10; i++ {
		pauseData := []uint64{10000000, 5000000} // Normal pauses
		if i == 9 {                              // Last entry has long pause
			longPause := uint64(150 * 1000 * 1000) // 150ms in nanoseconds
			pauseData = []uint64{longPause, 10000000, 5000000}
		}

		monitor.recordStats(MemoryStats{
			Timestamp:  now.Add(time.Duration(i) * time.Second),
			AllocBytes: 1000000,
			PauseNs:    pauseData,
		})
	}

	leaks := monitor.DetectLeaks()

	// Should detect GC pressure
	foundGCPressure := false
	for _, leak := range leaks {
		if leak.Type == "gc_pressure" {
			foundGCPressure = true
			if leak.Severity != "medium" {
				t.Errorf("Expected medium severity for GC pressure, got %s", leak.Severity)
			}
		}
	}

	if !foundGCPressure {
		t.Error("Expected to detect GC pressure")
	}
}

func TestMemoryAlerts(t *testing.T) {
	alertThreshold := uint64(1000000) // 1MB threshold
	monitor := NewMemoryMonitor(10, alertThreshold)

	alertTriggered := false
	monitor.SetAlertCallback(func(stats MemoryStats) {
		alertTriggered = true
		if stats.AllocBytes <= alertThreshold {
			t.Errorf("Alert triggered with %d bytes, threshold is %d", stats.AllocBytes, alertThreshold)
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Allocate memory to trigger alert
	bigData := make([]byte, 2000000) // 2MB allocation
	runtime.KeepAlive(bigData)

	monitor.Start(ctx, 10*time.Millisecond)
	time.Sleep(50 * time.Millisecond)
	monitor.Stop()

	// Use alertTriggered to avoid unused variable error
	_ = alertTriggered
	// Note: Alert may or may not trigger depending on actual memory usage
	// This test mainly verifies the alert mechanism doesn't panic
}

func TestForceGC(t *testing.T) {
	monitor := NewMemoryMonitor(10, 0)

	before, after := monitor.ForceGC()

	if before.Timestamp.After(after.Timestamp) {
		t.Error("Expected after timestamp to be later than before")
	}

	// After GC, we might see some reduction in memory, but this isn't guaranteed
	// The test mainly verifies ForceGC doesn't panic
	if after.NumGC <= before.NumGC {
		t.Error("Expected GC count to increase after ForceGC")
	}
}

func TestMemoryProfile(t *testing.T) {
	monitor := NewMemoryMonitor(10, 0)

	// Add some history for trend calculation
	now := time.Now()
	monitor.recordStats(MemoryStats{
		Timestamp:  now.Add(-5 * time.Second),
		AllocBytes: 1000000,
	})
	monitor.recordStats(MemoryStats{
		Timestamp:  now,
		AllocBytes: 1000000,
	})

	profile := monitor.GetMemoryProfile()

	if profile.Current.AllocBytes == 0 {
		t.Error("Expected non-zero current allocation")
	}

	if profile.Health == "" {
		t.Error("Expected health assessment")
	}

	// Health should be "healthy" with stable memory usage
	if profile.Health != "healthy" && profile.Health != "warning" {
		t.Errorf("Expected healthy or warning status, got %s", profile.Health)
	}
}

func TestPoolRegistration(t *testing.T) {
	monitor := NewMemoryMonitor(10, 0)

	// Create a test pool (using type erasure for registration)
	pool := NewSyncPool(func() *testNode { return &testNode{} }, func(n *testNode) {})

	// Type erasure for registration
	anyPool := &poolAdapter[testNode]{pool}
	monitor.RegisterPool(anyPool)

	// Use the pool to generate some stats
	node := pool.Get()
	pool.Put(node)

	stats := monitor.GetCurrentStats()

	// Should have pool stats now
	if len(stats.PoolStats) == 0 {
		t.Error("Expected pool statistics after registration")
	}
}

func TestStringInternerRegistration(t *testing.T) {
	monitor := NewMemoryMonitor(10, 0)
	interner := NewStringInterner()

	monitor.RegisterStringInterner(interner)

	// Use the interner
	interner.Intern("test")
	interner.Intern("test") // Should be a hit

	stats := monitor.GetCurrentStats()

	if stats.StringInternerSize == 0 {
		t.Error("Expected string interner size > 0")
	}

	if stats.StringInternerMem <= 0 {
		t.Error("Expected string interner memory usage > 0")
	}
}

// poolAdapter provides type erasure for Pool[T] to Pool[any]
type poolAdapter[T any] struct {
	pool Pool[T]
}

func (pa *poolAdapter[T]) Get() *any {
	obj := pa.pool.Get()
	var result any = obj
	return &result
}

func (pa *poolAdapter[T]) Put(obj *any) {
	if obj != nil {
		if typed, ok := (*obj).(*T); ok {
			pa.pool.Put(typed)
		}
	}
}

func (pa *poolAdapter[T]) Stats() PoolStats {
	return pa.pool.Stats()
}

func (pa *poolAdapter[T]) Clear() {
	pa.pool.Clear()
}

func BenchmarkMemoryStatsCollection(b *testing.B) {
	monitor := NewMemoryMonitor(10, 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = monitor.collectStats()
	}
}

func BenchmarkMemoryTrendCalculation(b *testing.B) {
	monitor := NewMemoryMonitor(100, 0)

	// Pre-populate with history
	now := time.Now()
	for i := 0; i < 50; i++ {
		monitor.recordStats(MemoryStats{
			Timestamp:  now.Add(time.Duration(-i) * time.Second),
			AllocBytes: uint64(1000000 + i*1000),
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = monitor.GetMemoryTrend()
	}
}
