// ABOUTME: Tests for the referent guard, ref(\$x) eq 'TYPE', which tests what $x itself holds.
// ABOUTME: Only GLOB narrows, to Glob, whatever $x was typed; a failed guard narrows nothing.

package types_test

import (
	"testing"

	"tamarou.com/pvm/internal/types"
)

// perl 5.42: `my $v = *STDOUT; ref \$v` is GLOB, and `my $v = "x"; ref
// \$v` SCALAR. A glob held in a scalar is typed Scalar, so the guard makes
// a Glob of a Scalar, an Unknown or an Any alike.
func TestReferentGuardNarrowsToGlob(t *testing.T) {
	glob := types.GuardPattern{Kind: types.GuardReferent, RefType: "GLOB"}
	for _, typ := range []types.Type{types.Scalar, types.Unknown, types.Any, types.Glob} {
		if got, _ := types.NarrowByGuard(typ, glob); got != types.Glob {
			t.Errorf("NarrowByGuard(%v, ref \\$x eq GLOB) = %v, want Glob", typ, got)
		}
		if got, ok := types.NegateGuard(typ, glob); ok {
			t.Errorf("NegateGuard(%v, ref \\$x eq GLOB) narrowed to %v, want no narrowing", typ, got)
		}
	}
	scalar := types.GuardPattern{Kind: types.GuardReferent, RefType: "SCALAR"}
	if got, ok := types.NarrowByGuard(types.Scalar, scalar); ok {
		t.Errorf("NarrowByGuard(Scalar, ref \\$x eq SCALAR) narrowed to %v, want no narrowing", got)
	}
}
