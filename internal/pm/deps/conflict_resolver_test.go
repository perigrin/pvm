// ABOUTME: Tests for the advanced conflict resolution system
// ABOUTME: Verifies sophisticated dependency conflict handling

package deps

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"tamarou.com/pvm/internal/cpan"
)

// Helper functions for the existing MockProvider

// Test helper functions

func createTestConflict(module string, requirements map[string][]string) *DependencyConflict {
	return &DependencyConflict{
		Module:       module,
		Requirements: requirements,
	}
}

// Helper to create test provider with versions support
func newTestProvider() *TestProvider {
	return &TestProvider{
		modules:  make(map[string]*cpan.ModuleInfo),
		versions: make(map[string][]string),
	}
}

type TestProvider struct {
	modules  map[string]*cpan.ModuleInfo
	versions map[string][]string
}

func (p *TestProvider) GetModuleInfo(ctx context.Context, moduleName string) (*cpan.ModuleInfo, error) {
	if info, ok := p.modules[moduleName]; ok {
		return info, nil
	}
	return &cpan.ModuleInfo{
		Name:    moduleName,
		Version: "1.0.0",
	}, nil
}

func (p *TestProvider) GetModuleVersions(ctx context.Context, moduleName string) ([]string, error) {
	if versions, ok := p.versions[moduleName]; ok {
		return versions, nil
	}
	return []string{"1.0.0", "1.1.0", "1.2.0", "2.0.0"}, nil
}

func (p *TestProvider) SearchModules(ctx context.Context, query string, limit int) (*cpan.SearchResults, error) {
	return &cpan.SearchResults{}, nil
}

func (p *TestProvider) GetDependencies(ctx context.Context, moduleName string) ([]*cpan.Dependency, error) {
	info, err := p.GetModuleInfo(ctx, moduleName)
	if err != nil {
		return nil, err
	}
	return info.Dependencies, nil
}

func (p *TestProvider) GetAuthorInfo(ctx context.Context, authorID string) (map[string]interface{}, error) {
	return make(map[string]interface{}), nil
}

func (p *TestProvider) IsCoreModule(ctx context.Context, moduleName string, perlVersion string) (bool, error) {
	return false, nil
}

func (p *TestProvider) Name() string {
	return "test"
}

func (p *TestProvider) BaseURL() string {
	return "http://test.local"
}

func (p *TestProvider) AddVersions(name string, versions []string) {
	p.versions[name] = versions
}

func createTestOptions(provider cpan.Provider) *StrategyOptions {
	return &StrategyOptions{
		Provider:          provider,
		PinnedVersions:    make(map[string]string),
		ExcludedVersions:  make(map[string]map[string]bool),
		PreferredVersions: make(map[string]string),
		MaxRetries:        3,
		Timeout:           30 * time.Second,
		Verbose:           false,
	}
}

// Test Cases

func TestConflictResolver_NewConflictResolver(t *testing.T) {
	provider := newTestProvider()
	resolver := NewConflictResolver(provider, StrategyFailFast)

	assert.NotNil(t, resolver)
	assert.Equal(t, provider, resolver.provider)
	assert.Equal(t, StrategyFailFast, resolver.strategy)
}

func TestConflictResolver_ResolveConflicts_NoConflicts(t *testing.T) {
	provider := newTestProvider()
	resolver := NewConflictResolver(provider, StrategyMinimalUpgrade)

	result := &DependencyResolutionResult{
		Conflicts: []*DependencyConflict{},
	}

	resResult, err := resolver.ResolveConflicts(context.Background(), result)
	require.NoError(t, err)
	assert.True(t, resResult.Success)
	assert.Empty(t, resResult.ResolvedVersions)
	assert.Empty(t, resResult.RemainingConflicts)
}

func TestConflictResolver_FailFastStrategy(t *testing.T) {
	provider := newTestProvider()
	resolver := NewConflictResolver(provider, StrategyFailFast)

	conflicts := []*DependencyConflict{
		createTestConflict("Module::A", map[string][]string{
			">= 1.0": {"Module::B"},
			"< 1.0":  {"Module::C"},
		}),
	}

	result := &DependencyResolutionResult{
		Conflicts: conflicts,
	}

	resResult, err := resolver.ResolveConflicts(context.Background(), result)
	assert.Error(t, err)
	assert.False(t, resResult.Success)
	assert.Equal(t, 1, len(resResult.RemainingConflicts))
}

