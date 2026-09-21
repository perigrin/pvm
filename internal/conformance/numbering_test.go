// ABOUTME: The corpus-wide numbering check: every tier declares how its files are ordered.
// ABOUTME: `derived` regenerates from names, `grouped` names the ranges, `accidental` admits neither.
package conformance

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// reFileOrder matches a tier README's machine-read block:
//
//	## FILE ORDER
//
//	    derived
//
// or, for a tier whose order was chosen rather than computed:
//
//	## FILE ORDER
//
//	    01-03	conditionals
//	    04-07	loops
//	    08-10	jumps
//
// Read exactly as INTRODUCES, DEPENDS ON and HARD MARKERS are: the FIRST
// indented block after the heading, so prose may follow inside the same
// section -- and for a grouped tier that prose is where the REASON lives,
// since no test can confirm that conditionals belong before loops.
//
// It lives in the README rather than in a Go table for the reason
// readTierOps gives: two lists that must agree are how drift starts.
var reFileOrder = regexp.MustCompile(`(?m)^## FILE ORDER\s*\n\s*\n((?:[ \t]+\S.*\n?)+)`)

// reGroupLine matches one grouping line: a range and the name of what the
// range holds.
var reGroupLine = regexp.MustCompile(`^(\d\d)-(\d\d)\s+(\S.*)$`)

// orderDerived is the word a tier uses when its numbers are a function of
// its names: sort the identities, count from 01. `regenerateNumbering`
// produces them, so the numbers could be deleted and rebuilt, and
// TestDerivedTierNumberingRegenerates verifies that they can.
const orderDerived = "derived"

// orderAccidental is the word a tier uses when its numbering is neither
// derived nor arranged -- the numbers are the order the files happened to
// be written in.
//
// This is a third state because the measurement found one, not because a
// third state is desirable. Running regenerateNumbering over all fourteen
// tiers found TWO derived, 01_literals and 02_variables. Reading the other
// twelve READMEs found exactly ONE that argues for its order -- 06_control,
// whose conditionals/loops/jumps runs the prose depends on -- and ELEVEN
// that mention their files only by name and never by position.
//
// Folding those eleven into `grouped` would mean writing eleven groupings
// nobody chose, and folding them into `derived` would mean renumbering
// eleven directories -- which is the right end state, since an accidental
// order is exactly the case where renumbering costs nothing, but it
// regenerates the ratchet and that is another issue's work. So the word
// exists to RECORD the debt rather than to bless it: a tier declaring it
// is saying its numbers mean nothing yet, which is a thing a reader can
// act on, unlike silence.
//
// It buys no check, and that is the point of its cost: it is the only
// spelling that asserts nothing, so a tier that could say something truer
// has a reason to.
const orderAccidental = "accidental"

// group is one declared run of construct files.
type group struct {
	lo, hi int
	name   string
}

// fileOrder is what one tier declares about its numbering: derived,
// accidental, or a list of groups saying what the arrangement is FOR.
type fileOrder struct {
	word   string // orderDerived, orderAccidental, or "" when grouped
	groups []group
}

// readFileOrders collects every tier's declared file ordering.
//
// Missing is an error rather than a default. A default would let a tier
// ship with no opinion at all, which is the state this check exists to
// end: before it, one tier was alphabetical on purpose, one was grouped on
// purpose, and twelve were neither on purpose or otherwise -- and nothing
// on disk distinguished the three.
//
// The three spellings are deliberately asymmetric in cost. `derived` is
// one word and buys a machine check. A grouping must NAME the ranges,
// which a tier whose order is an accident cannot write down without the
// falsehood being visible to a reader: there is no pair of ranges that
// describes `01_backslash_scalar.t` through `11_ref_builtin.t` as an
// arrangement, because nobody arranged them. `accidental` is one word and
// buys nothing, which is what makes it a record of debt rather than a
// rubber stamp -- a tier has an incentive to say something truer.
func readFileOrders(corpus string) (map[string]fileOrder, error) {
	entries, err := os.ReadDir(corpus)
	if err != nil {
		return nil, fmt.Errorf("reading the corpus root: %w", err)
	}

	out := map[string]fileOrder{}
	for _, e := range entries {
		if !e.IsDir() || !isTierDir(e.Name()) {
			continue
		}
		path := filepath.Join(corpus, e.Name(), "README.md")
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("tier %s has no README: %w", e.Name(), err)
		}
		m := reFileOrder.FindSubmatch(raw)
		if m == nil {
			return nil, fmt.Errorf("%s: no `## FILE ORDER` block", path)
		}
		order, err := parseFileOrder(string(m[1]))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		out[e.Name()] = order
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("%s holds no tiers", corpus)
	}
	return out, nil
}

