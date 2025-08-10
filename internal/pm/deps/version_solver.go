// ABOUTME: Advanced version constraint solver using CSP algorithms
// ABOUTME: Implements sophisticated version selection and constraint satisfaction

package deps

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"tamarou.com/pvm/internal/cpan"
	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/log"
)

// VersionSolver provides advanced version constraint solving capabilities
type VersionSolver struct {
	provider        cpan.Provider
	maxIterations   int
	timeoutDuration time.Duration
	verbose         bool

	// Constraint satisfaction problem state
	variables   map[string]*VersionVariable
	constraints []VersionConstraintRule

	// Resolution state
	assignments map[string]string
	domain      map[string][]string

	// Performance tracking
	metrics SolverMetrics
}

// VersionVariable represents a module whose version needs to be determined
type VersionVariable struct {
	Module       string
	CurrentValue string
	Domain       []string
	Constraints  []*VersionConstraintRule
	IsPinned     bool
	PinnedValue  string
}

// VersionConstraintRule represents a constraint between modules
type VersionConstraintRule struct {
	Source     string // Module that imposes the constraint
	Target     string // Module that must satisfy the constraint
	Constraint string // Version constraint string
	Parsed     *VersionConstraint
	Priority   int    // Higher priority constraints are enforced first
	Source_    string // Fallback for constraint identification
}

// SolverMetrics tracks solver performance
type SolverMetrics struct {
	StartTime          time.Time
	EndTime            time.Time
	IterationsUsed     int
	BacktrackCount     int
	ConstraintChecks   int
	DomainReductions   int
	VariablesProcessed int
}

// SolverResult contains the outcome of version solving
type SolverResult struct {
	Success                bool
	Assignments            map[string]string
	Explanation            string
	Metrics                SolverMetrics
	ConflictingConstraints []VersionConstraintRule
}

// NewVersionSolver creates a new version solver
func NewVersionSolver(provider cpan.Provider) *VersionSolver {
	return &VersionSolver{
		provider:        provider,
		maxIterations:   1000,
		timeoutDuration: 30 * time.Second,
		verbose:         false,
		variables:       make(map[string]*VersionVariable),
		constraints:     make([]VersionConstraintRule, 0),
		assignments:     make(map[string]string),
		domain:          make(map[string][]string),
	}
}

// SetMaxIterations sets the maximum number of solver iterations
func (vs *VersionSolver) SetMaxIterations(max int) {
	vs.maxIterations = max
}

// SetTimeout sets the solver timeout duration
func (vs *VersionSolver) SetTimeout(timeout time.Duration) {
	vs.timeoutDuration = timeout
}

// SetVerbose enables or disables verbose logging
func (vs *VersionSolver) SetVerbose(verbose bool) {
	vs.verbose = verbose
}

// AddVariable adds a variable (module) to be solved
func (vs *VersionSolver) AddVariable(module string, domain []string, pinned bool, pinnedValue string) error {
	if len(domain) == 0 {
		return errors.NewSystemError(
			ErrResolveFailure,
			fmt.Sprintf("Empty domain for module %s", module),
			nil)
	}

	variable := &VersionVariable{
		Module:      module,
		Domain:      make([]string, len(domain)),
		Constraints: make([]*VersionConstraintRule, 0),
		IsPinned:    pinned,
		PinnedValue: pinnedValue,
	}

	copy(variable.Domain, domain)

	// If pinned, domain should only contain the pinned value
	if pinned && pinnedValue != "" {
		variable.Domain = []string{pinnedValue}
	}

	vs.variables[module] = variable
	vs.domain[module] = variable.Domain

	if vs.verbose {
		log.Infof("Added variable %s with domain size %d (pinned: %v)", module, len(domain), pinned)
	}

	return nil
}

// AddConstraint adds a version constraint between modules
func (vs *VersionSolver) AddConstraint(source, target, constraint string, priority int) error {
	parsed, err := ParseVersionConstraint(constraint)
	if err != nil {
		return errors.NewSystemError(
			ErrInvalidVersionPattern,
			fmt.Sprintf("Failed to parse constraint '%s' from %s to %s: %v", constraint, source, target, err),
			err)
	}

	rule := VersionConstraintRule{
		Source:     source,
		Target:     target,
		Constraint: constraint,
		Parsed:     parsed,
		Priority:   priority,
	}

	vs.constraints = append(vs.constraints, rule)

	// Add constraint reference to target variable
	if targetVar, exists := vs.variables[target]; exists {
		targetVar.Constraints = append(targetVar.Constraints, &rule)
	}

	if vs.verbose {
		log.Infof("Added constraint: %s requires %s %s (priority %d)", source, target, constraint, priority)
	}

	return nil
}

