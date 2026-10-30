// ABOUTME: Tests the Earley recognizer over the cut Guacamole grammar: derivations, interning, the cache and Expected.
// ABOUTME: Also the toy-lattice rejoin, and the empty, truncated and past-the-end edge cases.

package recognize

import (
	"os"
	"slices"
	"sync"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/chart/lex"
)

const slif = "../grammar/testdata/guacamole/guacamole.slif"

var (
	guacOnce sync.Once
	guacTab  *grammar.Tables
	guacLex  *lex.Lexer
	guacErr  error
)

// guacamole is a fresh recognizer over the cut Guacamole grammar, and its
// lexer. The tables are built once per test binary; each call gets its own
// intern table and cache.
func guacamole(t testing.TB) (*Recognizer, *lex.Lexer) {
	t.Helper()
	guacOnce.Do(func() { guacTab, guacLex, guacErr = Guacamole(slif) })
	if guacErr != nil {
		t.Fatal(guacErr)
	}
	return New(guacTab, Options{}), guacLex
}

// run lexes src and recognizes it with r.
func run(t testing.TB, r *Recognizer, x *lex.Lexer, src string) *Chart {
	t.Helper()
	s, _ := x.Lex([]byte(src)) // an unmapped token is refused by the recognizer
	return Recognize(r, lex.Linear(s))
}

// TestRecognizeDerivations: spec §10's three, on the cut grammar. `1 2` and
// `$x $y` have no derivation; `print(1)` has one.
func TestRecognizeDerivations(t *testing.T) {
	r, x := guacamole(t)
	for src, want := range map[string]bool{"1 2": false, "$x $y": false, "print(1)": true, "print(1);": true, "my $x = 1;\nprint $x;\n": true} {
		if got := run(t, r, x, src).Accepted(); got != want {
			t.Errorf("Accepted(%q) = %v, want %v", src, got, want)
		}
	}
}

// TestRecognizeEmptyAndTruncated: a source with no significant token is the
// empty program, which perl accepts; one truncated mid-statement is refused,
// and what it expects at its end includes a term.
func TestRecognizeEmptyAndTruncated(t *testing.T) {
	r, x := guacamole(t)
	for name, src := range map[string]string{
		"empty":            "",
		"comments only":    "# nothing\n# here\n",
		"comments and POD": "# a comment\n\n=pod\n\nText.\n\n=cut\n\n# another\n",
	} {
		if !run(t, r, x, src).Accepted() {
			t.Errorf("%s: refused, want accepted", name)
		}
	}
	c := run(t, r, x, "my $x =")
	if c.Accepted() {
		t.Fatal("`my $x =` accepted, want refused")
	}
	exp, err := c.Expected(len(c.Sets) - 1)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(exp, "ScalarVar") || !slices.Contains(exp, "NumberDec") {
		t.Errorf("Expected at the end of `my $x =` = %v, want a term: ScalarVar and NumberDec among them", exp)
	}
	if slices.Contains(exp, "Semicolon") {
		t.Errorf("Expected at the end of `my $x =` holds Semicolon: %v", exp)
	}
}

// TestRecognizeSpacedDiamondRefused: perl parses `< $fh >` as a glob, not a
// readline, so the Guacamole parsent case holding it is refused, and the
// readline itself is accepted.
func TestRecognizeSpacedDiamondRefused(t *testing.T) {
	r, x := guacamole(t)
	src, err := os.ReadFile("../testdata/guacamole/parsent/Statements-Expressions-Value-NonLiteral-DiamondExpr-001.pl")
	if err != nil {
		t.Fatal(err)
	}
	if run(t, r, x, string(src)).Accepted() {
		t.Errorf("%q accepted, want refused", src)
	}
	if !run(t, r, x, "my $l = <$fh>;").Accepted() {
		t.Error("`my $l = <$fh>;` refused, want accepted")
	}
}

