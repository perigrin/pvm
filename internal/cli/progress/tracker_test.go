// ABOUTME: Tests for progress tracking implementations
// ABOUTME: Validates tracker functionality and concurrent safety

package progress

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestSimpleTracker_BasicOperation(t *testing.T) {
	tracker := NewTracker()

	// Start operation
	tracker.Start("test-operation", 100)

	if !tracker.IsRunning() {
		t.Error("Expected tracker to be running")
	}

	status := tracker.GetProgress()
	if status.Operation != "test-operation" {
		t.Errorf("Expected operation 'test-operation', got '%s'", status.Operation)
	}
	if status.Total != 100 {
		t.Errorf("Expected total 100, got %d", status.Total)
	}
	if status.Current != 0 {
		t.Errorf("Expected current 0, got %d", status.Current)
	}

	// Update progress
	tracker.Update(50, "halfway done")
	status = tracker.GetProgress()
	if status.Current != 50 {
		t.Errorf("Expected current 50, got %d", status.Current)
	}
	if status.Percentage != 50.0 {
		t.Errorf("Expected percentage 50.0, got %f", status.Percentage)
	}
	if status.Message != "halfway done" {
		t.Errorf("Expected message 'halfway done', got '%s'", status.Message)
	}

	// Finish operation
	result := &Result{
		Operation: "test-operation",
		Target:    "test-target",
		Success:   true,
		Duration:  time.Second,
		Message:   "completed successfully",
	}
	tracker.Finish(result)

	if tracker.IsRunning() {
		t.Error("Expected tracker to not be running after finish")
	}

	status = tracker.GetProgress()
	if status.Current != 100 {
		t.Errorf("Expected current 100, got %d", status.Current)
	}
	if status.Percentage != 100.0 {
		t.Errorf("Expected percentage 100.0, got %f", status.Percentage)
	}
}

func TestSimpleTracker_Callbacks(t *testing.T) {
	tracker := NewTracker()

	var callbackCalled bool
	var lastStatus *Status

	callback := func(status *Status) {
		callbackCalled = true
		lastStatus = status
	}

	tracker.Subscribe(callback)

	// Start operation should trigger callback
	tracker.Start("test-operation", 10)
	time.Sleep(10 * time.Millisecond) // Allow callback to run

	if !callbackCalled {
		t.Error("Expected callback to be called")
	}
	if lastStatus == nil {
		t.Error("Expected status to be set in callback")
	}
	if lastStatus.Operation != "test-operation" {
		t.Errorf("Expected operation 'test-operation', got '%s'", lastStatus.Operation)
	}

	// Reset for next test
	callbackCalled = false
	lastStatus = nil

	// Update should trigger callback
	tracker.Update(5, "updated")
	time.Sleep(10 * time.Millisecond) // Allow callback to run

	if !callbackCalled {
		t.Error("Expected callback to be called on update")
	}
	if lastStatus.Current != 5 {
		t.Errorf("Expected current 5, got %d", lastStatus.Current)
	}
}

func TestSimpleTracker_SetTotal(t *testing.T) {
	tracker := NewTracker()

	tracker.Start("test-operation", 100)
	tracker.Update(50, "halfway")

	// Change total
	tracker.SetTotal(200)

	status := tracker.GetProgress()
	if status.Total != 200 {
		t.Errorf("Expected total 200, got %d", status.Total)
	}
	if status.Percentage != 25.0 {
		t.Errorf("Expected percentage 25.0, got %f", status.Percentage)
	}
}

func TestSimpleTracker_SetMessage(t *testing.T) {
	tracker := NewTracker()

	tracker.Start("test-operation", 100)
	tracker.SetMessage("new message")

	status := tracker.GetProgress()
	if status.Message != "new message" {
		t.Errorf("Expected message 'new message', got '%s'", status.Message)
	}
}

func TestSimpleTracker_ConcurrentAccess(t *testing.T) {
	tracker := NewTracker()
	tracker.Start("concurrent-test", 1000)

	var wg sync.WaitGroup
	numGoroutines := 10

	// Start multiple goroutines updating progress
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				tracker.Update(id*100+j, "concurrent update")
				time.Sleep(time.Microsecond)
			}
		}(i)
	}

	// Start multiple goroutines reading progress
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_ = tracker.GetProgress()
				time.Sleep(time.Microsecond)
			}
		}()
	}

	wg.Wait()

	// Should not panic and should be in valid state
	status := tracker.GetProgress()
	if status == nil {
		t.Error("Expected status to be non-nil")
	}
}

func TestSimpleParallelTracker_BasicOperation(t *testing.T) {
	tracker := NewParallelTracker(5)

	operations := []string{"op1", "op2", "op3"}
	tracker.StartParallel(operations)

	status := tracker.GetOverallProgress()
	if status.TotalOperations != 3 {
		t.Errorf("Expected total operations 3, got %d", status.TotalOperations)
	}
	if status.CompletedOperations != 0 {
		t.Errorf("Expected completed operations 0, got %d", status.CompletedOperations)
	}
	if len(status.Operations) != 3 {
		t.Errorf("Expected 3 operations in status, got %d", len(status.Operations))
	}

	// Find an operation ID to update
	var opID string
	for id := range status.Operations {
		opID = id
		break
	}

	// Update operation
	opStatus := OperationStatus{
		Status:   StatusRunning,
		Progress: 50.0,
	}
	tracker.UpdateOperation(opID, opStatus, "running")

	status = tracker.GetOverallProgress()
	if status.RunningOperations != 1 {
		t.Errorf("Expected running operations 1, got %d", status.RunningOperations)
	}

	// Finish operation
	result := &Result{
		Operation: "op1",
		Target:    "target1",
		Success:   true,
		Duration:  time.Second,
	}
	tracker.FinishOperation(opID, result)

	status = tracker.GetOverallProgress()
	if status.CompletedOperations != 1 {
		t.Errorf("Expected completed operations 1, got %d", status.CompletedOperations)
	}
	if status.RunningOperations != 0 {
		t.Errorf("Expected running operations 0, got %d", status.RunningOperations)
	}
}

