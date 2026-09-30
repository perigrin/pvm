// ABOUTME: `BEGIN { *name = \&target }` gives name target's prototype: an alias of
// ABOUTME: CORE::push takes `\@@`, one of a declared sub takes that sub's.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestGlobAliasPrototype: measured on 5.42.0,
//
//	BEGIN { *my_push = \&CORE::push; }   prototype(\&my_push) is \@@
//	sub f ($$) {} BEGIN { *g = \&f }     prototype("g") is $$
//
// PerlOnJava unit/core_subroutine_refs.t:82-84.
func TestGlobAliasPrototype(t *testing.T) {
	src := "BEGIN { *my_push = \\&CORE::push; }\nmy @arr2;\nmy_push @arr2, 4, 5, 6;\n" +
		"sub f ($$) { } BEGIN { *g = \\&f; }\n"
	root := parse.Parse([]byte(src))
	if containsKind(root, parse.Unknown) {
		t.Errorf("perl accepts this; got %s", shape(root))
	}
	imps := parse.Imports(parse.ParseWithLoader([]byte(src), nil))
	for name, want := range map[string]string{"my_push": `(\@@)`, "g": "($$)"} {
		if imp := imps[name]; !imp.PrototypeKnown || imp.Prototype != want {
			t.Errorf("%s: %+v, want prototype %s", name, imp, want)
		}
	}
}
