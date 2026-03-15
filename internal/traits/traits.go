// ABOUTME: Trait system implementation for type capabilities
// ABOUTME: Based on Perl's overload.pm core operations

package traits

import (
	"sort"
)

// Trait represents a capability of a type, based on Perl's overload.pm
type Trait struct {
	// Operation is the overload operation symbol (e.g., "+", "\"\"", "0+")
	Operation string

	// ResultType is the type that this operation produces
	ResultType string
}

// TraitSet represents a collection of traits for a type
type TraitSet struct {
	// traits maps operation symbols to their result types
	traits map[string]string
}

// NewTraitSet creates a new empty trait set
func NewTraitSet() *TraitSet {
	return &TraitSet{
		traits: make(map[string]string),
	}
}

// AddTrait adds a trait to the set
func (ts *TraitSet) AddTrait(trait Trait) {
	ts.traits[trait.Operation] = trait.ResultType
}

// RemoveTrait removes a trait from the set
func (ts *TraitSet) RemoveTrait(operation string) {
	delete(ts.traits, operation)
}

// HasTrait checks if the set contains a trait for the given operation
func (ts *TraitSet) HasTrait(operation string) bool {
	_, exists := ts.traits[operation]
	return exists
}

// GetResultType returns the result type for an operation, or empty string if not supported
func (ts *TraitSet) GetResultType(operation string) string {
	return ts.traits[operation]
}

// GetAllTraits returns all traits as a slice
func (ts *TraitSet) GetAllTraits() []Trait {
	var traits []Trait
	for op, resultType := range ts.traits {
		traits = append(traits, Trait{
			Operation:  op,
			ResultType: resultType,
		})
	}

	// Sort for deterministic output
	sort.Slice(traits, func(i, j int) bool {
		return traits[i].Operation < traits[j].Operation
	})

	return traits
}

// Clone creates a copy of the trait set
func (ts *TraitSet) Clone() *TraitSet {
	clone := NewTraitSet()
	for op, resultType := range ts.traits {
		clone.traits[op] = resultType
	}
	return clone
}

// GetCoreOperations returns the core overload.pm operations that others derive from
func GetCoreOperations() []Trait {
	return []Trait{
		// Conversion operations - fundamental to Perl's type system
		{Operation: "\"\"", ResultType: "Str"},  // string conversion
		{Operation: "0+", ResultType: "Num"},    // numeric conversion
		{Operation: "bool", ResultType: "Bool"}, // boolean conversion

		// Arithmetic operations
		{Operation: "+", ResultType: "Num"},  // addition
		{Operation: "-", ResultType: "Num"},  // subtraction
		{Operation: "*", ResultType: "Num"},  // multiplication
		{Operation: "/", ResultType: "Num"},  // division
		{Operation: "%", ResultType: "Num"},  // modulus
		{Operation: "**", ResultType: "Num"}, // exponentiation

		// Comparison operations
		{Operation: "<=>", ResultType: "Int"}, // numeric comparison
		{Operation: "cmp", ResultType: "Int"}, // string comparison

		// Bitwise operations
		{Operation: "&", ResultType: "Int"},  // bitwise and
		{Operation: "|", ResultType: "Int"},  // bitwise or
		{Operation: "^", ResultType: "Int"},  // bitwise xor
		{Operation: "~", ResultType: "Int"},  // bitwise not
		{Operation: "<<", ResultType: "Int"}, // left shift
		{Operation: ">>", ResultType: "Int"}, // right shift

		// String comparison operations
		{Operation: "eq", ResultType: "Bool"}, // string equality
		{Operation: "ne", ResultType: "Bool"}, // string inequality
		{Operation: "lt", ResultType: "Bool"}, // string less than
		{Operation: "le", ResultType: "Bool"}, // string less than or equal
		{Operation: "gt", ResultType: "Bool"}, // string greater than
		{Operation: "ge", ResultType: "Bool"}, // string greater than or equal
	}
}

