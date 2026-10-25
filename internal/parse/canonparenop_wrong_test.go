// ABOUTME: Hand-built wrong trees for canon-added parens after an operator: each regroups its source's parse.
// ABOUTME: Faithful judges every one unfaithful against its source, so a misgrouping cannot pass as a paren canon added.
package parse

import (
	"slices"
	"testing"
)

// requireUnfaithful fails t unless Faithful rejects the wrong tree n of src,
// and canon wrote a paren the source lacks.
func requireUnfaithful(t *testing.T, n *Node, src string) {
	t.Helper()
	if ok, _ := Faithful(n, []byte(src)); ok {
		t.Errorf("Faithful accepted the wrong tree of %q, emitted %q", src, Canon(n, []byte(src)))
	}
	if !slices.Contains(canonToks(n, src), "(") {
		t.Errorf("the wrong tree of %q emitted %q, with no paren of its grouping", src, Canon(n, []byte(src)))
	}
	if ok, why := Faithful(Parse([]byte(src)), []byte(src)); !ok {
		t.Errorf("the parse of %q is unfaithful: %s", src, why)
	}
}

// TestCanonAddedParenOnNamedUnaryRejectsRegrouping: measured on perl
// 5.44.0, Deparse,-p reads `defined $x || $y` as `(defined($x) or $y)`, so
// a tree with defined over `$x || $y` writes `defined ($x || $y)`, a paren
// the source lacks, and stays unfaithful.
func TestCanonAddedParenOnNamedUnaryRejectsRegrouping(t *testing.T) {
	n, src := wrongTree(t, "defined over ||")
	requireCanon(t, n, src, `defined ($x || $y);`)
	requireUnfaithful(t, n, src)
}

// TestCanonAddedParenOnNotKeepsLowPrecedence: measured on perl 5.44.0,
// Deparse,-p reads `my $v = not $x or 1` as `((my $v = (!$x)) or '???')`.
// A tree with `not` over `$x or 1` writes `not($x or 1)` and stays
// unfaithful.
func TestCanonAddedParenOnNotKeepsLowPrecedence(t *testing.T) {
	src := `my $v = not $x or 1;`
	n := Parse([]byte(src))
	st := lastStmt(t, n)
	or := st.Children[0]
	if or.Kind != Binary || or.Text != "or" || len(or.Children) != 2 {
		t.Fatalf("%q: want or at the top, got %v %q", src, or.Kind, or.Text)
	}
	decl := or.Children[0]
	not := decl.Children[len(decl.Children)-1]
	if not.Kind != Unary || not.Text != "not" {
		t.Fatalf("%q: want not as the initialiser, got %v %q", src, not.Kind, not.Text)
	}
	inner := &Node{Kind: Binary, Text: "or", Start: not.Children[0].Start, End: or.End,
		Children: []*Node{not.Children[0], or.Children[1]}}
	not.Children = []*Node{inner}
	not.End, decl.End = or.End, or.End
	st.Children[0] = decl
	requireUnfaithful(t, n, src)
}

// TestCanonAddedParenOnNegateBindsLooserThanPow: unary minus binds looser
// than `**`: measured on perl 5.44.0, `-2 ** 2` folds to -4. A tree with
// the minus as `**`'s left operand writes `(-2) ** 2` and stays unfaithful.
func TestCanonAddedParenOnNegateBindsLooserThanPow(t *testing.T) {
	src := `-2 ** 2;`
	n := Parse([]byte(src))
	st := lastStmt(t, n)
	neg := st.Children[0]
	if neg.Kind != Unary || len(neg.Children) != 1 || neg.Children[0].Text != "**" {
		t.Fatalf("%q: want a minus over **, got %v %q", src, neg.Kind, neg.Text)
	}
	pow := neg.Children[0]
	neg.Children = []*Node{pow.Children[0]}
	neg.End = pow.Children[0].End
	pow.Children[0] = neg
	pow.Start = neg.Start
	st.Children[0] = pow
	requireCanon(t, n, src, `(-2) ** 2;`)
	requireUnfaithful(t, n, src)
}

// TestCanonAddedParenAfterBackslashIsNotForgiven: measured on perl 5.44.0,
// `\(@a)` is refgen over a list and `\@a` is srefgen, so a tree with `\`
// over a parenthesised `@a` writes `\(@a)`, which against source `\@a` is
// a different tree and stays unfaithful.
func TestCanonAddedParenAfterBackslashIsNotForgiven(t *testing.T) {
	src := `\@a;`
	n := Parse([]byte(src))
	st := lastStmt(t, n)
	ref := st.Children[0]
	if ref.Kind != Unary || len(ref.Children) != 1 {
		t.Fatalf("%q: want a backslash over @a, got %v %q", src, ref.Kind, ref.Text)
	}
	a := ref.Children[0]
	ref.Children = []*Node{{Kind: List, Paren: true, Start: a.Start, End: a.End, Children: []*Node{a}}}
	requireCanon(t, n, src, `\(@a);`)
	requireUnfaithful(t, n, src)
}
