// ABOUTME: Comprehensive tests for complex type implementations
// ABOUTME: Tests union types, intersection types, and enhanced parameterized types

package types

import (
	"testing"
)

func TestUnionType(t *testing.T) {
	intType := NewIntType()
	strType := NewStrType()
	boolType := NewBoolType()

	t.Run("creation and string representation", func(t *testing.T) {
		union := NewUnionType(intType, strType)
		expected := "Int|Str"
		if union.String() != expected {
			t.Errorf("Expected union string %s, got %s", expected, union.String())
		}
	})

	t.Run("three-way union", func(t *testing.T) {
		union := NewUnionType(intType, strType, boolType)
		expected := "Bool|Int|Str" // Alphabetically sorted
		if union.String() != expected {
			t.Errorf("Expected union string %s, got %s", expected, union.String())
		}
	})

	t.Run("equality", func(t *testing.T) {
		union1 := NewUnionType(intType, strType)
		union2 := NewUnionType(intType, strType)
		union3 := NewUnionType(strType, intType) // Different order
		union4 := NewUnionType(intType, boolType)

		if !union1.Equals(union2) {
			t.Error("Expected identical unions to be equal")
		}

		if !union1.Equals(union3) {
			t.Error("Expected unions with same types in different order to be equal")
		}

		if union1.Equals(union4) {
			t.Error("Expected unions with different types to not be equal")
		}
	})

	t.Run("compatibility with union members", func(t *testing.T) {
		union := NewUnionType(intType, strType)

		if !union.CompatibleWith(intType) {
			t.Error("Expected union to be compatible with member type Int")
		}

		if !union.CompatibleWith(strType) {
			t.Error("Expected union to be compatible with member type Str")
		}

		if union.CompatibleWith(boolType) {
			t.Error("Expected union to not be compatible with non-member type Bool")
		}
	})

	t.Run("member type compatibility with union", func(t *testing.T) {
		union := NewUnionType(intType, strType)

		if !intType.CompatibleWith(union) {
			t.Error("Expected Int to be compatible with union containing Int")
		}

		if !strType.CompatibleWith(union) {
			t.Error("Expected Str to be compatible with union containing Str")
		}

		if boolType.CompatibleWith(union) {
			t.Error("Expected Bool to not be compatible with union not containing Bool")
		}
	})

	t.Run("union with union compatibility", func(t *testing.T) {
		union1 := NewUnionType(intType, strType)
		union2 := NewUnionType(intType, strType, boolType)
		union3 := NewUnionType(strType, boolType)

		if !union1.CompatibleWith(union2) {
			t.Error("Expected smaller union to be compatible with larger union containing all its types")
		}

		if union2.CompatibleWith(union1) {
			t.Error("Expected larger union to not be compatible with smaller union missing some types")
		}

		if union1.CompatibleWith(union3) {
			t.Error("Expected unions with different type sets to not be compatible")
		}
	})

	t.Run("type classification", func(t *testing.T) {
		union := NewUnionType(intType, strType)

		if union.IsBasic() {
			t.Error("Expected union type to not be basic")
		}

		if !union.IsComplex() {
			t.Error("Expected union type to be complex")
		}
	})
}

func TestIntersectionType(t *testing.T) {
	// For testing, we'll need some mock types that can have meaningful intersections
	// In real Perl, this might be like Object&Serializable
	objectType := &BasicType{Name: "Object"}
	serializableType := &BasicType{Name: "Serializable"}
	readableType := &BasicType{Name: "Readable"}

	t.Run("creation and string representation", func(t *testing.T) {
		intersection := NewIntersectionType(objectType, serializableType)
		expected := "Object&Serializable"
		if intersection.String() != expected {
			t.Errorf("Expected intersection string %s, got %s", expected, intersection.String())
		}
	})

	t.Run("three-way intersection", func(t *testing.T) {
		intersection := NewIntersectionType(objectType, serializableType, readableType)
		expected := "Object&Readable&Serializable" // Alphabetically sorted
		if intersection.String() != expected {
			t.Errorf("Expected intersection string %s, got %s", expected, intersection.String())
		}
	})

	t.Run("equality", func(t *testing.T) {
		intersection1 := NewIntersectionType(objectType, serializableType)
		intersection2 := NewIntersectionType(objectType, serializableType)
		intersection3 := NewIntersectionType(serializableType, objectType) // Different order
		intersection4 := NewIntersectionType(objectType, readableType)

		if !intersection1.Equals(intersection2) {
			t.Error("Expected identical intersections to be equal")
		}

		if !intersection1.Equals(intersection3) {
			t.Error("Expected intersections with same types in different order to be equal")
		}

		if intersection1.Equals(intersection4) {
			t.Error("Expected intersections with different types to not be equal")
		}
	})

	t.Run("compatibility with intersection members", func(t *testing.T) {
		intersection := NewIntersectionType(objectType, serializableType)

		// An intersection type is compatible with something that satisfies ALL its requirements
		// For now, we'll implement basic compatibility checking
		if intersection.CompatibleWith(objectType) {
			t.Error("Expected intersection to not be compatible with single member type (needs all)")
		}

		if intersection.CompatibleWith(serializableType) {
			t.Error("Expected intersection to not be compatible with single member type (needs all)")
		}
	})

	t.Run("member type compatibility with intersection", func(t *testing.T) {
		intersection := NewIntersectionType(objectType, serializableType)

		// A type that satisfies the intersection must satisfy all parts
		// For basic implementation, we'll be conservative
		if objectType.CompatibleWith(intersection) {
			t.Error("Expected single type to not be compatible with intersection requiring multiple traits")
		}
	})

	t.Run("type classification", func(t *testing.T) {
		intersection := NewIntersectionType(objectType, serializableType)

		if intersection.IsBasic() {
			t.Error("Expected intersection type to not be basic")
		}

		if !intersection.IsComplex() {
			t.Error("Expected intersection type to be complex")
		}
	})
}

