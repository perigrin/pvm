// ABOUTME: Tests the named operators: UNIOP, FUNC0, FUNC0OP, FUNC1, LOOPEX, NOTOP, require and barewords.
// ABOUTME: Snippets under testdata/port/named-operators are recognized through internal/lexer, chart/lex and the recognizer; named-operators-refused ones refuse.

package chart

import (
	"os"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
)

// namedOperatorsLines are the perly.y alternatives this family ports or fits. None is left for a completing issue.
var namedOperatorsLines = []int{1661, 1664, 1666, 1668, 1670, 1672, 1674, 1676, 1683, 1685, 1687, 1689, 1693, 1697, 1712}

func TestPortNamedOperators(t *testing.T) {
	if err := grammar.CheckFamilyMarked(namedOperatorsLines, nil); err != nil {
		t.Error(err)
	}
	for _, f := range snippets(t, "named-operators") {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if v := recognizeSource(t, src); !v.accepted {
			t.Errorf("%s is refused at byte %d (expected %v); perl accepts it", f, v.at, v.expected)
		}
	}
}

func TestPortNamedOperatorsRefusals(t *testing.T) {
	for _, f := range snippets(t, "named-operators-refused") {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if v := recognizeSource(t, src); v.accepted {
			t.Errorf("%s is accepted; perl refuses it", f)
		}
	}
	for _, src := range []string{"time;", "defined($x);", "ref $x;", "not $x;", "last;"} {
		if v := recognizeSource(t, []byte(src)); !v.accepted {
			t.Errorf("%q is refused at %d", src, v.at)
		}
	}
}

// toke.c reads the next token to tell FUNC1 from UNIOP, and the grammar's position to tell x the
// operator from x the word.
func TestPortNamedOperatorsByPosition(t *testing.T) {
	for src, want := range map[string]bool{
		"ref($x);":     true, // FUNC1 ( expr )
		"ref ($x);":    true, // the next token, not adjacency
		"ref $x;":      true, // UNIOP term
		"ref;":         true, // UNIOP alone
		"ref() ;":      true, // FUNC1 ( ), which UNIOP term could not take
		"ref($x $y);":  false,
		"x x 3;":       true, // a word where a statement begins, the operator after a term
		"$y = $z x 3;": true,
		"$z x;":        false, // an operator needs its operand
		"x;":           true,
	} {
		if v := recognizeSource(t, []byte(src)); v.accepted != want {
			t.Errorf("%q: accepted %v, want %v (refused at %d, expecting %v)", src, v.accepted, want, v.at, v.expected)
		}
	}
}