// GetDefaultTraitsForType returns the default traits for basic Perl types
func GetDefaultTraitsForType(typeName string) *TraitSet {
	traits := NewTraitSet()

	switch typeName {
	case "Int":
		// Integers support numeric operations, conversions, and bitwise ops
		intTraits := []Trait{
			{Operation: "\"\"", ResultType: "Str"},
			{Operation: "0+", ResultType: "Num"},
			{Operation: "bool", ResultType: "Bool"},
			{Operation: "+", ResultType: "Num"},
			{Operation: "-", ResultType: "Num"},
			{Operation: "*", ResultType: "Num"},
			{Operation: "/", ResultType: "Num"},
			{Operation: "%", ResultType: "Num"},
			{Operation: "**", ResultType: "Num"},
			{Operation: "<=>", ResultType: "Int"},
			{Operation: "&", ResultType: "Int"},
			{Operation: "|", ResultType: "Int"},
			{Operation: "^", ResultType: "Int"},
			{Operation: "~", ResultType: "Int"},
			{Operation: "<<", ResultType: "Int"},
			{Operation: ">>", ResultType: "Int"},
		}
		for _, trait := range intTraits {
			traits.AddTrait(trait)
		}

	case "Str":
		// Strings support conversions and string operations
		strTraits := []Trait{
			{Operation: "\"\"", ResultType: "Str"},
			{Operation: "0+", ResultType: "Num"},
			{Operation: "bool", ResultType: "Bool"},
			{Operation: "cmp", ResultType: "Int"},
			{Operation: "eq", ResultType: "Bool"},
			{Operation: "ne", ResultType: "Bool"},
			{Operation: "lt", ResultType: "Bool"},
			{Operation: "le", ResultType: "Bool"},
			{Operation: "gt", ResultType: "Bool"},
			{Operation: "ge", ResultType: "Bool"},
		}
		for _, trait := range strTraits {
			traits.AddTrait(trait)
		}

	case "Num":
		// Numbers support arithmetic and conversions
		numTraits := []Trait{
			{Operation: "\"\"", ResultType: "Str"},
			{Operation: "0+", ResultType: "Num"},
			{Operation: "bool", ResultType: "Bool"},
			{Operation: "+", ResultType: "Num"},
			{Operation: "-", ResultType: "Num"},
			{Operation: "*", ResultType: "Num"},
			{Operation: "/", ResultType: "Num"},
			{Operation: "%", ResultType: "Num"},
			{Operation: "**", ResultType: "Num"},
			{Operation: "<=>", ResultType: "Int"},
		}
		for _, trait := range numTraits {
			traits.AddTrait(trait)
		}

	case "Bool":
		// Booleans support conversions and logical operations
		boolTraits := []Trait{
			{Operation: "\"\"", ResultType: "Str"},
			{Operation: "0+", ResultType: "Num"},
			{Operation: "bool", ResultType: "Bool"},
			{Operation: "&", ResultType: "Int"},
			{Operation: "|", ResultType: "Int"},
			{Operation: "^", ResultType: "Int"},
			{Operation: "~", ResultType: "Int"},
		}
		for _, trait := range boolTraits {
			traits.AddTrait(trait)
		}

	case "ArrayRef":
		// Array references support basic conversions only
		arrayTraits := []Trait{
			{Operation: "\"\"", ResultType: "Str"},
			{Operation: "0+", ResultType: "Num"},
			{Operation: "bool", ResultType: "Bool"},
		}
		for _, trait := range arrayTraits {
			traits.AddTrait(trait)
		}

	case "HashRef":
		// Hash references support basic conversions only
		hashTraits := []Trait{
			{Operation: "\"\"", ResultType: "Str"},
			{Operation: "0+", ResultType: "Num"},
			{Operation: "bool", ResultType: "Bool"},
		}
		for _, trait := range hashTraits {
			traits.AddTrait(trait)
		}

	default:
		// Unknown types get basic conversion traits only
		basicTraits := []Trait{
			{Operation: "\"\"", ResultType: "Str"},
			{Operation: "0+", ResultType: "Num"},
			{Operation: "bool", ResultType: "Bool"},
		}
		for _, trait := range basicTraits {
			traits.AddTrait(trait)
		}
	}

	return traits
}
