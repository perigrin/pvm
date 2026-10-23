// ABOUTME: A hash read through a reference folds into multideref or keeps its rv2hv,
// ABOUTME: and the shapes whose answer the text cannot give hedge rather than guess.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parseoracle"
)

// hashSites counts the decided and hedged hash sites Sites reports for src.
func hashSites(src string) (decided, hedged int) {
	for _, k := range decidedKinds(src) {
		if k == parseoracle.SiteKindHash {
			decided++
		}
	}
	for _, k := range hedgedKinds(src) {
		if k == parseoracle.SiteKindHash {
			hedged++
		}
	}
	return decided, hedged
}

// arrowCase is one measured snippet: perl 5.44.0's rv2hv count for it,
// which Sites must decide exactly and without a hash hedge.
type arrowCase struct {
	src   string
	rv2hv int
}

func checkDecided(t *testing.T, cases []arrowCase) {
	t.Helper()
	for _, c := range cases {
		decided, hedged := hashSites(c.src)
		if decided != c.rv2hv || hedged != 0 {
			t.Errorf("%s\n  perl 5.44.0 builds %d rv2hv; Sites decides %d and hedges %d",
				c.src, c.rv2hv, decided, hedged)
		}
	}
}

// TestArrowHashReadFollowsMultideref: perl's peephole optimiser folds a
// chain of subscripts into one multideref op, which builds no rv2hv, and
// three things stop it. Measured with perl 5.44.0, `-MO=Concise,-exec`,
// counting rv2hv and multideref ops:
//
//	my $h; my $x = $h->{a};              rv2hv=0 multideref=1
//	my $h; my $x = $h->{a}{b}{c};        rv2hv=0 multideref=1
//	my $h; my $x = $h->{a}->[0];         rv2hv=0 multideref=1
//	my $h; my $x = $$h{a};               rv2hv=0 multideref=1
//	my $h; my $x = ${$h}{a};             rv2hv=0 multideref=1
//	my $x; $x->{a} = 1;                  rv2hv=0 multideref=1
//	my $h; $h->{a}++;                    rv2hv=0 multideref=1
//	my $h; my $x = $h->{"a$b"};          rv2hv=1 multideref=0
//	my $r; my $x = $r->{$_[1]};          rv2hv=1 multideref=0
//	my $r; my $x = $r->{f()}{a}{g()};    rv2hv=1 multideref=1
//	my $r; my $x = $r->[0]{f()};         rv2hv=0 multideref=1
//	my $x = $_[0]->{$_[1]};              rv2hv=0 multideref=1
//	my $r; my @x = @{$r->{a}};           rv2hv=0 multideref=1
//	my $r; push @{$r->{a}}, 1;           rv2hv=1 multideref=0
//	my $r; push @{$r->{a}{b}{c}}, 1;     rv2hv=3 multideref=0
//	my $r; push @{$r->{a}{"x$r"}}, 1;    rv2hv=0 multideref=1
//	my $r; push @{$r->{"x$r"}{a}}, 1;    rv2hv=2 multideref=0
//	my $r; ${$r->{a}}++;                 rv2hv=1 multideref=0
//	my $r; ${$r->{a}{b}}++;              rv2hv=0 multideref=1
//	${$_[0]->{a}}++;                     rv2hv=0 multideref=1
//	push @{$_[0]->{a}}, 1;               rv2hv=1 multideref=0
//
// First, a key multideref cannot hold -- not a constant or a plain scalar,
// and not something that constant-folds -- ends the run of levels it
// folds. That level keeps its rv2hv unless an earlier level of the run
// folded, whose multideref then hands over the hash itself (`->{"a"}->`).
// Second, a dereference perl's op_lvalue modifies sets OPpDEREF on the
// chain's last element. Under `${...}` multideref folds the levels before
// the last and leaves the last out, so the last keeps its rv2hv only when
// nothing before it folded. Under `@{...}` and `%{...}` the whole run that
// holds the last element stays unfolded, and every brace level of it
// keeps its rv2hv. Third, which contexts modify is perl's op_lvalue, and
// the contexts measured below are the ones Sites decides.
func TestArrowHashReadFollowsMultideref(t *testing.T) {
	checkDecided(t, []arrowCase{
		// Chains read in place.
		{`my $h; my $x = $h->{a};`, 0},
		{`my $h; my $x = $h->{a}{b}{c};`, 0},
		{`my $h; my $x = $h->{a}->[0];`, 0},
		{`my $h; my $x = $$h{a};`, 0},
		{`my $h; my $x = ${$h}{a};`, 0},
		{`my $x; $x->{a} = 1;`, 0},
		{`my $h; $h->{a}++;`, 0},
		{`my $h; my $x = $h->{-bareword};`, 0},
		// Computed keys.
		{`my $r; my $x = $r->{$_[1]};`, 1},
		{`my $r; my $x = $r->[0]{2*$_[0]};`, 0},
		{`my $x = $_[0]->{$_[1]};`, 0},
		{`my $r; my $exp; $r->{t}{$exp =~ /^n/ ? 't' : 'T'}++;`, 0},
		{`my $r; my $x = $r->{"x$r"}{a}{"y$r"};`, 1},
		{`my $r; my $x = $r->{"x$r"}{"y$r"};`, 2},
		// Dereferences that only read: the chain folds.
		{`my $r; my @x = @{$r->{a}};`, 0},
		{`my $r; my $x = ${$r->{a}};`, 0},
		{`my $r; my $n = @{$r->{a}};`, 0},
		{`my $r; print @{$r->{a}};`, 0},
		{`my $r; my $x = join ",", @{$r->{a}{b}};`, 0},
		{`my $r; my @x = reverse @{$r->{a}};`, 0},
		{`my $r; my $n = scalar @{$r->{a}};`, 0},
		{`my $r; if (@{$r->{a}}) {}`, 0},
		{`my $r; print 1 if @{$r->{a}};`, 0},
		{`my $r; while (@{$r->{a}}) {}`, 0},
		{`my $r; my $x = 2 * @{$r->{a}} + @{$r->{b}};`, 0},
		{`my $r; my $x = 1 - @{$r->{a}} / 2;`, 0},
		{`my $r; my $x = 1 == @{$r->{a}};`, 0},
		{`my $r; my $x = 1 < @{$r->{a}};`, 0},
		{`my $r; my $x = @{$r->{a}} || 1;`, 0},
		{`my $r; my $x = !@{$r->{a}};`, 0},
		{`my $r; my $x; $x += @{$r->{a}};`, 0},
		{`my $r; my @x = (1, @{$r->{a}});`, 0},
		{`my $r; my ($x) = @{$r->{a}};`, 0},
		{`sub f { my $r; return @{$r->{a}} }`, 0},
		{`my $r; my @x = $r->{a}->@*;`, 0},
		{`my $r; my $x = $r->{a}->$*;`, 0},
		// Dereferences perl modifies: the last run stays unfolded.
		{`my $r; push @{$r->{a}}, 1;`, 1},
		{`my $r; push @{$r->{a}{b}{c}}, 1;`, 3},
		{`my $r; push @{$r->{a}[0]}, 1;`, 1},
		{`my $r; push @{$r->[0]{a}}, 1;`, 1},
		{`my $r; push @{$$r{a}}, 1;`, 1},
		{`push @{$_[0]->{a}}, 1;`, 1},
		{`my $r; push @{$r->{a}{"x$r"}}, 1;`, 0},
		{`my $r; push @{$r->{"x$r"}{a}}, 1;`, 2},
		{`my $r; push $r->{a}->@*, 1;`, 1},
		{`my $r; unshift @{$r->{a}}, 1;`, 1},
		{`my $r; splice @{$r->{a}}, 0, 1;`, 1},
		{`my $r; my $x = pop @{$r->{a}};`, 1},
		{`my $r; my $x = shift @{$r->{a}};`, 1},
		{`my $r; chomp(@{$r->{a}}, @{$r->{b}});`, 2},
		{`my $r; my @x = grep { 1 } @{$r->{a}};`, 1},
		{`my $r; my @x = grep 1, @{$r->{a}{b}};`, 2},
		{`my $r; my @x = map { 1 } @{$r->{a}};`, 1},
		{`my $r; foreach (@{$r->{a}}) {}`, 1},
		{`my $r; for my $e (1, @{$r->{a}}) {}`, 1},
		{`my $r; foo(1, @{$r->{a}});`, 1},
		{`my $r; my $o; $o->m(@{$r->{a}});`, 1},
		{`my $r; my $x = sprintf "%s", @{$r->{a}};`, 1},
		{`my $r; my @x = values %{$r->{a}};`, 2},
		{`my $r; my @x = each %{$r->{a}};`, 2},
		{`my $r; keys %{$r->{a}{b}};`, 3},
		{`my $r; my $x = \@{$r->{a}};`, 1},
		{`my $r; local @{$r->{a}};`, 1},
		{`my $r; @{$r->{a}} = (1);`, 1},
		{`my $r; my @x = @{$r->{a}}[0,1];`, 1},
		{`my $r; my @x = @{$r->{a}}{qw(a b)};`, 2},
		{`my $r; my $x = ${$r->{a}}[0];`, 1},
		{`my $r; my $x = ${$r->{a}}{b};`, 1},
		{`my $r; ${$r->{a}}++;`, 1},
		{`my $r; ${$r->{a}} = 1;`, 1},
		{`my $r; my $x = \${$r->{a}};`, 1},
		{`my $r; ${$r->{a}{b}}++;`, 0},
		{`${$_[0]->{a}}++;`, 0},
		{`my $r; ${$r->{a}{"x$r"}}++;`, 0},
		{`my $r; ${$r->{"x$r"}{a}}++;`, 2},
	})
}

