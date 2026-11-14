// ABOUTME: Holds forest identity to a committed golden: one digest per accepted file of the corpus sample.
// ABOUTME: The dump names no Go type, so a rewrite of the node types rewrites the dump and leaves the golden.

package forest

import (
	"cmp"
	"crypto/sha256"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
)

var updateForestDigest = flag.Bool("forest.update-forest-digest", false,
	"rewrite testdata/forest-digest.golden from this tree's forest")

var forestDigestPath = filepath.Join("testdata", "forest-digest.golden")

// forestText is f as text in a fixed order: each node reachable from the root
// by span, then symbol name or rule position, with its alternatives sorted,
// each its rule, split and both children's keys (label and span, - for none).
// Nothing in it depends on pointers, allocation or Alt order.
func forestText(tab *grammar.Tables, f *Forest) string {
	type line struct {
		from, to int
		text     string
	}
	key := func(i int32) string {
		n := f.Nodes[i]
		switch {
		case i == 0:
			return "-"
		case n.Sym >= 0:
			return fmt.Sprintf("s%s:%d-%d", tab.Symbols[n.Sym].Name, n.From, n.To)
		}
		return fmt.Sprintf("r%d.%d:%d-%d", n.Rule, n.Dot, n.From, n.To)
	}
	var lines []line
	for _, i := range reachable(f) {
		n := f.Nodes[i]
		var alts []string
		for x := n.Alt; x != 0; x = f.Alts[x].Next {
			a := f.Alts[x]
			split := n.To // an empty rule
			if a.Right != 0 {
				split = f.Nodes[a.Right].From
			}
			alts = append(alts, fmt.Sprintf("%d@%d(%s,%s)", a.Rule, split, key(a.Left), key(a.Right)))
		}
		slices.Sort(alts)
		lines = append(lines, line{int(n.From), int(n.To), key(i) + " " + strings.Join(alts, " ")})
	}
	slices.SortFunc(lines, func(x, y line) int {
		return cmp.Or(cmp.Compare(x.from, y.from), cmp.Compare(x.to, y.to), strings.Compare(x.text, y.text))
	})
	var b strings.Builder
	for _, l := range lines {
		fmt.Fprintf(&b, "%d-%d %s\n", l.from, l.to, l.text)
	}
	return b.String()
}

// forestDigest is the golden's text: the accepted count, then each accepted
// file's path and the hash of its forest's text.
func forestDigest(t *testing.T) (string, int) {
	t.Helper()
	tab := committedTables(t)
	var lines []string
	for _, p := range samplePaths(t) {
		src, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		if f, _ := Parse(tab, src); f != nil {
			lines = append(lines, fmt.Sprintf("%s %x", filepath.ToSlash(p), sha256.Sum256([]byte(forestText(tab, f)))))
		}
	}
	return fmt.Sprintf("accepted %d\n%s\n", len(lines), strings.Join(lines, "\n")), len(lines)
}

