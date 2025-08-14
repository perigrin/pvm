// ABOUTME: Implements constraint propagation for generic type inference
// ABOUTME: Handles advanced constraint resolution and type parameter bound checking

package typechecker

import (
	"fmt"
	"sort"

	"tamarou.com/pvm/internal/typedef"
)

// ConstraintPropagator handles constraint propagation during type inference
type ConstraintPropagator struct {
	hierarchy *typedef.TypeHierarchy
}

// NewConstraintPropagator creates a new constraint propagator
func NewConstraintPropagator(hierarchy *typedef.TypeHierarchy) *ConstraintPropagator {
	return &ConstraintPropagator{
		hierarchy: hierarchy,
	}
}

// PropagationContext tracks constraint propagation state
type PropagationContext struct {
	// Active constraints that need to be resolved
	activeConstraints []TypeConstraintInfo
	// All original constraints (for final validation)
	originalConstraints []TypeConstraintInfo
	// Type variable bounds discovered during propagation
	typeVarBounds map[string][]typedef.Type
	// Constraint dependencies between type variables
	dependencies map[string][]string
	// Resolved type assignments
	assignments map[string]typedef.Type
	// Constraint satisfaction cache
	satisfactionCache map[string]bool
}

// TypeConstraintInfo represents a constraint with context
type TypeConstraintInfo struct {
	Variable   string
	Constraint typedef.TypeConstraint
	Source     string // Where the constraint came from (for debugging)
	Priority   int    // Higher priority constraints are resolved first
}

// PropagateConstraints propagates constraints for a generic type instantiation
func (cp *ConstraintPropagator) PropagateConstraints(
	generic *typedef.GenericType,
	inferredTypes []typedef.Type,
) (*PropagationResult, error) {
	if len(generic.TypeParameters) != len(inferredTypes) {
		return nil, fmt.Errorf("type parameter count mismatch: expected %d, got %d",
			len(generic.TypeParameters), len(inferredTypes))
	}

	// Create propagation context
	ctx := &PropagationContext{
		activeConstraints:   []TypeConstraintInfo{},
		originalConstraints: []TypeConstraintInfo{},
		typeVarBounds:       make(map[string][]typedef.Type),
		dependencies:        make(map[string][]string),
		assignments:         make(map[string]typedef.Type),
		satisfactionCache:   make(map[string]bool),
	}

	// Initialize assignments with inferred types
	for i, param := range generic.TypeParameters {
		ctx.assignments[param.Name] = inferredTypes[i]
	}

	// Add constraints from the generic type
	for _, constraint := range generic.Constraints {
		constraintInfo := TypeConstraintInfo{
			Variable:   constraint.Parameter,
			Constraint: constraint,
			Source:     fmt.Sprintf("generic_%s", generic.Name),
			Priority:   getPriorityForConstraintKind(constraint.Kind),
		}
		ctx.activeConstraints = append(ctx.activeConstraints, constraintInfo)
		ctx.originalConstraints = append(ctx.originalConstraints, constraintInfo)
	}

	// Propagate constraints
	if err := cp.propagateConstraintsRecursive(ctx); err != nil {
		return nil, fmt.Errorf("constraint propagation failed: %w", err)
	}

	// Build result
	result := &PropagationResult{
		FinalAssignments: ctx.assignments,
		ConstraintChecks: make([]ConstraintCheckResult, 0),
		Dependencies:     ctx.dependencies,
	}

	// Validate all original constraints
	for _, constraintInfo := range ctx.originalConstraints {
		checkResult := cp.validateConstraintAssignment(constraintInfo, ctx)
		result.ConstraintChecks = append(result.ConstraintChecks, checkResult)
		if !checkResult.Satisfied {
			return result, fmt.Errorf("constraint validation failed: %s", checkResult.Error)
		}
	}

	return result, nil
}

// PropagationResult contains the results of constraint propagation
type PropagationResult struct {
	FinalAssignments map[string]typedef.Type
	ConstraintChecks []ConstraintCheckResult
	Dependencies     map[string][]string
}

// ConstraintCheckResult represents the result of checking a single constraint
type ConstraintCheckResult struct {
	Variable   string
	Constraint typedef.TypeConstraint
	Satisfied  bool
	Error      string
}

