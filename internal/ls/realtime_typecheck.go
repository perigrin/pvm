// ABOUTME: Real-time type checking integration for live error detection
// ABOUTME: Provides incremental type checking with debouncing and caching

package ls

import (
	"context"
	"sync"
	"time"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/binder"
)

// RealtimeTypeChecker provides incremental type checking
type RealtimeTypeChecker struct {
	ls             *LanguageService
	checkQueue     chan typeCheckRequest
	results        map[string]*TypeCheckResult
	resultsMu      sync.RWMutex
	debounceTimers map[string]*time.Timer
	debounceMs     int
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup
}

// typeCheckRequest represents a request to type check a document
type typeCheckRequest struct {
	URI      string
	Version  int
	Force    bool // Skip debouncing
	Response chan<- *TypeCheckResult
}

// TypeCheckResult contains the results of type checking
type TypeCheckResult struct {
	URI         string
	Version     int
	Diagnostics []EnhancedDiagnostic
	Timestamp   time.Time
	Duration    time.Duration
	Incremental bool // Whether this was an incremental check
}

// NewRealtimeTypeChecker creates a new real-time type checker
func NewRealtimeTypeChecker(ls *LanguageService, debounceMs int) *RealtimeTypeChecker {
	ctx, cancel := context.WithCancel(context.Background())

	rtc := &RealtimeTypeChecker{
		ls:             ls,
		checkQueue:     make(chan typeCheckRequest, 100),
		results:        make(map[string]*TypeCheckResult),
		debounceTimers: make(map[string]*time.Timer),
		debounceMs:     debounceMs,
		ctx:            ctx,
		cancel:         cancel,
	}

	// Start worker goroutines
	for i := 0; i < 4; i++ { // 4 concurrent type checkers
		rtc.wg.Add(1)
		go rtc.typeCheckWorker()
	}

	return rtc
}

// Stop gracefully shuts down the type checker
func (rtc *RealtimeTypeChecker) Stop() {
	rtc.cancel()
	close(rtc.checkQueue)
	rtc.wg.Wait()
}

// QueueTypeCheck queues a document for type checking with debouncing
func (rtc *RealtimeTypeChecker) QueueTypeCheck(uri string, version int) <-chan *TypeCheckResult {
	response := make(chan *TypeCheckResult, 1)

	// Cancel any existing timer for this URI
	if timer, exists := rtc.debounceTimers[uri]; exists {
		timer.Stop()
	}

	// Create debounced request
	rtc.debounceTimers[uri] = time.AfterFunc(time.Duration(rtc.debounceMs)*time.Millisecond, func() {
		select {
		case rtc.checkQueue <- typeCheckRequest{
			URI:      uri,
			Version:  version,
			Force:    false,
			Response: response,
		}:
		case <-rtc.ctx.Done():
		}
	})

	return response
}

// ForceTypeCheck immediately type checks a document (no debouncing)
func (rtc *RealtimeTypeChecker) ForceTypeCheck(uri string, version int) <-chan *TypeCheckResult {
	response := make(chan *TypeCheckResult, 1)

	// Cancel any pending debounced check
	if timer, exists := rtc.debounceTimers[uri]; exists {
		timer.Stop()
		delete(rtc.debounceTimers, uri)
	}

	select {
	case rtc.checkQueue <- typeCheckRequest{
		URI:      uri,
		Version:  version,
		Force:    true,
		Response: response,
	}:
	case <-rtc.ctx.Done():
		close(response)
	}

	return response
}

// GetCachedResult returns cached type check result if available
func (rtc *RealtimeTypeChecker) GetCachedResult(uri string) (*TypeCheckResult, bool) {
	rtc.resultsMu.RLock()
	defer rtc.resultsMu.RUnlock()

	result, ok := rtc.results[uri]
	return result, ok
}

// typeCheckWorker processes type check requests
func (rtc *RealtimeTypeChecker) typeCheckWorker() {
	defer rtc.wg.Done()

	for {
		select {
		case req, ok := <-rtc.checkQueue:
			if !ok {
				return
			}

			result := rtc.performTypeCheck(req)

			// Cache result
			rtc.resultsMu.Lock()
			rtc.results[req.URI] = result
			rtc.resultsMu.Unlock()

			// Send response
			select {
			case req.Response <- result:
			case <-rtc.ctx.Done():
			}
			close(req.Response)

		case <-rtc.ctx.Done():
			return
		}
	}
}

