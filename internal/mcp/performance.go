// ABOUTME: Performance optimization features for MCP server
// ABOUTME: Provides connection pooling, request queuing, and resource management

package mcp

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// PerformanceConfig configures performance optimization features
type PerformanceConfig struct {
	MaxConcurrentRequests int           `json:"max_concurrent_requests"`
	RequestTimeout        time.Duration `json:"request_timeout"`
	QueueSize             int           `json:"queue_size"`
	ConnectionPoolSize    int           `json:"connection_pool_size"`
	IdleConnTimeout       time.Duration `json:"idle_conn_timeout"`
	MaxIdleConns          int           `json:"max_idle_conns"`
	MaxIdleConnsPerHost   int           `json:"max_idle_conns_per_host"`
	KeepAlive             time.Duration `json:"keep_alive"`
	EnableMetrics         bool          `json:"enable_metrics"`
}

// DefaultPerformanceConfig returns sensible defaults
func DefaultPerformanceConfig() PerformanceConfig {
	return PerformanceConfig{
		MaxConcurrentRequests: 100,
		RequestTimeout:        30 * time.Second,
		QueueSize:             1000,
		ConnectionPoolSize:    50,
		IdleConnTimeout:       90 * time.Second,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		KeepAlive:             30 * time.Second,
		EnableMetrics:         true,
	}
}

// RequestQueue manages incoming requests with backpressure
type RequestQueue struct {
	semaphore         chan struct{}
	queue             chan *QueuedRequest
	activeRequests    int64
	processedRequests int64
	rejectedRequests  int64
	avgProcessingTime time.Duration
	mu                sync.RWMutex
	metrics           *PerformanceMetrics
}

// QueuedRequest represents a queued request
type QueuedRequest struct {
	ID        string
	Context   context.Context
	Handler   func(context.Context) error
	StartTime time.Time
	Result    chan error
}

// NewRequestQueue creates a new request queue
func NewRequestQueue(config PerformanceConfig, metrics *PerformanceMetrics) *RequestQueue {
	return &RequestQueue{
		semaphore: make(chan struct{}, config.MaxConcurrentRequests),
		queue:     make(chan *QueuedRequest, config.QueueSize),
		metrics:   metrics,
	}
}

// Submit submits a request to the queue
func (rq *RequestQueue) Submit(ctx context.Context, id string, handler func(context.Context) error) error {
	request := &QueuedRequest{
		ID:        id,
		Context:   ctx,
		Handler:   handler,
		StartTime: time.Now(),
		Result:    make(chan error, 1),
	}

	select {
	case rq.queue <- request:
		// Request queued successfully
		select {
		case err := <-request.Result:
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	case <-ctx.Done():
		return ctx.Err()
	default:
		// Queue is full
		atomic.AddInt64(&rq.rejectedRequests, 1)
		if rq.metrics != nil {
			rq.metrics.RecordRejectedRequest()
		}
		return fmt.Errorf("request queue is full")
	}
}

// Start starts the request queue processor
func (rq *RequestQueue) Start(ctx context.Context) {
	go rq.processRequests(ctx)
}

// processRequests processes queued requests
func (rq *RequestQueue) processRequests(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case request := <-rq.queue:
			go rq.handleRequest(request)
		}
	}
}

// handleRequest handles a single request
func (rq *RequestQueue) handleRequest(request *QueuedRequest) {
	// Acquire semaphore
	rq.semaphore <- struct{}{}
	defer func() { <-rq.semaphore }()

	atomic.AddInt64(&rq.activeRequests, 1)
	defer atomic.AddInt64(&rq.activeRequests, -1)

	start := time.Now()
	err := request.Handler(request.Context)
	duration := time.Since(start)

	// Update metrics
	atomic.AddInt64(&rq.processedRequests, 1)
	rq.updateAvgProcessingTime(duration)

	if rq.metrics != nil {
		rq.metrics.RecordRequestDuration(duration)
		if err != nil {
			rq.metrics.RecordError()
		}
	}

	// Send result
	select {
	case request.Result <- err:
	default:
		// Channel closed or context cancelled
	}
}

// updateAvgProcessingTime updates the average processing time
func (rq *RequestQueue) updateAvgProcessingTime(duration time.Duration) {
	rq.mu.Lock()
	defer rq.mu.Unlock()

	if rq.avgProcessingTime == 0 {
		rq.avgProcessingTime = duration
	} else {
		// Simple exponential moving average
		rq.avgProcessingTime = time.Duration(float64(rq.avgProcessingTime)*0.9 + float64(duration)*0.1)
	}
}

