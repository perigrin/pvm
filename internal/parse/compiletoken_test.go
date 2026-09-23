// ABOUTME: `__PACKAGE__` and friends are all-caps but are NOT filehandles, and print must not take them as one.
// ABOUTME: perl draws the line itself: `print FOO, 1` is a syntax error and `print __PACKAGE__, 1` is not.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestCompileTokensAreNotFilehandles: a compile-time token in `print`'s
// first slot is an ARGUMENT, not a handle.
//
// THE BUG THIS PINS. `isBarewordHandle` accepts any all-caps word,
// because perl's filehandle convention is all-caps and a static parser
// has no symbol table to consult. `__PACKAGE__`, `__LINE__`, `__FILE__`
// and `__CLASS__` are all-caps, so all four were taken into the handle
// slot -- and the comma that follows them then had nothing to attach to,
// which surfaced as `trailing_tokens` over the whole statement.
//
// PERL DRAWS THE LINE ITSELF, and the two readings are not both legal:
//
//	$ perl -MO=Deparse -e 'print FOO, "\n";'
//	No comma allowed after filehandle at -e line 1.
//
//	$ perl -MO=Deparse -e 'print __PACKAGE__, "\n";'
//	print __PACKAGE__, "\n";
//
// So the comma after a real bareword handle is an ERROR, and the comma
// after a compile-time token is ordinary. A parser that cannot tell them
// apart gets one of the two wrong whichever way it guesses.
//
// These four are the whole list perl documents as compile-time tokens
// that lex as barewords. `__DATA__` and `__END__` are not here: they
// terminate the program text rather than producing a value, and tier 13
// measures them as their own token kind.
func TestCompileTokensAreNotFilehandles(t *testing.T) {
	for _, src := range []string{
		`print __PACKAGE__, "\n";`,
		`print __LINE__, "\n";`,
		`print __FILE__, "\n";`,
		`print __LINE__, " ", __FILE__, "\n";`,
		`print __PACKAGE__;`,
		`print __CLASS__;`,
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q must parse: %v", src, kinds(root))
		}
	}
}

// TestBarewordHandleStillTakesTheSlot: the fix must not cost the handle.
//
// The narrow change is that FOUR NAMES stop qualifying, not that the
// all-caps rule goes away. `print STDERR "x"` has no comma and is the
// form the slot exists for; it must still parse with the handle taken.
func TestBarewordHandleStillTakesTheSlot(t *testing.T) {
	for _, src := range []string{
		`print STDERR "x";`,
		`print STDOUT "x";`,
		`printf STDERR "%s", "x";`,
		`print FH "x";`,
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q must parse: %v", src, kinds(root))
		}
	}
}
