// ABOUTME: A NAMED hash's rv2hv follows the declaration in effect, a DEREFERENCE always emits one,
// ABOUTME: and a named hash hedges only where its declaration may lie outside what the file shows.

package parse_test

import (
	"context"
	"testing"

	"tamarou.com/pvm/internal/parse"
	"tamarou.com/pvm/internal/parseoracle"
	"tamarou.com/pvm/internal/perltest"
)

// hedgedSite reports whether a site of this kind is present AND unresolved.
func hedgedSite(sites []parseoracle.SubjectCallSite, kind string) bool {
	for _, s := range sites {
		if s.Kind == kind && s.Unresolved {
			return true
		}
	}
	return false
}

// namedHashTable is perl 5.44.0's rv2hv count for each snippet, over the
// WHOLE snippet as parse_facts.pl's `sites` reports it: its B walk counts
// an `our` declaration's own unexecuted `rv2hv vK/OURINTR`, which Concise
// tree mode shows and `-exec` order omits. TestNamedHashRv2hvTableMatchesPerl
// re-measures every row. Read with -MO=Concise on 5.44.0:
//
//	my %h; keys %h                  padhv                       0
//	our %h; keys %h                 rv2hv vK/OURINTR, rv2hv     2
//	local %h; keys %h               rv2hv vKM/LVINTRO, rv2hv    2
//	keys %h                         gv, rv2hv                   1
//	state %h (feature on)           padhv                       0
//	state %h (feature off)          `state % h`, then gv, rv2hv 1
//	our %h; $h{a}{b}                multideref($h{"a"}{"b"})    1, the declaration
//	our %h; delete local $h{a}      rv2hv sKR, delete vK/LVINTRO  2
//	my %h; delete local $h{a}       padhv sR                    0
//	our %h; @h{"a","b"}             rv2hv, hslice               2
//	our %h; push @{$h{a}}, 1        rv2hv, helem, rv2av         2
//
// An element of a package hash folds into multideref as an element through
// a reference does, so its count is chainRv2hv's with the hash counted as
// one more level; a slice is one rv2hv for a package hash and none for a
// lexical.
var namedHashTable = []arrowCase{
	// my, our and local.
	{`my %h; my @k = keys %h;`, 0},
	{`our %h; my @k = keys %h;`, 2},
	{`local %h; my @k = keys %h;`, 2},
	{`my @k = keys %h;`, 1},
	{`our (%h, %g);`, 2},
	{`local (%h, %g);`, 2},
	{`our %h = (a => 1);`, 1},
	{`my (%h, %g); my @k = (keys %h, keys %g);`, 0},
	{`local our %h; my @k = keys %h;`, 2},
	// state, by the feature.
	{`use feature "state"; state %h; my @k = keys %h;`, 0},
	{`sub f { use feature "state"; state %h; my @k = keys %h; }`, 0},
	{`no feature "state"; state %h; my @k = keys %h;`, 1},
	{`use v5.10; state %h; my @k = keys %h;`, 0},
	{`use v5.36; state %h; my @k = keys %h;`, 0},
	// field and signature parameters are a my.
	{`use feature "signatures"; sub f (%o) { my @k = keys %o; }`, 0},
	{`use v5.40; use feature "class"; no warnings; class C { field %h; method m { my @k = keys %h; } }`, 0},
	// keys, values, each, exists, delete.
	{`our %h; my @v = values %h;`, 2},
	{`our %h; my @e = each %h;`, 2},
	{`my %h; my @v = values %h;`, 0},
	{`our %h; my $e = exists $h{a};`, 1},
	{`my %h; my $e = exists $h{a};`, 0},
	{`our %h; my $e = delete $h{a};`, 1},
	{`my %h; my $e = delete $h{a};`, 0},
	{`our %h; delete local $h{a};`, 2},
	{`my %h; delete local $h{a};`, 0},
	// Other uses of a whole hash.
	{`our %h; my %x = %h;`, 2},
	{`our %h; my $r = \%h;`, 2},
	{`our %h; foo(%h);`, 2},
	{`our %h; if (%h) { 1 }`, 2},
	{`my @x = keys %ENV;`, 1},
	{`my @x = keys %Foo::bar;`, 1},
	{`*h = \%x; my @k = keys %h;`, 2},
	{`eval q{my %h}; my @k = keys %h;`, 1},
	// Elements fold into multideref.
	{`our %h; my $v = $h{a};`, 1},
	{`our %h; my $v = $h{a}{b}{c};`, 1},
	{`our %h; $h{a}{b}++;`, 1},
	{`our %h; my $k; my $v = $h{"a$k"};`, 2},
	{`my %h; my $k; my $v = $h{"a$k"};`, 0},
	{`our %h; my $k; my $v = $h{"a$k"}{b};`, 2},
	{`our %h; my $x; my $v = $h{a}{"b$x"};`, 1},
	{`our %h; push @{$h{a}}, 1;`, 2},
	{`my %h; push @{$h{a}}, 1;`, 0},
	{`our %h; my @x = @{$h{a}};`, 1},
	{`my %h; my @x = @{$h{a}};`, 0},
	{`our %h; my @x = keys %{$h{a}};`, 3},
	{`my %h; my @x = keys %{$h{a}};`, 1},
	{`our %h; ${$h{a}}++;`, 2},
	{`our %h; my $x = ${$h{a}}[0];`, 2},
	{`my $x = $ENV{PATH};`, 0},
	// Slices.
	{`our %h; my @v = @h{"a","b"};`, 2},
	{`my %h; my @v = @h{"a","b"};`, 0},
	{`our %h; my %v = %h{"a","b"};`, 2},
	{`my %h; my %v = %h{"a","b"};`, 0},
	{`our %h; my @v = delete @h{"a","b"};`, 2},
	{`our %h; delete local @h{"a","b"};`, 2},
	{`my %h; delete local @h{"a","b"};`, 0},
	{`my %h; our %g; my @v = (@h{"a"}, @g{"b"});`, 2},
	// Scope.
	{`our %h; { my %h; my @k = keys %h; }`, 1},
	{`my %h; { our %h; } my @k = keys %h;`, 1},
	{`my %h; our %h; my @k = keys %h;`, 2},
	{`my @k = keys %h; my %h; @k = keys %h;`, 1},
	{`{ my %h; } my @k = keys %h;`, 1},
	{`sub f { my @k = keys %h; } my %h;`, 1},
	{`my %h; sub f { my @k = keys %h; }`, 0},
	// A BEGIN that runs no other code imports nothing.
	{`BEGIN { $| = 1 } my @k = keys %h;`, 1},
	// delete local through a reference keeps the last level out of
	// multideref, as `${...}` modified does.
	{`my $r; delete local $r->{a};`, 1},
	{`my $r; delete local $r->{a}{b};`, 0},
	// Nested shadowing, packages, a local in a sub, and mixed slices.
	{`our %h; {my %h; {our %h; keys %h} keys %h} keys %h;`, 4},
	{`package A; our %h; package B; keys %h;`, 2},
	{`sub f { local %h } f(); keys %h; my %h; keys %h;`, 2},
	{`our %a; my %b; my @x = (@a{1,2}, @b{1,2}, %a{3}, %b{4});`, 3},
	// Space after a sigil, and a braced name, still name the hash.
	{`my %h; my @k = keys % h;`, 0},
	{`my %h; my @x = @ h{1,2};`, 0},
	{`my %h; my %x = % h{1,2};`, 0},
	{`my %h; my @k = keys %{h};`, 0},
	{`my %h; my @k = keys %{ h };`, 0},
	{`my %h; my $x = ${h}{a};`, 0},
	{`my %h; my $x = $ h{a};`, 0},
	{`my %h; my @x = @{h}{1,2};`, 0},
	{`my %h; my $k; my $x = ${h}{"a$k"};`, 0},
	{`our %h; my $k; my $x = ${h}{"a$k"};`, 2},
	// A literal left operand settles a short circuit: `1 && X` is X.
	{`our %h; my $n = 1 && %h;`, 2},
	{`our %h; my $n = 0 || %h;`, 2},
	{`our %h; my $n = "" or %h;`, 2},
	// A true literal condition keeps what it guards.
	{`our %h; while (1) { my @k = keys %h; last; }`, 2},
	{`our %h; if (1) { my @k = keys %h; }`, 2},
	{`our %h; unless (0) { my @k = keys %h; }`, 2},
	{`our %h; my @k = keys %h if 1;`, 2},
	{`our %h; do { my @k = keys %h } until 1;`, 2},
	{`our %h; do { my @k = keys %h } while 0;`, 2},
	{`our %h; my $x = 1 ? keys %h : 0;`, 2},
	{`our %h; for (;;) { keys %h; last }`, 2},
	{`our %h; for (;1;) { keys %h; last }`, 2},
	{`our %h; my $c; if ($c) { keys %h } elsif ($c) { keys %h } else { keys %h }`, 4},
	// `use parent -norequire` loads nothing.
	{`use parent -norequire, "Lp"; keys %h;`, 1},
	// A use's arguments are resolved as any other code.
	{`use constant N => scalar(%h);`, 1},
}

