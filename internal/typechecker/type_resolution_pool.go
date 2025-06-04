// ABOUTME: Type resolution result pooling for efficient caching and memory management
// ABOUTME: Provides pooled allocation for type resolution results and caching strategies

package typechecker

import (
	"sync"
	"sync/atomic"
	"time"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/core"
)

// TypeResolutionResult represents the result of a type resolution operation
type TypeResolutionResult struct {
	// Key identifies the resolution request
	Key string

	// ResolvedType is the resolved type
	ResolvedType string

	// Confidence is the confidence level of the resolution
	Confidence float64

	// Success indicates if the resolution was successful
	Success bool

	// ErrorMessage contains error details if resolution failed
	ErrorMessage string

	// Context provides additional resolution context
	Context *TypeResolutionContext

	// CacheInfo contains caching metadata
	CacheInfo TypeResolutionCacheInfo

	// Dependencies lists types that this resolution depends on
	Dependencies []string

	// Timestamp records when this resolution was performed
	Timestamp time.Time
}

// TypeResolutionContext provides context for type resolution
type TypeResolutionContext struct {
	// SourceFile is the file being analyzed
	SourceFile string

	// Position is the location in the source
	Position ast.Position

	// SurroundingContext describes the code context
	SurroundingContext string

	// ExpectedType is the expected type if known
	ExpectedType string

	// Resolution chain for complex types
	ResolutionChain []string

	// Inference hints from static analysis
	InferenceHints map[string]string
}

// TypeResolutionCacheInfo contains metadata for caching
type TypeResolutionCacheInfo struct {
	// CacheHit indicates if this was served from cache
	CacheHit bool

	// CacheKey is the key used for caching
	CacheKey string

	// TTL is the time-to-live for the cache entry
	TTL time.Duration

	// AccessCount tracks how often this result is accessed
	AccessCount int64

	// LastAccessed records the last access time
	LastAccessed time.Time

	// CacheSize is the estimated memory size of the cached result
	CacheSize int64
}

// TypeResolutionPool manages pooled type resolution results
type TypeResolutionPool struct {
	hooks TypeResolutionPoolHooks

	// Core pools
	resultPool    core.Pool[TypeResolutionResult]
	contextPool   core.Pool[TypeResolutionContext]
	cacheInfoPool core.Pool[TypeResolutionCacheInfo]

	// Collection pools
	stringSlicePool      *core.Pool[[]string]
	inferenceHintMapPool *core.Pool[map[string]string]

	// Cache for resolution results
	cache        map[string]*TypeResolutionResult
	cacheMu      sync.RWMutex
	cacheSize    int64
	maxCacheSize int64

	// Statistics
	resolutionCount int64
	cacheHits       int64
	cacheMisses     int64
	cacheEvictions  int64
	poolHits        int64
	poolMisses      int64

	// Configuration
	config TypeResolutionPoolConfig

	mu sync.RWMutex
}

// TypeResolutionPoolConfig contains configuration options
type TypeResolutionPoolConfig struct {
	// MaxCacheSize is the maximum cache size in bytes
	MaxCacheSize int64

	// DefaultTTL is the default time-to-live for cache entries
	DefaultTTL time.Duration

	// MaxCacheEntries is the maximum number of cache entries
	MaxCacheEntries int

	// EvictionPolicy determines how cache entries are evicted
	EvictionPolicy string // "lru", "lfu", "ttl"

	// EnableStatistics controls whether detailed statistics are collected
	EnableStatistics bool

	// CacheWarming enables pre-population of common type resolutions
	CacheWarming bool

	// ConcurrentResolution allows concurrent type resolution
	ConcurrentResolution bool
}

// TypeResolutionPoolHooks provides lifecycle hooks
type TypeResolutionPoolHooks struct {
	OnResolution    func(result *TypeResolutionResult)             // Called when a resolution is performed
	OnCacheHit      func(key string, result *TypeResolutionResult) // Called on cache hit
	OnCacheMiss     func(key string)                               // Called on cache miss
	OnCacheEviction func(key string, result *TypeResolutionResult) // Called when entry is evicted
	OnPoolGrow      func(poolType string, newSize int)             // Called when pool grows
}

