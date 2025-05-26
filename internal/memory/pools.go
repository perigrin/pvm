// ABOUTME: Object pooling system for frequently allocated PVM objects
// ABOUTME: Provides memory pools for AST nodes, type definitions, and other hot path allocations

package memory

import (
	"sync"
	"unsafe"
)

// PoolStats tracks pool performance metrics
type PoolStats struct {
	Gets    uint64 // Number of Get() calls
	Puts    uint64 // Number of Put() calls
	Hits    uint64 // Number of times pool had available object
	Misses  uint64 // Number of times pool was empty
	Created uint64 // Number of new objects created
	MaxSize int    // Maximum pool size
	Current int    // Current pool size
}

// Pool interface for type-safe object pooling
type Pool[T any] interface {
	Get() *T
	Put(*T)
	Stats() PoolStats
	Clear()
}

// SyncPool implements Pool using sync.Pool with statistics
type SyncPool[T any] struct {
	pool      sync.Pool
	stats     *PoolStats
	statsMu   sync.RWMutex
	newFunc   func() *T
	resetFunc func(*T)
}

// NewSyncPool creates a new synchronized pool
func NewSyncPool[T any](newFunc func() *T, resetFunc func(*T)) *SyncPool[T] {
	p := &SyncPool[T]{
		stats:     &PoolStats{},
		newFunc:   newFunc,
		resetFunc: resetFunc,
	}
	p.pool.New = func() any {
		p.statsMu.Lock()
		p.stats.Created++
		p.statsMu.Unlock()
		return newFunc()
	}
	return p
}

// Get retrieves an object from the pool
func (p *SyncPool[T]) Get() *T {
	p.statsMu.Lock()
	p.stats.Gets++
	p.statsMu.Unlock()

	obj := p.pool.Get().(*T)

	p.statsMu.Lock()
	if p.stats.Created > p.stats.Gets-1 {
		p.stats.Hits++
	} else {
		p.stats.Misses++
	}
	p.statsMu.Unlock()

	return obj
}

// Put returns an object to the pool after resetting it
func (p *SyncPool[T]) Put(obj *T) {
	if obj == nil {
		return
	}

	if p.resetFunc != nil {
		p.resetFunc(obj)
	}

	p.pool.Put(obj)

	p.statsMu.Lock()
	p.stats.Puts++
	p.statsMu.Unlock()
}

// Stats returns current pool statistics
func (p *SyncPool[T]) Stats() PoolStats {
	p.statsMu.RLock()
	defer p.statsMu.RUnlock()
	return *p.stats
}

// Clear empties the pool and resets statistics
func (p *SyncPool[T]) Clear() {
	// Create new pool to clear old objects
	p.pool = sync.Pool{
		New: func() any {
			p.statsMu.Lock()
			p.stats.Created++
			p.statsMu.Unlock()
			return p.newFunc()
		},
	}

	p.statsMu.Lock()
	p.stats = &PoolStats{}
	p.statsMu.Unlock()
}

// SlicePool manages reusable slices with capacity management
type SlicePool[T any] struct {
	pools   []*SyncPool[[]T] // Pools for different capacity buckets
	buckets []int            // Capacity buckets
	stats   *PoolStats
	statsMu sync.RWMutex
}

// NewSlicePool creates a pool for slices with different capacity buckets
func NewSlicePool[T any](buckets []int) *SlicePool[T] {
	if len(buckets) == 0 {
		buckets = []int{8, 16, 32, 64, 128, 256, 512, 1024}
	}

	pools := make([]*SyncPool[[]T], len(buckets))
	for i, capacity := range buckets {
		cap := capacity // Capture for closure
		pools[i] = NewSyncPool(
			func() *[]T {
				s := make([]T, 0, cap)
				return &s
			},
			func(s *[]T) {
				*s = (*s)[:0] // Reset length but keep capacity
			},
		)
	}

	return &SlicePool[T]{
		pools:   pools,
		buckets: buckets,
		stats:   &PoolStats{},
	}
}

