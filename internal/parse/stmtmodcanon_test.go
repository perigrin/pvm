// ABOUTME: A statement modifier's canon: the expression is the BODY and the
// ABOUTME: trailing LIST or condition is the modifier's, not the other way round.

package parse_test

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestStatementModifierKeepsItsBody: `EXPR foreach LIST;` emits back as itself.
//
// The tree was already right -- applyModifier appends body then condition in
// SOURCE order and TestStatementModifierForms has checked the kinds and spans
// since it was written. What was wrong is the EMISSION: canon's Conditional
// and Loop case parenthesises the first non-block child as the condition,
// which is the block form's shape (`if (COND) BLOCK`) and the exact inverse of
// the modifier form's (`BODY if COND`). So every modifier canon'd with its two
// halves swapped, the keyword moved to the front, and the `;` dropped:
//
//	$s += $_ foreach 1..3;   ->   foreach ($s += $_) 1 .. 3
//	$y = 1 if $x;            ->   if ($y = 1) $x
//
// and all of them at Unknown = 0, because a swapped emission is not a refusal.
// This is the eleventh zero-Unknown-and-wrong defect in this parser and the
// assertion is therefore on the CANON: the count cannot see it and never could.
//
// perl reads them the other way round -- measured on 5.42.0:
//
//	$ perl -MO=Deparse -e 'my @a=(1,2); my $s=0; $s += $_ foreach 1..3;
//	                       print $_ foreach @a; push @r, $_ foreach @a;'
//	$s += $_ foreach (1 .. 3);
//	print $_ foreach (@a);
//	push @r, $_ foreach (@a);
//
// The parens perl's deparser writes round the list are its own habit, not
// grammar: the modifier's operand is a bare expression here and canon writes
// what the TREE holds, so an unparenthesised source stays unparenthesised.
func TestStatementModifierKeepsItsBody(t *testing.T) {
	for _, tc := range []struct{ src, want string }{
		// The six shapes the issue measured wrong, across both loop-list
		// keywords: a compound assignment, a parenless call, a call with
		// arguments, a plain assignment, a postfix increment, and a list
		// operator.
		{"$s += $_ foreach 1..3;", "$s += $_ foreach 1 .. 3;"},
		{"f() foreach @a;", "f() foreach @a;"},
		{"f($x) foreach @a;", "f($x) foreach @a;"},
		{"$s = 1 foreach @a;", "$s = 1 foreach @a;"},
		{"$x++ foreach @a;", "$x++ foreach @a;"},
		{"push @r, $_ foreach @a;", "push(@r , $_) foreach @a;"},

		// `for` is a synonym and shares the defect, so it shares the fix.
		{"$x++ for @a;", "$x++ for @a;"},
		{"push @r, $_ for @a;", "push(@r , $_) for @a;"},

		// The issue recorded that `if`, `unless`, `while` and `until`
		// "already canon correctly". They do NOT -- measured at 652f4792,
		// `$y = 1 if $x;` came back as `if ($y = 1) $x` at Unknown = 0, the
		// same swap. One canon case produced all six, so one fix ends all
		// six and these are not a regression guard but part of the defect.
		{"$y = 1 if $x;", "$y = 1 if $x;"},
		{"$y = 1 unless $x;", "$y = 1 unless $x;"},
		{"$y++ while $x;", "$y++ while $x;"},
		{"$y++ until $x;", "$y++ until $x;"},

		// A list operator body is where the swap was ugliest, because the
		// emission put a keyword in front of a call and a bare array after
		// it -- `foreach (push(@r , $_)) @a` is not a parse of anything.
		//
		// This case once expected `print $_ foreach @a;` verbatim, which is
		// what canon echoes for an UNKNOWN: `$_` was taken as print's handle,
		// the statement refused, and the test -- which asserts only the canon
		// -- matched the echo. Read correctly, the call gains its parens.
		{"print $_ foreach @a;", "print($_) foreach @a;"},

		// A body whose last token is a `}` but which is NOT a block form.
		// endsInBlock already declined the modifier for `defer { ... } if`;
		// this is the case it must keep accepting, and it is why the form
		// cannot be told from the presence of a Block in the subtree.
		{"$x = sub { 1 } if $y;", "$x = sub {1;} if $y;"},

		// `do BLOCK while COND` is a genuine modifier and perl spells it that
		// way. Measured on 5.42.0:
		//
		//	$ perl -MO=Deparse -e '$x=0; do {$x++;} while ($x < 10);'
		//	do { ++$x } while $x < 10;
		{"do { $x++; } while ($x < 10);", "do {$x++;} while $x < 10;"},

		// A PARENTHESISED condition. The parens are grouping here, not
		// grammar, so they belong to the child's Paren flag and canon writes
		// them back from there -- the modifier's own rule adds none.
		{"$y = 1 if ($x);", "$y = 1 if ($x);"},

		// The modifier binds looser than `or`, the loosest operator in the
		// expression grammar, so the whole `or` chain is the body. Measured:
		//
		//	$ perl -MO=Deparse -e 'print("a") or die("b") if $x;'
		//	print 'a' or die 'b' if $x;
		{"print \"a\" or die \"b\" if $x;", "print(\"a\") or die(\"b\") if $x;"},

		// A modifier on a DECLARATION, which reaches applyModifier by a
		// different path -- parseVarDecl stops at the keyword and the
		// statement site applies it.
		{"my $x = 1 if $c;", "my $x = 1 if $c;"},

		// A LABEL shifts the statement's start, which is why the form is
		// recorded by a flag and not by comparing the node's Start against
		// its first child's.
		{"LOOP: $x++ while $y;", "LOOP: $x++ while $y;"},
	} {
		src := []byte(tc.src)
		got := strings.TrimSpace(parse.Canon(parse.Parse(src), src))
		if got != tc.want {
			t.Errorf("Canon(parse(%q)):\n  got  %q\n  want %q", tc.src, got, tc.want)
		}
	}
}

// TestBlockFormLoopsKeepTheirParens is the regression guard the modifier fix
// needs, because both forms land in ONE canon case.
//
// `if (COND) BLOCK` really does parenthesise its first child, and the parens
// there are grammar rather than grouping -- perl will not accept `if $x {...}`.
// A fix that stopped writing them to cure the modifier would trade one wrong
// emission for another, so the block forms are pinned here. Every case below
// passes at 652f4792 and must keep passing.
func TestBlockFormLoopsKeepTheirParens(t *testing.T) {
	for _, tc := range []struct{ src, want string }{
		{"if ($x) { f(); }", "if ($x) {f();}"},
		{"unless ($x) { f(); }", "unless ($x) {f();}"},
		{"while ($x) { f(); }", "while ($x) {f();}"},
		{"until ($x) { f(); }", "until ($x) {f();}"},
		{"for (1..3) { f(); }", "for (1 .. 3) {f();}"},
		{"foreach my $x (@a) { f(); }", "foreach my $x (@a) {f();}"},
		{"if ($x) { f(); } else { g(); }", "if ($x) {f();} else {g();}"},
	} {
		src := []byte(tc.src)
		got := strings.TrimSpace(parse.Canon(parse.Parse(src), src))
		if got != tc.want {
			t.Errorf("Canon(parse(%q)):\n  got  %q\n  want %q", tc.src, got, tc.want)
		}
	}
}