// GetStats returns queue statistics
func (rq *RequestQueue) GetStats() map[string]interface{} {
	rq.mu.RLock()
	defer rq.mu.RUnlock()

	return map[string]interface{}{
		"active_requests":     atomic.LoadInt64(&rq.activeRequests),
		"processed_requests":  atomic.LoadInt64(&rq.processedRequests),
		"rejected_requests":   atomic.LoadInt64(&rq.rejectedRequests),
		"avg_processing_time": rq.avgProcessingTime.Milliseconds(),
		"queue_length":        len(rq.queue),
		"queue_capacity":      cap(rq.queue),
		"semaphore_available": cap(rq.semaphore) - len(rq.semaphore),
	}
}

// ConnectionPool manages HTTP connections for external services
type ConnectionPool struct {
	client *http.Client
	config PerformanceConfig
}

// NewConnectionPool creates a new connection pool
func NewConnectionPool(config PerformanceConfig) *ConnectionPool {
	transport := &http.Transport{
		MaxIdleConns:        config.MaxIdleConns,
		MaxIdleConnsPerHost: config.MaxIdleConnsPerHost,
		IdleConnTimeout:     config.IdleConnTimeout,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   config.RequestTimeout,
	}

	return &ConnectionPool{
		client: client,
		config: config,
	}
}

// GetClient returns the HTTP client
func (cp *ConnectionPool) GetClient() *http.Client {
	return cp.client
}

// GetStats returns connection pool statistics
func (cp *ConnectionPool) GetStats() map[string]interface{} {
	// Note: Go's http.Transport doesn't expose detailed connection stats
	// In production, you might want to use a library like go-connpool
	return map[string]interface{}{
		"max_idle_conns":          cp.config.MaxIdleConns,
		"max_idle_conns_per_host": cp.config.MaxIdleConnsPerHost,
		"idle_conn_timeout":       cp.config.IdleConnTimeout.Seconds(),
		"keep_alive":              cp.config.KeepAlive.Seconds(),
	}
}

// ResourceMonitor monitors system resources
type ResourceMonitor struct {
	mu             sync.RWMutex
	memoryUsage    uint64
	memoryLimit    uint64
	cpuUsage       float64
	goroutineCount int
	enableLimits   bool
	lastCheck      time.Time
	checkInterval  time.Duration
}

// NewResourceMonitor creates a new resource monitor
func NewResourceMonitor(memoryLimitMB uint64, enableLimits bool) *ResourceMonitor {
	return &ResourceMonitor{
		memoryLimit:   memoryLimitMB * 1024 * 1024, // Convert MB to bytes
		enableLimits:  enableLimits,
		checkInterval: 10 * time.Second,
	}
}

// Start starts the resource monitoring
func (rm *ResourceMonitor) Start(ctx context.Context) {
	ticker := time.NewTicker(rm.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			rm.updateStats()
		}
	}
}

// updateStats updates resource statistics
func (rm *ResourceMonitor) updateStats() {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// Get memory stats
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	rm.memoryUsage = memStats.Alloc
	rm.goroutineCount = runtime.NumGoroutine()
	rm.lastCheck = time.Now()

	// CPU usage would require additional implementation
	// For now, we'll just track memory and goroutines
}

// CheckResources checks if resource limits are exceeded
func (rm *ResourceMonitor) CheckResources() error {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	if !rm.enableLimits {
		return nil
	}

	if rm.memoryLimit > 0 && rm.memoryUsage > rm.memoryLimit {
		return fmt.Errorf("memory usage (%d bytes) exceeds limit (%d bytes)",
			rm.memoryUsage, rm.memoryLimit)
	}

	// Add more resource checks as needed
	return nil
}

// GetStats returns resource statistics
func (rm *ResourceMonitor) GetStats() map[string]interface{} {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	return map[string]interface{}{
		"memory_usage_bytes": rm.memoryUsage,
		"memory_usage_mb":    rm.memoryUsage / (1024 * 1024),
		"memory_limit_bytes": rm.memoryLimit,
		"memory_limit_mb":    rm.memoryLimit / (1024 * 1024),
		"goroutine_count":    rm.goroutineCount,
		"cpu_usage_percent":  rm.cpuUsage,
		"last_check":         rm.lastCheck,
		"enable_limits":      rm.enableLimits,
	}
}

// PerformanceMetrics tracks various performance metrics
type PerformanceMetrics struct {
	mu                sync.RWMutex
	requestCount      int64
	errorCount        int64
	rejectedCount     int64
	totalResponseTime time.Duration
	minResponseTime   time.Duration
	maxResponseTime   time.Duration
	lastRequestTime   time.Time
	startTime         time.Time
}

