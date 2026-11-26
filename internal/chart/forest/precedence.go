// ABOUTME: The precedence constraint (spec §8): bison's resolutions applied to a built forest, after recognition.
// ABOUTME: Constrain judges by state-free keys generated from bison's states; the test-only reference by parser state. Items are untouched.

package forest

import (
	"slices"

	"tamarou.com/pvm/internal/chart/prectab"
)

// Constrain applies p to f in place: it keeps each alternative some derivation of the file bison's
// resolutions would build passes through and drops the rest, true; or refuses the file, false, where the root
// keeps no derivation. It judges with no parser state (decision §3.3), by the keys of p's state-free tables: a
// reduction by its rule and the terminal after it, and the parent item for a pair that needs one; a shift by
// the spine-top item, the rule-position item before it and the terminal shifted. A span left with no
// derivation in its context fails every derivation holding it there. ConstrainReference is what it is
// compared against.
// shortcut: an alternative kept in one context is kept in all, so a node reached in two contexts that keep
// different alternatives keeps both; the verdict is exact, and only the derivations a later reader enumerates
// can include one bison builds in neither. Revisit if READINGS counts one.
func (f *Forest) Constrain(p *prectab.Precedence) bool {
	return newJudgement(f, p).constrain()
}

// constrain is Constrain on the judgement's forest, its memo as it stands.
func (c *judgement) constrain() bool {
	f := c.f
	if len(c.eval(f.Root, c.p.Tab.EOF, rootTop, -1)) == 0 {
		return false
	}
	c.mark(f.Root, c.p.Tab.EOF, rootTop, -1, -1)
	for n := range f.Nodes {
		a := &f.Nodes[n].Alt
		for *a != 0 {
			if c.kept[*a] {
				a = &f.Alts[*a].Next
			} else {
				*a = f.Alts[*a].Next
			}
		}
	}
	return true
}

// ConstrainReference is used only by tests. It is the judgement Constrain made before the state-free evaluator,
// kept under its own name as the fixed point the differential checks compare production Constrain against, so
// production can change without changing what it is judged by; TestPrecedenceReferenceAgreesWithBison holds its
// agreement with bison. It applies p to f in place as Constrain does, verdict, unlinking and shortcut alike, but
// judges a derivation in its edge context: the parser state the derivation holding it reaches at its span's start
// (state 0 at the file's start), and the terminal after it. Each reduction is checked against the terminal after
// it, each shift against the reductions the state before it holds. It must not call Constrain, nor Constrain it.
// Its known defect: shifts forbids a terminal the state does not shift wherever a reduction there with it in its
// lookahead resolves to reduce, where bison has no conflict and reduces. Observed on 423da4dd7: perly.y's
// generation finds 0 shift keys where it applies, and no in-tree or perl.git file diverges;
// TestPrecedenceReferenceKnownDivergence pins the toy where Constrain and bison accept what it refuses.
func (f *Forest) ConstrainReference(p *prectab.Precedence) bool {
	c := &constrainer{f: f, p: p, firsts: map[ctx][]int32{}, marked: map[ctx]bool{}, kept: make([]bool, len(f.Alts))}
	if len(c.eval(f.Root, 0, p.Tab.EOF)) == 0 {
		return false
	}
	c.mark(f.Root, 0, p.Tab.EOF, -1)
	for n := range f.Nodes {
		a := &f.Nodes[n].Alt
		for *a != 0 {
			if c.kept[*a] {
				a = &f.Alts[*a].Next
			} else {
				*a = f.Alts[*a].Next
			}
		}
	}
	return true
}

// constrainer is one application of a Precedence to a forest.
type constrainer struct {
	f      *Forest
	p      *prectab.Precedence
	firsts map[ctx][]int32 // eval's results
	marked map[ctx]bool    // mark's visits, first the terminal its derivations must begin with
	kept   []bool          // by alternative
}

// ctx is a node in an edge context: the parser state at its start and the terminal after it, and for mark the
// terminal its derivations must begin with (-1 any).
type ctx struct{ n, state, t, first int32 }

// rule is the rule of a symbol node's or rule position's alternative a, and how many of its symbols a covers.
func (c *constrainer) rule(n, a int32) (r, m int) {
	r = int(c.f.Alts[a].Rule)
	if nd := c.f.Nodes[n]; nd.Sym < 0 {
		return r, int(nd.Dot)
	}
	return r, len(c.p.Tab.Rules[r].RHS)
}

