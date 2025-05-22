// ABOUTME: Benchmarks for union type performance testing
// ABOUTME: Validates performance characteristics of union type operations

package typedef

import (
	"fmt"
	"testing"
)

// BenchmarkUnionTypeCreation benchmarks union type creation
func BenchmarkUnionTypeCreation(b *testing.B) {
	members := []string{"Int", "Str", "Bool"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewUnionType(members)
	}
}

// BenchmarkUnionTypeTraitComputation benchmarks trait intersection computation
func BenchmarkUnionTypeTraitComputation(b *testing.B) {
	unionType := NewUnionType([]string{"Int", "Str", "Bool", "Num"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		unionType.ClearTraitCache() // Force recomputation
		_ = unionType.GetTraits()
	}
}

// BenchmarkUnionTypeTraitCaching benchmarks cached trait access
func BenchmarkUnionTypeTraitCaching(b *testing.B) {
	unionType := NewUnionType([]string{"Int", "Str", "Bool", "Num"})
	_ = unionType.GetTraits() // Populate cache

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = unionType.GetTraits()
	}
}

// BenchmarkUnionTypeOperationCheck benchmarks operation support checking
func BenchmarkUnionTypeOperationCheck(b *testing.B) {
	unionType := NewUnionType([]string{"Int", "Str", "Bool"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = unionType.SupportsOperation("\"\"")
	}
}

// BenchmarkUnionTypeParsing benchmarks union type parsing
func BenchmarkUnionTypeParsing(b *testing.B) {
	storage, err := NewStorageWithPath(b.TempDir())
	if err != nil {
		b.Fatal(err)
	}
	hierarchy := NewTypeHierarchy(storage)

	testCases := []string{
		"Int|Str",
		"Union[Int, Str]",
		"Int|Str|Bool|Num",
		"ArrayRef[Int]|HashRef[Str]",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, testCase := range testCases {
			_ = hierarchy.ParseUnionType(testCase)
		}
	}
}

// BenchmarkLargeUnionOperations benchmarks operations on large unions
func BenchmarkLargeUnionOperations(b *testing.B) {
	// Create a large union
	members := make([]string, 50)
	for i := 0; i < 50; i++ {
		members[i] = fmt.Sprintf("Type%d", i)
	}

	unionType := NewUnionType(members)

	b.Run("TraitComputation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			unionType.ClearTraitCache()
			_ = unionType.GetTraits()
		}
	})

	b.Run("OperationCheck", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = unionType.SupportsOperation("\"\"")
		}
	})

	b.Run("MemberContainment", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = unionType.ContainsMember("Type25")
		}
	})
}

// BenchmarkUnionTypeString benchmarks string representation generation
func BenchmarkUnionTypeString(b *testing.B) {
	unionType := NewUnionType([]string{"Int", "Str", "Bool", "Num", "ArrayRef", "HashRef"})

	b.Run("String", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = unionType.String()
		}
	})

	b.Run("TypeName", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = unionType.TypeName()
		}
	})
}

// BenchmarkUnionTypeEquality benchmarks equality checking
func BenchmarkUnionTypeEquality(b *testing.B) {
	union1 := NewUnionType([]string{"Int", "Str", "Bool", "Num"})
	union2 := NewUnionType([]string{"Num", "Bool", "Str", "Int"})      // Same members, different order
	union3 := NewUnionType([]string{"Int", "Str", "Bool", "ArrayRef"}) // Different members

	b.Run("EqualUnions", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = union1.Equals(union2)
		}
	})

	b.Run("DifferentUnions", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = union1.Equals(union3)
		}
	})
}

// BenchmarkUnionTypeCompatibility benchmarks compatibility checking
func BenchmarkUnionTypeCompatibility(b *testing.B) {
	storage, err := NewStorageWithPath(b.TempDir())
	if err != nil {
		b.Fatal(err)
	}
	hierarchy := NewTypeHierarchy(storage)

	unionType := hierarchy.CreateUnionType([]string{"Int", "Str", "Bool"})

	b.Run("IsCompatibleWith", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = unionType.IsCompatibleWith("Scalar", hierarchy)
		}
	})

	b.Run("CanAssignFrom", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = unionType.CanAssignFrom("Int", hierarchy)
		}
	})
}
