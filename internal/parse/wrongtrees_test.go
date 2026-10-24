// ABOUTME: Hand-built wrong trees for named unary ops, `not` and list ops, each a regrouping of its source's parse.
// ABOUTME: Faithful judges every one unfaithful against its source, with no forgiveness to hide the extra paren.
package parse

import (
	"strings"
	"testing"
)

// canonWrongTrees are the wrong trees of the named-unary, `not` and list-op
// canon tests: each regroups the parse of src, and canon writes the regrouping
// as canon. Each canon is a different program from src, judged by the pinned
// perl 5.44.0 (`perl -MO=Deparse,-p`) in those tests' notes.
var canonWrongTrees = []struct {
	name, src, canon string
	wrong            func(t *testing.T, n *Node, src string)
}{
	{"defined over ||", `defined $x || $y;`, `defined ($x || $y);`, rotateLast},
	{"ref over eq", `ref $x eq 'A';`, `ref ($x eq 'A');`, rotateLast},
	{"defined over a leading paren", `defined ($x) . "a";`, `defined(($x) . "a");`, definedOverParenDot},
	{"not over or", `not $x or 1;`, `not ($x or 1);`, notOverLast},
	{"== over not", `not $x == 1;`, `(not $x) == 1;`, notUnderLast},
	{"not over a leading paren", `not (1) + 1;`, `not((1) + 1);`, notOverLast},
	{"push over or", `push @a, 1 or die;`, `push (@a, 1 or die);`, rotateLast},
	{"bar takes one argument", `sub foo; sub bar; foo 1, bar 2, 3;`, `sub foo; sub bar; foo 1, bar (2), 3;`, barTakesOne},
	{"print over a leading paren", `print ($a), $b;`, `print(($a), $b);`, printOverParenList},
}

// wrongTree is the parse of the canonWrongTrees entry named name, regrouped.
func wrongTree(t *testing.T, name string) (*Node, string) {
	t.Helper()
	for _, w := range canonWrongTrees {
		if w.name == name {
			n := Parse([]byte(w.src))
			w.wrong(t, n, w.src)
			return n, w.src
		}
	}
	t.Fatalf("no wrong tree named %q", name)
	return nil, ""
}

// TestCanonWrongTreesAreUnfaithful: every wrong tree's canon is a paren the
// source lacks, and Faithful, forgiving only a block's last `;`, rejects it.
func TestCanonWrongTreesAreUnfaithful(t *testing.T) {
	for _, w := range canonWrongTrees {
		t.Run(w.name, func(t *testing.T) {
			n, src := wrongTree(t, w.name)
			requireCanon(t, n, src, w.canon)
			if ok, _ := Faithful(n, []byte(src)); ok {
				t.Errorf("Faithful accepted the wrong tree of %q, emitted %q", src, Canon(n, []byte(src)))
			}
		})
	}
}

// rotateLast regroups the last statement of n with rotate.
func rotateLast(t *testing.T, n *Node, _ string) {
	st := lastStmt(t, n)
	st.Children[0] = rotate(t, st.Children[0])
}

// notOverLast regroups the last statement of n with notOver.
func notOverLast(t *testing.T, n *Node, _ string) {
	st := lastStmt(t, n)
	st.Children[0] = notOver(t, st.Children[0])
}

// notUnderLast regroups the last statement of n with notUnder.
func notUnderLast(t *testing.T, n *Node, _ string) {
	st := lastStmt(t, n)
	st.Children[0] = notUnder(t, st.Children[0])
}

// definedOverParenDot is defined over `($x) . "a"`, the `($x)` a
// parenthesised Term spanning the source's parens, and no call parens.
func definedOverParenDot(t *testing.T, n *Node, src string) {
	st := lastStmt(t, n)
	wrong := rotate(t, st.Children[0])
	wrong.ArgParen = false
	dot := wrong.Children[0]
	x := *dot.Children[0]
	x.Paren = true
	x.Start = strings.Index(src, "(")
	x.End = strings.Index(src, ")") + 1
	dot.Children[0] = &x
	st.Children[0] = wrong
}

// barTakesOne is foo(1, bar(2), 3) for `foo 1, bar 2, 3`: bar's comma
// follows it, so bar takes call parens.
func barTakesOne(t *testing.T, n *Node, _ string) {
	st := lastStmt(t, n)
	foo := st.Children[0]
	outer := foo.Children[0] // 1 , bar(2 , 3)
	bar := outer.Children[1]
	inner := bar.Children[0] // 2 , 3
	two, three := inner.Children[0], inner.Children[1]
	wrongBar := *bar
	wrongBar.Children = []*Node{two}
	left := *outer
	left.Children = []*Node{outer.Children[0], &wrongBar}
	whole := *outer
	whole.Children = []*Node{&left, three}
	foo.Children = []*Node{&whole}
}

// printOverParenList is print over `($a), $b`: the bare list would begin
// with a `(` that closes early, which perl reads as the call's parens.
func printOverParenList(t *testing.T, n *Node, src string) {
	st := lastStmt(t, n)
	wrong := rotate(t, st.Children[0])
	wrong.ArgParen = false
	comma := wrong.Children[0]
	a := *comma.Children[0]
	a.Paren = true
	a.Start = strings.Index(src, "(")
	a.End = strings.Index(src, ")") + 1
	comma.Children[0] = &a
	st.Children[0] = wrong
}

// wrongTreeCanon is the canon of the canonWrongTrees entry named name.
func wrongTreeCanon(t *testing.T, name string) string {
	t.Helper()
	for _, w := range canonWrongTrees {
		if w.name == name {
			return w.canon
		}
	}
	t.Fatalf("no wrong tree named %q", name)
	return ""
}
