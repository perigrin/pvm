// ABOUTME: Whitespace and comments may stand between a sigil and its name, as
// ABOUTME: toke.c's scan_ident skips them: `my $ bits` declares $bits.

package lexer

import "testing"

// TestSpaceAfterSigil: scan_ident skips space after the sigil,
//
//	if (isSPACE(*s) || !*s)
//	    s = skipspace(s);
//
// and skipspace skips comments too. Measured on 5.42.0:
//
//	$ perl -MO=Deparse -e 'my @ a = (1); my % h = (); my $ # c
//	    b = 2; print $ b, @ a;'
//	my(@a) = 1;
//	my(%h) = ();
//	my $b = 2;
//	print $b, @a;
//
// perl.git t/op/caller.t:279 has `my $ bits`, and op/signatures.t's t085 to
// t087 spell signature elements across lines and comments. The token keeps
// the bytes between, so the source round-trips.
func TestSpaceAfterSigil(t *testing.T) {
	for _, tc := range []struct{ src, want string }{
		{"my $ bits = 1;", "$ bits"},
		{"my @ a = (1);", "@ a"},
		{"my % h = ();", "% h"},
		{"my $ # c\n b = 2;", "$ # c\n b"},
	} {
		toks := significant(Tokenize([]byte(tc.src)))
		if len(toks) < 2 || toks[1].Kind != Variable || tc.src[toks[1].Start:toks[1].End] != tc.want {
			t.Errorf("%q: want one Variable %q", tc.src, tc.want)
		}
	}
	// No name after the space: the sigil does not reach across to an
	// operator. `$a % $b` is still a modulus.
	src := "my $n = $a % $b;"
	toks := significant(Tokenize([]byte(src)))
	found := false
	for _, tok := range toks {
		if src[tok.Start:tok.End] == "%" && tok.Kind == Operator {
			found = true
		}
	}
	if !found {
		t.Errorf("%q: `%%` after a term is the modulus operator", src)
	}
}

// TestSigilAtEOFIsAnError is the parser fuzzer's `$`: a sigil that ends the
// file has no name, and perl 5.42 says so -- `$`, `@`, `print $` and
// `$x = $` are each "syntax error at EOF" under `perl -c`. Lexed as a
// Variable, canon wrote `$` back with a `;` after it, and `$;` is the
// subscript-separator variable, a different program.
func TestSigilAtEOFIsAnError(t *testing.T) {
	for _, src := range []string{"$", "@", "%", "print $", "$x = $"} {
		toks := significant(Tokenize([]byte(src)))
		last := toks[len(toks)-1]
		if last.Kind != Error || last.Start != len(src)-1 {
			t.Errorf("%q: the closing sigil is %v %q, want an Error", src, last.Kind, src[last.Start:last.End])
		}
	}
}

// TestSpaceIsNoSigilName: past whitespace perl takes one punctuation byte as
// the name, so `print $ ;` and `$`, a newline and `;` are `$;` on 5.44.0.
// With nothing nameable there, or a heredoc body after the newline, the
// sigil stands alone: FuzzParser's `<<0%$` + newlines took the newline as
// the name, and the body began a byte late.
func TestSpaceIsNoSigilName(t *testing.T) {
	for src, want := range map[string]string{"print $ ;": "$ ;", "print $\n;": "$\n;", "my $x = @ ;": "@ ;"} {
		toks := significant(Tokenize([]byte(src)))
		last := toks[len(toks)-1]
		if last.Kind != Variable || src[last.Start:last.End] != want {
			t.Errorf("%q: last token %v %q, want Variable %q", src, last.Kind, src[last.Start:last.End], want)
		}
	}
	src := "<<0%$\n\n\n0"
	for _, tok := range Tokenize([]byte(src)) {
		if tok.Kind == Variable {
			t.Errorf("%q: %q is no variable; the newline begins the heredoc body", src, src[tok.Start:tok.End])
		}
	}
}

// TestHashCommentNeedsSpace: skipspace runs only when a space follows the
// sigil, so `%#x` is `%#` and `x`, and nothing crosses a newline a heredoc
// body starts after. FuzzParser's `<<0%%#x` + newline + `A` took `#x` as a
// comment and `A` as the name, so the body began a line late.
func TestHashCommentNeedsSpace(t *testing.T) {
	src := "<<0%%#x\nA\n0"
	for _, tok := range Tokenize([]byte(src)) {
		if tok.Kind == Variable && src[tok.Start:tok.End] != "%#" {
			t.Errorf("%q: Variable %q, want `%%#`", src, src[tok.Start:tok.End])
		}
	}
	src = "my $ # c\n b = 2;"
	if toks := significant(Tokenize([]byte(src))); src[toks[1].Start:toks[1].End] != "$ # c\n b" {
		t.Errorf("%q: a comment after a space still stands between sigil and name", src)
	}
}
