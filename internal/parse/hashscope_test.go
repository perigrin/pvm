// ABOUTME: Holds the named-hash scope resolver to perl 5.44: which declaration each `%h` use sees,
// ABOUTME: and whether the state feature is on there. Each fixture quotes perl's -MO=Concise ops.

package parse

import (
	"slices"
	"sort"
	"testing"
)

// resolvedHashes resolves src's named hashes and writes each in source
// order: a declaring variable as `DECL %h`, a use as `%h=DECL`, with `pkg`
// for a use no declaration reaches. A signature parameter is `sig %o`, and
// `sig! %o` or `sig? %o` where the signatures feature is off or unknown. A `state` declared with its feature off
// declares nothing and is `state! %h`; one declared where the feature is
// unknown is `state? %h`, and a use resolving to it `%h=state?`. withState
// appends ` +s` where the state feature is on and ` ?s` where it is unknown.
func resolvedHashes(src string, withState bool) []string {
	uses, _ := hashScopes(Parse([]byte(src)))
	nodes := make([]*Node, 0, len(uses))
	for n := range uses {
		nodes = append(nodes, n)
	}
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].Start < nodes[j].Start })
	var got []string
	for _, n := range nodes {
		u := uses[n]
		var s string
		switch {
		case u.Declares && u.Decl == "state" && u.State == stateOff:
			s = "state! " + n.Text
		case u.Declares && u.Decl == "state" && u.State == stateUnknown:
			s = "state? " + n.Text
		case u.Declares && u.Decl == "signature":
			s = map[featureState]string{stateOn: "sig ", stateOff: "sig! ", stateUnknown: "sig? "}[u.Signatures] + n.Text
		case u.Declares:
			s = u.Decl + " " + n.Text
		case u.Unsure:
			s = n.Text + "=" + u.Decl + "?"
		case u.Decl == "":
			s = n.Text + "=pkg"
		default:
			s = n.Text + "=" + u.Decl
		}
		if withState && u.State == stateOn {
			s += " +s"
		}
		if withState && u.State == stateUnknown {
			s += " ?s"
		}
		got = append(got, s)
	}
	return got
}

type hashScopeCase struct {
	name, src string
	want      []string
}

func checkHashScopes(t *testing.T, withState bool, cases []hashScopeCase) {
	t.Helper()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := resolvedHashes(tc.src, withState); !slices.Equal(got, tc.want) {
				t.Errorf("%s\n got  %q\n want %q", tc.src, got, tc.want)
			}
		})
	}
}

// TestHashScopeResolvesInnermostDeclaration: a use sees the innermost
// declaration in effect. Measured on perl 5.44.0, -MO=Concise:
//
//	our %h; { my %h; keys %h } keys %h
//	-     <1> rv2hv[t3] vK/OURINTR ->2
//	5           <0> padhv[%h:3,4] vM/LVINTRO ->6
//	7              <0> padhv[%h:3,4] vM/KEYS ->8
//	b        <1> rv2hv[t7] vKM/KEYS ->c
//
//	our %h; local %h; keys %h
//	-     <1> rv2hv[t3] vK/OURINTR ->2
//	4     <1> rv2hv[t5] vKM/LVINTRO ->5
//	7        <1> rv2hv[t7] vKM/KEYS ->8
//
//	keys %h; my %h; keys %h          rv2hv, then padhv[%h:1,2] twice
//	my %h = keys %h                  rv2hv, padhv[%h:1,2]
//	my %h = do { keys %h }           rv2hv, padhv[%h:3,4]
//	my %h; sub f { keys %h }         padhv[%h:1,4]
//	if ((my %g) = ()) { keys %g } keys %g     padhv[%g:2,5] twice, rv2hv
//	while (my %g = ()) { keys %g } keys %g    padhv[%g:2,5] twice, rv2hv
func TestHashScopeResolvesInnermostDeclaration(t *testing.T) {
	checkHashScopes(t, false, []hashScopeCase{
		{"my shadows our", "our %h; { my %h; keys %h } keys %h",
			[]string{"our %h", "my %h", "%h=my", "%h=our"}},
		{"closed block", "{ my %h; } keys %h",
			[]string{"my %h", "%h=pkg"}},
		{"use before my", "keys %h; my %h; keys %h",
			[]string{"%h=pkg", "my %h", "%h=my"}},
		{"local of an our hash", "our %h; local %h; keys %h",
			[]string{"our %h", "local %h", "%h=local"}},
		{"element and slices", "my %h; $h{a}; @h{1, 2}; %h{1}; %h[0]",
			[]string{"my %h", "$h=my", "@h=my", "%h=my"}},
		{"initialiser sees the outer hash", "my %h = keys %h",
			[]string{"my %h", "%h=pkg"}},
		{"do block in the initialiser", "my %h = do { keys %h }",
			[]string{"my %h", "%h=pkg"}},
		{"list declaration", "my (%a, $b) = (1); keys %a",
			[]string{"my %a", "%a=my"}},
		{"sub body sees the file's my", "my %h; sub f { keys %h }",
			[]string{"my %h", "%h=my"}},
		{"sub body's my ends with it", "sub f { my %h } keys %h",
			[]string{"my %h", "%h=pkg"}},
		{"if condition reaches its block, not past it", "if ((my %g) = ()) { keys %g } keys %g",
			[]string{"my %g", "%g=my", "%g=pkg"}},
		{"while condition reaches its block, not past it", "while (my %g = ()) { keys %g } keys %g",
			[]string{"my %g", "%g=my", "%g=pkg"}},
	})
}

