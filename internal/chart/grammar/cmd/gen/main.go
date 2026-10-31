// ABOUTME: The chart grammar's generator entry point: writes the LR(0) tables of the committed grammar, of which there is none yet.
// ABOUTME: Run by `make chart-tables` from the repository root; with no grammar it writes nothing and says so on stderr.

// Command gen writes the chart parser's tables. No grammar is committed, so
// it has nothing to generate: it writes no table file and exits 0.
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stderr, "chart-tables: no grammar is committed; nothing to generate, no tables written")
}