// NewTypeResolutionPool creates a new type resolution pool
func NewTypeResolutionPool(config TypeResolutionPoolConfig, hooks TypeResolutionPoolHooks) *TypeResolutionPool {
	if config.MaxCacheSize == 0 {
		config.MaxCacheSize = 64 * 1024 * 1024 // 64MB default
	}
	if config.DefaultTTL == 0 {
		config.DefaultTTL = 1 * time.Hour // 1 hour default
	}
	if config.MaxCacheEntries == 0 {
		config.MaxCacheEntries = 10000 // 10k entries default
	}
	if config.EvictionPolicy == "" {
		config.EvictionPolicy = "lru"
	}

	pool := &TypeResolutionPool{
		hooks:        hooks,
		cache:        make(map[string]*TypeResolutionResult),
		maxCacheSize: config.MaxCacheSize,
		config:       config,
	}

	// Initialize collection pools
	pool.stringSlicePool = &core.Pool[[]string]{}
	pool.inferenceHintMapPool = &core.Pool[map[string]string]{}

	// Register with global pool manager
	core.RegisterGlobalPool("type-resolution", pool)

	// Warm cache if enabled
	if config.CacheWarming {
		pool.warmCache()
	}

	return pool
}

// Stats returns pool allocation statistics
func (trp *TypeResolutionPool) Stats() core.PoolStats {
	return core.PoolStats{
		Allocations: atomic.LoadInt64(&trp.resolutionCount),
		Grows:       atomic.LoadInt64(&trp.poolHits),
		TotalSize:   atomic.LoadInt64(&trp.poolMisses),
		CurrentSize: atomic.LoadInt64(&trp.cacheSize),
		Capacity:    trp.maxCacheSize,
	}
}

// NewTypeResolutionResult creates a pooled type resolution result
func (trp *TypeResolutionPool) NewTypeResolutionResult(key, resolvedType string, confidence float64, success bool) *TypeResolutionResult {
	result := trp.resultPool.New()

	// Reset/initialize the pooled object
	trp.resetTypeResolutionResult(result)

	// Initialize fields
	result.Key = key
	result.ResolvedType = resolvedType
	result.Confidence = confidence
	result.Success = success
	result.Context = trp.NewTypeResolutionContext("", ast.Position{}, "", "")
	result.CacheInfo = *trp.NewTypeResolutionCacheInfo(false, key, trp.config.DefaultTTL)
	result.Dependencies = trp.getStringSlice()
	result.Timestamp = time.Now()

	atomic.AddInt64(&trp.resolutionCount, 1)
	atomic.AddInt64(&trp.poolHits, 1)

	if trp.hooks.OnResolution != nil {
		trp.hooks.OnResolution(result)
	}

	return result
}

// NewTypeResolutionContext creates a pooled type resolution context
func (trp *TypeResolutionPool) NewTypeResolutionContext(sourceFile string, position ast.Position, surroundingContext, expectedType string) *TypeResolutionContext {
	context := trp.contextPool.New()

	// Initialize fields
	context.SourceFile = sourceFile
	context.Position = position
	context.SurroundingContext = surroundingContext
	context.ExpectedType = expectedType
	context.ResolutionChain = trp.getStringSlice()
	context.InferenceHints = trp.getInferenceHintMap()

	atomic.AddInt64(&trp.poolHits, 1)

	return context
}

// NewTypeResolutionCacheInfo creates a pooled cache info
func (trp *TypeResolutionPool) NewTypeResolutionCacheInfo(cacheHit bool, cacheKey string, ttl time.Duration) *TypeResolutionCacheInfo {
	info := trp.cacheInfoPool.New()

	// Initialize fields
	info.CacheHit = cacheHit
	info.CacheKey = cacheKey
	info.TTL = ttl
	info.AccessCount = 0
	info.LastAccessed = time.Now()
	info.CacheSize = 0

	atomic.AddInt64(&trp.poolHits, 1)

	return info
}

