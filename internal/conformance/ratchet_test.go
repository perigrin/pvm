// ABOUTME: The corpus ratchet's tests: keyed on a case's name so a tier move is not a regression.
// ABOUTME: Includes the duplicate-key rejection the key's uniqueness assumption needs.

package conformance

import (
	"flag"
	"fmt"
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
		t.Logf("rewrote %s (%d cases); commit it with the change that moved it",
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

	// Every case the corpus marks `refuses:` must be non-clean in the
	// baseline. Counting is not enough: name the ones that are not.
	cases, err := AllCases(corpusDir)
	if err != nil {
		t.Fatal(err)
	}
	var blind []string
	declared := 0
	for _, c := range cases {
		if c.Refuses == "" {
			continue
		}
		declared++
		if state[c.Key] == ratchetClean {
			blind = append(blind, c.Key)
		}
	}
	if len(blind) > 0 {
		t.Errorf("%d case(s) are marked `refuses:` and the baseline "+
			"records them as clean:\n  %s\n\n"+
			"The ratchet is blind to their refusal.",
			len(blind), strings.Join(blind, "\n  "))
	}

	// DERIVED, not a constant. The old form pinned the number 14, which
	// was the count of `STATUS refuses` files on the day it was written
	// and was already wrong by the time the corpus reached 27 of them --
	// a floor that low stops being a floor. The claim was never about
	// that integer: it is that the baseline sees at least as many
	// refusals as the corpus DECLARES, and the corpus can be asked.
	if declared == 0 {
		t.Fatal("no case in the corpus is marked `refuses:`; this check " +
			"would pass over a corpus that declares nothing")
	}
	if refusing < declared {
		t.Errorf("the baseline records %d refusing cases; the corpus marks "+
			"%d. A baseline that sees fewer refusals than the corpus "+
			"declares is not measuring them.", refusing, declared)
	}
}

// TestRatchetKeysCarryNoTier pins why moving a case between tiers is not
// a regression.
//
// A ratchet key is `<topic>.md/<case title>`. Neither half names a tier,
// so re-declaring a topic's `**Tier NN name.**` line moves the case
// without renaming it, and the ratchet reports no drift.
//
// AT HEAD THIS WAS ACHIEVED RATHER THAN STRUCTURAL. Keys were file
// paths like `04_operators/06_and_cliff.t`, and `ratchetKey` stripped
// the tier directory and the numeric prefix to get the same property --
// so the test that guarded it built two files in different tiers and
// checked they produced one key.
//
// That form became TAUTOLOGICAL after the migration: it varied the tier
// line, which a key is no longer built from, so it could not fail
// whatever the keying code did. Replaced with the claim that can:
// no key mentions any tier, which fails the moment someone puts one
// back.
func TestRatchetKeysCarryNoTier(t *testing.T) {
	cases, err := AllCases(corpusDir)
	if err != nil {
		t.Fatalf("reading the corpus: %v", err)
	}

	tiers := map[string]bool{}
	for _, c := range cases {
		tiers[c.Tier] = true
	}
	if len(tiers) < 2 {
		t.Fatalf("corpus spans %d tier(s); this check needs several to "+
			"mean anything", len(tiers))
	}

	for _, c := range cases {
		// The adjacency topics are the deliberate exception, and they
		// are named `adjacency-<tier>.md` BECAUSE an adjacency case is
		// tier-specific by nature: its subject is "every construct THIS
		// tier introduces, each beside another". Such a case cannot
		// move tiers without becoming a different case, so a rename is
		// the honest report.
		if strings.HasPrefix(c.Topic, "adjacency-") {
			continue
		}
		for tier := range tiers {
			if strings.Contains(c.Key, tier) {
				t.Errorf("ratchet key %q names the tier %q.\n"+
					"\tA key that carries its tier turns MOVING a case "+
					"into a rename, and the ratchet reports a regression "+
					"for work that changed no behaviour.", c.Key, tier)
			}
		}
	}
}

// topicKeys reads a topic's cases and keys them at one refusal, so a
// test can state what it is about -- the topic's text -- rather than
// hand-building the keys the reader would have produced.
func topicKeys(topic, raw, state string) (map[string]string, error) {
	if _, err := topicTier(raw); err != nil {
		return nil, err
	}
	cases, err := ParseTopic(raw)
	if err != nil {
		return nil, err
	}
	byKey := map[string]string{}
	for _, c := range cases {
		byKey[topic+"/"+c.Title] = state
	}
	return ratchetKeys(byKey)
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
		"numeric-point.md/A leading decimal point": "not_a_term",
		"numeric-point.md/A plain decimal":         ratchetClean,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Direction 1: a file that STOPPED parsing.
	worse, err := ratchetKeys(map[string]string{
		"numeric-point.md/A leading decimal point": "not_a_term",
		"numeric-point.md/A plain decimal":         "missing_operand",
	})
	if err != nil {
		t.Fatal(err)
	}
	d := ratchetDrift(base, worse)
	if d.clean() {
		t.Error("a file that stopped parsing reported no drift")
	}
	if !strings.Contains(d.String(), "A plain decimal") {
		t.Errorf("drift does not name the case that broke:\n%s", d)
	}

	// Direction 2: a file that STARTED parsing while still baselined as
	// refusing. Good news, and it still fails: the baseline must be
	// regenerated by the change that earned it.
	better, err := ratchetKeys(map[string]string{
		"numeric-point.md/A leading decimal point": ratchetClean,
		"numeric-point.md/A plain decimal":         ratchetClean,
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
		"numeric-point.md/A plain decimal": ratchetClean,
	})
	if err != nil {
		t.Fatal(err)
	}
	d = ratchetDrift(base, gone)
	if d.clean() {
		t.Error("a file that vanished from the corpus reported no drift")
	}
	if !strings.Contains(d.String(), "A leading decimal point") {
		t.Errorf("drift does not name the vanished case:\n%s", d)
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
		"numeric-point.md/A plain decimal": ratchetClean,
	})
	if err != nil {
		t.Fatal(err)
	}

	withNew, err := ratchetKeys(map[string]string{
		"numeric-point.md/A plain decimal":      ratchetClean,
		"numeric-radix.md/Hex with underscores": "not_a_term",
	})
	if err != nil {
		t.Fatal(err)
	}
	d := ratchetDrift(base, withNew)
	if !d.clean() {
		t.Errorf("a new refusing file failed the baseline:\n%s", d)
	}
	if !strings.Contains(d.Notes(), "Hex with underscores") {
		t.Errorf("a new file was accepted SILENTLY; drift notes are %q", d.Notes())
	}

	// The converse: a new file that already passes is not free. Otherwise
	// this test passes against a ratchet that accepts every new name.
	withPassing, err := ratchetKeys(map[string]string{
		"numeric-point.md/A plain decimal":      ratchetClean,
		"numeric-radix.md/Hex with underscores": ratchetClean,
	})
	if err != nil {
		t.Fatal(err)
	}
	if d := ratchetDrift(base, withPassing); d.clean() {
		t.Error("a new file that already PASSES was accepted without a baseline entry")
	}
}

