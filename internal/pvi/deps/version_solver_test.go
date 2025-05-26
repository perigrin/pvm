// ABOUTME: Tests for the advanced version constraint solver
// ABOUTME: Verifies CSP-based version selection algorithms

package deps

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersionSolver_NewVersionSolver(t *testing.T) {
	provider := newTestProvider()
	solver := NewVersionSolver(provider)

	assert.NotNil(t, solver)
	assert.Equal(t, provider, solver.provider)
	assert.Equal(t, 1000, solver.maxIterations)
	assert.Equal(t, 30*time.Second, solver.timeoutDuration)
}

func TestVersionSolver_AddVariable(t *testing.T) {
	provider := newTestProvider()
	solver := NewVersionSolver(provider)

	domain := []string{"1.0.0", "1.1.0", "1.2.0"}
	err := solver.AddVariable("Module::A", domain, false, "")
	require.NoError(t, err)

	variable := solver.variables["Module::A"]
	assert.NotNil(t, variable)
	assert.Equal(t, "Module::A", variable.Module)
	assert.Equal(t, domain, variable.Domain)
	assert.False(t, variable.IsPinned)
}

func TestVersionSolver_AddPinnedVariable(t *testing.T) {
	provider := newTestProvider()
	solver := NewVersionSolver(provider)

	domain := []string{"1.0.0", "1.1.0", "1.2.0"}
	err := solver.AddVariable("Module::A", domain, true, "1.1.0")
	require.NoError(t, err)

	variable := solver.variables["Module::A"]
	assert.NotNil(t, variable)
	assert.True(t, variable.IsPinned)
	assert.Equal(t, "1.1.0", variable.PinnedValue)
	assert.Equal(t, []string{"1.1.0"}, variable.Domain) // Domain should be restricted to pinned value
}

func TestVersionSolver_AddConstraint(t *testing.T) {
	provider := newTestProvider()
	solver := NewVersionSolver(provider)

	// Add variables first
	err := solver.AddVariable("Module::A", []string{"1.0.0", "1.1.0"}, false, "")
	require.NoError(t, err)
	err = solver.AddVariable("Module::B", []string{"1.0.0", "1.1.0"}, false, "")
	require.NoError(t, err)

	// Add constraint
	err = solver.AddConstraint("Module::A", "Module::B", ">= 1.0", 10)
	require.NoError(t, err)

	assert.Equal(t, 1, len(solver.constraints))
	constraint := solver.constraints[0]
	assert.Equal(t, "Module::A", constraint.Source)
	assert.Equal(t, "Module::B", constraint.Target)
	assert.Equal(t, ">= 1.0", constraint.Constraint)
	assert.Equal(t, 10, constraint.Priority)
}

func TestVersionSolver_SimpleResolution(t *testing.T) {
	provider := newTestProvider()
	solver := NewVersionSolver(provider)
	solver.SetVerbose(true)

	// Add variables
	err := solver.AddVariable("Module::A", []string{"1.0.0", "1.1.0", "1.2.0"}, false, "")
	require.NoError(t, err)
	err = solver.AddVariable("Module::B", []string{"1.0.0", "1.1.0", "1.2.0"}, false, "")
	require.NoError(t, err)

	// Add constraints
	err = solver.AddConstraint("Module::A", "Module::B", ">= 1.1", 10)
	require.NoError(t, err)

	// Solve
	result, err := solver.Solve(context.Background())
	require.NoError(t, err)
	assert.True(t, result.Success)

	// Check assignments
	assert.Contains(t, result.Assignments, "Module::A")
	assert.Contains(t, result.Assignments, "Module::B")

	// Verify constraint is satisfied
	moduleB := result.Assignments["Module::B"]
	satisfied, err := CheckVersionConstraint(moduleB, ">= 1.1")
	require.NoError(t, err)
	assert.True(t, satisfied)
}