// Solve attempts to find a valid assignment for all variables
func (vs *VersionSolver) Solve(ctx context.Context) (*SolverResult, error) {
	vs.metrics = SolverMetrics{
		StartTime: time.Now(),
	}

	if vs.verbose {
		log.Infof("Starting version solving for %d variables and %d constraints", len(vs.variables), len(vs.constraints))
	}

	// Sort constraints by priority (higher priority first)
	sort.Slice(vs.constraints, func(i, j int) bool {
		return vs.constraints[i].Priority > vs.constraints[j].Priority
	})

	// Initialize domains and assignments
	vs.assignments = make(map[string]string)
	vs.domain = make(map[string][]string)

	for module, variable := range vs.variables {
		vs.domain[module] = make([]string, len(variable.Domain))
		copy(vs.domain[module], variable.Domain)

		// If variable is pinned, assign it immediately
		if variable.IsPinned && variable.PinnedValue != "" {
			vs.assignments[module] = variable.PinnedValue
		}
	}

	// Apply constraint propagation
	if err := vs.propagateConstraints(ctx); err != nil {
		return &SolverResult{
			Success:     false,
			Assignments: vs.assignments,
			Explanation: fmt.Sprintf("Constraint propagation failed: %v", err),
			Metrics:     vs.metrics,
		}, err
	}

	// Use backtracking search to find a solution
	success := vs.backtrackSearch(ctx)

	vs.metrics.EndTime = time.Now()

	result := &SolverResult{
		Success:     success,
		Assignments: make(map[string]string),
		Metrics:     vs.metrics,
	}

	// Copy assignments to result
	for module, version := range vs.assignments {
		result.Assignments[module] = version
	}

	if success {
		result.Explanation = fmt.Sprintf("Solution found in %d iterations", vs.metrics.IterationsUsed)
	} else {
		result.Explanation = "No solution found within iteration limit"
		result.ConflictingConstraints = vs.findConflictingConstraints()
	}

	if vs.verbose {
		log.Infof("Solving completed in %v: success=%v, iterations=%d",
			vs.metrics.EndTime.Sub(vs.metrics.StartTime),
			success,
			vs.metrics.IterationsUsed)
	}

	return result, nil
}

// propagateConstraints applies constraint propagation to reduce domains
func (vs *VersionSolver) propagateConstraints(ctx context.Context) error {
	changed := true
	iterations := 0

	for changed && iterations < vs.maxIterations {
		changed = false
		iterations++

		// Check for timeout
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		for i := range vs.constraints {
			constraint := &vs.constraints[i]

			if vs.reduceConstraintDomain(constraint) {
				changed = true
				vs.metrics.DomainReductions++
			}
		}

		// Check for empty domains
		for module, domain := range vs.domain {
			if len(domain) == 0 && vs.assignments[module] == "" {
				return errors.NewSystemError(
					ErrVersionConflict,
					fmt.Sprintf("No valid versions remain for module %s after constraint propagation", module),
					nil)
			}
		}
	}

	return nil
}

// reduceConstraintDomain applies arc consistency for a single constraint
func (vs *VersionSolver) reduceConstraintDomain(constraint *VersionConstraintRule) bool {
	targetDomain := vs.domain[constraint.Target]
	if len(targetDomain) <= 1 {
		return false // Domain already minimal
	}

	// If source is assigned, check if target domain needs reduction
	if sourceVersion, assigned := vs.assignments[constraint.Source]; assigned {
		return vs.reduceTargetDomain(constraint, sourceVersion)
	}

	// If source is not assigned, check all possible source values
	sourceDomain := vs.domain[constraint.Source]
	validTargetVersions := make(map[string]bool)

	for _, sourceVersion := range sourceDomain {
		for _, targetVersion := range targetDomain {
			vs.metrics.ConstraintChecks++

			if vs.constraintSatisfied(constraint, sourceVersion, targetVersion) {
				validTargetVersions[targetVersion] = true
			}
		}
	}

	// Update target domain to only include valid versions
	newTargetDomain := make([]string, 0, len(targetDomain))
	for _, version := range targetDomain {
		if validTargetVersions[version] {
			newTargetDomain = append(newTargetDomain, version)
		}
	}

	if len(newTargetDomain) < len(targetDomain) {
		vs.domain[constraint.Target] = newTargetDomain
		return true
	}

	return false
}

