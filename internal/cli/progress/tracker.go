// ABOUTME: Progress tracking implementation for CLI operations
// ABOUTME: Provides concrete implementations of progress tracking interfaces

package progress

import (
	"fmt"
	"sync"
	"time"
)

// SimpleTracker provides a basic implementation of the Tracker interface
type SimpleTracker struct {
	mu        sync.RWMutex
	status    *Status
	callbacks []Callback
	running   bool
}

// NewTracker creates a new simple progress tracker
func NewTracker() *SimpleTracker {
	return &SimpleTracker{
		status:    &Status{},
		callbacks: make([]Callback, 0),
		running:   false,
	}
}

// Start begins tracking an operation
func (t *SimpleTracker) Start(operation string, total int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	t.status = &Status{
		Operation:          operation,
		Current:            0,
		Total:              total,
		Message:            fmt.Sprintf("Starting %s", operation),
		Percentage:         0.0,
		StartTime:          now,
		ElapsedTime:        0,
		EstimatedRemaining: 0,
		Stage:              "initializing",
	}
	t.running = true

	t.notifyCallbacks()
}

// Update reports progress on the current operation
func (t *SimpleTracker) Update(current int, message string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.running {
		return
	}

	now := time.Now()
	t.status.Current = current
	t.status.Message = message
	t.status.ElapsedTime = now.Sub(t.status.StartTime)

	if t.status.Total > 0 {
		t.status.Percentage = float64(current) / float64(t.status.Total) * 100.0

		// Calculate estimated remaining time
		if current > 0 {
			avgTimePerItem := t.status.ElapsedTime / time.Duration(current)
			remaining := t.status.Total - current
			t.status.EstimatedRemaining = avgTimePerItem * time.Duration(remaining)
		}
	}

	t.status.Stage = "processing"
	t.notifyCallbacks()
}

// Finish completes the operation with final result
func (t *SimpleTracker) Finish(result *Result) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.running {
		return
	}

	now := time.Now()
	t.status.Current = t.status.Total
	t.status.Percentage = 100.0
	t.status.ElapsedTime = now.Sub(t.status.StartTime)
	t.status.EstimatedRemaining = 0
	t.status.Stage = "completed"

	if result != nil {
		if result.Success {
			t.status.Message = fmt.Sprintf("Completed %s successfully", t.status.Operation)
		} else {
			t.status.Message = fmt.Sprintf("Failed %s: %s", t.status.Operation, result.Message)
		}
	} else {
		t.status.Message = fmt.Sprintf("Completed %s", t.status.Operation)
	}

	t.running = false
	t.notifyCallbacks()
}

// SetTotal updates the total count
func (t *SimpleTracker) SetTotal(total int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.status.Total = total
	if t.status.Total > 0 {
		t.status.Percentage = float64(t.status.Current) / float64(t.status.Total) * 100.0
	}

	t.notifyCallbacks()
}

// SetMessage updates the current status message
func (t *SimpleTracker) SetMessage(message string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.status.Message = message
	t.notifyCallbacks()
}

// IsRunning returns true if tracking is currently active
func (t *SimpleTracker) IsRunning() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.running
}

// GetProgress returns the current progress information
func (t *SimpleTracker) GetProgress() *Status {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Return a copy to prevent external modification
	statusCopy := *t.status
	return &statusCopy
}

// Subscribe adds a progress callback
func (t *SimpleTracker) Subscribe(callback Callback) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.callbacks = append(t.callbacks, callback)
}

// Unsubscribe removes a progress callback
func (t *SimpleTracker) Unsubscribe(callback Callback) {
	t.mu.Lock()
	defer t.mu.Unlock()

	for i, cb := range t.callbacks {
		// Compare function pointers (this is a simple approach)
		// In production, you might want a more robust callback management system
		if fmt.Sprintf("%p", cb) == fmt.Sprintf("%p", callback) {
			t.callbacks = append(t.callbacks[:i], t.callbacks[i+1:]...)
			break
		}
	}
}

// notifyCallbacks sends progress updates to all subscribers
func (t *SimpleTracker) notifyCallbacks() {
	statusCopy := *t.status
	for _, callback := range t.callbacks {
		go callback(&statusCopy) // Run callbacks asynchronously
	}
}

