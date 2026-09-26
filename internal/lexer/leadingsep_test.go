// ABOUTME: A leading `::` belongs to the NAME -- `::ok` is one Word, the `main::ok` shorthand.
// ABOUTME: `::` with no name after it is perl's bareword string `'::'`, so the name byte is required.

package lexer_test

import "testing"

// TestLeadingPackageSeparatorInWord: `::ok` is ONE Word token.
//
// `scanIdentRunes`'s separator branch sits behind an `if !first` guard, so a
// `::` at the start of a name was never consumed: `::ok(1);` lexed as
// `Operator(::) Word(ok) ...` and the parser refused -- `trailing_tokens` as a
// statement, `not_a_term` as a term. The explicit `main::ok(1)` was always
// fine, which is why only the elided package name was affected.
//
// The sigil spelling is a DIFFERENT code path and was already correct:
// `scanVarName` has an explicit `leadingPackageSeparator` case, so `$::x`,
// `@::y` and `%::z` lex as one Variable each. This test covers the two paths
// that route through `scanIdentRunes`: a bareword, and the name after a `&`
// function sigil.
//
// Measured on perl 5.42.0:
//
//	$ perl -e 'sub main::ok {print "A\n"} ::ok(1);'          -> A
//	$ perl -e 'sub main::f {print "B\n"} &::f();'            -> B
//	$ perl -e 'sub main::f {1} my $c = \&::f; print $c->()'  -> 1
//	$ perl -e 'sub Foo::bar {print "C\n"} ::Foo::bar(1);'    -> C
//	$ perl -MO=Deparse -e 'sub main::ok {1} ::ok(1);'        -> ok 1;
func TestLeadingPackageSeparatorInWord(t *testing.T) {
	for _, c := range []struct {
		src  string
		want []string
	}{
		{"::ok", []string{"Word(::ok)"}},
		{"::ok(1);", []string{"Word(::ok)", "Operator(()", "Number(1)", "CloseBracket())", "Semicolon(;)"}},
		{"::ok 1;", []string{"Word(::ok)", "Number(1)", "Semicolon(;)"}},

		// A leading separator and an interior one compose: `::Foo::bar` is
		// `main::Foo::bar`, one name.
		{"::Foo::bar", []string{"Word(::Foo::bar)"}},

		// The `&` sigil is its own token; the name after it goes through the
		// same scanner, so it gets the same rule.
		{"&::f();", []string{"FuncSigil(&)", "Word(::f)", "Operator(()", "CloseBracket())", "Semicolon(;)"}},
		{`\&::f`, []string{`Operator(\)`, "FuncSigil(&)", "Word(::f)"}},

		// The explicit form must not move.
		{"main::ok(1);", []string{"Word(main::ok)", "Operator(()", "Number(1)", "CloseBracket())", "Semicolon(;)"}},

		// A name byte after the colons is REQUIRED, the same condition
		// `leadingPackageSeparator` already applies to `$::x`. Measured:
		// perl reads a bare `::` as the bareword STRING `'::'`, so there is
		// a real reading here to preserve rather than a name to take.
		//
		//	$ perl -MO=Deparse -e 'my $x = ::;'   ->  my $x = '::';
		//	$ perl -MO=Deparse -e '::1;'          ->  '???';
		{"my $x = ::;", []string{"Word(my)", "Variable($x)", "Operator(=)", "Operator(::)", "Semicolon(;)"}},
		{"::1;", []string{"Operator(::)", "Number(1)", "Semicolon(;)"}},

		// A single colon never starts a name -- it is the ternary arm, a
		// label boundary, or an attribute introducer.
		{":ok", []string{"Operator(:)", "Word(ok)"}},
	} {
		if !streamIs(c.src, c.want...) {
			t.Errorf("%q lexes as %v, want %v", c.src, significant(c.src), c.want)
		}
	}
}
