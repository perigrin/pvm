// ABOUTME: Without the isa feature `isa` is an ordinary word: `isa Local::Child('X')`
// ABOUTME: is the method call 'Local::Child'->isa('X'); with it, `$x isa Foo` is infix.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestIsaWithoutFeature: keywords.c:353 returns KEY_isa only under
// FEATURE_ISA_IS_ENABLED. Measured on 5.42.0 with -MO=Deparse,
//
//	{ package Local::Child; } ok(isa Local::Child('Local::Child'), 'x');
//	    &ok(scalar 'Local::Child'->isa('Local::Child'), 'x');
//	use feature "isa"; my $r = $x isa Foo;     my $r = $x isa 'Foo';
//	use v5.36; my $r = $x isa Foo;             my $r = ${'x'} isa 'Foo';
//
// PerlOnJava unit/indirect_isa_feature_gate.t:10.
func TestIsaWithoutFeature(t *testing.T) {
	src := `{ package Local::Child; } ok(isa Local::Child('Local::Child'), 'x');`
	root := parse.Parse([]byte(src))
	if call := indirectCall(root, "isa"); call == nil || call.Children[0].Text != "Local::Child" {
		t.Errorf("%q: want 'Local::Child'->isa(...); got %s", src, shape(root))
	}
	for _, src := range []string{
		`use feature "isa"; my $r = $x isa Foo;`,
		`use v5.36; my $r = $x isa Foo;`,
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) || indirectCall(root, "isa") != nil {
			t.Errorf("%q: `isa` is the infix operator here; got %s", src, shape(root))
		}
	}
}
