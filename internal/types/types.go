// ABOUTME: Type system for the PSC Perl static compiler type inference engine.
// ABOUTME: Defines Type as a uint32 bitset, type hierarchy masks, subtype checking, and polymorphic type satisfaction.

package types

import (
	"fmt"
	"strings"
)

// Type represents a Perl type in the PSC type system as a uint32 bitset.
// Each leaf type occupies a unique bit position. Parent/union types are
// expressed as the OR of their descendant leaf bits.
//
// The zero value (Unknown) means no type information is available.
// None (1 << 31) is the bottom type — a sentinel for unreachable branches.
type Type uint32

// Leaf type bits. Each concrete type a value can inhabit gets exactly one bit.
// Unexported leaf bits are used where the exported type name is a mask
// (a union of this leaf plus descendants).
const (
	Undef   Type = 1 << 0 // Undefined value
	Bool    Type = 1 << 1 // Boolean value
	Int     Type = 1 << 2 // Integer value
	numLeaf Type = 1 << 3 // Floating-point leaf (3.14 — Num but not Int)
	strLeaf Type = 1 << 4 // String leaf ("hello" — Str but not Num)
	DualVar Type = 1 << 5 // Dual-valued scalar (string + numeric)
	Regex   Type = 1 << 6 // Regular expression

	ScalarRef Type = 1 << 7  // Reference to a scalar
	ArrayRef  Type = 1 << 8  // Reference to an array
	HashRef   Type = 1 << 9  // Reference to a hash
	CodeRef   Type = 1 << 10 // Reference to code (subroutine)
	GlobRef   Type = 1 << 11 // Reference to a glob
	Object    Type = 1 << 12 // Blessed reference (object)

	Array Type = 1 << 13 // Array
	Hash  Type = 1 << 14 // Hash
	Code  Type = 1 << 15 // Subroutine/code
	Glob  Type = 1 << 16 // Typeglob
)

// Sentinel types.
const (
	// Unknown is the zero value — no type information is available.
	Unknown Type = 0

	// None is the bottom type — subtype of everything, produced when
	// guard narrowing yields the empty set (an unreachable branch).
	None Type = 1 << 31
)

// Parent/union type masks. Each is the bitwise OR of all descendant leaf bits.
// These are the "family" types used as type constraints and variable annotations.
//
// Num includes Int because Int <: Num (every integer is a valid number).
// Str includes Num and Int because Num <: Str (numbers stringify to strings).
// Ref is the union of all reference subtypes.
// Scalar is the full scalar family (all types that fit in a scalar variable).
// List is the aggregate family.
// Any is the top type — the union of all concrete types.
const (
	// Num is the numeric family: accepts floating-point values and integers.
	// A concrete float literal (3.14) is annotated as Num; Int is a strict subtype.
	Num Type = numLeaf | Int

	// Str is the string family: accepts string, float, and integer values.
	// A concrete string literal is annotated as Str; Num and Int are subtypes.
	Str Type = strLeaf | Num

	// Ref is the reference family mask.
	Ref Type = ScalarRef | ArrayRef | HashRef | CodeRef | GlobRef | Object

	// Scalar is the scalar family mask — all types that fit in a scalar variable.
	Scalar Type = Undef | Bool | Str | DualVar | Regex | Ref

	// List is the aggregate family mask.
	List Type = Array | Hash

	// Any is the top type — all concrete type bits.
	Any Type = Scalar | List | Code | Glob
)

// typeNames maps known Type masks/values to their canonical string names.
// String() checks this table first; arbitrary unions fall back to "A|B" format.
var typeNames = map[Type]string{
	Unknown:   "Unknown",
	Undef:     "Undef",
	Bool:      "Bool",
	Int:       "Int",
	Num:       "Num",
	Str:       "Str",
	DualVar:   "DualVar",
	Regex:     "Regex",
	ScalarRef: "ScalarRef",
	ArrayRef:  "ArrayRef",
	HashRef:   "HashRef",
	CodeRef:   "CodeRef",
	GlobRef:   "GlobRef",
	Object:    "Object",
	Array:     "Array",
	Hash:      "Hash",
	Code:      "Code",
	Glob:      "Glob",
	None:      "None",
	// Parent masks
	Ref:    "Ref",
	Scalar: "Scalar",
	List:   "List",
	Any:    "Any",
}

