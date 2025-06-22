// ABOUTME: Tests for core module management interfaces and types
// ABOUTME: Validates interface compliance and type operations

package modules

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

// TestModuleJSON tests JSON marshaling/unmarshaling of Module
func TestModuleJSON(t *testing.T) {
	original := &Module{
		Name:             "Test::Module",
		Version:          "1.0.0",
		Description:      "A test module",
		Author:           "Test Author",
		Path:             "/path/to/module.pm",
		InstallationTime: time.Now().Truncate(time.Second),
		CoreModule:       false,
		Dependencies:     []string{"Dep1", "Dep2"},
	}

	// Marshal to JSON
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal Module: %v", err)
	}

	// Unmarshal from JSON
	var unmarshaled Module
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal Module: %v", err)
	}

	// Verify fields
	if unmarshaled.Name != original.Name {
		t.Errorf("Name mismatch: got %s, want %s", unmarshaled.Name, original.Name)
	}
	if unmarshaled.Version != original.Version {
		t.Errorf("Version mismatch: got %s, want %s", unmarshaled.Version, original.Version)
	}
	if unmarshaled.Description != original.Description {
		t.Errorf("Description mismatch: got %s, want %s", unmarshaled.Description, original.Description)
	}
	if !unmarshaled.InstallationTime.Equal(original.InstallationTime) {
		t.Errorf("InstallationTime mismatch: got %v, want %v", unmarshaled.InstallationTime, original.InstallationTime)
	}
}

// TestInstallResultJSON tests JSON marshaling/unmarshaling of InstallResult
func TestInstallResultJSON(t *testing.T) {
	original := &InstallResult{
		ModuleName:   "Test::Module",
		Version:      "1.0.0",
		Success:      true,
		Duration:     5 * time.Second,
		Dependencies: []string{"Dep1", "Dep2"},
		Warnings:     []string{"Warning 1"},
		Errors:       []string{},
		Path:         "/path/to/module.pm",
	}

	// Marshal to JSON
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal InstallResult: %v", err)
	}

	// Unmarshal from JSON
	var unmarshaled InstallResult
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal InstallResult: %v", err)
	}

	// Verify fields
	if unmarshaled.ModuleName != original.ModuleName {
		t.Errorf("ModuleName mismatch: got %s, want %s", unmarshaled.ModuleName, original.ModuleName)
	}
	if unmarshaled.Success != original.Success {
		t.Errorf("Success mismatch: got %v, want %v", unmarshaled.Success, original.Success)
	}
	if unmarshaled.Duration != original.Duration {
		t.Errorf("Duration mismatch: got %v, want %v", unmarshaled.Duration, original.Duration)
	}
}

// TestOperationStatusString tests the String method of OperationStatus
func TestOperationStatusString(t *testing.T) {
	tests := []struct {
		status   OperationStatus
		expected string
	}{
		{StatusPending, "pending"},
		{StatusRunning, "running"},
		{StatusCompleted, "completed"},
		{StatusFailed, "failed"},
		{StatusCancelled, "cancelled"},
		{OperationStatus(999), "unknown"},
	}

	for _, test := range tests {
		if got := test.status.String(); got != test.expected {
			t.Errorf("OperationStatus(%d).String() = %s, want %s", int(test.status), got, test.expected)
		}
	}
}

// TestModuleFilterDefaults tests default values for ModuleFilter
func TestModuleFilterDefaults(t *testing.T) {
	filter := ModuleFilter{}

	// Verify default values
	if filter.Pattern != "" {
		t.Errorf("Default Pattern should be empty, got %s", filter.Pattern)
	}
	if filter.IncludeCore {
		t.Error("Default IncludeCore should be false")
	}
	if filter.IncludeDev {
		t.Error("Default IncludeDev should be false")
	}
	if filter.LatestOnly {
		t.Error("Default LatestOnly should be false")
	}
}

// TestInstallOptionsDefaults tests default values for InstallOptions
func TestInstallOptionsDefaults(t *testing.T) {
	opts := InstallOptions{}

	// Verify default values
	if opts.Force {
		t.Error("Default Force should be false")
	}
	if opts.RunTests {
		t.Error("Default RunTests should be false")
	}
	if opts.SkipDependencies {
		t.Error("Default SkipDependencies should be false")
	}
	if opts.Verbose {
		t.Error("Default Verbose should be false")
	}
	if opts.Cleanup {
		t.Error("Default Cleanup should be false")
	}
	if opts.Workers != 0 {
		t.Errorf("Default Workers should be 0, got %d", opts.Workers)
	}
}

// mockModuleManager implements ModuleManager for testing
type mockModuleManager struct {
	modules []Module
}

func (m *mockModuleManager) List(ctx context.Context, filter ModuleFilter) ([]*Module, error) {
	var result []*Module
	for i := range m.modules {
		if filter.Pattern == "" || filter.Pattern == m.modules[i].Name {
			result = append(result, &m.modules[i])
		}
	}
	return result, nil
}

func (m *mockModuleManager) Install(ctx context.Context, modules []string, opts InstallOptions) error {
	return nil
}

