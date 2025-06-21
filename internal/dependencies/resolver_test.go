// ABOUTME: This file provides comprehensive tests for the dependency resolver functionality.
// ABOUTME: It tests dependency resolution, conflict detection, install plan generation, and caching.

package dependencies

import (
	"context"
	"log"
	"os"
	"reflect"
	"testing"
	"time"

	"tamarou.com/pvm/internal/cpan"
)

// mockProvider implements cpan.Provider for testing
type mockProvider struct {
	modules map[string]*cpan.ModuleInfo
}

func newMockProvider() *mockProvider {
	return &mockProvider{
		modules: make(map[string]*cpan.ModuleInfo),
	}
}

func (mp *mockProvider) GetModuleInfo(ctx context.Context, moduleName string) (*cpan.ModuleInfo, error) {
	if info, ok := mp.modules[moduleName]; ok {
		return info, nil
	}
	return &cpan.ModuleInfo{
		Name:         moduleName,
		Version:      "1.0.0",
		Dependencies: []*cpan.Dependency{},
	}, nil
}

func (mp *mockProvider) addModule(name, version string, deps ...*cpan.Dependency) {
	mp.modules[name] = &cpan.ModuleInfo{
		Name:         name,
		Version:      version,
		Dependencies: deps,
	}
}

func (mp *mockProvider) SearchModules(ctx context.Context, query string, limit int) (*cpan.SearchResults, error) {
	return nil, nil
}

func (mp *mockProvider) GetDependencies(ctx context.Context, moduleName string) ([]*cpan.Dependency, error) {
	if info, ok := mp.modules[moduleName]; ok {
		return info.Dependencies, nil
	}
	return []*cpan.Dependency{}, nil
}

func (mp *mockProvider) GetModuleVersions(ctx context.Context, moduleName string) ([]string, error) {
	if info, ok := mp.modules[moduleName]; ok {
		return []string{info.Version}, nil
	}
	return []string{"1.0.0"}, nil
}

func (mp *mockProvider) GetAuthorInfo(ctx context.Context, authorID string) (map[string]interface{}, error) {
	return make(map[string]interface{}), nil
}

func (mp *mockProvider) IsCoreModule(ctx context.Context, moduleName string, perlVersion string) (bool, error) {
	return false, nil
}

func (mp *mockProvider) Name() string {
	return "MockProvider"
}

func (mp *mockProvider) BaseURL() string {
	return "http://mock.cpan.org"
}

func TestNewDependencyResolver(t *testing.T) {
	provider := newMockProvider()
	logger := log.New(os.Stdout, "", 0)

	resolver := NewDependencyResolver(provider, nil, logger)

	if resolver.provider.Name() != provider.Name() {
		t.Error("Provider not set correctly")
	}
	if resolver.logger != logger {
		t.Error("Logger not set correctly")
	}
	if resolver.cache == nil {
		t.Error("Cache not initialized")
	}
	if resolver.options.CacheEnabled != true {
		t.Error("Cache should be enabled by default")
	}
	if resolver.options.MaxDepth != 50 {
		t.Error("Default max depth should be 50")
	}
}

func TestResolverWithOptions(t *testing.T) {
	provider := newMockProvider()
	logger := log.New(os.Stdout, "", 0)

	options := ResolverOptions{
		CacheEnabled:     false,
		CacheTTL:         1 * time.Hour,
		MaxDepth:         25,
		ConflictStrategy: ConflictStrategyFailFast,
		ParallelEnabled:  false,
		MaxWorkers:       2,
	}

	resolver := NewDependencyResolver(provider, nil, logger).WithOptions(options)

	if resolver.options.CacheEnabled != false {
		t.Error("Cache enabled should be configurable")
	}
	if resolver.options.MaxDepth != 25 {
		t.Error("Max depth should be configurable")
	}
	if resolver.options.ConflictStrategy != ConflictStrategyFailFast {
		t.Error("Conflict strategy should be configurable")
	}
}

