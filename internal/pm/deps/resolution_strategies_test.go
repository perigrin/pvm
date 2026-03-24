// ABOUTME: Tests for dependency resolution strategies
// ABOUTME: Verifies different conflict resolution approaches

package deps

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFailFastStrategy(t *testing.T) {
	provider := newTestProvider()
	strategy := NewFailFastStrategy()

	assert.Equal(t, "fail-fast", strategy.Name())
	assert.NotEmpty(t, strategy.Description())

	// Test with no conflicts
	result, err := strategy.Resolve(context.Background(), []*DependencyConflict{}, createTestOptions(provider))
	require.NoError(t, err)
	assert.True(t, result.Success)

	// Test with conflicts
	conflicts := []*DependencyConflict{
		createTestConflict("Module::A", map[string][]string{
			">= 2.0": {"Module::B"},
			"< 1.0":  {"Module::C"},
		}),
	}

	result, err = strategy.Resolve(context.Background(), conflicts, createTestOptions(provider))
	assert.Error(t, err)
	assert.False(t, result.Success)
	assert.Equal(t, 1, len(result.UnresolvedConflicts))
	assert.Contains(t, result.Explanation, "Conflicts detected")
}

func TestMinimalUpgradeStrategy(t *testing.T) {
	provider := newTestProvider()
	provider.AddVersions("Module::A", []string{"1.0.0", "1.1.0", "1.2.0", "2.0.0"})

	strategy := NewMinimalUpgradeStrategy()

	assert.Equal(t, "minimal-upgrade", strategy.Name())
	assert.NotEmpty(t, strategy.Description())

	// Test resolvable conflict
	conflicts := []*DependencyConflict{
		createTestConflict("Module::A", map[string][]string{
			">= 1.1": {"Module::B"},
			"< 2.0":  {"Module::C"},
		}),
	}

	result, err := strategy.Resolve(context.Background(), conflicts, createTestOptions(provider))
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, "1.1.0", result.ResolvedVersions["Module::A"]) // Minimal compatible version
	assert.Empty(t, result.UnresolvedConflicts)
}

func TestMaximalUpgradeStrategy(t *testing.T) {
	provider := newTestProvider()
	provider.AddVersions("Module::A", []string{"1.0.0", "1.1.0", "1.2.0", "2.0.0"})

	strategy := NewMaximalUpgradeStrategy()

	assert.Equal(t, "maximal-upgrade", strategy.Name())
	assert.NotEmpty(t, strategy.Description())

	// Test resolvable conflict
	conflicts := []*DependencyConflict{
		createTestConflict("Module::A", map[string][]string{
			">= 1.1": {"Module::B"},
			"< 2.0":  {"Module::C"},
		}),
	}

	result, err := strategy.Resolve(context.Background(), conflicts, createTestOptions(provider))
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, "1.2.0", result.ResolvedVersions["Module::A"]) // Maximal compatible version
	assert.Empty(t, result.UnresolvedConflicts)
}

func TestBacktrackStrategy(t *testing.T) {
	provider := newTestProvider()
	provider.AddVersions("Module::A", []string{"1.0.0", "1.1.0", "1.2.0", "2.0.0"})
	provider.AddVersions("Module::B", []string{"1.0.0", "1.5.0", "2.0.0"})

	strategy := NewBacktrackStrategy(50)

	assert.Equal(t, "backtrack", strategy.Name())
	assert.NotEmpty(t, strategy.Description())

	// Test simple conflict that backtrack can resolve
	conflicts := []*DependencyConflict{
		createTestConflict("Module::A", map[string][]string{
			">= 1.1": {"Module::B"}, // A must be at least 1.1 (required by B)
		}),
	}

	result, err := strategy.Resolve(context.Background(), conflicts, createTestOptions(provider))
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Contains(t, result.ResolvedVersions, "Module::A")
	// Should resolve to a version >= 1.1
	resolvedVersion := result.ResolvedVersions["Module::A"]
	assert.True(t, resolvedVersion == "1.1.0" || resolvedVersion == "1.2.0" || resolvedVersion == "2.0.0")
}

