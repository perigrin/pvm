// ABOUTME: Context-based and guard-based type narrowing for the PSC type inference engine.
// ABOUTME: Implements NarrowByContext and NarrowByGuard following Chalk's narrow_type semantics.

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
//   - ScalarCtx: Array or Hash → Int (element/bucket count). List → Scalar. Everything else unchanged.
//   - ListCtx: Unchanged (pass through).
//   - VoidCtx: Returns (Unknown, false) — type is discarded.
//   - UnknownCtx: Unchanged (pass through).
func NarrowByContext(typ Type, ctx Context) (Type, bool) {
	switch ctx {
	case VoidCtx:
		return Unknown, false

	case ScalarCtx:
		switch typ {
		case Array, Hash:
			return Int, true
		case List:
			return Scalar, true
		default:
			return typ, true
		}

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

// NarrowByGuard returns the type that typ narrows to when a guard expression
// is known to be true. The second return value is true when narrowing occurred,
// false when the type is already more specific than the guard can determine.
//
// Rules:
//   - GuardDefined: If typ is Scalar, Undef, or Any (could be undef), narrows to Scalar.
//     Otherwise the type is already non-undef; returns (typ, false).
//   - GuardRef (plain): Narrows to Ref.
//   - GuardIsa: Narrows to Object.
//   - GuardRef with RefType: Maps the ref-type string to a specific Type:
//     "HASH"→HashRef, "ARRAY"→ArrayRef, "SCALAR"→ScalarRef, "CODE"→CodeRef,
//     "GLOB"→GlobRef, "REF"→Ref, anything else→Object (blessed reference).
func NarrowByGuard(typ Type, guard GuardPattern) (Type, bool) {
	switch guard.Kind {
	case GuardDefined:
		// Types that could hold an undef value at runtime
		if typ == Scalar || typ == Undef || typ == Any {
			return Scalar, true
		}
		// All other types are already known non-undef
		return typ, false

	case GuardRef:
		if guard.RefType != "" {
			// ref($x) eq 'TYPE' — narrow to specific reference subtype
			if specific, ok := refTypeMap[guard.RefType]; ok {
				return specific, true
			}
			// Unknown ref type string means a blessed reference (class name)
			return Object, true
		}
		// Plain ref($x) — narrows to generic Ref
		return Ref, true

	case GuardIsa:
		return Object, true

	default:
		return typ, false
	}
}