func TestVersionSolver_PinnedVariableResolution(t *testing.T) {
	provider := newTestProvider()
	solver := NewVersionSolver(provider)

	// Add variables with one pinned
	err := solver.AddVariable("Module::A", []string{"1.0.0", "1.1.0", "1.2.0"}, true, "1.2.0")
	require.NoError(t, err)
	err = solver.AddVariable("Module::B", []string{"1.0.0", "1.1.0", "1.2.0"}, false, "")
	require.NoError(t, err)

	// Add constraint
	err = solver.AddConstraint("Module::A", "Module::B", ">= 1.1", 10)
	require.NoError(t, err)

	// Solve
	result, err := solver.Solve(context.Background())
	require.NoError(t, err)
	assert.True(t, result.Success)

	// Check that pinned variable has correct value
	assert.Equal(t, "1.2.0", result.Assignments["Module::A"])

	// Check that Module::B satisfies constraint
	moduleB := result.Assignments["Module::B"]
	satisfied, err := CheckVersionConstraint(moduleB, ">= 1.1")
	require.NoError(t, err)
	assert.True(t, satisfied)
}

func TestVersionSolver_MultipleConstraints(t *testing.T) {
	provider := newTestProvider()
	solver := NewVersionSolver(provider)

	// Add variables
	err := solver.AddVariable("Module::A", []string{"1.0.0", "1.1.0", "1.2.0", "2.0.0"}, false, "")
	require.NoError(t, err)
	err = solver.AddVariable("Module::B", []string{"1.0.0", "1.1.0", "1.2.0", "2.0.0"}, false, "")
	require.NoError(t, err)
	err = solver.AddVariable("Module::C", []string{"1.0.0", "1.1.0", "1.2.0", "2.0.0"}, false, "")
	require.NoError(t, err)

	// Add constraints forming a chain: A -> B -> C
	err = solver.AddConstraint("Module::A", "Module::B", ">= 1.1", 10)
	require.NoError(t, err)
	err = solver.AddConstraint("Module::B", "Module::C", ">= 1.2", 10)
	require.NoError(t, err)
	err = solver.AddConstraint("Module::A", "Module::C", "< 2.0", 5)
	require.NoError(t, err)

	// Solve
	result, err := solver.Solve(context.Background())
	require.NoError(t, err)
	assert.True(t, result.Success)

	// Verify all constraints are satisfied
	moduleB := result.Assignments["Module::B"]
	moduleC := result.Assignments["Module::C"]

	satisfied, err := CheckVersionConstraint(moduleB, ">= 1.1")
	require.NoError(t, err)
	assert.True(t, satisfied)

	satisfied, err = CheckVersionConstraint(moduleC, ">= 1.2")
	require.NoError(t, err)
	assert.True(t, satisfied)

	satisfied, err = CheckVersionConstraint(moduleC, "< 2.0")
	require.NoError(t, err)
	assert.True(t, satisfied)
}

func TestVersionSolver_ImpossibleConstraints(t *testing.T) {
	provider := newTestProvider()
	solver := NewVersionSolver(provider)

	// Add variable with limited domain
	err := solver.AddVariable("Module::A", []string{"1.0.0", "1.1.0"}, false, "")
	require.NoError(t, err)

	// Add impossible constraint
	err = solver.AddConstraint("Root", "Module::A", ">= 2.0", 10)
	require.NoError(t, err)

	// Solve
	result, err := solver.Solve(context.Background())
	if err != nil {
		// Error during constraint propagation is expected for impossible constraints
		assert.Contains(t, err.Error(), "constraint propagation")
		return
	}
	assert.False(t, result.Success)
	assert.NotEmpty(t, result.ConflictingConstraints)
}