// TestArrowHashReadKeyShapesFollowPerl: which keys multideref holds, as
// the parser spells them. Measured on perl 5.44.0, counting rv2hv:
//
//	my $h; my $k; my $y = $h->{-$k};            1  (negate is computed)
//	my $h; my $k; my $y = $h->{-$k}{a};         1
//	my $h; my $k; push @{$h->{a}{-$k}},1;       0
//	my $h; my $y = $h->{-1}{a};                 0
//	my $h; my $k; my $y = $h->{+$k};            0  (unary plus is no op)
//	my $h; my $y = $h->{${k}};                  0
//	my $h; my $y = $h->{${^W}};                 0
//	my $h; my $y = $h->{${k}}{a};               0
//	my $h; my $y = $h->{1+1}{a};                0  (literals fold)
//	my $h; my $y = $h->{"a"."b"}{c};            0
//	my $h; my $y = $h->{2*3-1};                 0
//	my $h; my $y = $h->{-x};                    0  (the string "-x")
//	my $h; my $y = $h->{-e}{a};                 0
//	my $h; push @{$h->{-x}}, 1;                 1
//	my $h; my $y = $h->{lc $h};                 1  (a call on a variable)
//	my $h; my $y = $h->{scalar @$h};            1
func TestArrowHashReadKeyShapesFollowPerl(t *testing.T) {
	checkDecided(t, []arrowCase{
		{`my $h; my $k; my $y = $h->{-$k};`, 1},
		{`my $h; my $k; my $y = $h->{-$k}{a};`, 1},
		{`my $h; my $k; push @{$h->{a}{-$k}},1;`, 0},
		{`my $h; my $y = $h->{-1}{a};`, 0},
		{`my $h; my $k; my $y = $h->{+$k};`, 0},
		{`my $h; my $y = $h->{${k}};`, 0},
		{`my $h; my $y = $h->{${^W}};`, 0},
		{`my $h; my $y = $h->{${k}}{a};`, 0},
		{`my $h; my $y = $h->{1+1}{a};`, 0},
		{`my $h; my $y = $h->{"a"."b"}{c};`, 0},
		{`my $h; my $y = $h->{2*3-1};`, 0},
		{`my $h; my $y = $h->{-x};`, 0},
		{`my $h; my $y = $h->{-e}{a};`, 0},
		{`my $h; push @{$h->{-x}}, 1;`, 1},
		{`my $h; my $y = $h->{lc $h};`, 1},
		{`my $h; my $y = $h->{scalar @$h};`, 1},
	})
}

