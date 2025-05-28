// ABOUTME: Language service object pooling implementation for memory-efficient LS operations
// ABOUTME: Provides object pools for language service documents, caches, and analysis structures

package ls

import (
	"sync"
	"sync/atomic"
	"time"

	"tamarou.com/pvm/internal/core"
	"tamarou.com/pvm/internal/parser"
)

// LSPoolManager provides pooled allocation for Language Service objects
type LSPoolManager struct {
	hooks LSPoolHooks

	// Document and cache pools
	documentPool core.Pool[Document]
	// Note: DocumentCache, PerformanceMonitor, PerformanceStats already exist in the codebase

	// Analysis result pools
	analysisResultPool   core.Pool[AnalysisResult]
	completionResultPool core.Pool[CompletionResult]
	definitionResultPool core.Pool[DefinitionResult]
	diagnosticResultPool core.Pool[DiagnosticResult]
	hoverResultPool      core.Pool[HoverResult]

	// Operation context pools
	analysisContextPool core.Pool[AnalysisContext]
	requestContextPool  core.Pool[RequestContext]

	// Map pools for efficient reuse
	stringMapPool    *core.Pool[map[string]string]
	interfaceMapPool *core.Pool[map[string]interface{}]
	documentMapPool  *core.Pool[map[string]*Document]
	// Note: CacheEntry is generic, managed differently

	// Slice pools for collections
	stringSlicePool         *core.Pool[[]string]
	errorSlicePool          *core.Pool[[]error]
	typeCheckErrorSlicePool *core.Pool[[]parser.TypeCheckError]

	// Statistics and monitoring
	documentCount  int64
	analysisCount  int64
	cacheHits      int64
	cacheMisses    int64
	poolHits       int64
	poolMisses     int64
	memoryReused   int64
	objectsCreated int64
	objectsReused  int64

	mu sync.RWMutex
}

// LSPoolHooks provides lifecycle hooks for debugging and monitoring
type LSPoolHooks struct {
	OnDocumentCreate     func(doc *Document)                    // Called when a document is created
	OnAnalysisStart      func(uri string, analysisType string)  // Called when analysis starts
	OnAnalysisComplete   func(uri string, duration int64)       // Called when analysis completes
	OnCacheHit           func(uri, key string)                  // Called when cache hit occurs
	OnCacheMiss          func(uri, key string)                  // Called when cache miss occurs
	OnObjectCreate       func(objectType string)                // Called when any LS object is created
	OnObjectReset        func(objectType string)                // Called when an object is reset for pooling
	OnPoolWarming        func(poolType string)                  // Called during pool warming
	OnMemoryThreshold    func(usage int64)                      // Called when memory usage exceeds threshold
	OnPerformanceWarning func(operation string, duration int64) // Called when operation takes too long
}

// LSPoolCoercible allows types to provide an LSPoolManager
type LSPoolCoercible interface {
	AsLSPoolManager() *LSPoolManager
}

// NewLSPoolManager creates a new language service pool manager with the given hooks
func NewLSPoolManager(hooks LSPoolHooks) *LSPoolManager {
	manager := &LSPoolManager{
		hooks: hooks,
	}

	// Initialize map pools
	manager.stringMapPool = &core.Pool[map[string]string]{}
	manager.interfaceMapPool = &core.Pool[map[string]interface{}]{}
	manager.documentMapPool = &core.Pool[map[string]*Document]{}
	// Note: CacheEntry maps handled by existing cache system

	// Initialize slice pools
	manager.stringSlicePool = &core.Pool[[]string]{}
	manager.errorSlicePool = &core.Pool[[]error]{}
	manager.typeCheckErrorSlicePool = &core.Pool[[]parser.TypeCheckError]{}

	// Register with global pool manager for monitoring
	core.RegisterGlobalPool("ls-pool", manager)

	// Warm up pools with common LS objects
	manager.warmPools()

	return manager
}

// AsLSPoolManager implements LSPoolCoercible
func (lspm *LSPoolManager) AsLSPoolManager() *LSPoolManager {
	return lspm
}

