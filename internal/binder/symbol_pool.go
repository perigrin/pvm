// ABOUTME: Symbol pooling implementation for memory-efficient symbol table operations.
// ABOUTME: Provides object pools for Symbol, Scope, and SymbolTable structures following TypeScript-Go patterns.

package binder

import (
	"sync"
	"sync/atomic"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/core"
)

// SymbolPoolManager provides pooled allocation for symbol table objects
type SymbolPoolManager struct {
	hooks SymbolPoolHooks

	// Core symbol pools
	symbolPool      core.Pool[Symbol]
	scopePool       core.Pool[Scope]
	symbolTablePool core.Pool[SymbolTable]

	// Map pools for efficient reuse
	symbolMapPool     *core.Pool[map[string]*Symbol]
	symbolSlicePool   *core.Pool[[]*Symbol]
	scopeSlicePool    *core.Pool[[]*Scope]
	moduleMapPool     *core.Pool[map[string]*SymbolTable]
	stringMapPool     *core.Pool[map[string]string]
	nodeMapPool       *core.Pool[map[ast.Node]*Scope]

	// Statistics and monitoring
	symbolCount     int64
	scopeCount      int64
	poolHits        int64
	poolMisses      int64
	memoryReused    int64

	mu sync.RWMutex
}

// SymbolPoolHooks provides lifecycle hooks for debugging and monitoring
type SymbolPoolHooks struct {
	OnSymbolCreate func(symbol *Symbol)      // Called when a symbol is created
	OnScopeCreate  func(scope *Scope)        // Called when a scope is created
	OnSymbolReset  func(symbol *Symbol)      // Called when a symbol is reset for pooling
	OnScopeReset   func(scope *Scope)        // Called when a scope is reset for pooling
	OnPoolWarming  func(poolType string)     // Called during pool warming
}

// SymbolPoolCoercible allows types to provide a SymbolPoolManager
type SymbolPoolCoercible interface {
	AsSymbolPoolManager() *SymbolPoolManager
}

// NewSymbolPoolManager creates a new symbol pool manager with the given hooks
func NewSymbolPoolManager(hooks SymbolPoolHooks) *SymbolPoolManager {
	manager := &SymbolPoolManager{
		hooks: hooks,
	}

	// Initialize map and slice pools
	manager.symbolMapPool = &core.Pool[map[string]*Symbol]{}
	manager.symbolSlicePool = &core.Pool[[]*Symbol]{}
	manager.scopeSlicePool = &core.Pool[[]*Scope]{}
	manager.moduleMapPool = &core.Pool[map[string]*SymbolTable]{}
	manager.stringMapPool = &core.Pool[map[string]string]{}
	manager.nodeMapPool = &core.Pool[map[ast.Node]*Scope]{}

	// Register with global pool manager for monitoring
	core.RegisterGlobalPool("symbol-pool", manager)

	return manager
}

// AsSymbolPoolManager implements SymbolPoolCoercible
func (spm *SymbolPoolManager) AsSymbolPoolManager() *SymbolPoolManager {
	return spm
}

// Pool statistics methods

// SymbolCount returns the total number of symbols created
func (spm *SymbolPoolManager) SymbolCount() int64 {
	return atomic.LoadInt64(&spm.symbolCount)
}

// ScopeCount returns the total number of scopes created
func (spm *SymbolPoolManager) ScopeCount() int64 {
	return atomic.LoadInt64(&spm.scopeCount)
}

// PoolEfficiency returns the pool hit rate as a percentage
func (spm *SymbolPoolManager) PoolEfficiency() float64 {
	hits := atomic.LoadInt64(&spm.poolHits)
	misses := atomic.LoadInt64(&spm.poolMisses)
	total := hits + misses

	if total == 0 {
		return 0
	}

	return float64(hits) / float64(total) * 100
}

