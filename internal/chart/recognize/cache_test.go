// ABOUTME: Tests the transition cache and the Expected memo it holds: bounded per lookup as flat statements grow.
// ABOUTME: A lookup returns the first stored entry whose origin sets hold, as a linear scan of the entries would.

package recognize

import (
	"fmt"
	"math/rand"
	"slices"
	"strings"
	"sync"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/chart/lex"
)

// committedTables is the committed grammar's tables, generated once for the package's tests, none of which writes them.
func committedTables(t *testing.T) *grammar.Tables {
	t.Helper()
	committed.Do(func() {
		g, ref, err := grammar.Committed()
		if err == nil {
			committed.tab, err = grammar.Generate(g, ref)
		}
		committed.err = err
	})
	if committed.err != nil {
		t.Fatal(committed.err)
	}
	return committed.tab
}

var committed struct {
	sync.Once
	tab *grammar.Tables
	err error
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
// Expected whose origin sets hold. The corpus walk that holds the memo to
// no memo is TestCacheMatchesNoCache's, run once.
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
}

// TestCacheTrieRandomLookups: about 60k seeded lookups, between random
// stores, each answered by the trie as by a linear scan of the key's
// entries in store order for the first whose dependencies hold.
func TestCacheTrieRandomLookups(t *testing.T) {
	type entry struct {
		deps   []dep
		result *Set
	}
	rng := rand.New(rand.NewSource(1))
	r := New(toyStmt(t), Options{})
	ref := map[cacheKey][]entry{}
	c := &Chart{Sets: make([]*Set, 12)}
	lookups, hits := 0, 0
	for lookups < 60000 {
		key := cacheKey{set: rng.Intn(3), span: 1 + rng.Intn(2), tok: rng.Intn(8), la: rng.Intn(8)}
		if rng.Intn(20) == 0 {
			var deps []dep
			for range 1 + rng.Intn(4) {
				d := dep{dist: 1 + rng.Intn(5), set: rng.Intn(3)}
				if !slices.Contains(deps, d) {
					deps = append(deps, d)
				}
			}
			e := entry{deps, &Set{id: len(ref)}}
			ref[key] = append(ref[key], e)
			r.store(key, e.deps, e.result)
			continue
		}
		for i := range c.Sets {
			c.Sets[i] = nil
			if rng.Intn(8) != 0 {
				c.Sets[i] = &Set{id: rng.Intn(3)}
			}
		}
		pos := 4 + rng.Intn(len(c.Sets)-4)
		var want *Set
		for _, e := range ref[key] {
			holds := true
			for _, d := range e.deps {
				// The source set is the key's own, so the trie does not store it.
				if d.dist == key.span && d.set == key.set {
					continue
				}
				if at := pos - d.dist; at < 0 || c.Sets[at] == nil || c.Sets[at].id != d.set {
					holds = false
				}
			}
			if holds {
				want = e.result
				break
			}
		}
		if got := r.lookup(c, pos, key); got != want {
			t.Fatalf("lookup %d of %+v at node %d: set %v from the trie, %v from the scan", lookups, key, pos, got, want)
		}
		lookups++
		if want != nil {
			hits++
		}
	}
	if hits == 0 || hits == lookups {
		t.Errorf("%d of %d lookups hit: want both hits and misses", hits, lookups)
	}
	t.Logf("%d of %d lookups hit", hits, lookups)
}

// TestExpectedMemoRandomLookups: about 104k seeded Expected lookups, over
// random inputs to one warm recognizer, each answered by the memo as by a
// recognizer with no cache, which computes every answer.
func TestExpectedMemoRandomLookups(t *testing.T) {
	tab := toyStmt(t)
	warm := New(tab, Options{})
	rng := rand.New(rand.NewSource(1))
	lookups := 0
	for lookups < 104000 {
		b := make([]byte, rng.Intn(16))
		for i := range b {
			b[i] = byte(rng.Intn(len(stmtTerms)))
		}
		l := fuzzLattice(tab, string(b))
		c, cold := mustRecognize(t, warm, l), mustRecognize(t, New(tab, Options{NoCache: true}), l)
		for pos := range c.Sets {
			if got, want := expectedOf(t, c, pos), expectedOf(t, cold, pos); !slices.Equal(got, want) {
				t.Fatalf("%q node %d: Expected %v from the memo, %v computed", b, pos, got, want)
			}
			lookups++
		}
	}
	if warm.expectedComputed >= lookups {
		t.Errorf("the memo computed %d of %d lookups: it served none", warm.expectedComputed, lookups)
	}
}