// hedgedHashTable is perl 5.44.0's rv2hv count for snippets Sites must
// hedge, never deciding more than perl builds. Each is a case where the
// text alone cannot settle the count:
//
//   - a numeric fold over a string, which perl leaves unfolded because it
//     would warn (`"a" << 1`, `1+"a"`);
//   - a literal key after a module's import, which may install
//     overload::constant (`use bigint` makes `1` an object, and a `q`
//     handler does the same to `"a"`);
//   - `length` or `join` where the file may override it;
//   - code under a condition perl may fold away (`if (0)`, `keys %h if 0`,
//     `sub D(){0} if (D)`), and a `?:` whose condition may be constant.
var hedgedHashTable = []arrowCase{
	{`my $r; my $x = $r->{"a" << 1};`, 1},
	{`my $r; my $x = $r->{"a" ** 2};`, 1},
	{`my $r; my $x = $r->{"0x10" << 1};`, 1},
	{`my $r; my $x = $r->{1+"a"};`, 1},
	{`my $r; my $x = $r->{"a"*2};`, 1},
	{`my $r; my $x = $r->{"a"-2};`, 1},
	{`my $r; my $x = $r->{"1"+2};`, 0},
	{`use bigint; my $r; my $x = $r->{1<<2};`, 1},
	{`use bigint; my $r; my $x = $r->{~0};`, 1},
	{`use bigint; my $r; my $x = $r->{2**3};`, 1},
	{`use bigint; my $r; my $x = $r->{1};`, 1},
	{`use bignum; my $r; my $x = $r->{1.5};`, 1},
	{`BEGIN { $^H{integer} = sub { bless [], "X" }; $^H |= 0x1000 } my $r; my $x = $r->{1};`, 1},
	{`use bigint; my $r; my $x = $r->{"a"};`, 0},
	{`use subs "length"; sub length {1} my $r; my $x = $r->{length "abc"};`, 1},
	{`BEGIN { *CORE::GLOBAL::length = sub {1} } my $r; my $x = $r->{length "abc"};`, 1},
	{`BEGIN { *CORE::GLOBAL::join = sub {1} } my $r; my $x = $r->{join ",", 1, 2};`, 1},
	{`our %h; if (0) { keys %h }`, 1},
	{`our %h; keys %h if 0;`, 1},
	{`our %h; unless (1) { keys %h }`, 1},
	{`our %h; 0 and keys %h;`, 1},
	{`our %h; my $x = 1 || keys %h;`, 1},
	{`our %h; while (0) { keys %h }`, 1},
	{`our %h; my $x = 0 ? keys %h : 1;`, 1},
	{`our %h; my $x = (1 == 2) ? keys %h : 0;`, 1},
	{`our %h; my @x = (0 ? %h : ());`, 1},
	{`our %h; if (1) { 1 } else { keys %h }`, 1},
	{`our %h; if ($x) { keys %h } elsif (0) { keys %h }`, 2},
	{`our %h; sub D(){0} if (D) { keys %h }`, 1},
	{`my $r; keys %$r if 0;`, 0},
	{`my $r; push @{ 1 ? $r->{a} : $r->{b} }, 1;`, 1},
	{`my $r; push @{ 0 ? $r->{a} : $r->{b} }, 1;`, 1},
	{`my $r; push @{ "" ? $r->{a} : $r->{b} }, 1;`, 1},
	{`use constant C => 1; my $r; push @{ C ? $r->{a} : $r->{b} }, 1;`, 1},
	{`sub C(){1} my $r; push @{ C ? $r->{a} : $r->{b} }, 1;`, 1},
	// Code run at BEGIN that may declare a lexical %h.
	{`use lib do { builtin::export_lexically("%h" => {}); "." }; my @k = keys %h;`, 0},
	{`sub BEGIN { builtin::export_lexically("%h" => {}) } my @k = keys %h;`, 0},
	{`BEGIN { my $f = \&builtin::export_lexically; &$f("%h" => {}) } my @k = keys %h;`, 0},
	{`BEGIN { my $f = \&builtin::export_lexically; &{$f}("%h" => {}) } my @k = keys %h;`, 0},
	{`sub f { builtin::export_lexically("%h" => {}) } BEGIN { f } my @k = keys %h;`, 0},
	{`sub f { builtin::export_lexically("%h" => {}) } BEGIN { f 1 } my @k = keys %h;`, 0},
	{`sub f { builtin::export_lexically("%h" => {}) } BEGIN { 1 if f } my @k = keys %h;`, 0},
	// A later branch of an elsif chain, and a C-style for's condition.
	{`our %h; my $c; if ($c) { keys %h } elsif (1) { keys %h } else { keys %h }`, 3},
	{`our %h; my $c; if ($c) { keys %h } elsif ("x") { keys %h } else { keys %h }`, 3},
	{`our %h; for (;0;) { keys %h }`, 1},
	{`our %h; for (my $i = 0; 0; $i++) { keys %h }`, 1},
}

