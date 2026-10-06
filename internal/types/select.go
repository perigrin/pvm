// ABOUTME: Multi selection: which of a name's typed signatures a call picks, by arity and argument types.
// ABOUTME: Three outcomes -- a selected candidate, an undecided join of candidates, a failure.

package types

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
// 0001 "Multi declarations".
func Select(cands []Signature, args []Type) Selection {
	for i, c := range cands {
		if c.accepts(len(args)) && c.fits(args) {
			return Selection{Outcome: Selected, Candidate: i, Returns: c.Returns}
		}
	}
	return Selection{Outcome: Failed, Candidate: -1}
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
