// ABOUTME: A filetest with no operand defaults to $_, so the token that ends
// ABOUTME: the statement is not its argument.
package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestOperandlessFiletest covers `-e` with nothing after it.
//
// Perl supplies `$_`, verified on 5.42.0:
//
//	perl -MO=Deparse -e '$_ = "/etc"; print -e;'   ->  print -e $_;
//
// term.go's comment said an absent argument "needs no branch" because the node
// keeps its own span and has no child. The claim about perl is right and the
// conclusion was wrong: `parseExpr` does not return nil at a `;`, it returns an
// Unknown spanning it, so the terminator became the argument --
// `print(-e(;));` -- and in the `if` form the whole BLOCK was absorbed into the
// argument list with the parens left unbalanced.
//
// The assertions are on the CANON rather than the count. Eleven defects in this
// parser have scored Unknown=0 with a wrong tree, and here the count was 1 but
// said nothing about WHERE the bytes went.
func TestOperandlessFiletest(t *testing.T) {
	// A filetest with no operand: the statement terminator, closing paren or
	// block brace must NOT be swallowed as its argument. The assertion is
	// the whole emission: canon writes a bare filetest bare, so a swallowed
	// terminator shows as `-e ;;` or `-e ){`, not as a fixed substring.
	for _, tc := range []struct{ src, want string }{
		{"print -e;", "print(-e);"},
		{"my $x = -e;", "my $x = -e;"},
		{"print -d;", "print(-d);"},
		{"if (-e) { print 1 }", "if (-e) {print(1);} "},
		{"my @r = grep { -d } @list;", "my @r = grep({-d;}@list);"},
	} {
		b := []byte(tc.src)
		n := parse.Parse(b)
		if got := parse.Canon(n, b); got != tc.want {
			t.Errorf("Canon(%q) = %q, want %q: the terminator was swallowed", tc.src, got, tc.want)
		}
		if got := countUnknown(n); got != 0 {
			t.Errorf("Parse(%q): Unknown=%d, want 0", tc.src, got)
		}
	}

	// The operandless filetest emits bare, as the source wrote it, which
	// perl reads as the `$_` default -- measured on 5.44.0 with Deparse,-p:
	//
	//	my $x = -e;               (my($x) = (-e $_));
	//	my $x = -t;               (my($x) = -t(STDIN));
	//	my @x = grep -e, @l;      (my(@x) = grep((-e $_), @l));
	//	my @x = grep { -e } @l;   (my(@x) = grep({(-e $_);} @l));
	//
	// The grep forms are issue 01a0de97's: the comma, and the block's `}`,
	// each ended the filetest rather than being taken as its operand.
	//
	// An INFIX operator after the filetest also means it has no operand.
	// Perl supplies `$_` here too:
	//
	//	perl -MO=Deparse -e '$_="/etc"; -e or die;'
	//	  ->  die unless -e $_;
	//	perl -MO=Deparse -e '$_="/etc"; print -e ? 1 : 0;'
	//	  ->  print -e $_ ? 1 : 0;
	//
	// A first version of the fix stopped at `;` and closers and left both
	// of these refusing, so they are asserted rather than assumed. The bare
	// `-e or` is the source's own spelling, so the check is the whole
	// emission: a swallowed `or die` would show as its argument.
	for _, tc := range []struct{ src, want string }{
		{"my $x = -e;", "my $x = -e;"},
		{"my $x = -t;", "my $x = -t;"},
		{"my @x = grep -e, @l;", "my @x = grep(-e , @l);"},
		{"my @x = grep { -e } @l;", "my @x = grep({-e;}@l);"},
		{"-e or die;", "-e or die();"},
		{"print -e ? 1 : 0;", "print(-e ? 1 : 0);"},
	} {
		b := []byte(tc.src)
		n := parse.Parse(b)
		if got := parse.Canon(n, b); got != tc.want {
			t.Errorf("Canon(%q) = %q, want %q", tc.src, got, tc.want)
		}
		if got := countUnknown(n); got != 0 {
			t.Errorf("Parse(%q): Unknown=%d, want 0", tc.src, got)
		}
	}

	// REGRESSION GUARD, passes today: a filetest WITH an operand still takes
	// it, and stacking still nests. Written bare as the source wrote it;
	// perl 5.44.0 Deparse,-p: print((-e $f)), (my($x) = (-e $f)),
	// (my($x) = -t(STDIN)), print((-d (-e $f))).
	for _, tc := range []struct{ src, want string }{
		{"print -e $f;", "print(-e $f);"},
		{"my $x = -e $f;", "my $x = -e $f;"},
		{"my $x = -t STDIN;", "my $x = -t STDIN;"},
		{"print -d -e $f;", "print(-d -e $f);"},
	} {
		b := []byte(tc.src)
		n := parse.Parse(b)
		if got := parse.Canon(n, b); got != tc.want {
			t.Errorf("Canon(%q) = %q, want %q", tc.src, got, tc.want)
		}
	}
}