// TestNamedHashHedgesNeverOvercount: each hedgedHashTable row hedges, and
// decides no more rv2hv than perl builds.
func TestNamedHashHedgesNeverOvercount(t *testing.T) {
	for _, c := range hedgedHashTable {
		decided, hedged := hashSites(c.src)
		if decided > c.rv2hv || hedged == 0 {
			t.Errorf("%s\n  perl 5.44.0 builds %d rv2hv; Sites decides %d and hedges %d, "+
				"where it should hedge and decide no more than perl", c.src, c.rv2hv, decided, hedged)
		}
	}
}

// TestNamedHashRv2hvFollowsDeclaration: each row is decided, with no hash
// hedge, at perl's count.
func TestNamedHashRv2hvFollowsDeclaration(t *testing.T) {
	checkDecided(t, namedHashTable)
}

// TestNamedHashRv2hvTableMatchesPerl re-measures namedHashTable and
// hedgedHashTable on the
// pinned perl, so a row copied from an older measurement fails.
func TestNamedHashRv2hvTableMatchesPerl(t *testing.T) {
	perl := perltest.PinnedPerl(t)
	for _, c := range append(append([]arrowCase{}, namedHashTable...), hedgedHashTable...) {
		facts, err := parseoracle.Ask(context.Background(), []byte(c.src), parseoracle.Options{Perl: perl.Path})
		if err != nil || !facts.OK {
			t.Errorf("%s\n  perl did not compile it: %v %s", c.src, err, facts.Stderr)
			continue
		}
		if got := len(facts.Sites[parseoracle.MarkerRv2hv]); got != c.rv2hv {
			t.Errorf("%s\n  the table records %d rv2hv; perl 5.44 counts %d", c.src, c.rv2hv, got)
		}
	}
}

