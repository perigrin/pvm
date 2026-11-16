// ABOUTME: A sub body's { opens a block after a prototype or an attribute list, an empty one included.
// ABOUTME: toke.c leaves XATTRTERM after an anonymous sub's prototype and XBLOCK after COLONATTR.

package lexer

import "testing"

// TestSubBodyBraceOpensBlock: measured with -DT on the pinned perl, the `{`
// of each reads PERLY_BRACE_OPEN, from XATTRTERM after an anonymous sub's
// PROTOTYPE and from XBLOCK after a named sub's COLONATTR.
func TestSubBodyBraceOpensBlock(t *testing.T) {
	for _, src := range []string{
		`sub ($$) { 1 }`,
		`sub () { 1 }`,
		`my $c = sub ($$) { a => 1 };`,
		`sub ($$) {}`,
		`sub foo : { 1 }`,
		`sub foo :lvalue : { 1 }`,
		`sub foo : { a => 1 }`,
		`my $c = sub ($) : { 1 };`,
	} {
		open, closer := braceAndCloserOpenBlock(t, src)
		if !open || !closer {
			t.Errorf("%q: `{` OpensBlock=%v, `}` OpensBlock=%v; want a block (true, true)", src, open, closer)
		}
	}
}

// TestSubBodyBraceAnonIsTerm: an anonymous sub is a term, so its body's `}`
// leaves an operator expected, as toke.c's XATTRTERM does: the `/` divides.
func TestSubBodyBraceAnonIsTerm(t *testing.T) {
	src := `my $c = sub ($$) { 1 } / 2;`
	for _, tok := range Tokenize([]byte(src)) {
		if src[tok.Start:tok.End] == "/" {
			if tok.Kind != Operator {
				t.Errorf("%q: `/` is %v; want Operator", src, tok.Kind)
			}
			return
		}
	}
	t.Errorf("%q: no `/` token", src)
}

// TestSubBodyBraceLeavesHashAlone: the same neighbours outside a sub head
// keep the anonymous hash.
func TestSubBodyBraceLeavesHashAlone(t *testing.T) {
	for _, src := range []string{
		`$c ? foo : { a => 1 }`,
		`my $h = { a => 1 }`,
	} {
		open, closer := braceAndCloserOpenBlock(t, src)
		if open || closer {
			t.Errorf("%q: `{` OpensBlock=%v, `}` OpensBlock=%v; want an anonymous hash (false, false)", src, open, closer)
		}
	}
}
