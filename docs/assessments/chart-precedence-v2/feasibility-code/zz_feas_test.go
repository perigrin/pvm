package forest

import (
	"fmt"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
)

func TestZZStatic(t *testing.T) {
	tab := committedTables(t)
	p, err := ParsePrecedence(tab)
	if err != nil {
		t.Fatal(err)
	}
	for r := range tab.Rules {
		if it := tab.Items[p.base[r]]; it.Rule != r || it.Dot != 0 {
			t.Fatalf("base[%d] is not rule %d's dot-0 item", r, r)
		}
	}
	t.Logf("states=%d rules=%d items=%d (precedence items incl refused=%d) terminals+EOF=%d canonical tops=%d",
		len(tab.States), len(tab.Rules), len(tab.Items), len(p.topAt), p.nt, len(p.ZZTops()))

	// ---- M1
	same, conf, lines := ZZM1(tab, p, tab.States)
	t.Logf("M1: states reducing >=2 rules of one LHS=%d; with a differing V1 reduce verdict=%d", same, conf)
	for _, l := range lines {
		t.Logf("M1   %s", l)
	}
	// non-vacuity: the globally differing same-LHS pairs, each placed in a synthetic state, must be counted
	byLHS := map[int][]int{}
	for r, rule := range tab.Rules {
		byLHS[rule.LHS] = append(byLHS[rule.LHS], r)
	}
	var synth []grammar.State
	for lhs, rs := range byLHS {
		for i := range rs {
			for j := i + 1; j < len(rs); j++ {
				for _, x := range tab.Follow[lhs] {
					if p.ZZReduceDiffers(rs[i], rs[j], x) {
						synth = append(synth, grammar.State{Reduces: []int{rs[i], rs[j]}})
						break
					}
				}
			}
		}
	}
	_, sc, _ := ZZM1(tab, p, synth)
	t.Logf("M1 non-vacuity: same-LHS rule pairs anywhere in perly.y with a differing verdict=%d; synthetic states counted=%d", len(synth), sc)
	if sc != len(synth) || len(synth) == 0 {
		t.Errorf("M1 counter vacuous: %d of %d", sc, len(synth))
	}

	// ---- M2
	for _, l := range ZZMixedPairs(tab, p) {
		t.Logf("mixed pair: %s", l)
	}
	sh, needy, conf2, lines2 := ZZM2(tab, p, tab.States)
	t.Logf("M2: states with >=2 core items awaiting one nonterminal=%d; of them awaiting a needParent symbol=%d; with differing parentForbid=%d", sh, needy, conf2)
	for _, l := range lines2 {
		t.Logf("M2   %s", l)
	}
	// non-vacuity: every pair of items anywhere awaiting one needParent symbol with differing parentForbid, in a synthetic state
	await := map[int][]int{}
	for i, it := range tab.Items {
		if rhs := tab.Rules[it.Rule].RHS; it.Dot < len(rhs) && p.needParent[rhs[it.Dot]] {
			await[rhs[it.Dot]] = append(await[rhs[it.Dot]], i)
		}
	}
	var synth2 []grammar.State
	for b, is := range await {
		for x := range is {
			for y := x + 1; y < len(is); y++ {
				if len(p.ZZParentDiffers(tab, b, is[x], is[y])) > 0 {
					synth2 = append(synth2, grammar.State{Items: []int{is[x], is[y]}})
				}
			}
		}
	}
	_, _, sc2, _ := ZZM2(tab, p, synth2)
	t.Logf("M2 non-vacuity: item pairs anywhere awaiting one symbol with differing parentForbid=%d; synthetic states counted=%d", len(synth2), sc2)
	if sc2 != len(synth2) || len(synth2) == 0 {
		t.Errorf("M2 counter vacuous: %d of %d", sc2, len(synth2))
	}
	for b, is := range await {
		t.Logf("needParent symbol %s awaited by %d core items", tab.Symbols[b].Name, len(is))
	}
}

func TestZZStaticDetail(t *testing.T) {
	tab := committedTables(t)
	p, _ := ParsePrecedence(tab)
	for s, st := range tab.States {
		by := map[int][]int{}
		for _, r := range st.Reduces {
			by[tab.Rules[r].LHS] = append(by[tab.Rules[r].LHS], r)
		}
		for _, rs := range by {
			if len(rs) >= 2 {
				for _, r := range rs {
					t.Logf("M1 state %d (predicted=%v) reduces %s", s, st.Predicted, zzRule(tab, r))
				}
			}
		}
		aw := map[int][]int{}
		for _, i := range st.Items {
			it := tab.Items[i]
			if rhs := tab.Rules[it.Rule].RHS; it.Dot < len(rhs) && p.needParent[rhs[it.Dot]] {
				aw[rhs[it.Dot]] = append(aw[rhs[it.Dot]], i)
			}
		}
		for b, is := range aw {
			if len(is) >= 2 {
				for _, i := range is {
					t.Logf("M2 state %d (predicted=%v) awaits %s: [%s]", s, st.Predicted, tab.Symbols[b].Name, zzItem(tab, i))
				}
			}
		}
	}
}

