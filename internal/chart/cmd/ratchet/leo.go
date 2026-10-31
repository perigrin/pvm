// ABOUTME: The ratchet's leo-disagree: files whose recognition with Leo on and off differs in acceptance or Expected.
// ABOUTME: Until the superset population exists it reads every corpus file, as lex-overhead does.

package main

import (
	"fmt"
	"os"
	"strconv"

	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/chart/lex"
	"tamarou.com/pvm/internal/chart/recognize"
)

// measureLeo adds leo-disagree to measured, naming each disagreeing file on stderr.
func measureLeo(tab *grammar.Tables, files []File, srcs map[string][]byte, measured map[string]string) error {
	n := 0
	for _, f := range files {
		s, _ := lex.Lex(srcs[f.Path], tab.Tokens)
		d, err := recognize.LeoDisagree(tab, lex.Linear(s))
		if err != nil {
			return fmt.Errorf("leo-disagree: %s: %w", f.Path, err)
		}
		if d {
			fmt.Fprintf(os.Stderr, "chart-ratchet: leo-disagree: %s\n", f.Path)
			n++
		}
	}
	measured["leo-disagree"] = strconv.Itoa(n)
	return nil
}
