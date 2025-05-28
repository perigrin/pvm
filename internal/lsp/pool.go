// ABOUTME: LSP object pooling implementation for memory-efficient LSP operations
// ABOUTME: Provides object pools for LSP protocol objects, completion items, and language service structures

package lsp

import (
	"sync"
	"sync/atomic"
	"time"

	"tamarou.com/pvm/internal/core"
)

// LSPPoolManager provides pooled allocation for LSP protocol objects
type LSPPoolManager struct {
	hooks LSPPoolHooks

	// JSON-RPC message pools
	jsonRPCMessagePool      core.Pool[JSONRPCMessage]
	jsonRPCResponsePool     core.Pool[JSONRPCResponse]
	jsonRPCErrorPool        core.Pool[JSONRPCError]
	jsonRPCNotificationPool core.Pool[JSONRPCNotification]

	// LSP basic type pools
	positionPool                        core.Pool[Position]
	rangePool                           core.Pool[Range]
	locationPool                        core.Pool[Location]
	textDocumentIdentifierPool          core.Pool[TextDocumentIdentifier]
	versionedTextDocumentIdentifierPool core.Pool[VersionedTextDocumentIdentifier]
	textDocumentItemPool                core.Pool[TextDocumentItem]
	textDocumentContentChangePool       core.Pool[TextDocumentContentChangeEvent]

	// Completion pools
	completionItemPool    core.Pool[CompletionItem]
	completionListPool    core.Pool[CompletionList]
	completionParamsPool  core.Pool[CompletionParams]
	completionContextPool core.Pool[CompletionContext]

	// Diagnostic pools
	diagnosticPool            core.Pool[Diagnostic]
	diagnosticRelatedInfoPool core.Pool[DiagnosticRelatedInfo]
	publishDiagnosticsPool    core.Pool[PublishDiagnosticsParams]

	// Definition and navigation pools
	definitionParamsPool           core.Pool[DefinitionParams]
	textDocumentPositionParamsPool core.Pool[TextDocumentPositionParams]
	hoverParamsPool                core.Pool[HoverParams]
	hoverPool                      core.Pool[Hover]
	markupContentPool              core.Pool[MarkupContent]

	// Initialize pools
	initializeParamsPool   core.Pool[InitializeParams]
	initializeResultPool   core.Pool[InitializeResult]
	serverCapabilitiesPool core.Pool[ServerCapabilities]
	clientInfoPool         core.Pool[ClientInfo]

	// Slice pools for collections
	completionItemSlicePool *core.Pool[[]CompletionItem]
	diagnosticSlicePool     *core.Pool[[]Diagnostic]
	locationSlicePool       *core.Pool[[]Location]
	textEditSlicePool       *core.Pool[[]TextEdit]
	workspaceEditSlicePool  *core.Pool[[]WorkspaceEdit]
	stringSlicePool         *core.Pool[[]string]

	// Map pools for efficient reuse
	stringMapPool    *core.Pool[map[string]string]
	interfaceMapPool *core.Pool[map[string]interface{}]

	// Request-scoped allocation tracking
	requestPools map[string]*RequestScopedPool
	requestMu    sync.RWMutex

	// Statistics and monitoring
	requestCount    int64
	completionCount int64
	diagnosticCount int64
	poolHits        int64
	poolMisses      int64
	memoryReused    int64
	objectsCreated  int64
	objectsReused   int64

	mu sync.RWMutex
}

// LSPPoolHooks provides lifecycle hooks for debugging and monitoring
type LSPPoolHooks struct {
	OnRequestStart       func(requestID string, method string)       // Called when a request starts
	OnRequestEnd         func(requestID string, duration int64)      // Called when a request ends
	OnCompletionCreate   func(item *CompletionItem)                  // Called when a completion item is created
	OnDiagnosticCreate   func(diagnostic *Diagnostic)                // Called when a diagnostic is created
	OnObjectCreate       func(objectType string)                     // Called when any LSP object is created
	OnObjectReset        func(objectType string)                     // Called when an object is reset for pooling
	OnPoolWarming        func(poolType string)                       // Called during pool warming
	OnMemoryThreshold    func(usage int64)                           // Called when memory usage exceeds threshold
	OnRequestScopeCreate func(requestID string, objectCount int)     // Called when request scope is created
	OnRequestScopeClean  func(requestID string, objectsReleased int) // Called when request scope is cleaned
}

