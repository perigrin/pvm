// ABOUTME: The port ledger: which perly.y alternatives are ported, fitted, excluded or still open.
// ABOUTME: Reads the ledger and the grammar file and holds their citations equal to the reference list.

package grammar

import (
	_ "embed"
	"fmt"
	"slices"
	"strconv"
	"strings"
)

//go:embed testdata/perly/ledger.txt
var ledgerText string

//go:embed testdata/perly/refused-in-action.txt
var refusedText string

// Entry is one alternative: its perly.y line, its text ("lhs: rhs") and, in the ledger, its mark.
type Entry struct {
	Line       int
	Text, Mark string
}

const citePrefix = "perly.y:"

// comparisonChains are the four alternatives a ledger may mark refused-by-precedence (spec §5.2).
var comparisonChains = []int{1466, 1468, 1484, 1486}

// exclusionReasons is the closed list of reasons an `excluded` mark may give.
var exclusionReasons = []string{"keyword-plugin", "refused-in-action", "conditional-refusal", "error-recovery"}

// reasonLines pins a reason to the lines it is true of. refused-in-action is
// pinned to refused-in-action.txt; conditional-refusal is left to review.
var reasonLines = map[string][]int{"keyword-plugin": {878, 1714}, "error-recovery": {933}}

// ParseEntries reads `perly.y:<line> <text>` lines, each followed by a tab and
// its mark when marked. The mark is kept verbatim for checkMarks to judge.
func ParseEntries(text string, marked bool) ([]Entry, error) {
	var es []Entry
	for i, l := range strings.Split(strings.TrimSuffix(text, "\n"), "\n") {
		body, mark, hasMark := strings.Cut(l, "\t")
		ref, rest, ok := strings.Cut(body, " ")
		line, err := parseLine(ref)
		if !ok || err != nil || hasMark != marked {
			return nil, fmt.Errorf("line %d is not `perly.y:<line> <lhs>: <rhs>`, with a tab-separated mark only in the ledger: %q", i+1, l)
		}
		es = append(es, Entry{line, rest, mark})
	}
	return es, nil
}

func parseLine(ref string) (int, error) {
	n, err := strconv.Atoi(strings.TrimPrefix(ref, citePrefix))
	if err != nil || n <= 0 || !strings.HasPrefix(ref, citePrefix) {
		return 0, fmt.Errorf("%q is not %s<line>", ref, citePrefix)
	}
	return n, nil
}

// ParseRules reads the grammar file's data: one rule per line, `lhs: rhs # <citation>`,
// the citation `perly.y:N` or `perly.y:N, fitted: <what the lexer emits instead>`.
// Blank lines and lines starting with # (the placeholders) are skipped.
func ParseRules(text string) ([]Rule, error) {
	var rs []Rule
	for i, l := range strings.Split(text, "\n") {
		if l = strings.TrimSpace(l); l == "" || l[0] == '#' {
			continue
		}
		rule, cite, ok := strings.Cut(l, " # ")
		lhs, rhs, ok2 := strings.Cut(rule, ":")
		if _, _, err := parseCite(cite); !ok || !ok2 || err != nil {
			return nil, fmt.Errorf("line %d is not `lhs: rhs # perly.y:N[, fitted: <note>]`: %q", i+1, l)
		}
		r := Rule{LHS: strings.TrimSpace(lhs), Cite: cite}
		for _, s := range strings.Fields(rhs) {
			r.RHS = append(r.RHS, Symbol{Name, s})
		}
		rs = append(rs, r)
	}
	return rs, nil
}

// parseCite returns a citation's line and whether it is a fitted one.
func parseCite(cite string) (line int, fitted bool, err error) {
	ref, note, fitted := strings.Cut(cite, ", fitted: ")
	if fitted && strings.TrimSpace(note) == "" {
		return 0, false, fmt.Errorf("citation %q has an empty fitted note", cite)
	}
	line, err = parseLine(ref)
	return line, fitted, err
}

// checkMarks refuses a mark that is not well-formed: `-`, ported, fitted,
// refused-by-precedence on a comparison chain, or `excluded <reason>` with
// the reason on the closed list, no other word, and the alternative one the
// reason is true of.
func checkMarks(ledger []Entry) error {
	refused, err := parseRefused(refusedText)
	if err != nil {
		return err
	}
	for _, e := range ledger {
		f := strings.Fields(e.Mark)
		switch {
		case e.Mark == "-" || e.Mark == "ported" || e.Mark == "fitted":
		case e.Mark == "refused-by-precedence" && slices.Contains(comparisonChains, e.Line):
		case len(f) == 2 && f[0] == "excluded" && e.Mark == strings.Join(f, " ") && slices.Contains(exclusionReasons, f[1]) &&
			reasonFits(f[1], e.Line, refused):
		default:
			return fmt.Errorf("%s%d: mark %q is not -, ported, fitted, refused-by-precedence (comparison chains only) or excluded <%s> on a line that reason is true of",
				citePrefix, e.Line, e.Mark, strings.Join(exclusionReasons, "|"))
		}
	}
	return nil
}

