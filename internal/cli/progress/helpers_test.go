// ABOUTME: Tests for progress tracking helper functions
// ABOUTME: Validates helper utilities and configuration functions

package progress

import (
	"context"
	"testing"
	"time"
)

func TestCreateStandardTracker(t *testing.T) {
	ctx := context.Background()
	opTracker, tracker := CreateStandardTracker(ctx, "test-operation", 100)

	if opTracker == nil {
		t.Fatal("Expected operation tracker to be created")
	}
	if tracker == nil {
		t.Fatal("Expected tracker to be created")
	}

	// Check that the operation was started
	status := tracker.GetProgress()
	if status.Operation != "test-operation" {
		t.Errorf("Expected operation 'test-operation', got '%s'", status.Operation)
	}
	if status.Total != 100 {
		t.Errorf("Expected total 100, got %d", status.Total)
	}
}

func TestCreateStandardParallelTracker(t *testing.T) {
	operations := []string{"op1", "op2", "op3"}
	tracker := CreateStandardParallelTracker(3, operations)

	if tracker == nil {
		t.Fatal("Expected parallel tracker to be created")
	}

	status := tracker.GetOverallProgress()
	if status.TotalOperations != 3 {
		t.Errorf("Expected 3 total operations, got %d", status.TotalOperations)
	}
}

func TestCreateModuleInstallTracker(t *testing.T) {
	ctx := context.Background()
	modules := []string{"Module::A", "Module::B", "Module::C"}

	tracker, operationIDs := CreateModuleInstallTracker(ctx, modules)

	if tracker == nil {
		t.Fatal("Expected composite tracker to be created")
	}
	if len(operationIDs) != len(modules) {
		t.Errorf("Expected %d operation IDs, got %d", len(modules), len(operationIDs))
	}

	status := tracker.GetCompositeStatus()
	if status.TotalOperations != 3 {
		t.Errorf("Expected 3 total operations, got %d", status.TotalOperations)
	}
}

func TestCreateModuleListTracker(t *testing.T) {
	ctx := context.Background()
	opTracker, tracker := CreateModuleListTracker(ctx, 50)

	if opTracker == nil {
		t.Fatal("Expected operation tracker to be created")
	}
	if tracker == nil {
		t.Fatal("Expected tracker to be created")
	}

	status := tracker.GetProgress()
	if status.Operation != "list-modules" {
		t.Errorf("Expected operation 'list-modules', got '%s'", status.Operation)
	}
}

func TestCreateModuleSearchTracker(t *testing.T) {
	ctx := context.Background()
	opTracker, tracker := CreateModuleSearchTracker(ctx, "test-query")

	if opTracker == nil {
		t.Fatal("Expected operation tracker to be created")
	}
	if tracker == nil {
		t.Fatal("Expected tracker to be created")
	}

	status := tracker.GetProgress()
	if status.Operation != "search-test-query" {
		t.Errorf("Expected operation 'search-test-query', got '%s'", status.Operation)
	}
}

func TestCreateModuleRemoveTracker(t *testing.T) {
	ctx := context.Background()
	modules := []string{"Module::A", "Module::B"}

	tracker, operationIDs := CreateModuleRemoveTracker(ctx, modules)

	if tracker == nil {
		t.Fatal("Expected composite tracker to be created")
	}
	if len(operationIDs) != len(modules) {
		t.Errorf("Expected %d operation IDs, got %d", len(modules), len(operationIDs))
	}

	status := tracker.GetCompositeStatus()
	if status.TotalOperations != 2 {
		t.Errorf("Expected 2 total operations, got %d", status.TotalOperations)
	}
}

func TestCreateDependencyResolverTracker(t *testing.T) {
	ctx := context.Background()
	modules := []string{"Module::A", "Module::B"}

	opTracker, tracker := CreateDependencyResolverTracker(ctx, modules)

	if opTracker == nil {
		t.Fatal("Expected operation tracker to be created")
	}
	if tracker == nil {
		t.Fatal("Expected tracker to be created")
	}

	status := tracker.GetProgress()
	if status.Operation != "resolve-dependencies" {
		t.Errorf("Expected operation 'resolve-dependencies', got '%s'", status.Operation)
	}
	if status.Total != 20 { // 2 modules * 10 operations per module
		t.Errorf("Expected total 20, got %d", status.Total)
	}
}

func TestCreateInstallProgressCallback(t *testing.T) {
	tracker := NewTracker()
	callback := CreateInstallProgressCallback(tracker)

	if callback == nil {
		t.Fatal("Expected callback to be created")
	}

	// Start tracking
	tracker.Start("install-test", 100)

	// Test callback
	callback("downloading", "Test::Module", "downloading archive", 0.5)

	status := tracker.GetProgress()
	if status.Current == 0 {
		t.Error("Expected progress to be updated by callback")
	}
}

func TestCreateParallelInstallProgressCallback(t *testing.T) {
	parallelTracker := NewParallelTracker(3)
	callback := CreateParallelInstallProgressCallback(parallelTracker)

	if callback == nil {
		t.Fatal("Expected callback to be created")
	}

	// Start parallel tracking
	operations := []string{"mod1", "mod2", "mod3"}
	parallelTracker.StartParallel(operations)

	// Test callback
	callback(1, 3, "mod1", "downloading")

	status := parallelTracker.GetOverallProgress()
	if status.TotalOperations != 3 {
		t.Errorf("Expected 3 total operations, got %d", status.TotalOperations)
	}
}

