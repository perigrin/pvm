// ABOUTME: Tests of the port ledger: citation/ledger equality, the closed exclusion reasons, the precedence mark.
// ABOUTME: Also the refused-in-action list and the family helper, each shown to refuse as well as accept.

package grammar

import (
	"os"
	"slices"
	"strings"
	"testing"
)

func readTestdata(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile("testdata/perly/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func mustEntries(t *testing.T, text string, marked bool) []Entry {
	t.Helper()
	es, err := ParseEntries(text, marked)
	if err != nil {
		t.Fatal(err)
	}
	return es
}

// committed returns the reference list and the committed ledger.
func committed(t *testing.T) (alts, ledger []Entry) {
	return mustEntries(t, readTestdata(t, "alternatives.txt"), false), mustEntries(t, readTestdata(t, "ledger.txt"), true)
}

func wantErr(t *testing.T, what string, err error, words ...string) {
	t.Helper()
	if err == nil {
		t.Errorf("%s: accepted", what)
		return
	}
	for _, w := range words {
		if !strings.Contains(err.Error(), w) {
			t.Errorf("%s: error %q does not say %q", what, err, w)
		}
	}
}

// toyLedger is three alternatives with the given marks.
func toyLedger(t *testing.T, m10, m12, m14 string) []Entry {
	return mustEntries(t, "perly.y:10 a: X\t"+m10+"\nperly.y:12 a: Y\t"+m12+"\nperly.y:14 b: Z\t"+m14+"\n", true)
}

func toyRules(t *testing.T, text string) []Rule {
	t.Helper()
	rs, err := ParseRules(text)
	if err != nil {
		t.Fatal(err)
	}
	return rs
}

func TestPortLedger(t *testing.T) {
	alts, ledger := committed(t)
	if err := CheckLedger(alts, ledger); err != nil {
		t.Fatal(err)
	}
	if err := CheckCitations(ledger, toyRules(t, perlyRules)); err != nil {
		t.Fatal(err)
	}

	const good = "a: X # perly.y:10\na: Y Z # perly.y:12, fitted: Y then Z\n"
	if err := CheckCitations(toyLedger(t, "ported", "fitted", "-"), toyRules(t, good)); err != nil {
		t.Fatalf("a matching toy ledger: %v", err)
	}
	for _, c := range []struct {
		name, rules string
		marks       [3]string
		words       []string
	}{
		{"an uncited ported mark", "a: Y # perly.y:12, fitted: y\n", [3]string{"ported", "fitted", "-"}, []string{"perly.y:10", "no rule cites"}},
		{"an uncited fitted mark", "a: X # perly.y:10\n", [3]string{"ported", "fitted", "-"}, []string{"perly.y:12", "no rule cites"}},
		{"a citation outside the ledger", good + "b: Z # perly.y:99\n", [3]string{"ported", "fitted", "-"}, []string{"perly.y:99", "not in the ledger"}},
		{"a fitted citation on a ported mark", "a: X # perly.y:10, fitted: x\n", [3]string{"ported", "-", "-"}, []string{"perly.y:10", "fitted"}},
		{"a plain citation on a fitted mark", "a: X # perly.y:10\n", [3]string{"fitted", "-", "-"}, []string{"perly.y:10", "ported"}},
		{"a citation on an unmarked alternative", "b: Z # perly.y:14\n", [3]string{"-", "-", "-"}, []string{"perly.y:14", "ported"}},
		{"a citation on an excluded alternative", "b: Z # perly.y:14\n", [3]string{"-", "-", "excluded keyword-plugin"}, []string{"perly.y:14", "ported"}},
	} {
		wantErr(t, c.name, CheckCitations(toyLedger(t, c.marks[0], c.marks[1], c.marks[2]), toyRules(t, c.rules)), c.words...)
	}

	for name, text := range map[string]string{
		"a rule without a citation":     "a: X\n",
		"an empty fitted note":          "a: X # perly.y:10, fitted: \n",
		"a citation that is not a line": "a: X # perly.y:ten\n",
		"a rule without a colon":        "a X # perly.y:10\n",
	} {
		if _, err := ParseRules(text); err == nil {
			t.Errorf("ParseRules accepted %s", name)
		}
	}
	rs := toyRules(t, "# perly.y:10\n\na: # perly.y:12\n")
	if len(rs) != 1 || rs[0].LHS != "a" || len(rs[0].RHS) != 0 || rs[0].Cite != "perly.y:12" {
		t.Errorf("comments and blank lines are skipped and an empty right-hand side is allowed: got %+v", rs)
	}
}

func TestPortLedgerRefusesReorderAndDrift(t *testing.T) {
	alts, ledger := committed(t)
	if err := CheckLedger(alts, ledger); err != nil {
		t.Fatalf("the committed ledger: %v", err)
	}
	clone := func() []Entry { return append([]Entry(nil), ledger...) }

	swapped := clone()
	swapped[3], swapped[4] = swapped[4], swapped[3]
	wantErr(t, "two alternatives swapped", CheckLedger(alts, swapped), "perly.y:")

	dup := slices.Concat(clone()[:5], clone()[4:])
	wantErr(t, "one alternative duplicated", CheckLedger(alts, dup), "perly.y:")

	drift := clone()
	drift[7].Text += " EXTRA"
	wantErr(t, "an alternative whose text differs", CheckLedger(alts, drift), drift[7].Text)

	wantErr(t, "an alternative dropped", CheckLedger(alts, clone()[1:]), "perly.y:")
	wantErr(t, "an alternative added", CheckLedger(alts, append(clone(), Entry{Line: 99999, Text: "x: y", Mark: "-"})), "perly.y:99999")
}

func TestLedgerPrecedenceMarkOnlyComparisonChains(t *testing.T) {
	alts, ledger := committed(t)
	chains := map[int]bool{1466: true, 1468: true, 1484: true, 1486: true}
	for i := range ledger {
		l := append([]Entry(nil), ledger...)
		l[i].Mark = "refused-by-precedence"
		err := CheckLedger(alts, l)
		if chains[l[i].Line] {
			if err != nil {
				t.Errorf("refused-by-precedence on perly.y:%d: %v", l[i].Line, err)
			}
		} else if err == nil {
			t.Errorf("refused-by-precedence accepted on perly.y:%d", l[i].Line)
		}
	}
}

func TestLedgerExclusionReasons(t *testing.T) {
	alts, ledger := committed(t)
	if err := CheckLedger(alts, ledger); err != nil {
		t.Fatal(err)
	}
	wantErr(t, "an unknown reason", checkMarks(toyLedger(t, "excluded because", "-", "-")), "perly.y:10", "because")
}

func TestLedgerExclusionReasonsRefusesMalformed(t *testing.T) {
	for r, line := range map[string]int{"keyword-plugin": 878, "refused-in-action": 1105, "conditional-refusal": 10, "error-recovery": 933} {
		if err := checkMarks([]Entry{{Line: line, Mark: "excluded " + r}}); err != nil {
			t.Errorf("reason %s: %v", r, err)
		}
	}
	for _, m := range []string{"excluded", "excluded keyword-plugin extra", "excluded  ", "excluded: keyword-plugin", "Excluded keyword-plugin", "ported extra", "other", ""} {
		wantErr(t, "mark "+m, checkMarks(toyLedger(t, m, "-", "-")), "perly.y:10")
	}
}

func TestRefusedInActionList(t *testing.T) {
	alts, _ := committed(t)
	known := map[int]bool{}
	for _, a := range alts {
		known[a.Line] = true
	}
	prev, n := 0, 0
	lines, err := parseRefused(readTestdata(t, "refused-in-action.txt"))
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range lines {
		n++
		switch {
		case !known[line]:
			t.Errorf("perly.y:%d is not a reference alternative", line)
		case line <= prev:
			t.Errorf("perly.y:%d is out of line order or repeated (after perly.y:%d)", line, prev)
		}
		prev = line
	}
	if n == 0 {
		t.Error("the list is empty")
	}
	for _, bad := range []string{"1105", "perly.y:x", "perly.y:0"} {
		if _, err := parseRefused(bad); err == nil {
			t.Errorf("parseRefused accepted %q", bad)
		}
	}
}

func TestCheckFamilyMarked(t *testing.T) {
	ledger := toyLedger(t, "ported", "-", "-")
	ledger = append(ledger, mustEntries(t, "perly.y:16 c: W\t-\n", true)...)
	refused := []int{12}
	chk := func(lines, left []int) error { return checkFamilyMarked(ledger, refused, lines, left) }

	if err := chk([]int{10}, nil); err != nil {
		t.Errorf("a fully marked family: %v", err)
	}
	if err := chk([]int{10, 12}, []int{12}); err != nil {
		t.Errorf("a listed line named as left: %v", err)
	}
	wantErr(t, "a family line unmarked and not named as left", chk([]int{10, 12}, nil), "perly.y:12")
	wantErr(t, "a left line not on the refused list", chk([]int{10, 14}, []int{14}), "perly.y:14", "refused-in-action")
	wantErr(t, "a line missing from the ledger", chk([]int{77}, nil), "perly.y:77")

	// The four comparison-chain alternatives may be left, and the committed data answers to the same rules.
	chains := []int{1466, 1468, 1484, 1486}
	if err := CheckFamilyMarked(chains, chains); err != nil {
		t.Errorf("the comparison chains named as left: %v", err)
	}
	if err := CheckFamilyMarked([]int{878, 933, 1714}, nil); err != nil {
		t.Errorf("the marked exclusions: %v", err)
	}
	wantErr(t, "an ordinary alternative named as left", CheckFamilyMarked([]int{187}, []int{187}), "perly.y:187")
}

// Each exclusion reason is true only of the lines it names (P0-1); conditional-refusal is left to review.
func TestLedgerExclusionReasonsPinnedToLines(t *testing.T) {
	alts, ledger := committed(t)
	refused, err := parseRefused(readTestdata(t, "refused-in-action.txt"))
	if err != nil {
		t.Fatal(err)
	}
	allowed := map[string][]int{
		"refused-in-action":   refused,
		"error-recovery":      {933},
		"keyword-plugin":      {878, 1714},
		"conditional-refusal": nil,
	}
	for reason, lines := range allowed {
		for i := range ledger {
			l := slices.Clone(ledger)
			l[i].Mark = "excluded " + reason
			err := CheckLedger(alts, l)
			if reason == "conditional-refusal" || slices.Contains(lines, l[i].Line) {
				if err != nil {
					t.Errorf("excluded %s on perly.y:%d: %v", reason, l[i].Line, err)
				}
			} else if err == nil {
				t.Errorf("excluded %s accepted on perly.y:%d", reason, l[i].Line)
			}
		}
	}
}

// A blank line between entries keeps two families' adjacent edits from conflicting on rebase (P0-2).
func TestLedgerAndPlaceholdersSeparatedByBlankLines(t *testing.T) {
	for name, text := range map[string]string{"ledger.txt": readTestdata(t, "ledger.txt"), "perly.go": perlyRules} {
		lines := strings.Split(strings.TrimSpace(text), "\n")
		for i := 1; i < len(lines); i++ {
			if lines[i] != "" && lines[i-1] != "" {
				t.Errorf("%s: %q directly follows %q", name, lines[i], lines[i-1])
				break
			}
		}
	}
	es, err := ParseEntries("perly.y:1 a: X\t-\n\nperly.y:2 a: Y\tported\n", true)
	if err != nil || len(es) != 2 || es[1].Line != 2 {
		t.Errorf("ParseEntries over blank-separated entries = %v, %v", es, err)
	}
}