func TestZZM3NonVacuous(t *testing.T) {
	tab := committedTables(t)
	p, _ := ParsePrecedence(tab)
	var at, top, f int32 = -1, -1, -1
	n := 0
	for _, tp := range p.ZZTops() {
		for i := range tab.Items {
			for x := range p.nt {
				if p.ZZForbid(tp, int32(i), int32(x)) {
					n++
					if at < 0 {
						at, top, f = int32(i), tp, int32(x)
					}
				}
			}
		}
	}
	t.Logf("forbidden (top, at, f) triples in perly.y's tables: %d; using [%s] under %s, %s", n, zzItem(tab, int(at)), zzTopText(tab, top), zzSym(tab, int(f)))
	other := (at + 7) % int32(len(tab.Items))
	var cross ZZM3Stats
	ZZJudgeItem(tab, p, 1, []ZZShift{{0, 1, at, top, f}, {0, 1, other, -1, f}}, &cross)
	t.Logf("cross case: %+v", cross)
	if cross.MultiTop != 1 || cross.CrossCore != 1 || cross.NeedCopies != 1 || cross.NeedCopiesCross != 1 || cross.OverKeep != 1 || cross.SameCore != 0 {
		t.Errorf("cross case not counted: %+v", cross)
	}
	var same ZZM3Stats
	ZZJudgeItem(tab, p, 1, []ZZShift{{0, 1, at, top, f}, {0, 1, at, -1, f}}, &same)
	t.Logf("same case: %+v", same)
	if same.MultiTop != 1 || same.SameCore != 1 || same.CrossCore != 0 || same.NeedCopies != 1 || same.OverKeep != 0 {
		t.Errorf("same case miscounted: %+v", same)
	}
	var one ZZM3Stats
	ZZJudgeItem(tab, p, 1, []ZZShift{{0, 1, at, top, f}, {0, 1, other, top, f}}, &one)
	if one.MultiTop != 0 {
		t.Errorf("one-top case counted: %+v", one)
	}
}

func TestZZM3Static(t *testing.T) {
	tab := committedTables(t)
	p, _ := ParsePrecedence(tab)
	var st ZZM3Stats
	pairs, multi := ZZM3Static(tab, p, &st)
	t.Logf("M3 static: (state, core item, top) pairs=%d; core items with >=2 tops=%d; states judged=%d", pairs, multi, st.Items)
	t.Logf("M3 static: states with >=2 canonical tops=%d (same core item >=2: %d; core items differ: %d); needing >=2 verdict vectors=%d (cross-core: %d); over-keep=%d",
		st.MultiTop, st.SameCore, st.CrossCore, st.NeedCopies, st.NeedCopiesCross, st.OverKeep)
	for _, e := range ZZTop(st.Examples, 20) {
		t.Logf("M3s cross ex %s", e)
	}
	for _, e := range ZZTop(st.OverExamples, 20) {
		t.Logf("M3s over ex %s", e)
	}
}

func TestZZM3StaticDetail(t *testing.T) {
	tab := committedTables(t)
	p, _ := ParsePrecedence(tab)
	tops := ZZM3StaticTops(tab, p)
	for s, st := range tab.States {
		U := map[int32]bool{}
		per := map[int]map[int32]bool{}
		for _, i := range st.Items {
			it := tab.Items[i]
			if it.Dot == 0 || it.Dot >= len(tab.Rules[it.Rule].RHS) || len(tops[s][it.Rule]) == 0 {
				continue
			}
			per[i] = tops[s][it.Rule]
			for x := range tops[s][it.Rule] {
				U[x] = true
			}
		}
		cross, needs := false, false
		for _, r := range per {
			if len(r) < len(U) {
				cross = true
			}
		}
		var one ZZM3Stats
		var xs []ZZShift
		for i, rt := range per {
			x := tab.Rules[tab.Items[i].Rule].RHS[tab.Items[i].Dot]
			fs := tab.First[x]
			if tab.Symbols[x].Terminal {
				fs = []int{x}
			}
			for tp := range rt {
				for _, f := range fs {
					xs = append(xs, ZZShift{0, 0, int32(i), tp, int32(f)})
				}
			}
		}
		if len(xs) > 0 {
			ZZJudgeItem(tab, p, int32(s), xs, &one)
		}
		needs = one.NeedCopies > 0
		if cross || needs {
			t.Logf("state %d cross=%v needCopies=%v:", s, cross, needs)
			for i, r := range per {
				t.Logf("    [%s] tops %s", zzItem(tab, i), zzTopsText(tab, r))
			}
		}
	}
}