// ResolveType resolves a type with caching
func (trp *TypeResolutionPool) ResolveType(key string, resolver func() (*TypeResolutionResult, error)) (*TypeResolutionResult, error) {
	// Check cache first
	if cached := trp.getCachedResult(key); cached != nil {
		atomic.AddInt64(&cached.CacheInfo.AccessCount, 1)
		cached.CacheInfo.LastAccessed = time.Now()
		cached.CacheInfo.CacheHit = true

		atomic.AddInt64(&trp.cacheHits, 1)

		if trp.hooks.OnCacheHit != nil {
			trp.hooks.OnCacheHit(key, cached)
		}

		return cached, nil
	}

	// Cache miss - perform resolution
	atomic.AddInt64(&trp.cacheMisses, 1)

	if trp.hooks.OnCacheMiss != nil {
		trp.hooks.OnCacheMiss(key)
	}

	// Call the resolver function
	result, err := resolver()
	if err != nil {
		return nil, err
	}

	// Cache the result
	trp.cacheResult(key, result)

	return result, nil
}

// getCachedResult retrieves a result from cache
func (trp *TypeResolutionPool) getCachedResult(key string) *TypeResolutionResult {
	trp.cacheMu.RLock()
	defer trp.cacheMu.RUnlock()

	result, exists := trp.cache[key]
	if !exists {
		return nil
	}

	// Check TTL
	if time.Since(result.CacheInfo.LastAccessed) > result.CacheInfo.TTL {
		// Entry expired, remove it
		go trp.evictCacheEntryAsync(key)
		return nil
	}

	return result
}

// cacheResult stores a result in cache
func (trp *TypeResolutionPool) cacheResult(key string, result *TypeResolutionResult) {
	trp.cacheMu.Lock()
	defer trp.cacheMu.Unlock()

	// Check if cache is full
	if len(trp.cache) >= trp.config.MaxCacheEntries {
		trp.evictLRUEntry()
	}

	// Calculate estimated size
	estimatedSize := int64(len(key) + len(result.ResolvedType) + len(result.ErrorMessage) + 100) // approximate

	// Check cache size limit
	if atomic.LoadInt64(&trp.cacheSize)+estimatedSize > trp.maxCacheSize {
		trp.evictLargestEntry()
	}

	// Update cache info
	result.CacheInfo.CacheKey = key
	result.CacheInfo.CacheSize = estimatedSize
	result.CacheInfo.LastAccessed = time.Now()

	// Store in cache
	trp.cache[key] = result
	atomic.AddInt64(&trp.cacheSize, estimatedSize)
}

// evictCacheEntry removes an entry from cache
// Note: Caller must hold cacheMu.Lock()
func (trp *TypeResolutionPool) evictCacheEntry(key string) {
	if result, exists := trp.cache[key]; exists {
		delete(trp.cache, key)
		atomic.AddInt64(&trp.cacheSize, -result.CacheInfo.CacheSize)
		atomic.AddInt64(&trp.cacheEvictions, 1)

		if trp.hooks.OnCacheEviction != nil {
			trp.hooks.OnCacheEviction(key, result)
		}
	}
}

// evictCacheEntryAsync removes an entry from cache asynchronously
// This is safe to call while holding a read lock
func (trp *TypeResolutionPool) evictCacheEntryAsync(key string) {
	trp.cacheMu.Lock()
	defer trp.cacheMu.Unlock()
	trp.evictCacheEntry(key)
}

// evictLRUEntry removes the least recently used entry
func (trp *TypeResolutionPool) evictLRUEntry() {
	var oldestKey string
	var oldestTime time.Time = time.Now()

	for key, result := range trp.cache {
		if result.CacheInfo.LastAccessed.Before(oldestTime) {
			oldestTime = result.CacheInfo.LastAccessed
			oldestKey = key
		}
	}

	if oldestKey != "" {
		trp.evictCacheEntry(oldestKey)
	}
}

// evictLargestEntry removes the largest entry by size
func (trp *TypeResolutionPool) evictLargestEntry() {
	var largestKey string
	var largestSize int64 = 0

	for key, result := range trp.cache {
		if result.CacheInfo.CacheSize > largestSize {
			largestSize = result.CacheInfo.CacheSize
			largestKey = key
		}
	}

	if largestKey != "" {
		trp.evictCacheEntry(largestKey)
	}
}

