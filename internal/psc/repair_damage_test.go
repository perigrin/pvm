// ABOUTME: Tests damageFromRepair: the damage read off Repair's re-parsed region equals the tree diff's, and analysis a fresh open's.
// ABOUTME: Covers the fallback when Repair refuses, Class B, sub merges and splits, file: URIs with a local lib, and StrategyReuseAll.

package psc

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/infer"
	"tamarou.com/pvm/internal/parse"
)

// repairOn sets parse.IncrementalRepair for t. Set it before opening a
// document: an owner reads it as it is made.
func repairOn(t *testing.T) {
	t.Helper()
	was := parse.IncrementalRepair
	parse.IncrementalRepair = true
	t.Cleanup(func() { parse.IncrementalRepair = was })
}

// sameDamage reports a and b damage the same: both full, of the same
// class, or the same subs.
func sameDamage(a, b editDamage) bool {
	return a.Full == b.Full && a.ClassB == b.ClassB && slices.Equal(a.Subs, b.Subs)
}

// repairDamage repairs old, the tree of src, into edited, and returns the
// tree, the damage read off its region and off the tree diff, and whether
// Repair repaired; refused, the tree is a fresh parse and region is the
// tree diff's, as analyzeEdit falls back.
func repairDamage(t *testing.T, old *parse.Node, src, edited []byte) (tree *parse.Node, region, diff editDamage, ok bool) {
	t.Helper()
	tree, ok = parse.Repair(old, fullSyncEdit(src, edited), edited)
	if !ok {
		tree = parse.Parse(edited)
	}
	diff = damageFromTreeDiff(old, tree)
	region = diff
	if r := tree.Damage(); r != nil {
		region = damageFromRepair(old, tree, *r)
	} else if ok {
		t.Fatal("Repair repaired and reported no region")
	}
	return tree, region, diff, ok
}

// minimalRegion is the region a Repair re-parsing the fewest top-level
// statements would report for the edit from old to new, two full parses:
// between the statements equal before it and those equal, shifted, after.
func minimalRegion(old, new *parse.Node) parse.Damage {
	o, n, delta := old.Children, new.Children, new.End-old.End
	k := 0
	for k < min(len(o), len(n)) && parse.TreeDiff(o[k], n[k]) == "" {
		k++
	}
	m := 0
	for m < min(len(o), len(n))-k && parse.TreeDiffShifted(o[len(o)-1-m], n[len(n)-1-m], delta) == "" {
		m++
	}
	r := parse.Damage{End: new.End}
	if k > 0 {
		r.Start = n[k-1].End
	}
	if m > 0 {
		r.End = n[len(n)-m].Start
	}
	return r
}

// TestRepairDamageMatchesTreeDiff: on each §7.7.3 row and its undoing, on
// edit5k.pl's typing sequence, each edit Repair'd from the last repaired
// tree, and on every one-byte insertion and deletion of a small file,
// the subs Repair's region damages are the tree diff's.
func TestRepairDamageMatchesTreeDiff(t *testing.T) {
	repaired := 0
	check := func(t *testing.T, old *parse.Node, src, edited []byte) *parse.Node {
		t.Helper()
		tree, region, diff, ok := repairDamage(t, old, src, edited)
		if ok {
			repaired++
		}
		if !sameDamage(region, diff) {
			t.Errorf("%q -> %q: region damage %+v, tree diff %+v", src, edited, region, diff)
		}
		// Regions wider than one statement, as a Repair anchoring at top
		// level would report, over the fresh parses.
		o, n := parse.Parse(src), parse.Parse(edited)
		for _, r := range []parse.Damage{minimalRegion(o, n), {Start: 0, End: n.End}} {
			if d := damageFromRepair(o, n, r); !sameDamage(d, diff) {
				t.Errorf("%q -> %q: damage of region %+v %+v, tree diff %+v", src, edited, r, d, diff)
			}
		}
		return tree
	}
	for _, r := range seventhRows {
		t.Run(r.name, func(t *testing.T) {
			src, edited := []byte(r.src), []byte(strings.Replace(r.src, r.old, r.new, 1))
			back := check(t, parse.Parse(src), src, edited)
			check(t, back, edited, src)
		})
	}
	t.Run("edit5k.pl", func(t *testing.T) {
		src := readFixture(t)
		tree, before := parse.Parse(src), repaired
		for _, buf := range latencyEdits(src, 30) {
			tree, src = check(t, tree, src, buf), buf
		}
		if repaired-before != 30 {
			t.Errorf("Repair repaired %d of 30 edit5k.pl edits, want all", repaired-before)
		}
	})
	t.Run("every offset", func(t *testing.T) {
		const src = "my $g = 1;\nsub f {\n  my $x = 1;\n  return $x + $g;\n}\nsub h { my $y = 'a'; return $y . f() }\n# c\nsub k { return 2 }\nprint k();\n"
		old := parse.Parse([]byte(src))
		for at := 0; at <= len(src); at++ {
			for _, ins := range []string{" ", "\n", "x", "1", ";", "$", "}", "{"} {
				check(t, old, []byte(src), []byte(src[:at]+ins+src[at:]))
			}
			if at < len(src) {
				check(t, old, []byte(src), []byte(src[:at]+src[at+1:]))
			}
		}
	})
	if repaired == 0 {
		t.Fatal("Repair repaired no edit; the region path is unexercised")
	}
	t.Logf("%d edits repaired", repaired)
}

