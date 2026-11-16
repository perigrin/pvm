// ABOUTME: A `{` where a statement may begin is a hash when toke.c's yyl_leftcurly default case says so.
// ABOUTME: Empty, or a first term followed by `=>`, or by `,` unless the term starts lower-case.

package lexer

import (
	"strings"
	"testing"
)

// lastBraceAndCloserOpenBlock reports OpensBlock on the last `{` in src and
// on the `}` that closes it.
func lastBraceAndCloserOpenBlock(t *testing.T, src string) (open, closer bool) {
	t.Helper()
	at := strings.LastIndexByte(src, '{')
	depth := 0
	for _, tok := range Tokenize([]byte(src)) {
		if tok.End-tok.Start != 1 || tok.Start < at {
			continue
		}
		switch src[tok.Start] {
		case '{':
			if tok.Start == at {
				open = tok.OpensBlock
			}
			depth++
		case '}':
			depth--
			if depth == 0 {
				return open, tok.OpensBlock
			}
		}
	}
	t.Fatalf("%q: no matched brace", src)
	return
}

// TestStatementStartBraceIsAnonHash: measured with -DT on the pinned perl,
// each last `{` here reads HASHBRACK (toke.c:6715 and 6810-6814).
func TestStatementStartBraceIsAnonHash(t *testing.T) {
	for _, src := range []string{
		`{};`,
		`{ 1, 2 };`,
		`{ foo => 1 };`,
		`{ "a", 1 };`,
		`do { {1=>2} };`,
		`my $f = sub { {} };`,
		`{ Foo, 1 };`,
		`{ q(a), 1 };`,
		`{ q => 1 };`,
		"{ # c\n a => 1 };",
		`L: { a => 1 };`,
		`{ qux, 1 };`,
		`{ , 1 };`,
		// Only the skip before the first term crosses a line.
		"{\n a => 1 };",
		// map runs the same lookahead.
		`map { 1, 2 } @a;`,
	} {
		open, closer := lastBraceAndCloserOpenBlock(t, src)
		if open || closer {
			t.Errorf("%q: `{` OpensBlock=%v, `}` OpensBlock=%v; want an anonymous hash (false, false)", src, open, closer)
		}
	}
}

// TestStatementStartBraceStaysBlock: every other statement-start `{` is a
// block, and so is any `{` toke.c reaches at XBLOCK.
func TestStatementStartBraceStaysBlock(t *testing.T) {
	for _, src := range []string{
		`{ foo(); }`,
		`{ $x = 1; }`,
		`{ foo, 1 }`,
		`sub f { 1 }`,
		`if ($x) { 1 }`,
		`{ -foo => 1 }`,
		`sub f { }`,
		`if ($x) { a => 1 }`,
		`{ qw(a), 1 };`,
		`{ 1 };`,
		`map { foo, 1 } @a;`,
		// toke.c's PREBLOCK keywords leave XBLOCK: no lookahead.
		`try { 1, 2 }`,
		`try { 1 } catch ($e) { 3, 4 }`,
		`if (1) { } else { a => 1 }`,
		`defer { a => 1 }`,
		// toke.c runs the lookahead only in its default case, XSTATE or an
		// XREF list operator's brace. An anonymous sub's brace is XBLOCK,
		// and after any other word the -DT reads PERLY_BRACE_OPEN.
		`sub { 1, 2 };`,
		`my $f = sub { 1, 2 };`,
		`foo(sub { 1, 2 });`,
		`return sub { 1, 2 };`,
		`my $f = sub { a => 1 };`,
		`zzz { 1, 2 } print "b";`,
		`foo { 1, 2 };`,
		`Foo { 1, 2 };`,
		`any { 1, 2 } @a;`,
		`first { 1, 2 } @a;`,
		// The lookahead stops at the end of the first term's line: in a file
		// each of these is PERLY_BRACE_OPEN.
		"{ A\n => 1 };",
		"{ \"a\nb\", 1 };",
		"{ q(a\nb), 1 };",
		"{ A\n, 1 };",
		// shortcut pinned: ASCII word characters only, so a `use utf8` key
		// reads as a block where perl reads HASHBRACK. See braceIsHash.
		"use utf8; { \u00e9 => 1 };",
		// shortcut pinned: a declared sub named try, catch, finally, default
		// or defer still takes a block; perl reads HASHBRACK. See preblock.
		`sub try {} try { "a", 1 };`,
		`sub defer {} defer { "a", 1 };`,
	} {
		open, closer := lastBraceAndCloserOpenBlock(t, src)
		if !open || !closer {
			t.Errorf("%q: `{` OpensBlock=%v, `}` OpensBlock=%v; want a block (true, true)", src, open, closer)
		}
	}
}
