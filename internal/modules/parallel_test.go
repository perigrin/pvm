// ABOUTME: Tests for parallel installation coordination functionality
// ABOUTME: Validates dependency resolution and worker pool management

package modules

import (
	"context"
	"testing"
	"time"
)

// MockInstaller implements ModuleInstaller for testing
type MockInstaller struct {
	installResults map[string]*InstallResult
	installDelay   time.Duration
}

func NewMockInstaller() *MockInstaller {
	return &MockInstaller{
		installResults: make(map[string]*InstallResult),
		installDelay:   10 * time.Millisecond,
	}
}

func (m *MockInstaller) InstallModule(ctx context.Context, module string, opts InstallOptions) (*InstallResult, error) {
	if result, exists := m.installResults[module]; exists {
		time.Sleep(m.installDelay)
		return result, nil
	}

	// Default successful result
	time.Sleep(m.installDelay)
	return &InstallResult{
		ModuleName: module,
		Version:    "1.0.0",
		Success:    true,
		Duration:   m.installDelay,
	}, nil
}

func (m *MockInstaller) InstallBatch(ctx context.Context, modules []string, opts InstallOptions) ([]*InstallResult, error) {
	results := make([]*InstallResult, len(modules))
	for i, module := range modules {
		result, err := m.InstallModule(ctx, module, opts)
		if err != nil {
			return nil, err
		}
		results[i] = result
	}
	return results, nil
}

func (m *MockInstaller) SetResult(module string, result *InstallResult) {
	m.installResults[module] = result
}

// MockProgressTracker implements ParallelProgressTracker for testing
type MockProgressTracker struct {
	operations []string
	updates    map[string][]string
	finished   bool
}

func NewMockProgressTracker() *MockProgressTracker {
	return &MockProgressTracker{
		updates: make(map[string][]string),
	}
}

func (m *MockProgressTracker) StartParallel(operations []string) {
	m.operations = operations
}

func (m *MockProgressTracker) UpdateOperation(id string, status OperationStatus, message string) {
	if m.updates[id] == nil {
		m.updates[id] = []string{}
	}
	m.updates[id] = append(m.updates[id], message)
}

func (m *MockProgressTracker) FinishParallel(results []*OperationResult) {
	m.finished = true
}

func TestNewParallelCoordinator(t *testing.T) {
	installer := NewMockInstaller()
	tracker := NewMockProgressTracker()

	coordinator := NewParallelCoordinator(installer, 2, tracker)

	if coordinator == nil {
		t.Fatal("NewParallelCoordinator returned nil")
	}

	if coordinator.GetWorkerCount() != 2 {
		t.Errorf("Expected 2 workers, got %d", coordinator.GetWorkerCount())
	}
}

func TestNewParallelCoordinatorDefaultWorkers(t *testing.T) {
	installer := NewMockInstaller()
	tracker := NewMockProgressTracker()

	coordinator := NewParallelCoordinator(installer, 0, tracker)

	if coordinator.GetWorkerCount() != 4 {
		t.Errorf("Expected default 4 workers, got %d", coordinator.GetWorkerCount())
	}
}

func TestInstallModulesEmpty(t *testing.T) {
	installer := NewMockInstaller()
	tracker := NewMockProgressTracker()
	coordinator := NewParallelCoordinator(installer, 2, tracker)

	ctx := context.Background()
	opts := InstallOptions{}

	results, err := coordinator.InstallModules(ctx, []string{}, opts)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 results, got %d", len(results))
	}
}

func TestInstallModulesSingle(t *testing.T) {
	installer := NewMockInstaller()
	tracker := NewMockProgressTracker()
	coordinator := NewParallelCoordinator(installer, 2, tracker)

	ctx := context.Background()
	opts := InstallOptions{
		Parallel: true,
		Workers:  2,
	}

	results, err := coordinator.InstallModules(ctx, []string{"Test::Module"}, opts)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	if results[0].ModuleName != "Test::Module" {
		t.Errorf("Expected module name 'Test::Module', got '%s'", results[0].ModuleName)
	}

	if !results[0].Success {
		t.Errorf("Expected successful installation")
	}
}

func TestInstallModulesMultiple(t *testing.T) {
	installer := NewMockInstaller()
	tracker := NewMockProgressTracker()
	coordinator := NewParallelCoordinator(installer, 2, tracker)

	modules := []string{"DBI", "Moose", "Dancer2"}
	ctx := context.Background()
	opts := InstallOptions{
		Parallel: true,
		Workers:  2,
	}

	results, err := coordinator.InstallModules(ctx, modules, opts)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(results) != len(modules) {
		t.Errorf("Expected %d results, got %d", len(modules), len(results))
	}

	// Verify all modules were processed
	resultNames := make(map[string]bool)
	for _, result := range results {
		resultNames[result.ModuleName] = true
		if !result.Success {
			t.Errorf("Module %s failed to install", result.ModuleName)
		}
	}

	for _, module := range modules {
		if !resultNames[module] {
			t.Errorf("Module %s missing from results", module)
		}
	}
}

