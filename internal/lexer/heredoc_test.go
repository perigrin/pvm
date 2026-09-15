// ABOUTME: Heredocs: the pending queue, the terminator forms, and << as a left shift.
// ABOUTME: Every expectation ran against perl 5.42 in a FILE, because -e mangles the newlines that matter.

package lexer

import "testing"

// heredocTokens returns the heredoc-related tokens of src.
func heredocTokens(src string) []Token {
	var out []Token
	for _, tok := range Tokenize([]byte(src)) {
		if tok.Kind == HeredocOpen || tok.Kind == HeredocBody {
			out = append(out, tok)
		}
	}
	return out
}

// TestHeredocVersusLeftShift: the same two bytes, decided by the expect
// state. toke.c:7173-7177 gates heredoc recognition on
// `PL_expect != XOPERATOR`:
//
//	if (PL_expect != XOPERATOR) {
//	    if (s[1] == '<' && s[2] != '>')
//	        s = scan_heredoc(s);
//
// Measured: `my $n = 1 << 3` is 8.
//
// This is why heredocs are a separate issue from the quote-like operators --
// those are introduced by a keyword and need no state, while this one is
// nothing BUT state.
func TestHeredocVersusLeftShift(t *testing.T) {
	// Operator position: a left shift, no heredoc.
	for _, src := range []string{
		"my $n = 1 << 3;\n",
		"$x = $y << 2;\n",
	} {
		if got := heredocTokens(src); len(got) != 0 {
			t.Errorf("%q: got heredoc tokens %v; `<<` after a term is a left shift", src, got)
		}
	}

	// Term position: a heredoc.
	const src = "print <<EOF;\nbody\nEOF\n"
	got := heredocTokens(src)
	if len(got) != 2 {
		t.Fatalf("%q: got %d heredoc tokens, want an open and a body: %v", src, len(got), got)
	}
	if got[0].Kind != HeredocOpen {
		t.Errorf("first heredoc token is %v, want HeredocOpen", got[0].Kind)
	}
	if got[1].Kind != HeredocBody {
		t.Errorf("second heredoc token is %v, want HeredocBody", got[1].Kind)
	}
}

// TestHeredocSixForms covers §2.9.3's list. An earlier draft of this issue
// named four and claimed t/base/lex.t used four of them; the file uses three
// and contains no `<<~` at all, so that form has its own fixture here.
func TestHeredocSixForms(t *testing.T) {
	for _, tc := range []struct{ name, src string }{
		{"bare", "print <<EOF;\nbody\nEOF\n"},
		{"double quoted", "print <<\"EOF\";\nbody\nEOF\n"},
		{"single quoted", "print <<'EOF';\nbody\nEOF\n"},
		{"backslash quoted", "print <<\\EOF;\nbody\nEOF\n"},
		{"command", "print <<`EOF`;\nbody\nEOF\n"},
		{"indented", "print <<~EOT;\n  body\n  EOT\n"},
		// The tilde composes with the quoting forms.
		{"indented single quoted", "print <<~'EOT';\n  body\n  EOT\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := heredocTokens(tc.src)
			if len(got) != 2 {
				t.Fatalf("%q: got %d heredoc tokens, want 2: %v", tc.src, len(got), got)
			}
			// The body must reach the terminator line, not stop at the first
			// newline.
			if got[1].End < len(tc.src)-1 {
				t.Errorf("%q: body ends at %d, short of the terminator at %d",
					tc.src, got[1].End, len(tc.src))
			}
		})
	}
}

// TestHeredocQueueOrder: the terminator is read at the `<<` but the body
// starts after the LINE ends, so two heredocs opened on one line take their
// bodies in the order they were opened.
//
// t/base/lex.t:56 does exactly this. Spec §2.9.5, §2.15.1.
func TestHeredocQueueOrder(t *testing.T) {
	const src = "print <<A, <<B;\nfirst\nA\nsecond\nB\n"
	got := heredocTokens(src)
	if len(got) != 4 {
		t.Fatalf("got %d heredoc tokens, want 2 opens and 2 bodies: %v", len(got), got)
	}

	// Both opens come first, on the same line, before either body.
	if got[0].Kind != HeredocOpen || got[1].Kind != HeredocOpen {
		t.Fatalf("the two opens must precede both bodies, got %v %v", got[0].Kind, got[1].Kind)
	}
	if got[2].Kind != HeredocBody || got[3].Kind != HeredocBody {
		t.Fatalf("the two bodies must follow both opens, got %v %v", got[2].Kind, got[3].Kind)
	}

	// A's body comes before B's: the queue is FIFO, not a stack.
	firstBody := src[got[2].Start:got[2].End]
	if want := "first\nA\n"; firstBody != want {
		t.Errorf("first body is %q, want %q — the queue is FIFO", firstBody, want)
	}
}

// TestHeredocMidExpression: t/base/lex.t:45 opens one inside an argument
// list, `eval <<\EOE, print $@;`. The body still begins after the line ends,
// which is why the queue cannot be local to one scan.
func TestHeredocMidExpression(t *testing.T) {
	const src = "eval <<\\EOE, print $@;\nbody\nEOE\n"
	got := heredocTokens(src)
	if len(got) != 2 {
		t.Fatalf("got %d heredoc tokens, want 2: %v", len(got), got)
	}
	// The open is mid-line; the body starts on the next line.
	if got[0].End >= got[1].Start {
		t.Errorf("open ends at %d and body starts at %d: the body must follow the line",
			got[0].End, got[1].Start)
	}
	// Everything between them -- `, print $@;` -- is ordinary tokens, so the
	// open must not have swallowed the rest of the line.
	if got[0].End > 12 {
		t.Errorf("the open ends at %d, past the terminator: it swallowed the "+
			"rest of the expression", got[0].End)
	}
}

// TestHeredocWhitespaceAsymmetry, and the issue text I wrote for this was
// WRONG. It claimed `<<"EOF"` is a heredoc while `<< "EOF"` is a left shift.
// Measured in real files:
//
//	print <<"EOF";  ->  body
//	print << "EOF"; ->  body
//
// Both are heredocs. The asymmetry is on the BAREWORD form only:
//
//	print << EOF;   ->  Use of bare << to mean <<"" is forbidden
//
// So whitespace is permitted before a QUOTED terminator and forbidden before
// a bare one. Perl is explicit about why: `<< ` with nothing quoted would
// otherwise mean an empty terminator, which it refuses rather than guesses.
func TestHeredocWhitespaceAsymmetry(t *testing.T) {
	// A quoted terminator may be spaced.
	for _, src := range []string{
		"print <<\"EOF\";\nbody\nEOF\n",
		"print << \"EOF\";\nbody\nEOF\n",
		"print << 'EOF';\nbody\nEOF\n",
	} {
		if got := heredocTokens(src); len(got) != 2 {
			t.Errorf("%q: got %d heredoc tokens, want 2 — a quoted terminator "+
				"may be preceded by whitespace: %v", src, len(got), got)
		}
	}

	// A bare terminator may not: perl forbids it, so this is not a heredoc.
	const bare = "print << EOF;\nbody\nEOF\n"
	if got := heredocTokens(bare); len(got) != 0 {
		t.Errorf("%q: got heredoc tokens %v; `<< ` with a bare word is forbidden "+
			"by perl, not a heredoc", bare, got)
	}
}
