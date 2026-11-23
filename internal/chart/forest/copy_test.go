// ABOUTME: Tests the stage API the ratchet's stage keys time: Copy, LibraryOperators, Resolve and CountReadings.
// ABOUTME: A copy shares nothing a stage mutates, so each pass of a stage key times the same unpruned input.

package forest

import (
	"os"
	"testing"

	"tamarou.com/pvm/internal/chart/lex"
	"tamarou.com/pvm/internal/chart/recognize"
)

// TestForestCopySharesNothing: over the corpus sample, LibraryOperators, Constrain and Resolve on a copy leave the
// original's forest text unchanged, and a second copy of the original judges to the same kept set and readings count
// as the first. Some file's stages must change its copy, so a test that pruned nothing cannot pass.
func TestForestCopySharesNothing(t *testing.T) {
	tab := committedTables(t)
	prec, err := ParsePrecedence(tab)
	if err != nil {
		t.Fatal(err)
	}
	judged, changed := 0, 0
	for _, p := range samplePaths(t) {
		src, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		s, _ := lex.Lex(src, tab.Tokens)
		l := lex.Linear(s)
		c, err := recognize.Recognize(recognize.New(tab, recognize.Options{}), l)
		if err != nil {
			t.Fatal(err)
		}
		orig := Build(tab, c, l)
		if orig == nil {
			continue
		}
		before := forestText(tab, orig)
		judge := func() (string, Readings, bool) {
			f := orig.Copy()
			if LibraryOperators(tab, f, s, src, Scope{}) >= 0 || !f.Constrain(prec) || Resolve(tab, f, s, src, Scope{}) >= 0 {
				return "", Readings{}, false
			}
			return forestText(tab, f), f.CountReadings(tab), true
		}
		text1, r1, ok1 := judge()
		if forestText(tab, orig) != before {
			t.Fatalf("%s: judging a copy changed the original forest", p)
		}
		text2, r2, ok2 := judge()
		if ok1 != ok2 || text1 != text2 || r1 != r2 {
			t.Errorf("%s: a second copy judged %v %v, the first %v %v", p, ok2, r2, ok1, r1)
		}
		if ok1 {
			judged++
		}
		if !ok1 || text1 != before {
			changed++
		}
	}
	if judged == 0 || changed == 0 {
		t.Fatalf("%d files judged, %d changed by the stages; want some of each", judged, changed)
	}
}
