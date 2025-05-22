// ABOUTME: Tests for the trait intersection algorithm
// ABOUTME: Tests lazy computation, caching, and intersection operations

package traits

import (
	"testing"
)

func TestTraitIntersector(t *testing.T) {
	intersector := NewTraitIntersector()

	// Test basic intersection of two types
	intersection := intersector.IntersectTypes([]string{"Int", "Str"})

	// Int and Str should both support conversion operations
	expectedCommon := []string{"\"\"", "0+", "bool"}
	for _, operation := range expectedCommon {
		if !intersection.HasTrait(operation) {
			t.Errorf("Intersection should include operation '%s'", operation)
		}
	}

	// Int and Str should not both support arithmetic (only Int does)
	if intersection.HasTrait("+") {
		t.Error("Intersection should not include '+' operation (Str doesn't support it)")
	}

	// Int and Str should not both support string operations (only Str does)
	if intersection.HasTrait("cmp") {
		t.Error("Intersection should not include 'cmp' operation (Int doesn't support it)")
	}
}

func TestIntersectionOfIdenticalTypes(t *testing.T) {
	intersector := NewTraitIntersector()

	// Intersection of a type with itself should be the same as the type's traits
	intersection := intersector.IntersectTypes([]string{"Int", "Int"})
	intTraits := GetDefaultTraitsForType("Int")

	// Should have all Int traits
	for _, trait := range intTraits.GetAllTraits() {
		if !intersection.HasTrait(trait.Operation) {
			t.Errorf("Intersection should include Int operation '%s'", trait.Operation)
		}
	}
}

func TestIntersectionOfThreeTypes(t *testing.T) {
	intersector := NewTraitIntersector()

	// Test intersection of Int, Str, and ArrayRef
	intersection := intersector.IntersectTypes([]string{"Int", "Str", "ArrayRef"})

	// All three should support basic conversions
	expectedCommon := []string{"\"\"", "0+", "bool"}
	for _, operation := range expectedCommon {
		if !intersection.HasTrait(operation) {
			t.Errorf("Three-way intersection should include operation '%s'", operation)
		}
	}

	// None should support operations that only some types support
	forbiddenOps := []string{"+", "-", "*", "/", "cmp", "eq", "ne"}
	for _, operation := range forbiddenOps {
		if intersection.HasTrait(operation) {
			t.Errorf("Three-way intersection should not include operation '%s'", operation)
		}
	}
}

func TestEmptyIntersection(t *testing.T) {
	intersector := NewTraitIntersector()

	// Create a custom type with no common operations
	customTraits := NewTraitSet()
	customTraits.AddTrait(Trait{Operation: "custom_op", ResultType: "CustomType"})
	intersector.SetTraitsForType("CustomType", customTraits)

	// Intersection of CustomType with Int should be minimal
	intersection := intersector.IntersectTypes([]string{"CustomType", "Int"})

	// Should have no traits since CustomType doesn't have the basic conversion traits
	if len(intersection.GetAllTraits()) > 0 {
		t.Errorf("Expected empty intersection, got %d traits", len(intersection.GetAllTraits()))
	}
}

func TestSingleTypeIntersection(t *testing.T) {
	intersector := NewTraitIntersector()

	// Intersection of a single type should be the type itself
	intersection := intersector.IntersectTypes([]string{"Int"})
	intTraits := GetDefaultTraitsForType("Int")

	// Should have all Int traits
	for _, trait := range intTraits.GetAllTraits() {
		if !intersection.HasTrait(trait.Operation) {
			t.Errorf("Single type intersection should include operation '%s'", trait.Operation)
		}
	}
}

func TestLazyComputation(t *testing.T) {
	// Create a custom intersector that tracks computation calls
	computationCount := 0
	testIntersector := &TraitIntersector{
		cache:     make(map[string]*TraitSet),
		resolver:  NewOperationResolver(),
		onCompute: func() { computationCount++ },
	}

	// First call should trigger computation
	_ = testIntersector.IntersectTypes([]string{"Int", "Str"})
	if computationCount != 1 {
		t.Errorf("Expected 1 computation, got %d", computationCount)
	}

	// Second call with same types should use cache
	_ = testIntersector.IntersectTypes([]string{"Int", "Str"})
	if computationCount != 1 {
		t.Errorf("Expected cached result (1 computation), got %d", computationCount)
	}

	// Different order should still use cache (if normalized)
	_ = testIntersector.IntersectTypes([]string{"Str", "Int"})
	if computationCount > 2 {
		t.Errorf("Expected at most 2 computations (order normalization), got %d", computationCount)
	}
}

