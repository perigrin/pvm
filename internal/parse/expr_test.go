// ABOUTME: The expression parser's properties: 32 levels, associativity, nonassoc rejection, chaining.
// ABOUTME: Every grouping fact here was measured against real perl, not recalled from the table.

package parse_test

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestPrecedenceTableComplete: all 32 perly.y levels are accounted for.
//
// Counting map entries would not do it -- several levels share a binding
// power, so a missing level hides. The table records which levels carry no
// lexable operator and why, and this asserts the set is exactly 1..32.
func TestPrecedenceTableComplete(t *testing.T) {
	levels := parse.LevelsAccountedFor()
	if len(levels) != 32 {
		t.Errorf("got %d levels, want 32 (grep -cE '^%%(left|right|nonassoc)' perly.y)", len(levels))
	}
	for i := 1; i <= 32; i++ {
		if _, ok := levels[i]; !ok {
			t.Errorf("level %d is unaccounted for", i)
		}
	}

	// And the nonassoc count, also from perly.y:
	//   grep -cE '^%nonassoc' perly.y  ->  11
	// Of those 11, five carry no lexable operator (1, 3, 18, 28, 30) and
	// three are prefix or statement forms handled elsewhere (2, 7, 20), so
	// the infix table holds the remaining three: .. ... ++ --.
	if got := parse.NonassocOperators(); len(got) != 4 {
		t.Errorf("nonassoc infix operators = %v, want 4 (.. ... ++ --)", got)
	}
}

// TestRightAssociative: 2**3**2 is 512, not 64. Measured.
func TestRightAssociative(t *testing.T) {
	// $a ** $b ** $c groups as $a ** ($b ** $c), so the RIGHT operand of the
	// first ** is itself a ** node.
	root := parseOneExpr(t, "$a ** $b ** $c")
	assertShape(t, root, "(** $a (** $b $c))")
}

// TestLeftAssociative: 1-2-3 is -4, not 2. Measured.
func TestLeftAssociative(t *testing.T) {
	root := parseOneExpr(t, "$a - $b - $c")
	assertShape(t, root, "(- (- $a $b) $c)")
}

// TestNonassocRejects is the case a binding-power table cannot express.
//
// A power that stops the recursion still ACCEPTS the input; perly.y's
// %nonassoc makes it an error. Measured:
//
//	$ perl -e 'my $x = 1 .. 2 .. 3;'
//	syntax error at -e line 1, near "2 .."
//
// Eleven of the 32 levels are nonassoc, so this is not an edge case.
func TestNonassocRejects(t *testing.T) {
	for _, src := range []string{
		"$a .. $b .. $c",
		"$a ... $b ... $c",
	} {
		root := parseOneExpr(t, src)
		if !containsKind(root, parse.Unknown) {
			t.Errorf("%q: a repeated nonassoc operator is a syntax error; "+
				"got %s with no Unknown", src, shape(root))
		}
	}

	// And the single use is fine.
	root := parseOneExpr(t, "$a .. $b")
	if containsKind(root, parse.Unknown) {
		t.Errorf("$a .. $b is valid: got %s", shape(root))
	}
}

// TestComparisonChains: perl 5.32+ builds ONE n-ary node, not nested
// binaries, so each operand is evaluated once. Measured:
//
//	$ perl -MO=Deparse,-p -e 'my $q = $a < $b < $c'
//	(my($q) = ($a < $b < $c));
//
// §4.3 splits the comparison tokens into chaining and non-chaining classes;
// the two yyerror productions of perly.y:1466-1478 are the whole rule.
func TestComparisonChains(t *testing.T) {
	// Three operands under ONE node, not nested binaries. The shape string
	// cannot tell a chain from a binary by itself -- both print as
	// `(< a b ...)` -- so the kind is asserted separately.
	root := parseOneExpr(t, "$a < $b < $c")
	assertShape(t, root, "(< $a $b $c)")
	if n := firstOfKind(root, parse.CmpChain); n == nil {
		t.Error("$a < $b < $c must build a CmpChain, not nested binaries")
	} else if len(n.Children) != 3 {
		t.Errorf("chain has %d operands, want 3", len(n.Children))
	}

	// A non-chaining operator at the same level may not chain. Measured:
	//   $ perl -e 'my $q = $a <=> $b <=> $c'
	//   syntax error at -e line 1, near "$b <=>"
	for _, src := range []string{"$a <=> $b <=> $c", "$a cmp $b cmp $c"} {
		root := parseOneExpr(t, src)
		if !containsKind(root, parse.Unknown) {
			t.Errorf("%q: NCEQOP cannot chain; got %s", src, shape(root))
		}
	}
}