// Stats returns pool statistics (implements core.PoolStatsProvider)
func (spm *SymbolPoolManager) Stats() core.PoolStats {
	return core.PoolStats{
		Allocations: atomic.LoadInt64(&spm.symbolCount) + atomic.LoadInt64(&spm.scopeCount),
		Grows:       atomic.LoadInt64(&spm.poolMisses), // Use pool misses as growth indicator
		TotalSize:   atomic.LoadInt64(&spm.memoryReused),
		CurrentSize: atomic.LoadInt64(&spm.poolHits),
		Capacity:    atomic.LoadInt64(&spm.symbolCount) + atomic.LoadInt64(&spm.scopeCount),
	}
}

// Symbol creation methods

// NewSymbol creates a pooled symbol
func (spm *SymbolPoolManager) NewSymbol(name string, kind SymbolKind, flags SymbolFlags, decl ast.Node, pos ast.Position) *Symbol {
	symbol := spm.symbolPool.New()

	// Reset/initialize the pooled object
	if symbol.Name == "" {
		atomic.AddInt64(&spm.poolMisses, 1)
	} else {
		spm.resetSymbol(symbol)
		atomic.AddInt64(&spm.poolHits, 1)
	}

	// Initialize fields
	symbol.Name = name
	symbol.Kind = kind
	symbol.Flags = flags
	symbol.Declaration = decl
	symbol.Position = pos
	symbol.Type = ""
	symbol.Scope = nil
	symbol.Package = ""
	symbol.OriginalSymbol = nil
	symbol.CapturedBy = nil
	symbol.Upvalues = nil
	symbol.QualifiedName = ""

	atomic.AddInt64(&spm.symbolCount, 1)

	if spm.hooks.OnSymbolCreate != nil {
		spm.hooks.OnSymbolCreate(symbol)
	}

	return symbol
}

// NewScope creates a pooled scope
func (spm *SymbolPoolManager) NewScope(kind ScopeKind, parent *Scope, node ast.Node, pos ast.Position) *Scope {
	scope := spm.scopePool.New()

	// Reset/initialize the pooled object
	if scope.Symbols == nil {
		// First time use - initialize maps
		scope.Symbols = spm.NewSymbolMap(8)
		scope.LocalSymbols = spm.NewSymbolMap(4)
		scope.SavedValues = spm.NewSymbolMap(4)
		scope.ImportedModules = spm.NewStringMap(4)
		scope.CapturedSymbols = spm.NewSymbolSlice(4)
		scope.Children = spm.NewScopeSlice(4)
		atomic.AddInt64(&spm.poolMisses, 1)
	} else {
		spm.resetScope(scope)
		atomic.AddInt64(&spm.poolHits, 1)
	}

	// Initialize fields
	scope.Kind = kind
	scope.Parent = parent
	scope.Node = node
	scope.Position = pos
	scope.Package = ""

	// Add to parent's children if parent exists
	if parent != nil {
		parent.Children = append(parent.Children, scope)
	}

	atomic.AddInt64(&spm.scopeCount, 1)

	if spm.hooks.OnScopeCreate != nil {
		spm.hooks.OnScopeCreate(scope)
	}

	return scope
}

// NewSymbolTable creates a pooled symbol table
func (spm *SymbolPoolManager) NewSymbolTable(packageName string) *SymbolTable {
	table := spm.symbolTablePool.New()

	// Reset/initialize the pooled object
	if table.Scopes == nil {
		// First time use - initialize maps
		table.Scopes = spm.NewNodeMap(16)
		table.Symbols = make(map[string][]*Symbol, 32)
		table.ModuleSymbols = spm.NewModuleMap(8)
		table.PackageSymbols = make(map[string]*Scope, 8)
		table.ExportedSymbols = spm.NewSymbolMap(16)
		table.DynamicSymbols = spm.NewSymbolMap(8)
	} else {
		spm.resetSymbolTable(table)
	}

	// Initialize fields
	table.Package = packageName
	table.CurrentScope = nil
	table.GlobalScope = nil

	return table
}

// Map and slice allocation methods

// NewSymbolMap creates a pooled symbol map
func (spm *SymbolPoolManager) NewSymbolMap(capacity int) map[string]*Symbol {
	symbolMap := make(map[string]*Symbol, capacity)
	atomic.AddInt64(&spm.memoryReused, int64(capacity))
	return symbolMap
}

