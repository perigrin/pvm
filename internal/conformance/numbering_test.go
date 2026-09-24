// ABOUTME: The corpus-wide ordering check: every tier declares how its topics are arranged.
// ABOUTME: `grouped` names the topics, `accidental` admits no arrangement, `derived` no longer means anything.
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
//	    accidental
//
// or, for a tier whose arrangement was chosen rather than accidental:
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

// reGroupLine matches one grouping line: an OPTIONAL leading range, kept
// for the spelling's history, and the NAME of what the group holds.
//
// The name is the part that still means something, and with topics it
// is the only part there is: a group is a topic file, and a topic file
// has no number for a range to describe. A tier still spelling the
// range is not wrong, just older than the format.
var reGroupLine = regexp.MustCompile(`^(?:(\d\d)-(\d\d)\s+)?(\S.*)$`)

// orderDerived is the word a tier used when its numbers were a function
// of its file names: sort the identities, count from 01.
//
// IT NO LONGER NAMES ANYTHING CHECKABLE, and this constant survives only
// so a tier still spelling it gets told why. The corpus was 212 `.t`
// files whose `NN_` prefixes were a derived quantity -- the spec said so:
// "the number is DERIVED from the classification, so a reorder is a
// regeneration rather than a hand-edit". `regenerateNumbering` threw the
// numbers away and rebuilt them, and TestDerivedTierNumberingRegenerates
// checked that the rebuild matched disk.
//
// The corpus is now 65 topic files holding several cases each, and a
// topic carries NO NUMBER AT ALL. There is nothing to regenerate: the
// claim was about a naming convention, and the naming convention is
// gone. So this is not a check that got weaker -- it is a check whose
// subject stopped existing, which is different and worth saying out
// loud rather than leaving a test that passes over no files.
const orderDerived = "derived"

// orderAccidental is the word a tier uses when its topics sit in no
// arrangement anyone chose.
//
// This survived the format change unaltered, because it was never about
// numbers: it says the tier makes NO ordering claim. Eleven of fourteen
// tiers declare it, which was true of the `.t` corpus and is true of the
// topics that replaced them -- the eleven READMEs cite their material by
// name and never by position.
//
// It buys no check, and that is the point of its cost: it is the only
// spelling that asserts nothing, so a tier that could say something truer
// has a reason to.
const orderAccidental = "accidental"

// group is one declared, named part of a tier.
//
// WAS A RANGE OF FILE NUMBERS, IS NOW A TOPIC. Under the `.t` corpus a
// group was `01-04 conditionals`: a contiguous run of numbered files and
// a word for what the run held. The run is gone with the numbers, but
// the WORD survived the port literally -- tier 06's three groups
// `conditionals`, `loops` and `jumps` are now `conditionals.md`,
// `loops.md` and `jumps.md`, because the port cut topics along the lines
// the groupings already drew.
//
// So the claim a grouping makes is restated rather than dropped: it named
// the parts of the tier, and it still does. lo and hi are parsed so the
// README spelling keeps working, and are not checked against anything --
// nothing on disk carries a number for them to agree with.
type group struct {
	lo, hi int
	name   string
}

// fileOrder is what one tier declares about its arrangement: accidental,
// or a list of groups saying what the tier's parts ARE.
type fileOrder struct {
	word   string // orderDerived, orderAccidental, or "" when grouped
	groups []group
}

// readFileOrders collects every tier's declared ordering.
//
// Missing is an error rather than a default. A default would let a tier
// ship with no opinion at all, which is the state this check exists to
// end: before it, one tier was alphabetical on purpose, one was grouped on
// purpose, and twelve were neither on purpose or otherwise -- and nothing
// on disk distinguished the three.
//
// The spellings are deliberately asymmetric in cost. A grouping must NAME
// the tier's parts, which a tier whose arrangement is an accident cannot
// write down without the falsehood being visible to a reader.
// `accidental` is one word and buys nothing, which is what makes it a
// record of debt rather than a rubber stamp -- a tier has an incentive to
// say something truer.
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

