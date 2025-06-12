// ABOUTME: Parser pooling system for efficient reuse of parser instances
// ABOUTME: Provides thread-safe pooling of tree-sitter parsers to avoid segfaults and improve performance

package parser

import (
	"sync"
	"sync/atomic"

	"tamarou.com/pvm/internal/memory"
)

// ParserPool manages a pool of parser instances for concurrent use
type ParserPool struct {
	pool    sync.Pool
	created int64
	gets    int64
	puts    int64
	hits    int64
	misses  int64
}

// GlobalParserPool is the global parser pool instance
var GlobalParserPool = NewParserPool()

// NewParserPool creates a new parser pool
func NewParserPool() *ParserPool {
	pp := &ParserPool{}
	pp.pool.New = func() interface{} {
		atomic.AddInt64(&pp.created, 1)
		atomic.AddInt64(&pp.misses, 1)

		// Create a new parser directly (without using the pool to avoid recursion)
		parser, err := NewParser()
		if err != nil {
			return nil // This will cause Get() to return nil and caller should handle
		}

		return parser
	}

	// Register with global memory stats
	memory.RegisterPool(pp.asAnyPool())

	return pp
}

// Get retrieves a parser from the pool
func (pp *ParserPool) Get() Parser {
	atomic.AddInt64(&pp.gets, 1)

	item := pp.pool.Get()
	if item == nil {
		return nil
	}

	parser := item.(Parser)
	if parser != nil {
		atomic.AddInt64(&pp.hits, 1)
	}

	return parser
}

// Put returns a parser to the pool
func (pp *ParserPool) Put(parser Parser) {
	if parser == nil {
		return
	}

	// Reset parser state if it has a reset method
	// Note: Tree-sitter parsers don't need resetting as they're stateless
	if resettable, ok := parser.(interface{ Reset() }); ok {
		resettable.Reset()
	}

	atomic.AddInt64(&pp.puts, 1)
	pp.pool.Put(parser)
}

// Stats returns pool statistics
func (pp *ParserPool) Stats() memory.PoolStats {
	gets := atomic.LoadInt64(&pp.gets)
	puts := atomic.LoadInt64(&pp.puts)
	hits := atomic.LoadInt64(&pp.hits)
	misses := atomic.LoadInt64(&pp.misses)
	created := atomic.LoadInt64(&pp.created)

	return memory.PoolStats{
		Gets:    uint64(gets),
		Puts:    uint64(puts),
		Hits:    uint64(hits),
		Misses:  uint64(misses),
		Created: uint64(created),
		MaxSize: -1, // sync.Pool doesn't have a max size
		Current: -1, // sync.Pool doesn't expose current size
	}
}

// Clear clears the pool
func (pp *ParserPool) Clear() {
	// sync.Pool doesn't support clearing, but we can create a new one
	pp.pool = sync.Pool{
		New: pp.pool.New,
	}
}

// NewPooledParser creates a new parser using the global pool
func NewPooledParser() (Parser, error) {
	parser := GlobalParserPool.Get()
	if parser == nil {
		// Fallback to direct creation if pool fails
		return NewParser()
	}
	return parser, nil
}

// ReturnParser returns a parser to the global pool
func ReturnParser(parser Parser) {
	GlobalParserPool.Put(parser)
}

// PooledParserFunc wraps a function to automatically manage parser pooling
func PooledParserFunc[T any](fn func(Parser) (T, error)) (T, error) {
	var zero T

	parser, err := NewPooledParser()
	if err != nil {
		return zero, err
	}
	defer ReturnParser(parser)

	result, err := fn(parser)
	return result, err
}

// parserPoolAdapter adapts ParserPool to Pool[any] interface
type parserPoolAdapter struct {
	pool *ParserPool
}

func (pp *ParserPool) asAnyPool() memory.Pool[any] {
	return &parserPoolAdapter{pool: pp}
}

func (ppa *parserPoolAdapter) Get() *any {
	parser := ppa.pool.Get()
	if parser == nil {
		return nil
	}
	var result any = parser
	return &result
}

func (ppa *parserPoolAdapter) Put(item *any) {
	if item == nil {
		return
	}
	if parser, ok := (*item).(Parser); ok {
		ppa.pool.Put(parser)
	}
}

func (ppa *parserPoolAdapter) Stats() memory.PoolStats {
	return ppa.pool.Stats()
}

func (ppa *parserPoolAdapter) Clear() {
	ppa.pool.Clear()
}
