// ABOUTME: Tests the indirect objects: indirob, and the list operators and method calls that take one.
// ABOUTME: Snippets under testdata/port/indirect-objects are recognized; indirect-objects-refused ones refuse.

package chart

import (
	"os"
	"slices"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
)

// indirectObjectsLines are the perly.y alternatives this family ports or fits. None is left for a completing issue.
var indirectObjectsLines = []int{1277, 1285, 1312, 1318, 1888, 1890, 1892, 1895}

func TestPortIndirectObjects(t *testing.T) {
	if err := grammar.CheckFamilyMarked(indirectObjectsLines, nil); err != nil {
		t.Error(err)
	}
	for _, f := range snippets(t, "indirect-objects") {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if v := recognizeSource(t, src); !v.accepted {
			t.Errorf("%s is refused at byte %d (expected %v); perl accepts it", f, v.at, v.expected)
		}
	}
}

// After a list operator toke.c gives a filehandle (an XREF operator's or one whose first operand is
// OA_FILEREF), a bareword is an indirect object and PL_expect is XTERM (toke.c:8167), so an x after it
// is a word, and perl refuses each refused snippet's `FH x 3` (perl -c, 94e50866: "Number found where
// operator expected", then a syntax error). The bare statements below are the snippets' accepted forms.
//
// Pinned: `print(FH x 3);` is accepted, where perl refuses it. toke.c's yyl_leftparen lets the word after
// `print(` be the filehandle too (`allow print(STDOUT 123)`), but the x after it is three tokens past the
// operator, which no key part reaches. `new Foo x 3;` is accepted, where perl refuses it unless `new` is a
// declared sub (intuit_method leaves PL_expect at XTERM after Foo only when it is not): scope decides it.
func TestPortIndirectObjectsRefusals(t *testing.T) {
	for _, f := range snippets(t, "indirect-objects-refused") {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if v := recognizeSource(t, src); v.accepted {
			t.Errorf("%s is accepted; perl refuses it", f)
		}
	}
	for _, src := range []string{
		"print {$fh} 1;", "print STDERR 1;", "print $fh 1;", "new Foo(1);", "new Foo;", "print STDERR;",
		"print FH => 1;", "print $fh x 3;", "join foo x 3;", "close FH x 3;", "map foo x 3;", "x x 3;",
		"print(FH x 3);", "new Foo x 3;", // the limit and the scope reading above
	} {
		if v := recognizeSource(t, []byte(src)); !v.accepted {
			t.Errorf("%q is refused at %d, expecting %v", src, v.at, v.expected)
		}
	}
}

// A sub's name after & is PRIVATEREF when a lexical sub is in scope and BAREWORD otherwise (toke.c's
// pending_ident, reached through yyl_ampersand), so it is both. A glob's name is never in the pad, and a
// list operator's bareword filehandle is yyl_safe_bareword's BAREWORD, so those are one.
func TestPortIndirectObjectsSuperposition(t *testing.T) {
	for _, tc := range []struct {
		src       string
		at        int
		want, not []string
	}{
		{"$y = &foo(1);", 6, []string{"BAREWORD", "PRIVATEREF"}, nil},
		{"$y = *foo;", 6, []string{"BAREWORD"}, []string{"PRIVATEREF"}},
		{"print STDERR 1;", 6, []string{"BAREWORD"}, []string{"PRIVATEREF"}},
	} {
		got := terminalsAt(t, tc.src, tc.at)
		for _, w := range tc.want {
			if !slices.Contains(got, w) {
				t.Errorf("%q at %d admits %v, want %s", tc.src, tc.at, got, w)
			}
		}
		for _, n := range tc.not {
			if slices.Contains(got, n) {
				t.Errorf("%q at %d admits %v, not %s", tc.src, tc.at, got, n)
			}
		}
	}
}
