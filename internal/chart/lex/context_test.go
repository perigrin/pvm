// ABOUTME: Tests chart/lex's context facts: a sub declared earlier is declared later, and a constant sub as the lexer reads it.
// ABOUTME: The constant-sub fact is held to the production lexer's own decision, an operator after the name.

package lex

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/lexer"
)

// TestDeclaredSubLater: `sub foo` declares foo from there on, and not before.
func TestDeclaredSubLater(t *testing.T) {
	x := guacamole(t)
	src := "foo(); sub foo { 1 } foo();"
	s, err := x.Lex([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	first, last := strings.Index(src, "foo"), strings.LastIndex(src, "foo")
	if s.Facts.Declared("foo", first) {
		t.Error("foo is declared before its `sub foo`")
	}
	if !s.Facts.Declared("foo", last) {
		t.Error("foo is not declared after its `sub foo`")
	}
	if s.Facts.Constant("foo", last) {
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
	x := guacamole(t)
	for _, c := range []struct{ decl, name string }{
		{"sub PI () { 3 }", "PI"},
		{"use constant PI => 3;", "PI"},
		{"use constant { PI => 3, E => 2 };", "PI"},
		{"use constant { A => $#{$r}, B => 2 };", "B"},
	} {
		src := c.decl + " my $q = " + c.name + " / 2; # /"
		s, err := x.Lex([]byte(src))
		if err != nil && !strings.Contains(err.Error(), "01a1882d-d8a4") {
			t.Errorf("%s: %v", c.decl, err)
			continue
		}
		at := strings.LastIndex(src, c.name)
		if !s.Facts.Constant(c.name, at) {
			t.Errorf("%s: %s is not a constant sub at byte %d", c.decl, c.name, at)
		}
		slash := strings.Index(src[at:], "/") + at
		for _, tok := range s.Tokens {
			if tok.Start == slash && tok.Kind != lexer.Operator {
				t.Errorf("%s: the lexer read `/` after %s as %v; the fact disagrees with it", c.decl, c.name, tok.Kind)
			}
		}
	}
}
