// ABOUTME: The precedence referee (spec §8): the chart's verdict and grouping against bison's on the committed sentences,
// ABOUTME: and the fixtures' coverage of bison's (rule, token) resolutions; the ratchet's precedence-* keys read CountPrecedence.

package chart

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"tamarou.com/pvm/internal/chart/forest"
	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/chart/lex"
	"tamarou.com/pvm/internal/chart/recognize"
)

// bisonConflict is one line of bison's report of perly.y, as
// offline/referee/sentences.sh copies it into bison-conflicts.txt.
var bisonConflict = regexp.MustCompile(`^\s*Conflict between rule (\d+) and token (\S+) resolved as (shift|reduce|an error) \(.*\)\.$`)

// The two reasons spec §8 allows a resolution to stay uncovered.
const (
	reasonTimedOut    = "reason: generation timed out"
	reasonUnreachable = "reason: the rendering cannot reach it"
)

// fixtureLines returns a file's lines less blank ones and # comments.
func fixtureLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var lines []string
	sc := bufio.NewScanner(f)
	sc.Buffer(nil, 1<<20)
	for sc.Scan() {
		if l := sc.Text(); l != "" && !strings.HasPrefix(l, "#") {
			lines = append(lines, l)
		}
	}
	return lines, sc.Err()
}

// precedenceCoverage checks the four fixture files in dir and returns how many resolutions a kept sentence covers,
// the resolutions ("rule token") covered by no kept sentence and on no reviewed list, and how many there are. A
// sentence is kept when its rendering round-trips (TestPrecedenceRenderingRoundTrips).
func precedenceCoverage(dir string) (covered int, unlisted []string, total int, err error) {
	read := func(name string) []string {
		if err != nil {
			return nil
		}
		var l []string
		l, err = fixtureLines(filepath.Join(dir, name))
		return l
	}
	conflicts, resolutions, sentences, uncovered := read("bison-conflicts.txt"), read("resolutions.txt"), read("sentences.txt"), read("uncovered-reviewed.txt")
	if err != nil {
		return 0, nil, 0, err
	}
	fail := func(format string, a ...any) (int, []string, int, error) { return 0, nil, 0, fmt.Errorf(format, a...) }

	action := map[string]string{} // "rule token" => shift, reduce or error
	for _, l := range conflicts {
		m := bisonConflict.FindStringSubmatch(l)
		if m == nil {
			return fail("bison-conflicts.txt: not a resolution: %q", l)
		}
		action[m[1]+" "+m[2]] = strings.TrimPrefix(m[3], "an ")
	}
	if len(action) != len(conflicts) {
		return fail("bison-conflicts.txt: %d lines, %d resolutions", len(conflicts), len(action))
	}

	verdict, kept := map[string]string{}, map[string]bool{}
	for _, l := range sentences {
		f := strings.Split(l, "\t")
		if len(f) != 5 || f[1] == "" || f[3] == "" || f[4] == "" {
			return fail("sentences.txt: want id, terminals, verdict, grouping, rendering: %q", l)
		}
		if _, dup := verdict[f[0]]; dup {
			return fail("sentences.txt: %s twice", f[0])
		}
		verdict[f[0]], kept[f[0]] = f[2], f[4] != "-"
	}

	listed := map[string]string{} // "rule token" => its reason
	reason := ""
	for _, l := range uncovered {
		switch {
		case strings.HasPrefix(l, "reason:"):
			if l != reasonTimedOut && l != reasonUnreachable {
				return fail("uncovered-reviewed.txt: %q is not a reason spec §8 allows", l)
			}
			reason = l
		case reason == "":
			return fail("uncovered-reviewed.txt: %q is under no reason", l)
		case action[l] == "":
			return fail("uncovered-reviewed.txt: %q is not a resolution in bison's report", l)
		case listed[l] != "":
			return fail("uncovered-reviewed.txt: %q listed twice", l)
		default:
			listed[l] = reason
		}
	}

	seen := map[string]bool{}
	for _, l := range resolutions {
		f := strings.Split(l, "\t")
		if len(f) != 4 {
			return fail("resolutions.txt: want rule, token, action, sentences: %q", l)
		}
		key := f[0] + " " + f[1]
		switch {
		case action[key] == "":
			return fail("resolutions.txt: %s is not in bison's report", key)
		case action[key] != f[2]:
			return fail("resolutions.txt: %s resolved as %s, bison's report says %s", key, f[2], action[key])
		case seen[key]:
			return fail("resolutions.txt: %s listed twice", key)
		}
		seen[key] = true
		anyKept, accepted := false, false
		if f[3] != "-" {
			for _, id := range strings.Split(f[3], ",") {
				v, ok := verdict[id]
				if !ok {
					return fail("resolutions.txt: %s names sentence %s, which sentences.txt lacks", key, id)
				}
				anyKept = anyKept || kept[id]
				accepted = accepted || kept[id] && v == "accept"
			}
		}
		switch {
		case anyKept && listed[key] != "":
			return fail("resolutions.txt: %s is covered and also on the uncovered list", key)
		case anyKept && f[2] == "error" && !accepted:
			return fail("resolutions.txt: %s is resolved as an error and no accepted neighbour covers it", key)
		case anyKept:
			covered++
		case listed[key] == "":
			unlisted = append(unlisted, key)
		case (listed[key] == reasonTimedOut) != (f[3] == "-"):
			return fail("resolutions.txt: %s is under %q with sentences %s", key, listed[key], f[3])
		}
	}
	for key := range action {
		if !seen[key] {
			return fail("resolutions.txt: %s from bison's report is missing", key)
		}
	}
	return covered, unlisted, len(action), nil
}

