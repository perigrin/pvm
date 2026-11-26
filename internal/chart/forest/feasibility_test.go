package forest

// The V2 design's §3 feasibility counters, pinned as tests over the committed perly.y tables (design §5): each reads
// 0 on perly.y and counts every synthetic case built to trigger it. Ported from the scratch code measured in
// docs/assessments/chart-precedence-v2/feasibility.md.

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/chart/prectab"
)

func feasibilityTables(t *testing.T) (*grammar.Tables, *prectab.Precedence) {
	t.Helper()
	tab := committedTables(t)
	p, err := prectab.ParsePrecedence(tab)
	if err != nil {
		t.Fatal(err)
	}
	return tab, p
}

// perlyFeasibility is perly.y's tables, their Precedence and the counts prectab.Feasible makes over them.
func perlyFeasibility(t *testing.T) (*grammar.Tables, *prectab.Precedence, prectab.Feasibility) {
	t.Helper()
	tab, p := feasibilityTables(t)
	v, _ := prectab.Feasible(p)
	return tab, p, v
}

// toyFeasibility is prectab.Feasible's counts over a toy's Precedence, which the toy's loader built.
func toyFeasibility(_ *grammar.Tables, p *prectab.Precedence) prectab.Feasibility {
	v, _ := prectab.Feasible(p)
	return v
}

// TestFeasibleRefusesFailingToys holds prectab.Feasible, the six counters as one refusal: perly.y passes, and each
// toy built to trip a counter is refused, naming that counter, though NewPrecedence still loaded it, so Constrain
// still judges it (design §4.1, F2-10).
func TestFeasibleRefusesFailingToys(t *testing.T) {
	toyOf := func(name string) func(*testing.T) (*grammar.Tables, *prectab.Precedence) {
		return func(t *testing.T) (*grammar.Tables, *prectab.Precedence) {
			ty := loadToy(t, name)
			return ty.tab, ty.prec
		}
	}
	const offAt, nullable, scan = "forbidden shifts off at's next terminal", "all-nullable rules with a non-allow reduce verdict", "core items disagreeing on one scan"
	for _, c := range []struct {
		name, counter string // counter "" for a grammar Feasible accepts
		load          func(*testing.T) (*grammar.Tables, *prectab.Precedence)
	}{
		{"perly", "", feasibilityTables},
		{"emptyred", offAt, toyOf("emptyred")},
		{"nullableShift", offAt, nullableShiftToy},
		{"MarkFirstFilter", offAt, markFirstFilterToy},
		{"empty", nullable, toyOf("empty")},
		{"scancollide", scan, scancollideToy},
	} {
		t.Run(c.name, func(t *testing.T) {
			_, p := c.load(t)
			if p == nil || len(p.ReduceAt) == 0 {
				t.Fatalf("%s: NewPrecedence loaded no tables", c.name)
			}
			_, err := prectab.Feasible(p)
			t.Logf("%s: %v", c.name, err)
			switch {
			case c.counter == "" && err != nil:
				t.Errorf("%s: refused, %v; want it accepted", c.name, err)
			case c.counter != "" && (err == nil || !strings.Contains(err.Error(), c.counter)):
				t.Errorf("%s: %v; want a refusal naming %s", c.name, err, c.counter)
			}
		})
	}
}

// Counter 1 (feasibility.md §1): no state reduces two rules of one LHS whose V1 reduce verdicts differ before a
// terminal of FOLLOW(LHS), so completion can be judged per LHS.
func TestFeasibilityRuleReduceConflicts(t *testing.T) {
	tab, p, v := perlyFeasibility(t)
	n, agree := v.RuleReduceConflicts, v.TwinAgree
	for _, s := range agree {
		t.Logf("twin rule agrees: state %d", s)
	}
	t.Logf("perly.y: %d of %d states", n, len(tab.States))
	if n != 0 || len(tab.States) == 0 {
		t.Errorf("rule-reduce conflicts in %d of %d states", n, len(tab.States))
	}
	if len(agree) == 0 {
		t.Error("no state reduces two rules of one LHS: the counter examined nothing")
	}

	// Every same-LHS rule pair of perly.y whose verdicts differ before some terminal of FOLLOW, each in a state of its own.
	byLHS := map[int][]int{}
	for r, rule := range tab.Rules {
		byLHS[rule.LHS] = append(byLHS[rule.LHS], r)
	}
	var synth []grammar.State
	for lhs, rs := range byLHS {
		for i := range rs {
			for j := i + 1; j < len(rs); j++ {
				for _, x := range tab.Follow[lhs] {
					if prectab.ReduceDiffers(p, rs[i], x, rs[j], x) {
						synth = append(synth, grammar.State{Reduces: []int{rs[i], rs[j]}})
						break
					}
				}
			}
		}
	}
	sn, _ := prectab.RuleReduceConflicts(p, synth)
	t.Logf("synthetic: %d of %d", sn, len(synth))
	if sn != len(synth) || len(synth) != 649 {
		t.Errorf("synthetic same-LHS pairs counted: %d of %d, want 649 of 649", sn, len(synth))
	}
}

