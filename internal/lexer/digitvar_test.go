// ABOUTME: A variable named by digits takes every digit: `@119797` and `$10` are one
// ABOUTME: token each, not `@1` and a number.

package lexer

import "testing"

// TestDigitVariableName: scan_ident's parse_ident runs with
// STOP_AT_FIRST_NON_DIGIT, so a digit name is the whole run of digits.
// Measured on 5.42.0:
//
//	$ perl -MO=Deparse -e 'my @x = @119797; my $y = $119797[0]; my %h = %12;'
//	my(@x) = @119797;
//	my $y = $119797[0];
//	my(%h) = %12;
//
// perl.git t/op/sub_lval.t:1035 and after.
func TestDigitVariableName(t *testing.T) {
	for _, tc := range []struct{ src, want string }{
		{"my @x = @119797;", "@119797"},
		{"my $y = $119797[0];", "$119797"},
		{"my $z = $10;", "$10"},
		{"my $w = $1;", "$1"},
	} {
		toks := significant(Tokenize([]byte(tc.src)))
		if len(toks) < 4 || toks[3].Kind != Variable || tc.src[toks[3].Start:toks[3].End] != tc.want {
			t.Errorf("%q: want one Variable %q", tc.src, tc.want)
		}
	}
}