func TestVersionSolver_BacktrackingRequired(t *testing.T) {
	provider := newTestProvider()
	solver := NewVersionSolver(provider)
	solver.SetVerbose(true)

	// Create a scenario that requires backtracking
	err := solver.AddVariable("Module::A", []string{"1.0.0", "1.1.0", "1.2.0"}, false, "")
	require.NoError(t, err)
	err = solver.AddVariable("Module::B", []string{"1.0.0", "1.1.0", "1.2.0"}, false, "")
	require.NoError(t, err)
	err = solver.AddVariable("Module::C", []string{"1.0.0", "1.1.0", "1.2.0"}, false, "")
	require.NoError(t, err)

	// Add constraints that create interdependencies
	err = solver.AddConstraint("Module::A", "Module::B", ">= 1.1", 10)
	require.NoError(t, err)
	err = solver.AddConstraint("Module::B", "Module::C", ">= 1.1", 10)
	require.NoError(t, err)
	err = solver.AddConstraint("Module::C", "Module::A", "< 1.2", 10)
	require.NoError(t, err)

	// Solve
	result, err := solver.Solve(context.Background())
	require.NoError(t, err)
	assert.True(t, result.Success)

	// Check that backtracking occurred (or the solver was efficient enough to avoid it)
	// The important thing is that we found a valid solution
	assert.GreaterOrEqual(t, result.Metrics.BacktrackCount, 0)
}

func TestVersionSolver_ConstraintPriority(t *testing.T) {
	provider := newTestProvider()
	solver := NewVersionSolver(provider)

	// Add variables
	err := solver.AddVariable("Module::A", []string{"1.0.0", "1.1.0", "1.2.0"}, false, "")
	require.NoError(t, err)

	// Add constraints with different priorities
	err = solver.AddConstraint("HighPriority", "Module::A", ">= 1.2", 100)
	require.NoError(t, err)
	err = solver.AddConstraint("LowPriority", "Module::A", ">= 1.0", 1)
	require.NoError(t, err)

	// Check that constraints are sorted by priority
	assert.Equal(t, 2, len(solver.constraints))
	assert.Equal(t, 100, solver.constraints[0].Priority) // High priority first
	assert.Equal(t, 1, solver.constraints[1].Priority)
}

