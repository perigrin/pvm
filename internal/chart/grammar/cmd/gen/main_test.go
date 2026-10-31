// ABOUTME: Tests that gen exits non-zero on a refused grammar, naming the offence on stderr, and 0 on a consistent one.
// ABOUTME: Run on toy grammars through run, since the committed grammar has no rule to refuse yet.

package main

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/lexer"
)

// toy is s ::= A (perly.y:10) with A a classed operator, and the reference list it cites.
func toy() (*grammar.Grammar, []grammar.Entry) {
	kind, text := lexer.Operator, "+"
	g := &grammar.Grammar{
		G1:      []grammar.Rule{{LHS: "s", RHS: []grammar.Symbol{{Kind: grammar.Name, Text: "A"}}, Cite: "perly.y:10"}},
		G0:      []grammar.Rule{{LHS: "A"}},
		Classes: []grammar.TokenClass{{Terminals: []int{1}, Kind: &kind, Text: &text}},
	}
	return g, []grammar.Entry{{Line: 10, Text: "s: A"}}
}

func TestGenExitsNonZeroOnRefusal(t *testing.T) {
	dir, stderr := t.TempDir(), &bytes.Buffer{}
	g, ref := toy()
	if code := run(g, ref, dir, stderr); code != 0 {
		t.Fatalf("a consistent grammar exits %d: %s", code, stderr)
	}
	if entries, _ := os.ReadDir(dir); len(entries) == 0 {
		t.Error("a consistent grammar wrote no tables")
	}

	dir, stderr = t.TempDir(), &bytes.Buffer{}
	g.G1[0].Cite = ""
	if code := run(g, ref, dir, stderr); code == 0 {
		t.Error("a refused grammar exits 0")
	}
	if !strings.Contains(stderr.String(), "s ::= A") {
		t.Errorf("stderr does not name the rule: %q", stderr)
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 0 {
		t.Error("a refused grammar wrote tables")
	}
}
