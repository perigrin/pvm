// ABOUTME: Core object pooling infrastructure following Microsoft TypeScript-Go patterns
// ABOUTME: Provides arena-style allocators with automatic growth and statistics tracking

package core

import (
	"slices"
	"sync"
	"sync/atomic"
)

// Pool allocator with arena-style allocation following TypeScript-Go patterns
type Pool[T any] struct {
	data []T
	mu   sync.Mutex

	// Statistics
	allocations int64
	grows       int64
	totalSize   int64
}

// New allocates a single element in the pool and returns a pointer to it.
// If the pool is at capacity, it grows to the next size class.
func (p *Pool[T]) New() *T {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.data) == cap(p.data) {
		nextSize := nextPoolSize(len(p.data))
		// Use slices.Grow() for optimal memory allocation
		p.data = slices.Grow[[]T](nil, nextSize)
		atomic.AddInt64(&p.grows, 1)
		atomic.StoreInt64(&p.totalSize, int64(nextSize))
	}

	index := len(p.data)
	p.data = p.data[:index+1]
	atomic.AddInt64(&p.allocations, 1)

	return &p.data[index]
}

// NewSlice allocates a slice of the given size in the pool.
// If the requested size exceeds pool capacity and can't fit in the next size class,
// it allocates separately. Otherwise, it grows the pool and allocates from it.
func (p *Pool[T]) NewSlice(size int) []T {
	if size == 0 {
		return nil
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.data)+size > cap(p.data) {
		nextSize := nextPoolSize(len(p.data))
		if size > nextSize {
			// Size too large for pool, allocate separately
			atomic.AddInt64(&p.allocations, 1)
			return make([]T, size)
		}
		// Grow pool to accommodate
		p.data = slices.Grow[[]T](nil, nextSize)
		atomic.AddInt64(&p.grows, 1)
		atomic.StoreInt64(&p.totalSize, int64(nextSize))
	}

	newLen := len(p.data) + size
	slice := p.data[len(p.data):newLen:newLen]
	p.data = p.data[:newLen]
	atomic.AddInt64(&p.allocations, 1)

	return slice
}

// Stats returns pool allocation statistics
func (p *Pool[T]) Stats() PoolStats {
	return PoolStats{
		Allocations: atomic.LoadInt64(&p.allocations),
		Grows:       atomic.LoadInt64(&p.grows),
		TotalSize:   atomic.LoadInt64(&p.totalSize),
		CurrentSize: int64(len(p.data)),
		Capacity:    int64(cap(p.data)),
	}
}

// Reset clears the pool for reuse
func (p *Pool[T]) Reset() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.data = p.data[:0] // Keep capacity, reset length
}

// Clear completely empties the pool and resets statistics
func (p *Pool[T]) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.data = nil
	atomic.StoreInt64(&p.allocations, 0)
	atomic.StoreInt64(&p.grows, 0)
	atomic.StoreInt64(&p.totalSize, 0)
}

// PoolStats contains pool performance metrics
type PoolStats struct {
	Allocations int64 // Total number of allocations
	Grows       int64 // Number of pool growth events
	TotalSize   int64 // Total allocated pool size
	CurrentSize int64 // Current used size
	Capacity    int64 // Current pool capacity
}

// nextPoolSize calculates the next pool size following TypeScript-Go's algorithm
func nextPoolSize(size int) int {
	// Branch-free size calculation
	size = max(size, 1)
	size = min(size*2, 256*1024) // Cap at 256K elements
	return size
}

// PoolManager coordinates multiple pools and provides global statistics
type PoolManager struct {
	pools  map[string]PoolStatsProvider
	mu     sync.RWMutex
	config PoolConfig
}

// PoolStatsProvider interface for objects that can provide pool statistics
type PoolStatsProvider interface {
	Stats() PoolStats
}

// PoolConfig contains pool configuration options
type PoolConfig struct {
	InitialSize   int     // Initial pool size
	MaxSize       int     // Maximum pool size (0 = unlimited)
	GrowthFactor  float64 // Growth factor for pool expansion
	StatsEnabled  bool    // Enable detailed statistics collection
	WarnThreshold int64   // Threshold for allocation warnings
	ResetInterval int64   // Interval for automatic pool resets (0 = disabled)
}

// DefaultPoolConfig returns default pool configuration
func DefaultPoolConfig() PoolConfig {
	return PoolConfig{
		InitialSize:   8,
		MaxSize:       256 * 1024,
		GrowthFactor:  2.0,
		StatsEnabled:  true,
		WarnThreshold: 10000,
		ResetInterval: 0,
	}
}

// NewPoolManager creates a new pool manager with the given configuration
func NewPoolManager(config PoolConfig) *PoolManager {
	return &PoolManager{
		pools:  make(map[string]PoolStatsProvider),
		config: config,
	}
}

// RegisterPool registers a pool with the manager for monitoring
func (pm *PoolManager) RegisterPool(name string, pool PoolStatsProvider) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.pools[name] = pool
}

// UnregisterPool removes a pool from monitoring
func (pm *PoolManager) UnregisterPool(name string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	delete(pm.pools, name)
}

// GetPoolStats returns statistics for a specific pool
func (pm *PoolManager) GetPoolStats(name string) (PoolStats, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if pool, exists := pm.pools[name]; exists {
		return pool.Stats(), true
	}
	return PoolStats{}, false
}

// GetAllStats returns statistics for all registered pools
func (pm *PoolManager) GetAllStats() map[string]PoolStats {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	stats := make(map[string]PoolStats, len(pm.pools))
	for name, pool := range pm.pools {
		stats[name] = pool.Stats()
	}
	return stats
}

// GetAggregateStats returns aggregated statistics across all pools
func (pm *PoolManager) GetAggregateStats() PoolStats {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var aggregate PoolStats
	for _, pool := range pm.pools {
		stats := pool.Stats()
		aggregate.Allocations += stats.Allocations
		aggregate.Grows += stats.Grows
		aggregate.TotalSize += stats.TotalSize
		aggregate.CurrentSize += stats.CurrentSize
		aggregate.Capacity += stats.Capacity
	}
	return aggregate
}

// Config returns the current pool manager configuration
func (pm *PoolManager) Config() PoolConfig {
	return pm.config
}

// UpdateConfig updates the pool manager configuration
func (pm *PoolManager) UpdateConfig(config PoolConfig) {
	pm.config = config
}

// Global pool manager instance
var globalPoolManager = NewPoolManager(DefaultPoolConfig())

// GlobalPoolManager returns the global pool manager instance
func GlobalPoolManager() *PoolManager {
	return globalPoolManager
}

// RegisterGlobalPool registers a pool with the global manager
func RegisterGlobalPool(name string, pool PoolStatsProvider) {
	globalPoolManager.RegisterPool(name, pool)
}

// GetGlobalStats returns statistics for all globally registered pools
func GetGlobalStats() map[string]PoolStats {
	return globalPoolManager.GetAllStats()
}
