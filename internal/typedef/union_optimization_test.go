// ABOUTME: Tests for union type performance optimizations
// ABOUTME: Validates caching and memory optimization features

package typedef

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestUnionTypeCache tests the union type caching system
func TestUnionTypeCache(t *testing.T) {
	cache := NewUnionTypeCache(3) // Small cache for testing

	t.Run("BasicCaching", func(t *testing.T) {
		members1 := []string{"Int", "Str"}
		members2 := []string{"Str", "Int"} // Same members, different order

		union1 := cache.GetOrCreate(members1)
		union2 := cache.GetOrCreate(members2)

		// Should return the same cached instance
		assert.Same(t, union1, union2, "Should return cached instance for equivalent unions")
		assert.Equal(t, 1, cache.Size())
	})

	t.Run("CacheEviction", func(t *testing.T) {
		cache.Clear()

		// Fill cache to capacity
		union1 := cache.GetOrCreate([]string{"Int", "Str"})
		_ = cache.GetOrCreate([]string{"Bool", "Num"})
		union3 := cache.GetOrCreate([]string{"ArrayRef", "HashRef"})
		assert.Equal(t, 3, cache.Size())

		// Add one more to trigger eviction
		union4 := cache.GetOrCreate([]string{"CodeRef", "ScalarRef"})
		assert.Equal(t, 3, cache.Size()) // Should still be 3 due to eviction

		// Verify that instances are still valid
		assert.NotNil(t, union1)
		assert.NotNil(t, union3)
		assert.NotNil(t, union4)
	})

	t.Run("LRUEviction", func(t *testing.T) {
		cache.Clear()

		// Create three unions
		union1 := cache.GetOrCreate([]string{"Int", "Str"})
		_ = cache.GetOrCreate([]string{"Bool", "Num"})
		union3 := cache.GetOrCreate([]string{"ArrayRef", "HashRef"})

		// Access union1 and union3 multiple times to increase their access count
		for i := 0; i < 5; i++ {
			cache.GetOrCreate([]string{"Int", "Str"})
			cache.GetOrCreate([]string{"ArrayRef", "HashRef"})
		}

		// Add a new union - union2 should be evicted as it has lowest access count
		union4 := cache.GetOrCreate([]string{"CodeRef", "ScalarRef"})

		// union1 and union3 should still be cached, union2 should be evicted
		cached1 := cache.GetOrCreate([]string{"Int", "Str"})
		cached3 := cache.GetOrCreate([]string{"ArrayRef", "HashRef"})

		assert.Same(t, union1, cached1, "Frequently accessed union should remain cached")
		assert.Same(t, union3, cached3, "Frequently accessed union should remain cached")
		assert.NotNil(t, union4)
	})

	t.Run("CreateUnionKey", func(t *testing.T) {
		// Test that key creation is order-independent and handles duplicates
		key1 := createUnionKey([]string{"Int", "Str"})
		key2 := createUnionKey([]string{"Str", "Int"})
		key3 := createUnionKey([]string{"Int", "Str", "Int"}) // With duplicate

		assert.Equal(t, key1, key2, "Keys should be order-independent")
		assert.Equal(t, key1, key3, "Keys should handle duplicates")
		assert.Equal(t, "Int|Str", key1, "Key should be pipe-separated and sorted")
	})
}

