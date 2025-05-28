// ABOUTME: Tests for type resolution result pooling and caching functionality
// ABOUTME: Ensures efficient memory management and cache behavior for type resolution

package typechecker

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"tamarou.com/pvm/internal/ast"
)

func TestTypeResolutionPool_NewTypeResolutionResult(t *testing.T) {
	config := TypeResolutionPoolConfig{
		MaxCacheSize:    1024 * 1024,
		DefaultTTL:      1 * time.Hour,
		MaxCacheEntries: 100,
		EvictionPolicy:  "lru",
	}
	pool := NewTypeResolutionPool(config, TypeResolutionPoolHooks{})

	// Test basic creation
	result := pool.NewTypeResolutionResult("test_key", "Int", 0.9, true)
	if result == nil {
		t.Fatal("Expected non-nil TypeResolutionResult")
	}
	if result.Key != "test_key" {
		t.Errorf("Expected key 'test_key', got %s", result.Key)
	}
	if result.ResolvedType != "Int" {
		t.Errorf("Expected resolved type 'Int', got %s", result.ResolvedType)
	}
	if result.Confidence != 0.9 {
		t.Errorf("Expected confidence 0.9, got %f", result.Confidence)
	}
	if result.Success != true {
		t.Errorf("Expected success true, got %v", result.Success)
	}
	if result.Context == nil {
		t.Error("Expected non-nil Context")
	}
	if result.Dependencies == nil {
		t.Error("Expected non-nil Dependencies slice")
	}
	if result.Timestamp.IsZero() {
		t.Error("Expected non-zero Timestamp")
	}
}

func TestTypeResolutionPool_NewTypeResolutionContext(t *testing.T) {
	config := TypeResolutionPoolConfig{}
	pool := NewTypeResolutionPool(config, TypeResolutionPoolHooks{})

	pos := ast.Position{Line: 10, Column: 5}
	context := pool.NewTypeResolutionContext("test.pl", pos, "assignment", "Int")

	if context == nil {
		t.Fatal("Expected non-nil TypeResolutionContext")
	}
	if context.SourceFile != "test.pl" {
		t.Errorf("Expected source file 'test.pl', got %s", context.SourceFile)
	}
	if context.Position != pos {
		t.Errorf("Expected position %v, got %v", pos, context.Position)
	}
	if context.SurroundingContext != "assignment" {
		t.Errorf("Expected surrounding context 'assignment', got %s", context.SurroundingContext)
	}
	if context.ExpectedType != "Int" {
		t.Errorf("Expected expected type 'Int', got %s", context.ExpectedType)
	}
	if context.ResolutionChain == nil {
		t.Error("Expected non-nil ResolutionChain slice")
	}
	if context.InferenceHints == nil {
		t.Error("Expected non-nil InferenceHints map")
	}
}

func TestTypeResolutionPool_NewTypeResolutionCacheInfo(t *testing.T) {
	config := TypeResolutionPoolConfig{}
	pool := NewTypeResolutionPool(config, TypeResolutionPoolHooks{})

	ttl := 30 * time.Minute
	info := pool.NewTypeResolutionCacheInfo(true, "cache_key", ttl)

	if info == nil {
		t.Fatal("Expected non-nil TypeResolutionCacheInfo")
	}
	if info.CacheHit != true {
		t.Errorf("Expected cache hit true, got %v", info.CacheHit)
	}
	if info.CacheKey != "cache_key" {
		t.Errorf("Expected cache key 'cache_key', got %s", info.CacheKey)
	}
	if info.TTL != ttl {
		t.Errorf("Expected TTL %v, got %v", ttl, info.TTL)
	}
	if info.AccessCount != 0 {
		t.Errorf("Expected access count 0, got %d", info.AccessCount)
	}
	if info.LastAccessed.IsZero() {
		t.Error("Expected non-zero LastAccessed time")
	}
}

