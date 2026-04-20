// ABOUTME: Tests for the generic arena-style Pool type in the core package
// ABOUTME: Validates allocation, slice allocation, reset, clear, concurrent access, and the PoolManager registry

package core

import (
	"sync"
	"testing"
)

func TestPool_New(t *testing.T) {
	pool := &Pool[int]{}

	// First allocation should trigger initial growth
	ptr1 := pool.New()
	if ptr1 == nil {
		t.Fatal("Expected non-nil pointer from New()")
	}

	stats := pool.Stats()
	if stats.Allocations != 1 {
		t.Errorf("Expected 1 allocation, got %d", stats.Allocations)
	}
	if stats.Grows != 1 {
		t.Errorf("Expected 1 growth, got %d", stats.Grows)
	}

	// Second allocation should not trigger growth
	ptr2 := pool.New()
	if ptr2 == nil {
		t.Fatal("Expected non-nil pointer from second New()")
	}
	if ptr1 == ptr2 {
		t.Error("Expected different pointers from separate allocations")
	}

	stats = pool.Stats()
	if stats.Allocations != 2 {
		t.Errorf("Expected 2 allocations, got %d", stats.Allocations)
	}
	if stats.Grows != 1 {
		t.Errorf("Expected 1 growth, got %d", stats.Grows)
	}
}

func TestPool_NewSlice(t *testing.T) {
	pool := &Pool[int]{}

	// Allocate small slice
	slice1 := pool.NewSlice(5)
	if len(slice1) != 5 {
		t.Errorf("Expected slice length 5, got %d", len(slice1))
	}
	if cap(slice1) != 5 {
		t.Errorf("Expected slice capacity 5, got %d", cap(slice1))
	}

	// Allocate another slice
	slice2 := pool.NewSlice(3)
	if len(slice2) != 3 {
		t.Errorf("Expected slice length 3, got %d", len(slice2))
	}

	stats := pool.Stats()
	if stats.Allocations != 2 {
		t.Errorf("Expected 2 allocations, got %d", stats.Allocations)
	}
}

func TestPool_NewSlice_LargeSize(t *testing.T) {
	pool := &Pool[int]{}

	// Allocate very large slice that should bypass pool
	largeSize := 1000000
	slice := pool.NewSlice(largeSize)
	if len(slice) != largeSize {
		t.Errorf("Expected slice length %d, got %d", largeSize, len(slice))
	}

	stats := pool.Stats()
	if stats.Allocations != 1 {
		t.Errorf("Expected 1 allocation, got %d", stats.Allocations)
	}
}

func TestPool_NewSlice_ZeroSize(t *testing.T) {
	pool := &Pool[int]{}

	slice := pool.NewSlice(0)
	if slice != nil {
		t.Error("Expected nil slice for zero size")
	}

	stats := pool.Stats()
	if stats.Allocations != 0 {
		t.Errorf("Expected 0 allocations, got %d", stats.Allocations)
	}
}

func TestPool_Reset(t *testing.T) {
	pool := &Pool[int]{}

	// Allocate some objects
	pool.New()
	pool.New()
	pool.NewSlice(5)

	initialStats := pool.Stats()
	if initialStats.CurrentSize == 0 {
		t.Error("Expected non-zero current size before reset")
	}

	// Reset pool
	pool.Reset()

	stats := pool.Stats()
	if stats.CurrentSize != 0 {
		t.Errorf("Expected zero current size after reset, got %d", stats.CurrentSize)
	}
	if stats.Capacity == 0 {
		t.Error("Expected non-zero capacity after reset")
	}
	if stats.Allocations != initialStats.Allocations {
		t.Error("Reset should not change allocation counter")
	}
}

func TestPool_Clear(t *testing.T) {
	pool := &Pool[int]{}

	// Allocate some objects
	pool.New()
	pool.NewSlice(5)

	// Clear pool
	pool.Clear()

	stats := pool.Stats()
	if stats.CurrentSize != 0 {
		t.Errorf("Expected zero current size after clear, got %d", stats.CurrentSize)
	}
	if stats.Capacity != 0 {
		t.Errorf("Expected zero capacity after clear, got %d", stats.Capacity)
	}
	if stats.Allocations != 0 {
		t.Errorf("Expected zero allocations after clear, got %d", stats.Allocations)
	}
}