// reduceTargetDomain reduces target domain based on assigned source version
func (vs *VersionSolver) reduceTargetDomain(constraint *VersionConstraintRule, sourceVersion string) bool {
	targetDomain := vs.domain[constraint.Target]
	validVersions := make([]string, 0, len(targetDomain))

	for _, targetVersion := range targetDomain {
		vs.metrics.ConstraintChecks++

		if vs.constraintSatisfied(constraint, sourceVersion, targetVersion) {
			validVersions = append(validVersions, targetVersion)
		}
	}

	if len(validVersions) < len(targetDomain) {
		vs.domain[constraint.Target] = validVersions
		return true
	}

	return false
}

// constraintSatisfied checks if a constraint is satisfied by given versions
func (vs *VersionSolver) constraintSatisfied(constraint *VersionConstraintRule, sourceVersion, targetVersion string) bool {
	// The constraint is imposed by the source on the target
	// So we check if the target version satisfies the constraint
	satisfied, err := CheckVersionConstraint(targetVersion, constraint.Constraint)
	if err != nil {
		if vs.verbose {
			log.Warnf("Error checking constraint %s -> %s %s: %v", constraint.Source, constraint.Target, constraint.Constraint, err)
		}
		return false
	}
	return satisfied
}

// backtrackSearch performs backtracking search to find a valid assignment
func (vs *VersionSolver) backtrackSearch(ctx context.Context) bool {
	vs.metrics.IterationsUsed++

	// Check for timeout and iteration limit
	if vs.metrics.IterationsUsed >= vs.maxIterations {
		return false
	}

	select {
	case <-ctx.Done():
		return false
	default:
	}

	// Find next unassigned variable using MRV (Minimum Remaining Values) heuristic
	variable := vs.selectUnassignedVariable()
	if variable == "" {
		// All variables assigned - check if solution is valid
		return vs.validateSolution()
	}

	vs.metrics.VariablesProcessed++

	// Try values for this variable using LCV (Least Constraining Value) heuristic
	values := vs.orderDomainValues(variable)

	for _, value := range values {
		// Try assignment
		vs.assignments[variable] = value

		// Check consistency with current assignments
		if vs.isConsistent(variable, value) {
			// Forward checking: update domains based on this assignment
			savedDomains := vs.saveDomains()

			vs.updateDomainsAfterAssignment(variable, value)

			// Recursively try to assign remaining variables
			if vs.backtrackSearch(ctx) {
				return true
			}

			// Backtrack: restore domains
			vs.restoreDomains(savedDomains)
		}

		// Remove assignment
		delete(vs.assignments, variable)
		vs.metrics.BacktrackCount++
	}

	return false
}

// selectUnassignedVariable selects the next variable to assign using MRV heuristic
func (vs *VersionSolver) selectUnassignedVariable() string {
	var bestVariable string
	minDomainSize := int(^uint(0) >> 1) // Max int

	for module, domain := range vs.domain {
		// Skip if already assigned
		if _, assigned := vs.assignments[module]; assigned {
			continue
		}

		// Skip if pinned (should already be assigned)
		if variable := vs.variables[module]; variable.IsPinned {
			continue
		}

		// Choose variable with smallest domain
		if len(domain) < minDomainSize {
			bestVariable = module
			minDomainSize = len(domain)
		}
	}

	return bestVariable
}

// orderDomainValues orders domain values using LCV heuristic
func (vs *VersionSolver) orderDomainValues(variable string) []string {
	domain := vs.domain[variable]
	if len(domain) <= 1 {
		return domain
	}

	// For simplicity, we'll order by version (latest first)
	// A more sophisticated implementation would count constraint violations
	ordered := make([]string, len(domain))
	copy(ordered, domain)

	sort.Sort(sort.Reverse(VersionSlice(ordered)))
	return ordered
}

// isConsistent checks if an assignment is consistent with all constraints
func (vs *VersionSolver) isConsistent(variable, value string) bool {
	// Check all constraints involving this variable
	for i := range vs.constraints {
		constraint := &vs.constraints[i]

		// Check constraints where this variable is the target
		if constraint.Target == variable {
			// If source is assigned, check constraint
			if sourceValue, assigned := vs.assignments[constraint.Source]; assigned {
				if !vs.constraintSatisfied(constraint, sourceValue, value) {
					return false
				}
			}
		}

		// Check constraints where this variable is the source
		if constraint.Source == variable {
			// If target is assigned, check constraint
			if targetValue, assigned := vs.assignments[constraint.Target]; assigned {
				if !vs.constraintSatisfied(constraint, value, targetValue) {
					return false
				}
			}
		}
	}

	return true
}

