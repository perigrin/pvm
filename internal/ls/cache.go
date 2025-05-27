// ABOUTME: Document and operation caching system for LSP performance optimization
// ABOUTME: Provides multi-level caching with content hashing and change tracking for faster LSP responses

package ls

import (
	"crypto/sha256"
	"fmt"
	"sync"
	"time"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/binder"
	"tamarou.com/pvm/internal/memory"
	"tamarou.com/pvm/internal/parser"
)

// DocumentCache manages cached document analysis results
type DocumentCache struct {
	// Document-level caches
	asts         map[string]*CachedAST
	symbolTables map[string]*CachedSymbols
	typeResults  map[string]*CachedTypeCheck
	dependencies map[string][]string

	// Operation-level caches
	hoverCache      map[string]*CacheEntry[*Hover]
	completionCache map[string]*CacheEntry[[]CompletionItem]
	definitionCache map[string]*CacheEntry[*Definition]
	referencesCache map[string]*CacheEntry[[]Location]

	// Cache metadata
	mu          sync.RWMutex
	maxAge      time.Duration
	lastCleanup time.Time
	hitCount    int64
	missCount   int64

	// Memory pools for optimized allocations
	pools *LSPPools
}

// CachedAST represents a cached AST with metadata
type CachedAST struct {
	AST          *ast.AST
	ContentHash  string
	LastModified time.Time
	ParseTime    time.Duration
	Dependencies []string
}

// CachedSymbols represents cached symbol table data
type CachedSymbols struct {
	SymbolTable  *binder.SymbolTable
	ContentHash  string
	LastModified time.Time
	BindTime     time.Duration
	ASTHash      string // Hash of the AST this was generated from
}

// CachedTypeCheck represents cached type checking results
type CachedTypeCheck struct {
	Errors       []parser.TypeCheckError
	ContentHash  string
	LastModified time.Time
	CheckTime    time.Duration
	SymbolHash   string // Hash of the symbol table used
}

// CacheEntry represents a cached operation result with TTL
type CacheEntry[T any] struct {
	Value       T
	Timestamp   time.Time
	AccessTime  time.Time
	TTL         time.Duration
	ContentHash string
}

// LSPPools provides memory pools for frequently allocated LSP objects
type LSPPools struct {
	Completions    *memory.SlicePool[CompletionItem]
	Locations      *memory.SlicePool[Location]
	TextEdits      *memory.SlicePool[TextEdit]
	Symbols        *memory.SlicePool[*binder.Symbol]
	StringInterner *memory.StringInterner
}

// CacheStats provides cache performance statistics
type CacheStats struct {
	HitCount    int64
	MissCount   int64
	HitRate     float64
	CacheSize   int
	MemoryUsage int64
	LastCleanup time.Time
}

// NewDocumentCache creates a new document cache with default settings
func NewDocumentCache() *DocumentCache {
	return &DocumentCache{
		asts:            make(map[string]*CachedAST),
		symbolTables:    make(map[string]*CachedSymbols),
		typeResults:     make(map[string]*CachedTypeCheck),
		dependencies:    make(map[string][]string),
		hoverCache:      make(map[string]*CacheEntry[*Hover]),
		completionCache: make(map[string]*CacheEntry[[]CompletionItem]),
		definitionCache: make(map[string]*CacheEntry[*Definition]),
		referencesCache: make(map[string]*CacheEntry[[]Location]),
		maxAge:          5 * time.Minute,
		lastCleanup:     time.Now(),
		pools:           NewLSPPools(),
	}
}

// NewLSPPools creates memory pools for LSP operations
func NewLSPPools() *LSPPools {
	return &LSPPools{
		Completions:    memory.NewSlicePool[CompletionItem]([]int{8, 16, 32, 64, 128}),
		Locations:      memory.NewSlicePool[Location]([]int{4, 8, 16, 32}),
		TextEdits:      memory.NewSlicePool[TextEdit]([]int{4, 8, 16}),
		Symbols:        memory.NewSlicePool[*binder.Symbol]([]int{8, 16, 32, 64}),
		StringInterner: memory.NewStringInterner(),
	}
}

