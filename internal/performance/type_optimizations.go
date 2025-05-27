// ABOUTME: Type system performance optimizations for union types and type checking
// ABOUTME: Implements caching, lazy evaluation, and optimized algorithms for type operations

package performance

import (
	"sync"
	"time"
)

// TypeCache implements intelligent caching for type operations
type TypeCache struct {
	mu                 sync.RWMutex
	unionCache         map[string]*CachedUnionType
	intersectionCache  map[string]*CachedIntersectionType
	compatibilityCache map[string]bool
	maxCacheSize       int
	hits               int64
	misses             int64
}

// CachedUnionType represents a cached union type computation
type CachedUnionType struct {
	Components      []string
	ResultType      string
	ComputationTime time.Duration
	LastAccess      time.Time
	AccessCount     int64
}

// CachedIntersectionType represents a cached intersection type computation
type CachedIntersectionType struct {
	Components      []string
	ResultType      string
	ComputationTime time.Duration
	LastAccess      time.Time
	AccessCount     int64
}

// NewTypeCache creates a new type operation cache
func NewTypeCache(maxSize int) *TypeCache {
	return &TypeCache{
		unionCache:         make(map[string]*CachedUnionType),
		intersectionCache:  make(map[string]*CachedIntersectionType),
		compatibilityCache: make(map[string]bool),
		maxCacheSize:       maxSize,
	}
}

// GetUnionType retrieves or computes a union type
func (tc *TypeCache) GetUnionType(components []string, computeFunc func() (string, time.Duration)) (string, bool) {
	key := tc.makeUnionKey(components)

	tc.mu.RLock()
	cached, exists := tc.unionCache[key]
	tc.mu.RUnlock()

	if exists {
		tc.mu.Lock()
		cached.LastAccess = time.Now()
		cached.AccessCount++
		tc.mu.Unlock()
		tc.hits++
		return cached.ResultType, true
	}

	// Compute the union type
	tc.misses++
	result, computeTime := computeFunc()

	tc.mu.Lock()
	defer tc.mu.Unlock()

	// Evict entries if cache is full
	if len(tc.unionCache) >= tc.maxCacheSize {
		tc.evictLRUUnion()
	}

	tc.unionCache[key] = &CachedUnionType{
		Components:      components,
		ResultType:      result,
		ComputationTime: computeTime,
		LastAccess:      time.Now(),
		AccessCount:     1,
	}

	return result, false
}

// GetIntersectionType retrieves or computes an intersection type
func (tc *TypeCache) GetIntersectionType(components []string, computeFunc func() (string, time.Duration)) (string, bool) {
	key := tc.makeIntersectionKey(components)

	tc.mu.RLock()
	cached, exists := tc.intersectionCache[key]
	tc.mu.RUnlock()

	if exists {
		tc.mu.Lock()
		cached.LastAccess = time.Now()
		cached.AccessCount++
		tc.mu.Unlock()
		tc.hits++
		return cached.ResultType, true
	}

	// Compute the intersection type
	tc.misses++
	result, computeTime := computeFunc()

	tc.mu.Lock()
	defer tc.mu.Unlock()

	// Evict entries if cache is full
	if len(tc.intersectionCache) >= tc.maxCacheSize {
		tc.evictLRUIntersection()
	}

	tc.intersectionCache[key] = &CachedIntersectionType{
		Components:      components,
		ResultType:      result,
		ComputationTime: computeTime,
		LastAccess:      time.Now(),
		AccessCount:     1,
	}

	return result, false
}

// CheckCompatibility checks type compatibility with caching
func (tc *TypeCache) CheckCompatibility(typeA, typeB string, computeFunc func() bool) bool {
	key := typeA + "|" + typeB

	tc.mu.RLock()
	result, exists := tc.compatibilityCache[key]
	tc.mu.RUnlock()

	if exists {
		tc.hits++
		return result
	}

	// Compute compatibility
	tc.misses++
	result = computeFunc()

	tc.mu.Lock()
	defer tc.mu.Unlock()

	// Limit compatibility cache size
	if len(tc.compatibilityCache) >= tc.maxCacheSize*2 {
		// Simple eviction: clear half the cache
		count := 0
		for k := range tc.compatibilityCache {
			delete(tc.compatibilityCache, k)
			count++
			if count >= tc.maxCacheSize {
				break
			}
		}
	}

	tc.compatibilityCache[key] = result
	return result
}

// makeUnionKey creates a consistent key for union type caching
func (tc *TypeCache) makeUnionKey(components []string) string {
	// Sort components to ensure consistent keys
	sorted := make([]string, len(components))
	copy(sorted, components)

	// Simple insertion sort for small arrays
	for i := 1; i < len(sorted); i++ {
		key := sorted[i]
		j := i - 1
		for j >= 0 && sorted[j] > key {
			sorted[j+1] = sorted[j]
			j--
		}
		sorted[j+1] = key
	}

	result := "union:"
	for i, component := range sorted {
		if i > 0 {
			result += "|"
		}
		result += component
	}
	return result
}

