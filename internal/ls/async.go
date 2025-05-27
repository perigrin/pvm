// ABOUTME: Asynchronous request processing for LSP operations to improve responsiveness
// ABOUTME: Provides prioritized request queues and background processing for expensive operations

package ls

import (
	"context"
	"sync"
	"time"
)

// RequestPriority defines the priority level of LSP requests
type RequestPriority int

const (
	PriorityHigh   RequestPriority = iota // Hover, completion - immediate response needed
	PriorityMedium                        // Definition, references - quick response expected
	PriorityLow                           // Formatting, code actions - can be processed in background
)

// LSPRequest represents a queued LSP operation
type LSPRequest struct {
	ID         string
	Type       string
	Priority   RequestPriority
	Context    context.Context
	Handler    func() (interface{}, error)
	ResultChan chan LSPResult
	Timestamp  time.Time
	Timeout    time.Duration
}

// LSPResult represents the result of an LSP operation
type LSPResult struct {
	Data  interface{}
	Error error
}

// RequestQueue manages prioritized request processing
type RequestQueue struct {
	highQueue   chan *LSPRequest
	mediumQueue chan *LSPRequest
	lowQueue    chan *LSPRequest

	workers  int
	stopChan chan struct{}
	wg       sync.WaitGroup

	// Statistics
	processed  int64
	failed     int64
	avgLatency time.Duration
	mu         sync.RWMutex
}

// AsyncLanguageService wraps LanguageService with async capabilities
type AsyncLanguageService struct {
	*LanguageService
	queue *RequestQueue
}

// NewRequestQueue creates a new request queue with specified number of workers
func NewRequestQueue(workers int) *RequestQueue {
	return &RequestQueue{
		highQueue:   make(chan *LSPRequest, 100),
		mediumQueue: make(chan *LSPRequest, 200),
		lowQueue:    make(chan *LSPRequest, 500),
		workers:     workers,
		stopChan:    make(chan struct{}),
	}
}

// NewAsyncLanguageService creates a language service with async processing
func NewAsyncLanguageService() (*AsyncLanguageService, error) {
	ls, err := NewLanguageService()
	if err != nil {
		return nil, err
	}

	queue := NewRequestQueue(4) // 4 worker goroutines
	queue.Start()

	return &AsyncLanguageService{
		LanguageService: ls,
		queue:           queue,
	}, nil
}

// Start begins processing requests with worker goroutines
func (rq *RequestQueue) Start() {
	for i := 0; i < rq.workers; i++ {
		rq.wg.Add(1)
		go rq.worker(i)
	}
}

// Stop gracefully shuts down the request queue
func (rq *RequestQueue) Stop() {
	close(rq.stopChan)
	rq.wg.Wait()
}

// worker processes requests from the queues with priority ordering
func (rq *RequestQueue) worker(id int) {
	defer rq.wg.Done()

	for {
		select {
		case <-rq.stopChan:
			return

		// High priority requests are processed first
		case req := <-rq.highQueue:
			rq.processRequest(req)

		default:
			// If no high priority requests, check medium and low
			select {
			case <-rq.stopChan:
				return

			case req := <-rq.highQueue:
				rq.processRequest(req)

			case req := <-rq.mediumQueue:
				rq.processRequest(req)

			default:
				// If no high or medium priority requests, check low or wait
				select {
				case <-rq.stopChan:
					return

				case req := <-rq.highQueue:
					rq.processRequest(req)

				case req := <-rq.mediumQueue:
					rq.processRequest(req)

				case req := <-rq.lowQueue:
					rq.processRequest(req)

				case <-time.After(100 * time.Millisecond):
					// Brief idle period to prevent busy waiting
					continue
				}
			}
		}
	}
}

// processRequest executes a single request and sends the result
func (rq *RequestQueue) processRequest(req *LSPRequest) {
	start := time.Now()

	// Check if request has timed out
	if req.Timeout > 0 && time.Since(req.Timestamp) > req.Timeout {
		req.ResultChan <- LSPResult{Error: context.DeadlineExceeded}
		return
	}

	// Check if context is cancelled
	select {
	case <-req.Context.Done():
		req.ResultChan <- LSPResult{Error: req.Context.Err()}
		return
	default:
	}

	// Execute the request
	result, err := req.Handler()

	// Update statistics
	rq.mu.Lock()
	rq.processed++
	if err != nil {
		rq.failed++
	}

	latency := time.Since(start)
	// Simple moving average for latency
	if rq.avgLatency == 0 {
		rq.avgLatency = latency
	} else {
		rq.avgLatency = (rq.avgLatency + latency) / 2
	}
	rq.mu.Unlock()

	// Send result
	req.ResultChan <- LSPResult{Data: result, Error: err}
}

// SubmitRequest adds a request to the appropriate priority queue
func (rq *RequestQueue) SubmitRequest(req *LSPRequest) error {
	select {
	case <-rq.stopChan:
		return context.Canceled
	default:
	}

	var queue chan *LSPRequest
	switch req.Priority {
	case PriorityHigh:
		queue = rq.highQueue
	case PriorityMedium:
		queue = rq.mediumQueue
	case PriorityLow:
		queue = rq.lowQueue
	default:
		queue = rq.mediumQueue
	}

	select {
	case queue <- req:
		return nil
	case <-time.After(1 * time.Second):
		return context.DeadlineExceeded
	}
}