// Counter 2 (feasibility.md §2): no state holds two parent items waiting on one parent-dependent symbol whose
// parentForbid verdicts disagree, so a parent-dependent verdict is a function of the parent state.
func TestFeasibilityParentConflicts(t *testing.T) {
	tab, p, v := perlyFeasibility(t)
	n := v.ParentConflicts
	t.Logf("perly.y: %d of %d states", n, len(tab.States))
	if n != 0 || len(tab.States) == 0 {
		t.Errorf("parent conflicts in %d of %d states", n, len(tab.States))
	}
	// The states where two core items wait on one parent-dependent symbol: those the counter actually judged.
	meet := v.ParentMeet
	t.Logf("meeting states: %s", strings.Trim(fmt.Sprint(meet), "[]"))
	if !slices.Equal(meet, []int{105, 248}) {
		t.Errorf("states with two items waiting on one parent-dependent symbol: %v, want [105 248]", meet)
	}

	// Every pair of perly.y's items waiting on one parent-dependent symbol with differing verdicts, each in a state of its own.
	all := make([]int, len(tab.Items))
	for i := range all {
		all[i] = i
	}
	var synth []grammar.State
	for b, is := range prectab.Awaiting(p, all) {
		for x := range is {
			for y := x + 1; y < len(is); y++ {
				if prectab.ParentDiffers(p, b, is[x], is[y]) {
					synth = append(synth, grammar.State{Items: []int{is[x], is[y]}})
				}
			}
		}
	}
	sn := prectab.ParentConflicts(p, synth)
	t.Logf("synthetic: %d of %d", sn, len(synth))
	if sn != len(synth) || len(synth) != 18 {
		t.Errorf("synthetic parent-item pairs counted: %d of %d, want 18 of 18", sn, len(synth))
	}
}

