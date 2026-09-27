// ABOUTME: A `return`'s operand is an EXPRESSION, and a loop control may carry
// ABOUTME: a statement modifier -- two zero-Unknown wrong trees in one path.

package parse_test

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestReturnList: `return LIST;` emits its operands.
//
// Issue 01a0e013 measured `return 1, 2;` canoning as `return ,;` and
// `return ($a, $b);` as `return ;` -- both operands gone, at Unknown = 0.
//
// It attributed the loss to `parseReturn`'s `p.parseExpr(bpListOp)`, on the
// reading that the comma is not consumed at that binding power. That reading is
// WRONG and the tree proves it: the comma is level 8, BP 80, and bpListOp is
// 70, so the comma binds and the parse is already right --
//
//	loop_control "return"
//	  binary ","
//	    term "1"
//	    term "2"
//
// The loss is in CANON. Its LoopControl case writes `c.Text` for each child,
// which is the right thing for a Label -- `next L` names a loop, and emitting
// the Label node's own form would write `next L: ` -- and the wrong thing for
// anything with structure. A Binary's Text is the operator alone, a
// parenthesised List's is empty, and a Ternary's is "?:", so every one of
// those canon'd as its punctuation with the operands deleted.
//
// That is why the ternary case is here alongside the comma: it is the same
// defect, found by probing past the shape the issue named, and no comma is
// involved in it at all.
//
// perl keeps the operands -- measured on 5.42.0 at /home/perigrin/.local/bin/perl:
//
//	$ perl -MO=Deparse -e 'sub f { return 1, 2 } sub g { return ($a,$b) }
//	                       sub h { return wantarray ? @r : $r[0] }'
//	sub f { return 1, 2; }
//	sub g { return $a, $b; }
//	sub h { return wantarray ? @r : $r[0]; }
//
// The assertion is on the CANON because the Unknown count is zero for every
// spelling below and always was. This is the twelfth zero-Unknown wrong tree
// this milestone has found and the count cannot see any of them.
func TestReturnList(t *testing.T) {
	for _, tc := range []struct{ src, want string }{
		// The two shapes the issue measured. An UNPARENTHESISED list is a
		// Binary comma and emits ` , ` from the operator rule; a
		// PARENTHESISED one is a List and emits `, ` from
		// emitCommaSeparated. Two spellings because they are two node kinds,
		// and canon writes what the tree holds.
		{"return 1, 2;", "return 1 , 2;"},
		{"sub f { return ($a, $b); }", "sub f {return ($a, $b);}"},

		// The same loss with the `return` inside a sub body, which is where
		// the corpus meets it.
		{"sub f { return 1, 2; }", "sub f {return 1 , 2;}"},

		// A ternary operand: no comma, same deletion. `return ?:;` was the
		// canon at 34737c46. `wantarray` emits parenthesised because canon
		// writes every Call's argument boundary explicitly -- see its Call
		// case -- and that is not this issue's to change.
		{"return wantarray ? @r : $r[0];", "return wantarray() ? @r : $r[0];"},

		// An operand that is a call with its own arguments -- the Call's Text
		// is the callee name, so this one canon'd as `return f;` and looked
		// almost right, which is worse.
		{"return f($x, $y);", "return f($x , $y);"},

		// A binary that is not a comma. `return $a + $b;` canon'd as
		// `return +;`.
		{"return $a + $b;", "return $a + $b;"},
	} {
		src := []byte(tc.src)
		got := strings.TrimSpace(parse.Canon(parse.Parse(src), src))
		if got != tc.want {
			t.Errorf("Canon(parse(%q)):\n  got  %q\n  want %q", tc.src, got, tc.want)
		}
	}
}

