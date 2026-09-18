// ABOUTME: The quote-op predicates behind the match marker, tested on the leaf text directly.
// ABOUTME: perl takes any non-word character as a delimiter, so the rule is a shape, not a list.

package parse

import "testing"

// TestQuoteOpIs is the delimiter rule stated on its own, because the path
// through the parser does not reach every case.
//
// `sort` never arrives at isBarePattern -- it lexes as a Call -- so an
// integration test cannot show that the word-character guard is load
// bearing. It is: without it `sort` is `s` delimited by `o`, `mkdir` is `m`
// delimited by `k`, and `qw(a b)` is a `q` whose delimiter is `w`. Each
// would decide a marker perl never reported.
//
// Every expectation was measured on perl 5.42.0:
//
//	$ perl -MO=Concise,-exec -e 'my $a = m(x); my $b = m[x]; my $c = m!x!;'
//	-- three match ops
//	$ perl -MO=Concise,-exec -e 'my $x; $x =~ s(a)(b); $x =~ tr(a)(b);'
//	-- subst and trans, no match
func TestQuoteOpIs(t *testing.T) {
	for _, c := range []struct {
		text       string
		bare, subs bool
	}{
		// Any non-word delimiter, and whitespace skipped before it.
		{"m(x)", true, false},
		{"m[x]", true, false},
		{"m!x!", true, false},
		{"m'x'", true, false},
		{"m,x,", true, false},
		{"m#x#", true, false},
		{"m /x/", true, false},
		{"m{x}", true, false},
		{"/x/", true, false},

		// perl reports these with subst, trans and qr -- never match.
		{"s(a)(b)", false, true},
		{"s/a/b/", false, true},
		{"tr(a)(b)", false, true},
		{"y(a)(b)", false, true},
		{"qr(a)", false, false},
		{"qr/a/", false, false},

		// A NAME that begins with a quote-op keyword is a name. This is the
		// half the word-character guard exists for.
		{"sort", false, false},
		{"string", false, false},
		{"mkdir", false, false},
		{"yes", false, false},
		{"trim", false, false},
		{"qw(a b)", false, false},

		// A bare keyword with nothing to delimit is neither.
		{"s", false, false},
		{"m", false, false},
		{"y", false, false},
		{"tr", false, false},
	} {
		if got := isBarePattern(c.text); got != c.bare {
			t.Errorf("isBarePattern(%q) = %v, want %v", c.text, got, c.bare)
		}
		if got := isSubstOrTrans(c.text); got != c.subs {
			t.Errorf("isSubstOrTrans(%q) = %v, want %v", c.text, got, c.subs)
		}
	}
}
