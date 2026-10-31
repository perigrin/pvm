// ABOUTME: FuzzRecognize, the recognizer's fuzz target: it never panics, and acceptance does not depend on the cache.
// ABOUTME: Nor on lookahead, Leo or a dead-end branch, and Expected ignores lookahead. Run by `make fuzz-chart-recognize`.

package recognize

import (
	"slices"
	"strings"
	"sync"
	"testing"

	"tamarou.com/pvm/internal/chart/lex"
)

// withDeadEnd is l with a path that goes nowhere: edge i's token again, from
// node i to a new node placed after node i+1, so node i+1 and the new node
// are inside a branch. The last node is unchanged, so acceptance is too.
// l needs at least two edges, and i at most len(l.Edges)-2.
func withDeadEnd(l *lex.Lattice[lex.None], i int) *lex.Lattice[lex.None] {
	at := i + 2
	shift := func(n int) int {
		if n >= at {
			return n + 1
		}
		return n
	}
	out := &lex.Lattice[lex.None]{Nodes: slices.Insert(slices.Clone(l.Nodes), at, l.Nodes[i+1])}
	for _, e := range l.Edges {
		out.Edges = append(out.Edges, lex.Edge{From: shift(e.From), To: shift(e.To), Token: e.Token})
	}
	out.Edges = append(out.Edges, lex.Edge{From: i, To: at, Token: l.Edges[i].Token})
	return out
}

// stmtTerms are stmtGrammar's terminals; a fuzz input byte b is stmtTerms[b % 9].
var stmtTerms = []string{"My", "Var", "Assign", "Num", "Semi", "Print", "Add", "LBrace", "RBrace"}

// fuzzLattice is src, a byte to a terminal, as a linear lattice, so every
// input the fuzzer produces is a valid terminal string.
func fuzzLattice(src string) *lex.Lattice[lex.None] {
	terms := make([]string, len(src))
	for i := range src {
		terms[i] = stmtTerms[int(src[i])%len(stmtTerms)]
	}
	return tokens(terms...)
}

// enc is the fuzz input for the terminal names in words.
func enc(words string) string {
	var b []byte
	for _, w := range strings.Fields(words) {
		i := slices.Index(stmtTerms, w)
		if i < 0 {
			panic("enc: no terminal " + w)
		}
		b = append(b, byte(i))
	}
	return string(b)
}

// TestFuzzInputsAreTerminalStrings: whatever bytes the fuzzer produces, each
// token of the lattice names a terminal of stmtGrammar, so no mutation is
// refused at its first token for want of a terminal.
func TestFuzzInputsAreTerminalStrings(t *testing.T) {
	tab := toyStmt(t)
	term := map[string]bool{}
	for _, s := range tab.Symbols {
		term[s.Name] = s.Terminal
	}
	for _, src := range append([]string{"Print Num SQmi Semi", "0A", "\xff\x00 z"}, fuzzSeeds...) {
		l := fuzzLattice(src)
		if len(l.Edges) != len(src) {
			t.Errorf("%q: %d tokens, want one per byte", src, len(l.Edges))
		}
		for _, e := range l.Edges {
			if !term[e.Token.Terminals[0]] {
				t.Errorf("%q: token %q is no terminal", src, e.Token.Terminals[0])
			}
		}
	}
}

// fuzzSeeds is FuzzRecognize's seed corpus, some accepted and some refused.
var fuzzSeeds = []string{
	"", enc(decl), enc("Print Num"), enc("Num Num"), enc("Var Var"), enc("My Var Assign"),
	enc(decl + " Print Var Add Num Semi"),
	enc("LBrace Print Num Semi LBrace " + decl + " RBrace RBrace"),
	enc("LBrace LBrace Print Num Semi RBrace"),
}

func FuzzRecognize(f *testing.F) {
	for _, s := range fuzzSeeds {
		f.Add(s)
	}
	tab := toyStmt(f)
	// warm is shared across inputs, so its cache serves earlier inputs'
	// transitions to later ones.
	var mu sync.Mutex
	warm := New(tab, Options{})
	f.Fuzz(func(t *testing.T, src string) {
		lattices := []*lex.Lattice[lex.None]{fuzzLattice(src)}
		if n := len(lattices[0].Edges); n >= 2 {
			lattices = append(lattices, withDeadEnd(lattices[0], len(src)%(n-1)))
		}
		cold := Recognize(New(tab, Options{NoCache: true}), lattices[0])
		want := cold.Accepted()
		no := Recognize(New(tab, Options{NoCache: true, NoLookahead: true}), lattices[0])
		if no.Accepted() != want {
			t.Fatalf("%q: accepted %v without lookahead, %v with", src, no.Accepted(), want)
		}
		for _, la := range []bool{false, true} {
			if got := Recognize(New(tab, Options{NoCache: true, NoLookahead: la, NoLeo: true}), lattices[0]).Accepted(); got != want {
				t.Fatalf("%q: accepted %v without Leo (lookahead off %v), %v with", src, got, la, want)
			}
		}
		for pos := range cold.Sets {
			a, _ := cold.Expected(pos)
			b, _ := no.Expected(pos)
			if !slices.Equal(a, b) {
				t.Fatalf("%q node %d: Expected %v with lookahead, %v without", src, pos, a, b)
			}
		}
		mu.Lock()
		defer mu.Unlock()
		for i, l := range lattices {
			if got := Recognize(New(tab, Options{NoCache: true}), l).Accepted(); got != want {
				t.Fatalf("%q lattice %d: accepted %v cold, want %v", src, i, got, want)
			}
			for range 2 {
				c := Recognize(warm, l)
				if c.Accepted() != want {
					t.Fatalf("%q lattice %d: accepted %v with the cache, %v without", src, i, c.Accepted(), want)
				}
				for pos := -1; pos <= len(c.Sets); pos++ {
					_, _ = c.Expected(pos)
				}
			}
		}
	})
}

// TestFuzzSeedsAcceptAndRefuse: the seeds hold an accepted and a refused
// input, and the dead-end lattice keeps each verdict, so FuzzRecognize cannot
// pass by never reaching a refusal.
func TestFuzzSeedsAcceptAndRefuse(t *testing.T) {
	tab := toyStmt(t)
	var accepted, refused int
	for _, src := range fuzzSeeds {
		l := fuzzLattice(src)
		want := Recognize(New(tab, Options{}), l).Accepted()
		if want {
			accepted++
		} else {
			refused++
		}
		if len(l.Edges) < 2 {
			continue
		}
		d := withDeadEnd(l, 0)
		if len(d.Nodes) != len(l.Nodes)+1 {
			t.Errorf("%q: the dead end added %d nodes, want 1", src, len(d.Nodes)-len(l.Nodes))
		}
		if got := Recognize(New(tab, Options{}), d).Accepted(); got != want {
			t.Errorf("%q: accepted %v with a dead end, %v without", src, got, want)
		}
	}
	if accepted == 0 || refused == 0 {
		t.Errorf("%d accepted and %d refused seeds, want both", accepted, refused)
	}
}
