// ABOUTME: Progress tracking integration utilities for module operations
// ABOUTME: Provides adapters and utilities for standardized progress reporting

package progress

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// OperationTracker provides enhanced progress tracking for complex operations
type OperationTracker struct {
	tracker     Tracker
	context     context.Context
	cancel      context.CancelFunc
	subTrackers map[string]Tracker
	mu          sync.RWMutex
}

// NewOperationTracker creates a new operation tracker with context support
func NewOperationTracker(ctx context.Context, tracker Tracker) *OperationTracker {
	ctxWithCancel, cancel := context.WithCancel(ctx)
	return &OperationTracker{
		tracker:     tracker,
		context:     ctxWithCancel,
		cancel:      cancel,
		subTrackers: make(map[string]Tracker),
	}
}

// StartOperation begins tracking a top-level operation
func (ot *OperationTracker) StartOperation(operation string, total int) {
	ot.tracker.Start(operation, total)
}

// UpdateOperation updates the progress of the current operation
func (ot *OperationTracker) UpdateOperation(current int, message string) {
	select {
	case <-ot.context.Done():
		return // Operation was cancelled
	default:
		ot.tracker.Update(current, message)
	}
}

// FinishOperation completes the operation
func (ot *OperationTracker) FinishOperation(result *Result) {
	ot.tracker.Finish(result)
	ot.cancel() // Cancel context when operation finishes
}

// CreateSubTracker creates a sub-tracker for nested operations
func (ot *OperationTracker) CreateSubTracker(name string) Tracker {
	ot.mu.Lock()
	defer ot.mu.Unlock()

	subTracker := NewTracker()
	ot.subTrackers[name] = subTracker

	// Subscribe to sub-tracker updates to propagate to main tracker
	subTracker.Subscribe(func(status *Status) {
		// Update main tracker based on sub-tracker progress
		ot.UpdateOperation(int(status.Percentage), fmt.Sprintf("%s: %s", name, status.Message))
	})

	return subTracker
}

// GetSubTracker retrieves a sub-tracker by name
func (ot *OperationTracker) GetSubTracker(name string) Tracker {
	ot.mu.RLock()
	defer ot.mu.RUnlock()
	return ot.subTrackers[name]
}

// Cancel cancels the operation and all sub-trackers
func (ot *OperationTracker) Cancel() {
	ot.cancel()
}

// Context returns the operation context
func (ot *OperationTracker) Context() context.Context {
	return ot.context
}

// CompositeTracker manages multiple parallel operations with aggregated progress
type CompositeTracker struct {
	parallelTracker ParallelTracker
	operations      map[string]*OperationInfo
	mu              sync.RWMutex
	callbacks       []CompositeCallback
}

// OperationInfo contains information about a tracked operation
type OperationInfo struct {
	Name       string
	Total      int
	Current    int
	StartTime  time.Time
	LastUpdate time.Time
	Stage      string
	Tracker    Tracker
}

// CompositeCallback is called when composite progress updates
type CompositeCallback func(overall *CompositeStatus)

// CompositeStatus represents the status of composite operations
type CompositeStatus struct {
	TotalOperations     int                       `json:"total_operations"`
	ActiveOperations    int                       `json:"active_operations"`
	CompletedOperations int                       `json:"completed_operations"`
	FailedOperations    int                       `json:"failed_operations"`
	OverallProgress     float64                   `json:"overall_progress"`
	Operations          map[string]*OperationInfo `json:"operations"`
	StartTime           time.Time                 `json:"start_time"`
	ElapsedTime         time.Duration             `json:"elapsed_time"`
	EstimatedRemaining  time.Duration             `json:"estimated_remaining"`
}

// NewCompositeTracker creates a new composite tracker
func NewCompositeTracker(maxWorkers int) *CompositeTracker {
	return &CompositeTracker{
		parallelTracker: NewParallelTracker(maxWorkers),
		operations:      make(map[string]*OperationInfo),
		callbacks:       make([]CompositeCallback, 0),
	}
}