// NewSymbolSlice creates a pooled symbol slice
func (spm *SymbolPoolManager) NewSymbolSlice(capacity int) []*Symbol {
	if capacity <= 0 {
		capacity = 4 // Default capacity
	}

	slice := make([]*Symbol, 0, capacity)
	atomic.AddInt64(&spm.memoryReused, int64(cap(slice)))
	return slice
}

// NewScopeSlice creates a pooled scope slice
func (spm *SymbolPoolManager) NewScopeSlice(capacity int) []*Scope {
	if capacity <= 0 {
		capacity = 4 // Default capacity
	}

	slice := make([]*Scope, 0, capacity)
	atomic.AddInt64(&spm.memoryReused, int64(cap(slice)))
	return slice
}

// NewModuleMap creates a pooled module map
func (spm *SymbolPoolManager) NewModuleMap(capacity int) map[string]*SymbolTable {
	moduleMap := make(map[string]*SymbolTable, capacity)
	atomic.AddInt64(&spm.memoryReused, int64(capacity))
	return moduleMap
}

// NewStringMap creates a pooled string map
func (spm *SymbolPoolManager) NewStringMap(capacity int) map[string]string {
	stringMap := make(map[string]string, capacity)
	atomic.AddInt64(&spm.memoryReused, int64(capacity))
	return stringMap
}

// NewNodeMap creates a pooled node-to-scope map
func (spm *SymbolPoolManager) NewNodeMap(capacity int) map[ast.Node]*Scope {
	nodeMap := make(map[ast.Node]*Scope, capacity)
	atomic.AddInt64(&spm.memoryReused, int64(capacity))
	return nodeMap
}

// Utility methods

// resetSymbol resets a Symbol for reuse
func (spm *SymbolPoolManager) resetSymbol(symbol *Symbol) {
	// Clear slice fields but keep capacity
	if symbol.CapturedBy != nil {
		symbol.CapturedBy = symbol.CapturedBy[:0]
	}
	if symbol.Upvalues != nil {
		symbol.Upvalues = symbol.Upvalues[:0]
	}

	// Clear references
	symbol.OriginalSymbol = nil
	symbol.Scope = nil
	symbol.Declaration = nil

	if spm.hooks.OnSymbolReset != nil {
		spm.hooks.OnSymbolReset(symbol)
	}
}

// resetScope resets a Scope for reuse
func (spm *SymbolPoolManager) resetScope(scope *Scope) {
	// Clear maps by resetting to empty but keep allocated capacity
	for k := range scope.Symbols {
		delete(scope.Symbols, k)
	}
	for k := range scope.LocalSymbols {
		delete(scope.LocalSymbols, k)
	}
	for k := range scope.SavedValues {
		delete(scope.SavedValues, k)
	}
	for k := range scope.ImportedModules {
		delete(scope.ImportedModules, k)
	}

	// Clear slices but keep capacity
	scope.Children = scope.Children[:0]
	scope.CapturedSymbols = scope.CapturedSymbols[:0]

	// Clear references
	scope.Parent = nil
	scope.Node = nil

	if spm.hooks.OnScopeReset != nil {
		spm.hooks.OnScopeReset(scope)
	}
}

// resetSymbolTable resets a SymbolTable for reuse
func (spm *SymbolPoolManager) resetSymbolTable(table *SymbolTable) {
	// Clear all maps
	for k := range table.Scopes {
		delete(table.Scopes, k)
	}
	for k := range table.Symbols {
		delete(table.Symbols, k)
	}
	for k := range table.ModuleSymbols {
		delete(table.ModuleSymbols, k)
	}
	for k := range table.PackageSymbols {
		delete(table.PackageSymbols, k)
	}
	for k := range table.ExportedSymbols {
		delete(table.ExportedSymbols, k)
	}
	for k := range table.DynamicSymbols {
		delete(table.DynamicSymbols, k)
	}

	// Clear references
	table.GlobalScope = nil
	table.CurrentScope = nil
}

