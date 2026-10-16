// ABOUTME: `**` and `*=` after a term the lexer cannot see as one -- `__LINE__`, `do {}`, `eval {}`, an
// ABOUTME: undeclared bareword, a glob -- are operators, as perl 5.42's Deparse reads them; declared subs refuse.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestPowerAfterNiladicTerm: a `*` where a term is expected is a glob's
// sigil, so the lexer splits `**` and `*=` there. After each source below a
// term has just ended, and perl reads the operator. Measured on 5.42.0 with
// -MO=Deparse,-p:
//
//	my $y = __LINE__ ** 2;          (my($y) = 1.0);
//	my $y = __FILE__ ** 2;          (my($y) = (__FILE__ ** 2));
//	my $y = __PACKAGE__ ** 2;       (my($y) = (__PACKAGE__ ** 2));
//	sub f { my $y = __SUB__ ** 2 }  (my $y = (__SUB__ ** 2));
//	my $y = do {1} ** 2;            (my($y) = (do { 1 } ** 2));
//	my $y = eval {1} ** 2;          (my($y) = (eval { do { 1 } } ** 2));
//	my $y = pi ** 2;                (my($y) = ('pi' ** 2));
//	my $y = Foo::Bar ** 2;          (my($y) = ('Foo::Bar' ** 2));
//	my $y = pi *= 2;                (my $y = ('pi' *= 2));
//	print *STDOUT **2;              print((*STDOUT ** 2));
//
// `pi *= 2` is also "Can't modify constant item", a check after the parse;
// the operator is what perl read. The controls are perl's syntax errors,
// where a list operator takes `**` as a glob: `print **2;`, and `foo **2`
// and `foo *= 2` after `sub foo;`, each "Number found where operator
// expected". `sub foo(); foo **2` is a constant and raises it to a power.
func TestPowerAfterNiladicTerm(t *testing.T) {
	for src, op := range map[string]string{
		"my $y = __LINE__ ** 2;":                                    "**",
		"my $y = __FILE__ ** 2;":                                    "**",
		"my $y = __PACKAGE__ ** 2;":                                 "**",
		"use feature 'current_sub'; sub f { my $y = __SUB__ ** 2 }": "**",
		"my $y = do {1} ** 2;":                                      "**",
		"my $y = eval {1} ** 2;":                                    "**",
		"my $y = pi ** 2;":                                          "**",
		"my $y = Foo::Bar ** 2;":                                    "**",
		"my $y = pi *= 2;":                                          "*=",
		"print *STDOUT **2;":                                        "**",
		"sub foo(); my $y = foo **2;":                               "**",
	} {
		root := parse.Parse([]byte(src))
		if u := firstUnknown(root); u != nil {
			t.Errorf("%q: refused as %s at %d", src, u.Refusal, u.Start)
			continue
		}
		if !hasBinary(root, op) {
			t.Errorf("%q: no %q operator in the tree", src, op)
		}
	}
	for _, src := range []string{
		"print **2;",
		"sub foo; my $y = foo **2;",
		"sub foo; my $y = foo *= 2;",
	} {
		if firstUnknown(parse.Parse([]byte(src))) == nil {
			t.Errorf("%q: perl refuses it, but it parsed with no Unknown", src)
		}
	}
}

// hasBinary reports whether a Binary node with operator op is in the tree.
func hasBinary(n *parse.Node, op string) bool {
	if n.Kind == parse.Binary && n.Text == op {
		return true
	}
	for _, c := range n.Children {
		if hasBinary(c, op) {
			return true
		}
	}
	return false
}