// TestArrowHashReadForLoopHeads: a statement-modifier foreach modifies its
// list as a foreach does; a C-style for's three parts are not a foreach
// list. Measured on perl 5.44.0, counting rv2hv:
//
//	my $h; print for @{$h->{a}{b}};             2
//	my $h; print foreach @{$h->{a}};            1
//	my $h; for (;@{$h->{a}{b}};) {}             0
//	my $h; for (my $i;@{$h->{a}{b}};){}         0
//	my $h; for (;;@{$h->{a}{b}}){}              0
//	my $h; for (@{$h->{a}{b}};;){}              0
//	my $h; for (;%{$h->{a}};){}                 1  (the `%{}` alone)
//	my $h; for (;$h->{a}->@*;){}                0
//
// A C-style head is not a measured context, so its chains hedge: they
// must not decide the foreach count.
func TestArrowHashReadForLoopHeads(t *testing.T) {
	checkDecided(t, []arrowCase{
		{`my $h; print for @{$h->{a}{b}};`, 2},
		{`my $h; print foreach @{$h->{a}};`, 1},
	})
	for _, c := range []arrowCase{
		{`my $h; for (;@{$h->{a}{b}};) {}`, 0},
		{`my $h; for (my $i;@{$h->{a}{b}};){}`, 0},
		{`my $h; for (;;@{$h->{a}{b}}){}`, 0},
		{`my $h; for (@{$h->{a}{b}};;){}`, 0},
		{`my $h; for (;%{$h->{a}};){}`, 1},
		{`my $h; for (;$h->{a}->@*;){}`, 0},
	} {
		if decided, hedged := hashSites(c.src); decided != c.rv2hv || hedged != 1 {
			t.Errorf("%s\n  perl 5.44.0 builds %d rv2hv; Sites should decide %d and hedge "+
				"the chain once, and decides %d and hedges %d", c.src, c.rv2hv, c.rv2hv, decided, hedged)
		}
	}
}

