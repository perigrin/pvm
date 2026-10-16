// ABOUTME: A `package` or `class` with no name is refused, never a panic that kills psc check and the LSP.
// ABOUTME: Each refused case is perl 5.42's compile error; the named form is the clean control.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestNamelessPackageDoesNotPanic: the declaration's name is optional to the
// parser's reading but not to perl. Measured on 5.42.0 with `perl -c`:
//
//	package                       syntax error at line 1, at EOF
//	package;\nmy $x;              syntax error at line 1, near "package;"
//	use v5.38;\nclass;            Bareword "class" not allowed while "strict subs"
//	use feature 'class';\nclass;  syntax error at line 2, near "class;"
//	package Foo;\n...\npackage    syntax error at line 3, at EOF
//
// Every one is refused. `package Foo;` is the control, and parses clean.
func TestNamelessPackageDoesNotPanic(t *testing.T) {
	for _, src := range []string{
		"package",
		"package;\nmy $x;",
		"use v5.38;\nclass;",
		"use v5.38;\nuse feature 'class';\nclass;",
		"package Foo;\nmy $x = 1;\npackage ",
		"package {}",
	} {
		root := parse.Parse([]byte(src))
		if firstUnknown(root) == nil {
			t.Errorf("%q: perl refuses it, but it parsed with no Unknown", src)
		}
	}
	if u := firstUnknown(parse.Parse([]byte("package Foo;\nmy $x;"))); u != nil {
		t.Errorf("control `package Foo;`: refused at %d", u.Start)
	}
}