// parseFileOrder reads one FILE ORDER block's body.
func parseFileOrder(body string) (fileOrder, error) {
	var lines []string
	for _, l := range strings.Split(body, "\n") {
		if l = strings.TrimSpace(l); l != "" {
			lines = append(lines, l)
		}
	}
	if len(lines) == 0 {
		return fileOrder{}, fmt.Errorf("the FILE ORDER block is empty")
	}

	if len(lines) == 1 && (lines[0] == orderDerived || lines[0] == orderAccidental) {
		return fileOrder{word: lines[0]}, nil
	}

	var groups []group
	for _, l := range lines {
		g := reGroupLine.FindStringSubmatch(l)
		if g == nil {
			return fileOrder{}, fmt.Errorf(
				"%q is not %q, not %q, and not a `NN-NN name` grouping line",
				l, orderDerived, orderAccidental)
		}
		lo, _ := strconv.Atoi(g[1])
		hi, _ := strconv.Atoi(g[2])
		if lo > hi {
			return fileOrder{}, fmt.Errorf("group %q runs backwards", l)
		}
		groups = append(groups, group{lo: lo, hi: hi, name: strings.TrimSpace(g[3])})
	}
	return fileOrder{groups: groups}, nil
}

// tierConstructFiles returns a tier's `.t` file names, sorted, the
// adjacency file included -- regenerateNumbering drops it itself.
func tierConstructFiles(t *testing.T, tier string) []string {
	t.Helper()

	paths, err := filepath.Glob(filepath.Join(corpusDir, tier, "*.t"))
	if err != nil {
		t.Fatalf("globbing %s: %v", tier, err)
	}
	var names []string
	for _, p := range paths {
		names = append(names, filepath.Base(p))
	}
	sort.Strings(names)
	return names
}

// TestEveryTierDeclaresFileOrder is the corpus-wide half: every tier says
// which convention it uses, and no tier is silent.
//
// The issue this settles found two conventions and no record of which tier
// used which, discovered only when a tier's agent wrote the mirrored test,
// watched it fail, and deleted it. A declaration is what turns that
// discovery into a lookup.
func TestEveryTierDeclaresFileOrder(t *testing.T) {
	orders, err := readFileOrders(corpusDir)
	if err != nil {
		t.Fatalf("%v", err)
	}

	entries, err := os.ReadDir(corpusDir)
	if err != nil {
		t.Fatalf("reading the corpus root: %v", err)
	}
	for _, e := range entries {
		if !e.IsDir() || !isTierDir(e.Name()) {
			continue
		}
		if _, ok := orders[e.Name()]; !ok {
			t.Errorf("%s declares no file ordering", e.Name())
		}
	}
}

// TestDerivedTierNumberingRegenerates checks the tiers that CLAIM derived
// numbering against what a regeneration would produce.
//
// This is the corpus-wide form of tier 01's TestTierNumberingRegenerates,
// which asserted the same thing for one tier. The measurement that
// motivated widening it rather than copying it: of fourteen tiers, only
// TWO satisfy derived numbering. Asserting it corpus-wide would have
// renumbered twelve directories to satisfy a rule the spec never
// states -- the spec says a file's identity is its NAME and the number is
// its current position, and says nothing about positions being
// alphabetical.
//
// So the check is gated on the declaration. A tier that claims `derived`
// has promised its numbers are a function of its names, and this verifies
// the promise; a tier that declares groups has promised something else,
// and TestGroupedTierNumberingCoversItsFiles verifies that instead.
func TestDerivedTierNumberingRegenerates(t *testing.T) {
	orders, err := readFileOrders(corpusDir)
	if err != nil {
		t.Fatalf("%v", err)
	}

	for _, tier := range sortedTiers(orders) {
		if orders[tier].word != orderDerived {
			continue
		}
		names := tierConstructFiles(t, tier)
		if len(names) == 0 {
			t.Errorf("%s holds no corpus files", tier)
			continue
		}
		want := regenerateNumbering(names)
		for _, n := range names {
			if n == adjacencyFile {
				continue
			}
			m := reNumbered.FindStringSubmatch(n)
			if m == nil {
				t.Errorf("%s/%s is not numbered `NN_name.t`, so nothing can place it",
					tier, n)
				continue
			}
			if got := want[m[2]]; got != n {
				t.Errorf("%s declares `%s` numbering but %s would regenerate as %s",
					tier, orderDerived, n, got)
			}
		}
	}
}