// TestComparisonAcrossLevels: the yyerror productions of perly.y reject a
// relational operator after a termrelop and an equality operator after a
// termeqop, but relop binds tighter than eqop, so one of each is a
// termeqop over a termrelop. A parenthesised comparison is a term, so
// nothing stands to its left. Measured with perl 5.42:
//
//	$ perl -MO=Deparse,-p -e '$a < $b == $c; $a == $b < $c; ($a < $b) < $c;'
//	(($a < $b) == $c);
//	($a == ($b < $c));
//	(($a < $b) < $c);
//	$ perl -e '$a <=> $b == $c'
//	syntax error at -e line 1, near "$b =="
func TestComparisonAcrossLevels(t *testing.T) {
	for _, src := range []string{
		"$a < $b == $c",
		"$a == $b < $c",
		"$a lt $b <=> $c",
		"($a < $b) < $c",
		"($a == $b) == $c",
		"(-(-$x) < 0) == ($x < 0)",
	} {
		root := parseOneExpr(t, src)
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q: perl accepts this; got %s", src, shape(root))
		}
	}
	for _, src := range []string{"$a <=> $b == $c", "$a == $b <=> $c"} {
		root := parseOneExpr(t, src)
		if !containsKind(root, parse.Unknown) {
			t.Errorf("%q: two eqop-level operators; got %s", src, shape(root))
		}
	}
}

// TestParenthesisedComparisonIsATerm: parentheses end a termrelop or
// termeqop, so a comparison in them neither joins the chain beside it nor
// blocks a second operator at its own level. Measured with perl 5.42:
//
//	$ perl -MO=Deparse,-p -e '($a == $b) == 0; ($a cmp $b) == 0;'
//	(($a == $b) == 0);
//	(($a cmp $b) == 0);
//	$ perl -e 'my ($a,$b)=(1,2); print +(($a == $b) == 0) ? "y":"n", (($a == $b == 0) ? "y":"n")'
//	yn
//
// So the grouped form and the chain are different programs, and canon must
// keep the parentheses.
func TestParenthesisedComparisonIsATerm(t *testing.T) {
	for _, src := range []string{
		"ok(($a == $b) == 0);",
		"ok(($a cmp $b) == 0);",
		"ok(($a <=> $b) <=> 0);",
		"ok(($a < $b) < $c);",
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q: perl accepts this; got %s", src, shape(root))
			continue
		}
		if got := parse.Canon(root, []byte(src)); got != src {
			t.Errorf("%q: canon %q drops the grouping", src, got)
		}
	}
}

// TestTernaryAndAssignment: both are right associative and they nest at
// different levels. Measured:
//
//	$ perl -MO=Deparse -e 'my $x = $a ? $b : $c ? $d : $e;'
//	my $x = $a ? $b : ($c ? $d : $e);
func TestTernaryAndAssignment(t *testing.T) {
	root := parseOneExpr(t, "$a ? $b : $c ? $d : $e")
	assertShape(t, root, "(?: $a $b (?: $c $d $e))")

	// Assignment is right associative and BELOW the ternary, so a ternary on
	// the right of an `=` is the whole right operand.
	root = parseOneExpr(t, "$x = $a ? $b : $c")
	assertShape(t, root, "(= $x (?: $a $b $c))")

	root = parseOneExpr(t, "$a = $b = $c")
	assertShape(t, root, "(= $a (= $b $c))")
}

