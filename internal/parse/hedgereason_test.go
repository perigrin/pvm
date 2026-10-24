// ABOUTME: Every rv2hv hedge Sites makes names its reason, from a closed set,
// ABOUTME: and every reason in the set has a fixture that produces it.

package parse

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parseoracle"
)

// rv2hvHedgeFixtures is one snippet per reason, each measured on perl
// 5.44.0 or quoted from the test that measured it.
var rv2hvHedgeFixtures = map[hedgeReason]string{
	// An Unknown hides whatever it holds (TestUnknownHedgesEveryMarker).
	reasonUnread: "new Foo(\\@a);\n",
	// `$#{...}` is one token: rv2hv=1 (TestArrowHashReadArrayLastIndexKeepsRv2hv).
	reasonHiddenStructure: "my $r; my $x = $#{$r->{a}};\n",
	// perl compiles a string's code: `"@{[ %$r ]}"` is rv2hv=1.
	reasonInterpolatedCode: "my $r; my $s = \"@{[ %$r ]}\";\n",
	// Void context is not measured: rv2hv=0.
	reasonUnmeasuredContext: "my $p; @{$p->{S}};\n",
	// `$p->{f()}` is rv2hv=1, and 0 after `sub f(){1}`.
	reasonKeyMayFold: "my $p; my $x = $p->{f()};\n",
	// `use if 1, feature => "state"` is padhv; with `use if 0`, rv2hv.
	reasonFeatureUnknown: "use if 1, feature => \"state\"; state %h; my $n = keys %h;\n",
	// Lx's import may export %h lexically: padhv when it does.
	reasonMayBeImported: "use Lx; my $n = keys %h;\n",
	// `new Foo(my %h); keys %h` is padhv, the my unread.
	reasonDeclaredInUnread: "new Foo(my %h);\nmy $n = keys %h;\n",
	// `if (0)`'s block is folded away: rv2hv=1, the declaration's.
	reasonConstantCondition: "our %h; keys %h if 0;\n",
}

// rv2hvHedges returns the reason of every rv2hv hedge Sites makes for src.
func rv2hvHedges(src string) []hedgeReason {
	root := Parse([]byte(src))
	decls := newNamedHashes(root, []byte(src))
	var got []hedgeReason
	for _, n := range root.Children {
		for _, d := range decidedMarkers(n, []byte(src), decls) {
			if d.kind == parseoracle.SiteKindHash && d.unresolved {
				got = append(got, d.reason)
			}
		}
	}
	return got
}

// hedgeReasonProblems checks that every hedge hedges reports, over every
// source in fixtures and others, has a reason in set, and that each reason
// in set has a fixture that produces it.
func hedgeReasonProblems(set []hedgeReason, fixtures map[hedgeReason]string, others []string,
	hedges func(string) []hedgeReason) []string {
	var problems []string
	check := func(src string) []hedgeReason {
		got := hedges(src)
		for _, r := range got {
			if !slices.Contains(set, r) {
				problems = append(problems, fmt.Sprintf("%q hedges with reason %q, outside the closed set", src, r))
			}
		}
		return got
	}
	for _, r := range set {
		src, ok := fixtures[r]
		if !ok {
			problems = append(problems, fmt.Sprintf("reason %q has no fixture: it is dead", r))
			continue
		}
		if !slices.Contains(check(src), r) {
			problems = append(problems, fmt.Sprintf("reason %q: its fixture %q does not produce it: it is dead", r, src))
		}
	}
	for _, src := range others {
		check(src)
	}
	return problems
}

// TestEveryRv2hvHedgeNamesItsReason holds Sites to the closed set, over
// the fixtures and every probe snippet.
func TestEveryRv2hvHedgeNamesItsReason(t *testing.T) {
	others, err := probeSources("testdata/m3probes")
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range hedgeReasonProblems(hedgeReasons, rv2hvHedgeFixtures, others, rv2hvHedges) {
		t.Error(p)
	}
}

// TestRv2hvHedgeReasonsAreClosedAndLive: the check fails a hedge with no
// reason, one outside the set, and a reason no fixture produces.
func TestRv2hvHedgeReasonsAreClosedAndLive(t *testing.T) {
	set := []hedgeReason{"a", "b"}
	fixtures := map[hedgeReason]string{"a": "A", "b": "B"}
	fake := func(out map[string][]hedgeReason) func(string) []hedgeReason {
		return func(src string) []hedgeReason { return out[src] }
	}
	if p := hedgeReasonProblems(set, fixtures, nil, fake(map[string][]hedgeReason{"A": {"a"}, "B": {"b"}})); len(p) != 0 {
		t.Errorf("a live, closed set is flagged: %q", p)
	}
	for name, out := range map[string]map[string][]hedgeReason{
		"no reason":       {"A": {"a"}, "B": {"b"}, "C": {""}},
		"outside the set": {"A": {"a"}, "B": {"b"}, "C": {"z"}},
		"dead reason":     {"A": {"a"}, "B": {"a"}},
	} {
		if p := hedgeReasonProblems(set, fixtures, []string{"C"}, fake(out)); len(p) == 0 {
			t.Errorf("%s: not flagged", name)
		}
	}
	delete(fixtures, "b")
	if p := hedgeReasonProblems(set, fixtures, nil, fake(map[string][]hedgeReason{"A": {"a"}})); len(p) == 0 {
		t.Error("a reason with no fixture is not flagged")
	}
}

// probeSources reads each probe snippet's source from dir, in the format
// testdata/m3probes/README.md describes.
func probeSources(dir string) ([]string, error) {
	paths, err := filepath.Glob(filepath.Join(dir, "*.txt"))
	if err != nil || len(paths) == 0 {
		return nil, fmt.Errorf("no probe files in %s: %v", dir, err)
	}
	var srcs []string
	for _, path := range paths {
		text, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		for _, part := range strings.Split(string(text), "\n=== ")[1:] {
			_, body, _ := strings.Cut(part, "\n")
			src, _, _ := strings.Cut(body, "--- facts")
			srcs = append(srcs, src)
		}
	}
	return srcs, nil
}
