// ABOUTME: FuzzRecognize, the recognizer's fuzz target: it never panics, and acceptance does not depend on the cache.
// ABOUTME: Nor on lookahead. Run by `make fuzz-chart-recognize`; the seeds in testdata/fuzz/ replay in make test.

package recognize

import (
	"sync"
	"testing"

	"tamarou.com/pvm/internal/chart/lex"
)

func FuzzRecognize(f *testing.F) {
	for _, s := range []string{
		"", "my $x = 1;", "print(1)", "1 2", "$x $y", "my $x =",
		"sub f { my ($a) = @_; return $a + 1; }\nf(2);\n",
		"my @a = map { $_ * 2 } grep { defined($_) } @b;",
		"if ($x) { 1 } elsif ($y) { 2 } else { 3 }",
		"=pod\n\nx\n\n=cut\n",
	} {
		f.Add(s)
	}
	// warm is shared across inputs, so its cache serves earlier inputs'
	// transitions to later ones.
	var mu sync.Mutex
	warm, x := guacamole(f)
	f.Fuzz(func(t *testing.T, src string) {
		s, _ := x.Lex([]byte(src))
		l := lex.Linear(s)
		want := Recognize(New(guacTab, Options{NoCache: true}), l).Accepted()
		if got := Recognize(New(guacTab, Options{NoCache: true, NoLookahead: true}), l).Accepted(); got != want {
			t.Fatalf("%q: accepted %v without lookahead, %v with", src, got, want)
		}
		mu.Lock()
		defer mu.Unlock()
		for range 2 {
			c := Recognize(warm, l)
			if c.Accepted() != want {
				t.Fatalf("%q: accepted %v with the cache, %v without", src, c.Accepted(), want)
			}
			for pos := -1; pos <= len(c.Sets); pos++ {
				_, _ = c.Expected(pos)
			}
		}
	})
}
