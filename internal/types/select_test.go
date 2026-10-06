// ABOUTME: Tests for multi-candidate selection: arity, argument types, the most specific candidate.
// ABOUTME: Candidates are built in Go, as a .pmt's `multi sub` lines would declare them.
package types

import (
	"slices"
	"testing"
)

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

// wantSelected checks that args select candidate want, returning ret, in
// scalar context: a candidate stating no `:context` answers for every one.
func wantSelected(t *testing.T, cands []Signature, args []Type, want int, ret Type) {
	t.Helper()
	wantSelectedIn(t, cands, ScalarCtx, args, want, ret)
}

// wantSelectedIn checks that args in context ctx select candidate want,
// returning ret.
func wantSelectedIn(t *testing.T, cands []Signature, ctx Context, args []Type, want int, ret Type) {
	t.Helper()
	got := Select(cands, args, ctx)
	if got.Outcome != Selected || got.Candidate != want || got.Returns != ret {
		t.Errorf("%v in %v: got %+v, want candidate %d selected, returning %v", args, ctx, got, want, ret)
	}
}

// wantFailed checks that args select nothing and have no type, in scalar
// context.
func wantFailed(t *testing.T, cands []Signature, args []Type) {
	t.Helper()
	wantFailedIn(t, cands, ScalarCtx, args)
}

