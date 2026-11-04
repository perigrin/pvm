// ABOUTME: Tests the package declaration and block and the version tokens toke.c forces after package, use and no.
// ABOUTME: Snippets under testdata/port/package-use are recognized through internal/lexer, chart/lex and the recognizer; package-use-refused ones refuse.

package chart

import (
	"os"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
)

// packageUseLines are the perly.y alternatives this family ports or fits: both package forms and their
// dispatch. 714 is the skeleton's and marked there; this family adds a fitted line beside it, which the
// ledger test covers. None is left for a completing issue. `use VERSION MODULE` is not fitted (see
// grammar/versions.go).
var packageUseLines = []int{540, 556, 891, 892}

func TestPortPackageAndUse(t *testing.T) {
	if err := grammar.CheckFamilyMarked(packageUseLines, nil); err != nil {
		t.Error(err)
	}
	for _, f := range snippets(t, "package-use") {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if v := recognizeSource(t, src); !v.accepted {
			t.Errorf("%s is refused at byte %d (expected %v); perl accepts it", f, v.at, v.expected)
		}
	}
}

func TestPortPackageAndUseRefusals(t *testing.T) {
	for _, f := range snippets(t, "package-use-refused") {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if v := recognizeSource(t, src); v.accepted {
			t.Errorf("%s is accepted; perl refuses it", f)
		}
	}
	// Perl refuses these in toke.c, with "Invalid version format" and no syntax error, so they are not snippets.
	for _, src := range []string{"package Foo Bar;", "package Foo x;", "package Foo v1;", "package Foo Bar { }"} {
		if v := recognizeSource(t, []byte(src)); v.accepted {
			t.Errorf("%q is accepted; perl refuses it", src)
		}
	}
	// Accepted beside the refused forms, so a harness that refuses everything fails here. The bare
	// statements stay covered although a prefix of each is refused only at end of input.
	for _, src := range []string{"package Foo;", "package Foo 1.23;", "package Foo { 1 }", "package Foo 1.23 { 1 }",
		"use 5.036;", "use v5.36;", "no 5.036;", "use strict 1 2;", "use strict 1, 2;", "use strict 1.0 'refs';",
		"package Foo 0.1;", "package Foo 10.01 { }", "use 1_0;", "use 01;", "use 1.;", "use lib 1.2, 3;", "use lib 1.2 'a';"} {
		if v := recognizeSource(t, []byte(src)); !v.accepted {
			t.Errorf("%q is refused at %d", src, v.at)
		}
	}
}
