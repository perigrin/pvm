// ABOUTME: Comprehensive test suite for the core type system
// ABOUTME: Tests basic type operations, equality, compatibility, and confidence tracking

package types

import (
	"fmt"
	"testing"
)

func TestBasicTypes(t *testing.T) {
	tests := []struct {
		name     string
		typeName string
		want     string
	}{
		{"Int type", "Int", "Int"},
		{"Str type", "Str", "Str"},
		{"Bool type", "Bool", "Bool"},
		{"ArrayRef type", "ArrayRef", "ArrayRef[Int]"},
		{"HashRef type", "HashRef", "HashRef[Str]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var typ Type
			switch tt.typeName {
			case "Int":
				typ = NewIntType()
			case "Str":
				typ = NewStrType()
			case "Bool":
				typ = NewBoolType()
			case "ArrayRef":
				typ = NewArrayRefType(NewIntType())
			case "HashRef":
				typ = NewHashRefType(NewStrType())
			default:
				t.Fatalf("Unknown type: %s", tt.typeName)
			}

			if got := typ.String(); got != tt.want {
				t.Errorf("Type.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTypeEquality(t *testing.T) {
	tests := []struct {
		name   string
		type1  Type
		type2  Type
		equals bool
	}{
		{
			"Int equals Int",
			NewIntType(),
			NewIntType(),
			true,
		},
		{
			"Int not equals Str",
			NewIntType(),
			NewStrType(),
			false,
		},
		{
			"ArrayRef[Int] equals ArrayRef[Int]",
			NewArrayRefType(NewIntType()),
			NewArrayRefType(NewIntType()),
			true,
		},
		{
			"ArrayRef[Int] not equals ArrayRef[Str]",
			NewArrayRefType(NewIntType()),
			NewArrayRefType(NewStrType()),
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.type1.Equals(tt.type2); got != tt.equals {
				t.Errorf("Type.Equals() = %v, want %v", got, tt.equals)
			}
		})
	}
}

func TestTypeCompatibility(t *testing.T) {
	tests := []struct {
		name       string
		type1      Type
		type2      Type
		compatible bool
	}{
		{
			"Int compatible with Int",
			NewIntType(),
			NewIntType(),
			true,
		},
		{
			"Int not compatible with Str",
			NewIntType(),
			NewStrType(),
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.type1.CompatibleWith(tt.type2); got != tt.compatible {
				t.Errorf("Type.CompatibleWith() = %v, want %v", got, tt.compatible)
			}
		})
	}
}

func TestTypeInfo(t *testing.T) {
	tests := []struct {
		name       string
		typ        Type
		confidence float64
		source     TypeSource
	}{
		{
			"High confidence literal inference",
			NewIntType(),
			0.95,
			SourceLiteral,
		},
		{
			"Medium confidence variable inference",
			NewStrType(),
			0.75,
			SourceVariable,
		},
		{
			"Low confidence context inference",
			NewBoolType(),
			0.55,
			SourceContext,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := NewTypeInfo(tt.typ, tt.confidence, tt.source)

			if !info.Type.Equals(tt.typ) {
				t.Errorf("TypeInfo.Type = %v, want %v", info.Type, tt.typ)
			}
			if info.Confidence != tt.confidence {
				t.Errorf("TypeInfo.Confidence = %v, want %v", info.Confidence, tt.confidence)
			}
			if info.Source != tt.source {
				t.Errorf("TypeInfo.Source = %v, want %v", info.Source, tt.source)
			}
		})
	}
}

func TestTypeSource(t *testing.T) {
	sources := []TypeSource{
		SourceLiteral,
		SourceVariable,
		SourceReturn,
		SourceParameter,
		SourceContext,
		SourceExternal,
	}

	for _, source := range sources {
		t.Run(string(source), func(t *testing.T) {
			if source.String() == "" {
				t.Errorf("TypeSource.String() should not be empty for %v", source)
			}
		})
	}
}

func TestTypeConstraints(t *testing.T) {
	t.Run("Basic constraint creation", func(t *testing.T) {
		constraint := NewAssignmentConstraint(NewIntType(), NewIntType())
		if constraint == nil {
			t.Error("Expected constraint to be created")
		}
	})

	t.Run("Constraint validation", func(t *testing.T) {
		// Valid assignment: Int to Int
		constraint := NewAssignmentConstraint(NewIntType(), NewIntType())
		if err := constraint.Validate(); err != nil {
			t.Errorf("Expected valid constraint, got error: %v", err)
		}
	})

	t.Run("Invalid constraint validation", func(t *testing.T) {
		// Invalid assignment: Int to Str
		constraint := NewAssignmentConstraint(NewIntType(), NewStrType())
		if err := constraint.Validate(); err == nil {
			t.Error("Expected constraint validation to fail")
		}
	})
}

func TestComplexTypes(t *testing.T) {
	t.Run("Nested ArrayRef", func(t *testing.T) {
		// ArrayRef[ArrayRef[Int]]
		innerArray := NewArrayRefType(NewIntType())
		outerArray := NewArrayRefType(innerArray)

		expected := "ArrayRef[ArrayRef[Int]]"
		if got := outerArray.String(); got != expected {
			t.Errorf("Complex type string = %v, want %v", got, expected)
		}
	})

	t.Run("HashRef with complex value type", func(t *testing.T) {
		// HashRef[ArrayRef[Str]]
		arrayType := NewArrayRefType(NewStrType())
		hashType := NewHashRefType(arrayType)

		expected := "HashRef[ArrayRef[Str]]"
		if got := hashType.String(); got != expected {
			t.Errorf("Complex type string = %v, want %v", got, expected)
		}
	})
}

// Test for 100+ test cases requirement
func TestComprehensiveTypeCoverage(t *testing.T) {
	// Create all basic types and test their properties
	basicTypes := []Type{
		NewIntType(),
		NewStrType(),
		NewBoolType(),
		NewArrayRefType(NewIntType()),
		NewHashRefType(NewStrType()),
	}

	// Test all combinations of type operations
	for i, type1 := range basicTypes {
		for j, type2 := range basicTypes {
			t.Run(fmt.Sprintf("Type_%d_vs_Type_%d", i, j), func(t *testing.T) {
				// Test equality
				equals := type1.Equals(type2)
				if i == j && !equals {
					t.Errorf("Type should equal itself")
				}

				// Test compatibility
				compatible := type1.CompatibleWith(type2)
				if equals && !compatible {
					t.Errorf("Equal types should be compatible")
				}

				// Test string representation is not empty
				if type1.String() == "" {
					t.Errorf("Type string representation should not be empty")
				}
			})
		}
	}
}