// TestDuplicateCaseKeyRejected defends the key's own assumption.
//
// Name-keying is sound only while names are unique. Two cases sharing a
// key collide into one entry, and the baseline then records one of them
// and silently stops measuring the other -- the exact failure mode the
// corpus exists to prevent.
//
// WHAT CHANGED WITH TOPICS: the collision this used to describe was two
// TIERS holding `03_slice.t`, which `ratchetKey` manufactured by
// stripping the tier out of the path. Nothing strips now, so that
// collision cannot occur and testing it would be testing nothing. The
// collision that CAN occur moved up a level: two `##` headings with the
// same title in one topic file produce the same `<topic>.md/<title>`,
// and the corpus already holds fourteen cases titled "The whole tier in
// one body" -- one retitle away from being a real one.
//
// It is also why this goes through `ratchetKeysOf` and `corpusState`
// rather than a map. A Go map cannot hold a duplicate key, so a test
// that built one would have lost the second case before the check could
// see it -- the bug performed rather than caught.
func TestDuplicateCaseKeyRejected(t *testing.T) {
	_, err := ratchetKeysOf([]keyed{
		{key: "slices.md/A slice", state: ratchetClean, where: "case 1 of slices.md"},
		{key: "slices.md/A slice", state: "not_a_term", where: "case 2 of slices.md"},
	})
	if err == nil {
		t.Fatal("two cases titled `A slice` in one topic were accepted")
	}
	for _, want := range []string{"A slice", "case 1 of slices.md", "case 2 of slices.md"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not name %q", err, want)
		}
	}

	// Two distinct titles in one topic are not a collision.
	if _, err := ratchetKeysOf([]keyed{
		{key: "slices.md/A slice", state: ratchetClean, where: "slices.md"},
		{key: "slices.md/A dice", state: ratchetClean, where: "slices.md"},
	}); err != nil {
		t.Errorf("two distinct titles in one topic were rejected: %v", err)
	}

	// The corpus's ONE deliberate repetition must survive: fourteen tiers
	// each close with an adjacency case and all fourteen carry the same
	// TITLE. They are kept apart by the topic holding them, which is what
	// used to need the `tierFixture` carve-out and now needs nothing.
	if _, err := ratchetKeysOf([]keyed{
		{key: "adjacency-01_literals.md/The whole tier in one body",
			state: ratchetClean, where: "adjacency-01_literals.md"},
		{key: "adjacency-04_operators.md/The whole tier in one body",
			state: "missing_operand", where: "adjacency-04_operators.md"},
	}); err != nil {
		t.Errorf("the per-tier adjacency fixture was rejected as a duplicate: %v", err)
	}

	// And the real corpus on disk has no duplicate, which is the
	// observation rather than the assertion. `corpusState` is what reads
	// it as a slice, so this is the path a real collision would take.
	if _, err := corpusState(corpusDir); err != nil {
		t.Errorf("the corpus on disk has a colliding case key: %v", err)
	}

	// The same-titled adjacency cases are not hypothetical: assert the
	// corpus holds them, so the carve-out above is tested against a
	// repetition that exists rather than one this file imagined.
	cases, err := AllCases(corpusDir)
	if err != nil {
		t.Fatal(err)
	}
	titles := map[string]int{}
	for _, c := range cases {
		if strings.HasPrefix(c.Topic, "adjacency-") {
			titles[strings.TrimPrefix(c.Key, c.Topic+"/")]++
		}
	}
	repeated := 0
	for _, n := range titles {
		if n > 1 {
			repeated++
		}
	}
	if repeated == 0 {
		t.Error("no adjacency title is shared between topics; the " +
			"deliberate repetition this test protects is not in the corpus")
	}
}

