// ABOUTME: A `{` after `use`/`no` MODULE is an anonymous hash, as toke.c reads the import list as a term.
// ABOUTME: A `{` opening a block after a complete `use M;` statement is unaffected.

package lexer

import "testing"

// TestUseModuleBraceIsAnonHash: measured with -DT on the pinned perl, each
// last `{` here reads HASHBRACK (KW_USE_or_NO BAREWORD BAREWORD HASHBRACK).
func TestUseModuleBraceIsAnonHash(t *testing.T) {
	for _, src := range []string{
		`use strict { () };`,
		`use test_use { () };`,
		`no strict { () };`,
		`use strict 1.0 { () };`,
		`use strict {};`,
		`use Foo::Bar { a => 1 };`,
	} {
		open, closer := lastBraceAndCloserOpenBlock(t, src)
		if open || closer {
			t.Errorf("%q: { OpensBlock=%v, } OpensBlock=%v; want both false (HASHBRACK)", src, open, closer)
		}
	}
}

// TestUseModuleBraceAfterStatementOpensBlock: the negative half. Once the
// `use` statement is complete, a `{` at statement start is a bare block.
func TestUseModuleBraceAfterStatementOpensBlock(t *testing.T) {
	for _, src := range []string{
		`use strict; { 1; }`,
		`no strict; { 1; }`,
		`use strict 1.0; { 1; }`,
		"use Foo;\n{ 1; }",
	} {
		open, closer := lastBraceAndCloserOpenBlock(t, src)
		if !open || !closer {
			t.Errorf("%q: { OpensBlock=%v, } OpensBlock=%v; want both true (block)", src, open, closer)
		}
	}
}

// TestUseModuleBraceVersionKnownLimit pins useArg's shortcut: toke.c expects a
// term after the version too (-DT reads `use test_use 1.0 %h;` with
// PERLY_PERCENT_SIGN, a hash), but the lexer leaves an operator expected, so
// the `%` is modulus. A `{` there is no block all the same
// (TestUseModuleBraceIsAnonHash). Fails once the limit
// goes; move the row to a positive test then.
func TestUseModuleBraceVersionKnownLimit(t *testing.T) {
	// perl -c reads a term after each version: `%h` a hash, `<FH>` a readline, `.5` a
	// number (all syntax OK), `/ 3;` an unterminated pattern (refused). The lexer reads
	// an operator.
	for _, c := range []struct{ src, op string }{
		{`use test_use 1.0 %h;`, "%"},
		{`use test_use 1.0 <FH>;`, "<"},
		{`use test_use 1.0 .5;`, "."},
		{`use test_use 1.0 / 3;`, "/"},
	} {
		if got := significant(Tokenize([]byte(c.src)))[3]; got.Kind != Operator || c.src[got.Start:got.End] != c.op {
			t.Errorf("%q: token after the version is %v %q; the recorded limit moved", c.src, got.Kind, c.src[got.Start:got.End])
		}
	}
}
