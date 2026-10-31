// ABOUTME: Tests the Earley recognizer over a toy statement grammar: derivations, interning, the cache and Expected.
// ABOUTME: Also the empty, truncated and past-the-end edge cases; the toy-lattice rejoin is in lattice_test.go.

package recognize

import (
	"slices"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
)

// stmtGrammar is the shapes the cut grammar gives the recognizer, in small:
// a right-recursive nullable statement sequence, nested blocks, and a
// left-recursive term.
var stmtGrammar = []string{
	"Program ::= Seq",
	"Seq ::= Stmt Seq |",
	"Stmt ::= My Var Assign Term Semi | Print Term Semi | LBrace Seq RBrace",
	"Term ::= Var | Num | Term Add Num",
}

// toyStmt is stmtGrammar's tables.
func toyStmt(tb testing.TB) *grammar.Tables { return toyTables(tb, stmtGrammar...) }

// run recognizes src, terminal names separated by spaces, with r.
func run(r *Recognizer, src string) *Chart {
	return Recognize(r, tokens(strings.Fields(src)...))
}

const decl = "My Var Assign Num Semi" // a five-token statement

// TestRecognizeDerivations: spec §10's three, on the toy grammar. `Num Num`
// and `Var Var` have no derivation; `Print Num` has one.
func TestRecognizeDerivations(t *testing.T) {
	r := New(toyStmt(t), Options{})
	for src, want := range map[string]bool{"Num Num": false, "Var Var": false, "Print Num": false, "Print Num Semi": true, decl + " Print Var Semi": true} {
		if got := run(r, src).Accepted(); got != want {
			t.Errorf("Accepted(%q) = %v, want %v", src, got, want)
		}
	}
}

// TestRecognizeEmptyAndTruncated: no token is the empty program, which is
// accepted; one truncated mid-statement is refused, and what it expects at
// its end includes a term.
func TestRecognizeEmptyAndTruncated(t *testing.T) {
	r := New(toyStmt(t), Options{})
	if !run(r, "").Accepted() {
		t.Error("the empty program was refused, want accepted")
	}
	c := run(r, "My Var Assign")
	if c.Accepted() {
		t.Fatal("`My Var Assign` accepted, want refused")
	}
	exp, err := c.Expected(len(c.Sets) - 1)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(exp, "Var") || !slices.Contains(exp, "Num") {
		t.Errorf("Expected at the end of `My Var Assign` = %v, want a term: Var and Num among them", exp)
	}
	if slices.Contains(exp, "Semi") {
		t.Errorf("Expected at the end of `My Var Assign` holds Semi: %v", exp)
	}
}

// TestExpectedPastEnd: Expected outside the lattice's nodes is an error, and
// never a panic.
func TestExpectedPastEnd(t *testing.T) {
	r := New(toyStmt(t), Options{})
	c := run(r, "Print Num Semi")
	if _, err := c.Expected(len(c.Sets) - 1); err != nil {
		t.Fatalf("Expected at the last node: %v", err)
	}
	for _, pos := range []int{len(c.Sets), len(c.Sets) + 10, -1} {
		if exp, err := c.Expected(pos); err == nil && len(exp) > 0 {
			t.Errorf("Expected(%d) past the lattice = %v, nil; want an error or an empty set", pos, exp)
		}
	}
	// A refused input's dead positions have no set: empty, not a panic.
	d := run(r, "Num Num Num Num")
	for pos := range d.Sets {
		if _, err := d.Expected(pos); err != nil {
			t.Errorf("Expected(%d) on a refused input: %v", pos, err)
		}
	}
}

// TestInternSharesEqualSets: positions with equal content share one interned
// set (§10). A statement repeated after itself reaches the same sets.
func TestInternSharesEqualSets(t *testing.T) {
	r := New(toyStmt(t), Options{})
	// A fourth statement, so the third ends before the same lookahead.
	c := run(r, strings.Repeat(decl+" ", 4))
	if !c.Accepted() {
		t.Fatal("refused")
	}
	// 5 tokens a statement: node 5 ends the first, node 10 the second.
	for i := 1; i <= 5; i++ {
		if c.Sets[i] != c.Sets[i+5] || c.Sets[i+5] != c.Sets[i+10] {
			t.Errorf("node %d's set differs from nodes %d and %d", i, i+5, i+10)
		}
		if !c.Interned[i+5] || !c.Interned[i+10] {
			t.Errorf("nodes %d and %d were not interned hits", i+5, i+10)
		}
	}
	if hits, n := c.Reuse(); hits == 0 || hits >= n {
		t.Errorf("reuse %d of %d, want some but not all positions", hits, n)
	}
}

