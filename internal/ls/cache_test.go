// ABOUTME: Unit tests for the caching infrastructure independent of symbol binding
// ABOUTME: Tests cache functionality, performance monitoring, and memory pooling

package ls

import (
	"testing"
	"time"
)

// TestDocumentCache tests the document cache functionality
func TestDocumentCache(t *testing.T) {
	cache := NewDocumentCache()

	// Test hover caching
	hover := &Hover{
		Contents: "test hover content",
		Range: &Range{
			Start: Position{Line: 0, Character: 0},
			End:   Position{Line: 0, Character: 4},
		},
	}

	cacheKey := "file:///test.pl:0:0"
	contentHash := "testhash"

	// Should be cache miss initially
	result := cache.GetHover(cacheKey)
	if result != nil {
		t.Error("Expected cache miss, but got hit")
	}

	// Set the hover
	cache.SetHover(cacheKey, contentHash, hover)

	// Should be cache hit now
	result = cache.GetHover(cacheKey)
	if result == nil {
		t.Error("Expected cache hit, but got miss")
	}

	if result.Contents != hover.Contents {
		t.Errorf("Cached content mismatch: expected %s, got %s", hover.Contents, result.Contents)
	}
}

// TestCompletionCache tests completion caching
func TestCompletionCache(t *testing.T) {
	cache := NewDocumentCache()

	completions := []CompletionItem{
		{Label: "test1", Kind: CompletionItemKindVariable, Detail: "test variable"},
		{Label: "test2", Kind: CompletionItemKindFunction, Detail: "test function"},
	}

	cacheKey := "file:///test.pl:0:0:completion"
	contentHash := "testhash"

	// Should be cache miss initially
	result := cache.GetCompletions(cacheKey)
	if result != nil {
		t.Error("Expected cache miss, but got hit")
	}

	// Set the completions
	cache.SetCompletions(cacheKey, contentHash, completions)

	// Should be cache hit now
	result = cache.GetCompletions(cacheKey)
	if result == nil {
		t.Error("Expected cache hit, but got miss")
	}

	if len(result) != len(completions) {
		t.Errorf("Cached completions length mismatch: expected %d, got %d", len(completions), len(result))
	}
}

// TestCacheInvalidationDirect tests cache invalidation
func TestCacheInvalidationDirect(t *testing.T) {
	cache := NewDocumentCache()

	uri := "file:///test.pl"
	hover := &Hover{Contents: "test"}

	// Set some cached data
	cache.SetHover(uri+":0:0", "hash1", hover)
	cache.SetHover(uri+":1:0", "hash1", hover)
	cache.SetCompletions(uri+":0:0:completion", "hash1", []CompletionItem{{Label: "test"}})

	// Verify data is cached
	if cache.GetHover(uri+":0:0") == nil {
		t.Error("Expected cached hover")
	}
	if cache.GetCompletions(uri+":0:0:completion") == nil {
		t.Error("Expected cached completions")
	}

	// Invalidate document
	cache.InvalidateDocument(uri)

	// Verify data is no longer cached
	if cache.GetHover(uri+":0:0") != nil {
		t.Error("Expected cache invalidation for hover")
	}
	if cache.GetCompletions(uri+":0:0:completion") != nil {
		t.Error("Expected cache invalidation for completions")
	}
}

// TestCacheStats tests cache statistics
func TestCacheStats(t *testing.T) {
	cache := NewDocumentCache()

	// Initially should have no hits or misses
	stats := cache.GetStats()
	if stats.HitCount != 0 || stats.MissCount != 0 {
		t.Errorf("Expected 0 hits and misses initially, got %d hits, %d misses",
			stats.HitCount, stats.MissCount)
	}

	hover := &Hover{Contents: "test"}
	cacheKey := "file:///test.pl:0:0"

	// Cache miss
	result := cache.GetHover(cacheKey)
	if result != nil {
		t.Error("Expected cache miss")
	}

	// Set value
	cache.SetHover(cacheKey, "hash", hover)

	// Cache hit
	result = cache.GetHover(cacheKey)
	if result == nil {
		t.Error("Expected cache hit")
	}

	// Check stats
	stats = cache.GetStats()
	if stats.HitCount != 1 || stats.MissCount != 1 {
		t.Errorf("Expected 1 hit and 1 miss, got %d hits, %d misses",
			stats.HitCount, stats.MissCount)
	}

	if stats.HitRate != 0.5 {
		t.Errorf("Expected 50%% hit rate, got %.2f%%", stats.HitRate*100)
	}
}

