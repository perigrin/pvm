// ABOUTME: FuzzParser's pre-existing canon-fixpoint finds: `no &0`, `A*%x*(%A` and `;A&~ ~0`.
// ABOUTME: Each is read as perl 5.44.0 reads it, or refused where perl -c rejects it.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// clean parses src and fails unless it has no refusal and no error, then
// returns its canon.
func clean(t *testing.T, src string) string {
	t.Helper()
	root := parse.Parse([]byte(src))
	if found := collectUnknown(root); len(found) != 0 || root.HasError {
		t.Errorf("%q: perl compiles this, but it refuses: %v", src, refusalCodes(found))
	}
	return parse.Canon(root, []byte(src))
}

// TestStarPercentAfterWordParses: after an undeclared word, `*%x` is a
// multiplication by `%x`, not the glob `*%` followed by the repetition `x`.
// perl -c on 5.44.0 accepts `A*%x`, `A*@x`, `A*&x` and `A*%$h`, and rejects
// `A*%x*(%A` at EOF; the glob reading made the latter's canon re-read
// differently. A declared A takes the glob, `sub A {} A*%x3` being
// `A(*% x 3)`, which the parser declines rather than mis-read.
func TestStarPercentAfterWordParses(t *testing.T) {
	for src, want := range map[string]string{
		"A*%x":  "A * %x",
		"A*@x":  "A * @x",
		"A*&x":  "A * &x",
		"A*%$h": "A * %$h",
		// And a punctuation name, as perl's scan_ident takes one: `&;`
		// calls the sub `;`, so the statement still needs its terminator.
		"got*&;": "got * &;;",
		"got*%;": "got * %;",
		"A*@}":   "A * @}",
	} {
		if got := clean(t, src); got != want {
			t.Errorf("%q: canon %q, want %q", src, got, want)
		}
	}
	if found := collectUnknown(parse.Parse([]byte("sub A {} A *%x3"))); len(found) == 0 || found[0].Refusal != parse.AmbiguousSigil {
		t.Errorf("`sub A {} A *%%x3`: want an ambiguous_sigil refusal, got %v", refusalCodes(found))
	}
	if ok, why := canonMatches([]byte("A*%x*(%A")); !ok {
		t.Errorf("`A*%%x*(%%A`: canon is not a fixpoint: %s", why)
	}
}

// TestUseWithoutModuleRefuses: `use` and `no` take a module name or a
// version. perl -c on 5.44.0 rejects `no &0`, `no & ! no & 00`, `use
// 'strict'`, `use $x`, `use +strict` and `use;`; read on, `no &0` took `&`
// as an import list whose canon `&;` re-read differently. `use => 1` is the
// string "use", which perl compiles.
func TestUseWithoutModuleRefuses(t *testing.T) {
	for _, src := range []string{"no &0", "no & ! no & 00", "use 'strict';", "use $x;", "use +strict;", "use;", "no"} {
		if found := collectUnknown(parse.Parse([]byte(src))); len(found) == 0 {
			t.Errorf("%q: perl rejects this, but it parses", src)
		}
	}
	for _, src := range []string{"use => 1;", "no => 1;", "no 6;", "use v5.10;", "use 5.010;", "use strict;"} {
		clean(t, src)
	}
}

// TestAmpBeforeTildeAfterWord: `;A&~ ~0` is `'A' & ~~0` to perl 5.44.0, and
// so is its canon `;A & ~~0`. The lexer took the `&` for a sub's sigil and
// the `~~` after it for smartmatch, so the canon refused.
func TestAmpBeforeTildeAfterWord(t *testing.T) {
	for _, src := range []string{";A&~ ~0", ";A & ~~0", "A & ~~0"} {
		clean(t, src)
		if ok, why := canonMatches([]byte(src)); !ok {
			t.Errorf("%q: canon is not a fixpoint: %s", src, why)
		}
	}
}

// TestAmpDigitsNameASub: perl 5.44.0 compiles `goto&0`, `&0` and `f(&0)`,
// each calling the sub named `0`. Lexed as a sigil and a number, `goto&0`
// was `goto &` and a second statement, and its canon re-read differently.
func TestAmpDigitsNameASub(t *testing.T) {
	// The name may stand apart from the `&`: `& 0;` deparses as `&0;`, and
	// FuzzParser's `!& 0` was `!&` and then `0`.
	for src, want := range map[string]string{"goto&0": "goto &0", "&0": "&0", "f(&0)": "f(&0)", "&12;": "&12;", "!& 0": "!& 0"} {
		if got := clean(t, src); got != want {
			t.Errorf("%q: canon %q, want %q", src, got, want)
		}
	}
	// A longer name may not start with 0: "Numeric variables with more than
	// one digit may not start with '0'". FuzzParser's `no n!&00`.
	for _, src := range []string{"&00;", "f(&01)", "no n!&00", "& 00;"} {
		if found := collectUnknown(parse.Parse([]byte(src))); len(found) == 0 {
			t.Errorf("%q: perl rejects this, but it parses", src)
		}
	}
}

