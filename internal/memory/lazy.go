// ABOUTME: Lazy loading system for large data structures in PVM
// ABOUTME: Provides on-demand loading with caching and invalidation support

package memory

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// LazyLoader defines the interface for lazy loading implementations
type LazyLoader[T any] interface {
	Load(ctx context.Context) (T, error)
	Invalidate()
	IsLoaded() bool
	GetCached() (T, bool)
}

// LoadFunc represents a function that loads data
type LoadFunc[T any] func(ctx context.Context) (T, error)

// ValidateFunc represents a function that checks if cached data is still valid
type ValidateFunc[T any] func(data T) bool

// LazyValue holds a lazily loaded value with caching
type LazyValue[T any] struct {
	mu           sync.RWMutex
	loaded       bool
	loading      bool
	value        T
	err          error
	loadFunc     LoadFunc[T]
	validateFunc ValidateFunc[T]
	lastLoad     time.Time
	ttl          time.Duration
	loadCond     *sync.Cond
}

// NewLazyValue creates a new lazy value with the given load function
func NewLazyValue[T any](loadFunc LoadFunc[T]) *LazyValue[T] {
	lv := &LazyValue[T]{
		loadFunc: loadFunc,
	}
	lv.loadCond = sync.NewCond(&lv.mu)
	return lv
}

// WithTTL sets a time-to-live for the cached value
func (lv *LazyValue[T]) WithTTL(ttl time.Duration) *LazyValue[T] {
	lv.mu.Lock()
	defer lv.mu.Unlock()
	lv.ttl = ttl
	return lv
}

// WithValidator sets a validation function for the cached value
func (lv *LazyValue[T]) WithValidator(validateFunc ValidateFunc[T]) *LazyValue[T] {
	lv.mu.Lock()
	defer lv.mu.Unlock()
	lv.validateFunc = validateFunc
	return lv
}

// Load retrieves the value, loading it if necessary
func (lv *LazyValue[T]) Load(ctx context.Context) (T, error) {
	lv.mu.RLock()

	// Check if we need to reload
	needsReload := !lv.loaded ||
		(lv.ttl > 0 && time.Since(lv.lastLoad) > lv.ttl) ||
		(lv.validateFunc != nil && !lv.validateFunc(lv.value))

	if !needsReload {
		value, err := lv.value, lv.err
		lv.mu.RUnlock()
		return value, err
	}

	// Check if already loading
	if lv.loading {
		lv.mu.RUnlock()

		// Wait for loading to complete
		lv.mu.Lock()
		for lv.loading {
			select {
			case <-ctx.Done():
				lv.mu.Unlock()
				var zero T
				return zero, ctx.Err()
			default:
				lv.loadCond.Wait()
			}
		}
		value, err := lv.value, lv.err
		lv.mu.Unlock()
		return value, err
	}

	lv.mu.RUnlock()

	// Acquire write lock and load
	lv.mu.Lock()
	defer lv.mu.Unlock()

	// Double-check after acquiring write lock
	if lv.loaded &&
		(lv.ttl == 0 || time.Since(lv.lastLoad) <= lv.ttl) &&
		(lv.validateFunc == nil || lv.validateFunc(lv.value)) {
		return lv.value, lv.err
	}

	lv.loading = true
	defer func() {
		lv.loading = false
		lv.loadCond.Broadcast()
	}()

	// Release lock during loading
	lv.mu.Unlock()
	value, err := lv.loadFunc(ctx)
	lv.mu.Lock()

	lv.value = value
	lv.err = err
	lv.loaded = true
	lv.lastLoad = time.Now()

	return value, err
}

// Invalidate marks the cached value as invalid
func (lv *LazyValue[T]) Invalidate() {
	lv.mu.Lock()
	defer lv.mu.Unlock()

	lv.loaded = false
	var zero T
	lv.value = zero
	lv.err = nil
}

// IsLoaded returns true if the value has been loaded
func (lv *LazyValue[T]) IsLoaded() bool {
	lv.mu.RLock()
	defer lv.mu.RUnlock()
	return lv.loaded
}

// GetCached returns the cached value without loading
func (lv *LazyValue[T]) GetCached() (T, bool) {
	lv.mu.RLock()
	defer lv.mu.RUnlock()

	if !lv.loaded {
		var zero T
		return zero, false
	}

	return lv.value, true
}

// LazyMap provides lazy loading for map-like data structures
type LazyMap[K comparable, V any] struct {
	mu       sync.RWMutex
	items    map[K]*LazyValue[V]
	loadFunc func(ctx context.Context, key K) (V, error)
	ttl      time.Duration
}

// NewLazyMap creates a new lazy map
func NewLazyMap[K comparable, V any](loadFunc func(ctx context.Context, key K) (V, error)) *LazyMap[K, V] {
	return &LazyMap[K, V]{
		items:    make(map[K]*LazyValue[V]),
		loadFunc: loadFunc,
	}
}

// WithTTL sets a time-to-live for cached values
func (lm *LazyMap[K, V]) WithTTL(ttl time.Duration) *LazyMap[K, V] {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	lm.ttl = ttl
	return lm
}

