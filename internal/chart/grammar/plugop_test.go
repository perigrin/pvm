// ABOUTME: Tests LibraryTerminal: a library operator's class or level to the PLUGIN_* terminal tokentype_for_plugop gives it.
// ABOUTME: Each threshold's anchor level maps to its own bucket by perl's <=, the level above it to the next.

package grammar

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestPlugopLevelBoundaries: a level exactly on a threshold (its anchor's binding power) maps to that threshold's
// bucket, the level just below it to the same bucket, and the level just above to the next; past the last anchor is
// HIGH. A class perl names maps to its bucket whatever the level.
func TestPlugopLevelBoundaries(t *testing.T) {
	at := func(bp int) parse.LibraryOperator { return parse.LibraryOperator{Info: parse.OpInfo{BP: bp}} }
	var anchored []int // the plugopLevels with an anchor
	for i, l := range plugopLevels {
		if l.anchor != "" {
			anchored = append(anchored, i)
		}
	}
	for j, i := range anchored {
		l := plugopLevels[i]
		info, ok := parse.InfixInfo(l.anchor)
		if !ok {
			t.Fatalf("%s: no binding power for %q", l.class, l.anchor)
		}
		next := "PLUGIN_HIGH_OP"
		if j+1 < len(anchored) {
			next = plugopLevels[anchored[j+1]].terminal
		}
		for bp, want := range map[int]string{info.BP - 1: l.terminal, info.BP: l.terminal, info.BP + 1: next} {
			if got := LibraryTerminal(at(bp)); got != want {
				t.Errorf("level %d (%s's is %d): got %s, want %s", bp, l.anchor, info.BP, got, want)
			}
		}
	}
	for _, l := range plugopLevels {
		op := at(1 << 20)
		op.Class = l.class
		if got := LibraryTerminal(op); got != l.terminal {
			t.Errorf(":infix(%s): got %s, want %s", l.class, got, l.terminal)
		}
	}
	if got := LibraryTerminal(at(0)); got != "PLUGIN_LOGICAL_OR_LOW_OP" {
		t.Errorf("a level below or's, placed by relation: got %s, want PLUGIN_LOGICAL_OR_LOW_OP", got)
	}
}
