// ABOUTME: Hand-built wrong trees for canon-added parens on terms and argument lists: each regroups its source's parse.
// ABOUTME: Faithful judges every one unfaithful against its source, so a misgrouping cannot pass as a paren canon added.
package parse

import "testing"

// TestCanonAddedParenOnNestedListOpKeepsPerlGrouping: measured on perl
// 5.44.0, Deparse,-p reads `foo 1, bar 2, 3` (both subs predeclared) as
// `foo(1, bar(2, 3))`, so a tree with bar taking one argument writes `bar
// (2)` and stays unfaithful.
func TestCanonAddedParenOnNestedListOpKeepsPerlGrouping(t *testing.T) {
	n, src := wrongTree(t, "bar takes one argument")
	requireCanon(t, n, src, wrongTreeCanon(t, "bar takes one argument"))
	requireUnfaithful(t, n, src)
}

// TestCanonAddedParenOnCmpChainIsMisgrouping: measured on perl 5.44.0,
// Deparse,-p writes `1 < $x < 3` with no inner pair, while `(1 < $x) < 3`
// compiles to two lt ops, the inner one `lt sKP/2`. A tree with the chain
// split into two comparisons writes `(1 < $x) < 3` and stays unfaithful.
func TestCanonAddedParenOnCmpChainIsMisgrouping(t *testing.T) {
	src := `1 < $x < 3;`
	n := Parse([]byte(src))
	st := lastStmt(t, n)
	chain := st.Children[0]
	if chain.Kind != CmpChain || len(chain.Children) != 3 {
		t.Fatalf("%q: want a three-operand chain, got %v %q", src, chain.Kind, chain.Text)
	}
	a, x, b := chain.Children[0], chain.Children[1], chain.Children[2]
	inner := &Node{Kind: Binary, Text: "<", Start: a.Start, End: x.End, Children: []*Node{a, x}}
	st.Children[0] = &Node{Kind: Binary, Text: "<", Start: a.Start, End: b.End, Children: []*Node{inner, b}}
	requireCanon(t, n, src, `(1 < $x) < 3;`)
	requireUnfaithful(t, n, src)
}

// TestCanonArgParenDoesNotCoverInnerTermParens: measured on perl 5.44.0,
// Deparse,-p reads `print (1+2)*3` as `(print(3) * 3)`. A tree with print
// over the whole product writes `print ((1 + 2) * 3)`, a different token
// sequence from the source, and no forgiveness covers it.
func TestCanonArgParenDoesNotCoverInnerTermParens(t *testing.T) {
	src := `print (1+2)*3;`
	n := Parse([]byte(src))
	st := lastStmt(t, n)
	mul := st.Children[0]
	if mul.Kind != Binary || mul.Text != "*" || len(mul.Children) != 2 || mul.Children[0].Kind != Call {
		t.Fatalf("%q: want * over print, got %v %q", src, mul.Kind, mul.Text)
	}
	print := mul.Children[0]
	sum := print.Children[0]
	sum.Paren = true
	mul.Children[0] = sum
	mul.Start = sum.Start
	print.Children = []*Node{mul}
	print.ArgParen = false
	print.End = mul.End
	st.Children[0] = print
	requireUnfaithful(t, n, src)
}