// TestRatchetHeaderAgreesWithItsBody pins the summary line against the case
// lines it summarises.
//
// `parseRatchetFile` skips every `#` line, so the header was DERIVED on write
// and IGNORED on read. Measured: the committed baseline's header was edited to
// "999 cases, 1 clean, 998 refusing" and TestCorpusRatchet still passed, in
// 0.04s. The one number every report of this corpus quotes was the one number
// nothing verified.
//
// That is a baseline RECALLING rather than RECOMPUTING: right when written,
// wrong the moment the thing it counts changes, and silent either way. The
// same failure as a hardcoded denominator in a census script.
//
// Counted from the body rather than re-measured from the corpus on purpose --
// re-measuring would test the corpus, which TestCorpusRatchet already does.
// What was unguarded is the header AGREEING with the body beside it.
func TestRatchetHeaderAgreesWithItsBody(t *testing.T) {
	data, err := os.ReadFile(ratchetPath)
	if err != nil {
		t.Fatalf("reading %s: %v", ratchetPath, err)
	}

	var header string
	cases, clean := 0, 0
	for _, line := range strings.Split(string(data), "\n") {
		switch {
		case line == "":
			continue
		case strings.HasPrefix(line, "# ") && strings.Contains(line, " cases, "):
			header = strings.TrimPrefix(line, "# ")
		case strings.HasPrefix(line, "#"):
			continue
		default:
			cases++
			// A clean case's state is the "-" placeholder; anything else is
			// a refusal code, which is what refusalState writes.
			if strings.HasPrefix(line, "- ") {
				clean++
			}
		}
	}

	if header == "" {
		t.Fatal("no `# N cases, M clean, K refusing.` line in the baseline")
	}

	want := fmt.Sprintf("%d cases, %d clean, %d refusing.", cases, clean, cases-clean)
	if header != want {
		t.Errorf("header says %q, body holds %q.\n"+
			"\tThe summary is derived on write and skipped on read, so it can\n"+
			"\tdrift silently. Regenerate with -conformance.update-ratchet.",
			header, want)
	}
}
