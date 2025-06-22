// ABOUTME: Progress tracking helper functions for module operations
// ABOUTME: Provides utilities for consistent progress reporting across all module operations

package progress

import (
	"context"
	"fmt"
)

// CreateStandardTracker creates a progress tracker with standard configuration
func CreateStandardTracker(ctx context.Context, operationName string, total int) (*OperationTracker, Tracker) {
	tracker := NewTracker()
	opTracker := NewOperationTracker(ctx, tracker)
	opTracker.StartOperation(operationName, total)
	return opTracker, tracker
}

// CreateStandardParallelTracker creates a parallel tracker with standard configuration
func CreateStandardParallelTracker(maxWorkers int, operations []string) ParallelTracker {
	tracker := NewParallelTracker(maxWorkers)
	tracker.StartParallel(operations)
	return tracker
}

// CreateCompositeTracker creates a composite tracker for complex operations
func CreateCompositeTracker(maxWorkers int) *CompositeTracker {
	return NewCompositeTracker(maxWorkers)
}

// CreateModuleInstallTracker creates a specialized tracker for module installation
func CreateModuleInstallTracker(ctx context.Context, modules []string) (*CompositeTracker, []string) {
	tracker := NewCompositeTracker(len(modules))
	operationIDs := make([]string, len(modules))

	for i, module := range modules {
		operationIDs[i] = tracker.AddOperation(fmt.Sprintf("install-%s", module), 100)
	}

	return tracker, operationIDs
}

// CreateModuleListTracker creates a tracker for module listing operations
func CreateModuleListTracker(ctx context.Context, estimatedCount int) (*OperationTracker, Tracker) {
	return CreateStandardTracker(ctx, "list-modules", estimatedCount)
}

// CreateModuleSearchTracker creates a tracker for module search operations
func CreateModuleSearchTracker(ctx context.Context, query string) (*OperationTracker, Tracker) {
	return CreateStandardTracker(ctx, fmt.Sprintf("search-%s", query), 100)
}

// CreateModuleRemoveTracker creates a tracker for module removal operations
func CreateModuleRemoveTracker(ctx context.Context, modules []string) (*CompositeTracker, []string) {
	tracker := NewCompositeTracker(len(modules))
	operationIDs := make([]string, len(modules))

	for i, module := range modules {
		operationIDs[i] = tracker.AddOperation(fmt.Sprintf("remove-%s", module), 100)
	}

	return tracker, operationIDs
}

// CreateDependencyResolverTracker creates a tracker for dependency resolution
func CreateDependencyResolverTracker(ctx context.Context, modules []string) (*OperationTracker, Tracker) {
	return CreateStandardTracker(ctx, "resolve-dependencies", len(modules)*10) // Estimate 10 operations per module
}

// ProgressCallbackFactory provides factories for creating standard progress callbacks

// CreateInstallProgressCallback creates a standardized install progress callback
func CreateInstallProgressCallback(tracker Tracker) func(stage, module, details string, progress float64) {
	adapter := NewInstallProgressAdapter(tracker)
	return adapter.AdaptInstallCallback()
}

// CreateParallelInstallProgressCallback creates a standardized parallel install progress callback
func CreateParallelInstallProgressCallback(parallelTracker ParallelTracker) func(completed, total int, currentModule, stage string) {
	adapter := NewParallelProgressAdapter(parallelTracker)
	return adapter.AdaptParallelCallback()
}

// CreateUIProgressCallback creates a callback that integrates with UI systems
func CreateUIProgressCallback(uiProgressFunc func(current, total int, message string)) Callback {
	return func(status *Status) {
		if uiProgressFunc != nil {
			uiProgressFunc(status.Current, status.Total, status.Message)
		}
	}
}

// CreateVerboseProgressCallback creates a callback for verbose progress output
func CreateVerboseProgressCallback(logFunc func(format string, args ...interface{})) Callback {
	return func(status *Status) {
		if logFunc != nil {
			logFunc("[%s] %s: %s (%.1f%% - %v elapsed)",
				status.Stage,
				status.Operation,
				status.Message,
				status.Percentage,
				status.ElapsedTime)
		}
	}
}

