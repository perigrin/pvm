// ABOUTME: `|`, `^`, `>`, `>=` and `>>` cannot begin an argument, so a sub name before
// ABOUTME: one takes none: `opt_five | opt_three` is `opt_five() | opt_three()`.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestInfixEndsArguments: measured on 5.42.0 with -MO=Deparse, for `sub f { 1 }`
// and for `sub g (;$) { 1 }` alike,
//
//	my $v = f | 2;    f() | 2       my $v = f >= 2;   f() >= 2
//	my $v = f ^ 2;    f() ^ 2       my $v = f >> 2;   f() >> 2
//	my $v = f > 2;    f() > 2
//
// while `&` and `%` begin a term: `f & 2` is `f(&2)`. PerlOnJava
// unit/subroutine_prototype_args.t:263.
func TestInfixEndsArguments(t *testing.T) {
	for _, op := range []string{"|", "^", ">", ">=", ">>"} {
		for _, decl := range []string{"sub f { 1 }", "sub f (;$) { 1 }"} {
			src := decl + " my $v = f " + op + " 2;"
			root := parse.Parse([]byte(src))
			if containsKind(root, parse.Unknown) {
				t.Errorf("%q: perl accepts this; got %s", src, shape(root))
				continue
			}
			want := decl + " my $v = f() " + op + " 2;"
			if got, w := shape(root), shape(parse.Parse([]byte(want))); got != w {
				t.Errorf("%q: shape %s, want %s as for %q", src, got, w, want)
			}
		}
	}
}