// Counter 3 (feasibility.md §3.2): context conflicts, three checks over the static model of each rule's canonical
// spine tops (F7-A1).
func TestFeasibilityContextConflicts(t *testing.T) {
	tab, p, v := perlyFeasibility(t)
	// perly.y's first forbidden shift: f after rule-position item at, under spine top top. The synthetic cases use it.
	var at, top, f int32 = -1, -1, -1
find:
	for i := range p.TopAt {
		if p.TopAt[i] < 0 {
			continue
		}
		for x := range tab.Items {
			for y := range p.NT {
				if p.ShiftForbidden(int32(i), int32(x), int32(y)) {
					top, at, f = int32(i), int32(x), int32(y)
					break find
				}
			}
		}
	}
	if at < 0 {
		t.Fatal("perly.y's tables forbid no shift")
	}

	// (a) No state holds core items with different canonical tops that give different shift verdicts.
	t.Run("a", func(t *testing.T) {
		meet, n := v.CoreTopsMeet, v.ContextA
		t.Logf("core tops meet: %s", strings.Trim(fmt.Sprint(meet), "[]"))
		t.Logf("perly.y: %d of %d states", n, len(tab.States))
		if n != 0 || len(meet) == 0 {
			t.Errorf("context conflicts in %d of %d states (core tops meet in %d)", n, len(tab.States), len(meet))
		}

		// TestZZM3NonVacuous's cases, on perly.y's first forbidden shift.
		other := (at + 7) % int32(len(tab.Items))
		cases := []struct {
			name     string
			xs       []prectab.ShiftRec
			want     prectab.ShiftVerdict
			conflict bool
		}{
			{"cross-core", []prectab.ShiftRec{{At: at, Top: top, F: f}, {At: other, Top: -1, F: f}}, prectab.ShiftVerdict{Multi: true, Cross: true, Copies: true, OverKeep: true}, true},
			{"same-core", []prectab.ShiftRec{{At: at, Top: top, F: f}, {At: at, Top: -1, F: f}}, prectab.ShiftVerdict{Multi: true, Same: true, Copies: true}, false},
			{"one shared top", []prectab.ShiftRec{{At: at, Top: top, F: f}, {At: other, Top: top, F: f}}, prectab.ShiftVerdict{}, false},
		}
		counted := 0
		for _, c := range cases {
			got := prectab.JudgeShifts(p, c.xs)
			if got == c.want && got.ContextConflict() == c.conflict {
				counted++
			} else {
				t.Errorf("synthetic %s: got %+v (conflict %v), want %+v (conflict %v)", c.name, got, got.ContextConflict(), c.want, c.conflict)
			}
		}
		t.Logf("synthetic: %d of %d", counted, len(cases))
	})

	// (b) No predictor begins rules under two different non-none canonical tops.
	t.Run("b", func(t *testing.T) {
		n, mixed := v.ContextB, v.MixedNone
		t.Logf("predictors mixing a top with none: %s", strings.Trim(fmt.Sprint(mixed), "[]"))
		t.Logf("perly.y: %d of %d predictors", n, v.Predictors)
		if n != 0 || len(mixed) == 0 {
			t.Errorf("predictors beginning rules under two tops: %d of %d (mixing with none: %d)", n, v.Predictors, len(mixed))
		}

		// A predictor beginning one rule under top and another under a second non-none top.
		second := -1
		for i, row := range p.TopAt {
			if row >= 0 && int32(i) != top {
				second = i
				break
			}
		}
		synth := map[int]map[int32]bool{0: {top: true}, 1: {int32(second): true}}
		sn := 0
		if conflict, _ := prectab.PredictorTopsConflict(synth); conflict {
			sn++
		}
		t.Logf("synthetic: %d of 1", sn)
		if sn != 1 {
			t.Errorf("synthetic predictor under tops %d and %d not counted", top, second)
		}
	})

	// (c) No rule begun under one predictor carries two tops whose verdicts disagree.
	t.Run("c", func(t *testing.T) {
		n, multi, real := v.ContextC, v.MultiTop, v.RealShifts
		t.Logf("predictors beginning a rule under two tops: %d", multi)
		t.Logf("real shifts: %d", real)
		t.Logf("perly.y: %d of %d predictors", n, v.Predictors)
		if real == 0 {
			t.Error("no real (item, terminal) shift judged: the check examined nothing")
		}
		if n != 0 || multi == 0 {
			t.Errorf("predictors with a rule whose tops disagree: %d of %d (with a rule under two tops: %d)", n, v.Predictors, multi)
		}

		// at's rule begun under top, which forbids shifting f after at, and under none, which allows it.
		synth := map[int]map[int32]bool{tab.Items[at].Rule: {top: true, -1: true}}
		sn := 0
		if hit, _ := prectab.RuleTopsDisagree(p, synth, [][2]int{{int(at), int(f)}}); hit {
			sn++
		}
		t.Logf("synthetic: %d of 1", sn)
		if sn != 1 {
			t.Error("synthetic rule under two disagreeing tops not counted")
		}
	})
}

// Counter 4 (design §5): every forbidden (top, at, f) shift has f equal to at's own next terminal, so the item
// that scans f holds at.
func TestFeasibilityShiftsOwnNextTerminal(t *testing.T) {
	_, _, v := perlyFeasibility(t)
	forbidden, off := v.ForbiddenShifts, v.OffAtShifts
	t.Logf("perly.y: %d of %d", off, forbidden)
	if off != 0 || forbidden != 674 {
		t.Errorf("forbidden shifts whose terminal is not at's own next terminal: %d of %d, want 0 of 674", off, forbidden)
	}

	// The toys whose forbidden shifts are scanned by an item that does not hold at (design §4.1).
	for _, c := range []struct {
		name string
		load func(*testing.T) (*grammar.Tables, *prectab.Precedence)
	}{
		{"emptyred", func(t *testing.T) (*grammar.Tables, *prectab.Precedence) {
			ty := loadToy(t, "emptyred")
			return ty.tab, ty.prec
		}},
		{"nullableShift", nullableShiftToy},
		{"MarkFirstFilter", markFirstFilterToy},
	} {
		t.Run(c.name, func(t *testing.T) {
			v := toyFeasibility(c.load(t))
			forbidden, off := v.ForbiddenShifts, v.OffAtShifts
			t.Logf("%s: %d of %d", c.name, off, forbidden)
			if off == 0 {
				t.Errorf("%s: no forbidden shift counted off at's next terminal (of %d)", c.name, forbidden)
			}
		})
	}
}

