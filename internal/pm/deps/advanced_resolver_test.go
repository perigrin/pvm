// ABOUTME: Tests for advanced dependency resolver
// ABOUTME: Validates conflict resolution and optimization strategies

package deps

import (
	"context"
	"testing"
	"time"

	"tamarou.com/pvm/internal/cpan"
	"tamarou.com/pvm/internal/errors"
)

// mockProvider implements cpan.Provider for testing
type mockProvider struct {
	modules     map[string]*cpan.ModuleInfo
	versions    map[string][]string
	coreModules map[string]bool
}

func (m *mockProvider) Name() string {
	return "mock"
}

func (m *mockProvider) GetModuleInfo(ctx context.Context, moduleName string) (*cpan.ModuleInfo, error) {
	if info, ok := m.modules[moduleName]; ok {
		return info, nil
	}
	return nil, errors.NewSystemError(ErrModuleNotFound, "Module not found", nil)
}

func (m *mockProvider) SearchModules(ctx context.Context, query string, limit int) (*cpan.SearchResults, error) {
	return &cpan.SearchResults{
		Query:   query,
		Total:   0,
		Results: []*cpan.SearchResult{},
		Source:  "mock",
	}, nil
}

func (m *mockProvider) GetModuleVersions(ctx context.Context, moduleName string) ([]string, error) {
	if versions, ok := m.versions[moduleName]; ok {
		return versions, nil
	}
	return []string{"1.0.0"}, nil
}

func (m *mockProvider) IsCoreModule(ctx context.Context, moduleName string, perlVersion string) (bool, error) {
	return m.coreModules[moduleName], nil
}

func (m *mockProvider) GetDownloadURL(ctx context.Context, moduleName, version string) (string, error) {
	return "", nil
}

func (m *mockProvider) GetDependencies(ctx context.Context, moduleName string) ([]*cpan.Dependency, error) {
	if info, ok := m.modules[moduleName]; ok {
		return info.Dependencies, nil
	}
	return nil, nil
}

func (m *mockProvider) GetAuthorInfo(ctx context.Context, authorID string) (map[string]interface{}, error) {
	return nil, nil
}

func (m *mockProvider) BaseURL() string {
	return "http://mock.cpan.org"
}

func (m *mockProvider) Close() error {
	return nil
}

func TestAdvancedResolver_ConflictResolution(t *testing.T) {
	// Create mock provider with conflicting dependencies
	provider := &mockProvider{
		modules: map[string]*cpan.ModuleInfo{
			"Module::A": {
				Name:    "Module::A",
				Version: "2.0.0",
				Dependencies: []*cpan.Dependency{
					{Name: "Module::C", Version: ">= 1.5", Phase: "runtime"},
				},
			},
			"Module::B": {
				Name:    "Module::B",
				Version: "1.0.0",
				Dependencies: []*cpan.Dependency{
					{Name: "Module::C", Version: "< 1.8", Phase: "runtime"},
				},
			},
			"Module::C": {
				Name:    "Module::C",
				Version: "1.6.0",
			},
			"Test::Module": {
				Name:    "Test::Module",
				Version: "1.0.0",
				Dependencies: []*cpan.Dependency{
					{Name: "Module::A", Version: ">= 2.0", Phase: "runtime"},
					{Name: "Module::B", Version: ">= 1.0", Phase: "runtime"},
				},
			},
		},
		versions: map[string][]string{
			"Module::C": {"1.4.0", "1.5.0", "1.6.0", "1.7.0", "1.8.0", "1.9.0", "2.0.0"},
		},
	}

	resolver, err := NewAdvancedResolver("", 0)
	if err != nil {
		t.Fatalf("Failed to create resolver: %v", err)
	}

	tests := []struct {
		name            string
		strategy        ConflictResolutionStrategy
		expectedVersion string
		expectConflicts bool
	}{
		{
			name:            "Latest Compatible",
			strategy:        StrategyLatestCompatible,
			expectedVersion: "1.7.0", // Latest that satisfies both >= 1.5 and < 1.8
			expectConflicts: false,
		},
		{
			name:            "Minimal Version",
			strategy:        StrategyMinimalVersion,
			expectedVersion: "1.5.0", // Minimal that satisfies both >= 1.5 and < 1.8
			expectConflicts: false,
		},
		{
			name:            "Fail Fast",
			strategy:        StrategyFailFast,
			expectedVersion: "1.7.0", // Should succeed and use default latest logic
			expectConflicts: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := &AdvancedResolutionOptions{
				DependencyResolutionOptions: &DependencyResolutionOptions{
					Provider:    provider,
					IncludeCore: false,
				},
				ConflictStrategy:     tt.strategy,
				OptimizationStrategy: OptimizeSharedDependencies,
			}

			ctx := context.Background()

			// Cast to advanced resolver to use advanced features
			advancedResolver := resolver.(*advancedResolver)
			result, err := advancedResolver.ResolveDependenciesAdvanced(ctx, "Test::Module", options)

			if tt.expectConflicts {
				if err == nil && len(result.Conflicts) == 0 {
					t.Error("Expected conflicts but none found")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if len(result.Conflicts) > 0 {
					t.Errorf("Unexpected conflicts: %v", result.Conflicts)
				}

				// Check resolved version
				if tt.expectedVersion != "" {
					if node, ok := result.Modules["Module::C"]; ok {
						if node.Version != tt.expectedVersion {
							t.Errorf("Expected version %s, got %s", tt.expectedVersion, node.Version)
						}
					} else {
						t.Error("Module::C not found in results")
					}
				}
			}
		})
	}
}