// validateSolution checks if the current complete assignment is valid
func (vs *VersionSolver) validateSolution() bool {
	// Check all constraints
	for i := range vs.constraints {
		constraint := &vs.constraints[i]

		sourceValue, sourceAssigned := vs.assignments[constraint.Source]
		targetValue, targetAssigned := vs.assignments[constraint.Target]

		// Both must be assigned for complete solution
		if !sourceAssigned || !targetAssigned {
			return false
		}

		if !vs.constraintSatisfied(constraint, sourceValue, targetValue) {
			return false
		}
	}

	return true
}

// saveDomains creates a backup of current domains
func (vs *VersionSolver) saveDomains() map[string][]string {
	saved := make(map[string][]string)
	for module, domain := range vs.domain {
		saved[module] = make([]string, len(domain))
		copy(saved[module], domain)
	}
	return saved
}

// restoreDomains restores domains from backup
func (vs *VersionSolver) restoreDomains(saved map[string][]string) {
	vs.domain = saved
}

// updateDomainsAfterAssignment updates domains based on a new assignment
func (vs *VersionSolver) updateDomainsAfterAssignment(variable, value string) {
	// For each constraint involving this variable, update other domains
	for i := range vs.constraints {
		constraint := &vs.constraints[i]

		if constraint.Source == variable {
			// This variable is the source, update target domain
			vs.reduceTargetDomain(constraint, value)
		}
	}
}

// findConflictingConstraints identifies constraints that cannot be satisfied
func (vs *VersionSolver) findConflictingConstraints() []VersionConstraintRule {
	conflicting := make([]VersionConstraintRule, 0)

	for i := range vs.constraints {
		constraint := &vs.constraints[i]

		// Check if there's any valid assignment for this constraint
		sourceValues := vs.domain[constraint.Source]
		targetValues := vs.domain[constraint.Target]

		hasValidPair := false
		for _, sourceValue := range sourceValues {
			for _, targetValue := range targetValues {
				if vs.constraintSatisfied(constraint, sourceValue, targetValue) {
					hasValidPair = true
					break
				}
			}
			if hasValidPair {
				break
			}
		}

		if !hasValidPair {
			conflicting = append(conflicting, *constraint)
		}
	}

	return conflicting
}

// PrintSolution formats the solution for display
func (vs *VersionSolver) PrintSolution(result *SolverResult) string {
	var builder strings.Builder

	fmt.Fprintf(&builder, "Version Solution (Success: %v)\n", result.Success)
	fmt.Fprintf(&builder, "================================\n")

	if result.Success {
		// Sort modules for consistent output
		modules := make([]string, 0, len(result.Assignments))
		for module := range result.Assignments {
			modules = append(modules, module)
		}
		sort.Strings(modules)

		for _, module := range modules {
			version := result.Assignments[module]
			pinned := ""
			if variable := vs.variables[module]; variable != nil && variable.IsPinned {
				pinned = " (pinned)"
			}
			fmt.Fprintf(&builder, "  %s: %s%s\n", module, version, pinned)
		}
	} else {
		fmt.Fprintf(&builder, "No solution found.\n")
		if len(result.ConflictingConstraints) > 0 {
			fmt.Fprintf(&builder, "\nConflicting constraints:\n")
			for _, constraint := range result.ConflictingConstraints {
				fmt.Fprintf(&builder, "  %s requires %s %s\n", constraint.Source, constraint.Target, constraint.Constraint)
			}
		}
	}

	fmt.Fprintf(&builder, "\nMetrics:\n")
	fmt.Fprintf(&builder, "  Duration: %v\n", result.Metrics.EndTime.Sub(result.Metrics.StartTime))
	fmt.Fprintf(&builder, "  Iterations: %d\n", result.Metrics.IterationsUsed)
	fmt.Fprintf(&builder, "  Backtracks: %d\n", result.Metrics.BacktrackCount)
	fmt.Fprintf(&builder, "  Constraint checks: %d\n", result.Metrics.ConstraintChecks)
	fmt.Fprintf(&builder, "  Domain reductions: %d\n", result.Metrics.DomainReductions)

	return builder.String()
}
