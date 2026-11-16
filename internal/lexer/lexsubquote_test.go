// ABOUTME: A lexical sub named like a quote operator shadows it in its block:
// ABOUTME: after `my sub s`, `s(1)` is a call, and after the block `s///` is back.

package lexer

import (
	"slices"
	"testing"
)

// TestLexicalSubShadowsQuoteOp: measured on 5.42.0,
//
//	$ perl -e '{ my sub s { 42 } print s(1), "\n"; } $_ = "a"; s/a/b/;
//	      print "$_\n"; { my sub y { 43 } print y, "\n"; }'
//	42
//	b
//	43
//
// perl.git t/op/lvref.t:653 (`my sub s ($arg)`) and t/op/lexsub.t:869
// (`my sub y :prototype() {$x}; is y, 43`).
func TestLexicalSubShadowsQuoteOp(t *testing.T) {
	src := `{ my sub s { 42 } print s(1); } s/a/b/; { my sub y { 43 } is y, 43; }`
	var words, quotes []string
	for _, tok := range Tokenize([]byte(src)) {
		text := src[tok.Start:tok.End]
		switch tok.Kind {
		case Word:
			if text == "s" || text == "y" {
				words = append(words, text)
			}
		case Quote:
			quotes = append(quotes, text)
		}
	}
	// `sub s`, the call `s`, `sub y` and the call `y` are words; `s/a/b/`
	// outside the first block is the one substitution.
	if len(words) != 4 {
		t.Errorf("want s, s, y, y as words, got %q", words)
	}
	if len(quotes) != 1 || quotes[0] != "s/a/b/" {
		t.Errorf("want the one substitution s/a/b/ outside the block, got %q", quotes)
	}
}

// TestLexicalSubShadowsKeyword: toke.c checks the pad before the keyword
// table wherever it does not expect an operator (yyl_keylookup), so a
// lexical sub overrides every keyword there: perl.git t/op/lexsub.t:72,
// `our sub if() { 42 } my $x = if if if;`, is `$x = if() if if()`. The
// word is marked LexSub; the declaration's own name, the `if` after a
// niladic call, and an `if` after the sub's block has closed are not.
func TestLexicalSubShadowsKeyword(t *testing.T) {
	src := `{ our sub if() { 42 } my $x = if if if; } if (1) { 2 }`
	var got []bool
	for _, tok := range Tokenize([]byte(src)) {
		if tok.Kind == Word && src[tok.Start:tok.End] == "if" {
			got = append(got, tok.LexSub)
		}
	}
	want := []bool{false, true, false, true, false}
	if len(got) != len(want) {
		t.Fatalf("want %d ifs, got %v", len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("if %d: LexSub %v, want %v (all: %v)", i, got[i], want[i], got)
		}
	}
}

// TestLexicalSubNotLookedUp: where toke.c reads a word without consulting
// the pad, a lexical sub of that name does not shadow it, so the word is not
// LexSub. Each row's verdicts are the pinned perl's (94e50866):
//   - inside its own body a my, state or our sub is not yet in scope:
//     `sub foo {"P"} my sub foo { "L" . (@_ ? "" : foo(1)) } print foo`
//     prints LP, and `my sub lc { "L:" . lc("AB") }` calls CORE::lc; after
//     a forward declaration it is (`my sub foo; sub foo { foo(1) }`);
//   - a word before `=>` is a string, as toke.c checks the fat comma first;
//   - a lone word in a subscript is a string: `$h{foo}`, `$h{ foo }`, but
//     not `$h{+foo}`, nor `$h{foo` and a newline (perl prints 1, 2, 2 for
//     `$h{foo\n}, $h{ foo }, $h{foo\t}` with `my sub foo {"L"}` and
//     `%h = (L => 1, foo => 2)`);
//   - a label, and the label after last, next, redo or goto.
func TestLexicalSubNotLookedUp(t *testing.T) {
	for _, tc := range []struct {
		src, word string
		want      []bool
	}{
		{`my sub foo { foo(1) } foo;`, "foo", []bool{false, false, true}},
		{`state sub foo { foo(1) } foo;`, "foo", []bool{false, false, true}},
		{`our sub foo { foo(1) } foo;`, "foo", []bool{false, false, true}},
		{`my sub lc { lc("AB") } lc;`, "lc", []bool{false, false, true}},
		{`my sub foo; sub foo { foo(1) } foo;`, "foo", []bool{false, false, true, true}},
		{`{ my sub foo { 1 } } foo;`, "foo", []bool{false, false}},
		{`my sub foo { 1 } my %h = (foo => 1, foo=>2);`, "foo", []bool{false, false, false}},
		{`my sub foo { 1 } $h{foo}; $h{ foo }; $h{+foo};`, "foo", []bool{false, false, false, true}},
		// toke.c autoquotes only across spaces and tabs (yyl_leftcurly's
		// SPACE_OR_TAB): before a newline the word is the lexical sub's call.
		{"my sub foo { 1 } $h{foo\t}; $h{foo\n};", "foo", []bool{false, false, true}},
		{"my sub foo { 1 } foo: { last foo; next foo; redo foo; goto foo; } foo;", "foo",
			[]bool{false, false, false, false, false, false, true}},
	} {
		var got []bool
		for _, tok := range Tokenize([]byte(tc.src)) {
			if tok.Kind == Word && tc.src[tok.Start:tok.End] == tc.word {
				got = append(got, tok.LexSub)
			}
		}
		if !slices.Equal(got, tc.want) {
			t.Errorf("%s: LexSub on each %s %v, want %v", tc.src, tc.word, got, tc.want)
		}
	}
}