// TestNamedHashStateFeatureOffIsPackageHash: under `no feature "state"`,
// perl reads `state %h` as `state % h` and the later `%h` is the package's.
func TestNamedHashStateFeatureOffIsPackageHash(t *testing.T) {
	checkDecided(t, []arrowCase{
		{`no feature "state"; state %h; my @k = keys %h;`, 1},
		{`state %h; my @k = keys %h;`, 1},
	})
}

// TestNamedHashStateViaFeatureBundle: a `use VERSION` from 5.9.5 on turns
// the state feature on, and `state %h` is padhv.
func TestNamedHashStateViaFeatureBundle(t *testing.T) {
	checkDecided(t, []arrowCase{
		{`use v5.10; state %h; my @k = keys %h;`, 0},
		{`use v5.36; state %h; my @k = keys %h;`, 0},
	})
}

// TestNamedHashStateIsPadhv: `state %h` with the feature on is lexical, at
// file scope and in a sub: `padhv[%h:2,4] lRM` + `keys`, no rv2hv.
func TestNamedHashStateIsPadhv(t *testing.T) {
	checkDecided(t, []arrowCase{
		{`use feature "state"; state %h; my @k = keys %h;`, 0},
		{`use feature "state"; sub f { state %h; my @k = keys %h; }`, 0},
	})
}

// TestNamedHashDeleteLocalFollowsScope: `delete local $h{a}` builds
// `rv2hv sKR` under `delete vK/LVINTRO` for a package hash, 2 over the
// whole snippet with the declaration's own, and `padhv sR` for a lexical.
func TestNamedHashDeleteLocalFollowsScope(t *testing.T) {
	checkDecided(t, []arrowCase{
		{`our %h; delete local $h{a};`, 2},
		{`my %h; delete local $h{a};`, 0},
	})
}

// TestNamedHashOurChainFoldsToMultideref: `our %h; my $v = $h{a}{b}` is 1,
// the declaration's; the use folds into multideref($h{"a"}{"b"}).
func TestNamedHashOurChainFoldsToMultideref(t *testing.T) {
	checkDecided(t, []arrowCase{{`our %h; my $v = $h{a}{b};`, 1}})
}

// TestNamedHashInnerMyShadowsOur: the inner `keys %h` is padhv.
func TestNamedHashInnerMyShadowsOur(t *testing.T) {
	checkDecided(t, []arrowCase{{`our %h; { my %h; my @k = keys %h; }`, 1}})
}

