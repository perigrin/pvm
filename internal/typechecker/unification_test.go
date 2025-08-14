// ABOUTME: Tests for the type unification algorithm used in generic type inference
// ABOUTME: Validates constraint solving, type variable substitution, and generic instantiation

package typechecker

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/typedef"
)

func TestTypeUnifier_UnifySimpleTypes(t *testing.T) {
	hierarchy := typedef.NewTypeHierarchy(nil)
	unifier := NewTypeUnifier(hierarchy)

	// Test unifying identical simple types
	t1 := &typedef.SimpleType{Name: "Int"}
	t2 := &typedef.SimpleType{Name: "Int"}

	subs, err := unifier.Unify(t1, t2)
	if err != nil {
		t.Fatalf("Expected unification to succeed, got error: %v", err)
	}
	if len(subs) != 0 {
		t.Fatalf("Expected no substitutions for identical types, got: %v", subs)
	}
}

func TestTypeUnifier_UnifyTypeVariable(t *testing.T) {
	hierarchy := typedef.NewTypeHierarchy(nil)
	unifier := NewTypeUnifier(hierarchy)

	// Test unifying type variable with concrete type
	typeVar := &TypeVariable{Name: "T"}
	concrete := &typedef.SimpleType{Name: "Int"}

	subs, err := unifier.Unify(typeVar, concrete)
	if err != nil {
		t.Fatalf("Expected unification to succeed, got error: %v", err)
	}

	if len(subs) != 1 {
		t.Fatalf("Expected 1 substitution, got %d", len(subs))
	}

	if sub, exists := subs["T"]; !exists {
		t.Fatalf("Expected substitution for T")
	} else if sub.GetName() != "Int" {
		t.Fatalf("Expected T -> Int, got T -> %s", sub.GetName())
	}
}

func TestTypeUnifier_UnifyParameterizedTypes(t *testing.T) {
	hierarchy := typedef.NewTypeHierarchy(nil)
	unifier := NewTypeUnifier(hierarchy)

	// Test unifying Array[T] with Array[Int]
	baseType := &typedef.SimpleType{Name: "Array"}
	typeVar := &TypeVariable{Name: "T"}
	intType := &typedef.SimpleType{Name: "Int"}

	arrayT := &typedef.ParameterizedType{
		BaseType:  baseType,
		Arguments: []typedef.Type{typeVar},
	}
	arrayInt := &typedef.ParameterizedType{
		BaseType:  baseType,
		Arguments: []typedef.Type{intType},
	}

	subs, err := unifier.Unify(arrayT, arrayInt)
	if err != nil {
		t.Fatalf("Expected unification to succeed, got error: %v", err)
	}

	if len(subs) != 1 {
		t.Fatalf("Expected 1 substitution, got %d", len(subs))
	}

	if sub, exists := subs["T"]; !exists {
		t.Fatalf("Expected substitution for T")
	} else if sub.GetName() != "Int" {
		t.Fatalf("Expected T -> Int, got T -> %s", sub.GetName())
	}
}

func TestTypeUnifier_OccursCheck(t *testing.T) {
	hierarchy := typedef.NewTypeHierarchy(nil)
	unifier := NewTypeUnifier(hierarchy)

	// Test occurs check: prevent T = Array[T] (infinite type)
	typeVar := &TypeVariable{Name: "T"}
	baseType := &typedef.SimpleType{Name: "Array"}
	infiniteType := &typedef.ParameterizedType{
		BaseType:  baseType,
		Arguments: []typedef.Type{typeVar},
	}

	_, err := unifier.Unify(typeVar, infiniteType)
	if err == nil {
		t.Fatalf("Expected occurs check to fail, but unification succeeded")
	}

	if !strings.Contains(err.Error(), "occurs check failed") {
		t.Fatalf("Expected occurs check error, got: %v", err)
	}
}

func TestTypeUnifier_InstantiateGeneric(t *testing.T) {
	hierarchy := typedef.NewTypeHierarchy(nil)
	unifier := NewTypeUnifier(hierarchy)

	// Create a generic type List<T>
	baseType := &typedef.SimpleType{Name: "List"}
	genericType := &typedef.GenericType{
		Name: "List",
		TypeParameters: []typedef.TypeParameter{
			{Name: "T"},
		},
		BaseType: baseType,
	}

	// Try to unify with List[Int]
	intType := &typedef.SimpleType{Name: "Int"}
	concreteType := &typedef.ParameterizedType{
		BaseType:  baseType,
		Arguments: []typedef.Type{intType},
	}

	subs, err := unifier.Unify(genericType, concreteType)
	if err != nil {
		t.Fatalf("Expected generic instantiation to succeed, got error: %v", err)
	}

	// Should infer T = Int
	if len(subs) != 1 {
		t.Fatalf("Expected 1 substitution, got %d: %v", len(subs), subs)
	}
}