// RequestScopedPool manages pools for a specific request
type RequestScopedPool struct {
	requestID        string
	allocatedObjects []interface{}
	startTime        int64
	mu               sync.RWMutex
}

// LSPPoolCoercible allows types to provide an LSPPoolManager
type LSPPoolCoercible interface {
	AsLSPPoolManager() *LSPPoolManager
}

// NewLSPPoolManager creates a new LSP pool manager with the given hooks
func NewLSPPoolManager(hooks LSPPoolHooks) *LSPPoolManager {
	manager := &LSPPoolManager{
		hooks:        hooks,
		requestPools: make(map[string]*RequestScopedPool),
	}

	// Initialize slice pools
	manager.completionItemSlicePool = &core.Pool[[]CompletionItem]{}
	manager.diagnosticSlicePool = &core.Pool[[]Diagnostic]{}
	manager.locationSlicePool = &core.Pool[[]Location]{}
	manager.textEditSlicePool = &core.Pool[[]TextEdit]{}
	manager.workspaceEditSlicePool = &core.Pool[[]WorkspaceEdit]{}
	manager.stringSlicePool = &core.Pool[[]string]{}

	// Initialize map pools
	manager.stringMapPool = &core.Pool[map[string]string]{}
	manager.interfaceMapPool = &core.Pool[map[string]interface{}]{}

	// Register with global pool manager for monitoring
	core.RegisterGlobalPool("lsp-pool", manager)

	// Warm up pools with common LSP objects
	manager.warmPools()

	return manager
}

// AsLSPPoolManager implements LSPPoolCoercible
func (lpm *LSPPoolManager) AsLSPPoolManager() *LSPPoolManager {
	return lpm
}

// Stats returns pool allocation statistics
func (lpm *LSPPoolManager) Stats() core.PoolStats {
	return core.PoolStats{
		Allocations: atomic.LoadInt64(&lpm.requestCount) + atomic.LoadInt64(&lpm.completionCount),
		Grows:       atomic.LoadInt64(&lpm.poolHits),
		TotalSize:   atomic.LoadInt64(&lpm.poolMisses),
		CurrentSize: atomic.LoadInt64(&lpm.memoryReused),
		Capacity:    atomic.LoadInt64(&lpm.objectsCreated) + atomic.LoadInt64(&lpm.objectsReused),
	}
}

// Request-scoped pooling methods

// StartRequest creates a new request-scoped pool
func (lpm *LSPPoolManager) StartRequest(requestID, method string) *RequestScopedPool {
	lpm.requestMu.Lock()
	defer lpm.requestMu.Unlock()

	scope := &RequestScopedPool{
		requestID:        requestID,
		allocatedObjects: make([]interface{}, 0, 16),
		startTime:        time.Now().UnixNano(),
	}

	lpm.requestPools[requestID] = scope
	atomic.AddInt64(&lpm.requestCount, 1)

	if lpm.hooks.OnRequestStart != nil {
		lpm.hooks.OnRequestStart(requestID, method)
	}

	if lpm.hooks.OnRequestScopeCreate != nil {
		lpm.hooks.OnRequestScopeCreate(requestID, 0)
	}

	return scope
}

// EndRequest cleans up the request-scoped pool
func (lpm *LSPPoolManager) EndRequest(requestID string) {
	lpm.requestMu.Lock()
	scope, exists := lpm.requestPools[requestID]
	if exists {
		delete(lpm.requestPools, requestID)
	}
	lpm.requestMu.Unlock()

	if !exists {
		return
	}

	// Calculate duration
	duration := time.Now().UnixNano() - scope.startTime

	// Clean up allocated objects
	objectsReleased := len(scope.allocatedObjects)

	// Reset objects back to pools
	scope.mu.Lock()
	for _, obj := range scope.allocatedObjects {
		lpm.returnObjectToPool(obj)
	}
	scope.allocatedObjects = scope.allocatedObjects[:0]
	scope.mu.Unlock()

	if lpm.hooks.OnRequestEnd != nil {
		lpm.hooks.OnRequestEnd(requestID, duration)
	}

	if lpm.hooks.OnRequestScopeClean != nil {
		lpm.hooks.OnRequestScopeClean(requestID, objectsReleased)
	}
}

// Pool statistics methods

// RequestCount returns the total number of requests processed
func (lpm *LSPPoolManager) RequestCount() int64 {
	return atomic.LoadInt64(&lpm.requestCount)
}

