// ABOUTME: Implements a multi-level caching system with memory, disk, and distributed tiers
// ABOUTME: Provides efficient caching with automatic tier promotion and eviction policies

package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/log"
)

// CacheEntry represents an entry in the cache
type CacheEntry struct {
	Key        string
	Value      interface{}
	Size       int64
	TTL        time.Duration
	CreatedAt  time.Time
	AccessedAt time.Time
	HitCount   int64
}

// CacheTier represents a level in the cache hierarchy
type CacheTier interface {
	Get(ctx context.Context, key string) (*CacheEntry, error)
	Set(ctx context.Context, entry *CacheEntry) error
	Delete(ctx context.Context, key string) error
	Clear(ctx context.Context) error
	Stats() *TierStats
}

// TierStats contains statistics for a cache tier
type TierStats struct {
	Hits       int64
	Misses     int64
	Evictions  int64
	Size       int64
	MaxSize    int64
	EntryCount int64
}

// MultiLevelCache implements a hierarchical caching system
type MultiLevelCache struct {
	tiers      []CacheTier
	mu         sync.RWMutex
	config     *CacheConfig
	compressor Compressor
	logger     *log.Logger
}

// CacheConfig contains configuration for the multi-level cache
type CacheConfig struct {
	MemorySize      int64         // Max memory cache size in bytes
	DiskSize        int64         // Max disk cache size in bytes
	DefaultTTL      time.Duration // Default TTL for entries
	CompressionType string        // "none", "gzip", "zstd"
	EvictionPolicy  string        // "lru", "lfu", "fifo"
	TierPromotion   bool          // Enable automatic tier promotion
}

// NewMultiLevelCache creates a new multi-level cache
func NewMultiLevelCache(config *CacheConfig, logger *log.Logger) (*MultiLevelCache, error) {
	if config == nil {
		config = &CacheConfig{
			MemorySize:      100 * 1024 * 1024,  // 100MB
			DiskSize:        1024 * 1024 * 1024, // 1GB
			DefaultTTL:      24 * time.Hour,
			CompressionType: "zstd",
			EvictionPolicy:  "lru",
			TierPromotion:   true,
		}
	}

	compressor, err := NewCompressor(config.CompressionType)
	if err != nil {
		return nil, errors.Wrap(err, "PSC", "cache", "001", "failed to create compressor")
	}

	mlc := &MultiLevelCache{
		config:     config,
		compressor: compressor,
		logger:     logger,
		tiers:      make([]CacheTier, 0),
	}

	// Initialize memory tier
	memoryTier := NewMemoryTier(config.MemorySize, config.EvictionPolicy)
	mlc.tiers = append(mlc.tiers, memoryTier)

	// Initialize disk tier
	diskTier, err := NewDiskTier(config.DiskSize, config.EvictionPolicy, compressor)
	if err != nil {
		return nil, errors.Wrap(err, "PSC", "cache", "002", "failed to create disk tier")
	}
	mlc.tiers = append(mlc.tiers, diskTier)

	return mlc, nil
}

// Get retrieves a value from the cache
func (c *MultiLevelCache) Get(ctx context.Context, key string) (interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for i, tier := range c.tiers {
		entry, err := tier.Get(ctx, key)
		if err == nil && entry != nil {
			// Check TTL
			if c.isExpired(entry) {
				c.deleteFromAllTiers(ctx, key)
				return nil, errors.NewSystemError("007", "cache entry expired", nil)
			}

			// Promote to higher tiers if enabled
			if c.config.TierPromotion && i > 0 {
				c.promoteEntry(ctx, entry, i)
			}

			return entry.Value, nil
		}
	}

	return nil, errors.NewSystemError("008", "cache miss", nil)
}

