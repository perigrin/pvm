// ABOUTME: Debug tests to understand caching behavior and troubleshoot issues
// ABOUTME: Temporary file for debugging cache functionality

package ls

import (
	"fmt"
	"testing"

	"tamarou.com/pvm/internal/binder"
)

// TestCacheDebug debugs cache behavior step by step
func TestCacheDebug(t *testing.T) {
	ls, err := NewLanguageService()
	if err != nil {
		t.Fatalf("Failed to create language service: %v", err)
	}

	content := "my $var = 42;\nprint $var;\n"
	uri := "file:///debug.pl"
	pos := Position{Line: 1, Character: 6} // Position on $var in the print statement

	// Update document
	err = ls.UpdateDocument(uri, content, 1)
	if err != nil {
		t.Fatalf("UpdateDocument failed: %v", err)
	}

	// Check document and symbol table
	ls.mu.RLock()
	doc, exists := ls.documents[uri]
	ls.mu.RUnlock()

	if !exists {
		t.Fatal("Document not found")
	}

	t.Logf("Document loaded: %v", doc != nil)
	t.Logf("AST loaded: %v", doc.AST != nil)
	t.Logf("Symbol table loaded: %v", doc.SymbolTable != nil)

	if doc.SymbolTable != nil {
		symbols := []*binder.Symbol{}
		ls.collectSymbolsFromScope(doc.SymbolTable.GlobalScope, &symbols)
		t.Logf("Total symbols found: %d", len(symbols))
		for i, symbol := range symbols {
			if i < 5 {
				t.Logf("  Symbol %d: %s (%s)", i, symbol.Name, symbol.Kind)
			}
		}
	}

	// First hover call
	t.Log("=== First hover call ===")
	cacheKey := fmt.Sprintf("%s:%d:%d", uri, pos.Line, pos.Character)
	t.Logf("Cache key: %s", cacheKey)

	// Check if cached before call
	cached := ls.cache.GetHover(cacheKey)
	t.Logf("Cache before first call: %v", cached != nil)

	hover1, err := ls.GetHover(uri, pos)
	if err != nil {
		t.Fatalf("First GetHover failed: %v", err)
	}
	t.Logf("First hover result: %v", hover1 != nil)
	if hover1 != nil {
		t.Logf("Hover content: %s", hover1.Contents)
	}

	// Check cache after first call
	cached = ls.cache.GetHover(cacheKey)
	t.Logf("Cache after first call: %v", cached != nil)

	// Second hover call
	t.Log("=== Second hover call ===")
	hover2, err := ls.GetHover(uri, pos)
	if err != nil {
		t.Fatalf("Second GetHover failed: %v", err)
	}
	t.Logf("Second hover result: %v", hover2 != nil)

	// Check performance stats
	perfStats := ls.GetPerformanceStats()
	cacheStats := ls.GetCacheStats()

	t.Logf("Performance stats - Cache hits: %d, misses: %d, hit rate: %.2f%%",
		perfStats.TotalRequests, perfStats.TotalErrors, perfStats.CacheHitRate*100)
	t.Logf("Cache stats - Hits: %d, misses: %d, hit rate: %.2f%%",
		cacheStats.HitCount, cacheStats.MissCount, cacheStats.HitRate*100)
}
