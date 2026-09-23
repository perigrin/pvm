// ABOUTME: A word OPERATOR after a handle candidate does not start a term, so it denies the filehandle slot.
// ABOUTME: `print FOO eq "x"` is a comparison in perl, not a print to FOO, and the same holds for `print $fh eq "x"`.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestWordOperatorDeniesTheFilehandleSlot: `eq` is a Word and is not a
// term.
//
// THE BUG. `startsTerm` answered true for every `lexer.Word`, so any
// bareword operator -- `eq`, `ne`, `cmp`, `lt`, `x`, `and`, `or` -- read
// as the start of a new term. Both filehandle branches consult it, so
// both took the slot when they should not have:
//
//	print FOO eq "x" ? "a" : "b";      read FOO as a handle
//	print $fh eq "x" ? "a" : "b";      read $fh as a handle
//
// Measured on perl 5.42.0, NEITHER is a print to a handle:
//
//	$ perl -MO=Deparse -e 'print FOO eq "x" ? "a" : "b";'
//	print 'b';
//
//	$ perl -MO=Deparse -e 'print __PACKAGE__ eq "x" ? "a" : "b";'
//	print 'b';
//
// perl compares and prints the result. The bareword is the comparison's
// left operand, not a destination.
//
// THE FIX IS THE TABLE WE ALREADY HAVE. `infix` in precedence.go holds
// every one of perl's 32 levels including the word-spelled operators, so
// a Word that is an infix operator is not a term-starter. No new list:
// a new word operator arrives in one place and both callers follow.
//
// The `$fh` half was a PRE-EXISTING bug, older than the bareword half
// that exposed it -- `print $fh eq "x"` refused before any of this
// milestone's changes. Fixing `startsTerm` fixes both.
func TestWordOperatorDeniesTheFilehandleSlot(t *testing.T) {
	for _, src := range []string{
		`print FOO eq "x" ? "a" : "b";`,
		`print $fh eq "x" ? "a" : "b";`,
		`print __PACKAGE__ eq "x" ? "a" : "b";`,
		`print FOO ne "x" ? "a" : "b";`,
		`print FOO lt "x" ? "a" : "b";`,
		`print $fh cmp "x";`,
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q must parse: %v", src, kinds(root))
		}
	}
}

// TestHandleSlotSurvivesTheWordFix: a word that is NOT an operator still
// starts a term, so the slot still exists.
//
// The narrow change is that infix words stop qualifying. A bareword
// argument after a handle is the form the slot is for and must keep
// working.
func TestHandleSlotSurvivesTheWordFix(t *testing.T) {
	for _, src := range []string{
		`print STDERR "x";`,
		`print $fh "x";`,
		`printf STDERR "%s", "x";`,
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q must parse: %v", src, kinds(root))
		}
	}
}