// performTypeCheck executes the type checking
func (rtc *RealtimeTypeChecker) performTypeCheck(req typeCheckRequest) *TypeCheckResult {
	start := time.Now()

	result := &TypeCheckResult{
		URI:       req.URI,
		Version:   req.Version,
		Timestamp: start,
	}

	// Get document from language service
	doc, exists := rtc.ls.GetDocumentForDebug(req.URI)
	if !exists {
		return result
	}

	// Check if we can do incremental checking
	if rtc.canDoIncrementalCheck(doc, req) {
		result.Incremental = true
		diagnostics := rtc.performIncrementalTypeCheck(doc, req)
		result.Diagnostics = diagnostics
	} else {
		// Full type check
		diagnostics, err := rtc.ls.GetEnhancedDiagnostics(req.URI)
		if err == nil {
			result.Diagnostics = diagnostics
		}
	}

	result.Duration = time.Since(start)
	return result
}

// canDoIncrementalCheck determines if incremental checking is possible
func (rtc *RealtimeTypeChecker) canDoIncrementalCheck(doc *Document, req typeCheckRequest) bool {
	// Check if we have a previous result
	rtc.resultsMu.RLock()
	prevResult, hasPrev := rtc.results[req.URI]
	rtc.resultsMu.RUnlock()

	if !hasPrev {
		return false
	}

	// Check if AST structure has changed significantly
	// In a real implementation, this would compare AST hashes

	return doc.Version == prevResult.Version+1 // Simple version check
}

// performIncrementalTypeCheck does incremental type checking
func (rtc *RealtimeTypeChecker) performIncrementalTypeCheck(doc *Document, req typeCheckRequest) []EnhancedDiagnostic {
	// In a real implementation, this would:
	// 1. Identify changed regions of the document
	// 2. Re-type-check only affected scopes
	// 3. Merge with cached results for unchanged regions

	// For now, fall back to full type check
	diagnostics, _ := rtc.ls.GetEnhancedDiagnostics(req.URI)
	return diagnostics
}

// IncrementalTypeCheckContext tracks changes for incremental checking
type IncrementalTypeCheckContext struct {
	PreviousAST    *ast.AST
	CurrentAST     *ast.AST
	ChangedNodes   []ast.Node
	AffectedScopes []*ScopeChange
	CachedResults  map[string][]EnhancedDiagnostic
}

// ScopeChange represents a changed scope
type ScopeChange struct {
	Scope      *binder.Scope
	ChangeType RTChangeType
	Node       ast.Node
}

// RTChangeType represents the type of change in real-time checking
type RTChangeType int

const (
	RTChangeTypeAdded RTChangeType = iota
	RTChangeTypeModified
	RTChangeTypeDeleted
)

// TypeCheckStats provides statistics about type checking performance
type TypeCheckStats struct {
	TotalChecks       int
	IncrementalChecks int
	AverageTime       time.Duration
	CacheHitRate      float64
}

// GetStats returns type checking statistics
func (rtc *RealtimeTypeChecker) GetStats() TypeCheckStats {
	rtc.resultsMu.RLock()
	defer rtc.resultsMu.RUnlock()

	stats := TypeCheckStats{
		TotalChecks: len(rtc.results),
	}

	var totalTime time.Duration
	for _, result := range rtc.results {
		totalTime += result.Duration
		if result.Incremental {
			stats.IncrementalChecks++
		}
	}

	if stats.TotalChecks > 0 {
		stats.AverageTime = totalTime / time.Duration(stats.TotalChecks)
		stats.CacheHitRate = float64(stats.IncrementalChecks) / float64(stats.TotalChecks)
	}

	return stats
}

// DiagnosticPublisher handles publishing diagnostics to the client
type DiagnosticPublisher struct {
	sendFunc   func(uri string, diagnostics []Diagnostic) error
	batchDelay time.Duration
	batches    map[string]*diagnosticBatch
	mu         sync.Mutex
}

// diagnosticBatch represents a batch of diagnostics
type diagnosticBatch struct {
	diagnostics []Diagnostic
	timer       *time.Timer
	version     int
}

// NewDiagnosticPublisher creates a new diagnostic publisher
func NewDiagnosticPublisher(sendFunc func(uri string, diagnostics []Diagnostic) error, batchDelay time.Duration) *DiagnosticPublisher {
	return &DiagnosticPublisher{
		sendFunc:   sendFunc,
		batchDelay: batchDelay,
		batches:    make(map[string]*diagnosticBatch),
	}
}

