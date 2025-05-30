// ABOUTME: Performance benchmark tests for LSP operations and caching effectiveness
// ABOUTME: Validates performance targets and regression detection for language service operations

package ls

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// BenchmarkUpdateDocument tests document update performance with various file sizes
func BenchmarkUpdateDocument(b *testing.B) {
	ls, err := NewLanguageService()
	if err != nil {
		b.Fatalf("Failed to create language service: %v", err)
	}

	testCases := []struct {
		name string
		size int
	}{
		{"Small_100lines", 100},
		{"Medium_500lines", 500},
		{"Large_1000lines", 1000},
		{"XLarge_2000lines", 2000},
	}

	for _, tc := range testCases {
		content := generatePerlCode(tc.size)
		uri := fmt.Sprintf("file:///test_%s.pl", tc.name)

		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				err := ls.UpdateDocument(uri, content, i)
				if err != nil {
					b.Fatalf("UpdateDocument failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkUpdateDocumentCached tests caching effectiveness
func BenchmarkUpdateDocumentCached(b *testing.B) {
	ls, err := NewLanguageService()
	if err != nil {
		b.Fatalf("Failed to create language service: %v", err)
	}

	content := generatePerlCode(500)
	uri := "file:///test_cached.pl"

	// First update to populate cache
	err = ls.UpdateDocument(uri, content, 1)
	if err != nil {
		b.Fatalf("Initial UpdateDocument failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Same content should hit cache
		err := ls.UpdateDocument(uri, content, 1)
		if err != nil {
			b.Fatalf("Cached UpdateDocument failed: %v", err)
		}
	}
}

// BenchmarkHoverOperation tests hover performance
func BenchmarkHoverOperation(b *testing.B) {
	ls, err := NewLanguageService()
	if err != nil {
		b.Fatalf("Failed to create language service: %v", err)
	}

	content := generatePerlCode(500)
	uri := "file:///test_hover.pl"

	err = ls.UpdateDocument(uri, content, 1)
	if err != nil {
		b.Fatalf("UpdateDocument failed: %v", err)
	}

	pos := Position{Line: 10, Character: 5}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ls.GetHover(uri, pos)
		if err != nil {
			b.Fatalf("GetHover failed: %v", err)
		}
	}
}

// BenchmarkCompletionOperation tests completion performance
func BenchmarkCompletionOperation(b *testing.B) {
	ls, err := NewLanguageService()
	if err != nil {
		b.Fatalf("Failed to create language service: %v", err)
	}

	content := generatePerlCode(500)
	uri := "file:///test_completion.pl"

	err = ls.UpdateDocument(uri, content, 1)
	if err != nil {
		b.Fatalf("UpdateDocument failed: %v", err)
	}

	pos := Position{Line: 10, Character: 5}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ls.GetCompletions(uri, pos)
		if err != nil {
			b.Fatalf("GetCompletions failed: %v", err)
		}
	}
}

// BenchmarkDefinitionOperation tests definition lookup performance
func BenchmarkDefinitionOperation(b *testing.B) {
	ls, err := NewLanguageService()
	if err != nil {
		b.Fatalf("Failed to create language service: %v", err)
	}

	content := generatePerlCode(500)
	uri := "file:///test_definition.pl"

	err = ls.UpdateDocument(uri, content, 1)
	if err != nil {
		b.Fatalf("UpdateDocument failed: %v", err)
	}

	pos := Position{Line: 10, Character: 5}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ls.GetDefinition(uri, pos)
		if err != nil {
			b.Fatalf("GetDefinition failed: %v", err)
		}
	}
}

// BenchmarkMemoryPooling tests memory pool effectiveness
func BenchmarkMemoryPooling(b *testing.B) {
	cache := NewDocumentCache()

	b.Run("WithPooling", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			items := cache.GetCompletionItems(32)
			*items = append(*items, CompletionItem{
				Label: "test",
				Kind:  CompletionItemKindVariable,
			})
			cache.PutCompletionItems(items)
		}
	})

	b.Run("WithoutPooling", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			items := make([]CompletionItem, 0, 32)
			items = append(items, CompletionItem{
				Label: "test",
				Kind:  CompletionItemKindVariable,
			})
			_ = items // Simulate usage
		}
	})
}

