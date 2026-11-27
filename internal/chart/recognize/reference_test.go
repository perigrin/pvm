// ABOUTME: A textbook Earley recognizer, test-only, and the oracle that holds the recognizer to it.
// ABOUTME: It shares no code with the recognizer; the oracle compares acceptance and Expected under every Options combination.

package recognize

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/chart/prectab"
)

// refItem is a rule with a dot, the set it began in, and its context: the
// canonical spine top it was predicted under, an automaton item, or -1 for
// none, as every item's is where nothing is judged.
type refItem struct {
	rule, dot, origin int
	top               int32
}

// refChart is what the reference makes of a token sequence.
type refChart struct {
	sets     [][]refItem // by node: what is true after that many tokens
	inserted []int       // the nodes where a ; toke.c inserts was scanned, ascending
	accepted bool
	// expected is, by node, the terminals the judged reference expects there,
	// by name; nil where nothing is judged (refExpected reads the sets).
	expected [][]string
	// refusedCompletions and refusedScans count what precedence refused: a
	// completion into one parent before a lookahead some terminal of which can
	// follow its left-hand side, and an item's scan of one terminal.
	refusedCompletions, refusedScans int
}

// refRun is the unjudged reference's sets, inserted nodes and acceptance for
// the terminal-set ids ids (refJudged with no precedence).
func refRun(tab *grammar.Tables, ids []int) (sets [][]refItem, inserted []int, accepted bool) {
	rc := refJudged(tab, nil, ids)
	return rc.sets, rc.inserted, rc.accepted
}