func TestParameterizedTypeEnhancements(t *testing.T) {
	intType := NewIntType()
	strType := NewStrType()

	t.Run("nested parameterized types", func(t *testing.T) {
		// ArrayRef[HashRef[Int]]
		hashRefInt := NewHashRefType(intType)
		arrayRefHashRef := NewArrayRefType(hashRefInt)

		expected := "ArrayRef[HashRef[Int]]"
		if arrayRefHashRef.String() != expected {
			t.Errorf("Expected nested parameterized type string %s, got %s", expected, arrayRefHashRef.String())
		}
	})

	t.Run("multiple parameter types", func(t *testing.T) {
		// For types that need multiple parameters like HashRef[Str, Int] (key, value)
		mapType := NewMapType(strType, intType)
		expected := "Map[Str, Int]"
		if mapType.String() != expected {
			t.Errorf("Expected multi-parameter type string %s, got %s", expected, mapType.String())
		}
	})

	t.Run("parameterized with union", func(t *testing.T) {
		union := NewUnionType(intType, strType)
		arrayRef := NewArrayRefType(union)

		expected := "ArrayRef[Int|Str]"
		if arrayRef.String() != expected {
			t.Errorf("Expected parameterized union type string %s, got %s", expected, arrayRef.String())
		}
	})
}

func TestComplexTypeInteractions(t *testing.T) {
	intType := NewIntType()
	strType := NewStrType()

	t.Run("union of parameterized types", func(t *testing.T) {
		arrayRefInt := NewArrayRefType(intType)
		arrayRefStr := NewArrayRefType(strType)
		union := NewUnionType(arrayRefInt, arrayRefStr)

		expected := "ArrayRef[Int]|ArrayRef[Str]"
		if union.String() != expected {
			t.Errorf("Expected union of parameterized types string %s, got %s", expected, union.String())
		}
	})

	t.Run("parameterized type with union parameter", func(t *testing.T) {
		union := NewUnionType(intType, strType)
		hashRef := NewHashRefType(union)

		expected := "HashRef[Int|Str]"
		if hashRef.String() != expected {
			t.Errorf("Expected parameterized type with union parameter string %s, got %s", expected, hashRef.String())
		}
	})

	t.Run("complex nested combinations", func(t *testing.T) {
		// ArrayRef[HashRef[Int|Str]]
		union := NewUnionType(intType, strType)
		hashRef := NewHashRefType(union)
		arrayRef := NewArrayRefType(hashRef)

		expected := "ArrayRef[HashRef[Int|Str]]"
		if arrayRef.String() != expected {
			t.Errorf("Expected complex nested type string %s, got %s", expected, arrayRef.String())
		}
	})
}

func TestTypeCompatibilityEnhancements(t *testing.T) {
	intType := NewIntType()
	strType := NewStrType()

	t.Run("enhanced basic type compatibility", func(t *testing.T) {
		// Test that we've enhanced the basic types to work with complex types
		union := NewUnionType(intType, strType)

		// Enhanced compatibility: basic types should be compatible with unions containing them
		if !intType.CompatibleWith(union) {
			t.Error("Expected Int to be compatible with Int|Str union after enhancement")
		}

		if !strType.CompatibleWith(union) {
			t.Error("Expected Str to be compatible with Int|Str union after enhancement")
		}
	})

	t.Run("parameterized type enhanced compatibility", func(t *testing.T) {
		arrayRefInt := NewArrayRefType(intType)
		union := NewUnionType(intType, strType)
		arrayRefUnion := NewArrayRefType(union)

		// ArrayRef[Int] should be compatible with ArrayRef[Int|Str]
		if !arrayRefInt.CompatibleWith(arrayRefUnion) {
			t.Error("Expected ArrayRef[Int] to be compatible with ArrayRef[Int|Str]")
		}
	})
}
