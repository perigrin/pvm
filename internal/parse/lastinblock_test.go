// ABOUTME: A statement's last semicolon is optional before `}`, and two parsers hunted past it for an operand.
// ABOUTME: Measured in perl.git t/ as 10 files / 77 nodes for use/no/require and 4 files / 61 nodes for bare return.
package parse_test

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestLastInBlockWithoutSemicolon holds the two constructs that a backward
// walk from each bare-closer Unknown to its MATCHING opener named as the
// runners-up once the trailing comma was fixed: a `use`/`no`/`require` and a
// bare `return`, each last in a block with the final `;` left off.
//
// perl makes that semicolon optional -- `perl -e 'sub f { return } eval {
// require Errno }; print "ok\n"'` prints ok -- and perl's own suite leans on
// it: `eval { require Errno }` as a feature probe is why the use/require half
// clusters in the platform-dependent tests.
//
// Both parsers asked the SAME WRONG QUESTION. `parseUse`'s import list and
// `parseReturn`'s operand each read "the next token is not a Semicolon, so
// there is an argument here", and at the end of a block the next token is `}`.
// The operand hunt then consumed the closer.
//
// THE COUNT IS THE CHEAP HALF. The closer is gone from the enclosing block,
// which reads the next token as its own, and canon then emits a spurious `};`:
//
//	{ require Errno }   canon'd as `{require Errno }; }` before this fix
//	sub f { return }    canon'd as `sub f {return ; }`
//
// So every case asserts the canonical text, not only that the count is zero --
// a wrong tree that scores zero is the worse bug, and `{ no warnings 'all' }`
// below is exactly that shape: it ALREADY scored zero, because a non-empty
// import list gives parseExpr a real operand to stop the hunt.
func TestLastInBlockWithoutSemicolon(t *testing.T) {
	cases := []struct {
		name  string
		src   string
		canon string
	}{
		// The use/no/require half: 10 files, 77 nodes. The bareword-name
		// branch takes the module, then the import list runs into the `}`.
		{"require bareword", "{ require Errno }", "{require Errno; }"},
		{"use bareword", "{ use Errno }", "{use Errno; }"},
		{"no bareword", "{ no warnings }", "{no warnings; }"},
		{"use pragma", "{ use strict }", "{use strict; }"},

		// The idiom the perl suite actually writes, and the reason this
		// clusters in the platform-dependent tests.
		{"eval require probe", "eval { require Errno };", "eval {require Errno; };"},

		// The bare-return half: 4 files, 61 nodes. `op/attrs.t` and
		// `uni/attrs.t` are the same file in two encodings and both OPEN with
		// `sub A::MODIFY_SCALAR_ATTRIBUTES { return }`, so this is the first
		// failure in each.
		// The TRAILING SPACE these three used to expect was the BLOCK-FORM
		// spacer, and a LoopControl is not a block form. It was listed as one
		// in blockForm only because canon's LoopControl case wrote its own
		// `;`, and writing one there put a `;` in the middle of `last if $x`
		// -- issue 01a0dee8. With the terminator moved back to the enclosing
		// Statement, a `return` statement is spelled like every other
		// non-block statement: `{$x = 1;}` and `{return;}` are one rule, and
		// perl accepts both -- measured on 5.42.0,
		// `perl -e 'sub f {return;} sub g {return 1;}'` is syntax OK.
		{"bare return", "sub f { return }", "sub f {return;}"},
		{"bare return in bare block", "{ return }", "{return;}"},
		{"attrs idiom", "sub A::MODIFY_SCALAR_ATTRIBUTES { return }", "sub A::MODIFY_SCALAR_ATTRIBUTES {return;}"},

		// A COMMENT between the keyword and the closer is the same construct,
		// and it is how `op/closure.t:773` actually spells it -- `return # `
		// with the `}` on the next line. Found because that file moved and a
		// single-line grep for the construct could not see it, which is the
		// reason the spelling is pinned here rather than assumed equivalent:
		// peekSignificant skips the comment, so the token the guard examines
		// is the closer either way.
		{"bare return then comment", "sub f { return # c\n}", "sub f {return;}"},

		// WITH the semicolon, which is what already worked. Keeping these
		// beside the new cases is what proves the fix did not simply move the
		// damage onto the spelling that was correct.
		{"require with semicolon", "{ require Errno; }", "{require Errno; }"},
		{"return with semicolon", "sub f { return; }", "sub f {return;}"},

		// An operand PRESENT must still be read. These are the cases where
		// asking `!= Semicolon` happened to give the right answer, and a fix
		// that stops at any non-semicolon would break them.
		//
		// `return LIST` was deliberately NOT here, because it dropped its
		// operands -- `return 1, 2;` canon'd as `return ,;` -- at Unknown = 0
		// and did so WITH the semicolon as well as without, a separate cause
		// from this issue's. It was filed as 01a0e013 and fixed there, and the
		// cause turned out to be neither half of what the filing guessed:
		// parseReturn's `parseExpr(bpListOp)` was right all along -- the comma
		// is BP 80 and bpListOp is 70, so it binds -- and CANON's LoopControl
		// case was writing `c.Text` for every child, which spells an
		// expression as its bare punctuation. TestReturnList holds it now.
		{"return with argument", "sub f { return 1 }", "sub f {return 1;}"},
		{"import list present", "{ no warnings 'all' }", "{no warnings 'all'; }"},
		{"import list qw", "{ use POSIX qw(floor) }", "{use POSIX qw(floor); }"},

		// A `;` inside the block before the closer is the ordinary shape and
		// must be untouched by a change to what follows the last statement.
		// Canon writes no space after an interior `;`, which is its house
		// style and not this issue's to change.
		{"statement then use", "{ $x = 1; use strict }", "{$x = 1;use strict; }"},
		{"statement then return", "sub f { $x = 1; return }", "sub f {$x = 1;return;}"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			src := []byte(c.src)
			n := parse.Parse(src)
			if got := countUnknown(n); got != 0 {
				t.Errorf("Parse(%q) has %d Unknown, want 0", c.src, got)
			}
			got := strings.TrimSpace(parse.Canon(n, src))
			if got != c.canon {
				t.Errorf("Canon(%q):\n  got  %s\n  want %s", c.src, got, c.canon)
			}
		})
	}
}

