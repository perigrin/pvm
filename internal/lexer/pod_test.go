// ABOUTME: Line-oriented lexer modes: POD, __END__/__DATA__, format bodies, and /e replacement sublexing.
// ABOUTME: The plan's M0 prose names all four; an earlier draft of this chain covered none of them.

package lexer

import "testing"

// firstOfKind returns the first token of kind k, or a zero Token and false.
func firstOfKind(src string, k Kind) (Token, bool) {
	for _, tok := range Tokenize([]byte(src)) {
		if tok.Kind == k {
			return tok, true
		}
	}
	return Token{}, false
}

// TestPodSpan: a `=word` line where a STATEMENT could start opens POD, and
// everything through `=cut` inclusive is trivia.
//
// "Where a statement could start" is the expect state, which is why this
// issue follows the expect-state one. A `=` in the middle of an expression is
// an operator, not POD.
func TestPodSpan(t *testing.T) {
	const src = "print 1;\n=pod\nignored text\n=cut\nprint 2;\n"
	tok, ok := firstOfKind(src, Pod)
	if !ok {
		t.Fatalf("%q: no Pod token", src)
	}
	// From `=pod` through the newline after `=cut`.
	if tok.Start != 9 {
		t.Errorf("Pod starts at %d, want 9 (the `=` of `=pod`)", tok.Start)
	}
	if got := src[tok.Start:tok.End]; got != "=pod\nignored text\n=cut\n" {
		t.Errorf("Pod span is %q, want the block through `=cut`", got)
	}

	// An `=` in expression position is NOT pod.
	if _, ok := firstOfKind("my $x = 1;\n", Pod); ok {
		t.Error("`=` in expression position opened POD; it is assignment")
	}
}

// TestPodTerminatorBattery is t/base/lex.t:590-630, which exists to break a
// lexer that matches a `=cut` PREFIX.
//
// Spec §2.5.1's three-conjunct rule: `=cut` ends POD only when it is not
// followed by an identifier character. Measured -- a file whose POD opens at
// `=cute` and closes at a later `=cut` prints both statements around it:
//
//	print "a\n";
//	=cute
//	ignored
//	=cut
//	print "b\n";
//
//	$ perl pod1.pl
//	a
//	b
//
// If `=cute` had ended the POD, `ignored` would be a syntax error. This is
// the failure the golden token stream exists to catch, because a lexer that
// gets it wrong still round-trips perfectly.
func TestPodTerminatorBattery(t *testing.T) {
	for _, tc := range []struct {
		name, src string
		wantEnd   string // the text the Pod token must end with
	}{
		{"cute does not terminate", "=pod\na\n=cute\nb\n=cut\n", "=cut\n"},
		{"cut2 does not terminate", "=pod\na\n=cut2\nb\n=cut\n", "=cut\n"},
		{"cut_ does not terminate", "=pod\na\n=cut_\nb\n=cut\n", "=cut\n"},
		{"cut terminates", "=pod\na\n=cut\n", "=cut\n"},
		// A trailing space is not an identifier character, so this DOES end.
		{"cut with trailing space", "=pod\na\n=cut \n", "=cut \n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tok, ok := firstOfKind(tc.src, Pod)
			if !ok {
				t.Fatalf("%q: no Pod token", tc.src)
			}
			got := tc.src[tok.Start:tok.End]
			if len(got) < len(tc.wantEnd) || got[len(got)-len(tc.wantEnd):] != tc.wantEnd {
				t.Errorf("%q: Pod span %q does not end with %q — a `=cut` prefix "+
					"must not terminate", tc.src, got, tc.wantEnd)
			}
		})
	}
}

// TestPodAfterBareSigilBlock is t/base/lex.t:491-505: `=pod` immediately
// after a line ending in `map{` or `${`.
//
// Measured -- perl accepts it and the map still runs:
//
//	my @r = map{
//	=pod
//	ignored
//	=cut
//	$_ * 2 } (21);
//
//	$ perl pod2.pl
//	[42]
//
// This needs XSTATE inside a bare-sigil block (§3.1.4): after `{` a statement
// may start, so a `=` there is POD.
func TestPodAfterBareSigilBlock(t *testing.T) {
	for _, tc := range []struct{ name, src string }{
		{"after map brace", "my @r = map{\n=pod\nignored\n=cut\n$_ * 2 } (21);\n"},
		{"after bare brace", "{\n=pod\nignored\n=cut\n1; }\n"},
		{"after sub brace", "sub f {\n=pod\nignored\n=cut\n1; }\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tok, ok := firstOfKind(tc.src, Pod)
			if !ok {
				t.Fatalf("%q: no Pod token; a statement may start after `{`, so "+
					"a `=` there is POD", tc.src)
			}
			if got := tc.src[tok.Start:tok.End]; got != "=pod\nignored\n=cut\n" {
				t.Errorf("Pod span is %q, want the whole block", got)
			}
		})
	}
}