// TestGroupedTierNumberingCoversItsFiles checks a declared grouping
// against the directory.
//
// A grouping is a claim about WHICH FILES sit where, so the ranges must
// partition the tier's construct files exactly: start at 01, run
// contiguously, and end at the last file. A gap means a file belongs to no
// declared group and the arrangement does not describe the tier; an
// overlap means two groups claim one file; a range running past the end
// means the tier lost a file and the README did not notice.
//
// This is what makes the grouping spelling expensive enough to be honest.
// A tier whose numbering is an accident cannot satisfy it without writing
// down group names that a reader can check against the file names, and a
// tier that later gains or loses a file is told immediately rather than
// drifting -- which is the failure mode the derived check was written to
// prevent, recovered for the tiers that cannot be derived.
func TestGroupedTierNumberingCoversItsFiles(t *testing.T) {
	orders, err := readFileOrders(corpusDir)
	if err != nil {
		t.Fatalf("%v", err)
	}

	for _, tier := range sortedTiers(orders) {
		order := orders[tier]
		if len(order.groups) == 0 {
			continue
		}

		var nums []int
		for _, n := range tierConstructFiles(t, tier) {
			if n == adjacencyFile {
				continue
			}
			m := reNumbered.FindStringSubmatch(n)
			if m == nil {
				t.Errorf("%s/%s is not numbered `NN_name.t`, so nothing can place it",
					tier, n)
				continue
			}
			v, _ := strconv.Atoi(m[1])
			nums = append(nums, v)
		}
		sort.Ints(nums)
		if len(nums) == 0 {
			t.Errorf("%s holds no construct files", tier)
			continue
		}

		// Groups in declaration order must tile 01..last with no gap and
		// no overlap. Checking the boundaries rather than each file is
		// enough because the numbering itself is checked to be gapless
		// below.
		next := 1
		for _, g := range order.groups {
			if g.lo != next {
				t.Errorf("%s: group %q starts at %02d, but %02d is where the "+
					"previous group left off", tier, g.name, g.lo, next)
			}
			next = g.hi + 1
		}
		if last := nums[len(nums)-1]; next-1 != last {
			t.Errorf("%s: the groups run to %02d but the tier's last construct "+
				"file is %02d", tier, next-1, last)
		}

		// A grouping is a claim about contiguous runs, so a gap in the
		// numbering would make the claim meaningless even if the ranges
		// tiled: `04-07 loops` says nothing useful if 05 does not exist.
		for i, v := range nums {
			if v != i+1 {
				t.Errorf("%s declares groups but its numbering has a gap: "+
					"expected %02d at position %d, found %02d", tier, i+1, i+1, v)
				break
			}
		}
	}
}

// TestAccidentalTierNumberingIsNotDerivable checks that a tier claiming
// `accidental` really is not derivable.
//
// Without this, `accidental` is an opt-out: any tier could declare it and
// stop being checked, which is how a word that asserts nothing turns into
// a way to assert nothing on purpose. A tier whose files already sit in
// the order a regeneration produces has the derived property whether it
// meant to or not, and should say so, because then the stronger check
// applies and the numbering cannot drift out of it unnoticed.
//
// It is also how the debt gets paid down. When the issue that regenerates
// the ratchet renumbers one of these eleven tiers, this test fails on the
// stale `accidental` and names the tier, so the README is updated in the
// same change rather than left behind.
func TestAccidentalTierNumberingIsNotDerivable(t *testing.T) {
	orders, err := readFileOrders(corpusDir)
	if err != nil {
		t.Fatalf("%v", err)
	}

	for _, tier := range sortedTiers(orders) {
		if orders[tier].word != orderAccidental {
			continue
		}
		names := tierConstructFiles(t, tier)
		want := regenerateNumbering(names)
		differs := false
		for _, n := range names {
			if n == adjacencyFile {
				continue
			}
			m := reNumbered.FindStringSubmatch(n)
			if m == nil {
				t.Errorf("%s/%s is not numbered `NN_name.t`, so nothing can place it",
					tier, n)
				continue
			}
			if want[m[2]] != n {
				differs = true
			}
		}
		if !differs {
			t.Errorf("%s declares `%s` numbering but its files are already in the "+
				"order a regeneration produces, so it should declare `%s`",
				tier, orderAccidental, orderDerived)
		}
	}
}

// sortedTiers gives the tier names in order, so failures are reported in
// the order a reader would look for them rather than in map order.
func sortedTiers(orders map[string]fileOrder) []string {
	var out []string
	for tier := range orders {
		out = append(out, tier)
	}
	sort.Strings(out)
	return out
}
