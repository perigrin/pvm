// ABOUTME: The three word operators below assignment — `and`, `or`, `xor` — parsed both parenthesised and bare.
// ABOUTME: The paren is not a style choice: it changes which program perl compiles, so both shapes are asserted.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestParenthesisedWordInfixParses: `($a and $b)` is an ordinary `&&`.
//
// Measured on perl 5.42.0 -- the paren regroups, so the whole expression is
// the initialiser:
//
//	$ perl -MO=Deparse -e 'my $a=1; my $b=0; my $y = ($a and $b);'
//	my $y = $a && $b;
//
// These refused with `not_a_term` because parseParenList parsed each element
// at the COMMA's binding power, which is above levels 4 and 5, so the word
// operator was never reached as an infix. `cmp` and `x` bind tighter than
// the comma and were clean for that reason alone.
func TestParenthesisedWordInfixParses(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want string
	}{
		{"my $y = ($a and $b);", "(my $y (and $a $b))"},
		{"my $y = ($a or $b);", "(my $y (or $a $b))"},
		{"my $y = ($a xor $b);", "(my $y (xor $a $b))"},
	} {
		root := parseOneExpr(t, tc.src)
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q: refused, want a parse", tc.src)
			continue
		}
		assertShape(t, root, tc.want)
	}

	// Not every parenthesised occurrence is an initialiser. Nothing about
	// this shape needs an assignment to its left, and the measurement that
	// filed the issue used this form to prove it.
	root := parseOneExpr(t, "print(($a and 1), \"\\n\");")
	if containsKind(root, parse.Unknown) {
		t.Errorf("print(($a and 1), ...): refused, want a parse")
	}
}

// TestBareWordInfixStillRegroups: without the paren, the assignment wins.
//
// This is the other half of the same cliff, and it is the reason the paren
// matters. `and` binds LOOSER than `=`, so the assignment takes only `$a`
// and the `and` sits above it:
//
//	$ perl -MO=Deparse -e 'my $a=1; my $b=0; my $y = $a and $b;'
//	$b if my $y = $a;
//
// So the bare and parenthesised forms are different programs, and a fix that
// made one parse by flattening it into the other would be a regression
// dressed as a pass. The declarator's initialiser must therefore stop at
// level 4/5 rather than swallow it.
func TestBareWordInfixStillRegroups(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want string
	}{
		{"my $y = $a and $b;", "(and (my $y $a) $b)"},
		{"my $y = $a or $b;", "(or (my $y $a) $b)"},
		{"my $y = $a xor $b;", "(xor (my $y $a) $b)"},
	} {
		root := parseOneExpr(t, tc.src)
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q: refused, want a parse", tc.src)
			continue
		}
		assertShape(t, root, tc.want)
	}

	// The COMMA is not one of the three and must still be reached from an
	// initialiser. Stopping the initialiser at the assignment's own right
	// power excluded it too and made this refuse, which is legal perl:
	//
	//	$ perl -MO=Deparse -e 'my $x = 1, my $y = 2;'
	//	my $x = 1, my $y = 2;
	//
	// So the floor is bpBelowComma, and this is the case that says why.
	//
	// KNOWN WRONG SHAPE, asserted deliberately so it cannot be mistaken for
	// a correct one. perl puts the comma ABOVE both declarations -- measured:
	//
	//	$ perl -MO=Concise -e 'my $x = 1, my $y = 2;'
	//	-     <1> ex-list vK ->8
	//	4        <1> padsv_store[$x:1,2] vKS/LVINTRO ->5
	//	6        <1> padsv_store[$y:1,2] vKS/LVINTRO ->7
	//
	// Two SIBLING stores under one list, so the perl shape is
	// `(, (my $x 1) (my $y 2))`. Ours nests the comma inside `$x`'s
	// initialiser instead. The plain-assignment form is already right --
	// `$x = 1, $y = 2` gives `(, (= $x 1) (= $y 2))` -- so the declarator is
	// the outlier, not the comma.
	//
	// Not fixed here, and the cost is why. Making it right needs the
	// declarator to resume the Pratt loop at the COMMA as well as at the
	// three word operators, and a comma after a declarator is not always a
	// statement-level comma:
	//
	//	ok(!sysopen(my $fh, $link, O_WRONLY|O_CREAT, 0600), ...)
	//
	// Measured with that resume in place, the declarator swallowed the rest
	// of the ARGUMENT list and canon then emitted a paren the source does
	// not contain -- 18 T1 files went unfaithful, `open my $fh, ...` and
	// `sysopen(my $fh, ...)` among them, e.g. `fcntl_nofollow.t` token 38:
	// "source has \"my\", canon has \"(\"". Telling the two commas apart
	// needs context the declarator does not have, so the wrong shape is
	// recorded rather than traded for a worse one.
	root := parseOneExpr(t, "my $x = 1, my $y = 2;")
	if containsKind(root, parse.Unknown) {
		t.Errorf("my $x = 1, my $y = 2: refused, want a parse")
	}
	const commaDeclKnownWrong = "(my $x (, 1 (my $y 2)))"
	if got := shape(skipWrappers(root)); got != commaDeclKnownWrong {
		t.Errorf("my $x = 1, my $y = 2: shape = %s, want the KNOWN-WRONG %s.\n"+
			"If this now reads %s, perl's own shape, the nesting was fixed: "+
			"assert that instead and delete this comment.",
			got, commaDeclKnownWrong, "(, (my $x 1) (my $y 2))")
	}
}

