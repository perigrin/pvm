// ABOUTME: Tests for parallel installation coordination functionality
// ABOUTME: Validates dependency resolution and worker pool management

package modules

import (
	"context"
	"errors"
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

// MockDependencyResolver implements DependencyResolverInterface for testing
type MockDependencyResolver struct {
	should_fail       bool
	fail_message      string
	returned_graph    interface{}
	returned_plan     interface{}
	resolved_modules  []string
	plan_created_from interface{}
}

func NewMockDependencyResolver() *MockDependencyResolver {
	return &MockDependencyResolver{
		should_fail: false,
	}
}

func (m *MockDependencyResolver) ResolveDependencies(ctx context.Context, modules []string) (interface{}, error) {
	m.resolved_modules = modules
	if m.should_fail {
		return nil, errors.New(m.fail_message)
	}

	// Return a mock rich dependency graph
	mockGraph := &MockRichDependencyGraph{
		nodes: make(map[string]interface{}),
	}

	// Create mock nodes for each module
	for _, module := range modules {
		mockGraph.nodes[module] = map[string]interface{}{
			"name":    module,
			"version": "1.0.0",
		}
	}

	m.returned_graph = mockGraph
	return mockGraph, nil
}

func (m *MockDependencyResolver) CreateInstallPlan(dependencyGraph interface{}) (interface{}, error) {
	m.plan_created_from = dependencyGraph
	if m.should_fail {
		return nil, errors.New(m.fail_message)
	}

	// Create a mock install plan with dependency levels
	mockPlan := &MockRichInstallPlan{
		modules:      []interface{}{},
		dependencies: make(map[string][]string),
		levels:       [][]string{{"DBI"}, {"Moose"}}, // Two-level dependency example
	}

	// Extract modules from the graph if possible
	if graph, ok := dependencyGraph.(*MockRichDependencyGraph); ok {
		for name := range graph.nodes {
			mockPlan.modules = append(mockPlan.modules, map[string]interface{}{
				"name": name,
			})
			mockPlan.dependencies[name] = []string{} // No dependencies in simple mock
		}
	}

	m.returned_plan = mockPlan
	return mockPlan, nil
}

func (m *MockDependencyResolver) SetShouldFail(fail bool, message string) {
	m.should_fail = fail
	m.fail_message = message
}

// MockRichDependencyGraph implements RichDependencyGraph for testing
type MockRichDependencyGraph struct {
	nodes map[string]interface{}
	edges []interface{}
}

func (m *MockRichDependencyGraph) GetNodes() map[string]interface{} {
	return m.nodes
}

func (m *MockRichDependencyGraph) GetEdges() []interface{} {
	return m.edges
}

// MockRichInstallPlan implements RichInstallPlan for testing
type MockRichInstallPlan struct {
	modules      []interface{}
	dependencies map[string][]string
	levels       [][]string
}

func (m *MockRichInstallPlan) GetModules() []interface{} {
	return m.modules
}

func (m *MockRichInstallPlan) GetDependencies() map[string][]string {
	return m.dependencies
}

func (m *MockRichInstallPlan) GetLevels() [][]string {
	return m.levels
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

// Tests for advanced dependency resolution integration

func TestResolveDependenciesWithResolver(t *testing.T) {
	installer := NewMockInstaller()
	tracker := NewMockProgressTracker()
	coordinator := NewParallelCoordinator(installer, 2, tracker)
	resolver := NewMockDependencyResolver()

	modules := []string{"DBI", "Moose"}
	graph, err := coordinator.ResolveDependenciesWithResolver(modules, resolver)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if graph == nil {
		t.Fatal("ResolveDependenciesWithResolver returned nil graph")
	}

	// Verify that the resolver was called with correct modules
	if len(resolver.resolved_modules) != 2 {
		t.Errorf("Expected resolver to be called with 2 modules, got %d", len(resolver.resolved_modules))
	}

	for i, module := range modules {
		if resolver.resolved_modules[i] != module {
			t.Errorf("Expected module %s, got %s", module, resolver.resolved_modules[i])
		}
	}

	// Verify graph structure
	if len(graph.Nodes) != 2 {
		t.Errorf("Expected 2 nodes in graph, got %d", len(graph.Nodes))
	}

	for _, module := range modules {
		if _, exists := graph.Nodes[module]; !exists {
			t.Errorf("Module %s missing from dependency graph", module)
		}
	}
}

func TestResolveDependenciesWithResolverFailure(t *testing.T) {
	installer := NewMockInstaller()
	tracker := NewMockProgressTracker()
	coordinator := NewParallelCoordinator(installer, 2, tracker)
	resolver := NewMockDependencyResolver()

	// Configure resolver to fail
	resolver.SetShouldFail(true, "dependency resolution failed")

	modules := []string{"DBI", "Moose"}
	graph, err := coordinator.ResolveDependenciesWithResolver(modules, resolver)

	if err == nil {
		t.Fatal("Expected error from failing resolver")
	}

	if graph != nil {
		t.Errorf("Expected nil graph on resolver failure")
	}

	if err.Error() != "dependency resolution failed: dependency resolution failed" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestResolveDependenciesWithoutResolver(t *testing.T) {
	installer := NewMockInstaller()
	tracker := NewMockProgressTracker()
	coordinator := NewParallelCoordinator(installer, 2, tracker)

	// Test with nil resolver (fallback mode)
	modules := []string{"DBI", "Moose"}
	graph, err := coordinator.ResolveDependenciesWithResolver(modules, nil)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if graph == nil {
		t.Fatal("ResolveDependenciesWithResolver returned nil graph")
	}

	// Verify fallback behavior creates simple graph
	if len(graph.Nodes) != 2 {
		t.Errorf("Expected 2 nodes in fallback graph, got %d", len(graph.Nodes))
	}

	for _, module := range modules {
		node := graph.Nodes[module]
		if node == nil {
			t.Errorf("Module %s missing from fallback graph", module)
			continue
		}
		if len(node.Dependencies) != 0 {
			t.Errorf("Expected no dependencies in fallback graph, got %d for %s", len(node.Dependencies), module)
		}
	}
}

func TestCreateInstallPlanWithResolver(t *testing.T) {
	installer := NewMockInstaller()
	tracker := NewMockProgressTracker()
	coordinator := NewParallelCoordinator(installer, 2, tracker)
	resolver := NewMockDependencyResolver()

	// Create a simple dependency graph
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

	plan, err := coordinator.CreateInstallPlanWithResolver(graph, resolver)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if plan == nil {
		t.Fatal("CreateInstallPlanWithResolver returned nil plan")
	}

	// Verify that resolver was called to create the plan
	if resolver.plan_created_from == nil {
		t.Error("Expected resolver to be called with dependency graph")
	}

	// Verify plan structure uses sophisticated parallel levels
	if len(plan.ParallelBatches) != 2 {
		t.Errorf("Expected 2 parallel batches from advanced resolver, got %d", len(plan.ParallelBatches))
	}

	// First batch should have DBI (no dependencies)
	if len(plan.ParallelBatches[0]) != 1 || plan.ParallelBatches[0][0] != "DBI" {
		t.Errorf("Expected first batch to contain DBI only, got %v", plan.ParallelBatches[0])
	}

	// Second batch should have Moose (depends on DBI in this mock)
	if len(plan.ParallelBatches[1]) != 1 || plan.ParallelBatches[1][0] != "Moose" {
		t.Errorf("Expected second batch to contain Moose only, got %v", plan.ParallelBatches[1])
	}
}

func TestCreateInstallPlanWithResolverFailure(t *testing.T) {
	installer := NewMockInstaller()
	tracker := NewMockProgressTracker()
	coordinator := NewParallelCoordinator(installer, 2, tracker)
	resolver := NewMockDependencyResolver()

	// Configure resolver to fail
	resolver.SetShouldFail(true, "install plan creation failed")

	graph := &DependencyGraph{
		Nodes: map[string]*DependencyNode{
			"DBI": {Name: "DBI"},
		},
		Edges: map[string][]string{
			"DBI": {},
		},
	}

	plan, err := coordinator.CreateInstallPlanWithResolver(graph, resolver)

	if err == nil {
		t.Fatal("Expected error from failing resolver")
	}

	if plan != nil {
		t.Errorf("Expected nil plan on resolver failure")
	}

	if err.Error() != "install plan creation failed: install plan creation failed" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestCreateInstallPlanWithoutResolver(t *testing.T) {
	installer := NewMockInstaller()
	tracker := NewMockProgressTracker()
	coordinator := NewParallelCoordinator(installer, 2, tracker)

	// Test with nil resolver (fallback mode)
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

	plan, err := coordinator.CreateInstallPlanWithResolver(graph, nil)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if plan == nil {
		t.Fatal("CreateInstallPlanWithResolver returned nil plan")
	}

	// Verify fallback behavior creates simple plan
	if len(plan.Modules) != 2 {
		t.Errorf("Expected 2 modules in fallback plan, got %d", len(plan.Modules))
	}

	// Fallback should put all modules in one batch
	if len(plan.ParallelBatches) != 1 {
		t.Errorf("Expected 1 parallel batch in fallback plan, got %d", len(plan.ParallelBatches))
	}

	if len(plan.ParallelBatches[0]) != 2 {
		t.Errorf("Expected 2 modules in fallback batch, got %d", len(plan.ParallelBatches[0]))
	}
}

func TestExecuteInstallPlanWithDependencyBatches(t *testing.T) {
	installer := NewMockInstaller()
	tracker := NewMockProgressTracker()
	coordinator := NewParallelCoordinator(installer, 2, tracker)

	// Create a plan with multiple dependency-aware batches
	plan := &InstallPlan{
		Modules: []string{"DBI", "DBD::SQLite", "Moose", "Dancer2"},
		Dependencies: map[string][]string{
			"DBI":         {},
			"DBD::SQLite": {"DBI"},
			"Moose":       {},
			"Dancer2":     {"Moose"},
		},
		InstallationOrder: []string{"DBI", "Moose", "DBD::SQLite", "Dancer2"},
		ParallelBatches: [][]string{
			{"DBI", "Moose"},           // First batch: modules with no dependencies
			{"DBD::SQLite", "Dancer2"}, // Second batch: modules that depend on first batch
		},
	}

	ctx := context.Background()
	results, err := coordinator.ExecuteInstallPlan(ctx, plan)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(results) != 4 {
		t.Errorf("Expected 4 results, got %d", len(results))
	}

	// Verify all modules were successfully installed
	successCount := 0
	for _, result := range results {
		if result.Success {
			successCount++
		}
	}

	if successCount != 4 {
		t.Errorf("Expected 4 successful installations, got %d", successCount)
	}
}

func TestExecuteInstallPlanWithBatchFailure(t *testing.T) {
	installer := NewMockInstaller()
	tracker := NewMockProgressTracker()
	coordinator := NewParallelCoordinator(installer, 2, tracker)

	// Set up a failing module in the first batch
	installer.SetResult("DBI", &InstallResult{
		ModuleName: "DBI",
		Version:    "",
		Success:    false,
		Duration:   10 * time.Millisecond,
		Errors:     []string{"Database connection failed"},
	})

	// Create a plan with dependency batches
	plan := &InstallPlan{
		Modules: []string{"DBI", "DBD::SQLite"},
		Dependencies: map[string][]string{
			"DBI":         {},
			"DBD::SQLite": {"DBI"},
		},
		InstallationOrder: []string{"DBI", "DBD::SQLite"},
		ParallelBatches: [][]string{
			{"DBI"},         // First batch: DBI (will fail)
			{"DBD::SQLite"}, // Second batch: should not be executed due to failure
		},
	}

	ctx := context.Background()
	results, err := coordinator.ExecuteInstallPlan(ctx, plan)

	// Should return error due to batch failure
	if err == nil {
		t.Error("Expected error due to batch failure")
	}

	// Should only have results from the first batch
	if len(results) != 1 {
		t.Errorf("Expected 1 result from failed batch, got %d", len(results))
	}

	if results[0].Success {
		t.Error("Expected DBI installation to fail")
	}

	if results[0].ModuleName != "DBI" {
		t.Errorf("Expected DBI result, got %s", results[0].ModuleName)
	}
}

func TestBackwardCompatibilityMethods(t *testing.T) {
	installer := NewMockInstaller()
	tracker := NewMockProgressTracker()
	coordinator := NewParallelCoordinator(installer, 2, tracker)

	// Test that the original methods still work (backward compatibility)
	modules := []string{"DBI", "Moose"}

	// Test ResolveDependencies (should call ResolveDependenciesWithResolver with nil)
	graph, err := coordinator.ResolveDependencies(modules)
	if err != nil {
		t.Errorf("Backward compatibility ResolveDependencies failed: %v", err)
	}
	if graph == nil {
		t.Error("Expected non-nil graph from backward compatibility method")
	}

	// Test CreateInstallPlan (should call CreateInstallPlanWithResolver with nil)
	plan, err := coordinator.CreateInstallPlan(graph)
	if err != nil {
		t.Errorf("Backward compatibility CreateInstallPlan failed: %v", err)
	}
	if plan == nil {
		t.Error("Expected non-nil plan from backward compatibility method")
	}

	// Test ExecuteInstallPlan (should work as before)
	ctx := context.Background()
	results, err := coordinator.ExecuteInstallPlan(ctx, plan)
	if err != nil {
		t.Errorf("Backward compatibility ExecuteInstallPlan failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 results from backward compatibility, got %d", len(results))
	}
}
