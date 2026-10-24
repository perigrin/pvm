// ABOUTME: Canon writes `not` by the follow power: bare, its operand read above NOTOP,
// ABOUTME: grouped only when a tighter operator follows, and call parens for a leading `(`.
package parse

import "testing"

// The perl facts these tests rest on, measured with the pinned perl 5.44.0
// (`perl -MO=Deparse,-p -e SRC`):
//
//	not $x or $z                  ($x and $z);
//	my $v = not $x or 1           ((my $v = (!$x)) or '???');
//	not ($y) + 1                  ((!$y) + 1);
//	not (($y) + 1)                not(($y + 1));
//	print not $x                  print((!$x));
//	print not $x, $y              print((!($x, $y)));
//	not $x == 1                   not(($x == 1));
//	(not $x) == 1                 ((!$x) == 1);
//	my @l = (not $x, 1)           (my(@l) = (!($x, 1)));
//	[(not $x), 1]                 [(!$x), 1];
//	[not $x, 1]                   [(!($x, 1))];
//	not ($a == 1) + 1             ((!($a == 1)) + 1);
//	not ($x or 1)                 not(($x || 1));
//	!not $x                       not((!$x));
//	$a + not $x, $y               ($a + (!($x, $y)));
//	$c ? not $x : 1               ($c ? (!$x) : '???');
//
// `not` is a term (perly.y's `NOTOP listexpr`): what comes before it never
// splits it, and its operand runs right until a looser `and`, `or` or `xor`.
// So only the operator AFTER it decides whether it is grouped.
//
// "Emits exactly" compares significant tokens, the comparison Faithful
// makes, so canon's ` , ` spacing is not pinned.

// findNot is the first `not` under n, in source order.
func findNot(n *Node) *Node {
	if n.Kind == Unary && n.Text == "not" {
		return n
	}
	for _, c := range n.Children {
		if f := findNot(c); f != nil {
			return f
		}
	}
	return nil
}

// notOver turns OP(not(a), b) into not(OP(a, b)): `not` taking the whole
// expression.
func notOver(t *testing.T, x *Node) *Node {
	t.Helper()
	if x.Kind != Binary || len(x.Children) != 2 {
		t.Fatalf("notOver: %s %q is not a binary", x.Kind, x.Text)
	}
	l, r := x.Children[0], x.Children[1]
	if l.Kind != Unary || l.Text != "not" {
		t.Fatalf("notOver: left operand %s %q is not a not", l.Kind, l.Text)
	}
	inner := *x
	inner.Children = []*Node{l.Children[0], r}
	inner.Start = l.Children[0].Start
	not := *l
	not.Children = []*Node{&inner}
	not.End = r.End
	return &not
}

// notUnder turns not(OP(a, b)) into OP(not(a), b): `not` taking only the
// left operand.
func notUnder(t *testing.T, not *Node) *Node {
	t.Helper()
	if not.Kind != Unary || not.Text != "not" || len(not.Children) != 1 {
		t.Fatalf("notUnder: %s %q is not a not", not.Kind, not.Text)
	}
	x := not.Children[0]
	if (x.Kind != Binary && x.Kind != CmpChain) || len(x.Children) != 2 {
		t.Fatalf("notUnder: operand %s %q is not a two-operand operator", x.Kind, x.Text)
	}
	inner := *not
	inner.Children = []*Node{x.Children[0]}
	inner.End = x.Children[0].End
	outer := *x
	outer.Children = []*Node{&inner, x.Children[1]}
	outer.Start = not.Start
	return &outer
}

// TestCanonWritesNotByPrecedence: a `not` the source wrote bare is
// re-emitted bare, and its operand is the one perl reads.
func TestCanonWritesNotByPrecedence(t *testing.T) {
	for _, src := range []string{
		`not $x or 1;`,
		`my $v = not $x or 1;`,
		`not (1) + 1;`,
		`not $x == 1;`,
		`not ($a == 1) + 1;`,
		`!not $x;`,
		`$c ? not $x : 1;`,
		`$a + not $x, $y;`,
		`(not $x) == 1;`,
		`my @l = (not $x, 1);`,
	} {
		n := Parse([]byte(src))
		requireCanon(t, n, src, src)
		if ok, why := Faithful(n, []byte(src)); !ok {
			t.Errorf("Faithful(%q) = false: %s", src, why)
		}
	}
	requireNotBareUnderListOp(t)
}

