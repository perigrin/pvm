// ABOUTME: The corpus ratchet, keyed on a file's NAME so a tier move is not a regression.
// ABOUTME: Records which refusal each file has, and fails when one moves in either direction.

package conformance

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"tamarou.com/pvm/internal/parse"
)

// ratchetClean is the state of a file our parser handles: no refusal.
//
// A sentinel rather than the empty string, because the baseline is a text
// file and a blank field there is indistinguishable from a truncated
// line.
const ratchetClean = "-"

// ratchetPath is the committed baseline, beside this package.
const ratchetPath = "testdata/corpus.ratchet"

// tierFixture is the one file name the corpus repeats ON PURPOSE.
//
// The spec gives every tier an adjacency file -- "Adjacency files, one
// closing each tier" -- so fourteen tiers hold fourteen `00_adjacency.t`.
// That is not the collision AC4 guards against: those files are not one
// file at fourteen positions, they are fourteen different files that
// happen to share a role. Each belongs to its tier and is keyed with it.
//
// Any OTHER repeated name is a genuine collision and rejected.
const tierFixture = "adjacency.t"

// ratchetKey turns a corpus-relative path into the baseline's key.
//
// NAME-KEYED, and that is the whole design. The spec says it outright:
// "A file's identity is its NAME; the number is its current position, and
// positions move." Both numbers move -- the tier renumbers and the file
// renumbers with it -- so both are stripped, and the tier directory goes
// with them for every file but the per-tier fixture.
//
// Path-keyed, `04_operators/06_and_cliff.t` moving to `02_variables/` is
// one deletion plus one addition: a regression report for a file whose
// bytes did not change, produced by the `t/` sweep doing exactly what the
// spec built it to do.
func ratchetKey(rel string) string {
	rel = filepath.ToSlash(rel)
	tier, name, ok := strings.Cut(rel, "/")
	if !ok {
		return stripNumber(rel)
	}
	name = stripNumber(name)
	if name == tierFixture {
		return stripNumber(tier) + "/" + name
	}
	return name
}

// stripNumber removes a leading `NN_` position prefix.
//
// The prefix is DERIVED from the classification, so a reorder regenerates
// it. Keeping it in the key would make every renumbering read as churn.
func stripNumber(s string) string {
	if len(s) > 3 && s[2] == '_' &&
		s[0] >= '0' && s[0] <= '9' && s[1] >= '0' && s[1] <= '9' {
		return s[3:]
	}
	return s
}

// ratchetKeys rekeys a corpus-relative path -> refusal map by file name,
// rejecting a name two tiers both hold.
//
// The rejection is the key's own precondition made loud. Name-keying is
// sound only while names are unique; a silent collision records one file
// and stops measuring the other, which is the failure the corpus exists
// to prevent rather than to commit.
func ratchetKeys(state map[string]string) (map[string]string, error) {
	out := make(map[string]string, len(state))
	from := make(map[string]string, len(state))

	var collisions []string
	for _, rel := range sortedKeys(state) {
		key := ratchetKey(rel)
		if prev, ok := from[key]; ok {
			collisions = append(collisions, fmt.Sprintf(
				"%s is held by both %s and %s", key, prev, rel))
			continue
		}
		from[key] = rel
		out[key] = state[rel]
	}
	if len(collisions) > 0 {
		return nil, fmt.Errorf(
			"the corpus holds %d colliding file name(s), and the baseline "+
				"is keyed on NAMES:\n  %s\n\n"+
				"Rename one of each pair. A collision records one file and "+
				"silently stops measuring the other.",
			len(collisions), strings.Join(collisions, "\n  "))
	}
	return out, nil
}