func TestIntersectionCaching(t *testing.T) {
	intersector := NewTraitIntersector()

	// Perform intersection multiple times
	result1 := intersector.IntersectTypes([]string{"Int", "Str"})
	result2 := intersector.IntersectTypes([]string{"Int", "Str"})

	// Results should be equivalent (though may be different objects)
	traits1 := result1.GetAllTraits()
	traits2 := result2.GetAllTraits()

	if len(traits1) != len(traits2) {
		t.Errorf("Cached results should be equivalent: %d vs %d traits", len(traits1), len(traits2))
	}

	// Check that all traits match
	for _, trait1 := range traits1 {
		found := false
		for _, trait2 := range traits2 {
			if trait1.Operation == trait2.Operation && trait1.ResultType == trait2.ResultType {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Cached result missing trait: %+v", trait1)
		}
	}
}

func TestIntersectionResultTypes(t *testing.T) {
	intersector := NewTraitIntersector()

	// Test that intersection preserves correct result types
	intersection := intersector.IntersectTypes([]string{"Int", "Str"})

	tests := []struct {
		operation  string
		resultType string
	}{
		{"\"\"", "Str"},
		{"0+", "Num"},
		{"bool", "Bool"},
	}

	for _, test := range tests {
		if !intersection.HasTrait(test.operation) {
			t.Errorf("Intersection should include operation '%s'", test.operation)
			continue
		}

		actualResult := intersection.GetResultType(test.operation)
		if actualResult != test.resultType {
			t.Errorf("Operation '%s' should result in '%s', got '%s'",
				test.operation, test.resultType, actualResult)
		}
	}
}

func TestIntersectionWithUnknownTypes(t *testing.T) {
	intersector := NewTraitIntersector()

	// Test intersection with unknown types
	intersection := intersector.IntersectTypes([]string{"Int", "UnknownType"})

	// Should include operations that both Int and unknown types support
	// Unknown types get basic conversions
	expectedCommon := []string{"\"\"", "0+", "bool"}
	for _, operation := range expectedCommon {
		if !intersection.HasTrait(operation) {
			t.Errorf("Intersection with unknown type should include operation '%s'", operation)
		}
	}

	// Should not include Int-specific operations
	if intersection.HasTrait("+") {
		t.Error("Intersection with unknown type should not include '+' operation")
	}
}

func TestCacheInvalidation(t *testing.T) {
	intersector := NewTraitIntersector()

	// Perform intersection
	result1 := intersector.IntersectTypes([]string{"Int", "Str"})

	// Clear cache
	intersector.ClearCache()

	// Perform same intersection again
	result2 := intersector.IntersectTypes([]string{"Int", "Str"})

	// Results should still be equivalent
	if len(result1.GetAllTraits()) != len(result2.GetAllTraits()) {
		t.Error("Results should be equivalent after cache clear")
	}
}

func TestIntersectionEdgeCases(t *testing.T) {
	intersector := NewTraitIntersector()

	// Test empty type list
	intersection := intersector.IntersectTypes([]string{})
	if len(intersection.GetAllTraits()) != 0 {
		t.Error("Empty type list should result in empty intersection")
	}

	// Test with duplicate types
	intersection = intersector.IntersectTypes([]string{"Int", "Int", "Str", "Str"})

	// Should be same as Int ∩ Str
	normalIntersection := intersector.IntersectTypes([]string{"Int", "Str"})
	if len(intersection.GetAllTraits()) != len(normalIntersection.GetAllTraits()) {
		t.Error("Duplicate types should not affect intersection result")
	}
}

func TestIntersectionUtilities(t *testing.T) {
	intersector := NewTraitIntersector()

	// Test utility functions
	types := []string{"Int", "Str", "Num"}

	// Test getting common operations
	commonOps := intersector.GetCommonOperations(types)

	// All three support conversions
	expectedCommon := []string{"\"\"", "0+", "bool"}
	for _, expected := range expectedCommon {
		found := false
		for _, op := range commonOps {
			if op == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Common operations should include '%s'", expected)
		}
	}

	// Test checking if operation is supported by all types
	if !intersector.IsOperationSupportedByAll(types, "\"\"") {
		t.Error("String conversion should be supported by all types")
	}

	if intersector.IsOperationSupportedByAll(types, "+") {
		t.Error("Addition should not be supported by all types (Str doesn't support it)")
	}
}
