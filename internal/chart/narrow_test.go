// ABOUTME: Tests narrowing before recognition: a library operator's dead PLUGIN_* readings never reach the lattice.
// ABOUTME: A kept operator's set takes an id the committed tables lack, fixed for the tables' life, written nowhere shared.

package chart

import (
	"slices"
	"strings"
	"sync"
	"testing"

	"tamarou.com/pvm/internal/chart/forest"
	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/chart/lex"
	"tamarou.com/pvm/internal/chart/recognize"
)

// narrowTables is the committed tables, failing the test where they do not load.
func narrowTables(t *testing.T) *grammar.Tables {
	t.Helper()
	tab, err := committed()
	if err != nil {
		t.Fatal(err)
	}
	return tab
}

// isPlugin is whether symbol x is a PLUGIN_* terminal.
func isPlugin(tab *grammar.Tables, x int) bool {
	return tab.Symbols[x].Terminal && strings.HasPrefix(tab.Symbols[x].Name, "PLUGIN_")
}

// narrowedTerminals is each edge's terminals of src's lattice narrowed under scope, read through a chart of it.
func narrowedTerminals(t *testing.T, tab *grammar.Tables, src string, scope forest.Scope) (*lex.Lattice[lex.None], [][]int) {
	t.Helper()
	s, _ := lex.Lex([]byte(src), tab.Tokens)
	l := forest.Narrow(tab, s, []byte(src), scope)
	c, err := recognize.Recognize(recognize.New(tab, recognize.Options{}), l)
	if err != nil {
		t.Fatal(err)
	}
	var terms [][]int
	for _, e := range l.Edges {
		terms = append(terms, c.Terminals(e.Token.TerminalSet))
	}
	return l, terms
}

// TestNarrowBeforeRecognition: with no library operator in scope, no PLUGIN_* reading of `plus` reaches the lattice
// Narrow gives the recognizer, nor any node of production's forest before resolve, where the un-narrowed lattice and
// V1's forest, which prunes after Build, both have one.
func TestNarrowBeforeRecognition(t *testing.T) {
	tab := narrowTables(t)
	const src = "foo plus 2;"
	s, _ := lex.Lex([]byte(src), tab.Tokens)
	if !slices.ContainsFunc(tab.TerminalSets[s.Tokens[1].TerminalSet], func(x int) bool { return isPlugin(tab, x) }) {
		t.Fatalf("%q: plus's un-narrowed set has no PLUGIN_* reading, so the test proves nothing", src)
	}
	_, terms := narrowedTerminals(t, tab, src, forest.Scope{})
	for k, set := range terms {
		if slices.ContainsFunc(set, func(x int) bool { return isPlugin(tab, x) }) {
			t.Errorf("%q: token %d keeps a PLUGIN_* reading on the narrowed lattice: %v", src, k, set)
		}
	}
	pluginNode := func(f *forest.Forest) bool {
		return slices.ContainsFunc(f.Nodes[1:], func(n forest.Node) bool { return n.Sym >= 0 && isPlugin(tab, int(n.Sym)) })
	}
	cand, ref := forest.CandidateForest(tab, []byte(src), forest.Scope{})
	if cand == nil {
		t.Fatalf("%q: production refuses it (%+v)", src, ref)
	}
	if pluginNode(cand) {
		t.Errorf("%q: production's forest has a PLUGIN_* node: the reading reached recognition", src)
	}
	if v1, _ := forest.V1Forest(tab, []byte(src), forest.Scope{}); v1 == nil || !pluginNode(v1) {
		t.Errorf("%q: V1's forest has no PLUGIN_* node, so the check above proves nothing", src)
	}
}

