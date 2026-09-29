// ABOUTME: A leading apostrophe in a declared sub or format name is the old package
// ABOUTME: separator, not a string: `sub 'Hello'_he_said` names Hello::_he_said.

package lexer

import "testing"

// TestDeclNameLeadingApostrophe: toke.c scans the name after `sub` and
// `format` with scan_word, whose parse_ident takes `'` before an identifier
// start as `::` in the first position as in any other. Measured on 5.42.0:
//
//	$ perl -MO=Deparse -e "no warnings; sub 'Hello'_he_said (_);"
//	sub Hello::_he_said (_);
//
// and a file with "format 'one =", a picture line and "." deparses as
// `format one =`. perl.git t/comp/parser.t:480 and :506.
func TestDeclNameLeadingApostrophe(t *testing.T) {
	for src, name := range map[string]string{
		"sub 'Hello'_he_said (_);":       "'Hello'_he_said",
		"format 'one =\nok @<<\n$t\n.\n": "'one",
	} {
		found := false
		for _, tok := range Tokenize([]byte(src)) {
			text := src[tok.Start:tok.End]
			if tok.Kind == Word && text == name {
				found = true
			}
			if tok.Kind == Quote || tok.Kind == UnknownRest {
				t.Errorf("%q: got %v %q; the apostrophe is a separator", src, tok.Kind, text)
			}
		}
		if !found {
			t.Errorf("%q: no Word %q", src, name)
		}
	}
	// Anywhere else a leading apostrophe opens a string.
	if _, ok := firstOfKind("print 'x';", Quote); !ok {
		t.Errorf("`print 'x'` prints a string")
	}
}