// tierTopics returns the topic file names a tier holds, sorted, the
// adjacency topic included.
//
// REPLACES tierConstructFiles, which globbed `<tier>/*.t`. A tier is no
// longer a directory of cases -- a topic DECLARES its tier in its own
// prose -- so the tier's membership is read from the corpus rather than
// from a path.
func tierTopics(t *testing.T, tier string) []string {
	t.Helper()

	cases, err := AllCases(corpusDir)
	if err != nil {
		t.Fatalf("reading the corpus: %v", err)
	}
	seen := map[string]bool{}
	var names []string
	for _, c := range cases {
		if c.Tier != tier || seen[c.Topic] {
			continue
		}
		seen[c.Topic] = true
		names = append(names, c.Topic)
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
//
// Untouched by the topic port: it reads READMEs and never the corpus, so
// what it asserts -- that fourteen tiers each declare something -- means
// the same over topics as it did over files.
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

// TestDeclaredOrderingIsStillMeaningful reports any tier still claiming
// `derived`, whose subject the topic format removed.
//
// REPLACES TestDerivedTierNumberingRegenerates, and it is the one claim
// in this file that could not be carried across. That test threw a tier's
// `NN_` prefixes away, rebuilt them by sorting the identities, and
// compared. A topic has no prefix to throw away, so the rebuild has no
// input and the comparison no subject: kept as it was, it would have
// iterated over zero files and passed, which is the vacuous-pass failure
// this package has been bitten by before.
//
// What stands in its place is a check that no tier is still making the
// claim. `derived` promised a machine could reproduce the corpus's
// layout from its names; nothing about topics makes that promise, so a
// README that still says the word is describing a corpus that is gone,
// and the reader it misleads is the next person deciding whether they may
// rename a topic.
//
// It is deliberately NOT a weaker version of the old check. It asserts
// about the README, not about the corpus, because the README is the only
// place the stale claim can now live.
func TestDeclaredOrderingIsStillMeaningful(t *testing.T) {
	orders, err := readFileOrders(corpusDir)
	if err != nil {
		t.Fatalf("%v", err)
	}

	for _, tier := range sortedTiers(orders) {
		if orders[tier].word != orderDerived {
			continue
		}
		t.Errorf("%s declares `%s` numbering, but topics carry no numbers to "+
			"derive -- say `%s`, or name the tier's topics as groups",
			tier, orderDerived, orderAccidental)
	}
}

// TestGroupedTierNamesItsTopics checks a declared grouping against the
// tier's topics.
//
// REPLACES TestGroupedTierNumberingCoversItsFiles, whose claim was that
// the declared ranges partitioned the tier's numbered files exactly:
// start at 01, run contiguously, end at the last file. Three quarters of
// that was arithmetic over numbers that no longer exist -- no gaps, no
// overlaps, no range past the end.
//
// The part that survived is the part that was never about numbers: a
// grouping NAMES THE PARTS OF A TIER, and a reader must be able to find
// each named part. Under the port the names came across literally --
// tier 06's `conditionals`, `loops` and `jumps` are `conditionals.md`,
// `loops.md` and `jumps.md` -- because the topics were cut along the
// lines the groupings already drew.
//
// So the check is: every group names a topic this tier holds. That keeps
// what made the grouping spelling expensive enough to be honest -- a tier
// whose arrangement is an accident still cannot write down names a reader
// can check -- and it keeps the drift alarm, since renaming or merging a
// topic without touching the README is reported. What it no longer
// checks, because nothing on disk says it, is that the parts are
// CONTIGUOUS or that they COVER the tier: a topic in no group is now
// invisible to this test.
func TestGroupedTierNamesItsTopics(t *testing.T) {
	orders, err := readFileOrders(corpusDir)
	if err != nil {
		t.Fatalf("%v", err)
	}

	for _, tier := range sortedTiers(orders) {
		order := orders[tier]
		if len(order.groups) == 0 {
			continue
		}

		topics := tierTopics(t, tier)
		if len(topics) == 0 {
			t.Errorf("%s holds no topics", tier)
			continue
		}
		have := map[string]bool{}
		for _, topic := range topics {
			have[strings.TrimSuffix(topic, ".md")] = true
		}

		for _, g := range order.groups {
			if !have[groupTopic(g.name)] {
				t.Errorf("%s: group %q names no topic; %s holds %s",
					tier, g.name, tier, strings.Join(topics, ", "))
			}
		}
	}
}

// groupTopic turns a group's prose name into the topic stem it would be
// filed under: lowercased, spaces hyphenated.
//
// A group name is PROSE a README author wrote -- `block-valued
// expressions` -- and a topic name is a file stem. Mapping one to the
// other rather than demanding they match exactly keeps the README
// readable, which is the whole reason the declaration lives there.
func groupTopic(name string) string {
	return strings.ReplaceAll(strings.ToLower(strings.TrimSpace(name)), " ", "-")
}

// TestAccidentalTierClaimsNoArrangement checks that a tier claiming
// `accidental` is not quietly describing one.
//
// REPLACES TestAccidentalTierNumberingIsNotDerivable, which ran the
// regeneration over a tier's `.t` names and complained if the numbers
// already matched, since such a tier had the derived property whether it
// meant to or not and should have said so. With `derived` gone there is
// no stronger spelling for it to be pushed toward by that route.
//
// What remains is the other half of the same job: `accidental` must not
// become an opt-out. A tier declaring it while its README names its parts
// as groups would be asserting nothing while a truer spelling sat one
// edit away, so the check is that the declaration and the block agree --
// a tier cannot say `accidental` and list groups.
//
// This is WEAKER than what it replaces, and that is not a choice: the old
// check compared the tier's on-disk order against a computed one, and no
// computed order over topics exists to compare against.
func TestAccidentalTierClaimsNoArrangement(t *testing.T) {
	orders, err := readFileOrders(corpusDir)
	if err != nil {
		t.Fatalf("%v", err)
	}

	for _, tier := range sortedTiers(orders) {
		order := orders[tier]
		if order.word != orderAccidental {
			continue
		}
		if len(order.groups) != 0 {
			t.Errorf("%s declares `%s` and also names %d groups; a tier that can "+
				"name its parts is arranged", tier, orderAccidental, len(order.groups))
		}
		if len(tierTopics(t, tier)) == 0 {
			t.Errorf("%s declares `%s` but holds no topics", tier, orderAccidental)
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