// CompletionCount returns the total number of completion requests
func (lpm *LSPPoolManager) CompletionCount() int64 {
	return atomic.LoadInt64(&lpm.completionCount)
}

// DiagnosticCount returns the total number of diagnostic publications
func (lpm *LSPPoolManager) DiagnosticCount() int64 {
	return atomic.LoadInt64(&lpm.diagnosticCount)
}

// PoolEfficiency returns the pool hit rate as a percentage
func (lpm *LSPPoolManager) PoolEfficiency() float64 {
	hits := atomic.LoadInt64(&lpm.poolHits)
	misses := atomic.LoadInt64(&lpm.poolMisses)
	total := hits + misses

	if total == 0 {
		return 0
	}
	return float64(hits) / float64(total) * 100
}

// ObjectReuseRate returns the object reuse rate as a percentage
func (lpm *LSPPoolManager) ObjectReuseRate() float64 {
	created := atomic.LoadInt64(&lpm.objectsCreated)
	reused := atomic.LoadInt64(&lpm.objectsReused)
	total := created + reused

	if total == 0 {
		return 0
	}
	return float64(reused) / float64(total) * 100
}

// JSON-RPC message creation methods

// NewJSONRPCMessage creates a pooled JSON-RPC message
func (lpm *LSPPoolManager) NewJSONRPCMessage(requestID string) *JSONRPCMessage {
	msg := lpm.jsonRPCMessagePool.New()

	// Reset/initialize the pooled object
	lpm.resetJSONRPCMessage(msg)

	// Initialize fields
	msg.JSONRPC = "2.0"

	atomic.AddInt64(&lpm.poolHits, 1)
	lpm.trackObject(requestID, msg)

	if lpm.hooks.OnObjectCreate != nil {
		lpm.hooks.OnObjectCreate("JSONRPCMessage")
	}

	return msg
}

// NewJSONRPCResponse creates a pooled JSON-RPC response
func (lpm *LSPPoolManager) NewJSONRPCResponse(requestID string) *JSONRPCResponse {
	resp := lpm.jsonRPCResponsePool.New()

	// Reset/initialize the pooled object
	lpm.resetJSONRPCResponse(resp)

	// Initialize fields
	resp.JSONRPC = "2.0"

	atomic.AddInt64(&lpm.poolHits, 1)
	lpm.trackObject(requestID, resp)

	if lpm.hooks.OnObjectCreate != nil {
		lpm.hooks.OnObjectCreate("JSONRPCResponse")
	}

	return resp
}

// NewJSONRPCError creates a pooled JSON-RPC error
func (lpm *LSPPoolManager) NewJSONRPCError(requestID string, code int, message string) *JSONRPCError {
	err := lpm.jsonRPCErrorPool.New()

	// Initialize fields
	err.Code = code
	err.Message = message
	err.Data = nil

	atomic.AddInt64(&lpm.poolHits, 1)
	lpm.trackObject(requestID, err)

	if lpm.hooks.OnObjectCreate != nil {
		lpm.hooks.OnObjectCreate("JSONRPCError")
	}

	return err
}

// LSP basic type creation methods

// NewPosition creates a pooled LSP position
func (lpm *LSPPoolManager) NewPosition(requestID string, line, character int) *Position {
	pos := lpm.positionPool.New()

	// Initialize fields
	pos.Line = line
	pos.Character = character

	atomic.AddInt64(&lpm.poolHits, 1)
	lpm.trackObject(requestID, pos)

	if lpm.hooks.OnObjectCreate != nil {
		lpm.hooks.OnObjectCreate("Position")
	}

	return pos
}

// NewRange creates a pooled LSP range
func (lpm *LSPPoolManager) NewRange(requestID string, start, end Position) *Range {
	rng := lpm.rangePool.New()

	// Initialize fields
	rng.Start = start
	rng.End = end

	atomic.AddInt64(&lpm.poolHits, 1)
	lpm.trackObject(requestID, rng)

	if lpm.hooks.OnObjectCreate != nil {
		lpm.hooks.OnObjectCreate("Range")
	}

	return rng
}

// NewLocation creates a pooled LSP location
func (lpm *LSPPoolManager) NewLocation(requestID, uri string, rng Range) *Location {
	loc := lpm.locationPool.New()

	// Initialize fields
	loc.URI = uri
	loc.Range = rng

	atomic.AddInt64(&lpm.poolHits, 1)
	lpm.trackObject(requestID, loc)

	if lpm.hooks.OnObjectCreate != nil {
		lpm.hooks.OnObjectCreate("Location")
	}

	return loc
}