// AddOperation adds an operation to track
func (ct *CompositeTracker) AddOperation(name string, total int) string {
	operationID := fmt.Sprintf("comp_%d_%s", time.Now().UnixNano(), name)
	tracker := NewTracker()

	// Subscribe to individual tracker updates
	tracker.Subscribe(func(status *Status) {
		ct.updateOperationStatus(operationID, status)
	})

	ct.mu.Lock()
	ct.operations[operationID] = &OperationInfo{
		Name:       name,
		Total:      total,
		Current:    0,
		StartTime:  time.Now(),
		LastUpdate: time.Now(),
		Stage:      "pending",
		Tracker:    tracker,
	}
	ct.mu.Unlock()

	// Notify callbacks outside of lock
	go ct.notifyCallbacks()
	return operationID
}

// StartOperation starts tracking a specific operation
func (ct *CompositeTracker) StartOperation(operationID string) {
	ct.mu.RLock()
	op, exists := ct.operations[operationID]
	ct.mu.RUnlock()

	if exists {
		op.Tracker.Start(op.Name, op.Total)
	}
}

// UpdateOperation updates a specific operation's progress
func (ct *CompositeTracker) UpdateOperation(operationID string, current int, message string) {
	ct.mu.RLock()
	op, exists := ct.operations[operationID]
	ct.mu.RUnlock()

	if exists {
		op.Tracker.Update(current, message)
	}
}

// FinishOperation completes a specific operation
func (ct *CompositeTracker) FinishOperation(operationID string, result *Result) {
	ct.mu.RLock()
	op, exists := ct.operations[operationID]
	ct.mu.RUnlock()

	if exists {
		op.Tracker.Finish(result)
	}
}

// GetCompositeStatus returns the current composite status
func (ct *CompositeTracker) GetCompositeStatus() *CompositeStatus {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	now := time.Now()
	status := &CompositeStatus{
		TotalOperations:     len(ct.operations),
		ActiveOperations:    0,
		CompletedOperations: 0,
		FailedOperations:    0,
		OverallProgress:     0.0,
		Operations:          make(map[string]*OperationInfo),
		EstimatedRemaining:  0,
	}

	var earliestStart time.Time
	totalProgress := 0.0

	for id, op := range ct.operations {
		// Copy operation info
		opCopy := *op
		status.Operations[id] = &opCopy

		if earliestStart.IsZero() || op.StartTime.Before(earliestStart) {
			earliestStart = op.StartTime
		}

		progress := op.Tracker.GetProgress()
		totalProgress += progress.Percentage

		switch {
		case progress.Percentage >= 100.0:
			status.CompletedOperations++
		case progress.Percentage > 0.0:
			status.ActiveOperations++
		}
	}

	status.StartTime = earliestStart
	status.ElapsedTime = now.Sub(earliestStart)

	if status.TotalOperations > 0 {
		status.OverallProgress = totalProgress / float64(status.TotalOperations)
	}

	// Estimate remaining time based on current progress rate
	if status.OverallProgress > 0 && status.OverallProgress < 100 {
		avgTimePerPercent := status.ElapsedTime / time.Duration(status.OverallProgress)
		remainingPercent := 100.0 - status.OverallProgress
		status.EstimatedRemaining = avgTimePerPercent * time.Duration(remainingPercent)
	}

	return status
}

// Subscribe adds a callback for composite progress updates
func (ct *CompositeTracker) Subscribe(callback CompositeCallback) {
	ct.mu.Lock()
	defer ct.mu.Unlock()
	ct.callbacks = append(ct.callbacks, callback)
}

// updateOperationStatus updates an operation's status and notifies callbacks
func (ct *CompositeTracker) updateOperationStatus(operationID string, status *Status) {
	ct.mu.Lock()
	op, exists := ct.operations[operationID]
	if exists {
		op.Current = status.Current
		op.LastUpdate = time.Now()
		op.Stage = status.Stage
	}
	ct.mu.Unlock()

	if exists {
		// Use a separate goroutine to avoid blocking
		go ct.notifyCallbacks()
	}
}