// TestRequireExpressionPathUnaffected is the regression guard the issue asks
// for: `require "foo.pl"` and `require $x` never reach the bareword-name
// branch, so no change to it may touch them. Both pass before the fix as well
// as after -- which is the point.
func TestRequireExpressionPathUnaffected(t *testing.T) {
	cases := []struct {
		name  string
		src   string
		canon string
	}{
		{"require literal", `{ require "foo.pl" }`, `{require "foo.pl"; }`},
		{"require scalar", "{ require $x }", "{require $x; }"},
		{"require literal semicolon", `require "foo.pl";`, `require "foo.pl";`},
		// A version bundle is ONE Quote token, so it misses the Word-shaped
		// name branch and reads as an expression -- the same path as the two
		// above, and likewise already correct.
		{"use version bundle", "{ use v5.36 }", "{use v5.36; }"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			src := []byte(c.src)
			n := parse.Parse(src)
			if got := countUnknown(n); got != 0 {
				t.Errorf("Parse(%q) has %d Unknown, want 0", c.src, got)
			}
			got := strings.TrimSpace(parse.Canon(n, src))
			if got != c.canon {
				t.Errorf("Canon(%q):\n  got  %s\n  want %s", c.src, got, c.canon)
			}
		})
	}
}

// TestBareLoopControlLastInBlock is the other regression guard: `last`, `next`
// and `redo` bare at the end of a block already stop at the closer, because
// parseLoopControl asks a POSITIVE question -- is the next token a Word that
// could be a label -- rather than the negative `!= Semicolon` that `return`
// and `use` asked. Widening a shared terminator set must not disturb them.
//
// The canon here lost its trailing space when 01a0dee8 moved the `;` out of
// canon's LoopControl case and back to the enclosing Statement -- a
// LoopControl is not a block form, and the space was the block-form spacer.
// The property this guard exists for is unchanged: a bare loop control stops
// at the closer rather than consuming it.
func TestBareLoopControlLastInBlock(t *testing.T) {
	cases := []struct {
		name  string
		src   string
		canon string
	}{
		{"last", "{ last }", "{last;}"},
		{"next", "{ next }", "{next;}"},
		{"redo", "{ redo }", "{redo;}"},
		{"goto", "{ goto }", "{goto;}"},
		{"last with label", "L: while (1) { last L }", "L: while (1) {last L;}"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			src := []byte(c.src)
			n := parse.Parse(src)
			if got := countUnknown(n); got != 0 {
				t.Errorf("Parse(%q) has %d Unknown, want 0", c.src, got)
			}
			got := strings.TrimSpace(parse.Canon(n, src))
			if got != c.canon {
				t.Errorf("Canon(%q):\n  got  %s\n  want %s", c.src, got, c.canon)
			}
		})
	}
}
