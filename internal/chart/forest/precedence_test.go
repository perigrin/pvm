// ABOUTME: Tests the precedence constraint against bison builds of toy grammars (testdata/precedence/toys, from offline/referee/toys).
// ABOUTME: Same verdict and grouping on every sentence; doubled prefixes kept, inner forbidden spans refuse, and no dependence on Leo or alt order.

package forest

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/chart/recognize"
)

// toyDir holds the toy fixtures, written by offline/referee/toys/build.sh.
var toyDir = filepath.Join("..", "testdata", "precedence", "toys")

// toyNames are the toys build.sh writes; a missing or extra fixture fails TestPrecedenceToyFixturesNonEmpty.
var toyNames = []string{"empty", "emptyred", "follow", "indirob", "indirobsort", "ladder", "lalr", "mixed", "nonassoc", "nullpred", "nullsuf", "prefix", "twostates"}

// toySentence is one fixture line: bison's verdict, its grouping (- where refused) and the terminals.
type toySentence struct {
	verdict, grouping string
	words             []string
}

// toy is a fixture's grammar, its precedence and its sentences.
type toy struct {
	tab       *grammar.Tables
	prec      *Precedence
	sentences []toySentence
}

// loadToy reads a toy fixture: `level <assoc> TOKEN...` loosest first, `rule <n> <lhs>: <rhs> [%prec X]` in order,
// then `<verdict>\t<grouping>\t<terminals>` per sentence.
func loadToy(t *testing.T, name string) toy {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(toyDir, name+".txt"))
	if err != nil {
		t.Fatal(err)
	}
	var levels []grammar.Level
	var statements, rulePrec []string
	var out toy
	for _, line := range strings.Split(strings.TrimSuffix(string(data), "\n"), "\n") {
		f := strings.Fields(line)
		switch {
		case strings.HasPrefix(line, "#"):
		case f[0] == "level":
			levels = append(levels, grammar.Level{N: len(levels) + 1, Assoc: f[1], Tokens: f[2:]})
		case f[0] == "rule":
			lhs, rhs, _ := strings.Cut(strings.Join(f[2:], " "), ":")
			rhs, prec, _ := strings.Cut(rhs, "%prec")
			statements = append(statements, lhs+" ::= "+rhs)
			rulePrec = append(rulePrec, strings.TrimSpace(prec))
		default:
			p := strings.Split(line, "\t")
			if len(p) != 3 {
				t.Fatalf("%s: not verdict, grouping and terminals: %q", name, line)
			}
			out.sentences = append(out.sentences, toySentence{p[0], p[1], strings.Fields(p[2])})
		}
	}
	out.tab = toyTables(t, statements...)
	prec, err := NewPrecedence(out.tab, levels, rulePrec, nil)
	if err != nil {
		t.Fatal(err)
	}
	out.prec = prec
	return out
}

// constrained is the chart's verdict on words under opts with the precedence constraint, and its grouping:
// each rule's reduction as rule:from-to in bison's order (children first, left to right), - where refused,
// and "ambiguous" where more than one derivation is left. reverse reverses every node's alternatives first.
// constrain is the judgement applied: (*Forest).Constrain or (*Forest).ConstrainReference.
func (ty toy) constrained(t *testing.T, constrain func(*Forest, *Precedence) bool, words []string, opts recognize.Options, reverse bool) (verdict, grouping string) {
	t.Helper()
	f, _, _ := toyForest(t, ty.tab, words, opts)
	if f == nil {
		return "refuse", "-"
	}
	if reverse {
		reverseAlts(f)
	}
	if !constrain(f, ty.prec) {
		return "refuse", "-"
	}
	trees, _ := f.Derivations(f.Root, 2)
	if len(trees) != 1 {
		return "accept", fmt.Sprintf("ambiguous (%d derivations)", len(trees))
	}
	var g []string
	var walk func(*Tree)
	walk = func(tr *Tree) {
		for _, k := range tr.Kids {
			walk(k)
		}
		if tr.Rule >= 0 {
			g = append(g, fmt.Sprintf("%d:%d-%d", tr.Rule, tr.From, tr.To))
		}
	}
	walk(trees[0])
	return "accept", strings.Join(g, ",")
}