// TestArrowHashReadSubArgElementFoldsToMultideref: a sub's arguments are
// modified, but an element argument is not dereferenced, so nothing sets
// OPpDEREF and the chain folds. Only a whole dereference keeps the chain
// unfolded. Measured on perl 5.44.0:
//
//	my $r; foo($r->{a});         rv2hv=0 multideref=1
//	my $r; foo(%{$r->{a}});      rv2hv=2 multideref=0
//	my $r; foo(@{$r->{a}{b}});   rv2hv=2 multideref=0
func TestArrowHashReadSubArgElementFoldsToMultideref(t *testing.T) {
	checkDecided(t, []arrowCase{
		{`my $r; foo($r->{a});`, 0},
		{`my $r; foo($r->{a}{b}, $r->{c});`, 0},
		{`my $r; foo(%{$r->{a}});`, 2},
		{`my $r; foo(@{$r->{a}{b}});`, 2},
	})
}

// TestArrowHashReadCountFollowsOuterDerefContext: the same chain under the
// same `%{...}` or `@{...}` builds a different number of rv2hv by what is
// done with the dereference. Measured on perl 5.44.0:
//
//	my $r; my %x = %{$r->{a}};      rv2hv=1 multideref=1
//	my $r; keys %{$r->{a}};         rv2hv=2 multideref=0
//	my $r; push @{$r->{a}}, 1;      rv2hv=1 multideref=0
//	my $r; my @x = @{$r->{a}};      rv2hv=0 multideref=1
//
// `%x` is a named hash, whose rv2hv turns on its declaration and which
// hedges for that reason alone; the chain decides its own count.
func TestArrowHashReadCountFollowsOuterDerefContext(t *testing.T) {
	if decided, hedged := hashSites(`my $r; my %x = %{$r->{a}};`); decided != 1 || hedged != 1 {
		t.Errorf("my %%x = %%{$r->{a}}: perl 5.44.0 builds 1 rv2hv; Sites decides %d "+
			"and hedges %d, where only the named `%%x` should hedge", decided, hedged)
	}
	checkDecided(t, []arrowCase{
		{`my $r; keys %{$r->{a}};`, 2},
		{`my $r; push @{$r->{a}}, 1;`, 1},
		{`my $r; my @x = @{$r->{a}};`, 0},
	})
}

