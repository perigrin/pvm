// ABOUTME: Tests the grammar-only tables New builds once per tables value and shares between its recognizers.
// ABOUTME: Shared read-only and race-free, extended for sets added later, and never shared across tables values.

package recognize

import (
	"fmt"
	"slices"
	"sync"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/chart/lex"
)

// BenchmarkNew is the allocations of one New over the committed tables,
// after one warm-up New.
func BenchmarkNew(b *testing.B) {
	tab := committedTables(b)
	New(tab, Options{})
	b.ReportAllocs()
	for b.Loop() {
		New(tab, Options{})
	}
}

// ownTables is a copy of tab whose set table this test may append to.
func ownTables(tab *grammar.Tables) *grammar.Tables {
	c := *tab
	c.TerminalSets = slices.Clone(tab.TerminalSets)
	return &c
}

// outcome is a recognition's verdict and each node's items, to compare two.
func outcome(tb testing.TB, r *Recognizer, l *lex.Lattice[lex.None]) string {
	tb.Helper()
	c := mustRecognize(tb, r, l)
	items := make([][]Item, len(c.Sets))
	for i, s := range c.Sets {
		items[i] = s.Items()
	}
	return fmt.Sprint(c.Accepted(), items)
}

// withSet is l with every token of set id from given set id to instead.
func withSet(l *lex.Lattice[lex.None], from, to int) *lex.Lattice[lex.None] {
	c := &lex.Lattice[lex.None]{Nodes: l.Nodes, Edges: slices.Clone(l.Edges)}
	for i := range c.Edges {
		if c.Edges[i].Token.TerminalSet == from {
			c.Edges[i].Token.TerminalSet = to
		}
	}
	return c
}

// TestRecognizersShareGrammarTables: two recognizers over one tables value
// share its grammar-only tables, recognize the corpus sample in parallel
// goroutines exactly as one each alone does, and both see a set added to
// the tables after both were built. Run under -race.
func TestRecognizersShareGrammarTables(t *testing.T) {
	tab := ownTables(committedTables(t))
	var files []*lex.Lattice[lex.None]
	for _, p := range samplePaths(t) {
		s, _ := lex.Lex(readFile(t, p), tab.Tokens)
		files = append(files, lex.Linear(s))
	}
	want := make([]string, len(files))
	for i, l := range files {
		want[i] = outcome(t, New(ownTables(tab), Options{}), l)
	}
	rs := []*Recognizer{New(tab, Options{}), New(tab, Options{})}
	if rs[0].g != rs[1].g {
		t.Fatal("two recognizers over one tables value built their grammar-only tables twice")
	}
	var wg sync.WaitGroup
	for _, r := range rs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i, l := range files {
				if got := outcome(t, r, l); got != want[i] {
					t.Errorf("file %d, recognized beside another recognizer, differs from one alone", i)
				}
			}
		}()
	}
	wg.Wait()

	s, _ := lex.Lex([]byte("print 1;\n"), tab.Tokens)
	l := lex.Linear(s)
	semi := l.Edges[len(l.Edges)-1].Token.TerminalSet
	tab.TerminalSets = append(tab.TerminalSets, tab.TerminalSets[semi])
	added := withSet(l, semi, len(tab.TerminalSets)-1)
	for i, r := range rs {
		if got, want := outcome(t, r, added), outcome(t, r, l); got != want {
			t.Errorf("recognizer %d: the ; as a set added after New reads %s, as its own set %s", i, got, want)
		}
	}
}

// TestRecognizersOverDistinctTablesShareNothing: recognizers over two tables
// values, equal in content or one a copy of the other made after New, share
// no grammar-only tables, and a set added to one is outside the other's table.
func TestRecognizersOverDistinctTablesShareNothing(t *testing.T) {
	a := toyStmt(t)
	ra := New(a, Options{})
	for name, b := range map[string]*grammar.Tables{"built apart": toyStmt(t), "copied": ownTables(a)} {
		if New(b, Options{}).g == ra.g {
			t.Errorf("recognizers over two tables values, %s, share their grammar-only tables", name)
		}
	}
	b := ownTables(a)
	rb := New(b, Options{})
	semi := addSet(a, "Semi")
	l := setLattice(termSet(a, "Print"), termSet(a, "Num"), semi)
	if !mustRecognize(t, ra, l).Accepted() {
		t.Errorf("Print Num and set %d, added to its own tables after New, refused", semi)
	}
	if _, err := Recognize(rb, l); err == nil {
		t.Errorf("set %d, added to the other tables value, recognized without an error", semi)
	}
}

// TestRecognizerStateNotShared: the intern table, the cache trie and the
// counters stay per recognizer: one recognizer's are unchanged by another
// recognizing different input over the same tables.
func TestRecognizerStateNotShared(t *testing.T) {
	tab := toyStmt(t)
	r1, r2 := New(tab, Options{}), New(tab, Options{})
	state := func(r *Recognizer) string {
		return fmt.Sprint(len(r.sets.sets), len(r.sets.cores), len(r.sets.dists), len(r.cache), len(r.trie), r.stored, r.longestScan, r.expectedComputed)
	}
	c := run(t, r1, decl+" "+decl)
	if _, err := c.Expected(2); err != nil {
		t.Fatal(err)
	}
	before := state(r1)
	c = run(t, r2, "Print Var Semi Print Num Semi "+decl)
	if _, err := c.Expected(3); err != nil {
		t.Fatal(err)
	}
	if after := state(r1); after != before {
		t.Errorf("recognizer 1's state went from %s to %s while recognizer 2 recognized", before, after)
	}
	if state(r2) == state(New(tab, Options{})) {
		t.Error("recognizer 2's state did not change as it recognized")
	}
}

// TestRecognizersExtendSharedTablesConcurrently: two recognizers built before
// a set is added each extend the shared part for it at once, race-free under
// -race, and agree with a recognizer built after it was added.
func TestRecognizersExtendSharedTablesConcurrently(t *testing.T) {
	tab := toyStmt(t)
	prt, num := termSet(tab, "Print"), termSet(tab, "Num")
	for range 20 {
		rs := []*Recognizer{New(tab, Options{}), New(tab, Options{})}
		semi := addSet(tab, "Semi", "RBrace")
		l := setLattice(prt, num, semi)
		got := make([]string, len(rs))
		start := make(chan struct{})
		var wg sync.WaitGroup
		for i, r := range rs {
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-start
				got[i] = outcome(t, r, l)
			}()
		}
		close(start)
		wg.Wait()
		want := outcome(t, New(ownTables(tab), Options{}), l)
		for i, g := range got {
			if g != want {
				t.Errorf("set %d: recognizer %d read %s, one built after the set was added %s", semi, i, g, want)
			}
		}
	}
}

// packageTables is a tables value the linker allocates, not the heap.
var packageTables grammar.Tables

// TestNewOverAnyTablesValue: New and Recognize work over a tables value
// wherever it lives, a local variable's or a package variable's, not only
// one the heap holds.
func TestNewOverAnyTablesValue(t *testing.T) {
	var local grammar.Tables
	for name, tab := range map[string]*grammar.Tables{"local": &local, "package": &packageTables} {
		*tab = *toyStmt(t)
		if !run(t, New(tab, Options{}), "Print Num Semi").Accepted() {
			t.Errorf("%s tables value: Print Num Semi refused", name)
		}
	}
}
