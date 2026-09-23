// ABOUTME: A bareword takes print's filehandle slot only when NO COMMA follows it, which is perl's own rule.
// ABOUTME: The name of the word is not the test: `print FOO, 1` and `print __CLASS__, 1` are both errors.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestCommaDeniesTheFilehandleSlot: what separates a handle from an
// argument is the COMMA, not the word.
//
// THE BUG THIS PINS. `isBarewordHandle` took any all-caps word, because
// perl's filehandle convention is all-caps and a static parser has no
// symbol table. So `print __PACKAGE__, "\n"` put the token in the handle
// slot and the comma after it had nothing to attach to, which surfaced
// as `trailing_tokens` over the whole statement.
//
// PERL'S RULE IS THE COMMA, measured on 5.42.0:
//
//	$ perl -e 'print FOO, "x";'           No comma allowed after filehandle
//	$ perl -e 'print __CLASS__, "x";'     No comma allowed after filehandle
//	$ perl -e 'print __PACKAGE__, "x";'   syntax OK
//	$ perl -e 'print FOO "x";'            syntax OK -- a handle
//
// A COMMA AFTER A HANDLE IS AN ERROR. So a bareword followed by a comma
// was never a handle, whatever it is called, and the word that follows
// the slot must start a new term for the slot to exist at all.
//
// WHY NOT A LIST OF TOKEN NAMES, which an earlier fix used. It was
// wrong on one entry and incomplete on another, and both faults came
// from the same place -- the list encodes a judgement about each name
// where perl encodes none:
//
//   - `__CLASS__` is a value only under `use feature 'class'` and a
//     bareword FILEHANDLE everywhere else. A list that exempts it
//     unconditionally adopts the gated reading with no feature state to
//     justify it, and contradicts `11_oo/16_classname_ungated.t`, whose
//     whole subject is that ungated `print __CLASS__;` is a filehandle
//     print.
//   - `__SUB__` behaves identically and was simply missing.
//
// The comma test needs neither name and gets both right in both feature
// states, because it is the rule perl actually applies.
func TestCommaDeniesTheFilehandleSlot(t *testing.T) {
	for _, src := range []string{
		`print __PACKAGE__, "\n";`,
		`print __LINE__, "\n";`,
		`print __FILE__, "\n";`,
		`print __LINE__, " ", __FILE__, "\n";`,
		`print __SUB__, "\n";`,
		`print __CLASS__, "\n";`,
		// A plain bareword with a comma is the same shape. perl calls it
		// an error; we decline to call it a handle, which is the part
		// this parser is responsible for.
		`print FOO, "\n";`,
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q must parse: %v", src, kinds(root))
		}
	}
}

// TestBarewordHandleStillTakesTheSlot: no comma, so the slot stands.
//
// The change is that a FOLLOWING COMMA denies the slot, not that the
// all-caps rule goes away. These are the forms the slot exists for and
// every one of them must still take it.
//
// `print __CLASS__;` is here deliberately. Ungated it IS a filehandle
// print -- `print __CLASS__ $_` -- which is what
// `11_oo/16_classname_ungated.t` measures, and no comma follows, so the
// slot is correct for it.
func TestBarewordHandleStillTakesTheSlot(t *testing.T) {
	for _, src := range []string{
		`print STDERR "x";`,
		`print STDOUT "x";`,
		`printf STDERR "%s", "x";`,
		`print FH "x";`,
		`print __CLASS__;`,
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q must parse: %v", src, kinds(root))
		}
	}
}
