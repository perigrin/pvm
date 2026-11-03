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
	for _, src := range []string{"time;", "defined($x);", "ref $x;", "not $x;", "last;",
		"require;", "__PACKAGE__;", "require 5.006;", "require v5.10;", "require 5.10.1;"} {
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

// toke.c returns lt, gt, le, ge, eq, ne and cmp as operators whatever precedes them; only x falls
// through to a word. perl -c (94e50866): each refused row is a syntax error, its nearby form compiles.
func TestPortNamedOperatorsOperatorWords(t *testing.T) {
	for _, tc := range []struct{ refused, nearby string }{
		{"lt;", "ltx;"}, {"cmp;", "cmpx;"}, {"$y = lt;", "$y = ltx;"}, {"$y = ge;", "$y = gex;"},
	} {
		if v := recognizeSource(t, []byte(tc.refused)); v.accepted {
			t.Errorf("%q is accepted; perl refuses it", tc.refused)
		}
		if v := recognizeSource(t, []byte(tc.nearby)); !v.accepted {
			t.Errorf("%q, near %q, is refused at %d", tc.nearby, tc.refused, v.at)
		}
	}
}

// x after a UNIOP word, a LOOPEX word, require, not or a filetest is a word, because toke.c expects a
// term there; after a term it is the repetition operator. perl -c (94e50866): each refused row is a
// syntax error, its nearby form compiles.
func TestPortNamedOperatorsXAfterNamedOperator(t *testing.T) {
	for _, tc := range []struct{ refused, nearby string }{
		{"shift x 3;", "time x 3;"}, {"ref x 3;", "ref x;"}, {"last x 3;", "last x;"},
		{"require x 3;", "require x;"}, {"not x 3;", "not x;"}, {"-e x 3;", "-e x;"},
	} {
		if v := recognizeSource(t, []byte(tc.refused)); v.accepted {
			t.Errorf("%q is accepted; perl refuses it", tc.refused)
		}
		if v := recognizeSource(t, []byte(tc.nearby)); !v.accepted {
			t.Errorf("%q, near %q, is refused at %d", tc.nearby, tc.refused, v.at)
		}
	}
	// A builtin's name called with & is a sub call, so x after it is repetition. perl -c accepts all four.
	for _, src := range []string{"&ref x 3;", "$y = &shift x 3;", "&last x 3;", "&require x 3;"} {
		if v := recognizeSource(t, []byte(src)); !v.accepted {
			t.Errorf("%q is refused at %d", src, v.at)
		}
	}
}
