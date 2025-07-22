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