func TestResolveDependenciesSimple(t *testing.T) {
	provider := newMockProvider()
	logger := log.New(os.Stdout, "", 0)

	// Set up test modules
	provider.addModule("ModuleA", "1.0.0",
		&cpan.Dependency{Name: "ModuleB", Version: "2.0.0"},
	)
	provider.addModule("ModuleB", "2.0.0")

	resolver := NewDependencyResolver(provider, nil, logger)
	ctx := context.Background()

	graph, err := resolver.ResolveDependencies(ctx, []string{"ModuleA"})
	if err != nil {
		t.Fatalf("Failed to resolve dependencies: %v", err)
	}

	if len(graph.Nodes) != 2 {
		t.Errorf("Expected 2 nodes, got %d", len(graph.Nodes))
	}
	if len(graph.RootNodes) != 1 {
		t.Errorf("Expected 1 root node, got %d", len(graph.RootNodes))
	}

	nodeA := graph.Nodes["ModuleA"]
	if nodeA == nil {
		t.Fatal("ModuleA not found in graph")
	}
	if nodeA.Version != "1.0.0" {
		t.Errorf("Expected ModuleA version 1.0.0, got %s", nodeA.Version)
	}
	if len(nodeA.Dependencies) != 1 {
		t.Errorf("Expected ModuleA to have 1 dependency, got %d", len(nodeA.Dependencies))
	}

	nodeB := graph.Nodes["ModuleB"]
	if nodeB == nil {
		t.Fatal("ModuleB not found in graph")
	}
	if nodeB.Version != "2.0.0" {
		t.Errorf("Expected ModuleB version 2.0.0, got %s", nodeB.Version)
	}
}

func TestResolveDependenciesEmpty(t *testing.T) {
	provider := newMockProvider()
	logger := log.New(os.Stdout, "", 0)

	resolver := NewDependencyResolver(provider, nil, logger)
	ctx := context.Background()

	graph, err := resolver.ResolveDependencies(ctx, []string{})
	if err != nil {
		t.Fatalf("Failed to resolve empty dependencies: %v", err)
	}

	if len(graph.Nodes) != 0 {
		t.Errorf("Expected 0 nodes for empty modules, got %d", len(graph.Nodes))
	}
	if len(graph.RootNodes) != 0 {
		t.Errorf("Expected 0 root nodes for empty modules, got %d", len(graph.RootNodes))
	}
}

func TestResolveDependenciesMultipleRoots(t *testing.T) {
	provider := newMockProvider()
	logger := log.New(os.Stdout, "", 0)

	// Set up test modules
	provider.addModule("ModuleA", "1.0.0")
	provider.addModule("ModuleB", "2.0.0")

	resolver := NewDependencyResolver(provider, nil, logger)
	ctx := context.Background()

	graph, err := resolver.ResolveDependencies(ctx, []string{"ModuleA", "ModuleB"})
	if err != nil {
		t.Fatalf("Failed to resolve dependencies: %v", err)
	}

	if len(graph.Nodes) != 2 {
		t.Errorf("Expected 2 nodes, got %d", len(graph.Nodes))
	}
	if len(graph.RootNodes) != 2 {
		t.Errorf("Expected 2 root nodes, got %d", len(graph.RootNodes))
	}
}

func TestDetectConflicts(t *testing.T) {
	// Create a graph with version conflicts
	graph := &DependencyGraph{
		Nodes: map[string]*DependencyNode{
			"ModuleA": {
				Name:    "ModuleA",
				Version: "1.0.0",
			},
			"ModuleA-v2": {
				Name:    "ModuleA",
				Version: "2.0.0",
			},
		},
	}

	provider := newMockProvider()
	logger := log.New(os.Stdout, "", 0)
	resolver := NewDependencyResolver(provider, nil, logger)

	conflicts, err := resolver.DetectConflicts(graph)
	if err != nil {
		t.Fatalf("Failed to detect conflicts: %v", err)
	}

	if len(conflicts) != 1 {
		t.Errorf("Expected 1 conflict, got %d", len(conflicts))
	}

	conflict := conflicts[0]
	if conflict.Module != "ModuleA" {
		t.Errorf("Expected conflict for ModuleA, got %s", conflict.Module)
	}
	if len(conflict.Versions) != 2 {
		t.Errorf("Expected 2 conflicting versions, got %d", len(conflict.Versions))
	}

	expectedVersions := []string{"1.0.0", "2.0.0"}
	if !reflect.DeepEqual(conflict.Versions, expectedVersions) {
		t.Errorf("Expected versions %v, got %v", expectedVersions, conflict.Versions)
	}
}

