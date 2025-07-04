// ABOUTME: Performance optimization utilities for CST processing and compilation
// ABOUTME: Provides caching, memory management, and efficient traversal patterns

package compiler

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"sync"
	"time"

	sitter "github.com/tree-sitter/go-tree-sitter"
)

// CompilationCache provides caching for compilation results to avoid redundant processing
type CompilationCache struct {
	cleanCache map[string]*CachedResult
	typedCache map[string]*CachedResult
	mutex      sync.RWMutex
	maxSize    int
	hitCount   int64
	missCount  int64
	evictCount int64
}

// CachedResult stores a compilation result with metadata
type CachedResult struct {
	Code        string
	Timestamp   time.Time
	AccessCount int64
	Size        int
}

// NewCompilationCache creates a new compilation cache with the specified maximum size
func NewCompilationCache(maxSize int) *CompilationCache {
	return &CompilationCache{
		cleanCache: make(map[string]*CachedResult),
		typedCache: make(map[string]*CachedResult),
		maxSize:    maxSize,
	}
}

// generateCacheKey creates a cache key from content and target
func (cc *CompilationCache) generateCacheKey(content []byte, target Target) string {
	hash := sha256.Sum256(content)
	return fmt.Sprintf("%s_%x", target, hash[:16])
}

// Get retrieves a cached compilation result
func (cc *CompilationCache) Get(content []byte, target Target) (string, bool) {
	key := cc.generateCacheKey(content, target)

	cc.mutex.RLock()
	defer cc.mutex.RUnlock()

	cache := cc.cleanCache
	if target == TargetTypedPerl {
		cache = cc.typedCache
	}

	if result, exists := cache[key]; exists {
		cc.hitCount++
		result.AccessCount++
		return result.Code, true
	}

	cc.missCount++
	return "", false
}

// Put stores a compilation result in the cache
func (cc *CompilationCache) Put(content []byte, target Target, code string) {
	key := cc.generateCacheKey(content, target)

	cc.mutex.Lock()
	defer cc.mutex.Unlock()

	cache := cc.cleanCache
	if target == TargetTypedPerl {
		cache = cc.typedCache
	}

	// Check if we need to evict entries
	if len(cache) >= cc.maxSize {
		cc.evictOldest(cache)
	}

	cache[key] = &CachedResult{
		Code:        code,
		Timestamp:   time.Now(),
		AccessCount: 1,
		Size:        len(code),
	}
}

// evictOldest removes the oldest cache entry
func (cc *CompilationCache) evictOldest(cache map[string]*CachedResult) {
	var oldestKey string
	var oldestTime time.Time

	for key, result := range cache {
		if oldestKey == "" || result.Timestamp.Before(oldestTime) {
			oldestKey = key
			oldestTime = result.Timestamp
		}
	}

	if oldestKey != "" {
		delete(cache, oldestKey)
		cc.evictCount++
	}
}

// GetStats returns cache statistics
func (cc *CompilationCache) GetStats() CacheStats {
	cc.mutex.RLock()
	defer cc.mutex.RUnlock()

	return CacheStats{
		HitCount:   cc.hitCount,
		MissCount:  cc.missCount,
		EvictCount: cc.evictCount,
		CleanSize:  len(cc.cleanCache),
		TypedSize:  len(cc.typedCache),
		HitRatio:   float64(cc.hitCount) / float64(cc.hitCount+cc.missCount),
	}
}

// CacheStats provides cache performance metrics
type CacheStats struct {
	HitCount   int64
	MissCount  int64
	EvictCount int64
	CleanSize  int
	TypedSize  int
	HitRatio   float64
}

// OptimizedCSTTransformer provides performance-optimized CST transformation
type OptimizedCSTTransformer struct {
	*CSTTransformer
	nodeCache     map[*sitter.Node]string   // Cache for repeated node transformations
	pathCache     map[string][]*sitter.Node // Cache for frequently used navigation paths
	stringBuilder *strings.Builder
	reuseBuilder  bool
}

// NewOptimizedCSTTransformer creates a performance-optimized CST transformer
func NewOptimizedCSTTransformer(content []byte, options TransformationOptions) *OptimizedCSTTransformer {
	base := NewCSTTransformer(content, options)
	return &OptimizedCSTTransformer{
		CSTTransformer: base,
		nodeCache:      make(map[*sitter.Node]string),
		pathCache:      make(map[string][]*sitter.Node),
		stringBuilder:  &strings.Builder{},
		reuseBuilder:   true,
	}
}

// TransformOptimized performs optimized transformation with caching
func (oct *OptimizedCSTTransformer) TransformOptimized(root *sitter.Node) (string, error) {
	if root == nil {
		return "", fmt.Errorf("cannot transform nil root node")
	}

	// Check node cache first
	if cached, exists := oct.nodeCache[root]; exists {
		return cached, nil
	}

	// Use optimized string building
	if oct.reuseBuilder {
		oct.stringBuilder.Reset()
	} else {
		oct.stringBuilder = &strings.Builder{}
	}

	result, err := oct.transformNodeOptimized(root)
	if err != nil {
		return "", err
	}

	// Cache the result for future use
	oct.nodeCache[root] = result

	return result, nil
}

