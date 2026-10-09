// ABOUTME: Multi selection: which of a name's typed signatures a call picks, by arity and argument types.
// ABOUTME: Three outcomes -- a selected candidate, an undecided join of candidates, a failure.

package types

import (
	"fmt"
	"slices"
	"strings"
)

// Outcome is what selection over a name's candidates concluded.
type Outcome int

const (
	// Failed: no candidate accepts the call. It has no type.
	Failed Outcome = iota

	// Selected: one candidate is the call's.
	Selected

	// Undecided: the call site cannot decide between candidates, so its
	// type is the join of theirs.
	Undecided
)

// Selection is the result of Select. Candidate is the selected candidate's
// index, -1 unless the outcome is Selected; Returns is the call's type,
// Unknown on failure.
type Selection struct {
	Outcome   Outcome
	Candidate int
	Returns   Type
}

// Select picks the candidate for a call in context ctx with arguments of
// types args, RFC 0001 "Multi declarations". Of the candidates the call
// fits, those whose return type answers ctx are taken ("Context selects by
// return type"; see Taking). Of those, the most specific wins (perigrin,
// 2026-10-02): the one whose parameters are all subtypes of every other's,
// as multi dispatch does in Raku, CLOS and Julia. When the call site cannot decide
// -- an argument of unknown type fits more than one candidate, or none is
// most specific -- the call's type is the join of theirs. A call no
// candidate takes fails, and is never joined.
//
// ctx is the context the call is in: ScalarCtx (BooleanCtx is the same),
// ListCtx or VoidCtx. What a call whose context is undetermined passes is
// RFC 0001 "Call sites"'s; no candidate takes UnknownCtx.
func Select(cands []Signature, args []Type, ctx Context) Selection {
	return SelectShaped(cands, args, "", ctx)
}

// SelectShaped is Select for a call whose operands' shapes the call site
// states, RFC 0001 "Operators that fork": shapes[i] is `@` when the i'th
// operand is a parenthesised list, `(1,2) x 2`, and `$` otherwise. The
// parenthesis is a parse fact, so types cannot stand in for it. A `@`
// parameter takes only a `@` operand. A `$` parameter takes either, as perl
// evaluates a parenthesised list in scalar context: `my $x = (1,2) x 2` is
// `22`. Candidates taking every `@` operand as a list rule out those taking
// one as a scalar, before the most specific is sought. With no shapes
// stated, as for a call to a sub, none is ruled out by shape.
func SelectShaped(cands []Signature, args []Type, shapes string, ctx Context) Selection {
	fit := Taking(cands, args, shapes, ctx)
	if len(fit) == 0 {
		return Selection{Outcome: Failed, Candidate: -1}
	}
	if len(fit) == 1 || !slices.Contains(args, Unknown) {
		if best, ok := mostSpecific(cands, fit, len(args)); ok {
			return Selection{Outcome: Selected, Candidate: best, Returns: cands[best].Returns}
		}
	}
	joined := Unknown
	for _, i := range fit {
		joined = Join(joined, cands[i].Returns)
	}
	return Selection{Outcome: Undecided, Candidate: -1, Returns: joined}
}

// Taking is the candidates of cands, by index, a call in context ctx with
// arguments of types args and operands of the given shapes can be, before
// the most specific is sought: those that take its arity, argument types
// and shapes, then of those the ones Answering ctx, then of those the ones
// taking every `@` operand as a list, if any do. Context comes before
// shape: `my $x = (1,2) x 2` is `22`, the Str candidate's, though the List
// one takes the parenthesised operand as a list.
func Taking(cands []Signature, args []Type, shapes string, ctx Context) []int {
	var fit []int
	for i, c := range cands {
		if takes, _ := c.takesShapes(shapes); takes && c.accepts(len(args)) && c.fits(args) {
			fit = append(fit, i)
		}
	}
	fit = answering(cands, fit, ctx)
	asLists := slices.DeleteFunc(slices.Clone(fit), func(i int) bool {
		_, lists := cands[i].takesShapes(shapes)
		return !lists
	})
	if len(asLists) > 0 {
		return asLists
	}
	return fit
}

