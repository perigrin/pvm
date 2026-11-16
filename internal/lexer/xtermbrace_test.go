// ABOUTME: A `{` after a word toke.c leaves at XTERM is an anonymous hash, with no lookahead.
// ABOUTME: bless, return, the XTERM list operators and the named unaries; other words still guess.

package lexer

import "testing"

// braceAndCloserOpenBlock reports OpensBlock on the first `{` and on the `}`
// that closes it.
func braceAndCloserOpenBlock(t *testing.T, src string) (open, closer bool) {
	t.Helper()
	depth, opened := 0, false
	for _, tok := range Tokenize([]byte(src)) {
		if tok.End-tok.Start != 1 {
			continue
		}
		switch src[tok.Start] {
		case '{':
			if !opened {
				open, opened = tok.OpensBlock, true
			}
			depth++
		case '}':
			depth--
			if opened && depth == 0 {
				return open, tok.OpensBlock
			}
		}
	}
	t.Fatalf("%q: no matched brace", src)
	return
}

// TestXTermBraceIsAnonHash: toke.c's yyl_leftcurly returns HASHBRACK at
// XTERM without looking inside (toke.c:6650-6654). bless is
// LOP(OP_BLESS,XTERM), return is OLDLOP, ref is UNI. Measured with -DT on
// the pinned perl, each of these reads HASHBRACK at the `{`.
func TestXTermBraceIsAnonHash(t *testing.T) {
	for _, src := range []string{
		`bless {}, "X"`,
		`bless { a => 1 }, $c`,
		`return { 1, 2 }`,
		`return {}`,
		`return { $x }`,
		`push @a, { 1 }`,
		`die { code => 1 }`,
		`ref { 1 }`,
		`CORE::bless { 1 }, "X"`,
		`atan2 {}, 1`,
		`shift { 1 }`,
	} {
		open, closer := braceAndCloserOpenBlock(t, src)
		if open || closer {
			t.Errorf("%q: `{` OpensBlock=%v, `}` OpensBlock=%v; want an anonymous hash (false, false)", src, open, closer)
		}
	}
}

// TestXTermBraceLeavesOtherWordsAlone: a word toke.c does not leave at XTERM
// still has its brace guessed, or settled by its own rule.
func TestXTermBraceLeavesOtherWordsAlone(t *testing.T) {
	for _, src := range []string{
		`sub f { 1 }`,
		`map { $_ } @x`,
		`if ($x) { 1 }`,
		`foo { 1 }`,
		`sub bless { 1 }`,
		`sub return { 1 }`,
	} {
		open, closer := braceAndCloserOpenBlock(t, src)
		if !open || !closer {
			t.Errorf("%q: `{` OpensBlock=%v, `}` OpensBlock=%v; want a block (true, true)", src, open, closer)
		}
	}
}