// startSelector is the terminal that selects the grammar's start rule and is no token (perly.go: "GRAMPROG is a start
// selector"), so it renders to nothing.
const startSelector = "GRAMPROG"

// precedenceSentence is one row of sentences.txt: its id, terminals, bison's verdict and grouping, and its rendering
// ("-" when not kept).
type precedenceSentence struct {
	id, verdict, grouping, rendering string
	terms                            []string
}

func readSentences(dir string) ([]precedenceSentence, error) {
	lines, err := fixtureLines(filepath.Join(dir, "sentences.txt"))
	if err != nil {
		return nil, err
	}
	var out []precedenceSentence
	for _, l := range lines {
		f := strings.Split(l, "\t")
		if len(f) < 4 {
			return nil, fmt.Errorf("sentences.txt: %q", l)
		}
		s := precedenceSentence{id: f[0], terms: strings.Fields(f[1]), verdict: f[2], grouping: f[3], rendering: "-"}
		if len(f) > 4 {
			s.rendering = f[4]
		}
		out = append(out, s)
	}
	return out, nil
}

// lexeme is a terminal's rendering: its one word, and the source that must precede a sentence holding it ("" for none)
// for the lexer to read the word as that terminal, such as `use feature "isa";` for isa.
type lexeme struct{ text, preamble string }

// readLexemes reads rendering.txt: per line a terminal, its one lexeme and optionally its preamble, tab separated.
func readLexemes(dir string) (map[string]lexeme, error) {
	lines, err := fixtureLines(filepath.Join(dir, "rendering.txt"))
	if err != nil {
		return nil, err
	}
	table := map[string]lexeme{}
	for _, l := range lines {
		f := strings.Split(l, "\t")
		if len(f) < 2 || len(f) > 3 || f[1] == "" || strings.Contains(f[1], " ") {
			return nil, fmt.Errorf("rendering.txt: want a terminal, one lexeme and optionally a preamble: %q", l)
		}
		if _, dup := table[f[0]]; dup {
			return nil, fmt.Errorf("rendering.txt: %s has two lexemes", f[0])
		}
		x := lexeme{text: f[1]}
		if len(f) == 3 {
			x.preamble = f[2]
		}
		table[f[0]] = x
	}
	return table, nil
}

// renderTerms is the lexemes of terms, one word per terminal, and the preamble they need (each distinct preamble once,
// in order); ok is false when a terminal has none.
func renderTerms(lexemes map[string]lexeme, terms []string) (preamble string, words []string, missing string, ok bool) {
	var preambles []string
	for _, t := range terms {
		if t == startSelector {
			continue
		}
		x, found := lexemes[t]
		if !found {
			return "", nil, t, false
		}
		words = append(words, x.text)
		if x.preamble != "" && !slices.Contains(preambles, x.preamble) {
			preambles = append(preambles, x.preamble)
		}
	}
	return strings.Join(preambles, " "), words, "", true
}