// TestHashScopeTracksStateFeature: `state %h` declares only where the state
// feature is on, and the feature is lexical. Measured on perl 5.44.0,
// -MO=Concise:
//
//	use feature "state"; state %h; keys %h
//	<0> padhv[%h:2,3] vM/LVINTRO,STATE
//	<0> padhv[%h:2,3] vM/KEYS
//
//	no feature "state"; state %h; keys %h      <1> rv2hv[t3] vKM/KEYS
//	state %h; keys %h                          <1> rv2hv[t3] vKM/KEYS
//	use v5.10; state %h; keys %h               padhv[%h:2,3] twice
//	use 5.010; state %h; keys %h               padhv[%h:2,3] twice
//	use v5.36; state %h; keys %h               padhv[%h:2,3] twice
//	use feature ":5.10"; state %h; keys %h     padhv[%h:2,3] twice
//	use feature ":all"; state %h; keys %h      padhv[%h:2,3] twice
//	use feature qw(say state); state %h; keys %h      padhv[%h:2,3] twice
//	{ use feature "state"; state %h; keys %h }        padhv[%h:3,4] twice
//	{ use feature "state"; } state %h; keys %h        <1> rv2hv[t3]
//	use feature "state"; { no feature "state"; state %h; keys %h }   <1> rv2hv[t3]
//	use feature "state"; no feature; state %h; keys %h               <1> rv2hv[t3]
//	use feature "state"; use v5.8; state %h; keys %h                 <1> rv2hv[t3]
//	use feature "state"; no feature ":default"; state %h; keys %h    padhv[%h:3,4] twice
func TestHashScopeTracksStateFeature(t *testing.T) {
	declared := []string{"state %h +s", "%h=state +s"}
	undeclared := []string{"state! %h", "%h=pkg"}
	checkHashScopes(t, true, []hashScopeCase{
		{"use feature state", `use feature "state"; state %h; keys %h`, declared},
		{"no feature state", `no feature "state"; state %h; keys %h`, undeclared},
		{"no feature at all", `state %h; keys %h`, undeclared},
		{"use v5.10", `use v5.10; state %h; keys %h`, declared},
		{"use 5.010", `use 5.010; state %h; keys %h`, declared},
		{"use v5.36", `use v5.36; state %h; keys %h`, declared},
		{"bundle :5.10", `use feature ":5.10"; state %h; keys %h`, declared},
		{"bundle :all", `use feature ":all"; state %h; keys %h`, declared},
		{"qw list", `use feature qw(say state); state %h; keys %h`, declared},
		{"on inside a block", `{ use feature "state"; state %h; keys %h }`, declared},
		{"on inside a block, off after it", `{ use feature "state"; } state %h; keys %h`, undeclared},
		{"off inside a block", `use feature "state"; { no feature "state"; state %h; keys %h }`, undeclared},
		{"bare no feature resets", `use feature "state"; no feature; state %h; keys %h`, undeclared},
		{"use v5.8 resets", `use feature "state"; use v5.8; state %h; keys %h`, undeclared},
		{"no feature :default keeps it", `use feature "state"; no feature ":default"; state %h; keys %h`, declared},
		{"a use reports the feature", `{ use feature "state"; keys %h } keys %h`,
			[]string{"%h=pkg +s", "%h=pkg"}},
	})
}