// drift is what the baseline and the corpus disagree about.
//
// Four buckets, and the split is what lets AC3 hold: `arrived` is a name
// the baseline has never seen, which is a NOTE rather than a failure when
// the file refuses, because a corpus file is written BECAUSE something
// failed and requiring a second commit to bless it is the step that gets
// skipped. The other three are failures.
type drift struct {
	// changed: a file whose refusal is not the one recorded. Covers both
	// directions -- a file that stopped parsing and one that started --
	// because either means the baseline no longer describes the parser.
	changed []string

	// vanished: a name in the baseline with no file behind it. THE
	// ratchet's own contribution: verdict() runs per file and cannot
	// report on a file that is not there, so deleting a refusing file
	// deletes its check along with it.
	vanished []string

	// arrivedPassing: a name the baseline has never seen, already
	// parsing. Not free: a construct that works needs no new corpus file
	// in a refusing state, so the baseline should carry it before the
	// file lands.
	arrivedPassing []string

	// arrived: a new name that refuses. Reported, never failed.
	arrived []string
}

// clean reports whether the corpus still matches its baseline.
func (d drift) clean() bool {
	return len(d.changed) == 0 && len(d.vanished) == 0 &&
		len(d.arrivedPassing) == 0
}

// Notes renders what the drift observed without failing: new refusing
// files. A new file is accepted and never SILENT.
func (d drift) Notes() string {
	if len(d.arrived) == 0 {
		return ""
	}
	return fmt.Sprintf("%d new refusing file(s), accepted:\n  %s",
		len(d.arrived), strings.Join(d.arrived, "\n  "))
}

// String renders the failures.
func (d drift) String() string {
	var b strings.Builder
	if len(d.changed) > 0 {
		fmt.Fprintf(&b, "%d file(s) no longer match the baseline:\n  %s\n\n"+
			"This fails in BOTH directions. A file that started parsing is "+
			"good news and still fails: regenerate with "+
			"-conformance.update-ratchet and commit the baseline with the "+
			"change that earned it.\n",
			len(d.changed), strings.Join(d.changed, "\n  "))
	}
	if len(d.vanished) > 0 {
		fmt.Fprintf(&b, "%d file(s) are in the baseline and not the corpus:\n  %s\n\n"+
			"Leaving the denominator is a regression, not a cleanup.\n",
			len(d.vanished), strings.Join(d.vanished, "\n  "))
	}
	if len(d.arrivedPassing) > 0 {
		fmt.Fprintf(&b, "%d new file(s) already PASS:\n  %s\n\n"+
			"A new corpus file is written because something failed. One that "+
			"passes on arrival belongs in the baseline before it lands.\n",
			len(d.arrivedPassing), strings.Join(d.arrivedPassing, "\n  "))
	}
	return b.String()
}

// ratchetDrift compares a measurement against its baseline, both already
// keyed by ratchetKeys.
func ratchetDrift(base, now map[string]string) drift {
	var d drift
	for _, key := range sortedKeys(now) {
		was, ok := base[key]
		switch {
		case !ok && now[key] == ratchetClean:
			d.arrivedPassing = append(d.arrivedPassing, key)
		case !ok:
			d.arrived = append(d.arrived, fmt.Sprintf("%s: %s", key, now[key]))
		case was != now[key]:
			d.changed = append(d.changed,
				fmt.Sprintf("%s: %s -> %s", key, was, now[key]))
		}
	}
	for _, key := range sortedKeys(base) {
		if _, ok := now[key]; !ok {
			d.vanished = append(d.vanished, fmt.Sprintf("%s: %s", key, base[key]))
		}
	}
	return d
}

