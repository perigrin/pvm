// ABOUTME: Tests recognition over a branched toy lattice: the paths rejoin where offset and state agree.
// ABOUTME: Inside the branch no set is interned and no transition cache entry is used; outside it both are.

package recognize

import (
	"slices"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/chart/lex"
	"tamarou.com/pvm/internal/lexer"
)

// toyTables is the tables of a toy grammar stated as Go data, one string per
// statement: "LHS ::= alt | alt", symbols space-separated, an empty
// alternative empty, and `X*` or `X+` a sequence of one symbol. A symbol no
// statement defines is a terminal.
// Its terminal-set table is the empty set at id 0, then one singleton set
// per terminal, in symbol order.
func toyTables(tb testing.TB, statements ...string) *grammar.Tables {
	tb.Helper()
	g := &grammar.Grammar{}
	defined := map[string]bool{}
	for _, st := range statements {
		lhs, rhs, _ := strings.Cut(st, "::=")
		lhs = strings.TrimSpace(lhs)
		defined[lhs] = true
		for _, alt := range strings.Split(rhs, "|") {
			r := grammar.Rule{LHS: lhs}
			for _, w := range strings.Fields(alt) {
				if base, ok := strings.CutSuffix(w, "*"); ok {
					r.Seq, w = grammar.Star, base
				} else if base, ok := strings.CutSuffix(w, "+"); ok {
					r.Seq, w = grammar.Plus, base
				}
				r.RHS = append(r.RHS, grammar.Symbol{Kind: grammar.Name, Text: w})
			}
			g.G1 = append(g.G1, r)
		}
	}
	seen := map[string]bool{}
	for _, r := range g.G1 {
		for _, sym := range r.RHS {
			if !defined[sym.Text] && !seen[sym.Text] {
				seen[sym.Text] = true
				g.G0 = append(g.G0, grammar.Rule{LHS: sym.Text})
			}
		}
	}
	tab, err := grammar.BuildTables(g)
	if err != nil {
		tb.Fatal(err)
	}
	tab.TerminalSets = [][]int{{}}
	for id, s := range tab.Symbols {
		if s.Terminal {
			tab.TerminalSets = append(tab.TerminalSets, []int{id})
		}
	}
	return tab
}

// termSet is the id of the singleton terminal set toyTables gave name, or 0,
// the empty set, where name is no terminal.
func termSet(tab *grammar.Tables, name string) int {
	return max(0, slices.IndexFunc(tab.TerminalSets, func(s []int) bool { return len(s) == 1 && tab.Symbols[s[0]].Name == name }))
}

// toyEdge is a token of term's singleton set from node from to node to.
func toyEdge(tab *grammar.Tables, from, to int, term string, start, end int) lex.Edge {
	return lex.Edge{From: from, To: to, Token: lex.Token{Token: lexer.Token{Start: start, End: end}, TerminalSet: termSet(tab, term)}}
}

// mustRecognize is Recognize(r, l), failing tb on a validation error.
func mustRecognize(tb testing.TB, r *Recognizer, l *lex.Lattice[lex.None]) *Chart {
	tb.Helper()
	c, err := Recognize(r, l)
	if err != nil {
		tb.Fatal(err)
	}
	return c
}

// toyLinear is one path as a linear lattice, a node per byte offset given.
func toyLinear(offsets []int, edges ...lex.Edge) *lex.Lattice[lex.None] {
	l := &lex.Lattice[lex.None]{Edges: edges}
	for _, off := range offsets {
		l.Nodes = append(l.Nodes, lex.Node[lex.None]{Offset: off})
	}
	return l
}