// InvalidateCache invalidates cache entries matching a pattern
func (trp *TypeResolutionPool) InvalidateCache(pattern string) int {
	trp.cacheMu.Lock()
	defer trp.cacheMu.Unlock()

	evicted := 0
	for key := range trp.cache {
		// Simple pattern matching - in production, might use regex
		if pattern == "*" || key == pattern {
			trp.evictCacheEntry(key)
			evicted++
		}
	}

	return evicted
}

// ClearCache clears the entire cache
func (trp *TypeResolutionPool) ClearCache() {
	trp.cacheMu.Lock()
	defer trp.cacheMu.Unlock()

	trp.cache = make(map[string]*TypeResolutionResult)
	atomic.StoreInt64(&trp.cacheSize, 0)
}

// GetCacheStats returns cache statistics
func (trp *TypeResolutionPool) GetCacheStats() TypeResolutionCacheStats {
	trp.cacheMu.RLock()
	defer trp.cacheMu.RUnlock()

	totalHits := atomic.LoadInt64(&trp.cacheHits)
	totalMisses := atomic.LoadInt64(&trp.cacheMisses)
	totalRequests := totalHits + totalMisses

	hitRate := float64(0)
	if totalRequests > 0 {
		hitRate = float64(totalHits) / float64(totalRequests) * 100
	}

	return TypeResolutionCacheStats{
		Entries:     len(trp.cache),
		Size:        atomic.LoadInt64(&trp.cacheSize),
		MaxSize:     trp.maxCacheSize,
		Hits:        totalHits,
		Misses:      totalMisses,
		Evictions:   atomic.LoadInt64(&trp.cacheEvictions),
		HitRate:     hitRate,
		Utilization: float64(atomic.LoadInt64(&trp.cacheSize)) / float64(trp.maxCacheSize) * 100,
	}
}

// TypeResolutionCacheStats contains cache statistics
type TypeResolutionCacheStats struct {
	Entries     int     // Number of cache entries
	Size        int64   // Current cache size in bytes
	MaxSize     int64   // Maximum cache size in bytes
	Hits        int64   // Number of cache hits
	Misses      int64   // Number of cache misses
	Evictions   int64   // Number of cache evictions
	HitRate     float64 // Cache hit rate as percentage
	Utilization float64 // Cache utilization as percentage
}

// Helper methods for pooled collections

// getStringSlice returns a pooled string slice
func (trp *TypeResolutionPool) getStringSlice() []string {
	s := trp.stringSlicePool.New()
	if *s == nil {
		*s = make([]string, 0, 4)
		atomic.AddInt64(&trp.poolMisses, 1)
	} else {
		*s = (*s)[:0] // Reset length but keep capacity
		atomic.AddInt64(&trp.poolHits, 1)
	}
	return *s
}

// getInferenceHintMap returns a pooled inference hint map
func (trp *TypeResolutionPool) getInferenceHintMap() map[string]string {
	m := trp.inferenceHintMapPool.New()
	if *m == nil {
		*m = make(map[string]string)
		atomic.AddInt64(&trp.poolMisses, 1)
	} else {
		trp.clearStringMap(*m)
		atomic.AddInt64(&trp.poolHits, 1)
	}
	return *m
}

// clearStringMap clears a string map efficiently
func (trp *TypeResolutionPool) clearStringMap(m map[string]string) {
	for k := range m {
		delete(m, k)
	}
}

// Reset methods for proper object cleanup and reuse

// resetTypeResolutionResult resets a type resolution result for reuse
func (trp *TypeResolutionPool) resetTypeResolutionResult(result *TypeResolutionResult) {
	result.Key = ""
	result.ResolvedType = ""
	result.Confidence = 0.0
	result.Success = false
	result.ErrorMessage = ""
	result.Context = nil
	result.Timestamp = time.Time{}

	if result.Dependencies != nil {
		result.Dependencies = result.Dependencies[:0]
	}
}

// Pool management methods