// Completion creation methods

// NewCompletionItem creates a pooled completion item
func (lpm *LSPPoolManager) NewCompletionItem(requestID, label, detail string) *CompletionItem {
	item := lpm.completionItemPool.New()

	// Reset/initialize the pooled object
	lpm.resetCompletionItem(item)

	// Initialize fields
	item.Label = label
	item.Detail = detail

	atomic.AddInt64(&lpm.completionCount, 1)
	atomic.AddInt64(&lpm.poolHits, 1)
	lpm.trackObject(requestID, item)

	if lpm.hooks.OnCompletionCreate != nil {
		lpm.hooks.OnCompletionCreate(item)
	}

	if lpm.hooks.OnObjectCreate != nil {
		lpm.hooks.OnObjectCreate("CompletionItem")
	}

	return item
}

// NewCompletionList creates a pooled completion list
func (lpm *LSPPoolManager) NewCompletionList(requestID string, isIncomplete bool) *CompletionList {
	list := lpm.completionListPool.New()

	// Reset/initialize the pooled object
	lpm.resetCompletionList(list)

	// Initialize fields
	list.IsIncomplete = isIncomplete
	list.Items = lpm.getCompletionItemSlice()

	atomic.AddInt64(&lpm.poolHits, 1)
	lpm.trackObject(requestID, list)

	if lpm.hooks.OnObjectCreate != nil {
		lpm.hooks.OnObjectCreate("CompletionList")
	}

	return list
}

// Diagnostic creation methods

// NewDiagnostic creates a pooled diagnostic
func (lpm *LSPPoolManager) NewDiagnostic(requestID string, rng Range, message string, severity *DiagnosticSeverity) *Diagnostic {
	diag := lpm.diagnosticPool.New()

	// Reset/initialize the pooled object
	lpm.resetDiagnostic(diag)

	// Initialize fields
	diag.Range = rng
	diag.Message = message
	diag.Severity = severity

	atomic.AddInt64(&lpm.diagnosticCount, 1)
	atomic.AddInt64(&lpm.poolHits, 1)
	lpm.trackObject(requestID, diag)

	if lpm.hooks.OnDiagnosticCreate != nil {
		lpm.hooks.OnDiagnosticCreate(diag)
	}

	if lpm.hooks.OnObjectCreate != nil {
		lpm.hooks.OnObjectCreate("Diagnostic")
	}

	return diag
}

// NewPublishDiagnosticsParams creates a pooled publish diagnostics params
func (lpm *LSPPoolManager) NewPublishDiagnosticsParams(requestID, uri string) *PublishDiagnosticsParams {
	params := lpm.publishDiagnosticsPool.New()

	// Reset/initialize the pooled object
	lpm.resetPublishDiagnosticsParams(params)

	// Initialize fields
	params.URI = uri
	params.Diagnostics = lpm.getDiagnosticSlice()

	atomic.AddInt64(&lpm.poolHits, 1)
	lpm.trackObject(requestID, params)

	if lpm.hooks.OnObjectCreate != nil {
		lpm.hooks.OnObjectCreate("PublishDiagnosticsParams")
	}

	return params
}

// Helper methods for getting pooled collections

// getCompletionItemSlice returns a pooled completion item slice
func (lpm *LSPPoolManager) getCompletionItemSlice() []CompletionItem {
	s := lpm.completionItemSlicePool.New()
	if *s == nil {
		*s = make([]CompletionItem, 0, 16)
		atomic.AddInt64(&lpm.poolMisses, 1)
	} else {
		*s = (*s)[:0] // Reset length but keep capacity
		atomic.AddInt64(&lpm.poolHits, 1)
	}
	return *s
}

// getDiagnosticSlice returns a pooled diagnostic slice
func (lpm *LSPPoolManager) getDiagnosticSlice() []Diagnostic {
	s := lpm.diagnosticSlicePool.New()
	if *s == nil {
		*s = make([]Diagnostic, 0, 8)
		atomic.AddInt64(&lpm.poolMisses, 1)
	} else {
		*s = (*s)[:0] // Reset length but keep capacity
		atomic.AddInt64(&lpm.poolHits, 1)
	}
	return *s
}

