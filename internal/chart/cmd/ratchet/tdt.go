// ABOUTME: The ratchet's tdt keys (spec §12): the token table's width against the aligned toke.c pairs in TDT_FIXTURES.
// ABOUTME: tdt-aligned, tdt-mismatch, tdt-unseen, tdt-unseen-unlisted and tdt-unaligned-unlisted, counted over whole captures.

package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/chart/lex"
	"tamarou.com/pvm/internal/lexer"
)

// defaultTDT is where this milestone generates the fixtures; TDT_FIXTURES overrides it, as T1_CORPUS does defaultT1.
const defaultTDT = "/var/tmp/pvm-chart-tdt"

var (
	familiesFile = filepath.Join(testdata, "tdt", "fitting-families.txt")
	reviewedFile = filepath.Join(testdata, "tdt", "unseen-reviewed.txt")
)

// tdtFile is a corpus file and its aligned-pairs fixture (internal/chart/offline/README.md).
type tdtFile struct {
	Path    string
	Src     []byte
	Fixture string
}

// span is a token's byte range, a fixture line's key.
type span struct{ start, end int }

// pair is a fixture line: the toke.c type its token aligned to ("-" if none) and its family.
type pair struct{ typ, family string }

// readFixture is a whole capture's pair lines by span; ok is false for any other capture,
// which has no alignment. shortcut: this parser is a few lines of its own, since
// internal/chart's fixture parser lives in a _test file and checks more than it reads.
func readFixture(text string) (pairs map[span]pair, ok bool) {
	pairs = map[span]pair{}
	for _, l := range strings.Split(text, "\n") {
		if c, found := strings.CutPrefix(l, "# capture: "); found {
			ok = c == "whole"
			continue
		}
		f := strings.Fields(l)
		if len(f) < 3 || f[0] == "unmatched" || strings.HasPrefix(l, "#") {
			continue
		}
		s, e1 := strconv.Atoi(f[0])
		e, e2 := strconv.Atoi(f[1])
		if e1 != nil || e2 != nil {
			continue
		}
		p := pair{typ: f[2]}
		if p.typ == "-" && len(f) > 3 {
			p.family = f[3]
		}
		pairs[span{s, e}] = p
	}
	return pairs, ok
}

// termSets is the terminal-set id of each significant token and heredoc body of s, by start byte;
// a body has its opener's.
func termSets(s *lex.Stream) map[int]int {
	m := map[int]int{}
	for _, t := range s.Tokens {
		m[s.Raw[t.Index].Start] = t.TerminalSet
	}
	for _, b := range s.Bodies {
		m[b.Start] = s.Tokens[b.Opener].TerminalSet
	}
	return m
}

// keyText writes a class's key as the reviewed list names it, parts in the struct's order.
func keyText(c grammar.TokenClass) string {
	var p []string
	add := func(name string, v any, set bool) {
		if set {
			p = append(p, fmt.Sprintf("%s=%v", name, v))
		}
	}
	if c.Kind != nil {
		add("Kind", *c.Kind, true)
	}
	if c.Text != nil {
		add("Text", strconv.Quote(*c.Text), true)
	}
	if c.OpensBlock != nil {
		add("OpensBlock", *c.OpensBlock, true)
	}
	if c.TermFollows != nil {
		add("TermFollows", *c.TermFollows, true)
	}
	if c.Prefix != nil {
		add("Prefix", strconv.Quote(string(rune(*c.Prefix))), true)
	}
	for i, n := range []*grammar.Neighbour{c.Prev2, c.Prev1, c.Next} {
		if n != nil {
			add([]string{"Prev2", "Prev1", "Next"}[i], fmt.Sprintf("%+v", *n), true)
		}
	}
	if c.Global != nil {
		add("Global", *c.Global, true)
	}
	if c.LineColon != nil {
		add("LineColon", *c.LineColon, true)
	}
	if c.Blanks != nil {
		add("Blanks", *c.Blanks, true)
	}
	add("NotText", digest(c.NotText), len(c.NotText) > 0)
	var prev []string
	for _, n := range c.NotPrev1 {
		prev = append(prev, fmt.Sprintf("%+v", n))
	}
	add("NotPrev1", digest(prev), len(prev) > 0)
	var prev2 []string
	for _, n := range c.NotPrev2 {
		prev2 = append(prev2, fmt.Sprintf("%+v", n))
	}
	add("NotPrev2", digest(prev2), len(prev2) > 0)
	var next []string
	for _, n := range c.NotNext {
		next = append(next, fmt.Sprintf("%+v", n))
	}
	if len(next) == 1 { // a lone neighbour reads as it did before NotNext was a list
		add("NotNext", next[0], true)
	} else {
		add("NotNext", digest(next), len(next) > 1)
	}
	add("NotNextIdent", true, c.NotNextIdent)
	add("NextName", true, c.NextName)
	add("NotNextName", true, c.NotNextName)
	add("Lone", true, c.Lone)
	return strings.Join(p, " ")
}

// digest writes an exclusion list in full when it is short, and as its length and a hash of its sorted
// members when it is not, so a key stays readable and two lists never share one.
func digest(items []string) string {
	s := strings.Join(slices.Sorted(slices.Values(items)), ",")
	if len(s) <= 60 {
		return "[" + s + "]"
	}
	return fmt.Sprintf("#%d:%x", len(items), sha256.Sum256([]byte(s)))[:16]
}