// Set stores a value in the cache
func (c *MultiLevelCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ttl == 0 {
		ttl = c.config.DefaultTTL
	}

	// Calculate size
	size, err := c.calculateSize(value)
	if err != nil {
		return errors.Wrap(err, "PSC", "cache", "003", "failed to calculate size")
	}

	entry := &CacheEntry{
		Key:        key,
		Value:      value,
		Size:       size,
		TTL:        ttl,
		CreatedAt:  time.Now(),
		AccessedAt: time.Now(),
		HitCount:   0,
	}

	// Try to set in each tier, starting from the highest performance tier
	for _, tier := range c.tiers {
		if err := tier.Set(ctx, entry); err == nil {
			return nil
		}
	}

	return errors.NewSystemError("009", "failed to set in any cache tier", nil)
}

// Delete removes a key from all cache tiers
func (c *MultiLevelCache) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.deleteFromAllTiers(ctx, key)
}

// Clear removes all entries from all cache tiers
func (c *MultiLevelCache) Clear(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var errs []error
	for _, tier := range c.tiers {
		if err := tier.Clear(ctx); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errors.NewSystemError("010", fmt.Sprintf("failed to clear all tiers: %v", errs), nil)
	}

	return nil
}

// Stats returns statistics for all cache tiers
func (c *MultiLevelCache) Stats() map[string]*TierStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := make(map[string]*TierStats)
	stats["memory"] = c.tiers[0].Stats()
	if len(c.tiers) > 1 {
		stats["disk"] = c.tiers[1].Stats()
	}
	if len(c.tiers) > 2 {
		stats["distributed"] = c.tiers[2].Stats()
	}

	return stats
}

// AddDistributedTier adds a distributed cache tier
func (c *MultiLevelCache) AddDistributedTier(tier CacheTier) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.tiers = append(c.tiers, tier)
}

// isExpired checks if a cache entry has expired
func (c *MultiLevelCache) isExpired(entry *CacheEntry) bool {
	return time.Since(entry.CreatedAt) > entry.TTL
}

// deleteFromAllTiers removes a key from all tiers
func (c *MultiLevelCache) deleteFromAllTiers(ctx context.Context, key string) error {
	var errs []error
	for _, tier := range c.tiers {
		if err := tier.Delete(ctx, key); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errors.NewSystemError("011", fmt.Sprintf("failed to delete from all tiers: %v", errs), nil)
	}

	return nil
}

// promoteEntry promotes an entry to higher performance tiers
func (c *MultiLevelCache) promoteEntry(ctx context.Context, entry *CacheEntry, currentTier int) {
	// Update access time and hit count
	entry.AccessedAt = time.Now()
	entry.HitCount++

	// Promote to all higher tiers
	for i := currentTier - 1; i >= 0; i-- {
		if err := c.tiers[i].Set(ctx, entry); err != nil {
			c.logger.Debugf("Failed to promote entry to tier %d: %v", i, err)
			break
		}
	}
}

// calculateSize estimates the size of a value
func (c *MultiLevelCache) calculateSize(value interface{}) (int64, error) {
	// Use JSON encoding to estimate size
	data, err := json.Marshal(value)
	if err != nil {
		return 0, err
	}
	return int64(len(data)), nil
}

// MemoryTier implements an in-memory cache tier
type MemoryTier struct {
	mu             sync.RWMutex
	cache          map[string]*CacheEntry
	evictionPolicy string
	maxSize        int64
	currentSize    int64
	stats          *TierStats
	lru            *LRUList
}

// NewMemoryTier creates a new memory cache tier
func NewMemoryTier(maxSize int64, evictionPolicy string) *MemoryTier {
	return &MemoryTier{
		cache:          make(map[string]*CacheEntry),
		evictionPolicy: evictionPolicy,
		maxSize:        maxSize,
		stats:          &TierStats{MaxSize: maxSize},
		lru:            NewLRUList(),
	}
}

// Get retrieves an entry from the memory tier
func (m *MemoryTier) Get(ctx context.Context, key string) (*CacheEntry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entry, exists := m.cache[key]
	if !exists {
		m.stats.Misses++
		return nil, errors.NewSystemError("012", "key not found", nil)
	}

	m.stats.Hits++
	if m.evictionPolicy == "lru" {
		m.lru.MoveToFront(key)
	}

	return entry, nil
}