// TestExpectedPastEnd: Expected outside the lattice's nodes is an error, and
// never a panic.
func TestExpectedPastEnd(t *testing.T) {
	r, x := guacamole(t)
	c := run(t, r, x, "print(1);")
	if _, err := c.Expected(len(c.Sets) - 1); err != nil {
		t.Fatalf("Expected at the last node: %v", err)
	}
	for _, pos := range []int{len(c.Sets), len(c.Sets) + 10, -1} {
		if exp, err := c.Expected(pos); err == nil && len(exp) > 0 {
			t.Errorf("Expected(%d) past the lattice = %v, nil; want an error or an empty set", pos, exp)
		}
	}
	// A refused input's dead positions have no set: empty, not a panic.
	d := run(t, r, x, "1 2 3 4")
	for pos := range d.Sets {
		if _, err := d.Expected(pos); err != nil {
			t.Errorf("Expected(%d) on a refused input: %v", pos, err)
		}
	}
}

// TestInternSharesEqualSets: positions with equal content share one interned
// set (§10). A statement repeated after itself reaches the same sets.
func TestInternSharesEqualSets(t *testing.T) {
	r, x := guacamole(t)
	// A fourth statement, so the third ends before the same lookahead.
	c := run(t, r, x, "my $x = 1;\nmy $x = 1;\nmy $x = 1;\nmy $x = 1;\n")
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
	r, x := guacamole(t)
	c := run(t, r, x, "my $x = 1;\nprint($x);\n{ my $y = [1, 2]; }\n")
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
	r, x := guacamole(t)
	src := "my $x = 1;\nprint($x);\nsub f { return 2; }\n"
	first := run(t, r, x, src)
	second := run(t, r, x, src)
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
	warm, x := guacamole(t)
	srcs := []string{
		"my $x = 1;", "my $x = 1", "my $x =", "print(1);", "1 2;", "$x $y;",
		"sub f { my $x = 1; return $x; }", "{ my $x = 1; }", "my @a = (1, 2);", "my @a = (1, 2;",
	}
	hits := 0
	for range 2 {
		for _, src := range srcs {
			c := run(t, warm, x, src)
			cold := run(t, New(guacTab, Options{NoCache: true}), x, src)
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
	r, x := guacamole(t)
	for _, src := range []string{"my $x = 1;", "my $x =", "print(1)", "1 2", "{ 1; } 2;", "sub f { 1 }"} {
		la := run(t, r, x, src).Accepted()
		no := run(t, New(guacTab, Options{NoLookahead: true}), x, src).Accepted()
		if la != no {
			t.Errorf("%q: accepted %v with lookahead, %v without", src, la, no)
		}
	}
}

// TestExpectedAfterAssign: after `my $x =` a term is expected, and `;` is not.
func TestExpectedAfterAssign(t *testing.T) {
	r, x := guacamole(t)
	c := run(t, r, x, "my $x = 1;")
	exp, err := c.Expected(3)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(exp, "ScalarVar") || slices.Contains(exp, "Semicolon") {
		t.Errorf("Expected after `my $x =` = %v, want ScalarVar and no Semicolon", exp)
	}
}

// TestExpectedIgnoresLookahead: Expected(pos) is the union over the Earley
// set as recognition without lookahead builds it, at every node, refusal
// points included: after `my $x = 1` before `2`, lookahead prunes every
// completion, but an operator is still expected there.
func TestExpectedIgnoresLookahead(t *testing.T) {
	r, x := guacamole(t)
	for _, src := range []string{"my $x = 1 2;", "my $x = 1;\nprint($x + 2);\nsub f { return [1, 2]; }\n"} {
		la := run(t, r, x, src)
		no := run(t, New(guacTab, Options{NoLookahead: true}), x, src)
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
	c := run(t, r, x, "my $x = 1 2;")
	if exp, _ := c.Expected(4); !slices.Contains(exp, "OpAdd") {
		t.Errorf("Expected after `my $x = 1` = %v, want OpAdd among them", exp)
	}
}
