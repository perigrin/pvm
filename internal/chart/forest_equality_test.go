// ABOUTME: The forest-equality differential over every waypoint2 row and every referee sentence, with V1's side from forest/equality.go.
// ABOUTME: The referee's V1 path is Build, narrow and Constrain with no LibraryOperators; its comparison drops zero-width slipper ; leaves.

package chart

import (
	"fmt"
	"testing"

	"tamarou.com/pvm/internal/chart/forest"
)

// waypoint2Sources is every source the waypoint2 tables hold, scope rows after their preamble, with the verdicts
// dropped: the differential compares forests, not perl's verdicts.
func waypoint2Sources() []string {
	var srcs []string
	for _, c := range waypoint2Refusals {
		srcs = append(srcs, c.src)
	}
	for _, c := range waypoint2ScopeRefusals {
		srcs = append(srcs, c.pre+c.src)
	}
	return srcs
}

// TestForestEqualityWaypoint2: on every waypoint2 row, refused rows included, the candidate's forest under the
// empty scope equals V1's section 5 pipeline. It logs how many rows it compared of how many it loaded and how many
// V1 refused, and requires every row compared and some refused, so a skipped row or an empty table cannot pass.
func TestForestEqualityWaypoint2(t *testing.T) {
	tab, err := committed()
	if err != nil {
		t.Fatal(err)
	}
	srcs := waypoint2Sources()
	compared, refused := 0, 0
	var diffs []string
	for _, src := range srcs {
		cand, _ := forest.CandidateForest(tab, []byte(src), forest.Scope{})
		v1, _ := forest.V1Forest(tab, []byte(src), forest.Scope{})
		compared++
		if v1 == nil {
			refused++
		}
		if d := forest.ForestDiff(cand, v1); d != "" {
			diffs = append(diffs, fmt.Sprintf("%q: %s", src, d))
		}
	}
	for _, d := range diffs[:min(10, len(diffs))] {
		t.Error(d)
	}
	t.Logf("waypoint2: compared %d of %d rows, refused: %d", compared, len(srcs), refused)
	if len(diffs) > 0 {
		t.Errorf("%d of %d forests differ from V1's", len(diffs), compared)
	}
	if compared == 0 || compared != len(srcs) || refused == 0 {
		t.Errorf("compared %d of %d rows, V1 refused %d; want every row compared, and some refused", compared, len(srcs), refused)
	}
}
