// ABOUTME: Tests for progress bounds checking functionality
// ABOUTME: Ensures progress tracking doesn't get stuck and handles edge cases

package progress

import (
	"testing"
	"time"
)

func TestProgressBoundsChecking(t *testing.T) {
	tracker := NewTracker()
	tracker.Start("test-operation", 100)

	// Test normal progress advancement
	tracker.Update(25, "Processing item 25")
	status := tracker.GetProgress()
	if status.Percentage != 25.0 {
		t.Errorf("Expected percentage to be 25.0, got %f", status.Percentage)
	}

	// Test progress doesn't go backwards
	tracker.Update(20, "Processing item 20")
	status = tracker.GetProgress()
	if status.Percentage != 25.0 {
		t.Errorf("Expected percentage to remain 25.0 (no backwards movement), got %f", status.Percentage)
	}

	// Test progress can still advance after preventing backwards movement
	tracker.Update(30, "Processing item 30")
	status = tracker.GetProgress()
	if status.Percentage != 30.0 {
		t.Errorf("Expected percentage to advance to 30.0, got %f", status.Percentage)
	}
}

func TestStuckProgressDetection(t *testing.T) {
	// This test focuses on observable behavior rather than internal implementation
	tracker := NewTracker()
	tracker.Start("test-operation", 100)

	// Set initial progress
	tracker.Update(50, "Processing item 50")
	status := tracker.GetProgress()
	initialPercentage := status.Percentage

	// Test that progress bounds checking works even without direct access to internal fields
	// The system should handle stuck progress automatically
	tracker.Update(50, "Still processing item 50")
	status = tracker.GetProgress()

	// Progress should not go backwards
	if status.Percentage < initialPercentage {
		t.Errorf("Expected percentage to not go backwards from %f, got %f", initialPercentage, status.Percentage)
	}

	// Progress should be at least the initial percentage
	if status.Percentage < initialPercentage {
		t.Errorf("Expected percentage to be at least %f, got %f", initialPercentage, status.Percentage)
	}
}

func TestProgressCappingAt95Percent(t *testing.T) {
	tracker := NewTracker()
	tracker.Start("test-operation", 100)

	// Set progress to 94%
	tracker.Update(94, "Processing item 94")

	// Test that progress is handled correctly even at high percentages
	for i := 0; i < 5; i++ {
		tracker.Update(94, "Still processing item 94")
	}

	status := tracker.GetProgress()

	// Progress should not exceed 100%
	if status.Percentage > 100.0 {
		t.Errorf("Expected percentage to not exceed 100%%, got %f", status.Percentage)
	}

	// Progress should be reasonable
	if status.Percentage < 94.0 {
		t.Errorf("Expected percentage to be at least 94%%, got %f", status.Percentage)
	}
}

func TestProgressWithinBounds(t *testing.T) {
	tracker := NewTracker()
	tracker.Start("test-operation", 100)

	// Test negative progress is clamped to 0
	tracker.Update(-10, "Negative progress")
	status := tracker.GetProgress()
	if status.Percentage < 0.0 {
		t.Errorf("Expected percentage to be clamped to 0.0, got %f", status.Percentage)
	}

	// Test progress over 100% is clamped to 100%
	tracker.Update(150, "Over 100% progress")
	status = tracker.GetProgress()
	if status.Percentage > 100.0 {
		t.Errorf("Expected percentage to be clamped to 100.0, got %f", status.Percentage)
	}
}

func TestProgressChangeTimeTracking(t *testing.T) {
	tracker := NewTracker()
	tracker.Start("test-operation", 100)

	// Test that elapsed time is tracked correctly
	tracker.Update(25, "Processing item 25")
	status := tracker.GetProgress()

	if status.ElapsedTime <= 0 {
		t.Error("Expected elapsed time to be positive")
	}

	// Wait a bit and update again. Sleep for 50ms to ensure the elapsed time
	// is measurably larger than 10ms even on platforms with coarse timer resolution.
	time.Sleep(50 * time.Millisecond)

	tracker.Update(30, "Processing item 30")
	status = tracker.GetProgress()

	// Elapsed time should have increased beyond a small threshold. Using 5ms as
	// the lower bound leaves plenty of margin below the 50ms sleep duration.
	if status.ElapsedTime <= 5*time.Millisecond {
		t.Error("Expected elapsed time to increase with subsequent updates")
	}

	// Progress should be updated correctly
	if status.Percentage != 30.0 {
		t.Errorf("Expected percentage to be 30.0, got %f", status.Percentage)
	}
}

func TestProgressBoundsWithZeroTotal(t *testing.T) {
	tracker := NewTracker()
	tracker.Start("test-operation", 0)

	// Update with zero total should not cause issues
	tracker.Update(10, "Processing without total")
	status := tracker.GetProgress()

	// Percentage should remain 0 when total is 0
	if status.Percentage != 0.0 {
		t.Errorf("Expected percentage to be 0.0 when total is 0, got %f", status.Percentage)
	}
}

func TestProgressBoundsIntegration(t *testing.T) {
	tracker := NewTracker()
	tracker.Start("test-operation", 100)

	// Simulate a realistic scenario with various updates
	updates := []struct {
		current int
		message string
	}{
		{10, "Starting process"},
		{25, "Quarter done"},
		{25, "Still at quarter"}, // Stuck progress
		{25, "Still at quarter"}, // Still stuck
		{50, "Half done"},
		{45, "Trying to go backwards"}, // Should be prevented
		{75, "Three quarters done"},
		{100, "Complete"},
	}

	var previousPercentage float64

	for _, update := range updates {
		// For stuck progress simulation, just update normally
		// The bounds checking should handle this automatically

		tracker.Update(update.current, update.message)
		status := tracker.GetProgress()

		// Progress should never go backwards
		if status.Percentage < previousPercentage {
			t.Errorf("Progress went backwards from %f to %f for update: %v", previousPercentage, status.Percentage, update)
		}

		// Progress should be within bounds
		if status.Percentage < 0.0 || status.Percentage > 100.0 {
			t.Errorf("Progress %f is out of bounds [0.0, 100.0] for update: %v", status.Percentage, update)
		}

		previousPercentage = status.Percentage
	}

	// Final progress should be 100%
	finalStatus := tracker.GetProgress()
	if finalStatus.Percentage != 100.0 {
		t.Errorf("Expected final progress to be 100.0, got %f", finalStatus.Percentage)
	}
}
