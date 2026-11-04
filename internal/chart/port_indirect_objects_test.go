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

// After sort, a bareword is the sort sub's name whether or not a sub of that name is declared, and
// PL_expect is XTERM after it (toke.c:8158-8167, "indir obj regardless for sort"), so an x after it is a
// word and perl refuses `sort foo x 3;`. A ( after sort lets the word after it be one too (yyl_leftparen,
// "allow print(STDOUT 123)"). After the other list operators that take a filehandle, the bareword is one
// only when no sub of that name is declared: `print FH x 3;` is refused by perl without a declaration
// and accepted with `use constant FH => 1;`, so scope decides it and the chart accepts it.
// `new Foo x 3;` is the same: perl refuses it unless `new` is a declared sub (intuit_method leaves
// PL_expect at XTERM after Foo only when it is not), and so is `foo { a => 1 } x 3;`, a block after a method
// and a hash after a declared sub.
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
		"$o->sort(foo x 3);", "sort(byname @a);", "sort byname @a;",
		"print FH x 3;", "print(FH x 3);", "open(FH x 3);", "print x x 3;", "new Foo x 3;", "foo { a => 1 } x 3;", // scope's, above
		"BEGIN : lvalue { 1 }", // a label and a call with a block, as perl reads it
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
