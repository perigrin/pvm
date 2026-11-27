// ABOUTME: The refusals waypoint 2 builds, through the whole pipeline (chart/lex, recognition, the forest, its constraints), each
// ABOUTME: beside a nearby form perl accepts: the precedence refusals and the scope refusals the ledger and spec §5.2 name.

package chart

import (
	"testing"

	"tamarou.com/pvm/internal/chart/forest"
	"tamarou.com/pvm/internal/chart/lex"
	"tamarou.com/pvm/internal/chart/prectab"
	"tamarou.com/pvm/internal/chart/recognize"
)

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

// classOn is the preamble the class rows take: perl reads field, method and ADJUST as keywords only under it.
const classOn = "use feature 'class'; no warnings; "

// waypoint2ScopeRefusals are perl -c's verdicts (94e50866) on the refusals resolve makes under scope and features,
// each src read after its pre, beside a nearby form perl accepts: the class feature's keywords outside a class
// ("Cannot 'field' outside of a 'class'"), and a signature after an attribute list with signatures off.
var waypoint2ScopeRefusals = []struct {
	src, pre string
	accepted bool
}{
	{src: "field $x;", pre: classOn},
	{src: "class Foo { field $x; }", pre: classOn, accepted: true},
	{src: "method m { 1 }", pre: classOn},
	{src: "class Foo { method m { 1 } }", pre: classOn, accepted: true},
	{src: "my $c = method { 1 };", pre: classOn},
	{src: "class Foo { method m { my $c = method { 1 }; } }", pre: classOn, accepted: true},
	{src: "ADJUST { 1 }", pre: classOn},
	{src: "class Foo { ADJUST { 1 } }", pre: classOn, accepted: true},
	{src: "sub f :lvalue ($x) {}"},
	{src: "sub f :lvalue ($x) {}", pre: "use v5.36; ", accepted: true},
	{src: "sub f ($x) {}", accepted: true},
}

func TestWaypoint2Refusals(t *testing.T) {
	for _, c := range waypoint2Refusals {
		if got := pipelineSource(t, c.src); got != c.accepted {
			t.Errorf("%q: the chart accepts=%v, perl -c accepts=%v", c.src, got, c.accepted)
		}
	}
	for _, c := range waypoint2ScopeRefusals {
		if got := pipelineSource(t, c.pre+c.src); got != c.accepted {
			t.Errorf("%q: the chart accepts=%v, perl -c accepts=%v", c.pre+c.src, got, c.accepted)
		}
	}
}

// pipelineSource is whether src survives the whole pipeline: forest.Parse, which applies the precedence constraint
// and resolves scope with nothing imported.
func pipelineSource(t *testing.T, src string) bool {
	t.Helper()
	tab, err := committed()
	if err != nil {
		t.Fatal(err)
	}
	f, _ := forest.Parse(tab, []byte(src), forest.Scope{})
	return f != nil
}

// TestWaypoint2RefusalsMatchReference requires production's forest of every row of both tables under Scope{}, before
// resolve (forest.CandidateForest: the judged chart of the narrowed lattice, then Build, with no Constrain), to equal
// ConstrainReference's judgement of the no-precedence forest after LibraryOperators under Scope{}. A row the judged
// chart refuses is compared by verdict alone (d247's rule): the reference must refuse it too, by its own verdict or
// because the no-precedence chart or LibraryOperators does. A row production accepts and the reference has no forest
// for fails, so none drops out of the count.
func TestWaypoint2RefusalsMatchReference(t *testing.T) {
	tab, err := committed()
	if err != nil {
		t.Fatal(err)
	}
	p, err := prectab.ParsePrecedence(tab)
	if err != nil {
		t.Fatal(err)
	}
	entries, compared, verdictOnly := 0, 0, 0
	compare := func(src string) {
		entries++
		var ref *forest.Forest
		s, _ := lex.Lex([]byte(src), tab.Tokens)
		l := lex.Linear(s)
		if c, err := recognize.Recognize(recognize.NewNoPrecedence(tab, recognize.Options{}), l); err == nil && c.Accepted() {
			if ref = forest.Build(tab, c, l); forest.LibraryOperators(tab, ref, s, []byte(src), forest.Scope{}) >= 0 || !ref.ConstrainReference(p) {
				ref = nil
			}
		}
		cand, _ := forest.CandidateForest(tab, []byte(src), forest.Scope{})
		if cand == nil {
			verdictOnly++
			if ref != nil {
				t.Errorf("%q: the judged chart refuses it, ConstrainReference accepts it", src)
			}
			return
		}
		compared++
		if ref == nil {
			t.Errorf("%q: production accepts it, the reference refuses it", src)
		} else if d := forest.ForestDiff(cand, ref); d != "" {
			t.Errorf("%q: %s", src, d)
		}
	}
	for _, c := range waypoint2Refusals {
		compare(c.src)
	}
	for _, c := range waypoint2ScopeRefusals {
		compare(c.pre + c.src)
	}
	t.Logf("entries=%d compared=%d", entries, compared)
	t.Logf("%s verdict only: %d", t.Name(), verdictOnly)
	if compared == 0 || compared+verdictOnly != entries {
		t.Errorf("compared %d and %d by verdict only of %d rows; want every row, and some compared", compared, verdictOnly, entries)
	}
}

// TestForestEqualityCandidateNeverConstrains: counted by Constrain's own hook, the waypoint2 differential's
// candidate (forest.CandidateForest) enters Constrain no time over every row, and accepts some.
func TestForestEqualityCandidateNeverConstrains(t *testing.T) {
	tab, err := committed()
	if err != nil {
		t.Fatal(err)
	}
	accepted, n0 := 0, forest.ConstrainEntries()
	for _, src := range waypoint2Sources() {
		if f, _ := forest.CandidateForest(tab, []byte(src), forest.Scope{}); f != nil {
			accepted++
		}
	}
	entries := forest.ConstrainEntries() - n0
	t.Logf("waypoint2: %d rows, %d accepted, Constrain entered %d times", len(waypoint2Sources()), accepted, entries)
	if accepted == 0 || entries != 0 {
		t.Errorf("the candidate accepted %d rows and entered Constrain %d times; want some accepted, and no entry", accepted, entries)
	}
}