// TestNarrowLeavesCommittedTerminalSetsUnmutated: two recognizers over the shared committed tables parse files that
// keep different library operators at once, so each narrows to sets the tables lack; afterwards the tables still
// hold their 1,880 sets, each as it was. Run under -race, a write to what both read is a failure.
func TestNarrowLeavesCommittedTerminalSetsUnmutated(t *testing.T) {
	tab := narrowTables(t)
	before := make([][]int, len(tab.TerminalSets))
	for i, set := range tab.TerminalSets {
		before[i] = slices.Clone(set)
	}
	srcs := []string{"use My::Ops; 1 plus 2;", "use My::Ops; 1 mul 2;", "use My::Ops; 1 pow 2;", "use My::Ops; 1 lowop 2;"}
	scopes := make([]forest.Scope, len(srcs))
	for i, src := range srcs {
		scopes[i] = libopsScope(t, src)
	}
	var wg sync.WaitGroup
	for w := range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r := recognize.New(tab, recognize.Options{})
			for i := range srcs {
				i = (i + w) % len(srcs) // the two meet each set first in a different order
				if f, ref := forest.ParseWith(tab, r, []byte(srcs[i]), scopes[i]); f == nil {
					t.Errorf("%q refused: %+v", srcs[i], ref)
				}
			}
		}()
	}
	wg.Wait()
	if len(tab.TerminalSets) != 1880 || !slices.EqualFunc(tab.TerminalSets, before, slices.Equal) {
		t.Errorf("the committed tables hold %d sets after narrowing, or one changed; want the 1880 as they were", len(tab.TerminalSets))
	}
}

// TestNarrowWithoutLibraryOperatorsIsIdentity: a file no token of which admits a PLUGIN_* reading narrows to its own
// lattice, and recognizes to the same sets the un-narrowed lattice gives.
func TestNarrowWithoutLibraryOperatorsIsIdentity(t *testing.T) {
	tab := narrowTables(t)
	const src = "my $x = 1;\nprint $x + 2;\n"
	s, _ := lex.Lex([]byte(src), tab.Tokens)
	for _, tk := range s.Tokens {
		if slices.ContainsFunc(tab.TerminalSets[tk.TerminalSet], func(x int) bool { return isPlugin(tab, x) }) {
			t.Fatalf("%q has a PLUGIN_* reading, so it is no fixture for the identity", src)
		}
	}
	l, un := forest.Narrow(tab, s, []byte(src), forest.Scope{}), lex.Linear(s)
	if !slices.Equal(l.Edges, un.Edges) {
		t.Errorf("%q narrows to %v, want its own lattice %v", src, l.Edges, un.Edges)
	}
	nc, err := recognize.Recognize(recognize.New(tab, recognize.Options{}), l)
	if err != nil {
		t.Fatal(err)
	}
	uc, err := recognize.Recognize(recognize.New(tab, recognize.Options{}), un)
	if err != nil {
		t.Fatal(err)
	}
	if !nc.Accepted() || !slices.EqualFunc(nc.Sets, uc.Sets, func(a, b *recognize.Set) bool { return slices.Equal(a.Items(), b.Items()) }) {
		t.Errorf("%q: the narrowed chart's sets differ from the un-narrowed one's, or it refuses", src)
	}
}

// TestNarrowKeptLibraryOperator: with My::Ops in scope, `plus` keeps its one PLUGIN_* reading, PLUGIN_ADD_OP, through
// a set id the committed tables lack (C12: the corpus keeps none), and the file parses as V1 parses it.
func TestNarrowKeptLibraryOperator(t *testing.T) {
	tab := narrowTables(t)
	const src = "use My::Ops; 1 plus 2;"
	scope := libopsScope(t, src)
	l, terms := narrowedTerminals(t, tab, src, scope)
	k := 4 // use My::Ops ; 1 plus
	id, set := l.Edges[k].Token.TerminalSet, terms[k]
	var plugins []string
	for _, x := range set {
		if isPlugin(tab, x) {
			plugins = append(plugins, tab.Symbols[x].Name)
		}
	}
	t.Logf("plus: set id %d, PLUGIN_* readings %v", id, plugins)
	if !slices.Equal(plugins, []string{"PLUGIN_ADD_OP"}) {
		t.Errorf("plus keeps %v, want [PLUGIN_ADD_OP]", plugins)
	}
	if id < len(tab.TerminalSets) || slices.ContainsFunc(tab.TerminalSets, func(c []int) bool { return slices.Equal(c, set) }) {
		t.Errorf("plus's set id %d, %v: want an id past the committed %d, for a set they lack", id, set, len(tab.TerminalSets))
	}
	f, ref := forest.Parse(tab, []byte(src), scope)
	if f == nil {
		t.Fatalf("%q refused: %+v", src, ref)
	}
	cand, _ := forest.CandidateForest(tab, []byte(src), scope)
	v1, _ := forest.V1Forest(tab, []byte(src), scope)
	if d := forest.ForestDiff(cand, v1); d != "" {
		t.Errorf("%q: production's forest differs from V1's: %s", src, d)
	}
}