// BenchmarkCachePerformance tests cache hit/miss performance
func BenchmarkCachePerformance(b *testing.B) {
	cache := NewDocumentCache()

	// Populate cache
	hover := &Hover{Contents: "test hover"}
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("test:%d:5", i)
		cache.SetHover(key, "hash", hover)
	}

	b.Run("CacheHit", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("test:%d:5", i%1000)
			result := cache.GetHover(key)
			if result == nil {
				b.Fatal("Expected cache hit")
			}
		}
	})

	b.Run("CacheMiss", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("miss:%d:5", i)
			result := cache.GetHover(key)
			if result != nil {
				b.Fatal("Expected cache miss")
			}
		}
	})
}

// TestPerformanceTargets validates that operations meet performance targets
func TestPerformanceTargets(t *testing.T) {
	ls, err := NewLanguageService()
	if err != nil {
		t.Fatalf("Failed to create language service: %v", err)
	}

	content := generatePerlCode(500)
	uri := "file:///test_targets.pl"

	err = ls.UpdateDocument(uri, content, 1)
	if err != nil {
		t.Fatalf("UpdateDocument failed: %v", err)
	}

	pos := Position{Line: 10, Character: 5}

	// Test hover performance target
	start := time.Now()
	_, err = ls.GetHover(uri, pos)
	if err != nil {
		t.Fatalf("GetHover failed: %v", err)
	}
	hoverDuration := time.Since(start)

	if hoverDuration > time.Duration(DefaultTargets.HoverMaxMs)*time.Millisecond {
		t.Errorf("Hover operation took %v, exceeds target of %dms",
			hoverDuration, DefaultTargets.HoverMaxMs)
	}

	// Test completion performance target
	start = time.Now()
	_, err = ls.GetCompletions(uri, pos)
	if err != nil {
		t.Fatalf("GetCompletions failed: %v", err)
	}
	completionDuration := time.Since(start)

	if completionDuration > time.Duration(DefaultTargets.CompletionMaxMs)*time.Millisecond {
		t.Errorf("Completion operation took %v, exceeds target of %dms",
			completionDuration, DefaultTargets.CompletionMaxMs)
	}

	// Check overall performance stats
	stats := ls.GetPerformanceStats()

	if stats.ErrorRate > DefaultTargets.MaxErrorRate {
		t.Errorf("Error rate %.2f exceeds target %.2f",
			stats.ErrorRate, DefaultTargets.MaxErrorRate)
	}

	// After some operations, we should have cache hits
	if stats.CacheHitRate > 0 && stats.CacheHitRate < DefaultTargets.MinCacheHitRate {
		t.Errorf("Cache hit rate %.2f below target %.2f",
			stats.CacheHitRate, DefaultTargets.MinCacheHitRate)
	}
}

// TestCacheEffectiveness validates that caching provides performance benefits
func TestCacheEffectiveness(t *testing.T) {
	ls, err := NewLanguageService()
	if err != nil {
		t.Fatalf("Failed to create language service: %v", err)
	}

	content := generatePerlCode(500)
	uri := "file:///test_cache_effectiveness.pl"
	pos := Position{Line: 10, Character: 5}

	// First call (cache miss)
	start := time.Now()
	err = ls.UpdateDocument(uri, content, 1)
	if err != nil {
		t.Fatalf("UpdateDocument failed: %v", err)
	}
	_, err = ls.GetHover(uri, pos)
	if err != nil {
		t.Fatalf("GetHover failed: %v", err)
	}
	firstCallDuration := time.Since(start)

	// Second call (should hit cache)
	start = time.Now()
	_, err = ls.GetHover(uri, pos)
	if err != nil {
		t.Fatalf("Second GetHover failed: %v", err)
	}
	secondCallDuration := time.Since(start)

	// Second call should be significantly faster
	if secondCallDuration >= firstCallDuration {
		t.Errorf("Second call (%v) should be faster than first call (%v)",
			secondCallDuration, firstCallDuration)
	}

	// Verify cache statistics from performance monitor
	perfStats := ls.GetPerformanceStats()
	if perfStats.CacheHitRate == 0 {
		t.Error("Expected cache hits, but got none")
	}

	t.Logf("Cache effectiveness: %.2f%% hit rate",
		perfStats.CacheHitRate*100)
}