func TestZZM4Static(t *testing.T) {
	tab := committedTables(t)
	p, _ := ParsePrecedence(tab)
	semi, brace := -1, -1
	for i, s := range tab.Symbols {
		switch s.Name {
		case "PERLY_SEMICOLON":
			semi = i
		case "PERLY_BRACE_CLOSE":
			brace = i
		}
	}
	n := 0
	for r := range tab.Rules {
		if zzReduceDiff(p, r, semi, brace) {
			n++
			t.Logf("reduce %s differs before ; and }", zzRule(tab, r))
		}
	}
	t.Logf("rules whose V1 reduce verdict differs between PERLY_SEMICOLON and PERLY_BRACE_CLOSE: %d; levels ; %d } %d", n, p.term[semi], p.term[brace])
	// shift side: an item whose next symbol X is a nonterminal, a canonical top, and f1, f2 in FIRST(X) with differing verdicts
	cnt := 0
	ex := map[string]int{}
	for i, it := range tab.Items {
		rhs := tab.Rules[it.Rule].RHS
		if it.Dot == 0 || it.Dot >= len(rhs) || tab.Symbols[rhs[it.Dot]].Terminal {
			continue
		}
		fs := tab.First[rhs[it.Dot]]
		for _, tp := range p.ZZTops() {
			var forb, allow []string
			for _, f := range fs {
				if p.ZZForbid(tp, int32(i), int32(f)) {
					forb = append(forb, zzSym(tab, f))
				} else {
					allow = append(allow, zzSym(tab, f))
				}
			}
			if len(forb) > 0 && len(allow) > 0 {
				cnt++
				ex[fmt.Sprintf("after [%s] under %s forbids %v", zzItem(tab, i), zzTopText(tab, tp), forb)]++
			}
		}
	}
	t.Logf("(item with nonterminal next, top) pairs whose FIRST terminals get differing shift verdicts: %d", cnt)
	for _, e := range ZZTop(ex, 10) {
		t.Logf("  %s", e)
	}
	// reduce side: rule pairs of terminals in one FOLLOW with differing verdicts (scale)
	rd := 0
	for r, rule := range tab.Rules {
		fl := tab.Follow[rule.LHS]
		for a := range fl {
			for b := a + 1; b < len(fl); b++ {
				if zzReduceDiff(p, r, fl[a], fl[b]) {
					rd++
				}
			}
		}
	}
	t.Logf("(rule, t1, t2 in FOLLOW) triples with differing reduce verdicts: %d", rd)
	var plug []string
	for x := range tab.Symbols {
		if ZZIsPlugin(tab, x) {
			plug = append(plug, fmt.Sprintf("%s=%d", tab.Symbols[x].Name, p.term[x]))
		}
	}
	t.Logf("PLUGIN_* levels: %v", plug)
}

func TestZZM4StaticSets(t *testing.T) {
	tab := committedTables(t)
	p, _ := ParsePrecedence(tab)
	semi, brace := -1, -1
	for i, s := range tab.Symbols {
		switch s.Name {
		case "PERLY_SEMICOLON":
			semi = i
		case "PERLY_BRACE_CLOSE":
			brace = i
		}
	}
	inFollow := func(lhs, x int) bool {
		for _, y := range tab.Follow[lhs] {
			if y == x {
				return true
			}
		}
		return false
	}
	multi, bad, badUn := 0, 0, 0
	ex := map[string]int{}
	check := func(set []int) string {
		la := append([]int{}, set...)
		has := func(x int) bool {
			for _, y := range la {
				if y == x {
					return true
				}
			}
			return false
		}
		if has(brace) && !has(semi) {
			la = append(la, semi)
		}
		for r, rule := range tab.Rules {
			for a := range la {
				for b := a + 1; b < len(la); b++ {
					if inFollow(rule.LHS, la[a]) && inFollow(rule.LHS, la[b]) && zzReduceDiff(p, r, la[a], la[b]) {
						return fmt.Sprintf("%s before %s vs %s", zzRule(tab, r), zzSym(tab, la[a]), zzSym(tab, la[b]))
					}
				}
			}
		}
		return ""
	}
	for id, set := range tab.TerminalSets {
		if len(set) >= 2 {
			multi++
		}
		if d := check(set); d != "" {
			badUn++
			_ = id
		}
		var base, plugs []int
		for _, x := range set {
			if ZZIsPlugin(tab, x) {
				plugs = append(plugs, x)
			} else {
				base = append(base, x)
			}
		}
		choices := [][]int{base}
		for _, pl := range plugs {
			choices = append(choices, append(append([]int{}, base...), pl))
		}
		for _, c := range choices {
			if d := check(c); d != "" {
				bad++
				ex[d]++
				break
			}
		}
	}
	t.Logf("terminal sets=%d with >=2 terminals=%d; differing reduce verdict unnarrowed=%d; after narrowing to <=1 PLUGIN_* (any choice)=%d", len(tab.TerminalSets), multi, badUn, bad)
	for _, e := range ZZTop(ex, 10) {
		t.Logf("  %s", e)
	}
}