func TestTypeResolutionPool_ResolveTypeWithCache(t *testing.T) {
	resolverCalled := 0
	hooksCalled := false

	hooks := TypeResolutionPoolHooks{
		OnCacheHit: func(key string, result *TypeResolutionResult) {
			hooksCalled = true
		},
	}

	config := TypeResolutionPoolConfig{
		MaxCacheSize:    1024 * 1024,
		DefaultTTL:      1 * time.Hour,
		MaxCacheEntries: 100,
		EvictionPolicy:  "lru",
	}
	pool := NewTypeResolutionPool(config, hooks)

	resolver := func() (*TypeResolutionResult, error) {
		resolverCalled++
		return pool.NewTypeResolutionResult("test_resolve", "Str", 0.8, true), nil
	}

	// First call should miss cache and call resolver
	result1, err := pool.ResolveType("test_resolve", resolver)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result1 == nil {
		t.Fatal("Expected non-nil result")
	}
	if resolverCalled != 1 {
		t.Errorf("Expected resolver to be called once, was called %d times", resolverCalled)
	}

	// Second call should hit cache and not call resolver
	result2, err := pool.ResolveType("test_resolve", resolver)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result2 == nil {
		t.Fatal("Expected non-nil result")
	}
	if resolverCalled != 1 {
		t.Errorf("Expected resolver to be called once total, was called %d times", resolverCalled)
	}
	if !hooksCalled {
		t.Error("Expected cache hit hook to be called")
	}

	// Verify cache hit was recorded
	if !result2.CacheInfo.CacheHit {
		t.Error("Expected second result to be marked as cache hit")
	}
}

func TestTypeResolutionPool_CacheEviction(t *testing.T) {
	evictionCalled := false
	hooks := TypeResolutionPoolHooks{
		OnCacheEviction: func(key string, result *TypeResolutionResult) {
			evictionCalled = true
		},
	}

	config := TypeResolutionPoolConfig{
		MaxCacheSize:    1024, // Very small cache
		DefaultTTL:      1 * time.Hour,
		MaxCacheEntries: 2, // Very few entries
		EvictionPolicy:  "lru",
	}
	pool := NewTypeResolutionPool(config, hooks)

	// Fill cache beyond capacity
	resolver1 := func() (*TypeResolutionResult, error) {
		return pool.NewTypeResolutionResult("key1", "Type1", 0.9, true), nil
	}
	resolver2 := func() (*TypeResolutionResult, error) {
		return pool.NewTypeResolutionResult("key2", "Type2", 0.9, true), nil
	}
	resolver3 := func() (*TypeResolutionResult, error) {
		return pool.NewTypeResolutionResult("key3", "Type3", 0.9, true), nil
	}

	_, err := pool.ResolveType("key1", resolver1)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	_, err = pool.ResolveType("key2", resolver2)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Third entry should trigger eviction
	_, err = pool.ResolveType("key3", resolver3)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Give some time for eviction to happen (it might be async)
	time.Sleep(10 * time.Millisecond)

	if !evictionCalled {
		t.Error("Expected cache eviction hook to be called")
	}

	// Check cache stats
	stats := pool.GetCacheStats()
	if stats.Evictions == 0 {
		t.Error("Expected at least one cache eviction")
	}
}

func TestTypeResolutionPool_TTLExpiration(t *testing.T) {
	config := TypeResolutionPoolConfig{
		MaxCacheSize:    1024 * 1024,
		DefaultTTL:      10 * time.Millisecond, // Very short TTL
		MaxCacheEntries: 100,
		EvictionPolicy:  "ttl",
	}
	pool := NewTypeResolutionPool(config, TypeResolutionPoolHooks{})

	resolverCalled := 0
	resolver := func() (*TypeResolutionResult, error) {
		resolverCalled++
		return pool.NewTypeResolutionResult("ttl_test", "ExpiredType", 0.9, true), nil
	}

	// First call
	_, err := pool.ResolveType("ttl_test", resolver)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resolverCalled != 1 {
		t.Errorf("Expected resolver to be called once, was called %d times", resolverCalled)
	}

	// Wait for TTL to expire
	time.Sleep(20 * time.Millisecond)

	// Second call should miss cache due to TTL expiration
	_, err = pool.ResolveType("ttl_test", resolver)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resolverCalled != 2 {
		t.Errorf("Expected resolver to be called twice after TTL expiration, was called %d times", resolverCalled)
	}
}

