// ABOUTME: Parenless builtins keep their arguments — an empty call in canon is a wrong tree at Unknown=0.
// ABOUTME: The audit of the three arity tables against perl's regen/keywords.pl lives here.

package parse_test

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestParenlessBuiltinArity: a parenless builtin takes its arguments, and the
// canon proves it.
//
// The Unknown count cannot see this failure. A name in none of the three
// arity tables falls through parseWordTerm's `n.Resolved` branch to an
// argumentless call, and the argument list becomes a statement of its own:
//
//	my $d = atan2 $var, 1 ;   ->   my $d = atan2();$var , 1;
//
// That scores Unknown=0 and is not a parse of its source. The signature is
// `NAME()` in the canon -- an empty call with the arguments detached -- so
// that is what this test looks for, not the node count.
//
// Each classification below was measured on perl 5.42.0 by where the comma
// lands, the same method that built namedUnary:
//
//	perl -MO=Deparse,-p -e 'our ($a,$b); my $z = atan2 $a, $b;'
//	  (my($z) = atan2($a, $b));        comma INSIDE  -> list operator
//
//	perl -MO=Deparse,-p -e 'our ($a,$b); my $z = readpipe $a, $b;'
//	  ((my($z) = `$a`), $b);           comma OUTSIDE -> named unary
//
//	perl -MO=Deparse,-p -e 'our ($a,$b); my $z = glob $a;'
//	  use File::Glob (); (my($z) = glob($a));
//	perl -MO=Deparse,-p -e 'our ($a,$b); my $z = glob $a, $b;'
//	  Too many arguments for glob      -> named unary, one argument only
func TestParenlessBuiltinArity(t *testing.T) {
	for _, tc := range []struct{ src, want string }{
		// atan2 is a two-argument list operator.
		{"my $d = atan2 $var, 1;", "my $d = atan2($var , 1);"},
		{"$dummy = atan2 $var, 1;", "$dummy = atan2($var , 1);"},

		// readpipe and glob are named unaries: the comma stays outside, so
		// the second element belongs to the enclosing list, not to the call.
		{"my $out = readpipe $cmd;", "my $out = readpipe($cmd);"},
		{"my @f = glob $pattern;", "my @f = glob($pattern);"},

		// REGRESSION GUARD: three list operators already tabled. These pass
		// today and must keep passing -- a table edit that reclassified one
		// of them would show here.
		{"my $i = index $str, $sub;", "my $i = index($str , $sub);"},
		{"my $s = sprintf $fmt, $x;", "my $s = sprintf($fmt , $x);"},
		{"my $c = crypt $pw, $salt;", "my $c = crypt($pw , $salt);"},

		// REGRESSION GUARD: a named unary already tabled keeps the comma
		// outside, so `ord $a, $b` is `ord($a), $b` and NOT `ord($a, $b)`.
		{"my $n = ord $char, 1;", "my $n = ord($char) , 1;"},
	} {
		root := parse.Parse([]byte(tc.src))
		got := parse.Canon(root, []byte(tc.src))
		if got != tc.want {
			t.Errorf("%q\n got  %q\n want %q", tc.src, got, tc.want)
		}
		if n := countKind(root, parse.Unknown); n != 0 {
			t.Errorf("%q: %d Unknown, want 0", tc.src, n)
		}
	}
}

// TestParenlessArityNoDetachedCall: no tabled builtin emits an empty call.
//
// The generalisation of the case above. Every name in the three arity tables
// that takes at least one argument must, given one, put it INSIDE the call.
// `NAME()` immediately followed by `;` is the detached-argument signature.
func TestParenlessArityNoDetachedCall(t *testing.T) {
	for _, name := range []string{
		"atan2", "glob", "readpipe", "index", "sprintf", "crypt", "vec",
		"join", "ord", "uc", "lc", "length", "ref", "hex", "oct", "sqrt",
	} {
		src := "my $z = " + name + " $a;"
		root := parse.Parse([]byte(src))
		got := parse.Canon(root, []byte(src))
		if strings.Contains(got, name+"()") {
			t.Errorf("%q: canon %q holds an empty %s() -- the argument is detached",
				src, got, name)
		}
	}
}
