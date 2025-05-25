// ABOUTME: Implements a high-performance worker pool for parallel task execution
// ABOUTME: Provides configurable concurrency with automatic load balancing and backpressure

package parallel

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/log"
)

// Task represents a unit of work to be executed
type Task interface {
	Execute(ctx context.Context) error
	Priority() int
	ID() string
}

// TaskResult contains the result of a task execution
type TaskResult struct {
	TaskID    string
	Error     error
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration
}

// WorkerPool manages a pool of workers for parallel task execution
type WorkerPool struct {
	config      *PoolConfig
	workers     []*Worker
	taskQueue   chan Task
	resultQueue chan *TaskResult
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
	stats       *PoolStats
	logger      *log.Logger
	started     atomic.Bool
	stopped     atomic.Bool
}

// PoolConfig contains configuration for the worker pool
type PoolConfig struct {
	NumWorkers         int           // Number of worker goroutines
	QueueSize          int           // Size of the task queue
	MaxRetries         int           // Maximum retry attempts for failed tasks
	RetryDelay         time.Duration // Delay between retries
	IdleTimeout        time.Duration // Worker idle timeout
	EnableProfiling    bool          // Enable performance profiling
	EnableAdaptive     bool          // Enable adaptive worker scaling
	MinWorkers         int           // Minimum workers for adaptive scaling
	MaxWorkers         int           // Maximum workers for adaptive scaling
	ScaleUpThreshold   float64       // Queue utilization threshold for scaling up
	ScaleDownThreshold float64       // Queue utilization threshold for scaling down
}

// PoolStats tracks worker pool statistics
type PoolStats struct {
	TasksSubmitted atomic.Int64
	TasksCompleted atomic.Int64
	TasksFailed    atomic.Int64
	TasksRetried   atomic.Int64
	ActiveWorkers  atomic.Int32
	QueueDepth     atomic.Int32
	TotalDuration  atomic.Int64 // nanoseconds
	mu             sync.RWMutex
	workerStats    map[int]*WorkerStats
}

// WorkerStats tracks individual worker statistics
type WorkerStats struct {
	ID             int
	TasksProcessed int64
	TotalDuration  time.Duration
	LastTaskTime   time.Time
	Status         string
}