// Set stores an entry in the memory tier
func (m *MemoryTier) Set(ctx context.Context, entry *CacheEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if we need to evict entries
	if m.currentSize+entry.Size > m.maxSize {
		if err := m.evict(entry.Size); err != nil {
			return err
		}
	}

	m.cache[entry.Key] = entry
	m.currentSize += entry.Size
	m.stats.EntryCount = int64(len(m.cache))

	if m.evictionPolicy == "lru" {
		m.lru.Add(entry.Key)
	}

	return nil
}

// Delete removes an entry from the memory tier
func (m *MemoryTier) Delete(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry, exists := m.cache[key]
	if !exists {
		return nil
	}

	delete(m.cache, key)
	m.currentSize -= entry.Size
	m.stats.EntryCount = int64(len(m.cache))

	if m.evictionPolicy == "lru" {
		m.lru.Remove(key)
	}

	return nil
}

// Clear removes all entries from the memory tier
func (m *MemoryTier) Clear(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.cache = make(map[string]*CacheEntry)
	m.currentSize = 0
	m.stats.EntryCount = 0
	m.lru = NewLRUList()

	return nil
}

// Stats returns statistics for the memory tier
func (m *MemoryTier) Stats() *TierStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	m.stats.Size = m.currentSize
	return m.stats
}

// evict removes entries to make space
func (m *MemoryTier) evict(requiredSpace int64) error {
	switch m.evictionPolicy {
	case "lru":
		return m.evictLRU(requiredSpace)
	case "lfu":
		return m.evictLFU(requiredSpace)
	case "fifo":
		return m.evictFIFO(requiredSpace)
	default:
		return errors.NewSystemError("013", fmt.Sprintf("unknown eviction policy: %s", m.evictionPolicy), nil)
	}
}

// evictLRU implements LRU eviction
func (m *MemoryTier) evictLRU(requiredSpace int64) error {
	freedSpace := int64(0)

	for freedSpace < requiredSpace && m.lru.Len() > 0 {
		key := m.lru.RemoveLast()
		if entry, exists := m.cache[key]; exists {
			delete(m.cache, key)
			freedSpace += entry.Size
			m.currentSize -= entry.Size
			m.stats.Evictions++
		}
	}

	if freedSpace < requiredSpace {
		return errors.NewSystemError("014", "insufficient space after eviction", nil)
	}

	return nil
}

// evictLFU implements LFU eviction (simplified)
func (m *MemoryTier) evictLFU(requiredSpace int64) error {
	// Find entries with lowest hit count
	type entryHit struct {
		key   string
		count int64
	}

	entries := make([]entryHit, 0, len(m.cache))
	for key, entry := range m.cache {
		entries = append(entries, entryHit{key: key, count: entry.HitCount})
	}

	// Sort by hit count (ascending)
	// Simple bubble sort for demonstration
	for i := 0; i < len(entries)-1; i++ {
		for j := 0; j < len(entries)-i-1; j++ {
			if entries[j].count > entries[j+1].count {
				entries[j], entries[j+1] = entries[j+1], entries[j]
			}
		}
	}

	freedSpace := int64(0)
	for _, eh := range entries {
		if freedSpace >= requiredSpace {
			break
		}

		if entry, exists := m.cache[eh.key]; exists {
			delete(m.cache, eh.key)
			freedSpace += entry.Size
			m.currentSize -= entry.Size
			m.stats.Evictions++
		}
	}

	if freedSpace < requiredSpace {
		return errors.NewSystemError("014", "insufficient space after eviction", nil)
	}

	return nil
}