func TestSuggestResolutions(t *testing.T) {
	provider := newMockProvider()
	logger := log.New(os.Stdout, "", 0)
	resolver := NewDependencyResolver(provider, nil, logger)

	conflicts := []*Conflict{
		{
			Module:   "ModuleA",
			Type:     ConflictTypeVersion,
			Versions: []string{"1.0.0", "2.0.0"},
		},
	}

	resolutions, err := resolver.SuggestResolutions(conflicts)
	if err != nil {
		t.Fatalf("Failed to suggest resolutions: %v", err)
	}

	if len(resolutions) != 1 {
		t.Errorf("Expected 1 resolution, got %d", len(resolutions))
	}

	resolution := resolutions[0]
	if resolution.Module != "ModuleA" {
		t.Errorf("Expected resolution for ModuleA, got %s", resolution.Module)
	}
	if len(resolution.Suggested) == 0 {
		t.Error("Expected suggested resolution options")
	}
}

func TestCreateInstallPlan(t *testing.T) {
	// Create a simple dependency graph
	nodeB := &DependencyNode{
		Name:         "ModuleB",
		Version:      "2.0.0",
		Dependencies: []*DependencyNode{},
	}

	nodeA := &DependencyNode{
		Name:         "ModuleA",
		Version:      "1.0.0",
		Dependencies: []*DependencyNode{nodeB},
	}

	graph := &DependencyGraph{
		Nodes: map[string]*DependencyNode{
			"ModuleA": nodeA,
			"ModuleB": nodeB,
		},
	}

	provider := newMockProvider()
	logger := log.New(os.Stdout, "", 0)
	resolver := NewDependencyResolver(provider, nil, logger)

	plan, err := resolver.CreateInstallPlan(graph)
	if err != nil {
		t.Fatalf("Failed to create install plan: %v", err)
	}

	if len(plan.Modules) != 2 {
		t.Errorf("Expected 2 modules in plan, got %d", len(plan.Modules))
	}

	// Check that dependencies are installed before dependents
	var bIndex, aIndex int = -1, -1
	for i, module := range plan.Modules {
		if module.Name == "ModuleB" {
			bIndex = i
		}
		if module.Name == "ModuleA" {
			aIndex = i
		}
	}

	if bIndex == -1 || aIndex == -1 {
		t.Fatalf("ModuleA index: %d, ModuleB index: %d", aIndex, bIndex)
	}

	if bIndex >= aIndex {
		t.Errorf("ModuleB should be installed before ModuleA (B at %d, A at %d)", bIndex, aIndex)
	}
}

func TestCompareVersions(t *testing.T) {
	provider := newMockProvider()
	logger := log.New(os.Stdout, "", 0)
	resolver := NewDependencyResolver(provider, nil, logger)

	tests := []struct {
		v1       string
		v2       string
		expected int
	}{
		{"1.0.0", "1.0.0", 0},
		{"1.0.0", "1.0.1", -1},
		{"1.0.1", "1.0.0", 1},
		{"1.0", "1.0.0", 0}, // "1.0" and "1.0.0" should be equal
		{"2.0.0", "1.9.9", 1},
		{"1.10.0", "1.9.0", 1},
	}

	for _, test := range tests {
		result := resolver.compareVersions(test.v1, test.v2)
		if result != test.expected {
			t.Errorf("compareVersions(%s, %s) = %d, expected %d", test.v1, test.v2, result, test.expected)
		}
	}
}

func TestFindLatestVersion(t *testing.T) {
	provider := newMockProvider()
	logger := log.New(os.Stdout, "", 0)
	resolver := NewDependencyResolver(provider, nil, logger)

	versions := []string{"1.0.0", "2.0.0", "1.5.0", "1.10.0"}
	latest := resolver.findLatestVersion(versions)

	expected := "2.0.0"
	if latest != expected {
		t.Errorf("findLatestVersion(%v) = %s, expected %s", versions, latest, expected)
	}
}