// Answering is the candidates of cands, by index, that answer a call in
// context ctx, RFC 0001 "Context selects by return type": those whose
// return type answers ctx (see Signature.Answers), or, when none does,
// every one, whose result the context coerces as it coerces any value --
// grep's List is a count in scalar context. Void context is a form of
// scalar context (perlglossary, "void context"), so with no Void candidate
// it is answered as scalar context is. No candidate answers UnknownCtx, a
// context the call site has not determined.
func Answering(cands []Signature, ctx Context) []int {
	all := make([]int, len(cands))
	for i := range cands {
		all[i] = i
	}
	return answering(cands, all, ctx)
}

// answering is Answering over the candidates among.
func answering(cands []Signature, among []int, ctx Context) []int {
	if ctx == UnknownCtx {
		return nil
	}
	wants := []Context{ctx}
	if ctx == VoidCtx {
		wants = append(wants, ScalarCtx)
	}
	for _, want := range wants {
		matched := slices.DeleteFunc(slices.Clone(among), func(i int) bool { return cands[i].Answers() != want })
		if len(matched) > 0 {
			return matched
		}
	}
	return among
}

// Answers is the calling context s's return type answers: VoidCtx for
// Void, ListCtx for a list -- a return holding an Array or a Hash, as
// List, List[...], Array and Hash do -- and ScalarCtx for any other
// stated return. A list is not a subtype test: Scalar <: List, so every
// scalar return is under List too. UnknownCtx when s states no return.
func (s Signature) Answers() Context {
	switch {
	case s.Returns == Void:
		return VoidCtx
	case s.Returns&(Array|Hash) != 0:
		return ListCtx
	case s.Returns == Unknown:
		return UnknownCtx
	}
	return ScalarCtx
}

// mostSpecific is the candidate of fit as specific as every other, if one
// is, for a call of n arguments.
func mostSpecific(cands []Signature, fit []int, n int) (int, bool) {
	for _, i := range fit {
		if !slices.ContainsFunc(fit, func(j int) bool { return !cands[i].asSpecific(cands[j], n) }) {
			return i, true
		}
	}
	return -1, false
}

// asSpecific reports whether, for a call of n arguments, each of s's
// parameter types is a subtype of o's.
func (s Signature) asSpecific(o Signature, n int) bool {
	for i := range n {
		if !IsSubtype(s.paramType(i), o.paramType(i)) {
			return false
		}
	}
	return true
}

// accepts reports whether s takes n arguments: at least its required
// parameters, and no more than it has unless one is a slurpy. A required
// slurpy, `List @list = die`, takes at least one. The n arguments are a
// call's comma-separated ones, which write no invocant, so a candidate
// whose invocant is required takes none.
func (s Signature) accepts(n int) bool {
	if s.Invocant != nil && s.Invocant.Required {
		return false
	}
	least, slurpy := 0, false
	for _, p := range s.Params {
		if p.Required {
			least++
		}
		slurpy = slurpy || p.Slurpy()
	}
	return n >= least && (slurpy || n <= len(s.Params))
}

// Accepts reports whether some candidate takes a call of n arguments, in
// any context: the arity half of Select. A call that writes an invocant
// counts it among its n, as Ambiguity's byInvocant does.
func Accepts(cands []Signature, n int) bool {
	for _, written := range []bool{false, true} {
		view, _ := byInvocant(cands, written)
		if slices.ContainsFunc(view, func(c Signature) bool { return c.accepts(n) }) {
			return true
		}
	}
	return false
}

// fits reports whether every argument's type is a subtype of its
// parameter's (RFC 0001, "What a parameter type means"). A parameter that
// states no type takes any argument, and an argument of unknown type may be
// any parameter's; arguments past the last parameter are the slurpy's.
func (s Signature) fits(args []Type) bool {
	for i, a := range args {
		if a != Unknown && !IsSubtype(a, s.paramType(i)) {
			return false
		}
	}
	return true
}

// takesShapes reports whether s takes operands of the given shapes, a `@`
// parameter only a `@` operand, and whether it takes every `@` operand as
// a list.
func (s Signature) takesShapes(shapes string) (takes, asLists bool) {
	asLists = true
	for i, shape := range shapes {
		list := len(s.Params) > 0 && s.Params[min(i, len(s.Params)-1)].Sigil == '@'
		if list && shape != '@' {
			return false, false
		}
		asLists = asLists && (list || shape != '@')
	}
	return true, asLists
}