func reasonFits(reason string, line int, refused []int) bool {
	switch reason {
	case "refused-in-action":
		return slices.Contains(refused, line)
	case "conditional-refusal":
		return true
	}
	return slices.Contains(reasonLines[reason], line)
}

// CheckLedger holds the ledger to the reference list: the same alternatives,
// in the same order, each text equal, each mark well-formed.
func CheckLedger(alts, ledger []Entry) error {
	for i := 0; i < max(len(alts), len(ledger)); i++ {
		switch {
		case i >= len(ledger):
			return fmt.Errorf("the ledger ends before %s%d", citePrefix, alts[i].Line)
		case i >= len(alts):
			return fmt.Errorf("%s%d is in the ledger and not in the reference list", citePrefix, ledger[i].Line)
		case alts[i].Line != ledger[i].Line || alts[i].Text != ledger[i].Text:
			return fmt.Errorf("ledger entry %d is %s%d %s, the reference list has %s%d %s",
				i+1, citePrefix, ledger[i].Line, ledger[i].Text, citePrefix, alts[i].Line, alts[i].Text)
		}
	}
	return checkMarks(ledger)
}

// CheckCitations holds the rules' citations to the ledger: each names a ledger
// alternative marked ported (plain) or fitted (fitted), and each ported or
// fitted mark is cited by some rule.
func CheckCitations(ledger []Entry, rules []Rule) error {
	mark := map[int]string{}
	for _, e := range ledger {
		mark[e.Line] = e.Mark
	}
	cited := map[int]bool{}
	for _, r := range rules {
		line, fitted, err := parseCite(r.Cite)
		if err != nil {
			return err
		}
		m, ok := mark[line]
		want := "ported"
		if fitted {
			want = "fitted"
		}
		if !ok {
			return fmt.Errorf("%q cites %s%d, which is not in the ledger", r.LHS, citePrefix, line)
		}
		if m != want {
			return fmt.Errorf("%q cites %s%d as %s, but the ledger marks it %q", r.LHS, citePrefix, line, want, m)
		}
		cited[line] = true
	}
	for _, e := range ledger {
		if (e.Mark == "ported" || e.Mark == "fitted") && !cited[e.Line] {
			return fmt.Errorf("%s%d is marked %s and no rule cites it", citePrefix, e.Line, e.Mark)
		}
	}
	return nil
}

// LedgerCounts is the ratchet's ledger-gaps (unmarked alternatives plus rule
// citations not in the ledger) and ledger-excluded, read from the committed
// ledger and grammar file.
func LedgerCounts() (gaps, excluded int, err error) {
	ledger, err := ParseEntries(ledgerText, true)
	if err != nil {
		return 0, 0, err
	}
	rules, err := ParseRules(perlyRules)
	if err != nil {
		return 0, 0, err
	}
	in := map[int]bool{}
	for _, e := range ledger {
		in[e.Line] = true
		switch {
		case e.Mark == "-":
			gaps++
		case strings.HasPrefix(e.Mark, "excluded"):
			excluded++
		}
	}
	for _, r := range rules {
		if line, _, err := parseCite(r.Cite); err != nil || !in[line] {
			gaps++
		}
	}
	return gaps, excluded, nil
}

// parseRefused reads refused-in-action.txt: `perly.y:<line>` per line.
func parseRefused(text string) ([]int, error) {
	var lines []int
	for _, f := range strings.Fields(text) {
		n, err := parseLine(f)
		if err != nil {
			return nil, err
		}
		lines = append(lines, n)
	}
	return lines, nil
}

// CheckFamilyMarked is what a family's test calls with the perly.y lines of
// its alternatives and the lines it leaves for the completing issue. A family
// line must be marked in the ledger or named in left, and a left line must be
// an alternative whose action refuses unconditionally (refused-in-action.txt)
// or one of the four refused-by-precedence comparison chains.
func CheckFamilyMarked(lines, left []int) error {
	ledger, err := ParseEntries(ledgerText, true)
	if err != nil {
		return err
	}
	refused, err := parseRefused(refusedText)
	if err != nil {
		return err
	}
	return checkFamilyMarked(ledger, refused, lines, left)
}

func checkFamilyMarked(ledger []Entry, refused, lines, left []int) error {
	mark := map[int]string{}
	for _, e := range ledger {
		mark[e.Line] = e.Mark
	}
	for _, l := range lines {
		m, ok := mark[l]
		switch {
		case !ok:
			return fmt.Errorf("%s%d is not in the ledger", citePrefix, l)
		case m == "-" && !slices.Contains(left, l):
			return fmt.Errorf("%s%d is unmarked and not named as left for the completing issue", citePrefix, l)
		}
	}
	for _, l := range left {
		if !slices.Contains(refused, l) && !slices.Contains(comparisonChains, l) {
			return fmt.Errorf("%s%d is left, but is neither on refused-in-action.txt nor a comparison chain", citePrefix, l)
		}
	}
	return nil
}