// DefaultPoolConfig returns a default configuration
func DefaultPoolConfig() *PoolConfig {
	numCPU := runtime.NumCPU()
	return &PoolConfig{
		NumWorkers:         numCPU,
		QueueSize:          numCPU * 100,
		MaxRetries:         3,
		RetryDelay:         time.Second,
		IdleTimeout:        30 * time.Second,
		EnableProfiling:    false,
		EnableAdaptive:     true,
		MinWorkers:         2,
		MaxWorkers:         numCPU * 2,
		ScaleUpThreshold:   0.8,
		ScaleDownThreshold: 0.2,
	}
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(config *PoolConfig, logger *log.Logger) *WorkerPool {
	if config == nil {
		config = DefaultPoolConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	pool := &WorkerPool{
		config:      config,
		workers:     make([]*Worker, 0, config.MaxWorkers),
		taskQueue:   make(chan Task, config.QueueSize),
		resultQueue: make(chan *TaskResult, config.QueueSize),
		ctx:         ctx,
		cancel:      cancel,
		stats: &PoolStats{
			workerStats: make(map[int]*WorkerStats),
		},
		logger: logger,
	}

	return pool
}

// Start starts the worker pool
func (p *WorkerPool) Start() error {
	if p.started.Load() {
		return errors.NewSystemError("001", "worker pool already started", nil)
	}

	p.started.Store(true)
	p.logger.Infof("Starting worker pool with %d workers", p.config.NumWorkers)

	// Start initial workers
	for i := 0; i < p.config.NumWorkers; i++ {
		p.startWorker(i)
	}

	// Start adaptive scaling if enabled
	if p.config.EnableAdaptive {
		go p.adaptiveScaling()
	}

	// Start statistics collector
	if p.config.EnableProfiling {
		go p.collectStats()
	}

	return nil
}

// Stop gracefully stops the worker pool
func (p *WorkerPool) Stop() error {
	if p.stopped.Load() {
		return errors.NewSystemError("002", "worker pool already stopped", nil)
	}

	p.stopped.Store(true)
	p.logger.Infof("Stopping worker pool")

	// Cancel context to signal workers to stop
	p.cancel()

	// Close task queue to prevent new submissions
	close(p.taskQueue)

	// Wait for all workers to finish
	p.wg.Wait()

	// Close result queue
	close(p.resultQueue)

	p.logger.Infof("Worker pool stopped")
	return nil
}

// Submit submits a task to the worker pool
func (p *WorkerPool) Submit(task Task) error {
	if p.stopped.Load() {
		return errors.NewSystemError("003", "worker pool is stopped", nil)
	}

	if !p.started.Load() {
		return errors.NewSystemError("004", "worker pool not started", nil)
	}

	select {
	case p.taskQueue <- task:
		p.stats.TasksSubmitted.Add(1)
		p.stats.QueueDepth.Store(int32(len(p.taskQueue)))
		return nil
	case <-p.ctx.Done():
		return errors.NewSystemError("005", "worker pool shutting down", nil)
	default:
		return errors.NewSystemError("006", "task queue full", nil)
	}
}

// SubmitBatch submits multiple tasks to the worker pool
func (p *WorkerPool) SubmitBatch(tasks []Task) error {
	for _, task := range tasks {
		if err := p.Submit(task); err != nil {
			return errors.Wrap(err, "PSC", "parallel", "001", "failed to submit task batch")
		}
	}
	return nil
}

// Results returns the result channel
func (p *WorkerPool) Results() <-chan *TaskResult {
	return p.resultQueue
}

// WaitForCompletion waits for all submitted tasks to complete
func (p *WorkerPool) WaitForCompletion() {
	for p.stats.TasksSubmitted.Load() > (p.stats.TasksCompleted.Load() + p.stats.TasksFailed.Load()) {
		time.Sleep(10 * time.Millisecond)
	}
}

// GetStats returns current pool statistics
func (p *WorkerPool) GetStats() *PoolStats {
	p.stats.mu.RLock()
	defer p.stats.mu.RUnlock()

	stats := PoolStats{
		workerStats: make(map[int]*WorkerStats),
	}

	stats.TasksSubmitted.Store(p.stats.TasksSubmitted.Load())
	stats.TasksCompleted.Store(p.stats.TasksCompleted.Load())
	stats.TasksFailed.Store(p.stats.TasksFailed.Load())
	stats.TasksRetried.Store(p.stats.TasksRetried.Load())
	stats.ActiveWorkers.Store(p.stats.ActiveWorkers.Load())
	stats.QueueDepth.Store(p.stats.QueueDepth.Load())
	stats.TotalDuration.Store(p.stats.TotalDuration.Load())

	for id, ws := range p.stats.workerStats {
		stats.workerStats[id] = &WorkerStats{
			ID:             ws.ID,
			TasksProcessed: ws.TasksProcessed,
			TotalDuration:  ws.TotalDuration,
			LastTaskTime:   ws.LastTaskTime,
			Status:         ws.Status,
		}
	}

	return &stats
}

// startWorker starts a new worker
func (p *WorkerPool) startWorker(id int) {
	worker := &Worker{
		id:          id,
		pool:        p,
		taskQueue:   p.taskQueue,
		resultQueue: p.resultQueue,
		logger:      p.logger,
		stats: &WorkerStats{
			ID:     id,
			Status: "idle",
		},
	}

	p.stats.mu.Lock()
	p.workers = append(p.workers, worker)
	p.stats.workerStats[id] = worker.stats
	p.stats.mu.Unlock()

	p.wg.Add(1)
	go worker.run(p.ctx, &p.wg)
	p.stats.ActiveWorkers.Add(1)
}

// stopWorker stops a specific worker
func (p *WorkerPool) stopWorker(id int) {
	p.stats.mu.Lock()
	defer p.stats.mu.Unlock()

	for i, worker := range p.workers {
		if worker.id == id {
			worker.stop()
			p.workers = append(p.workers[:i], p.workers[i+1:]...)
			delete(p.stats.workerStats, id)
			p.stats.ActiveWorkers.Add(-1)
			break
		}
	}
}

// adaptiveScaling dynamically adjusts the number of workers
func (p *WorkerPool) adaptiveScaling() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.scaleWorkers()
		case <-p.ctx.Done():
			return
		}
	}
}

