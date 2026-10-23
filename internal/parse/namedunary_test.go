// ABOUTME: Canon writes named unary ops by precedence: bare where the source was bare,
// ABOUTME: call parens only where the source wrote them (ArgParen) or binding power needs them.
package parse

import (
	"strings"
	"testing"
)

// The perl facts these tests rest on, measured with the pinned perl 5.44.0
// (`perl -MO=Deparse,-p -e SRC`):
//
//	defined $x || $y              (defined($x) or $y);
//	ref $x eq 'A'                 (ref($x) eq 'A');
//	!defined $x                   not(defined($x));
//	$a . lc $b                    ($a . lc($b));
//	chdir 't' if -d 't'           chdir('t') if (-d 't');
//	shift;                        shift(@ARGV);
//	-e $f                         (-e $f);
//	defined($x) || $y             (defined($x) or $y);
//	exists $h{a}                  exists $h{'a'};
//	defined $x ** 2               defined(($x ** 2));
//	defined($x) ** 2              (defined($x) ** 2);
//	defined ($x) . "a"            (defined($x) . 'a');
//	sub one($){1} one $x, 2       (one($x), '???');
//	use builtin qw(blessed); blessed $o, 1
//	                              (builtin::blessed($o), '???');
//	$o->shift(1); $o->keys        $o->shift(1); $o->keys;
//
// and `perl -MO=Concise` builds the same `chdir vK/1` for `chdir "t"` and
// `chdir("t")`: the call parens are not in perl's tree, so canon may write
// the bare spelling the source wrote.
//
// "Emits exactly" compares significant tokens, the comparison Faithful
// makes, so canon's ` , ` spacing is not pinned.

// canonToks gives the significant tokens of canon's emission of n.
func canonToks(n *Node, src string) []string {
	return significantTokens([]byte(Canon(n, []byte(src))))
}

func sameToks(a, b []string) bool {
	return strings.Join(a, "\x00") == strings.Join(b, "\x00")
}

// requireCanon fails unless canon of n emits exactly want's tokens.
func requireCanon(t *testing.T, n *Node, src, want string) {
	t.Helper()
	got := canonToks(n, src)
	if !sameToks(got, significantTokens([]byte(want))) {
		t.Errorf("Canon(%q) = %q, want %q", src, Canon(n, []byte(src)), want)
	}
}

// lastStmt is n's last non-empty statement.
func lastStmt(t *testing.T, n *Node) *Node {
	t.Helper()
	for i := len(n.Children) - 1; i >= 0; i-- {
		if c := n.Children[i]; c.Kind == Statement && len(c.Children) > 0 {
			return c
		}
	}
	t.Fatalf("no statement in %s", n.Kind)
	return nil
}

// findCall is the first Call named name under n, in source order.
func findCall(n *Node, name string) *Node {
	if n.Kind == Call && n.Text == name {
		return n
	}
	for _, c := range n.Children {
		if f := findCall(c, name); f != nil {
			return f
		}
	}
	return nil
}

// rotate turns OP(NAME(a), b) into NAME(OP(a, b)): the named op taking the
// whole expression, which is the wrong tree a misgrouping parser would build.
func rotate(t *testing.T, x *Node) *Node {
	t.Helper()
	if (x.Kind != Binary && x.Kind != CmpChain) || len(x.Children) != 2 {
		t.Fatalf("rotate: %s %q is not a two-operand operator", x.Kind, x.Text)
	}
	l, r := x.Children[0], x.Children[1]
	if l.Kind != Call || len(l.Children) != 1 {
		t.Fatalf("rotate: left operand %s %q is not a one-argument call", l.Kind, l.Text)
	}
	a := l.Children[0]
	inner := *x
	inner.Children = []*Node{a, r}
	inner.Start = a.Start
	c := *l
	c.Children = []*Node{&inner}
	c.End = r.End
	return &c
}

// TestCanonWritesNamedUnaryByPrecedence: a named unary op the source wrote
// bare is re-emitted bare, and its operand is the one perl reads.
//
// A user sub with a `($)` prototype the file cannot see is parsed as a list
// operator, and is not covered here: that is M4 prototype territory, and the
// list-op issue writes it.
func TestCanonWritesNamedUnaryByPrecedence(t *testing.T) {
	for _, src := range []string{
		`defined $x || $y;`,
		`ref $x eq 'A';`,
		`!defined $x;`,
		`$a . lc $b;`,
		`chdir 't' if -d 't';`,
		`chdir "t";`,
		`shift;`,
		`-e $f;`,
		`exists $h{a};`,
		`defined $x ** 2;`,
		`lc $s;`,
		`scalar @a;`,
		`sub one($){1;} one $x, 2;`,
		`use builtin qw(blessed); blessed $o, 1;`,
	} {
		n := Parse([]byte(src))
		requireCanon(t, n, src, src)
		if ok, why := Faithful(n, []byte(src)); !ok {
			t.Errorf("Faithful(%q) = false: %s", src, why)
		}
	}
}