// TestForestDigestGolden: the forests of the corpus sample's accepted files
// equal testdata/forest-digest.golden. Never regenerate it to make a change
// pass: a change that moves it has changed the forest the chart builds.
func TestForestDigestGolden(t *testing.T) {
	got, _ := forestDigest(t)
	if *updateForestDigest {
		if err := os.WriteFile(forestDigestPath, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(forestDigestPath)
	if err != nil {
		t.Fatalf("%v (regenerate with -forest.update-forest-digest)", err)
	}
	if got != string(want) {
		t.Errorf("forest identity moved from %s:\ngot:\n%swant:\n%s", forestDigestPath, got, want)
	}
}

// TestForestDigestDetectsChange: the digest cannot pass vacuously and keys on
// content, not construction. The sample has accepted files; one alternative's
// split, Left (to a same-span node, or nil), Right (to a same-start node), or
// Left and Right swapped, or one node's span, changes the text; the same nodes
// built in another order, with their Alts reversed, do not.
func TestForestDigestDetectsChange(t *testing.T) {
	if _, n := forestDigest(t); n == 0 {
		t.Fatal("the chart accepts no file of the sample")
	}
	tab := committedTables(t)
	f, _ := Parse(tab, []byte("my $x = 1 + 2;\nprint $x;\n"))
	if f == nil {
		t.Fatal("no forest")
	}
	want := forestText(tab, f)

	if got := forestText(tab, rebuilt(f)); got != want {
		t.Errorf("the same nodes built in another order changed the text:\n%s\nwant:\n%s", got, want)
	}

	// Only alternative 0 changes: every node stays reachable through the
	// others, so no node's own line changes, only that alternative's.
	const a, b, c, d, pos, e = 1, 2, 3, 4, 5, 6 // pos has a's span, e b's, each another label
	alt0 := func(l, r int32) string {
		g := handForest()
		for _, n := range []Node{
			{Sym: 1, Rule: -1, From: 0, To: 1}, {Sym: 2, Rule: -1, From: 1, To: 2},
			{Sym: 1, Rule: -1, From: 0, To: 0}, {Sym: 2, Rule: -1, From: 0, To: 2},
			{Sym: -1, Rule: 5, Dot: 1, From: 0, To: 1}, {Sym: 3, Rule: -1, From: 1, To: 2},
		} {
			g.newNode(n)
		}
		g.Root = g.newNode(Node{Sym: 0, Rule: -1, From: 0, To: 2})
		for _, x := range []Alt{{Rule: 1, Left: l, Right: r}, {Rule: 2, Left: c, Right: d}, {Rule: 3, Left: a, Right: b}, {Rule: 4, Left: pos, Right: b}, {Rule: 6, Left: a, Right: e}} {
			g.add(g.Root, x)
		}
		return forestText(tab, g)
	}
	for name, pair := range map[string][2][2]int32{
		"split":                       {{a, b}, {c, d}},
		"Left, to a same-span node":   {{a, b}, {pos, b}},
		"Left and Right, swapped":     {{c, d}, {d, c}}, // a zero-width Left: the split stays 0
		"Left, to none":               {{a, b}, {0, b}},
		"Right, to a same-start node": {{a, b}, {a, e}},
	} {
		if alt0(pair[0][0], pair[0][1]) == alt0(pair[1][0], pair[1][1]) {
			t.Errorf("changing one alternative's %s left the text unchanged", name)
		}
	}

	g := rebuilt(f)
	g.Nodes[g.Root].To++
	if forestText(tab, g) == want {
		t.Error("changing one node's span left the text unchanged")
	}
	if forestText(tab, f) != want {
		t.Error("changing a rebuilt copy changed the original's text")
	}
}

// rebuilt is a copy of f whose nodes are numbered in the reverse of f's
// discovery order and whose Alts are in the reverse of f's order.
func rebuilt(f *Forest) *Forest {
	order := reachable(f)
	g := handForest()
	copied := make([]int32, len(f.Nodes)) // 0 stays 0
	for _, n := range slices.Backward(order) {
		nd := f.Nodes[n]
		nd.Alt = 0
		copied[n] = g.newNode(nd)
	}
	for _, n := range order {
		for x := f.Nodes[n].Alt; x != 0; x = f.Alts[x].Next {
			a := f.Alts[x]
			g.add(copied[n], Alt{Rule: a.Rule, Left: copied[a.Left], Right: copied[a.Right]})
		}
	}
	g.Root = copied[f.Root]
	return g
}

// reachable is every node reachable from f's root, each once, in discovery order.
func reachable(f *Forest) []int32 {
	var order []int32
	seen := make([]bool, len(f.Nodes))
	for work := []int32{f.Root}; len(work) > 0; {
		n := work[len(work)-1]
		work = work[:len(work)-1]
		if n == 0 || seen[n] {
			continue
		}
		seen[n] = true
		order = append(order, n)
		for x := f.Nodes[n].Alt; x != 0; x = f.Alts[x].Next {
			work = append(work, f.Alts[x].Left, f.Alts[x].Right)
		}
	}
	return order
}