// Stats returns pool allocation statistics
func (lspm *LSPoolManager) Stats() core.PoolStats {
	return core.PoolStats{
		Allocations: atomic.LoadInt64(&lspm.documentCount) + atomic.LoadInt64(&lspm.analysisCount),
		Grows:       atomic.LoadInt64(&lspm.poolHits),
		TotalSize:   atomic.LoadInt64(&lspm.poolMisses),
		CurrentSize: atomic.LoadInt64(&lspm.memoryReused),
		Capacity:    atomic.LoadInt64(&lspm.objectsCreated) + atomic.LoadInt64(&lspm.objectsReused),
	}
}

// Pool statistics methods

// DocumentCount returns the total number of documents managed
func (lspm *LSPoolManager) DocumentCount() int64 {
	return atomic.LoadInt64(&lspm.documentCount)
}

// AnalysisCount returns the total number of analysis operations
func (lspm *LSPoolManager) AnalysisCount() int64 {
	return atomic.LoadInt64(&lspm.analysisCount)
}

// CacheEfficiency returns the cache hit rate as a percentage
func (lspm *LSPoolManager) CacheEfficiency() float64 {
	hits := atomic.LoadInt64(&lspm.cacheHits)
	misses := atomic.LoadInt64(&lspm.cacheMisses)
	total := hits + misses

	if total == 0 {
		return 0
	}
	return float64(hits) / float64(total) * 100
}

// PoolEfficiency returns the pool hit rate as a percentage
func (lspm *LSPoolManager) PoolEfficiency() float64 {
	hits := atomic.LoadInt64(&lspm.poolHits)
	misses := atomic.LoadInt64(&lspm.poolMisses)
	total := hits + misses

	if total == 0 {
		return 0
	}
	return float64(hits) / float64(total) * 100
}

// ObjectReuseRate returns the object reuse rate as a percentage
func (lspm *LSPoolManager) ObjectReuseRate() float64 {
	created := atomic.LoadInt64(&lspm.objectsCreated)
	reused := atomic.LoadInt64(&lspm.objectsReused)
	total := created + reused

	if total == 0 {
		return 0
	}
	return float64(reused) / float64(total) * 100
}

// Document creation methods

// NewDocument creates a pooled document
func (lspm *LSPoolManager) NewDocument(uri, text string, version int) *Document {
	doc := lspm.documentPool.New()

	// Reset/initialize the pooled object
	lspm.resetDocument(doc)

	// Initialize fields
	doc.URI = uri
	doc.Text = text
	doc.Version = version
	doc.LastChanged = time.Now()
	doc.Errors = lspm.getTypeCheckErrorSlice()

	atomic.AddInt64(&lspm.documentCount, 1)
	atomic.AddInt64(&lspm.poolHits, 1)

	if lspm.hooks.OnDocumentCreate != nil {
		lspm.hooks.OnDocumentCreate(doc)
	}

	if lspm.hooks.OnObjectCreate != nil {
		lspm.hooks.OnObjectCreate("Document")
	}

	return doc
}

// Note: DocumentCache creation handled by existing cache system

// Note: CacheEntry creation handled by existing generic cache system

// Analysis result creation methods

// NewAnalysisResult creates a pooled analysis result
func (lspm *LSPoolManager) NewAnalysisResult(uri string) *AnalysisResult {
	result := lspm.analysisResultPool.New()

	// Reset/initialize the pooled object
	lspm.resetAnalysisResult(result)

	// Initialize fields
	result.URI = uri
	result.Errors = lspm.getErrorSlice()
	result.StartTime = time.Now()

	atomic.AddInt64(&lspm.analysisCount, 1)
	atomic.AddInt64(&lspm.poolHits, 1)

	if lspm.hooks.OnObjectCreate != nil {
		lspm.hooks.OnObjectCreate("AnalysisResult")
	}

	return result
}

// NewCompletionResult creates a pooled completion result
func (lspm *LSPoolManager) NewCompletionResult(uri string, position Position) *CompletionResult {
	result := lspm.completionResultPool.New()

	// Reset/initialize the pooled object
	lspm.resetCompletionResult(result)

	// Initialize fields
	result.URI = uri
	result.Position = position
	result.Items = lspm.getStringSlice()

	atomic.AddInt64(&lspm.poolHits, 1)

	if lspm.hooks.OnObjectCreate != nil {
		lspm.hooks.OnObjectCreate("CompletionResult")
	}

	return result
}