// propagateConstraintsRecursive recursively propagates constraints
func (cp *ConstraintPropagator) propagateConstraintsRecursive(ctx *PropagationContext) error {
	maxIterations := 100 // Prevent infinite loops
	iteration := 0

	for len(ctx.activeConstraints) > 0 && iteration < maxIterations {
		iteration++

		// Sort constraints by priority
		cp.sortConstraintsByPriority(ctx.activeConstraints)

		// Process highest priority constraint
		constraintInfo := ctx.activeConstraints[0]
		ctx.activeConstraints = ctx.activeConstraints[1:]

		if err := cp.processConstraint(constraintInfo, ctx); err != nil {
			return fmt.Errorf("failed to process constraint for %s: %w",
				constraintInfo.Variable, err)
		}
	}

	if iteration >= maxIterations {
		return fmt.Errorf("constraint propagation exceeded maximum iterations (%d)", maxIterations)
	}

	return nil
}

// processConstraint processes a single constraint and may add new constraints
func (cp *ConstraintPropagator) processConstraint(
	constraintInfo TypeConstraintInfo,
	ctx *PropagationContext,
) error {
	variable := constraintInfo.Variable
	constraint := constraintInfo.Constraint

	// Get current assignment for the variable
	assignment, hasAssignment := ctx.assignments[variable]
	if !hasAssignment {
		// If no assignment, add bounds instead
		return cp.addConstraintBounds(variable, constraint, ctx)
	}

	// Check if constraint is satisfied by current assignment
	satisfied, err := cp.isConstraintSatisfied(assignment, constraint)
	if err != nil {
		return fmt.Errorf("error checking constraint satisfaction: %w", err)
	}

	if satisfied {
		// Constraint is satisfied, no further action needed
		return nil
	}

	// Try to refine the assignment to satisfy the constraint
	refinedType, err := cp.refineTypeForConstraint(assignment, constraint)
	if err != nil {
		return fmt.Errorf("failed to refine type for constraint: %w", err)
	}

	if refinedType != nil {
		// Update assignment
		ctx.assignments[variable] = refinedType
		// Add dependencies for any type variables in the refined type
		cp.extractDependencies(variable, refinedType, ctx)
	}

	return nil
}

// addConstraintBounds adds constraint bounds for unassigned variables
func (cp *ConstraintPropagator) addConstraintBounds(
	variable string,
	constraint typedef.TypeConstraint,
	ctx *PropagationContext,
) error {
	// Add the constraint expression as a bound
	if constraint.Expression != nil {
		ctx.typeVarBounds[variable] = append(ctx.typeVarBounds[variable], constraint.Expression)
	}

	// For trait constraints, add the trait as a bound
	if constraint.Kind == typedef.TraitConstraint {
		// The variable must implement this trait
		// This creates a subtype relationship
		cp.addDependency(variable, constraint.Expression.GetName(), ctx)
	}

	return nil
}

// isConstraintSatisfied checks if a type satisfies a constraint
func (cp *ConstraintPropagator) isConstraintSatisfied(
	typ typedef.Type,
	constraint typedef.TypeConstraint,
) (bool, error) {
	switch constraint.Kind {
	case typedef.TraitConstraint:
		// Check if type implements the trait
		return cp.checkTraitImplementation(typ, constraint.Expression)
	case typedef.ProtocolConstraint:
		// Check if type implements the protocol
		return cp.checkProtocolImplementation(typ, constraint.Expression)
	case typedef.ValueConstraint:
		// Value constraints are runtime checks, assume satisfied for now
		return true, nil
	default:
		return false, fmt.Errorf("unknown constraint kind: %d", constraint.Kind)
	}
}

// checkTraitImplementation checks if a type implements a trait
func (cp *ConstraintPropagator) checkTraitImplementation(
	typ typedef.Type,
	trait typedef.Type,
) (bool, error) {
	// For now, use type hierarchy to check compatibility
	// In a full implementation, this would check method signatures, etc.
	if err := cp.hierarchy.CheckTypeCompatibility(typ.GetName(), trait.GetName()); err != nil {
		return false, nil // Not compatible
	}
	return true, nil
}

// checkProtocolImplementation checks if a type implements a protocol
func (cp *ConstraintPropagator) checkProtocolImplementation(
	typ typedef.Type,
	protocol typedef.Type,
) (bool, error) {
	// Similar to trait implementation but with protocol semantics
	if err := cp.hierarchy.CheckTypeCompatibility(typ.GetName(), protocol.GetName()); err != nil {
		return false, nil // Not compatible
	}
	return true, nil
}

