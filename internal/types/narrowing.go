// ABOUTME: Context-based and guard-based type narrowing for the PSC type inference engine.
// ABOUTME: Implements NarrowByContext and NarrowByGuard using bitset operations on the uint32 Type.

package types

// Context represents the evaluation context in which a Perl expression appears.
// Context determines how a value is coerced or discarded.
type Context int

const (
	UnknownCtx Context = iota // Context is not yet determined
	ScalarCtx                 // Expression is used in scalar context
	ListCtx                   // Expression is used in list context
	VoidCtx                   // Expression result is discarded (void context)
)

// contextNames maps Context values to their human-readable string representations.
var contextNames = map[Context]string{
	UnknownCtx: "Unknown",
	ScalarCtx:  "Scalar",
	ListCtx:    "List",
	VoidCtx:    "Void",
}

// String returns the human-readable name for the context.
func (c Context) String() string {
	if name, ok := contextNames[c]; ok {
		return name
	}
	return "Unknown"
}

// NarrowByContext returns the type a value of typ would have when evaluated in ctx.
// The second return value is false when the type is discarded (void context), true otherwise.
//
// Rules (matching Chalk's narrow_type):
//   - ScalarCtx: Array or Hash bits become Int (element/bucket count). Other bits pass through.
//     For union types the narrowing is applied per-bit: each Array or Hash bit in the mask
//     is replaced by an Int bit, all other bits pass through unchanged.
//   - ListCtx: Unchanged (pass through).
//   - VoidCtx: Returns (Unknown, false) — type is discarded.
//   - UnknownCtx: Unchanged (pass through).
func NarrowByContext(typ Type, ctx Context) (Type, bool) {
	switch ctx {
	case VoidCtx:
		return Unknown, false

	case ScalarCtx:
		// For union types we apply scalar context per-bit: Array and Hash bits
		// become Int; all other bits pass through unchanged.
		if typ&(Array|Hash) == 0 {
			// No Array or Hash bits — pass through unchanged.
			// Special case: List = Array|Hash exactly, but if typ == List the
			// check above passes since List contains both bits. Re-check below.
			return typ, true
		}
		// Remove Array and Hash bits, add Int for each that was present.
		result := typ &^ (Array | Hash)
		result |= Int
		return result, true

	case ListCtx, UnknownCtx:
		return typ, true

	default:
		return typ, true
	}
}

// GuardKind identifies the category of a runtime type guard expression.
type GuardKind int

const (
	GuardDefined GuardKind = iota // defined($x) — tests that value is not undef
	GuardRef                      // ref($x) — tests that value is a reference
	GuardIsa                      // $x isa Foo — tests that value is an instance of a class
)

// GuardPattern describes a guard expression used to narrow a type at a branch point.
// For a plain ref() check, RefType is empty. For ref($x) eq 'HASH' style checks,
// RefType holds the right-hand string literal (e.g. "HASH", "ARRAY", "MyClass").
type GuardPattern struct {
	Kind    GuardKind
	RefType string // Non-empty for ref($x) eq 'TYPE' comparisons
}

// refTypeMap maps ref() string return values to their corresponding Type constants.
// Keys are the strings Perl's ref() builtin returns for core reference types.
var refTypeMap = map[string]Type{
	"HASH":   HashRef,
	"ARRAY":  ArrayRef,
	"SCALAR": ScalarRef,
	"CODE":   CodeRef,
	"GLOB":   GlobRef,
	"REF":    Ref,
}

// narrowResult encodes the three possible outcomes of a bitset narrowing operation.
// When result == 0 (empty set), the branch is unreachable.
// When result == typ, the type is already more specific than the guard can determine.
// Otherwise, narrowing occurred.
func narrowResult(result, typ Type) (Type, bool) {
	if result == 0 {
		// Empty set — the branch is unreachable.
		return None, true
	}
	if result == typ {
		// No change — already as specific as this guard can determine.
		return typ, false
	}
	return result, true
}

// NarrowByGuard returns the type that typ narrows to when a guard expression
// is known to be true. The second return value is true when narrowing occurred,
// false when the type is already more specific than the guard can determine.
//
// Rules (bitset operations):
//   - GuardDefined: result = typ &^ Undef (remove the Undef bit).
//     Empty result → (None, true) — unreachable branch.
//     Result == typ → (typ, false) — type already has no Undef bit.
//   - GuardRef plain: treat Unknown as Any; result = typ & Ref (intersection with Ref mask).
//     Empty → (None, true). Result == typ → (typ, false).
//   - GuardRef with RefType: look up specific type in refTypeMap; result = typ & specificType.
//     For an unknown class name the specific type is Object.
//   - GuardIsa: result = typ & Object.
func NarrowByGuard(typ Type, guard GuardPattern) (Type, bool) {
	switch guard.Kind {
	case GuardDefined:
		effective := typ
		if effective == Unknown {
			// Unknown means no type information — treat as Any (top type).
			effective = Any
		}
		result := effective &^ Undef
		return narrowResult(result, effective)

	case GuardRef:
		effective := typ
		if effective == Unknown {
			effective = Any
		}
		if guard.RefType != "" {
			// ref($x) eq 'TYPE' — narrow to specific reference subtype.
			specificType, ok := refTypeMap[guard.RefType]
			if !ok {
				// Unknown ref type string means a blessed reference (class name).
				specificType = Object
			}
			result := effective & specificType
			return narrowResult(result, effective)
		}
		// Plain ref($x) — intersection with the full Ref mask.
		result := effective & Ref
		return narrowResult(result, effective)

	case GuardIsa:
		effective := typ
		if effective == Unknown {
			effective = Any
		}
		result := effective & Object
		return narrowResult(result, effective)

	default:
		return typ, false
	}
}

// NegateGuard returns the type that typ narrows to when a guard expression is
// known to be FALSE. The second return value is true when useful narrowing
// occurred.
//
// Rules (bitset operations):
//   - GuardDefined: result = typ & Undef (keep only the Undef bit).
//     Empty result → (None, true) — type has no Undef bit, negated branch unreachable.
//     Result == typ → (typ, false) — type is already pure Undef, no change.
//   - GuardRef plain: result = typ &^ Ref (remove all Ref bits).
//     Empty → (None, true). Result == typ → (typ, false).
//   - GuardRef with RefType: return (typ, false) — "not a HashRef" could be anything.
//   - GuardIsa: return (typ, false) — "not a Foo" could be anything.
func NegateGuard(typ Type, guard GuardPattern) (Type, bool) {
	switch guard.Kind {
	case GuardDefined:
		result := typ & Undef
		return narrowResult(result, typ)

	case GuardRef:
		if guard.RefType != "" {
			// "not ref eq TYPE" — not useful, anything could be not-that-type.
			return typ, false
		}
		result := typ &^ Ref
		return narrowResult(result, typ)

	case GuardIsa:
		// "not isa Foo" — not useful, anything could be not-that-class.
		return typ, false

	default:
		return typ, false
	}
}
