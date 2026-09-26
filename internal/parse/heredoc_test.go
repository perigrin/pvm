// ABOUTME: Heredocs in statement position: the body token arrives after the `;` that ended its statement.
// ABOUTME: The source-order inversion is the whole difficulty — the opener is a term, the body follows the terminator.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestHeredocInStatementPosition: the statement and the body that follows it
// are ONE statement, with no Unknown.
//
// Measured at 2963c47d, before this landed:
//
//	my $here = <<"EOT";\nheredoc line\nEOT\n
//	  -> statement [0,19), trivia [19,20), unknown not_a_term [20,37)
//
// The statement ALREADY parses -- the opener is a Term, as `heredocs.md`
// records. What fails is the body: it is a token arriving where a new
// statement should start, and `statement()` reaches `parseTerm`, finds a kind
// it has no term rule for, and declines.
//
// With a `print` after the body the refusal widens to `trailing_tokens`,
// because the expression parser then reads the body as a term-less statement
// and keeps going. The two codes are one gap, which is what the corpus cases
// say in prose.
func TestHeredocInStatementPosition(t *testing.T) {
	src := []byte("my $here = <<\"EOT\";\nheredoc line\nEOT\n")

	root := parse.Parse(src)
	if n := countUnknown(root); n != 0 {
		t.Errorf("%d Unknown node(s), want 0: %v", n, kinds(root))
	}
	if got := root.SourceText(src); got != string(src) {
		t.Errorf("round-trip: got %q, want %q", got, src)
	}
}

// TestHeredocQuotingForms: all four spellings of the opener, each with a
// statement after the body so the trailing-token path is covered too.
//
// The quoting is the LEXER's business -- whether the body interpolates is
// decided there and the opener's text carries the answer. The parser sees one
// HeredocOpen token and one HeredocBody token in every form, so a parser that
// handles one handles all four. Asserting it anyway is what makes that claim
// falsifiable rather than assumed: the day a form lexes differently, this
// reports it.
func TestHeredocQuotingForms(t *testing.T) {
	for _, src := range []string{
		"my $h = <<EOT;\nbody\nEOT\nprint $h;\n",
		"my $h = <<\"EOT\";\nbody\nEOT\nprint $h;\n",
		"my $h = <<'EOT';\nbody\nEOT\nprint $h;\n",
		"my $h = <<~EOT;\n    body\n    EOT\nprint $h;\n",

		// The body as a call ARGUMENT rather than an assignment, which is
		// how `t/base/lex.t` writes it.
		"print <<'EOF';\nbody\nEOF\n",

		// Two openers on one line: the bodies arrive in opener order, and
		// both belong to the same statement.
		"print <<A, <<B;\nfirst\nA\nsecond\nB\n",
	} {
		root := parse.Parse([]byte(src))
		if n := countUnknown(root); n != 0 {
			t.Errorf("%q: %d Unknown node(s), want 0: %v", src, n, kinds(root))
		}
		if got := root.SourceText([]byte(src)); got != src {
			t.Errorf("%q: round-trip got %q", src, got)
		}
	}
}