func TestConflictResolver_MinimalUpgradeStrategy(t *testing.T) {
	provider := newTestProvider()
	provider.AddVersions("Module::A", []string{"1.0.0", "1.1.0", "1.2.0", "2.0.0"})

	resolver := NewConflictResolver(provider, StrategyMinimalUpgrade)

	conflicts := []*DependencyConflict{
		createTestConflict("Module::A", map[string][]string{
			">= 1.1": {"Module::B"},
			"< 2.0":  {"Module::C"},
		}),
	}

	result := &DependencyResolutionResult{
		Conflicts: conflicts,
	}

	resResult, err := resolver.ResolveConflicts(context.Background(), result)
	require.NoError(t, err)
	assert.True(t, resResult.Success)
	assert.Equal(t, "1.1.0", resResult.ResolvedVersions["Module::A"])
	assert.Empty(t, resResult.RemainingConflicts)
}

func TestConflictResolver_MaximalUpgradeStrategy(t *testing.T) {
	provider := newTestProvider()
	provider.AddVersions("Module::A", []string{"1.0.0", "1.1.0", "1.2.0", "2.0.0"})

	resolver := NewConflictResolver(provider, StrategyMaximalUpgrade)

	conflicts := []*DependencyConflict{
		createTestConflict("Module::A", map[string][]string{
			">= 1.1": {"Module::B"},
			"< 2.0":  {"Module::C"},
		}),
	}

	result := &DependencyResolutionResult{
		Conflicts: conflicts,
	}

	resResult, err := resolver.ResolveConflicts(context.Background(), result)
	require.NoError(t, err)
	assert.True(t, resResult.Success)
	assert.Equal(t, "1.2.0", resResult.ResolvedVersions["Module::A"])
	assert.Empty(t, resResult.RemainingConflicts)
}

func TestConflictResolver_PinnedVersions(t *testing.T) {
	provider := newTestProvider()
	provider.AddVersions("Module::A", []string{"1.0.0", "1.1.0", "1.2.0", "2.0.0"})

	resolver := NewConflictResolver(provider, StrategyMinimalUpgrade)
	resolver.SetPinnedVersions(map[string]string{
		"Module::A": "1.2.0",
	})

	conflicts := []*DependencyConflict{
		createTestConflict("Module::A", map[string][]string{
			">= 1.1": {"Module::B"},
			"< 2.0":  {"Module::C"},
		}),
	}

	result := &DependencyResolutionResult{
		Conflicts: conflicts,
	}

	resResult, err := resolver.ResolveConflicts(context.Background(), result)
	require.NoError(t, err)
	assert.True(t, resResult.Success)
	assert.Equal(t, "1.2.0", resResult.ResolvedVersions["Module::A"])
}

func TestConflictResolver_ExcludedVersions(t *testing.T) {
	provider := newTestProvider()
	provider.AddVersions("Module::A", []string{"1.0.0", "1.1.0", "1.2.0", "2.0.0"})

	resolver := NewConflictResolver(provider, StrategyMaximalUpgrade)
	resolver.SetExcludedVersions(map[string]map[string]bool{
		"Module::A": {
			"1.2.0": true, // Exclude the latest compatible version
		},
	})

	conflicts := []*DependencyConflict{
		createTestConflict("Module::A", map[string][]string{
			">= 1.1": {"Module::B"},
			"< 2.0":  {"Module::C"},
		}),
	}

	result := &DependencyResolutionResult{
		Conflicts: conflicts,
	}

	resResult, err := resolver.ResolveConflicts(context.Background(), result)
	require.NoError(t, err)
	assert.True(t, resResult.Success)
	assert.Equal(t, "1.1.0", resResult.ResolvedVersions["Module::A"]) // Should pick next best
}