// reverseAlts reverses the order of every node's alternatives.
func reverseAlts(f *Forest) {
	for n := range f.Nodes {
		var alts []int32
		for a := f.Nodes[n].Alt; a != 0; a = f.Alts[a].Next {
			alts = append(alts, a)
		}
		f.Nodes[n].Alt = 0
		for _, a := range alts {
			f.Alts[a].Next = f.Nodes[n].Alt
			f.Nodes[n].Alt = a
		}
	}
}

// TestPrecedenceToyReferees runs the constraint on every toy sentence and requires bison's verdict and grouping.
func TestPrecedenceToyReferees(t *testing.T) {
	toyReferees(t, (*Forest).Constrain)
}

// TestPrecedenceReferenceAgreesWithBison holds the reference, the fixed point the differential checks compare
// production against, to bison's verdict and grouping on every toy sentence.
func TestPrecedenceReferenceAgreesWithBison(t *testing.T) {
	toyReferees(t, (*Forest).ConstrainReference)
}

// toyReferees requires constrain to give bison's verdict and grouping on every toy sentence, one subtest a toy.
func toyReferees(t *testing.T, constrain func(*Forest, *Precedence) bool) {
	for _, name := range toyNames {
		t.Run(name, func(t *testing.T) {
			ty := loadToy(t, name)
			bad := 0
			for _, s := range ty.sentences {
				v, g := ty.constrained(t, constrain, s.words, recognize.Options{}, false)
				if v != s.verdict || g != s.grouping {
					if bad++; bad <= 10 {
						t.Errorf("%s: got %s %s, bison %s %s", strings.Join(s.words, " "), v, g, s.verdict, s.grouping)
					}
				}
			}
			if bad > 0 {
				t.Errorf("%d of %d sentences differ from bison", bad, len(ty.sentences))
			}
		})
	}
}

// TestPrecedenceToyFixturesNonEmpty requires each toy's fixture, with accepted and refused sentences, so the
// referee test cannot pass on nothing.
func TestPrecedenceToyFixturesNonEmpty(t *testing.T) {
	paths, err := filepath.Glob(filepath.Join(toyDir, "*.txt"))
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, p := range paths {
		got = append(got, strings.TrimSuffix(filepath.Base(p), ".txt"))
	}
	if !slices.Equal(got, toyNames) {
		t.Fatalf("toy fixtures %v, want %v", got, toyNames)
	}
	for _, name := range toyNames {
		ty := loadToy(t, name)
		n := map[string]int{}
		for _, s := range ty.sentences {
			n[s.verdict]++
		}
		if n["accept"] == 0 || n["refuse"] == 0 {
			t.Errorf("%s: %d accepted, %d refused; want both above 0", name, n["accept"], n["refuse"])
		}
	}
}

// sentence is the fixture sentence words of ty, failing where there is none.
func (ty toy) sentence(t *testing.T, words string) toySentence {
	t.Helper()
	for _, s := range ty.sentences {
		if strings.Join(s.words, " ") == words {
			return s
		}
	}
	t.Fatalf("no fixture sentence %q", words)
	return toySentence{}
}