// TestRepairDamageFallbackOnNotOk: a comment added to a sub body Repair
// refuses (a comment is a trigger token), so the owner's tree carries no
// region; the owner reads the damage off the tree diff, re-analyzes that
// sub alone, and equals a fresh open.
func TestRepairDamageFallbackOnNotOk(t *testing.T) {
	repairOn(t)
	src := readFixture(t)
	edited, _ := replaceOnce(t, src, "sub scale {", "    my @kept", "    1; # x\n    my @kept")
	if _, ok := parse.Repair(parse.Parse(src), fullSyncEdit(src, edited), edited); ok {
		t.Fatal("Repair repaired the edit; the case does not exercise the fallback")
	}
	d := openLive(t, src)
	got := d.edit(t, edited)
	if r := d.s.snapshot(liveURI).tree.Damage(); r != nil {
		t.Errorf("the owner's tree carries region %+v after Repair refused", r)
	}
	if len(got.Walked) != 1 || !strings.HasPrefix(got.Walked[0], "sub scale#") {
		t.Errorf("walked %v, want the one edited sub", got.Walked)
	}
	sameAnalysis(t, got, fresh(t, edited))
}

// TestRepairDamageClassBStillFull: `my @kept` becoming `our @kept` in a sub
// body is an edit Repair repairs locally, and still a Class B declaration,
// so the region's damage is full, as the tree diff's is, and the owner
// takes the full path.
func TestRepairDamageClassBStillFull(t *testing.T) {
	src := readFixture(t)
	edited, _ := replaceOnce(t, src, "sub scale {", "my @kept", "our @kept")
	_, region, diff, ok := repairDamage(t, parse.Parse(src), src, edited)
	if !ok {
		t.Fatal("Repair refused the edit; the case does not exercise the region")
	}
	if !region.Full || !region.ClassB || !sameDamage(region, diff) {
		t.Errorf("region damage %+v, tree diff %+v; want both full, Class B", region, diff)
	}
	repairOn(t)
	d := openLive(t, src)
	got := d.edit(t, edited)
	if d.s.snapshot(liveURI).tree.Damage() == nil || got.Walked != nil {
		t.Errorf("walked %v; want a repaired tree analysed on the full path", got.Walked)
	}
	sameAnalysis(t, got, fresh(t, edited))
}

// TestRepairDamageAcrossSubBoundary: an edit that merges f and g, or splits
// f in two, damages both subs or goes full, through Repair's region or the
// fallback, and through the owner equals a fresh open.
func TestRepairDamageAcrossSubBoundary(t *testing.T) {
	const src = "sub f { my $a = 1; return $a }\nsub g { return 2 }\nsub k { return 3 }\nprint k();\n"
	for _, c := range []struct{ name, old, new string }{
		{"merge", " }\nsub g {", ";"},
		{"split", "my $a = 1;", "my $a = 1; }\nsub h {"},
		{"move the boundary", "return $a }\nsub g {", "}\nsub g { return $a"},
	} {
		t.Run(c.name, func(t *testing.T) {
			edited := []byte(strings.Replace(src, c.old, c.new, 1))
			_, region, diff, _ := repairDamage(t, parse.Parse([]byte(src)), []byte(src), edited)
			for _, d := range []editDamage{region, diff} {
				if !d.Full && !(slices.Contains(d.Subs, "sub f#0") && slices.Contains(d.Subs, "sub g#0")) {
					t.Errorf("damage %+v; want f and g, or full", d)
				}
			}
			repairOn(t)
			live := openLive(t, []byte(src))
			sameAnalysis(t, live.edit(t, edited), fresh(t, edited))
		})
	}
}

