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
// by span, then symbol name or rule position, with its alternatives' rules and
// splits sorted. Nothing in it depends on pointers, allocation or Alt order.
func forestText(tab *grammar.Tables, f *Forest) string {
	type line struct {
		from, to int
		text     string
	}
	var lines []line
	for _, n := range reachable(f) {
		label := fmt.Sprintf("rule %d dot %d", n.Rule, n.Dot)
		if n.Sym >= 0 {
			label = "sym " + tab.Symbols[n.Sym].Name
		}
		var alts []string
		for _, a := range n.Alts {
			split := n.To // an empty rule
			if a.Right != nil {
				split = a.Right.From
			}
			alts = append(alts, fmt.Sprintf("%d@%d", a.Rule, split))
		}
		slices.Sort(alts)
		lines = append(lines, line{n.From, n.To, label + " " + strings.Join(alts, " ")})
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
// split or one node's span changed changes the text; the same nodes built in
// another order, with their Alts reversed, do not.
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

	// Only alternative 0's split moves, from 1 to 0: every node stays reachable
	// through the third alternative, so no node's own line changes.
	a, b := &Node{Sym: 1, Rule: -1, From: 0, To: 1}, &Node{Sym: 2, Rule: -1, From: 1, To: 2}
	c, d := &Node{Sym: 1, Rule: -1, From: 0, To: 0}, &Node{Sym: 2, Rule: -1, From: 0, To: 2}
	split := func(l, r *Node) string {
		root := &Node{Sym: 0, Rule: -1, From: 0, To: 2, Alts: []Alt{{1, l, r}, {2, c, d}, {3, a, b}}}
		return forestText(tab, &Forest{Root: root})
	}
	if split(a, b) == split(c, d) {
		t.Error("changing one alternative's split left the text unchanged")
	}

	g := rebuilt(f)
	g.Root.To++
	if forestText(tab, g) == want {
		t.Error("changing one node's span left the text unchanged")
	}
	if forestText(tab, f) != want {
		t.Error("changing a rebuilt copy changed the original's text")
	}
}

// rebuilt is a deep copy of f whose nodes are allocated in the reverse of
// f's discovery order and whose Alts are reversed.
func rebuilt(f *Forest) *Forest {
	order := reachable(f)
	copied := map[*Node]*Node{nil: nil}
	for _, n := range slices.Backward(order) {
		copied[n] = &Node{Sym: n.Sym, Rule: n.Rule, Dot: n.Dot, From: n.From, To: n.To}
	}
	for _, n := range order {
		for _, a := range slices.Backward(n.Alts) {
			copied[n].Alts = append(copied[n].Alts, Alt{Rule: a.Rule, Left: copied[a.Left], Right: copied[a.Right]})
		}
	}
	return &Forest{Root: copied[f.Root]}
}

// reachable is every node reachable from f's root, each once, in discovery order.
func reachable(f *Forest) []*Node {
	var order []*Node
	seen := map[*Node]bool{}
	for work := []*Node{f.Root}; len(work) > 0; {
		n := work[len(work)-1]
		work = work[:len(work)-1]
		if n == nil || seen[n] {
			continue
		}
		seen[n] = true
		order = append(order, n)
		for _, a := range n.Alts {
			work = append(work, a.Left, a.Right)
		}
	}
	return order
}
