package forest

// The V2 design's §3 feasibility counters, pinned as tests over the committed perly.y tables (design §5): each reads
// 0 on perly.y and counts every synthetic case built to trigger it. Ported from the scratch code measured in
// docs/assessments/chart-precedence-v2/feasibility.md.

import (
	"slices"
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
