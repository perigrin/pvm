// ABOUTME: A hash read through a reference folds into multideref or keeps its rv2hv,
// ABOUTME: and the shapes whose answer the text cannot give hedge rather than guess.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parseoracle"
)

// TestArrowHashReadHedgesRv2hv: perl's optimiser folds `$r->{k}` into one
// multideref op, which emits no rv2hv, but only when it can. Two things stop
// it, and neither is in the text. Measured on 5.42.0 with -MO=Concise
// (live ops only):
//
//	my $p; my @x = @{$p->{S}};        rv2av multideref
//	my $p; push @{$p->{S}}, 1;        rv2av helem rv2hv
//	my $p; chomp @{$p->{S}};          rv2av helem rv2hv
//	my $p; my %h = %{$p->{S}};        rv2hv multideref
//	my $p; keys %{$p->{S}};           rv2hv helem rv2hv
//	my $p; ${$p->{S}} = 1;            helem rv2hv
//	my $p; push $p->{S}->@*, 1;       rv2av helem rv2hv
//	my $p; my $x = $p->{a};           multideref
//	my $p; my $x = $p->{f()};         helem rv2hv
//	sub f(){1} my $p; $p->{f()};      multideref
//	my $p; my $x = $p->{a}{f()};      helem multideref
//	my $p; my $x = $p->{f()}{g()};    helem rv2hv helem rv2hv
//
// A dereferenced element, `@{$p->{S}}`, keeps its rv2hv when the context
// around the dereference modifies or vivifies it -- push, chomp, grep, a
// sub's arguments, keys -- and folds when it only reads. A computed key
// keeps its rv2hv unless it constant-folds, which turns on declarations,
// and unless the element below it folded. japh/abigail.t:191,
// op/gmagic.t:233, op/ref.t:343 and porting/libperl.t:212,307 are these
// shapes.
func TestArrowHashReadHedgesRv2hv(t *testing.T) {
	for _, src := range []string{
		`my $p; chomp @{$p -> {SKIP_OS}};`,
		`my $p; if (@{$p -> {SKIP_OS}}) {1}`,
		`my $self; push @{ $self->{values} }, 1;`,
		`my $p; my %h = %{$p->{S}};`,
		`my $p; ${$p->{S}} = 1;`,
		`my $p; my $x = ${$p->{S}}[0];`,
		`my $p; my @x = @{$p->{S}}[0,1];`,
		`my $p; my @x = @{$p->{a}}{qw(a b)};`,
		`my $p; push $p->{S}->@*, 1;`,
		`my $p; my @x = keys $p->{S}->%*;`,
		`my $s; push @{$s->{xref}{$sym}}, 1;`,
		`my $ref; $ref->{shift()};`,
		`my $p; my $x = $p->{f()}{g()};`,
		`my $p; my $x = $$p{f()};`,
		`my $p; my $x = $p->{"a$p"};`,
		`my $p; my $x = $p->{$x[0]};`,
		`my $p; $p->{text}{$sym}{$p->{o}}{$exp =~ /^non/ ? 't' : 'T'}++;`,
	} {
		if got := hedgedKinds(src); !has(got, parseoracle.SiteKindHash) {
			t.Errorf("%s\n  does not hedge a hash site; whether perl folds this "+
				"read into multideref is not in the text (hedged %v)", src, got)
		}
	}
}

// TestArrowHashReadStaysDecided is the negative half: a chain of simple
// keys read in place folds into multideref every time, measured, so the
// subject keeps saying there is no rv2hv and those statements stay exact.
// A chained slice of an element, `@{$s->{x}{$y}}`, is not the bare
// `@{$::{$k}}` stash deref either, so it decides no hash site of its own.
func TestArrowHashReadStaysDecided(t *testing.T) {
	for _, src := range []string{
		`my $p; my $x = $p->{a};`,
		`our $p; my $x = $p->{a};`,
		`my ($p, $k); my $x = $p->{$k};`,
		`my $p; my $x = $p->{a}{b}{c};`,
		`my $p; my $x = $p->{-bar};`,
		`my $p; my $x = $p->{"abc"};`,
		`my $p; my $x = $p->{'a$b'};`,
		`my $p; my $x = $p->{1}{-1}{1.5};`,
		`my $p; my $x = $p->{$_}{$1}{$::y}{$Foo::x};`,
		`my $p; my $x = $p->{a}[0];`,
		`my $p; my $x = $p->[0]{a};`,
		`my $p; my $x = $$p{a};`,
		`my $p; exists $p->{a};`,
		`my $p; delete $p->{a};`,
		`my $p; $p->{a}{b} = 1;`,
		`my $p; foo($p->{a});`,
		`my $p; my $x = &{$p->{S}};`,
		`my $p; my @x = @$p;`,
	} {
		if got := hedgedKinds(src); has(got, parseoracle.SiteKindHash) {
			t.Errorf("%s\n  perl folds this into multideref every time, yet it hedges %v", src, got)
		}
		if got := decidedKinds(src); has(got, parseoracle.SiteKindHash) {
			t.Errorf("%s\n  perl emits no rv2hv here, yet it decides %v", src, got)
		}
	}
	src := `my $s; my @x = @{$s->{xref}{$sym}};`
	if got := decidedKinds(src); has(got, parseoracle.SiteKindHash) {
		t.Errorf("%s\n  perl emits no rv2hv here, yet it decides %v", src, got)
	}
}
