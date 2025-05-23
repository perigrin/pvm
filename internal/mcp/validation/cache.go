// ABOUTME: LRU cache implementation for validation results
// ABOUTME: Provides caching with size limits and smart invalidation

package validation

import (
	"container/list"
	"crypto/sha256"
	"fmt"
	"strings"
	"sync"
	"time"
)

// ValidationCache provides LRU caching for validation results
type ValidationCache struct {
	mu           sync.RWMutex
	maxSize      int64                      // Maximum cache size in bytes
	currentSize  int64                      // Current cache size in bytes
	cache        map[string]*list.Element   // Map of cache keys to list elements
	lru          *list.List                 // LRU list
	projectIndex map[string]map[string]bool // Index of cache keys by project
}

// cacheEntry represents an entry in the cache
type cacheEntry struct {
	key       string
	result    *ValidationResult
	size      int64
	timestamp time.Time
	project   string
}

// NewValidationCache creates a new validation cache
func NewValidationCache(maxSizeStr string) (*ValidationCache, error) {
	maxSize, err := parseSize(maxSizeStr)
	if err != nil {
		return nil, fmt.Errorf("invalid cache size: %w", err)
	}

	return &ValidationCache{
		maxSize:      maxSize,
		cache:        make(map[string]*list.Element),
		lru:          list.New(),
		projectIndex: make(map[string]map[string]bool),
	}, nil
}

// GenerateKey generates a cache key from code content and project context
func (c *ValidationCache) GenerateKey(code string, projectPath string) string {
	h := sha256.New()
	h.Write([]byte(code))
	h.Write([]byte("|"))
	h.Write([]byte(projectPath))
	return fmt.Sprintf("%s:%x", projectPath, h.Sum(nil))
}

// Get retrieves a validation result from the cache
func (c *ValidationCache) Get(key string) (*ValidationResult, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, found := c.cache[key]
	if !found {
		return nil, false
	}

	// Move to front (most recently used)
	c.lru.MoveToFront(elem)
	entry := elem.Value.(*cacheEntry)

	return entry.result, true
}

// Set stores a validation result in the cache
func (c *ValidationCache) Set(key string, result *ValidationResult) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Calculate size of the result
	size := c.estimateSize(result)

	// Check if we already have this key
	if elem, exists := c.cache[key]; exists {
		// Update existing entry
		entry := elem.Value.(*cacheEntry)
		c.currentSize -= entry.size
		entry.result = result
		entry.size = size
		entry.timestamp = time.Now()
		c.currentSize += size
		c.lru.MoveToFront(elem)
		return
	}

	// Create new entry
	project := extractProject(key)
	entry := &cacheEntry{
		key:       key,
		result:    result,
		size:      size,
		timestamp: time.Now(),
		project:   project,
	}

	// Add to cache
	elem := c.lru.PushFront(entry)
	c.cache[key] = elem
	c.currentSize += size

	// Update project index
	if entry.project != "" {
		if c.projectIndex[entry.project] == nil {
			c.projectIndex[entry.project] = make(map[string]bool)
		}
		c.projectIndex[entry.project][key] = true
	}

	// Evict entries if necessary
	c.evictIfNeeded()
}

// ClearProject clears all cache entries for a specific project
func (c *ValidationCache) ClearProject(projectPath string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	keys, exists := c.projectIndex[projectPath]
	if !exists {
		return
	}

	// Remove all entries for this project
	for key := range keys {
		if elem, exists := c.cache[key]; exists {
			c.removeElement(elem)
		}
	}

	// Clear project index
	delete(c.projectIndex, projectPath)
}

// Clear removes all entries from the cache
func (c *ValidationCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[string]*list.Element)
	c.lru = list.New()
	c.projectIndex = make(map[string]map[string]bool)
	c.currentSize = 0
}

// GetStats returns cache statistics
func (c *ValidationCache) GetStats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return CacheStats{
		MaxSize:     c.maxSize,
		CurrentSize: c.currentSize,
		EntryCount:  len(c.cache),
		HitRate:     0, // Would need to track hits/misses for this
	}
}

