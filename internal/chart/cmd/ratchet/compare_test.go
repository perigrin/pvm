// ABOUTME: Tests the chart ratchet's comparator and line printer against hand-built baselines.
// ABOUTME: Each reset rule must fire on a perturbed run and stay quiet on an equal one.

package main

import (
	"regexp"
	"slices"
	"strings"
	"testing"
)

func TestCompareQuietOnEqualBaseline(t *testing.T) {
	m := []Measure{{"tables-drift", "0"}, {"tdt-aligned", "12"}, {"set-reuse", "na"}, {"lex-overhead", "1.1"}}
	if got := Compare(m, m, nil, ""); len(got) != 0 {
		t.Fatalf("Compare on an equal baseline = %q, want none", got)
	}
}

// compareOne compares a single key, with ledger-gaps added to the run when
// ledgerGaps is not empty.
func compareOne(key, baseline, measured, ledgerGaps string) []string {
	base, run := []Measure{{key, baseline}}, []Measure{{key, measured}}
	if ledgerGaps != "" {
		base = append(base, Measure{"ledger-gaps", ledgerGaps})
		run = append(run, Measure{"ledger-gaps", ledgerGaps})
	}
	return Compare(base, run, nil, "")
}

func TestCompareResetRules(t *testing.T) {
	for _, tc := range []struct {
		name                     string
		key, base, measure, gaps string
		regress                  bool
	}{
		{"tables-drift rises", "tables-drift", "0", "1", "", true},
		{"leo-disagree rises", "leo-disagree", "0", "1", "", true},
		{"ledger-gaps rises", "ledger-gaps", "40", "41", "", true},
		{"ledger-gaps falls", "ledger-gaps", "40", "39", "", false},
		{"rd-accepted-chart-refused rises", "rd-accepted-chart-refused", "2", "3", "", true},
		{"chart-accepted-both-refused rises", "chart-accepted-both-refused", "2", "3", "", true},
		{"lex-unmapped rises", "lex-unmapped", "2", "3", "", true},
		{"lex-unmapped falls", "lex-unmapped", "2", "1", "", false},
		{"tdt-aligned falls", "tdt-aligned", "100", "99", "", true},
		{"tdt-aligned rises", "tdt-aligned", "100", "101", "", false},
		{"lex-overhead above 1.3", "lex-overhead", "1.0", "1.4", "", true},
		{"lex-overhead moves within 1.3", "lex-overhead", "1.0", "1.25", "", false},
		{"lex-overhead-std5k is timing", "lex-overhead-std5k", "1.0", "2.5", "", false},
		{"speed ratio is timing", "speed-ratio-std5k-longlived", "1.0", "9.0", "", false},
		{"alloc per token is timing", "alloc-per-token-chart", "1.0", "9.0", "", false},
		{"number becomes na", "items-per-token", "1.2", "na", "", true},
		{"reported key moves", "items-per-token", "1.2", "9.9", "", false},
		{"tdt-mismatch before completion", "tdt-mismatch", "4", "9", "3", false},
		{"tdt-mismatch at completion", "tdt-mismatch", "0", "1", "0", true},
		{"terminals-per-token rises before completion", "terminals-per-token", "2.0", "2.5", "3", false},
		{"terminals-per-token rises at completion", "terminals-per-token", "2.0", "2.5", "0", true},
		{"terminals-per-token falls at completion", "terminals-per-token", "2.0", "1.5", "0", false},
		{"tdt-unseen rises before completion", "tdt-unseen", "5", "6", "na", false},
		{"tdt-unseen rises at completion", "tdt-unseen", "5", "6", "0", true},
		{"tdt-unseen-unlisted before completion", "tdt-unseen-unlisted", "5", "5", "3", false},
		{"tdt-unseen-unlisted at completion", "tdt-unseen-unlisted", "1", "1", "0", true},
		{"tdt-unaligned-unlisted before completion", "tdt-unaligned-unlisted", "5", "5", "3", false},
		{"tdt-unaligned-unlisted at completion", "tdt-unaligned-unlisted", "1", "1", "0", true},
		{"tdt-unaligned-unlisted zero at completion", "tdt-unaligned-unlisted", "0", "0", "0", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := compareOne(tc.key, tc.base, tc.measure, tc.gaps)
			if (len(got) > 0) != tc.regress {
				t.Fatalf("Compare(%s %s -> %s, ledger-gaps %q) = %q, want regression %v", tc.key, tc.base, tc.measure, tc.gaps, got, tc.regress)
			}
			if tc.regress && !(strings.Contains(got[0], tc.key) && strings.Contains(got[0], tc.measure)) {
				t.Fatalf("regression %q does not name the key and the value", got[0])
			}
		})
	}
	// Q-list deletion: a rule exists only for a key the ratchet prints.
	for _, rules := range []map[string]bool{mayNotRise, mayNotFall, zeroWhenComplete, mayNotRiseWhenComplete} {
		for k := range rules {
			if !slices.Contains(keys, k) {
				t.Errorf("a rule names %s, which the ratchet does not print", k)
			}
		}
	}
}

