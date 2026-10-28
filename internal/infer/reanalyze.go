// ABOUTME: Reanalyze: an edited file's analysis that walks only its damaged top-level subs and the code outside every sub.
// ABOUTME: Each walk records what a top-level sub left behind -- scopes, diagnostics, outer writes -- for the next edit to replay.

package infer

import (
	"maps"
	"slices"
	"strings"

	"tamarou.com/pvm/internal/lower"
	"tamarou.com/pvm/internal/parse"
	"tamarou.com/pvm/internal/types"
)

// subTrace is what walking one top-level sub -- a named sub not inside
// another, as lower.AssignIDs anchors it -- did to the analysis, besides
// the annotations of its own nodes, which its ids find.
type subTrace struct {
	decl *lower.Decl
	// pass1 is the scopes pass 1 made for the sub, as pass 2 left them,
	// and pkg the package name a `package` inside it left the file at, or
	// "" if it held none that changed it.
	pass1 []*Scope
	pkg   string
	// pass2 is the scopes pass 2 made for the sub around its body, and
	// diags the diagnostics its walk raised, in order.
	pass2 []*Scope
	diags []Diagnostic
	// outer is each name in a scope around the sub that the walk changed,
	// and what it left there: the widened writes, the sub's return and
	// parameter types.
	outer []outerWrite
	// ok is whether the trace replays as the walk it records: false when
	// the walk ran where something other than the symbol table carries
	// state past it (a loop, an unreached scope, an `if` modifier), for a
	// method, or when it added a name around the sub or moved one's span.
	ok bool
}

// outerWrite is one name a sub's walk left changed in a scope around it,
// the scope named by its depth on the chain.
type outerWrite struct {
	depth   int
	name    string
	sym     Symbol
	deleted bool
}

// recorder records a trace of every top-level sub an analysis walks, and,
// for an edited file, replays the trace prev holds of every sub the edit
// did not damage instead of walking it again.
type recorder struct {
	traces  map[string]*subTrace
	prev    map[string]*subTrace
	damaged map[string]bool
	edit    parse.Edit
	// walked is the top-level subs walked rather than replayed, in walk
	// order; failed is set when a replay found the file not as its trace.
	walked []string
	failed bool
}

// topSub reports whether d is a top-level sub's declaration -- the node
// its anchor opens at -- and returns the anchor.
func topSub(d *lower.Decl) (string, bool) {
	id := d.ID
	if id.Local != 0 || id.Role != 0 || id.Anchor == "" || strings.Contains(id.Anchor, "/") {
		return "", false
	}
	return id.Anchor, true
}

// replays reports whether the sub anchor names is replayed, not walked.
func (r *recorder) replays(anchor string) bool {
	return r.prev[anchor] != nil && !r.damaged[anchor]
}

// trace is anchor's trace in the analysis being made.
func (r *recorder) trace(d *lower.Decl, anchor string) *subTrace {
	t := r.traces[anchor]
	if t == nil {
		t = &subTrace{decl: d, ok: d.Kind == lower.DeclSub}
		r.traces[anchor] = t
	}
	return t
}

// collect runs pass 1 over the top-level sub d, through live, recording
// the scopes it made.
func (r *recorder) collect(st *SymbolTable, d *lower.Decl, anchor string, live func()) {
	t := r.trace(d, anchor)
	cur, n0, pkg := st.current, len(st.current.children), st.root.Name
	live()
	t.pass1 = slices.Clone(cur.children[n0:])
	if st.root.Name != pkg {
		t.pkg = st.root.Name
	}
}

// graft is pass 1 over the top-level sub d replayed from its trace: the
// sub's own symbol defined as collectDecl defines it, and copies of the
// scopes the trace holds where collectDecl would have made them.
func (r *recorder) graft(st *SymbolTable, pos *lower.Positions, d *lower.Decl, anchor string) {
	p, t := r.prev[anchor], r.trace(d, anchor)
	cur, n0 := st.current, len(st.current.children)
	collectDecl(st, pos, d, func(lower.Node) {})
	cur.children = cur.children[:n0]
	for _, s := range p.pass1 {
		cur.children = append(cur.children, copyScope(s, cur, r.edit))
	}
	t.pass1 = slices.Clone(cur.children[n0:])
	if t.pkg = p.pkg; p.pkg != "" {
		st.root.Name = p.pkg
	}
}

