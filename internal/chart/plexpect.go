// ABOUTME: Compares expectation classes with spec 03's eleven PL_expect states, through testdata/pl-expect-map.txt.
// ABOUTME: A class is Expected(pos), the union over a position's Earley set, projected onto the conflicting pairs.

package chart

import (
	"errors"
	"fmt"
	"maps"
	"path/filepath"
	"slices"
	"strings"

	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/chart/recognize"
)

// PLExpectMapFile is the map from each PL_expect state to the conflicts it
// settles, relative to this package.
var PLExpectMapFile = filepath.Join("testdata", "pl-expect-map.txt")

// PLExpectStates is spec 03 §3.1.2's eleven states, in perl.h order.
var PLExpectStates = []string{
	"XOPERATOR", "XTERM", "XREF", "XSTATE", "XBLOCK", "XATTRBLOCK",
	"XATTRTERM", "XTERMBLOCK", "XBLOCKTERM", "XPOSTDEREF", "XTERMORDORDOR",
}

// lateJoiners are the conflicts the generator emits only once the grammar
// makes them conflict (spec §5): { as block or hash, ? as ternary or ?PATTERN?.
var lateJoiners = []string{"{", "?"}

// ReadPLExpectMap reads the map named name: one state per line, its name
// then the texts of the conflicts it settles; # lines and blank lines are
// skipped. Every state appears exactly once, with at least one conflict, and
// every conflict is one the generator emits in conflicts or a late joiner.
// Each error names the map and the state at fault.
func ReadPLExpectMap(name string, text []byte, conflicts []grammar.Conflict) (map[string][]string, error) {
	known := slices.Clone(lateJoiners)
	for _, c := range conflicts {
		known = append(known, c.Text)
	}
	m := map[string][]string{}
	var errs []error
	for i, line := range strings.Split(string(text), "\n") {
		f := strings.Fields(line)
		if len(f) == 0 || strings.HasPrefix(f[0], "#") {
			continue
		}
		state, settles := f[0], f[1:]
		switch _, dup := m[state]; {
		case !slices.Contains(PLExpectStates, state):
			errs = append(errs, fmt.Errorf("%s:%d: %s is not a PL_expect state", name, i+1, state))
			continue
		case dup:
			errs = append(errs, fmt.Errorf("%s:%d: PL_expect state %s is a duplicate", name, i+1, state))
			continue
		case len(settles) == 0:
			errs = append(errs, fmt.Errorf("%s:%d: PL_expect state %s names no conflict", name, i+1, state))
		}
		for _, c := range settles {
			if !slices.Contains(known, c) {
				errs = append(errs, fmt.Errorf("%s:%d: PL_expect state %s names %q, a conflict the generator cannot emit", name, i+1, state, c))
			}
		}
		m[state] = settles
	}
	for _, s := range PLExpectStates {
		if _, ok := m[s]; !ok {
			errs = append(errs, fmt.Errorf("%s: PL_expect state %s is missing", name, s))
		}
	}
	return m, errors.Join(errs...)
}

// Settled is the texts of the conflicts an expected-terminal set settles,
// sorted. A pair is settled when expected holds a terminal of exactly one of
// its two readings; a text is settled when every pair named by it is.
func Settled(expected []string, conflicts []grammar.Conflict) []string {
	holds := func(terms string) bool {
		return slices.ContainsFunc(strings.Split(terms, ","), func(t string) bool { return slices.Contains(expected, t) })
	}
	open := map[string]bool{}
	for _, c := range conflicts {
		open[c.Text] = open[c.Text] || holds(c.Operator) == holds(c.Term)
	}
	var out []string
	for text, o := range open {
		if !o {
			out = append(out, text)
		}
	}
	slices.Sort(out)
	return out
}

// Class is the expectation class at node pos of c: Expected(pos), the union
// over every state in the node's Earley set, projected onto conflicts. It is
// never formed per state. Expected recloses a set lookahead pruned, so the
// class is the same with lookahead on or off.
func Class(c *recognize.Chart, pos int, conflicts []grammar.Conflict) ([]string, error) {
	exp, err := c.Expected(pos)
	if err != nil {
		return nil, err
	}
	return Settled(exp, conflicts), nil
}

// ClassKey is a class as a set key.
func ClassKey(settled []string) string { return strings.Join(settled, " ") }

// Unreproduced compares the map with the classes seen, keyed by ClassKey.
// A state's conflicts in force are those conflicts names; a state with none
// is out of force and not counted. A counted state is reproduced when some
// class settles exactly its conflicts in force. It returns the counted states
// not reproduced, sorted, and how many states were counted.
func Unreproduced(m map[string][]string, conflicts []grammar.Conflict, classes map[string]bool) (states []string, counted int) {
	inForce := map[string]bool{}
	for _, c := range conflicts {
		inForce[c.Text] = true
	}
	for _, state := range slices.Sorted(maps.Keys(m)) {
		var want []string
		for _, c := range m[state] {
			if inForce[c] && !slices.Contains(want, c) {
				want = append(want, c)
			}
		}
		if len(want) == 0 {
			continue
		}
		counted++
		slices.Sort(want)
		if !classes[ClassKey(want)] {
			states = append(states, state)
		}
	}
	return states, counted
}
