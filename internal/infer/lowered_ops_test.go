// ABOUTME: Tests for AnalyzeLowered's operand and argument rules where perl 5.42 -w is silent and PSC was not.
// ABOUTME: Each pairs the silent cases with a control perl warns about or rejects, which must still be reported.

package infer_test

import (
	"testing"

	"tamarou.com/pvm/internal/types"
)

// A Scalar is some scalar, so it may be any member of a union slot that
// has scalar members: perl 5.42 -w is silent on `chdir $dir` and
// `prototype($dir)` for `my $dir = shift // "."`. Any and a Hole's Any
// satisfy it too. A slot with no scalar member still rejects a Scalar:
// perl 5.42 dies "Experimental push on scalar is now forbidden" on
// `push $s, 1`.
func TestUntypedScalarSatisfiesUnionParameter(t *testing.T) {
	for _, src := range []string{
		"my $dir = shift // \".\";\nchdir $dir or die;\nprint defined prototype($dir) ? \"p\\n\" : \"none\\n\";\n",
		"open my $fh, \"+>\", \"t.out\" or die;\ntruncate $fh, 0;\n",
	} {
		for _, d := range loweredDiags(src) {
			t.Errorf("%q: diagnostic %q, want none", src, d.Message)
		}
	}

	handleOrStr := types.Int | types.Num | types.Str | types.GlobRef | types.Glob | types.NaN | types.Inf | types.IO
	for _, actual := range []types.Type{types.Scalar, types.Any} {
		if !types.TypeSatisfies(actual, handleOrStr) {
			t.Errorf("TypeSatisfies(%v, %v) = false, want true", actual, handleOrStr)
		}
	}
	for _, required := range []types.Type{types.Array, types.Code | types.Glob | types.IO} {
		if types.TypeSatisfies(types.Scalar, required) {
			t.Errorf("TypeSatisfies(Scalar, %v) = true, want false", required)
		}
	}

	src := "my $s = shift // \".\";\npush $s, 1;\n"
	if diags := loweredDiags(src); len(diags) != 1 {
		t.Errorf("%q: diagnostics %v, want one", src, diags)
	}
}