// eval is the first terminals of n's derivations that bison's resolutions pass in context (s, t), sorted;
// none where every one is forbidden. A span with no tokens begins with t.
func (c *constrainer) eval(n, s int32, t int) []int32 {
	nd := c.f.Nodes[n]
	if nd.Alt == 0 {
		return []int32{nd.Sym} // a leaf
	}
	k := ctx{n, s, int32(t), -1}
	if out, ok := c.firsts[k]; ok {
		return out
	}
	var out []int32
	for a := nd.Alt; a != 0; a = c.f.Alts[a].Next {
		r, m := c.rule(n, a)
		if nd.Sym >= 0 && !c.p.Reduces(r, s, t) {
			continue
		}
		out = append(out, c.seq(a, r, m, s, t, nil)...)
	}
	slices.Sort(out)
	out = slices.Compact(out)
	c.firsts[k] = out
	return out
}

// seq is the first terminals of alternative a's derivations, rule r's first m symbols begun in state s with t
// next, that bison's resolutions pass. With keep it also marks, through keep, the children of each derivation
// whose first terminal it accepts.
func (c *constrainer) seq(a int32, r, m int, s int32, t int, keep func(first int32) bool) []int32 {
	if m == 0 {
		return []int32{int32(t)}
	}
	alt := c.f.Alts[a]
	before := c.p.Walk(s, c.p.Tab.Rules[r].RHS[:m-1]) // the state the right child begins in
	right := c.eval(alt.Right, before, t)
	if m == 1 {
		for _, f := range right {
			if keep != nil && keep(f) {
				c.mark(alt.Right, before, t, f)
			}
		}
		return right
	}
	var out []int32
	empty := c.f.Nodes[alt.Right].From == c.f.Nodes[alt.Right].To && c.f.Nodes[alt.Right].Alt != 0
	for _, f := range right {
		// A shift of f where the right child begins, unless the right child is empty and reduces before f is read.
		if !empty && !c.p.Shifts(before, int(f)) {
			continue
		}
		left := c.eval(alt.Left, s, int(f))
		out = append(out, left...)
		if keep == nil {
			continue
		}
		for _, l := range left {
			if keep(l) {
				c.mark(alt.Right, before, t, f)
				c.mark(alt.Left, s, int(f), l)
			}
		}
	}
	return out
}

// mark keeps the alternatives of n's derivations that bison's resolutions pass in context (s, t) and that
// begin with first (-1 any), and those of their children.
func (c *constrainer) mark(n, s int32, t int, first int32) {
	nd := c.f.Nodes[n]
	k := ctx{n, s, int32(t), first}
	if nd.Alt == 0 || c.marked[k] {
		return
	}
	c.marked[k] = true
	keep := func(f int32) bool { return first < 0 || f == first }
	for a := nd.Alt; a != 0; a = c.f.Alts[a].Next {
		r, m := c.rule(n, a)
		if nd.Sym >= 0 && !c.p.Reduces(r, s, t) {
			continue
		}
		if slices.ContainsFunc(c.seq(a, r, m, s, t, keep), keep) {
			c.kept[a] = true
		}
	}
}

// rootTop stands for the root's spine top: the start item of each root alternative's own rule.
const rootTop = -2

// judgement is one state-free application of a Precedence to a forest. Everything it writes lives here, on the
// call, never on the shared Precedence. Its memo is per node and uses no hashing: evalAt[n] and markAt[n] head
// node n's lists in evals and marks (0 none, so entry 0 is unused), linked by next.
type judgement struct {
	f              *Forest
	p              *prectab.Precedence
	evalAt, markAt []int32
	evals          []evalEntry
	marks          []markEntry
	kept           []bool // by alternative
}

// evalEntry is eval's result in one context: the terminal after the node, its spine top and its parent item.
type evalEntry struct {
	t, top, pi, next int32
	out              []int32 // shared with every parent that reads it; never appended into
}

// markEntry is one mark visit: eval's context and the terminal its derivations must begin with (-1 any).
type markEntry struct{ t, top, pi, first, next int32 }

func newJudgement(f *Forest, p *prectab.Precedence) *judgement {
	return &judgement{f: f, p: p, evalAt: make([]int32, len(f.Nodes)), markAt: make([]int32, len(f.Nodes)),
		evals: make([]evalEntry, 1, len(f.Nodes)), marks: make([]markEntry, 1, len(f.Nodes)), kept: make([]bool, len(f.Alts))}
}

