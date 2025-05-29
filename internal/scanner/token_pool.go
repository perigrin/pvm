// ABOUTME: Scanner token pooling infrastructure following Microsoft TypeScript-Go patterns
// ABOUTME: Provides pooled allocation for tokens, iterators, and tokenization buffers

package scanner

import (
	"sync"
	"sync/atomic"

	"tamarou.com/pvm/internal/core"
	"tamarou.com/pvm/internal/memory"
)

// TokenPoolManager provides pooled allocation for scanner tokens and related structures
type TokenPoolManager struct {
	// Token object pools
	tokenPool    *memory.SyncPool[treeSitterToken]
	positionPool *memory.SyncPool[Position]

	// Iterator pools
	iteratorPool *memory.SyncPool[tokenIterator]

	// Buffer pools for batch operations
	tokenSlicePool  *memory.SlicePool[Token]
	tokenBufferPool *memory.SlicePool[*treeSitterToken]

	// String interning for token values
	stringInterner *memory.StringInterner

	// Statistics
	tokenCount    int64
	iteratorCount int64
	poolHits      int64
	poolMisses    int64
	memoryReused  int64

	mu sync.RWMutex
}

// TokenPoolHooks provides lifecycle hooks for debugging and monitoring
type TokenPoolHooks struct {
	OnCreateToken    func(token Token)        // Called when a token is created
	OnReuseToken     func(token Token)        // Called when a token is reused from pool
	OnCreateIterator func(iter TokenIterator) // Called when an iterator is created
	OnResetToken     func(token Token)        // Called when a token is reset for pooling
}

// NewTokenPoolManager creates a new token pool manager
func NewTokenPoolManager(hooks TokenPoolHooks) *TokenPoolManager {
	manager := &TokenPoolManager{
		stringInterner: memory.NewStringInterner(),
	}

	// Initialize token pool with reset function
	manager.tokenPool = memory.NewSyncPool(
		func() *treeSitterToken {
			atomic.AddInt64(&manager.tokenCount, 1)
			return &treeSitterToken{}
		},
		func(token *treeSitterToken) {
			// Reset token state for reuse
			token.tokenType = TokenEOF
			token.value = ""
			token.position = Position{}
			token.length = 0

			if hooks.OnResetToken != nil {
				hooks.OnResetToken(token)
			}
		},
	)

	// Initialize position pool
	manager.positionPool = memory.NewSyncPool(
		func() *Position {
			return &Position{}
		},
		func(pos *Position) {
			pos.Line = 0
			pos.Column = 0
			pos.Offset = 0
		},
	)

	// Initialize iterator pool
	manager.iteratorPool = memory.NewSyncPool(
		func() *tokenIterator {
			atomic.AddInt64(&manager.iteratorCount, 1)
			return &tokenIterator{}
		},
		func(iter *tokenIterator) {
			// Return token slices to pool if they came from pool
			if len(iter.tokens) > 0 {
				// Only put back if it was a pooled slice
				if slice := manager.tokenSlicePool.Get(len(iter.tokens)); slice != nil {
					manager.tokenSlicePool.Put(slice)
				}
			}
			iter.tokens = nil
			iter.pos = 0
		},
	)

	// Initialize slice pools with common token count buckets
	manager.tokenSlicePool = memory.NewSlicePool[Token]([]int{8, 16, 32, 64, 128, 256, 512, 1024, 2048})
	manager.tokenBufferPool = memory.NewSlicePool[*treeSitterToken]([]int{8, 16, 32, 64, 128, 256, 512, 1024, 2048})

	// Register with global pool manager for monitoring
	core.RegisterGlobalPool("scanner-tokens", manager)

	return manager
}

// CreateToken creates a new token with pooled allocation
func (tm *TokenPoolManager) CreateToken(tokenType TokenType, value string, pos Position, length int) Token {
	token := tm.tokenPool.Get()

	token.tokenType = tokenType
	token.value = tm.stringInterner.Intern(value) // Intern string values to reduce memory
	token.position = pos
	token.length = length

	atomic.AddInt64(&tm.poolHits, 1)
	atomic.AddInt64(&tm.memoryReused, 1)
	return token
}

// CreateTokenDirect creates a token with direct allocation (for performance comparison)
func (tm *TokenPoolManager) CreateTokenDirect(tokenType TokenType, value string, pos Position, length int) Token {
	atomic.AddInt64(&tm.poolMisses, 1)
	return &treeSitterToken{
		tokenType: tokenType,
		value:     value,
		position:  pos,
		length:    length,
	}
}

// ReleaseToken returns a token to the pool
func (tm *TokenPoolManager) ReleaseToken(token Token) {
	// Only handle treeSitterToken instances (safe type check)
	if token == nil {
		return
	}
	if t, ok := token.(*treeSitterToken); ok { //nolint:gocritic
		tm.tokenPool.Put(t)
	}
}