// refJudged is the reference's Earley sets for the terminal-set ids ids, one
// per node of a linear lattice: sets[i] holds what is true after i tokens; and
// whether it accepts. It reads tab's rules and terminal sets only, never its
// states. Each set is closed by predicting and completing until nothing is
// added, so a nullable symbol needs no special case. Where the next token
// holds PERLY_BRACE_CLOSE, or the input ends, a ; toke.c inserts is scanned
// into a set of its own past the node's (spec §5.2); the next node scans its
// token from both, and at the end either may accept.
//
// With p it judges p's precedence (V2 design §4), from p's state-free tables
// alone. A start item's context is itself, a child predicted at its rule's
// first position inherits its parent's, and one predicted later takes the
// predicting item; a context that forbids no shift is none. A scan is refused
// where V1's shift verdict for (context, item, terminal) forbids it. A
// completion advances only the parents that predicted it under its context,
// and is refused where V1's reduce verdict forbids its rule before every
// terminal of the lookahead that can follow its left-hand side, or where none
// can, the parent item deciding a pair that needs it. The lookahead is the
// next token's terminals, with ; where one may be inserted, and at the end of
// input EOF and ;. Expected judges each candidate as the lookahead.
func refJudged(tab *grammar.Tables, p *prectab.Precedence, ids []int) (rc refChart) {
	nsym := tab.EOF + 1
	byLHS := make([][]int, nsym)
	for r, rule := range tab.Rules {
		byLHS[rule.LHS] = append(byLHS[rule.LHS], r)
	}
	follow := refFollow(tab)
	// canon is rule at dot as a context: its automaton item where that
	// forbids some shift, none otherwise.
	canon := func(rule, dot int) int32 {
		if p == nil {
			return -1
		}
		if i := p.Base[rule] + int32(dot); p.TopAt[i] >= 0 {
			return i
		}
		return -1
	}
	// under is the context it predicts its children under.
	under := func(it refItem) int32 {
		if it.dot == 0 {
			return it.top
		}
		return canon(it.rule, it.dot)
	}
	shiftForbidden := func(it refItem, t int) bool {
		return p != nil && p.ShiftForbidden(it.top, p.Base[it.rule]+int32(it.dot), int32(t))
	}
	// reduceForbidden judges completing it into its parent w before la, nil
	// judging nothing; vacuous where no terminal of la can follow it.
	reduceForbidden := func(it, w refItem, la []bool) (forbidden, vacuous bool) {
		if p == nil || la == nil {
			return false, false
		}
		lhs, pi := tab.Rules[it.rule].LHS, p.Base[w.rule]+int32(w.dot)
		vacuous = true
		for t, ok := range la {
			if ok && follow[lhs][t] {
				vacuous = false
				if p.ReducesFree(it.rule, t, pi) {
					return false, false
				}
			}
		}
		return true, vacuous
	}
	var all [][]refItem // every set, a node's, an inserted ;'s or Expected's; origins index it
	counting := true    // Expected's closures count nothing
	open := func(seed []refItem, la []bool) []refItem {
		k := len(all)
		all = append(all, nil)
		seen := map[refItem]bool{}
		refused := map[[2]refItem]bool{}
		put := func(it refItem) {
			if !seen[it] {
				seen[it] = true
				all[k] = append(all[k], it)
			}
		}
		for _, it := range seed {
			put(it)
		}
		for grew := true; grew; {
			grew = false
			n := len(all[k])
			for i := 0; i < len(all[k]); i++ {
				it := all[k][i]
				rule := tab.Rules[it.rule]
				if it.dot < len(rule.RHS) {
					for _, r := range byLHS[rule.RHS[it.dot]] { // predict
						put(refItem{r, 0, k, under(it)})
					}
					continue
				}
				for _, w := range all[it.origin] { // complete
					if rhs := tab.Rules[w.rule].RHS; w.dot < len(rhs) && rhs[w.dot] == rule.LHS && under(w) == it.top {
						if forbidden, vacuous := reduceForbidden(it, w, la); forbidden {
							if counting && !vacuous && !refused[[2]refItem{it, w}] {
								refused[[2]refItem{it, w}] = true
								rc.refusedCompletions++
							}
							continue
						}
						put(refItem{w.rule, w.dot + 1, w.origin, w.top})
					}
				}
			}
			grew = len(all[k]) > n
		}
		return all[k]
	}
	advance := func(from []refItem, terms []int) (out []refItem) {
		for _, it := range from {
			if rhs := tab.Rules[it.rule].RHS; it.dot < len(rhs) && slices.Contains(terms, rhs[it.dot]) {
				if shiftForbidden(it, rhs[it.dot]) {
					if counting {
						rc.refusedScans++
					}
					continue
				}
				out = append(out, refItem{it.rule, it.dot + 1, it.origin, it.top})
			}
		}
		return out
	}
	named := func(name string) int {
		return slices.IndexFunc(tab.Symbols, func(s grammar.TableSymbol) bool { return s.Terminal && s.Name == name })
	}
	semi, brace := named("PERLY_SEMICOLON"), named("PERLY_BRACE_CLOSE")
	laAt := func(i int) []bool {
		if p == nil {
			return nil
		}
		la := make([]bool, nsym)
		if i == len(ids) {
			la[tab.EOF] = true
		}
		for _, t := range terminalsAt(tab, ids, i) {
			la[t] = true
		}
		if semi >= 0 && (i == len(ids) || brace >= 0 && la[brace]) {
			la[semi] = true
		}
		return la
	}
	var start, ins []refItem
	for r, rule := range tab.Rules {
		if rule.LHS == tab.Start {
			start = append(start, refItem{r, 0, 0, canon(r, 0)})
		}
	}
	seeds := [][]refItem{start} // by node: what the start or a token's scan made
	rc.sets = append(rc.sets, open(start, laAt(0)))
	for i := 0; ; i++ {
		ins = nil
		if i == len(ids) || slices.Contains(tab.TerminalSets[ids[i]], brace) {
			if seed := advance(rc.sets[i], []int{semi}); len(seed) > 0 {
				ins = open(seed, laAt(i))
				rc.inserted = append(rc.inserted, i)
			}
		}
		if i == len(ids) {
			break
		}
		terms := tab.TerminalSets[ids[i]]
		seeds = append(seeds, append(advance(rc.sets[i], terms), advance(ins, terms)...))
		rc.sets = append(rc.sets, open(seeds[i+1], laAt(i+1)))
	}
	done := func(it refItem) bool {
		return it.origin == 0 && it.dot == len(tab.Rules[it.rule].RHS) && tab.Rules[it.rule].LHS == tab.Start
	}
	rc.accepted = slices.ContainsFunc(rc.sets[len(ids)], done) || slices.ContainsFunc(ins, done)
	if p == nil {
		return rc
	}
	counting = false
	awaits := func(items []refItem, t int) bool {
		return slices.ContainsFunc(items, func(it refItem) bool {
			rhs := tab.Rules[it.rule].RHS
			return it.dot < len(rhs) && rhs[it.dot] == t && !shiftForbidden(it, t)
		})
	}
	for _, seed := range seeds {
		var names []string
		candidates := open(seed, nil) // judging no completion, a superset of every candidate's closure
		// Its completions, each a completed item and the parent it advanced.
		var made [][2]refItem
		for _, it := range candidates {
			if rule := tab.Rules[it.rule]; it.dot == len(rule.RHS) {
				for _, w := range all[it.origin] {
					if rhs := tab.Rules[w.rule].RHS; w.dot < len(rhs) && rhs[w.dot] == rule.LHS && under(w) == it.top {
						made = append(made, [2]refItem{it, w})
					}
				}
			}
		}
		for t := range nsym {
			if t >= len(tab.Symbols) || !tab.Symbols[t].Terminal || !awaits(candidates, t) {
				continue
			}
			la := make([]bool, nsym)
			la[t] = true
			// Where t forbids none of those completions but vacuously, the closure before t awaits t exactly as
			// candidates does: an item awaiting t past a completion of A has t in FOLLOW(A).
			if !slices.ContainsFunc(made, func(c [2]refItem) bool { f, v := reduceForbidden(c[0], c[1], la); return f && !v }) ||
				awaits(open(seed, la), t) {
				names = append(names, tab.Symbols[t].Name)
			}
		}
		slices.Sort(names)
		rc.expected = append(rc.expected, names)
	}
	return rc
}

