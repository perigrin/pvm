// ABOUTME: Tests that canon adds no paren after an operator, a declarator or `sub` that the source lacks.
// ABOUTME: Each case is a canon-adds-paren-op row of the re-emission queue, reduced and held to the pinned perl's Deparse,-p.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/perltest"
)

// canonParenOpCases are the canon-adds-paren-op rows of faithful_queue.txt
// (issue 01a16602-608d), each reduced to its first difference. Every one
// was the tree read as perl reads it with canon writing a pair the grouping
// does not need; the lvref row was a misgrouping. Measured on perl 5.44.0:
//
//   - op/negate.t:17, op/numconvert.t:265. `- -10` and `- -$nz`: canon
//     wrote `-(-10)` so that `-` `-` could not read as `--`. A blank does
//     that as well, and is the source's tree: `- -$x` is `negate sK` over
//     a negate and `-(-$x)` is `negate sKP`, the pair perl's OPf_PARENS.
//     `! ~$x` is the same case for `!~`, and `- A()` for a filetest.
//   - op/ver.t:225. `10**-8`: a prefix operator on the right of `**` is
//     read as a term's start, so it cannot be split there. Deparse,-p of
//     `2**-1**2` folds to 0.5, `2 ** (-(1 ** 2))`.
//   - op/filetest.t:170. `-X => sub {...}`: canon wrote `-(X)` to keep it
//     from reading as a filetest. toke.c reads `-` and a letter followed
//     by `=>` as a minus on the word: Concise is `const[PV "-X"]`.
//   - op/coresubs.t:119. `$p ? () = $' =~ /x/g : 0`: the middle of `?:` is
//     perly.y's term, which takes an assignment. Concise of `$p ? $a = 1 :
//     0` is a cond_expr over a sassign.
//   - op/gmagic.t:141. `() = sub {...}->()->[3]`: `->` and the subscripts
//     are one left-associative level, so a `->` call is a subscript's base
//     as written. Deparse,-p writes `$o->meth->[0]` bare.
//   - op/lvref.t:42. `\my $n = \$y` is a refassign of `\my $n`: Concise is
//     `refassign` over `srefgen`. The parser read `\(my $n = \$y)`. A
//     declaration that is the operand of an operator tighter than
//     assignment takes no initialiser; `!my $x = 1` is "Can't modify not".
var canonParenOpCases = []string{
	"is(- -10, 10, 'neg');\n",
	"my $nz = 0; my $s = sprintf('%+.f', - -$nz);\n",
	"my $x; my $y = ! ~$x;\n",
	"sub A { 1; } my $y = - A();\n",
	"my $v; ok(abs($v - $]) < 10**-8, 'v');\n",
	"my $z = 2**-1**2;\n",
	"my $w = 2 ** !1;\n",
	"my @o = (fallback => 1, -X => sub { 1; });\n",
	"my ($p, $q); my $n = $p ? () = $q =~ /x/g : 0;\n",
	"my ($p, $a); my $n = $p ? $a = 1 : 0;\n",
	"() = sub { 1; }->()->[3];\n",
	"my $o; my $m = $o->meth->[0];\n",
	"my $o; my $m = $o->meth()->{k};\n",
	"my $f; my $e = $f->()[3];\n",
	"use feature 'refaliasing'; no warnings; my $y; \\my $n = \\$y;\n",
	"use feature 'refaliasing'; no warnings; my $y; \\my ($n) = \\$y;\n",
	"use feature 'refaliasing'; no warnings; our $n; my $y; \\local $n = \\$y;\n",
	"my $s = \\my $x . 1;\n",
}

// TestCanonAddedParensOnOperatorsMatchPerl: each reduced case re-emits
// faithfully with nothing forgiven, and perl reads its canon as it reads
// the source.
func TestCanonAddedParensOnOperatorsMatchPerl(t *testing.T) {
	for _, src := range canonParenOpCases {
		faithfulUnforgiven(t, src)
	}
	t.Run("perl", func(t *testing.T) {
		deparsesAlike(t, perltest.PinnedPerl(t), canonParenOpCases)
	})
}
