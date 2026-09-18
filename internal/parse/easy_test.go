// ABOUTME: T1-easy defined by source-side markers, each tracing to a known parser gap.
// ABOUTME: The definition is the marker list; the rate it produces is measured, never targeted.

package parse_test

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// hardMarkers are constructs this parser is known not to handle, each
// detectable in SOURCE without parsing.
//
// Source-side detection is the whole point. Defining "easy" by how well the
// parser does on a file would make the T1-easy rate a measurement of itself:
// the subset would move every time the parser did, and ">= 70% of T1-easy"
// would be satisfiable by shrinking the denominator. A criterion the parser
// cannot influence is the only kind worth ratcheting against.
//
// Every marker names a row in spec §4.14.2 that is marked PARSER, or an open
// issue. None was chosen to make a number come out:
//
//	marker         issue / spec row          files  clean
//	delete         01a0afc0 keyword table       55    3.6%
//	exists         01a0afc0 keyword table       59   10.2%
//	goto           01a0afc0 keyword table       22    4.5%
//	qx             01a0afc0 scanner rows         5   20.0%
//	glob-angle     01a0afc0 scanner rows         2    0.0%
//	pkg-colon      01a0afc0 scanner rows        14    0.0%
//	deref-brace    01a0ad52 sigils structural   70   27.1%
//	deref-at       01a0ad52 sigils structural   55   14.5%
//	signature      01a0afc1 wrong trees          5   20.0%
//	indirect-new   §4.14.2 MethodCall.Indirect 119   28.6%
//	heredoc        §2 lexer, sublexing          61   11.5%
//	format         §5 format bodies are opaque   2    0.0%
//
// Measured at 4bc971ec over the 986. Files carrying NO marker are 50.6%
// clean against 41.4% for the corpus as a whole, so the markers do separate
// the population -- that is the evidence the list is real rather than
// decorative.
//
// The plan says "easy tier (~400)" and the M1 gate issue already notes that
// "~400" is an estimate rather than a definition. This is the definition.
var hardMarkers = map[string]string{
	"heredoc":      "<<",
	"format":       "\nformat ",
	"qx":           "qx",
	"delete":       "delete ",
	"exists":       "exists ",
	"goto":         "goto ",
	"deref-brace":  "${",
	"deref-at":     "@{",
	"signature":    "use v5.3",
	"indirect-new": "new ",
	"glob-angle":   "<*",
	"pkg-colon":    "$::",
}

// t1Easy returns the files carrying no hard marker, and the markers found.
func t1Easy(t *testing.T) (dir string, easy []string, found map[string][]string) {
	t.Helper()
	dir, files := t1Files(t)
	found = make(map[string][]string, len(files))
	for _, rel := range files {
		src, err := os.ReadFile(filepath.Join(dir, rel))
		if err != nil {
			continue
		}
		text := string(src)
		var hits []string
		for name, probe := range hardMarkers {
			if strings.Contains(text, probe) {
				hits = append(hits, name)
			}
		}
		sort.Strings(hits)
		found[rel] = hits
		if len(hits) == 0 {
			easy = append(easy, rel)
		}
	}
	sort.Strings(easy)
	return dir, easy, found
}

// TestT1EasyIsDefinedByMeasurement records the definition and the command
// that produces it, which is what the M1 gate asks for.
//
// It pins the SIZE of the subset, not its parse rate. The size moves only
// when the corpus or the marker list moves, and both are deliberate acts.
func TestT1EasyIsDefinedByMeasurement(t *testing.T) {
	_, easy, found := t1Easy(t)

	// Pinned at 4bc971ec over the 986 top-level files.
	const want = 660
	if len(easy) != want {
		t.Errorf("T1-easy is %d files, want %d\n"+
			"The corpus or the marker list moved. Both are deliberate: update\n"+
			"this constant in the commit that moved it, and say which.",
			len(easy), want)
	}

	// The markers must actually separate the population. A marker list that
	// selected an arbitrary subset would satisfy the count above and mean
	// nothing, so assert the thing that makes it a definition rather than a
	// filter: marked files parse measurably worse than unmarked ones.
	var markedClean, marked, easyClean int
	dir, _ := t1Files(t)
	for rel, hits := range found {
		src, err := os.ReadFile(filepath.Join(dir, rel))
		if err != nil {
			continue
		}
		clean := countUnknown(parse.Parse(src)) == 0
		if len(hits) > 0 {
			marked++
			if clean {
				markedClean++
			}
		} else if clean {
			easyClean++
		}
	}
	easyRate := 100 * float64(easyClean) / float64(len(easy))
	markedRate := 100 * float64(markedClean) / float64(marked)
	if easyRate <= markedRate {
		t.Errorf("the markers do not separate the population:\n"+
			"  easy   %d of %d clean (%.1f%%)\n"+
			"  marked %d of %d clean (%.1f%%)\n"+
			"A marker list that does not predict failure is a filter, not a "+
			"definition.", easyClean, len(easy), easyRate,
			markedClean, marked, markedRate)
	}
	t.Logf("T1-easy: %d of 986 files; %d clean (%.1f%%). "+
		"Marked files: %d clean of %d (%.1f%%).",
		len(easy), easyClean, easyRate, markedClean, marked, markedRate)
}

