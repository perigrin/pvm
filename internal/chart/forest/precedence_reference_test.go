// ABOUTME: Compares production Constrain against ConstrainReference, verdict and kept set, on two copies of one forest.
// ABOUTME: Covers the toy fixture forests.

package forest

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/chart/recognize"
)

// keptAlts is node n's alternatives, in order.
func keptAlts(f *Forest, n int) []int32 {
	var out []int32
	for a := f.Nodes[n].Alt; a != 0; a = f.Alts[a].Next {
		out = append(out, a)
	}
	return out
}

// altCount is how many alternatives f's nodes hold.
func altCount(f *Forest) int {
	k := 0
	for n := range f.Nodes {
		k += len(keptAlts(f, n))
	}
	return k
}

// precedenceDiff is where production's judgement of one copy of a forest, prodOK and prod, differs from the
// reference's of another copy, refOK and ref: the verdict, or the first node whose kept alternatives differ; ""
// where the two agree.
func precedenceDiff(prodOK bool, prod *Forest, refOK bool, ref *Forest) string {
	if prodOK != refOK {
		return fmt.Sprintf("node %d (root): production verdict %v, reference %v", prod.Root, prodOK, refOK)
	}
	if len(prod.Nodes) != len(ref.Nodes) {
		return fmt.Sprintf("production has %d nodes, reference %d", len(prod.Nodes), len(ref.Nodes))
	}
	for n := range prod.Nodes {
		if p, r := keptAlts(prod, n), keptAlts(ref, n); !slices.Equal(p, r) {
			return fmt.Sprintf("node %d: production keeps alternatives %v, reference %v", n, p, r)
		}
	}
	return ""
}

// judgeBoth makes two copies of f before either is judged, runs Constrain on one and ConstrainReference on the
// other, and is where they differ ("" none), the reference's verdict, and whether it accepted with an alternative
// dropped.
func judgeBoth(f *Forest, p *Precedence) (diff string, ok, pruned bool) {
	prod, ref := f.Copy(), f.Copy()
	prodOK := prod.Constrain(p)
	refOK := ref.ConstrainReference(p)
	return precedenceDiff(prodOK, prod, refOK, ref), refOK, refOK && altCount(ref) < altCount(f)
}

// TestPrecedenceDiffDetectsDifference: the comparison reports a node that lost one alternative on one side, a verdict
// that differs alone, and a copy Constrain judged against one it did not; and nothing where both sides agree. The
// sentence is one the reference accepts and prunes, so a broken Constrain cannot choose it.
func TestPrecedenceDiffDetectsDifference(t *testing.T) {
	ty := loadToy(t, "ladder")
	var f, judged *Forest
	for _, s := range ty.sentences {
		if g, _, _ := toyForest(t, ty.tab, s.words, recognize.Options{}); g != nil {
			if j := g.Copy(); j.ConstrainReference(ty.prec) && altCount(j) < altCount(g) {
				f, judged = g, j
				break
			}
		}
	}
	if f == nil {
		t.Fatal("no ladder sentence the reference accepts and prunes")
	}
	if d := precedenceDiff(true, judged, true, judged.Copy()); d != "" {
		t.Errorf("two equal copies: %s, want no difference", d)
	}

	dropped := judged.Copy()
	n := slices.IndexFunc(dropped.Nodes, func(nd Node) bool { return nd.Alt != 0 })
	dropped.Nodes[n].Alt = dropped.Alts[dropped.Nodes[n].Alt].Next
	if d := precedenceDiff(true, dropped, true, judged); !strings.Contains(d, fmt.Sprintf("node %d:", n)) {
		t.Errorf("one alternative dropped on node %d: %q, want that node reported", n, d)
	}

	if d := precedenceDiff(false, judged, true, judged.Copy()); !strings.Contains(d, fmt.Sprintf("node %d (root)", judged.Root)) {
		t.Errorf("verdict alone differing: %q, want the root reported", d)
	}

	prod := f.Copy()
	if d := precedenceDiff(prod.Constrain(ty.prec), prod, true, f.Copy()); d == "" {
		t.Error("a copy Constrain judged against an unjudged one: no difference reported")
	}
}

// TestPrecedenceToyForestsMatchReference requires production and the reference to keep the same alternatives, and
// reach the same verdict, on every toy fixture sentence's forest.
func TestPrecedenceToyForestsMatchReference(t *testing.T) {
	for _, name := range toyNames {
		ty := loadToy(t, name)
		bad := 0
		for _, s := range ty.sentences {
			f, _, _ := toyForest(t, ty.tab, s.words, recognize.Options{})
			if f == nil {
				continue
			}
			if diff, _, _ := judgeBoth(f, ty.prec); diff != "" {
				if bad++; bad <= 10 {
					t.Errorf("%s %s: %s", name, strings.Join(s.words, " "), diff)
				}
			}
		}
	}
}