// terminalsAt is the terminals of token i of ids, none past the last.
func terminalsAt(tab *grammar.Tables, ids []int, i int) []int {
	if i == len(ids) {
		return nil
	}
	return tab.TerminalSets[ids[i]]
}

// refFollow is, by symbol, the terminals that can follow it, EOF after the
// start symbol, from tab's rules alone: the textbook nullable, FIRST and
// FOLLOW fixed points.
func refFollow(tab *grammar.Tables) [][]bool {
	nsym := tab.EOF + 1
	nullable := make([]bool, nsym)
	first, follow := make([][]bool, nsym), make([][]bool, nsym)
	for s := range nsym {
		first[s], follow[s] = make([]bool, nsym), make([]bool, nsym)
		if s < len(tab.Symbols) && tab.Symbols[s].Terminal {
			first[s][s] = true
		}
	}
	follow[tab.Start][tab.EOF] = true
	union := func(dst, src []bool) (grew bool) {
		for t, ok := range src {
			if ok && !dst[t] {
				dst[t], grew = true, true
			}
		}
		return grew
	}
	for grew := true; grew; {
		grew = false
		for _, rule := range tab.Rules {
			all := true
			for _, s := range rule.RHS {
				grew = union(first[rule.LHS], first[s]) || grew
				if all = nullable[s]; !all {
					break
				}
			}
			if all && !nullable[rule.LHS] {
				nullable[rule.LHS], grew = true, true
			}
			rest := follow[rule.LHS] // what follows RHS[i:] as the loop walks back
			restNullable := true
			after := make([]bool, nsym)
			for i := len(rule.RHS) - 1; i >= 0; i-- {
				s := rule.RHS[i]
				grew = union(follow[s], after) || grew
				if restNullable {
					grew = union(follow[s], rest) || grew
				}
				if !nullable[s] {
					after = slices.Clone(first[s])
					restNullable = false
				} else {
					union(after, first[s])
				}
			}
		}
	}
	return follow
}

// refExpected is the names of the terminals that some item of items awaits,
// sorted.
func refExpected(tab *grammar.Tables, items []refItem) []string {
	var names []string
	for _, it := range items {
		if rhs := tab.Rules[it.rule].RHS; it.dot < len(rhs) && tab.Symbols[rhs[it.dot]].Terminal {
			names = append(names, tab.Symbols[rhs[it.dot]].Name)
		}
	}
	slices.Sort(names)
	return slices.Compact(names)
}

// refAccepted is whether the reference accepts the terminal names words.
func refAccepted(tab *grammar.Tables, words string) bool {
	var ids []int
	for _, w := range strings.Fields(words) {
		ids = append(ids, termSet(tab, w))
	}
	_, _, ok := refRun(tab, ids)
	return ok
}