// SimpleParallelTracker provides parallel operation tracking
type SimpleParallelTracker struct {
	mu            sync.RWMutex
	operations    map[string]*OperationStatus
	status        *ParallelStatus
	callbacks     []ParallelCallback
	running       bool
	maxWorkers    int
	activeWorkers int
}

// NewParallelTracker creates a new parallel progress tracker
func NewParallelTracker(maxWorkers int) *SimpleParallelTracker {
	return &SimpleParallelTracker{
		operations:    make(map[string]*OperationStatus),
		status:        &ParallelStatus{Operations: make(map[string]*OperationStatus)},
		callbacks:     make([]ParallelCallback, 0),
		running:       false,
		maxWorkers:    maxWorkers,
		activeWorkers: 0,
	}
}

// StartParallel begins tracking multiple parallel operations
func (pt *SimpleParallelTracker) StartParallel(operations []string) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	now := time.Now()
	pt.operations = make(map[string]*OperationStatus)

	for _, opName := range operations {
		opStatus := &OperationStatus{
			ID:        fmt.Sprintf("op_%d_%s", now.UnixNano(), opName),
			Name:      opName,
			Status:    StatusPending,
			Message:   "Pending",
			Progress:  0.0,
			StartTime: time.Time{},
		}
		pt.operations[opStatus.ID] = opStatus
	}

	pt.status = &ParallelStatus{
		Operations:          make(map[string]*OperationStatus),
		TotalOperations:     len(operations),
		CompletedOperations: 0,
		FailedOperations:    0,
		RunningOperations:   0,
		OverallPercentage:   0.0,
		StartTime:           now,
		ElapsedTime:         0,
	}

	// Copy operations to status
	for id, op := range pt.operations {
		pt.status.Operations[id] = op
	}

	pt.running = true
	pt.notifyParallelCallbacks()
}

// UpdateOperation updates the status of a specific operation
func (pt *SimpleParallelTracker) UpdateOperation(id string, status OperationStatus, message string) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	if !pt.running {
		return
	}

	op, exists := pt.operations[id]
	if !exists {
		return
	}

	now := time.Now()

	// Update operation status
	if op.Status == StatusPending && status.Status == StatusRunning {
		op.StartTime = now
		pt.activeWorkers++
	}

	op.Status = status.Status
	op.Message = message
	op.Progress = status.Progress

	// Update overall status
	pt.updateOverallStatus()
	pt.notifyParallelCallbacks()
}

// FinishOperation completes a specific operation
func (pt *SimpleParallelTracker) FinishOperation(id string, result *Result) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	if !pt.running {
		return
	}

	op, exists := pt.operations[id]
	if !exists {
		return
	}

	now := time.Now()
	op.EndTime = now
	if !op.StartTime.IsZero() {
		op.Duration = now.Sub(op.StartTime)
		pt.activeWorkers--
	}

	if result != nil {
		op.Result = result
		if result.Success {
			op.Status = StatusCompleted
		} else {
			op.Status = StatusFailed
			op.Error = result.Error
		}
		op.Progress = 100.0 // Always 100% when finished, regardless of success
	} else {
		op.Status = StatusCompleted
		op.Progress = 100.0
	}

	pt.updateOverallStatus()
	pt.notifyParallelCallbacks()
}

// FinishAll completes all parallel operations
func (pt *SimpleParallelTracker) FinishAll(results []*Result) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	now := time.Now()
	pt.status.ElapsedTime = now.Sub(pt.status.StartTime)
	pt.running = false

	// Update any remaining operations
	i := 0
	for _, op := range pt.operations {
		if op.Status == StatusRunning || op.Status == StatusPending {
			if i < len(results) {
				pt.finishOperationInternal(op, results[i])
				i++
			} else {
				op.Status = StatusCompleted
				op.Progress = 100.0
				op.EndTime = now
			}
		}
	}

	pt.updateOverallStatus()
	pt.notifyParallelCallbacks()
}

// GetOperationStatus returns status for a specific operation
func (pt *SimpleParallelTracker) GetOperationStatus(id string) *OperationStatus {
	pt.mu.RLock()
	defer pt.mu.RUnlock()

	if op, exists := pt.operations[id]; exists {
		// Return a copy
		opCopy := *op
		return &opCopy
	}
	return nil
}