// getLocationSlice returns a pooled location slice
func (lpm *LSPPoolManager) getLocationSlice() []Location {
	s := lpm.locationSlicePool.New()
	if *s == nil {
		*s = make([]Location, 0, 4)
		atomic.AddInt64(&lpm.poolMisses, 1)
	} else {
		*s = (*s)[:0] // Reset length but keep capacity
		atomic.AddInt64(&lpm.poolHits, 1)
	}
	return *s
}

// getStringMap returns a pooled string map
func (lpm *LSPPoolManager) getStringMap() map[string]string {
	m := lpm.stringMapPool.New()
	if *m == nil {
		*m = make(map[string]string)
		atomic.AddInt64(&lpm.poolMisses, 1)
	} else {
		lpm.clearStringMap(*m)
		atomic.AddInt64(&lpm.poolHits, 1)
	}
	return *m
}

// Pool warming methods

// warmPools pre-allocates common LSP objects for better performance
func (lpm *LSPPoolManager) warmPools() {
	if lpm.hooks.OnPoolWarming != nil {
		lpm.hooks.OnPoolWarming("lsp-protocol")
	}

	// Warm up pools with common LSP patterns
	lpm.warmJSONRPCObjects()
	lpm.warmBasicTypes()
	lpm.warmCompletionObjects()
	lpm.warmDiagnosticObjects()

	if lpm.hooks.OnPoolWarming != nil {
		lpm.hooks.OnPoolWarming("lsp-protocol-complete")
	}
}

// warmJSONRPCObjects pre-allocates JSON-RPC message structures
func (lpm *LSPPoolManager) warmJSONRPCObjects() {
	// Create a few JSON-RPC objects to warm the pools
	for i := 0; i < 4; i++ {
		msg := lpm.NewJSONRPCMessage("warmup")
		lpm.resetJSONRPCMessage(msg)

		resp := lpm.NewJSONRPCResponse("warmup")
		lpm.resetJSONRPCResponse(resp)
	}
}

// warmBasicTypes pre-allocates common LSP basic type structures
func (lpm *LSPPoolManager) warmBasicTypes() {
	// Create basic LSP types to warm the pools
	for i := 0; i < 8; i++ {
		pos := lpm.NewPosition("warmup", 0, 0)
		rng := lpm.NewRange("warmup", *pos, *pos)
		loc := lpm.NewLocation("warmup", "", *rng)
		_ = loc // Use the variable to avoid compiler warnings
	}
}

// warmCompletionObjects pre-allocates completion structures
func (lpm *LSPPoolManager) warmCompletionObjects() {
	// Create completion objects to warm the pools
	for i := 0; i < 8; i++ {
		item := lpm.NewCompletionItem("warmup", "", "")
		lpm.resetCompletionItem(item)

		list := lpm.NewCompletionList("warmup", false)
		lpm.resetCompletionList(list)
	}
}

// warmDiagnosticObjects pre-allocates diagnostic structures
func (lpm *LSPPoolManager) warmDiagnosticObjects() {
	// Create diagnostic objects to warm the pools
	for i := 0; i < 4; i++ {
		pos := lpm.NewPosition("warmup", 0, 0)
		rng := lpm.NewRange("warmup", *pos, *pos)
		diag := lpm.NewDiagnostic("warmup", *rng, "", nil)
		lpm.resetDiagnostic(diag)

		params := lpm.NewPublishDiagnosticsParams("warmup", "")
		lpm.resetPublishDiagnosticsParams(params)
	}
}

// Reset methods for proper object cleanup and reuse

// resetJSONRPCMessage resets a JSON-RPC message for reuse
func (lpm *LSPPoolManager) resetJSONRPCMessage(msg *JSONRPCMessage) {
	msg.JSONRPC = ""
	msg.ID = nil
	msg.Method = ""
	msg.Params = nil

	if lpm.hooks.OnObjectReset != nil {
		lpm.hooks.OnObjectReset("JSONRPCMessage")
	}
}

// resetJSONRPCResponse resets a JSON-RPC response for reuse
func (lpm *LSPPoolManager) resetJSONRPCResponse(resp *JSONRPCResponse) {
	resp.JSONRPC = ""
	resp.ID = nil
	resp.Result = nil
	resp.Error = nil

	if lpm.hooks.OnObjectReset != nil {
		lpm.hooks.OnObjectReset("JSONRPCResponse")
	}
}

