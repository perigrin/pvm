// ABOUTME: The braces after a deref sigil hold a block, and statements in them
// ABOUTME: parse as one: `${no strict; \$_}`, `${; do { ... } }`.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestDerefHoldsStatements: perly.y's indirob is `BAREWORD | scalar |
// block | PRIVATEREF`, so `${ ... }` holds a block. Read as one expression,
// a `;` inside refused. Measured on 5.42.0:
//
//	$ perl -MO=Deparse -e '$_ = "r"; my $a = ${no strict; \$_};
//	      my $b = ${; do { j: \(my $foo = "foo") } }; my @c = @{; [1,2] };'
//	my $a = ${\$_;};
//	my $b = ${do {
//
// perl.git t/op/lex.t:99 and t/op/goto.t's `${; do { LABEL: ... } }` tests.
// An ordinary `${$x}` stays an expression.
func TestDerefHoldsStatements(t *testing.T) {
	for _, src := range []string{
		`my $a = ${no strict; \$_};`,
		`my $b = ${; do { j: \(my $foo = "foo") } };`,
		`my @c = @{; [1,2] };`,
		`my $d = ${$x};`,
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q: perl accepts this; got %s", src, shape(root))
			continue
		}
		if got := root.SourceText([]byte(src)); got != src {
			t.Errorf("%q: round-trip gave %q", src, got)
		}
		canon := parse.Canon(root, []byte(src))
		again := parse.Canon(parse.Parse([]byte(canon)), []byte(canon))
		if canon != again {
			t.Errorf("%q: canon %q is not a fixpoint, reparses to %q", src, canon, again)
		}
	}
	if root := parse.Parse([]byte(`my $d = ${$x};`)); firstOfKind(root, parse.Block) != nil {
		t.Errorf("`${$x}` is one expression, not a block; got %s", shape(root))
	}
}
