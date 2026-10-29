// ABOUTME: The chart grammar's generator: reads the vendored grammar, applies the token cut, writes the LR(0) tables.
// ABOUTME: Run by `make chart-tables` from the repository root; the output is byte-identical on every run.

// Command gen writes the chart parser's tables. Its output directory holds
// exactly the generated files: any other .txt file there is removed.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"tamarou.com/pvm/internal/chart/grammar"
)

func main() {
	dir := filepath.Join("internal", "chart", "grammar", "testdata")
	slif := flag.String("grammar", filepath.Join(dir, "guacamole", "guacamole.slif"), "the SLIF grammar to read")
	out := flag.String("o", filepath.Join(dir, "tables"), "the directory to write the tables into")
	flag.Parse()
	files, err := grammar.GenerateTables(*slif)
	if err == nil {
		err = grammar.WriteTables(*out, files)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "chart-tables: %v\n", err)
		os.Exit(1)
	}
}