// resetCompletionItem resets a completion item for reuse
func (lpm *LSPPoolManager) resetCompletionItem(item *CompletionItem) {
	item.Label = ""
	item.Kind = nil
	item.Tags = nil
	item.Detail = ""
	item.Documentation = nil
	item.Deprecated = false
	item.Preselect = false
	item.SortText = ""
	item.FilterText = ""
	item.InsertText = ""
	item.InsertTextFormat = nil
	item.InsertTextMode = nil
	item.TextEdit = nil
	item.AdditionalTextEdits = nil
	item.CommitCharacters = nil
	item.Command = nil
	item.Data = nil

	if lpm.hooks.OnObjectReset != nil {
		lpm.hooks.OnObjectReset("CompletionItem")
	}
}

// resetCompletionList resets a completion list for reuse
func (lpm *LSPPoolManager) resetCompletionList(list *CompletionList) {
	list.IsIncomplete = false
	if list.Items != nil {
		list.Items = list.Items[:0]
	}

	if lpm.hooks.OnObjectReset != nil {
		lpm.hooks.OnObjectReset("CompletionList")
	}
}

// resetDiagnostic resets a diagnostic for reuse
func (lpm *LSPPoolManager) resetDiagnostic(diag *Diagnostic) {
	diag.Range = Range{}
	diag.Severity = nil
	diag.Code = nil
	diag.CodeDescription = nil
	diag.Source = ""
	diag.Message = ""
	diag.Tags = nil
	diag.RelatedInformation = nil
	diag.Data = nil

	if lpm.hooks.OnObjectReset != nil {
		lpm.hooks.OnObjectReset("Diagnostic")
	}
}

// resetPublishDiagnosticsParams resets publish diagnostics params for reuse
func (lpm *LSPPoolManager) resetPublishDiagnosticsParams(params *PublishDiagnosticsParams) {
	params.URI = ""
	params.Version = nil
	if params.Diagnostics != nil {
		params.Diagnostics = params.Diagnostics[:0]
	}

	if lpm.hooks.OnObjectReset != nil {
		lpm.hooks.OnObjectReset("PublishDiagnosticsParams")
	}
}

// Helper methods to clear maps efficiently

// clearStringMap clears a string map efficiently
func (lpm *LSPPoolManager) clearStringMap(m map[string]string) {
	for k := range m {
		delete(m, k)
	}
}

// Request tracking methods

// trackObject adds an object to the request-scoped pool
func (lpm *LSPPoolManager) trackObject(requestID string, obj interface{}) {
	if requestID == "" || requestID == "warmup" {
		return
	}

	lpm.requestMu.RLock()
	scope, exists := lpm.requestPools[requestID]
	lpm.requestMu.RUnlock()

	if exists {
		scope.mu.Lock()
		scope.allocatedObjects = append(scope.allocatedObjects, obj)
		scope.mu.Unlock()
	}
}

// returnObjectToPool returns an object to the appropriate pool
func (lpm *LSPPoolManager) returnObjectToPool(obj interface{}) {
	// Use type assertions instead of switch to avoid go-critic caseOrder issues
	if v, ok := obj.(*CompletionItem); ok {
		lpm.resetCompletionItem(v)
		atomic.AddInt64(&lpm.objectsReused, 1)
		return
	}
	if v, ok := obj.(*CompletionList); ok {
		lpm.resetCompletionList(v)
		atomic.AddInt64(&lpm.objectsReused, 1)
		return
	}
	if v, ok := obj.(*Diagnostic); ok {
		lpm.resetDiagnostic(v)
		atomic.AddInt64(&lpm.objectsReused, 1)
		return
	}
	if v, ok := obj.(*JSONRPCMessage); ok {
		lpm.resetJSONRPCMessage(v)
		atomic.AddInt64(&lpm.objectsReused, 1)
		return
	}
	if v, ok := obj.(*JSONRPCResponse); ok {
		lpm.resetJSONRPCResponse(v)
		atomic.AddInt64(&lpm.objectsReused, 1)
		return
	}
	if v, ok := obj.(*PublishDiagnosticsParams); ok {
		lpm.resetPublishDiagnosticsParams(v)
		atomic.AddInt64(&lpm.objectsReused, 1)
		return
	}
	// Add other types as needed
}

// Pool management methods

