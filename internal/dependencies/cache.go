// ABOUTME: This file provides dependency caching functionality for improved resolution performance.
// ABOUTME: It implements a simple in-memory cache with TTL support for dependency resolution results.

package dependencies

import (
	"sync"
	"time"
)

// dependencyCache provides caching for dependency resolution results
type dependencyCache struct {
	cache map[string]*cacheEntry
	mutex sync.RWMutex
	ttl   time.Duration
}

// cacheEntry represents a cached dependency node with expiration
type cacheEntry struct {
	node      *DependencyNode
	timestamp time.Time
}

// newDependencyCache creates a new dependency cache with default TTL
func newDependencyCache() *dependencyCache {
	return &dependencyCache{
		cache: make(map[string]*cacheEntry),
		ttl:   24 * time.Hour,
	}
}

// get retrieves a dependency node from cache if it exists and is not expired
func (dc *dependencyCache) get(moduleName string) *DependencyNode {
	dc.mutex.RLock()
	defer dc.mutex.RUnlock()

	entry, exists := dc.cache[moduleName]
	if !exists {
		return nil
	}

	// Check if entry has expired
	if time.Since(entry.timestamp) > dc.ttl {
		// Remove expired entry
		delete(dc.cache, moduleName)
		return nil
	}

	return entry.node
}

// put stores a dependency node in cache with current timestamp
func (dc *dependencyCache) put(moduleName string, node *DependencyNode) {
	dc.mutex.Lock()
	defer dc.mutex.Unlock()

	dc.cache[moduleName] = &cacheEntry{
		node:      node,
		timestamp: time.Now(),
	}
}

// clear removes all entries from the cache
func (dc *dependencyCache) clear() {
	dc.mutex.Lock()
	defer dc.mutex.Unlock()

	dc.cache = make(map[string]*cacheEntry)
}

// cleanup removes expired entries from the cache
func (dc *dependencyCache) cleanup() {
	dc.mutex.Lock()
	defer dc.mutex.Unlock()

	now := time.Now()
	for moduleName, entry := range dc.cache {
		if now.Sub(entry.timestamp) > dc.ttl {
			delete(dc.cache, moduleName)
		}
	}
}

// size returns the number of cached entries
func (dc *dependencyCache) size() int {
	dc.mutex.RLock()
	defer dc.mutex.RUnlock()

	return len(dc.cache)
}

// setTTL updates the cache TTL
func (dc *dependencyCache) setTTL(ttl time.Duration) {
	dc.mutex.Lock()
	defer dc.mutex.Unlock()

	dc.ttl = ttl
}