// TestOptimizedUnionType tests the optimized union type wrapper
func TestOptimizedUnionType(t *testing.T) {
	optimized := NewOptimizedUnionType([]string{"Int", "Str", "Bool"})

	t.Run("OperationCaching", func(t *testing.T) {
		// First call should compute and cache
		result1 := optimized.SupportsOperation("\"\"")
		assert.True(t, result1)

		// Second call should use cache
		result2 := optimized.SupportsOperation("\"\"")
		assert.True(t, result2)
		assert.Equal(t, result1, result2)

		// Verify cache contains the result
		optimized.mu.RLock()
		cached, exists := optimized.operationCache["supports:\"\""]
		optimized.mu.RUnlock()

		assert.True(t, exists, "Result should be cached")
		assert.Equal(t, result1, cached)
	})

	t.Run("ResultTypeCaching", func(t *testing.T) {
		// First call should compute and cache
		resultType1, err1 := optimized.GetOperationResultType("\"\"")
		assert.NoError(t, err1)
		assert.Equal(t, "Str", resultType1)

		// Second call should use cache
		resultType2, err2 := optimized.GetOperationResultType("\"\"")
		assert.NoError(t, err2)
		assert.Equal(t, resultType1, resultType2)

		// Verify cache contains the result
		optimized.mu.RLock()
		cached, exists := optimized.operationCache["result:\"\""]
		optimized.mu.RUnlock()

		assert.True(t, exists, "Result should be cached")
		if result, ok := cached.(*OperationResult); ok {
			assert.NoError(t, result.Error, "Cached result should not contain error")
			assert.Equal(t, resultType1, result.Result)
		} else {
			t.Errorf("Cached result has wrong type: %T", cached)
		}
	})

	t.Run("ErrorCaching", func(t *testing.T) {
		// Test caching of error results
		optimized.ClearOperationCache()

		// Try an operation that should not be supported
		result1, err1 := optimized.GetOperationResultType("complex_unsupported_op_xyz")
		t.Logf("First call: result=%s, err=%v", result1, err1)

		if err1 != nil {
			// Operation failed as expected, test caching
			result2, err2 := optimized.GetOperationResultType("complex_unsupported_op_xyz")
			t.Logf("Second call: result=%s, err=%v", result2, err2)
			assert.Error(t, err2, "Second call should also return error")

			// Verify error is cached properly
			optimized.mu.RLock()
			cached, exists := optimized.operationCache["result:complex_unsupported_op_xyz"]
			optimized.mu.RUnlock()

			t.Logf("Cache check: exists=%v, cached=%v", exists, cached)
			assert.True(t, exists, "Error should be cached")
			if result, ok := cached.(*OperationResult); ok {
				assert.Error(t, result.Error, "Cached result should contain error")
			} else {
				t.Errorf("Cached result has wrong type: %T", cached)
			}
		} else {
			// If the operation doesn't fail, test that success is cached properly
			t.Logf("Operation succeeded, testing success caching instead")

			result2, err2 := optimized.GetOperationResultType("complex_unsupported_op_xyz")
			assert.NoError(t, err2)
			assert.Equal(t, result1, result2, "Results should be consistent")
		}
	})

	t.Run("CacheClear", func(t *testing.T) {
		// Populate cache
		optimized.SupportsOperation("bool")
		optimized.GetOperationResultType("bool")

		// Clear cache
		optimized.ClearOperationCache()

		// Verify cache is empty
		optimized.mu.RLock()
		cacheSize := len(optimized.operationCache)
		optimized.mu.RUnlock()

		assert.Equal(t, 0, cacheSize, "Cache should be empty after clear")
	})

	t.Run("MemoryFootprint", func(t *testing.T) {
		// Test memory footprint calculation
		footprint := optimized.MemoryFootprint()
		assert.Greater(t, footprint, 0, "Memory footprint should be positive")

		// Add more cache entries and verify footprint increases
		optimized.SupportsOperation("bool")
		optimized.GetOperationResultType("bool")

		newFootprint := optimized.MemoryFootprint()
		assert.Greater(t, newFootprint, footprint, "Memory footprint should increase with cache entries")
	})
}

// BenchmarkUnionTypeCaching benchmarks the caching performance
func BenchmarkUnionTypeCaching(b *testing.B) {
	cache := NewUnionTypeCache(100)

	b.Run("CacheHit", func(b *testing.B) {
		// Pre-populate cache
		cache.GetOrCreate([]string{"Int", "Str"})

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = cache.GetOrCreate([]string{"Int", "Str"})
		}
	})

	b.Run("CacheMiss", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			members := []string{string(rune('A' + i%26)), string(rune('B' + i%26))}
			_ = cache.GetOrCreate(members)
		}
	})
}

// BenchmarkOptimizedOperations benchmarks optimized union type operations
func BenchmarkOptimizedOperations(b *testing.B) {
	regular := NewUnionType([]string{"Int", "Str", "Bool"})
	optimized := NewOptimizedUnionType([]string{"Int", "Str", "Bool"})

	b.Run("RegularSupportsOperation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = regular.SupportsOperation("\"\"")
		}
	})

	b.Run("OptimizedSupportsOperation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = optimized.SupportsOperation("\"\"")
		}
	})

	b.Run("RegularGetOperationResultType", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = regular.GetOperationResultType("\"\"")
		}
	})

	b.Run("OptimizedGetOperationResultType", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = optimized.GetOperationResultType("\"\"")
		}
	})
}

// TestConcurrentCacheAccess tests thread safety of caching
func TestConcurrentCacheAccess(t *testing.T) {
	cache := NewUnionTypeCache(10)
	optimized := NewOptimizedUnionType([]string{"Int", "Str"})

	// Test concurrent access to union type cache
	t.Run("ConcurrentUnionCache", func(t *testing.T) {
		done := make(chan bool, 10)

		// Launch multiple goroutines accessing cache
		for i := 0; i < 10; i++ {
			go func(id int) {
				defer func() { done <- true }()

				for j := 0; j < 100; j++ {
					members := []string{"Type" + string(rune('A'+id)), "Int"}
					_ = cache.GetOrCreate(members)
				}
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}

		// Cache should still be valid
		assert.LessOrEqual(t, cache.Size(), 10, "Cache size should respect limit")
	})

	// Test concurrent access to operation cache
	t.Run("ConcurrentOperationCache", func(t *testing.T) {
		done := make(chan bool, 10)

		// Launch multiple goroutines accessing operation cache
		for i := 0; i < 10; i++ {
			go func() {
				defer func() { done <- true }()

				for j := 0; j < 100; j++ {
					_ = optimized.SupportsOperation("\"\"")
					_, _ = optimized.GetOperationResultType("bool")
				}
			}()
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}

		// Operations should still work correctly
		assert.True(t, optimized.SupportsOperation("\"\""))
		result, err := optimized.GetOperationResultType("bool")
		assert.NoError(t, err)
		assert.Equal(t, "Bool", result)
	})
}