func TestHybridStrategy(t *testing.T) {
	provider := newTestProvider()
	provider.AddVersions("Module::A", []string{"1.0.0", "1.1.0", "1.2.0", "2.0.0"})

	// Create hybrid strategy with fail-fast first, then minimal upgrade
	strategy := NewHybridStrategy(
		NewFailFastStrategy(),
		NewMinimalUpgradeStrategy(),
	)

	assert.Equal(t, "hybrid", strategy.Name())
	assert.NotEmpty(t, strategy.Description())

	// Test with resolvable conflict (should succeed with minimal upgrade after fail-fast fails)
	conflicts := []*DependencyConflict{
		createTestConflict("Module::A", map[string][]string{
			">= 1.1": {"Module::B"},
			"< 2.0":  {"Module::C"},
		}),
	}

	options := createTestOptions(provider)
	options.Verbose = true

	result, err := strategy.Resolve(context.Background(), conflicts, options)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, "1.1.0", result.ResolvedVersions["Module::A"])
	assert.Contains(t, result.Explanation, "minimal-upgrade") // Should indicate which strategy succeeded
}

func TestStrategyWithPinnedVersions(t *testing.T) {
	provider := newTestProvider()
	provider.AddVersions("Module::A", []string{"1.0.0", "1.1.0", "1.2.0", "2.0.0"})

	strategy := NewMaximalUpgradeStrategy()

	conflicts := []*DependencyConflict{
		createTestConflict("Module::A", map[string][]string{
			">= 1.1": {"Module::B"},
			"< 2.0":  {"Module::C"},
		}),
	}

	options := createTestOptions(provider)
	options.PinnedVersions = map[string]string{
		"Module::A": "1.1.0", // Pin to non-maximal version
	}

	result, err := strategy.Resolve(context.Background(), conflicts, options)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, "1.1.0", result.ResolvedVersions["Module::A"]) // Should use pinned version
}

func TestStrategyWithExcludedVersions(t *testing.T) {
	provider := newTestProvider()
	provider.AddVersions("Module::A", []string{"1.0.0", "1.1.0", "1.2.0", "2.0.0"})

	strategy := NewMaximalUpgradeStrategy()

	conflicts := []*DependencyConflict{
		createTestConflict("Module::A", map[string][]string{
			">= 1.1": {"Module::B"},
			"< 2.0":  {"Module::C"},
		}),
	}

	options := createTestOptions(provider)
	options.ExcludedVersions = map[string]map[string]bool{
		"Module::A": {
			"1.2.0": true, // Exclude the maximal version
		},
	}

	result, err := strategy.Resolve(context.Background(), conflicts, options)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, "1.1.0", result.ResolvedVersions["Module::A"]) // Should use next best version
}

func TestStrategyWithStablePreference(t *testing.T) {
	provider := newTestProvider()
	provider.AddVersions("Module::A", []string{"1.0.0", "1.1.0", "1.2.0_01", "1.2.0"}) // 1.2.0_01 is unstable

	strategy := NewMaximalUpgradeStrategy()

	conflicts := []*DependencyConflict{
		createTestConflict("Module::A", map[string][]string{
			">= 1.0": {"Module::B"},
		}),
	}

	options := createTestOptions(provider)
	options.PreferStable = true

	result, err := strategy.Resolve(context.Background(), conflicts, options)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, "1.2.0", result.ResolvedVersions["Module::A"]) // Should prefer stable over unstable _01
}

func TestStrategyWithImpossibleConstraints(t *testing.T) {
	provider := newTestProvider()
	provider.AddVersions("Module::A", []string{"1.0.0", "1.1.0", "1.2.0"})

	strategy := NewMaximalUpgradeStrategy()

	// Create impossible conflict
	conflicts := []*DependencyConflict{
		createTestConflict("Module::A", map[string][]string{
			">= 2.0": {"Module::B"},
			"< 1.0":  {"Module::C"},
		}),
	}

	result, err := strategy.Resolve(context.Background(), conflicts, createTestOptions(provider))
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Equal(t, 1, len(result.UnresolvedConflicts))
}

