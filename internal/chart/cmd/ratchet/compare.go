// ABOUTME: The chart ratchet's output contract: the `chart-ratchet: <key> <value>` line form and the regression rules.
// ABOUTME: Compare checks a run against the committed baseline key by key; only a regression fails the target.

package main

import (
	"fmt"
	"maps"
	"math"
	"slices"
	"strconv"
	"strings"
)

// Measure is one key and its value: a number, or "na" for a component not built yet.
type Measure struct {
	Key, Value string
}

const (
	linePrefix = "chart-ratchet: "
	na         = "na"
)

// mayNotRise are the gated counts.
var mayNotRise = map[string]bool{
	"tables-drift": true, "leo-disagree": true, "ledger-gaps": true,
	"rd-accepted-chart-refused": true, "chart-accepted-both-refused": true, "lex-unmapped": true,
}

var mayNotFall = map[string]bool{"tdt-aligned": true}

// From the port's completion (the same run's ledger-gaps is 0) these must be
// 0, and these may not rise.
var (
	zeroWhenComplete       = map[string]bool{"tdt-mismatch": true, "tdt-unseen-unlisted": true, "tdt-unaligned-unlisted": true}
	mayNotRiseWhenComplete = map[string]bool{"terminals-per-token": true, "tdt-unseen": true}
)

// lexOverheadMax gates lex-overhead absolutely, whatever its baseline.
const lexOverheadMax = 1.3

// Format renders measures in the contract's line form, one per line.
func Format(ms []Measure) string {
	var b strings.Builder
	for _, m := range ms {
		fmt.Fprintf(&b, "%s%s %s\n", linePrefix, m.Key, m.Value)
	}
	return b.String()
}

// Parse reads the line form back.
func Parse(text string) ([]Measure, error) {
	var ms []Measure
	for i, line := range strings.Split(strings.TrimSuffix(text, "\n"), "\n") {
		if line == "" {
			continue
		}
		f := strings.Fields(strings.TrimPrefix(line, linePrefix))
		if !strings.HasPrefix(line, linePrefix) || len(f) != 2 {
			return nil, fmt.Errorf("line %d is not %q<key> <value>: %q", i+1, linePrefix, line)
		}
		if v, err := strconv.ParseFloat(f[1], 64); f[1] != na && (err != nil || math.IsNaN(v) || math.IsInf(v, 0)) {
			return nil, fmt.Errorf("line %d: %s's value %q is neither a finite number nor %s", i+1, f[0], f[1], na)
		}
		ms = append(ms, Measure{f[0], f[1]})
	}
	return ms, nil
}

// Compare returns one message per regression of measured against baseline.
// Timing keys have no rule here, so they are never compared with the
// baseline; lex-overhead's gate is absolute. From the port's completion,
// unmapped (lex-unmapped's files as "<path> <byte>" pairs) must equal the
// first two fields of known's lines.
func Compare(baseline, measured []Measure, unmapped []string, known string) []string {
	base := map[string]string{}
	for _, m := range baseline {
		base[m.Key] = m.Value
	}
	complete := slices.ContainsFunc(measured, func(m Measure) bool {
		return m.Key == "ledger-gaps" && m.Value != na && number(m.Value) == 0
	})
	var out []string
	regress := func(key, b, v, why string) {
		out = append(out, fmt.Sprintf("regression: %s: baseline %s, measured %s (%s)", key, b, v, why))
	}
	for _, m := range measured {
		b, ok := base[m.Key]
		switch {
		case !ok, b == na && m.Value == na:
		case b == na:
			regress(m.Key, b, m.Value, "baseline still na; take a baseline")
		case m.Value == na:
			regress(m.Key, b, m.Value, "a number may not become na")
		case (mayNotRise[m.Key] || complete && mayNotRiseWhenComplete[m.Key]) && number(m.Value) > number(b):
			regress(m.Key, b, m.Value, "may not rise")
		case mayNotFall[m.Key] && number(m.Value) < number(b):
			regress(m.Key, b, m.Value, "may not fall")
		}
		if m.Key == "lex-overhead" && m.Value != na && number(m.Value) > lexOverheadMax {
			regress(m.Key, b, m.Value, fmt.Sprintf("above %g, whatever the baseline", lexOverheadMax))
		}
		if complete && zeroWhenComplete[m.Key] && (m.Value == na || number(m.Value) != 0) {
			regress(m.Key, b, m.Value, "must be 0 once ledger-gaps is 0")
		}
	}
	if complete {
		out = append(out, knownMismatch(unmapped, known)...)
	}
	return out
}

// knownMismatch holds lex-unmapped's "<path> <byte>" pairs to the known
// list, whose lines are `<path> <byte> <issue ref>`, as sets: each pair the
// list lacks, and each listed pair no file fails at.
func knownMismatch(unmapped []string, known string) []string {
	listed := map[string]bool{}
	for _, line := range strings.Split(known, "\n") {
		if f := strings.Fields(line); len(f) > 0 {
			listed[strings.Join(f[:min(2, len(f))], " ")] = true
		}
	}
	var out []string
	for _, u := range unmapped {
		if !listed[u] {
			out = append(out, fmt.Sprintf("lex-unmapped: %s fails but is not listed in lex-unmapped-known.txt", u))
		}
		delete(listed, u)
	}
	for _, l := range slices.Sorted(maps.Keys(listed)) {
		out = append(out, fmt.Sprintf("lex-unmapped: %s is listed in lex-unmapped-known.txt but no longer fails", l))
	}
	slices.Sort(out)
	return out
}

// checkBaseline is an error unless baseline holds every printed key exactly
// once and nothing else.
func checkBaseline(baseline []Measure) error {
	seen := map[string]bool{}
	for _, m := range baseline {
		if !slices.Contains(keys, m.Key) {
			return fmt.Errorf("%s is not a key the ratchet prints", m.Key)
		}
		if seen[m.Key] {
			return fmt.Errorf("%s appears twice", m.Key)
		}
		seen[m.Key] = true
	}
	for _, k := range keys {
		if !seen[k] {
			return fmt.Errorf("%s is missing", k)
		}
	}
	return nil
}

// number reads a value Parse has already checked.
func number(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}