// TestEndAndDataSection: `__END__` and `__DATA__` end the program. Everything
// after is data, not Perl, and must not be lexed as code.
func TestEndAndDataSection(t *testing.T) {
	for _, marker := range []string{"__END__", "__DATA__"} {
		src := "print 1;\n" + marker + "\nthis is ) not ( perl $$$\n"
		tok, ok := firstOfKind(src, DataSection)
		if !ok {
			t.Fatalf("%q: no DataSection token", src)
		}
		if tok.End != len(src) {
			t.Errorf("%s: data section ends at %d, want %d — it runs to EOF",
				marker, tok.End, len(src))
		}
		// And nothing after the marker is lexed as code: no Error tokens from
		// the deliberately unbalanced punctuation.
		for _, other := range Tokenize([]byte(src)) {
			if other.Start >= tok.Start && other.Kind == Error {
				t.Errorf("%s: an Error token inside the data section; its contents "+
					"are not Perl", marker)
			}
		}
	}
}

// TestFormatBody: §0.13 rank 3, 8 corpus files, 3 of them in T2.
//
// A `format NAME =` line opens a body ending at a lone `.` on its own line --
// the same shape as a heredoc, which is why §0.13 says it "wants the same
// lexer mode". Picture lines are not Perl:
//
//	format STDOUT =
//	@<<<<< @>>>>>
//	$a,     $b
//	.
//
// The picture line `@<<<<< @>>>>>` would otherwise lex as an array sigil
// followed by left shifts.
func TestFormatBody(t *testing.T) {
	const src = "format STDOUT =\n@<<<<< @>>>>>\n$a,     $b\n.\nprint 1;\n"
	tok, ok := firstOfKind(src, FormatBody)
	if !ok {
		t.Fatalf("%q: no FormatBody token", src)
	}
	if got := src[tok.Start:tok.End]; got != "@<<<<< @>>>>>\n$a,     $b\n.\n" {
		t.Errorf("format body is %q, want the picture lines through the lone `.`", got)
	}

	// The picture lines are not lexed as Perl: no Error and no Readline
	// tokens from `@<<<<<`.
	for _, other := range Tokenize([]byte(src)) {
		if other.Start >= tok.Start && other.End <= tok.End {
			if other.Kind == Error || other.Kind == Readline {
				t.Errorf("token %v inside the format body; picture lines are not Perl",
					other.Kind)
			}
		}
	}
}

// TestSubstitutionEvalSublex is t/base/lex.t:114, and it is the case that
// makes /e non-optional:
//
//	$foo =~ s/^not /substr(<<EOF, 0, 0)/e;
//	  Ignored
//	EOF
//
// A heredoc opened INSIDE the replacement, with its body on the lines that
// follow. Treat the replacement as opaque and the `<<EOF` is missed, the body
// lexes as barewords, and the file still round-trips -- which is why round
// trip alone cannot gate this milestone.
func TestSubstitutionEvalSublex(t *testing.T) {
	const src = "$foo =~ s/^not /substr(<<EOF, 0, 0)/e;\n  Ignored\nEOF\n"

	// The heredoc opener is NOT a separate token, and that is deliberate: it
	// sits inside the Quote span, so emitting it too would cover those bytes
	// twice and break the lossless invariant. What has to escape the span is
	// the QUEUE ENTRY -- the effect on where the body lands -- not a token.
	//
	// An earlier version of this test demanded a HeredocOpen token and was
	// asserting something incompatible with invariant 3.
	body, ok := firstOfKind(src, HeredocBody)
	if !ok {
		t.Fatal("no HeredocBody: the body follows the line the substitution sits on")
	}
	if got := src[body.Start:body.End]; got != "  Ignored\nEOF\n" {
		t.Errorf("heredoc body is %q, want the two lines after the substitution", got)
	}

	// Without /e the replacement is a STRING: `<<EOF` in it is literal text,
	// so no body is taken and the following lines stay ordinary code. This
	// is the control that keeps the test above from passing on a lexer that
	// queues a heredoc for every `<<` it sees anywhere.
	const noE = "$foo =~ s/^not /substr(<<EOF, 0, 0)/;\n  Ignored\nEOF\n"
	if _, ok := firstOfKind(noE, HeredocBody); ok {
		t.Error("a heredoc body was taken for a replacement with no /e; there " +
			"the `<<EOF` is literal text, not code")
	}
}