// Performance monitoring creation methods

// Note: PerformanceMonitor creation handled by existing performance system

// NewAnalysisContext creates a pooled analysis context
func (lspm *LSPoolManager) NewAnalysisContext(uri string) *AnalysisContext {
	ctx := lspm.analysisContextPool.New()

	// Reset/initialize the pooled object
	lspm.resetAnalysisContext(ctx)

	// Initialize fields
	ctx.URI = uri
	ctx.StartTime = time.Now()
	ctx.Metadata = lspm.getInterfaceMap()

	atomic.AddInt64(&lspm.poolHits, 1)

	if lspm.hooks.OnObjectCreate != nil {
		lspm.hooks.OnObjectCreate("AnalysisContext")
	}

	return ctx
}

// Helper methods for getting pooled collections

// getStringMap returns a pooled string map
func (lspm *LSPoolManager) getStringMap() map[string]string {
	m := lspm.stringMapPool.New()
	if *m == nil {
		*m = make(map[string]string)
		atomic.AddInt64(&lspm.poolMisses, 1)
	} else {
		lspm.clearStringMap(*m)
		atomic.AddInt64(&lspm.poolHits, 1)
	}
	return *m
}

// getInterfaceMap returns a pooled interface map
func (lspm *LSPoolManager) getInterfaceMap() map[string]interface{} {
	m := lspm.interfaceMapPool.New()
	if *m == nil {
		*m = make(map[string]interface{})
		atomic.AddInt64(&lspm.poolMisses, 1)
	} else {
		lspm.clearInterfaceMap(*m)
		atomic.AddInt64(&lspm.poolHits, 1)
	}
	return *m
}

// getDocumentMap returns a pooled document map
func (lspm *LSPoolManager) getDocumentMap() map[string]*Document {
	m := lspm.documentMapPool.New()
	if *m == nil {
		*m = make(map[string]*Document)
		atomic.AddInt64(&lspm.poolMisses, 1)
	} else {
		lspm.clearDocumentMap(*m)
		atomic.AddInt64(&lspm.poolHits, 1)
	}
	return *m
}

// Note: CacheEntry maps managed by existing generic cache system

// Slice pool methods

// getStringSlice returns a pooled string slice
func (lspm *LSPoolManager) getStringSlice() []string {
	s := lspm.stringSlicePool.New()
	if *s == nil {
		*s = make([]string, 0, 8)
		atomic.AddInt64(&lspm.poolMisses, 1)
	} else {
		*s = (*s)[:0] // Reset length but keep capacity
		atomic.AddInt64(&lspm.poolHits, 1)
	}
	return *s
}

// getErrorSlice returns a pooled error slice
func (lspm *LSPoolManager) getErrorSlice() []error {
	s := lspm.errorSlicePool.New()
	if *s == nil {
		*s = make([]error, 0, 4)
		atomic.AddInt64(&lspm.poolMisses, 1)
	} else {
		*s = (*s)[:0] // Reset length but keep capacity
		atomic.AddInt64(&lspm.poolHits, 1)
	}
	return *s
}

// getTypeCheckErrorSlice returns a pooled type check error slice
func (lspm *LSPoolManager) getTypeCheckErrorSlice() []parser.TypeCheckError {
	s := lspm.typeCheckErrorSlicePool.New()
	if *s == nil {
		*s = make([]parser.TypeCheckError, 0, 4)
		atomic.AddInt64(&lspm.poolMisses, 1)
	} else {
		*s = (*s)[:0] // Reset length but keep capacity
		atomic.AddInt64(&lspm.poolHits, 1)
	}
	return *s
}

// Pool warming methods

// warmPools pre-allocates common LS objects for better performance
func (lspm *LSPoolManager) warmPools() {
	if lspm.hooks.OnPoolWarming != nil {
		lspm.hooks.OnPoolWarming("language-service")
	}

	// Warm up pools with common LS patterns
	lspm.warmDocuments()
	// Note: Cache warming handled by existing cache system
	lspm.warmAnalysisResults()
	lspm.warmPerformanceMonitors()

	if lspm.hooks.OnPoolWarming != nil {
		lspm.hooks.OnPoolWarming("language-service-complete")
	}
}