// Counter 5 (design §5): no rule whose right-hand side is all nullable, empty or not, has a non-allow reduce verdict
// before any terminal. The recognizer completes such a rule over an empty span inside the nullable closure, with no
// completion event to judge (design §4.1).
func TestFeasibilityAllNullableRulesAllow(t *testing.T) {
	_, _, v := perlyFeasibility(t)
	empty, nonEmpty, bad := v.NullableEmpty, v.NullableNonEmpty, v.NullableNonAllow
	t.Logf("perly.y: %d of %d (%d empty, %d non-empty)", len(bad), empty+nonEmpty, empty, nonEmpty)
	if len(bad) != 0 || empty != 13 || nonEmpty != 16 {
		t.Errorf("all-nullable rules with a non-allow verdict: rules %v of %d empty and %d non-empty, want none of 13 and 16", bad, empty, nonEmpty)
	}

	// Toy empty's opt ::= (empty) %prec LOW is forbidden before LB.
	t.Run("empty", func(t *testing.T) {
		ty := loadToy(t, "empty")
		v := toyFeasibility(ty.tab, ty.prec)
		empty, nonEmpty, bad := v.NullableEmpty, v.NullableNonEmpty, v.NullableNonAllow
		t.Logf("empty: %d of %d (%d empty, %d non-empty)", len(bad), empty+nonEmpty, empty, nonEmpty)
		if len(bad) == 0 {
			t.Errorf("toy empty: no all-nullable rule with a non-allow verdict counted (of %d)", empty+nonEmpty)
		}
	})

	// a ::= (empty) %prec LOW before T is forbidden after ONE, where s: ONE T shifts T, and reduced after TWO, so its
	// only non-allow verdict is 2, the parent item decides: a counter that reads only verdict 1 misses it.
	t.Run("verdict2", func(t *testing.T) {
		tab, p := pinnedToy(t, []grammar.Level{{N: 1, Assoc: "left", Tokens: []string{"LOW"}}, {N: 2, Assoc: "left", Tokens: []string{"T"}}},
			[]string{"", "", "", "", "", "LOW"}, "s ::= ONE e", "s ::= ONE T", "s ::= TWO f", "e ::= a T", "f ::= a T", "a ::= ")
		v := toyFeasibility(tab, p)
		empty, nonEmpty, bad := v.NullableEmpty, v.NullableNonEmpty, v.NullableNonAllow
		t.Logf("verdict-2 toy: %d of %d (%d empty, %d non-empty)", len(bad), empty+nonEmpty, empty, nonEmpty)
		row := p.ReduceAt[5*p.NT : 6*p.NT]
		if !slices.Equal(bad, []int{5}) || !slices.Contains(row, 2) || slices.Contains(row, 1) {
			t.Errorf("toy verdict2: counted rules %v, a's verdicts %v; want rule 5, a ::= (empty), with a 2 and no 1", bad, row)
		}
	})
}

// Counter 6 (design §5): the core items of one state, under one context, that scan one terminal get one shift
// verdict, since one transition carries them all. A dot-0 item, which V1 never checks, counts as allowed (F6-R3).
func TestFeasibilityCoreItemsShiftAgree(t *testing.T) {
	_, _, v := perlyFeasibility(t)
	groups, conflicts, dot0 := v.ScanGroups, v.ScanConflicts, v.ScanDot0
	t.Logf("perly.y: %d of %d groups (%d with a dot-0 item)", len(conflicts), groups, dot0)
	if len(conflicts) != 0 || groups == 0 {
		t.Errorf("(state, context, terminal) groups whose core items disagree: %d of %d", len(conflicts), groups)
	}

	// Its one collision is under the spine top s: P . a, where a: n . X R is forbidden X and the dot-0 c: . X Q is
	// allowed. Under none, a dot-0 item judged forbidden would collide instead, so the context pins F6-R3.
	t.Run("scancollide", func(t *testing.T) {
		tab, p := scancollideToy(t)
		v := toyFeasibility(tab, p)
		groups, conflicts, dot0 := v.ScanGroups, v.ScanConflicts, v.ScanDot0
		top := p.Base[1] + 1 // s: P . a
		t.Logf("scancollide: %d of %d groups, %d with a dot-0 item, under contexts %v (spine top s: P . a is %d)", len(conflicts), groups, dot0, conflicts, top)
		if !slices.Equal(conflicts, []int32{top}) || dot0 != 1 {
			t.Errorf("toy scancollide: conflicts under contexts %v, %d with a dot-0 item; want one dot-0 collision under the spine top %d, none under none (-1)", conflicts, dot0, top)
		}
	})
}

