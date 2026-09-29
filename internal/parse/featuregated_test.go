// ABOUTME: `fc` and `evalbytes` are named unaries only where their feature is on -- by
// ABOUTME: `use feature`, a 5.16+ bundle, or a `CORE::` prefix -- as perl reads them.
package parse_test

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestFeatureGatedFcArity holds issue 01a0dfbd-f8d9's decision, option (b):
// the parser tracks the features `use feature`, `no feature` and a version
// bundle turn on, and reads `fc` and `evalbytes` as named unaries only where
// they are on. Measured on 5.42.0 with -MO=Deparse,-p:
//
//	use feature "fc"; my $z = fc $a, $b;       ((my($z) = fc($a)), $b)
//	use 5.016; my $z = evalbytes $a, $b;      ((my($z) = evalbytes($a)), $b)
//	use v5.14; my $z = fc $a;                 (my($z) = $a->fc)
//
// A version from 5.15 up enables both; `CORE::fc` names the builtin whatever
// is enabled. Unfeatured, perl reads a METHOD call, which this parser does not
// model; that reading is left as it was, and is not asserted here.
func TestFeatureGatedFcArity(t *testing.T) {
	cases := []struct {
		src   string
		canon string
	}{
		{`use feature 'fc'; my $z = fc $a, $b;`, `use feature 'fc'; my $z = fc($a) , $b;`},
		{`use feature qw(unicode_strings evalbytes); evalbytes $p;`, `use feature qw(unicode_strings evalbytes); evalbytes($p);`},
		{`use 5.016; my $z = evalbytes $a, $b;`, `use 5.016; my $z = evalbytes($a) , $b;`},
		{`use v5.36; my $z = fc $a;`, `use v5.36; my $z = fc($a);`},
		{`my $m = ord CORE::fc $c;`, `my $m = ord(CORE::fc($c));`},
	}
	for _, c := range cases {
		src := []byte(c.src)
		n := parse.Parse(src)
		if got := countUnknown(n); got != 0 {
			t.Errorf("Parse(%q) has %d Unknown, want 0", c.src, got)
		}
		got := strings.TrimSpace(parse.Canon(n, src))
		if got != c.canon {
			t.Errorf("Canon(%q):\n  got  %s\n  want %s", c.src, got, c.canon)
		}
	}

	// `no feature` turns it back off, and a bundle below 5.15 never turned it
	// on: the builtin reading must not apply.
	for _, src := range []string{
		`use v5.16; no feature 'fc'; my $z = fc $a, $b;`,
		`use v5.14; my $z = fc $a, $b;`,
	} {
		got := strings.TrimSpace(parse.Canon(parse.Parse([]byte(src)), []byte(src)))
		if strings.Contains(got, "fc($a) , $b") {
			t.Errorf("Canon(%q) = %s: fc read as the builtin where it is not enabled", src, got)
		}
	}
}