// TestCanonNotStaysBareUnderListOp: `not` is never wrapped as a list op's
// operand and never given call parens. perl reads `print not $x, $y` as
// `print((!($x, $y)))`: `not` takes the whole comma list.
func TestCanonNotStaysBareUnderListOp(t *testing.T) {
	requireNotBareUnderListOp(t)
}

// requireNotBareUnderListOp checks `not` under `print`. The list op's own
// argument parens are still canon's, so they are the only parens written:
// the list-op issue writes the list op by precedence and drops them.
func requireNotBareUnderListOp(t *testing.T) {
	t.Helper()
	for _, tc := range []struct{ src, want string }{
		{`print not $x;`, `print(not $x);`},
		{`print not $x, $y;`, `print(not $x, $y);`},
	} {
		n := Parse([]byte(tc.src))
		requireCanon(t, n, tc.src, tc.want)
		if ok, why := Faithful(n, []byte(tc.src)); !ok {
			t.Errorf("Faithful(%q) = false: %s", tc.src, why)
		}
	}
}

// TestCanonNotWrapsLooserOperand: a wrong tree's grouping is a paren the
// source lacks. Canon's text only; the verdict is asserted once the
// call-paren forgiveness is removed.
func TestCanonNotWrapsLooserOperand(t *testing.T) {
	src := `not $x or 1;`
	n := Parse([]byte(src))
	st := lastStmt(t, n)
	st.Children[0] = notOver(t, st.Children[0])
	requireCanon(t, n, src, `not ($x or 1);`)
}

// TestCanonNotKeepsGroupingUnderTighterFollow: `not` is grouped when the
// operator after it binds tighter than NOTOP, unless its operand is
// parenthesised, which perl reads as `not (X)`, tighter. And an operand
// whose emission opens with a `(` that closes early takes call parens, or
// perl would read that `(` as the whole operand.
func TestCanonNotKeepsGroupingUnderTighterFollow(t *testing.T) {
	// `==` over `(not $x)`: perl reads `not $x == 1` as `not(($x == 1))`,
	// so the grouping must be written.
	src := `not $x == 1;`
	n := Parse([]byte(src))
	requireCanon(t, n, src, src)
	st := lastStmt(t, n)
	st.Children[0] = notUnder(t, st.Children[0])
	requireCanon(t, n, src, `(not $x) == 1;`)

	// `not` over `(1) + 1`: perl reads `not (1) + 1` as `(!1) + 1`, so the
	// source's own spelling would be a different program.
	src = `not (1) + 1;`
	n = Parse([]byte(src))
	requireCanon(t, n, src, src)
	st = lastStmt(t, n)
	st.Children[0] = notOver(t, st.Children[0])
	requireCanon(t, n, src, `not((1) + 1);`)

	// A bracketed list's comma follows each element but the last, so a
	// `not` element is grouped there even without the source's parens:
	// `[not $x, 1]` would be `[(!($x, 1))]`.
	src = `[(not $x), 1];`
	n = Parse([]byte(src))
	requireCanon(t, n, src, src)
	not := findNot(n)
	if not == nil || !not.Paren {
		t.Fatalf("%q: want a parenthesised not, got %+v", src, not)
	}
	not.Paren = false
	requireCanon(t, n, src, src)
}

// TestCanonNotKeepsArrowCallParens: the parens of `->(` are the call's
// argument list, and a parenthesised `not` inside them is written once.
// Measured on perl 5.44.0, `$f->(not $x)` is `&$f((!$x))`.
func TestCanonNotKeepsArrowCallParens(t *testing.T) {
	for _, src := range []string{
		`$f->(not $x);`,
		`$f->(not $x, 1);`,
		`$f->(not ($x));`,
	} {
		n := Parse([]byte(src))
		requireCanon(t, n, src, src)
		if ok, why := Faithful(n, []byte(src)); !ok {
			t.Errorf("Faithful(%q) = false: %s", src, why)
		}
	}
}