func TestZZIndirob(t *testing.T) {
	tab := committedTables(t)
	p, _ := ParsePrecedence(tab)
	id := map[string]int{}
	for i, s := range tab.Symbols {
		id[s.Name] = i
	}
	for _, n := range []string{"HASHBRACK", "PERLY_BRACE_OPEN"} {
		t.Logf("indirob ::= scalar before %s: reduceAt=%d", n, p.reduceAt[347*p.nt+id[n]])
	}
	for sid, set := range tab.TerminalSets {
		h, b := false, false
		for _, x := range set {
			h = h || x == id["HASHBRACK"]
			b = b || x == id["PERLY_BRACE_OPEN"]
		}
		if h && b {
			var w []string
			for _, x := range set {
				w = append(w, tab.Symbols[x].Name)
			}
			t.Logf("set %d: %v", sid, w)
		}
	}
}

// TestZZM3StaticPerE: an item predicted in E at origin o, if it carries one context chosen from U_E (every
// canonical top any rule of E has), keeps that context in every state reached from E. Over-keep: some core item
// at of a reached state, whose real tops (its rule's in E) all forbid shifting f, is allowed it under a C in U_E.
func TestZZM3StaticPerE(t *testing.T) {
	tab := committedTables(t)
	p, _ := ParsePrecedence(tab)
	type res struct{ es, multiU, over int }
	var r res
	ex := map[string]int{}
	ZZPerE = func(e int, ruleTop map[int]map[int32]bool) {
		r.es++
		U := map[int32]bool{}
		for _, ts := range ruleTop {
			for x := range ts {
				U[x] = true
			}
		}
		if len(U) >= 2 {
			r.multiU++
		}
		seen := map[int]bool{}
		work := []int{e}
		hit := false
		for len(work) > 0 {
			s := work[len(work)-1]
			work = work[:len(work)-1]
			if seen[s] {
				continue
			}
			seen[s] = true
			for _, i := range tab.States[s].Items {
				it := tab.Items[i]
				rhs := tab.Rules[it.Rule].RHS
				rt := ruleTop[it.Rule]
				if it.Dot == 0 || it.Dot >= len(rhs) || len(rt) == 0 {
					continue
				}
				x := rhs[it.Dot]
				fs := tab.First[x]
				if tab.Symbols[x].Terminal {
					fs = []int{x}
				}
				for _, f := range fs {
					all := true
					for tp := range rt {
						if !p.ZZForbid(tp, int32(i), int32(f)) {
							all = false
						}
					}
					if !all {
						continue
					}
					for c := range U {
						if !rt[c] && !p.ZZForbid(c, int32(i), int32(f)) {
							hit = true
							ex[fmt.Sprintf("E=%d state %d: shift %s after [%s] forbidden under %s, allowed under %s", e, s, zzSym(tab, f), zzItem(tab, i), zzTopsText(tab, rt), zzTopText(tab, c))]++
						}
					}
				}
			}
			for _, tr := range tab.States[s].Transitions {
				if !tab.States[tr.Target].Predicted {
					work = append(work, tr.Target)
				}
			}
		}
		if hit {
			r.over++
		}
	}
	defer func() { ZZPerE = nil }()
	ZZM3StaticTops(tab, p)
	t.Logf("predicted states=%d; with >=2 canonical tops among their rules=%d; over-keep if one context from U_E is carried through its descendants=%d", r.es, r.multiU, r.over)
	for _, e := range ZZTop(ex, 30) {
		t.Logf("  %s", e)
	}
}

