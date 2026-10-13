// ABOUTME: AnalyzeLowered: PSC type inference over the lowered tree, keyed by node id rather than byte offset.
// ABOUTME: It runs alongside the legacy tree-sitter Analyze; the ports move one rule slice at a time onto it.

package infer

import (
	"tamarou.com/pvm/internal/lower"
	"tamarou.com/pvm/internal/types"
)

// AnalyzeLowered runs inference over f, which lower.Lower returned, and
// returns each node's type keyed by its id, the diagnostics, and the symbol
// table. pos is f's position map: diagnostics and symbols carry byte spans,
// and the lowered tree has none of its own. idx is the project index for
// cross-file analysis, nil in single-file mode.
//
// It is the legacy Analyze's two passes, rewritten against the lowered
// kinds: collectLoweredDeclarations fills the symbol table, then walk types
// every node bottom-up. Each rule slice is ported on its own, and proven by
// TestLoweredGoldenRatchet against what the legacy path reported. A kind no
// port handles yet is typed Any, raises no diagnostic, and has its children
// walked.
func AnalyzeLowered(f *lower.File, pos *lower.Positions, idx *ProjectIndex) (map[lower.ID]types.Type, []Diagnostic, *SymbolTable) {
	a := &loweredAnalysis{
		pos:         pos,
		idx:         idx,
		st:          collectLoweredDeclarations(f, pos),
		annotations: map[lower.ID]types.Type{},
		subs:        fileSubs(f),
	}
	a.walk(f)
	return a.annotations, a.diags, a.st
}

// loweredAnalysis is pass 2's state: what the legacy walkers thread through
// every call as arguments.
type loweredAnalysis struct {
	pos         *lower.Positions
	idx         *ProjectIndex
	st          *SymbolTable
	annotations map[lower.ID]types.Type
	diags       []Diagnostic
	// subs is each file-level named sub's body, which a user-defined
	// guard reads.
	subs map[string]*lower.AnonSub
}

// walk types n, children first, and records the type under n's id. Unknown
// is not recorded: as on the legacy path, an absent annotation is how
// "no type" reads.
func (a *loweredAnalysis) walk(n lower.Node) types.Type {
	var t types.Type
	switch n := n.(type) {
	case *lower.Hole:
		// A Hole is typed Any and never inspected (§4.14.2, Hole row).
		t = types.Any
	case *lower.NumLit:
		t = types.Num
		if n.IsInt {
			t = types.Int
		}
	case *lower.StrLit:
		// A command's output is not typed: the legacy path has no rule
		// for it, and qx// gives undef when the command cannot run.
		t = types.Str
		if n.Backtick {
			t = types.Unknown
		}
	case *lower.QwList, *lower.Bareword:
		// Neither is typed on the legacy path: qw's value depends on
		// context, and a bareword is a name, not a value.
		t = types.Unknown
	case *lower.List:
		a.walkChildren(n)
		t = types.Unknown
	case *lower.AnonSub:
		a.walkChildren(n)
		// A CodeRef, not a Code: Code is the compiled CV itself, which
		// no perl scalar ever holds.
		t = types.CodeRef
	case *lower.Decl:
		t = a.decl(n)
	case *lower.IfStmt:
		a.ifStmt(n)
		t = types.Unknown
	case *lower.WhileStmt:
		a.whileStmt(n)
		t = types.Unknown
	case *lower.ForeachStmt:
		a.foreachStmt(n)
		t = types.Unknown
	case *lower.UseStmt:
		a.useStmt(n)
		t = types.Unknown
	case *lower.LoopEx:
		// return, last, next, redo and goto are not values the legacy
		// path typed; a return's value is read where the sub is.
		a.walkChildren(n)
		t = types.Unknown
	default:
		a.walkChildren(n)
		t = types.Any
	}
	if t != types.Unknown {
		a.annotations[n.Meta().ID] = t
	}
	return t
}

// walkChildren walks each of n's children.
func (a *loweredAnalysis) walkChildren(n lower.Node) {
	for _, c := range lower.Children(n) {
		a.walk(c)
	}
}

