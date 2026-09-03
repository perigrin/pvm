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
	Undef   Type = 1 << 0  // Undefined value
	Bool    Type = 1 << 1  // Boolean value
	Int     Type = 1 << 2  // Integer value
	numLeaf Type = 1 << 3  // Floating-point leaf (3.14 — Num but not Int)
	strLeaf Type = 1 << 4  // String leaf ("hello" — Str but not Num)
	DualVar Type = 1 << 5  // Dual-valued scalar (string + numeric)
	NaN     Type = 1 << 17 // IEEE 754 Not-a-Number (Scalar, not Str or Num)
	Inf     Type = 1 << 18 // IEEE 754 Infinity (Scalar, not Str or Num)
	Regex   Type = 1 << 6  // Regular expression

	ScalarRef  Type = 1 << 7  // Reference to a scalar
	ArrayRef   Type = 1 << 8  // Reference to an array
	HashRef    Type = 1 << 9  // Reference to a hash
	CodeRef    Type = 1 << 10 // Reference to code (subroutine)
	GlobRef    Type = 1 << 11 // Reference to a glob
	objectLeaf Type = 1 << 12 // Blessed-reference leaf (blessed, but not a Regexp)

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
	//
	// NaN and Inf are members of Str, not Num. Both pass syntactic
	// preservation ("NaN" -> NaN -> "NaN" round-trips) and satisfy the string
	// operation contracts; they are excluded from Num by the semantic
	// component alone — NaN violates Contract_== and Contract_-, Inf violates
	// Contract_- (Inf - Inf = NaN).
	Str Type = strLeaf | Num | NaN | Inf

	// Object is the blessed-reference family mask. Regex is a subtype of
	// Object: a compiled pattern is blessed into Regexp and participates in
	// method dispatch.
	Object Type = objectLeaf | Regex

	// Ref is the reference family mask.
	Ref Type = ScalarRef | ArrayRef | HashRef | CodeRef | GlobRef | Object

	// Scalar is the scalar family mask — all types that fit in a scalar variable.
	Scalar Type = Undef | Bool | Str | DualVar | Ref

	// List is the aggregate family mask. Arity orders the top of the lattice:
	// Scalar denotes {1} and List denotes {0,1,2,...}, so Scalar <: List by
	// subset inclusion on arities. This states Perl's list-flattening rule as
	// a subtype fact — a scalar satisfies a list position because one value is
	// one of the arities a list admits.
	List Type = Array | Hash | Scalar

	// Any is the top type — all concrete type bits.
	Any Type = List | Code | Glob
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
	NaN:       "NaN",
	Inf:       "Inf",
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
	{objectLeaf, "Object"},
	{Array, "Array"},
	{Hash, "Hash"},
	{Code, "Code"},
	{Glob, "Glob"},
	{NaN, "NaN"},
	{Inf, "Inf"},
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
//
// List is handled separately by listSatisfiesAggregate rather than listed
// here: it stands above Scalar in the arity ordering rather than beside it,
// so a blanket polymorphic check would let every scalar requirement be
// satisfied by a list-returning expression.
var polymorphicMasks = map[Type]bool{
	Any:    true,
	Scalar: true,
}

// listSatisfiesAggregate reports whether a List actual can satisfy required.
//
// A list-producing expression may stand where an aggregate is expected —
// sort() returns List and can feed an Array position. It may NOT stand where
// a scalar is expected: under the arity ordering Scalar <: List, so a blanket
// reverse-subtype check would make `keys(%h) + 1` type-check as arithmetic.
// Arity is a claim about how many values arrive, not about which scalar type
// one of them has, so only aggregate requirements are satisfied this way.
func listSatisfiesAggregate(actual, required Type) bool {
	if actual != List {
		return false
	}
	return required&^(Array|Hash) == 0 && required != Unknown
}

// TypeSatisfies reports whether a value of actual type can satisfy a required
// type at a use site — a call argument, an operator operand.
//
// It asks two questions:
//
//  1. Is actual already a member of required? (IsSubtype — nothing happens at
//     runtime and the value survives unchanged.)
//  2. Could a value of the required type be hiding inside actual? (The union
//     and polymorphic cases below — a variable whose inferred type is a broad
//     family may hold a value of the required type.)
//
// It deliberately does NOT ask whether Perl would coerce. Coercibility is far
// too permissive to gate a diagnostic on: Perl stringifies and numifies
// everything, so `IsCoercible(Undef, Num)` and `IsCoercible(NaN, Int)` are
// both true, and accepting them here silently deletes exactly the diagnostics
// PSC exists to emit — `chr($n)` on a float, `undef` in arithmetic, NaN where
// an integer is wanted. Measured: routing this function through IsCoercible
// turns four such diagnostics off.
//
// So the two relations serve opposite ends. IsCoercible answers "will this
// run?", which is the question a code GENERATOR asks. TypeSatisfies answers
// "does this mean what it says?", which is the question a CHECKER asks. The
// gap between them — coercible but not a subtype — is where a
// coercion-mismatch diagnostic belongs, and CoercionMismatch below names it.
func TypeSatisfies(actual, required Type) bool {
	// required == Any accepts everything.
	if required == Any {
		return true
	}

	// Unknown type passes permissively (type not yet determined).
	if actual == Unknown {
		return true
	}

	// (1) Membership: all of actual's bits are within required.
	if IsSubtype(actual, required) {
		return true
	}

	// (2) The value might be of the required type at runtime.
	//
	// Union containment: actual is an ad-hoc union (not a named parent mask)
	// whose bits include all of required's. Object|HashRef satisfies Object
	// because the union carries the Object bit. Named masks are excluded
	// because Str contains Int's bits while a Str should not satisfy an Int
	// requirement — it may hold non-numeric data.
	if _, isNamed := typeNames[actual]; !isNamed && actual&required == required {
		return true
	}

	// Polymorphic: actual is a general container that could hold any of its
	// subtypes at runtime, so a required subtype might be what it holds.
	if polymorphicMasks[actual] && IsSubtype(required, actual) {
		return true
	}

	// A List can stand where an aggregate is expected, but not where a scalar
	// is — see listSatisfiesAggregate.
	if listSatisfiesAggregate(actual, required) {
		return true
	}

	return false
}