func TestTypeResolutionPool_InvalidateCache(t *testing.T) {
	config := TypeResolutionPoolConfig{
		MaxCacheSize:    1024 * 1024,
		DefaultTTL:      1 * time.Hour,
		MaxCacheEntries: 100,
		EvictionPolicy:  "lru",
	}
	pool := NewTypeResolutionPool(config, TypeResolutionPoolHooks{})

	// Add some entries to cache
	resolver1 := func() (*TypeResolutionResult, error) {
		return pool.NewTypeResolutionResult("inval1", "Type1", 0.9, true), nil
	}
	resolver2 := func() (*TypeResolutionResult, error) {
		return pool.NewTypeResolutionResult("inval2", "Type2", 0.9, true), nil
	}

	_, err := pool.ResolveType("inval1", resolver1)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	_, err = pool.ResolveType("inval2", resolver2)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check cache has entries
	stats := pool.GetCacheStats()
	if stats.Entries < 2 {
		t.Errorf("Expected at least 2 cache entries, got %d", stats.Entries)
	}

	// Invalidate specific entry
	evicted := pool.InvalidateCache("inval1")
	if evicted != 1 {
		t.Errorf("Expected 1 entry to be evicted, got %d", evicted)
	}

	// Invalidate all entries
	evicted = pool.InvalidateCache("*")
	if evicted < 1 {
		t.Errorf("Expected at least 1 entry to be evicted, got %d", evicted)
	}
}

func TestTypeResolutionPool_ClearCache(t *testing.T) {
	config := TypeResolutionPoolConfig{}
	pool := NewTypeResolutionPool(config, TypeResolutionPoolHooks{})

	// Add entry to cache
	resolver := func() (*TypeResolutionResult, error) {
		return pool.NewTypeResolutionResult("clear_test", "ClearType", 0.9, true), nil
	}

	_, err := pool.ResolveType("clear_test", resolver)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify cache has entries
	stats := pool.GetCacheStats()
	if stats.Entries == 0 {
		t.Error("Expected cache to have entries before clear")
	}

	// Clear cache
	pool.ClearCache()

	// Verify cache is empty
	statsAfter := pool.GetCacheStats()
	if statsAfter.Entries != 0 {
		t.Errorf("Expected 0 cache entries after clear, got %d", statsAfter.Entries)
	}
	if statsAfter.Size != 0 {
		t.Errorf("Expected 0 cache size after clear, got %d", statsAfter.Size)
	}
}

func TestTypeResolutionPool_GetCacheStats(t *testing.T) {
	config := TypeResolutionPoolConfig{
		MaxCacheSize:    1024 * 1024,
		DefaultTTL:      1 * time.Hour,
		MaxCacheEntries: 100,
		EvictionPolicy:  "lru",
	}
	pool := NewTypeResolutionPool(config, TypeResolutionPoolHooks{})

	// Add some entries
	for i := 0; i < 5; i++ {
		key := fmt.Sprintf("stats_test_%d", i)
		resolver := func() (*TypeResolutionResult, error) {
			return pool.NewTypeResolutionResult(key, "StatsType", 0.9, true), nil
		}
		_, err := pool.ResolveType(key, resolver)
		if err != nil {
			t.Fatalf("Expected no error for key %s, got %v", key, err)
		}
	}

	// Get some cache hits
	resolver := func() (*TypeResolutionResult, error) {
		return pool.NewTypeResolutionResult("stats_test_0", "StatsType", 0.9, true), nil
	}
	_, err := pool.ResolveType("stats_test_0", resolver)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	stats := pool.GetCacheStats()

	if stats.Entries < 5 {
		t.Errorf("Expected at least 5 cache entries, got %d", stats.Entries)
	}
	if stats.Size <= 0 {
		t.Errorf("Expected positive cache size, got %d", stats.Size)
	}
	if stats.MaxSize != config.MaxCacheSize {
		t.Errorf("Expected max size %d, got %d", config.MaxCacheSize, stats.MaxSize)
	}
	if stats.Hits < 1 {
		t.Errorf("Expected at least 1 cache hit, got %d", stats.Hits)
	}
	if stats.Misses < 5 {
		t.Errorf("Expected at least 5 cache misses, got %d", stats.Misses)
	}
	if stats.HitRate < 0 || stats.HitRate > 100 {
		t.Errorf("Expected hit rate between 0 and 100, got %f", stats.HitRate)
	}
	if stats.Utilization < 0 || stats.Utilization > 100 {
		t.Errorf("Expected utilization between 0 and 100, got %f", stats.Utilization)
	}
}