func TestPool_ConcurrentAccess(t *testing.T) {
	pool := &Pool[int]{}
	const numGoroutines = 10
	const allocationsPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < allocationsPerGoroutine; j++ {
				ptr := pool.New()
				if ptr == nil {
					t.Error("Got nil pointer from concurrent allocation")
					return
				}
				*ptr = j
			}
		}()
	}

	wg.Wait()

	stats := pool.Stats()
	expected := int64(numGoroutines * allocationsPerGoroutine)
	if stats.Allocations != expected {
		t.Errorf("Expected %d allocations, got %d", expected, stats.Allocations)
	}
}

func TestNextPoolSize(t *testing.T) {
	tests := []struct {
		input    int
		expected int
	}{
		{0, 2},
		{1, 2},
		{2, 4},
		{4, 8},
		{8, 16},
		{100, 200},
		{256*1024 - 1, 256 * 1024}, // Should cap at max
		{256 * 1024, 256 * 1024},   // Should remain at max
		{300 * 1024, 256 * 1024},   // Should cap at max
	}

	for _, test := range tests {
		result := nextPoolSize(test.input)
		if result != test.expected {
			t.Errorf("nextPoolSize(%d) = %d, expected %d", test.input, result, test.expected)
		}
	}
}

func TestPoolManager(t *testing.T) {
	config := DefaultPoolConfig()
	config.StatsEnabled = true

	manager := NewPoolManager(config)

	// Create test pools
	pool1 := &Pool[int]{}
	pool2 := &Pool[string]{}

	// Register pools
	manager.RegisterPool("pool1", pool1)
	manager.RegisterPool("pool2", pool2)

	// Allocate from pools
	pool1.New()
	pool1.NewSlice(5)
	pool2.New()

	// Check individual stats
	stats1, exists := manager.GetPoolStats("pool1")
	if !exists {
		t.Error("Expected pool1 to exist in manager")
	}
	if stats1.Allocations != 2 {
		t.Errorf("Expected 2 allocations for pool1, got %d", stats1.Allocations)
	}

	stats2, exists := manager.GetPoolStats("pool2")
	if !exists {
		t.Error("Expected pool2 to exist in manager")
	}
	if stats2.Allocations != 1 {
		t.Errorf("Expected 1 allocation for pool2, got %d", stats2.Allocations)
	}

	// Check aggregate stats
	allStats := manager.GetAllStats()
	if len(allStats) != 2 {
		t.Errorf("Expected 2 pools in all stats, got %d", len(allStats))
	}

	aggregate := manager.GetAggregateStats()
	if aggregate.Allocations != 3 {
		t.Errorf("Expected 3 total allocations, got %d", aggregate.Allocations)
	}

	// Unregister pool
	manager.UnregisterPool("pool1")
	_, exists = manager.GetPoolStats("pool1")
	if exists {
		t.Error("Expected pool1 to be unregistered")
	}
}

func TestGlobalPoolManager(t *testing.T) {
	// Test global pool manager
	pool := &Pool[int]{}
	RegisterGlobalPool("test-pool", pool)

	// Allocate something
	pool.New()

	// Check global stats
	globalStats := GetGlobalStats()
	if _, exists := globalStats["test-pool"]; !exists {
		t.Error("Expected test-pool in global stats")
	}

	if globalStats["test-pool"].Allocations != 1 {
		t.Errorf("Expected 1 allocation in global stats, got %d", globalStats["test-pool"].Allocations)
	}
}

func TestDefaultPoolConfig(t *testing.T) {
	config := DefaultPoolConfig()

	if config.InitialSize <= 0 {
		t.Error("Expected positive initial size")
	}
	if config.MaxSize <= 0 {
		t.Error("Expected positive max size")
	}
	if config.GrowthFactor <= 1.0 {
		t.Error("Expected growth factor > 1.0")
	}
	if !config.StatsEnabled {
		t.Error("Expected stats enabled by default")
	}
}

func BenchmarkPool_New(b *testing.B) {
	pool := &Pool[int]{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ptr := pool.New()
		*ptr = i
	}
}

func BenchmarkPool_NewSlice(b *testing.B) {
	pool := &Pool[int]{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		slice := pool.NewSlice(10)
		for j := 0; j < len(slice); j++ {
			slice[j] = i + j
		}
	}
}

func BenchmarkPool_ConcurrentNew(b *testing.B) {
	pool := &Pool[int]{}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ptr := pool.New()
			*ptr = 42
		}
	})
}
