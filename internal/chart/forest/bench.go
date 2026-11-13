// ABOUTME: The body of BenchmarkForest, outside a _test file, so the ratchet runs the same code (cmd/ratchet measureBench).
// ABOUTME: It times building a file's forest from its chart, recognized once before the timer starts; make test never runs it.

package forest

import (
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/chart/lex"
	"tamarou.com/pvm/internal/chart/recognize"
)

// Bench times building src's forest after recognition: the chart is
// recognized once, outside the timed loop. A file the chart refuses has no
// forest, so it fails the benchmark rather than timing nothing.
func Bench(b *testing.B, tab *grammar.Tables, src []byte) {
	s, _ := lex.Lex(src, tab.Tokens)
	l := lex.Linear(s)
	c, err := recognize.Recognize(recognize.New(tab, recognize.Options{}), l)
	if err != nil {
		b.Fatal(err)
	}
	if !c.Accepted() {
		b.Fatal("the chart refuses the source, so it has no forest to time")
	}
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		Build(tab, c, l)
	}
}
