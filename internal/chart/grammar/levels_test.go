// ABOUTME: Holds the grammar's precedence levels' differences from chapter 4 and parse equal to the committed list.
// ABOUTME: Also shows the comparison can fail, and holds every declaration and %prec to the perly.y line it cites.

package grammar

import (
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

const (
	levelList = "testdata/level-differences.txt"
	chapter4  = "../../../docs/specs/perl-parser/04-expressions.md"
)

// chapter4Row is a row of chapter 4's level table: `| # | BP | tokens | assoc | ...`, each token in backticks.
var chapter4Row = regexp.MustCompile("^\\| (\\d+) \\| \\d+ \\| ([^|]+) \\| \\**(left|right|nonassoc)\\** \\|")

// chapter4Levels reads chapter 4's level table, read and never changed.
func chapter4Levels(t *testing.T) []Level {
	t.Helper()
	b, err := os.ReadFile(chapter4)
	if err != nil {
		t.Fatal(err)
	}
	var ls []Level
	for _, l := range strings.Split(string(b), "\n") {
		m := chapter4Row.FindStringSubmatch(l)
		if m == nil {
			continue
		}
		n, _ := strconv.Atoi(m[1])
		var tokens []string
		for _, tok := range strings.Split(m[2], "`")[1:] {
			if tok = strings.TrimSpace(tok); tok != "" {
				tokens = append(tokens, tok)
			}
		}
		ls = append(ls, Level{N: n, Assoc: m[3], Tokens: tokens})
	}
	if len(ls) == 0 {
		t.Fatalf("%s has no level table rows", chapter4)
	}
	return ls
}

func TestLevelDifferences(t *testing.T) {
	levels, err := ParseLevels(perlyPrecedence)
	if err != nil {
		t.Fatal(err)
	}
	list, _ := listEntries(t, levelList)
	if d := diffAgainstList(LevelDifferences(levels, chapter4Levels(t), parse.LevelsAccountedFor()), list); d != "" {
		t.Errorf("the grammar's levels differ from chapter 4 and parse.LevelsAccountedFor() other than as %s lists; review each and edit the list:\n%s", levelList, d)
	}
}

func TestLevelDifferencesDetects(t *testing.T) {
	lv := func(n int, assoc string, tokens ...string) Level { return Level{N: n, Assoc: assoc, Tokens: tokens} }
	grammar := []Level{lv(1, "left", "A"), lv(2, "right", "B"), lv(3, "nonassoc", "C"), lv(4, "left", "D"), lv(5, "left", "E", "F")}
	chapter := []Level{lv(1, "left", "A"), lv(3, "nonassoc", "C"), lv(4, "right", "D"), lv(5, "left", "E")}
	accounted := map[int]string{1: "a", 2: "b", 4: "d", 5: "e"}
	want := []string{
		"level 2\tgrammar: right B\tchapter 4: none\tparse: b",
		"level 3\tgrammar: nonassoc C\tchapter 4: nonassoc C\tparse: none",
		"level 4\tgrammar: left D\tchapter 4: right D\tparse: d",
		"level 5\tgrammar: left E F\tchapter 4: left E\tparse: e",
	}
	got := LevelDifferences(grammar, chapter, accounted)
	if !slices.Equal(got, want) {
		t.Errorf("synthetic differences:\n got %q\nwant %q", got, want)
	}
	if got := LevelDifferences(grammar[:1], nil, map[int]string{1: "a"}); !slices.Equal(got, []string{"level 1\tgrammar: left A\tchapter 4: none\tparse: a"}) {
		t.Errorf("a level only chapter 4 lacks: %q", got)
	}
	if got := LevelDifferences(nil, chapter[:1], nil); !slices.Equal(got, []string{"level 1\tgrammar: none\tchapter 4: left A\tparse: none"}) {
		t.Errorf("a level the grammar lacks: %q", got)
	}
	if d := diffAgainstList(got, want); d != "" {
		t.Errorf("the equal list was refused:\n%s", d)
	}
	if diffAgainstList(got, want[1:]) == "" {
		t.Error("a list omitting an entry passed")
	}
	if diffAgainstList(got, append(slices.Clone(want), "level 9\tgrammar: none\tchapter 4: none\tparse: x")) == "" {
		t.Error("a list adding an entry passed")
	}
}

var levelListForm = regexp.MustCompile(`^level \d+\tgrammar: [^\t]+\tchapter 4: [^\t]+\tparse: [^\t]+\t# \S.*$`)

func TestLevelDifferencesListForm(t *testing.T) {
	_, raw := listEntries(t, levelList)
	for _, l := range raw {
		if !levelListForm.MatchString(l) {
			t.Errorf("entry does not name each side's account and a reason (`level N<TAB>grammar: ..<TAB>chapter 4: ..<TAB>parse: ..<TAB># reason`): %q", l)
		}
	}
}

// perlyPrecUses are perly.y's %prec uses, by the alternative that carries each (`grep -n '%prec' perly.y` at
// 94e50866; 1336's is on the alternative starting at 1330). The reference the grammar file is held to.
var perlyPrecUses = map[int]string{
	1165: "PREC_LOW", 1262: "PREC_LOW", 1273: "PREC_LOW", 1330: "LSTOP", 1436: "PREC_LOW", 1438: "PREC_LOW",
	1462: "PREC_LOW", 1480: "PREC_LOW", 1497: "UMINUS", 1499: "UMINUS", 1536: "PERLY_PAREN_OPEN",
	1538: "PERLY_PAREN_OPEN", 1541: "PERLY_PAREN_OPEN", 1543: "PERLY_PAREN_OPEN", 1546: "PERLY_PAREN_OPEN",
	1548: "PERLY_PAREN_OPEN", 1553: "PERLY_PAREN_OPEN", 1558: "UNIOP", 1560: "PERLY_PAREN_OPEN", 1572: "UNIOP",
	1574: "UNIOP", 1582: "PERLY_PAREN_OPEN", 1584: "PERLY_PAREN_OPEN", 1586: "PERLY_PAREN_OPEN",
	1588: "PERLY_PAREN_OPEN", 1590: "PERLY_PAREN_OPEN", 1634: "PERLY_PAREN_OPEN", 1659: "PERLY_PAREN_OPEN",
	1735: "PERLY_PAREN_OPEN", 1737: "PERLY_PAREN_OPEN", 1739: "PERLY_PAREN_OPEN", 1744: "PERLY_PAREN_OPEN",
	1749: "PERLY_PAREN_OPEN", 1754: "PERLY_PAREN_OPEN", 1787: "PREC_LOW", 1788: "PREC_LOW", 1818: "PREC_LOW",
	1819: "PREC_LOW", 1890: "PREC_LOW",
}

func TestPrecedenceCitations(t *testing.T) {
	if len(perlyPrecUses) != 39 {
		t.Fatalf("the reference holds %d %%prec uses, perly.y has 39", len(perlyPrecUses))
	}
	levels, err := ParseLevels(perlyPrecedence)
	if err != nil {
		t.Fatal(err)
	}
	cited := map[int]bool{}
	declared := map[string]bool{}
	for _, l := range levels {
		line, _, _ := parseCite(l.Cite)
		if line < 150 || line > 182 {
			t.Errorf("level %d cites %s, outside perly.y:150-182", l.N, l.Cite)
		}
		if cited[line] {
			t.Errorf("level %d cites perly.y:%d, which another level cites", l.N, line)
		}
		cited[line] = true
		for _, tok := range l.Tokens {
			declared[tok] = true
		}
	}
	ledger, err := ParseEntries(ledgerText, true)
	if err != nil {
		t.Fatal(err)
	}
	rules, err := ParseRules(perlyRules)
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range rules {
		line, _, _ := parseCite(r.Cite)
		switch {
		case r.Prec != perlyPrecUses[line]:
			t.Errorf("%s carries %%prec %q; perly.y:%d carries %q", ruleText(r, ":"), r.Prec, line, perlyPrecUses[line])
		case r.Prec != "" && !slices.ContainsFunc(ledger, func(e Entry) bool { return e.Line == line }):
			t.Errorf("%s carries %%prec on a citation that is no ledger alternative", ruleText(r, ":"))
		case r.Prec != "" && !declared[r.Prec]:
			t.Errorf("%s carries %%prec %s, which no level declares", ruleText(r, ":"), r.Prec)
		}
	}
}

// ParseRules reads a %prec only as the right-hand side's last two words.
func TestParseRulesPrec(t *testing.T) {
	rs, err := ParseRules("a: B c %prec D # perly.y:1\n\na: B # perly.y:2\n")
	if err != nil || len(rs) != 2 || rs[0].Prec != "D" || len(rs[0].RHS) != 2 || rs[1].Prec != "" {
		t.Errorf("ParseRules = %+v, %v", rs, err)
	}
	for _, bad := range []string{"a: B %prec # perly.y:1", "a: %prec D B # perly.y:1", "a: B %prec D %prec E # perly.y:1"} {
		if _, err := ParseRules(bad); err == nil {
			t.Errorf("ParseRules accepts %q", bad)
		}
	}
}

func TestParseLevels(t *testing.T) {
	ls, err := ParseLevels("%left A B # perly.y:150\n\n%nonassoc C # perly.y:151, fitted: x\n")
	if err != nil || len(ls) != 2 || ls[1].N != 2 || ls[1].Assoc != "nonassoc" || !slices.Equal(ls[0].Tokens, []string{"A", "B"}) {
		t.Errorf("ParseLevels = %+v, %v", ls, err)
	}
	for _, bad := range []string{"%left A", "%left # perly.y:150", "%token A # perly.y:150", "%left A # 150"} {
		if _, err := ParseLevels(bad); err == nil {
			t.Errorf("ParseLevels accepts %q", bad)
		}
	}
}
