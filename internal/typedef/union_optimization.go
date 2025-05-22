// ABOUTME: Performance optimizations for union types
// ABOUTME: Additional caching and memory optimizations for union type operations

package typedef

import (
	"sync"
)

// UnionTypeCache provides optimized caching for union type instances
type UnionTypeCache struct {
	// cache stores union type instances by their string representation
	cache map[string]*UnionType

	// mutex protects concurrent access to the cache
	mu sync.RWMutex

	// maxSize limits cache size to prevent memory bloat
	maxSize int

	// accessCount tracks access frequency for LRU eviction
	accessCount map[string]int
}

// NewUnionTypeCache creates a new union type cache with specified max size
func NewUnionTypeCache(maxSize int) *UnionTypeCache {
	return &UnionTypeCache{
		cache:       make(map[string]*UnionType),
		maxSize:     maxSize,
		accessCount: make(map[string]int),
	}
}

// GetOrCreate returns an existing union type or creates a new one
func (utc *UnionTypeCache) GetOrCreate(members []string) *UnionType {
	// Create a canonical key for the union type
	key := createUnionKey(members)

	// Try to get from cache first (read lock)
	utc.mu.RLock()
	if cached, exists := utc.cache[key]; exists {
		utc.mu.RUnlock()

		// Update access count with write lock
		utc.mu.Lock()
		utc.accessCount[key]++
		utc.mu.Unlock()

		return cached
	}
	utc.mu.RUnlock()

	// Not in cache, create new one (write lock)
	utc.mu.Lock()
	defer utc.mu.Unlock()

	// Double-check after acquiring write lock
	if cached, exists := utc.cache[key]; exists {
		utc.accessCount[key]++
		return cached
	}

	// Create new union type
	unionType := NewUnionType(members)

	// Add to cache with eviction if needed
	if len(utc.cache) >= utc.maxSize {
		utc.evictLRU()
	}

	utc.cache[key] = unionType
	utc.accessCount[key] = 1

	return unionType
}

// evictLRU removes the least recently used item from cache
func (utc *UnionTypeCache) evictLRU() {
	minCount := int(^uint(0) >> 1) // Max int
	var evictKey string

	for key, count := range utc.accessCount {
		if count < minCount {
			minCount = count
			evictKey = key
		}
	}

	if evictKey != "" {
		delete(utc.cache, evictKey)
		delete(utc.accessCount, evictKey)
	}
}

// Clear removes all cached union types
func (utc *UnionTypeCache) Clear() {
	utc.mu.Lock()
	defer utc.mu.Unlock()

	utc.cache = make(map[string]*UnionType)
	utc.accessCount = make(map[string]int)
}

// Size returns the current cache size
func (utc *UnionTypeCache) Size() int {
	utc.mu.RLock()
	defer utc.mu.RUnlock()
	return len(utc.cache)
}

// createUnionKey creates a canonical key for union type caching
func createUnionKey(members []string) string {
	// Remove duplicates and sort for canonical representation
	seen := make(map[string]bool)
	unique := make([]string, 0, len(members))

	for _, member := range members {
		if !seen[member] {
			unique = append(unique, member)
			seen[member] = true
		}
	}

	// Sort to ensure consistent key regardless of input order
	for i := 0; i < len(unique)-1; i++ {
		for j := i + 1; j < len(unique); j++ {
			if unique[i] > unique[j] {
				unique[i], unique[j] = unique[j], unique[i]
			}
		}
	}

	// Create pipe-separated key
	result := ""
	for i, member := range unique {
		if i > 0 {
			result += "|"
		}
		result += member
	}

	return result
}

// OptimizedUnionType provides memory and performance optimizations for union types
type OptimizedUnionType struct {
	*UnionType

	// operationCache caches frequently used operation results
	operationCache map[string]interface{}
	mu             sync.RWMutex
}

// NewOptimizedUnionType creates an optimized union type wrapper
func NewOptimizedUnionType(members []string) *OptimizedUnionType {
	return &OptimizedUnionType{
		UnionType:      NewUnionType(members),
		operationCache: make(map[string]interface{}),
	}
}

// SupportsOperation with caching for frequently called operations
func (out *OptimizedUnionType) SupportsOperation(operation string) bool {
	cacheKey := "supports:" + operation

	// Check cache first
	out.mu.RLock()
	if cached, exists := out.operationCache[cacheKey]; exists {
		out.mu.RUnlock()
		return cached.(bool)
	}
	out.mu.RUnlock()

	// Compute and cache result
	result := out.UnionType.SupportsOperation(operation)

	out.mu.Lock()
	out.operationCache[cacheKey] = result
	out.mu.Unlock()

	return result
}

// OperationResult represents a cached operation result
type OperationResult struct {
	Result string
	Error  error
}

// GetOperationResultType with caching for frequently called operations
func (out *OptimizedUnionType) GetOperationResultType(operation string) (string, error) {
	cacheKey := "result:" + operation

	// Check cache first
	out.mu.RLock()
	if cached, exists := out.operationCache[cacheKey]; exists {
		out.mu.RUnlock()
		if result, ok := cached.(*OperationResult); ok {
			return result.Result, result.Error
		}
		// Fallback for old cache format
		return cached.(string), nil
	}
	out.mu.RUnlock()

	// Compute and cache result
	result, err := out.UnionType.GetOperationResultType(operation)

	out.mu.Lock()
	out.operationCache[cacheKey] = &OperationResult{
		Result: result,
		Error:  err,
	}
	out.mu.Unlock()

	return result, err
}

// ClearOperationCache clears the operation result cache
func (out *OptimizedUnionType) ClearOperationCache() {
	out.mu.Lock()
	defer out.mu.Unlock()
	out.operationCache = make(map[string]interface{})
}

// MemoryFootprint returns an estimate of memory usage
func (out *OptimizedUnionType) MemoryFootprint() int {
	out.mu.RLock()
	defer out.mu.RUnlock()

	// Rough estimation of memory usage
	base := len(out.Members) * 50          // Estimated per-member overhead
	cache := len(out.operationCache) * 100 // Estimated per-cache-entry overhead

	return base + cache
}