// makeIntersectionKey creates a consistent key for intersection type caching
func (tc *TypeCache) makeIntersectionKey(components []string) string {
	// Sort components to ensure consistent keys
	sorted := make([]string, len(components))
	copy(sorted, components)

	// Simple insertion sort
	for i := 1; i < len(sorted); i++ {
		key := sorted[i]
		j := i - 1
		for j >= 0 && sorted[j] > key {
			sorted[j+1] = sorted[j]
			j--
		}
		sorted[j+1] = key
	}

	result := "intersection:"
	for i, component := range sorted {
		if i > 0 {
			result += "&"
		}
		result += component
	}
	return result
}

// evictLRUUnion removes the least recently used union type entry
func (tc *TypeCache) evictLRUUnion() {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range tc.unionCache {
		if oldestKey == "" || entry.LastAccess.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.LastAccess
		}
	}

	if oldestKey != "" {
		delete(tc.unionCache, oldestKey)
	}
}

// evictLRUIntersection removes the least recently used intersection type entry
func (tc *TypeCache) evictLRUIntersection() {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range tc.intersectionCache {
		if oldestKey == "" || entry.LastAccess.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.LastAccess
		}
	}

	if oldestKey != "" {
		delete(tc.intersectionCache, oldestKey)
	}
}

// GetCacheStats returns cache performance statistics
func (tc *TypeCache) GetCacheStats() (hits, misses int64, hitRateResult float64, cacheSize int) {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	total := tc.hits + tc.misses
	if total > 0 {
		hitRateResult = float64(tc.hits) / float64(total)
	}

	totalCacheSize := len(tc.unionCache) + len(tc.intersectionCache) + len(tc.compatibilityCache)
	return tc.hits, tc.misses, hitRateResult, totalCacheSize
}

// LazyTypeResolver implements lazy evaluation for type resolution
type LazyTypeResolver struct {
	typeCache       *TypeCache
	unresolvedTypes map[string]*LazyType
	resolutionQueue []*LazyType
	isResolving     bool
	mu              sync.Mutex
}

// LazyType represents a type that will be resolved on demand
type LazyType struct {
	Name            string
	Dependencies    []string
	ResolutionFunc  func() (string, error)
	Resolved        bool
	ResolvedType    string
	ResolutionError error
	ResolutionTime  time.Duration
}

// NewLazyTypeResolver creates a new lazy type resolver
func NewLazyTypeResolver() *LazyTypeResolver {
	return &LazyTypeResolver{
		typeCache:       NewTypeCache(5000),
		unresolvedTypes: make(map[string]*LazyType),
		resolutionQueue: make([]*LazyType, 0, 100),
	}
}

// AddLazyType adds a type for lazy resolution
func (ltr *LazyTypeResolver) AddLazyType(name string, dependencies []string, resolutionFunc func() (string, error)) {
	ltr.mu.Lock()
	defer ltr.mu.Unlock()

	lazyType := &LazyType{
		Name:           name,
		Dependencies:   dependencies,
		ResolutionFunc: resolutionFunc,
		Resolved:       false,
	}

	ltr.unresolvedTypes[name] = lazyType
	ltr.resolutionQueue = append(ltr.resolutionQueue, lazyType)
}

// ResolveType resolves a type on demand
func (ltr *LazyTypeResolver) ResolveType(name string) (string, error) {
	ltr.mu.Lock()
	defer ltr.mu.Unlock()

	lazyType, exists := ltr.unresolvedTypes[name]
	if !exists {
		return "", nil // Type not found
	}

	if lazyType.Resolved {
		return lazyType.ResolvedType, lazyType.ResolutionError
	}

	// Resolve dependencies first
	for _, dep := range lazyType.Dependencies {
		if _, err := ltr.ResolveType(dep); err != nil {
			return "", err
		}
	}

	// Resolve the type
	start := time.Now()
	resolvedType, err := lazyType.ResolutionFunc()

	lazyType.Resolved = true
	lazyType.ResolvedType = resolvedType
	lazyType.ResolutionError = err
	lazyType.ResolutionTime = time.Since(start)

	return resolvedType, err
}

// ResolveBatch resolves multiple types efficiently
func (ltr *LazyTypeResolver) ResolveBatch(names []string) map[string]string {
	results := make(map[string]string)

	for _, name := range names {
		if resolvedType, err := ltr.ResolveType(name); err == nil {
			results[name] = resolvedType
		}
	}

	return results
}

// OptimizedUnionType implements fast union type operations
type OptimizedUnionType struct {
	components    []string
	componentSet  map[string]bool // For O(1) membership testing
	normalized    bool
	normalizedStr string
}

