// ABOUTME: A `&` in term position always names the sub it calls; with no name
// ABOUTME: after it, perl takes the next punctuation character as the name.
package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestBareAmpersandIsRefused: a `&` with no Word, scalar or block after it was
// a term of its own, so `f 1, &= 2` parsed as an assignment to it. perl has no
// bare `&`: toke.c hands the `&` to scan_ident, which takes one punctuation
// character as a name the way `$;` and `$=` are names, so `&=` is a call of
// the sub named `=` and a term after it is a term after a term. Measured on
// 5.42.0 with -MO=Deparse:
//
//	sub f {} f 1, &= 2;   Number found where operator expected ... near "&= 2"
//	&= 2;                 Number found where operator expected ... near "&= 2"
//	& = 2;                Number found where operator expected ... near "& = 2"
//	f(1, &, 2);           Number found where operator expected ... near "&, 2"
//	&+ 1;                 Number found where operator expected ... near "&+ 1"
//	f(&);                 syntax error at -e line 1, near "&);"
//	[&];                  syntax error at -e line 1, near "&];"
//	&&                    syntax error at -e line 1, near "&&"
//	&                     syntax error at -e line 1, at EOF
//
// and, deparsing clean:
//
//	&;            &;;
//	& ;           &;;
//	&=;           &=;
//	f(&));        syntax OK: the first `)` is the name
//	$x &= 2;      $x &= 2;
//	& $x;         &$x;
//	&{$x};        &{$x;};
//	&CORE::print; &CORE::print;
func TestBareAmpersandIsRefused(t *testing.T) {
	for _, src := range []string{
		"sub f {} f 1, &= 2;",
		"&= 2;",
		"& = 2;",
		"f(1, &, 2);",
		"&+ 1;",
		"f(&);",
		"[&];",
		"&&",
		"&",
	} {
		if got := countUnknown(parse.Parse([]byte(src))); got == 0 {
			t.Errorf("Parse(%q): Unknown=0, but perl rejects it; want a refusal", src)
		}
	}

	for _, src := range []string{
		"&;",
		"& ;",
		"&=;",
		"f(&));",
		"[&]];",
		"my $x; $x &= 2;",
		"my $x; & $x;",
		"my $x; &$x;",
		"my $x; &{$x};",
		"my $x; & {$x};",
		"sub f {} &f;",
		"sub f {} &f();",
		"&CORE::print;",
		"sub f {} \\&f;",
		"sub f {} defined &f;",
		"my ($a, $b); $a & $b;",
		"my ($a, $b); $a && $b;",
	} {
		if got := countUnknown(parse.Parse([]byte(src))); got != 0 {
			t.Errorf("Parse(%q): Unknown=%d, want 0", src, got)
		}
	}
}
