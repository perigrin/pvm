// ABOUTME: A `&` before a scalar is a call through a code reference, and the
// ABOUTME: forms perl rejects must keep refusing.
package parse_test

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestAmpersandScalarCall covers `&$coderef` in its four spellings.
//
// The lexer was already right: `&$s` arrives as `FuncSigil(&) Variable($s)`,
// which is what `&f` arrives as with a Word in place of the Variable. The gap
// was in term.go's FuncSigil arm, which joined only a following Word, so the
// scalar fell through to a bare-sigil term and its name had nowhere to go.
//
// The three forms perl REJECTS are asserted here too, because the fix must not
// widen to them. Verified on 5.42.0:
//
//	&$s[0];   syntax error at -e line 1, near "$s["
//	&$s{k};   syntax error at -e line 1, near "$s{k"
//	&@a;      Bareword found where operator expected
//
// So `&` takes a `$` and not an `@`, and nothing subscripted. Refusing those is
// agreement with perl rather than a shortfall.
func TestAmpersandScalarCall(t *testing.T) {
	clean := []string{
		"&$subref;",
		"&$s;",
		"&$s();",
		"&$s(1,2);",
		"my $s = sub {1}; &$s;",
		"my $c = \\&$s;",
	}
	for _, src := range clean {
		b := []byte(src)
		n := parse.Parse(b)
		if got := countUnknown(n); got != 0 {
			t.Errorf("Parse(%q): Unknown=%d, want 0", src, got)
		}
		// Canon must re-emit the ampersand. Today these echo their bytes
		// because the statement is Unknown; once it parses, the `&` has to
		// come from the tree.
		if canon := parse.Canon(n, b); !strings.Contains(canon, "&$") {
			t.Errorf("Canon(%q) = %q, lost the `&$`", src, canon)
		}
	}

	// Still bitwise-and, and still the low-precedence boolean.
	for _, src := range []string{"my $x = $a & $b;", "my $y = $a && $b;"} {
		b := []byte(src)
		if got := countUnknown(parse.Parse(b)); got != 0 {
			t.Errorf("Parse(%q): Unknown=%d, want 0", src, got)
		}
	}

	// The forms perl REJECTS must keep refusing. This half is not padding:
	// the first version of this fix joined any Variable, and all three of
	// these went to Unknown=0 -- agreeing with nothing, since perl calls each
	// a syntax error. A parser that accepts what perl refuses is as wrong as
	// one that refuses what perl accepts, and only this assertion says so.
	for _, src := range []string{"&$s[0];", "&$s{k};", "&@a;"} {
		b := []byte(src)
		if got := countUnknown(parse.Parse(b)); got == 0 {
			t.Errorf("Parse(%q): Unknown=0, but perl rejects it; want a refusal", src)
		}
	}
}