func TestTypeUnifier_SubstitutionApplication(t *testing.T) {
	hierarchy := typedef.NewTypeHierarchy(nil)
	unifier := NewTypeUnifier(hierarchy)

	// Create substitutions map
	subs := map[string]typedef.Type{
		"T": &typedef.SimpleType{Name: "Int"},
		"U": &typedef.SimpleType{Name: "String"},
	}

	// Test applying substitutions to Array[T]
	baseType := &typedef.SimpleType{Name: "Array"}
	typeVar := &TypeVariable{Name: "T"}
	arrayT := &typedef.ParameterizedType{
		BaseType:  baseType,
		Arguments: []typedef.Type{typeVar},
	}

	result := unifier.applySubstitutions(arrayT, subs)
	if resultParam, ok := result.(*typedef.ParameterizedType); ok {
		if len(resultParam.Arguments) != 1 {
			t.Fatalf("Expected 1 argument, got %d", len(resultParam.Arguments))
		}
		if resultParam.Arguments[0].GetName() != "Int" {
			t.Fatalf("Expected Int, got %s", resultParam.Arguments[0].GetName())
		}
	} else {
		t.Fatalf("Expected ParameterizedType result, got %T", result)
	}
}

func TestTypeUnifier_ChainedSubstitutions(t *testing.T) {
	hierarchy := typedef.NewTypeHierarchy(nil)
	unifier := NewTypeUnifier(hierarchy)

	// Test T -> U -> Int (chained substitutions)
	subs := map[string]typedef.Type{
		"T": &TypeVariable{Name: "U"},
		"U": &typedef.SimpleType{Name: "Int"},
	}

	typeVar := &TypeVariable{Name: "T"}
	result := unifier.applySubstitutions(typeVar, subs)

	if result.GetName() != "Int" {
		t.Fatalf("Expected Int after chained substitution, got %s", result.GetName())
	}
}