// TestConcatIsAnAddop pins §4.1's most surprising row: `.` is level 22, the
// same as `+`. Measured:
//
//	$ perl -MO=Deparse -e 'my $x = "a" . 1 + 2;'
//	my $x = 'a1' + 2;
func TestConcatIsAnAddop(t *testing.T) {
	root := parseOneExpr(t, "$a . $b + $c")
	assertShape(t, root, "(+ (. $a $b) $c)")
}

// TestPowerBindsTighterThanUnaryMinus: -2**2 is -4, not 4. Measured.
func TestPowerBindsTighterThanUnaryMinus(t *testing.T) {
	root := parseOneExpr(t, "- $a ** $b")
	assertShape(t, root, "(neg (** $a $b))")
}

// TestTermForms: §4.4's term forms all parse. The Pratt loop needs a nud for
// each of these or it has nothing to bind.
//
// §4.4.5's `\` is not optional scaffolding -- it is the srefgen signal the
// whole fidelity harness measures.
func TestTermForms(t *testing.T) {
	for _, src := range []string{
		// 4.4.1 literals
		"42", "3.14", "0x1f", "'str'", `"str"`,
		// 4.4.2 variables and sigils
		"$x", "@a", "%h", "$#a", "$$r", "@$r", "${$r}",
		// 4.4.4 anonymous constructors
		"[1, 2]", "{ a => 1 }", "[]", "{}",
		// 4.4.5 references
		`\$x`, `\@a`, `\%h`, `\&f`,
		// 4.4.6 dereference
		"$r->[0]", "$r->{k}", "$$r[0]", "${$r}{k}",
		// 4.4.7 slices
		"@a[0,1]", "@h{'a','b'}",
		// parens
		"(1, 2, 3)", "($x)",
	} {
		root := parseOneExpr(t, src)
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q is a §4.4 term form and must parse: got %s", src, shape(root))
		}
	}
}

// TestExpressionUnknownNotGuess: an operand the parser cannot read becomes
// Unknown rather than a guess. The plan's rule applies inside expressions as
// much as at statement level.
func TestExpressionUnknownNotGuess(t *testing.T) {
	root := parseOneExpr(t, "$a + ")
	if !containsKind(root, parse.Unknown) {
		t.Errorf("a missing operand must be Unknown, not invented: %s", shape(root))
	}
}

// --- helpers ---

// parseOneExpr parses src as a single expression statement and returns the
// expression node.
func parseOneExpr(t *testing.T, src string) *parse.Node {
	t.Helper()
	root := parse.Parse([]byte(src))
	if got := leafText(root, []byte(src)); got != src {
		t.Fatalf("%q: round-trip failed before shape could be checked, got %q", src, got)
	}
	return root
}

func shape(n *parse.Node) string {
	var b strings.Builder
	var walk func(*parse.Node)
	walk = func(n *parse.Node) {
		if len(n.Children) == 0 {
			b.WriteString(n.Text)
			return
		}
		b.WriteString("(")
		b.WriteString(n.Text)
		for _, c := range n.Children {
			b.WriteString(" ")
			walk(c)
		}
		b.WriteString(")")
	}
	walk(n)
	return b.String()
}

// skipWrappers descends past the source_file and statement nodes to the
// expression itself, so a shape reads as grouping rather than as tree
// plumbing.
func skipWrappers(root *parse.Node) *parse.Node {
	n := root
	for len(n.Children) == 1 && n.Text == "" {
		n = n.Children[0]
	}
	return n
}

// assertShape compares the expression's tree against an s-expression.
func assertShape(t *testing.T, root *parse.Node, want string) {
	t.Helper()
	if got := shape(skipWrappers(root)); got != want {
		t.Errorf("shape = %s, want %s", got, want)
	}
}

func firstOfKind(n *parse.Node, k parse.Kind) *parse.Node {
	if n.Kind == k {
		return n
	}
	for _, c := range n.Children {
		if got := firstOfKind(c, k); got != nil {
			return got
		}
	}
	return nil
}

func containsKind(n *parse.Node, k parse.Kind) bool {
	if n.Kind == k {
		return true
	}
	for _, c := range n.Children {
		if containsKind(c, k) {
			return true
		}
	}
	return false
}