// topSub walks the top-level sub d, recording what the walk does, or
// replays it from prev's trace when the edit did not damage it.
func (a *analysis) topSub(d *lower.Decl, anchor string) types.Type {
	r := a.rec
	t := r.trace(d, anchor)
	if r.replays(anchor) {
		a.replay(t, r.prev[anchor])
		return types.Code
	}
	if r.prev != nil {
		r.walked = append(r.walked, anchor)
	}
	if len(a.loops) > 0 || len(a.loopExits) > 0 || a.st.unreached > 0 || a.modifier != nil {
		t.ok = false
	}
	cur, n0, d0 := a.st.current, len(a.st.current.children), len(a.diags)
	a.st.PushFrame()
	typ := a.walkNode(d)
	for _, touched := range []map[scopedName]Symbol{a.st.Touched(), a.st.TouchedElems()} {
		for n, before := range touched {
			if !n.scope.open {
				continue // the sub's own scope
			}
			after, held := n.scope.symbols[n.name]
			if before.Name == "" || held && (after.StartByte != before.StartByte || after.EndByte != before.EndByte) {
				t.ok = false
			}
			t.outer = append(t.outer, outerWrite{n.scope.depth, n.name, after, !held})
		}
	}
	a.st.PopFrame()
	t.pass2 = slices.Clone(cur.children[n0:])
	t.diags = slices.Clone(a.diags[d0:])
	return typ
}

// replay does to the analysis what walking the sub p traces did, its
// spans moved by the edit, and records it as t.
func (a *analysis) replay(t, p *subTrace) {
	e := a.rec.edit
	t.ok, t.outer = p.ok, p.outer
	cur := a.st.current
	for _, s := range p.pass2 {
		c := copyScope(s, cur, e)
		cur.children = append(cur.children, c)
		t.pass2 = append(t.pass2, c)
	}
	t.diags = shiftDiags(p.diags, e)
	a.diags = append(a.diags, t.diags...)
	for _, w := range p.outer {
		s := a.st.current
		for s != nil && s.depth > w.depth {
			s = s.parent
		}
		now, held := Symbol{}, false
		if s != nil {
			now, held = s.symbols[w.name]
		}
		if s == nil || s.depth != w.depth || !held {
			a.rec.failed = true
			return
		}
		n := scopedName{s, w.name}
		if w.deleted {
			a.st.remove(n)
			continue
		}
		sym := w.sym
		sym.StartByte, sym.EndByte = now.StartByte, now.EndByte
		a.st.set(n, sym)
	}
}

// copyScope is a copy of s and every scope under it, closed, under parent,
// each symbol's span moved by e.
func copyScope(s, parent *Scope, e parse.Edit) *Scope {
	c := &Scope{Name: s.Name, parent: parent, symbols: make(map[string]Symbol, len(s.symbols)), depth: s.depth}
	for k, sym := range s.symbols {
		sym.StartByte, sym.EndByte = shiftSpan(e, sym.StartByte, sym.EndByte)
		c.symbols[k] = sym
	}
	for _, ch := range s.children {
		c.children = append(c.children, copyScope(ch, c, e))
	}
	return c
}

// shiftSpan is where the bytes [start, end) of the source before e are in
// the source after it, for a span that is not inside the bytes e replaced:
// a start at or past the replaced bytes moves, and an end past their start,
// which ends after them or around them, moves.
func shiftSpan(e parse.Edit, start, end uint32) (uint32, uint32) {
	delta := len(e.NewText) - (e.OldEnd - e.Start)
	if int(start) >= e.OldEnd {
		start = uint32(int(start) + delta)
	}
	if int(end) > e.Start {
		end = uint32(int(end) + delta)
	}
	return start, end
}