func TestConflictResolver_BacktrackingStrategy(t *testing.T) {
	provider := newTestProvider()
	provider.AddVersions("Module::A", []string{"1.0.0", "1.1.0", "2.0.0"})
	provider.AddVersions("Module::B", []string{"1.0.0", "1.5.0", "2.0.0"})

	resolver := NewConflictResolver(provider, StrategyBacktrack)

	// Create a complex conflict scenario that requires backtracking
	conflicts := []*DependencyConflict{
		createTestConflict("Module::A", map[string][]string{
			">= 1.0": {"Root"},
			"< 2.0":  {"Module::B"},
		}),
		createTestConflict("Module::B", map[string][]string{
			">= 1.0": {"Root"},
			"< 2.0":  {"Module::A"},
		}),
	}

	result := &DependencyResolutionResult{
		Conflicts: conflicts,
	}

	resResult, err := resolver.ResolveConflicts(context.Background(), result)
	require.NoError(t, err)
	assert.True(t, resResult.Success)
	assert.Contains(t, resResult.ResolvedVersions, "Module::A")
	assert.Contains(t, resResult.ResolvedVersions, "Module::B")
}

func TestConflictResolver_ImpossibleConflict(t *testing.T) {
	provider := newTestProvider()
	provider.AddVersions("Module::A", []string{"1.0.0", "1.1.0", "1.2.0"})

	resolver := NewConflictResolver(provider, StrategyMaximalUpgrade)

	// Create an impossible conflict
	conflicts := []*DependencyConflict{
		createTestConflict("Module::A", map[string][]string{
			">= 2.0": {"Module::B"},
			"< 1.0":  {"Module::C"},
		}),
	}

	result := &DependencyResolutionResult{
		Conflicts: conflicts,
	}

	resResult, err := resolver.ResolveConflicts(context.Background(), result)
	require.NoError(t, err)
	assert.False(t, resResult.Success)
	assert.Equal(t, 1, len(resResult.RemainingConflicts))
}

func TestConflictResolver_ComplexScenario(t *testing.T) {
	provider := newTestProvider()

	// Set up a complex dependency scenario similar to real CPAN modules
	provider.AddVersions("DBI", []string{"1.631", "1.640", "1.641", "1.642", "1.643"})
	provider.AddVersions("DBD::SQLite", []string{"1.58", "1.60", "1.62", "1.64", "1.66"})
	provider.AddVersions("Test::More", []string{"1.001014", "1.302073", "1.302075", "1.302085"})

	resolver := NewConflictResolver(provider, StrategyMaximalUpgrade)
	resolver.SetVerbose(true)

	conflicts := []*DependencyConflict{
		createTestConflict("DBI", map[string][]string{
			">= 1.631": {"DBD::SQLite"},
			">= 1.640": {"SomeApp"},
		}),
		createTestConflict("Test::More", map[string][]string{
			">= 1.001014": {"DBD::SQLite"},
			">= 1.302073": {"SomeApp"},
		}),
	}

	result := &DependencyResolutionResult{
		Conflicts: conflicts,
	}

	resResult, err := resolver.ResolveConflicts(context.Background(), result)
	require.NoError(t, err)
	assert.True(t, resResult.Success)

	// Should resolve to latest compatible versions
	assert.Equal(t, "1.643", resResult.ResolvedVersions["DBI"])
	assert.Equal(t, "1.302085", resResult.ResolvedVersions["Test::More"])
}

func TestConflictResolver_ConflictExplanation(t *testing.T) {
	provider := newTestProvider()
	provider.AddVersions("Module::A", []string{"1.0.0", "1.1.0", "1.2.0", "2.0.0"})

	resolver := NewConflictResolver(provider, StrategyFailFast)

	conflicts := []*DependencyConflict{
		createTestConflict("Module::A", map[string][]string{
			">= 2.0": {"Module::B"},
			"< 1.0":  {"Module::C"},
		}),
	}

	result := &DependencyResolutionResult{
		Conflicts: conflicts,
	}

	resResult, err := resolver.ResolveConflicts(context.Background(), result)
	assert.Error(t, err)
	assert.False(t, resResult.Success)
	assert.NotEmpty(t, resResult.Explanations)

	explanation := resResult.Explanations[0]
	assert.Equal(t, "Module::A", explanation.Module)
	assert.Contains(t, explanation.ConflictingVersions, ">= 2.0")
	assert.Contains(t, explanation.ConflictingVersions, "< 1.0")
	assert.Equal(t, "", explanation.RecommendedVersion) // No compatible version exists
}

