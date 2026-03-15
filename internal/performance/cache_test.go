// ABOUTME: Tests for performance caching system
// ABOUTME: Validates cache behavior, TTL, LRU eviction, and performance

package performance

import (
	"testing"
	"time"
)

func TestCache_BasicOperations(t *testing.T) {
	cache := NewCache(3, 0) // No TTL

	// Test Set and Get
	cache.Set("key1", "value1")
	value, found := cache.Get("key1")

	if !found {
		t.Error("Expected to find key1")
	}

	if value != "value1" {
		t.Errorf("Expected value1, got %v", value)
	}

	// Test non-existent key
	_, found = cache.Get("nonexistent")
	if found {
		t.Error("Should not find nonexistent key")
	}
}

func TestCache_LRUEviction(t *testing.T) {
	cache := NewCache(2, 0) // Max size 2, no TTL

	// Fill cache to capacity
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")

	// Access key1 to make it recently used
	cache.Get("key1")

	// Add third item - should evict key2 (least recently used)
	cache.Set("key3", "value3")

	// key1 and key3 should exist, key2 should be evicted
	_, found1 := cache.Get("key1")
	_, found2 := cache.Get("key2")
	_, found3 := cache.Get("key3")

	if !found1 {
		t.Error("key1 should still exist")
	}

	if found2 {
		t.Error("key2 should have been evicted")
	}

	if !found3 {
		t.Error("key3 should exist")
	}
}

func TestCache_TTLExpiration(t *testing.T) {
	cache := NewCache(10, 50*time.Millisecond)

	// Set a value
	cache.Set("key1", "value1")

	// Should be available immediately
	_, found := cache.Get("key1")
	if !found {
		t.Error("key1 should be found immediately")
	}

	// Wait for expiration
	time.Sleep(60 * time.Millisecond)

	// Should be expired now
	_, found = cache.Get("key1")
	if found {
		t.Error("key1 should have expired")
	}
}

func TestCache_Stats(t *testing.T) {
	cache := NewCache(10, 0)

	// Initial stats
	stats := cache.Stats()
	if stats.Items != 0 || stats.HitCount != 0 || stats.MissCount != 0 {
		t.Error("Initial stats should be zero")
	}

	// Add some data and access
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")

	// Hit
	cache.Get("key1")
	// Miss
	cache.Get("nonexistent")

	stats = cache.Stats()

	if stats.Items != 2 {
		t.Errorf("Expected 2 items, got %d", stats.Items)
	}

	if stats.HitCount != 1 {
		t.Errorf("Expected 1 hit, got %d", stats.HitCount)
	}

	if stats.MissCount != 1 {
		t.Errorf("Expected 1 miss, got %d", stats.MissCount)
	}

	if stats.HitRatio != 0.5 {
		t.Errorf("Expected hit ratio 0.5, got %f", stats.HitRatio)
	}
}

func TestCache_Delete(t *testing.T) {
	cache := NewCache(10, 0)

	cache.Set("key1", "value1")

	// Verify it exists
	_, found := cache.Get("key1")
	if !found {
		t.Error("key1 should exist before delete")
	}

	// Delete it
	cache.Delete("key1")

	// Verify it's gone
	_, found = cache.Get("key1")
	if found {
		t.Error("key1 should not exist after delete")
	}
}

func TestCache_Clear(t *testing.T) {
	cache := NewCache(10, 0)

	// Add multiple items
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")

	stats := cache.Stats()
	if stats.Items != 3 {
		t.Errorf("Expected 3 items before clear, got %d", stats.Items)
	}

	// Clear cache
	cache.Clear()

	stats = cache.Stats()
	if stats.Items != 0 || stats.HitCount != 0 || stats.MissCount != 0 {
		t.Error("All stats should be zero after clear")
	}

	// Verify items are gone
	_, found := cache.Get("key1")
	if found {
		t.Error("No items should exist after clear")
	}
}