// GetAST retrieves a cached AST or returns nil if not found/expired
func (dc *DocumentCache) GetAST(uri, contentHash string) *ast.AST {
	dc.mu.RLock()
	defer dc.mu.RUnlock()

	if cached, exists := dc.asts[uri]; exists {
		if cached.ContentHash == contentHash && time.Since(cached.LastModified) < dc.maxAge {
			dc.hitCount++
			return cached.AST
		}
	}
	dc.missCount++
	return nil
}

// SetAST stores an AST in the cache
func (dc *DocumentCache) SetAST(uri, contentHash string, astNode *ast.AST, parseTime time.Duration, dependencies []string) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	dc.asts[uri] = &CachedAST{
		AST:          astNode,
		ContentHash:  contentHash,
		LastModified: time.Now(),
		ParseTime:    parseTime,
		Dependencies: dependencies,
	}

	// Update dependencies
	dc.dependencies[uri] = dependencies

	// Trigger cleanup if needed
	dc.maybeCleanup()
}

// GetSymbolTable retrieves a cached symbol table
func (dc *DocumentCache) GetSymbolTable(uri, contentHash, astHash string) *binder.SymbolTable {
	dc.mu.RLock()
	defer dc.mu.RUnlock()

	if cached, exists := dc.symbolTables[uri]; exists {
		if cached.ContentHash == contentHash && cached.ASTHash == astHash &&
			time.Since(cached.LastModified) < dc.maxAge {
			dc.hitCount++
			return cached.SymbolTable
		}
	}
	dc.missCount++
	return nil
}

// SetSymbolTable stores a symbol table in the cache
func (dc *DocumentCache) SetSymbolTable(uri, contentHash, astHash string, symbolTable *binder.SymbolTable, bindTime time.Duration) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	dc.symbolTables[uri] = &CachedSymbols{
		SymbolTable:  symbolTable,
		ContentHash:  contentHash,
		LastModified: time.Now(),
		BindTime:     bindTime,
		ASTHash:      astHash,
	}

	dc.maybeCleanup()
}

// GetTypeCheckResult retrieves cached type checking results
func (dc *DocumentCache) GetTypeCheckResult(uri, contentHash, symbolHash string) []parser.TypeCheckError {
	dc.mu.RLock()
	defer dc.mu.RUnlock()

	if cached, exists := dc.typeResults[uri]; exists {
		if cached.ContentHash == contentHash && cached.SymbolHash == symbolHash &&
			time.Since(cached.LastModified) < dc.maxAge {
			dc.hitCount++
			return cached.Errors
		}
	}
	dc.missCount++
	return nil
}

// SetTypeCheckResult stores type checking results in the cache
func (dc *DocumentCache) SetTypeCheckResult(uri, contentHash, symbolHash string, errors []parser.TypeCheckError, checkTime time.Duration) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	dc.typeResults[uri] = &CachedTypeCheck{
		Errors:       errors,
		ContentHash:  contentHash,
		LastModified: time.Now(),
		CheckTime:    checkTime,
		SymbolHash:   symbolHash,
	}

	dc.maybeCleanup()
}

// GetHover retrieves cached hover information
func (dc *DocumentCache) GetHover(cacheKey string) *Hover {
	dc.mu.RLock()
	defer dc.mu.RUnlock()

	if entry, exists := dc.hoverCache[cacheKey]; exists {
		if time.Since(entry.Timestamp) < entry.TTL {
			entry.AccessTime = time.Now()
			dc.hitCount++
			return entry.Value
		}
	}
	dc.missCount++
	return nil
}

// SetHover stores hover information in the cache
func (dc *DocumentCache) SetHover(cacheKey, contentHash string, hover *Hover) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	dc.hoverCache[cacheKey] = &CacheEntry[*Hover]{
		Value:       hover,
		Timestamp:   time.Now(),
		AccessTime:  time.Now(),
		TTL:         2 * time.Minute,
		ContentHash: contentHash,
	}

	dc.maybeCleanup()
}