// The closing terminals carry no precedence level (F7-C3). The recognizer adds ; to a set where } or the end of input
// can come next, and V1 judges such a derivation through the slipper before ;; the two agree because none of the three
// has a level, so no resolution names any of them.
func TestFeasibilityClosingTerminalsHaveNoLevel(t *testing.T) {
	_, p := feasibilityTables(t)
	for _, name := range []string{"PERLY_SEMICOLON", "PERLY_BRACE_CLOSE", "EOF"} {
		level, err := terminalLevel(p, name)
		t.Logf("%s: level %d", name, level)
		if err != nil || level != 0 {
			t.Errorf("%s: level %d, %v; want no level", name, level, err)
		}
	}
}

// TestFeasibilityTerminalLevelLookupControl: terminalLevel cannot read no level for a terminal it failed to find. A
// name that is no terminal is an error, and ADDOP, which perly.y declares, reads its level.
func TestFeasibilityTerminalLevelLookupControl(t *testing.T) {
	_, p := feasibilityTables(t)
	for _, name := range []string{"NO_SUCH_TERMINAL", "term"} {
		if level, err := terminalLevel(p, name); err == nil {
			t.Errorf("%s: level %d, no error; want an error", name, level)
		}
	}
	level, err := terminalLevel(p, "ADDOP")
	t.Logf("ADDOP: level %d", level)
	if err != nil || level == 0 {
		t.Errorf("ADDOP: level %d, %v; want its declared level", level, err)
	}
}

// terminalLevel is the precedence level of the terminal named name, EOF the end of input, read as resolve reads it:
// 0 is no level, and EOF, past the symbols, has none. A name that is no terminal is an error, never level 0.
func terminalLevel(p *prectab.Precedence, name string) (int, error) {
	tab := p.Tab
	id := tab.EOF
	if name != "EOF" {
		id = symbol(tab, name)
		if id < 0 || !tab.Symbols[id].Terminal {
			return 0, fmt.Errorf("no terminal %s", name)
		}
	}
	if id < len(p.TermLevel) {
		return p.TermLevel[id], nil
	}
	return 0, nil
}