// builtForest is src's forest as forest.Build makes it from the chart, before resolve: the precedence constraint and
// its bison referee read the token sequence's grammar, which scope does not decide.
// A recognition error gives no forest, as forest.Parse's refusal did.
func builtForest(tab *grammar.Tables, src []byte) *forest.Forest {
	s, _ := lex.Lex(src, tab.Tokens)
	l := lex.Linear(s)
	c, err := recognize.Recognize(recognize.New(tab, recognize.Options{}), l)
	if err != nil {
		return nil
	}
	return forest.Build(tab, c, l)
}

// newPrecedence is g's precedence over tab: its levels, each rule's %prec, and the rules refused by precedence, which
// the tables leave out.
func newPrecedence(g *grammar.Grammar, tab *grammar.Tables) (*forest.Precedence, error) {
	if len(tab.Rules) != len(g.G1) {
		return nil, fmt.Errorf("%d table rules for %d grammar rules: a sequence rule moved the rule ids", len(tab.Rules), len(g.G1))
	}
	rulePrec := make([]string, len(g.G1))
	for r, rule := range g.G1 {
		rulePrec[r] = rule.Prec
	}
	return forest.NewPrecedence(tab, g.Levels, rulePrec, g.Refused)
}

// referee compares the chart with bison's committed verdict and grouping on the generated sentences.
type referee struct {
	tab      *grammar.Tables
	p        *forest.Precedence
	lexemes  map[string]lexeme
	terminal map[string]int32
	line     []int  // by rule: the perly.y alternative it cites
	carries  []bool // by rule: a %prec, or a last terminal with a level, as the referee's grouping trace reads it
}

// newReferee is g's referee over tab, rendering with dir's rendering.txt.
func newReferee(g *grammar.Grammar, tab *grammar.Tables, dir string) (*referee, error) {
	p, err := newPrecedence(g, tab)
	if err != nil {
		return nil, err
	}
	lexemes, err := readLexemes(dir)
	if err != nil {
		return nil, err
	}
	r := &referee{tab: tab, p: p, lexemes: lexemes, terminal: map[string]int32{}}
	for id, s := range tab.Symbols {
		if s.Terminal {
			r.terminal[s.Name] = int32(id)
		}
	}
	for i, rule := range g.G1 {
		ref, _, _ := strings.Cut(strings.TrimPrefix(rule.Cite, "perly.y:"), ",")
		n, err := strconv.Atoi(ref)
		if err != nil {
			return nil, fmt.Errorf("rule %d cites %q", i, rule.Cite)
		}
		last := ""
		for _, s := range slices.Backward(tab.Rules[i].RHS) {
			if tab.Symbols[s].Terminal {
				last = tab.Symbols[s].Name
				break
			}
		}
		r.line = append(r.line, n)
		r.carries = append(r.carries, rule.Prec != "" || slices.Contains(g.Prec, last))
	}
	return r, nil
}

// narrowed is s's rendering's Build forest before resolve, narrowed to s's own terminals (a token admits a set of
// terminals, and bison read one of them); nil where recognition or the narrowing leaves no derivation. skip is the
// token count of s's preamble.
func (r *referee) narrowed(s precedenceSentence) (*forest.Forest, int) {
	preamble, _, _, _ := renderTerms(r.lexemes, s.terms)
	skip := 0
	if preamble != "" {
		pre, _ := lex.Lex([]byte(preamble), r.tab.Tokens)
		skip = len(pre.Tokens)
	}
	var want []int32
	for _, name := range s.terms[1:] { // after GRAMPROG
		want = append(want, r.terminal[name])
	}
	f := builtForest(r.tab, []byte(s.rendering))
	if f == nil || !narrow(f, want, skip) {
		return nil, skip
	}
	return f, skip
}

// judge is the chart's verdict ("accept" or "refuse") and grouping on s's rendering: Build's forest before resolve,
// since bison's token choice stands in for scope, its derivations
// narrowed to s's own terminals (a token admits a set of terminals, and bison read one of them), then the constraint.
// The grouping is the referee's: each reduction of a rule carrying a precedence, line:from-to over terminal indexes
// (GRAMPROG is 0), children first, comma-separated, - for none; "ambiguous" where the derivations left disagree on it.
func (r *referee) judge(s precedenceSentence) (verdict, grouping string) {
	f, skip := r.narrowed(s)
	if f == nil || !f.Constrain(r.p) {
		return "refuse", "-"
	}
	trees, cut := f.Derivations(f.Root, 16)
	if cut {
		return "accept", "more than 16 derivations left, not all compared"
	}
	groupings := map[string]bool{}
	for _, tr := range trees {
		var g []string
		var walk func(*forest.Tree)
		walk = func(tr *forest.Tree) {
			for _, k := range tr.Kids {
				walk(k)
			}
			if tr.Rule >= 0 && r.carries[tr.Rule] && int(tr.From) >= skip { // bison reads no preamble
				g = append(g, fmt.Sprintf("%d:%d-%d", r.line[tr.Rule], tr.From-skip+1, tr.To-skip+1))
			}
		}
		walk(tr)
		if len(g) == 0 {
			g = []string{"-"}
		}
		groupings[strings.Join(g, ",")] = true
	}
	if len(groupings) != 1 {
		return "accept", fmt.Sprintf("ambiguous (%d groupings) %v", len(groupings), groupings)
	}
	for g := range groupings {
		grouping = g
	}
	return "accept", grouping
}

