// ABOUTME: The first `{` inside an XREF block (map, grep, sort, the print family) is an anonymous hash.
// ABOUTME: toke.c's block_expectation; a first brace anywhere else stays a statement-start brace.

package lexer

import "testing"

// TestBlockExpectationBraceIsAnonHash: measured with -DT on the pinned perl
// (94e50866), each last `{` here reads HASHBRACK. A block opened in XREF
// leaves XTERM, not XSTATE, when a `{` comes next (toke.c block_expectation),
// so the inner brace is a hash whatever is inside it. `map { { 1; } } @x` is
// a syntax error for the same reason: the `;` ends a hash's term list.
func TestBlockExpectationBraceIsAnonHash(t *testing.T) {
	for _, src := range []string{
		`map { { $_->get_columns } } @x`,
		`map { {1} } @x`,
		`map { { 1; } } @x`,
		`map { {; 1} } @x`,
		`map { {} } @x`,
		"map {\n{ 1 } } @x",
		"map { # c\n { 1 } } @x",
		`grep { { a=>1 } } @x`,
		`sort { { 1 } } @x`,
		`map({ {1} } @x)`,
		`grep({ {1} } @x)`,
		`print { {1} } 1`,
		`printf { {1} } 1`,
		`system { {1} } 1`,
		`exec { {1} } 1`,
		`CORE::say { {1} } 1`,
	} {
		open, closer := lastBraceAndCloserOpenBlock(t, src)
		if open || closer {
			t.Errorf("%q: { OpensBlock=%v, } OpensBlock=%v; want both false (HASHBRACK)", src, open, closer)
		}
	}
}

// TestBlockExpectationBraceOpensBlock: the negative half, each measured with
// -DT on the pinned perl. A first brace after anything but an XREF block's `{`
// is a statement-start brace, which braceIsHash decides; so is one that is not
// the block's first token.
func TestBlockExpectationBraceOpensBlock(t *testing.T) {
	for _, src := range []string{
		`map {; {1} } @x`,
		`map { 1; {1} } @x`,
		`say { {1} } 1`,
		`sub foo(&@){} foo { { 1 } } 1`,
		`do { { 1 } }`,
		`eval { { 1 } }`,
		`if (1) { { 1 } }`,
		`sub f { { 1 } }`,
	} {
		open, closer := lastBraceAndCloserOpenBlock(t, src)
		if !open || !closer {
			t.Errorf("%q: { OpensBlock=%v, } OpensBlock=%v; want both true (block)", src, open, closer)
		}
	}
}

// TestBlockExpectationBraceSubColon: block_expectation's other term, `sub`
// then `:`. -DT reads `map { sub : lvalue { 1 } } @x` as KW_SUB_anon COLONATTR,
// an anonymous sub with an attribute, where `map {; sub : lvalue ...` reads
// LABEL. So `sub` here is not a label word, and `lvalue` is the attribute.
func TestBlockExpectationBraceSubColon(t *testing.T) {
	for _, c := range []struct {
		src  string
		attr bool
	}{
		{`map { sub : lvalue { 1 } } @x`, true},
		{`map { sub: lvalue { 1 } } @x`, true},
		{`map {; sub : lvalue { 1 } } @x`, false},
	} {
		for _, tok := range Tokenize([]byte(c.src)) {
			if c.src[tok.Start:tok.End] == "lvalue" && tok.AttrName != c.attr {
				t.Errorf("%q: lvalue AttrName=%v; want %v", c.src, tok.AttrName, c.attr)
			}
		}
	}
}