func TestCreateUIProgressCallback(t *testing.T) {
	var lastCurrent, lastTotal int
	var lastMessage string

	uiFunc := func(current, total int, message string) {
		lastCurrent = current
		lastTotal = total
		lastMessage = message
	}

	callback := CreateUIProgressCallback(uiFunc)
	if callback == nil {
		t.Fatal("Expected callback to be created")
	}

	// Test callback
	status := &Status{
		Current: 50,
		Total:   100,
		Message: "test message",
	}
	callback(status)

	if lastCurrent != 50 {
		t.Errorf("Expected current 50, got %d", lastCurrent)
	}
	if lastTotal != 100 {
		t.Errorf("Expected total 100, got %d", lastTotal)
	}
	if lastMessage != "test message" {
		t.Errorf("Expected message 'test message', got '%s'", lastMessage)
	}
}

func TestCreateVerboseProgressCallback(t *testing.T) {
	var lastFormat string
	var lastArgs []interface{}

	logFunc := func(format string, args ...interface{}) {
		lastFormat = format
		lastArgs = args
	}

	callback := CreateVerboseProgressCallback(logFunc)
	if callback == nil {
		t.Fatal("Expected callback to be created")
	}

	// Test callback
	status := &Status{
		Operation:   "test-op",
		Current:     50,
		Total:       100,
		Message:     "test message",
		Stage:       "processing",
		Percentage:  50.0,
		ElapsedTime: time.Second,
	}
	callback(status)

	if lastFormat == "" {
		t.Error("Expected log format to be set")
	}
	if len(lastArgs) == 0 {
		t.Error("Expected log args to be set")
	}
}

func TestCreateJSONProgressCallback(t *testing.T) {
	var lastData []byte

	outputFunc := func(data []byte) {
		lastData = data
	}

	callback := CreateJSONProgressCallback(outputFunc)
	if callback == nil {
		t.Fatal("Expected callback to be created")
	}

	// Test callback
	status := &Status{
		Operation:  "test-op",
		Current:    50,
		Total:      100,
		Message:    "test message",
		Stage:      "processing",
		Percentage: 50.0,
	}
	callback(status)

	if lastData == nil {
		t.Error("Expected JSON data to be output")
	}
	if len(lastData) == 0 {
		t.Error("Expected JSON data to be non-empty")
	}
}

func TestUpdateProgressWithStage(t *testing.T) {
	tracker := NewTracker()
	tracker.Start("test-operation", 100)

	UpdateProgressWithStage(tracker, 50, StageProcessing, "halfway done")

	status := tracker.GetProgress()
	if status.Current != 50 {
		t.Errorf("Expected current 50, got %d", status.Current)
	}
	expectedMessage := "[processing] halfway done"
	if status.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, status.Message)
	}
}

func TestFinishProgressWithResult(t *testing.T) {
	tracker := NewTracker()
	tracker.Start("test-operation", 100)

	// Test successful completion
	FinishProgressWithResult(tracker, true, "install", "Test::Module", nil)

	if tracker.IsRunning() {
		t.Error("Expected tracker to be finished")
	}

	status := tracker.GetProgress()
	if status.Percentage != 100.0 {
		t.Errorf("Expected percentage 100.0, got %f", status.Percentage)
	}
}

func TestEstimateOperationSteps(t *testing.T) {
	tests := []struct {
		operationType string
		itemCount     int
		expectedSteps int
	}{
		{"install", 2, 14},             // 2 * 7
		{"remove", 3, 9},               // 3 * 3
		{"update", 2, 10},              // 2 * 5
		{"search", 5, 100},             // fixed 100
		{"list", 10, 10},               // 10 * 1
		{"resolve-dependencies", 3, 6}, // 3 * 2
		{"unknown", 5, 5},              // default 5 * 1
	}

	for _, test := range tests {
		steps := EstimateOperationSteps(test.operationType, test.itemCount)
		if steps != test.expectedSteps {
			t.Errorf("For operation '%s' with %d items, expected %d steps, got %d",
				test.operationType, test.itemCount, test.expectedSteps, steps)
		}
	}
}

func TestProgressConfigurations(t *testing.T) {
	// Test default config
	defaultConfig := DefaultProgressConfig()
	if defaultConfig == nil {
		t.Fatal("Expected default config to be created")
	}
	if !defaultConfig.ShowPercentage {
		t.Error("Expected default config to show percentage")
	}

	// Test verbose config
	verboseConfig := VerboseProgressConfig()
	if verboseConfig == nil {
		t.Fatal("Expected verbose config to be created")
	}
	if !verboseConfig.Verbose {
		t.Error("Expected verbose config to be verbose")
	}
	if !verboseConfig.ShowThroughput {
		t.Error("Expected verbose config to show throughput")
	}

	// Test quiet config
	quietConfig := QuietProgressConfig()
	if quietConfig == nil {
		t.Fatal("Expected quiet config to be created")
	}
	if quietConfig.ShowPercentage {
		t.Error("Expected quiet config to not show percentage")
	}

	// Test JSON config
	jsonConfig := JSONProgressConfig()
	if jsonConfig == nil {
		t.Fatal("Expected JSON config to be created")
	}
	if !jsonConfig.JSONOutput {
		t.Error("Expected JSON config to enable JSON output")
	}
	if jsonConfig.ShowPercentage {
		t.Error("Expected JSON config to not show percentage in UI")
	}
}

func TestProgressStageConstants(t *testing.T) {
	// Test that stage constants are defined
	stages := []string{
		StageInitializing,
		StageProcessing,
		StageCompleting,
		StageCompleted,
		StageFailed,
		StageCancelled,
		StageResolving,
		StageDownloading,
		StageExtracting,
		StageBuilding,
		StageTesting,
		StageInstalling,
		StageCleaningUp,
		StageValidating,
		StageSearching,
		StageFiltering,
		StageRemoving,
		StageUpdating,
	}

	for _, stage := range stages {
		if stage == "" {
			t.Error("Expected stage constant to be non-empty")
		}
	}
}
