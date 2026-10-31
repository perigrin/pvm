// ABOUTME: The chart ratchet: measures the chart spike against recursive descent and prints the contract lines.
// ABOUTME: Run by `make chart-ratchet`; exits 1 on a regression against testdata/ratchet/baseline.txt, 2 on an error.

// Command ratchet is the chart spike's ratchet. Its output contract is in
// internal/chart/README.md. It reads only the working tree, from the
// repository root: no commit, no refs/zhi, no network.
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
)

var (
	testdata     = filepath.Join("internal", "chart", "testdata")
	baselineFile = filepath.Join(testdata, "ratchet", "baseline.txt")
	knownFile    = filepath.Join(testdata, "lex-unmapped-known.txt")
)

// keys is spec §9.5's key set in its fixed print order. It was fixed at the
// reset: no key is added, renamed or retired afterwards.
var keys = []string{
	"tables-drift", "leo-disagree", "ledger-gaps", "rd-accepted-chart-refused", "chart-accepted-both-refused",
	"lex-unmapped", "tdt-aligned", "tdt-mismatch", "tdt-unseen", "terminals-per-token",
	"tdt-unseen-unlisted", "tdt-unaligned-unlisted", "lex-overhead", "lex-overhead-std5k",
	"speed-ratio-std5k-longlived", "speed-ratio-std5k-cold", "speed-ratio-aggregate-longlived", "speed-ratio-aggregate-cold",
	"alloc-per-token-chart", "alloc-per-token-rd", "items-per-token", "set-reuse", "ledger-excluded",
	"superset-population", "superset-recognized", "rd-accepted-perl-refused", "chart-accepted-perl-refused",
}

func main() {
	measured, unmapped, err := measure()
	if err != nil {
		fmt.Fprintf(os.Stderr, "chart-ratchet: %v\n", err)
		os.Exit(2)
	}
	ms := report(measured)
	fmt.Print(Format(ms))

	baseline, err := os.ReadFile(baselineFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "chart-ratchet: reading the baseline: %v\n", err)
		os.Exit(2)
	}
	known, err := os.ReadFile(knownFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "chart-ratchet: %v\n", err)
		os.Exit(2)
	}
	os.Exit(check(ms, unmapped, string(baseline), string(known), os.Stderr))
}

// check compares one run with the baseline text and returns the exit code:
// 0, 1 on a regression, 2 when the baseline is malformed or its key set is
// not exactly the printed one.
func check(ms []Measure, unmapped []string, baselineText, known string, stderr io.Writer) int {
	baseline, err := Parse(baselineText)
	if err == nil {
		err = checkBaseline(baseline)
	}
	if err != nil {
		fmt.Fprintf(stderr, "chart-ratchet: %s: %v\n", baselineFile, err)
		return 2
	}
	regressions := Compare(baseline, ms, unmapped, known)
	for _, r := range regressions {
		fmt.Fprintf(stderr, "chart-ratchet: %s\n", r)
	}
	if len(regressions) > 0 {
		return 1
	}
	return 0
}

// report lays measured values out in the fixed order, na for every key
// nothing measured yet. A measured key with no slot is a bug, so it panics.
func report(measured map[string]string) []Measure {
	for k := range measured {
		if !slices.Contains(keys, k) {
			panic("chart-ratchet: measured key " + k + " has no place in the print order")
		}
	}
	ms := make([]Measure, len(keys))
	for i, k := range keys {
		v, ok := measured[k]
		if !ok {
			v = na
		}
		ms[i] = Measure{k, v}
	}
	return ms
}

// measure takes every measurement there is a component for, and returns
// lex-unmapped's files as "<path> <byte>" pairs. At the reset nothing is
// measured: each later issue makes its keys print a number.
func measure() (map[string]string, []string, error) {
	return nil, nil, nil
}
