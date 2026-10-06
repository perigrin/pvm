// ABOUTME: Tests for multi-candidate selection: arity, argument types, the most specific candidate.
// ABOUTME: Candidates are built in Go, as a .pmt's `multi sub` lines would declare them.
package types

import "testing"

// scalar is a required scalar parameter of type t, `Int $x`; Unknown is
// a parameter that states no type, `$r`.
func scalar(name string, t Type) Param {
	return Param{Name: name, Sigil: '$', Type: t, Required: true}
}

// sig is a candidate returning ret.
func sig(ret Type, params ...Param) Signature {
	return Signature{Params: params, Returns: ret}
}

// selectCandidates are RFC 0001's `select`:
//
//	multi sub select () Str;
//	multi sub select (FileHandle $fh) Str;
//	multi sub select ($r, $w, $e, Num $timeout) Int;
var selectCandidates = []Signature{
	sig(Str),
	sig(Str, scalar("fh", FileHandle)),
	sig(Int, scalar("r", Unknown), scalar("w", Unknown), scalar("e", Unknown), scalar("timeout", Num)),
}

// eachCandidates are RFC 0001's `each`, `(Hash \%h) List` and `(Array \@a)
// List`. How a `.pmt` records a backslashed parameter is the backslash
// parse's; each takes exactly one argument, so here it is a scalar
// parameter of the container's type.
var eachCandidates = []Signature{
	sig(List, scalar("h", Hash)),
	sig(List, scalar("a", Array)),
}

// wantSelected checks that args select candidate want, returning ret.
func wantSelected(t *testing.T, cands []Signature, args []Type, want int, ret Type) {
	t.Helper()
	got := Select(cands, args)
	if got.Outcome != Selected || got.Candidate != want || got.Returns != ret {
		t.Errorf("%v: got %+v, want candidate %d selected, returning %v", args, got, want, ret)
	}
}

// wantFailed checks that args select nothing and have no type.
func wantFailed(t *testing.T, cands []Signature, args []Type) {
	t.Helper()
	got := Select(cands, args)
	if got.Outcome != Failed || got.Candidate != -1 || got.Returns != Unknown {
		t.Errorf("%v: got %+v, want a failure with no candidate and no type", args, got)
	}
}

// TestSelectByArity: RFC 0001 "Multi declarations", a call picks the
// candidate whose parameters accept its number of arguments.
func TestSelectByArity(t *testing.T) {
	wantSelected(t, selectCandidates, nil, 0, Str)
	wantSelected(t, selectCandidates, []Type{GlobRef}, 1, Str)
	wantSelected(t, selectCandidates, []Type{Undef, Undef, Undef, Num}, 2, Int)
}

// TestSelectByArgumentTypes: among candidates that take the call's number
// of arguments, the one whose parameter types the arguments fit, `each %h`
// against `each @a`.
func TestSelectByArgumentTypes(t *testing.T) {
	wantSelected(t, eachCandidates, []Type{Hash}, 0, List)
	wantSelected(t, eachCandidates, []Type{Array}, 1, List)
}

// intOrNum are `multi sub f (Int $x) Str;` beside `multi sub f (Num $x)
// Int;`, whose parameters are related by subtyping: Int <: Num.
var intOrNum = []Signature{
	sig(Str, scalar("x", Int)),
	sig(Int, scalar("x", Num)),
}

// TestSelectMostSpecificCandidate: RFC 0001 "Multi declarations", the
// most specific candidate wins (perigrin, 2026-10-02). An Int fits both
// candidates, and the Int one, whose parameters are all subtypes of the
// other's, is selected; a Num fits only the Num one. Declared the other
// way round, the Int candidate still wins.
func TestSelectMostSpecificCandidate(t *testing.T) {
	wantSelected(t, intOrNum, []Type{Int}, 0, Str)
	wantSelected(t, intOrNum, []Type{Num}, 1, Int)
	numFirst := []Signature{intOrNum[1], intOrNum[0]}
	wantSelected(t, numFirst, []Type{Int}, 1, Str)
}

// TestMultiAmbiguityEdgeCases: candidates with disjoint parameter types
// share no call, so `each`'s Hash and Array are not ambiguous; identical
// candidates fit every call alike, and neither is more specific.
func TestMultiAmbiguityEdgeCases(t *testing.T) {
	if err := Ambiguity(eachCandidates); err != nil {
		t.Errorf("each: got %v, want no ambiguity", err)
	}
	if err := Ambiguity(intOrNum); err != nil {
		t.Errorf("Int beside Num: got %v, want no ambiguity", err)
	}
	dup := []Signature{intOrNum[0], intOrNum[0]}
	want := "candidates (Int $x) and (Int $x) are ambiguous for (Int)"
	if err := Ambiguity(dup); err == nil || err.Error() != want {
		t.Errorf("duplicates: got %v, want %q", err, want)
	}
}