// The post-narrowing terminal-set counter (design §5, feasibility.md's static check): a terminal set, with ; added
// beside }, is mixed where some rule gets different reduce verdicts before two of its terminals that are both in
// FOLLOW(LHS). Un-narrowed, 63 of perly.y's 1,880 sets are mixed, all through PLUGIN_* readings; narrowed to at most
// one PLUGIN_* reading, whichever is kept, exactly one is: set 1626, {HASHBRACK, PERLY_BRACE_OPEN}, for
// indirob ::= scalar, which the recognizer cannot judge by one verdict (design §4.1). The set is named by its members
// and rule, so a renumbered table cannot pass by coincidence.
func TestFeasibilityNarrowedMixedSets(t *testing.T) {
	tab, p := feasibilityTables(t)
	un, narrowed := mixedSets(tab, p, tab.TerminalSets)
	t.Logf("sets: %d un-narrowed: %d narrowed: %d", len(tab.TerminalSets), len(un), len(narrowed))
	for _, m := range narrowed {
		t.Logf("narrowed mixed: set %d %s after %s", m.set, setText(tab, tab.TerminalSets[m.set]), ruleText(tab, m.rule))
	}
	if len(tab.TerminalSets) != 1880 || len(un) != 63 {
		t.Errorf("%d sets, %d mixed un-narrowed; want 1880 and 63", len(tab.TerminalSets), len(un))
	}
	if len(narrowed) != 1 || setText(tab, tab.TerminalSets[narrowed[0].set]) != "HASHBRACK PERLY_BRACE_OPEN" ||
		ruleText(tab, narrowed[0].rule) != "indirob ::= scalar" {
		t.Errorf("narrowed mixed sets %v; want exactly {HASHBRACK PERLY_BRACE_OPEN} for indirob ::= scalar", narrowed)
	}

	// Synthetic sets: two non-PLUGIN_* terminals of one FOLLOW(LHS) whose verdicts differ are mixed narrowed or not; a
	// non-PLUGIN_* and a PLUGIN_* terminal stay mixed when narrowing keeps that one reading, which dropping every
	// PLUGIN_* would hide; two PLUGIN_* terminals are mixed only un-narrowed.
	plugin := grammar.Derived(tab, buildPluginGrammar).is
	var base, kept, plugins [][]int
	for r, rule := range tab.Rules {
		fl := tab.Follow[rule.LHS]
		for i, a := range fl {
			for _, b := range fl[i+1:] {
				if a >= len(plugin) || b >= len(plugin) || !prectab.ReduceDiffers(p, r, a, r, b) {
					continue
				}
				switch {
				case !plugin[a] && !plugin[b] && len(base) < 20:
					base = append(base, []int{a, b})
				case plugin[a] != plugin[b] && len(kept) < 20:
					kept = append(kept, []int{a, b})
				case plugin[a] && plugin[b] && len(plugins) < 20:
					plugins = append(plugins, []int{a, b})
				}
			}
		}
	}
	for _, c := range []struct {
		name     string
		sets     [][]int
		narrowed bool // whether narrowing keeps every set mixed; un-narrowed, every set is
	}{{"base", base, true}, {"kept", kept, true}, {"plugins", plugins, false}} {
		su, sn := mixedSets(tab, p, c.sets)
		t.Logf("synthetic %s: %d sets, un-narrowed %d, narrowed %d", c.name, len(c.sets), len(su), len(sn))
		want := 0
		if c.narrowed {
			want = len(c.sets)
		}
		if len(c.sets) == 0 || len(su) != len(c.sets) || len(sn) != want {
			t.Errorf("synthetic %s: %d sets, un-narrowed %d, narrowed %d", c.name, len(c.sets), len(su), len(sn))
		}
	}
}

// mixedSet is a terminal set (by index) and the first rule found mixed before it.
type mixedSet struct{ set, rule int }

// mixedSets lists the sets that are mixed whole (un) and those mixed under some narrowing to at most one PLUGIN_*
// reading (narrowed).
func mixedSets(tab *grammar.Tables, p *prectab.Precedence, sets [][]int) (un, narrowed []mixedSet) {
	semi, brace := symbol(tab, "PERLY_SEMICOLON"), symbol(tab, "PERLY_BRACE_CLOSE")
	plugin := grammar.Derived(tab, buildPluginGrammar).is
	mixed := func(la []int) int {
		if slices.Contains(la, brace) && !slices.Contains(la, semi) {
			la = append(slices.Clone(la), semi)
		}
		for r, rule := range tab.Rules {
			fl := tab.Follow[rule.LHS]
			for i, a := range la {
				for _, b := range la[i+1:] {
					if slices.Contains(fl, a) && slices.Contains(fl, b) && prectab.ReduceDiffers(p, r, a, r, b) {
						return r
					}
				}
			}
		}
		return -1
	}
	for id, set := range sets {
		if r := mixed(set); r >= 0 {
			un = append(un, mixedSet{id, r})
		}
		var base, plugs []int
		for _, x := range set {
			if plugin[x] {
				plugs = append(plugs, x)
			} else {
				base = append(base, x)
			}
		}
		readings := [][]int{base}
		for _, x := range plugs {
			readings = append(readings, append(slices.Clone(base), x))
		}
		for _, la := range readings {
			if r := mixed(la); r >= 0 {
				narrowed = append(narrowed, mixedSet{id, r})
				break
			}
		}
	}
	return un, narrowed
}

// setText is a terminal set's member names, in id order.
func setText(tab *grammar.Tables, set []int) string {
	names := make([]string, len(set))
	for i, x := range set {
		names[i] = tab.Symbols[x].Name
	}
	return strings.Join(names, " ")
}

// ruleText is rule r as "LHS ::= RHS".
func ruleText(tab *grammar.Tables, r int) string {
	parts := []string{tab.Symbols[tab.Rules[r].LHS].Name, "::="}
	for _, x := range tab.Rules[r].RHS {
		parts = append(parts, tab.Symbols[x].Name)
	}
	return strings.Join(parts, " ")
}