func TestTypeResolutionPool_ConcurrentAccess(t *testing.T) {
	config := TypeResolutionPoolConfig{
		MaxCacheSize:         1024 * 1024,
		DefaultTTL:           1 * time.Hour,
		MaxCacheEntries:      1000,
		EvictionPolicy:       "lru",
		ConcurrentResolution: true,
	}
	pool := NewTypeResolutionPool(config, TypeResolutionPoolHooks{})

	const numGoroutines = 10
	const numOperations = 50

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Run concurrent operations
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			for j := 0; j < numOperations; j++ {
				key := fmt.Sprintf("concurrent_%d_%d", id, j)
				resolver := func() (*TypeResolutionResult, error) {
					return pool.NewTypeResolutionResult(key, "ConcurrentType", 0.9, true), nil
				}

				result, err := pool.ResolveType(key, resolver)
				if err != nil {
					t.Errorf("Goroutine %d: Expected no error for key %s, got %v", id, key, err)
					return
				}
				if result == nil {
					t.Errorf("Goroutine %d: Expected non-nil result for key %s", id, key)
					return
				}
				if result.ResolvedType != "ConcurrentType" {
					t.Errorf("Goroutine %d: Expected resolved type 'ConcurrentType', got %s", id, result.ResolvedType)
					return
				}
			}
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Check final statistics
	stats := pool.GetCacheStats()
	expectedOps := int64(numGoroutines * numOperations)
	totalRequests := stats.Hits + stats.Misses
	if totalRequests < expectedOps {
		t.Errorf("Expected at least %d total requests, got %d", expectedOps, totalRequests)
	}
}

func TestTypeResolutionPool_Reset(t *testing.T) {
	config := TypeResolutionPoolConfig{}
	pool := NewTypeResolutionPool(config, TypeResolutionPoolHooks{})

	// Add some entries
	resolver := func() (*TypeResolutionResult, error) {
		return pool.NewTypeResolutionResult("reset_test", "ResetType", 0.9, true), nil
	}

	_, err := pool.ResolveType("reset_test", resolver)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify cache has entries
	statsBefore := pool.GetCacheStats()
	if statsBefore.Entries == 0 {
		t.Error("Expected cache to have entries before reset")
	}

	// Reset the pool
	pool.Reset()

	// Verify cache is cleared
	statsAfter := pool.GetCacheStats()
	if statsAfter.Entries != 0 {
		t.Errorf("Expected 0 cache entries after reset, got %d", statsAfter.Entries)
	}

	// Verify pool still works
	_, err = pool.ResolveType("reset_test_after", resolver)
	if err != nil {
		t.Fatalf("Expected pool to work after reset, got %v", err)
	}
}

