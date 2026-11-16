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
// Measured with -DT on the pinned perl, each reads XOPERATOR after the `}`,
// an attribute word in the head included.
func TestSubBodyBraceAnonIsTerm(t *testing.T) {
	for _, src := range []string{
		`my $c = sub ($$) { 1 } / 2;`,
		`my $c = sub :lvalue : { 1 } / 2;`,
		`my $c = sub ($) :lvalue : { 1 } / 2;`,
		`my $c = sub ($$) : lvalue { 1 } / 2;`,
		`my $c = sub :lvalue { 1 } / 2;`,
	} {
		found := false
		for _, tok := range Tokenize([]byte(src)) {
			if src[tok.Start:tok.End] == "/" {
				found = true
				if tok.Kind != Operator {
					t.Errorf("%q: `/` is %v; want Operator", src, tok.Kind)
				}
			}
		}
		if !found {
			t.Errorf("%q: no `/` token", src)
		}
	}
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

// TestSubBodyBraceAttrName: an attribute's name in a sub head is marked, as
// toke.c's yyl_colon scans it as ATTRLIST; the same word after a ternary's
// colon is not.
func TestSubBodyBraceAttrName(t *testing.T) {
	for src, want := range map[string]bool{
		`sub foo :lvalue { 1 }`:            true,
		`my $c = sub ($) : lvalue { 1 };`:  true,
		`my $c = sub :lvalue : { 1 } / 2;`: true,
		`$c ? f : lvalue { 1 }`:            false,
	} {
		for _, tok := range Tokenize([]byte(src)) {
			if src[tok.Start:tok.End] == "lvalue" && tok.AttrName != want {
				t.Errorf("%q: lvalue AttrName=%v; want %v", src, tok.AttrName, want)
			}
		}
	}
}
