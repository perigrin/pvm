// ABOUTME: A leading `::` names main, so `::ok(1)` parses as the call `main::ok(1)` does,
// ABOUTME: as a statement and as a term; held by tree shape and canon, not by a zero alone.

package parse_test

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestLeadingPackageSeparatorCall holds issue 01a0de97-3c5d at the parser.
// The lexer's TestLeadingPackageSeparatorInWord holds the token stream; this
// holds what the parser builds from it. Before the lexer fix, the two
// spellings refused with two different codes:
//
//	::ok(1);         Unknown=1  trailing_tokens
//	my $r = ::f();   Unknown=1  not_a_term "::"
//
// Valid perl 5.42.0, and Deparse reads both as the main:: sub:
//
//	$ perl -MO=Deparse -e 'sub main::ok {1} ::ok(1);'          -> ok 1;
//	$ perl -MO=Deparse -e 'sub main::f {1} my $r = ::f();'     -> my $r = f();
func TestLeadingPackageSeparatorCall(t *testing.T) {
	for _, tc := range []struct{ src, shape, canon string }{
		// As a statement: one call whose name carries the separator.
		{"::ok(1);", "( ( (::ok 1)))", "::ok(1);"},
		// As a term: the call is the initialiser.
		{"my $r = ::f();", "( ( (my $r ::f)))", "my $r = ::f();"},
		// REGRESSION GUARD: the explicit spelling was always fine.
		{"main::ok(1);", "( ( (main::ok 1)))", "main::ok(1);"},
	} {
		n := parse.Parse([]byte(tc.src))
		if got := countUnknown(n); got != 0 {
			t.Errorf("Parse(%q): Unknown=%d, want 0", tc.src, got)
		}
		if got := shape(n); got != tc.shape {
			t.Errorf("Parse(%q): shape %s, want %s", tc.src, got, tc.shape)
		}
		if got := strings.TrimSpace(parse.Canon(n, []byte(tc.src))); got != tc.canon {
			t.Errorf("Canon(parse(%q)) = %q, want %q", tc.src, got, tc.canon)
		}
	}
}
