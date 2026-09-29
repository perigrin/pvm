// ABOUTME: A foreach loop variable may be a reference to alias through:
// ABOUTME: `for \my $x (...)`, `for my \$x (...)`, `for \$x (...)`, `for \@a (...)`.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestRefaliasLoopVariable: perly.y's KW_FOR productions take `my_refgen
// my_var` and `REFGEN refgen_topic` before the list. The parser left the
// `foreach` as an empty loop and read the rest as an expression statement,
// the `\` swallowing the list as a call on the variable -- a wrong tree at
// Unknown=0. Measured on 5.42.0:
//
//	$ perl -MO=Deparse -e 'use feature qw(refaliasing declared_refs); no warnings;
//	      for my \$a (\1) { 1 } for \my $b (\2) { 2 } for \$c (\3) { 3 }
//	      for \@d ([4]) { 4 } for \%e ({}) { 5 } for \&f (sub {}) { 6 }'
//	foreach \my $a (\1) {
//	foreach \my $b (\2) {
//	foreach \$c (\3) {
//	foreach \@d ([4]) {
//	foreach \%e ({}) {
//	foreach \&f (sub {
//
// perl.git t/op/decl-refs.t:127 and t/op/lvref.t.
func TestRefaliasLoopVariable(t *testing.T) {
	for _, src := range []string{
		`for my \$a (\1) { 1 }`,
		`for \my $b (\2) { 2 }`,
		`foreach \$slexical ( \1, \2, \3 ) { }`,
		`for \@d ([4]) { 4 }`,
		`for \%e ({}) { 5 }`,
		`for \&f (sub {}) { 6 }`,
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q: perl accepts this; got %s", src, shape(root))
			continue
		}
		loop := firstOfKind(root, parse.Loop)
		if loop == nil || len(loop.Children) < 3 || loop.Children[len(loop.Children)-1].Kind != parse.Block {
			t.Errorf("%q: want one loop with a variable, a list and a block; got %s", src, shape(root))
			continue
		}
		if got := root.SourceText([]byte(src)); got != src {
			t.Errorf("%q: round-trip gave %q", src, got)
		}
		canon := parse.Canon(root, []byte(src))
		if again := parse.Canon(parse.Parse([]byte(canon)), []byte(canon)); again != canon {
			t.Errorf("%q: canon %q is not a fixpoint, reparses to %q", src, canon, again)
		}
	}
}
