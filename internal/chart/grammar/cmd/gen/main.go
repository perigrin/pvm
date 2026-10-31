// ABOUTME: The chart grammar's generator entry point: writes the LR(0) tables of the committed grammar.
// ABOUTME: Run by `make chart-tables` from the repository root; a refused grammar writes nothing and exits 1.

// Command gen writes the chart parser's tables. It refuses an inconsistent
// grammar: the offences go to stderr, no table is written and it exits 1.
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"tamarou.com/pvm/internal/chart/grammar"
)

func main() {
	g, ref, err := grammar.Committed()
	if err != nil {
		fmt.Fprintf(os.Stderr, "chart-tables: %v\n", err)
		os.Exit(1)
	}
	os.Exit(run(g, ref, filepath.Join("internal", "chart", "grammar", "testdata", "tables"), os.Stderr))
}

// run generates g's tables into dir and returns the exit code.
func run(g *grammar.Grammar, ref []grammar.Entry, dir string, stderr io.Writer) int {
	if len(g.G1) == 0 {
		fmt.Fprintln(stderr, "chart-tables: no rule is committed; nothing to generate, no tables written")
		return 0
	}
	if err := grammar.GenerateTables(dir, g, ref); err != nil {
		fmt.Fprintf(stderr, "chart-tables: %v\n", err)
		return 1
	}
	return 0
}