// Get retrieves a slice with at least the requested capacity
func (p *SlicePool[T]) Get(minCap int) *[]T {
	p.statsMu.Lock()
	p.stats.Gets++
	p.statsMu.Unlock()

	// Find appropriate bucket
	for i, capacity := range p.buckets {
		if capacity >= minCap {
			slice := p.pools[i].Get()

			p.statsMu.Lock()
			p.stats.Hits++
			p.statsMu.Unlock()

			return slice
		}
	}

	// No suitable bucket, create new slice
	p.statsMu.Lock()
	p.stats.Misses++
	p.stats.Created++
	p.statsMu.Unlock()

	s := make([]T, 0, minCap)
	return &s
}

// Put returns a slice to the appropriate pool
func (p *SlicePool[T]) Put(slice *[]T) {
	if slice == nil {
		return
	}

	cap := cap(*slice)

	// Find appropriate bucket
	for i, capacity := range p.buckets {
		if capacity >= cap {
			p.pools[i].Put(slice)

			p.statsMu.Lock()
			p.stats.Puts++
			p.statsMu.Unlock()

			return
		}
	}

	// Slice too large for any bucket, let GC handle it
}

// Stats returns aggregated statistics from all buckets
func (p *SlicePool[T]) Stats() PoolStats {
	p.statsMu.RLock()
	stats := *p.stats
	p.statsMu.RUnlock()

	// Add stats from individual pools
	for _, pool := range p.pools {
		poolStats := pool.Stats()
		stats.Gets += poolStats.Gets
		stats.Puts += poolStats.Puts
		stats.Hits += poolStats.Hits
		stats.Misses += poolStats.Misses
		stats.Created += poolStats.Created
	}

	return stats
}

// Clear empties all pools
func (p *SlicePool[T]) Clear() {
	for _, pool := range p.pools {
		pool.Clear()
	}

	p.statsMu.Lock()
	p.stats = &PoolStats{}
	p.statsMu.Unlock()
}

// StringInterner manages string deduplication
type StringInterner struct {
	strings sync.Map // map[string]string for interned strings
	stats   *PoolStats
	statsMu sync.RWMutex
}

// NewStringInterner creates a new string interner
func NewStringInterner() *StringInterner {
	return &StringInterner{
		stats: &PoolStats{},
	}
}

// Intern returns the canonical instance of the string
func (si *StringInterner) Intern(s string) string {
	si.statsMu.Lock()
	si.stats.Gets++
	si.statsMu.Unlock()

	if existing, ok := si.strings.Load(s); ok {
		si.statsMu.Lock()
		si.stats.Hits++
		si.statsMu.Unlock()
		return existing.(string)
	}

	// Store the string
	si.strings.Store(s, s)

	si.statsMu.Lock()
	si.stats.Misses++
	si.stats.Created++
	si.statsMu.Unlock()

	return s
}

// Stats returns string interning statistics
func (si *StringInterner) Stats() PoolStats {
	si.statsMu.RLock()
	defer si.statsMu.RUnlock()
	return *si.stats
}

// Clear removes all interned strings
func (si *StringInterner) Clear() {
	si.strings = sync.Map{}

	si.statsMu.Lock()
	si.stats = &PoolStats{}
	si.statsMu.Unlock()
}

// Size estimates the number of interned strings
func (si *StringInterner) Size() int {
	count := 0
	si.strings.Range(func(key, value any) bool {
		count++
		return true
	})
	return count
}

// MemoryUsage estimates memory usage of interned strings
func (si *StringInterner) MemoryUsage() int64 {
	var total int64
	si.strings.Range(func(key, value any) bool {
		s := key.(string)
		total += int64(len(s)) + int64(unsafe.Sizeof(s))
		return true
	})
	return total
}