// Reset resets all pools for reuse
func (trp *TypeResolutionPool) Reset() {
	trp.mu.Lock()
	defer trp.mu.Unlock()

	// Reset core pools
	trp.resultPool.Reset()
	trp.contextPool.Reset()
	trp.cacheInfoPool.Reset()

	// Clear cache but keep pool allocations
	trp.ClearCache()
}

// Clear completely empties all pools and resets statistics
func (trp *TypeResolutionPool) Clear() {
	trp.mu.Lock()
	defer trp.mu.Unlock()

	// Clear core pools
	trp.resultPool.Clear()
	trp.contextPool.Clear()
	trp.cacheInfoPool.Clear()

	// Clear cache
	trp.ClearCache()

	// Reset statistics
	atomic.StoreInt64(&trp.resolutionCount, 0)
	atomic.StoreInt64(&trp.cacheHits, 0)
	atomic.StoreInt64(&trp.cacheMisses, 0)
	atomic.StoreInt64(&trp.cacheEvictions, 0)
	atomic.StoreInt64(&trp.poolHits, 0)
	atomic.StoreInt64(&trp.poolMisses, 0)
}

// warmCache pre-populates cache with common type resolutions
func (trp *TypeResolutionPool) warmCache() {
	// Common Perl types
	commonTypes := []struct {
		key        string
		typeName   string
		confidence float64
	}{
		{"scalar_int", "Int", 1.0},
		{"scalar_str", "Str", 1.0},
		{"scalar_num", "Num", 1.0},
		{"scalar_bool", "Bool", 1.0},
		{"scalar_undef", "Undef", 1.0},
		{"ref_array", "ArrayRef", 1.0},
		{"ref_hash", "HashRef", 1.0},
		{"ref_code", "CodeRef", 1.0},
		{"scalar_any", "Any", 1.0},
		{"object_base", "Object", 1.0},
	}

	for _, commonType := range commonTypes {
		result := trp.NewTypeResolutionResult(commonType.key, commonType.typeName, commonType.confidence, true)
		trp.cacheResult(commonType.key, result)
	}
}

// GetDetailedStats returns detailed statistics about the resolution pool
func (trp *TypeResolutionPool) GetDetailedStats() TypeResolutionDetailedStats {
	cacheStats := trp.GetCacheStats()
	poolStats := trp.Stats()

	return TypeResolutionDetailedStats{
		ResolutionCount: atomic.LoadInt64(&trp.resolutionCount),
		PoolHits:        atomic.LoadInt64(&trp.poolHits),
		PoolMisses:      atomic.LoadInt64(&trp.poolMisses),
		CacheStats:      cacheStats,
		PoolStats:       poolStats,
	}
}

// TypeResolutionDetailedStats contains detailed resolution pool statistics
type TypeResolutionDetailedStats struct {
	ResolutionCount int64                    // Total number of resolutions performed
	PoolHits        int64                    // Number of pool hits
	PoolMisses      int64                    // Number of pool misses
	CacheStats      TypeResolutionCacheStats // Cache-specific statistics
	PoolStats       core.PoolStats           // Pool-specific statistics
}

// Global type resolution pool instance
var globalTypeResolutionPool *TypeResolutionPool
var typeResolutionPoolOnce sync.Once

// GlobalTypeResolutionPool returns the global type resolution pool instance
func GlobalTypeResolutionPool() *TypeResolutionPool {
	typeResolutionPoolOnce.Do(func() {
		config := TypeResolutionPoolConfig{
			MaxCacheSize:         64 * 1024 * 1024, // 64MB
			DefaultTTL:           1 * time.Hour,
			MaxCacheEntries:      10000,
			EvictionPolicy:       "lru",
			EnableStatistics:     true,
			CacheWarming:         true,
			ConcurrentResolution: true,
		}

		globalTypeResolutionPool = NewTypeResolutionPool(config, TypeResolutionPoolHooks{
			// Default hooks can be set here
		})
	})
	return globalTypeResolutionPool
}

// SetGlobalTypeResolutionPoolHooks sets hooks for the global type resolution pool
func SetGlobalTypeResolutionPoolHooks(hooks TypeResolutionPoolHooks) {
	pool := GlobalTypeResolutionPool()
	pool.hooks = hooks
}
