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
// types args, RFC 0001 "Multi declarations". A candidate takes only calls
// in the contexts its `:context(...)` names ("`:context(...)`"). Of the
// candidates the call fits, the most specific wins (perigrin, 2026-10-02):
// the one whose parameters are all subtypes of every other's, as multi
// dispatch does in Raku, CLOS and Julia. When the call site cannot decide
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
	var fit, asLists []int
	for i, c := range cands {
		if c.contexts()&ContextSet(ctx) == 0 || !c.accepts(len(args)) || !c.fits(args) {
			continue
		}
		if takes, lists := c.takesShapes(shapes); takes {
			fit = append(fit, i)
			if lists {
				asLists = append(asLists, i)
			}
		}
	}
	if len(asLists) > 0 {
		fit = asLists
	}
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
// arguments, and the argument types both fit. Only a third candidate can
// decide between them, in every context both answer for. Candidates with
// no context in common share no call: context selects between them.
func ambiguous(cands []Signature, i, j, n int) (string, bool) {
	a, b := cands[i], cands[j]
	shared := a.contexts() & b.contexts()
	if !a.accepts(n) || !b.accepts(n) || shared == 0 {
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
		if k != i && k != j && c.accepts(n) && c.takesExactly(overlap) && c.contexts()&shared == shared {
			return "", false
		}
	}
	return strings.Join(names, ", "), true
}

// contexts is the set of contexts s answers for: its `:context(...)`, or
// every context when it states none.
func (s Signature) contexts() Contexts {
	if s.Context == EveryContext {
		return ContextSet(ScalarCtx, ListCtx, VoidCtx)
	}
	return s.Context
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
