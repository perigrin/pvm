// ABOUTME: Tests chart/lex's context facts: a sub declared earlier is declared later, and a constant sub as the lexer reads it.
// ABOUTME: The constant-sub fact is held to the production lexer's own decision, an operator after the name.

package lex

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/lexer"
)

// factsOf is the facts of src's tokens as the production lexer reads them.
func factsOf(src string) *Context {
	return facts([]byte(src), lexer.TokenizeBarewords([]byte(src), BarewordOperators))
}

// TestDeclaredSubLater: `sub foo` declares foo from there on, and not before.
func TestDeclaredSubLater(t *testing.T) {
	src := "foo(); sub foo { 1 } foo();"
	f := factsOf(src)
	first, last := strings.Index(src, "foo"), strings.LastIndex(src, "foo")
	if f.Declared("foo", first) {
		t.Error("foo is declared before its `sub foo`")
	}
	if !f.Declared("foo", last) {
		t.Error("foo is not declared after its `sub foo`")
	}
	if f.Constant("foo", last) {
		t.Error("foo has no empty prototype, so it is no constant")
	}
}

// TestConstantSubAsTheLexerDecides: each form the production lexer reads as
// a constant sub is one here, and the lexer's own decision follows: after the
// name it expects an operator, so `/ 2` is division and not a pattern.
//
// The facts come from the production lexer's own tokens, before the `$#{`
// split: the split's `}` would close the `{` of `use constant {` early and
// lose B (gate review of #541, F2). That file is refused for its `$#{`
// interior (F1), and its facts still hold.
func TestConstantSubAsTheLexerDecides(t *testing.T) {
	for _, c := range []struct{ decl, name string }{
		{"sub PI () { 3 }", "PI"},
		{"use constant PI => 3;", "PI"},
		{"use constant { PI => 3, E => 2 };", "PI"},
		{"use constant { A => $#{$r}, B => 2 };", "B"},
	} {
		src := c.decl + " my $q = " + c.name + " / 2; # /"
		raw := lexer.TokenizeBarewords([]byte(src), BarewordOperators)
		at := strings.LastIndex(src, c.name)
		if !facts([]byte(src), raw).Constant(c.name, at) {
			t.Errorf("%s: %s is not a constant sub at byte %d", c.decl, c.name, at)
		}
		slash := strings.Index(src[at:], "/") + at
		for _, tok := range raw {
			if tok.Start == slash && tok.Kind != lexer.Operator {
				t.Errorf("%s: the lexer read `/` after %s as %v; the fact disagrees with it", c.decl, c.name, tok.Kind)
			}
		}
	}
}