// TestHashScopeClosedBlockDoesNotLeak: a block's `my %h` ends with the
// block. Measured on perl 5.44.0, -MO=Concise:
//
//	{ my %h; } keys %h
//	<0> padhv[%h:2,3] vM/LVINTRO
//	<1> rv2hv[t3] vKM/KEYS
func TestHashScopeClosedBlockDoesNotLeak(t *testing.T) {
	checkHashScopes(t, false, []hashScopeCase{
		{"bare block", "{ my %h; } keys %h", []string{"my %h", "%h=pkg"}},
		{"nested blocks", "{ { my %h; } keys %h }", []string{"my %h", "%h=pkg"}},
		{"if block", "if (1) { my %h } keys %h", []string{"my %h", "%h=pkg"}},
		{"back to the our", "our %h; { my %h; } keys %h", []string{"our %h", "my %h", "%h=our"}},
	})
}

// TestHashScopeUseBeforeMyIsPackage: a use before `my %h` is the package
// hash. Measured on perl 5.44.0, -MO=Concise:
//
//	keys %h; my %h; keys %h
//	<1> rv2hv[t2] vKM/KEYS
//	<0> padhv[%h:1,2] vM/LVINTRO
//	<0> padhv[%h:1,2] vM/KEYS
//
//	my %h; my $n = (keys %h, my %h)     keys %h is the first %h, padhv[%h:1,3]
func TestHashScopeUseBeforeMyIsPackage(t *testing.T) {
	checkHashScopes(t, false, []hashScopeCase{
		{"statement before", "keys %h; my %h; keys %h", []string{"%h=pkg", "my %h", "%h=my"}},
		{"same statement", "my $n = (keys %h, my %h); keys %h", []string{"%h=pkg", "my %h", "%h=my"}},
	})
}

// TestHashScopeNoFeatureStateIsNotADeclaration: with the state feature off,
// `state %h` is `state % h` and the later `%h` is the package hash.
// Measured on perl 5.44.0, -MO=Concise:
//
//	no feature "state"; state %h; keys %h
//	<1> rv2hv[t3] vKM/KEYS
//
// and that rv2hv is the only hash op in the program.
func TestHashScopeNoFeatureStateIsNotADeclaration(t *testing.T) {
	checkHashScopes(t, true, []hashScopeCase{
		{"no feature state", `no feature "state"; state %h; keys %h`, []string{"state! %h", "%h=pkg"}},
		{"off after a bundle", `use v5.10; no feature "state"; state %h; keys %h`, []string{"state! %h", "%h=pkg"}},
		{"does not hide the our", `our %h; no feature "state"; state %h; keys %h`, []string{"our %h", "state! %h", "%h=our"}},
	})
}

// TestHashScopeFeatureEndsWithItsBlock: `use feature "state"` inside a block
// does not reach a `state %h` after it. Measured on perl 5.44.0,
// -MO=Concise:
//
//	{ use feature "state"; } state %h; keys %h
//	<1> rv2hv[t3] vKM/KEYS
//
//	sub f { use v5.10; } state %h; keys %h     <1> rv2hv
func TestHashScopeFeatureEndsWithItsBlock(t *testing.T) {
	checkHashScopes(t, true, []hashScopeCase{
		{"bare block", `{ use feature "state"; } state %h; keys %h`, []string{"state! %h", "%h=pkg"}},
		{"sub body", `sub f { use v5.10; } state %h; keys %h`, []string{"state! %h", "%h=pkg"}},
	})
}

