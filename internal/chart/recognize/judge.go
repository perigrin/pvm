// ABOUTME: The judged recognizer's reduce side (V2 design §4.1): a completion V1's precedence forbids is refused.
// ABOUTME: NewJudged's explicit and derived forms, the verdict per (state, LHS) and parent state, and the refusal records.

package recognize

import (
	"cmp"
	"math/bits"
	"slices"

	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/chart/prectab"
)

// judgement is what the reduce side reads, built from one Precedence over the recognizer's tables. Its rules and
// items are the automaton's, which keep the tables' ids (prectab.Item, TestPrectabItemMapping).
type judgement struct {
	p *prectab.Precedence
	// rule is, per state and parallel to grammarTables.lhs, a rule the state reduces for that LHS: Feasible's
	// rule-reduce counter is 0, so every such rule has its verdict before FOLLOW(LHS).
	rule [][]int32
	// need is each parent-dependent symbol's index, -1 for the rest; parent[state*nneed+need[sym]] is the automaton
	// item of a core item of state awaiting sym, -1 for none. Feasible's parent counter is 0, so any one stands for all.
	need   []int32
	parent []int32
	nneed  int
}

// newJudgement is the reduce side of p over tab, or Feasible's refusal of p, unwrapped.
func newJudgement(tab *grammar.Tables, p *prectab.Precedence) (*judgement, error) {
	if _, err := prectab.Feasible(p); err != nil {
		return nil, err
	}
	g := grammar.Derived(tab, buildGrammar)
	j := &judgement{p: p, rule: make([][]int32, len(tab.States)), need: slices.Repeat([]int32{-1}, len(tab.Symbols))}
	for s, st := range tab.States {
		j.rule[s] = make([]int32, len(g.lhs[s]))
		for i, lhs := range g.lhs[s] {
			k := slices.IndexFunc(st.Reduces, func(r int) bool { return tab.Rules[r].LHS == lhs })
			j.rule[s][i] = int32(st.Reduces[k])
		}
	}
	for sym, need := range p.NeedParent {
		if need {
			j.need[sym] = int32(j.nneed)
			j.nneed++
		}
	}
	j.parent = slices.Repeat([]int32{-1}, len(tab.States)*j.nneed)
	for s, st := range tab.States {
		for _, i := range st.Items {
			it := tab.Items[i]
			if rhs := tab.Rules[it.Rule].RHS; it.Dot < len(rhs) && j.need[rhs[it.Dot]] >= 0 {
				j.parent[s*j.nneed+int(j.need[rhs[it.Dot]])] = p.Item(i)
			}
		}
	}
	return j, nil
}

// derived is the derived form's judgement over one tables value (grammar.Derived): nil where the tables are not
// perly.y's committed grammar, rule for rule, and so declare no levels.
type derived struct {
	j   *judgement
	err error
}

// deriveJudgement decides by content, never by pointer: tables whose rules are perly.y's, each with the same LHS
// and RHS symbol names as grammar.Committed()'s G1, are judged with perly.y's precedence (prectab.ParsePrecedence,
// whose failure, Feasible's included, is the error); any other tables are judged with nothing.
func deriveJudgement(tab *grammar.Tables) derived {
	g, _, err := grammar.Committed()
	if err != nil {
		return derived{err: err}
	}
	if len(tab.Rules) != len(g.G1) {
		return derived{}
	}
	for r, rule := range g.G1 {
		tr := tab.Rules[r]
		if tab.Symbols[tr.LHS].Name != rule.LHS || len(tr.RHS) != len(rule.RHS) {
			return derived{}
		}
		for i, s := range rule.RHS {
			if tab.Symbols[tr.RHS[i]].Name != s.Text {
				return derived{}
			}
		}
	}
	p, err := prectab.ParsePrecedence(tab)
	if err != nil {
		return derived{err: err}
	}
	j, err := newJudgement(tab, p)
	return derived{j, err}
}

// NewNoPrecedence is a recognizer over tab that judges no precedence: the referees' unpruned recognizer and the
// refusal rerun's. recognize.New is this recognizer until the switch-over makes it the judged one.
func NewNoPrecedence(tab *grammar.Tables, opts Options) *Recognizer {
	g := grammar.Derived(tab, buildGrammar)
	return &Recognizer{tab: tab, opts: opts, g: g, cls: g.cls.Load(), ncls: g.narrowed.cls.Load(), sets: newInterner(), cache: map[cacheKey]*cacheNode{}, trie: map[cacheEdge]*cacheNode{}}
}

// NewJudged is a recognizer over tab that refuses the completions precedence forbids (the reduce side). With p it
// judges with p (explicit); with p nil it derives perly.y's precedence where tab's rules are perly.y's and judges
// nothing otherwise (deriveJudgement). An error, Feasible's refusal or a failed derivation, is returned here, and
// the recognizer returns it from Recognize for every input: it never recognizes unjudged.
func NewJudged(tab *grammar.Tables, p *prectab.Precedence, opts Options) (*Recognizer, error) {
	r := NewNoPrecedence(tab, opts)
	if p == nil {
		d := grammar.Derived(tab, deriveJudgement)
		r.judge, r.err = d.j, d.err
	} else {
		r.judge, r.err = newJudgement(tab, p)
	}
	return r, r.err
}