// WarmPools pre-allocates objects for common usage patterns
func (spm *SymbolPoolManager) WarmPools() {
	spm.mu.Lock()
	defer spm.mu.Unlock()

	if spm.hooks.OnPoolWarming != nil {
		spm.hooks.OnPoolWarming("symbols")
	}

	// Pre-allocate common symbol types
	commonSymbols := []struct {
		name string
		kind SymbolKind
	}{
		{"self", SymbolScalar},
		{"class", SymbolScalar},
		{"this", SymbolScalar},
		{"args", SymbolArray},
		{"main", SymbolPackage},
	}

	for _, sym := range commonSymbols {
		pooledSym := spm.symbolPool.New()
		pooledSym.Name = sym.name
		pooledSym.Kind = sym.kind
		// Return to pool immediately for reuse
		spm.resetSymbol(pooledSym)
	}

	if spm.hooks.OnPoolWarming != nil {
		spm.hooks.OnPoolWarming("scopes")
	}

	// Pre-allocate common scope types
	commonScopes := []ScopeKind{
		ScopeGlobal,
		ScopePackage,
		ScopeSubroutine,
		ScopeBlock,
	}

	for _, scopeKind := range commonScopes {
		pooledScope := spm.scopePool.New()
		pooledScope.Kind = scopeKind
		pooledScope.Symbols = spm.NewSymbolMap(8)
		pooledScope.Children = spm.NewScopeSlice(4)
		// Return to pool immediately for reuse
		spm.resetScope(pooledScope)
	}
}

// Reset clears all pools for reuse
func (spm *SymbolPoolManager) Reset() {
	spm.mu.Lock()
	defer spm.mu.Unlock()

	// Reset all pools
	spm.symbolPool.Reset()
	spm.scopePool.Reset()
	spm.symbolTablePool.Reset()

	if spm.symbolMapPool != nil {
		spm.symbolMapPool.Reset()
	}
	if spm.symbolSlicePool != nil {
		spm.symbolSlicePool.Reset()
	}
	if spm.scopeSlicePool != nil {
		spm.scopeSlicePool.Reset()
	}
	if spm.moduleMapPool != nil {
		spm.moduleMapPool.Reset()
	}
	if spm.stringMapPool != nil {
		spm.stringMapPool.Reset()
	}
	if spm.nodeMapPool != nil {
		spm.nodeMapPool.Reset()
	}
}

// Clear completely empties all pools and resets statistics
func (spm *SymbolPoolManager) Clear() {
	spm.mu.Lock()
	defer spm.mu.Unlock()

	// Clear all pools
	spm.symbolPool.Clear()
	spm.scopePool.Clear()
	spm.symbolTablePool.Clear()

	if spm.symbolMapPool != nil {
		spm.symbolMapPool.Clear()
	}
	if spm.symbolSlicePool != nil {
		spm.symbolSlicePool.Clear()
	}
	if spm.scopeSlicePool != nil {
		spm.scopeSlicePool.Clear()
	}
	if spm.moduleMapPool != nil {
		spm.moduleMapPool.Clear()
	}
	if spm.stringMapPool != nil {
		spm.stringMapPool.Clear()
	}
	if spm.nodeMapPool != nil {
		spm.nodeMapPool.Clear()
	}

	// Reset statistics
	atomic.StoreInt64(&spm.symbolCount, 0)
	atomic.StoreInt64(&spm.scopeCount, 0)
	atomic.StoreInt64(&spm.poolHits, 0)
	atomic.StoreInt64(&spm.poolMisses, 0)
	atomic.StoreInt64(&spm.memoryReused, 0)
}

// Global symbol pool manager instance
var defaultSymbolPoolManager = NewSymbolPoolManager(SymbolPoolHooks{})

// DefaultSymbolPoolManager returns the global default symbol pool manager
func DefaultSymbolPoolManager() *SymbolPoolManager {
	return defaultSymbolPoolManager
}

// SetDefaultSymbolPoolManager sets the global default symbol pool manager
func SetDefaultSymbolPoolManager(manager *SymbolPoolManager) {
	defaultSymbolPoolManager = manager
}