func TestAdvancedResolver_OptimizationStrategies(t *testing.T) {
	// Create a complex dependency tree for testing optimization
	provider := &mockProvider{
		modules: map[string]*cpan.ModuleInfo{
			"App::Main": {
				Name:    "App::Main",
				Version: "1.0.0",
				Dependencies: []*cpan.Dependency{
					{Name: "Module::A", Version: ">= 1.0", Phase: "runtime"},
					{Name: "Module::B", Version: ">= 1.0", Phase: "runtime"},
				},
			},
			"Module::A": {
				Name:    "Module::A",
				Version: "2.0.0",
				Dependencies: []*cpan.Dependency{
					{Name: "Module::Common", Version: ">= 1.0", Phase: "runtime"},
					{Name: "Module::Util", Version: ">= 1.0", Phase: "runtime"},
				},
			},
			"Module::B": {
				Name:    "Module::B",
				Version: "2.0.0",
				Dependencies: []*cpan.Dependency{
					{Name: "Module::Common", Version: ">= 1.5", Phase: "runtime"},
					{Name: "Module::Util", Version: ">= 2.0", Phase: "runtime"},
				},
			},
			"Module::Common": {
				Name:    "Module::Common",
				Version: "2.0.0",
			},
			"Module::Util": {
				Name:    "Module::Util",
				Version: "2.5.0",
			},
		},
		versions: map[string][]string{
			"Module::Common": {"1.0.0", "1.5.0", "2.0.0", "2.5.0"},
			"Module::Util":   {"1.0.0", "1.5.0", "2.0.0", "2.5.0", "3.0.0"},
		},
	}

	resolver, err := NewAdvancedResolver("", 0)
	if err != nil {
		t.Fatalf("Failed to create resolver: %v", err)
	}

	tests := []struct {
		name     string
		strategy OptimizationStrategy
	}{
		{
			name:     "Parallel Resolution",
			strategy: OptimizeParallel,
		},
		{
			name:     "Minimal Tree",
			strategy: OptimizeMinimalTree,
		},
		{
			name:     "Shared Dependencies",
			strategy: OptimizeSharedDependencies,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := &AdvancedResolutionOptions{
				DependencyResolutionOptions: &DependencyResolutionOptions{
					Provider:    provider,
					IncludeCore: false,
				},
				OptimizationStrategy: tt.strategy,
				ParallelWorkers:      2,
			}

			ctx := context.Background()
			start := time.Now()
			result, err := resolver.ResolveDependencies(ctx, "App::Main", options.DependencyResolutionOptions)
			elapsed := time.Since(start)

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Verify all expected modules are present
			expectedModules := []string{"App::Main", "Module::A", "Module::B", "Module::Common", "Module::Util"}
			for _, moduleName := range expectedModules {
				if _, ok := result.Modules[moduleName]; !ok {
					t.Errorf("Expected module %s not found", moduleName)
				}
			}

			t.Logf("Strategy %s completed in %v with %d modules", tt.name, elapsed, len(result.Modules))
		})
	}
}