// TestInternDistinctSets: positions whose sets differ in content never
// share an interned set.
func TestInternDistinctSets(t *testing.T) {
	r := New(toyStmt(t), Options{})
	c := run(r, decl+" Print Var Semi LBrace My Var Assign Var Add Num Semi RBrace")
	if !c.Accepted() {
		t.Fatal("refused")
	}
	for i, a := range c.Sets {
		for j, b := range c.Sets[i+1:] {
			if a == b && !slices.Equal(a.Items(), b.Items()) {
				t.Errorf("nodes %d and %d share a set but differ in content", i, i+1+j)
			}
			if a != b && slices.Equal(a.Items(), b.Items()) {
				t.Errorf("nodes %d and %d are equal in content but were not shared", i, i+1+j)
			}
		}
	}
}

// TestCacheHitsOnSecondRecognition: recognizing the same input again hits
// the transition cache at every position after the first.
func TestCacheHitsOnSecondRecognition(t *testing.T) {
	r := New(toyStmt(t), Options{})
	src := decl + " Print Var Semi LBrace " + decl + " RBrace"
	first, second := run(r, src), run(r, src)
	if !first.Accepted() || !second.Accepted() {
		t.Fatal("refused")
	}
	for i := 1; i < len(second.Sets); i++ {
		if !second.Cached[i] {
			t.Errorf("node %d missed the transition cache on the second recognition", i)
		}
		if second.Sets[i] != first.Sets[i] {
			t.Errorf("node %d's set differs between the two recognitions", i)
		}
	}
}

// TestCacheHitKeepsAcceptance: a cache hit never changes acceptance. Every
// input is recognized by a recognizer warmed on all the others, and by one
// with no cache.
func TestCacheHitKeepsAcceptance(t *testing.T) {
	tab := toyStmt(t)
	warm := New(tab, Options{})
	srcs := []string{
		decl, "My Var Assign Num", "My Var Assign", "Print Num Semi", "Num Num Semi", "Var Var Semi",
		"LBrace " + decl + " RBrace", "LBrace My Var Assign Num Semi", "Print Var Add Num Semi", "Print Var Add Semi",
	}
	hits := 0
	for range 2 {
		for _, src := range srcs {
			c := run(warm, src)
			cold := run(New(tab, Options{NoCache: true}), src)
			if c.Accepted() != cold.Accepted() {
				t.Errorf("%q: accepted %v with the cache, %v without", src, c.Accepted(), cold.Accepted())
			}
			for _, h := range c.Cached {
				if h {
					hits++
				}
			}
		}
	}
	if hits == 0 {
		t.Error("no position hit the cache, so nothing was tested")
	}
}

// TestLookaheadKeepsAcceptance: lookahead only prunes completions that
// cannot continue, so acceptance is the same with it off.
func TestLookaheadKeepsAcceptance(t *testing.T) {
	tab := toyStmt(t)
	for _, src := range []string{decl, "My Var Assign", "Print Num", "Num Num", "LBrace Print Num Semi RBrace Num Semi", "LBrace RBrace"} {
		la := run(New(tab, Options{}), src).Accepted()
		no := run(New(tab, Options{NoLookahead: true}), src).Accepted()
		if la != no {
			t.Errorf("%q: accepted %v with lookahead, %v without", src, la, no)
		}
	}
}

// TestExpectedAfterAssign: after `My Var Assign` a term is expected, and
// `Semi` is not.
func TestExpectedAfterAssign(t *testing.T) {
	c := run(New(toyStmt(t), Options{}), decl)
	exp, err := c.Expected(3)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(exp, "Var") || slices.Contains(exp, "Semi") {
		t.Errorf("Expected after `My Var Assign` = %v, want Var and no Semi", exp)
	}
}

// TestExpectedIgnoresLookahead: Expected(pos) is the union over the Earley
// set as recognition without lookahead builds it, at every node, refusal
// points included: after `My Var Assign Num` before a second `Num`,
// lookahead prunes every completion, but an operator is still expected there.
func TestExpectedIgnoresLookahead(t *testing.T) {
	tab := toyStmt(t)
	for _, src := range []string{"My Var Assign Num Num Semi", decl + " Print Var Add Num Semi LBrace " + decl + " RBrace"} {
		la := run(New(tab, Options{}), src)
		no := run(New(tab, Options{NoLookahead: true}), src)
		for pos := range la.Sets {
			a, err := la.Expected(pos)
			if err != nil {
				t.Fatal(err)
			}
			b, err := no.Expected(pos)
			if err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(a, b) {
				t.Errorf("%q node %d: Expected has %d terminals with lookahead, %d without", src, pos, len(a), len(b))
			}
		}
	}
	c := run(New(tab, Options{}), "My Var Assign Num Num Semi")
	if exp, _ := c.Expected(4); !slices.Contains(exp, "Add") {
		t.Errorf("Expected after `My Var Assign Num` = %v, want Add among them", exp)
	}
}