// refineTypeForConstraint attempts to refine a type to satisfy a constraint
func (cp *ConstraintPropagator) refineTypeForConstraint(
	typ typedef.Type,
	constraint typedef.TypeConstraint,
) (typedef.Type, error) {
	switch constraint.Kind {
	case typedef.TraitConstraint:
		// Try to create an intersection type that includes both the original type and the trait
		return cp.createIntersectionWithTrait(typ, constraint.Expression)
	case typedef.ProtocolConstraint:
		// Similar to trait but with protocol semantics
		return cp.createIntersectionWithProtocol(typ, constraint.Expression)
	default:
		// Can't refine for other constraint types
		return nil, fmt.Errorf("cannot refine type for constraint kind %d", constraint.Kind)
	}
}

// createIntersectionWithTrait creates an intersection type with a trait
func (cp *ConstraintPropagator) createIntersectionWithTrait(
	typ typedef.Type,
	trait typedef.Type,
) (typedef.Type, error) {
	// If the type already satisfies the trait, no refinement needed
	compatible, err := cp.checkTraitImplementation(typ, trait)
	if err != nil {
		return nil, err
	}
	if compatible {
		return typ, nil // No refinement needed
	}

	// Create an intersection type
	intersectionType := cp.hierarchy.CreateIntersectionType([]string{
		typ.GetName(),
		trait.GetName(),
	})

	return &typedef.SimpleType{Name: intersectionType.TypeName()}, nil
}

// createIntersectionWithProtocol creates an intersection type with a protocol
func (cp *ConstraintPropagator) createIntersectionWithProtocol(
	typ typedef.Type,
	protocol typedef.Type,
) (typedef.Type, error) {
	// Similar logic to trait intersection
	compatible, err := cp.checkProtocolImplementation(typ, protocol)
	if err != nil {
		return nil, err
	}
	if compatible {
		return typ, nil // No refinement needed
	}

	// Create an intersection type
	intersectionType := cp.hierarchy.CreateIntersectionType([]string{
		typ.GetName(),
		protocol.GetName(),
	})

	return &typedef.SimpleType{Name: intersectionType.TypeName()}, nil
}

// extractDependencies extracts type variable dependencies from a type
func (cp *ConstraintPropagator) extractDependencies(
	variable string,
	typ typedef.Type,
	ctx *PropagationContext,
) {
	switch t := typ.(type) {
	case *TypeVariable:
		cp.addDependency(variable, t.Name, ctx)
	case *typedef.ParameterizedType:
		for _, arg := range t.Arguments {
			cp.extractDependencies(variable, arg, ctx)
		}
	}
}

// addDependency adds a dependency relationship between variables
func (cp *ConstraintPropagator) addDependency(from, to string, ctx *PropagationContext) {
	if ctx.dependencies[from] == nil {
		ctx.dependencies[from] = []string{}
	}
	// Check if dependency already exists
	for _, existing := range ctx.dependencies[from] {
		if existing == to {
			return // Already exists
		}
	}
	ctx.dependencies[from] = append(ctx.dependencies[from], to)
}

// validateConstraintAssignment validates that an assignment satisfies a constraint
func (cp *ConstraintPropagator) validateConstraintAssignment(
	constraintInfo TypeConstraintInfo,
	ctx *PropagationContext,
) ConstraintCheckResult {
	result := ConstraintCheckResult{
		Variable:   constraintInfo.Variable,
		Constraint: constraintInfo.Constraint,
		Satisfied:  false,
		Error:      "",
	}

	assignment, hasAssignment := ctx.assignments[constraintInfo.Variable]
	if !hasAssignment {
		result.Error = fmt.Sprintf("no assignment for variable %s", constraintInfo.Variable)
		return result
	}

	satisfied, err := cp.isConstraintSatisfied(assignment, constraintInfo.Constraint)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	result.Satisfied = satisfied
	if !satisfied {
		result.Error = fmt.Sprintf("assignment %s does not satisfy constraint %s",
			assignment.GetName(), constraintInfo.Constraint.Expression.GetName())
	}

	return result
}

// sortConstraintsByPriority sorts constraints by priority (highest first)
func (cp *ConstraintPropagator) sortConstraintsByPriority(constraints []TypeConstraintInfo) {
	// Use efficient sort.Slice for O(n log n) performance
	sort.Slice(constraints, func(i, j int) bool {
		return constraints[i].Priority > constraints[j].Priority
	})
}

// getPriorityForConstraintKind returns the priority for a constraint kind
func getPriorityForConstraintKind(kind typedef.ConstraintKind) int {
	switch kind {
	case typedef.TraitConstraint:
		return 100 // High priority - affects type structure
	case typedef.ProtocolConstraint:
		return 90 // High priority - affects behavior
	case typedef.ValueConstraint:
		return 50 // Medium priority - runtime constraint
	default:
		return 10 // Low priority
	}
}
