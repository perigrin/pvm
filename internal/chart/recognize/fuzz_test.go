// ABOUTME: FuzzRecognize, the recognizer's fuzz target: it never panics, and acceptance does not depend on the cache.
// ABOUTME: Nor on lookahead or a dead-end branch, and Expected ignores lookahead. Run by `make fuzz-chart-recognize`.

package recognize

import (
	"slices"
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

// fuzzSeeds is FuzzRecognize's seed corpus.
var fuzzSeeds = []string{
	"", "my $x = 1;", "print(1)", "1 2", "$x $y", "my $x =",
	"sub f { my ($a) = @_; return $a + 1; }\nf(2);\n",
	"my @a = map { $_ * 2 } grep { defined($_) } @b;",
	"if ($x) { 1 } elsif ($y) { 2 } else { 3 }",
	"=pod\n\nx\n\n=cut\n",
}

func FuzzRecognize(f *testing.F) {
	for _, s := range fuzzSeeds {
		f.Add(s)
	}
	// warm is shared across inputs, so its cache serves earlier inputs'
	// transitions to later ones.
	var mu sync.Mutex
	warm, x := guacamole(f)
	f.Fuzz(func(t *testing.T, src string) {
		s, _ := x.Lex([]byte(src))
		lattices := []*lex.Lattice[lex.None]{lex.Linear(s)}
		if n := len(s.Tokens); n >= 2 {
			lattices = append(lattices, withDeadEnd(lattices[0], len(src)%(n-1)))
		}
		cold := Recognize(New(guacTab, Options{NoCache: true}), lattices[0])
		want := cold.Accepted()
		no := Recognize(New(guacTab, Options{NoCache: true, NoLookahead: true}), lattices[0])
		if no.Accepted() != want {
			t.Fatalf("%q: accepted %v without lookahead, %v with", src, no.Accepted(), want)
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
			if got := Recognize(New(guacTab, Options{NoCache: true}), l).Accepted(); got != want {
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
