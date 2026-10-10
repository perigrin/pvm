// ABOUTME: Tests for the Void type: arity 0, under List beside Scalar, coerced to Undef in scalar context.
// ABOUTME: Void returns and yields nothing; None does not return. The two are kept apart.
package types

import "testing"

// TestVoidInLattice: the paper's "Arity orders the top of the lattice".
// Void {0} and Scalar {1} are subsets of List {0,1,2,...} and disjoint
// from each other, so Void <: List and Void and Scalar are incomparable.
//
// The paper's List is exactly Void ⊔ Scalar. pvm's List also holds Array
// and Hash as leaves of their own (the paper places them under List
// too, but as a sequence's shapes rather than beside its arities), so
// here the join of Void and Scalar is the union Void|Scalar, a proper
// subtype of List, and List is that union with Array and Hash. pvm has
// no meet function; in the bitset the meet is the intersection, and
// Void and Scalar share no leaf, so Void ⊓ Scalar is the empty set,
// None.
func TestVoidInLattice(t *testing.T) {
	if !IsSubtype(Void, List) {
		t.Errorf("Void is not a subtype of List")
	}
	if IsSubtype(Void, Scalar) || IsSubtype(Scalar, Void) {
		t.Errorf("Void and Scalar are comparable")
	}
	join := Join(Void, Scalar)
	if join != Void|Scalar || !IsSubtype(join, List) {
		t.Errorf("Join(Void, Scalar) = %v, want Void|Scalar under List", join)
	}
	if join|Array|Hash != List {
		t.Errorf("List = %v, want Void|Scalar|Array|Hash", List)
	}
	if Void&Scalar != 0 {
		t.Errorf("Void and Scalar share leaves %v; their meet is not None", Void&Scalar)
	}
	if Void&(Code|Glob|IO|None) != 0 || !IsSubtype(Void, Any) {
		t.Errorf("Void overlaps another top-level type or is not under Any: %032b", Void)
	}
	if got, err := FromName("Void"); err != nil || got != Void || Void.String() != "Void" {
		t.Errorf("FromName(\"Void\") = %v, %v; Void prints %q", got, err, Void.String())
	}
}

// TestVoidCoercesToUndefInScalarContext: the paper's `Void ⇓^Scalar undef`.
// Measured on 5.42, `sub f(+) {...} f(())` passes one argument, undef:
// the empty list in a scalar slot arrives as undef. A union coerces per
// leaf. A None result, from `die`, has no value to coerce: scalar context
// leaves it None, and None is still vacuously accepted everywhere.
func TestVoidCoercesToUndefInScalarContext(t *testing.T) {
	if got, ok := NarrowByContext(Void, ScalarCtx); got != Undef || !ok {
		t.Errorf("NarrowByContext(Void, ScalarCtx) = %v, %v; want Undef", got, ok)
	}
	if got, _ := NarrowByContext(Void|Int, ScalarCtx); got != Undef|Int {
		t.Errorf("NarrowByContext(Void|Int, ScalarCtx) = %v, want Undef|Int", got)
	}
	if got, _ := NarrowByContext(Void, ListCtx); got != Void {
		t.Errorf("NarrowByContext(Void, ListCtx) = %v, want Void", got)
	}
	if got, ok := NarrowByContext(None, ScalarCtx); got != None || !ok {
		t.Errorf("NarrowByContext(None, ScalarCtx) = %v, %v; want None", got, ok)
	}
	for _, to := range []Type{Undef, Int, Str, Void, Scalar} {
		if !IsCoercible(None, to) || !TypeSatisfies(None, to) {
			t.Errorf("None is not vacuously accepted as %v", to)
		}
	}
}

// TestVoidIsNotNone: Void returns and yields nothing, `{()}`; None does
// not return, `∅`. None is the bottom, so None <: Void, and not the
// other way.
func TestVoidIsNotNone(t *testing.T) {
	if Void == None {
		t.Fatalf("Void is None")
	}
	if !IsSubtype(None, Void) {
		t.Errorf("None is not a subtype of Void")
	}
	if IsSubtype(Void, None) {
		t.Errorf("Void is a subtype of None")
	}
	if Join(Void, None) != Void {
		t.Errorf("Join(Void, None) = %v, want Void", Join(Void, None))
	}
}
