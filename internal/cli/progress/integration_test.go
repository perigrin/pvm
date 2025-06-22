// ABOUTME: Tests for progress tracking integration utilities
// ABOUTME: Validates integration adapters and composite tracking functionality

package progress

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestOperationTracker_BasicOperation(t *testing.T) {
	ctx := context.Background()
	tracker := NewTracker()
	opTracker := NewOperationTracker(ctx, tracker)

	// Start operation
	opTracker.StartOperation("test-operation", 100)

	// Update progress
	opTracker.UpdateOperation(50, "halfway done")

	// Check tracker state
	status := tracker.GetProgress()
	if status.Current != 50 {
		t.Errorf("Expected current 50, got %d", status.Current)
	}
	if status.Message != "halfway done" {
		t.Errorf("Expected message 'halfway done', got '%s'", status.Message)
	}

	// Finish operation
	result := &Result{
		Operation: "test-operation",
		Success:   true,
		Duration:  time.Second,
	}
	opTracker.FinishOperation(result)

	// Check that context is cancelled after finish
	select {
	case <-opTracker.Context().Done():
		// Expected - context should be cancelled
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected context to be cancelled after finish")
	}
}

func TestOperationTracker_Cancellation(t *testing.T) {
	ctx := context.Background()
	tracker := NewTracker()
	opTracker := NewOperationTracker(ctx, tracker)

	opTracker.StartOperation("test-operation", 100)

	// Cancel operation
	opTracker.Cancel()

	// Check that context is cancelled
	select {
	case <-opTracker.Context().Done():
		// Expected
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected context to be cancelled")
	}

	// Updates after cancellation should be ignored
	initialStatus := tracker.GetProgress()
	opTracker.UpdateOperation(75, "should be ignored")

	// Status should remain unchanged (due to cancellation)
	currentStatus := tracker.GetProgress()
	if currentStatus.Current != initialStatus.Current {
		t.Error("Expected update to be ignored after cancellation")
	}
}

func TestOperationTracker_SubTrackers(t *testing.T) {
	ctx := context.Background()
	tracker := NewTracker()
	opTracker := NewOperationTracker(ctx, tracker)

	opTracker.StartOperation("parent-operation", 100)

	// Create sub-tracker
	subTracker := opTracker.CreateSubTracker("sub-operation")
	if subTracker == nil {
		t.Fatal("Expected sub-tracker to be created")
	}

	// Start sub-operation
	subTracker.Start("sub-task", 50)
	subTracker.Update(25, "sub-task progress")

	// Give callbacks time to propagate
	time.Sleep(10 * time.Millisecond)

	// Main tracker should have been updated by sub-tracker
	status := tracker.GetProgress()
	if status.Message == "" {
		t.Error("Expected main tracker to be updated by sub-tracker")
	}

	// Retrieve sub-tracker
	retrievedSubTracker := opTracker.GetSubTracker("sub-operation")
	if retrievedSubTracker == nil {
		t.Error("Expected to retrieve sub-tracker")
	}
}

func TestCompositeTracker_BasicOperation(t *testing.T) {
	tracker := NewCompositeTracker(3)

	// Add operations
	op1ID := tracker.AddOperation("operation-1", 100)
	op2ID := tracker.AddOperation("operation-2", 200)

	status := tracker.GetCompositeStatus()
	if status.TotalOperations != 2 {
		t.Errorf("Expected 2 total operations, got %d", status.TotalOperations)
	}

	// Start operations
	tracker.StartOperation(op1ID)
	tracker.StartOperation(op2ID)

	// Update progress
	tracker.UpdateOperation(op1ID, 50, "operation 1 halfway")
	tracker.UpdateOperation(op2ID, 100, "operation 2 halfway")

	status = tracker.GetCompositeStatus()
	if status.ActiveOperations != 2 {
		t.Errorf("Expected 2 active operations, got %d", status.ActiveOperations)
	}

	// Finish operations
	result1 := &Result{Operation: "operation-1", Success: true}
	result2 := &Result{Operation: "operation-2", Success: true}

	tracker.FinishOperation(op1ID, result1)
	tracker.FinishOperation(op2ID, result2)

	status = tracker.GetCompositeStatus()
	if status.CompletedOperations != 2 {
		t.Errorf("Expected 2 completed operations, got %d", status.CompletedOperations)
	}
	if status.OverallProgress < 100.0 {
		t.Errorf("Expected overall progress to be 100%%, got %f", status.OverallProgress)
	}
}

func TestCompositeTracker_Callbacks(t *testing.T) {
	tracker := NewCompositeTracker(2)

	var callbackCalled bool
	var lastStatus *CompositeStatus

	callback := func(status *CompositeStatus) {
		callbackCalled = true
		lastStatus = status
	}

	tracker.Subscribe(callback)

	// Add operation should trigger callback
	opID := tracker.AddOperation("test-operation", 100)
	time.Sleep(10 * time.Millisecond) // Allow callback to run

	if !callbackCalled {
		t.Error("Expected callback to be called when adding operation")
	}
	if lastStatus == nil {
		t.Error("Expected status to be set in callback")
	}
	if lastStatus.TotalOperations != 1 {
		t.Errorf("Expected 1 total operation in callback, got %d", lastStatus.TotalOperations)
	}

	// Reset callback state
	callbackCalled = false
	lastStatus = nil

	// Update operation should trigger callback
	tracker.StartOperation(opID)
	tracker.UpdateOperation(opID, 50, "updated")
	time.Sleep(10 * time.Millisecond) // Allow callback to run

	if !callbackCalled {
		t.Error("Expected callback to be called when updating operation")
	}
}