func (m *mockModuleManager) Remove(ctx context.Context, modules []string) error {
	return nil
}

func (m *mockModuleManager) Update(ctx context.Context, modules []string) error {
	return nil
}

// TestModuleManagerInterface tests that mockModuleManager implements ModuleManager
func TestModuleManagerInterface(t *testing.T) {
	var _ ModuleManager = &mockModuleManager{}

	manager := &mockModuleManager{
		modules: []Module{
			{Name: "Module1", Version: "1.0.0"},
			{Name: "Module2", Version: "2.0.0"},
		},
	}

	ctx := context.Background()

	// Test List method
	modules, err := manager.List(ctx, ModuleFilter{})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(modules) != 2 {
		t.Errorf("Expected 2 modules, got %d", len(modules))
	}

	// Test List with filter
	modules, err = manager.List(ctx, ModuleFilter{Pattern: "Module1"})
	if err != nil {
		t.Fatalf("List with filter failed: %v", err)
	}
	if len(modules) != 1 || modules[0].Name != "Module1" {
		t.Errorf("Filter failed: expected 1 module 'Module1', got %d modules", len(modules))
	}
}

// mockModuleInstaller implements ModuleInstaller for testing
type mockModuleInstaller struct{}

func (m *mockModuleInstaller) InstallModule(ctx context.Context, module string, opts InstallOptions) (*InstallResult, error) {
	return &InstallResult{
		ModuleName: module,
		Version:    "1.0.0",
		Success:    true,
		Duration:   time.Second,
	}, nil
}

func (m *mockModuleInstaller) InstallBatch(ctx context.Context, modules []string, opts InstallOptions) ([]*InstallResult, error) {
	var results []*InstallResult
	for _, module := range modules {
		result, err := m.InstallModule(ctx, module, opts)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, nil
}

// TestModuleInstallerInterface tests that mockModuleInstaller implements ModuleInstaller
func TestModuleInstallerInterface(t *testing.T) {
	var _ ModuleInstaller = &mockModuleInstaller{}

	installer := &mockModuleInstaller{}
	ctx := context.Background()
	opts := InstallOptions{}

	// Test InstallModule
	result, err := installer.InstallModule(ctx, "Test::Module", opts)
	if err != nil {
		t.Fatalf("InstallModule failed: %v", err)
	}
	if result.ModuleName != "Test::Module" {
		t.Errorf("Expected module name 'Test::Module', got %s", result.ModuleName)
	}
	if !result.Success {
		t.Error("Expected successful installation")
	}

	// Test InstallBatch
	modules := []string{"Module1", "Module2"}
	results, err := installer.InstallBatch(ctx, modules, opts)
	if err != nil {
		t.Fatalf("InstallBatch failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
}

// mockProgressTracker implements ProgressTracker for testing
type mockProgressTracker struct {
	started   bool
	operation string
	total     int
	current   int
	message   string
	finished  bool
}

func (m *mockProgressTracker) Start(operation string, total int) {
	m.started = true
	m.operation = operation
	m.total = total
	m.current = 0
	m.finished = false
}

func (m *mockProgressTracker) Update(current int, message string) {
	m.current = current
	m.message = message
}

func (m *mockProgressTracker) Finish(result *OperationResult) {
	m.finished = true
}

// TestProgressTrackerInterface tests that mockProgressTracker implements ProgressTracker
func TestProgressTrackerInterface(t *testing.T) {
	var _ ProgressTracker = &mockProgressTracker{}

	tracker := &mockProgressTracker{}

	// Test Start
	tracker.Start("install", 5)
	if !tracker.started {
		t.Error("Expected tracker to be started")
	}
	if tracker.operation != "install" {
		t.Errorf("Expected operation 'install', got %s", tracker.operation)
	}
	if tracker.total != 5 {
		t.Errorf("Expected total 5, got %d", tracker.total)
	}

	// Test Update
	tracker.Update(3, "Installing module 3")
	if tracker.current != 3 {
		t.Errorf("Expected current 3, got %d", tracker.current)
	}
	if tracker.message != "Installing module 3" {
		t.Errorf("Expected message 'Installing module 3', got %s", tracker.message)
	}

	// Test Finish
	result := &OperationResult{Operation: "install", Success: true}
	tracker.Finish(result)
	if !tracker.finished {
		t.Error("Expected tracker to be finished")
	}
}

// BenchmarkModuleJSON benchmarks JSON operations on Module
func BenchmarkModuleJSON(b *testing.B) {
	module := &Module{
		Name:             "Benchmark::Module",
		Version:          "1.0.0",
		Description:      "A benchmark module",
		Author:           "Benchmark Author",
		Path:             "/path/to/benchmark/module.pm",
		InstallationTime: time.Now(),
		CoreModule:       false,
		Dependencies:     []string{"Dep1", "Dep2", "Dep3"},
	}

	b.Run("Marshal", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := json.Marshal(module)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	data, _ := json.Marshal(module)
	b.Run("Unmarshal", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var m Module
			err := json.Unmarshal(data, &m)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
