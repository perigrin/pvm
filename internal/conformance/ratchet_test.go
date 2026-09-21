// ABOUTME: The corpus ratchet's tests: name-keyed so a tier move is not a regression.
// ABOUTME: Includes the duplicate-name rejection the key's uniqueness assumption needs.

package conformance

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var updateRatchet = flag.Bool("conformance.update-ratchet", false,
	"rewrite the committed per-file refusal baseline")

// TestCorpusRatchet is the gate: the corpus's refusals, as measured,
// against the committed baseline.
//
// Everything below this is a test OF the mechanism; this is the
// mechanism pointed at the corpus. Without it the rest is a well-tested
// library nothing calls.
func TestCorpusRatchet(t *testing.T) {
	state, err := corpusState(corpusDir)
	if err != nil {
		t.Fatal(err)
	}
	now, err := ratchetKeys(state)
	if err != nil {
		t.Fatal(err)
	}

	if *updateRatchet {
		if err := os.MkdirAll(filepath.Dir(ratchetPath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(ratchetPath, []byte(renderRatchet(now)), 0o644); err != nil {
			t.Fatal(err)
		}
		t.Logf("rewrote %s (%d files); commit it with the change that moved it",
			ratchetPath, len(now))
		return
	}

	base, err := parseRatchetFile(ratchetPath)
	if err != nil {
		t.Fatalf("%v\nrun with -conformance.update-ratchet to create it", err)
	}

	d := ratchetDrift(base, now)
	if !d.clean() {
		t.Errorf("%s", d)
	}
	if notes := d.Notes(); notes != "" {
		t.Logf("%s", notes)
	}
}

// TestRatchetSeesTokenFactRefusals keeps the baseline measuring what the
// corpus measures.
//
// Four of the fourteen files marked `STATUS refuses` at 6036de4e produce
// no Unknown node: their refusal is the TOKEN claim, not the parse. A
// baseline built from Unknowns alone records those as clean and stops
// measuring the gap they were written to name -- a ratchet silently
// blind to 29% of the corpus's known refusals.
func TestRatchetSeesTokenFactRefusals(t *testing.T) {
	state, err := corpusState(corpusDir)
	if err != nil {
		t.Fatal(err)
	}

	refusing := 0
	for _, s := range state {
		if s != ratchetClean {
			refusing++
		}
	}

	// Every file the corpus marks `STATUS refuses` must be non-clean in
	// the baseline. Counting is not enough: name the ones that are not.
	paths, err := filepath.Glob(filepath.Join(corpusDir, "*_*", "*.t"))
	if err != nil {
		t.Fatal(err)
	}
	var blind []string
	for _, p := range paths {
		raw, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		f, err := ParseFile(string(raw))
		if err != nil {
			t.Fatal(err)
		}
		rel := filepath.ToSlash(strings.TrimPrefix(p,
			corpusDir+string(filepath.Separator)))
		if f.Refuses != "" && state[rel] == ratchetClean {
			blind = append(blind, rel)
		}
	}
	if len(blind) > 0 {
		t.Errorf("%d file(s) are marked `STATUS refuses` and the baseline "+
			"records them as clean:\n  %s\n\n"+
			"The ratchet is blind to their refusal.",
			len(blind), strings.Join(blind, "\n  "))
	}

	if refusing < 14 {
		t.Errorf("the baseline records %d refusing files; the corpus marks "+
			"14. A baseline that sees fewer refusals than the corpus "+
			"declares is not measuring them.", refusing)
	}
}

// TestTierMoveIsNotARegression is the reason the key is a NAME.
//
// A file moving from `04_operators/` to `02_variables/` because the `t/`
// sweep proved it belongs earlier is the sweep WORKING. Path-keyed, that
// reads as one deletion plus one addition -- a regression plus an
// unexplained new entry -- for a file whose bytes did not change.
func TestTierMoveIsNotARegression(t *testing.T) {
	before := map[string]string{"04_operators/06_and_cliff.t": "missing_operand"}
	after := map[string]string{"02_variables/06_and_cliff.t": "missing_operand"}

	base, err := ratchetKeys(before)
	if err != nil {
		t.Fatal(err)
	}
	now, err := ratchetKeys(after)
	if err != nil {
		t.Fatal(err)
	}

	if d := ratchetDrift(base, now); !d.clean() {
		t.Errorf("a tier move reported drift:\n%s", d)
	}

	// The converse, without which this passes against a ratchet that
	// never reports anything: a genuine change in the same move is not
	// silent.
	changed, err := ratchetKeys(map[string]string{
		"02_variables/06_and_cliff.t": "not_a_term"})
	if err != nil {
		t.Fatal(err)
	}
	if d := ratchetDrift(base, changed); d.clean() {
		t.Error("a moved file whose refusal CHANGED reported no drift")
	}
}

// TestRatchetFailsBothDirections is the ratchet's whole point.
//
// What this adds beyond verdict()'s per-file `staleMarker` check is
// AGGREGATE: staleMarker sees one file at a time and only a file that
// carries a `STATUS refuses` header. It cannot see a file that VANISHED
// from the corpus -- deleting the file deletes its check with it -- and
// it cannot see a refusal whose CAUSE changed on a file whose header
// names no code. The baseline sees both, because it is a list of names
// that must all still be there, each with the refusal it had.
func TestRatchetFailsBothDirections(t *testing.T) {
	base, err := ratchetKeys(map[string]string{
		"01_literals/03_leading_decimal.t": "not_a_term",
		"01_literals/02_decimal.t":         ratchetClean,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Direction 1: a file that STOPPED parsing.
	worse, err := ratchetKeys(map[string]string{
		"01_literals/03_leading_decimal.t": "not_a_term",
		"01_literals/02_decimal.t":         "missing_operand",
	})
	if err != nil {
		t.Fatal(err)
	}
	d := ratchetDrift(base, worse)
	if d.clean() {
		t.Error("a file that stopped parsing reported no drift")
	}
	if !strings.Contains(d.String(), "decimal.t") {
		t.Errorf("drift does not name the file that broke:\n%s", d)
	}

	// Direction 2: a file that STARTED parsing while still baselined as
	// refusing. Good news, and it still fails: the baseline must be
	// regenerated by the change that earned it.
	better, err := ratchetKeys(map[string]string{
		"01_literals/03_leading_decimal.t": ratchetClean,
		"01_literals/02_decimal.t":         ratchetClean,
	})
	if err != nil {
		t.Fatal(err)
	}
	if d := ratchetDrift(base, better); d.clean() {
		t.Error("a file that started parsing reported no drift")
	}

	// Direction 3, the one no per-file check can reach: the file is GONE.
	// verdict() cannot report on a file that is not there.
	gone, err := ratchetKeys(map[string]string{
		"01_literals/02_decimal.t": ratchetClean,
	})
	if err != nil {
		t.Fatal(err)
	}
	d = ratchetDrift(base, gone)
	if d.clean() {
		t.Error("a file that vanished from the corpus reported no drift")
	}
	if !strings.Contains(d.String(), "leading_decimal.t") {
		t.Errorf("drift does not name the vanished file:\n%s", d)
	}
}

// TestNewFileNeedsNoBaselineEdit keeps adding a corpus file a ONE-step
// act.
//
// A corpus file is written because something failed. If arriving in a
// refusing state also failed the build, every addition would be a
// two-step dance -- add the file, regenerate the baseline -- and the
// regeneration is the step that gets skipped. So a name the baseline has
// never seen, refusing, is accepted.
//
// It is not invisible: the drift REPORTS it in its notes, and a new file
// that PASSES still fails, because a construct that already works needs
// no corpus file written in a refusing state and the baseline should say
// so before the file lands.
func TestNewFileNeedsNoBaselineEdit(t *testing.T) {
	base, err := ratchetKeys(map[string]string{
		"01_literals/02_decimal.t": ratchetClean,
	})
	if err != nil {
		t.Fatal(err)
	}

	withNew, err := ratchetKeys(map[string]string{
		"01_literals/02_decimal.t":        ratchetClean,
		"01_literals/05_hex_underscore.t": "not_a_term",
	})
	if err != nil {
		t.Fatal(err)
	}
	d := ratchetDrift(base, withNew)
	if !d.clean() {
		t.Errorf("a new refusing file failed the baseline:\n%s", d)
	}
	if !strings.Contains(d.Notes(), "hex_underscore.t") {
		t.Errorf("a new file was accepted SILENTLY; drift notes are %q", d.Notes())
	}

	// The converse: a new file that already passes is not free. Otherwise
	// this test passes against a ratchet that accepts every new name.
	withPassing, err := ratchetKeys(map[string]string{
		"01_literals/02_decimal.t":        ratchetClean,
		"01_literals/05_hex_underscore.t": ratchetClean,
	})
	if err != nil {
		t.Fatal(err)
	}
	if d := ratchetDrift(base, withPassing); d.clean() {
		t.Error("a new file that already PASSES was accepted without a baseline entry")
	}
}

// TestDuplicateFileNameRejected defends the key's own assumption.
//
// Name-keying is sound only while names are unique. Two tiers holding
// `03_slice.t` collide into one key, and the baseline then records one of
// them and silently stops measuring the other -- the exact failure mode
// the corpus exists to prevent.
func TestDuplicateFileNameRejected(t *testing.T) {
	_, err := ratchetKeys(map[string]string{
		"01_literals/03_slice.t":  ratchetClean,
		"02_variables/03_slice.t": "not_a_term",
	})
	if err == nil {
		t.Fatal("two tiers holding `03_slice.t` were accepted")
	}
	for _, want := range []string{"slice.t", "01_literals", "02_variables"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not name %q", err, want)
		}
	}

	// Two distinct names in one tier are not a collision.
	if _, err := ratchetKeys(map[string]string{
		"01_literals/03_slice.t": ratchetClean,
		"01_literals/07_dice.t":  ratchetClean,
	}); err != nil {
		t.Errorf("two distinct names in one tier were rejected: %v", err)
	}

	// The per-tier adjacency fixture is the corpus's ONE deliberate
	// repetition and must survive: 14 tiers each close with one.
	if _, err := ratchetKeys(map[string]string{
		"01_literals/00_adjacency.t":  ratchetClean,
		"04_operators/00_adjacency.t": "missing_operand",
	}); err != nil {
		t.Errorf("the per-tier adjacency fixture was rejected as a duplicate: %v", err)
	}

	// And the real corpus on disk has no duplicate, which is the
	// observation rather than the assertion.
	paths, err := filepath.Glob(filepath.Join(corpusDir, "*_*", "*.t"))
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) == 0 {
		t.Fatal("no corpus files found")
	}
	state := make(map[string]string, len(paths))
	for _, p := range paths {
		state[filepath.ToSlash(strings.TrimPrefix(p,
			corpusDir+string(filepath.Separator)))] = ratchetClean
	}
	if _, err := ratchetKeys(state); err != nil {
		t.Errorf("the corpus on disk has a colliding name: %v", err)
	}
}