// NewPerformanceMetrics creates new performance metrics
func NewPerformanceMetrics() *PerformanceMetrics {
	return &PerformanceMetrics{
		startTime:       time.Now(),
		minResponseTime: time.Hour, // Start with a high value
	}
}

// RecordRequestDuration records a request duration
func (pm *PerformanceMetrics) RecordRequestDuration(duration time.Duration) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.requestCount++
	pm.totalResponseTime += duration
	pm.lastRequestTime = time.Now()

	if duration < pm.minResponseTime {
		pm.minResponseTime = duration
	}
	if duration > pm.maxResponseTime {
		pm.maxResponseTime = duration
	}
}

// RecordError records an error
func (pm *PerformanceMetrics) RecordError() {
	atomic.AddInt64(&pm.errorCount, 1)
}

// RecordRejectedRequest records a rejected request
func (pm *PerformanceMetrics) RecordRejectedRequest() {
	atomic.AddInt64(&pm.rejectedCount, 1)
}

// GetMetrics returns current metrics
func (pm *PerformanceMetrics) GetMetrics() map[string]interface{} {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var avgResponseTime time.Duration
	if pm.requestCount > 0 {
		avgResponseTime = pm.totalResponseTime / time.Duration(pm.requestCount)
	}

	return map[string]interface{}{
		"request_count":        pm.requestCount,
		"error_count":          atomic.LoadInt64(&pm.errorCount),
		"rejected_count":       atomic.LoadInt64(&pm.rejectedCount),
		"avg_response_time_ms": avgResponseTime.Milliseconds(),
		"min_response_time_ms": pm.minResponseTime.Milliseconds(),
		"max_response_time_ms": pm.maxResponseTime.Milliseconds(),
		"last_request_time":    pm.lastRequestTime,
		"uptime_seconds":       time.Since(pm.startTime).Seconds(),
		"requests_per_second":  float64(pm.requestCount) / time.Since(pm.startTime).Seconds(),
	}
}

// PerformanceManager combines all performance optimization features
type PerformanceManager struct {
	config          PerformanceConfig
	requestQueue    *RequestQueue
	connectionPool  *ConnectionPool
	resourceMonitor *ResourceMonitor
	metrics         *PerformanceMetrics
	circuitManager  *CircuitBreakerManager
}

// NewPerformanceManager creates a new performance manager
func NewPerformanceManager(config PerformanceConfig) *PerformanceManager {
	metrics := NewPerformanceMetrics()

	return &PerformanceManager{
		config:          config,
		requestQueue:    NewRequestQueue(config, metrics),
		connectionPool:  NewConnectionPool(config),
		resourceMonitor: NewResourceMonitor(100, true), // 100MB limit by default
		metrics:         metrics,
		circuitManager:  NewCircuitBreakerManager(),
	}
}

// Start starts all performance optimization features
func (pm *PerformanceManager) Start(ctx context.Context) {
	pm.requestQueue.Start(ctx)
	go pm.resourceMonitor.Start(ctx)
}

// ExecuteWithOptimization executes a request with all optimizations
func (pm *PerformanceManager) ExecuteWithOptimization(ctx context.Context, requestID string,
	handler func(context.Context) error) error {

	// Check resource limits
	if err := pm.resourceMonitor.CheckResources(); err != nil {
		pm.metrics.RecordError()
		return fmt.Errorf("resource limit exceeded: %w", err)
	}

	// Submit to request queue
	return pm.requestQueue.Submit(ctx, requestID, handler)
}

// GetAllStats returns comprehensive performance statistics
func (pm *PerformanceManager) GetAllStats() map[string]interface{} {
	return map[string]interface{}{
		"request_queue":    pm.requestQueue.GetStats(),
		"connection_pool":  pm.connectionPool.GetStats(),
		"resource_monitor": pm.resourceMonitor.GetStats(),
		"metrics":          pm.metrics.GetMetrics(),
		"circuit_breakers": pm.circuitManager.GetBreakersStatus(),
	}
}

// GetRequestQueue returns the request queue
func (pm *PerformanceManager) GetRequestQueue() *RequestQueue {
	return pm.requestQueue
}

// GetConnectionPool returns the connection pool
func (pm *PerformanceManager) GetConnectionPool() *ConnectionPool {
	return pm.connectionPool
}

// GetResourceMonitor returns the resource monitor
func (pm *PerformanceManager) GetResourceMonitor() *ResourceMonitor {
	return pm.resourceMonitor
}

// GetMetrics returns the performance metrics
func (pm *PerformanceManager) GetMetrics() *PerformanceMetrics {
	return pm.metrics
}

// GetCircuitManager returns the circuit breaker manager
func (pm *PerformanceManager) GetCircuitManager() *CircuitBreakerManager {
	return pm.circuitManager
}