// TestPrecedenceToyPrefixAfterPrefixAccepted requires the prefix toy's doubled prefixes (- - 1; and ! ! $x;'s shape,
// and ++ ++ at a %nonassoc level) accepted as bison accepts them: prefix after prefix is never a conflict.
func TestPrecedenceToyPrefixAfterPrefixAccepted(t *testing.T) {
	ty := loadToy(t, "prefix")
	prefixes := []string{"MINUS", "BANG", "PREINC"}
	n := 0
	for _, s := range ty.sentences {
		doubled := false
		for i := 1; i < len(s.words); i++ {
			doubled = doubled || slices.Contains(prefixes, s.words[i-1]) && slices.Contains(prefixes, s.words[i])
		}
		if !doubled || s.verdict != "accept" {
			continue
		}
		n++
		if v, g := ty.constrained(t, (*Forest).Constrain, s.words, recognize.Options{}, false); v != "accept" || g != s.grouping {
			t.Errorf("%s: got %s %s, bison accepts %s", strings.Join(s.words, " "), v, g, s.grouping)
		}
	}
	for _, w := range []string{"MINUS MINUS NUM", "BANG BANG NUM", "PREINC PREINC NUM"} {
		if ty.sentence(t, w).verdict != "accept" {
			t.Errorf("bison refuses %s", w)
		}
	}
	if n == 0 {
		t.Fatal("no accepted sentence holds a doubled prefix")
	}
}

// TestPrecedenceForbiddenInnerSpanRefusesFile requires a refusal where one span's every derivation is
// forbidden (1 .. 2 .. 3, a %nonassoc chain) though the file's span has derivations that do not hold it.
func TestPrecedenceForbiddenInnerSpanRefusesFile(t *testing.T) {
	ty := loadToy(t, "nonassoc")
	words := strings.Fields("NUM DOTDOT NUM DOTDOT NUM OR NUM")
	if s := ty.sentence(t, strings.Join(words, " ")); s.verdict != "refuse" {
		t.Fatalf("bison %s, want refuse", s.verdict)
	}
	f, _, _ := toyForest(t, ty.tab, words, recognize.Options{})
	if f == nil {
		t.Fatal("the chart refuses it")
	}
	trees, _ := f.Derivations(f.Root, 100)
	without := 0
	for _, tr := range trees {
		if !slices.ContainsFunc(tr.Kids, func(k *Tree) bool { return k.From == 0 && k.To == 5 }) {
			without++
		}
	}
	if without == 0 || without == len(trees) {
		t.Fatalf("%d of %d derivations hold no node over 1 .. 2 .. 3; want some, not all", without, len(trees))
	}
	if inner, _ := ty.constrained(t, (*Forest).Constrain, words[:5], recognize.Options{}, false); inner != "refuse" {
		t.Fatalf("1 .. 2 .. 3 alone: %s, want refuse", inner)
	}
	if v, _ := ty.constrained(t, (*Forest).Constrain, words, recognize.Options{}, false); v != "refuse" {
		t.Errorf("got %s, want refuse", v)
	}
}

// TestPrecedenceConstraintOrderInvariant requires every toy sentence's verdict and grouping under each recognizer
// option set, Leo on and off, with each node's alternatives in either order.
func TestPrecedenceConstraintOrderInvariant(t *testing.T) {
	for _, name := range toyNames {
		ty := loadToy(t, name)
		for _, s := range ty.sentences {
			v0, g0 := ty.constrained(t, (*Forest).Constrain, s.words, recognize.Options{}, false)
			for _, opts := range optionSets {
				for _, rev := range []bool{false, true} {
					if v, g := ty.constrained(t, (*Forest).Constrain, s.words, opts, rev); v != v0 || g != g0 {
						t.Errorf("%s %s under %+v, reversed %v: %s %s, want %s %s", name, strings.Join(s.words, " "), opts, rev, v, g, v0, g0)
					}
				}
			}
		}
	}
}