func TestFindMinimalVersion(t *testing.T) {
	provider := newMockProvider()
	logger := log.New(os.Stdout, "", 0)
	resolver := NewDependencyResolver(provider, nil, logger)

	versions := []string{"1.0.0", "2.0.0", "1.5.0", "1.10.0"}
	minimal := resolver.findMinimalVersion(versions)

	expected := "1.0.0"
	if minimal != expected {
		t.Errorf("findMinimalVersion(%v) = %s, expected %s", versions, minimal, expected)
	}
}

func TestDependencyCache(t *testing.T) {
	cache := newDependencyCache()

	// Test empty cache
	if cache.get("test") != nil {
		t.Error("Expected nil for non-existent cache entry")
	}

	// Test put and get
	node := &DependencyNode{
		Name:    "TestModule",
		Version: "1.0.0",
	}
	cache.put("test", node)

	retrieved := cache.get("test")
	if retrieved == nil {
		t.Fatal("Expected cached node to be retrieved")
	}
	if retrieved.Name != "TestModule" {
		t.Errorf("Expected TestModule, got %s", retrieved.Name)
	}

	// Test cache size
	if cache.size() != 1 {
		t.Errorf("Expected cache size 1, got %d", cache.size())
	}

	// Test clear
	cache.clear()
	if cache.size() != 0 {
		t.Errorf("Expected cache size 0 after clear, got %d", cache.size())
	}
}

func TestCacheExpiration(t *testing.T) {
	cache := newDependencyCache()
	cache.setTTL(1 * time.Millisecond)

	node := &DependencyNode{
		Name:    "TestModule",
		Version: "1.0.0",
	}
	cache.put("test", node)

	// Immediately retrieve - should exist
	if cache.get("test") == nil {
		t.Error("Expected cached node to exist immediately after put")
	}

	// Wait for expiration
	time.Sleep(2 * time.Millisecond)

	// Should be expired now
	if cache.get("test") != nil {
		t.Error("Expected cached node to be expired")
	}
}

func TestMaxDepthExceeded(t *testing.T) {
	provider := newMockProvider()
	logger := log.New(os.Stdout, "", 0)

	// Create a deep dependency chain
	provider.addModule("Module1", "1.0.0",
		&cpan.Dependency{Name: "Module2", Version: "1.0.0"},
	)
	provider.addModule("Module2", "1.0.0",
		&cpan.Dependency{Name: "Module3", Version: "1.0.0"},
	)

	resolver := NewDependencyResolver(provider, nil, logger).WithOptions(ResolverOptions{
		MaxDepth: 1, // Very low depth limit
	})

	ctx := context.Background()
	_, err := resolver.ResolveDependencies(ctx, []string{"Module1"})

	if err == nil {
		t.Error("Expected error for exceeded max depth")
	}
	if !contains(err.Error(), "maximum dependency depth exceeded") {
		t.Errorf("Expected max depth error, got: %v", err)
	}
}

func TestConflictStrategies(t *testing.T) {
	provider := newMockProvider()
	logger := log.New(os.Stdout, "", 0)

	tests := []struct {
		strategy ConflictStrategy
		name     string
	}{
		{ConflictStrategyFailFast, "FailFast"},
		{ConflictStrategyLatestCompatible, "LatestCompatible"},
		{ConflictStrategyMinimalVersion, "MinimalVersion"},
		{ConflictStrategyPreferExisting, "PreferExisting"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resolver := NewDependencyResolver(provider, nil, logger).WithOptions(ResolverOptions{
				ConflictStrategy: test.strategy,
			})

			conflicts := []*Conflict{
				{
					Module:   "TestModule",
					Type:     ConflictTypeVersion,
					Versions: []string{"1.0.0", "2.0.0"},
				},
			}

			resolutions, err := resolver.SuggestResolutions(conflicts)
			if err != nil {
				t.Fatalf("Failed to suggest resolutions for %s: %v", test.name, err)
			}

			if len(resolutions) != 1 {
				t.Errorf("Expected 1 resolution for %s, got %d", test.name, len(resolutions))
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || s[0:len(substr)] == substr || contains(s[1:], substr))
}