// TestLoopControlStatementModifier: a modifier on a loop control is ONE
// statement.
//
// Issue 01a0dee8 measured four spellings, each producing a wrong tree with
// zero Unknowns:
//
//	goto HERE if $x;   ->  goto HERE; if $x;
//	last if $x;        ->  last if; $x;
//	next LOOP if $x;   ->  next LOOP; if $x;
//	return 1 if $x;    ->  return 1; if $x;
//
// Three statements out of one, and `if` arriving as a Conditional with no
// children. `last if $x` is worse: `if` becomes the LABEL, because
// parseLoopControl's label test asks only `!statementKeywords[...]` where
// parseGoto asks `!modifiers[...]` as well.
//
// perl parses all of them as one statement -- measured on 5.42.0:
//
//	$ perl -MO=Deparse -e 'my $x=1; L: while(1) { last if $x; next L if $x;
//	      redo unless $x; } HERE: ; goto HERE if !$x++;'
//	L: while (1) { last if $x; next L if $x; redo unless $x; }
//	HERE: ;
//	goto HERE unless $x++;
//
// Two causes, both required: the label test must refuse a modifier word, and
// the statement dispatch's control-flow exit must reach applyModifier the way
// its declaration and expression siblings already do.
func TestLoopControlStatementModifier(t *testing.T) {
	for _, tc := range []struct{ src, want string }{
		// The four the issue measured.
		{"goto HERE if $x;", "goto HERE if $x;"},
		{"last if $x;", "last if $x;"},
		{"next LOOP if $x;", "next LOOP if $x;"},
		{"return 1 if $x;", "return 1 if $x;"},

		// All six modifier words, not just `if`. `while` and `until` on a
		// loop control are rare and legal, and they share the label bug.
		{"last unless $x;", "last unless $x;"},
		{"next while $x;", "next while $x;"},
		{"redo until $x;", "redo until $x;"},
		{"last for @a;", "last for @a;"},
		{"next foreach @a;", "next foreach @a;"},

		// A label AND a modifier: the label is the loop's name and the
		// modifier is the statement's, so both survive.
		{"last LOOP if $x;", "last LOOP if $x;"},
		{"redo LOOP unless $x;", "redo LOOP unless $x;"},

		// goto's expression forms, which take the other branch of parseGoto.
		{"goto &f if $x;", "goto &f if $x;"},
		{"goto $where unless $x;", "goto $where unless $x;"},

		// A bare `return` with a modifier and no operand.
		{"return if $x;", "return if $x;"},

		// A LIST return with a modifier: 01a0e013's canon and 01a0dee8's
		// modifier in one statement, which is why the two fixes ship
		// together.
		{"return 1, 2 if $x;", "return 1 , 2 if $x;"},
	} {
		src := []byte(tc.src)
		got := strings.TrimSpace(parse.Canon(parse.Parse(src), src))
		if got != tc.want {
			t.Errorf("Canon(parse(%q)):\n  got  %q\n  want %q", tc.src, got, tc.want)
		}
	}
}

// TestLoopControlWithoutModifier is the regression guard both fixes need.
//
// A bareword after `last` really is a label, and a modifier check that went
// too wide would turn `next LOOP` into a modifier hunt. Every case below
// passes at 34737c46 and must keep passing.
func TestLoopControlWithoutModifier(t *testing.T) {
	for _, tc := range []struct{ src, want string }{
		{"last;", "last;"},
		{"next;", "next;"},
		{"redo;", "redo;"},
		{"last LOOP;", "last LOOP;"},
		{"next OUTER;", "next OUTER;"},
		{"goto HERE;", "goto HERE;"},
		{"goto &f;", "goto &f;"},
		{"return;", "return;"},
		{"return 1;", "return 1;"},
		// A LoopControl is NOT a block form -- see blockForm's own case for
		// why the classification changed with this fix -- so its statement
		// takes a `;` and no spacer, which is what every other non-block
		// statement gets: `{$x = 1;}` and `{return;}` are one rule.
		{"sub f { return 1 }", "sub f {return 1;}"},
	} {
		src := []byte(tc.src)
		got := strings.TrimSpace(parse.Canon(parse.Parse(src), src))
		if got != tc.want {
			t.Errorf("Canon(parse(%q)):\n  got  %q\n  want %q", tc.src, got, tc.want)
		}
	}
}