// GetOverallProgress returns aggregate progress information
func (pt *SimpleParallelTracker) GetOverallProgress() *ParallelStatus {
	pt.mu.RLock()
	defer pt.mu.RUnlock()

	// Return a copy
	statusCopy := *pt.status
	statusCopy.Operations = make(map[string]*OperationStatus)
	for id, op := range pt.status.Operations {
		opCopy := *op
		statusCopy.Operations[id] = &opCopy
	}
	return &statusCopy
}

// Subscribe adds a parallel progress callback
func (pt *SimpleParallelTracker) Subscribe(callback ParallelCallback) {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	pt.callbacks = append(pt.callbacks, callback)
}

// Unsubscribe removes a parallel progress callback
func (pt *SimpleParallelTracker) Unsubscribe(callback ParallelCallback) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	for i, cb := range pt.callbacks {
		if fmt.Sprintf("%p", cb) == fmt.Sprintf("%p", callback) {
			pt.callbacks = append(pt.callbacks[:i], pt.callbacks[i+1:]...)
			break
		}
	}
}

// updateOverallStatus recalculates the overall progress status
func (pt *SimpleParallelTracker) updateOverallStatus() {
	now := time.Now()
	pt.status.ElapsedTime = now.Sub(pt.status.StartTime)

	completed := 0
	failed := 0
	running := 0
	totalProgress := 0.0

	for _, op := range pt.operations {
		switch op.Status {
		case StatusCompleted:
			completed++
		case StatusFailed:
			failed++
		case StatusRunning:
			running++
		}
		totalProgress += op.Progress
	}

	pt.status.CompletedOperations = completed
	pt.status.FailedOperations = failed
	pt.status.RunningOperations = running

	if pt.status.TotalOperations > 0 {
		pt.status.OverallPercentage = totalProgress / float64(pt.status.TotalOperations)
	}

	// Estimate remaining time
	if completed > 0 && running > 0 {
		avgTimePerOp := pt.status.ElapsedTime / time.Duration(completed)
		remaining := pt.status.TotalOperations - completed - failed
		pt.status.EstimatedRemaining = avgTimePerOp * time.Duration(remaining)
	}

	// Update operations in status
	for id, op := range pt.operations {
		pt.status.Operations[id] = op
	}
}

// finishOperationInternal completes an operation (internal method, assumes lock is held)
func (pt *SimpleParallelTracker) finishOperationInternal(op *OperationStatus, result *Result) {
	now := time.Now()
	op.EndTime = now
	if !op.StartTime.IsZero() {
		op.Duration = now.Sub(op.StartTime)
	}

	if result != nil {
		op.Result = result
		if result.Success {
			op.Status = StatusCompleted
		} else {
			op.Status = StatusFailed
			op.Error = result.Error
		}
		op.Progress = 100.0 // Always 100% when finished, regardless of success
	} else {
		op.Status = StatusCompleted
		op.Progress = 100.0
	}
}

// notifyParallelCallbacks sends parallel progress updates to all subscribers
func (pt *SimpleParallelTracker) notifyParallelCallbacks() {
	// Create a copy without acquiring another lock (lock is already held)
	statusCopy := *pt.status
	statusCopy.Operations = make(map[string]*OperationStatus)
	for id, op := range pt.status.Operations {
		opCopy := *op
		statusCopy.Operations[id] = &opCopy
	}

	for _, callback := range pt.callbacks {
		go callback(&statusCopy) // Run callbacks asynchronously
	}
}

// NullTracker provides a no-op implementation for when progress tracking is disabled
type NullTracker struct{}

// NewNullTracker creates a tracker that does nothing
func NewNullTracker() *NullTracker {
	return &NullTracker{}
}

// Start does nothing
func (nt *NullTracker) Start(operation string, total int) {}

// Update does nothing
func (nt *NullTracker) Update(current int, message string) {}

// Finish does nothing
func (nt *NullTracker) Finish(result *Result) {}

// SetTotal does nothing
func (nt *NullTracker) SetTotal(total int) {}

// SetMessage does nothing
func (nt *NullTracker) SetMessage(message string) {}

// IsRunning always returns false
func (nt *NullTracker) IsRunning() bool { return false }

// GetProgress returns an empty status
func (nt *NullTracker) GetProgress() *Status {
	return &Status{}
}
