// ABOUTME: Reads the grammar file's precedence declarations and compares their levels with chapter 4 and parse.
// ABOUTME: Spec §3 rule 9 and §9.2: a level held twice is checked by a test that lists each disagreement for review.

package grammar

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

// Level is one precedence declaration: its number counting from the loosest (1), its associativity (left, right or
// nonassoc), its tokens and, for the grammar file's, its citation.
type Level struct {
	N      int
	Assoc  string
	Tokens []string
	Cite   string
}

// ParseLevels reads precedence declarations, loosest first, one per line: `%<assoc> TOKEN... # <citation>`.
// Blank lines are skipped.
func ParseLevels(text string) ([]Level, error) {
	var ls []Level
	for i, l := range strings.Split(text, "\n") {
		if l = strings.TrimSpace(l); l == "" {
			continue
		}
		decl, cite, _ := strings.Cut(l, " # ")
		f := strings.Fields(decl)
		if _, _, err := parseCite(cite); err != nil || len(f) < 2 || !slices.Contains([]string{"%left", "%right", "%nonassoc"}, f[0]) {
			return nil, fmt.Errorf("line %d is not `%%left|%%right|%%nonassoc TOKEN... # perly.y:N[, fitted: <note>]`: %q", i+1, l)
		}
		ls = append(ls, Level{len(ls) + 1, f[0][1:], f[1:], cite})
	}
	return ls, nil
}

// LevelDifferences lists every level the grammar, chapter 4's level table and parse.LevelsAccountedFor() do not
// agree on, by number: `level N<TAB>grammar: <assoc tokens|none><TAB>chapter 4: <assoc tokens|none><TAB>parse:
// <account|none>`. A level differs when one side lacks it, or chapter 4 gives it another associativity or other
// tokens; parse's account is prose and names no associativity, so only its presence is compared.
func LevelDifferences(grammar, chapter4 []Level, parse map[int]string) []string {
	byN := func(ls []Level) map[int]Level {
		m := map[int]Level{}
		for _, l := range ls {
			m[l.N] = l
		}
		return m
	}
	g, c := byN(grammar), byN(chapter4)
	ns := map[int]bool{}
	for _, m := range []map[int]Level{g, c} {
		for n := range m {
			ns[n] = true
		}
	}
	for n := range parse {
		ns[n] = true
	}
	account := func(l Level, ok bool) string {
		if !ok {
			return "none"
		}
		return l.Assoc + " " + strings.Join(l.Tokens, " ")
	}
	var diffs []string
	for _, n := range slices.Sorted(maps.Keys(ns)) {
		gl, inG := g[n]
		cl, inC := c[n]
		p, inP := parse[n]
		if inG && inC && inP && account(gl, true) == account(cl, true) {
			continue
		}
		if !inP {
			p = "none"
		}
		diffs = append(diffs, fmt.Sprintf("level %d\tgrammar: %s\tchapter 4: %s\tparse: %s", n, account(gl, inG), account(cl, inC), p))
	}
	return diffs
}
