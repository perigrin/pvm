// ABOUTME: Tests the conditionals: if, unless, while, until, their else, elsif and continue parts, and the statement modifiers.
// ABOUTME: Snippets under testdata/port/conditionals are recognized through internal/lexer, chart/lex and the recognizer; conditionals-refused ones refuse.

package chart

import (
	"os"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
)

// conditionalsLines are the perly.y alternatives this family ports. None is left for a completing issue.
// 993, 994 and 1015 (nexpr, mnexpr) are reached only through the C-style for, so they are that family's.
var conditionalsLines = []int{517, 684, 699, 744, 799, 806, 889, 897, 898, 901, 929, 937, 940, 943, 945, 947, 950,
	956, 957, 962, 974, 986, 998, 1002, 1006, 1011}

func TestPortConditionals(t *testing.T) {
	if err := grammar.CheckFamilyMarked(conditionalsLines, nil); err != nil {
		t.Error(err)
	}
	for _, f := range snippets(t, "conditionals") {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if v := recognizeSource(t, src); !v.accepted {
			t.Errorf("%s is refused at byte %d (expected %v); perl accepts it", f, v.at, v.expected)
		}
	}
}

func TestPortConditionalsRefusals(t *testing.T) {
	for _, f := range snippets(t, "conditionals-refused") {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if v := recognizeSource(t, src); v.accepted {
			t.Errorf("%s is accepted; perl refuses it", f)
		}
	}
	for _, src := range []string{"if ($x) { $y }", "unless ($x) { 1 } else { 2 }", "while ($x) { 1 } continue { 2 }", "$y if $x;",
		"$y if 1", "if (1) { $x }", "continue;", "$y = continue;"} {
		if v := recognizeSource(t, []byte(src)); !v.accepted {
			t.Errorf("%q is refused at %d", src, v.at)
		}
	}
}