// evictFIFO implements FIFO eviction
func (m *MemoryTier) evictFIFO(requiredSpace int64) error {
	// Find oldest entries
	type entryTime struct {
		key  string
		time time.Time
	}

	entries := make([]entryTime, 0, len(m.cache))
	for key, entry := range m.cache {
		entries = append(entries, entryTime{key: key, time: entry.CreatedAt})
	}

	// Sort by creation time (ascending)
	for i := 0; i < len(entries)-1; i++ {
		for j := 0; j < len(entries)-i-1; j++ {
			if entries[j].time.After(entries[j+1].time) {
				entries[j], entries[j+1] = entries[j+1], entries[j]
			}
		}
	}

	freedSpace := int64(0)
	for _, et := range entries {
		if freedSpace >= requiredSpace {
			break
		}

		if entry, exists := m.cache[et.key]; exists {
			delete(m.cache, et.key)
			freedSpace += entry.Size
			m.currentSize -= entry.Size
			m.stats.Evictions++
		}
	}

	if freedSpace < requiredSpace {
		return errors.NewSystemError("014", "insufficient space after eviction", nil)
	}

	return nil
}

// LRUList implements a doubly linked list for LRU tracking
type LRUList struct {
	head   *LRUNode
	tail   *LRUNode
	nodes  map[string]*LRUNode
	length int
}

// LRUNode represents a node in the LRU list
type LRUNode struct {
	key  string
	prev *LRUNode
	next *LRUNode
}

// NewLRUList creates a new LRU list
func NewLRUList() *LRUList {
	return &LRUList{
		nodes: make(map[string]*LRUNode),
	}
}

// Add adds a key to the front of the list
func (l *LRUList) Add(key string) {
	if node, exists := l.nodes[key]; exists {
		l.removeNode(node)
	}

	node := &LRUNode{key: key}
	l.nodes[key] = node

	if l.head == nil {
		l.head = node
		l.tail = node
	} else {
		node.next = l.head
		l.head.prev = node
		l.head = node
	}

	l.length++
}

// MoveToFront moves a key to the front of the list
func (l *LRUList) MoveToFront(key string) {
	if node, exists := l.nodes[key]; exists {
		l.removeNode(node)
		l.Add(key)
	}
}

// Remove removes a key from the list
func (l *LRUList) Remove(key string) {
	if node, exists := l.nodes[key]; exists {
		l.removeNode(node)
		delete(l.nodes, key)
		l.length--
	}
}

// RemoveLast removes and returns the last key
func (l *LRUList) RemoveLast() string {
	if l.tail == nil {
		return ""
	}

	key := l.tail.key
	l.removeNode(l.tail)
	delete(l.nodes, key)
	l.length--

	return key
}

// Len returns the length of the list
func (l *LRUList) Len() int {
	return l.length
}

// removeNode removes a node from the list
func (l *LRUList) removeNode(node *LRUNode) {
	if node.prev != nil {
		node.prev.next = node.next
	} else {
		l.head = node.next
	}

	if node.next != nil {
		node.next.prev = node.prev
	} else {
		l.tail = node.prev
	}
}

// DiskTier implements a disk-based cache tier
type DiskTier struct {
	// Implementation would include file-based storage
	// Placeholder for now
	stats *TierStats
}

// NewDiskTier creates a new disk cache tier
func NewDiskTier(maxSize int64, evictionPolicy string, compressor Compressor) (*DiskTier, error) {
	return &DiskTier{
		stats: &TierStats{MaxSize: maxSize},
	}, nil
}

// Get retrieves an entry from the disk tier
func (d *DiskTier) Get(ctx context.Context, key string) (*CacheEntry, error) {
	// Placeholder implementation
	return nil, errors.NewSystemError("015", "not implemented", nil)
}

// Set stores an entry in the disk tier
func (d *DiskTier) Set(ctx context.Context, entry *CacheEntry) error {
	// Placeholder implementation
	return nil
}

// Delete removes an entry from the disk tier
func (d *DiskTier) Delete(ctx context.Context, key string) error {
	// Placeholder implementation
	return nil
}

// Clear removes all entries from the disk tier
func (d *DiskTier) Clear(ctx context.Context) error {
	// Placeholder implementation
	return nil
}

// Stats returns statistics for the disk tier
func (d *DiskTier) Stats() *TierStats {
	return d.stats
}