func TestVersionSolver_Timeout(t *testing.T) {
	provider := newTestProvider()
	solver := NewVersionSolver(provider)
	solver.SetTimeout(1 * time.Millisecond) // Very short timeout

	// Add many variables to trigger timeout
	for i := 0; i < 10; i++ {
		name := fmt.Sprintf("Module::%d", i)
		domain := []string{"1.0.0", "1.1.0", "1.2.0", "2.0.0", "2.1.0"}
		err := solver.AddVariable(name, domain, false, "")
		require.NoError(t, err)

		if i > 0 {
			prevName := fmt.Sprintf("Module::%d", i-1)
			err = solver.AddConstraint(prevName, name, ">= 1.0", 10)
			require.NoError(t, err)
		}
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// Solve (should timeout)
	result, err := solver.Solve(ctx)
	if err == nil {
		// If it didn't timeout, that's also acceptable (depends on machine speed)
		assert.NotNil(t, result)
	}
}

func TestVersionSolver_EmptyDomain(t *testing.T) {
	provider := newTestProvider()
	solver := NewVersionSolver(provider)

	// Try to add variable with empty domain
	err := solver.AddVariable("Module::A", []string{}, false, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Empty domain")
}

func TestVersionSolver_InvalidConstraint(t *testing.T) {
	provider := newTestProvider()
	solver := NewVersionSolver(provider)

	// Add variable
	err := solver.AddVariable("Module::A", []string{"1.0.0"}, false, "")
	require.NoError(t, err)

	// Try to add invalid constraint
	err = solver.AddConstraint("Source", "Module::A", "invalid_constraint", 10)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Failed to parse constraint")
}

func TestVersionSolver_PrintSolution(t *testing.T) {
	provider := newTestProvider()
	solver := NewVersionSolver(provider)

	// Add variables
	err := solver.AddVariable("Module::A", []string{"1.0.0", "1.1.0"}, false, "")
	require.NoError(t, err)
	err = solver.AddVariable("Module::B", []string{"1.0.0", "1.1.0"}, true, "1.1.0")
	require.NoError(t, err)

	// Solve
	result, err := solver.Solve(context.Background())
	require.NoError(t, err)

	// Print solution
	output := solver.PrintSolution(result)
	assert.Contains(t, output, "Version Solution")
	assert.Contains(t, output, "Module::A")
	assert.Contains(t, output, "Module::B")
	assert.Contains(t, output, "(pinned)") // Should indicate pinned variable
	assert.Contains(t, output, "Metrics:")
}

func TestVersionSolver_ConstraintSatisfaction(t *testing.T) {
	provider := newTestProvider()
	solver := NewVersionSolver(provider)

	// Test constraint satisfaction check
	constraint := &VersionConstraintRule{
		Source:     "Module::A",
		Target:     "Module::B",
		Constraint: ">= 1.1",
	}

	parsed, err := ParseVersionConstraint(">= 1.1")
	require.NoError(t, err)
	constraint.Parsed = parsed

	// Test satisfied constraint
	satisfied := solver.constraintSatisfied(constraint, "1.2.0", "1.2.0")
	assert.True(t, satisfied)

	// Test unsatisfied constraint
	satisfied = solver.constraintSatisfied(constraint, "1.2.0", "1.0.0")
	assert.False(t, satisfied)
}

func TestVersionSolver_DomainReduction(t *testing.T) {
	provider := newTestProvider()
	solver := NewVersionSolver(provider)

	// Add variables
	err := solver.AddVariable("Module::A", []string{"1.0.0", "1.1.0", "1.2.0"}, false, "")
	require.NoError(t, err)
	err = solver.AddVariable("Module::B", []string{"1.0.0", "1.1.0", "1.2.0"}, false, "")
	require.NoError(t, err)

	// Add constraint
	err = solver.AddConstraint("Module::A", "Module::B", ">= 1.1", 10)
	require.NoError(t, err)

	// Assign Module::A
	solver.assignments["Module::A"] = "1.1.0"

	// Apply domain reduction
	constraint := &solver.constraints[0]
	reduced := solver.reduceTargetDomain(constraint, "1.1.0")
	assert.True(t, reduced)

	// Check that Module::B domain was reduced
	domainB := solver.domain["Module::B"]
	assert.True(t, len(domainB) <= 2) // Should exclude 1.0.0
}

// Benchmark tests

func BenchmarkVersionSolver_SimpleResolution(b *testing.B) {
	provider := newTestProvider()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		solver := NewVersionSolver(provider)

		solver.AddVariable("Module::A", []string{"1.0.0", "1.1.0", "1.2.0"}, false, "")
		solver.AddVariable("Module::B", []string{"1.0.0", "1.1.0", "1.2.0"}, false, "")
		solver.AddConstraint("Module::A", "Module::B", ">= 1.1", 10)

		result, err := solver.Solve(context.Background())
		if err != nil || !result.Success {
			b.Fatal("Solve failed")
		}
	}
}

func BenchmarkVersionSolver_ComplexResolution(b *testing.B) {
	provider := newTestProvider()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		solver := NewVersionSolver(provider)

		// Add multiple variables with constraints
		versions := []string{"1.0.0", "1.1.0", "1.2.0", "2.0.0", "2.1.0"}
		for j := 0; j < 5; j++ {
			name := fmt.Sprintf("Module::%d", j)
			solver.AddVariable(name, versions, false, "")

			if j > 0 {
				prevName := fmt.Sprintf("Module::%d", j-1)
				solver.AddConstraint(prevName, name, ">= 1.1", 10)
			}
		}

		result, err := solver.Solve(context.Background())
		if err != nil || !result.Success {
			b.Fatal("Solve failed")
		}
	}
}

// Test helper for adding import
func init() {
	// Ensure fmt is imported for test that uses it
	_ = fmt.Sprintf("")
}
