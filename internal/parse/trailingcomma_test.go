// ABOUTME: A trailing comma before a closer is a separator with no element, not a missing operand.
// ABOUTME: Measured as the largest single first-failure cause in perl.git t/: 27 files, 417 nodes.
package parse_test

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestTrailingCommaBeforeCloserIsNotATerm holds the construct that bucketing
// every dirty perl.git t/ file's FIRST Unknown span named as the single largest
// cause: a list whose last element is followed by a comma.
//
// perl allows it everywhere a list is written, and the whole `runperl(` /
// `foreach my $x (` / `test_opcount(` multi-line house style of perl's own
// suite depends on it -- an argument added at the end of such a list is a
// one-line diff precisely because the comma is already there.
//
// The parser read the comma as an INFIX OPERATOR demanding a right operand,
// so parseTerm was called at the closer and consumed it as `not_a_term`. That
// is the cheap half of the damage. The expensive half is that the closer is
// then GONE from the enclosing construct, which reads the next token as its
// own -- and the tree that results is not a wrong count, it is a wrong shape:
//
//	foreach my $x ($a,) { 1 }
//
// canon'd as `foreach my $x ($a , ){1})` before this fix, the `{ 1 }` having
// become a HASH SUBSCRIPT on the comma expression rather than the loop body.
// An Unknown count cannot see that; canon can, which is why every case here
// asserts the canonical text and not just the node count.
func TestTrailingCommaBeforeCloserIsNotATerm(t *testing.T) {
	cases := []struct {
		name  string
		src   string
		canon string
	}{
		// The call argument list: `parseCallArgs` hands the whole inside to
		// parseExpr, so the comma's operand hunt runs straight into the `)`.
		// 27 of the 67 bare-closer files in perl.git t/ are this shape.
		{"call one arg", "f($a,);", "f($a,);"},
		{"call two args", "f($a, $b,);", "f($a, $b,);"},

		// A named list operator's parenthesised form takes the same path.
		{"list op", `is(index($a, "bar",    ), 5);`, `is(index($a, "bar",), 5);`},

		// The loop list, where the damage is structural rather than a count:
		// the swallowed `)` left the body `{ 1 }` to be read as a subscript.
		{"foreach list", "foreach my $x ($a,) { 1 }", "foreach my $x ($a,) {1;}"},

		// A hash subscript is a list too, and it was the third caller. With
		// the comma gone the key is a BAREWORD KEY rather than a call, which
		// is what perl reads there -- `$h{a}` auto-quotes -- so the canon
		// loses the `()` the broken parse had invented around it.
		{"hash subscript", "$h{a,};", "$h{a,};"},

		// The fat comma is the same separator and must behave the same way.
		{"fat comma", "f(a => 1,);", "f(a => 1,);"},

		// A comma with nothing at all in front of it is NOT this case: there
		// is no element for the separator to follow, so the refusal stands.
		// Guarding that keeps the fix from turning a real syntax error into a
		// silent empty list.
		{"only a comma", "f(,);", "f(,);"},

		// A parenthesised list of ONE element is that element, its span
		// widened over the parens, so a leaf's source text already holds
		// the comma. Written again after it, `(0,)` came back `(0,) ,`, a
		// comma expression the next pass read differently -- found by the
		// parser fuzzer from `(0,)`. perl 5.42 Deparses `my @a = (0,);` as
		// `my(@a) = 0;`.
		{"paren one leaf", "my @a = (0,);", "my @a = (0,);"},
		{"paren one variable", "my @a = ($x,);", "my @a = ($x,);"},
		// A fat comma's word is the same leaf, written as its word: in
		// parens it is written as its source, `=>` and all, or canon wrote
		// `A`, which the next pass read as a call. perl 5.42 Deparses
		// `my @a = (A=>);` as `my(@a) = 'A';`. Found by the parser fuzzer
		// from `(A=>)`.
		{"paren one fat word", "my @a = (A=>);", "my @a = (A=>);"},
		// Any other single element keeps its parens around the comma, or
		// the comma lands bare after it: `(A,)/0` came back `A() , / 0`,
		// which the next pass read as a comma before an unterminated
		// pattern -- found by the parser fuzzer. perl 5.42 Deparses
		// `(A,)/0` as `'A' / 0` and `my @x = (f(1),)` as `my(@x) = f(1)`.
		{"paren one call", "(A,)/0;", "(A,) / 0;"},
		{"paren one call assigned", "my @x = (f(1),);", "my @x = (f(1),);"},
		{"paren one element", "($a[0],)[0];", "($a[0],)[0];"},
	}

	// Canon writes the trailing comma back where the source had it
	// (perigrin, 2026-09-30); the tree drops it, as perl does.
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			src := []byte(c.src)
			n := parse.Parse(src)
			if c.name != "only a comma" {
				if got := countUnknown(n); got != 0 {
					t.Errorf("Parse(%q) has %d Unknown, want 0", c.src, got)
				}
			}
			got := strings.TrimSpace(parse.Canon(n, src))
			if got != c.canon {
				t.Errorf("Canon(%q):\n  got  %s\n  want %s", c.src, got, c.canon)
			}
		})
	}
}