// narrow keeps the derivations of f whose leaves past the first skip tokens are want's terminals, one token each, and
// reports whether the root keeps one. A slipper ; is zero-width, and bison reads no ; that is not written.
func narrow(f *forest.Forest, want []int32, skip int) bool {
	alive := make([]bool, len(f.Nodes))
	for n, nd := range f.Nodes {
		i := int(nd.From) - skip
		alive[n] = n > 0 && nd.Alt == 0 && (i < 0 || nd.To == nd.From+1 && i < len(want) && nd.Sym == want[i])
	}
	ok := func(a forest.Alt) bool { return (a.Left == 0 || alive[a.Left]) && (a.Right == 0 || alive[a.Right]) }
	for changed := true; changed; { // to a fixed point: a unit rule's cycle keeps a node only through another way
		changed = false
		for n := range f.Nodes {
			for a := f.Nodes[n].Alt; a != 0 && !alive[n]; a = f.Alts[a].Next {
				if ok(f.Alts[a]) {
					alive[n], changed = true, true
				}
			}
		}
	}
	for n := range f.Nodes {
		if !alive[n] {
			continue // unreachable from a kept alternative
		}
		for a := &f.Nodes[n].Alt; *a != 0; {
			if ok(f.Alts[*a]) {
				a = &f.Alts[*a].Next
			} else {
				*a = f.Alts[*a].Next
			}
		}
	}
	return alive[f.Root]
}

// disagreement is how the chart differs from bison's verdict and grouping on s, or "" where it agrees. Two
// refusals are not compared by grouping, since bison's stops at its first error.
func (r *referee) disagreement(s precedenceSentence) string {
	v, g := r.judge(s)
	switch {
	case v == "refuse" && s.verdict == "accept":
		return fmt.Sprintf("%s %q: the chart refuses, bison accepts %s", s.id, s.rendering, s.grouping)
	case v == "accept" && s.verdict != "accept":
		return fmt.Sprintf("%s %q: the chart accepts %s, bison %s", s.id, s.rendering, g, s.verdict)
	case v == "accept" && g != s.grouping:
		return fmt.Sprintf("%s %q: the chart groups %s, bison %s", s.id, s.rendering, g, s.grouping)
	}
	return ""
}

// PrecedenceCounts are the referee's counts over a precedence fixture directory: the kept sentences whose chart
// verdict or grouping differs from bison's, and each disagreement; the bison resolutions a kept sentence covers; and
// those covered by no kept sentence and on no reviewed list.
type PrecedenceCounts struct {
	Disagree          []string
	Covered           int
	UncoveredUnlisted []string
}

// CountPrecedence referees every kept sentence in dir (testdata/precedence's layout) under g over tab. A fixture
// file that is missing, unreadable or malformed is an error, never a 0 count.
func CountPrecedence(g *grammar.Grammar, tab *grammar.Tables, dir string) (PrecedenceCounts, error) {
	covered, unlisted, _, err := precedenceCoverage(dir)
	if err != nil {
		return PrecedenceCounts{}, err
	}
	r, err := newReferee(g, tab, dir)
	if err != nil {
		return PrecedenceCounts{}, err
	}
	sentences, err := readSentences(dir)
	if err != nil {
		return PrecedenceCounts{}, err
	}
	c := PrecedenceCounts{Covered: covered, UncoveredUnlisted: unlisted}
	for _, s := range sentences {
		if s.rendering == "-" {
			continue
		}
		if d := r.disagreement(s); d != "" {
			c.Disagree = append(c.Disagree, d)
		}
	}
	return c, nil
}
