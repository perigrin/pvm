// ABOUTME: A declaration statement that stops before its end refuses, as an
// ABOUTME: expression statement does, rather than starting a second statement.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestDeclarationInitialiserDoesNotSplit: `WORD $var` with an unknown WORD is
// an indirect method call to perl -- measured on 5.42.0,
// `my $w = Foo::bar $r, 1;` deparses as `my $w = $r->Foo::bar, '???';` --
// and the parser declines to read it (issue 01a0ebb0-6329-7a13-b398-a5120ec9bbe8).
// The call stops before `$r`. In every position but a declaration's
// initialiser that left tokens before the terminator and the statement
// refused; after `my $w =` the rest became a second statement with no `;`
// between them, at Unknown=0:
//
//	my $w = bar $r, 1;    canon: my $w = bar();$r , 1;
func TestDeclarationInitialiserDoesNotSplit(t *testing.T) {
	for _, src := range []string{
		"my $w = bar $r;",
		"my $w = bar $r, 1;",
		"our $w = Foo::bar $r;",
	} {
		root := parse.Parse([]byte(src))
		if !containsKind(root, parse.Unknown) {
			t.Errorf("%q: must refuse, not split; canon %q", src, parse.Canon(root, []byte(src)))
		}
	}
	// Statements the check must leave alone: a declaration that took its
	// own `;`, one ending in a block, and the comma list at statement level.
	for _, src := range []string{
		"my $x = 1;\nmy $y = 2;\n",
		"my sub f { 1 }\nf();\n",
		"my $aa, $bb, $cc;\n",
		"my $x = 1 if $c;\n",
		"sub f { my $x }\n",
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q: perl accepts this; got %s", src, shape(root))
		}
	}
}
