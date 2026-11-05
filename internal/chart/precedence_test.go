// ABOUTME: Recognition ignores precedence at waypoint 1 (spec §8): with perly.y's declarations and %prec uses in the
// ABOUTME: grammar, forms perl refuses only by precedence are still accepted; waypoint 2's constraint refuses them.

package chart

import (
	"slices"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
)

func TestPrecedenceIgnoredByRecognition(t *testing.T) {
	g, _, err := grammar.Committed()
	if err != nil {
		t.Fatal(err)
	}
	for _, tok := range []string{"NCEQOP", "DOTDOT", "PERLY_BRACKET_OPEN", "ADDOP", "MULOP"} {
		if !slices.Contains(g.Prec, tok) {
			t.Errorf("the committed grammar declares no precedence for %s", tok)
		}
	}
	if !slices.ContainsFunc(g.G1, func(r grammar.Rule) bool { return r.Prec != "" }) {
		t.Error("no committed rule carries a precedence token (Rule.Prec)")
	}
	// The first three perl refuses by precedence alone: <=> chained by termbinop's `termeqop %prec PREC_LOW`
	// (perly.y:1438) losing to NCEQOP, which shifts into the error rule perly.y:1484 (166 is %left);
	// 1 .. 2 .. 3 by DOTDOT's %nonassoc; and print $fh [] by indirob's PREC_LOW.
	for _, src := range []string{"$a <=> $b <=> $c;", "1 .. 2 .. 3;", "print $fh [];", "$a + $b * $c;"} {
		if v := recognizeSource(t, []byte(src)); !v.accepted {
			t.Errorf("%q is refused at byte %d (expected %v); recognition must ignore precedence", src, v.at, v.expected)
		}
	}
}