// TestHashScopeSignatureDeclaresMy: a sub's or method's signature hash is a
// `my` of its body. Measured on perl 5.44.0, -MO=Concise,-exec,-main,f:
//
//	use feature "signatures"; sub f ($x, %o) { keys %o } keys %o
//	padhv[%o:4,5] (in f), rv2hv (in main)
//
//	use v5.40; use feature "class"; class C { method m (%o) { keys %o } }
//	padhv[%o:8,9] (in C::m)
func TestHashScopeSignatureDeclaresMy(t *testing.T) {
	checkHashScopes(t, false, []hashScopeCase{
		{"named sub", `use feature "signatures"; sub f ($x, %o) { keys %o } keys %o`,
			[]string{"sig %o", "%o=my", "%o=pkg"}},
		{"lone hash parameter", `use feature "signatures"; sub f (%o) { keys %o } keys %o`,
			[]string{"sig %o", "%o=my", "%o=pkg"}},
		{"anonymous sub", `use feature "signatures"; my $f = sub (%p) { keys %p }; keys %p`,
			[]string{"sig %p", "%p=my", "%p=pkg"}},
		{"method", `use v5.40; use feature "class"; class C { method m (%o) { keys %o } }`,
			[]string{"sig %o", "%o=my"}},
	})
}

// TestHashScopeStateUnknownThroughModule: a module or a BEGIN block can turn
// state on, which the resolver cannot see, so the feature is unknown after
// one rather than off. Core pragmas that touch no feature leave it as it
// was, and an explicit `use`/`no feature` decides it again. Measured on perl
// 5.44.0, -MO=Concise,-exec:
//
//	use if 1, feature => "state"; state %h; keys %h               padhv
//	BEGIN { feature->import("state") } state %h; keys %h          padhv
//	BEGIN { 1 } state %h; keys %h                                 rv2hv
//	use warnings; state %h; keys %h                               rv2hv
//	use Data::Dumper; use feature "state"; state %h; keys %h      padhv
//	use Data::Dumper; no feature "state"; state %h; keys %h       rv2hv
//	{ use if 1, feature => "state"; } state %h; keys %h           rv2hv
//	use feature "st" . "ate"; state %h; keys %h                   padhv
//
// `use Modern::Perl` and `no Foo` name modules the pinned perl does not
// have; they stand for any module whose import this resolver cannot read.
func TestHashScopeStateUnknownThroughModule(t *testing.T) {
	unknown := []string{"state? %h ?s", "%h=state? ?s"}
	checkHashScopes(t, true, []hashScopeCase{
		{"use if", `use if 1, feature => "state"; state %h; keys %h`, unknown},
		{"BEGIN import", `BEGIN { feature->import("state") } state %h; keys %h`, unknown},
		{"any BEGIN", `BEGIN { 1 } state %h; keys %h`, unknown},
		{"Modern::Perl", `use Modern::Perl; state %h; keys %h`, unknown},
		{"no of a module", `use feature "state"; no Foo; state %h; keys %h`, unknown},
		{"quiet pragma", `use warnings; state %h; keys %h`, []string{"state! %h", "%h=pkg"}},
		{"explicit use after", `use Data::Dumper; use feature "state"; state %h; keys %h`,
			[]string{"state %h +s", "%h=state +s"}},
		{"explicit no after", `use Data::Dumper; no feature "state"; state %h; keys %h`,
			[]string{"state! %h", "%h=pkg"}},
		{"ends with its block", `{ use if 1, feature => "state"; } state %h; keys %h`,
			[]string{"state! %h", "%h=pkg"}},
		{"computed feature list", `use feature "st" . "ate"; state %h; keys %h`, unknown},
	})
}

// TestHashScopeCStyleForInit: a C-style for's init declaration reaches its
// condition, step and block, and ends with the loop. Measured on perl
// 5.44.0, -MO=Concise,-exec:
//
//	for (my %h = (); keys %h; ) { keys %h } keys %h
//	padhv padhv padhv rv2hv
func TestHashScopeCStyleForInit(t *testing.T) {
	checkHashScopes(t, false, []hashScopeCase{
		{"init reaches condition", "for (my %h = (); keys %h; ) { keys %h } keys %h",
			[]string{"my %h", "%h=my", "%h=my", "%h=pkg"}},
	})
}

// TestHashScopeFieldDeclaresMy: a class field is a lexical its methods see.
// Measured on perl 5.44.0, -MO=Concise,-exec,C::m:
//
//	use feature "class"; class C { field %h; method m { keys %h } }
//	padhv[%h:FAKE:]
func TestHashScopeFieldDeclaresMy(t *testing.T) {
	checkHashScopes(t, false, []hashScopeCase{
		{"field in a method", `use feature "class"; class C { field %h; method m { keys %h } }`,
			[]string{"field %h", "%h=field"}},
	})
}