// decl types a declaration and narrows the variable it declares to its
// initialiser's type.
//
// This is the declaration half of the legacy inferAssignmentNarrowing: the
// first declared variable takes the initialiser's type, unless that is
// Unknown. The rest of that rule -- an aggregate in scalar context is its
// count, the `() =` count-of idiom, a list-context match binding captures,
// element and class types -- reads the kinds the operator, variable and call
// ports type, and moves with them.
func (a *loweredAnalysis) decl(d *lower.Decl) types.Type {
	switch d.Kind {
	case lower.DeclMy, lower.DeclOur, lower.DeclState, lower.DeclLocal:
		for _, v := range d.Vars {
			a.walk(v)
		}
		if d.Init == nil {
			return types.Unknown
		}
		t := a.walk(d.Init)
		// local neither declares nor narrows: the legacy path reads no
		// name through it.
		if vars := declaredVars(d); len(vars) > 0 && t != types.Unknown && d.Kind != lower.DeclLocal {
			a.st.UpdateType(vars[0].Sigil+vars[0].Name, t)
		}
		return t
	case lower.DeclPackage:
		a.walkChildren(d)
		return types.Unknown
	case lower.DeclSub:
		// The body is walked in a fresh scope, as the legacy pass 2 does,
		// not the one collectLoweredDeclarations filled: a name declared
		// in a sub body is not found from here, and keeps its sigil type.
		// ponytail: that loses `my $x = 1` inside every sub; re-entering
		// pass 1's scope fixes it, and changes what PSC reports.
		for _, v := range d.Vars {
			a.walk(v)
		}
		if d.Init != nil {
			a.st.EnterScope(subName(d))
			a.walk(d.Init)
			ret := types.Unknown
			if body, ok := d.Init.(*lower.AnonSub); ok && body.Body != nil {
				b, _ := body.Body.(*lower.Block)
				ret = a.explicitReturns(body.Body) | a.implicitReturn(b)
			}
			a.st.ExitScope()
			if name := subName(d); name != "" && ret != types.Unknown {
				a.st.UpdateReturnType(name, ret)
			}
		}
		return types.Code
	}
	// field, class and method: no fixture has captured what the legacy
	// path does with them yet.
	a.walkChildren(d)
	return types.Any
}

// declaredVars returns the variables a my, our, state or local declares, in
// source order: the one Var, or the Vars of the parenthesised form. A Hole
// or an `undef` placeholder in the list declares nothing.
func declaredVars(d *lower.Decl) []*lower.Var {
	var out []*lower.Var
	for _, n := range d.Vars {
		switch n := n.(type) {
		case *lower.Var:
			out = append(out, n)
		case *lower.List:
			for _, e := range n.Elems {
				if v, ok := e.(*lower.Var); ok {
					out = append(out, v)
				}
			}
		}
	}
	return out
}

// subName is a named sub's name, without its & sigil, as the symbol table
// keys subs.
func subName(d *lower.Decl) string {
	if vars := declaredVars(d); len(vars) > 0 {
		return vars[0].Name
	}
	return ""
}

// sigilTypes is a declared variable's type before anything narrows it.
var sigilTypes = map[string]types.Type{"$": types.Scalar, "@": types.Array, "%": types.Hash}

// collectLoweredDeclarations is pass 1 on the lowered tree, the legacy
// CollectDeclarations: every my, our and state variable, named sub and
// package, defined in the scope the legacy path gives it.
//
// A Block is a scope, a named sub's body a scope of the sub's name, and a
// bare block's body is not one: the legacy path never saw it as a block.
func collectLoweredDeclarations(f *lower.File, pos *lower.Positions) *SymbolTable {
	st := NewSymbolTable()
	var walk func(n lower.Node)
	walkStmts := func(b lower.Node) {
		for _, c := range lower.Children(b) {
			walk(c)
		}
	}
	walk = func(n lower.Node) {
		switch n := n.(type) {
		case *lower.Block:
			st.EnterScope("block")
			walkStmts(n)
			st.ExitScope()
			return
		case *lower.BlockStmt:
			for _, c := range lower.Children(n) {
				if c == n.Body {
					walkStmts(c)
				} else {
					walk(c)
				}
			}
			return
		case *lower.ForeachStmt:
			// The loop variable is not declared here: pass 2 defines it
			// in the loop's own scope, as the legacy path does.
			for _, c := range lower.Children(n) {
				if c != n.Var {
					walk(c)
				}
			}
			return
		case *lower.Decl:
			if collectDecl(st, pos, n, walkStmts) {
				return
			}
		}
		for _, c := range lower.Children(n) {
			walk(c)
		}
	}
	walk(f)
	return st
}

// collectDecl defines what one declaration declares, and reports whether it
// walked the declaration's body itself.
func collectDecl(st *SymbolTable, pos *lower.Positions, d *lower.Decl, walkStmts func(lower.Node)) bool {
	start, end, _ := pos.Span(d.ID)
	span := func(s Symbol) Symbol {
		s.StartByte, s.EndByte = uint32(start), uint32(end)
		return s
	}
	switch d.Kind {
	case lower.DeclMy, lower.DeclOur, lower.DeclState:
		for _, v := range declaredVars(d) {
			if t, ok := sigilTypes[v.Sigil]; ok {
				st.Define(span(Symbol{Name: v.Sigil + v.Name, Type: t, Kind: SymVariable}))
			}
		}
	case lower.DeclPackage:
		if len(d.Vars) > 0 {
			if w, ok := d.Vars[0].(*lower.Bareword); ok {
				st.root.Name = w.Name
				st.Define(span(Symbol{Name: w.Name, Type: types.Unknown, Kind: SymPackage}))
			}
		}
	case lower.DeclSub:
		// `sub f;` declares without defining, and the legacy path records
		// no symbol for it.
		body, ok := d.Init.(*lower.AnonSub)
		if !ok {
			return true
		}
		name := subName(d)
		st.Define(span(Symbol{Name: name, Type: types.Code, Kind: SymSubroutine}))
		st.EnterScope(name)
		if body.Body != nil {
			walkStmts(body.Body)
		}
		st.ExitScope()
		return true
	}
	return false
}