// TestRepairDamageFileURIWithLocalModule: on a file: URI whose script uses
// a module from a lib on PERL5LIB, Session.Repair with the document's
// roots repairs an edit in a sub, its region damages the subs the tree
// diff does, and the owner's analysis, read off that region, equals a
// fresh open's.
func TestRepairDamageFileURIWithLocalModule(t *testing.T) {
	dir := t.TempDir()
	lib := filepath.Join(dir, "lib")
	if err := os.MkdirAll(filepath.Join(lib, "My"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(lib, "My", "Mod.pm"), []byte("package My::Mod;\nour @EXPORT = qw(twice);\nsub twice($) { }\n1;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PERL5LIB", lib)
	path := filepath.Join(dir, "main.pl")
	uri := (&url.URL{Scheme: "file", Path: filepath.ToSlash(path)}).String()
	roots, _ := incRoots(path)
	if !slices.Contains(roots, lib) {
		t.Fatalf("roots %v lack the local lib", roots)
	}
	src := []byte("use My::Mod;\nsub f { my @r = (twice 1, 2); return 1 }\nsub g { return f() }\nsub k { return 3 }\nprint k();\n")
	edited := bytes.Replace(src, []byte("return 1 }"), []byte("return 12 }"), 1)

	sess := parse.NewSession()
	old := sess.ParseSourceFrom(path, src, roots...)
	tree, ok := sess.Repair(old, fullSyncEdit(src, edited), path, edited, roots...)
	if !ok {
		t.Fatal("Session.Repair refused the edit; the case does not exercise the region")
	}
	region, diff := damageFromRepair(old, tree, *tree.Damage()), damageFromTreeDiff(old, tree)
	if !sameDamage(region, diff) || !slices.Equal(region.Subs, []string{"sub f#0", "sub g#0"}) {
		t.Errorf("region damage %+v, tree diff %+v; want both [sub f#0 sub g#0]", region, diff)
	}

	repairOn(t)
	s := NewLSPServer()
	s.parseDelay, s.diagDelay = 0, 0
	t.Cleanup(s.Shutdown)
	if err := s.OpenDocument(uri, src); err != nil {
		t.Fatal(err)
	}
	s.change(uri, 1, edited)
	doc := s.snapshot(uri)
	if doc == nil || !bytes.Equal(doc.text.Text, edited) || doc.tree.Damage() == nil {
		t.Fatal("the owner did not land a repaired tree of the edit")
	}
	if !slices.Equal(doc.analysis.Walked, []string{"sub f#0", "sub g#0"}) {
		t.Errorf("walked %v, want [sub f#0 sub g#0]", doc.analysis.Walked)
	}
	want, err := analyze(context.Background(), parse.NewSession().ParseSourceFrom(path, edited, roots...), infer.Options{})
	if err != nil {
		t.Fatal(err)
	}
	sameAnalysis(t, doc.analysis, want)
}

// TestRepairDamageReuseAllHasNoDamagedSubs: StrategyReuseAll edits at top
// level -- inside a comment in a top-level statement, and whitespace
// between two top-level subs -- damage no sub through Repair's region, as
// through the tree diff, and the owner re-analyzes no sub, shifts every
// position, and equals a fresh open. One inside a comment in a sub damages
// that sub alone, on both paths: a replayed sub's diagnostics move only as
// its declaration does, so one the edit moves within it must be walked.
func TestRepairDamageReuseAllHasNoDamagedSubs(t *testing.T) {
	const src = "my %h = (\n  # a comment\n  k => 1,\n);\nsub f {\n  # in f\n  my $x = undef;\n  return $x eq 'n';\n}\n\nsub g { my $u = undef; return $u eq 'n' }\n"
	for _, c := range []struct {
		name, old, new string
		subs           []string
	}{
		{"a comment in a top-level statement", "a comment", "a  comment", nil},
		{"whitespace between subs", "}\n\nsub g", "}\n \nsub g", nil},
		{"a comment in a sub", "in f", "in  f", []string{"sub f#0"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			edited := []byte(strings.Replace(src, c.old, c.new, 1))
			_, region, diff, ok := repairDamage(t, parse.Parse([]byte(src)), []byte(src), edited)
			t.Logf("Repair repaired: %v", ok)
			if region.Full || !slices.Equal(region.Subs, c.subs) || !sameDamage(region, diff) {
				t.Errorf("region damage %+v, tree diff %+v; want subs %v from both", region, diff, c.subs)
			}
			repairOn(t)
			live := openLive(t, []byte(src))
			before := live.s.snapshot(liveURI).analysis
			got := live.edit(t, edited)
			if got.Walked == nil || !slices.Equal(got.Walked, c.subs) {
				t.Errorf("walked %v, want %v", got.Walked, c.subs)
			}
			want := fresh(t, edited)
			if len(want.Diagnostics) != 2 || fmt.Sprint(before.Diagnostics) == fmt.Sprint(want.Diagnostics) {
				t.Fatalf("diagnostics %v became %v; the case does not check the shift", before.Diagnostics, want.Diagnostics)
			}
			sameAnalysis(t, got, want)
		})
	}
}
