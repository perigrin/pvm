// ABOUTME: Tests that the ratchet prints exactly the reset key set once each, in order, na where nothing measured it.
// ABOUTME: Also that a bad baseline is an error (exit 2), never a regression (exit 1) or a silent pass.

package main

import (
	"io"
	"slices"
	"strings"
	"testing"
)

// resetKeys is spec §9.5's key set, fixed at the reset and never changed afterwards.
var resetKeys = []string{
	"tables-drift", "leo-disagree", "ledger-gaps", "rd-accepted-chart-refused", "chart-accepted-both-refused",
	"lex-unmapped", "tdt-aligned", "tdt-mismatch", "tdt-unseen", "terminals-per-token",
	"tdt-unseen-unlisted", "tdt-unaligned-unlisted", "lex-overhead", "lex-overhead-std5k",
	"speed-ratio-std5k-longlived", "speed-ratio-std5k-cold", "speed-ratio-aggregate-longlived", "speed-ratio-aggregate-cold",
	"alloc-per-token-chart", "alloc-per-token-rd", "items-per-token", "set-reuse", "ledger-excluded",
	"superset-population", "superset-recognized", "rd-accepted-perl-refused", "chart-accepted-perl-refused",
}

func TestResetKeySet(t *testing.T) {
	if !slices.Equal(keys, resetKeys) {
		t.Fatalf("keys = %q, want the reset's set in order %q", keys, resetKeys)
	}
	ms := report(map[string]string{"tdt-aligned": "12"})
	got := make([]string, len(ms))
	for i, m := range ms {
		got[i] = m.Key
		want := na
		if m.Key == "tdt-aligned" {
			want = "12"
		}
		if m.Value != want {
			t.Errorf("%s printed %s, want %s", m.Key, m.Value, want)
		}
	}
	if !slices.Equal(got, resetKeys) {
		t.Fatalf("report printed %q, want each reset key once, in order", got)
	}
}

// report must not invent a key: a measure with no slot in the order is a bug in the ratchet.
func TestReportRejectsUnknownKey(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("report accepted a key that is not in the fixed order")
		}
	}()
	report(map[string]string{"not-a-key": "1"})
}

func TestRatchetBaselineErrorsExit2(t *testing.T) {
	good := Format(report(nil))
	if code := check(report(nil), nil, good, "", io.Discard); code != 0 {
		t.Fatalf("a well-formed all-na baseline exits %d, want 0", code)
	}
	for name, baseline := range map[string]string{
		"malformed line": good + "chart-ratchet: tdt-aligned\n",
		"bad value":      strings.Replace(good, "tdt-aligned na", "tdt-aligned many", 1),
		"retired key":    good + "chart-ratchet: rejects-disagree 0\n",
		"omitted key":    strings.Replace(good, "chart-ratchet: tdt-aligned na\n", "", 1),
		"key twice":      good + "chart-ratchet: tdt-aligned na\n",
	} {
		var stderr strings.Builder
		if code := check(report(nil), nil, baseline, "", &stderr); code != 2 {
			t.Errorf("%s: exit %d, want 2 (stderr %q)", name, code, stderr.String())
		}
	}
	// A regression is exit 1, which the baseline errors above are not.
	if code := check(report(map[string]string{"tdt-aligned": "1"}), nil, good, "", io.Discard); code != 1 {
		t.Errorf("na -> a number without a baseline exits %d, want 1", code)
	}
}
