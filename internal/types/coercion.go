// ABOUTME: The coercion relation — what Perl will convert to what, per the paper's Coercion Rules.
// ABOUTME: Kept separate from subtyping: coercion is lossy and directional, subtyping is neither.

package types

// Coercibility and subtyping answer different questions, and a checker that
// conflates them is either silent or wrong.
//
//	IsSubtype(A, B)   — a value of A ALREADY IS a member of B. Nothing happens
//	                    at runtime, and the value survives unchanged.
//	IsCoercible(A, B) — Perl will CONVERT an A into a B, possibly destroying
//	                    the value in the process.
//
// Every reference stringifies to "HASH(0x55f1...)", so HashRef is coercible to
// Str; it is not a subtype of Str, because that string is metadata and cannot
// be turned back into the reference. Likewise "hello" numifies to 0, so Str is
// coercible to Num while emphatically not a subtype of it.
//
// Subtyping implies coercibility — if a value already is a B, the conversion is
// the identity — so IsCoercible answers true for every subtype pair without
// needing an entry here.
//
// The rules below are transcribed from the paper's "Coercion Rules" section
// (2-perl-types-formal.md, "## Implementation Notes"), which enumerates the
// coercions to Boolean, Str, Num and Int. Where the paper describes a coercion
// this table names it; where it does not, the absence is deliberate rather
// than an oversight, and coerciblePairs records why.

// stringifiable is every type Perl will stringify. The paper's "To String"
// rules cover numbers, references, undef, objects and booleans — which is to
// say all of Scalar. Aggregates are excluded: `"@a"` interpolates and
// `scalar(@a)` counts, but neither is a value-preserving conversion of the
// array itself, and PSC models context in NarrowByContext.
const stringifiable = Scalar

// numifiable is every type Perl will numify. Same domain as stringification:
// the paper's "To Number" rules take numeric strings, non-numeric strings
// (0, with a warning), references (the address), undef (0), objects and
// booleans.
const numifiable = Scalar

// booleanisable is every type with a truth value, which in Perl is all of
// them. The paper's "To Boolean" rules map 0, ” and undef to false and every
// other value to true.
const booleanisable = Scalar

// coercionTargets lists, for each coercion target, the domain that converts to
// it. A target absent from this map is one nothing coerces INTO: Perl will not
// fabricate a reference from a string, so no entry exists for HashRef and its
// siblings. (Symbolic references are a separate mechanism and are off under
// `use strict`, which PSC assumes.)
var coercionTargets = map[Type]Type{
	Str:  stringifiable,
	Num:  numifiable,
	Int:  numifiable, // "To Integer": as Num, then truncated toward zero
	Bool: booleanisable,
}

// IsCoercible reports whether Perl will convert a value of type from into a
// value of type to.
//
// The conversion may be lossy — that is precisely what distinguishes this from
// IsSubtype, which reports membership rather than convertibility. A diagnostic
// that fires on `!IsCoercible` is reporting code that cannot run; one that
// fires on `IsCoercible && !IsSubtype` is reporting code that runs but
// probably does not mean what it says.
//
// Unknown never participates: it is the statement that inference has not
// determined a type, not a claim about a value, so it can neither be converted
// nor be a conversion's target.
func IsCoercible(from, to Type) bool {
	if from == Unknown || to == Unknown {
		return false
	}

	// None is the empty set — vacuously coercible to anything, matching its
	// role as the bottom type in IsSubtype.
	if from == None {
		return true
	}

	// Subtyping implies coercibility: the identity coercion.
	if IsSubtype(from, to) {
		return true
	}

	// Otherwise the target must name a coercion whose domain includes from.
	domain, ok := coercionTargets[to]
	if !ok {
		return false
	}
	return IsSubtype(from, domain)
}

// CoercionMismatch reports whether a use of actual where required is wanted
// will RUN but probably does not mean what it says.
//
// That is exactly the gap between the two relations: Perl will perform the
// conversion (IsCoercible) but the value is not a member of the required type
// (TypeSatisfies), so something is lost on the way. `chr(3.7)` runs and gives
// the character for 3; `undef + 1` runs and gives 1. Both are legal Perl and
// both are usually a bug, which is why the diagnostic is a warning about
// meaning rather than an error about legality.
//
// A use that is neither satisfying nor coercible is a harder problem — the
// code cannot do anything sensible at all — and is not a coercion mismatch.
func CoercionMismatch(actual, required Type) bool {
	if actual == Unknown || required == Unknown {
		return false
	}
	return !TypeSatisfies(actual, required) && IsCoercible(actual, required)
}
