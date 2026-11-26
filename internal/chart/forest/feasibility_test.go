package forest

// The V2 design's §3 feasibility counters, pinned as tests over the committed perly.y tables (design §5): each reads
// 0 on perly.y and counts every synthetic case built to trigger it. Ported from the scratch code measured in
// docs/assessments/chart-precedence-v2/feasibility.md.

import (
	"fmt"
	"maps"
	"slices"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
)

func feasibilityTables(t *testing.T) (*grammar.Tables, *Precedence) {
	t.Helper()
	tab := committedTables(t)
	p, err := ParsePrecedence(tab)
	if err != nil {
		t.Fatal(err)
	}
	return tab, p
}

// Counter 1 (feasibility.md §1): no state reduces two rules of one LHS whose V1 reduce verdicts differ before a
// terminal of FOLLOW(LHS), so completion can be judged per LHS.
func TestFeasibilityRuleReduceConflicts(t *testing.T) {
	tab, p := feasibilityTables(t)
	n, agree := ruleReduceConflicts(tab, p, tab.States)
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
					if reduceDiffers(p, rs[i], rs[j], x) {
						synth = append(synth, grammar.State{Reduces: []int{rs[i], rs[j]}})
						break
					}
				}
			}
		}
	}
	sn, _ := ruleReduceConflicts(tab, p, synth)
	t.Logf("synthetic: %d of %d", sn, len(synth))
	if sn != len(synth) || len(synth) == 0 {
		t.Errorf("synthetic same-LHS pairs counted: %d of %d", sn, len(synth))
	}
}

// Counter 2 (feasibility.md §2): no state holds two parent items waiting on one parent-dependent symbol whose
// parentForbid verdicts disagree, so a parent-dependent verdict is a function of the parent state.
func TestFeasibilityParentConflicts(t *testing.T) {
	tab, p := feasibilityTables(t)
	n := parentConflicts(tab, p, tab.States)
	t.Logf("perly.y: %d of %d states", n, len(tab.States))
	if n != 0 || len(tab.States) == 0 {
		t.Errorf("parent conflicts in %d of %d states", n, len(tab.States))
	}

	// Every pair of perly.y's items waiting on one parent-dependent symbol with differing verdicts, each in a state of its own.
	all := make([]int, len(tab.Items))
	for i := range all {
		all[i] = i
	}
	var synth []grammar.State
	for b, is := range awaiting(tab, p, all) {
		for x := range is {
			for y := x + 1; y < len(is); y++ {
				if parentDiffers(tab, p, b, is[x], is[y]) {
					synth = append(synth, grammar.State{Items: []int{is[x], is[y]}})
				}
			}
		}
	}
	sn := parentConflicts(tab, p, synth)
	t.Logf("synthetic: %d of %d", sn, len(synth))
	if sn != len(synth) || len(synth) == 0 {
		t.Errorf("synthetic parent-item pairs counted: %d of %d", sn, len(synth))
	}
}

