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

		// TestZZM3NonVacuous's cases, on perly.y's first forbidden shift (at, f) under top.
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
		other := (at + 7) % int32(len(tab.Items))
		cases := []struct {
			name string
			xs   []shiftRec
			want shiftVerdict
		}{
			{"cross-core", []shiftRec{{at, top, f}, {other, -1, f}}, shiftVerdict{multi: true, cross: true, copies: true, overKeep: true}},
			{"same-core", []shiftRec{{at, top, f}, {at, -1, f}}, shiftVerdict{multi: true, same: true, copies: true}},
			{"one shared top", []shiftRec{{at, top, f}, {other, top, f}}, shiftVerdict{}},
		}
		counted := 0
		for _, c := range cases {
			if got := judgeShifts(p, c.xs); got == c.want {
				counted++
			} else {
				t.Errorf("synthetic %s: got %+v, want %+v", c.name, got, c.want)
			}
		}
		t.Logf("synthetic: %d of %d", counted, len(cases))
	})

	// (b) No predictor begins rules under two different non-none canonical tops.
	t.Run("b", func(t *testing.T) {
		n := 0
		var mixed []int
		for j, k := range preds {
			real := map[int32]bool{}
			for _, ts := range rts[j] {
				maps.Copy(real, ts)
			}
			none := real[-1]
			delete(real, -1)
			if len(real) >= 2 {
				n++
			} else if len(real) == 1 && none {
				mixed = append(mixed, k.kernel)
			}
		}
		t.Logf("predictors mixing a top with none: %s", strings.Trim(fmt.Sprint(mixed), "[]"))
		t.Logf("perly.y: %d of %d predictors", n, len(preds))
		if n != 0 || len(mixed) == 0 {
			t.Errorf("predictors beginning rules under two tops: %d of %d (mixing with none: %d)", n, len(preds), len(mixed))
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
			hit, two := false, false
			for r, ts := range rts[j] {
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
	})
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
