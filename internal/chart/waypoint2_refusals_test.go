// ABOUTME: Waypoint 2's refusals through the whole pipeline (chart/lex, recognition, the forest, its constraints), each
// ABOUTME: beside a nearby form perl accepts: the precedence refusals the ledger and spec §5.2 name.

package chart

import "testing"

// waypoint2Refusals are perl -c's verdicts (94e50866) on the precedence refusals, each beside a nearby form it accepts.
var waypoint2Refusals = []struct {
	src      string
	accepted bool
}{
	// The dangling comparison chains, perly.y:1466-1487: a non-chaining comparison chained with another.
	{"$a <=> $b <=> $c;", false},
	{"$a <=> $b;", true},
	{"$a == $b <=> $c;", false},
	{"$a <=> $b == $c;", false},
	{"$a cmp $b eq $c;", false},
	{"$a == $b == $c;", true},
	{`use feature "isa"; $a isa Foo isa Bar;`, false},
	{`use feature "isa"; $a < $b isa Foo;`, false},
	{`use feature "isa"; $a isa Foo < $b;`, false},
	{`use feature "isa"; $a isa Foo;`, true},
	{"$a < $b < $c;", true},
	// DOTDOT's %nonassoc.
	{"1 .. 2 .. 3;", false},
	{"1 .. 2;", true},
	// The split chains: a tighter operator between two of the same chain.
	{"1 .. 2 + 3 .. 4;", false},
	{"1 .. 2 + 3;", true},
	{"$a <=> $b + 1 <=> $c;", false},
	{"$a <=> $b + 1;", true},
	// indirob: scalar %prec PREC_LOW loses to [ after print.
	{"print $fh [];", false},
	{"print $$fh[0];", true},
}

func TestWaypoint2Refusals(t *testing.T) {
	for _, c := range waypoint2Refusals {
		if got := constrainSource(t, []byte(c.src)); got != c.accepted {
			t.Errorf("%q: the chart accepts=%v, perl -c accepts=%v", c.src, got, c.accepted)
		}
	}
}