// TestPrecedenceToyResolutionNeedsConflict requires indirob: scalar %prec PREC_LOW's resolution against [ only
// where bison's conflict arises: after LSTOP (print $fh [] refused), never after SNAIL (@$array[1..3], T1/array.t).
func TestPrecedenceToyResolutionNeedsConflict(t *testing.T) {
	ty := loadToy(t, "indirob")
	for _, c := range []struct{ words, verdict string }{
		{"SNAIL VAR LB NUM RB", "accept"},
		{"LSTOP SNAIL VAR LB NUM RB", "accept"},
		{"LSTOP VAR LB NUM RB", "accept"},
		{"LSTOP VAR LB RB", "refuse"},
	} {
		s := ty.sentence(t, c.words)
		if s.verdict != c.verdict {
			t.Fatalf("%s: bison %s, want %s", c.words, s.verdict, c.verdict)
		}
		if v, g := ty.constrained(t, (*Forest).Constrain, s.words, recognize.Options{}, false); v != s.verdict || g != s.grouping {
			t.Errorf("%s: got %s %s, bison %s %s", c.words, v, g, s.verdict, s.grouping)
		}
	}
}

// TestPrecedenceRefusedRuleUnknownSymbol requires a refused rule naming a symbol the tables lack to be an error naming
// it, never a silent symbol 0.
func TestPrecedenceRefusedRuleUnknownSymbol(t *testing.T) {
	ty := loadToy(t, "nonassoc")
	bad := grammar.Rule{LHS: ty.tab.Symbols[ty.tab.Start].Name, RHS: []grammar.Symbol{{Kind: grammar.Name, Text: "NOSUCHSYMBOL"}}}
	if _, err := NewPrecedence(ty.tab, nil, make([]string, len(ty.tab.Rules)), []grammar.Rule{bad}); err == nil || !strings.Contains(err.Error(), "NOSUCHSYMBOL") {
		t.Fatalf("got %v, want an error naming NOSUCHSYMBOL", err)
	}
}

// guardToy is a toy grammar with a state-dependent key: a: X %prec LOW meets T in a conflict the shift wins after
// ONE (s: ONE X T) and in no conflict after TWO, and both states hold the parent item e: . a, so not even the
// parent item settles it. With settled, TWO is followed by f, whose parent item f: . a settles it.
func guardToy(t *testing.T, settled bool) (*grammar.Tables, []grammar.Level, []string) {
	after := "e"
	if settled {
		after = "f"
	}
	tab := toyTables(t, "s ::= ONE e T", "s ::= ONE X T", "s ::= TWO "+after, "e ::= a", "f ::= a", "a ::= X")
	levels := []grammar.Level{{N: 1, Assoc: "left", Tokens: []string{"LOW"}}, {N: 2, Assoc: "left", Tokens: []string{"T"}}}
	return tab, levels, []string{"", "", "", "", "", "LOW"}
}

// TestPrecedenceGenerationGuard requires generation to refuse the guard toy with an error naming its key, and
// perly.y to generate without one.
func TestPrecedenceGenerationGuard(t *testing.T) {
	tab, levels, rulePrec := guardToy(t, false)
	_, err := NewPrecedence(tab, levels, rulePrec, nil)
	if err == nil {
		t.Fatal("the guard toy generated tables; want an error naming its state-dependent key")
	}
	for _, part := range []string{"a: X .", "before T", "e: . a"} {
		if !strings.Contains(err.Error(), part) {
			t.Errorf("error %q does not name %q", err, part)
		}
	}
	if _, err := ParsePrecedence(committedTables(t)); err != nil {
		t.Errorf("perly.y: %v", err)
	}
}

