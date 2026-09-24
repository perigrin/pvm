// ABOUTME: The twelve hardMarkers must each be exercised by a case in the tier that claims it.
// ABOUTME: Claims live in the tier READMEs, so the list cannot drift from the corpus it describes.
package conformance

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// reHardMarkers matches a tier README's machine-read block:
//
//	## HARD MARKERS
//
//	    goto	goto·
//
// Read exactly as INTRODUCES and DEPENDS ON are: the FIRST indented block
// after the heading, so prose may follow inside the same section. One
// marker per line, `name` then the SOURCE PROBE that detects it.
//
// The probe is the one from easy_test.go verbatim, and it is the load-
// bearing half. A name alone would let a README claim `heredoc` while the
// tier contained nothing resembling one; the probe is what a file in the
// tier must actually contain, so the claim is checkable against the
// directory rather than against itself.
var reHardMarkers = regexp.MustCompile(`(?m)^## HARD MARKERS\s*\n\s*\n((?:[ \t]+\S.*\n?)+)`)

// hardMarkerCount is the number of markers the corpus must account for.
//
// The twelve come from `hardMarkers` in internal/parse/easy_test.go, deleted
// by issue 01a0c436-39e4 once this test passed. The spec's tier-list section
// states the placement that made the deletion safe:
//
//	The twelve `hardMarkers` all place [...] heredoc/format/qx/glob-angle to
//	tier 13, deref-brace and deref-at to tier 08, signature to tier 07,
//	indirect-new to tier 11, goto to tier 06, delete/exists/pkg-colon to
//	tier 02. No residue, so no misc tier.
//
// Only the COUNT lives here. The names, the probes and the tiers live in the
// READMEs of the tiers that claim them, because a Go table repeating them
// would be a second list that must agree with the first -- the drift this
// package already watched a hand-copied op list produce on its first day.
// A count cannot drift silently: it is one number, and moving it is the
// deliberate act of adding or retiring a marker.
const hardMarkerCount = 12

// hardMarker is one claim: a tier says it covers this construct, and names
// the source probe that finds it.
type hardMarker struct {
	name  string
	probe string
	tier  string
}

// readHardMarkers collects every tier's HARD MARKERS claims.
//
// Tiers without the block are not an error -- most tiers claim no marker.
// A marker claimed TWICE is, because then no single tier owns it and
// retiring either claim would silently drop the coverage.
func readHardMarkers(corpus string) ([]hardMarker, error) {
	entries, err := os.ReadDir(corpus)
	if err != nil {
		return nil, fmt.Errorf("reading the corpus root: %w", err)
	}

	var out []hardMarker
	seen := map[string]string{}
	for _, e := range entries {
		if !e.IsDir() || !isTierDir(e.Name()) {
			continue
		}
		path := filepath.Join(corpus, e.Name(), "README.md")
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("tier %s has no README: %w", e.Name(), err)
		}
		m := reHardMarkers.FindSubmatch(raw)
		if m == nil {
			continue
		}
		for _, line := range strings.Split(string(m[1]), "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			name, probe, ok := strings.Cut(line, "\t")
			if !ok {
				return nil, fmt.Errorf(
					"%s: HARD MARKERS line %q takes a name and a tab and a probe",
					path, line)
			}
			name = strings.TrimSpace(name)
			// The probe is NOT trimmed: `delete ` and `exists ` end in a
			// significant space, and `format` begins with a newline.
			probe = decodeProbe(probe)
			if probe == "" {
				return nil, fmt.Errorf("%s: marker %q has an empty probe", path, name)
			}
			if prev, dup := seen[name]; dup {
				return nil, fmt.Errorf(
					"marker %q is claimed by both %s and %s; one tier owns each",
					name, prev, e.Name())
			}
			seen[name] = e.Name()
			out = append(out, hardMarker{name: name, probe: probe, tier: e.Name()})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].name < out[j].name })
	return out, nil
}

// decodeProbe turns a README's written probe into the bytes to search for.
//
// Two of the twelve probes are whitespace that Markdown cannot carry: a
// trailing space (`delete `, `exists `, `goto `, `new `) and a leading
// newline (`\nformat `). Writing them literally in an indented block would
// leave them at the mercy of any editor that strips trailing whitespace --
// which would silently WIDEN the probe from `delete ` to `delete`, and a
// widened probe matches more, which is the direction that lets a tier pass
// without the construct.
func decodeProbe(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\\n", "\n")
	s = strings.ReplaceAll(s, "·", " ")
	return s
}

// TestEveryHardMarkerPlaced is the gate that made deleting easy_test.go safe.
//
// `hardMarkers` encoded twelve constructs this parser failed on at 4bc971ec.
// The spec calls that scaffolding: in a graded corpus each one is a tier
// PLACEMENT rather than a filter. Deleting the file loses the information
// unless the corpus already carries it, so this test asserts the corpus
// does -- and it asserts the STRONG reading.
//
// The weak reading is that a tier README MENTIONS the marker. That was
// measured before choosing, and it fails on its own terms: `glob-angle`,
// `indirect-new` and `pkg-colon` are coinages from easy_test.go that no
// README writes, while `exists` -- an ordinary English word -- appears in
// ten of the fourteen. Mentioning is simultaneously too weak to prove
// coverage and too noisy to disprove it.
//
// The strong reading is that the tier contains a CASE the marker's own
// source probe finds. That is the thing easy_test.go measured: the probe is
// what marked a T1 file hard, so a corpus case the same probe matches is a
// corpus case exercising the same construct. The probe reads the case's
// ```perl block, which is the Perl a `.t` file used to be all of -- the
// prose and the other blocks around it are not source and never were.
func TestEveryHardMarkerPlaced(t *testing.T) {
	markers, err := readHardMarkers(corpusDir)
	if err != nil {
		t.Fatalf("reading tier READMEs: %v", err)
	}

	if len(markers) != hardMarkerCount {
		var have []string
		for _, m := range markers {
			have = append(have, m.name+" ("+m.tier+")")
		}
		t.Fatalf("the corpus claims %d hard markers, want %d:\n  %s\n\n"+
			"Every marker needs a `## HARD MARKERS` block in the README of "+
			"the tier that covers it.",
			len(markers), hardMarkerCount, strings.Join(have, "\n  "))
	}

	cases, err := AllCases(corpusDir)
	if err != nil {
		t.Fatalf("reading the corpus: %v", err)
	}

	for _, m := range markers {
		var hits []string
		for _, c := range cases {
			if c.Tier != m.tier {
				continue
			}
			if strings.Contains(c.Source, m.probe) {
				hits = append(hits, c.Key)
			}
		}
		if len(hits) == 0 {
			t.Errorf("marker %q claims %s, but no case there contains %q.\n"+
				"A README that names a construct its tier does not exercise "+
				"is a claim about itself.", m.name, m.tier, m.probe)
			continue
		}
		// Comma-joined: a case key contains spaces, where a file
		// name did not.
		t.Logf("%-13s %s  %s", m.name, m.tier, strings.Join(hits, ", "))
	}
}