// TestBracketedWordInfixParses: `[...]` and `{...}` hold a full expr too.
//
// parseParenList was fixed for `($a and $b)` and parseBracketed -- the same
// element loop, one function away -- was not, so the bracketed forms still
// refused. Measured on perl 5.42.0, all three are legal and all three put the
// operator INSIDE the constructor:
//
//	$ perl -MO=Deparse -e 'my ($a,$b); my $y = [$a and $b];'
//	my $y = [$a && $b];
//	$ perl -MO=Deparse -e 'my ($a,$b); my $y = {$a and $b};'
//	my $y = {$a && $b};
//	$ perl -MO=Deparse -e 'my $y = [1, 2 and 3];'
//	my $y = [('???', 2) && 3];
//
// The symptom was worse than the paren's. The operand ESCAPED the bracket:
// `my $y = [$a and $b]` gave `(and (my $y [$a]) $b)`, a tree claiming `and`
// sat above the declaration with a ONE-element arrayref as its left operand.
// That is not a parse of the source -- the source has both operands inside
// the brackets.
func TestBracketedWordInfixParses(t *testing.T) {
	for _, tc := range []struct {
		src  string
		kind parse.Kind
		want string
	}{
		// `shape` prints a constructor's own text, which is empty for both
		// AnonArray and AnonHash, so the leading "(" with nothing after it
		// is the bracket node. The Kind check below is what separates the
		// array from the hash.
		{"my $y = [$a and $b];", parse.AnonArray, "(my $y ( (and $a $b)))"},
		{"my $y = {$a and $b};", parse.AnonHash, "(my $y ( (and $a $b)))"},
		{"my $y = [$a or $b];", parse.AnonArray, "(my $y ( (or $a $b)))"},
		{"my $y = [$a xor $b];", parse.AnonArray, "(my $y ( (xor $a $b)))"},

		// The comma case, which is the one that shows the list is the word
		// operator's LEFT operand rather than a sibling of it. The `( 1 2)`
		// is a List node -- Concise's `list` op -- so there are two
		// text-less nodes here: the outer bracket and the inner list.
		{"my $y = [1, 2 and 3];", parse.AnonArray, "(my $y ( (and ( 1 2) 3)))"},

		// A comma on the RIGHT belongs to the operator's right operand,
		// because `and` takes a listexpr there. Measured -- one `anonlist`,
		// `$a` alone on the left and a two-element list on the right:
		//
		//	7  <|> and(other->8) lK/1
		//	6     <0> padsv[$a]
		//	-     <1> ex-list lK
		//	8        <0> padsv[$b]
		//	9        <0> padsv[$c]
		{"my $y = [$a and $b, $c];", parse.AnonArray, "(my $y ( (and $a (, $b $c))))"},
	} {
		root := parseOneExpr(t, tc.src)
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q: refused, want a parse", tc.src)
			continue
		}
		if firstOfKind(root, tc.kind) == nil {
			t.Errorf("%q: no %v node, so the constructor itself was lost", tc.src, tc.kind)
		}
		assertShape(t, root, tc.want)
	}

	// Punctuation was never part of this -- `&&` is level 14, far above the
	// comma -- and it is asserted so a fix that broke it would fail here
	// rather than somewhere downstream.
	assertShape(t, parseOneExpr(t, "my $y = [$a && $b];"), "(my $y ( (&& $a $b)))")
}

// TestWordInfixGrouping: the measurements the design rests on.
//
// `(1, 2 and 3)` is quoted in parseInfix's doc comment, in parseParenList's
// code comment and in the commit message that introduced both, as the
// evidence that these three operators sit ABOVE the comma and take the whole
// list as a left operand. Nothing asserted it until now.
//
// Every shape here was confirmed against perl 5.42.0 before it was written.
// Associativity needed Concise or an explicit re-paren, because Deparse
// prints `$a ^^ $b ^^ $c` for the flat form and only re-parenthesises the
// grouping it does NOT get by default:
//
//	$ perl -MO=Deparse -e 'my ($a,$b,$c); my $y = (($a xor $b) xor $c);'
//	my $y = $a ^^ $b ^^ $c;
//	$ perl -MO=Deparse -e 'my ($a,$b,$c); my $y = ($a xor ($b xor $c));'
//	my $y = $a ^^ ($b ^^ $c);
//
// The left-grouped source deparses flat and the right-grouped one keeps its
// parens, so the bare form is LEFT associative.
func TestWordInfixGrouping(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want string
	}{
		// The design's own measurement:
		//	$ perl -MO=Deparse -e 'my @x = (1, 2 and 3);'
		//	my(@x) = ('???', 2) && 3;
		// The `and` took `(1, 2)` as its left operand.
		{"my $y = (1, 2 and 3);", "(my $y (and ( 1 2) 3))"},

		// Left associative, per the re-paren measurement above.
		{"my $y = ($a xor $b xor $c);", "(my $y (xor (xor $a $b) $c))"},

		// `and` (level 5) binds tighter than `or`/`xor` (level 4), so the
		// grouping is the same whichever order they appear in. Measured:
		//	$ perl -MO=Deparse -e 'my ($a,$b,$c); my $y = ($a and $b or $c);'
		//	my $y = $a && $b || $c;
		//	$ perl -MO=Deparse -e 'my ($a,$b,$c); my $y = ($a or $b and $c);'
		//	my $y = $a || $b && $c;
		{"my $y = ($a and $b or $c);", "(my $y (or (and $a $b) $c))"},
		{"my $y = ($a or $b and $c);", "(my $y (or $a (and $b $c)))"},
	} {
		root := parseOneExpr(t, tc.src)
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q: refused, want a parse", tc.src)
			continue
		}
		assertShape(t, root, tc.want)
	}
}
