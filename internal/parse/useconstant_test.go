// ABOUTME: `use constant` declares each name as a sub with the empty prototype, so
// ABOUTME: `Y / X` divides; read with or without a loader, as `use builtin` is.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestUseConstantDeclares: perl's constant.pm defines each constant in the
// caller with prototype "" -- measured on 5.42.0, `prototype "main::X"` is ""
// for `use constant X => 2`, for `use constant { A => 1 }` and for the list
// constant `use constant L => (1, 2)` -- so a following `/` divides:
//
//	$ perl -MO=Deparse -e 'use constant X => 2; use constant Y => 4; print Y / X;'
//	print 2.0;
//
// PerlOnJava unit/constant.t:60.
func TestUseConstantDeclares(t *testing.T) {
	src := "use constant X => 2;\nuse constant 'Y' => 4;\nuse constant { A => 1, B => 3 };\n" +
		"use constant L => (1, 2);\nis(Y / X, 2, 'div');\nis(A + B, 4, 'sum');\nmy @l = (L);\n"
	root := parse.Parse([]byte(src))
	if containsKind(root, parse.Unknown) {
		t.Errorf("perl accepts this; got %s", shape(root))
	}
	// Imports reports only for a resolving parse; a nil loader resolves
	// nothing but keeps the record.
	imps := parse.Imports(parse.ParseWithLoader([]byte(src), nil))
	for _, name := range []string{"X", "Y", "A", "B", "L"} {
		if imp, ok := imps[name]; !ok || !imp.PrototypeKnown || imp.Prototype != "()" {
			t.Errorf("%s declared as %+v (found %v), want prototype ()", name, imp, ok)
		}
	}
}
