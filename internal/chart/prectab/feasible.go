// ABOUTME: The V2 design's §3 feasibility checks as one refusal over a grammar's precedence tables (design §5).
// ABOUTME: Feasible counts the six counters the recognizer's judgement rests on; NewPrecedence never calls it.

package prectab

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"tamarou.com/pvm/internal/chart/grammar"
)

// Feasibility is the six §3 counters over one Precedence's tables, each 0 where the recognizer can judge the
// grammar's precedence, and beside each what it examined, so a counter that examined nothing can be told from a 0.
type Feasibility struct {
	RuleReduceConflicts int   // states reducing two rules of one LHS with different verdicts before FOLLOW(LHS)
	TwinAgree           []int // states reducing two rules of one LHS whose verdicts all agree

	ParentConflicts int   // states holding two parent items of one parent-dependent symbol with different verdicts
	ParentMeet      []int // states where two core items wait on one parent-dependent symbol

	ContextA     int   // states whose core items carry different canonical tops that give different shift verdicts
	CoreTopsMeet []int // states whose core items carry different canonical tops
	ContextB     int   // predictors beginning rules under two different non-none canonical tops
	Predictors   int
	MixedNone    []int // the kernels of predictors beginning rules under one non-none top beside none
	ContextC     int   // predictors beginning a rule under two tops whose shift verdicts disagree
	MultiTop     int   // predictors beginning some rule under two tops
	RealShifts   int   // the (item, terminal) shifts check c judged

	OffAtShifts, ForbiddenShifts int // forbidden (top, at, f) shifts whose f is not at's next terminal, of all

	NullableNonAllow                []int // all-nullable rules with a non-allow reduce verdict before some terminal
	NullableEmpty, NullableNonEmpty int   // the all-nullable rules, empty and not

	ScanConflicts []int32 // the context of each (state, context, terminal) group whose core items disagree
	ScanGroups    int
	ScanDot0      int // the conflicts holding a dot-0 item
}

// Feasible is the §3 counters over p's tables, and an error naming each counter that is not 0: the refusal the
// recognizer applies before judging a grammar's precedence. NewPrecedence does not call it, so a grammar that
// fails it still loads, and the forest's Constrain still judges it.
func Feasible(p *Precedence) (Feasibility, error) {
	tab := p.Tab
	var v Feasibility
	v.RuleReduceConflicts, v.TwinAgree = RuleReduceConflicts(p, tab.States)

	v.ParentConflicts = ParentConflicts(p, tab.States)
	for s, st := range tab.States {
		for _, is := range Awaiting(p, st.Items) {
			if len(is) >= 2 {
				v.ParentMeet = append(v.ParentMeet, s)
				break
			}
		}
	}

	preds := predictors(tab)
	v.Predictors = len(preds)
	rts := make([]map[int]map[int32]bool, len(preds)) // by predictor: its rules' tops
	for n, k := range preds {
		rts[n] = ruleTops(p, k)
	}
	tops := staticTops(tab, preds, rts)
	for s, st := range tab.States {
		var xs []ShiftRec
		for _, i := range st.Items {
			for top := range tops[s][tab.Items[i].Rule] {
				for _, f := range itemShifts(tab, i) {
					xs = append(xs, ShiftRec{int32(i), top, int32(f)})
				}
			}
		}
		j := JudgeShifts(p, xs)
		if j.Cross {
			v.CoreTopsMeet = append(v.CoreTopsMeet, s)
		}
		if j.ContextConflict() {
			v.ContextA++
		}
	}
	for j, k := range preds {
		conflict, withNone := PredictorTopsConflict(rts[j])
		if conflict {
			v.ContextB++
		} else if withNone {
			v.MixedNone = append(v.MixedNone, k.kernel)
		}
		var shifts [][2]int
		for _, s := range descendants(tab, k.epsilon) {
			for _, i := range tab.States[s].Items {
				for _, f := range itemShifts(tab, i) {
					shifts = append(shifts, [2]int{i, f})
				}
			}
		}
		v.RealShifts += len(shifts)
		hit, two := RuleTopsDisagree(p, rts[j], shifts)
		if two {
			v.MultiTop++
		}
		if hit {
			v.ContextC++
		}
	}

	v.ForbiddenShifts, v.OffAtShifts = offAtShifts(p)
	v.NullableEmpty, v.NullableNonEmpty, v.NullableNonAllow = nullableNonAllow(p)
	v.ScanGroups, v.ScanConflicts, v.ScanDot0 = scanDisagreements(p)

	var fail []string
	for _, c := range []struct {
		n    int
		name string
	}{
		{v.RuleReduceConflicts, "rule-reduce conflicts"},
		{v.ParentConflicts, "parent conflicts"},
		{v.ContextA + v.ContextB + v.ContextC, "context conflicts"},
		{v.OffAtShifts, "forbidden shifts off at's next terminal"},
		{len(v.NullableNonAllow), "all-nullable rules with a non-allow reduce verdict"},
		{len(v.ScanConflicts), "core items disagreeing on one scan"},
	} {
		if c.n != 0 {
			fail = append(fail, fmt.Sprintf("%d %s", c.n, c.name))
		}
	}
	if len(fail) > 0 {
		return v, fmt.Errorf("prectab: the recognizer cannot judge this grammar's precedence: %s", strings.Join(fail, "; "))
	}
	return v, nil
}