// TestZZM3PerK: for each kernel state K (and the start state), the rules its predicted state E begins get their
// canonical spine tops from K's own items. Does one canonical context C, carried by the (E, o) item and every
// item reached from it at origin o, give every shift of every such rule the verdict its real top gives?
func TestZZM3PerK(t *testing.T) {
	tab := committedTables(t)
	p, _ := ParsePrecedence(tab)
	cands := append([]int32{-1}, p.ZZTops()...)
	type pk struct{ k, e int }
	var preds []pk
	preds = append(preds, pk{-1, 0})
	for k, st := range tab.States {
		if !st.Predicted && st.Epsilon >= 0 {
			preds = append(preds, pk{k, st.Epsilon})
		}
	}
	multiTopRule, multiTopRuleDiffer, multiU, noSingle := 0, 0, 0, 0
	desc := map[int]map[[2]int32]bool{}
	for _, x := range preds {
		var ks []int
		if x.k >= 0 {
			ks = []int{x.k}
		}
		rt := ZZRuleTops(tab, p, x.e, ks)
		if desc[x.e] == nil {
			desc[x.e] = ZZDescShifts(tab, x.e)
		}
		sh := desc[x.e]
		U := map[int32]bool{}
		mt, mtd := false, false
		for r, ts := range rt {
			for c := range ts {
				U[c] = true
			}
			if len(ts) >= 2 {
				mt = true
				// differing verdicts among this rule's own tops on its shifts
				for s := range sh {
					if tab.Items[s[0]].Rule != r {
						continue
					}
					var v []bool
					for c := range ts {
						v = append(v, p.ZZForbid(c, s[0], s[1]))
					}
					for _, b := range v[1:] {
						if b != v[0] {
							mtd = true
						}
					}
				}
			}
		}
		if mt {
			multiTopRule++
			t.Logf("rule with >=2 tops under K=%d E=%d", x.k, x.e)
		}
		if mtd {
			multiTopRuleDiffer++
			t.Logf("K=%d E=%d: a rule with >=2 tops whose verdicts differ (needs per-parent copies)", x.k, x.e)
		}
		if len(U) < 2 {
			continue
		}
		multiU++
		// exists one C exact for every rule with a single top
		found := int32(-3)
		order := []int32{}
		for c := range U {
			order = append(order, c)
		}
		order = append(order, cands...)
		for _, c := range order {
			ok := true
			for s := range sh {
				ts := rt[tab.Items[s[0]].Rule]
				if len(ts) != 1 {
					continue
				}
				for real := range ts {
					if p.ZZForbid(c, s[0], s[1]) != p.ZZForbid(real, s[0], s[1]) {
						ok = false
					}
				}
			}
			if ok {
				found = c
				break
			}
		}
		var w []string
		for c := range U {
			w = append(w, zzTopText(tab, c))
		}
		kname := "start"
		if x.k >= 0 {
			var ki []string
			for _, i := range tab.States[x.k].Items {
				ki = append(ki, zzItem(tab, i))
			}
			kname = fmt.Sprintf("%d %v", x.k, ki)
		}
		if found == -3 {
			noSingle++
			t.Logf("NO single context: K=%s E=%d tops %v", kname, x.e, w)
		} else {
			t.Logf("single context %s serves K=%s E=%d tops %v", zzTopText(tab, found), kname, x.e, w)
		}
	}
	t.Logf("predictors (kernel states with a predicted state, + start)=%d; with a rule carrying >=2 tops=%d (verdicts differ: %d); with >=2 tops among rules=%d; of them no single context exact=%d",
		len(preds), multiTopRule, multiTopRuleDiffer, multiU, noSingle)
}

func TestZZForbidShape(t *testing.T) {
	tab := committedTables(t)
	p, _ := ParsePrecedence(tab)
	term, nonterm, notFirst := 0, 0, 0
	for _, tp := range p.ZZTops() {
		for i, it := range tab.Items {
			rhs := tab.Rules[it.Rule].RHS
			for x := range p.nt {
				if !p.ZZForbid(tp, int32(i), int32(x)) {
					continue
				}
				switch {
				case it.Dot >= len(rhs):
					notFirst++
				case tab.Symbols[rhs[it.Dot]].Terminal:
					term++
					if rhs[it.Dot] != x {
						notFirst++
					}
				default:
					nonterm++
				}
			}
		}
	}
	t.Logf("forbidden (top, at, f): at's next symbol a terminal=%d, a nonterminal=%d; f not at's next terminal / at complete=%d", term, nonterm, notFirst)
}
