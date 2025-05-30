// ABOUTME: Performance optimization implementations for CPU and memory efficiency
// ABOUTME: Provides caching, object pooling, and algorithmic optimizations based on profiling data

package performance

import (
	"sync"
	"time"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/binder"
	"tamarou.com/pvm/internal/parser"
)

// ParseCache implements intelligent caching for parsing operations
type ParseCache struct {
	mu      sync.RWMutex
	entries map[string]*CacheEntry
	maxSize int
	hits    int64
	misses  int64
}

// CacheEntry represents a cached parse result
type CacheEntry struct {
	AST         *ast.AST
	SymbolTable *binder.SymbolTable
	Hash        uint64
	LastUsed    time.Time
	UseCount    int64
}

// NewParseCache creates a new parse cache with specified maximum size
func NewParseCache(maxSize int) *ParseCache {
	return &ParseCache{
		entries: make(map[string]*CacheEntry),
		maxSize: maxSize,
	}
}

// Get retrieves cached parse result by content hash
func (pc *ParseCache) Get(key string, contentHash uint64) (*CacheEntry, bool) {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	entry, exists := pc.entries[key]
	if !exists || entry.Hash != contentHash {
		pc.misses++
		return nil, false
	}

	// Update usage statistics
	entry.LastUsed = time.Now()
	entry.UseCount++
	pc.hits++

	return entry, true
}

// Put stores a parse result in the cache
func (pc *ParseCache) Put(key string, ast *ast.AST, symbolTable *binder.SymbolTable, contentHash uint64) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	// Evict old entries if cache is full
	if len(pc.entries) >= pc.maxSize {
		pc.evictLRU()
	}

	pc.entries[key] = &CacheEntry{
		AST:         ast,
		SymbolTable: symbolTable,
		Hash:        contentHash,
		LastUsed:    time.Now(),
		UseCount:    1,
	}
}

// evictLRU removes the least recently used entry
func (pc *ParseCache) evictLRU() {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range pc.entries {
		if oldestKey == "" || entry.LastUsed.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.LastUsed
		}
	}

	if oldestKey != "" {
		delete(pc.entries, oldestKey)
	}
}

// GetStats returns cache performance statistics
func (pc *ParseCache) GetStats() (hits, misses int64, hitRate float64) {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	total := pc.hits + pc.misses
	if total == 0 {
		return pc.hits, pc.misses, 0
	}

	return pc.hits, pc.misses, float64(pc.hits) / float64(total)
}

// OptimizedParser wraps the regular parser with performance optimizations
type OptimizedParser struct {
	cache      *ParseCache
	objectPool *ObjectPool
	// Note: Using parser pool instead of shared instance for thread safety
}

// NewOptimizedParser creates a performance-optimized parser
func NewOptimizedParser() *OptimizedParser {
	return &OptimizedParser{
		cache:      NewParseCache(1000), // Cache up to 1000 parse results
		objectPool: NewObjectPool(),
	}
}

// ParseString parses with caching and optimization
func (op *OptimizedParser) ParseString(content string) (*ast.AST, error) {
	// Calculate content hash for cache key
	contentHash := fastHash([]byte(content))
	cacheKey := content

	// Try cache first
	if entry, found := op.cache.Get(cacheKey, contentHash); found {
		return entry.AST, nil
	}

	// Parse with parser pool for thread safety
	result, err := parser.PooledParserFunc(func(p parser.Parser) (*ast.AST, error) {
		return p.ParseString(content)
	})
	if err != nil {
		return nil, err
	}

	// Cache the result (symbol table will be added separately)
	op.cache.Put(cacheKey, result, nil, contentHash)

	return result, nil
}

// ObjectPool manages reusable objects to reduce allocation overhead
type ObjectPool struct {
	astNodes       sync.Pool
	symbolTables   sync.Pool
	stringBuilders sync.Pool
}

// NewObjectPool creates a new object pool
func NewObjectPool() *ObjectPool {
	return &ObjectPool{
		astNodes: sync.Pool{
			New: func() interface{} {
				return make([]*ast.Node, 0, 100) // Pre-allocate capacity
			},
		},
		symbolTables: sync.Pool{
			New: func() interface{} {
				return make(map[string]*binder.Symbol, 50)
			},
		},
		stringBuilders: sync.Pool{
			New: func() interface{} {
				return make([]byte, 0, 1024) // 1KB initial capacity
			},
		},
	}
}

// GetASTNodes returns a pre-allocated slice for AST nodes
func (op *ObjectPool) GetASTNodes() []*ast.Node {
	nodes := op.astNodes.Get().([]*ast.Node)
	return nodes[:0] // Reset length but keep capacity
}

// PutASTNodes returns AST nodes to the pool
func (op *ObjectPool) PutASTNodes(nodes []*ast.Node) {
	if cap(nodes) > 1000 { // Don't keep very large slices
		return
	}
	op.astNodes.Put(&nodes)
}

// GetSymbolMap returns a pre-allocated map for symbols
func (op *ObjectPool) GetSymbolMap() map[string]*binder.Symbol {
	symbols := op.symbolTables.Get().(map[string]*binder.Symbol)
	// Clear the map but keep capacity
	for k := range symbols {
		delete(symbols, k)
	}
	return symbols
}

// PutSymbolMap returns a symbol map to the pool
func (op *ObjectPool) PutSymbolMap(symbols map[string]*binder.Symbol) {
	if len(symbols) > 500 { // Don't keep very large maps
		return
	}
	op.symbolTables.Put(symbols)
}

// GetStringBuffer returns a pre-allocated byte buffer
func (op *ObjectPool) GetStringBuffer() []byte {
	buffer := op.stringBuilders.Get().([]byte)
	return buffer[:0] // Reset length but keep capacity
}

// PutStringBuffer returns a buffer to the pool
func (op *ObjectPool) PutStringBuffer(buffer []byte) {
	if cap(buffer) > 10240 { // Don't keep buffers larger than 10KB
		return
	}
	op.stringBuilders.Put(&buffer)
}

// fastHash implements a fast hash function for content caching
func fastHash(data []byte) uint64 {
	// FNV-1a hash function - fast and good distribution
	const prime = 1099511628211
	var hash uint64 = 14695981039346656037

	for _, b := range data {
		hash ^= uint64(b)
		hash *= prime
	}

	return hash
}

// PerformanceMonitor tracks optimization effectiveness
type PerformanceMonitor struct {
	mu              sync.RWMutex
	parseOperations int64
	cacheHits       int64
	totalParseTime  time.Duration
	avgParseTime    time.Duration
	lastUpdate      time.Time
}

// NewPerformanceMonitor creates a new performance monitor
func NewPerformanceMonitor() *PerformanceMonitor {
	return &PerformanceMonitor{
		lastUpdate: time.Now(),
	}
}

// RecordParse records a parse operation for monitoring
func (pm *PerformanceMonitor) RecordParse(duration time.Duration, cacheHit bool) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.parseOperations++
	pm.totalParseTime += duration

	if cacheHit {
		pm.cacheHits++
	}

	// Update running average
	pm.avgParseTime = pm.totalParseTime / time.Duration(pm.parseOperations)
	pm.lastUpdate = time.Now()
}

// GetStats returns current performance statistics
func (pm *PerformanceMonitor) GetStats() (operations int64, cacheHitRate float64, avgTime time.Duration) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var hitRate float64
	if pm.parseOperations > 0 {
		hitRate = float64(pm.cacheHits) / float64(pm.parseOperations)
	}

	return pm.parseOperations, hitRate, pm.avgParseTime
}
