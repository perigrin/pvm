// ABOUTME: Tests for dependency resolver
// ABOUTME: Verifies the functionality of the dependency resolution system

package deps

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"tamarou.com/pvm/internal/cpan"
)

// MockProvider is a mock implementation of the cpan.Provider interface
type MockProvider struct {
	Modules map[string]*cpan.ModuleInfo
}

func (p *MockProvider) GetModuleInfo(ctx context.Context, moduleName string) (*cpan.ModuleInfo, error) {
	if info, ok := p.Modules[moduleName]; ok {
		return info, nil
	}
	return &cpan.ModuleInfo{
		Name:    moduleName,
		Version: "1.0.0",
		Dependencies: []*cpan.Dependency{
			{
				Name:    "Test::More",
				Version: "1.0.0",
				Phase:   "test",
				Type:    "requires",
			},
		},
	}, nil
}

func (p *MockProvider) SearchModules(ctx context.Context, query string, limit int) (*cpan.SearchResults, error) {
	return &cpan.SearchResults{}, nil
}

func (p *MockProvider) GetDependencies(ctx context.Context, moduleName string) ([]*cpan.Dependency, error) {
	info, err := p.GetModuleInfo(ctx, moduleName)
	if err != nil {
		return nil, err
	}
	return info.Dependencies, nil
}

func (p *MockProvider) GetModuleVersions(ctx context.Context, moduleName string) ([]string, error) {
	return []string{"1.0.0"}, nil
}

func (p *MockProvider) GetAuthorInfo(ctx context.Context, authorID string) (map[string]interface{}, error) {
	return map[string]interface{}{}, nil
}

func (p *MockProvider) IsCoreModule(ctx context.Context, moduleName string, perlVersion string) (bool, error) {
	return moduleName == "Test::More", nil
}

func (p *MockProvider) Name() string {
	return "mock"
}

func (p *MockProvider) BaseURL() string {
	return "https://mock.example.com"
}

// Create a mock provider with test data
func createMockProvider() *MockProvider {
	return &MockProvider{
		Modules: map[string]*cpan.ModuleInfo{
			"Moose": {
				Name:    "Moose",
				Version: "2.2014",
				Dependencies: []*cpan.Dependency{
					{
						Name:    "Class::MOP",
						Version: ">= 2.0",
						Phase:   "runtime",
						Type:    "requires",
					},
					{
						Name:    "Test::More",
						Version: "1.0.0",
						Phase:   "test",
						Type:    "requires",
					},
				},
			},
			"Class::MOP": {
				Name:    "Class::MOP",
				Version: "2.1",
				Dependencies: []*cpan.Dependency{
					{
						Name:    "Module::Runtime",
						Version: "0.014",
						Phase:   "runtime",
						Type:    "requires",
					},
				},
			},
			"Module::Runtime": {
				Name:         "Module::Runtime",
				Version:      "0.016",
				Dependencies: []*cpan.Dependency{
					// No dependencies
				},
			},
			"Test::More": {
				Name:    "Test::More",
				Version: "1.302183",
				Dependencies: []*cpan.Dependency{
					{
						Name:    "Test::Simple",
						Version: "1.302183",
						Phase:   "runtime",
						Type:    "requires",
					},
				},
			},
			"Test::Simple": {
				Name:         "Test::Simple",
				Version:      "1.302183",
				Dependencies: []*cpan.Dependency{
					// No dependencies
				},
			},
			"Circular::A": {
				Name:    "Circular::A",
				Version: "1.0",
				Dependencies: []*cpan.Dependency{
					{
						Name:    "Circular::B",
						Version: "1.0",
						Phase:   "runtime",
						Type:    "requires",
					},
				},
			},
			"Circular::B": {
				Name:    "Circular::B",
				Version: "1.0",
				Dependencies: []*cpan.Dependency{
					{
						Name:    "Circular::C",
						Version: "1.0",
						Phase:   "runtime",
						Type:    "requires",
					},
				},
			},
			"Circular::C": {
				Name:    "Circular::C",
				Version: "1.0",
				Dependencies: []*cpan.Dependency{
					{
						Name:    "Circular::A",
						Version: "1.0",
						Phase:   "runtime",
						Type:    "requires",
					},
				},
			},
			"Version::Conflict": {
				Name:    "Version::Conflict",
				Version: "2.0",
				Dependencies: []*cpan.Dependency{
					{
						Name:    "Dep::A",
						Version: ">= 2.0",
						Phase:   "runtime",
						Type:    "requires",
					},
					{
						Name:    "Dep::B",
						Version: "1.0",
						Phase:   "runtime",
						Type:    "requires",
					},
				},
			},
			"Dep::A": {
				Name:         "Dep::A",
				Version:      "2.5",
				Dependencies: []*cpan.Dependency{
					// No dependencies
				},
			},
			"Dep::B": {
				Name:    "Dep::B",
				Version: "1.0",
				Dependencies: []*cpan.Dependency{
					{
						Name:    "Dep::A",
						Version: "< 2.0",
						Phase:   "runtime",
						Type:    "requires",
					},
				},
			},
		},
	}
}