func TestStrategyWithPinnedVersionConflict(t *testing.T) {
	provider := newTestProvider()
	provider.AddVersions("Module::A", []string{"1.0.0", "1.1.0", "1.2.0"})

	strategy := NewMinimalUpgradeStrategy()

	conflicts := []*DependencyConflict{
		createTestConflict("Module::A", map[string][]string{
			">= 1.2": {"Module::B"},
		}),
	}

	options := createTestOptions(provider)
	options.PinnedVersions = map[string]string{
		"Module::A": "1.0.0", // Pinned version doesn't satisfy constraint
	}

	result, err := strategy.Resolve(context.Background(), conflicts, options)
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Equal(t, 1, len(result.UnresolvedConflicts))
	assert.NotEmpty(t, result.Warnings)
	assert.Contains(t, result.Warnings[0], "Pinned version")
}

func TestStrategyMetrics(t *testing.T) {
	provider := newTestProvider()
	provider.AddVersions("Module::A", []string{"1.0.0", "1.1.0", "1.2.0"})

	strategy := NewMaximalUpgradeStrategy()

	conflicts := []*DependencyConflict{
		createTestConflict("Module::A", map[string][]string{
			">= 1.1": {"Module::B"},
		}),
	}

	start := time.Now()
	result, err := strategy.Resolve(context.Background(), conflicts, createTestOptions(provider))
	end := time.Now()

	require.NoError(t, err)
	assert.True(t, result.Success)

	// Check metrics
	// Allow equal times: Windows timer resolution is ~15ms, fast operations may start and end in same tick
	assert.True(t, !result.Metrics.EndTime.Before(result.Metrics.StartTime))
	assert.True(t, !result.Metrics.EndTime.After(end))
	assert.True(t, !result.Metrics.StartTime.Before(start))
	assert.Equal(t, 1, result.Metrics.ConflictsProcessed)
	assert.Equal(t, 1, result.Metrics.ConflictsResolved)
	assert.Greater(t, result.Metrics.VersionsEvaluated, 0)
}

func TestGetAvailableStrategies(t *testing.T) {
	strategies := GetAvailableStrategies()
	assert.Greater(t, len(strategies), 0)

	// Check that all expected strategies are present
	names := make(map[string]bool)
	for _, strategy := range strategies {
		names[strategy.Name()] = true
	}

	assert.True(t, names["fail-fast"])
	assert.True(t, names["minimal-upgrade"])
	assert.True(t, names["maximal-upgrade"])
	assert.True(t, names["backtrack"])
	assert.True(t, names["hybrid"])
}

func TestGetStrategyByName(t *testing.T) {
	testCases := []struct {
		name     string
		expected string
	}{
		{"fail-fast", "fail-fast"},
		{"failfast", "fail-fast"},
		{"minimal-upgrade", "minimal-upgrade"},
		{"minimal", "minimal-upgrade"},
		{"min", "minimal-upgrade"},
		{"maximal-upgrade", "maximal-upgrade"},
		{"maximal", "maximal-upgrade"},
		{"max", "maximal-upgrade"},
		{"backtrack", "backtrack"},
		{"bt", "backtrack"},
		{"hybrid", "hybrid"},
		{"unknown", "maximal-upgrade"}, // Default
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			strategy := GetStrategyByName(tc.name)
			assert.Equal(t, tc.expected, strategy.Name())
		})
	}
}

