// ABOUTME: LeoDisagree: Leo on and off compared on acceptance and on Expected at every position (spec §9.2).
// ABOUTME: The ratchet's leo-disagree counts this function over the corpora rather than restating it.

package recognize

import (
	"slices"

	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/chart/lex"
)

// LeoDisagree recognizes l over tab with Leo on and with it off, on fresh
// recognizers, and reports whether they differ in acceptance or in Expected
// at any node. An invalid lattice is an error, never "no disagreement", so
// a file that could not be run cannot count toward zero.
func LeoDisagree[S lex.State[S]](tab *grammar.Tables, l *lex.Lattice[S]) (bool, error) {
	on, err := Recognize(New(tab, Options{}), l)
	if err != nil {
		return false, err
	}
	off, err := Recognize(New(tab, Options{NoLeo: true}), l)
	if err != nil {
		return false, err
	}
	return chartsDisagree(on, off), nil
}

func chartsDisagree(a, b *Chart) bool {
	if a.Accepted() != b.Accepted() || len(a.Sets) != len(b.Sets) {
		return true
	}
	for pos := range a.Sets {
		x, _ := a.Expected(pos)
		y, _ := b.Expected(pos)
		if !slices.Equal(x, y) {
			return true
		}
	}
	return false
}