// TestT1EasyParseRate is the M1 gate's rate metric.
//
// The gate's target is >= 70%. The measured rate at e0d124d1 is 50.8%, so
// this test RECORDS the gap rather than asserting a pass -- a test that
// asserted 70% today would be red for reasons no commit caused, and a test
// that lowered the bar to today's number would make the target meaningless.
//
// It fails when the rate FALLS, which is the ratchet discipline applied to a
// second metric. Raising the floor is the work of the issues the markers
// name: 01a0afc0 (keywords and scanner rows), 01a0ad52 (sigils), 01a0afc1
// (wrong trees).
func TestT1EasyParseRate(t *testing.T) {
	dir, easy, _ := t1Easy(t)

	var clean int
	for _, rel := range easy {
		src, err := os.ReadFile(filepath.Join(dir, rel))
		if err != nil {
			continue
		}
		if countUnknown(parse.Parse(src)) == 0 {
			clean++
		}
	}
	rate := 100 * float64(clean) / float64(len(easy))

	// Measured at 4bc971ec. The gate wants 70.0.
	const floor = 50.8
	const target = 70.0

	if rate+0.05 < floor {
		t.Errorf("T1-easy parse rate fell: %.1f%% of %d files, floor %.1f%%",
			rate, len(easy), floor)
	}
	if rate > floor+0.05 {
		t.Errorf("T1-easy parse rate ROSE to %.1f%% (floor %.1f%%).\n"+
			"Good news, and it still fails: raise the floor in the commit "+
			"that earned it.", rate, floor)
	}
	if rate < target {
		t.Logf("T1-easy parse rate is %.1f%%, below the gate's %.1f%% target. "+
			"%d of %d files clean; %.0f more files needed.",
			rate, target, clean, len(easy),
			target/100*float64(len(easy))-float64(clean))
	}
}

// TestT1EasyMarkersAreTraceable keeps the marker list honest: every marker
// must name a construct the parser actually fails on, measured here rather
// than asserted in a comment.
//
// Without this, a marker could be added to move the subset -- which is
// exactly the failure mode the source-side rule exists to prevent.
func TestT1EasyMarkersAreTraceable(t *testing.T) {
	dir, _, found := t1Easy(t)

	// Corpus-wide clean rate, the bar each marker must fall below.
	var total, totalClean int
	rates := make(map[string][2]int)
	for rel, hits := range found {
		src, err := os.ReadFile(filepath.Join(dir, rel))
		if err != nil {
			continue
		}
		clean := countUnknown(parse.Parse(src)) == 0
		total++
		if clean {
			totalClean++
		}
		for _, h := range hits {
			c := rates[h]
			c[0]++
			if clean {
				c[1]++
			}
			rates[h] = c
		}
	}
	corpusRate := 100 * float64(totalClean) / float64(total)

	var names []string
	for name := range hardMarkers {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		c, ok := rates[name]
		if !ok || c[0] == 0 {
			t.Errorf("marker %q matches no file in the corpus; "+
				"a marker that selects nothing is not evidence of anything",
				name)
			continue
		}
		rate := 100 * float64(c[1]) / float64(c[0])
		if rate >= corpusRate {
			t.Errorf("marker %q: %d files, %d clean (%.1f%%), "+
				"corpus-wide %.1f%%.\n"+
				"This marker does not predict failure, so it does not belong "+
				"in the list.", name, c[0], c[1], rate, corpusRate)
		}
	}
	t.Logf("%d markers, all below the corpus-wide %.1f%% clean rate",
		len(names), corpusRate)
}