func TestIsStableVersion(t *testing.T) {
	testCases := []struct {
		version  string
		expected bool
	}{
		{"1.0.0", true},
		{"1.2.3", true},
		{"2.0", true},
		{"1.0.0_01", false},  // Development release
		{"1.2.3_004", false}, // Development release
		{"1.0-TRIAL", false}, // Trial release
		{"1.0-RC1", false},   // Release candidate
		{"1.0-ALPHA", false}, // Alpha release
		{"1.0-BETA", false},  // Beta release
		{"1.0-DEV", false},   // Development release
		{"1.0-rc1", false},   // Case insensitive
	}

	for _, tc := range testCases {
		t.Run(tc.version, func(t *testing.T) {
			result := isStableVersion(tc.version)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestSortWithStablePreference(t *testing.T) {
	versions := []string{"1.0.0_01", "1.2.0", "1.1.0_02", "1.1.0", "2.0-TRIAL", "2.0.0"}

	sorted := sortWithStablePreference(versions)

	// Should have stable versions first
	assert.Equal(t, "2.0.0", sorted[0])
	assert.Equal(t, "1.2.0", sorted[1])
	assert.Equal(t, "1.1.0", sorted[2])

	// Then unstable versions
	assert.Contains(t, sorted[3:], "2.0-TRIAL")
	assert.Contains(t, sorted[3:], "1.1.0_02")
	assert.Contains(t, sorted[3:], "1.0.0_01")
}

func TestFilterExcludedVersions(t *testing.T) {
	versions := []string{"1.0.0", "1.1.0", "1.2.0", "2.0.0"}
	excluded := map[string]map[string]bool{
		"Module::A": {
			"1.1.0": true,
			"2.0.0": true,
		},
	}

	filtered := filterExcludedVersions("Module::A", versions, excluded)
	assert.Equal(t, []string{"1.0.0", "1.2.0"}, filtered)

	// Test module with no exclusions
	filtered = filterExcludedVersions("Module::B", versions, excluded)
	assert.Equal(t, versions, filtered)
}

func TestSatisfiesAllConstraints(t *testing.T) {
	conflict := createTestConflict("Module::A", map[string][]string{
		">= 1.1": {"Module::B"},
		"< 2.0":  {"Module::C"},
	})

	options := createTestOptions(newTestProvider())

	// Test satisfying version
	assert.True(t, satisfiesAllConstraints("1.5.0", conflict, options))

	// Test non-satisfying versions
	assert.False(t, satisfiesAllConstraints("1.0.0", conflict, options)) // Too low
	assert.False(t, satisfiesAllConstraints("2.0.0", conflict, options)) // Too high
}

// Benchmark tests

func BenchmarkMinimalUpgradeStrategy(b *testing.B) {
	provider := newTestProvider()
	provider.AddVersions("Module::A", []string{"1.0.0", "1.1.0", "1.2.0", "2.0.0"})

	strategy := NewMinimalUpgradeStrategy()
	conflicts := []*DependencyConflict{
		createTestConflict("Module::A", map[string][]string{
			">= 1.1": {"Module::B"},
			"< 2.0":  {"Module::C"},
		}),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, err := strategy.Resolve(context.Background(), conflicts, createTestOptions(provider))
		if err != nil || !result.Success {
			b.Fatal("Strategy failed")
		}
	}
}

func BenchmarkBacktrackStrategy(b *testing.B) {
	provider := newTestProvider()
	provider.AddVersions("Module::A", []string{"1.0.0", "1.1.0", "2.0.0"})
	provider.AddVersions("Module::B", []string{"1.0.0", "1.5.0", "2.0.0"})

	strategy := NewBacktrackStrategy(50)
	conflicts := []*DependencyConflict{
		createTestConflict("Module::A", map[string][]string{
			">= 1.1": {"Module::B"},
		}),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, err := strategy.Resolve(context.Background(), conflicts, createTestOptions(provider))
		if err != nil || !result.Success {
			b.Fatal("Strategy failed")
		}
	}
}

func BenchmarkHybridStrategy(b *testing.B) {
	provider := newTestProvider()
	provider.AddVersions("Module::A", []string{"1.0.0", "1.1.0", "1.2.0", "2.0.0"})

	strategy := NewHybridStrategy(
		NewMinimalUpgradeStrategy(),
		NewMaximalUpgradeStrategy(),
	)

	conflicts := []*DependencyConflict{
		createTestConflict("Module::A", map[string][]string{
			">= 1.1": {"Module::B"},
			"< 2.0":  {"Module::C"},
		}),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, err := strategy.Resolve(context.Background(), conflicts, createTestOptions(provider))
		if err != nil || !result.Success {
			b.Fatal("Strategy failed")
		}
	}
}