func TestAdvancedResolver_LockedVersions(t *testing.T) {
	provider := &mockProvider{
		modules: map[string]*cpan.ModuleInfo{
			"App::Test": {
				Name:    "App::Test",
				Version: "1.0.0",
				Dependencies: []*cpan.Dependency{
					{Name: "Module::Locked", Version: ">= 2.0", Phase: "runtime"},
				},
			},
			"Module::Locked": {
				Name:    "Module::Locked",
				Version: "1.5.0", // Default version doesn't satisfy constraint
			},
		},
		versions: map[string][]string{
			"Module::Locked": {"1.0.0", "1.5.0", "2.0.0", "2.5.0", "3.0.0"},
		},
	}

	resolver, err := NewAdvancedResolver("", 0)
	if err != nil {
		t.Fatalf("Failed to create resolver: %v", err)
	}

	// Test with locked version that satisfies constraint
	options := &AdvancedResolutionOptions{
		DependencyResolutionOptions: &DependencyResolutionOptions{
			Provider:    provider,
			IncludeCore: false,
		},
		ConflictStrategy:     StrategyPreferExisting,
		OptimizationStrategy: OptimizeSharedDependencies,
		LockedVersions: map[string]string{
			"Module::Locked": "2.0.0",
		},
	}

	ctx := context.Background()

	// Cast to advanced resolver to use locked versions feature
	advancedResolver := resolver.(*advancedResolver)
	result, err := advancedResolver.ResolveDependenciesAdvanced(ctx, "App::Test", options)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if node, ok := result.Modules["Module::Locked"]; ok {
		if node.Version != "2.0.0" {
			t.Errorf("Expected locked version 2.0.0, got %s", node.Version)
		}
	} else {
		t.Error("Module::Locked not found in results")
	}
}

func TestAdvancedResolver_ExcludedVersions(t *testing.T) {
	provider := &mockProvider{
		modules: map[string]*cpan.ModuleInfo{
			"App::Exclude": {
				Name:    "App::Exclude",
				Version: "1.0.0",
				Dependencies: []*cpan.Dependency{
					{Name: "Module::Exclude", Version: ">= 2.0", Phase: "runtime"},
				},
			},
			"Module::Exclude": {
				Name:    "Module::Exclude",
				Version: "2.5.0",
			},
		},
		versions: map[string][]string{
			"Module::Exclude": {"1.0.0", "2.0.0", "2.1.0", "2.5.0", "3.0.0"},
		},
	}

	resolver, err := NewAdvancedResolver("", 0)
	if err != nil {
		t.Fatalf("Failed to create resolver: %v", err)
	}

	// Exclude specific versions
	options := &AdvancedResolutionOptions{
		DependencyResolutionOptions: &DependencyResolutionOptions{
			Provider:    provider,
			IncludeCore: false,
		},
		ConflictStrategy:     StrategyLatestCompatible,
		OptimizationStrategy: OptimizeSharedDependencies,
		ExcludedVersions: map[string]map[string]bool{
			"Module::Exclude": {
				"2.5.0": true, // Exclude the default version
				"3.0.0": true, // Exclude the latest version
			},
		},
	}

	ctx := context.Background()

	// Cast to advanced resolver to use excluded versions feature
	advancedResolver := resolver.(*advancedResolver)
	result, err := advancedResolver.ResolveDependenciesAdvanced(ctx, "App::Exclude", options)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if node, ok := result.Modules["Module::Exclude"]; ok {
		if node.Version == "2.5.0" || node.Version == "3.0.0" {
			t.Errorf("Got excluded version %s", node.Version)
		}
		if node.Version != "2.1.0" {
			t.Errorf("Expected version 2.1.0, got %s", node.Version)
		}
	} else {
		t.Error("Module::Exclude not found in results")
	}
}
