// ABOUTME: Tests try, catch, finally, defer, labels and the yada-yada statement.
// ABOUTME: Snippets under testdata/port/try-defer-labels are recognized through internal/lexer, chart/lex and the recognizer; try-defer-labels-refused ones refuse.

package chart

import (
	"os"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
)

// tryDeferLabelsLines are the perly.y alternatives this family ports. 812 (catch_paren: empty) is left
// unmarked by the ledger's `excluded conditional-refusal` instead: toke.c's action refuses a catch with no
// variable, so the rule is fitted to exclude it and no alternative here accepts it.
var tryDeferLabelsLines = []int{325, 662, 759, 815, 850, 854, 861, 883, 896, 902, 979, 981}

func TestPortTryDeferLabels(t *testing.T) {
	if err := grammar.CheckFamilyMarked(append(tryDeferLabelsLines, 812), nil); err != nil {
		t.Error(err)
	}
	for _, f := range snippets(t, "try-defer-labels") {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if v := recognizeSource(t, src); !v.accepted {
			t.Errorf("%s is refused at byte %d (expected %v); perl accepts it", f, v.at, v.expected)
		}
	}
}

func TestPortTryDeferLabelsRefusals(t *testing.T) {
	for _, f := range snippets(t, "try-defer-labels-refused") {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if v := recognizeSource(t, src); v.accepted {
			t.Errorf("%s is accepted; perl refuses it", f)
		}
	}
	for _, src := range []string{"use feature 'try'; $y = 0; try { 1 } catch ($e) { 2 }",
		"use feature 'try'; $y = 0; try { 1 } catch ($e) { 2 } finally { 3 }", "use feature 'defer'; $y = 0; defer { 1 }",
		"...;", "L: 1;", "$y = 1 ... 2;"} {
		if v := recognizeSource(t, []byte(src)); !v.accepted {
			t.Errorf("%q is refused at %d", src, v.at)
		}
	}
}