// TestCanonNamedUnaryKeepsSourceArgParen: the source's `NAME(` survives,
// and a tree that lost the ArgParen flag emits the bare spelling, which is
// a difference from the source.
func TestCanonNamedUnaryKeepsSourceArgParen(t *testing.T) {
	src := `defined($x) || $y;`
	n := Parse([]byte(src))
	requireCanon(t, n, src, src)

	call := findCall(n, "defined")
	if call == nil || !call.ArgParen || !call.NamedUnary {
		t.Fatalf("defined in %q: want a Call with ArgParen and NamedUnary, got %+v", src, call)
	}
	call.ArgParen = false
	requireCanon(t, n, src, `defined $x || $y;`)
	if sameToks(canonToks(n, src), significantTokens([]byte(src))) {
		t.Errorf("a tree without ArgParen re-emits %q unchanged", src)
	}
}

// TestCanonNamedUnaryWrapsLooserOperand: a wrong tree's grouping is a paren
// the source lacks. Canon's text only; the verdict is asserted once the
// call-paren forgiveness is removed.
func TestCanonNamedUnaryWrapsLooserOperand(t *testing.T) {
	for _, tc := range []struct{ src, want string }{
		{`defined $x || $y;`, `defined ($x || $y);`},
		{`ref $x eq 'A';`, `ref ($x eq 'A');`},
	} {
		n := Parse([]byte(tc.src))
		st := lastStmt(t, n)
		st.Children[0] = rotate(t, st.Children[0])
		requireCanon(t, n, tc.src, tc.want)
	}
}

// TestCanonNamedUnaryParenthesisesLeadingParenOperand: perl reads `NAME (`
// as call parens, so an operand whose emission begins with a `(` that does
// not close at its end would re-read as the call's argument list. Canon
// writes the call parens there.
func TestCanonNamedUnaryParenthesisesLeadingParenOperand(t *testing.T) {
	src := `defined ($x) . "a";`
	n := Parse([]byte(src))
	requireCanon(t, n, src, src)

	// The wrong tree: defined over `($x) . "a"`, the `($x)` a parenthesised
	// Term spanning the source's parens, and no call parens.
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
	requireCanon(t, n, src, `defined(($x) . "a");`)
}

// TestCanonNamedUnaryUsesFollowPower: a prefix op's parens depend on the
// operator written after it, not on the power its parent reads it at.
func TestCanonNamedUnaryUsesFollowPower(t *testing.T) {
	for _, src := range []string{
		`!defined $x;`,
		`$a . lc $b;`,
		`defined $x ** 2;`,
		`defined($x) ** 2;`,
		`$c ? defined $x : lc $y;`,
	} {
		requireCanon(t, Parse([]byte(src)), src, src)
	}

	// Without the source's parens, `**` after the call is tighter than a
	// named unary and would bind into its operand, so the parens come back.
	src := `defined($x) ** 2;`
	n := Parse([]byte(src))
	findCall(n, "defined").ArgParen = false
	requireCanon(t, n, src, src)
}

// TestCanonVisibleUnaryPrototypeIsNamedUnary: a sub the file declares with a
// unary prototype, and a `builtin::` unary import, parse as named unaries
// and are written as them; a list prototype, or none, is not one.
func TestCanonVisibleUnaryPrototypeIsNamedUnary(t *testing.T) {
	for _, tc := range []struct {
		src, name string
		unary     bool
	}{
		{`sub one($){1;} one $x, 2;`, "one", true},
		{`sub one(_){1;} one $x, 2;`, "one", true},
		{`sub one(;$){1;} one $x, 2;`, "one", true},
		{`use builtin qw(blessed); blessed $o, 1;`, "blessed", true},
		{`sub two($$){1;} two $x, 2;`, "two", false},
		{`sub three{1;} three $x, 2;`, "three", false},
	} {
		n := Parse([]byte(tc.src))
		// A list operator is written by the list-op issue, not here.
		if tc.unary {
			requireCanon(t, n, tc.src, tc.src)
		}
		call := findCall(lastStmt(t, n), tc.name)
		if call == nil {
			t.Fatalf("%q: no Call %s", tc.src, tc.name)
		}
		if call.NamedUnary != tc.unary {
			t.Errorf("%q: %s NamedUnary = %v, want %v", tc.src, tc.name, call.NamedUnary, tc.unary)
		}
	}
}

// TestCanonNamedUnaryIgnoresMethodNames: after `->` a word is a method name,
// never a named unary, and keeps the spelling it was written with.
func TestCanonNamedUnaryIgnoresMethodNames(t *testing.T) {
	for _, src := range []string{`$o->shift(1);`, `$o->keys;`, `$o->defined($x) || $y;`} {
		n := Parse([]byte(src))
		requireCanon(t, n, src, src)
		if c := findCall(n, "shift"); c != nil && c.NamedUnary {
			t.Errorf("%q: method shift is marked NamedUnary", src)
		}
	}
}