// Get retrieves a value by key, loading it if necessary
func (lm *LazyMap[K, V]) Get(ctx context.Context, key K) (V, error) {
	lm.mu.RLock()
	item, exists := lm.items[key]
	lm.mu.RUnlock()

	if !exists {
		lm.mu.Lock()
		// Double-check after acquiring write lock
		if item, exists = lm.items[key]; !exists {
			item = NewLazyValue(func(ctx context.Context) (V, error) {
				return lm.loadFunc(ctx, key)
			})
			if lm.ttl > 0 {
				item.WithTTL(lm.ttl)
			}
			lm.items[key] = item
		}
		lm.mu.Unlock()
	}

	return item.Load(ctx)
}

// Invalidate removes a cached value
func (lm *LazyMap[K, V]) Invalidate(key K) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	if item, exists := lm.items[key]; exists {
		item.Invalidate()
		delete(lm.items, key)
	}
}

// InvalidateAll removes all cached values
func (lm *LazyMap[K, V]) InvalidateAll() {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	for _, item := range lm.items {
		item.Invalidate()
	}
	lm.items = make(map[K]*LazyValue[V])
}

// GetCached returns a cached value without loading
func (lm *LazyMap[K, V]) GetCached(key K) (V, bool) {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	if item, exists := lm.items[key]; exists {
		return item.GetCached()
	}

	var zero V
	return zero, false
}

// Keys returns all currently cached keys
func (lm *LazyMap[K, V]) Keys() []K {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	keys := make([]K, 0, len(lm.items))
	for key := range lm.items {
		keys = append(keys, key)
	}
	return keys
}

// Size returns the number of cached items
func (lm *LazyMap[K, V]) Size() int {
	lm.mu.RLock()
	defer lm.mu.RUnlock()
	return len(lm.items)
}

// LazySlice provides lazy loading for slice-like data structures
type LazySlice[T any] struct {
	mu       sync.RWMutex
	items    []*LazyValue[T]
	loadFunc func(ctx context.Context, index int) (T, error)
	size     int
	ttl      time.Duration
}

// NewLazySlice creates a new lazy slice with the given size
func NewLazySlice[T any](size int, loadFunc func(ctx context.Context, index int) (T, error)) *LazySlice[T] {
	return &LazySlice[T]{
		items:    make([]*LazyValue[T], size),
		loadFunc: loadFunc,
		size:     size,
	}
}

// WithTTL sets a time-to-live for cached values
func (ls *LazySlice[T]) WithTTL(ttl time.Duration) *LazySlice[T] {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	ls.ttl = ttl
	return ls
}

// Get retrieves a value by index, loading it if necessary
func (ls *LazySlice[T]) Get(ctx context.Context, index int) (T, error) {
	if index < 0 || index >= ls.size {
		var zero T
		return zero, ErrIndexOutOfRange
	}

	ls.mu.RLock()
	item := ls.items[index]
	ls.mu.RUnlock()

	if item == nil {
		ls.mu.Lock()
		// Double-check after acquiring write lock
		if ls.items[index] == nil {
			item = NewLazyValue(func(ctx context.Context) (T, error) {
				return ls.loadFunc(ctx, index)
			})
			if ls.ttl > 0 {
				item.WithTTL(ls.ttl)
			}
			ls.items[index] = item
		} else {
			item = ls.items[index]
		}
		ls.mu.Unlock()
	}

	return item.Load(ctx)
}

// Invalidate removes a cached value at the given index
func (ls *LazySlice[T]) Invalidate(index int) {
	if index < 0 || index >= ls.size {
		return
	}

	ls.mu.Lock()
	defer ls.mu.Unlock()

	if ls.items[index] != nil {
		ls.items[index].Invalidate()
		ls.items[index] = nil
	}
}

// Size returns the size of the slice
func (ls *LazySlice[T]) Size() int {
	return ls.size
}

// LoadedCount returns the number of loaded items
func (ls *LazySlice[T]) LoadedCount() int {
	ls.mu.RLock()
	defer ls.mu.RUnlock()

	count := 0
	for _, item := range ls.items {
		if item != nil && item.IsLoaded() {
			count++
		}
	}
	return count
}

// Common errors
var (
	ErrIndexOutOfRange = fmt.Errorf("index out of range")
)

// LazyInitializer provides one-time initialization with lazy loading
type LazyInitializer[T any] struct {
	once     sync.Once
	value    T
	err      error
	initFunc func() (T, error)
}

// NewLazyInitializer creates a new lazy initializer
func NewLazyInitializer[T any](initFunc func() (T, error)) *LazyInitializer[T] {
	return &LazyInitializer[T]{
		initFunc: initFunc,
	}
}

// Get returns the initialized value, initializing it if necessary
func (li *LazyInitializer[T]) Get() (T, error) {
	li.once.Do(func() {
		li.value, li.err = li.initFunc()
	})
	return li.value, li.err
}

// IsInitialized returns true if the value has been initialized
func (li *LazyInitializer[T]) IsInitialized() bool {
	// Check if once.Do has been called by checking if we have an error or value
	return li.err != nil || li.initFunc == nil
}