// tdtKeys measures the tdt keys of decls (the grammar's token classes; names are the terminals by id)
// over files. A key is a class and the terminal it declares; it is seen when a token it admits aligned
// to that terminal's toke.c type. reviewed lists the unseen keys a person has accepted, each written as
// "<terminal> <keyText>"; an entry naming no key of the table is an error.
func tdtKeys(decls []grammar.ClassDecl, names []string, files []tdtFile, fams map[string]bool, reviewed []string) (map[string]string, error) {
	ids := map[string]int{}
	for i, n := range names {
		ids[n] = i
	}
	classes := make([]grammar.TokenClass, len(decls))
	var keyNames []string // the distinct (terminal, key) pairs, in declaration order
	byName := map[string]int{}
	for i, d := range decls {
		id, ok := ids[d.Terminal]
		if !ok {
			return nil, fmt.Errorf("a token class names %s, which is no terminal", d.Terminal)
		}
		classes[i] = d.Key
		classes[i].Terminals = []int{id}
		k := strings.TrimSpace(d.Terminal + " " + keyText(d.Key))
		if _, dup := byName[k]; !dup {
			byName[k] = i
			keyNames = append(keyNames, k)
		}
	}
	for _, r := range reviewed {
		if _, ok := byName[r]; !ok {
			return nil, fmt.Errorf("%s: %q names no (terminal, key) pair of the table", reviewedFile, r)
		}
	}
	tt, err := grammar.CompileTokens(classes)
	if err != nil {
		return nil, err
	}
	alone := make([]*grammar.TokenTable, len(decls)) // each class compiled by itself, built when first needed
	seen := map[string]bool{}
	var total, aligned, mismatch, unlisted int
	for _, f := range files {
		pairs, whole := readFixture(f.Fixture)
		if !whole {
			continue
		}
		raw := lexer.Tokenize(f.Src)
		stream, _ := lex.Map(f.Src, raw, tt)
		sets := termSets(stream)
		byType := map[string][]int{} // aligned toke.c type -> the start bytes of its tokens
		for _, t := range raw {
			if lexer.IsTrivia(t.Kind) {
				continue
			}
			total++
			p, ok := pairs[span{t.Start, t.End}]
			if !ok || p.typ == "-" {
				if !fams[p.family] {
					unlisted++
				}
				continue
			}
			aligned++
			if id, known := ids[p.typ]; !known || !slices.Contains(tt.Sets[sets[t.Start]], id) {
				mismatch++
			}
			byType[p.typ] = append(byType[p.typ], t.Start)
		}
		for _, k := range keyNames {
			i := byName[k]
			starts := byType[decls[i].Terminal]
			if seen[k] || len(starts) == 0 {
				continue
			}
			if alone[i] == nil {
				if alone[i], err = grammar.CompileTokens(classes[i : i+1]); err != nil {
					return nil, err
				}
			}
			one, _ := lex.Map(f.Src, raw, alone[i])
			own := termSets(one)
			seen[k] = slices.ContainsFunc(starts, func(s int) bool { return own[s] != 0 })
		}
	}
	var unseen, unseenUnlisted int
	for _, k := range keyNames {
		if !seen[k] {
			unseen++
			if !slices.Contains(reviewed, k) {
				unseenUnlisted++
			}
		}
	}
	frac := 0.0
	if total > 0 {
		frac = float64(aligned) / float64(total)
	}
	return map[string]string{
		"tdt-aligned": strconv.FormatFloat(frac, 'f', 4, 64), "tdt-mismatch": strconv.Itoa(mismatch),
		"tdt-unseen": strconv.Itoa(unseen), "tdt-unseen-unlisted": strconv.Itoa(unseenUnlisted),
		"tdt-unaligned-unlisted": strconv.Itoa(unlisted),
	}, nil
}

// loadTDT reads each corpus file's fixture from dir, keyed by the file's path as tdt-align.sh writes it.
// A file with no fixture is an error naming it, never a smaller count.
func loadTDT(dir string, files []File, srcs map[string][]byte) ([]tdtFile, error) {
	out := make([]tdtFile, len(files))
	for i, f := range files {
		fx, err := os.ReadFile(filepath.Join(dir, f.Path+".tdt"))
		if err != nil {
			return nil, fmt.Errorf("%s has no aligned-pairs fixture in TDT_FIXTURES: %w", f.Path, err)
		}
		out[i] = tdtFile{f.Path, srcs[f.Path], string(fx)}
	}
	return out, nil
}

// listLines are a list file's lines that are neither blank nor a comment.
func listLines(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, l := range strings.Split(string(data), "\n") {
		if l = strings.TrimSpace(l); l != "" && !strings.HasPrefix(l, "#") {
			out = append(out, l)
		}
	}
	return out, nil
}

// fixturesDir is the TDT_FIXTURES value set, or def when it is unset.
func fixturesDir(set, def string) string {
	if set == "" {
		return def
	}
	return set
}

// measureTDT adds the tdt keys to measured, reading the fixtures in dir; a dir that is no directory is an error.
func measureTDT(dir string, g *grammar.Grammar, tab *grammar.Tables, files []File, srcs map[string][]byte, measured map[string]string) error {
	if fi, err := os.Stat(dir); err != nil || !fi.IsDir() {
		return fmt.Errorf("TDT_FIXTURES=%s names no directory (%v)", dir, err)
	}
	tfiles, err := loadTDT(dir, files, srcs)
	if err != nil {
		return err
	}
	lines, err := listLines(familiesFile)
	if err != nil {
		return err
	}
	fams := map[string]bool{}
	for _, l := range lines {
		fams[strings.Fields(l)[0]] = true
	}
	reviewed, err := listLines(reviewedFile)
	if err != nil {
		return err
	}
	names := make([]string, len(tab.Symbols))
	for i, s := range tab.Symbols {
		names[i] = s.Name
	}
	m, err := tdtKeys(g.Classes, names, tfiles, fams, reviewed)
	for k, v := range m {
		measured[k] = v
	}
	return err
}