// PublishDiagnostics publishes diagnostics with batching
func (dp *DiagnosticPublisher) PublishDiagnostics(uri string, diagnostics []EnhancedDiagnostic, version int) {
	dp.mu.Lock()
	defer dp.mu.Unlock()

	// Convert enhanced diagnostics to basic diagnostics
	basicDiags := make([]Diagnostic, len(diagnostics))
	for i, diag := range diagnostics {
		basicDiags[i] = diag.Diagnostic
	}

	// Cancel existing batch for this URI
	if batch, exists := dp.batches[uri]; exists {
		batch.timer.Stop()
	}

	// Create new batch
	timer := time.AfterFunc(dp.batchDelay, func() {
		dp.sendBatch(uri)
	})

	dp.batches[uri] = &diagnosticBatch{
		diagnostics: basicDiags,
		timer:       timer,
		version:     version,
	}
}

// sendBatch sends a batched set of diagnostics
func (dp *DiagnosticPublisher) sendBatch(uri string) {
	dp.mu.Lock()
	batch, exists := dp.batches[uri]
	if !exists {
		dp.mu.Unlock()
		return
	}
	delete(dp.batches, uri)
	dp.mu.Unlock()

	// Send diagnostics
	_ = dp.sendFunc(uri, batch.diagnostics)
}

// TypeCheckCoordinator coordinates type checking across multiple files
type TypeCheckCoordinator struct {
	checker   *RealtimeTypeChecker
	publisher *DiagnosticPublisher
	queue     *PriorityQueue
	mu        sync.Mutex
}

// PriorityQueue manages type check priorities
type PriorityQueue struct {
	items []QueueItem
	mu    sync.Mutex
}

// QueueItem represents an item in the priority queue
type QueueItem struct {
	URI      string
	Priority int
	Time     time.Time
}

// NewTypeCheckCoordinator creates a new coordinator
func NewTypeCheckCoordinator(ls *LanguageService, sendFunc func(uri string, diagnostics []Diagnostic) error) *TypeCheckCoordinator {
	return &TypeCheckCoordinator{
		checker:   NewRealtimeTypeChecker(ls, 300), // 300ms debounce
		publisher: NewDiagnosticPublisher(sendFunc, 100*time.Millisecond),
		queue:     &PriorityQueue{},
	}
}

// HandleDocumentChange handles document changes for type checking
func (tcc *TypeCheckCoordinator) HandleDocumentChange(uri string, version int, priority int) {
	// Queue for type checking based on priority
	tcc.queue.Push(QueueItem{
		URI:      uri,
		Priority: priority,
		Time:     time.Now(),
	})

	// Process queue
	go tcc.processQueue()
}

// processQueue processes the type check queue
func (tcc *TypeCheckCoordinator) processQueue() {
	for {
		item, ok := tcc.queue.Pop()
		if !ok {
			break
		}

		// Type check the document
		resultChan := tcc.checker.QueueTypeCheck(item.URI, 0) // Version from item

		go func(uri string) {
			result := <-resultChan
			if result != nil {
				// Publish diagnostics
				tcc.publisher.PublishDiagnostics(uri, result.Diagnostics, result.Version)
			}
		}(item.URI)
	}
}

// Priority queue methods
func (pq *PriorityQueue) Push(item QueueItem) {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	// Insert in priority order
	inserted := false
	for i, existing := range pq.items {
		if item.Priority > existing.Priority {
			pq.items = append(pq.items[:i], append([]QueueItem{item}, pq.items[i:]...)...)
			inserted = true
			break
		}
	}

	if !inserted {
		pq.items = append(pq.items, item)
	}
}

func (pq *PriorityQueue) Pop() (QueueItem, bool) {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	if len(pq.items) == 0 {
		return QueueItem{}, false
	}

	item := pq.items[0]
	pq.items = pq.items[1:]
	return item, true
}

// DocumentPriority calculates priority for a document
func DocumentPriority(uri string, isActive bool, hasErrors bool) int {
	priority := 0

	// Active document gets highest priority
	if isActive {
		priority += 100
	}

	// Documents with errors get higher priority
	if hasErrors {
		priority += 50
	}

	// Recently modified documents get higher priority
	// (would need modification time tracking)

	return priority
}