func TestSimpleParallelTracker_FinishAll(t *testing.T) {
	tracker := NewParallelTracker(3)

	operations := []string{"op1", "op2", "op3"}
	tracker.StartParallel(operations)

	results := []*Result{
		{Operation: "op1", Success: true},
		{Operation: "op2", Success: true},
		{Operation: "op3", Success: false},
	}

	tracker.FinishAll(results)

	status := tracker.GetOverallProgress()
	if status.CompletedOperations != 2 {
		t.Errorf("Expected completed operations 2, got %d", status.CompletedOperations)
	}
	if status.FailedOperations != 1 {
		t.Errorf("Expected failed operations 1, got %d", status.FailedOperations)
	}
	// 2 operations at 100% (completed) + 1 operation at 100% (failed but finished) = 300/3 = 100%
	// But our current logic calculates based on progress, not completion
	expectedPercentage := (100.0 + 100.0 + 100.0) / 3.0
	if status.OverallPercentage != expectedPercentage {
		t.Errorf("Expected overall percentage %f, got %f", expectedPercentage, status.OverallPercentage)
	}
}

func TestSimpleParallelTracker_GetOperationStatus(t *testing.T) {
	tracker := NewParallelTracker(2)

	operations := []string{"op1", "op2"}
	tracker.StartParallel(operations)

	status := tracker.GetOverallProgress()
	var opID string
	for id := range status.Operations {
		opID = id
		break
	}

	opStatus := tracker.GetOperationStatus(opID)
	if opStatus == nil {
		t.Error("Expected operation status to be non-nil")
	}
	if opStatus.Status != StatusPending {
		t.Errorf("Expected status pending, got %s", opStatus.Status.String())
	}

	// Test non-existent operation
	nonExistentStatus := tracker.GetOperationStatus("non-existent")
	if nonExistentStatus != nil {
		t.Error("Expected non-existent operation status to be nil")
	}
}

func TestSimpleParallelTracker_ConcurrentAccess(t *testing.T) {
	tracker := NewParallelTracker(10)

	operations := make([]string, 20)
	for i := 0; i < 20; i++ {
		operations[i] = fmt.Sprintf("op%d", i)
	}

	tracker.StartParallel(operations)

	var wg sync.WaitGroup

	// Start goroutines to update operations
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			status := tracker.GetOverallProgress()
			for opID := range status.Operations {
				opStatus := OperationStatus{
					Status:   StatusRunning,
					Progress: 50.0,
				}
				tracker.UpdateOperation(opID, opStatus, "concurrent update")
				break
			}
		}(i)
	}

	// Start goroutines to read status
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				_ = tracker.GetOverallProgress()
				time.Sleep(time.Microsecond)
			}
		}()
	}

	wg.Wait()

	// Should not panic
	status := tracker.GetOverallProgress()
	if status == nil {
		t.Error("Expected status to be non-nil")
	}
}

func TestNullTracker(t *testing.T) {
	tracker := NewNullTracker()

	// All operations should be no-ops
	tracker.Start("test", 100)
	tracker.Update(50, "test")
	tracker.Finish(&Result{})
	tracker.SetTotal(200)
	tracker.SetMessage("test")

	if tracker.IsRunning() {
		t.Error("Expected null tracker to never be running")
	}

	status := tracker.GetProgress()
	if status == nil {
		t.Error("Expected status to be non-nil")
	}
	if status.Operation != "" {
		t.Errorf("Expected empty operation, got '%s'", status.Operation)
	}
}

func TestSimpleTracker_EstimatedTime(t *testing.T) {
	tracker := NewTracker()

	tracker.Start("test-operation", 100)

	// Update with some progress
	time.Sleep(10 * time.Millisecond)
	tracker.Update(25, "quarter done")

	status := tracker.GetProgress()
	if status.EstimatedRemaining <= 0 {
		t.Error("Expected estimated remaining time to be positive")
	}

	// The estimated time should be approximately 3 times the elapsed time
	// (since we're 25% done, we need 75% more time)
	expectedRatio := 3.0
	actualRatio := float64(status.EstimatedRemaining) / float64(status.ElapsedTime)

	// Allow for some variance due to timing
	if actualRatio < expectedRatio*0.5 || actualRatio > expectedRatio*2.0 {
		t.Errorf("Expected ratio around %f, got %f", expectedRatio, actualRatio)
	}
}

func TestSimpleParallelTracker_ProgressCalculation(t *testing.T) {
	tracker := NewParallelTracker(3)

	operations := []string{"op1", "op2", "op3"}
	tracker.StartParallel(operations)

	status := tracker.GetOverallProgress()
	var opIDs []string
	for id := range status.Operations {
		opIDs = append(opIDs, id)
	}

	// Update operations with different progress
	tracker.UpdateOperation(opIDs[0], OperationStatus{Status: StatusRunning, Progress: 100.0}, "done")
	tracker.UpdateOperation(opIDs[1], OperationStatus{Status: StatusRunning, Progress: 50.0}, "half")
	tracker.UpdateOperation(opIDs[2], OperationStatus{Status: StatusRunning, Progress: 0.0}, "starting")

	status = tracker.GetOverallProgress()
	expectedPercentage := (100.0 + 50.0 + 0.0) / 3.0
	if status.OverallPercentage != expectedPercentage {
		t.Errorf("Expected overall percentage %f, got %f", expectedPercentage, status.OverallPercentage)
	}
}