func TestConflictResolver_PerformanceMetrics(t *testing.T) {
	provider := newTestProvider()
	provider.AddVersions("Module::A", []string{"1.0.0", "1.1.0", "1.2.0", "2.0.0"})

	resolver := NewConflictResolver(provider, StrategyMaximalUpgrade)

	conflicts := []*DependencyConflict{
		createTestConflict("Module::A", map[string][]string{
			">= 1.1": {"Module::B"},
			"< 2.0":  {"Module::C"},
		}),
	}

	result := &DependencyResolutionResult{
		Conflicts: conflicts,
	}

	start := time.Now()
	resResult, err := resolver.ResolveConflicts(context.Background(), result)
	end := time.Now()

	require.NoError(t, err)
	assert.True(t, resResult.Success)

	// Check that metrics are populated
	assert.True(t, resResult.Metrics.EndTime.After(resResult.Metrics.StartTime))
	assert.True(t, resResult.Metrics.EndTime.Before(end) || resResult.Metrics.EndTime.Equal(end))
	assert.True(t, resResult.Metrics.StartTime.After(start) || resResult.Metrics.StartTime.Equal(start))
	assert.Equal(t, 1, resResult.Metrics.ConflictsFound)
	assert.Equal(t, 1, resResult.Metrics.ConflictsResolved)
}

// Benchmark tests

func BenchmarkConflictResolver_MinimalStrategy(b *testing.B) {
	provider := newTestProvider()
	provider.AddVersions("Module::A", []string{"1.0.0", "1.1.0", "1.2.0", "2.0.0"})

	resolver := NewConflictResolver(provider, StrategyMinimalUpgrade)

	conflicts := []*DependencyConflict{
		createTestConflict("Module::A", map[string][]string{
			">= 1.1": {"Module::B"},
			"< 2.0":  {"Module::C"},
		}),
	}

	result := &DependencyResolutionResult{
		Conflicts: conflicts,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := resolver.ResolveConflicts(context.Background(), result)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkConflictResolver_BacktrackStrategy(b *testing.B) {
	provider := newTestProvider()
	provider.AddVersions("Module::A", []string{"1.0.0", "1.1.0", "2.0.0"})
	provider.AddVersions("Module::B", []string{"1.0.0", "1.5.0", "2.0.0"})

	resolver := NewConflictResolver(provider, StrategyBacktrack)

	conflicts := []*DependencyConflict{
		createTestConflict("Module::A", map[string][]string{
			">= 1.0": {"Root"},
			"< 2.0":  {"Module::B"},
		}),
		createTestConflict("Module::B", map[string][]string{
			">= 1.0": {"Root"},
			"< 2.0":  {"Module::A"},
		}),
	}

	result := &DependencyResolutionResult{
		Conflicts: conflicts,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := resolver.ResolveConflicts(context.Background(), result)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Test edge cases

func TestConflictResolver_EmptyVersionDomain(t *testing.T) {
	provider := newTestProvider()
	provider.AddVersions("Module::A", []string{}) // Empty versions

	resolver := NewConflictResolver(provider, StrategyMaximalUpgrade)

	conflicts := []*DependencyConflict{
		createTestConflict("Module::A", map[string][]string{
			">= 1.0": {"Module::B"},
		}),
	}

	result := &DependencyResolutionResult{
		Conflicts: conflicts,
	}

	resResult, err := resolver.ResolveConflicts(context.Background(), result)
	require.NoError(t, err)
	assert.False(t, resResult.Success)
	assert.Equal(t, 1, len(resResult.RemainingConflicts))
}

func TestConflictResolver_InvalidConstraints(t *testing.T) {
	provider := newTestProvider()
	provider.AddVersions("Module::A", []string{"1.0.0", "1.1.0", "1.2.0"})

	resolver := NewConflictResolver(provider, StrategyMaximalUpgrade)

	conflicts := []*DependencyConflict{
		createTestConflict("Module::A", map[string][]string{
			"invalid_constraint": {"Module::B"},
		}),
	}

	result := &DependencyResolutionResult{
		Conflicts: conflicts,
	}

	resResult, err := resolver.ResolveConflicts(context.Background(), result)
	require.NoError(t, err)
	assert.False(t, resResult.Success)
	assert.Equal(t, 1, len(resResult.RemainingConflicts))
}