// TestArrowHashReadComputedKeyKeepsRv2hv: multideref takes a constant or a
// plain scalar as a key and nothing else. A key holding a variable cannot
// constant-fold, so the level keeps its rv2hv. Measured on perl 5.44.0:
//
//	my $h; my $b; my $x = $h->{"a$b"};    rv2hv=1 multideref=0
//	my $r; my $x = $r->{shift()};         rv2hv=1 multideref=0
//	my $r; exists $r->{shift()};          rv2hv=1 multideref=0
//	my $r; delete $r->{shift()};          rv2hv=1 multideref=0
//	my $r; $r->{shift()} = 1;             rv2hv=1 multideref=0
//	my $r; my $x = $r->{$x->{y}};         rv2hv=1 multideref=1
//
// op/ref.t:343, `$ref->{shift()}`, is this shape.
func TestArrowHashReadComputedKeyKeepsRv2hv(t *testing.T) {
	checkDecided(t, []arrowCase{
		{`my $h; my $b; my $x = $h->{"a$b"};`, 1},
		{`my $r; my $x = $r->{shift()};`, 1},
		{`my $r; exists $r->{shift()};`, 1},
		{`my $r; delete $r->{shift()};`, 1},
		{`my $r; $r->{shift()} = 1;`, 1},
		{`my $r; my $x = $r->{$x->{y}};`, 1},
	})
}

// TestArrowHashReadLvalueElementContextsFold: an element in a context that
// looks modifying is still not dereferenced, so the chain folds. Measured
// on perl 5.44.0, each one multideref and no rv2hv:
//
//	my $r; local $r->{a} = 1;
//	my $r; my $x = \$r->{a};
//	my $r; for ($r->{a}) {}
//	my $r; exists $r->{a}{b};
//	my $r; delete $r->{a}{b};
func TestArrowHashReadLvalueElementContextsFold(t *testing.T) {
	checkDecided(t, []arrowCase{
		{`my $r; local $r->{a} = 1;`, 0},
		{`my $r; my $x = \$r->{a};`, 0},
		{`my $r; for ($r->{a}) {}`, 0},
		{`my $r; exists $r->{a}{b};`, 0},
		{`my $r; delete $r->{a}{b};`, 0},
	})
}

// TestArrowHashReadArrayLastIndexKeepsRv2hv: `$#{$r->{a}}` keeps its rv2hv
// where the plain read `@{$r->{a}}` folds. Measured on perl 5.44.0:
//
//	my $r; my $x = $#{$r->{a}};     rv2hv=1 multideref=0
//	my $r; my @x = @{$r->{a}};      rv2hv=0 multideref=1
//
// The lexer reads `$#{ ... }` as one Variable token (hidesStructure,
// 01a0ad52), so no node holds the chain and Sites cannot count its levels.
// It hedges the hash site, which scores wider and never says "none".
func TestArrowHashReadArrayLastIndexKeepsRv2hv(t *testing.T) {
	for _, src := range []string{
		`my $r; my $x = $#{$r->{a}};`,
		`my $self; my $x = $self->{values}[$#{ $self->{values} }];`,
	} {
		if _, hedged := hashSites(src); hedged == 0 {
			t.Errorf("%s\n  perl keeps an rv2hv here, and Sites does not hedge it", src)
		}
	}
	if decided, hedged := hashSites(`my $r; my @x = @{$r->{a}};`); decided != 0 || hedged != 0 {
		t.Errorf("the plain read folds, yet Sites decides %d and hedges %d", decided, hedged)
	}
}

