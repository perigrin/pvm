// ABOUTME: `CORE::my`, `CORE::our` and `CORE::state` declare in expression position
// ABOUTME: exactly as the bare spellings do.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestCoreDeclaratorInExpression: the statement path names a declarator
// through keywordName, so `CORE::state $x = 1;` already parsed, but the term
// path looked the spelling up as written and `ok(ref(CORE::state $y = ...))`
// refused. Measured on 5.42.0:
//
//	$ perl -MO=Deparse -e 'use feature "state"; ok(ref(CORE::state $y = "a"));
//	      f(CORE::my $x); my $n = (CORE::our $z = 1) + 1;'
//	ok(ref(state $y = 'a'));
//	f(my $x);
//	my $n = (our $z = 1) + 1;
//
// perl.git t/opbasic/concat.t:203.
func TestCoreDeclaratorInExpression(t *testing.T) {
	for _, tc := range []struct{ src, want string }{
		{`ok(ref(CORE::state $y = "a"));`, `ok(ref(CORE::state $y = "a"));`},
		{`f(CORE::my $x);`, `f(CORE::my $x);`},
		{`my $n = (CORE::our $z = 1) + 1;`, `my $n = (CORE::our $z = 1) + 1;`},
	} {
		root := parse.Parse([]byte(tc.src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q: perl accepts this; got %s", tc.src, shape(root))
			continue
		}
		if got := parse.Canon(root, []byte(tc.src)); got != tc.want {
			t.Errorf("%q: canon %q, want %q", tc.src, got, tc.want)
		}
	}
}
