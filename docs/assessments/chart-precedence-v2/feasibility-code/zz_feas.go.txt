package forest

// Scratch feasibility instruments for chart precedence V2 (not for commit).

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/chart/lex"
	"tamarou.com/pvm/internal/chart/recognize"
)

func zzItem(tab *grammar.Tables, i int) string {
	it := tab.Items[i]
	rule := tab.Rules[it.Rule]
	w := []string{tab.Symbols[rule.LHS].Name + ":"}
	for d, s := range rule.RHS {
		if d == it.Dot {
			w = append(w, ".")
		}
		w = append(w, tab.Symbols[s].Name)
	}
	if it.Dot == len(rule.RHS) {
		w = append(w, ".")
	}
	return strings.Join(w, " ")
}

func zzRule(tab *grammar.Tables, r int) string {
	w := []string{tab.Symbols[tab.Rules[r].LHS].Name, "::="}
	for _, s := range tab.Rules[r].RHS {
		w = append(w, tab.Symbols[s].Name)
	}
	return fmt.Sprintf("r%d %s", r, strings.Join(w, " "))
}

func zzSym(tab *grammar.Tables, t int) string {
	if t == tab.EOF {
		return "EOF"
	}
	return tab.Symbols[t].Name
}

// ZZReduceDiffers: V1's reduce verdicts for (r1,t) and (r2,t) differ. Value 2 rows compared over every parent item.
func (p *Precedence) ZZReduceDiffers(r1, r2, t int) bool {
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

// ZZM1 is measure 1 over states: states reducing >=2 rules of one LHS whose V1 reduce verdict differs for some t
// in FOLLOW(LHS) (EOF included).
func ZZM1(tab *grammar.Tables, p *Precedence, states []grammar.State) (sameLHS, conflict int, lines []string) {
	for s, st := range states {
		by := map[int][]int{}
		for _, r := range st.Reduces {
			by[tab.Rules[r].LHS] = append(by[tab.Rules[r].LHS], r)
		}
		hasSame, hit := false, false
		for lhs, rs := range by {
			if len(rs) < 2 {
				continue
			}
			hasSame = true
			for i := range rs {
				for j := i + 1; j < len(rs); j++ {
					var ts []string
					for _, t := range tab.Follow[lhs] {
						if p.ZZReduceDiffers(rs[i], rs[j], t) {
							ts = append(ts, fmt.Sprintf("%s(%d/%d)", zzSym(tab, t), p.reduceAt[rs[i]*p.nt+t], p.reduceAt[rs[j]*p.nt+t]))
						}
					}
					if len(ts) > 0 {
						hit = true
						lines = append(lines, fmt.Sprintf("state %d: %s | %s | before %s", s, zzRule(tab, rs[i]), zzRule(tab, rs[j]), strings.Join(ts, " ")))
					}
				}
			}
		}
		if hasSame {
			sameLHS++
		}
		if hit {
			conflict++
		}
	}
	return
}

// ZZParentDiffers lists the mixed pairs (r,t) with LHS b whose parentForbid differs between parent items p1 and p2.
func (p *Precedence) ZZParentDiffers(tab *grammar.Tables, b, p1, p2 int) []string {
	ni := len(p.topAt)
	var out []string
	for r, rule := range tab.Rules {
		if rule.LHS != b {
			continue
		}
		for t := range p.nt {
			k := r*p.nt + t
			if p.reduceAt[k] != 2 {
				continue
			}
			row := int(p.mixAt[k]) * ni
			if p.parentForbid[row+p1] != p.parentForbid[row+p2] {
				out = append(out, fmt.Sprintf("%s before %s (forbid %v/%v)", zzRule(tab, r), zzSym(tab, t), p.parentForbid[row+p1], p.parentForbid[row+p2]))
			}
		}
	}
	return out
}

// ZZM2 is measure 2: states holding >=2 core items with the dot before one nonterminal B whose parentForbid differs
// for some mixed pair of B.
func ZZM2(tab *grammar.Tables, p *Precedence, states []grammar.State) (shared, needy, conflict int, lines []string) {
	for s, st := range states {
		by := map[int][]int{}
		for _, i := range st.Items {
			it := tab.Items[i]
			if rhs := tab.Rules[it.Rule].RHS; it.Dot < len(rhs) && !tab.Symbols[rhs[it.Dot]].Terminal {
				by[rhs[it.Dot]] = append(by[rhs[it.Dot]], i)
			}
		}
		hasShared, hasNeedy, hit := false, false, false
		for b, is := range by {
			if len(is) < 2 {
				continue
			}
			hasShared = true
			if p.needParent[b] {
				hasNeedy = true
			}
			for x := range is {
				for y := x + 1; y < len(is); y++ {
					if d := p.ZZParentDiffers(tab, b, is[x], is[y]); len(d) > 0 {
						hit = true
						lines = append(lines, fmt.Sprintf("state %d: [%s] vs [%s]: %s", s, zzItem(tab, is[x]), zzItem(tab, is[y]), strings.Join(d, "; ")))
					}
				}
			}
		}
		if hasShared {
			shared++
		}
		if hasNeedy {
			needy++
		}
		if hit {
			conflict++
		}
	}
	return
}

// ZZMixedPairs lists V1's mixed (parent-dependent) pairs and, per pair, the parent items forbidding / allowing.
func ZZMixedPairs(tab *grammar.Tables, p *Precedence) []string {
	ni := len(p.topAt)
	var out []string
	for r := range tab.Rules {
		for t := range p.nt {
			k := r*p.nt + t
			if p.reduceAt[k] != 2 {
				continue
			}
			row := int(p.mixAt[k]) * ni
			var forb []string
			for pi := range ni {
				if pi < len(tab.Items) && p.parentForbid[row+pi] {
					forb = append(forb, zzItem(tab, pi))
				}
			}
			out = append(out, fmt.Sprintf("%s before %s: forbidden under %d parent items: %s", zzRule(tab, r), zzSym(tab, t), len(forb), strings.Join(forb, " ; ")))
		}
	}
	return out
}

// ZZTops is the spine-top items with a shiftAt row (the canonical non-none tops).
func (p *Precedence) ZZTops() []int32 {
	var out []int32
	for i, row := range p.topAt {
		if row >= 0 {
			out = append(out, int32(i))
		}
	}
	return out
}

// ZZForbid is shiftForbidden exported.
func (p *Precedence) ZZForbid(top, at, f int32) bool { return p.shiftForbidden(top, at, f) }

// ---- measure 3: the shifts a V1 judgement makes, with their canonical spine tops ----

// ZZShift is one shift V1 judges: from rule-position item At, whose rule began at From, at position J, of
// terminal F, under canonical spine top Top (-1 none).
type ZZShift struct{ From, J, At, Top, F int32 }

type zzJudge struct {
	*judgement
	rec map[ZZShift]bool
}

// ZZShifts runs V1's state-free eval over f from the root and records every shift check seq makes.
func ZZShifts(f *Forest, p *Precedence) map[ZZShift]bool {
	z := &zzJudge{judgement: newJudgement(f, p), rec: map[ZZShift]bool{}}
	z.eval(f.Root, p.tab.EOF, rootTop, -1)
	return z.rec
}

func (c *zzJudge) eval(n int32, t int, top, pi int32) []int32 {
	nd := c.f.Nodes[n]
	if nd.Alt == 0 {
		return []int32{nd.Sym}
	}
	top, pi = c.canon(n, top, pi)
	for e := c.evalAt[n]; e != 0; e = c.evals[e].next {
		if x := &c.evals[e]; x.t == int32(t) && x.top == top && x.pi == pi {
			return x.out
		}
	}
	var out []int32
	for a := nd.Alt; a != 0; a = c.f.Alts[a].Next {
		r, m, at := c.alt(n, a, top)
		if nd.Sym >= 0 && !c.p.reducesFree(r, t, pi) {
			continue
		}
		out = append(out, c.seq(n, a, r, m, t, at)...)
	}
	slices.Sort(out)
	out = slices.Compact(out)
	c.evals = append(c.evals, evalEntry{int32(t), top, pi, c.evalAt[n], out})
	c.evalAt[n] = int32(len(c.evals) - 1)
	return out
}

func (c *zzJudge) seq(n, a int32, r, m int, t int, top int32) []int32 {
	if m == 0 {
		return []int32{int32(t)}
	}
	alt := c.f.Alts[a]
	at := c.p.base[r] + int32(m-1)
	rtop := top
	if m > 1 {
		rtop = at
	}
	right := c.eval(alt.Right, t, rtop, at)
	if m == 1 {
		return right
	}
	lpi := int32(-1)
	if m == 2 {
		lpi = c.p.base[r]
	}
	ctop := top
	if ctop >= 0 && c.p.topAt[ctop] < 0 {
		ctop = -1
	}
	var out []int32
	rn := c.f.Nodes[alt.Right]
	empty := rn.From == rn.To && rn.Alt != 0
	for _, f := range right {
		if !empty {
			c.rec[ZZShift{c.f.Nodes[n].From, rn.From, at, ctop, f}] = true
		}
		if !empty && c.p.shiftForbidden(top, at, f) {
			continue
		}
		out = append(out, c.eval(alt.Left, int(f), top, lpi)...)
	}
	return out
}

// ZZM3Stats aggregates measure 3 over chart items (state, origin, position).
type ZZM3Stats struct {
	Shifts, Items, Unmatched, Ambiguous   int // shift records; chart items holding one; records with no chart item; records in >1 state item
	MultiTop, SameCore, CrossCore         int // items whose shifting core items carry >=2 canonical tops; ...because one core item has >=2; ...because core items differ
	NeedCopies, NeedCopiesCross, OverKeep int // items whose tops give >=2 verdict vectors over the item's shifts; of them cross-core; items where some top of U would keep a shift V1 forbids for that core item
	Examples, UnmatchedEx, StaticMissEx   map[string]int
	StaticMiss                            int
	OverExamples                          map[string]int
}

// ZZPerE, when set, sees each predicted state's rule tops.
var ZZPerE func(e int, ruleTop map[int]map[int32]bool)

// ZZStaticTops, when set, is checked against every record ZZM3 attributes.
var ZZStaticTops []map[int]map[int32]bool

// ZZM3 attributes rec to chart c's items and judges them.
func ZZM3(tab *grammar.Tables, p *Precedence, c *recognize.Chart, next func(int) int, rec map[ZZShift]bool, s *ZZM3Stats) {
	type key struct{ from, j, at int32 }
	byKey := map[key][]ZZShift{}
	for x := range rec {
		byKey[key{x.From, x.J, x.At}] = append(byKey[key{x.From, x.J, x.At}], x)
	}
	type sitem struct{ state, from, j int32 }
	items := map[sitem][]ZZShift{}
	matched := map[key]int{}
	okTop := map[[2]int64]bool{}
	var missKeys [][2]int64
	for j, set := range c.Sets {
		if set == nil {
			continue
		}
		its := set.Items()
		if slices.Contains(c.Inserted(), j) { // the slipper's closure after its zero-width ; at j
			its = append(slices.Clone(its), c.Slipped(j, next(j))...)
		}
		for _, it := range its {
			o := int32(j - it.Dist)
			for _, ci := range tab.States[it.State].Items {
				k := key{o, int32(j), int32(ci)}
				if xs, ok := byKey[k]; ok {
					matched[k]++
					if ZZStaticTops != nil {
						for _, x := range xs {
							if ZZStaticTops[it.State][tab.Items[ci].Rule][x.Top] {
								okTop[[2]int64{int64(k.from)<<40 | int64(k.j)<<20 | int64(k.at), int64(x.Top)}] = true
							} else {
								missKeys = append(missKeys, [2]int64{int64(k.from)<<40 | int64(k.j)<<20 | int64(k.at), int64(x.Top)})
								if s.StaticMissEx == nil {
									s.StaticMissEx = map[string]int{}
								}
								s.StaticMissEx[fmt.Sprintf("state %d [%s] top %s", it.State, zzItem(tab, ci), zzTopText(tab, x.Top))]++
							}
						}
					}
					si := sitem{int32(it.State), o, int32(j)}
					items[si] = append(items[si], xs...)
				}
			}
		}
	}
	for _, m := range missKeys {
		if !okTop[m] { // no state item holding this core item here has the top statically
			s.StaticMiss++
		}
	}
	s.Shifts += len(rec)
	for k, xs := range byKey {
		switch matched[k] {
		case 0:
			s.Unmatched += len(xs)
			if s.UnmatchedEx == nil {
				s.UnmatchedEx = map[string]int{}
			}
			states := ""
			if set := c.Sets[k.j]; set != nil {
				for _, it := range set.Items() {
					if int32(k.j-int32(it.Dist)) == k.from {
						states += fmt.Sprintf(" %d", it.State)
					}
				}
			}
			s.UnmatchedEx[fmt.Sprintf("[%s] span %d; states at that origin:%s; inserted-; at j=%v", zzItem(tab, int(k.at)), k.j-k.from, states, slices.Contains(c.Inserted(), int(k.j)))]++
		case 1:
		default:
			s.Ambiguous += len(xs)
		}
	}
	for si, xs := range items {
		ZZJudgeItem(tab, p, si.state, xs, s)
	}
}

// ZZJudgeItem judges one chart item (state, origin, position) holding the shift records xs.
func ZZJudgeItem(tab *grammar.Tables, p *Precedence, state int32, xs []ZZShift, s *ZZM3Stats) {
	type sitem struct{ state int32 }
	si := sitem{state}
	if s.Examples == nil {
		s.Examples, s.OverExamples = map[string]int{}, map[string]int{}
	}
	{
		s.Items++
		real := map[int32]map[int32]bool{} // at -> its tops
		fs := map[int32]map[int32]bool{}   // at -> terminals shifted
		U := map[int32]bool{}
		for _, x := range xs {
			if real[x.At] == nil {
				real[x.At], fs[x.At] = map[int32]bool{}, map[int32]bool{}
			}
			real[x.At][x.Top] = true
			fs[x.At][x.F] = true
			U[x.Top] = true
		}
		if len(U) < 2 {
			return
		}
		s.MultiTop++
		same, cross := false, false
		for at := range real {
			if len(real[at]) >= 2 {
				same = true
			}
			if len(real[at]) < len(U) {
				cross = true
			}
		}
		if same {
			s.SameCore++
		}
		if cross {
			s.CrossCore++
		}
		// verdict vector of each top in U over every (at, f) of the item
		vec := map[string]bool{}
		for top := range U {
			var b strings.Builder
			for _, at := range slices.Sorted(mapsKeys(fs)) {
				for _, f := range slices.Sorted(mapsKeys(fs[at])) {
					if p.shiftForbidden(top, at, f) {
						b.WriteByte('1')
					} else {
						b.WriteByte('0')
					}
				}
			}
			vec[b.String()] = true
		}
		if len(vec) >= 2 {
			s.NeedCopies++
			if cross {
				s.NeedCopiesCross++
			}
		}
		// over-keep: a top C in U that is no real top of at allows (at,f) that every real top of at forbids
		over := ""
		for at, tops := range real {
			for f := range fs[at] {
				all := true
				for tp := range tops {
					if !p.shiftForbidden(tp, at, f) {
						all = false
					}
				}
				if !all {
					continue
				}
				for cc := range U {
					if !tops[cc] && !p.shiftForbidden(cc, at, f) {
						over = fmt.Sprintf("state %d: shift %s after [%s] forbidden under its tops %s, allowed under %s (another core item's)", si.state, zzSym(tab, int(f)), zzItem(tab, int(at)), zzTopsText(tab, tops), zzTopText(tab, cc))
					}
				}
			}
		}
		if over != "" {
			s.OverKeep++
			s.OverExamples[over]++
		}
		if len(vec) >= 2 && cross {
			var b strings.Builder
			fmt.Fprintf(&b, "state %d:", si.state)
			for _, at := range slices.Sorted(mapsKeys(real)) {
				fmt.Fprintf(&b, " [%s] tops %s;", zzItem(tab, int(at)), zzTopsText(tab, real[at]))
			}
			s.Examples[b.String()]++
		}
	}
}

func mapsKeys[K int32, V any](m map[K]V) func(func(K) bool) {
	return func(yield func(K) bool) {
		for k := range m {
			if !yield(k) {
				return
			}
		}
	}
}

func zzTopText(tab *grammar.Tables, t int32) string {
	if t < 0 {
		return "{none}"
	}
	return "{" + zzItem(tab, int(t)) + "}"
}

func zzTopsText(tab *grammar.Tables, ts map[int32]bool) string {
	var w []string
	for _, t := range slices.Sorted(mapsKeys(ts)) {
		w = append(w, zzTopText(tab, t))
	}
	return strings.Join(w, ",")
}

// ---- measure 4: terminal sets after libraryOperators' narrowing ----

// ZZDead is libraryOperators' dead (token, terminal) pairs, computed as it does.
func ZZDead(tab *grammar.Tables, s *lex.Stream, src []byte, scope Scope) map[[2]int]bool {
	g := grammar.Derived(tab, buildPluginGrammar)
	r := &resolver{tab: tab, s: s, src: src, n: len(s.Tokens), dead: map[[2]int]bool{}}
	var from []int
	for k, t := range s.Tokens {
		if !g.sets[t.TerminalSet] {
			continue
		}
		if from == nil {
			for _, op := range scope.LibraryOperators {
				from = append(from, r.statementEnd(op.From))
			}
		}
		keep := -1
		for i, op := range scope.LibraryOperators {
			if from[i] <= s.Raw[t.Index].Start && op.Name == r.text(k) {
				keep = g.id[grammar.LibraryTerminal(op)]
			}
		}
		for _, x := range tab.TerminalSets[t.TerminalSet] {
			if g.is[x] && x != keep {
				r.dead[[2]int{k, x}] = true
			}
		}
	}
	return r.dead
}

// ZZIsPlugin reports whether terminal x is a PLUGIN_* terminal.
func ZZIsPlugin(tab *grammar.Tables, x int) bool {
	return grammar.Derived(tab, buildPluginGrammar).is[x]
}

// ZZM4Stats aggregates measure 4.
type ZZM4Stats struct {
	Tokens, Plugin, PluginAfter, Multi                       int
	ReduceDiff, ShiftDiffAny, ShiftDiffObserved, AnyDiff     int
	ReduceDiffSlip                                           int // reduce differences that need the slipper's ; beside }
	Examples                                                 map[string]int
}

// ZZM4 judges, for each token k with effective terminal set ts(k), whether two of its terminals get different V1
// verdicts at the completions and shifts chart set k actually holds. observed maps (from, j, at) to the canonical
// tops V1's eval met there (measure 3's records), nil for none.
func ZZM4(tab *grammar.Tables, p *Precedence, c *recognize.Chart, s *lex.Stream, dead map[[2]int]bool, observed map[ZZShift]bool, st *ZZM4Stats) {
	if st.Examples == nil {
		st.Examples = map[string]int{}
	}
	type key struct{ from, j, at int32 }
	obs := map[key]map[int32]bool{}
	for x := range observed {
		k := key{x.From, x.J, x.At}
		if obs[k] == nil {
			obs[k] = map[int32]bool{}
		}
		obs[k][x.Top] = true
	}
	tops := p.ZZTops()
	semi := slices.IndexFunc(tab.Symbols, func(x grammar.TableSymbol) bool { return x.Terminal && x.Name == "PERLY_SEMICOLON" })
	brace := slices.IndexFunc(tab.Symbols, func(x grammar.TableSymbol) bool { return x.Terminal && x.Name == "PERLY_BRACE_CLOSE" })
	inFollow := func(lhs, t int) bool { _, ok := slices.BinarySearch(tab.Follow[lhs], t); return ok }
	inFirst := func(x, t int) bool { _, ok := slices.BinarySearch(tab.First[x], t); return ok }
	for k, tok := range s.Tokens {
		st.Tokens++
		full := tab.TerminalSets[tok.TerminalSet]
		var eff []int
		plug, plugAfter := false, false
		for _, x := range full {
			if ZZIsPlugin(tab, x) {
				plug = true
			}
			if dead[[2]int{k, x}] {
				continue
			}
			if ZZIsPlugin(tab, x) {
				plugAfter = true
			}
			eff = append(eff, x)
		}
		if plug {
			st.Plugin++
		}
		if plugAfter {
			st.PluginAfter++
		}
		la := slices.Clone(eff)
		slip := brace >= 0 && slices.Contains(eff, brace) && semi >= 0 && !slices.Contains(eff, semi)
		if slip {
			la = append(la, semi)
		}
		if len(la) < 2 || k >= len(c.Sets) || c.Sets[k] == nil {
			continue
		}
		if len(eff) >= 2 {
			st.Multi++
		}
		set := c.Sets[k]
		rd, rdSlip, sa, so := false, false, false, false
		for _, it := range set.Items() {
			o := int32(k - it.Dist)
			for _, ci := range tab.States[it.State].Items {
				item := tab.Items[ci]
				rhs := tab.Rules[item.Rule].RHS
				if item.Dot == len(rhs) {
					lhs := tab.Rules[item.Rule].LHS
					for a := range la {
						for b := a + 1; b < len(la); b++ {
							if inFollow(lhs, la[a]) && inFollow(lhs, la[b]) && zzReduceDiff(p, item.Rule, la[a], la[b]) {
								if la[a] == semi && slip || la[b] == semi && slip {
									if !rdSlip {
										st.Examples[fmt.Sprintf("reduce(slip ;) %s before %s vs %s", zzRule(tab, item.Rule), zzSym(tab, la[a]), zzSym(tab, la[b]))]++
									}
									rdSlip = true
								} else {
									if !rd {
										st.Examples[fmt.Sprintf("reduce %s before %s vs %s", zzRule(tab, item.Rule), zzSym(tab, la[a]), zzSym(tab, la[b]))]++
									}
									rd = true
								}
							}
						}
					}
					continue
				}
				x := rhs[item.Dot]
				for a := range eff {
					for b := a + 1; b < len(eff); b++ {
						if !inFirst(x, eff[a]) || !inFirst(x, eff[b]) {
							continue
						}
						for _, tp := range tops {
							if p.shiftForbidden(tp, int32(ci), int32(eff[a])) != p.shiftForbidden(tp, int32(ci), int32(eff[b])) {
								if !sa {
									st.Examples[fmt.Sprintf("shift(any top) after [%s] %s vs %s under %s", zzItem(tab, ci), zzSym(tab, eff[a]), zzSym(tab, eff[b]), zzTopText(tab, tp))]++
								}
								sa = true
								break
							}
						}
						for tp := range obs[key{o, int32(k), int32(ci)}] {
							if p.shiftForbidden(tp, int32(ci), int32(eff[a])) != p.shiftForbidden(tp, int32(ci), int32(eff[b])) {
								if !so {
									st.Examples[fmt.Sprintf("shift(observed top) after [%s] %s vs %s under %s", zzItem(tab, ci), zzSym(tab, eff[a]), zzSym(tab, eff[b]), zzTopText(tab, tp))]++
								}
								so = true
							}
						}
					}
				}
			}
		}
		if rd {
			st.ReduceDiff++
		}
		if rdSlip {
			st.ReduceDiffSlip++
		}
		if sa {
			st.ShiftDiffAny++
		}
		if so {
			st.ShiftDiffObserved++
		}
		if rd || sa {
			st.AnyDiff++
		}
	}
}

// zzReduceDiff: rule r's V1 reduce verdict before t1 and before t2 differ (value 2 rows compared over parent items).
func zzReduceDiff(p *Precedence, r, t1, t2 int) bool {
	v1, v2 := p.reduceAt[r*p.nt+t1], p.reduceAt[r*p.nt+t2]
	if v1 != v2 {
		return true
	}
	if v1 != 2 {
		return false
	}
	ni := len(p.topAt)
	a, b := int(p.mixAt[r*p.nt+t1])*ni, int(p.mixAt[r*p.nt+t2])*ni
	return !slices.Equal(p.parentForbid[a:a+ni], p.parentForbid[b:b+ni])
}

// ZZTop returns the top examples of m, most frequent first.
func ZZTop(m map[string]int, n int) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	sort.Slice(ks, func(i, j int) bool {
		if m[ks[i]] != m[ks[j]] {
			return m[ks[i]] > m[ks[j]]
		}
		return ks[i] < ks[j]
	})
	if len(ks) > n {
		ks = ks[:n]
	}
	out := make([]string, len(ks))
	for i, k := range ks {
		out[i] = fmt.Sprintf("%6d  %s", m[k], k)
	}
	return out
}

