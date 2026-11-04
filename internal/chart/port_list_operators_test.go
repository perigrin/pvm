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
// 1277 (LSTOP indirob listexpr) and 1892 (indirob: block) are written too, borrowed from the indirect-object
// family (01a18fa6-f3c8) so that `sort { $a <=> $b } @a` has its one path; that family marks them.
var listOperatorsLines = []int{1281, 1289, 1295, 1324, 1326, 1344, 1345, 1713}

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

// terminalsAt names the terminals that admit the token at byte offset of src.
func terminalsAt(t *testing.T, src string, offset int) []string {
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
		got := terminalsAt(t, tc.src, at)
		if !slices.Contains(got, tc.want) || slices.Contains(got, tc.not) {
			t.Errorf("%q: the first word admits %v, want %s and not %s", tc.src, got, tc.want, tc.not)
		}
	}
}

// toke.c's KEY_x is the repetition operator exactly when PL_expect is XOPERATOR, which a token that
// completes a term leaves, and a word otherwise. perl -c (94e50866) agrees with every row.
//
// Pinned for when do blocks land (remaining terms): `$y = do { 1 } * ref x 3;` will be accepted, where perl
// refuses it, because the neighbour cannot tell a do block's } from a statement block's. Add it as an
// accepted row marked with this issue then, until the lexer marks what a } closes
// (01a1a422-9176-7153-8625-6535b483a4b4). It cannot be a row here: do blocks are not ported on this base.
func TestPortListOperatorsXRule(t *testing.T) {
	for src, want := range map[string]bool{
		"$z x 3;":                   true,
		"$i++ x 3;":                 true,
		"time x 3;":                 true,
		"x x x;":                    true,
		"$y = 1 x 2 x 3;":           true,
		"$y = push(@a, 1) x 3;":     true,
		"$o->ref x 3;":              true,
		"$o->print x 3;":            true,
		"&ref x 3;":                 true,
		"&print x 3;":               true,
		"*print x 3;":               true,  // a glob's name, whatever it spells
		"$y = 2 * *print x 3;":      true,  // a glob after a multiplication
		"if (1) { } *print x 3;":    true,  // a glob after a block
		"$y = 2 * ref x 3;":         false, // a multiplication, so ref waits for a term
		"$y = $a * shift x 3;":      false,
		"$y = 2 * x 3;":             false,
		"foo * ref x 3;":            false, // a bareword completes a term, so the * multiplies
		"$y = foo * ref x 3;":       false,
		"$y = *print * ref x 3;":    false, // a glob's name completes a term, spelled like a keyword or not
		"$y = &print * ref x 3;":    false,
		"$y = $o->print * ref x 3;": false,
		"$y = *print * *print x 3;": true,
		"*open x 3;":                true,
		"$y = *print x 3;":          true,
		"ref x 3;":                  false,
		"shift x 3;":                false,
		"last x 3;":                 false,
		"push @a, x 3;":             false,
		"$y = 1, x 3;":              false,
		"$z x;":                     false,
	} {
		if v := recognizeSource(t, []byte(src)); v.accepted != want {
			t.Errorf("%q: accepted %v, want %v (refused at %d, expecting %v)", src, v.accepted, want, v.at, v.expected)
		}
	}
}

// Known over-acceptance, pinned so it is noticed when it ends. After an XREF list operator toke.c reads
// the bareword as a filehandle and expects a term after it, so x is a word there and perl refuses each
// of these (perl -c, 94e50866: "Number found where operator expected"). The chart reads the bareword as
// a term and x as repetition. That is the indirect-object family's to fix (01a18fa6-f3c8): when it does,
// this test fails, and its rows move to the refused snippets.
func TestPortListOperatorsFilehandleBarewordKnownLimit(t *testing.T) {
	for _, src := range []string{"print FH x 3;", "printf FH x 3;", "say FH x 3;", "exec foo x 3;", "sort foo x 3;"} {
		if v := recognizeSource(t, []byte("use feature 'say';\n"+src)); !v.accepted {
			t.Errorf("%q is refused at %d: the known limit has ended, so move it to the refused snippets", src, v.at)
		}
	}
}