// NewOptimizedUnionType creates an optimized union type
func NewOptimizedUnionType(components []string) *OptimizedUnionType {
	// Remove duplicates and create set for fast lookup
	componentSet := make(map[string]bool)
	uniqueComponents := make([]string, 0, len(components))

	for _, component := range components {
		if !componentSet[component] {
			componentSet[component] = true
			uniqueComponents = append(uniqueComponents, component)
		}
	}

	return &OptimizedUnionType{
		components:   uniqueComponents,
		componentSet: componentSet,
		normalized:   false,
	}
}

// Contains checks if a type is part of the union in O(1) time
func (out *OptimizedUnionType) Contains(typeName string) bool {
	return out.componentSet[typeName]
}

// Union combines with another union type efficiently
func (out *OptimizedUnionType) Union(other *OptimizedUnionType) *OptimizedUnionType {
	newComponentSet := make(map[string]bool)
	newComponents := make([]string, 0, len(out.components)+len(other.components))

	// Add components from first union
	for _, component := range out.components {
		if !newComponentSet[component] {
			newComponentSet[component] = true
			newComponents = append(newComponents, component)
		}
	}

	// Add components from second union
	for _, component := range other.components {
		if !newComponentSet[component] {
			newComponentSet[component] = true
			newComponents = append(newComponents, component)
		}
	}

	return &OptimizedUnionType{
		components:   newComponents,
		componentSet: newComponentSet,
		normalized:   false,
	}
}

// Intersection computes intersection with another union type
func (out *OptimizedUnionType) Intersection(other *OptimizedUnionType) *OptimizedUnionType {
	newComponentSet := make(map[string]bool)
	newComponents := make([]string, 0, len(out.components))

	// Find common components
	for _, component := range out.components {
		if other.componentSet[component] && !newComponentSet[component] {
			newComponentSet[component] = true
			newComponents = append(newComponents, component)
		}
	}

	return &OptimizedUnionType{
		components:   newComponents,
		componentSet: newComponentSet,
		normalized:   false,
	}
}

// String returns a normalized string representation
func (out *OptimizedUnionType) String() string {
	if out.normalized {
		return out.normalizedStr
	}

	// Sort components for consistent representation
	sorted := make([]string, len(out.components))
	copy(sorted, out.components)

	// Simple insertion sort for small arrays
	for i := 1; i < len(sorted); i++ {
		key := sorted[i]
		j := i - 1
		for j >= 0 && sorted[j] > key {
			sorted[j+1] = sorted[j]
			j--
		}
		sorted[j+1] = key
	}

	result := ""
	for i, component := range sorted {
		if i > 0 {
			result += "|"
		}
		result += component
	}

	out.normalizedStr = result
	out.normalized = true
	return result
}

// TypeCompatibilityOptimizer speeds up type compatibility checking
type TypeCompatibilityOptimizer struct {
	compatibilityMatrix map[string]map[string]bool
	hierarchyCache      map[string][]string // Type -> parent types
	mu                  sync.RWMutex
}

// NewTypeCompatibilityOptimizer creates a new compatibility optimizer
func NewTypeCompatibilityOptimizer() *TypeCompatibilityOptimizer {
	return &TypeCompatibilityOptimizer{
		compatibilityMatrix: make(map[string]map[string]bool),
		hierarchyCache:      make(map[string][]string),
	}
}

// PrecomputeCompatibility precomputes compatibility for common type pairs
func (tco *TypeCompatibilityOptimizer) PrecomputeCompatibility(commonTypes []string) {
	tco.mu.Lock()
	defer tco.mu.Unlock()

	// Precompute compatibility matrix for common types
	for _, typeA := range commonTypes {
		if tco.compatibilityMatrix[typeA] == nil {
			tco.compatibilityMatrix[typeA] = make(map[string]bool)
		}

		for _, typeB := range commonTypes {
			// This would call actual compatibility checking logic
			compatible := tco.computeCompatibility(typeA, typeB)
			tco.compatibilityMatrix[typeA][typeB] = compatible
		}
	}
}

// IsCompatible checks compatibility using precomputed matrix when possible
func (tco *TypeCompatibilityOptimizer) IsCompatible(typeA, typeB string) (bool, bool) {
	tco.mu.RLock()
	defer tco.mu.RUnlock()

	if matrix, exists := tco.compatibilityMatrix[typeA]; exists {
		if result, found := matrix[typeB]; found {
			return result, true // Found in cache
		}
	}

	return false, false // Not in cache
}

// computeCompatibility implements actual compatibility logic
func (tco *TypeCompatibilityOptimizer) computeCompatibility(typeA, typeB string) bool {
	// Simplified compatibility logic
	// In practice, this would implement full type compatibility rules

	// Same types are compatible
	if typeA == typeB {
		return true
	}

	// Basic Perl type compatibility
	switch typeA {
	case "Str":
		return typeB == "Any" || typeB == "Scalar"
	case "Int":
		return typeB == "Num" || typeB == "Any" || typeB == "Scalar"
	case "Num":
		return typeB == "Any" || typeB == "Scalar"
	case "Any":
		return true
	default:
		return false
	}
}