// CacheStats represents cache statistics
type CacheStats struct {
	MaxSize     int64   `json:"max_size"`
	CurrentSize int64   `json:"current_size"`
	EntryCount  int     `json:"entry_count"`
	HitRate     float64 `json:"hit_rate"`
}

// evictIfNeeded removes least recently used entries if cache is full
func (c *ValidationCache) evictIfNeeded() {
	for c.currentSize > c.maxSize && c.lru.Len() > 0 {
		// Remove least recently used
		elem := c.lru.Back()
		if elem != nil {
			c.removeElement(elem)
		}
	}
}

// removeElement removes an element from the cache
func (c *ValidationCache) removeElement(elem *list.Element) {
	entry := elem.Value.(*cacheEntry)

	// Remove from cache map
	delete(c.cache, entry.key)

	// Remove from LRU list
	c.lru.Remove(elem)

	// Update size
	c.currentSize -= entry.size

	// Remove from project index
	if entry.project != "" && c.projectIndex[entry.project] != nil {
		delete(c.projectIndex[entry.project], entry.key)
		if len(c.projectIndex[entry.project]) == 0 {
			delete(c.projectIndex, entry.project)
		}
	}
}

// estimateSize estimates the memory size of a validation result
func (c *ValidationCache) estimateSize(result *ValidationResult) int64 {
	// Basic estimation based on content
	size := int64(0)

	// Add size for basic fields
	size += 100 // Base overhead

	// Add size for errors
	size += int64(len(result.Errors) * 200)

	// Add size for warnings
	size += int64(len(result.Warnings) * 150)

	// Add size for type info
	for _, info := range result.TypeInfo {
		size += int64(len(info.Name) + len(info.Type) + 50)
	}

	return size
}

// parseSize parses a size string like "50MB" to bytes
func parseSize(sizeStr string) (int64, error) {
	sizeStr = strings.TrimSpace(strings.ToUpper(sizeStr))

	// Check suffixes in order from longest to shortest
	suffixes := []struct {
		suffix     string
		multiplier int64
	}{
		{"GB", 1024 * 1024 * 1024},
		{"MB", 1024 * 1024},
		{"KB", 1024},
		{"B", 1},
	}

	// Check for known suffixes
	for _, s := range suffixes {
		if strings.HasSuffix(sizeStr, s.suffix) {
			numStr := strings.TrimSuffix(sizeStr, s.suffix)
			if numStr == "" {
				return 0, fmt.Errorf("invalid size format: %s", sizeStr)
			}
			var num int64
			if _, err := fmt.Sscanf(numStr, "%d", &num); err != nil {
				return 0, fmt.Errorf("invalid size format: %s", sizeStr)
			}
			// Additional validation: ensure numStr only contains digits
			for _, c := range numStr {
				if c < '0' || c > '9' {
					return 0, fmt.Errorf("invalid size format: %s", sizeStr)
				}
			}
			return num * s.multiplier, nil
		}
	}

	// Check if it has any alphabetic characters
	for i := 0; i < len(sizeStr); i++ {
		if (sizeStr[i] >= 'A' && sizeStr[i] <= 'Z') || (sizeStr[i] >= 'a' && sizeStr[i] <= 'z') {
			// Has letters but not a known suffix
			return 0, fmt.Errorf("invalid size format: %s", sizeStr)
		}
	}

	// Try parsing as raw number (bytes)
	var num int64
	if _, err := fmt.Sscanf(sizeStr, "%d", &num); err != nil {
		return 0, fmt.Errorf("invalid size format: %s", sizeStr)
	}
	return num, nil
}

// extractProject extracts project path from cache key
func extractProject(key string) string {
	// Key format is "projectPath:hash"
	parts := strings.SplitN(key, ":", 2)
	if len(parts) >= 1 {
		return parts[0]
	}
	return ""
}