// TestAmpPunctNameEndsTerm: `&!` calls the sub named `!`, so an operator
// follows it. perl 5.44.0 compiles `A*&!*&%*!x &0` as `A * &! * &% * !x &
// 0`; with the `!` lexed as an operator, the `*` after it read as a glob.
func TestAmpPunctNameEndsTerm(t *testing.T) {
	for src, want := range map[string]string{
		"A*&!*&%*!x &0": "A * &! * &% * !x & 0",
		"&!*2;":         "&! * 2;",
		// And a keyword's spelling: `&gt` is a call, not the operator.
		"&gt*&;": "&gt * &;;",
	} {
		if got := clean(t, src); got != want {
			t.Errorf("%q: canon %q, want %q", src, got, want)
		}
	}
}

// TestWordOperatorIsNoTerm: perl 5.44.0 rejects `gt*&;`, `gt*2;` and
// `print -eq;`, and compiles `$h{eq}`, `{eq => 1}` and `x*&;`, the last
// `'x' * &;`. Read as a call, `gt*&;`'s canon re-read differently.
func TestWordOperatorIsNoTerm(t *testing.T) {
	for _, src := range []string{"gt*&;", "gt*2;", "print -eq;", "CORE::eq*2"} {
		if found := collectUnknown(parse.Parse([]byte(src))); len(found) == 0 {
			t.Errorf("%q: perl rejects this, but it parses", src)
		}
	}
	for _, src := range []string{"my %h; $h{eq}", "my $r = { eq => 1 };", "f(-lt => 1)"} {
		clean(t, src)
	}
	if got := clean(t, "x*&;"); got != "x * &;;" {
		t.Errorf("`x*&;`: canon %q, want %q", got, "x * &;;")
	}
}

// TestListSeparatorAfterInnerTrailingComma: perl 5.44.0 compiles
// `((A,),%y)`. The comma inside `(A,)` is that element's, and canon took it
// for the list's separator, writing `((A,) %y)`, a modulus.
func TestListSeparatorAfterInnerTrailingComma(t *testing.T) {
	for src, want := range map[string]string{"((A,),%y)": "((A,), %y)", "((1,),2)": "((1,), 2)", "((A),,%y)": "((A), , %y)", "(0,A,,0)": "(0, A, , 0)"} {
		if got := clean(t, src); got != want {
			t.Errorf("%q: canon %q, want %q", src, got, want)
		}
	}
}

// TestIncrementOverPrefixCanon: canon wrote `+++y=>` as `++(+y)=>`, and
// `y)=>` re-lexed as a transliteration. A blank keeps `++` and the prefix
// operator apart without a paren the source lacks; perl 5.44.0 compiles
// `my $y; ++ +$y;`.
func TestIncrementOverPrefixCanon(t *testing.T) {
	for src, want := range map[string]string{"my $y; ++ +$y;": "my $y;++ +$y;", "my $y; --+$y;": "my $y;-- +$y;"} {
		if got := clean(t, src); got != want {
			t.Errorf("%q: canon %q, want %q", src, got, want)
		}
	}
	if ok, why := canonMatches([]byte("+++y=>")); !ok {
		t.Errorf("`+++y=>`: canon is not a fixpoint: %s", why)
	}
}

// TestNumberTakesNoSubscript: perl 5.44.0 rejects `A{},(0{},)` near "0{".
// Read as a hash subscript on 0, its canon doubled the trailing comma and
// re-read differently. A list slice, `(1)[0]`, is no number.
func TestNumberTakesNoSubscript(t *testing.T) {
	for _, src := range []string{"A{},(0{},)", "(0{},)", "(0[1])"} {
		if found := collectUnknown(parse.Parse([]byte(src))); len(found) == 0 {
			t.Errorf("%q: perl rejects this, but it parses", src)
		}
	}
	for _, src := range []string{"my @a = (1)[0];", "my %x; ($x{a},)", "my @x; ($x[0],)"} {
		clean(t, src)
	}
}