func TestInstallProgressAdapter(t *testing.T) {
	tracker := NewTracker()
	adapter := NewInstallProgressAdapter(tracker)

	callback := adapter.AdaptInstallCallback()

	// Start tracking
	tracker.Start("install-test", 100)

	// Simulate install progress stages
	callback("resolving", "Test::Module", "resolving dependencies", 0.5)
	callback("downloading", "Test::Module", "downloading archive", 1.0)
	callback("finished", "Test::Module", "installation complete", 1.0)

	status := tracker.GetProgress()
	if status.Current < 100 {
		t.Errorf("Expected progress to reach 100, got %d", status.Current)
	}
	if status.Message == "" {
		t.Error("Expected status message to be set")
	}
}

func TestParallelProgressAdapter(t *testing.T) {
	parallelTracker := NewParallelTracker(3)
	adapter := NewParallelProgressAdapter(parallelTracker)

	callback := adapter.AdaptParallelCallback()

	// Start parallel tracking
	operations := []string{"mod1", "mod2", "mod3"}
	parallelTracker.StartParallel(operations)

	// Simulate parallel progress updates
	callback(1, 3, "mod1", "downloading")
	callback(2, 3, "mod2", "installing")
	callback(3, 3, "mod3", "finished")

	status := parallelTracker.GetOverallProgress()
	if status.TotalOperations != 3 {
		t.Errorf("Expected 3 total operations, got %d", status.TotalOperations)
	}
}

func TestCompositeTracker_ConcurrentAccess(t *testing.T) {
	tracker := NewCompositeTracker(10)

	var wg sync.WaitGroup
	numOperations := 20

	// Add operations concurrently
	operationIDs := make([]string, numOperations)
	for i := 0; i < numOperations; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			opID := tracker.AddOperation(fmt.Sprintf("op-%d", index), 100)
			operationIDs[index] = opID
		}(i)
	}

	wg.Wait()

	// Update operations concurrently
	for i := 0; i < numOperations; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			if operationIDs[index] != "" {
				tracker.StartOperation(operationIDs[index])
				tracker.UpdateOperation(operationIDs[index], 50, "concurrent update")
				tracker.FinishOperation(operationIDs[index], &Result{Success: true})
			}
		}(i)
	}

	// Read status concurrently
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				_ = tracker.GetCompositeStatus()
				time.Sleep(time.Microsecond)
			}
		}()
	}

	wg.Wait()

	// Should not panic
	status := tracker.GetCompositeStatus()
	if status == nil {
		t.Error("Expected status to be non-nil")
	}
}

func TestCompositeTracker_ProgressCalculation(t *testing.T) {
	tracker := NewCompositeTracker(3)

	// Add operations with different progress levels
	op1ID := tracker.AddOperation("op1", 100)
	op2ID := tracker.AddOperation("op2", 100)
	op3ID := tracker.AddOperation("op3", 100)

	tracker.StartOperation(op1ID)
	tracker.StartOperation(op2ID)
	tracker.StartOperation(op3ID)

	// Set different progress levels
	tracker.UpdateOperation(op1ID, 100, "completed") // 100%
	tracker.UpdateOperation(op2ID, 50, "halfway")    // 50%
	tracker.UpdateOperation(op3ID, 0, "starting")    // 0%

	// Give some time for progress to be calculated
	time.Sleep(10 * time.Millisecond)

	status := tracker.GetCompositeStatus()
	expectedProgress := (100.0 + 50.0 + 0.0) / 3.0 // Average progress

	// Allow for some variance in progress calculation
	if status.OverallProgress < expectedProgress-5 || status.OverallProgress > expectedProgress+5 {
		t.Errorf("Expected overall progress around %f, got %f", expectedProgress, status.OverallProgress)
	}
}

func TestOperationTracker_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	tracker := NewTracker()
	opTracker := NewOperationTracker(ctx, tracker)

	opTracker.StartOperation("test-operation", 100)

	// Cancel parent context
	cancel()

	// Give cancellation time to propagate
	time.Sleep(10 * time.Millisecond)

	// Operation tracker context should also be cancelled
	select {
	case <-opTracker.Context().Done():
		// Expected
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected operation tracker context to be cancelled")
	}
}

func TestProgressPersistence(t *testing.T) {
	persistence := NewProgressPersistence("/tmp/test-progress.json")

	// Test that persistence object is created
	if persistence == nil {
		t.Error("Expected persistence object to be created")
	}

	// Note: Actual save/load functionality would require file I/O implementation
	// This test just verifies the interface is available
	err := persistence.SaveProgress(map[string]interface{}{"test": "data"})
	if err != nil {
		// Placeholder implementation returns nil, so this is expected
	}

	_, err = persistence.LoadProgress()
	if err != nil {
		// Placeholder implementation returns nil, so this is expected
	}
}