// scaleWorkers adjusts the number of workers based on queue utilization
func (p *WorkerPool) scaleWorkers() {
	queueDepth := len(p.taskQueue)
	queueUtilization := float64(queueDepth) / float64(p.config.QueueSize)
	currentWorkers := int(p.stats.ActiveWorkers.Load())

	if queueUtilization > p.config.ScaleUpThreshold && currentWorkers < p.config.MaxWorkers {
		// Scale up
		newWorkers := min(currentWorkers*2, p.config.MaxWorkers) - currentWorkers
		p.logger.Infof("Scaling up: adding %d workers (queue utilization: %.2f%%)",
			newWorkers, queueUtilization*100)

		for i := 0; i < newWorkers; i++ {
			p.startWorker(currentWorkers + i)
		}
	} else if queueUtilization < p.config.ScaleDownThreshold && currentWorkers > p.config.MinWorkers {
		// Scale down
		removeWorkers := max(1, (currentWorkers-p.config.MinWorkers)/2)
		p.logger.Infof("Scaling down: removing %d workers (queue utilization: %.2f%%)",
			removeWorkers, queueUtilization*100)

		// Remove idle workers first
		p.stats.mu.RLock()
		idleWorkers := make([]int, 0)
		for id, stats := range p.stats.workerStats {
			if stats.Status == "idle" && len(idleWorkers) < removeWorkers {
				idleWorkers = append(idleWorkers, id)
			}
		}
		p.stats.mu.RUnlock()

		for _, id := range idleWorkers {
			p.stopWorker(id)
		}
	}
}

// collectStats periodically collects and logs statistics
func (p *WorkerPool) collectStats() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			stats := p.GetStats()
			avgDuration := time.Duration(0)
			if stats.TasksCompleted.Load() > 0 {
				avgDuration = time.Duration(stats.TotalDuration.Load() / stats.TasksCompleted.Load())
			}

			p.logger.Infof("Worker Pool Stats - Submitted: %d, Completed: %d, Failed: %d, Avg Duration: %v, Queue Depth: %d, Active Workers: %d",
				stats.TasksSubmitted.Load(),
				stats.TasksCompleted.Load(),
				stats.TasksFailed.Load(),
				avgDuration,
				stats.QueueDepth.Load(),
				stats.ActiveWorkers.Load())
		case <-p.ctx.Done():
			return
		}
	}
}

// Worker represents a worker in the pool
type Worker struct {
	id          int
	pool        *WorkerPool
	taskQueue   <-chan Task
	resultQueue chan<- *TaskResult
	logger      *log.Logger
	stats       *WorkerStats
	stopCh      chan struct{}
}

// run executes the worker's main loop
func (w *Worker) run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	w.stopCh = make(chan struct{})

	idleTimer := time.NewTimer(w.pool.config.IdleTimeout)
	defer idleTimer.Stop()

	for {
		select {
		case task, ok := <-w.taskQueue:
			if !ok {
				return // Queue closed
			}

			idleTimer.Stop()
			w.executeTask(ctx, task)
			idleTimer.Reset(w.pool.config.IdleTimeout)

		case <-idleTimer.C:
			// Worker idle for too long
			if w.pool.config.EnableAdaptive && int(w.pool.stats.ActiveWorkers.Load()) > w.pool.config.MinWorkers {
				w.logger.Debugf("Worker %d idle timeout", w.id)
				return
			}
			idleTimer.Reset(w.pool.config.IdleTimeout)

		case <-w.stopCh:
			return

		case <-ctx.Done():
			return
		}
	}
}

// stop signals the worker to stop
func (w *Worker) stop() {
	close(w.stopCh)
}

// executeTask executes a single task with retry logic
func (w *Worker) executeTask(ctx context.Context, task Task) {
	w.updateStatus("busy")
	defer w.updateStatus("idle")

	result := &TaskResult{
		TaskID:    task.ID(),
		StartTime: time.Now(),
	}

	var err error
	for attempt := 0; attempt <= w.pool.config.MaxRetries; attempt++ {
		if attempt > 0 {
			w.pool.stats.TasksRetried.Add(1)
			time.Sleep(w.pool.config.RetryDelay * time.Duration(attempt))
		}

		err = task.Execute(ctx)
		if err == nil {
			break
		}

		if attempt < w.pool.config.MaxRetries {
			w.logger.Warningf("Task %s failed (attempt %d/%d): %v",
				task.ID(), attempt+1, w.pool.config.MaxRetries+1, err)
		}
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.Error = err

	// Update statistics
	w.stats.TasksProcessed++
	w.stats.TotalDuration += result.Duration
	w.stats.LastTaskTime = result.EndTime

	w.pool.stats.TotalDuration.Add(int64(result.Duration))
	if err != nil {
		w.pool.stats.TasksFailed.Add(1)
	} else {
		w.pool.stats.TasksCompleted.Add(1)
	}

	// Send result
	select {
	case w.resultQueue <- result:
	case <-ctx.Done():
	}

	w.pool.stats.QueueDepth.Store(int32(len(w.pool.taskQueue)))
}

// updateStatus updates the worker's status
func (w *Worker) updateStatus(status string) {
	w.pool.stats.mu.Lock()
	defer w.pool.stats.mu.Unlock()

	if ws, exists := w.pool.stats.workerStats[w.id]; exists {
		ws.Status = status
	}
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