func TestCache_AccessCount(t *testing.T) {
	cache := NewCache(10, 0)

	cache.Set("key1", "value1")

	// Access multiple times
	for i := 0; i < 5; i++ {
		cache.Get("key1")
	}

	// Check internal access count
	cache.mu.RLock()
	item := cache.items["key1"]
	accessCount := item.AccessCount
	cache.mu.RUnlock()

	// Should be 6 (1 from Set + 5 from Get)
	if accessCount != 6 {
		t.Errorf("Expected access count 6, got %d", accessCount)
	}
}

func TestCache_SizeEstimation(t *testing.T) {
	cache := NewCache(10, 0)

	// Test different value types
	cache.Set("string", "hello world")
	cache.Set("int", 42)
	cache.Set("bytes", []byte("test bytes"))
	cache.Set("bool", true)

	stats := cache.Stats()

	// Should have some size calculated
	if stats.TotalSize <= 0 {
		t.Error("Total size should be greater than 0")
	}

	// String should contribute its length
	stringSize := estimateSize("hello world")
	if stringSize != 11 {
		t.Errorf("Expected string size 11, got %d", stringSize)
	}
}

func TestHashKey(t *testing.T) {
	// Test hash key generation
	key1 := HashKey("test", 123, true)
	key2 := HashKey("test", 123, true)
	key3 := HashKey("test", 123, false)

	// Same inputs should produce same hash
	if key1 != key2 {
		t.Error("Same inputs should produce same hash")
	}

	// Different inputs should produce different hash
	if key1 == key3 {
		t.Error("Different inputs should produce different hash")
	}

	// Hash should be consistent length
	if len(key1) != 16 {
		t.Errorf("Expected hash length 16, got %d", len(key1))
	}
}

func TestCache_ConcurrentAccess(t *testing.T) {
	cache := NewCache(100, 0)

	// Concurrent read/write operations
	done := make(chan bool, 20)

	// Writers
	for i := 0; i < 10; i++ {
		go func(id int) {
			cache.Set(HashKey("key", id), HashKey("value", id))
			done <- true
		}(i)
	}

	// Readers
	for i := 0; i < 10; i++ {
		go func(id int) {
			cache.Get(HashKey("key", id))
			done <- true
		}(i)
	}

	// Wait for all operations
	for i := 0; i < 20; i++ {
		<-done
	}

	// Should complete without data races
	stats := cache.Stats()
	if stats.Items == 0 {
		t.Error("Should have some items after concurrent operations")
	}
}

func TestGlobalCaches(t *testing.T) {
	// Test that global caches are initialized
	if ParserCache == nil {
		t.Error("ParserCache should be initialized")
	}

	if TypeCache == nil {
		t.Error("TypeCache should be initialized")
	}

	if FileCache == nil {
		t.Error("FileCache should be initialized")
	}

	// Test basic operation on global cache
	ParserCache.Set("test", "value")
	value, found := ParserCache.Get("test")

	if !found || value != "value" {
		t.Error("Global cache should work correctly")
	}
}

// Benchmark tests
func BenchmarkCache_Set(b *testing.B) {
	cache := NewCache(1000, 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set(HashKey("key", i), HashKey("value", i))
	}
}

func BenchmarkCache_Get(b *testing.B) {
	cache := NewCache(1000, 0)

	// Populate cache
	for i := 0; i < 100; i++ {
		cache.Set(HashKey("key", i), HashKey("value", i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get(HashKey("key", i%100))
	}
}

func BenchmarkCache_Concurrent(b *testing.B) {
	cache := NewCache(1000, 0)

	// Pre-populate
	for i := 0; i < 100; i++ {
		cache.Set(HashKey("key", i), HashKey("value", i))
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			if i%2 == 0 {
				cache.Get(HashKey("key", i%100))
			} else {
				cache.Set(HashKey("key", i), HashKey("value", i))
			}
			i++
		}
	})
}

func BenchmarkHashKey(b *testing.B) {
	values := []interface{}{"test", 123, true, []byte("data")}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		HashKey(values...)
	}
}