// Reset resets all pools for reuse
func (lpm *LSPPoolManager) Reset() {
	lpm.mu.Lock()
	defer lpm.mu.Unlock()

	// Reset all core pools
	lpm.jsonRPCMessagePool.Reset()
	lpm.jsonRPCResponsePool.Reset()
	lpm.jsonRPCErrorPool.Reset()
	lpm.jsonRPCNotificationPool.Reset()
	lpm.positionPool.Reset()
	lpm.rangePool.Reset()
	lpm.locationPool.Reset()
	lpm.completionItemPool.Reset()
	lpm.completionListPool.Reset()
	lpm.diagnosticPool.Reset()
	lpm.publishDiagnosticsPool.Reset()
}

// Clear completely empties all pools and resets statistics
func (lpm *LSPPoolManager) Clear() {
	lpm.mu.Lock()
	defer lpm.mu.Unlock()

	// Clear all core pools
	lpm.jsonRPCMessagePool.Clear()
	lpm.jsonRPCResponsePool.Clear()
	lpm.jsonRPCErrorPool.Clear()
	lpm.jsonRPCNotificationPool.Clear()
	lpm.positionPool.Clear()
	lpm.rangePool.Clear()
	lpm.locationPool.Clear()
	lpm.completionItemPool.Clear()
	lpm.completionListPool.Clear()
	lpm.diagnosticPool.Clear()
	lpm.publishDiagnosticsPool.Clear()

	// Reset statistics
	atomic.StoreInt64(&lpm.requestCount, 0)
	atomic.StoreInt64(&lpm.completionCount, 0)
	atomic.StoreInt64(&lpm.diagnosticCount, 0)
	atomic.StoreInt64(&lpm.poolHits, 0)
	atomic.StoreInt64(&lpm.poolMisses, 0)
	atomic.StoreInt64(&lpm.memoryReused, 0)
	atomic.StoreInt64(&lpm.objectsCreated, 0)
	atomic.StoreInt64(&lpm.objectsReused, 0)
}

// GetDetailedStats returns detailed statistics about all LSP pools
func (lpm *LSPPoolManager) GetDetailedStats() LSPPoolStats {
	return LSPPoolStats{
		RequestCount:    atomic.LoadInt64(&lpm.requestCount),
		CompletionCount: atomic.LoadInt64(&lpm.completionCount),
		DiagnosticCount: atomic.LoadInt64(&lpm.diagnosticCount),
		PoolHits:        atomic.LoadInt64(&lpm.poolHits),
		PoolMisses:      atomic.LoadInt64(&lpm.poolMisses),
		MemoryReused:    atomic.LoadInt64(&lpm.memoryReused),
		ObjectsCreated:  atomic.LoadInt64(&lpm.objectsCreated),
		ObjectsReused:   atomic.LoadInt64(&lpm.objectsReused),
		PoolEfficiency:  lpm.PoolEfficiency(),
		ObjectReuseRate: lpm.ObjectReuseRate(),
	}
}

// LSPPoolStats contains detailed LSP pool statistics
type LSPPoolStats struct {
	RequestCount    int64   // Total number of LSP requests processed
	CompletionCount int64   // Total number of completion requests
	DiagnosticCount int64   // Total number of diagnostic publications
	PoolHits        int64   // Number of successful pool allocations
	PoolMisses      int64   // Number of pool misses requiring new allocation
	MemoryReused    int64   // Total amount of memory reused
	ObjectsCreated  int64   // Number of objects created
	ObjectsReused   int64   // Number of objects reused from pool
	PoolEfficiency  float64 // Pool hit rate as percentage
	ObjectReuseRate float64 // Object reuse rate as percentage
}

// Global LSP pool manager instance
var globalLSPPoolManager *LSPPoolManager
var lspPoolOnce sync.Once

// GlobalLSPPoolManager returns the global LSP pool manager instance
func GlobalLSPPoolManager() *LSPPoolManager {
	lspPoolOnce.Do(func() {
		globalLSPPoolManager = NewLSPPoolManager(LSPPoolHooks{
			// Default hooks can be set here
			OnPoolWarming: func(poolType string) {
				// Default pool warming notification
			},
		})
	})
	return globalLSPPoolManager
}

// SetGlobalLSPPoolHooks sets hooks for the global LSP pool manager
func SetGlobalLSPPoolHooks(hooks LSPPoolHooks) {
	lpm := GlobalLSPPoolManager()
	lpm.hooks = hooks
}
