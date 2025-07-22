// ABOUTME: Tests for constraint propagation in generic type inference
// ABOUTME: Validates constraint resolution, type refinement, and dependency tracking

package typechecker

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/typedef"
)

func TestConstraintPropagator_Basic(t *testing.T) {
	// Create test hierarchy
	store, err := typedef.NewStorage()
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	hierarchy := typedef.NewTypeHierarchy(store)
	propagator := NewConstraintPropagator(hierarchy)

	// Create a simple generic type with trait constraint
	// Generic: Container<T> where T: Serializable
	genericType := &typedef.GenericType{
		Name: "Container",
		TypeParameters: []typedef.TypeParameter{
			{Name: "T"},
		},
		Constraints: []typedef.TypeConstraint{
			{
				Parameter:  "T",
				Kind:       typedef.TraitConstraint,
				Expression: &typedef.SimpleType{Name: "Serializable"},
			},
		},
		BaseType: &typedef.SimpleType{Name: "Container"},
	}

	// Test with a compatible type
	inferredTypes := []typedef.Type{
		&typedef.SimpleType{Name: "Int"}, // Assume Int implements Serializable
	}

	result, err := propagator.PropagateConstraints(genericType, inferredTypes)
	if err != nil {
		t.Fatalf("Expected constraint propagation to succeed, got error: %v", err)
	}

	// Check final assignments
	if len(result.FinalAssignments) != 1 {
		t.Fatalf("Expected 1 final assignment, got %d", len(result.FinalAssignments))
	}

	if assignment, ok := result.FinalAssignments["T"]; !ok {
		t.Fatalf("Expected assignment for T")
	} else {
		// The assignment should be either "Int" or an intersection type containing Int
		assignmentName := assignment.GetName()
		if assignmentName != "Int" && !strings.Contains(assignmentName, "Int") {
			t.Fatalf("Expected T -> Int or intersection with Int, got T -> %s", assignmentName)
		}
	}

	// Check constraint validation results
	if len(result.ConstraintChecks) != 1 {
		t.Fatalf("Expected 1 constraint check, got %d", len(result.ConstraintChecks))
	}

	constraintCheck := result.ConstraintChecks[0]
	if constraintCheck.Variable != "T" {
		t.Fatalf("Expected constraint check for T, got %s", constraintCheck.Variable)
	}

	if !constraintCheck.Satisfied {
		t.Fatalf("Expected constraint to be satisfied, got error: %s", constraintCheck.Error)
	}
}

func TestConstraintPropagator_MultipleConstraints(t *testing.T) {
	store, err := typedef.NewStorage()
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	hierarchy := typedef.NewTypeHierarchy(store)
	propagator := NewConstraintPropagator(hierarchy)

	// Generic with multiple constraints: Map<K, V> where K: Hashable, V: Serializable
	genericType := &typedef.GenericType{
		Name: "Map",
		TypeParameters: []typedef.TypeParameter{
			{Name: "K"},
			{Name: "V"},
		},
		Constraints: []typedef.TypeConstraint{
			{
				Parameter:  "K",
				Kind:       typedef.TraitConstraint,
				Expression: &typedef.SimpleType{Name: "Hashable"},
			},
			{
				Parameter:  "V",
				Kind:       typedef.TraitConstraint,
				Expression: &typedef.SimpleType{Name: "Serializable"},
			},
		},
		BaseType: &typedef.SimpleType{Name: "Map"},
	}

	inferredTypes := []typedef.Type{
		&typedef.SimpleType{Name: "Str"}, // K: Str implements Hashable
		&typedef.SimpleType{Name: "Int"}, // V: Int implements Serializable
	}

	result, err := propagator.PropagateConstraints(genericType, inferredTypes)
	if err != nil {
		t.Fatalf("Expected constraint propagation to succeed, got error: %v", err)
	}

	// Check assignments
	if len(result.FinalAssignments) != 2 {
		t.Fatalf("Expected 2 final assignments, got %d", len(result.FinalAssignments))
	}

	if kAssignment, ok := result.FinalAssignments["K"]; !ok {
		t.Fatalf("Expected assignment for K")
	} else {
		kName := kAssignment.GetName()
		if kName != "Str" && !strings.Contains(kName, "Str") {
			t.Fatalf("Expected K -> Str or intersection with Str, got K -> %s", kName)
		}
	}

	if vAssignment, ok := result.FinalAssignments["V"]; !ok {
		t.Fatalf("Expected assignment for V")
	} else {
		vName := vAssignment.GetName()
		if vName != "Int" && !strings.Contains(vName, "Int") {
			t.Fatalf("Expected V -> Int or intersection with Int, got V -> %s", vName)
		}
	}

	// Check constraint validations
	if len(result.ConstraintChecks) != 2 {
		t.Fatalf("Expected 2 constraint checks, got %d", len(result.ConstraintChecks))
	}

	for _, check := range result.ConstraintChecks {
		if !check.Satisfied {
			t.Fatalf("Expected constraint for %s to be satisfied, got error: %s",
				check.Variable, check.Error)
		}
	}
}

