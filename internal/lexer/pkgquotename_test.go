// ABOUTME: After `package` a quote-operator word is the package's name:
// ABOUTME: `package y;` is the package y, not a transliteration.

package lexer

import "testing"

// TestPackageQuoteOpName: measured on 5.44.0, `perl -c` accepts `package W;`,
// `package W 1.2;` and `package W { 1 }` for every quote operator word W.
// After `package` perl reads a name, so the word is never a quote there.
func TestPackageQuoteOpName(t *testing.T) {
	for _, name := range []string{"m", "q", "qq", "qr", "qw", "qx", "s", "tr", "y"} {
		for _, src := range []string{
			"package " + name + "; $n = 1;",
			"package " + name + " 1.2; $n = 1;",
			"package " + name + " { 1 }",
		} {
			for _, tok := range Tokenize([]byte(src)) {
				if tok.Kind == Quote || tok.Kind == UnknownRest || tok.Kind == Error {
					t.Errorf("%q: `%s` after package is a package name, got %v %q",
						src, name, tok.Kind, src[tok.Start:tok.End])
				}
			}
		}
	}
	// Anywhere else the word still quotes.
	if _, ok := firstOfKind("package Foo; q(x);", Quote); !ok {
		t.Errorf("`q(x)` after a package statement is still a quote")
	}
}