// TestMemoryUsage validates memory usage patterns
func TestMemoryUsage(t *testing.T) {
	ls, err := NewLanguageService()
	if err != nil {
		t.Fatalf("Failed to create language service: %v", err)
	}

	// Process multiple documents to build up cache
	for i := 0; i < 10; i++ {
		content := generatePerlCode(100)
		uri := fmt.Sprintf("file:///test_memory_%d.pl", i)

		err = ls.UpdateDocument(uri, content, 1)
		if err != nil {
			t.Fatalf("UpdateDocument failed: %v", err)
		}

		pos := Position{Line: 6, Character: 4} // Should be on $var0
		_, err = ls.GetHover(uri, pos)
		if err != nil {
			t.Fatalf("GetHover failed: %v", err)
		}
	}

	cacheStats := ls.GetCacheStats()
	perfStats := ls.GetPerformanceStats()

	t.Logf("Memory usage: %.2f MB current, %.2f MB peak",
		perfStats.CurrentMemoryMB, perfStats.PeakMemoryMB)
	t.Logf("Cache size: %d entries, estimated %d bytes",
		cacheStats.CacheSize, cacheStats.MemoryUsage)

	// Basic sanity check - memory usage should be reasonable
	if perfStats.CurrentMemoryMB > 100 { // 100MB seems excessive for test
		t.Errorf("Memory usage seems high: %.2f MB", perfStats.CurrentMemoryMB)
	}
}

// TestCacheInvalidation tests that cache invalidation works correctly
func TestCacheInvalidation(t *testing.T) {
	ls, err := NewLanguageService()
	if err != nil {
		t.Fatalf("Failed to create language service: %v", err)
	}

	uri := "file:///test_invalidation.pl"
	content1 := generatePerlCode(100)
	content2 := generatePerlCode(100) + "\n# Additional line"
	pos := Position{Line: 6, Character: 4} // Should be on $var0

	// First document version
	err = ls.UpdateDocument(uri, content1, 1)
	if err != nil {
		t.Fatalf("UpdateDocument failed: %v", err)
	}

	hover1, err := ls.GetHover(uri, pos)
	if err != nil {
		t.Fatalf("GetHover failed: %v", err)
	}

	// Update document (should invalidate cache)
	err = ls.UpdateDocument(uri, content2, 2)
	if err != nil {
		t.Fatalf("UpdateDocument failed: %v", err)
	}

	hover2, err := ls.GetHover(uri, pos)
	if err != nil {
		t.Fatalf("GetHover failed: %v", err)
	}

	// Results might be the same, but cache should have been invalidated
	cacheStats := ls.GetCacheStats()
	if cacheStats.MissCount == 0 {
		t.Error("Expected cache misses due to invalidation")
	}

	// Both calls should succeed
	if hover1 == nil || hover2 == nil {
		t.Error("Hover results should not be nil")
	}
}

// generatePerlCode creates a synthetic Perl file with the specified number of lines
func generatePerlCode(lines int) string {
	var builder strings.Builder

	builder.WriteString("#!/usr/bin/perl\n")
	builder.WriteString("use strict;\n")
	builder.WriteString("use warnings;\n\n")

	builder.WriteString("package TestCode;\n\n")

	// Generate some variables
	for i := 0; i < lines/10; i++ {
		builder.WriteString(fmt.Sprintf("my $var%d = %d;\n", i, i))
	}
	builder.WriteString("\n")

	// Generate some subroutines
	for i := 0; i < lines/20; i++ {
		builder.WriteString(fmt.Sprintf("sub function%d {\n", i))
		builder.WriteString("    my ($param) = @_;\n")
		builder.WriteString(fmt.Sprintf("    return $param * %d;\n", i+1))
		builder.WriteString("}\n\n")
	}

	// Fill remaining lines with comments and simple statements
	remaining := lines - (lines / 10) - (lines/20)*4 - 6
	for i := 0; i < remaining; i++ {
		if i%3 == 0 {
			builder.WriteString(fmt.Sprintf("# Comment line %d\n", i))
		} else {
			builder.WriteString(fmt.Sprintf("print \"Line %d\\n\";\n", i))
		}
	}

	return builder.String()
}