// The lex-overhead gate is absolute: the baseline does not move it.
func TestCompareLexOverheadBoundary(t *testing.T) {
	for _, tc := range []struct {
		measure string
		regress bool
	}{{"1.31", true}, {"1.3", false}, {"na", false}} {
		base := "2.0"
		if tc.measure == "na" {
			base = "na"
		}
		got := compareOne("lex-overhead", base, tc.measure, "")
		if (len(got) > 0) != tc.regress {
			t.Errorf("lex-overhead %s against baseline %s = %q, want regression %v", tc.measure, base, got, tc.regress)
		}
		if tc.regress && len(got) > 0 && !strings.Contains(got[0], "1.3") {
			t.Errorf("regression %q does not name the 1.3 gate", got[0])
		}
	}
}

func TestCompareCompletionNeedsLedgerGapsZero(t *testing.T) {
	known := "a.pl 10 01a18582-65bf\n"
	run := func(gaps, mismatch string, unmapped []string) []string {
		base := []Measure{{"ledger-gaps", gaps}, {"tdt-mismatch", mismatch}}
		return Compare(base, base, unmapped, known)
	}
	for _, gaps := range []string{"na", "3"} {
		if got := run(gaps, "2", []string{"b.pl 7"}); len(got) != 0 {
			t.Errorf("ledger-gaps %s: tdt-mismatch 2 and an off-list file = %q, want none before completion", gaps, got)
		}
	}
	if got := run("0", "2", []string{"a.pl 10"}); len(got) != 1 || !strings.Contains(got[0], "tdt-mismatch") {
		t.Errorf("ledger-gaps 0: tdt-mismatch 2 = %q, want one tdt-mismatch regression", got)
	}
	if got := run("0", "0", []string{"a.pl 10", "b.pl 7"}); len(got) != 1 || !strings.Contains(got[0], "b.pl 7") {
		t.Errorf("ledger-gaps 0: an off-list file = %q, want one naming b.pl 7", got)
	}
	if got := run("0", "0", []string{"a.pl 11"}); len(got) != 2 {
		t.Errorf("ledger-gaps 0: a listed file refused at another byte = %q, want two mismatches", got)
	}
	if got := run("0", "0", []string{"a.pl 10"}); len(got) != 0 {
		t.Errorf("ledger-gaps 0: the listed file at the listed byte = %q, want none", got)
	}
	if got := run("0", "0", nil); len(got) != 1 || !strings.Contains(got[0], "no longer fails") {
		t.Errorf("ledger-gaps 0: a listed pair no file fails at = %q, want one naming it", got)
	}
}

func TestCompareNaToNumberNeedsBaseline(t *testing.T) {
	if got := compareOne("items-per-token", "na", "1.2", ""); len(got) != 1 || !strings.Contains(got[0], "baseline still na; take a baseline") {
		t.Errorf("na -> 1.2 = %q, want one regression saying to take a baseline", got)
	}
	if got := compareOne("items-per-token", "1.2", "1.2", ""); len(got) != 0 {
		t.Errorf("the same key with its baseline taken = %q, want none", got)
	}
	if got := compareOne("items-per-token", "na", "na", ""); len(got) != 0 {
		t.Errorf("na stays na = %q, want none", got)
	}
}

var contractLine = regexp.MustCompile(`^chart-ratchet: [a-z0-9-]+ [^ ]+$`)

func TestFormatAndParseRoundTrip(t *testing.T) {
	m := []Measure{{"tdt-aligned", "746"}, {"set-reuse", "na"}}
	text := Format(m)
	for _, line := range strings.Split(strings.TrimSuffix(text, "\n"), "\n") {
		if !contractLine.MatchString(line) {
			t.Fatalf("line %q is not the contract's form", line)
		}
	}
	if text != "chart-ratchet: tdt-aligned 746\nchart-ratchet: set-reuse na\n" {
		t.Fatalf("Format = %q", text)
	}
	got, err := Parse(text)
	if err != nil || len(got) != 2 || got[0] != m[0] || got[1] != m[1] {
		t.Fatalf("Parse(Format(m)) = %v, %v", got, err)
	}
	if _, err := Parse("chart-ratchet: tdt-aligned\n"); err == nil {
		t.Fatal("Parse accepted a line with no value")
	}
	if _, err := Parse("chart-ratchet: tdt-aligned many\n"); err == nil {
		t.Fatal("Parse accepted a value that is neither a number nor na")
	}
	for _, v := range []string{"NaN", "nan", "Inf", "+Inf", "-inf", "infinity"} {
		if _, err := Parse("chart-ratchet: set-reuse " + v + "\n"); err == nil {
			t.Fatalf("Parse accepted %s, which no comparison can order", v)
		}
	}
	if _, err := Parse("tdt-aligned 746\n"); err == nil {
		t.Fatal("Parse accepted a line without the chart-ratchet: prefix")
	}
}