func TestTypeResolutionPool_Clear(t *testing.T) {
	config := TypeResolutionPoolConfig{}
	pool := NewTypeResolutionPool(config, TypeResolutionPoolHooks{})

	// Add some entries to populate statistics
	for i := 0; i < 3; i++ {
		key := fmt.Sprintf("clear_test_%d", i)
		resolver := func() (*TypeResolutionResult, error) {
			return pool.NewTypeResolutionResult(key, "ClearType", 0.9, true), nil
		}
		_, err := pool.ResolveType(key, resolver)
		if err != nil {
			t.Fatalf("Expected no error for key %s, got %v", key, err)
		}
	}

	// Get initial stats
	statsBefore := pool.GetDetailedStats()
	if statsBefore.ResolutionCount == 0 {
		t.Error("Expected some resolution operations before clear")
	}

	// Clear the pool
	pool.Clear()

	// Check that statistics are reset
	statsAfter := pool.GetDetailedStats()
	if statsAfter.ResolutionCount != 0 {
		t.Errorf("Expected resolution count to be 0 after clear, got %d", statsAfter.ResolutionCount)
	}
	if statsAfter.PoolHits != 0 {
		t.Errorf("Expected pool hits to be 0 after clear, got %d", statsAfter.PoolHits)
	}
	if statsAfter.PoolMisses != 0 {
		t.Errorf("Expected pool misses to be 0 after clear, got %d", statsAfter.PoolMisses)
	}
}

func TestTypeResolutionPool_WarmCache(t *testing.T) {
	config := TypeResolutionPoolConfig{
		CacheWarming: true,
	}
	pool := NewTypeResolutionPool(config, TypeResolutionPoolHooks{})

	// Cache should be warmed with common types
	stats := pool.GetCacheStats()
	if stats.Entries == 0 {
		t.Error("Expected cache to be warmed with common types")
	}

	// Test that we can resolve common types from cache
	resolver := func() (*TypeResolutionResult, error) {
		return pool.NewTypeResolutionResult("scalar_int", "Int", 1.0, true), nil
	}

	result, err := pool.ResolveType("scalar_int", resolver)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	if result.ResolvedType != "Int" {
		t.Errorf("Expected resolved type 'Int', got %s", result.ResolvedType)
	}

	// Should be a cache hit since we warmed the cache
	if !result.CacheInfo.CacheHit {
		t.Error("Expected cache hit for warmed type")
	}
}

func TestTypeResolutionPool_GlobalInstance(t *testing.T) {
	// Test that global instance is created correctly
	global1 := GlobalTypeResolutionPool()
	if global1 == nil {
		t.Fatal("Expected non-nil global type resolution pool")
	}

	// Test that subsequent calls return the same instance
	global2 := GlobalTypeResolutionPool()
	if global1 != global2 {
		t.Error("Expected same instance from multiple calls to GlobalTypeResolutionPool")
	}
}

func TestTypeResolutionPool_SetGlobalTypeResolutionPoolHooks(t *testing.T) {
	hookCalled := false
	hooks := TypeResolutionPoolHooks{
		OnResolution: func(result *TypeResolutionResult) {
			hookCalled = true
		},
	}

	// Set hooks on global pool
	SetGlobalTypeResolutionPoolHooks(hooks)

	// Use global pool to create a result
	global := GlobalTypeResolutionPool()
	_ = global.NewTypeResolutionResult("global_test", "GlobalType", 0.9, true)

	// Verify hook was called
	if !hookCalled {
		t.Error("Expected hook to be called when using global pool")
	}
}

func TestTypeResolutionPool_GetDetailedStats(t *testing.T) {
	config := TypeResolutionPoolConfig{}
	pool := NewTypeResolutionPool(config, TypeResolutionPoolHooks{})

	// Perform some operations
	resolver := func() (*TypeResolutionResult, error) {
		return pool.NewTypeResolutionResult("detailed_stats", "DetailedType", 0.9, true), nil
	}

	_, err := pool.ResolveType("detailed_stats", resolver)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Get detailed stats
	detailedStats := pool.GetDetailedStats()

	if detailedStats.ResolutionCount == 0 {
		t.Error("Expected at least one resolution operation")
	}
	if detailedStats.PoolHits == 0 {
		t.Error("Expected at least one pool hit")
	}

	// Check that cache stats are included
	if detailedStats.CacheStats.Entries == 0 {
		t.Error("Expected cache stats to be included")
	}

	// Check that pool stats are included
	if detailedStats.PoolStats.Allocations == 0 {
		t.Error("Expected pool stats to be included")
	}
}