// GetCompletions retrieves cached completion items
func (dc *DocumentCache) GetCompletions(cacheKey string) []CompletionItem {
	dc.mu.RLock()
	defer dc.mu.RUnlock()

	if entry, exists := dc.completionCache[cacheKey]; exists {
		if time.Since(entry.Timestamp) < entry.TTL {
			entry.AccessTime = time.Now()
			dc.hitCount++
			return entry.Value
		}
	}
	dc.missCount++
	return nil
}

// SetCompletions stores completion items in the cache
func (dc *DocumentCache) SetCompletions(cacheKey, contentHash string, completions []CompletionItem) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	dc.completionCache[cacheKey] = &CacheEntry[[]CompletionItem]{
		Value:       completions,
		Timestamp:   time.Now(),
		AccessTime:  time.Now(),
		TTL:         1 * time.Minute,
		ContentHash: contentHash,
	}

	dc.maybeCleanup()
}

// InvalidateDocument removes all cached data for a document and its dependents
func (dc *DocumentCache) InvalidateDocument(uri string) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	// Remove direct caches
	delete(dc.asts, uri)
	delete(dc.symbolTables, uri)
	delete(dc.typeResults, uri)

	// Find and invalidate dependents
	dependents := dc.findDependents(uri)
	for _, dependent := range dependents {
		delete(dc.asts, dependent)
		delete(dc.symbolTables, dependent)
		delete(dc.typeResults, dependent)
	}

	// Remove operation caches for this URI
	for key := range dc.hoverCache {
		if dc.keyBelongsToURI(key, uri) {
			delete(dc.hoverCache, key)
		}
	}
	for key := range dc.completionCache {
		if dc.keyBelongsToURI(key, uri) {
			delete(dc.completionCache, key)
		}
	}
	for key := range dc.definitionCache {
		if dc.keyBelongsToURI(key, uri) {
			delete(dc.definitionCache, key)
		}
	}
	for key := range dc.referencesCache {
		if dc.keyBelongsToURI(key, uri) {
			delete(dc.referencesCache, key)
		}
	}

	delete(dc.dependencies, uri)
}

// findDependents finds all documents that depend on the given URI
func (dc *DocumentCache) findDependents(uri string) []string {
	var dependents []string
	for docURI, deps := range dc.dependencies {
		for _, dep := range deps {
			if dep == uri {
				dependents = append(dependents, docURI)
				break
			}
		}
	}
	return dependents
}

// keyBelongsToURI checks if a cache key belongs to a specific URI
func (dc *DocumentCache) keyBelongsToURI(key, uri string) bool {
	// Cache keys are typically formatted as "uri:line:char" or "uri:operation"
	return len(key) > len(uri) && key[:len(uri)] == uri && key[len(uri)] == ':'
}

// maybeCleanup performs cleanup if enough time has passed
func (dc *DocumentCache) maybeCleanup() {
	if time.Since(dc.lastCleanup) > 10*time.Minute {
		go dc.cleanup()
	}
}

// cleanup removes expired entries from all caches
func (dc *DocumentCache) cleanup() {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	now := time.Now()
	dc.lastCleanup = now

	// Clean up document-level caches
	for uri, cached := range dc.asts {
		if now.Sub(cached.LastModified) > dc.maxAge {
			delete(dc.asts, uri)
		}
	}

	for uri, cached := range dc.symbolTables {
		if now.Sub(cached.LastModified) > dc.maxAge {
			delete(dc.symbolTables, uri)
		}
	}

	for uri, cached := range dc.typeResults {
		if now.Sub(cached.LastModified) > dc.maxAge {
			delete(dc.typeResults, uri)
		}
	}

	// Clean up operation caches
	for key, entry := range dc.hoverCache {
		if now.Sub(entry.Timestamp) > entry.TTL {
			delete(dc.hoverCache, key)
		}
	}

	for key, entry := range dc.completionCache {
		if now.Sub(entry.Timestamp) > entry.TTL {
			delete(dc.completionCache, key)
		}
	}

	for key, entry := range dc.definitionCache {
		if now.Sub(entry.Timestamp) > entry.TTL {
			delete(dc.definitionCache, key)
		}
	}

	for key, entry := range dc.referencesCache {
		if now.Sub(entry.Timestamp) > entry.TTL {
			delete(dc.referencesCache, key)
		}
	}
}

