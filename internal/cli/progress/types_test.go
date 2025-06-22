// ABOUTME: Tests for CLI progress tracking types
// ABOUTME: Validates type definitions and methods for progress tracking

package progress

import (
	"testing"
	"time"
)

func TestOpStatus_String(t *testing.T) {
	tests := []struct {
		status   OpStatus
		expected string
	}{
		{StatusPending, "pending"},
		{StatusRunning, "running"},
		{StatusCompleted, "completed"},
		{StatusFailed, "failed"},
		{StatusCancelled, "cancelled"},
		{StatusSkipped, "skipped"},
		{OpStatus(999), "unknown"},
	}

	for _, test := range tests {
		result := test.status.String()
		if result != test.expected {
			t.Errorf("OpStatus(%d).String() = %q, expected %q",
				test.status, result, test.expected)
		}
	}
}

func TestDisplayStyle_String(t *testing.T) {
	tests := []struct {
		style    DisplayStyle
		expected string
	}{
		{StyleBar, "bar"},
		{StyleSpinner, "spinner"},
		{StyleDots, "dots"},
		{StyleLines, "lines"},
		{StyleJSON, "json"},
		{StyleSilent, "silent"},
		{DisplayStyle(999), "unknown"},
	}

	for _, test := range tests {
		result := test.style.String()
		if result != test.expected {
			t.Errorf("DisplayStyle(%d).String() = %q, expected %q",
				test.style, result, test.expected)
		}
	}
}

func TestStatus_Basic(t *testing.T) {
	now := time.Now()
	status := Status{
		Operation:    "install",
		Current:      5,
		Total:        10,
		Message:      "Installing modules",
		Percentage:   50.0,
		StartTime:    now,
		ElapsedTime:  time.Minute,
		Stage:        "downloading",
		SubOperation: "fetching tarball",
	}

	if status.Operation != "install" {
		t.Errorf("Expected operation 'install', got %q", status.Operation)
	}
	if status.Current != 5 {
		t.Errorf("Expected current 5, got %d", status.Current)
	}
	if status.Total != 10 {
		t.Errorf("Expected total 10, got %d", status.Total)
	}
	if status.Percentage != 50.0 {
		t.Errorf("Expected percentage 50.0, got %f", status.Percentage)
	}
}

func TestOperationStatus_Basic(t *testing.T) {
	now := time.Now()
	opStatus := OperationStatus{
		ID:        "op-1",
		Name:      "Install DBI",
		Status:    StatusRunning,
		Message:   "Building module",
		Progress:  75.0,
		StartTime: now,
		Duration:  time.Minute * 2,
	}

	if opStatus.ID != "op-1" {
		t.Errorf("Expected ID 'op-1', got %q", opStatus.ID)
	}
	if opStatus.Status != StatusRunning {
		t.Errorf("Expected status StatusRunning, got %v", opStatus.Status)
	}
	if opStatus.Progress != 75.0 {
		t.Errorf("Expected progress 75.0, got %f", opStatus.Progress)
	}
}

func TestParallelStatus_Basic(t *testing.T) {
	status := ParallelStatus{
		Operations:          make(map[string]*OperationStatus),
		TotalOperations:     5,
		CompletedOperations: 2,
		FailedOperations:    1,
		RunningOperations:   2,
		OverallPercentage:   60.0,
		StartTime:           time.Now(),
		ElapsedTime:         time.Minute * 3,
	}

	if status.TotalOperations != 5 {
		t.Errorf("Expected 5 total operations, got %d", status.TotalOperations)
	}
	if status.CompletedOperations != 2 {
		t.Errorf("Expected 2 completed operations, got %d", status.CompletedOperations)
	}
	if status.FailedOperations != 1 {
		t.Errorf("Expected 1 failed operation, got %d", status.FailedOperations)
	}
	if status.RunningOperations != 2 {
		t.Errorf("Expected 2 running operations, got %d", status.RunningOperations)
	}
}

func TestResult_Basic(t *testing.T) {
	result := Result{
		Operation: "install",
		Target:    "DBI",
		Success:   true,
		Duration:  time.Minute * 2,
		Message:   "Successfully installed DBI",
		Warnings:  []string{"deprecated API used"},
		Metadata:  map[string]interface{}{"version": "1.643"},
	}

	if result.Operation != "install" {
		t.Errorf("Expected operation 'install', got %q", result.Operation)
	}
	if result.Target != "DBI" {
		t.Errorf("Expected target 'DBI', got %q", result.Target)
	}
	if !result.Success {
		t.Error("Expected success to be true")
	}
	if len(result.Warnings) != 1 {
		t.Errorf("Expected 1 warning, got %d", len(result.Warnings))
	}
	if result.Metadata["version"] != "1.643" {
		t.Errorf("Expected version '1.643' in metadata, got %v", result.Metadata["version"])
	}
}

func TestStats_Basic(t *testing.T) {
	stats := Stats{
		ItemsProcessed:      100,
		ItemsSuccessful:     95,
		ItemsFailed:         3,
		ItemsSkipped:        2,
		BytesProcessed:      1024 * 1024,
		AverageTime:         time.Millisecond * 100,
		ThroughputPerSecond: 10.5,
	}

	if stats.ItemsProcessed != 100 {
		t.Errorf("Expected 100 items processed, got %d", stats.ItemsProcessed)
	}
	if stats.ItemsSuccessful != 95 {
		t.Errorf("Expected 95 successful items, got %d", stats.ItemsSuccessful)
	}
	if stats.ItemsFailed != 3 {
		t.Errorf("Expected 3 failed items, got %d", stats.ItemsFailed)
	}
	if stats.ItemsSkipped != 2 {
		t.Errorf("Expected 2 skipped items, got %d", stats.ItemsSkipped)
	}
}

func TestDisplayOptions_Basic(t *testing.T) {
	options := DisplayOptions{
		ShowPercentage: true,
		ShowETA:        true,
		ShowThroughput: false,
		ShowElapsed:    true,
		RefreshRate:    time.Millisecond * 100,
		Width:          80,
		Style:          StyleBar,
		Color:          true,
		Quiet:          false,
		Verbose:        false,
	}

	if !options.ShowPercentage {
		t.Error("Expected ShowPercentage to be true")
	}
	if !options.ShowETA {
		t.Error("Expected ShowETA to be true")
	}
	if options.ShowThroughput {
		t.Error("Expected ShowThroughput to be false")
	}
	if options.Style != StyleBar {
		t.Errorf("Expected style StyleBar, got %v", options.Style)
	}
	if options.Width != 80 {
		t.Errorf("Expected width 80, got %d", options.Width)
	}
}
