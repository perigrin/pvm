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
	// Perl refuses these in toke.c with "Invalid version format" and no syntax error, so they are not snippets.
	for _, src := range []string{"package Foo Bar;", "package Foo x;", "package Foo v1;", "package Foo Bar { }"} {
		if v := recognizeSource(t, []byte(src)); v.accepted {
			t.Errorf("%q is accepted; perl refuses it", src)
		}
	}
	// Limits: perl refuses these and the chart accepts them, because the word x after a word is a BAREWORD to
	// the key (a refusal would also refuse `use lib foo x 3;`, which perl accepts). `use lib 1.2 x;` is the
	// other way: perl accepts it and the chart refuses it. A row fails once its limit goes; move it then.
	for _, src := range []string{"use lib foo x;", "use lib foo x, 1;", "use lib foo x => 1;", "use lib x x;", "use lib Foo::Bar x;", "no lib foo x;", "use lib 1.2 x;"} {
		if v := recognizeSource(t, []byte(src)); v.accepted == (src == "use lib 1.2 x;") {
			t.Errorf("%q: the recorded limit moved (accepted=%v)", src, v.accepted)
		}
	}
	if v := recognizeSource(t, []byte("use lib foo x 3;")); !v.accepted {
		t.Errorf("use lib foo x 3; is refused at %d; perl accepts it", v.at)
	}
	// Accepted beside the refused forms, so a harness that refuses everything fails here. The bare
	// statements stay covered although a prefix of each is refused only at end of input.
	for _, src := range []string{"package Foo;", "package Foo 1.23;", "package Foo { 1 }", "package Foo 1.23 { 1 }",
		"use 5.036;", "use v5.36;", "no 5.036;", "use strict 1 2;", "use strict 1, 2;", "use strict 1.0 'refs';",
		"package Foo 0.1;", "package Foo 10.01 { }", "use 1_0;", "use 01;", "use 1.;", "use lib 1.2, 3;", "use lib 1.2 'a';",
		// a word after a module is a term: perl lexes `foo bar` and `foo $x` as method calls and `foo (3)` as a call
		"use lib foo bar;", "use lib foo $x;", "use lib x;", "use lib x, 3;", "use lib foo (3);", "use lib 1.2 - 3;", "use lib 1.2 + 3;", "use lib 1.2 * 3;",
		// ** and // open a glob and a match after a version, which perl's grammar accepts
		"my @b = @a x 3;", "$s = $#a x 3;", "use lib 0.1 **= 3;", "use lib 0.1 //= 3;", "use lib foo %h;", "use lib foo eq 1;",
		// the indirect-object forms: perl -c is syntax OK for both (they fail in BEGIN, at run time)
		"use lib foo { }", "use lib foo $$x"} {
		if v := recognizeSource(t, []byte(src)); !v.accepted {
			t.Errorf("%q is refused at %d", src, v.at)
		}
	}
}