func TestResolveDependencies(t *testing.T) {
	// Create a resolver with a mock provider
	provider := createMockProvider()
	resolver, err := NewDefaultResolver("", 0) // No caching for this test
	require.NoError(t, err)

	// Test basic resolution
	ctx := context.Background()
	options := &DependencyResolutionOptions{
		Provider:     provider,
		IncludeCore:  false,
		IncludeTest:  false,
		IncludeBuild: true,
		IncludeDev:   false,
		MaxDepth:     0, // No limit
		Verbose:      true,
	}

	result, err := resolver.ResolveDependencies(ctx, "Moose", options)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Check that we have the expected modules
	assert.Equal(t, 3, len(result.Modules), "Should find 3 modules (Moose, Class::MOP, Module::Runtime)")
	assert.Contains(t, result.Modules, "Moose")
	assert.Contains(t, result.Modules, "Class::MOP")
	assert.Contains(t, result.Modules, "Module::Runtime")

	// Check that test dependencies are excluded
	assert.NotContains(t, result.Modules, "Test::More", "Test dependencies should be excluded")

	// Check with test dependencies included
	options.IncludeTest = true
	result, err = resolver.ResolveDependencies(ctx, "Moose", options)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, 5, len(result.Modules), "Should find 5 modules including test dependencies")
	assert.Contains(t, result.Modules, "Test::More")
	assert.Contains(t, result.Modules, "Test::Simple")
}

func TestCircularDependencies(t *testing.T) {
	// Create a resolver with a mock provider
	provider := createMockProvider()
	resolver, err := NewDefaultResolver("", 0) // No caching for this test
	require.NoError(t, err)

	// Test circular dependency detection
	ctx := context.Background()
	options := &DependencyResolutionOptions{
		Provider:     provider,
		IncludeCore:  false,
		IncludeTest:  false,
		IncludeBuild: true,
		IncludeDev:   false,
		MaxDepth:     0, // No limit
		Verbose:      true,
	}

	result, err := resolver.ResolveDependencies(ctx, "Circular::A", options)
	require.NoError(t, err) // Should not error, just warn
	require.NotNil(t, result)

	// Check that the implementation handles circular dependencies
	// We're not asserting specific warning messages, just that the resolver completed
	assert.Equal(t, 3, len(result.Modules), "Should find 3 modules in the circular dependency")
}

func TestVersionConflicts(t *testing.T) {
	// Create a resolver with a mock provider
	provider := createMockProvider()
	resolver, err := NewDefaultResolver("", 0) // No caching for this test
	require.NoError(t, err)

	// Test version conflict detection
	ctx := context.Background()
	options := &DependencyResolutionOptions{
		Provider:     provider,
		IncludeCore:  false,
		IncludeTest:  false,
		IncludeBuild: true,
		IncludeDev:   false,
		MaxDepth:     0, // No limit
		Verbose:      true,
	}

	result, err := resolver.ResolveDependencies(ctx, "Version::Conflict", options)
	require.NoError(t, err) // Should not error, just warn
	require.NotNil(t, result)

	// Check that the version conflict is detected
	assert.GreaterOrEqual(t, len(result.Conflicts), 1, "Should have at least one conflict")
	assert.GreaterOrEqual(t, len(result.Warnings), 1, "Should have at least one warning")
}