// Reanalyze is AnalyzeLowered(f, pos, idx, opts) for f, a file
// lower.Relower made from prev.File by re-lowering the top-level subs
// damaged names, after e, the edit between their sources. Every other
// top-level sub is not walked: what its walk in prev did is replayed, its
// spans moved by e. That is sound only for subs the edit cannot have
// changed the analysis of, as psc's damage decides. The code outside every
// sub is walked, since it threads state into the subs.
//
// It reports false when prev cannot be replayed -- it holds no trace of a
// sub, or one that does not replay -- and the caller analyzes afresh.
//
// prev is not changed. Its Diagnostics are not read: every diagnostic is
// the walk's own or a trace's, as inference raised it.
func Reanalyze(prev *FileAnalysis, f *lower.File, pos *lower.Positions, damaged []string, e parse.Edit, idx *ProjectIndex, opts Options) (*FileAnalysis, bool) {
	if prev == nil || prev.traces == nil {
		return nil, false
	}
	r := &recorder{traces: map[string]*subTrace{}, prev: prev.traces, damaged: map[string]bool{}, edit: e, walked: []string{}}
	for _, s := range damaged {
		r.damaged[s] = true
	}
	for anchor, t := range prev.traces {
		if !r.damaged[anchor] && !t.ok {
			return nil, false
		}
	}
	a := newAnalysisFrom(f, pos, idx, opts, r, prev.assigned)
	a.walk(f)
	if r.failed || len(r.traces) != len(prev.traces) {
		return nil, false
	}
	// The annotations are prev's, less every damaged sub's old nodes', plus
	// what this walk typed. The walk starts from none, as a fresh one does,
	// since a rule may read a node's earlier type.
	ann := maps.Clone(prev.Annotations)
	for s := range r.damaged {
		if t := prev.traces[s]; t != nil {
			forget(ann, t.decl)
		}
	}
	maps.Copy(ann, a.annotations)
	a.annotations = ann
	out := a.result(f, pos)
	out.Walked = r.walked
	return out, true
}

// Shifted is prev moved by e, an edit that changed nothing the analysis
// reads, such as whitespace between two subs: nothing is walked, and every
// span prev holds -- a diagnostic's, a symbol's, a trace's -- moves by e.
// pos is prev.File's position map over the edited tree.
//
// Its Diagnostics are prev's moved, whatever a caller made of them.
func Shifted(prev *FileAnalysis, pos *lower.Positions, e parse.Edit) *FileAnalysis {
	copies := map[*Scope]*Scope{}
	var cp func(s, parent *Scope) *Scope
	cp = func(s, parent *Scope) *Scope {
		c := copyScope(s, parent, e)
		var pair func(s, c *Scope)
		pair = func(s, c *Scope) {
			copies[s] = c
			for i := range s.children {
				pair(s.children[i], c.children[i])
			}
		}
		pair(s, c)
		return c
	}
	root := cp(prev.Symbols.root, nil)
	root.open = true
	st := &SymbolTable{root: root, current: root, blockScopes: map[lower.ID]*Scope{}, holders: map[string][]*Scope{}}
	for name := range root.symbols {
		st.holders[name] = []*Scope{root}
	}
	traces := make(map[string]*subTrace, len(prev.traces))
	for anchor, p := range prev.traces {
		t := *p
		t.pass1, t.pass2, t.diags = nil, nil, shiftDiags(p.diags, e)
		for _, s := range p.pass1 {
			t.pass1 = append(t.pass1, copies[s])
		}
		for _, s := range p.pass2 {
			t.pass2 = append(t.pass2, copies[s])
		}
		traces[anchor] = &t
	}
	return &FileAnalysis{
		File:        prev.File,
		Positions:   pos,
		Annotations: prev.Annotations,
		Diagnostics: shiftDiags(prev.Diagnostics, e),
		Symbols:     st,
		Package:     prev.Package,
		Walked:      []string{},
		traces:      traces,
		assigned:    prev.assigned,
	}
}

func shiftDiags(diags []Diagnostic, e parse.Edit) []Diagnostic {
	out := make([]Diagnostic, len(diags))
	for i, g := range diags {
		g.StartByte, g.EndByte = shiftSpan(e, g.StartByte, g.EndByte)
		out[i] = g
	}
	return out
}

// forget deletes the annotation of n and of every node under it.
func forget(ann map[lower.ID]types.Type, n lower.Node) {
	delete(ann, n.Meta().ID)
	for _, c := range lower.Children(n) {
		forget(ann, c)
	}
}
