// ABOUTME: Tests CORE.pmt's delete and four-argument select rows against what perl 5.42 returns.
// ABOUTME: A slice's delete and the select system call yield lists; an element's delete one scalar.
package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
	"tamarou.com/pvm/internal/types"
)

// TestCoreDeleteSliceAndSelectListRows: delete of a slice and four-argument
// select are typed as the lists perl returns. Measured on 5.42:
//
//   - `delete $h{a}` is (1) and `delete $a[0]` (10), one value in either
//     context. `delete @h{qw(a b)}` is (1, 2) and `delete @a[0,1]` (10, 20);
//     `delete %h{qw(a b)}` is (a, 1, b, 2). In scalar context each slice
//     is its last value, one of the list's, so no `:context` fork: the
//     slice is a parse-shape fork, a `@` operand (RFC 0001, "Operators
//     that fork"), as `x`'s parenthesised left operand is. `delete()` is
//     "Not enough arguments", so neither candidate takes none.
//   - `select(undef, undef, undef, 0.01)` is (0, 0), nfound and timeleft;
//     on a ready pipe with a timeout of 5 it is (1, 4.999974) where the
//     scalar call is 1, nfound, again one of the list's values.
func TestCoreDeleteSliceAndSelectListRows(t *testing.T) {
	core := parse.CoreSignatures()
	del := core["delete"]
	for _, c := range []struct {
		arg    types.Type
		shape  string
		ctx    types.Context
		args   int
		want   types.Type
		reason string
	}{
		{types.List, "@", types.ListCtx, 1, types.List, "delete @h{qw(a b)} in list context"},
		{types.List, "@", types.ScalarCtx, 1, types.List, "delete @h{qw(a b)} in scalar context, its last value"},
		{types.Scalar, "$", types.ListCtx, 1, types.Scalar, "delete $h{a} in list context"},
		{types.Scalar, "$", types.ScalarCtx, 1, types.Scalar, "delete $h{a} in scalar context"},
	} {
		sel := types.SelectShaped(del, []types.Type{c.arg}, c.shape, c.ctx)
		if sel.Outcome != types.Selected || sel.Returns != c.want {
			t.Errorf("%s: selects %+v; want %v", c.reason, sel, c.want)
		}
	}
	if sel := types.Select(del, nil, types.ListCtx); sel.Outcome != types.Failed {
		t.Errorf("delete(): selects %+v; perl says Not enough arguments", sel)
	}
	if !del[0].Unary || !del[len(del)-1].Unary {
		t.Errorf("delete: candidates %v; want each :unary", del)
	}

	sel4 := []types.Type{types.Unknown, types.Unknown, types.Unknown, types.Num}
	for _, ctx := range []types.Context{types.ListCtx, types.ScalarCtx} {
		if sel := types.Select(core["select"], sel4, ctx); sel.Outcome != types.Selected || sel.Returns != types.List {
			t.Errorf("select(undef, undef, undef, 0.01) in %v selects %+v; want List, (nfound, timeleft)", ctx, sel)
		}
	}
}
