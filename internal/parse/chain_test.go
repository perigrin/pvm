// ABOUTME: A comparison chain keeps each link's own operator, in the tree and in canon:
// ABOUTME: `1 < 2 <= 3` is not `1 < 2 < 3`, and a chain never mixes the relational and equality classes.

package parse_test

import (
	"slices"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestMixedComparisonChainKeepsItsOperators: a chain may mix operators of
// one class, and each link compares with its own. Measured on perl 5.42.0,
// `perl -MO=Deparse -e SRC` writes every case back exactly as `want`:
//
//	my $r = 1 < 2 <= 3;
//	my $s = $a < $b > $c;
//
// When the classes mix, perl builds two chains, the relational one an
// operand of the equality one (perly.y's termeqop over termrelop); with -p:
//
//	$ perl -MO=Deparse,-p -e 'my $y = $a == $b < $c <= $d != $e;'
//	(my($y) = ($a == ($b < $c <= $d) != $e));
func TestMixedComparisonChainKeepsItsOperators(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want string
		ops  []string // the outermost chain's operators, one per link
	}{
		{"my $r = 1 < 2 <= 3;", "my $r = 1 < 2 <= 3;", []string{"<", "<="}},
		{"my $s = $a<$b>$c;", "my $s = $a < $b > $c;", []string{"<", ">"}},
		{"my $t = $a >= $b < $c > $d;", "my $t = $a >= $b < $c > $d;", []string{">=", "<", ">"}},
		{"my $u = $a lt $b ge $c le $d gt $e;", "my $u = $a lt $b ge $c le $d gt $e;", []string{"lt", "ge", "le", "gt"}},
		{"my $q = $a == $b != $c;", "my $q = $a == $b != $c;", []string{"==", "!="}},
		{"my $p = $a eq $b ne $c == $d;", "my $p = $a eq $b ne $c == $d;", []string{"eq", "ne", "=="}},
		{"my $v = $a < $b == $c;", "my $v = $a < $b == $c;", []string{"=="}},
		{"my $y = $a == $b < $c <= $d != $e;", "my $y = $a == $b < $c <= $d != $e;", []string{"==", "!="}},
	} {
		src := []byte(tc.src)
		root := parse.Parse(src)
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q: perl accepts this; got %s", tc.src, shape(root))
			continue
		}
		chain := firstOfKind(root, parse.CmpChain)
		if chain == nil {
			t.Errorf("%q: no CmpChain; got %s", tc.src, shape(root))
			continue
		}
		if !slices.Equal(chain.Ops, tc.ops) {
			t.Errorf("%q: chain operators = %q, want %q", tc.src, chain.Ops, tc.ops)
		}
		if got := parse.Canon(root, src); got != tc.want {
			t.Errorf("%q: canon = %q, want %q", tc.src, got, tc.want)
		}
	}

	// The relational chain inside an equality chain keeps its own links.
	src := []byte("my $y = $a == $b < $c <= $d != $e;")
	outer := firstOfKind(parse.Parse(src), parse.CmpChain)
	if outer == nil || len(outer.Children) != 3 || outer.Children[1].Kind != parse.CmpChain {
		t.Fatalf("want `$b < $c <= $d` as the middle operand of the equality chain")
	}
	if inner := outer.Children[1].Ops; !slices.Equal(inner, []string{"<", "<="}) {
		t.Errorf("inner chain operators = %q, want [\"<\" \"<=\"]", inner)
	}
}
