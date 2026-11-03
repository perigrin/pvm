// ABOUTME: Tests the list operators and method calls: LSTOP, FUNC, BLKLSTOP, method calls and their method names.
// ABOUTME: Snippets under testdata/port/list-operators are recognized through internal/lexer, chart/lex and the recognizer; list-operators-refused ones refuse.

package chart

import (
	"os"
	"slices"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/chart/lex"
)

// listOperatorsLines are the perly.y alternatives this family ports or fits. None is left for a completing issue.
// 1277 and 1892 are the indirect-object family's, written as they are so that `sort { $a <=> $b } @a`
// has the one path perly.y gives it (LSTOP indirob listexpr, indirob a block); that family keeps them
// and adds the other indirob forms.
var listOperatorsLines = []int{1277, 1281, 1289, 1295, 1324, 1326, 1344, 1345, 1713, 1892}

func TestPortListOperators(t *testing.T) {
	if err := grammar.CheckFamilyMarked(listOperatorsLines, nil); err != nil {
		t.Error(err)
	}
	for _, f := range snippets(t, "list-operators") {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if v := recognizeSource(t, src); !v.accepted {
			t.Errorf("%s is refused at byte %d (expected %v); perl accepts it", f, v.at, v.expected)
		}
	}
}

func TestPortListOperatorsRefusals(t *testing.T) {
	for _, f := range snippets(t, "list-operators-refused") {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if v := recognizeSource(t, src); v.accepted {
			t.Errorf("%s is accepted; perl refuses it", f)
		}
	}
	for _, src := range []string{"$o->m(1);", "$o->m;", "$o->$m();", "push @a, 1;", "sort { $a <=> $b } @a;"} {
		if v := recognizeSource(t, []byte(src)); !v.accepted {
			t.Errorf("%q is refused at %d, expecting %v", src, v.at, v.expected)
		}
	}
}

// terminalsOf names the terminals that admit the token at byte offset of src.
func terminalsOf(t *testing.T, src string, offset int) []string {
	t.Helper()
	tab, err := committed()
	if err != nil || tab == nil {
		t.Fatalf("the committed grammar generates no tables: %v", err)
	}
	s, _ := lex.Lex([]byte(src), tab.Tokens)
	for _, tok := range s.Tokens {
		if s.Raw[tok.Index].Start == offset {
			var names []string
			for _, id := range tab.TerminalSets[tok.TerminalSet] {
				names = append(names, tab.Symbols[id].Name)
			}
			return names
		}
	}
	t.Fatalf("%q has no token at %d", src, offset)
	return nil
}

// toke.c's lop() returns FUNC when the next significant token is a parenthesis and LSTOP otherwise
// (BLKLSTOP for all and any), so the same word is told apart by what follows it. return is OLDLOP,
// LSTOP either way.
func TestPortListOperatorsFuncVersusLstop(t *testing.T) {
	for _, tc := range []struct{ src, want, not string }{
		{"push(@a, 1);", "FUNC", "LSTOP"},
		{"push (@a, 1);", "FUNC", "LSTOP"},
		{"push @a, 1;", "LSTOP", "FUNC"},
		{"print();", "FUNC", "LSTOP"},
		{"print $x;", "LSTOP", "FUNC"},
		{"$r = any(@a);", "FUNC", "BLKLSTOP"},
		{"$r = any { 1 } @a;", "BLKLSTOP", "FUNC"},
		{"return(1);", "LSTOP", "FUNC"},
	} {
		at := 0
		if tc.src[0] == '$' {
			at = 5
		}
		got := terminalsOf(t, tc.src, at)
		if !slices.Contains(got, tc.want) || slices.Contains(got, tc.not) {
			t.Errorf("%q: the first word admits %v, want %s and not %s", tc.src, got, tc.want, tc.not)
		}
	}
}