// warmDocuments pre-allocates document structures
func (lspm *LSPoolManager) warmDocuments() {
	// Create a few documents to warm the pool
	for i := 0; i < 4; i++ {
		doc := lspm.NewDocument("", "", 0)
		lspm.resetDocument(doc)
	}
}

// warmCaches pre-allocates cache structures
func (lspm *LSPoolManager) warmCaches() {
	// Note: Cache warming handled by existing cache system
}

// warmAnalysisResults pre-allocates analysis result structures
func (lspm *LSPoolManager) warmAnalysisResults() {
	// Create analysis results to warm the pools
	for i := 0; i < 8; i++ {
		result := lspm.NewAnalysisResult("")
		lspm.resetAnalysisResult(result)

		completion := lspm.NewCompletionResult("", Position{})
		lspm.resetCompletionResult(completion)
	}
}

// warmPerformanceMonitors pre-allocates performance monitoring structures
func (lspm *LSPoolManager) warmPerformanceMonitors() {
	// Create performance monitors to warm the pools
	for i := 0; i < 2; i++ {
		// Note: Performance monitor warming handled by existing system

		ctx := lspm.NewAnalysisContext("")
		lspm.resetAnalysisContext(ctx)
	}
}

// Reset methods for proper object cleanup and reuse

// resetDocument resets a document for reuse
func (lspm *LSPoolManager) resetDocument(doc *Document) {
	doc.URI = ""
	doc.Text = ""
	doc.Version = 0
	doc.AST = nil
	doc.SymbolTable = nil
	doc.LastChecked = time.Time{}
	doc.LastChanged = time.Time{}
	doc.ContentHash = ""
	doc.ASTHash = ""
	doc.SymbolHash = ""

	// Reset slices but keep capacity
	if doc.Errors != nil {
		doc.Errors = doc.Errors[:0]
	}

	if lspm.hooks.OnObjectReset != nil {
		lspm.hooks.OnObjectReset("Document")
	}
}

// Note: DocumentCache reset handled by existing cache system

// Note: CacheEntry reset handled by existing generic cache system

// resetAnalysisResult resets an analysis result for reuse
func (lspm *LSPoolManager) resetAnalysisResult(result *AnalysisResult) {
	result.URI = ""
	result.StartTime = time.Time{}
	result.EndTime = time.Time{}

	// Reset slices but keep capacity
	if result.Errors != nil {
		result.Errors = result.Errors[:0]
	}

	if lspm.hooks.OnObjectReset != nil {
		lspm.hooks.OnObjectReset("AnalysisResult")
	}
}

// resetCompletionResult resets a completion result for reuse
func (lspm *LSPoolManager) resetCompletionResult(result *CompletionResult) {
	result.URI = ""
	result.Position = Position{}

	// Reset slices but keep capacity
	if result.Items != nil {
		result.Items = result.Items[:0]
	}

	if lspm.hooks.OnObjectReset != nil {
		lspm.hooks.OnObjectReset("CompletionResult")
	}
}

// Note: PerformanceMonitor reset handled by existing performance system

// resetAnalysisContext resets an analysis context for reuse
func (lspm *LSPoolManager) resetAnalysisContext(ctx *AnalysisContext) {
	ctx.URI = ""
	ctx.StartTime = time.Time{}

	// Clear maps
	lspm.clearInterfaceMap(ctx.Metadata)

	if lspm.hooks.OnObjectReset != nil {
		lspm.hooks.OnObjectReset("AnalysisContext")
	}
}

// Helper methods to clear maps efficiently

// clearStringMap clears a string map efficiently
func (lspm *LSPoolManager) clearStringMap(m map[string]string) {
	for k := range m {
		delete(m, k)
	}
}

// clearInterfaceMap clears an interface map efficiently
func (lspm *LSPoolManager) clearInterfaceMap(m map[string]interface{}) {
	for k := range m {
		delete(m, k)
	}
}

// clearDocumentMap clears a document map efficiently
func (lspm *LSPoolManager) clearDocumentMap(m map[string]*Document) {
	for k := range m {
		delete(m, k)
	}
}

// Note: CacheEntry map clearing handled by existing cache system

// Pool management methods