// Counter 3 (feasibility.md §3.2): context conflicts, three checks over the static model of each rule's canonical
// spine tops (F7-A1).
func TestFeasibilityContextConflicts(t *testing.T) {
	tab, p := feasibilityTables(t)
	preds := predictors(tab)
	rts := make([]map[int]map[int32]bool, len(preds)) // by predictor: its rules' tops
	for n, k := range preds {
		rts[n] = ruleTops(tab, p, k)
	}
	// perly.y's first forbidden shift: f after rule-position item at, under spine top top. The synthetic cases use it.
	var at, top, f int32 = -1, -1, -1
find:
	for i := range p.topAt {
		if p.topAt[i] < 0 {
			continue
		}
		for x := range tab.Items {
			for y := range p.nt {
				if p.shiftForbidden(int32(i), int32(x), int32(y)) {
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
		tops := staticTops(tab, preds, rts)
		var meet []int
		n := 0
		for s, st := range tab.States {
			var xs []shiftRec
			for _, i := range st.Items {
				for top := range tops[s][tab.Items[i].Rule] {
					for _, f := range itemShifts(tab, i) {
						xs = append(xs, shiftRec{int32(i), top, int32(f)})
					}
				}
			}
			v := judgeShifts(p, xs)
			if v.cross {
				meet = append(meet, s)
			}
			if v.contextConflict() {
				n++
			}
		}
		t.Logf("core tops meet: %s", strings.Trim(fmt.Sprint(meet), "[]"))
		t.Logf("perly.y: %d of %d states", n, len(tab.States))
		if n != 0 || len(meet) == 0 {
			t.Errorf("context conflicts in %d of %d states (core tops meet in %d)", n, len(tab.States), len(meet))
		}

		// TestZZM3NonVacuous's cases, on perly.y's first forbidden shift.
		other := (at + 7) % int32(len(tab.Items))
		cases := []struct {
			name     string
			xs       []shiftRec
			want     shiftVerdict
			conflict bool
		}{
			{"cross-core", []shiftRec{{at, top, f}, {other, -1, f}}, shiftVerdict{multi: true, cross: true, copies: true, overKeep: true}, true},
			{"same-core", []shiftRec{{at, top, f}, {at, -1, f}}, shiftVerdict{multi: true, same: true, copies: true}, false},
			{"one shared top", []shiftRec{{at, top, f}, {other, top, f}}, shiftVerdict{}, false},
		}
		counted := 0
		for _, c := range cases {
			got := judgeShifts(p, c.xs)
			if got == c.want && got.contextConflict() == c.conflict {
				counted++
			} else {
				t.Errorf("synthetic %s: got %+v (conflict %v), want %+v (conflict %v)", c.name, got, got.contextConflict(), c.want, c.conflict)
			}
		}
		t.Logf("synthetic: %d of %d", counted, len(cases))
	})

	// (b) No predictor begins rules under two different non-none canonical tops.
	t.Run("b", func(t *testing.T) {
		n := 0
		var mixed []int
		for j, k := range preds {
			conflict, withNone := predictorTopsConflict(rts[j])
			if conflict {
				n++
			} else if withNone {
				mixed = append(mixed, k.kernel)
			}
		}
		t.Logf("predictors mixing a top with none: %s", strings.Trim(fmt.Sprint(mixed), "[]"))
		t.Logf("perly.y: %d of %d predictors", n, len(preds))
		if n != 0 || len(mixed) == 0 {
			t.Errorf("predictors beginning rules under two tops: %d of %d (mixing with none: %d)", n, len(preds), len(mixed))
		}

		// A predictor beginning one rule under top and another under a second non-none top.
		second := -1
		for i, row := range p.topAt {
			if row >= 0 && int32(i) != top {
				second = i
				break
			}
		}
		synth := map[int]map[int32]bool{0: {top: true}, 1: {int32(second): true}}
		sn := 0
		if conflict, _ := predictorTopsConflict(synth); conflict {
			sn++
		}
		t.Logf("synthetic: %d of 1", sn)
		if sn != 1 {
			t.Errorf("synthetic predictor under tops %d and %d not counted", top, second)
		}
	})

	// (c) No rule begun under one predictor carries two tops whose verdicts disagree.
	t.Run("c", func(t *testing.T) {
		n, multi := 0, 0
		for j, k := range preds {
			var shifts [][2]int
			for _, s := range descendants(tab, k.epsilon) {
				for _, i := range tab.States[s].Items {
					for _, f := range itemShifts(tab, i) {
						shifts = append(shifts, [2]int{i, f})
					}
				}
			}
			hit, two := ruleTopsDisagree(tab, p, rts[j], shifts)
			if two {
				multi++
			}
			if hit {
				n++
			}
		}
		t.Logf("predictors beginning a rule under two tops: %d", multi)
		t.Logf("perly.y: %d of %d predictors", n, len(preds))
		if n != 0 || multi == 0 {
			t.Errorf("predictors with a rule whose tops disagree: %d of %d (with a rule under two tops: %d)", n, len(preds), multi)
		}

		// at's rule begun under top, which forbids shifting f after at, and under none, which allows it.
		synth := map[int]map[int32]bool{tab.Items[at].Rule: {top: true, -1: true}}
		sn := 0
		if hit, _ := ruleTopsDisagree(tab, p, synth, [][2]int{{int(at), int(f)}}); hit {
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
	tab, p := feasibilityTables(t)
	forbidden, off := offAtShifts(tab, p)
	t.Logf("perly.y: %d of %d", off, forbidden)
	if off != 0 || forbidden != 674 {
		t.Errorf("forbidden shifts whose terminal is not at's own next terminal: %d of %d, want 0 of 674", off, forbidden)
	}

	// The toys whose forbidden shifts are scanned by an item that does not hold at (design §4.1).
	for _, c := range []struct {
		name string
		load func(*testing.T) (*grammar.Tables, *Precedence)
	}{
		{"emptyred", func(t *testing.T) (*grammar.Tables, *Precedence) {
			ty := loadToy(t, "emptyred")
			return ty.tab, ty.prec
		}},
		{"nullableShift", nullableShiftToy},
		{"MarkFirstFilter", markFirstFilterToy},
	} {
		t.Run(c.name, func(t *testing.T) {
			forbidden, off := offAtShifts(c.load(t))
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
	tab, p := feasibilityTables(t)
	empty, nonEmpty, bad := nullableNonAllow(tab, p)
	t.Logf("perly.y: %d of %d (%d empty, %d non-empty)", len(bad), empty+nonEmpty, empty, nonEmpty)
	if len(bad) != 0 || empty != 13 || nonEmpty != 16 {
		t.Errorf("all-nullable rules with a non-allow verdict: rules %v of %d empty and %d non-empty, want none of 13 and 16", bad, empty, nonEmpty)
	}

	// Toy empty's opt ::= (empty) %prec LOW is forbidden before LB.
	t.Run("empty", func(t *testing.T) {
		ty := loadToy(t, "empty")
		empty, nonEmpty, bad := nullableNonAllow(ty.tab, ty.prec)
		t.Logf("empty: %d of %d (%d empty, %d non-empty)", len(bad), empty+nonEmpty, empty, nonEmpty)
		if len(bad) == 0 {
			t.Errorf("toy empty: no all-nullable rule with a non-allow verdict counted (of %d)", empty+nonEmpty)
		}
	})
}

// Counter 6 (design §5): the core items of one state, under one context, that scan one terminal get one shift
// verdict, since one transition carries them all. A dot-0 item, which V1 never checks, counts as allowed (F6-R3).
func TestFeasibilityCoreItemsShiftAgree(t *testing.T) {
	tab, p := feasibilityTables(t)
	groups, conflicts, dot0 := scanDisagreements(tab, p)
	t.Logf("perly.y: %d of %d groups (%d with a dot-0 item)", len(conflicts), groups, dot0)
	if len(conflicts) != 0 || groups == 0 {
		t.Errorf("(state, context, terminal) groups whose core items disagree: %d of %d", len(conflicts), groups)
	}

	// Its one collision is under the spine top s: P . a, where a: n . X R is forbidden X and the dot-0 c: . X Q is
	// allowed. Under none, a dot-0 item judged forbidden would collide instead, so the context pins F6-R3.
	t.Run("scancollide", func(t *testing.T) {
		tab, p := scancollideToy(t)
		groups, conflicts, dot0 := scanDisagreements(tab, p)
		top := p.base[1] + 1 // s: P . a
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

// reduceDiffers reports whether V1's reduce verdicts for rules r1 and r2 before t differ; where both depend on the
// parent item (2), their whole parentForbid rows are compared.
func reduceDiffers(p *Precedence, r1, r2, t int) bool {
	v1, v2 := p.reduceAt[r1*p.nt+t], p.reduceAt[r2*p.nt+t]
	if v1 != v2 {
		return true
	}
	if v1 != 2 {
		return false
	}
	ni := len(p.topAt)
	a, b := int(p.mixAt[r1*p.nt+t])*ni, int(p.mixAt[r2*p.nt+t])*ni
	return !slices.Equal(p.parentForbid[a:a+ni], p.parentForbid[b:b+ni])
}

// ruleReduceConflicts counts the states reducing two rules of one LHS whose verdicts differ before some terminal
// of FOLLOW(LHS), EOF included, and lists the states where every such pair agrees.
func ruleReduceConflicts(tab *grammar.Tables, p *Precedence, states []grammar.State) (conflicts int, agree []int) {
	for s, st := range states {
		by := map[int][]int{}
		for _, r := range st.Reduces {
			by[tab.Rules[r].LHS] = append(by[tab.Rules[r].LHS], r)
		}
		paired, hit := false, false
		for lhs, rs := range by {
			for i := range rs {
				for j := i + 1; j < len(rs); j++ {
					paired = true
					for _, x := range tab.Follow[lhs] {
						if reduceDiffers(p, rs[i], rs[j], x) {
							hit = true
						}
					}
				}
			}
		}
		switch {
		case hit:
			conflicts++
		case paired:
			agree = append(agree, s)
		}
	}
	return conflicts, agree
}

// parentDiffers reports whether some parent-dependent pair (r, t) with LHS(r) = b is forbidden under one of the
// parent items p1 and p2 and not the other.
func parentDiffers(tab *grammar.Tables, p *Precedence, b, p1, p2 int) bool {
	ni := len(p.topAt)
	for r, rule := range tab.Rules {
		if rule.LHS != b {
			continue
		}
		for t := range p.nt {
			k := r*p.nt + t
			if p.reduceAt[k] != 2 {
				continue
			}
			if row := int(p.mixAt[k]) * ni; p.parentForbid[row+p1] != p.parentForbid[row+p2] {
				return true
			}
		}
	}
	return false
}

// awaiting groups items by the parent-dependent symbol after their dot.
func awaiting(tab *grammar.Tables, p *Precedence, items []int) map[int][]int {
	by := map[int][]int{}
	for _, i := range items {
		it := tab.Items[i]
		if rhs := tab.Rules[it.Rule].RHS; it.Dot < len(rhs) && p.needParent[rhs[it.Dot]] {
			by[rhs[it.Dot]] = append(by[rhs[it.Dot]], i)
		}
	}
	return by
}

// parentConflicts counts the states holding two core items with the dot before one parent-dependent symbol whose
// parentForbid verdicts differ.
func parentConflicts(tab *grammar.Tables, p *Precedence, states []grammar.State) (conflicts int) {
	for _, st := range states {
	state:
		for b, is := range awaiting(tab, p, st.Items) {
			for x := range is {
				for y := x + 1; y < len(is); y++ {
					if parentDiffers(tab, p, b, is[x], is[y]) {
						conflicts++
						break state
					}
				}
			}
		}
	}
	return conflicts
}

// A predictor is a kernel state and the predicted state it begins rules in; the start state is kernel -1,
// whose start rules hang from their own start item.
type predictor struct{ kernel, epsilon int }

func predictors(tab *grammar.Tables) []predictor {
	out := []predictor{{-1, 0}}
	for k, st := range tab.States {
		if !st.Predicted && st.Epsilon >= 0 {
			out = append(out, predictor{k, st.Epsilon})
		}
	}
	return out
}

// canonTop is item i as a canonical spine top: itself where it forbids some shift, -1 (none) otherwise.
func canonTop(p *Precedence, i int) int32 {
	if i >= 0 && p.topAt[i] < 0 {
		return -1
	}
	return int32(i)
}

// ruleTops is, by rule, the canonical spine tops of the rules k begins: each hangs from a kernel item of k, or
// from an item of k's predicted state advanced over a nullable, and a left-corner prediction inherits its tops.
func ruleTops(tab *grammar.Tables, p *Precedence, k predictor) map[int]map[int32]bool {
	byLHS := make([][]int, len(tab.Symbols))
	for r, rule := range tab.Rules {
		byLHS[rule.LHS] = append(byLHS[rule.LHS], r)
	}
	tops := map[int]map[int32]bool{}
	changed := false
	add := func(r int, t int32) {
		if tops[r] == nil {
			tops[r] = map[int32]bool{}
		}
		if !tops[r][t] {
			tops[r][t], changed = true, true
		}
	}
	if k.epsilon == 0 {
		for _, r := range byLHS[tab.Start] {
			add(r, canonTop(p, int(p.base[r])))
		}
	}
	predict := func(i int, t int32) {
		it := tab.Items[i]
		if rhs := tab.Rules[it.Rule].RHS; it.Dot < len(rhs) && !tab.Symbols[rhs[it.Dot]].Terminal {
			for _, r := range byLHS[rhs[it.Dot]] {
				add(r, t)
			}
		}
	}
	for changed = true; changed; {
		changed = false
		if k.kernel >= 0 {
			for _, i := range tab.States[k.kernel].Items {
				predict(i, canonTop(p, i))
			}
		}
		for _, i := range tab.States[k.epsilon].Items {
			if tab.Items[i].Dot > 0 {
				predict(i, canonTop(p, i))
				continue
			}
			for t := range tops[tab.Items[i].Rule] {
				predict(i, t)
			}
		}
	}
	return tops
}

// descendants is predicted state e and every state reached from it by transitions at the same origin.
func descendants(tab *grammar.Tables, e int) []int {
	seen := map[int]bool{e: true}
	out := []int{e}
	for n := 0; n < len(out); n++ {
		for _, tr := range tab.States[out[n]].Transitions {
			if !tab.States[tr.Target].Predicted && !seen[tr.Target] {
				seen[tr.Target] = true
				out = append(out, tr.Target)
			}
		}
	}
	return out
}

// staticTops over-approximates, by state and rule, the canonical spine tops of every rule a state's core items
// belong to: the union of rts, by predictor, over every predictor whose predicted state reaches it (feasibility.md §3.2).
func staticTops(tab *grammar.Tables, preds []predictor, rts []map[int]map[int32]bool) []map[int]map[int32]bool {
	out := make([]map[int]map[int32]bool, len(tab.States))
	for n, k := range preds {
		rt := rts[n]
		for _, s := range descendants(tab, k.epsilon) {
			if out[s] == nil {
				out[s] = map[int]map[int32]bool{}
			}
			for _, i := range tab.States[s].Items {
				r := tab.Items[i].Rule
				if len(rt[r]) > 0 && out[s][r] == nil {
					out[s][r] = map[int32]bool{}
				}
				maps.Copy(out[s][r], rt[r])
			}
		}
	}
	return out
}

// itemShifts is the terminals item i can shift next: none at dot 0 or at its end.
func itemShifts(tab *grammar.Tables, i int) []int {
	it := tab.Items[i]
	rhs := tab.Rules[it.Rule].RHS
	if it.Dot == 0 || it.Dot >= len(rhs) {
		return nil
	}
	if x := rhs[it.Dot]; !tab.Symbols[x].Terminal {
		return tab.First[x]
	}
	return []int{rhs[it.Dot]}
}

// shiftRec is a shift of f after rule-position item at under canonical spine top top (-1 none).
type shiftRec struct{ at, top, f int32 }

// shiftVerdict judges one item's shifts: its core items carry two or more tops (multi), because one core item
// carries several (same) or because core items carry different ones (cross); the tops give two or more verdict
// vectors over its shifts (copies); a top of another core item allows a shift every top of at forbids (overKeep).
type shiftVerdict struct{ multi, same, cross, copies, overKeep bool }

// contextConflict is check (a)'s conflict: core items with different tops whose verdicts differ.
func (v shiftVerdict) contextConflict() bool { return v.cross && (v.copies || v.overKeep) }

func judgeShifts(p *Precedence, xs []shiftRec) (v shiftVerdict) {
	real := map[int32]map[int32]bool{} // at -> its tops
	all := map[int32]bool{}
	for _, x := range xs {
		if real[x.at] == nil {
			real[x.at] = map[int32]bool{}
		}
		real[x.at][x.top], all[x.top] = true, true
	}
	if len(all) < 2 {
		return v
	}
	v.multi = true
	for _, tops := range real {
		v.same = v.same || len(tops) >= 2
		v.cross = v.cross || len(tops) < len(all)
	}
	vectors := map[string]bool{}
	for top := range all {
		var b strings.Builder
		for _, x := range xs {
			fmt.Fprint(&b, p.shiftForbidden(top, x.at, x.f))
		}
		vectors[b.String()] = true
	}
	v.copies = len(vectors) >= 2
	for _, x := range xs {
		forbidden := true
		for tp := range real[x.at] {
			forbidden = forbidden && p.shiftForbidden(tp, x.at, x.f)
		}
		for c := range all {
			v.overKeep = v.overKeep || forbidden && !real[x.at][c] && !p.shiftForbidden(c, x.at, x.f)
		}
	}
	return v
}

// predictorTopsConflict is check (b) on one predictor's rule tops: conflict where they hold two different non-none
// tops; withNone where they hold one non-none top beside none.
func predictorTopsConflict(rt map[int]map[int32]bool) (conflict, withNone bool) {
	real := map[int32]bool{}
	for _, ts := range rt {
		maps.Copy(real, ts)
	}
	none := real[-1]
	delete(real, -1)
	return len(real) >= 2, len(real) == 1 && none
}

// ruleTopsDisagree is check (c) on one predictor: hit where a rule it begins under two or more tops gets different
// verdicts from them on one of shifts, its (rule-position item, terminal) pairs; two where any rule has two tops.
func ruleTopsDisagree(tab *grammar.Tables, p *Precedence, rt map[int]map[int32]bool, shifts [][2]int) (hit, two bool) {
	for r, ts := range rt {
		if len(ts) < 2 {
			continue
		}
		two = true
		for _, sh := range shifts {
			if tab.Items[sh[0]].Rule != r {
				continue
			}
			seen := map[bool]bool{}
			for c := range ts {
				seen[p.shiftForbidden(c, int32(sh[0]), int32(sh[1]))] = true
			}
			hit = hit || len(seen) > 1
		}
	}
	return hit, two
}

// offAtShifts counts p's forbidden (top, at, f) shifts over every automaton item at, and of them those whose f is
// not at's own next terminal: at is complete, a nonterminal or another terminal follows its dot, or at is a
// refused rule's item, which no item of the tables holds.
func offAtShifts(tab *grammar.Tables, p *Precedence) (forbidden, off int) {
	for top, row := range p.topAt {
		if row < 0 {
			continue
		}
		for at := range p.topAt {
			for f := range p.nt {
				if !p.shiftForbidden(int32(top), int32(at), int32(f)) {
					continue
				}
				forbidden++
				if at >= len(tab.Items) {
					off++
					continue
				}
				it := tab.Items[at]
				if rhs := tab.Rules[it.Rule].RHS; it.Dot >= len(rhs) || rhs[it.Dot] != f {
					off++
				}
			}
		}
	}
	return forbidden, off
}

// nullableNonAllow counts tab's rules whose right-hand side is all nullable, empty and non-empty, and lists those
// with a non-allow reduce verdict (forbidden, or decided by the parent item) before some terminal, EOF included.
func nullableNonAllow(tab *grammar.Tables, p *Precedence) (empty, nonEmpty int, bad []int) {
	for r, rule := range tab.Rules {
		if !slices.ContainsFunc(rule.RHS, func(s int) bool { return !tab.Symbols[s].Nullable }) {
			if len(rule.RHS) == 0 {
				empty++
			} else {
				nonEmpty++
			}
			if slices.ContainsFunc(p.reduceAt[r*p.nt:(r+1)*p.nt], func(v uint8) bool { return v != 0 }) {
				bad = append(bad, r)
			}
		}
	}
	return empty, nonEmpty, bad
}

// scanDisagreements groups the core items of each state by the terminal after their dot, and judges each group of
// two or more under every context an item can carry: none and each canonical spine top. A group conflicts where its
// items' shift verdicts differ; a dot-0 item is allowed, never skipped. conflicts is each conflicting group's
// context, -1 for none; dot0 counts the conflicts with a dot-0 item.
func scanDisagreements(tab *grammar.Tables, p *Precedence) (groups int, conflicts []int32, dot0 int) {
	contexts := []int32{-1}
	for i, row := range p.topAt {
		if row >= 0 {
			contexts = append(contexts, int32(i))
		}
	}
	for _, st := range tab.States {
		by := map[int][]int{}
		for _, i := range st.Items {
			it := tab.Items[i]
			if rhs := tab.Rules[it.Rule].RHS; it.Dot < len(rhs) && tab.Symbols[rhs[it.Dot]].Terminal {
				by[rhs[it.Dot]] = append(by[rhs[it.Dot]], i)
			}
		}
		for f, is := range by {
			if len(is) < 2 {
				continue
			}
			for _, c := range contexts {
				groups++
				verdicts, zero := map[bool]bool{}, false
				for _, i := range is {
					dot := tab.Items[i].Dot
					verdicts[dot > 0 && p.shiftForbidden(c, int32(i), int32(f))] = true
					zero = zero || dot == 0
				}
				if len(verdicts) > 1 {
					conflicts = append(conflicts, c)
					if zero {
						dot0++
					}
				}
			}
		}
	}
	return groups, conflicts, dot0
}

// terminalLevel is the precedence level of the terminal named name, EOF the end of input, read as resolve reads it:
// 0 is no level, and EOF, past the symbols, has none. A name that is no terminal is an error, never level 0.
func terminalLevel(p *Precedence, name string) (int, error) {
	tab := p.tab
	id := tab.EOF
	if name != "EOF" {
		id = symbol(tab, name)
		if id < 0 || !tab.Symbols[id].Terminal {
			return 0, fmt.Errorf("no terminal %s", name)
		}
	}
	if id < len(p.term) {
		return p.term[id], nil
	}
	return 0, nil
}