// ZZM3Static over-approximates, from the tables alone, each split state's core items' canonical spine tops: a
// rule begun at an item's origin has its tops fixed in the predicted state E it was predicted in, from the
// kernel items of every kernel state whose Epsilon is E (they merge into one (E, o) item) or from E's own items
// advanced over nullables; left-corner predictions inherit. Every state reachable from E by transitions holds
// items of rules begun in E, with those tops. A state reached from several E is the union (over-approximate:
// two E at one origin). It reports states where two core items carry different top sets and, for some shift
// (at, f) with f in FIRST of at's next symbol and at's dot > 0, a top of the other set allows a shift every top
// of at's own set forbids (over-keep), and states needing >= 2 verdict vectors with differing core items.
func ZZM3StaticTops(tab *grammar.Tables, p *Precedence) []map[int]map[int32]bool {
	canon := func(t int) int32 {
		if t >= 0 && p.topAt[t] < 0 {
			return -1
		}
		return int32(t)
	}
	byLHS := make([][]int, len(tab.Symbols))
	for r, rule := range tab.Rules {
		byLHS[rule.LHS] = append(byLHS[rule.LHS], r)
	}
	kernelsOf := map[int][]int{}
	for k, st := range tab.States {
		if !st.Predicted && st.Epsilon >= 0 {
			kernelsOf[st.Epsilon] = append(kernelsOf[st.Epsilon], k)
		}
	}
	tops := make([]map[int]map[int32]bool, len(tab.States)) // state -> rule -> tops
	for e, st := range tab.States {
		if !st.Predicted {
			continue
		}
		ruleTop := map[int]map[int32]bool{}
		addR := func(r int, t int32) bool {
			if ruleTop[r] == nil {
				ruleTop[r] = map[int32]bool{}
			}
			if ruleTop[r][t] {
				return false
			}
			ruleTop[r][t] = true
			return true
		}
		if e == 0 {
			for _, r := range byLHS[tab.Start] {
				addR(r, canon(int(p.base[r])))
			}
		}
		var outer []int // kernel items: sources for what they predict
		for _, k := range kernelsOf[e] {
			outer = append(outer, tab.States[k].Items...)
		}
		for changed := true; changed; {
			changed = false
			src := func(i int, t int32) {
				it := tab.Items[i]
				rhs := tab.Rules[it.Rule].RHS
				if it.Dot >= len(rhs) || tab.Symbols[rhs[it.Dot]].Terminal {
					return
				}
				for _, r := range byLHS[rhs[it.Dot]] {
					if addR(r, t) {
						changed = true
					}
				}
			}
			for _, i := range outer {
				src(i, canon(i)) // kernel items have dot > 0
			}
			for _, i := range st.Items {
				it := tab.Items[i]
				if it.Dot > 0 {
					src(i, canon(i))
				} else {
					for t := range ruleTop[it.Rule] {
						src(i, t)
					}
				}
			}
		}
		if ZZPerE != nil {
			ZZPerE(e, ruleTop)
		}
		// propagate over transitions from e (origin kept)
		seen := map[int]bool{}
		work := []int{e}
		for len(work) > 0 {
			s := work[len(work)-1]
			work = work[:len(work)-1]
			if seen[s] {
				continue
			}
			seen[s] = true
			if tops[s] == nil {
				tops[s] = map[int]map[int32]bool{}
			}
			for _, i := range tab.States[s].Items {
				r := tab.Items[i].Rule
				for t := range ruleTop[r] {
					if tops[s][r] == nil {
						tops[s][r] = map[int32]bool{}
					}
					tops[s][r][t] = true
				}
			}
			for _, tr := range tab.States[s].Transitions {
				if !tab.States[tr.Target].Predicted {
					work = append(work, tr.Target)
				}
			}
		}
	}
	return tops
}

