// ABOUTME: Type system for the PSC Perl static compiler type inference engine.
// ABOUTME: Defines Type enum, type hierarchy, subtype checking, and polymorphic type satisfaction.

package types

// Type represents a Perl type in the PSC type system.
type Type int

// Type constants in hierarchy order. The zero value is Unknown (uninitialized sentinel).
const (
	Unknown   Type = iota // Zero value sentinel for uninitialized/unknown types
	Any                   // Top type — accepts any value
	Scalar                // Any scalar value
	Undef                 // Undefined value
	Bool                  // Boolean value
	Str                   // String value
	Num                   // Numeric value
	Int                   // Integer value
	DualVar               // Dual-valued scalar (string + numeric)
	Regex                 // Regular expression
	Ref                   // Any reference
	ScalarRef             // Reference to a scalar
	ArrayRef              // Reference to an array
	HashRef               // Reference to a hash
	CodeRef               // Reference to code (subroutine)
	GlobRef               // Reference to a glob
	Object                // Blessed reference (object)
	List                  // Any list
	Array                 // Array
	Hash                  // Hash
	Code                  // Subroutine/code
	Glob                  // Typeglob
	None                  // Bottom type — subtype of everything, no concrete value
)

// typeNames maps Type values to their human-readable string representations.
var typeNames = map[Type]string{
	Unknown:   "Unknown",
	Any:       "Any",
	Scalar:    "Scalar",
	Undef:     "Undef",
	Bool:      "Bool",
	Str:       "Str",
	Num:       "Num",
	Int:       "Int",
	DualVar:   "DualVar",
	Regex:     "Regex",
	Ref:       "Ref",
	ScalarRef: "ScalarRef",
	ArrayRef:  "ArrayRef",
	HashRef:   "HashRef",
	CodeRef:   "CodeRef",
	GlobRef:   "GlobRef",
	Object:    "Object",
	List:      "List",
	Array:     "Array",
	Hash:      "Hash",
	Code:      "Code",
	Glob:      "Glob",
	None:      "None",
}

// String returns the human-readable name for the type.
func (t Type) String() string {
	if name, ok := typeNames[t]; ok {
		return name
	}
	return "Unknown"
}

// parentMap records the direct parent of each type in the type hierarchy.
// Types not present (Unknown, Any, None) have no parent in the DAG.
var parentMap = map[Type]Type{
	Scalar:    Any,
	Undef:     Scalar,
	Bool:      Scalar,
	Str:       Scalar,
	Num:       Str,
	Int:       Num,
	DualVar:   Scalar,
	Regex:     Scalar,
	Ref:       Scalar,
	ScalarRef: Ref,
	ArrayRef:  Ref,
	HashRef:   Ref,
	CodeRef:   Ref,
	GlobRef:   Ref,
	Object:    Ref,
	List:      Any,
	Array:     List,
	Hash:      List,
	Code:      Any,
	Glob:      Any,
}

// polymorphicTypes is the set of types that are polymorphic: a variable of one
// of these types could hold any value of a subtype at runtime.
var polymorphicTypes = map[Type]bool{
	Any:    true,
	Scalar: true,
	List:   true,
}

// IsSubtype reports whether child is a subtype of parent in the type hierarchy.
//
// Rules:
//   - A type is a subtype of itself (identity).
//   - None is a subtype of every type (bottom type).
//   - Subtyping is transitive: walking the parent chain from child to parent.
func IsSubtype(child, parent Type) bool {
	// Identity check
	if child == parent {
		return true
	}

	// None is the bottom type — subtype of everything
	if child == None {
		return true
	}

	// Walk the parent chain from child upward
	current := child
	for {
		p, ok := parentMap[current]
		if !ok {
			// Reached the top without finding parent
			return false
		}
		if p == parent {
			return true
		}
		current = p
	}
}

// TypeSatisfies reports whether a value of actual type can satisfy a required type.
//
// Rules:
//   - required == Any accepts everything.
//   - actual == Unknown passes permissively (type not yet determined).
//   - IsSubtype(actual, required) covers normal subtype relationships.
//   - For polymorphic types (Any, Scalar, List): also accepts when required is a
//     subtype of actual, because a polymorphic variable could hold the required type
//     at runtime.
func TypeSatisfies(actual, required Type) bool {
	// required == Any accepts everything
	if required == Any {
		return true
	}

	// Unknown type passes permissively (type not yet determined)
	if actual == Unknown {
		return true
	}

	// Normal subtype check
	if IsSubtype(actual, required) {
		return true
	}

	// Polymorphic types: actual could hold a value of any subtype
	if polymorphicTypes[actual] && IsSubtype(required, actual) {
		return true
	}

	return false
}