// GetStats returns queue statistics
func (rq *RequestQueue) GetStats() RequestQueueStats {
	rq.mu.RLock()
	defer rq.mu.RUnlock()

	return RequestQueueStats{
		Processed:   rq.processed,
		Failed:      rq.failed,
		AvgLatency:  rq.avgLatency,
		HighQueue:   len(rq.highQueue),
		MediumQueue: len(rq.mediumQueue),
		LowQueue:    len(rq.lowQueue),
		Workers:     rq.workers,
	}
}

// RequestQueueStats provides statistics about request processing
type RequestQueueStats struct {
	Processed   int64
	Failed      int64
	AvgLatency  time.Duration
	HighQueue   int
	MediumQueue int
	LowQueue    int
	Workers     int
}

// Async wrapper methods for LanguageService operations

// GetHoverAsync provides hover information asynchronously
func (als *AsyncLanguageService) GetHoverAsync(ctx context.Context, uri string, pos Position) (*Hover, error) {
	resultChan := make(chan LSPResult, 1)

	req := &LSPRequest{
		ID:       "hover",
		Type:     "hover",
		Priority: PriorityHigh,
		Context:  ctx,
		Handler: func() (interface{}, error) {
			return als.LanguageService.GetHover(uri, pos)
		},
		ResultChan: resultChan,
		Timestamp:  time.Now(),
		Timeout:    5 * time.Second,
	}

	err := als.queue.SubmitRequest(req)
	if err != nil {
		return nil, err
	}

	select {
	case result := <-resultChan:
		if result.Error != nil {
			return nil, result.Error
		}
		return result.Data.(*Hover), nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// GetCompletionsAsync provides completion suggestions asynchronously
func (als *AsyncLanguageService) GetCompletionsAsync(ctx context.Context, uri string, pos Position) ([]CompletionItem, error) {
	resultChan := make(chan LSPResult, 1)

	req := &LSPRequest{
		ID:       "completion",
		Type:     "completion",
		Priority: PriorityHigh,
		Context:  ctx,
		Handler: func() (interface{}, error) {
			return als.LanguageService.GetCompletions(uri, pos)
		},
		ResultChan: resultChan,
		Timestamp:  time.Now(),
		Timeout:    5 * time.Second,
	}

	err := als.queue.SubmitRequest(req)
	if err != nil {
		return nil, err
	}

	select {
	case result := <-resultChan:
		if result.Error != nil {
			return nil, result.Error
		}
		return result.Data.([]CompletionItem), nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// GetDefinitionAsync finds symbol definition asynchronously
func (als *AsyncLanguageService) GetDefinitionAsync(ctx context.Context, uri string, pos Position) (*Definition, error) {
	resultChan := make(chan LSPResult, 1)

	req := &LSPRequest{
		ID:       "definition",
		Type:     "definition",
		Priority: PriorityMedium,
		Context:  ctx,
		Handler: func() (interface{}, error) {
			return als.LanguageService.GetDefinition(uri, pos)
		},
		ResultChan: resultChan,
		Timestamp:  time.Now(),
		Timeout:    10 * time.Second,
	}

	err := als.queue.SubmitRequest(req)
	if err != nil {
		return nil, err
	}

	select {
	case result := <-resultChan:
		if result.Error != nil {
			return nil, result.Error
		}
		return result.Data.(*Definition), nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// FindReferencesAsync finds symbol references asynchronously
func (als *AsyncLanguageService) FindReferencesAsync(ctx context.Context, uri string, pos Position, includeDeclaration bool) ([]Location, error) {
	resultChan := make(chan LSPResult, 1)

	req := &LSPRequest{
		ID:       "references",
		Type:     "references",
		Priority: PriorityMedium,
		Context:  ctx,
		Handler: func() (interface{}, error) {
			return als.LanguageService.FindReferences(uri, pos, includeDeclaration)
		},
		ResultChan: resultChan,
		Timestamp:  time.Now(),
		Timeout:    15 * time.Second,
	}

	err := als.queue.SubmitRequest(req)
	if err != nil {
		return nil, err
	}

	select {
	case result := <-resultChan:
		if result.Error != nil {
			return nil, result.Error
		}
		return result.Data.([]Location), nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// UpdateDocumentAsync updates document asynchronously (background operation)
func (als *AsyncLanguageService) UpdateDocumentAsync(ctx context.Context, uri, text string, version int) error {
	resultChan := make(chan LSPResult, 1)

	req := &LSPRequest{
		ID:       "update",
		Type:     "update",
		Priority: PriorityLow, // Background operation
		Context:  ctx,
		Handler: func() (interface{}, error) {
			return nil, als.LanguageService.UpdateDocument(uri, text, version)
		},
		ResultChan: resultChan,
		Timestamp:  time.Now(),
		Timeout:    30 * time.Second,
	}

	err := als.queue.SubmitRequest(req)
	if err != nil {
		return err
	}

	select {
	case result := <-resultChan:
		return result.Error
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Convenience method to check queue health
func (als *AsyncLanguageService) GetQueueStats() RequestQueueStats {
	return als.queue.GetStats()
}

// Shutdown gracefully stops the async service
func (als *AsyncLanguageService) Shutdown() {
	als.queue.Stop()
}
