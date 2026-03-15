// ABOUTME: Performance caching system for PVM operations
// ABOUTME: Provides intelligent caching with TTL and memory-aware eviction

package performance

import (
	"crypto/sha256"
	"fmt"
	"sync"
	"time"
)

// Cache provides high-performance caching with TTL and LRU eviction
type Cache struct {
	mu        sync.RWMutex
	items     map[string]*CacheItem
	access    []string // LRU order
	maxSize   int
	maxAge    time.Duration
	hitCount  int64
	missCount int64
}

// CacheItem represents a cached value with metadata
type CacheItem struct {
	Value       interface{}
	CreatedAt   time.Time
	LastAccess  time.Time
	AccessCount int64
	Size        int64
}

// NewCache creates a new cache with specified max size and TTL
func NewCache(maxSize int, maxAge time.Duration) *Cache {
	cache := &Cache{
		items:   make(map[string]*CacheItem),
		access:  make([]string, 0, maxSize),
		maxSize: maxSize,
		maxAge:  maxAge,
	}

	// Start cleanup goroutine
	go cache.periodicCleanup()

	return cache
}

// Get retrieves a value from the cache
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, exists := c.items[key]
	if !exists {
		c.missCount++
		return nil, false
	}

	// Check if item has expired
	if c.maxAge > 0 && time.Since(item.CreatedAt) > c.maxAge {
		delete(c.items, key)
		c.removeFromAccess(key)
		c.missCount++
		return nil, false
	}

	// Update access information
	item.LastAccess = time.Now()
	item.AccessCount++
	c.updateAccess(key)
	c.hitCount++

	return item.Value, true
}

// Set stores a value in the cache
func (c *Cache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	size := estimateSize(value)

	// Create or update item
	now := time.Now()
	item := &CacheItem{
		Value:       value,
		CreatedAt:   now,
		LastAccess:  now,
		AccessCount: 1,
		Size:        size,
	}

	c.items[key] = item
	c.updateAccess(key)

	// Evict if necessary
	c.evictIfNecessary()
}

// Delete removes a key from the cache
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
	c.removeFromAccess(key)
}

// Clear removes all items from the cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*CacheItem)
	c.access = c.access[:0]
	c.hitCount = 0
	c.missCount = 0
}

// Stats returns cache statistics
func (c *Cache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var totalSize int64
	for _, item := range c.items {
		totalSize += item.Size
	}

	total := c.hitCount + c.missCount
	hitRatio := float64(0)
	if total > 0 {
		hitRatio = float64(c.hitCount) / float64(total)
	}

	return CacheStats{
		Items:     len(c.items),
		HitCount:  c.hitCount,
		MissCount: c.missCount,
		HitRatio:  hitRatio,
		TotalSize: totalSize,
	}
}

// CacheStats holds cache performance statistics
type CacheStats struct {
	Items     int
	HitCount  int64
	MissCount int64
	HitRatio  float64
	TotalSize int64
}

// updateAccess moves a key to the front of the access list
func (c *Cache) updateAccess(key string) {
	// Remove from current position
	c.removeFromAccess(key)

	// Add to front
	c.access = append([]string{key}, c.access...)
}

// removeFromAccess removes a key from the access list
func (c *Cache) removeFromAccess(key string) {
	for i, k := range c.access {
		if k == key {
			c.access = append(c.access[:i], c.access[i+1:]...)
			break
		}
	}
}

// evictIfNecessary removes items if cache is full
func (c *Cache) evictIfNecessary() {
	for len(c.items) > c.maxSize && len(c.access) > 0 {
		// Remove least recently used item
		lru := c.access[len(c.access)-1]
		delete(c.items, lru)
		c.access = c.access[:len(c.access)-1]
	}
}

// periodicCleanup removes expired items
func (c *Cache) periodicCleanup() {
	if c.maxAge <= 0 {
		return // No TTL, no cleanup needed
	}

	ticker := time.NewTicker(c.maxAge / 4) // Cleanup 4 times per TTL period
	defer ticker.Stop()

	for range ticker.C {
		c.cleanupExpired()
	}
}

// cleanupExpired removes expired items
func (c *Cache) cleanupExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	var keysToDelete []string

	for key, item := range c.items {
		if time.Since(item.CreatedAt) > c.maxAge {
			keysToDelete = append(keysToDelete, key)
		}
	}

	for _, key := range keysToDelete {
		delete(c.items, key)
		c.removeFromAccess(key)
	}
}

// estimateSize provides a rough estimate of memory usage
func estimateSize(value interface{}) int64 {
	switch v := value.(type) {
	case string:
		return int64(len(v))
	case []byte:
		return int64(len(v))
	case int, int32, int64, float32, float64:
		return 8
	case bool:
		return 1
	default:
		// Rough estimate for complex types
		return 100
	}
}

// HashKey creates a consistent hash key from multiple values
func HashKey(values ...interface{}) string {
	hasher := sha256.New()
	for _, v := range values {
		fmt.Fprintf(hasher, "%v", v)
	}
	return fmt.Sprintf("%x", hasher.Sum(nil))[:16] // Use first 16 characters
}

// Global caches for common use cases
var (
	// ParserCache caches parsing results
	ParserCache = NewCache(1000, 30*time.Minute)

	// TypeCache caches type checking results
	TypeCache = NewCache(500, 15*time.Minute)

	// FileCache caches file content and metadata
	FileCache = NewCache(200, 10*time.Minute)
)
