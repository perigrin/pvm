// ABOUTME: Tests HasQuoteOperator, the exported quote-op predicate.
// ABOUTME: Exported so the conformance corpus need not keep a second table.
package lexer

import "testing"

// TestHasQuoteOperator covers the question the conformance corpus asks:
// does this Quote token RUN an operator, or is it a plain string?
//
// Both lex to Kind Quote, so the text is what separates them. The corpus
// needs the distinction because `qw(a b)` is a LIST and `"hi"` is a
// string, and a file asserting one must not be satisfied by the other.
func TestHasQuoteOperator(t *testing.T) {
	for _, tc := range []struct {
		text string
		want bool
	}{
		// Every operator in quoteOps, so a name added there without a
		// case here is visible.
		{"q(a b)", true},
		{"qq{hi}", true},
		{"qw(a b)", true},
		{"qr/pat/", true},
		{"qx/ls/", true},
		{"m{pat}", true},
		{"s/a/b/", true},
		{"tr/a/b/", true},
		{"y/a/b/", true},

		// Backticks run a command with no operator name at all, so a
		// prefix test alone would miss them.
		{"`ls`", true},

		// Plain strings.
		{`"hi"`, false},
		{"'hi'", false},

		// The text CONTAINS an operator name but does not begin with
		// one: the delimiters decide.
		{`"qw stuff"`, false},

		// A longer word beginning with an operator's name. Not a Quote
		// in practice, but the predicate must not say otherwise.
		{"sort", false},
		{"qty", false},
		{"my", false},

		// A bare name with no delimiter is not an operator either.
		{"q", false},
	} {
		if got := HasQuoteOperator(tc.text); got != tc.want {
			t.Errorf("HasQuoteOperator(%q) = %v, want %v", tc.text, got, tc.want)
		}
	}
}

// TestHasQuoteOperatorCoversEveryOp keeps the predicate honest as quoteOps
// grows. `qx` was once missing from a hand-copied list elsewhere, which
// made `qx/ls/` read as a plain string; this is the check that would have
// caught it at the source.
func TestHasQuoteOperatorCoversEveryOp(t *testing.T) {
	for _, op := range quoteOps {
		text := op.name + "/x/"
		if !HasQuoteOperator(text) {
			t.Errorf("HasQuoteOperator(%q) = false, but %q is in quoteOps", text, op.name)
		}
	}
}