// sortedKeys gives a map's keys in order, so every message this file
// produces is stable between runs.
func sortedKeys(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// renderRatchet writes the baseline: one `state key` line per file,
// sorted, with a header naming what the states are and how to move them.
//
// The format follows internal/parse/testdata/t1.ratchet -- comment
// header, then one measurement per line, regenerated only under a flag.
// What differs is the measurement: a refusal CODE rather than a count,
// because a count cannot tell a refusal that changed cause from one that
// did not, and the cause is what a corpus file documents.
func renderRatchet(state map[string]string) string {
	keys := sortedKeys(state)
	clean := 0
	for _, k := range keys {
		if state[k] == ratchetClean {
			clean++
		}
	}

	var b strings.Builder
	b.WriteString("# The refusal each conformance corpus file has, as our parser\n")
	b.WriteString("# produces it. `-` is a file that parses clean; anything else is\n")
	b.WriteString("# the distinct refusal codes it produced, in tree order.\n")
	b.WriteString("#\n")
	b.WriteString("# KEYED ON FILE NAME, not path, and the numeric prefixes are\n")
	b.WriteString("# stripped. A file moving between tiers or renumbering within one\n")
	b.WriteString("# is the `t/` sweep working, and must not read as a regression.\n")
	b.WriteString("# The per-tier adjacency fixture carries its tier, since fourteen\n")
	b.WriteString("# tiers each hold one and they are different files.\n")
	b.WriteString("#\n")
	fmt.Fprintf(&b, "# %d files, %d clean, %d refusing.\n",
		len(keys), clean, len(keys)-clean)
	b.WriteString("#\n")
	b.WriteString("# Regenerate with -conformance.update-ratchet and commit the result\n")
	b.WriteString("# WITH the change that moved it. A baseline updated on its own is a\n")
	b.WriteString("# number nobody can attribute.\n")
	for _, k := range keys {
		b.WriteString(state[k])
		b.WriteByte(' ')
		b.WriteString(k)
		b.WriteByte('\n')
	}
	return b.String()
}

// parseRatchetFile reads a baseline written by renderRatchet.
func parseRatchetFile(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading the baseline at %s: %w", path, err)
	}
	out := make(map[string]string)
	for i, line := range strings.Split(string(data), "\n") {
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		state, key, ok := strings.Cut(line, " ")
		if !ok {
			return nil, fmt.Errorf("%s:%d: no space in %q", path, i+1, line)
		}
		out[key] = state
	}
	return out, nil
}

// corpusState measures every corpus file: what our parser does with it.
//
// Our parser only. perl does not run here, which is deliberate rather
// than a shortcut: perl adjudicates the corpus in verdict(), and a
// ratchet that re-ran it would double a thirty-second suite to re-derive
// a fact the runner already checks. What the baseline guards is OUR
// answer, and that needs no subprocess.
func corpusState(corpus string) (map[string]string, error) {
	paths, err := filepath.Glob(filepath.Join(corpus, "*_*", "*.t"))
	if err != nil {
		return nil, err
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("no corpus files under %s", corpus)
	}

	state := make(map[string]string, len(paths))
	for _, path := range paths {
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		f, err := ParseFile(string(raw))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		rel := filepath.ToSlash(strings.TrimPrefix(path,
			corpus+string(filepath.Separator)))
		state[rel] = refusalState(f)
	}
	return state, nil
}

// refusalState renders one file's refusal for the baseline: the distinct
// parser refusal codes, plus `tokens` when a declared token fact fails.
//
// The token fact is not an afterthought. Measured over the corpus at
// 6036de4e, FOUR of the fourteen files marked `STATUS refuses` produce no
// Unknown at all -- `06_and_cliff.t` and its neighbours refuse on the
// token claim, because our lexer gives `&&` a kind the file says is
// wrong while the parser reads the expression perfectly well. A baseline
// built from Unknowns alone would call those four clean and stop
// measuring the very gap they were written to name.
//
// A file the corpus expects perl to REJECT is not parsed here -- the
// runner never asks our parser to -- so it has no refusal to record, and
// recording one would invent a measurement.
func refusalState(f *File) string {
	if !f.ExpectParses {
		return ratchetClean
	}

	var seen []string
	for _, c := range refusalCodes(parse.Parse([]byte(f.Source))) {
		name := string(c)
		if name == "" {
			name = "(no code)"
		}
		if !slices.Contains(seen, name) {
			seen = append(seen, name)
		}
	}

	facts := &recorder{}
	for _, fact := range f.TokenFacts {
		checkTokenFact(facts, fact, []byte(f.Source))
	}
	if len(facts.msgs) > 0 {
		seen = append(seen, "tokens")
	}

	if len(seen) == 0 {
		return ratchetClean
	}
	return strings.Join(seen, ",")
}
