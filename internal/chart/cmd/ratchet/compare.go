// ABOUTME: The chart ratchet's output contract: the `chart-ratchet: <key> <value>` line form and the regression rules.
// ABOUTME: Compare checks a run against the committed baseline key by key; only a regression fails the target.

package main

import (
	"fmt"
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

// mayNotRise are the exit counts, the sweep keys among them.
var mayNotRise = map[string]bool{
	"parses-refused": true, "parses-wrong-sites": true, "yield-violations": true,
	"oracle-below-rd": true, "std5k-refused": true, "parsent-accepted": true,
	"rejects-disagree": true, "lex-unmapped": true, "leo-disagree": true,
	"recovery-on-accepted": true,
}

// mayNotFall are the corpus sizes and the chart's side of defect yield.
var mayNotFall = map[string]bool{
	"parses-scored": true, "parsent-scored": true,
	"defect-yield-settled": true, "defect-yield-chart-exact": true,
}

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
		if _, err := strconv.ParseFloat(f[1], 64); err != nil && f[1] != na {
			return nil, fmt.Errorf("line %d: %s's value %q is neither a number nor %s", i+1, f[0], f[1], na)
		}
		ms = append(ms, Measure{f[0], f[1]})
	}
	return ms, nil
}

// Compare returns one message per regression of measured against baseline.
// A key may not appear in one and not the other: an added key joins the
// baseline in the commit that adds it, and a key is never renamed.
func Compare(baseline, measured []Measure) []string {
	base := map[string]string{}
	for _, m := range baseline {
		base[m.Key] = m.Value
	}
	var out []string
	regress := func(key, b, v, why string) {
		out = append(out, fmt.Sprintf("regression: %s: baseline %s, measured %s (%s)", key, b, v, why))
	}
	for _, m := range measured {
		b, ok := base[m.Key]
		delete(base, m.Key)
		switch {
		case !ok:
			regress(m.Key, "absent", m.Value, "every key the ratchet prints must be in baseline.txt")
		case b == na:
		case m.Value == na:
			regress(m.Key, b, m.Value, "a number may not become na")
		case mayNotRise[m.Key] && number(m.Value) > number(b):
			regress(m.Key, b, m.Value, "may not rise")
		case mayNotFall[m.Key] && number(m.Value) < number(b):
			regress(m.Key, b, m.Value, "may not fall")
		}
	}
	for _, m := range baseline {
		if _, gone := base[m.Key]; gone {
			regress(m.Key, m.Value, "absent", "a key is never removed or renamed")
		}
	}
	return out
}

// number reads a value Parse has already checked.
func number(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}
