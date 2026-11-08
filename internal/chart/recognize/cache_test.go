// ABOUTME: Tests the transition cache and the Expected memo it holds: bounded per lookup as flat statements grow.
// ABOUTME: A lookup returns the first stored entry whose origin sets hold, as a linear scan of the entries would.

package recognize

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/chart/lex"
)

// committedTables is the committed grammar's tables.
func committedTables(t *testing.T) *grammar.Tables {
	t.Helper()
	g, ref, err := grammar.Committed()
	if err != nil {
		t.Fatal(err)
	}
	tab, err := grammar.Generate(g, ref)
	if err != nil {
		t.Fatal(err)
	}
	return tab
}

// flat is n top-level statements in the shape of T1/code_too_large.t, each
// starting at a set of its own, so each completion reads a different one.
func flat(n int) []byte {
	var b strings.Builder
	for i := 1; i <= n; i++ {
		fmt.Fprintf(&b, "$v = (\" \" / \"%d\"); print \"not \" if $v ne q{0}; print 'ok %d';\n", i, i)
	}
	return []byte(b.String())
}

// TestCacheLookupBounded: the most candidates one lookup examines does not
// grow from 100 flat statements to 400. A key's entries grew by one per
// statement, each differing only in the statement's start set, and every
// lookup scanned them all: quadratic in the statement count.
func TestCacheLookupBounded(t *testing.T) {
	tab := committedTables(t)
	scan := func(n int) int {
		r := New(tab, Options{})
		s, _ := lex.Lex(flat(n), tab.Tokens)
		if c := mustRecognize(t, r, lex.Linear(s)); !c.Accepted() {
			t.Fatalf("%d flat statements are refused", n)
		}
		return r.longestScan
	}
	small, big := scan(100), scan(400)
	t.Logf("the longest scan of one lookup: %d at 100 statements, %d at 400", small, big)
	if small == 0 {
		t.Fatal("no lookup examined a candidate: the cache was not used")
	}
	if big > small {
		t.Errorf("the longest scan of one lookup grew from %d at 100 statements to %d at 400", small, big)
	}
}

// TestCacheMatchesNoCache: a lookup returns the first stored entry whose
// origin sets hold, whatever its dependencies' shape; and over the corpus
// sample's whole files and flat statements, one warm recognizer agrees with
// no cache on acceptance and on Expected at every node.
func TestCacheMatchesNoCache(t *testing.T) {
	c := &Chart{Sets: []*Set{{id: 5}, {id: 6}, {id: 7}, {id: 8}}}
	r := New(toyStmt(t), Options{})
	key := cacheKey{set: 7, span: 1}
	entries := []struct {
		deps   []dep
		result *Set
	}{
		{deps: []dep{{1, 99}}, result: &Set{id: 1}},        // fails at node 3
		{deps: []dep{{1, 7}, {2, 6}}, result: &Set{id: 2}}, // the first that holds
		{deps: []dep{{1, 7}}, result: &Set{id: 3}},         // holds, but stored later
	}
	for _, e := range entries {
		r.store(key, e.deps, e.result)
	}
	if got := r.lookup(c, 3, key); got != entries[1].result {
		t.Errorf("lookup returned set %v, want the first entry whose dependencies hold, set 2", got)
	}
	warmMatchesNoCache(t)
}

// warmMatchesNoCache: over the corpus sample's whole files and flat
// statements, one warm recognizer, whose cache holds the Expected memo,
// agrees with no cache on acceptance and on Expected at every node.
func warmMatchesNoCache(t *testing.T) {
	tab := committedTables(t)
	warm := New(tab, Options{})
	srcs := [][]byte{flat(50)}
	for _, p := range samplePaths(t) {
		srcs = append(srcs, readFile(t, p))
	}
	srcs = append(srcs, flat(50)) // the second time through a warm cache
	for i, src := range srcs {
		s, _ := lex.Lex(src, tab.Tokens)
		l := lex.Linear(s)
		on, off := mustRecognize(t, warm, l), mustRecognize(t, New(tab, Options{NoCache: true}), l)
		if on.Accepted() != off.Accepted() {
			t.Errorf("source %d: accepted %v with the cache, %v without", i, on.Accepted(), off.Accepted())
		}
		for pos := range on.Sets {
			if a, b := expectedOf(t, on, pos), expectedOf(t, off, pos); !slices.Equal(a, b) {
				t.Errorf("source %d node %d: Expected %v with the cache, %v without", i, pos, a, b)
				break
			}
		}
	}
}

// TestExpectedMemoBounded: the most entries one Expected's memo lookup
// examines does not grow from 100 flat statements to 400. The memo kept a
// list per set, which grew by one entry per statement, and every lookup
// scanned it: quadratic in the statement count over a file's nodes.
func TestExpectedMemoBounded(t *testing.T) {
	tab := committedTables(t)
	scan := func(n int) int {
		r := New(tab, Options{})
		s, _ := lex.Lex(flat(n), tab.Tokens)
		c := mustRecognize(t, r, lex.Linear(s))
		if !c.Accepted() {
			t.Fatalf("%d flat statements are refused", n)
		}
		r.longestScan = 0 // count the memo's lookups only
		for pos := range c.Sets {
			expectedOf(t, c, pos)
		}
		return r.longestScan
	}
	small, big := scan(100), scan(400)
	t.Logf("the longest scan of one Expected memo lookup: %d at 100 statements, %d at 400", small, big)
	if small == 0 {
		t.Fatal("no Expected lookup examined a candidate: the memo was not used")
	}
	if big > small {
		t.Errorf("the longest scan of one Expected memo lookup grew from %d at 100 statements to %d at 400", small, big)
	}
}

// TestExpectedMemoMatchesNoCache: a memo lookup returns the first stored
// Expected whose origin sets hold; and over the corpus sample's whole
// files and flat statements, one warm recognizer's memo answers Expected
// at every node as no memo does.
func TestExpectedMemoMatchesNoCache(t *testing.T) {
	c := &Chart{Sets: []*Set{{id: 5}, {id: 6}, {id: 7}, {id: 8}}}
	r := New(toyStmt(t), Options{})
	key := cacheKey{set: 8}
	for _, e := range []struct {
		deps  []dep
		names []string
	}{
		{deps: []dep{{1, 99}}, names: []string{"a"}},        // fails at node 3
		{deps: []dep{{1, 7}, {2, 6}}, names: []string{"b"}}, // the first that holds
		{deps: []dep{{1, 7}}, names: []string{"c"}},         // holds, but stored later
	} {
		r.slot(key, e.deps).names = e.names
	}
	var got []string
	if n := r.hit(c, 3, key); n != nil {
		got = n.names
	}
	if !slices.Equal(got, []string{"b"}) {
		t.Errorf("memo lookup returned %v, want the first entry whose dependencies hold, [b]", got)
	}

	warmMatchesNoCache(t)
}