func TestProgressTracking(t *testing.T) {
	installer := NewMockInstaller()
	tracker := NewMockProgressTracker()
	coordinator := NewParallelCoordinator(installer, 2, tracker)

	modules := []string{"DBI", "Moose"}
	ctx := context.Background()
	opts := InstallOptions{
		Parallel: true,
		Workers:  2,
	}

	_, err := coordinator.InstallModules(ctx, modules, opts)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify progress tracking was called
	if len(tracker.operations) != 2 {
		t.Errorf("Expected 2 operations tracked, got %d", len(tracker.operations))
	}

	if !tracker.finished {
		t.Errorf("Expected progress tracking to be finished")
	}
}

func TestResolveDependencies(t *testing.T) {
	installer := NewMockInstaller()
	tracker := NewMockProgressTracker()
	coordinator := NewParallelCoordinator(installer, 2, tracker)

	modules := []string{"DBI", "Moose"}
	graph, err := coordinator.ResolveDependencies(modules)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if graph == nil {
		t.Fatal("ResolveDependencies returned nil graph")
	}

	if len(graph.Nodes) != 2 {
		t.Errorf("Expected 2 nodes in graph, got %d", len(graph.Nodes))
	}

	for _, module := range modules {
		if _, exists := graph.Nodes[module]; !exists {
			t.Errorf("Module %s missing from dependency graph", module)
		}
	}
}

func TestCreateInstallPlan(t *testing.T) {
	installer := NewMockInstaller()
	tracker := NewMockProgressTracker()
	coordinator := NewParallelCoordinator(installer, 2, tracker)

	graph := &DependencyGraph{
		Nodes: map[string]*DependencyNode{
			"DBI":   {Name: "DBI"},
			"Moose": {Name: "Moose"},
		},
		Edges: map[string][]string{
			"DBI":   {},
			"Moose": {},
		},
	}

	plan, err := coordinator.CreateInstallPlan(graph)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if plan == nil {
		t.Fatal("CreateInstallPlan returned nil plan")
	}

	if len(plan.Modules) != 2 {
		t.Errorf("Expected 2 modules in plan, got %d", len(plan.Modules))
	}
}

func TestExecuteInstallPlan(t *testing.T) {
	installer := NewMockInstaller()
	tracker := NewMockProgressTracker()
	coordinator := NewParallelCoordinator(installer, 2, tracker)

	plan := &InstallPlan{
		Modules:           []string{"DBI", "Moose"},
		Dependencies:      map[string][]string{},
		InstallationOrder: []string{"DBI", "Moose"},
		ParallelBatches:   [][]string{{"DBI", "Moose"}},
	}

	ctx := context.Background()
	results, err := coordinator.ExecuteInstallPlan(ctx, plan)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
}

func TestSetWorkerCount(t *testing.T) {
	installer := NewMockInstaller()
	tracker := NewMockProgressTracker()
	coordinator := NewParallelCoordinator(installer, 2, tracker)

	coordinator.SetWorkerCount(8)

	if coordinator.GetWorkerCount() != 8 {
		t.Errorf("Expected 8 workers, got %d", coordinator.GetWorkerCount())
	}

	// Test invalid worker count (should be ignored)
	coordinator.SetWorkerCount(0)

	if coordinator.GetWorkerCount() != 8 {
		t.Errorf("Expected worker count to remain 8, got %d", coordinator.GetWorkerCount())
	}
}

func TestInstallModulesWithFailure(t *testing.T) {
	installer := NewMockInstaller()
	tracker := NewMockProgressTracker()
	coordinator := NewParallelCoordinator(installer, 2, tracker)

	// Set up a failing module
	installer.SetResult("FailingModule", &InstallResult{
		ModuleName: "FailingModule",
		Version:    "",
		Success:    false,
		Duration:   10 * time.Millisecond,
		Errors:     []string{"Installation failed"},
	})

	modules := []string{"DBI", "FailingModule"}
	ctx := context.Background()
	opts := InstallOptions{
		Parallel: true,
		Workers:  2,
	}

	results, err := coordinator.InstallModules(ctx, modules, opts)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	// Find the failing module result
	var failingResult *InstallResult
	for _, result := range results {
		if result.ModuleName == "FailingModule" {
			failingResult = result
			break
		}
	}

	if failingResult == nil {
		t.Fatal("FailingModule result not found")
	}

	if failingResult.Success {
		t.Errorf("Expected FailingModule to fail")
	}

	if len(failingResult.Errors) == 0 {
		t.Errorf("Expected error information in failing result")
	}
}
