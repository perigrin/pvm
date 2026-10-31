// ABOUTME: The bodies of BenchmarkLex, BenchmarkRecognize and BenchmarkExpected, outside a _test file.
// ABOUTME: The ratchet can run them with testing.Benchmark; make test never runs a benchmark.

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
func BenchRecognize(b *testing.B, tab *grammar.Tables, src []byte) {
	s, _ := lex.Lex(src, tab.Tokens)
	l := lex.Linear(s)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		if _, err := recognize.Recognize(recognize.New(tab, recognize.Options{}), l); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchExpected times Expected at every node of src's chart. Each run
// recognizes on a fresh recognizer outside the timer, so no run reads the
// memo an earlier one filled.
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
	}
}