func TestMaxDepth(t *testing.T) {
	// Create a resolver with a mock provider
	provider := createMockProvider()
	resolver, err := NewDefaultResolver("", 0) // No caching for this test
	require.NoError(t, err)

	// Test max depth limiting
	ctx := context.Background()
	options := &DependencyResolutionOptions{
		Provider:     provider,
		IncludeCore:  false,
		IncludeTest:  false,
		IncludeBuild: true,
		IncludeDev:   false,
		MaxDepth:     1, // Limit to depth 1
		Verbose:      true,
	}

	result, err := resolver.ResolveDependencies(ctx, "Moose", options)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Check that we only find the root module and its direct dependencies
	assert.Equal(t, 2, len(result.Modules), "Should find only 2 modules due to max depth")
	assert.Contains(t, result.Modules, "Moose")
	assert.Contains(t, result.Modules, "Class::MOP")
	assert.NotContains(t, result.Modules, "Module::Runtime", "Should not find transitive dependencies due to max depth")
}

func TestCaching(t *testing.T) {
	// Create a temporary directory for testing
	cacheDir := t.TempDir()

	// Create a resolver with caching
	provider := createMockProvider()
	resolver, err := NewDefaultResolver(cacheDir, 1) // 1 hour TTL
	require.NoError(t, err)

	// Test caching
	ctx := context.Background()
	options := &DependencyResolutionOptions{
		Provider:     provider,
		IncludeCore:  false,
		IncludeTest:  false,
		IncludeBuild: true,
		IncludeDev:   false,
		UseCache:     true,
		CacheDir:     cacheDir,
		CacheTTL:     1,
		Verbose:      true,
	}

	// First resolution should hit the provider
	result1, err := resolver.ResolveDependencies(ctx, "Moose", options)
	require.NoError(t, err)
	require.NotNil(t, result1)

	// Second resolution should use the cache
	result2, err := resolver.ResolveDependencies(ctx, "Moose", options)
	require.NoError(t, err)
	require.NotNil(t, result2)

	// Results should be equivalent
	assert.Equal(t, len(result1.Modules), len(result2.Modules), "Cached result should have same number of modules")
	assert.Equal(t, result1.Root.Name, result2.Root.Name, "Cached result should have same root module")
}

func TestVersionConstraintParsing(t *testing.T) {
	// Test cases for constraint parsing
	testCases := []struct {
		constraint    string
		expectedCount int
		expectedOps   []ConstraintOperator
		expectedVers  []string
		expectedError bool
	}{
		// Single constraints
		{"== 1.0", 1, []ConstraintOperator{OpEqual}, []string{"1.0"}, false},
		{">= 2.0", 1, []ConstraintOperator{OpGreaterThanOrEqual}, []string{"2.0"}, false},
		{"> 1.0", 1, []ConstraintOperator{OpGreaterThan}, []string{"1.0"}, false},
		{"< 3.0", 1, []ConstraintOperator{OpLessThan}, []string{"3.0"}, false},
		{"<= 3.0", 1, []ConstraintOperator{OpLessThanOrEqual}, []string{"3.0"}, false},
		{"!= 1.5", 1, []ConstraintOperator{OpNotEqual}, []string{"1.5"}, false},
		{"1.0", 1, []ConstraintOperator{OpGreaterThanOrEqual}, []string{"1.0"}, false},
		{"v2.3.4", 1, []ConstraintOperator{OpGreaterThanOrEqual}, []string{"2.3.4"}, false},

		// Multiple constraints (cpanfile style)
		{">= 1.2, != 1.5, < 2.0", 3,
			[]ConstraintOperator{OpGreaterThanOrEqual, OpNotEqual, OpLessThan},
			[]string{"1.2", "1.5", "2.0"}, false},
		{">= 2.00, < 2.80", 2,
			[]ConstraintOperator{OpGreaterThanOrEqual, OpLessThan},
			[]string{"2.00", "2.80"}, false},

		// Empty constraint
		{"", 0, []ConstraintOperator{}, []string{}, false},

		// Invalid constraints
		{"invalid", 0, []ConstraintOperator{}, []string{}, true},
		{"<> 1.0", 0, []ConstraintOperator{}, []string{}, true}, // Invalid operator
	}

	for _, tc := range testCases {
		t.Run(tc.constraint, func(t *testing.T) {
			constraint, err := ParseVersionConstraint(tc.constraint)
			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedCount, len(constraint.Constraints))

				for i := 0; i < tc.expectedCount; i++ {
					assert.Equal(t, tc.expectedOps[i], constraint.Constraints[i].Operator)
					assert.Equal(t, tc.expectedVers[i], constraint.Constraints[i].Version)
				}
			}
		})
	}
}

