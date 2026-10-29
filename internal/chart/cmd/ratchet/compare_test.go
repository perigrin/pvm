// ABOUTME: Tests the chart ratchet's comparator and line printer against hand-built baselines.
// ABOUTME: Each regression rule must fire on a perturbed baseline and stay quiet on an equal one.

package main

import (
	"regexp"
	"strings"
	"testing"
)

func TestCompareQuietOnEqualBaseline(t *testing.T) {
	m := []Measure{{"parses-scored", "746"}, {"rejects-disagree", "0"}, {"set-reuse", "na"}, {"rd-oracle-exact", "0.8123"}}
	if got := Compare(m, m); len(got) != 0 {
		t.Fatalf("Compare on an equal baseline = %q, want none", got)
	}
}

func TestCompareRules(t *testing.T) {
	for _, tc := range []struct {
		name               string
		key, base, measure string
		regress            bool
	}{
		{"exit count rises", "rejects-disagree", "0", "1", true},
		{"exit count falls", "parses-refused", "5", "4", false},
		{"floor falls", "parses-scored", "747", "746", true},
		{"floor rises", "parsent-scored", "124", "125", false},
		{"defect-yield-settled falls", "defect-yield-settled", "3", "2", true},
		{"defect-yield-chart-exact falls", "defect-yield-chart-exact", "3", "2", true},
		{"number becomes na", "speed-ratio-std5k", "1.2", "na", true},
		{"na becomes a number", "speed-ratio-std5k", "na", "1.2", false},
		{"reported-only key moves", "speed-ratio-std5k", "1.2", "9.9", false},
		{"defect-yield-sites falls, reported only", "defect-yield-sites", "7", "3", false},
		{"rd-oracle-exact falls, reported only", "rd-oracle-exact", "0.9", "0.1", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := Compare([]Measure{{tc.key, tc.base}}, []Measure{{tc.key, tc.measure}})
			if (len(got) > 0) != tc.regress {
				t.Fatalf("Compare(%s %s -> %s) = %q, want regression %v", tc.key, tc.base, tc.measure, got, tc.regress)
			}
			if tc.regress && !(strings.Contains(got[0], tc.key) && strings.Contains(got[0], tc.base) && strings.Contains(got[0], tc.measure)) {
				t.Fatalf("regression %q does not name the key, the baseline and the value", got[0])
			}
		})
	}
}

// A key may be added only together with its baseline, and never renamed.
func TestCompareKeySetsMustMatch(t *testing.T) {
	base := []Measure{{"parses-scored", "746"}}
	if got := Compare(base, []Measure{{"parses-scored", "746"}, {"new-key", "1"}}); len(got) != 1 || !strings.Contains(got[0], "new-key") {
		t.Fatalf("a key missing from the baseline: got %q", got)
	}
	if got := Compare(base, nil); len(got) != 1 || !strings.Contains(got[0], "parses-scored") {
		t.Fatalf("a baseline key no longer printed: got %q", got)
	}
}

var contractLine = regexp.MustCompile(`^chart-ratchet: [a-z0-9-]+ [^ ]+$`)

func TestFormatAndParseRoundTrip(t *testing.T) {
	m := []Measure{{"parses-scored", "746"}, {"set-reuse", "na"}}
	text := Format(m)
	for _, line := range strings.Split(strings.TrimSuffix(text, "\n"), "\n") {
		if !contractLine.MatchString(line) {
			t.Fatalf("line %q is not the contract's form", line)
		}
	}
	if text != "chart-ratchet: parses-scored 746\nchart-ratchet: set-reuse na\n" {
		t.Fatalf("Format = %q", text)
	}
	got, err := Parse(text)
	if err != nil || len(got) != 2 || got[0] != m[0] || got[1] != m[1] {
		t.Fatalf("Parse(Format(m)) = %v, %v", got, err)
	}
	if _, err := Parse("chart-ratchet: parses-scored\n"); err == nil {
		t.Fatal("Parse accepted a line with no value")
	}
	if _, err := Parse("chart-ratchet: parses-scored many\n"); err == nil {
		t.Fatal("Parse accepted a value that is neither a number nor na")
	}
	if _, err := Parse("parses-scored 746\n"); err == nil {
		t.Fatal("Parse accepted a line without the chart-ratchet: prefix")
	}
}