// CreateJSONProgressCallback creates a callback for JSON progress output
func CreateJSONProgressCallback(outputFunc func(data []byte)) Callback {
	return func(status *Status) {
		if outputFunc != nil {
			// This would typically use a JSON encoder
			// For now, just format as a simple string representation
			data := fmt.Sprintf(`{"operation":"%s","current":%d,"total":%d,"percentage":%.2f,"message":"%s","stage":"%s"}`,
				status.Operation, status.Current, status.Total, status.Percentage, status.Message, status.Stage)
			outputFunc([]byte(data))
		}
	}
}

// Progress Stage Constants for standardization across operations
const (
	StageInitializing = "initializing"
	StageProcessing   = "processing"
	StageCompleting   = "completing"
	StageCompleted    = "completed"
	StageFailed       = "failed"
	StageCancelled    = "cancelled"
)

// Module-specific stage constants
const (
	StageResolving   = "resolving"
	StageDownloading = "downloading"
	StageExtracting  = "extracting"
	StageBuilding    = "building"
	StageTesting     = "testing"
	StageInstalling  = "installing"
	StageCleaningUp  = "cleaning-up"
	StageValidating  = "validating"
	StageSearching   = "searching"
	StageFiltering   = "filtering"
	StageRemoving    = "removing"
	StageUpdating    = "updating"
)

// ProgressUtils provides utility functions for common progress operations

// UpdateProgressWithStage updates progress with a specific stage
func UpdateProgressWithStage(tracker Tracker, current int, stage, message string) {
	tracker.Update(current, message)
	tracker.SetMessage(fmt.Sprintf("[%s] %s", stage, message))
}

// FinishProgressWithResult finishes progress with a standardized result
func FinishProgressWithResult(tracker Tracker, success bool, operation, target string, err error) {
	var message string
	if success {
		message = fmt.Sprintf("Successfully completed %s for %s", operation, target)
	} else {
		if err != nil {
			message = fmt.Sprintf("Failed %s for %s: %v", operation, target, err)
		} else {
			message = fmt.Sprintf("Failed %s for %s", operation, target)
		}
	}

	result := &Result{
		Operation: operation,
		Target:    target,
		Success:   success,
		Message:   message,
		Error:     err,
	}

	tracker.Finish(result)
}

// EstimateOperationSteps provides standard estimates for different operation types
func EstimateOperationSteps(operationType string, itemCount int) int {
	switch operationType {
	case "install":
		return itemCount * 7 // 7 stages per module install
	case "remove":
		return itemCount * 3 // 3 stages per module remove
	case "update":
		return itemCount * 5 // 5 stages per module update
	case "search":
		return 100 // Fixed estimate for search
	case "list":
		return itemCount // 1 step per module to list
	case "resolve-dependencies":
		return itemCount * 2 // 2 steps per module for dependency resolution
	default:
		return itemCount // Default to 1 step per item
	}
}

// CreateProgressConfig creates a standard progress configuration
type ProgressConfig struct {
	ShowPercentage bool
	ShowETA        bool
	ShowThroughput bool
	ShowElapsed    bool
	Verbose        bool
	JSONOutput     bool
	RefreshRate    int // milliseconds
}

// DefaultProgressConfig returns a default progress configuration
func DefaultProgressConfig() *ProgressConfig {
	return &ProgressConfig{
		ShowPercentage: true,
		ShowETA:        true,
		ShowThroughput: false,
		ShowElapsed:    true,
		Verbose:        false,
		JSONOutput:     false,
		RefreshRate:    100, // 100ms
	}
}

// VerboseProgressConfig returns a verbose progress configuration
func VerboseProgressConfig() *ProgressConfig {
	return &ProgressConfig{
		ShowPercentage: true,
		ShowETA:        true,
		ShowThroughput: true,
		ShowElapsed:    true,
		Verbose:        true,
		JSONOutput:     false,
		RefreshRate:    50, // 50ms for more frequent updates
	}
}

// QuietProgressConfig returns a minimal progress configuration
func QuietProgressConfig() *ProgressConfig {
	return &ProgressConfig{
		ShowPercentage: false,
		ShowETA:        false,
		ShowThroughput: false,
		ShowElapsed:    false,
		Verbose:        false,
		JSONOutput:     false,
		RefreshRate:    1000, // 1 second for minimal updates
	}
}

// JSONProgressConfig returns a JSON-only progress configuration
func JSONProgressConfig() *ProgressConfig {
	return &ProgressConfig{
		ShowPercentage: false,
		ShowETA:        false,
		ShowThroughput: false,
		ShowElapsed:    false,
		Verbose:        false,
		JSONOutput:     true,
		RefreshRate:    200, // 200ms
	}
}