func TestVersionConstraintChecking(t *testing.T) {
	// Test cases for constraint checking
	testCases := []struct {
		version       string
		constraint    string
		expectedMatch bool
		expectedError bool
	}{
		{"1.0", "== 1.0", true, false},
		{"1.0", "!= 1.0", false, false},
		{"2.0", "> 1.0", true, false},
		{"1.0", "> 1.0", false, false},
		{"2.0", ">= 2.0", true, false},
		{"1.9", ">= 2.0", false, false},
		{"1.5", "< 2.0", true, false},
		{"2.0", "< 2.0", false, false},
		{"1.5", "<= 1.5", true, false},
		{"1.6", "<= 1.5", false, false},
		{"1.5", "1.5", true, false},      // 1.5 satisfies >= 1.5
		{"1.5", "1.6", false, false},     // 1.5 does not satisfy >= 1.6
		{"1.2.3", "v1.2.3", true, false}, // 1.2.3 satisfies >= 1.2.3
		{"1.6", "1.5", true, false},      // 1.6 satisfies >= 1.5
		{"2.0", "1.0", true, false},      // 2.0 satisfies >= 1.0

		// Multiple constraints (cpanfile style)
		{"2.0.0", ">= 1, < 3", true, false},            // 2.0.0 is between 1 and 3
		{"0.5.0", ">= 1, < 3", false, false},           // 0.5.0 is less than 1
		{"3.0.0", ">= 1, < 3", false, false},           // 3.0.0 is not less than 3
		{"1.5", ">= 1.2, != 1.5, < 2.0", false, false}, // 1.5 is excluded
		{"1.6", ">= 1.2, != 1.5, < 2.0", true, false},  // 1.6 satisfies all
		{"2.5", ">= 2.00, < 2.80", true, false},        // cpanfile example

		{"1.0", "", true, false}, // Empty constraint matches anything
	}

	for _, tc := range testCases {
		t.Run(tc.version+"-"+tc.constraint, func(t *testing.T) {
			match, err := CheckVersionConstraint(tc.version, tc.constraint)
			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedMatch, match)
			}
		})
	}
}

func TestPrintDependencyTree(t *testing.T) {
	// Create a sample dependency tree
	root := &DependencyNode{
		Name:     "Root",
		Version:  "1.0",
		IsRoot:   true,
		Children: []*DependencyNode{},
	}

	child1 := &DependencyNode{
		Name:              "Child1",
		Version:           "2.0",
		VersionConstraint: ">= 1.5",
		Parent:            root,
		Children:          []*DependencyNode{},
	}

	child2 := &DependencyNode{
		Name:              "Child2",
		Version:           "3.0",
		VersionConstraint: "3.0",
		Parent:            root,
		Children:          []*DependencyNode{},
	}

	grandchild := &DependencyNode{
		Name:              "Grandchild",
		Version:           "4.0",
		VersionConstraint: "4.0",
		Parent:            child1,
		Children:          []*DependencyNode{},
	}

	root.Children = append(root.Children, child1, child2)
	child1.Children = append(child1.Children, grandchild)

	// Create resolver
	resolver, err := NewDefaultResolver("", 0)
	require.NoError(t, err)

	// Print the tree
	tree := resolver.PrintDependencyTree(root)

	// Check that the tree contains all module names
	assert.Contains(t, tree, "Root")
	assert.Contains(t, tree, "Child1")
	assert.Contains(t, tree, "Child2")
	assert.Contains(t, tree, "Grandchild")

	// Check that version constraints are included
	assert.Contains(t, tree, ">= 1.5")

	// Check tree structure indicators
	assert.Contains(t, tree, "├──")
	assert.Contains(t, tree, "└──")
}
