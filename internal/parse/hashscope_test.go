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
// for a use no declaration reaches. A declaration perl does not make -- a
// `state` with its feature off -- is `state! %h`. withState appends ` +s`
// where the state feature is on.
func resolvedHashes(src string, withState bool) []string {
	uses := hashScopes(Parse([]byte(src)))
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
		case u.Declares && u.Decl == "state" && !u.StateOn:
			s = "state! " + n.Text
		case u.Declares:
			s = u.Decl + " " + n.Text
		case u.Decl == "":
			s = n.Text + "=pkg"
		default:
			s = n.Text + "=" + u.Decl
		}
		if withState && u.StateOn {
			s += " +s"
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