// CreatePosition creates a new position with pooled allocation
func (tm *TokenPoolManager) CreatePosition(line, column, offset int) Position {
	pos := tm.positionPool.Get()
	pos.Line = line
	pos.Column = column
	pos.Offset = offset
	return *pos
}

// ReleasePosition returns a position to the pool
func (tm *TokenPoolManager) ReleasePosition(pos *Position) {
	if pos != nil {
		tm.positionPool.Put(pos)
	}
}

// CreateTokenIterator creates a new token iterator with pooled allocation
func (tm *TokenPoolManager) CreateTokenIterator(tokens []Token) TokenIterator {
	iter := tm.iteratorPool.Get()
	iter.tokens = tokens
	iter.pos = 0

	return iter
}

// CreateTokenIteratorPooled creates a token iterator with a pooled token slice
func (tm *TokenPoolManager) CreateTokenIteratorPooled(tokenCount int) (*tokenIterator, *[]Token) {
	iter := tm.iteratorPool.Get()
	tokenSlice := tm.tokenSlicePool.Get(tokenCount)

	// Clear the slice
	*tokenSlice = (*tokenSlice)[:0]

	iter.tokens = nil // Will be set once tokens are added
	iter.pos = 0

	return iter, tokenSlice
}

// ReleaseTokenIterator returns an iterator to the pool
func (tm *TokenPoolManager) ReleaseTokenIterator(iter TokenIterator) {
	// Only handle tokenIterator instances (safe type check)
	if iter == nil {
		return
	}
	if ti, ok := iter.(*tokenIterator); ok { //nolint:gocritic
		tm.iteratorPool.Put(ti)
	}
}

// GetTokenBuffer gets a pooled buffer for batch token operations
func (tm *TokenPoolManager) GetTokenBuffer(minCapacity int) *[]*treeSitterToken {
	return tm.tokenBufferPool.Get(minCapacity)
}

// ReleaseTokenBuffer returns a token buffer to the pool
func (tm *TokenPoolManager) ReleaseTokenBuffer(buffer *[]*treeSitterToken) {
	tm.tokenBufferPool.Put(buffer)
}

// GetTokenSlice gets a pooled slice for storing tokens
func (tm *TokenPoolManager) GetTokenSlice(minCapacity int) *[]Token {
	return tm.tokenSlicePool.Get(minCapacity)
}

// ReleaseTokenSlice returns a token slice to the pool
func (tm *TokenPoolManager) ReleaseTokenSlice(slice *[]Token) {
	tm.tokenSlicePool.Put(slice)
}

// InternString interns a string value to reduce memory usage
func (tm *TokenPoolManager) InternString(s string) string {
	return tm.stringInterner.Intern(s)
}

// Stats returns pool allocation statistics
func (tm *TokenPoolManager) Stats() core.PoolStats {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	tokenStats := tm.tokenPool.Stats()
	iteratorStats := tm.iteratorPool.Stats()
	sliceStats := tm.tokenSlicePool.Stats()
	stringStats := tm.stringInterner.Stats()

	return core.PoolStats{
		Allocations: atomic.LoadInt64(&tm.tokenCount) + atomic.LoadInt64(&tm.iteratorCount),
		Grows:       int64(tokenStats.Gets + iteratorStats.Gets + sliceStats.Gets),
		TotalSize:   int64(tokenStats.Created + iteratorStats.Created + sliceStats.Created + stringStats.Created),
		CurrentSize: atomic.LoadInt64(&tm.poolHits),
		Capacity:    atomic.LoadInt64(&tm.poolMisses),
	}
}

// Reset clears all pools for reuse
func (tm *TokenPoolManager) Reset() {
	tm.tokenPool.Clear()
	tm.positionPool.Clear()
	tm.iteratorPool.Clear()
	tm.tokenSlicePool.Clear()
	tm.tokenBufferPool.Clear()
	tm.stringInterner.Clear()

	tm.mu.Lock()
	atomic.StoreInt64(&tm.tokenCount, 0)
	atomic.StoreInt64(&tm.iteratorCount, 0)
	atomic.StoreInt64(&tm.poolHits, 0)
	atomic.StoreInt64(&tm.poolMisses, 0)
	atomic.StoreInt64(&tm.memoryReused, 0)
	tm.mu.Unlock()
}

// MemoryUsage estimates total memory usage of the token pools
func (tm *TokenPoolManager) MemoryUsage() int64 {
	return tm.stringInterner.MemoryUsage()
}

// Global token pool manager instance
var globalTokenPoolManager *TokenPoolManager
var tokenPoolOnce sync.Once

// GetGlobalTokenPoolManager returns the global token pool manager
func GetGlobalTokenPoolManager() *TokenPoolManager {
	tokenPoolOnce.Do(func() {
		globalTokenPoolManager = NewTokenPoolManager(TokenPoolHooks{})
	})
	return globalTokenPoolManager
}

// SetGlobalTokenPoolManager sets a custom global token pool manager
func SetGlobalTokenPoolManager(manager *TokenPoolManager) {
	globalTokenPoolManager = manager
}