// notifyCallbacks notifies all registered callbacks
func (ct *CompositeTracker) notifyCallbacks() {
	// Get callbacks without holding lock to avoid deadlock
	ct.mu.RLock()
	callbacks := make([]CompositeCallback, len(ct.callbacks))
	copy(callbacks, ct.callbacks)
	ct.mu.RUnlock()

	// Get status separately
	status := ct.GetCompositeStatus()

	// Notify callbacks
	for _, callback := range callbacks {
		go callback(status) // Run callbacks asynchronously
	}
}

// ProgressAdapters provide utility functions for converting between different progress callback types

// InstallProgressAdapter converts PVI install progress callbacks to unified progress tracking
type InstallProgressAdapter struct {
	tracker Tracker
	stages  map[string]int
}

// NewInstallProgressAdapter creates a new install progress adapter
func NewInstallProgressAdapter(tracker Tracker) *InstallProgressAdapter {
	return &InstallProgressAdapter{
		tracker: tracker,
		stages: map[string]int{
			"resolving":   10,
			"downloading": 30,
			"extracting":  50,
			"building":    70,
			"testing":     80,
			"installing":  90,
			"finished":    100,
		},
	}
}

// AdaptInstallCallback adapts a PVI install callback to unified progress tracking
func (ipa *InstallProgressAdapter) AdaptInstallCallback() func(stage, module, details string, progress float64) {
	return func(stage, module, details string, progress float64) {
		// Convert stage-based progress to unified progress
		stageProgress, exists := ipa.stages[stage]
		if !exists {
			stageProgress = 0
		}

		// Combine stage progress with detailed progress
		overallProgress := int(float64(stageProgress) + (progress * 10)) // 10% per stage detail
		message := fmt.Sprintf("[%s] %s: %s", stage, module, details)

		ipa.tracker.Update(overallProgress, message)
	}
}

// ParallelProgressAdapter converts parallel installation callbacks to unified progress tracking
type ParallelProgressAdapter struct {
	parallelTracker ParallelTracker
	operationMap    map[string]string
	mu              sync.RWMutex
}

// NewParallelProgressAdapter creates a new parallel progress adapter
func NewParallelProgressAdapter(parallelTracker ParallelTracker) *ParallelProgressAdapter {
	return &ParallelProgressAdapter{
		parallelTracker: parallelTracker,
		operationMap:    make(map[string]string),
	}
}

// AdaptParallelCallback adapts a PVI parallel callback to unified progress tracking
func (ppa *ParallelProgressAdapter) AdaptParallelCallback() func(completed, total int, currentModule, stage string) {
	return func(completed, total int, currentModule, stage string) {
		ppa.mu.Lock()
		operationID, exists := ppa.operationMap[currentModule]
		if !exists {
			// Create operation ID for new module
			operationID = fmt.Sprintf("mod_%s", currentModule)
			ppa.operationMap[currentModule] = operationID
		}
		ppa.mu.Unlock()

		// Update operation status
		opStatus := OperationStatus{
			ID:       operationID,
			Name:     currentModule,
			Status:   StatusRunning,
			Message:  stage,
			Progress: float64(completed) / float64(total) * 100.0,
		}

		ppa.parallelTracker.UpdateOperation(operationID, opStatus, stage)
	}
}

// ProgressPersistence provides functionality for saving and restoring progress state
type ProgressPersistence struct {
	filePath string
}

// NewProgressPersistence creates a new progress persistence manager
func NewProgressPersistence(filePath string) *ProgressPersistence {
	return &ProgressPersistence{
		filePath: filePath,
	}
}

// SaveProgress saves progress state to file (implementation would depend on specific needs)
func (pp *ProgressPersistence) SaveProgress(status interface{}) error {
	// Implementation would serialize progress state to file
	// This is a placeholder for the interface
	return nil
}

// LoadProgress loads progress state from file (implementation would depend on specific needs)
func (pp *ProgressPersistence) LoadProgress() (interface{}, error) {
	// Implementation would deserialize progress state from file
	// This is a placeholder for the interface
	return nil, nil
}