// wantFailedIn checks that args in context ctx select nothing and have no
// type.
func wantFailedIn(t *testing.T, cands []Signature, ctx Context, args []Type) {
	t.Helper()
	got := Select(cands, args, ctx)
	if got.Outcome != Failed || got.Candidate != -1 || got.Returns != Unknown {
		t.Errorf("%v in %v: got %+v, want a failure with no candidate and no type", args, ctx, got)
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

// TestSelectNoArityMatch: RFC 0001 "Multi declarations", a call whose
// arity no candidate accepts fails (perigrin, 2026-10-02): `each` with two
// arguments has no candidate and no type, never the join of the two.
func TestSelectNoArityMatch(t *testing.T) {
	wantFailed(t, eachCandidates, []Type{Hash, Hash})
}

// TestSelectCodeAndGlobAreNotSlurpy: a code or glob slot, `Code &block`
// or `Glob *fh`, takes one argument as a scalar does; only `@` and `%`
// take the rest of the call.
func TestSelectCodeAndGlobAreNotSlurpy(t *testing.T) {
	wantFailed(t, []Signature{sig(Str, Param{Name: "block", Sigil: '&', Type: Code, Required: true})}, []Type{Code, Code})
	wantFailed(t, []Signature{sig(Str, Param{Name: "fh", Sigil: '*', Type: Glob, Required: true})}, []Type{Glob, Glob})
}

// TestSelectEmptyCandidates: selection over no candidates, or a call with
// no arguments that every candidate requires, has no type, and does not
// panic.
func TestSelectEmptyCandidates(t *testing.T) {
	wantFailed(t, nil, nil)
	wantFailed(t, nil, []Type{Int})
	wantFailed(t, []Signature{}, nil)
	wantFailed(t, eachCandidates, nil)
}

// TestSelectJoinsWhenUndecided: RFC 0001 "Multi declarations", when the
// call site cannot decide -- an argument of unknown type -- the call's type
// is the join of the candidates' returns.
func TestSelectJoinsWhenUndecided(t *testing.T) {
	cands := []Signature{
		sig(Str, scalar("h", Hash)),
		sig(ArrayRef, scalar("a", Array)),
	}
	got := Select(cands, []Type{Unknown}, ScalarCtx)
	if got.Outcome != Undecided || got.Candidate != -1 || got.Returns != Str|ArrayRef {
		t.Errorf("got %+v, want undecided, returning %v", got, Str|ArrayRef)
	}

	// Nor can it decide between candidates with no most specific one,
	// which a .pmt refuses (Ambiguity) and Go may still build.
	crossed := []Signature{
		sig(Str, scalar("a", Int), scalar("b", Num)),
		sig(ArrayRef, scalar("a", Num), scalar("b", Int)),
	}
	got = Select(crossed, []Type{Int, Int}, ScalarCtx)
	if got.Outcome != Undecided || got.Candidate != -1 || got.Returns != Str|ArrayRef {
		t.Errorf("crossed: got %+v, want undecided, returning %v", got, Str|ArrayRef)
	}
}

// TestSelectArityFailureIsNotJoined: an unknown argument type does not
// rescue an arity failure. `each` with two arguments of unknown type still
// fails rather than joining.
func TestSelectArityFailureIsNotJoined(t *testing.T) {
	wantFailed(t, eachCandidates, []Type{Unknown, Unknown})
}

// TestResultIsDeclaredReturnType: RFC 0001 "Multi declarations", a call's
// result is the selected candidate's declared return type (perigrin). An Int
// reaches `sqrt (Num $x) Num` and `abs (Num $x) Num` through Int <: Num, and
// each result is a Num: measured on 5.42.0, sqrt(4) is a float.
func TestResultIsDeclaredReturnType(t *testing.T) {
	sqrt := []Signature{sig(Num, scalar("x", Num))}
	wantSelected(t, sqrt, []Type{Int}, 0, Num)
	abs := []Signature{sig(Num, scalar("x", Num))}
	wantSelected(t, abs, []Type{Int}, 0, Num)
}

// in is sig answering only for the contexts cs, `:context(...)`.
func in(s Signature, cs ...Context) Signature {
	s.Context = ContextSet(cs...)
	return s
}

// localtimeCandidates are RFC 0001's `localtime`, the same parameters
// forked by context:
//
//	multi sub localtime :prototype(;$) :context($) (Int $time = time) Str;
//	multi sub localtime :prototype(;$) :context(@) (Int $time = time) List[Int];
var localtimeCandidates = []Signature{
	in(sig(Str, Param{Name: "time", Sigil: '$', Type: Int, Default: "time"}), ScalarCtx),
	in(sig(List, Param{Name: "time", Sigil: '$', Type: Int, Default: "time"}), ListCtx),
}

// TestMultiContextForkNotAmbiguous: RFC 0001 "`:context(...)`", context
// selects a declaration, so candidates whose parameters are the same and
// whose contexts are disjoint share no call. Candidates that both answer
// for scalar context are still ambiguous there, and the error names it.
func TestMultiContextForkNotAmbiguous(t *testing.T) {
	if err := Ambiguity(localtimeCandidates); err != nil {
		t.Errorf("localtime: got %v, want no ambiguity", err)
	}
	both := []Signature{localtimeCandidates[0], in(localtimeCandidates[1], ScalarCtx, ListCtx)}
	want := "candidates (Int $time) and (Int $time) are ambiguous for ()"
	if err := Ambiguity(both); err == nil || err.Error() != want {
		t.Errorf("overlapping contexts: got %v, want %q", err, want)
	}
	if err := Ambiguity([]Signature{localtimeCandidates[0], localtimeCandidates[0]}); err == nil {
		t.Errorf("duplicate scalar candidates: got no ambiguity")
	}

	// A third candidate decides a pair only in the contexts it answers
	// for: `(Int $a, Int $b)` in list context leaves two Ints in scalar
	// context ambiguous.
	crossed := []Signature{
		sig(Str, scalar("a", Int), scalar("b", Num)),
		sig(Str, scalar("a", Num), scalar("b", Int)),
		in(sig(Str, scalar("a", Int), scalar("b", Int)), ListCtx),
	}
	if err := Ambiguity(crossed); err == nil {
		t.Errorf("decided in list context only: got no ambiguity")
	}
}

// TestSelectByContext: RFC 0001 "`:context(...)`", a call selects the
// candidate that answers for its calling context. Measured on 5.42.0,
// `my $t = localtime(0)` is "Thu Jan  1 00:00:00 1970" and `my @l =
// localtime(0)` is nine integers. Boolean context is scalar: `wantarray`
// reports scalar inside `if (f())`, `!f()` and `f() and ...`.
func TestSelectByContext(t *testing.T) {
	for _, args := range [][]Type{nil, {Int}} {
		wantSelectedIn(t, localtimeCandidates, ScalarCtx, args, 0, Str)
		wantSelectedIn(t, localtimeCandidates, ListCtx, args, 1, List)
		wantSelectedIn(t, localtimeCandidates, BooleanCtx, args, 0, Str)
	}
}

// sortCandidates is RFC 0001's list-only `sort`, `sub sort :context(@)
// (...) List;`, whose scalar call is undefined: measured on 5.42.0,
// `scalar(sort(1,2))` is undef.
var sortCandidates = []Signature{
	in(sig(List, Param{Name: "list", Sigil: '@', Type: List}), ListCtx),
}

// TestSelectNoContextAnswerHasNoType: RFC 0001 "`:context(...)`", a call
// in a context no candidate answers for has no type. Scalar sort fails; it
// is not the list candidate's List.
func TestSelectNoContextAnswerHasNoType(t *testing.T) {
	wantFailedIn(t, sortCandidates, ScalarCtx, []Type{Int, Int})
	wantSelectedIn(t, sortCandidates, ListCtx, []Type{Int, Int}, 0, List)
}

// TestSelectVoidContextUnanswered: only a declaration stating no
// `:context`, or `:context()`, answers for void context, so a void call to
// localtime's scalar and list candidates has no type. A `:context()`
// candidate beside them takes it, and so does one stating no `:context`.
func TestSelectVoidContextUnanswered(t *testing.T) {
	wantFailedIn(t, localtimeCandidates, VoidCtx, []Type{Int})
	withVoid := append(slices.Clone(localtimeCandidates), in(sig(Undef, Param{Name: "time", Sigil: '$', Type: Int, Default: "time"}), VoidCtx))
	wantSelectedIn(t, withVoid, VoidCtx, []Type{Int}, 2, Undef)
	wantSelectedIn(t, intOrNum, VoidCtx, []Type{Int}, 0, Str)
}

// TestSelectBooleanIsNotList: boolean context is scalar, so a boolean call
// to the list-only sort has no type; it does not fall back to the list
// candidate.
func TestSelectBooleanIsNotList(t *testing.T) {
	wantFailedIn(t, sortCandidates, BooleanCtx, []Type{Int, Int})
}