// TestRecognizeToyLatticeRejoin: on S ::= X C, X ::= A | B B, bytes 0-2
// read as one A or as two Bs, the kind of token branch the interleaved
// lexer will emit.
func TestRecognizeToyLatticeRejoin(t *testing.T) {
	tab := toyTables(t, "S ::= X C", "X ::= A | B B")
	// Nodes 0 (the fork), 1 (inside the B B path), 2 (byte 2), 3 (the end).
	l := toyLinear([]int{0, 1, 2, 3},
		toyEdge(tab, 0, 2, "A", 0, 2), toyEdge(tab, 0, 1, "B", 0, 1), toyEdge(tab, 1, 2, "B", 1, 2), toyEdge(tab, 2, 3, "C", 2, 3))
	i, _, err := lex.Rejoin([]lex.Node[lex.None]{l.Nodes[0], l.Nodes[2]}, l.Nodes[:3], 4)
	if err != nil || i != 1 {
		t.Fatalf("Rejoin = %d, %v; want the A path's node 1, lattice node 2", i, err)
	}
	const join, inside = 2, 1

	r := New(tab, Options{})
	first, second := mustRecognize(t, r, l), mustRecognize(t, r, l)
	for run, c := range []*Chart{first, second} {
		if !c.Accepted() {
			t.Fatalf("run %d: the branched lattice was refused", run)
		}
		if c.Sets[inside] == nil || c.Sets[inside].id != -1 || c.Interned[inside] || c.Cached[inside] {
			t.Errorf("run %d: the node inside the branch was interned or cached", run)
		}
		if c.Sets[join] == nil || c.Sets[join].id < 0 || c.Cached[join] {
			t.Errorf("run %d: the rejoin node's set is not interned, or was cached though two edges reach it", run)
		}
	}
	if !second.Cached[3] {
		t.Error("the node after the rejoin missed the cache on the second run: the cache is in use outside the branch")
	}

	// Both paths' items meet in the one set at the rejoin node: it is the
	// union of each path's own set there. Distances count lattice nodes, so
	// the A path's single node to the fork counts as 2 in the lattice.
	a := mustRecognize(t, r, toyLinear([]int{0, 2, 3}, toyEdge(tab, 0, 1, "A", 0, 2), toyEdge(tab, 1, 2, "C", 2, 3)))
	b := mustRecognize(t, r, toyLinear([]int{0, 1, 2, 3}, toyEdge(tab, 0, 1, "B", 0, 1), toyEdge(tab, 1, 2, "B", 1, 2), toyEdge(tab, 2, 3, "C", 2, 3)))
	want := slices.Clone(b.Sets[2].Items())
	for _, it := range a.Sets[1].Items() {
		if it.Dist > 0 {
			it.Dist++
		}
		if !slices.Contains(want, it) {
			want = append(want, it)
		}
	}
	got := first.Sets[join].Items()
	if len(got) != len(want) || slices.ContainsFunc(want, func(it Item) bool { return !slices.Contains(got, it) }) {
		t.Errorf("the rejoin node's set is %v, want the union of both paths' %v", got, want)
	}
	// The B B path alone interns the set the branch holds uninterned.
	if first.Sets[inside] == b.Sets[1] || !slices.Equal(first.Sets[inside].Items(), b.Sets[1].Items()) {
		t.Error("the node inside the branch should equal the B B path's set in content but not be the interned set")
	}
}

// TestCacheChecksOriginSets: two inputs reach the same set before the same
// token and lookahead, but completing E reads the set where E began, and
// that differs. `p a b x` is accepted; `q a b x` needs a second x. A cache
// keyed on (set, token, lookahead) alone would accept both.
func TestCacheChecksOriginSets(t *testing.T) {
	tab := toyTables(t, "S ::= P E X | Q E X X", "E ::= A B")
	line := func(terms ...string) *lex.Lattice[lex.None] {
		l := &lex.Lattice[lex.None]{Nodes: []lex.Node[lex.None]{{}}}
		for i, term := range terms {
			l.Nodes = append(l.Nodes, lex.Node[lex.None]{Offset: i + 1})
			l.Edges = append(l.Edges, toyEdge(tab, i, i+1, term, i, i+1))
		}
		return l
	}
	r := New(tab, Options{})
	p := mustRecognize(t, r, line("P", "A", "B", "X"))
	q := mustRecognize(t, r, line("Q", "A", "B", "X"))
	if !p.Accepted() || q.Accepted() {
		t.Errorf("p a b x accepted %v, q a b x accepted %v; want true, false", p.Accepted(), q.Accepted())
	}
	if q.Sets[2] != p.Sets[2] || q.Cached[3] {
		t.Errorf("node 2 shared %v, node 3 cached %v; want the same set before b, and no cache hit after it", q.Sets[2] == p.Sets[2], q.Cached[3])
	}
}

// TestCacheNotKeyedOnBranchSet: a node with one in-edge from a node inside
// a branch reads a set that is not interned. Its result must not be cached:
// every branch set has id -1, so a key on it would serve the B lattice's
// set to the D lattice, which needs a further X. Warm and cold must agree.
func TestCacheNotKeyedOnBranchSet(t *testing.T) {
	tab := toyTables(t, "S ::= B T | D T X", "S ::= Z")
	// 0 -first-> 1 -T-> 3, and a path 0 -Z-> 2 that goes nowhere, which puts
	// node 1 inside a branch.
	branched := func(first string) *lex.Lattice[lex.None] {
		return toyLinear([]int{0, 1, 2, 3}, toyEdge(tab, 0, 1, first, 0, 1), toyEdge(tab, 1, 3, "T", 1, 2), toyEdge(tab, 0, 2, "Z", 0, 1))
	}
	warm := New(tab, Options{})
	mustRecognize(t, warm, branched("B"))
	d := mustRecognize(t, warm, branched("D"))
	cold := mustRecognize(t, New(tab, Options{NoCache: true}), branched("D"))
	if d.Accepted() != cold.Accepted() || d.Cached[3] {
		t.Errorf("d t: accepted %v warm, %v cold; cached at node 3 %v, want no cache hit", d.Accepted(), cold.Accepted(), d.Cached[3])
	}
}