// transformNodeOptimized performs optimized node transformation
func (oct *OptimizedCSTTransformer) transformNodeOptimized(node *sitter.Node) (string, error) {
	if node == nil {
		return "", nil
	}

	// Check cache first for repeated nodes
	if cached, exists := oct.nodeCache[node]; exists {
		return cached, nil
	}

	// Use the base transformer's rules but with optimized string handling
	for _, rule := range oct.rules {
		if rule.CanTransform(node) {
			result, err := rule.Transform(node, oct.content, oct.CSTTransformer)
			if err != nil {
				return "", err
			}

			// Cache the result
			oct.nodeCache[node] = result
			return result, nil
		}
	}

	// If no rule matches, use optimized text extraction
	result := oct.getNodeTextOptimized(node)
	oct.nodeCache[node] = result
	return result, nil
}

// getNodeTextOptimized provides optimized text extraction with bounds checking
func (oct *OptimizedCSTTransformer) getNodeTextOptimized(node *sitter.Node) string {
	if node == nil || oct.content == nil {
		return ""
	}

	start := node.StartByte()
	end := node.EndByte()

	// Optimized bounds checking
	contentLen := uint(len(oct.content))
	if start >= contentLen || end > contentLen || start > end {
		return ""
	}

	return string(oct.content[start:end])
}

// ClearCache clears the transformation cache to free memory
func (oct *OptimizedCSTTransformer) ClearCache() {
	oct.nodeCache = make(map[*sitter.Node]string)
	oct.pathCache = make(map[string][]*sitter.Node)
}

// GetCacheStats returns transformation cache statistics
func (oct *OptimizedCSTTransformer) GetCacheStats() TransformationCacheStats {
	return TransformationCacheStats{
		NodeCacheSize: len(oct.nodeCache),
		PathCacheSize: len(oct.pathCache),
	}
}

// TransformationCacheStats provides transformation cache metrics
type TransformationCacheStats struct {
	NodeCacheSize int
	PathCacheSize int
}

// MemoryPool manages reusable string builders and byte slices for performance
type MemoryPool struct {
	stringBuilders sync.Pool
	byteSlices     sync.Pool
}

// NewMemoryPool creates a new memory pool for efficient allocation
func NewMemoryPool() *MemoryPool {
	return &MemoryPool{
		stringBuilders: sync.Pool{
			New: func() interface{} {
				return &strings.Builder{}
			},
		},
		byteSlices: sync.Pool{
			New: func() interface{} {
				return make([]byte, 0, 1024) // Start with 1KB capacity
			},
		},
	}
}

// GetStringBuilder retrieves a reusable string builder
func (mp *MemoryPool) GetStringBuilder() *strings.Builder {
	sb := mp.stringBuilders.Get().(*strings.Builder)
	sb.Reset()
	return sb
}

// PutStringBuilder returns a string builder to the pool
func (mp *MemoryPool) PutStringBuilder(sb *strings.Builder) {
	if sb.Cap() < 64*1024 { // Don't pool very large builders
		mp.stringBuilders.Put(sb)
	}
}

// GetByteSlice retrieves a reusable byte slice
func (mp *MemoryPool) GetByteSlice() []byte {
	return mp.byteSlices.Get().([]byte)[:0]
}

// PutByteSlice returns a byte slice to the pool
func (mp *MemoryPool) PutByteSlice(slice []byte) {
	if cap(slice) < 64*1024 { // Don't pool very large slices
		mp.byteSlices.Put(slice)
	}
}

// PerformanceMonitor tracks compilation performance metrics
type PerformanceMonitor struct {
	mutex            sync.RWMutex
	compilationTime  time.Duration
	compilationCount int64
	cacheHits        int64
	cacheMisses      int64
	memoryUsage      int64
}

// RecordCompilation records a compilation operation
func (pm *PerformanceMonitor) RecordCompilation(duration time.Duration, cacheHit bool, memoryUsed int64) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	pm.compilationTime += duration
	pm.compilationCount++
	pm.memoryUsage += memoryUsed

	if cacheHit {
		pm.cacheHits++
	} else {
		pm.cacheMisses++
	}
}

// GetStats returns performance statistics
func (pm *PerformanceMonitor) GetStats() PerformanceStats {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	avgTime := time.Duration(0)
	if pm.compilationCount > 0 {
		avgTime = pm.compilationTime / time.Duration(pm.compilationCount)
	}

	cacheHitRatio := float64(0)
	if pm.cacheHits+pm.cacheMisses > 0 {
		cacheHitRatio = float64(pm.cacheHits) / float64(pm.cacheHits+pm.cacheMisses)
	}

	return PerformanceStats{
		TotalCompilations: pm.compilationCount,
		AverageTime:       avgTime,
		TotalTime:         pm.compilationTime,
		CacheHitRatio:     cacheHitRatio,
		MemoryUsage:       pm.memoryUsage,
	}
}

// PerformanceStats provides compilation performance metrics
type PerformanceStats struct {
	TotalCompilations int64
	AverageTime       time.Duration
	TotalTime         time.Duration
	CacheHitRatio     float64
	MemoryUsage       int64
}

// GlobalPerformanceManager provides global performance tracking
var GlobalPerformanceManager = &PerformanceMonitor{}

// Reset clears all performance statistics
func (pm *PerformanceMonitor) Reset() {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	pm.compilationTime = 0
	pm.compilationCount = 0
	pm.cacheHits = 0
	pm.cacheMisses = 0
	pm.memoryUsage = 0
}