// TestNamedHashUseBeforeMyIsPackageHash: the first `keys` is gv + rv2hv,
// the second padhv.
func TestNamedHashUseBeforeMyIsPackageHash(t *testing.T) {
	checkDecided(t, []arrowCase{{`my @k = keys %h; my %h; @k = keys %h;`, 1}})
}

// TestNamedHashClosedBlockMyDoesNotLeak: a `my` ends with its block.
func TestNamedHashClosedBlockMyDoesNotLeak(t *testing.T) {
	checkDecided(t, []arrowCase{{`{ my %h; } my @k = keys %h;`, 1}})
}

// TestNamedHashMixedScopeSliceCountsEach: the `my` slice is 0, the `our`
// slice 1, and the `our` declaration 1.
func TestNamedHashMixedScopeSliceCountsEach(t *testing.T) {
	checkDecided(t, []arrowCase{{`my %h; our %g; my @v = (@h{"a"}, @g{"b"});`, 2}})
}

// TestUndecidableHashHedges: a named hash whose declaration may not be in
// the file hedges. Each was measured on 5.44.0 with a module or file that
// exports %h lexically (builtin::export_lexically), where the same text is
// rv2hv without one:
//
//	use Lx; keys %h                        padhv
//	our %h; use Lx; keys %h                padhv
//	BEGIN { require "./x.pl" } keys %h     padhv
//	new Foo(my %h); keys %h                padhv, a declaration in unread code
//	use if 1, feature => "state"; state %h; keys %h    padhv; rv2hv with `use if 0`
func TestUndecidableHashHedges(t *testing.T) {
	for _, src := range []string{
		`use Lx; my $n = keys %h;`,
		`our %h; use Lx; my $n = keys %h;`,
		`use Lx; my @x = @h{"a","b"};`,
		`use Lx; delete local $h{a};`,
		`BEGIN { require "./x.pl" } my $n = keys %h;`,
		`new Foo(my %h); my $n = keys %h;`,
		`use if 1, feature => "state"; state %h; my $n = keys %h;`,
		`use parent "Lp"; my $n = keys %h;`,
		`use base "Lp"; my $n = keys %h;`,
	} {
		sites := parse.Sites(parse.Parse([]byte(src)), []byte(src))
		if !hedgedSite(sites, parseoracle.SiteKindHash) {
			t.Errorf("%s\n  does not hedge; the declaration %%h resolves to may "+
				"not be in the file", src)
		}
	}
}

// TestDecidableHashStillCommits: an explicit DEREFERENCE always emits the
// op, so the subject commits and keeps earning exact verdicts.
//
// This is the half that must not collapse. Hedging every hash site would
// take WRONG to zero and exact with it, which is the vacuous gate
// TestGoParserExactFloor exists to prevent. Measured -- every one of these
// is 1 or more regardless of how anything was declared:
//
//	keys %$r            1      keys %{$s{a}}      1
//	keys %{$r}          1      keys %{$s{a}{b}}   2
//	@$r{"a","b"}        1      $r->%*             1
//	$r->@{"a","b"}      1      keys %ENV          1
func TestDecidableHashStillCommits(t *testing.T) {
	for _, src := range []string{
		`my $n = keys %$r;`,
		`my $n = keys %{$r};`,
		`my @x = @$r{"a","b"};`,
	} {
		sites := parse.Sites(parse.Parse([]byte(src)), []byte(src))
		if !hasSite(sites, parseoracle.SiteKindHash) {
			t.Errorf("%s\n  reports no decided hash site; a dereference emits "+
				"rv2hv whatever the declaration", src)
		}
		if hedgedSite(sites, parseoracle.SiteKindHash) {
			t.Errorf("%s\n  hedges a site it can decide; that costs an exact "+
				"verdict for nothing", src)
		}
	}

	// A NESTED access is both at once: `%{$s{a}}` dereferences -- always
	// an op -- an element whose own rv2hv follows how %s was declared.
	// Measured on 5.44.0:
	//
	//	my %s;  keys %{$s{a}}      1 rv2hv   (the element folds to multideref)
	//	my %s;  keys %{$s{a}{b}}   2 rv2hv   (the inner element does not fold)
	//	        keys %{$s{a}}      2 rv2hv   (%s is the package's)
	checkDecided(t, []arrowCase{
		{`my %s; my @x = keys %{$s{a}};`, 1},
		{`my %s; my @x = keys %{$s{a}{b}};`, 2},
		{`my @x = keys %{$s{a}};`, 2},
	})
}