// TestHashScopeDevelopmentBundles: state is in the bundle from 5.9.5, the
// development release it arrived in. Measured on perl 5.44.0,
// -MO=Concise,-exec, each followed by `state %h; keys %h`:
//
//	use 5.9.5       padhv        use 5.009005    padhv
//	use v5.9.5      padhv        use 5.009_005   padhv
//	use 5.9.4       rv2hv        use 5.009004    rv2hv
//	use 5.010_001   padhv        use 5.01        padhv
func TestHashScopeDevelopmentBundles(t *testing.T) {
	on := []string{"state %h +s", "%h=state +s"}
	off := []string{"state! %h", "%h=pkg"}
	var cases []hashScopeCase
	for v, want := range map[string][]string{
		"5.9.5": on, "5.009005": on, "v5.9.5": on, "5.009_005": on, "5.010_001": on, "5.01": on,
		"5.9.4": off, "5.009004": off,
	} {
		cases = append(cases, hashScopeCase{"use " + v, "use " + v + "; state %h; keys %h", want})
	}
	checkHashScopes(t, true, cases)
}

// TestHashScopeDeclaredRefs: `my \%h` declares %h. Measured on perl 5.44.0,
// -MO=Concise,-exec:
//
//	use feature "declared_refs","refaliasing"; my \%h = {}; keys %h             padhv
//	use feature "declared_refs","refaliasing"; my (\%h, $x) = ({}, 1); keys %h  padhv
func TestHashScopeDeclaredRefs(t *testing.T) {
	checkHashScopes(t, false, []hashScopeCase{
		{"my backslash", `use feature "declared_refs","refaliasing"; my \%h = {}; keys %h`,
			[]string{"my %h", "%h=my"}},
		{"in a list", `use feature "declared_refs","refaliasing"; my (\%h, $x) = ({}, 1); keys %h`,
			[]string{"my %h", "%h=my"}},
	})
}

// TestHashScopeSignaturesFeature: a parenthesised sub head is a signature
// where the signatures feature is on, and an illegal prototype ("Illegal
// character in prototype") where it is off. The feature is lexical, a
// `use VERSION` sets its bundle's (5.35 on), and a module or BEGIN block
// leaves it unknown, where a head's hashes are declared Unsure. A method's
// head is always a signature. The parser's own choice of prototype or
// signature is file-wide; the resolver follows perl's lexical feature
// either way. Measured on perl 5.44.0, -MO=Concise,-exec,-main,f:
//
//	use experimental "signatures"; sub f (%o) { keys %o }             padhv
//	use experimental "class"; class C { method m (%o) { keys %o } }   padhv
//	use feature "class"; class C { method m (%o) { keys %o } }        padhv[%o:5,6]
//	sub f (%o) { keys %o }                                            rv2hv
//	{ use feature "signatures"; } sub f (%o) { keys %o }              rv2hv
//	use feature "signatures"; { no feature "signatures"; }
//	sub f (%o) { keys %o }                                            padhv[%o:6,7]
//	use 5.035; sub f (%o) { keys %o }                                 padhv
//	use feature "signatures"; use v5.10; sub f (%o) { keys %o }       rv2hv
//
// `use Mojo::Base -signatures` turns the feature on through a module the
// pinned perl does not have; it stands for any import this resolver cannot
// read, as `BEGIN { 1 }` does.
func TestHashScopeSignaturesFeature(t *testing.T) {
	checkHashScopes(t, false, []hashScopeCase{
		{"use experimental", `use experimental "signatures"; sub f (%o) { keys %o } keys %o`,
			[]string{"sig %o", "%o=my", "%o=pkg"}},
		{"experimental class method", `use experimental "class"; class C { method m (%o) { keys %o } }`,
			[]string{"sig %o", "%o=my"}},
		{"class feature method", `use feature "class"; class C { method m (%o) { keys %o } }`,
			[]string{"sig %o", "%o=my"}},
		{"Mojo::Base", `use Mojo::Base -strict, -signatures; sub f (%o) { keys %o }`,
			[]string{"%o=my?"}},
		{"BEGIN", `BEGIN { 1 } sub f (%o) { keys %o }`,
			[]string{"%o=my?"}},
		{"off", `sub f (%o) { keys %o }`,
			[]string{"%o=pkg"}},
		{"on inside a block, off after it", `{ use feature "signatures"; } sub f (%o) { keys %o }`,
			[]string{"sig! %o", "%o=pkg"}},
		{"off inside a block, on after it", `use feature "signatures"; { no feature "signatures"; } sub f (%o) { keys %o }`,
			[]string{"%o=my"}},
		{"5.35 bundle", `use 5.035; sub f (%o) { keys %o }`,
			[]string{"sig %o", "%o=my"}},
		{"use VERSION resets it", `use feature "signatures"; use v5.10; sub f (%o) { keys %o }`,
			[]string{"sig! %o", "%o=pkg"}},
	})
}