// ZZM3Static judges every state with ZZJudgeItem over its static (at, top, f) shifts.
func ZZM3Static(tab *grammar.Tables, p *Precedence, st *ZZM3Stats) (pairs, multiRule int) {
	tops := ZZM3StaticTops(tab, p)
	for s, state := range tab.States {
		var xs []ZZShift
		for _, i := range state.Items {
			it := tab.Items[i]
			rhs := tab.Rules[it.Rule].RHS
			rt := tops[s][it.Rule]
			pairs += len(rt)
			if len(rt) >= 2 {
				multiRule++
			}
			if it.Dot == 0 || it.Dot >= len(rhs) {
				continue
			}
			x := rhs[it.Dot]
			fs := tab.First[x]
			if tab.Symbols[x].Terminal {
				fs = []int{x}
			}
			for t := range rt {
				for _, f := range fs {
					xs = append(xs, ZZShift{0, 0, int32(i), t, int32(f)})
				}
			}
		}
		if len(xs) > 0 {
			ZZJudgeItem(tab, p, int32(s), xs, st)
		}
	}
	return
}

// ZZRuleTops is the canonical spine tops of the rules predicted in predicted state e from the kernel states ks
// (none for the start state, whose start rules hang from their own start item).
func ZZRuleTops(tab *grammar.Tables, p *Precedence, e int, ks []int) map[int]map[int32]bool {
	canon := func(t int) int32 {
		if t >= 0 && p.topAt[t] < 0 {
			return -1
		}
		return int32(t)
	}
	byLHS := make([][]int, len(tab.Symbols))
	for r, rule := range tab.Rules {
		byLHS[rule.LHS] = append(byLHS[rule.LHS], r)
	}
	ruleTop := map[int]map[int32]bool{}
	addR := func(r int, t int32) bool {
		if ruleTop[r] == nil {
			ruleTop[r] = map[int32]bool{}
		}
		if ruleTop[r][t] {
			return false
		}
		ruleTop[r][t] = true
		return true
	}
	if e == 0 {
		for _, r := range byLHS[tab.Start] {
			addR(r, canon(int(p.base[r])))
		}
	}
	var outer []int
	for _, k := range ks {
		outer = append(outer, tab.States[k].Items...)
	}
	for changed := true; changed; {
		changed = false
		src := func(i int, t int32) {
			it := tab.Items[i]
			rhs := tab.Rules[it.Rule].RHS
			if it.Dot >= len(rhs) || tab.Symbols[rhs[it.Dot]].Terminal {
				return
			}
			for _, r := range byLHS[rhs[it.Dot]] {
				if addR(r, t) {
					changed = true
				}
			}
		}
		for _, i := range outer {
			src(i, canon(i))
		}
		for _, i := range tab.States[e].Items {
			if it := tab.Items[i]; it.Dot > 0 {
				src(i, canon(i))
			} else {
				for t := range ruleTop[it.Rule] {
					src(i, t)
				}
			}
		}
	}
	return ruleTop
}

// ZZDescShifts is every shift (at, f) of a rule begun in e, over every state reachable from e (origin kept).
func ZZDescShifts(tab *grammar.Tables, e int) map[[2]int32]bool {
	out := map[[2]int32]bool{}
	seen := map[int]bool{}
	work := []int{e}
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
			if it.Dot == 0 || it.Dot >= len(rhs) {
				continue
			}
			x := rhs[it.Dot]
			fs := tab.First[x]
			if tab.Symbols[x].Terminal {
				fs = []int{x}
			}
			for _, f := range fs {
				out[[2]int32{int32(i), int32(f)}] = true
			}
		}
		for _, tr := range tab.States[s].Transitions {
			if !tab.States[tr.Target].Predicted {
				work = append(work, tr.Target)
			}
		}
	}
	return out
}