// paramType is the type the i'th argument is taken as: its parameter's,
// the slurpy's past the end, and Any where no type is stated.
func (s Signature) paramType(i int) Type {
	p := s.Params[min(i, len(s.Params)-1)]
	if p.Type == Unknown {
		return Any
	}
	return p.Type
}

// Ambiguity reports a pair of candidates with no single most specific one
// for some call, RFC 0001 "Multi declarations": `(Int $a, Num $b)` beside
// `(Num $a, Int $b)`, both fitting two Ints. A pair is ambiguous for n
// arguments when both take n, every parameter's types overlap, neither is
// strictly more specific, and no third candidate takes exactly the overlap
// -- `(Int $a, Int $b)` decides the call above. Identical candidates are
// ambiguous: each is as specific as the other, and neither is more.
//
// A call either writes an invocant or does not (RFC 0001, "Builtins that
// keep their own parse"), so candidates are compared over each kind of
// call apart: sort's `(Code|Str $by: List @list)` and `(List @list = die)`
// share none.
func Ambiguity(cands []Signature) error {
	for _, written := range []bool{false, true} {
		view, orig := byInvocant(cands, written)
		most := 0
		for _, c := range view {
			most = max(most, len(c.Params))
		}
		for i := range view {
			for j := i + 1; j < len(view); j++ {
				// A slurpy takes any count past its parameters, so one past
				// the longest stands for them all.
				for n := range most + 2 {
					if overlap, ok := ambiguous(view, i, j, n); ok {
						return fmt.Errorf("candidates %s and %s are ambiguous for (%s)", cands[orig[i]].params(), cands[orig[j]].params(), overlap)
					}
				}
			}
		}
	}
	return nil
}

// byInvocant is the candidates that take a call which writes an invocant,
// or which writes none, each as a signature over that call's arguments,
// with the index in cands of the candidate it stands for. A written
// invocant is the call's first argument. A required invocant takes only a
// call that writes one; one with a default, print's, takes either.
func byInvocant(cands []Signature, written bool) (view []Signature, orig []int) {
	for i, c := range cands {
		switch {
		case written && c.Invocant != nil:
			inv := *c.Invocant
			inv.Required = true
			c.Params = append([]Param{inv}, c.Params...)
		case !written && (c.Invocant == nil || !c.Invocant.Required):
		default:
			continue
		}
		c.Invocant = nil
		view, orig = append(view, c), append(orig, i)
	}
	return view, orig
}

// ambiguous reports whether candidates i and j are ambiguous for n
// arguments, and the argument types both fit. Only a third candidate
// answering the same context can decide between them. Candidates whose
// return types answer different contexts share no call: context selects
// between them.
func ambiguous(cands []Signature, i, j, n int) (string, bool) {
	a, b := cands[i], cands[j]
	if !a.accepts(n) || !b.accepts(n) || a.Answers() != b.Answers() {
		return "", false
	}
	if a.asSpecific(b, n) != b.asSpecific(a, n) {
		return "", false
	}
	overlap := make([]Type, n)
	names := make([]string, n)
	for k := range n {
		overlap[k] = a.paramType(k) & b.paramType(k)
		if overlap[k] == 0 {
			return "", false
		}
		names[k] = overlap[k].String()
	}
	for k, c := range cands {
		if k != i && k != j && c.accepts(n) && c.takesExactly(overlap) && c.Answers() == a.Answers() {
			return "", false
		}
	}
	return strings.Join(names, ", "), true
}

// takesExactly reports whether s's parameter types are exactly ts.
func (s Signature) takesExactly(ts []Type) bool {
	for i, t := range ts {
		if s.paramType(i) != t {
			return false
		}
	}
	return true
}

// params renders s's parameter list as a .pmt states it, `(Int $a, Num $b)`
// or, with an invocant, `(Code $c: List @l)`.
func (s Signature) params() string {
	ps := make([]string, len(s.Params))
	for i, p := range s.Params {
		ps[i] = p.decl()
	}
	if s.Invocant != nil {
		return "(" + s.Invocant.decl() + ": " + strings.Join(ps, ", ") + ")"
	}
	return "(" + strings.Join(ps, ", ") + ")"
}

// decl renders p as a .pmt states it, without its default: `Int $a`, or
// `Array \@a` for an aliased parameter.
func (p Param) decl() string {
	if p.Type == Unknown {
		return p.Variable()
	}
	return p.Type.String() + " " + p.Variable()
}