// Reset resets all pools for reuse
func (lspm *LSPoolManager) Reset() {
	lspm.mu.Lock()
	defer lspm.mu.Unlock()

	// Reset all core pools
	lspm.documentPool.Reset()
	lspm.analysisResultPool.Reset()
	lspm.completionResultPool.Reset()
	lspm.definitionResultPool.Reset()
	lspm.diagnosticResultPool.Reset()
	lspm.hoverResultPool.Reset()
	lspm.analysisContextPool.Reset()
	lspm.requestContextPool.Reset()
}

// Clear completely empties all pools and resets statistics
func (lspm *LSPoolManager) Clear() {
	lspm.mu.Lock()
	defer lspm.mu.Unlock()

	// Clear all core pools
	lspm.documentPool.Clear()
	lspm.analysisResultPool.Clear()
	lspm.completionResultPool.Clear()
	lspm.definitionResultPool.Clear()
	lspm.diagnosticResultPool.Clear()
	lspm.hoverResultPool.Clear()
	lspm.analysisContextPool.Clear()
	lspm.requestContextPool.Clear()

	// Reset statistics
	atomic.StoreInt64(&lspm.documentCount, 0)
	atomic.StoreInt64(&lspm.analysisCount, 0)
	atomic.StoreInt64(&lspm.cacheHits, 0)
	atomic.StoreInt64(&lspm.cacheMisses, 0)
	atomic.StoreInt64(&lspm.poolHits, 0)
	atomic.StoreInt64(&lspm.poolMisses, 0)
	atomic.StoreInt64(&lspm.memoryReused, 0)
	atomic.StoreInt64(&lspm.objectsCreated, 0)
	atomic.StoreInt64(&lspm.objectsReused, 0)
}

// GetDetailedStats returns detailed statistics about all LS pools
func (lspm *LSPoolManager) GetDetailedStats() LSPoolStats {
	return LSPoolStats{
		DocumentCount:   atomic.LoadInt64(&lspm.documentCount),
		AnalysisCount:   atomic.LoadInt64(&lspm.analysisCount),
		CacheHits:       atomic.LoadInt64(&lspm.cacheHits),
		CacheMisses:     atomic.LoadInt64(&lspm.cacheMisses),
		PoolHits:        atomic.LoadInt64(&lspm.poolHits),
		PoolMisses:      atomic.LoadInt64(&lspm.poolMisses),
		MemoryReused:    atomic.LoadInt64(&lspm.memoryReused),
		ObjectsCreated:  atomic.LoadInt64(&lspm.objectsCreated),
		ObjectsReused:   atomic.LoadInt64(&lspm.objectsReused),
		CacheEfficiency: lspm.CacheEfficiency(),
		PoolEfficiency:  lspm.PoolEfficiency(),
		ObjectReuseRate: lspm.ObjectReuseRate(),
	}
}

// LSPoolStats contains detailed LS pool statistics
type LSPoolStats struct {
	DocumentCount   int64   // Total number of documents managed
	AnalysisCount   int64   // Total number of analysis operations
	CacheHits       int64   // Number of successful cache lookups
	CacheMisses     int64   // Number of cache misses
	PoolHits        int64   // Number of successful pool allocations
	PoolMisses      int64   // Number of pool misses requiring new allocation
	MemoryReused    int64   // Total amount of memory reused
	ObjectsCreated  int64   // Number of objects created
	ObjectsReused   int64   // Number of objects reused from pool
	CacheEfficiency float64 // Cache hit rate as percentage
	PoolEfficiency  float64 // Pool hit rate as percentage
	ObjectReuseRate float64 // Object reuse rate as percentage
}

// Global LS pool manager instance
var globalLSPoolManager *LSPoolManager
var lsPoolOnce sync.Once

// GlobalLSPoolManager returns the global LS pool manager instance
func GlobalLSPoolManager() *LSPoolManager {
	lsPoolOnce.Do(func() {
		globalLSPoolManager = NewLSPoolManager(LSPoolHooks{
			// Default hooks can be set here
			OnPoolWarming: func(poolType string) {
				// Default pool warming notification
			},
		})
	})
	return globalLSPoolManager
}

// SetGlobalLSPoolHooks sets hooks for the global LS pool manager
func SetGlobalLSPoolHooks(hooks LSPoolHooks) {
	lspm := GlobalLSPoolManager()
	lspm.hooks = hooks
}
