// ABOUTME: Multi selection: which of a name's typed signatures a call picks, by arity and argument types.
// ABOUTME: Three outcomes -- a selected candidate, an undecided join of candidates, a failure.

package types

import (
	"fmt"
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

// Select picks the candidate for a call with arguments of types args, RFC
// 0001 "Multi declarations". Of the candidates the call fits, the most
// specific wins (perigrin, 2026-10-02): the one whose parameters are all
// subtypes of every other's, as multi dispatch does in Raku, CLOS and Julia.
func Select(cands []Signature, args []Type) Selection {
	var fit []int
	for i, c := range cands {
		if c.accepts(len(args)) && c.fits(args) {
			fit = append(fit, i)
		}
	}
	if len(fit) == 0 {
		return Selection{Outcome: Failed, Candidate: -1}
	}
	best := fit[0]
	for _, i := range fit {
		if cands[i].asSpecific(cands[best], len(args)) {
			best = i
		}
	}
	return Selection{Outcome: Selected, Candidate: best, Returns: cands[best].Returns}
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
// slurpy, `List @list = die`, takes at least one.
func (s Signature) accepts(n int) bool {
	least, slurpy := 0, false
	for _, p := range s.Params {
		if p.Required {
			least++
		}
		slurpy = slurpy || p.Sigil != '$'
	}
	return n >= least && (slurpy || n <= len(s.Params))
}

// fits reports whether every argument's type is a subtype of its
// parameter's (RFC 0001, "What a parameter type means"). A parameter that
// states no type takes any argument; arguments past the last parameter are
// the slurpy's.
func (s Signature) fits(args []Type) bool {
	for i, a := range args {
		if !IsSubtype(a, s.paramType(i)) {
			return false
		}
	}
	return true
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
func Ambiguity(cands []Signature) error {
	most := 0
	for _, c := range cands {
		most = max(most, len(c.Params))
	}
	for i, a := range cands {
		for j := i + 1; j < len(cands); j++ {
			// A slurpy takes any count past its parameters, so one past
			// the longest stands for them all.
			for n := range most + 2 {
				if overlap, ok := ambiguous(cands, i, j, n); ok {
					return fmt.Errorf("candidates %s and %s are ambiguous for (%s)", a.params(), cands[j].params(), overlap)
				}
			}
		}
	}
	return nil
}

// ambiguous reports whether candidates i and j are ambiguous for n
// arguments, and the argument types both fit. Only a third candidate can
// decide between them.
func ambiguous(cands []Signature, i, j, n int) (string, bool) {
	a, b := cands[i], cands[j]
	if !a.accepts(n) || !b.accepts(n) {
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
		if k != i && k != j && c.accepts(n) && c.takesExactly(overlap) {
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

// params renders s's parameter list as a .pmt states it, `(Int $a, Num $b)`.
func (s Signature) params() string {
	ps := make([]string, len(s.Params))
	for i, p := range s.Params {
		ps[i] = fmt.Sprintf("%c%s", p.Sigil, p.Name)
		if p.Type != Unknown {
			ps[i] = p.Type.String() + " " + ps[i]
		}
	}
	return "(" + strings.Join(ps, ", ") + ")"
}