func TestConstraintPropagator_ConstraintPriority(t *testing.T) {
	store, err := typedef.NewStorage()
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	hierarchy := typedef.NewTypeHierarchy(store)
	propagator := NewConstraintPropagator(hierarchy)

	// Test constraint priority ordering
	constraints := []TypeConstraintInfo{
		{
			Variable:   "T",
			Constraint: typedef.TypeConstraint{Kind: typedef.ValueConstraint},
			Priority:   getPriorityForConstraintKind(typedef.ValueConstraint),
		},
		{
			Variable:   "U",
			Constraint: typedef.TypeConstraint{Kind: typedef.TraitConstraint},
			Priority:   getPriorityForConstraintKind(typedef.TraitConstraint),
		},
		{
			Variable:   "V",
			Constraint: typedef.TypeConstraint{Kind: typedef.ProtocolConstraint},
			Priority:   getPriorityForConstraintKind(typedef.ProtocolConstraint),
		},
	}

	// Sort by priority
	propagator.sortConstraintsByPriority(constraints)

	// Check order: Trait (100), Protocol (90), Value (50)
	expectedOrder := []typedef.ConstraintKind{
		typedef.TraitConstraint,
		typedef.ProtocolConstraint,
		typedef.ValueConstraint,
	}

	for i, expected := range expectedOrder {
		if constraints[i].Constraint.Kind != expected {
			t.Fatalf("Expected constraint %d to be %v, got %v",
				i, expected, constraints[i].Constraint.Kind)
		}
	}
}

func TestConstraintPropagator_DependencyTracking(t *testing.T) {
	store, err := typedef.NewStorage()
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	hierarchy := typedef.NewTypeHierarchy(store)
	propagator := NewConstraintPropagator(hierarchy)

	// Create propagation context
	ctx := &PropagationContext{
		dependencies: make(map[string][]string),
	}

	// Add some dependencies
	propagator.addDependency("T", "Serializable", ctx)
	propagator.addDependency("T", "Hashable", ctx)
	propagator.addDependency("U", "T", ctx)

	// Check dependencies were recorded
	tDeps := ctx.dependencies["T"]
	if len(tDeps) != 2 {
		t.Fatalf("Expected 2 dependencies for T, got %d", len(tDeps))
	}

	expectedDeps := map[string]bool{"Serializable": true, "Hashable": true}
	for _, dep := range tDeps {
		if !expectedDeps[dep] {
			t.Fatalf("Unexpected dependency for T: %s", dep)
		}
	}

	uDeps := ctx.dependencies["U"]
	if len(uDeps) != 1 || uDeps[0] != "T" {
		t.Fatalf("Expected U to depend on T, got %v", uDeps)
	}
}

func TestConstraintPropagator_TypeRefinement(t *testing.T) {
	store, err := typedef.NewStorage()
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	hierarchy := typedef.NewTypeHierarchy(store)
	propagator := NewConstraintPropagator(hierarchy)

	// Test type refinement for trait constraints
	originalType := &typedef.SimpleType{Name: "MyClass"}
	traitConstraint := typedef.TypeConstraint{
		Kind:       typedef.TraitConstraint,
		Expression: &typedef.SimpleType{Name: "Serializable"},
	}

	refinedType, err := propagator.refineTypeForConstraint(originalType, traitConstraint)
	if err != nil {
		t.Fatalf("Expected type refinement to succeed, got error: %v", err)
	}

	if refinedType == nil {
		t.Fatalf("Expected refined type, got nil")
	}

	// The refined type should be an intersection type or the original type if compatible
	refinedName := refinedType.GetName()
	if refinedName != "MyClass" && !strings.Contains(refinedName, "MyClass") {
		t.Fatalf("Expected refined type to contain MyClass, got %s", refinedName)
	}
}

func TestConstraintPropagator_ConstraintSatisfaction(t *testing.T) {
	store, err := typedef.NewStorage()
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	hierarchy := typedef.NewTypeHierarchy(store)
	propagator := NewConstraintPropagator(hierarchy)

	// Test trait constraint satisfaction
	typ := &typedef.SimpleType{Name: "Str"}
	traitConstraint := typedef.TypeConstraint{
		Kind:       typedef.TraitConstraint,
		Expression: &typedef.SimpleType{Name: "Scalar"},
	}

	satisfied, err := propagator.isConstraintSatisfied(typ, traitConstraint)
	if err != nil {
		t.Fatalf("Expected constraint check to succeed, got error: %v", err)
	}

	if !satisfied {
		t.Fatalf("Expected Str to satisfy Scalar trait constraint")
	}
}

func TestConstraintPropagator_ErrorHandling(t *testing.T) {
	store, err := typedef.NewStorage()
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	hierarchy := typedef.NewTypeHierarchy(store)
	propagator := NewConstraintPropagator(hierarchy)

	// Test parameter count mismatch
	genericType := &typedef.GenericType{
		Name:           "Test",
		TypeParameters: []typedef.TypeParameter{{Name: "T"}},
		Constraints:    []typedef.TypeConstraint{},
	}

	// Provide wrong number of inferred types
	inferredTypes := []typedef.Type{
		&typedef.SimpleType{Name: "Int"},
		&typedef.SimpleType{Name: "Str"}, // One too many
	}

	_, err = propagator.PropagateConstraints(genericType, inferredTypes)
	if err == nil {
		t.Fatalf("Expected error for parameter count mismatch, but propagation succeeded")
	}

	if err.Error() != "type parameter count mismatch: expected 1, got 2" {
		t.Fatalf("Expected parameter count mismatch error, got: %v", err)
	}
}

func TestConstraintPropagator_ValueConstraints(t *testing.T) {
	store, err := typedef.NewStorage()
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	hierarchy := typedef.NewTypeHierarchy(store)
	propagator := NewConstraintPropagator(hierarchy)

	// Test value constraint (runtime check)
	typ := &typedef.SimpleType{Name: "Int"}
	valueConstraint := typedef.TypeConstraint{
		Kind:       typedef.ValueConstraint,
		Expression: &typedef.SimpleType{Name: "PositiveInt"},
	}

	satisfied, err := propagator.isConstraintSatisfied(typ, valueConstraint)
	if err != nil {
		t.Fatalf("Expected value constraint check to succeed, got error: %v", err)
	}

	// Value constraints are assumed to be satisfied at compile time
	if !satisfied {
		t.Fatalf("Expected value constraint to be satisfied")
	}
}