// GetStats returns cache performance statistics
func (dc *DocumentCache) GetStats() CacheStats {
	dc.mu.RLock()
	defer dc.mu.RUnlock()

	totalRequests := dc.hitCount + dc.missCount
	hitRate := 0.0
	if totalRequests > 0 {
		hitRate = float64(dc.hitCount) / float64(totalRequests)
	}

	cacheSize := len(dc.asts) + len(dc.symbolTables) + len(dc.typeResults) +
		len(dc.hoverCache) + len(dc.completionCache) +
		len(dc.definitionCache) + len(dc.referencesCache)

	return CacheStats{
		HitCount:    dc.hitCount,
		MissCount:   dc.missCount,
		HitRate:     hitRate,
		CacheSize:   cacheSize,
		MemoryUsage: dc.estimateMemoryUsage(),
		LastCleanup: dc.lastCleanup,
	}
}

// estimateMemoryUsage provides a rough estimate of cache memory usage
func (dc *DocumentCache) estimateMemoryUsage() int64 {
	// This is a simplified estimation
	var usage int64

	// Estimate AST cache usage
	usage += int64(len(dc.asts)) * 1024 // Rough estimate per AST

	// Estimate symbol table usage
	usage += int64(len(dc.symbolTables)) * 512

	// Estimate operation caches
	usage += int64(len(dc.hoverCache)) * 256
	usage += int64(len(dc.completionCache)) * 512

	// Add string interner usage
	usage += dc.pools.StringInterner.MemoryUsage()

	return usage
}

// Clear removes all cached data
func (dc *DocumentCache) Clear() {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	dc.asts = make(map[string]*CachedAST)
	dc.symbolTables = make(map[string]*CachedSymbols)
	dc.typeResults = make(map[string]*CachedTypeCheck)
	dc.dependencies = make(map[string][]string)
	dc.hoverCache = make(map[string]*CacheEntry[*Hover])
	dc.completionCache = make(map[string]*CacheEntry[[]CompletionItem])
	dc.definitionCache = make(map[string]*CacheEntry[*Definition])
	dc.referencesCache = make(map[string]*CacheEntry[[]Location])

	dc.hitCount = 0
	dc.missCount = 0

	// Clear memory pools
	dc.pools.Completions.Clear()
	dc.pools.Locations.Clear()
	dc.pools.TextEdits.Clear()
	dc.pools.Symbols.Clear()
	dc.pools.StringInterner.Clear()
}

// HashContent generates a SHA256 hash of the given content
func (dc *DocumentCache) HashContent(content string) string {
	hash := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", hash)
}

// GetCompletionItems retrieves completion items from pool
func (dc *DocumentCache) GetCompletionItems(minCap int) *[]CompletionItem {
	return dc.pools.Completions.Get(minCap)
}

// PutCompletionItems returns completion items to pool
func (dc *DocumentCache) PutCompletionItems(items *[]CompletionItem) {
	dc.pools.Completions.Put(items)
}

// GetLocations retrieves locations slice from pool
func (dc *DocumentCache) GetLocations(minCap int) *[]Location {
	return dc.pools.Locations.Get(minCap)
}

// PutLocations returns locations slice to pool
func (dc *DocumentCache) PutLocations(locations *[]Location) {
	dc.pools.Locations.Put(locations)
}

// InternString interns a string to reduce memory usage
func (dc *DocumentCache) InternString(s string) string {
	return dc.pools.StringInterner.Intern(s)
}