// verdict judges reducing rule r before the terminals of la that can follow its LHS (follow): refused where its
// reduce verdict forbids it before every one, or where there are none (vacuously), and kept where some allows it.
// Under the parent item pi, a parent-dependent terminal is judged by pi; with pi -1 it is left to the parent, and
// dep says one was.
func (j *judgement) verdict(r int, follow, la bitset, pi int32) (refused, dep bool) {
	p := j.p
	for w := range la {
		for x := la[w] & follow[w]; x != 0; x &= x - 1 {
			k := r*p.NT + w*64 + bits.TrailingZeros64(x)
			switch p.ReduceAt[k] {
			case 0:
				return false, false
			case 2:
				if pi < 0 {
					dep = true
				} else if !p.ParentForbid[int(p.MixAt[k])*len(p.TopAt)+int(pi)] {
					return false, false
				}
			}
		}
	}
	return !dep, dep
}

// parentItem is the automaton item of a core item of state awaiting sym, -1 where sym is not parent-dependent.
func (j *judgement) parentItem(state, sym int) int32 {
	if j.need[sym] < 0 {
		return -1
	}
	return j.parent[state*j.nneed+int(j.need[sym])]
}

// completes reports whether the judgement keeps completing the li-th LHS of state into parent, a state awaiting it.
func (j *judgement) completes(state, li int, follow, la bitset, lhs, parent int) bool {
	refused, _ := j.verdict(int(j.rule[state][li]), follow, la, j.parentItem(parent, lhs))
	return !refused
}

// refusal is a completion the judged recognizer refused: sym, completed from dist nodes back, into the items of the
// parent state there.
type refusal struct{ sym, dist, parent int32 }

// cmpRefusal orders refusals by symbol, distance and parent state.
func cmpRefusal(a, b refusal) int {
	return cmp.Or(cmp.Compare(a.sym, b.sym), cmp.Compare(a.dist, b.dist), cmp.Compare(a.parent, b.parent))
}

// refuse records that completing lhs from origin into the state parent there was refused.
func (b *builder) refuse(lhs, origin, parent int) {
	b.refused = append(b.refused, refusal{int32(lhs), int32(b.pos - origin), int32(parent)})
}

// refuseAll records the refusal of completing lhs from origin into every state of src, the set there, awaiting it.
func (b *builder) refuseAll(lhs, origin int, src *Set) {
	if src.core != nil {
		for _, en := range b.r.awaiting(src, lhs) {
			b.refuse(lhs, origin, src.items[en.idx].State)
		}
		return
	}
	for _, it := range src.items {
		if _, ok := b.r.goTo(it.State, lhs); ok {
			b.refuse(lhs, origin, it.State)
		}
	}
}

// settled is the node's refusals as records: relative, sorted, each once, less any another item of the set allowed,
// completing the same LHS from the same origin into the same parent state with another rule. Build pairs a child
// node with a parent, not a rule, so such a pair stands.
func (b *builder) settled() []refusal {
	j := b.r.judge
	out := b.refused[:0]
	for _, x := range b.refused {
		origin := b.pos - int(x.dist)
		if !slices.ContainsFunc(b.items, func(it [2]int) bool {
			if it[1] != origin || origin == b.pos && b.r.tab.States[it[0]].Predicted {
				return false
			}
			li := slices.Index(b.r.g.lhs[it[0]], int(x.sym))
			return li >= 0 && j.completes(it[0], li, b.r.g.follow[x.sym], b.la, int(x.sym), int(x.parent))
		}) {
			out = append(out, x)
		}
	}
	slices.SortFunc(out, cmpRefusal)
	return slices.Clone(slices.Compact(out))
}

// keepRefused records recs as the refusals at node pos, built there or served with the transition.
func (c *Chart) keepRefused(pos int, recs []refusal) {
	if len(recs) == 0 {
		return
	}
	if c.refused == nil {
		c.refused = make([][]refusal, len(c.Sets))
	}
	c.refused[pos] = recs
}

// RefusedAt reports whether the judged recognizer refused any completion at node to.
func (c *Chart) RefusedAt(to int) bool {
	return c.refused != nil && to >= 0 && to < len(c.refused) && len(c.refused[to]) > 0
}

// Refused reports whether the judged recognizer refused completing sym from node origin, at node to, into the
// items of the state parent at origin. Build reads it so that it never pairs a completion with such a parent.
func (c *Chart) Refused(to, sym, origin, parent int) bool {
	if c.refused == nil || to < 0 || to >= len(c.refused) {
		return false
	}
	_, ok := slices.BinarySearchFunc(c.refused[to], refusal{int32(sym), int32(to - origin), int32(parent)}, cmpRefusal)
	return ok
}

// leoCrosses reports whether the Leo chain from m, which completing sym into m's item begins, crosses a completion
// the judgement refuses before la: one of a link's rule into the next link's item. The first completion, sym's, is
// judged by the caller. Each memo keeps its verdict for the last la, so a walk stops at a link already judged before
// the same la, and a chain taken at every node of a right recursion is walked once rather than once per node.
func (b *builder) leoCrosses(m *memo) bool {
	if m.next == nil {
		return false
	}
	if m.crossLA != nil && slices.Equal(m.crossLA, b.la) {
		return m.crosses
	}
	b.r.leoLinks++
	li := slices.Index(b.r.g.lhs[m.adv], m.lhs)
	crosses := !b.r.judge.completes(m.adv, li, b.r.g.follow[m.lhs], b.la, m.lhs, m.next.item) || b.leoCrosses(m.next)
	m.crossLA, m.crosses = b.la, crosses // b.la is never written once a node's lookahead is made
	return crosses
}
