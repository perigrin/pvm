// ABOUTME: The caret control variables -- `$^O`, `$^W`, `$^_` -- are ONE token.
// ABOUTME: `$^` alone is a real punctuation variable, so the caret only binds a name when one follows.

package lexer_test

import "testing"

// TestCaretControlVariable: `$^O` is one Variable token.
//
// `scanVarName`'s punctuation-variable case takes exactly one byte, so
// without the caret branch above it `$^O` lexes as `Variable($^)` and a
// separate `Word(O)`. Perl reads it as one variable:
//
//	$ perl -MO=Deparse -e 'my $x = $^O;'
//	my $x = $^O;
//
// The Unknown count could not see this. Four of six spellings scored
// Unknown=0 on the wrong tree -- `my $x = $^O;` canonicalised to `$^;O()`, a
// punctuation variable followed by a call to `O()`, which is a parseable
// statement and so drew no refusal. That is why this test asserts the TOKEN
// STREAM rather than a node count.
//
// The braced sibling `${^TAINT}` was already right: `bracedNameFollowsAt` has
// a caret branch added when `TestLexDotTGoldenStream` caught `${^TEST}`
// splitting. One spelling of the rule was repaired and the bare one was left.
func TestCaretControlVariable(t *testing.T) {
	// Measured on perl 5.42.0 -- every uppercase letter, `_`, `^` and `[`
	// after `$^` compiles, so the caret binds the character that follows it.
	for _, c := range []struct {
		src  string
		want []string
	}{
		{"$^O", []string{"Variable($^O)"}},
		{"$^W", []string{"Variable($^W)"}},
		{"$^A", []string{"Variable($^A)"}},
		{"$^X", []string{"Variable($^X)"}},
		{"$^_", []string{"Variable($^_)"}},
		{"my $x = $^O;", []string{
			"Word(my)", "Variable($x)", "Operator(=)", "Variable($^O)", "Semicolon(;)",
		}},
		{"$^O eq 'VMS'", []string{
			"Variable($^O)", "Word(eq)", "Quote('VMS')",
		}},

		// All four sigils take the caret name. Measured: `@^O`, `%^O` and
		// `&^O` all compile -- `&^O` reaches runtime as
		// `Undefined subroutine &main::` with a control character for a
		// name, so the caret went into the name there too.
		{"@^O", []string{"Variable(@^O)"}},

		// `$^[` is the least obvious member of the accept set, because `[`
		// also opens a subscript. There is no "`$^` subscripted" reading to
		// preserve: perl takes the bracket into the NAME, so `$^[0]` is a
		// syntax error rather than element 0 of `@^`. Measured:
		//
		//	$ perl -e 'my $x = $^[;'      compiles
		//	$ perl -e 'my $x = $^[1];'    Number found where operator expected
		{"$^[", []string{"Variable($^[)"}},
	} {
		if !streamIs(c.src, c.want...) {
			t.Errorf("%q lexes as %v, want %v",
				c.src, significant(c.src), c.want)
		}
	}
}

// TestBareCaretIsStillOneToken is the negative. `$^` ALONE is the format
// top-of-page name, a real punctuation variable, so the caret must not reach
// past a character that cannot be part of a name.
//
// Measured on perl 5.42.0:
//
//	$ perl -e '$^ = "foo"; print $^;'    compiles
//	$ perl -e 'my $x = $^o;'             Bareword found where operator expected
//	$ perl -e 'my $x = $^1;'             Number found where operator expected
//
// A LOWERCASE letter after `$^` is NOT part of the name -- perl reads `$^`
// and then a bareword, which is a syntax error in that position but proves
// where the token ends. Digits behave the same. So the rule is uppercase,
// `_`, `^` and `[`, not "any identifier byte".
func TestBareCaretIsStillOneToken(t *testing.T) {
	for _, c := range []struct {
		src  string
		want []string
	}{
		{"$^", []string{"Variable($^)"}},
		{"$^ = 1", []string{"Variable($^)", "Operator(=)", "Number(1)"}},
		{"$^;", []string{"Variable($^)", "Semicolon(;)"}},

		// Lowercase: `$^` then a word, which is what perl reports.
		{"$^o", []string{"Variable($^)", "Word(o)"}},

		// A digit is not part of the name either.
		{"$^1", []string{"Variable($^)", "Number(1)"}},
	} {
		if !streamIs(c.src, c.want...) {
			t.Errorf("%q lexes as %v, want %v",
				c.src, significant(c.src), c.want)
		}
	}
}

// TestBracedCaretStillLexes is the regression guard for the sibling that
// already worked. `${^TAINT}` must keep lexing as it does today -- widening
// the bare form must not disturb the braced one.
func TestBracedCaretStillLexes(t *testing.T) {
	for _, src := range []string{"${^TAINT}", "${^UNICODE}", "${^TEST}"} {
		if got := significant(src); len(got) != 1 {
			t.Errorf("%q lexes as %v, want one token", src, got)
		}
	}
}