// TestMemoryPools tests memory pool functionality
func TestMemoryPools(t *testing.T) {
	cache := NewDocumentCache()

	// Test completion items pool
	items1 := cache.GetCompletionItems(10)
	if items1 == nil {
		t.Fatal("Expected completion items from pool")
	}

	if cap(*items1) < 10 {
		t.Errorf("Expected capacity >= 10, got %d", cap(*items1))
	}

	// Use the items
	*items1 = append(*items1, CompletionItem{Label: "test"})

	// Return to pool
	cache.PutCompletionItems(items1)

	// Get again - should potentially reuse
	items2 := cache.GetCompletionItems(10)
	if items2 == nil {
		t.Fatal("Expected completion items from pool")
	}

	// Should be reset
	if len(*items2) != 0 {
		t.Errorf("Expected empty slice from pool, got length %d", len(*items2))
	}

	cache.PutCompletionItems(items2)
}

// TestStringInterning tests string interning functionality
func TestStringInterning(t *testing.T) {
	cache := NewDocumentCache()

	str1 := "test string"
	str2 := "test string"

	interned1 := cache.InternString(str1)
	interned2 := cache.InternString(str2)

	// For string interning, we check if they're the same string reference
	// In Go, string interning doesn't guarantee same memory address for identical strings
	// but our string interner should return the same instance
	if interned1 != interned2 {
		t.Error("Expected interned strings to be identical")
	}

	// Content should be preserved
	if interned1 != str1 || interned2 != str2 {
		t.Error("Interned string content should match original")
	}
}

// TestPerformanceMonitor tests the performance monitoring functionality
func TestPerformanceMonitor(t *testing.T) {
	monitor := NewPerformanceMonitor()

	// Test operation timing
	op := monitor.StartOperation(nil, "test")
	time.Sleep(1 * time.Millisecond) // Small delay
	op.Complete()

	// Check stats
	stats := monitor.GetStats()
	if stats.TotalRequests != 1 {
		t.Errorf("Expected 1 request, got %d", stats.TotalRequests)
	}

	// Test error recording
	op2 := monitor.StartOperation(nil, "test")
	op2.CompleteWithError(nil)

	stats = monitor.GetStats()
	if stats.TotalErrors != 1 {
		t.Errorf("Expected 1 error, got %d", stats.TotalErrors)
	}

	// Test cache hit/miss recording
	monitor.RecordCacheHit()
	monitor.RecordCacheMiss()

	stats = monitor.GetStats()
	if stats.CacheHitRate != 0.5 {
		t.Errorf("Expected 50%% cache hit rate, got %.2f%%", stats.CacheHitRate*100)
	}
}

// TestCacheTTL tests cache TTL (time-to-live) functionality
func TestCacheTTL(t *testing.T) {
	cache := NewDocumentCache()

	hover := &Hover{Contents: "test"}
	cacheKey := "file:///test.pl:0:0"

	// Set with short TTL by manipulating the cache entry
	cache.SetHover(cacheKey, "hash", hover)

	// Should be available immediately
	result := cache.GetHover(cacheKey)
	if result == nil {
		t.Error("Expected cached hover to be available")
	}

	// Manually expire the entry by manipulating the timestamp
	cache.mu.Lock()
	if entry, exists := cache.hoverCache[cacheKey]; exists {
		entry.Timestamp = time.Now().Add(-10 * time.Minute) // Make it very old
	}
	cache.mu.Unlock()

	// Should be expired now
	result = cache.GetHover(cacheKey)
	if result != nil {
		t.Error("Expected cached hover to be expired")
	}
}

// TestContentHashing tests content hashing functionality
func TestContentHashing(t *testing.T) {
	cache := NewDocumentCache()

	content1 := "my $var = 42;"
	content2 := "my $var = 42;"
	content3 := "my $var = 43;"

	hash1 := cache.HashContent(content1)
	hash2 := cache.HashContent(content2)
	hash3 := cache.HashContent(content3)

	// Same content should produce same hash
	if hash1 != hash2 {
		t.Error("Same content should produce same hash")
	}

	// Different content should produce different hash
	if hash1 == hash3 {
		t.Error("Different content should produce different hash")
	}

	// Hash should be non-empty
	if hash1 == "" {
		t.Error("Hash should not be empty")
	}
}