// TestReferenceRecognizerEdgeCases: the edges a textbook recognizer gets
// wrong. The empty input is accepted by a nullable start and refused by a
// non-nullable one, a unit cycle terminates, and left and right recursion
// accept their language and refuse a token past it.
func TestReferenceRecognizerEdgeCases(t *testing.T) {
	for _, tc := range []struct {
		name    string
		rules   []string
		accepts []string
		refuses []string
	}{
		{"nullable start", []string{"S ::= A S |"}, []string{"", "A", "A A"}, []string{"B"}},
		{"non-nullable start", []string{"S ::= A S | A"}, []string{"A", "A A"}, []string{""}},
		{"nullable through a chain", []string{"S ::= X Y", "X ::= Z", "Z ::=", "Y ::= A |"}, []string{"", "A"}, []string{"A A"}},
		{"nullable fixed point", []string{"S ::= X Y", "X ::=", "Y ::= X B"}, []string{"B"}, []string{"", "B B"}},
		{"start on a right-hand side", []string{"S ::= B S C | A"}, []string{"A", "B A C"}, []string{"B A", "B B A C"}},
		{"unit cycle", []string{"S ::= S | A"}, []string{"A"}, []string{"", "A A"}},
		{"left recursion", []string{"S ::= S A | A", "Z ::= B"}, []string{"A", "A A A"}, []string{"", "A A A B"}},
		{"right recursion", []string{"S ::= A S | A", "Z ::= B"}, []string{"A", "A A A"}, []string{"", "A A A B"}},
	} {
		tab := toyTables(t, tc.rules...)
		for _, in := range tc.accepts {
			if !refAccepted(tab, in) {
				t.Errorf("%s: the reference refused %q", tc.name, in)
			}
		}
		for _, in := range tc.refuses {
			if refAccepted(tab, in) {
				t.Errorf("%s: the reference accepted %q", tc.name, in)
			}
		}
	}
}

// declared is the names of the functions, methods and types file declares at
// package scope.
func declared(file string) (map[string]bool, error) {
	f, err := parser.ParseFile(token.NewFileSet(), file, nil, 0)
	if err != nil {
		return nil, err
	}
	out := map[string]bool{}
	for _, d := range f.Decls {
		switch d := d.(type) {
		case *ast.FuncDecl:
			out[d.Name.Name] = true
		case *ast.GenDecl:
			for _, s := range d.Specs {
				if ts, ok := s.(*ast.TypeSpec); ok {
					out[ts.Name.Name] = true
				}
			}
		}
	}
	return out, nil
}

// usedFrom is the names in banned that src, a file called file, mentions,
// sorted.
func usedFrom(file string, src any, banned map[string]bool) ([]string, error) {
	f, err := parser.ParseFile(token.NewFileSet(), file, src, 0)
	if err != nil {
		return nil, err
	}
	var out []string
	ast.Inspect(f, func(n ast.Node) bool {
		if id, ok := n.(*ast.Ident); ok && banned[id.Name] {
			out = append(out, id.Name)
		}
		return true
	})
	slices.Sort(out)
	return slices.Compact(out), nil
}

// TestReferenceSharesNoCode: reference_test.go mentions no function, method
// or type the package's non-test files declare, so the reference cannot
// agree with the recognizer by borrowing from it. Fed a synthetic reference
// that calls one, the check names it.
func TestReferenceSharesNoCode(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	banned := map[string]bool{}
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		names, err := declared(file)
		if err != nil {
			t.Fatal(err)
		}
		for n := range names {
			banned[n] = true
		}
	}
	if !banned["Recognize"] || !banned["Expected"] {
		t.Fatalf("the check found %d declarations and not Recognize and Expected", len(banned))
	}
	got, err := usedFrom("reference_test.go", nil, banned)
	if err != nil || len(got) > 0 {
		t.Errorf("reference_test.go uses the recognizer's %v (err %v)", got, err)
	}
	got, err = usedFrom("synthetic.go", "package p\nfunc ref() { c.Accepted(); New(nil, Options{}); mine() }", banned)
	if err != nil || !slices.Equal(got, []string{"Accepted", "New", "Options"}) {
		t.Errorf("the synthetic reference's borrowings were %v (err %v), want Accepted, New and Options", got, err)
	}
}