// TestPrecedenceGenerationErrorRefusesEveryFile requires the guard toy's generation error, standing as the
// tables' precedence, to refuse every file the parse path is given, each of which it accepts without it.
func TestPrecedenceGenerationErrorRefusesEveryFile(t *testing.T) {
	toyTab, levels, rulePrec := guardToy(t, false)
	_, genErr := NewPrecedence(toyTab, levels, rulePrec, nil)
	if genErr == nil {
		t.Fatal("the guard toy generated tables")
	}
	g, ref, err := grammar.Committed()
	if err != nil {
		t.Fatal(err)
	}
	tab, err := grammar.Generate(g, ref) // its own value, whose precedence the toy's error stands as
	if err != nil {
		t.Fatal(err)
	}
	grammar.Derived(tab, func(*grammar.Tables) parsePrecedence { return parsePrecedence{err: genErr} })
	for _, src := range []string{"my $x = 1;\n", "print $x + 2 * 3;\n", "foo(1, 2) if $y;\n", "{ 1 }"} {
		if f, _ := Parse(committedTables(t), []byte(src), Scope{}); f == nil {
			t.Fatalf("%q: refused without the error", src)
		}
		if f, r := Parse(tab, []byte(src), Scope{}); f != nil || r != (Refusal{Reason: genErr.Error()}) {
			t.Errorf("%q: forest %v, refusal %+v; want none and the generation error", src, f != nil, r)
		}
	}
}

// parentPairs is p's (rule, terminal) pairs whose reduction its parent item decides, as "rule before terminal".
func parentPairs(p *Precedence) []string {
	var out []string
	for k, v := range p.reduceAt {
		if v != 2 {
			continue
		}
		r, x := k/p.nt, k%p.nt
		rhs := p.tab.Rules[r].RHS
		words := []string{p.tab.Symbols[p.tab.Rules[r].LHS].Name + ":"}
		for _, s := range rhs {
			words = append(words, p.tab.Symbols[s].Name)
		}
		out = append(out, strings.Join(words, " ")+" before "+p.tab.Symbols[x].Name)
	}
	return out
}

// TestPrecedenceParentItemPairsDerived requires generation to find the pairs that need the parent item: on perly.y
// six, every rule of them indirob's or myterm's and every terminal [ or {, both of each; on the settled guard toy
// its own pair. A list written for perly.y passes on nothing else.
func TestPrecedenceParentItemPairsDerived(t *testing.T) {
	p, err := ParsePrecedence(committedTables(t))
	if err != nil {
		t.Fatal(err)
	}
	got := parentPairs(p)
	if len(got) != 6 {
		t.Errorf("perly.y: %d pairs %q, want 6", len(got), got)
	}
	seen := map[string]bool{}
	for _, pair := range got {
		lhs, _, _ := strings.Cut(pair, ":")
		_, x, _ := strings.Cut(pair, " before ")
		if !slices.Contains([]string{"indirob", "myterm"}, lhs) || !slices.Contains([]string{"PERLY_BRACKET_OPEN", "PERLY_BRACE_OPEN"}, x) {
			t.Errorf("perly.y: %s is not indirob or myterm before [ or {", pair)
		}
		seen[lhs+" "+x] = true
	}
	if len(seen) != 4 {
		t.Errorf("perly.y: %d of indirob and myterm before [ and {, want all 4", len(seen))
	}
	tab, levels, rulePrec := guardToy(t, true)
	toy, err := NewPrecedence(tab, levels, rulePrec, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := parentPairs(toy), []string{"a: X before T"}; !slices.Equal(got, want) {
		t.Errorf("settled toy: %q, want %q", got, want)
	}
}

// TestPrecedenceTablesDense requires every state-free table on Precedence to be a slice, read by index: a map
// there would put a hash lookup on the evaluation path.
func TestPrecedenceTablesDense(t *testing.T) {
	ty := reflect.TypeFor[Precedence]()
	for _, name := range []string{"base", "reduceAt", "mixAt", "parentForbid", "needParent", "topAt", "shiftAt", "shiftForbid"} {
		f, ok := ty.FieldByName(name)
		if !ok {
			t.Errorf("Precedence has no %s", name)
			continue
		}
		elem := f.Type
		for elem.Kind() == reflect.Slice {
			elem = elem.Elem()
		}
		if f.Type.Kind() != reflect.Slice || elem.Kind() == reflect.Map {
			t.Errorf("%s is %s; want a slice holding no map", name, f.Type)
		}
	}
}
