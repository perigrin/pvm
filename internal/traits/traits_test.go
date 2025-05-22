// ABOUTME: Tests for the trait system implementation
// ABOUTME: Tests trait definitions, assignments, and basic operations

package traits

import (
	"testing"
)

func TestTraitDefinition(t *testing.T) {
	// Test creating a basic trait
	trait := Trait{
		Operation:  "\"\"",
		ResultType: "Str",
	}

	if trait.Operation != "\"\"" {
		t.Errorf("Expected operation '\"\"', got %s", trait.Operation)
	}

	if trait.ResultType != "Str" {
		t.Errorf("Expected result type 'Str', got %s", trait.ResultType)
	}
}

func TestCoreOperationDefinitions(t *testing.T) {
	// Test that all core overload.pm operations are defined
	coreOps := GetCoreOperations()

	expectedOps := []string{
		"\"\"", // string conversion
		"0+",   // numeric conversion
		"bool", // boolean conversion
		"+",    // addition
		"-",    // subtraction
		"*",    // multiplication
		"/",    // division
		"%",    // modulus
		"**",   // exponentiation
		"<=>",  // numeric comparison
		"cmp",  // string comparison
		"&",    // bitwise and
		"|",    // bitwise or
		"^",    // bitwise xor
		"~",    // bitwise not
		"<<",   // left shift
		">>",   // right shift
		"eq",   // string equality
		"ne",   // string inequality
		"lt",   // string less than
		"le",   // string less than or equal
		"gt",   // string greater than
		"ge",   // string greater than or equal
	}

	if len(coreOps) != len(expectedOps) {
		t.Errorf("Expected %d core operations, got %d", len(expectedOps), len(coreOps))
	}

	// Check that all expected operations are present
	for _, expectedOp := range expectedOps {
		found := false
		for _, op := range coreOps {
			if op.Operation == expectedOp {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Core operation '%s' not found", expectedOp)
		}
	}
}

func TestTraitAssignment(t *testing.T) {
	// Test assigning traits to a type
	traits := NewTraitSet()

	// Add stringification trait
	stringTrait := Trait{
		Operation:  "\"\"",
		ResultType: "Str",
	}
	traits.AddTrait(stringTrait)

	// Add numeric conversion trait
	numTrait := Trait{
		Operation:  "0+",
		ResultType: "Num",
	}
	traits.AddTrait(numTrait)

	// Test that traits were added
	if !traits.HasTrait("\"\"") {
		t.Error("Expected trait '\"\"' to be present")
	}

	if !traits.HasTrait("0+") {
		t.Error("Expected trait '0+' to be present")
	}

	// Test getting trait result type
	if resultType := traits.GetResultType("\"\""); resultType != "Str" {
		t.Errorf("Expected result type 'Str' for '\"\"', got %s", resultType)
	}

	if resultType := traits.GetResultType("0+"); resultType != "Num" {
		t.Errorf("Expected result type 'Num' for '0+', got %s", resultType)
	}
}

func TestBasicTypeDefaultTraits(t *testing.T) {
	tests := []struct {
		typeName       string
		expectedTraits []string
	}{
		{
			typeName:       "Int",
			expectedTraits: []string{"\"\"", "0+", "bool", "+", "-", "*", "/", "%", "**", "<=>", "&", "|", "^", "~", "<<", ">>"},
		},
		{
			typeName:       "Str",
			expectedTraits: []string{"\"\"", "0+", "bool", "cmp", "eq", "ne", "lt", "le", "gt", "ge"},
		},
		{
			typeName:       "Num",
			expectedTraits: []string{"\"\"", "0+", "bool", "+", "-", "*", "/", "%", "**", "<=>"},
		},
		{
			typeName:       "Bool",
			expectedTraits: []string{"\"\"", "0+", "bool", "&", "|", "^", "~"},
		},
		{
			typeName:       "ArrayRef",
			expectedTraits: []string{"\"\"", "0+", "bool"},
		},
		{
			typeName:       "HashRef",
			expectedTraits: []string{"\"\"", "0+", "bool"},
		},
	}

	for _, test := range tests {
		t.Run(test.typeName, func(t *testing.T) {
			traits := GetDefaultTraitsForType(test.typeName)

			for _, expectedTrait := range test.expectedTraits {
				if !traits.HasTrait(expectedTrait) {
					t.Errorf("Type %s should have trait '%s'", test.typeName, expectedTrait)
				}
			}
		})
	}
}

func TestTraitQueryOperations(t *testing.T) {
	// Test querying traits on a type
	traits := GetDefaultTraitsForType("Int")

	// Test basic queries
	if !traits.HasTrait("+") {
		t.Error("Int should support addition")
	}

	if traits.HasTrait("cmp") {
		t.Error("Int should not support string comparison")
	}

	// Test result type queries
	if resultType := traits.GetResultType("+"); resultType != "Num" {
		t.Errorf("Int addition should result in Num, got %s", resultType)
	}

	if resultType := traits.GetResultType("\"\""); resultType != "Str" {
		t.Errorf("Int stringification should result in Str, got %s", resultType)
	}
}

func TestTraitSetOperations(t *testing.T) {
	// Test basic trait set operations
	traits := NewTraitSet()

	// Test empty set
	if traits.HasTrait("+") {
		t.Error("Empty trait set should not have any traits")
	}

	// Add a trait
	trait := Trait{
		Operation:  "+",
		ResultType: "Num",
	}
	traits.AddTrait(trait)

	if !traits.HasTrait("+") {
		t.Error("Trait set should have added trait")
	}

	// Test removing a trait
	traits.RemoveTrait("+")

	if traits.HasTrait("+") {
		t.Error("Trait should be removed")
	}
}

func TestTraitSetClone(t *testing.T) {
	// Test cloning trait sets
	original := NewTraitSet()
	original.AddTrait(Trait{Operation: "+", ResultType: "Num"})
	original.AddTrait(Trait{Operation: "\"\"", ResultType: "Str"})

	clone := original.Clone()

	// Verify clone has same traits
	if !clone.HasTrait("+") {
		t.Error("Clone should have '+' trait")
	}

	if !clone.HasTrait("\"\"") {
		t.Error("Clone should have '\"\"' trait")
	}

	// Verify modifying clone doesn't affect original
	clone.RemoveTrait("+")

	if !original.HasTrait("+") {
		t.Error("Original should still have '+' trait after clone modification")
	}

	if clone.HasTrait("+") {
		t.Error("Clone should not have '+' trait after removal")
	}
}

func TestGetAllTraits(t *testing.T) {
	// Test getting all traits from a set
	traits := NewTraitSet()
	traits.AddTrait(Trait{Operation: "+", ResultType: "Num"})
	traits.AddTrait(Trait{Operation: "\"\"", ResultType: "Str"})

	allTraits := traits.GetAllTraits()

	if len(allTraits) != 2 {
		t.Errorf("Expected 2 traits, got %d", len(allTraits))
	}

	// Traits should be sorted by operation
	if allTraits[0].Operation != "\"\"" {
		t.Errorf("Expected first trait to be '\"\"', got %s", allTraits[0].Operation)
	}

	if allTraits[1].Operation != "+" {
		t.Errorf("Expected second trait to be '+', got %s", allTraits[1].Operation)
	}
}
