// ABOUTME: Tests for the lowered path's variable, subscript and match rules: what a bound match, a hash element and a slot's context are typed.
// ABOUTME: Each case cites perl 5.42 -w's verdict, and each fix carries a control perl reports that must stay reported.

package infer_test

import (
	"testing"

	"tamarou.com/pvm/internal/types"
)

// A named unary operator binds looser than `=~`, so `chr $u =~ /re/` is
// chr($u =~ /re/): perl 5.42 deparses it so, and -w is silent, printing
// "ascii" for $u = 65. A `$` slot takes the match in scalar context, where
// it is a Boolean, not a List. A match in list context is still a List, and
// an aggregate in a `$` slot is still the mistake perl warns about:
// `length(1 ? @a : "x")` draws "length() used on @a".
func TestChrOfBoundMatchIsScalar(t *testing.T) {
	for _, src := range []string{
		"my $u = 65;\nprint \"ascii\\n\" if chr $u =~ /[[:^ascii:]]/;\n",
		"my $u = 65;\nmy $c = chr($u =~ /x/);\n",
		"my $m;\nmy $c = chr($m =~ /x/);\n",
		"my $c = chr(;",
	} {
		_, diags := analyzeSource(t, []byte(src))
		for _, d := range diags {
			t.Errorf("%q: diagnostic %q, want none", src, d.Message)
		}
	}

	src := []byte("my $u = 65;\nmy @m = ($u =~ /(\\d)/);\n")
	ty, _ := analyzeSource(t, src)
	if got, ok := findNodeType(ty, src, `($u =~ /(\d)/)`); !ok || got != types.List {
		t.Errorf("%q: the match in list context is %v (found %v), want List", src, got, ok)
	}

	for _, src := range []string{
		"my @a = (1, 2);\nmy $n = length(1 ? @a : \"x\");\n",
		"my $r = {};\nmy $c = chr($r);\n",
	} {
		if _, diags := analyzeSource(t, []byte(src)); len(diags) == 0 {
			t.Errorf("%q: no diagnostic, want the argument reported", src)
		}
	}
}