// TestArrowHashReadHedgesNameTheirShape lists every shape the arrow-read
// hedges covered, each with what Sites does now and why.
func TestArrowHashReadHedgesNameTheirShape(t *testing.T) {
	// Decided: the hedge is gone, and the measurement on perl 5.44.0 that
	// settled it is quoted beside each.
	checkDecided(t, []arrowCase{
		// japh/abigail.t:191, chomp modifies: rv2hv=1.
		{`my $p; chomp @{$p -> {SKIP_OS}};`, 1},
		// japh/abigail.t:192, a condition reads: rv2hv=0.
		{`my $p; if (@{$p -> {SKIP_OS}}) {1}`, 0},
		// op/gmagic.t:233, push modifies: rv2hv=1.
		{`my $self; push @{ $self->{values} }, 1;`, 1},
		// A hash deref in list assignment reads: rv2hv=1, the `%{}` alone.
		{`my $p; my @h = %{$p->{S}};`, 1},
		// `${...}` assigned: rv2hv=1.
		{`my $p; ${$p->{S}} = 1;`, 1},
		// An element of an array deref: rv2hv=1.
		{`my $p; my $x = ${$p->{S}}[0];`, 1},
		// Slices of a deref: rv2hv=1 for `[...]`, 2 for `{...}`.
		{`my $p; my @x = @{$p->{S}}[0,1];`, 1},
		{`my $p; my @x = @{$p->{a}}{qw(a b)};`, 2},
		// Postfix deref pushed: rv2hv=1.
		{`my $p; push $p->{S}->@*, 1;`, 1},
		// porting/libperl.t:212's shape, pushed: rv2hv=2.
		{`my $s; my $sym; push @{$s->{xref}{$sym}}, 1;`, 2},
		// op/ref.t:343, a computed key: rv2hv=1.
		{`my $ref; $ref->{shift()};`, 1},
		// Interpolating and element keys: rv2hv=1.
		{`my $p; my $x = $p->{"a$p"};`, 1},
		{`my $p; my @x; my $x = $p->{$x[0]};`, 1},
		// A `?:` key after keys that fold: rv2hv=1 for the `?:` level,
		// whose run starts after the element key.
		{`my $p; my $sym; my $exp; $p->{text}{$sym}{$p->{o}}{$exp =~ /^non/ ? 't' : 'T'}++;`, 1},
	})

	// Hedged: what each hedge covers and why the text cannot settle it.
	for _, c := range []struct{ src, why string }{
		{`my $p; my $x = $p->{f()}{g()};`,
			"a call key constant-folds when `sub f(){1}` is in scope: `$r->{f()}` builds rv2hv=1, " +
				"but rv2hv=0 after `sub f(){1}`"},
		{`my $p; my $x = $$p{f()};`,
			"a call key constant-folds when the sub is a constant"},
		{`my $p; my @x = keys $p->{S}->%*;`,
			"keys of a postfix `->%*` (rv2hv=2) is not among the contexts Sites decides"},
		{`my $p; my @x = sort @{$p->{S}};`,
			"sort with a block modifies (rv2hv=1) and without one reads (rv2hv=0)"},
		{`my $p; @{$p->{S}};`,
			"void context (rv2hv=0) is not among the contexts Sites decides"},
		{`my $p; my $x = $p->{S}->$#*;`,
			"`->$#*` reads (rv2hv=0) where `$#{...}` keeps it (rv2hv=1); only the read is measured"},
		{`my $p; for (;@{$p->{S}};) {}`,
			"a C-style for's head is not a measured context (rv2hv=0 in the condition)"},
		{`my $p; my $x = $#{$p->{S}};`,
			"the lexer reads `$#{...}` as one token, so no node holds the chain (01a0ad52)"},
	} {
		if _, hedged := hashSites(c.src); hedged == 0 {
			t.Errorf("%s\n  should hedge its hash site: %s", c.src, c.why)
		}
	}
}

// TestArrowHashReadStaysDecided is the regression guard: a chain of simple
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
