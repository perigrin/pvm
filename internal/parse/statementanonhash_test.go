// ABOUTME: An anonymous hash that begins an expression statement is written `+{`, as perl's Deparse writes it.
// ABOUTME: Written bare, a `{` at statement position re-reads as a block.

package parse_test

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestCanonStatementAnonHashKeepsItsPlus is the parser fuzzer's `({0})`. The
// parens told perl the brace is a hash; written `{0};`, perl reads a block.
// Measured on 5.42.0, Deparse writes `({0})` as `+{0};`, and `+{0}->{a};`,
// `+{0} x 2;` and `+{0} . 1;` each compile, while `{0} . 1;` is "syntax
// error ... near "} ."". So canon writes a `+` before an expression
// statement whose emission begins with a `{` that would re-read as a block;
// `{a => 1};` and `{};` are hashes there already. A parenthesised hash keeps
// its parens, as every Paren node does, and needs no `+`.
func TestCanonStatementAnonHashKeepsItsPlus(t *testing.T) {
	for _, tc := range []struct{ src, want string }{
		{"({0})", "({0});"},
		{"({0}) . 1;", "({0}) . 1;"},
		{"+{0};", "+{0};"},
		{"+{0} . 1;", "+{0} . 1;"},
		{"+{0}->{a};", "+{0}->{a};"},
		{"+{0} x 2;", "+{0} x 2;"},
		{"{0};", "{0;} ;"},
		// A brace perl already reads as a hash at statement start stays bare.
		{"({a=>1});", "({a => 1});"},
		{"{a=>1};", "{a => 1};"},
		{"{};", "{};"},
	} {
		got := strings.TrimSpace(parse.Canon(parse.Parse([]byte(tc.src)), []byte(tc.src)))
		if got != tc.want {
			t.Errorf("Canon(parse(%q)):\n  got  %q\n  want %q", tc.src, got, tc.want)
		}
		again := strings.TrimSpace(parse.Canon(parse.Parse([]byte(got)), []byte(got)))
		if again != got {
			t.Errorf("canon is not a fixpoint on %q:\n  once  %q\n  twice %q", tc.src, got, again)
		}
	}
}