// TestHashScopeImportPoints: a module's import, or a BEGIN block that
// loads or calls code, may declare a lexical %h (builtin::export_lexically),
// so a package-hash use after one is Imported, until a declaration read
// after it. Measured on 5.44.0, with Lx's import and x.pl exporting %h:
//
//	use Lx; keys %h                          padhv
//	our %h; use Lx; keys %h                  padhv
//	use Lx; our %h; keys %h                  rv2hv
//	{ use Lx; } keys %h                      rv2hv
//	BEGIN { require "./x.pl" } keys %h       padhv
//	BEGIN { $| = 1 } keys %h                 rv2hv
//	use strict; keys %h                      rv2hv (with no strict vars)
//	use parent -norequire, "Lp"; keys %h     rv2hv: it loads nothing
//	use lib do { builtin::export_lexically("%h" => {}); "." }; keys %h   padhv
//	sub BEGIN { builtin::export_lexically("%h" => {}) } keys %h          padhv
//	BEGIN { my $f = \&builtin::export_lexically; &$f("%h" => {}) } keys %h  padhv
func TestHashScopeImportPoints(t *testing.T) {
	for _, c := range []struct {
		src      string
		imported bool
	}{
		{`use Lx; keys %h;`, true},
		{`our %h; use Lx; keys %h;`, true},
		{`local %h; use Lx; keys %h;`, true},
		{`use Lx; our %h; keys %h;`, false},
		{`{ use Lx; } keys %h;`, false},
		{`BEGIN { require "./x.pl" } keys %h;`, true},
		{`BEGIN { f() } keys %h;`, true},
		{`BEGIN { Foo->import } keys %h;`, true},
		{`BEGIN { $| = 1 } keys %h;`, false},
		{`use strict; use warnings; keys %h;`, false},
		{`use Lx; keys %Foo::h;`, false},
		{`use Lx; my %h; keys %h;`, false},
		{`keys %h; use Lx;`, false},
		{`use parent "Lp"; keys %h;`, true},
		{`use base "Lp"; keys %h;`, true},
		{`use parent -norequire, "Lp"; keys %h;`, false},
		{`use lib do { builtin::export_lexically("%h" => {}); "." }; keys %h;`, true},
		{`use lib "."; keys %h;`, false},
		{`sub BEGIN { builtin::export_lexically("%h" => {}) } keys %h;`, true},
		{`BEGIN { my $f; &$f("%h" => {}) } keys %h;`, true},
		{`BEGIN { my $f; &{$f}("%h" => {}) } keys %h;`, true},
		{`BEGIN { print 1 if defined &DynaLoader::boot_DynaLoader } keys %h;`, false},
		{`BEGIN { my $f = \&foo } keys %h;`, false},
		{`BEGIN { &foo; } keys %h;`, true},
		{`BEGIN { $^H{FOO} = 1 } keys %h;`, false},
	} {
		root := Parse([]byte(c.src))
		uses, _ := hashScopes(root)
		var last *Node
		for n, u := range uses {
			if !u.Declares && (last == nil || n.Start > last.Start) {
				last = n
			}
		}
		if last == nil {
			t.Errorf("%s: no use resolved", c.src)
			continue
		}
		if got := uses[last].Imported; got != c.imported {
			t.Errorf("%s: the last use's Imported is %v, want %v", c.src, got, c.imported)
		}
	}
}
