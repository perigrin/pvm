// ABOUTME: BenchmarkLex, BenchmarkRecognize and BenchmarkExpected over std5k.pl and the committed grammar's tables.
// ABOUTME: Their bodies are in bench.go, so the ratchet runs the same code; make test never runs them.

package chart

import (
	"os"
	"path/filepath"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
)

func benchInput(b *testing.B) (*grammar.Tables, []byte) {
	b.Helper()
	tab, err := committed()
	if err != nil || tab == nil {
		b.Fatalf("the committed grammar generates no tables: %v", err)
	}
	src, err := os.ReadFile(filepath.Join("testdata", "std5k.pl"))
	if err != nil {
		b.Fatal(err)
	}
	return tab, src
}

func BenchmarkLex(b *testing.B) {
	tab, src := benchInput(b)
	BenchLex(b, tab, src)
}

func BenchmarkRecognize(b *testing.B) {
	tab, src := benchInput(b)
	BenchRecognize(b, tab, src)
}

func BenchmarkExpected(b *testing.B) {
	tab, src := benchInput(b)
	BenchExpected(b, tab, src)
}