// TestTypeUnifier_UnifyGenericsWithCompatibleConstraints tests unifying two generic types
// with compatible constraints
func TestTypeUnifier_UnifyGenericsWithCompatibleConstraints(t *testing.T) {
	hierarchy := typedef.NewTypeHierarchy(nil)
	unifier := NewTypeUnifier(hierarchy)

	// Create first generic: Container<T: Serializable>
	g1 := &typedef.GenericType{
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

	// Create second generic: Container<U: Serializable> (same constraint)
	g2 := &typedef.GenericType{
		Name: "Container",
		TypeParameters: []typedef.TypeParameter{
			{Name: "U"},
		},
		Constraints: []typedef.TypeConstraint{
			{
				Parameter:  "U",
				Kind:       typedef.TraitConstraint,
				Expression: &typedef.SimpleType{Name: "Serializable"},
			},
		},
		BaseType: &typedef.SimpleType{Name: "Container"},
	}

	subs, err := unifier.Unify(g1, g2)
	if err != nil {
		t.Fatalf("Expected unification to succeed for compatible constraints, got error: %v", err)
	}
	if len(subs) != 0 {
		t.Fatalf("Expected no substitutions for compatible generics, got: %v", subs)
	}
}

// TestTypeUnifier_UnifyGenericsWithIncompatibleConstraints tests unifying two generic types
// with incompatible constraints
func TestTypeUnifier_UnifyGenericsWithIncompatibleConstraints(t *testing.T) {
	hierarchy := typedef.NewTypeHierarchy(nil)
	unifier := NewTypeUnifier(hierarchy)

	// Create first generic: Container<T: Serializable>
	g1 := &typedef.GenericType{
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

	// Create second generic: Container<U: Display> (incompatible constraint)
	g2 := &typedef.GenericType{
		Name: "Container",
		TypeParameters: []typedef.TypeParameter{
			{Name: "U"},
		},
		Constraints: []typedef.TypeConstraint{
			{
				Parameter:  "U",
				Kind:       typedef.TraitConstraint,
				Expression: &typedef.SimpleType{Name: "Display"},
			},
		},
		BaseType: &typedef.SimpleType{Name: "Container"},
	}

	_, err := unifier.Unify(g1, g2)
	if err == nil {
		t.Fatalf("Expected unification to fail for incompatible constraints")
	}
	if !strings.Contains(err.Error(), "constraint unification failed") {
		t.Fatalf("Expected constraint unification error, got: %v", err)
	}
}

// TestTypeUnifier_UnifyGenericsWithMultipleConstraints tests unifying generics with
// multiple constraints per parameter
func TestTypeUnifier_UnifyGenericsWithMultipleConstraints(t *testing.T) {
	hierarchy := typedef.NewTypeHierarchy(nil)
	unifier := NewTypeUnifier(hierarchy)

	// Create first generic: Handler<T: Serializable, T: Display>
	g1 := &typedef.GenericType{
		Name: "Handler",
		TypeParameters: []typedef.TypeParameter{
			{Name: "T"},
		},
		Constraints: []typedef.TypeConstraint{
			{
				Parameter:  "T",
				Kind:       typedef.TraitConstraint,
				Expression: &typedef.SimpleType{Name: "Serializable"},
			},
			{
				Parameter:  "T",
				Kind:       typedef.TraitConstraint,
				Expression: &typedef.SimpleType{Name: "Display"},
			},
		},
		BaseType: &typedef.SimpleType{Name: "Handler"},
	}

	// Create second generic: Handler<U: Display, U: Serializable> (same constraints, different order)
	g2 := &typedef.GenericType{
		Name: "Handler",
		TypeParameters: []typedef.TypeParameter{
			{Name: "U"},
		},
		Constraints: []typedef.TypeConstraint{
			{
				Parameter:  "U",
				Kind:       typedef.TraitConstraint,
				Expression: &typedef.SimpleType{Name: "Display"},
			},
			{
				Parameter:  "U",
				Kind:       typedef.TraitConstraint,
				Expression: &typedef.SimpleType{Name: "Serializable"},
			},
		},
		BaseType: &typedef.SimpleType{Name: "Handler"},
	}

	subs, err := unifier.Unify(g1, g2)
	if err != nil {
		t.Fatalf("Expected unification to succeed for compatible multiple constraints, got error: %v", err)
	}
	if len(subs) != 0 {
		t.Fatalf("Expected no substitutions for compatible generics, got: %v", subs)
	}
}

// TestTypeUnifier_UnifyGenericsWithProtocolConstraints tests unifying generics with
// protocol constraints
func TestTypeUnifier_UnifyGenericsWithProtocolConstraints(t *testing.T) {
	hierarchy := typedef.NewTypeHierarchy(nil)
	unifier := NewTypeUnifier(hierarchy)

	// Create first generic: Listener<T does EventHandler>
	g1 := &typedef.GenericType{
		Name: "Listener",
		TypeParameters: []typedef.TypeParameter{
			{Name: "T"},
		},
		Constraints: []typedef.TypeConstraint{
			{
				Parameter:  "T",
				Kind:       typedef.ProtocolConstraint,
				Expression: &typedef.SimpleType{Name: "EventHandler"},
			},
		},
		BaseType: &typedef.SimpleType{Name: "Listener"},
	}

	// Create second generic: Listener<U does EventHandler>
	g2 := &typedef.GenericType{
		Name: "Listener",
		TypeParameters: []typedef.TypeParameter{
			{Name: "U"},
		},
		Constraints: []typedef.TypeConstraint{
			{
				Parameter:  "U",
				Kind:       typedef.ProtocolConstraint,
				Expression: &typedef.SimpleType{Name: "EventHandler"},
			},
		},
		BaseType: &typedef.SimpleType{Name: "Listener"},
	}

	subs, err := unifier.Unify(g1, g2)
	if err != nil {
		t.Fatalf("Expected unification to succeed for compatible protocol constraints, got error: %v", err)
	}
	if len(subs) != 0 {
		t.Fatalf("Expected no substitutions for compatible generics, got: %v", subs)
	}
}

// TestTypeUnifier_UnifyGenericsWithMixedConstraints tests unifying generics with
// different constraint types
func TestTypeUnifier_UnifyGenericsWithMixedConstraints(t *testing.T) {
	hierarchy := typedef.NewTypeHierarchy(nil)
	unifier := NewTypeUnifier(hierarchy)

	// Create first generic: Processor<T: Serializable>
	g1 := &typedef.GenericType{
		Name: "Processor",
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
		BaseType: &typedef.SimpleType{Name: "Processor"},
	}

	// Create second generic: Processor<U does EventHandler> (different constraint type)
	g2 := &typedef.GenericType{
		Name: "Processor",
		TypeParameters: []typedef.TypeParameter{
			{Name: "U"},
		},
		Constraints: []typedef.TypeConstraint{
			{
				Parameter:  "U",
				Kind:       typedef.ProtocolConstraint,
				Expression: &typedef.SimpleType{Name: "EventHandler"},
			},
		},
		BaseType: &typedef.SimpleType{Name: "Processor"},
	}

	_, err := unifier.Unify(g1, g2)
	if err == nil {
		t.Fatalf("Expected unification to fail for different constraint types")
	}
	if !strings.Contains(err.Error(), "constraint unification failed") {
		t.Fatalf("Expected constraint unification error, got: %v", err)
	}
}

// TestTypeUnifier_UnifyGenericsWithNoConstraints tests unifying a generic with
// constraints against one without constraints
func TestTypeUnifier_UnifyGenericsWithNoConstraints(t *testing.T) {
	hierarchy := typedef.NewTypeHierarchy(nil)
	unifier := NewTypeUnifier(hierarchy)

	// Create first generic: Container<T: Serializable>
	g1 := &typedef.GenericType{
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

	// Create second generic: Container<U> (no constraints)
	g2 := &typedef.GenericType{
		Name: "Container",
		TypeParameters: []typedef.TypeParameter{
			{Name: "U"},
		},
		Constraints: []typedef.TypeConstraint{},
		BaseType:    &typedef.SimpleType{Name: "Container"},
	}

	subs, err := unifier.Unify(g1, g2)
	if err != nil {
		t.Fatalf("Expected unification to succeed when one generic has no constraints, got error: %v", err)
	}
	if len(subs) != 0 {
		t.Fatalf("Expected no substitutions, got: %v", subs)
	}
}
