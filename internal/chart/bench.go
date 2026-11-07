// ABOUTME: The bodies of BenchmarkLex, BenchmarkRecognize and BenchmarkExpected, outside a _test file.
// ABOUTME: The ratchet runs them with testing.Benchmark (cmd/ratchet measureBench); make test never runs a benchmark.

package chart

import (
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/chart/lex"
	"tamarou.com/pvm/internal/chart/recognize"
)

// BenchLex times chart/lex over src, the production lexer included.
func BenchLex(b *testing.B, tab *grammar.Tables, src []byte) {
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		lex.Lex(src, tab.Tokens)
	}
}

// BenchRecognize times recognizing src's lattice, lexed once outside the
// timed loop, with a fresh recognizer per run (the cold lifetime, §9.3).
// On a file the chart refuses it times only the prefix the grammar reaches;
// nodes-reached says how far recognition got, out of nodes.
func BenchRecognize(b *testing.B, tab *grammar.Tables, src []byte) {
	s, _ := lex.Lex(src, tab.Tokens)
	l := lex.Linear(s)
	var c *recognize.Chart
	var err error
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		if c, err = recognize.Recognize(recognize.New(tab, recognize.Options{}), l); err != nil {
			b.Fatal(err)
		}
	}
	reportReach(b, c)
}

// BenchExpected times Expected at every node of src's chart. Each run
// recognizes on a fresh recognizer outside the timer, so no run reads the
// memo an earlier one filled. A node with no set expects nothing at no
// cost, so on a file the chart refuses it times only the prefix the grammar
// reaches.
func BenchExpected(b *testing.B, tab *grammar.Tables, src []byte) {
	s, _ := lex.Lex(src, tab.Tokens)
	l := lex.Linear(s)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		b.StopTimer()
		c, err := recognize.Recognize(recognize.New(tab, recognize.Options{}), l)
		if err != nil {
			b.Fatal(err)
		}
		b.StartTimer()
		for pos := range c.Sets {
			if _, err := c.Expected(pos); err != nil {
				b.Fatal(err)
			}
		}
		b.StopTimer()
		reportReach(b, c)
		b.StartTimer()
	}
}

// reportReach reports nodes-reached, the first node of c with no Earley set
// (every node, once the grammar recognizes the input), beside nodes, all of them.
func reportReach(b *testing.B, c *recognize.Chart) {
	n := len(c.Sets)
	for i, s := range c.Sets {
		if s == nil {
			n = i
			break
		}
	}
	b.ReportMetric(float64(n), "nodes-reached")
	b.ReportMetric(float64(len(c.Sets)), "nodes")
}