// ReduceDiffers reports whether V1's reduce verdicts for rule r1 before t1 and rule r2 before t2 differ; where both
// depend on the parent item (2), their whole parentForbid rows are compared.
func ReduceDiffers(p *Precedence, r1, t1, r2, t2 int) bool {
	v1, v2 := p.ReduceAt[r1*p.NT+t1], p.ReduceAt[r2*p.NT+t2]
	if v1 != v2 {
		return true
	}
	if v1 != 2 {
		return false
	}
	ni := len(p.TopAt)
	a, b := int(p.MixAt[r1*p.NT+t1])*ni, int(p.MixAt[r2*p.NT+t2])*ni
	return !slices.Equal(p.ParentForbid[a:a+ni], p.ParentForbid[b:b+ni])
}

// RuleReduceConflicts counts the states reducing two rules of one LHS whose verdicts differ before some terminal
// of FOLLOW(LHS), EOF included, and lists the states where every such pair agrees.
func RuleReduceConflicts(p *Precedence, states []grammar.State) (conflicts int, agree []int) {
	tab := p.Tab
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
						if ReduceDiffers(p, rs[i], x, rs[j], x) {
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

// ParentDiffers reports whether some parent-dependent pair (r, t) with LHS(r) = b is forbidden under one of the
// parent items p1 and p2 and not the other.
func ParentDiffers(p *Precedence, b, p1, p2 int) bool {
	tab := p.Tab
	ni := len(p.TopAt)
	for r, rule := range tab.Rules {
		if rule.LHS != b {
			continue
		}
		for t := range p.NT {
			k := r*p.NT + t
			if p.ReduceAt[k] != 2 {
				continue
			}
			if row := int(p.MixAt[k]) * ni; p.ParentForbid[row+p1] != p.ParentForbid[row+p2] {
				return true
			}
		}
	}
	return false
}

// Awaiting groups items by the parent-dependent symbol after their dot.
func Awaiting(p *Precedence, items []int) map[int][]int {
	tab := p.Tab
	by := map[int][]int{}
	for _, i := range items {
		it := tab.Items[i]
		if rhs := tab.Rules[it.Rule].RHS; it.Dot < len(rhs) && p.NeedParent[rhs[it.Dot]] {
			by[rhs[it.Dot]] = append(by[rhs[it.Dot]], i)
		}
	}
	return by
}

// ParentConflicts counts the states holding two core items with the dot before one parent-dependent symbol whose
// parentForbid verdicts differ.
func ParentConflicts(p *Precedence, states []grammar.State) (conflicts int) {
	for _, st := range states {
	state:
		for b, is := range Awaiting(p, st.Items) {
			for x := range is {
				for y := x + 1; y < len(is); y++ {
					if ParentDiffers(p, b, is[x], is[y]) {
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
	if i >= 0 && p.TopAt[i] < 0 {
		return -1
	}
	return int32(i)
}

// ruleTops is, by rule, the canonical spine tops of the rules k begins: each hangs from a kernel item of k, or
// from an item of k's predicted state advanced over a nullable, and a left-corner prediction inherits its tops.
func ruleTops(p *Precedence, k predictor) map[int]map[int32]bool {
	tab := p.Tab
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
			add(r, canonTop(p, int(p.Base[r])))
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

// ShiftRec is a shift of f after rule-position item at under canonical spine top top (-1 none).
type ShiftRec struct{ At, Top, F int32 }

// ShiftVerdict judges one item's shifts: its core items carry two or more tops (multi), because one core item
// carries several (same) or because core items carry different ones (cross); the tops give two or more verdict
// vectors over its shifts (copies); a top of another core item allows a shift every top of at forbids (overKeep).
type ShiftVerdict struct{ Multi, Same, Cross, Copies, OverKeep bool }

// ContextConflict is check (a)'s conflict: core items with different tops whose verdicts differ.
func (v ShiftVerdict) ContextConflict() bool { return v.Cross && (v.Copies || v.OverKeep) }

func JudgeShifts(p *Precedence, xs []ShiftRec) (v ShiftVerdict) {
	real := map[int32]map[int32]bool{} // at -> its tops
	all := map[int32]bool{}
	for _, x := range xs {
		if real[x.At] == nil {
			real[x.At] = map[int32]bool{}
		}
		real[x.At][x.Top], all[x.Top] = true, true
	}
	if len(all) < 2 {
		return v
	}
	v.Multi = true
	for _, tops := range real {
		v.Same = v.Same || len(tops) >= 2
		v.Cross = v.Cross || len(tops) < len(all)
	}
	vectors := map[string]bool{}
	for top := range all {
		var b strings.Builder
		for _, x := range xs {
			fmt.Fprint(&b, p.ShiftForbidden(top, x.At, x.F))
		}
		vectors[b.String()] = true
	}
	v.Copies = len(vectors) >= 2
	for _, x := range xs {
		forbidden := true
		for tp := range real[x.At] {
			forbidden = forbidden && p.ShiftForbidden(tp, x.At, x.F)
		}
		for c := range all {
			v.OverKeep = v.OverKeep || forbidden && !real[x.At][c] && !p.ShiftForbidden(c, x.At, x.F)
		}
	}
	return v
}

// PredictorTopsConflict is check (b) on one predictor's rule tops: conflict where they hold two different non-none
// tops; withNone where they hold one non-none top beside none.
func PredictorTopsConflict(rt map[int]map[int32]bool) (conflict, withNone bool) {
	real := map[int32]bool{}
	for _, ts := range rt {
		maps.Copy(real, ts)
	}
	none := real[-1]
	delete(real, -1)
	return len(real) >= 2, len(real) == 1 && none
}

// RuleTopsDisagree is check (c) on one predictor: hit where a rule it begins under two or more tops gets different
// verdicts from them on one of shifts, its (rule-position item, terminal) pairs; two where any rule has two tops.
func RuleTopsDisagree(p *Precedence, rt map[int]map[int32]bool, shifts [][2]int) (hit, two bool) {
	tab := p.Tab
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
				seen[p.ShiftForbidden(c, int32(sh[0]), int32(sh[1]))] = true
			}
			hit = hit || len(seen) > 1
		}
	}
	return hit, two
}

// offAtShifts counts p's forbidden (top, at, f) shifts over every automaton item at, and of them those whose f is
// not at's own next terminal: at is complete, a nonterminal or another terminal follows its dot, or at is a
// refused rule's item, which no item of the tables holds.
func offAtShifts(p *Precedence) (forbidden, off int) {
	tab := p.Tab
	for top, row := range p.TopAt {
		if row < 0 {
			continue
		}
		for at := range p.TopAt {
			for f := range p.NT {
				if !p.ShiftForbidden(int32(top), int32(at), int32(f)) {
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
func nullableNonAllow(p *Precedence) (empty, nonEmpty int, bad []int) {
	tab := p.Tab
	for r, rule := range tab.Rules {
		if !slices.ContainsFunc(rule.RHS, func(s int) bool { return !tab.Symbols[s].Nullable }) {
			if len(rule.RHS) == 0 {
				empty++
			} else {
				nonEmpty++
			}
			if slices.ContainsFunc(p.ReduceAt[r*p.NT:(r+1)*p.NT], func(v uint8) bool { return v != 0 }) {
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
func scanDisagreements(p *Precedence) (groups int, conflicts []int32, dot0 int) {
	tab := p.Tab
	contexts := []int32{-1}
	for i, row := range p.TopAt {
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
					verdicts[dot > 0 && p.ShiftForbidden(c, int32(i), int32(f))] = true
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
