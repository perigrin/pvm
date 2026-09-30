// ABOUTME: Faithful forgives a call's added parens inside another call's added parens,
// ABOUTME: the same difference one level down; any other extra paren still fails it.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestFaithfulNestedCallParens: canon gives every call its argument parens,
// and Faithful forgives the ones the source left out -- but only at the top
// level, so `print create PrintIndirectFactory sub {...}`, emitted
// `print(create PrintIndirectFactory(sub {...}))`, failed on the inner pair
// though it differs from the source exactly as the outer one does.
// PerlOnJava unit/print_indirect_method_postfix.t:16.
func TestFaithfulNestedCallParens(t *testing.T) {
	src := "package Foo; sub create { } package main; print create Foo sub { 1 };"
	root := parse.Parse([]byte(src))
	if containsKind(root, parse.Unknown) {
		t.Fatalf("%q: got %s", src, shape(root))
	}
	if ok, why := parse.Faithful(root, []byte(src)); !ok {
		t.Errorf("%q: canon %q is not faithful: %s", src, parse.Canon(root, []byte(src)), why)
	}
	// A paren that follows no name is not forgiven, nested or not: the
	// emission `print((1 + 2) * 3)` for source `print((1 + 2) * 3)`'s
	// misgrouped twin must still fail -- TestCanonFidelityCatchesMisgrouping
	// holds that at the top level.
}
