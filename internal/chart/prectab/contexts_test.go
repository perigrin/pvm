// ABOUTME: Tests PredictionContexts on perly.y: §4.2's canonical tops, §4.3's collapse against a brute-force witness search,
// ABOUTME: and the three predictors (do, UNIOP, my \) that begin rules under a real top beside none.

package prectab

import (
	"slices"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
)

// coreText is the tables' core item i as "lhs: x . y".
func coreText(tab *grammar.Tables, i int) string {
	it := tab.Items[i]
	w := []string{tab.Symbols[tab.Rules[it.Rule].LHS].Name + ":"}
	for d, s := range tab.Rules[it.Rule].RHS {
		if d == it.Dot {
			w = append(w, ".")
		}
		w = append(w, tab.Symbols[s].Name)
	}
	if it.Dot == len(tab.Rules[it.Rule].RHS) {
		w = append(w, ".")
	}
	return strings.Join(w, " ")
}

// witness is a core item, in a state reached from predicted state e by kernel transitions, after which top forbids
// some shift, as its text, or "": read off top's whole ShiftAt row, every terminal, not only the item's next ones.
func witness(p *Precedence, top int32, e int) string {
	tab := p.Tab
	if p.TopAt[top] < 0 {
		return ""
	}
	row := int(p.TopAt[top]) * len(p.TopAt)
	seen, work := map[int]bool{e: true}, []int{e}
	for len(work) > 0 {
		s := work[0]
		work = work[1:]
		for _, i := range tab.States[s].Items {
			if x := p.ShiftAt[row+int(p.Item(i))]; x >= 0 && slices.ContainsFunc(p.ShiftForbid[x], func(w uint64) bool { return w != 0 }) {
				return coreText(tab, i)
			}
		}
		for _, tr := range tab.States[s].Transitions {
			if !tab.States[tr.Target].Predicted && !seen[tr.Target] {
				seen[tr.Target] = true
				work = append(work, tr.Target)
			}
		}
	}
	return ""
}

// TestContextCollapseTable: on perly.y every predictor's context is none or a canonical spine top, the predictor's
// own (§4.2); it is that top exactly where the top forbids a shift some item reachable at its origin makes, and none
// otherwise (§4.3), each judged by a brute-force search over every terminal; the three predictors that begin rules
// under a real top beside none (do, UNIOP, my \) give their real top. Each population is counted, so none is empty.
func TestContextCollapseTable(t *testing.T) {
	_, tab, p := committed(t)
	c := PredictionContexts(p)
	if len(c.Kernel) != len(tab.States) {
		t.Fatalf("%d kernel entries for %d states", len(c.Kernel), len(tab.States))
	}
	canonical := func(x int32) bool { return x == -1 || x >= 0 && int(x) < len(p.TopAt) && p.TopAt[x] >= 0 }
	var kept, collapsed43, collapsed42, predictors int
	check := func(name string, ctx, raw int32, e int) {
		predictors++
		if !canonical(ctx) || !canonical(raw) {
			t.Errorf("%s: context %d, raw %d; want none or a canonical spine top", name, ctx, raw)
		}
		switch w := ""; {
		case raw < 0:
			collapsed42++
			if ctx != -1 {
				t.Errorf("%s: context %d under no real top", name, ctx)
			}
		case func() bool { w = witness(p, raw, e); return w != "" }():
			kept++
			if ctx != raw {
				t.Errorf("%s: context %d, want its top %d, which forbids a shift after %s", name, ctx, raw, w)
			}
		default:
			collapsed43++
			if ctx != -1 {
				t.Errorf("%s: context %d, want none: its top forbids no reachable shift", name, ctx)
			}
		}
	}
	check("the start", c.Start, rawTop(p, predictor{-1, 0}), 0)
	for s, st := range tab.States {
		if st.Predicted || st.Epsilon < 0 {
			if c.Kernel[s] != -1 {
				t.Errorf("state %d predicts nothing but has context %d", s, c.Kernel[s])
			}
			continue
		}
		check(coreText(tab, st.Items[0]), c.Kernel[s], rawTop(p, predictor{s, st.Epsilon}), st.Epsilon)
	}
	t.Logf("predictors %d: kept %d, collapsed by §4.3 %d, by §4.2 %d", predictors, kept, collapsed43, collapsed42)
	if kept == 0 || collapsed42 == 0 {
		t.Errorf("kept %d, collapsed by §4.2 %d: a population is empty", kept, collapsed42)
	}

	// §4.3's predicate on its own, every canonical top against every predictor's region: perly.y gives each predictor
	// a top that forbids a shift in its own region, so the pairs where it forbids none come from other predictors.
	var regions []int
	for _, st := range tab.States {
		if !st.Predicted && st.Epsilon >= 0 {
			regions = append(regions, st.Epsilon)
		}
	}
	regions = append(regions, 0)
	forbids, none := 0, 0
	for top, row := range p.TopAt {
		if row < 0 {
			continue
		}
		for _, e := range regions {
			got, want := forbidsAt(p, int32(top), e), witness(p, int32(top), e) != ""
			if got != want {
				t.Errorf("top %d over the region of predicted state %d: forbidsAt %v, the witness search %v", top, e, got, want)
			}
			if want {
				forbids++
			} else {
				none++
			}
		}
	}
	t.Logf("(top, region) pairs: %d forbid a shift there, %d none", forbids, none)
	if forbids == 0 || none == 0 {
		t.Errorf("(top, region) pairs: %d forbid, %d none; want some of each", forbids, none)
	}

	// The three mixed predictors: a kernel state holding the item gives that item, beside rules it begins under none.
	for _, top := range []string{"termdo: KW_DO . term", "term: UNIOP . term", "myattrterm: KW_MY REFGEN . term"} {
		found := 0
		for s, st := range tab.States {
			for _, i := range st.Items {
				if st.Predicted || st.Epsilon < 0 || coreText(tab, i) != top {
					continue
				}
				found++
				if _, withNone := PredictorTopsConflict(ruleTops(p, predictor{s, st.Epsilon})); !withNone {
					t.Errorf("state %d, holding %s, begins no rule under none", s, top)
				}
				if c.Kernel[s] != p.Item(i) {
					t.Errorf("state %d gives context %d, want %s's %d", s, c.Kernel[s], top, p.Item(i))
				}
			}
		}
		if found == 0 {
			t.Errorf("no kernel state holds %s", top)
		}
	}
}
