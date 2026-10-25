// ABOUTME: Tests that canon adds no paren after a call name, a term or a bareword that the source lacks.
// ABOUTME: Each case is a canon-adds-paren-term row of the re-emission queue, reduced and held to the pinned perl's Deparse,-p.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/perltest"
)

// canonParenTermCases are the canon-adds-paren-term rows of
// faithful_queue.txt (issue 01a16602-60e9), reduced, with the shapes beside
// them that the same cause reaches. The one row at the issue's start,
// op/postfixderef.t:379, `is $cr->()->{ppp}, "pp"`, was canon writing
// `($cr->())->{ppp}`: an Index wrote its base at an atom's power, so a `->`
// call there took a pair. Measured on perl 5.44.0, Deparse,-p writes it
// `is(&$cr()->{'ppp'}, 'pp')`, the call the subscript's base with no pair.
// 01a16602-608d's subscript-base rule fixed it, with op/gmagic.t's
// `sub {...}->()->[3]`.
//
// The bareword cases are a call's arguments perl reads as a filehandle or a
// class name, not a call: Deparse,-p writes `print(STDERR 'x')`, `open(FH,
// '<', $f)`, `close(FH)` and `bless({}, 'Foo')`. Canon gives none of them
// call parens.
var canonParenTermCases = []string{
	"sub is { 1; } my $cr = sub { +{ppp => 'pp'}; }; is $cr->()->{ppp}, 'pp';\n",
	"sub is { 1; } my $cr = sub { +{rrr => 7}; }; is $cr->()->{'rrr'} - 2, 5;\n",
	"my $ar = [{ppp => 1}]; my $v = $ar->[0]->{ppp};\n",
	"my $o; my $v = $o->meth->{k}{j};\n",
	"my $f; my @v = @{$f->()};\n",
	"print STDERR 'x';\n",
	"my $f; open FH, '<', $f;\n",
	"close FH;\n",
	"my $fh; print {$fh;} 'x';\n",
	"my $r = bless {}, 'Foo';\n",
}

// TestCanonAddedParensOnTermsMatchPerl: each reduced case re-emits
// faithfully with nothing forgiven, and perl reads its canon as it reads
// the source.
func TestCanonAddedParensOnTermsMatchPerl(t *testing.T) {
	for _, src := range canonParenTermCases {
		faithfulUnforgiven(t, src)
	}
	t.Run("perl", func(t *testing.T) {
		deparsesAlike(t, perltest.PinnedPerl(t), canonParenTermCases)
	})
}

// TestCanonAddedParenOnBarewordMatchesPerl: a bareword or filehandle perl
// reads as a call's argument is written bare, not given call parens and
// then forgiven, and perl reads the canon as the source.
func TestCanonAddedParenOnBarewordMatchesPerl(t *testing.T) {
	barewords := []string{
		"print STDERR 'x';\n",
		"my $f; open FH, '<', $f;\n",
		"close FH;\n",
		"my $r = bless {}, 'Foo';\n",
		"printf STDOUT '%s', 1;\n",
		"binmode STDOUT, ':raw';\n",
	}
	for _, src := range barewords {
		faithfulUnforgiven(t, src)
	}
	t.Run("perl", func(t *testing.T) {
		deparsesAlike(t, perltest.PinnedPerl(t), barewords)
	})
}