// canon drops the context node n cannot use, so two encodings of one context share a memo entry: a spine top
// that forbids no shift, and a parent item where n's symbol has no pair that needs one.
func (c *judgement) canon(n, top, pi int32) (int32, int32) {
	if top >= 0 && c.p.TopAt[top] < 0 {
		top = -1
	}
	if sym := c.f.Nodes[n].Sym; sym < 0 || !c.p.NeedParent[sym] {
		pi = -1
	}
	return top, pi
}

// alt is the rule of node n's alternative a, how many of its symbols a covers, and the spine top its own
// symbols hang from: top, or at the root the start item of a's rule.
func (c *judgement) alt(n, a, top int32) (r, m int, at int32) {
	r = int(c.f.Alts[a].Rule)
	m = len(c.p.Tab.Rules[r].RHS)
	if nd := c.f.Nodes[n]; nd.Sym < 0 {
		m = int(nd.Dot)
	}
	if top == rootTop {
		top = c.p.Base[r]
	}
	return r, m, top
}

// eval is the first terminals of n's derivations that bison's resolutions pass with t next, under the spine top
// top and the parent item pi, sorted; none where every one is forbidden. A span with no tokens begins with t.
func (c *judgement) eval(n int32, t int, top, pi int32) []int32 {
	nd := c.f.Nodes[n]
	if nd.Alt == 0 {
		return []int32{nd.Sym} // a leaf
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
		if nd.Sym >= 0 && !c.p.ReducesFree(r, t, pi) {
			continue
		}
		out = append(out, c.seq(a, r, m, t, at, nil)...)
	}
	slices.Sort(out)
	out = slices.Compact(out)
	c.evals = append(c.evals, evalEntry{int32(t), top, pi, c.evalAt[n], out})
	c.evalAt[n] = int32(len(c.evals) - 1)
	return out
}

// seq is the first terminals of alternative a's derivations, rule r's first m symbols under the spine top top
// with t next, that bison's resolutions pass. The right child hangs from the rule-position item before it,
// unless it is r's first symbol and inherits top; the left inherits top. With keep it also marks, through keep,
// the children of each derivation whose first terminal it accepts.
func (c *judgement) seq(a int32, r, m int, t int, top int32, keep func(first int32) bool) []int32 {
	if m == 0 {
		return []int32{int32(t)}
	}
	alt := c.f.Alts[a]
	at := c.p.Base[r] + int32(m-1) // the rule-position item before the right child
	rtop := top
	if m > 1 {
		rtop = at
	}
	right := c.eval(alt.Right, t, rtop, at)
	if m == 1 {
		for _, f := range right {
			if keep != nil && keep(f) {
				c.mark(alt.Right, t, rtop, at, f)
			}
		}
		return right
	}
	lpi := int32(-1) // the left child's parent item, where it is r's first symbol
	if m == 2 {
		lpi = c.p.Base[r]
	}
	var out []int32
	empty := c.f.Nodes[alt.Right].From == c.f.Nodes[alt.Right].To && c.f.Nodes[alt.Right].Alt != 0
	for _, f := range right {
		// A shift of f where the right child begins, unless the right child is empty and reduces before f is read.
		if !empty && c.p.ShiftForbidden(top, at, f) {
			continue
		}
		left := c.eval(alt.Left, int(f), top, lpi)
		out = append(out, left...)
		if keep == nil {
			continue
		}
		for _, l := range left {
			if keep(l) {
				c.mark(alt.Right, t, rtop, at, f)
				c.mark(alt.Left, int(f), top, lpi, l)
			}
		}
	}
	return out
}

// mark keeps the alternatives of n's derivations that bison's resolutions pass in eval's context and that begin
// with first (-1 any), and those of their children.
func (c *judgement) mark(n int32, t int, top, pi, first int32) {
	if c.f.Nodes[n].Alt == 0 {
		return
	}
	top, pi = c.canon(n, top, pi)
	for e := c.markAt[n]; e != 0; e = c.marks[e].next {
		if x := c.marks[e]; x.t == int32(t) && x.top == top && x.pi == pi && x.first == first {
			return
		}
	}
	c.marks = append(c.marks, markEntry{int32(t), top, pi, first, c.markAt[n]})
	c.markAt[n] = int32(len(c.marks) - 1)
	keep := func(f int32) bool { return first < 0 || f == first }
	nd := c.f.Nodes[n]
	for a := nd.Alt; a != 0; a = c.f.Alts[a].Next {
		r, m, at := c.alt(n, a, top)
		if nd.Sym >= 0 && !c.p.ReducesFree(r, t, pi) {
			continue
		}
		if slices.ContainsFunc(c.seq(a, r, m, t, at, keep), keep) {
			c.kept[a] = true
		}
	}
}