// allLeafBits lists all leaf type bits (both exported and internal) in ascending
// bit-position order, used for deterministic "A|B" display of arbitrary union types.
var allLeafBits = []struct {
	bit  Type
	name string
}{
	{Undef, "Undef"},
	{Bool, "Bool"},
	{Int, "Int"},
	{numLeaf, "Num"},
	{strLeaf, "Str"},
	{DualVar, "DualVar"},
	{Regex, "Regex"},
	{ScalarRef, "ScalarRef"},
	{ArrayRef, "ArrayRef"},
	{HashRef, "HashRef"},
	{CodeRef, "CodeRef"},
	{GlobRef, "GlobRef"},
	{Object, "Object"},
	{Array, "Array"},
	{Hash, "Hash"},
	{Code, "Code"},
	{Glob, "Glob"},
}

// String returns the human-readable name for the type. Known masks return
// their canonical name (e.g. "Scalar", "Ref", "Num"). Unknown arbitrary
// unions return a deterministic "A|B|C" representation in bit-position order.
func (t Type) String() string {
	// Fast path: exact match in named mask table.
	if name, ok := typeNames[t]; ok {
		return name
	}

	// None is handled by typeNames above, but guard against it to avoid
	// decomposing its bit as an unknown bit.
	if t == None {
		return "None"
	}

	// Decompose into leaf bit names for arbitrary unions.
	var parts []string
	remaining := t
	for _, leaf := range allLeafBits {
		if remaining&leaf.bit != 0 {
			parts = append(parts, leaf.name)
			remaining &^= leaf.bit
		}
	}
	if remaining != 0 {
		// Unrecognized bits.
		parts = append(parts, fmt.Sprintf("0x%x", uint32(remaining)))
	}

	if len(parts) == 0 {
		return "Unknown"
	}
	return strings.Join(parts, "|")
}

// IsSubtype reports whether child is a subtype of parent in the type lattice.
//
// With bitsets, subtyping is containment: all of child's bits must be present
// in parent's bits. None is the bottom type — subtype of everything. Unknown
// (zero) is only a subtype of Unknown itself.
//
// Rules:
//   - None is a subtype of every type (bottom type sentinel).
//   - Unknown (zero) is a subtype of Unknown only.
//   - child is a subtype of parent when all of child's bits are present in parent.
func IsSubtype(child, parent Type) bool {
	// None is the bottom type — subtype of everything.
	if child == None {
		return true
	}

	// Unknown (zero) has no bits — subtype of Unknown only.
	if child == Unknown {
		return parent == Unknown
	}

	// Strip None sentinel bit from parent before the containment check.
	parentMask := parent &^ None

	// child is a subtype of parent iff all of child's bits are within parent.
	return parentMask&child == child
}

// polymorphicMasks is the set of types for which TypeSatisfies uses the
// polymorphic (reverse-subtype) check. These are general container types
// that could hold any of their subtypes at runtime. Numeric and string
// family types (Num, Str) are intentionally excluded: a Str variable
// cannot satisfy an Int requirement because it may hold non-numeric data.
// Ref is also excluded: a generic Ref could be any reference subtype at
// runtime, but Perl code that expects a specific ref type (e.g. HashRef)
// should get a diagnostic when passed a generic Ref. This matches the
// pre-bitset behavior.
var polymorphicMasks = map[Type]bool{
	Any:    true,
	Scalar: true,
	List:   true,
}

// TypeSatisfies reports whether a value of actual type can satisfy a required type.
//
// Rules:
//   - required == Any accepts everything.
//   - actual == Unknown passes permissively (type not yet determined).
//   - IsSubtype(actual, required) covers exact and subtype relationships.
//   - For polymorphic types (Any, Scalar, List): required is a subtype of actual,
//     meaning the actual variable could hold a value of the required type at runtime.
func TypeSatisfies(actual, required Type) bool {
	// required == Any accepts everything.
	if required == Any {
		return true
	}

	// Unknown type passes permissively (type not yet determined).
	if actual == Unknown {
		return true
	}

	// Exact subtype check: all of actual's bits are within required.
	if IsSubtype(actual, required) {
		return true
	}

	// Polymorphic check: actual is a general container type that could hold
	// any of its subtypes at runtime. If required is a subtype of actual,
	// the variable might hold a value of that required type.
	if polymorphicMasks[actual] && IsSubtype(required, actual) {
		return true
	}

	return false
}
