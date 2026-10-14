// ABOUTME: AnalyzeLowered: PSC type inference over the lowered tree, keyed by node id rather than byte offset.
// ABOUTME: It runs alongside the legacy tree-sitter Analyze; the ports move one rule slice at a time onto it.

package infer

import (
	"tamarou.com/pvm/internal/lower"
	"tamarou.com/pvm/internal/parse"
	"tamarou.com/pvm/internal/types"
)

// AnalyzeLowered runs inference over f, which lower.Lower returned, and
// returns each node's type keyed by its id, the diagnostics, and the symbol
// table. pos is f's position map: diagnostics and symbols carry byte spans,
// and the lowered tree has none of its own. idx is the project index for
// cross-file analysis, nil in single-file mode. opts is how an operand of no
// known type is checked, as AnalyzeWithOptions takes it.
//
// It is the legacy Analyze's two passes, rewritten against the lowered
// kinds: collectLoweredDeclarations fills the symbol table, then walk types
// every node bottom-up. Each rule slice is ported on its own, and proven by
// TestLoweredGoldenRatchet against what the legacy path reported. A kind no
// port handles yet is typed Any, raises no diagnostic, and has its children
// walked.
func AnalyzeLowered(f *lower.File, pos *lower.Positions, idx *ProjectIndex, opts Options) (map[lower.ID]types.Type, []Diagnostic, *SymbolTable) {
	a := &loweredAnalysis{
		pos:         pos,
		idx:         idx,
		opts:        opts,
		st:          collectLoweredDeclarations(f, pos),
		annotations: map[lower.ID]types.Type{},
		diags:       []Diagnostic{},
		subs:        fileSubs(f),
		assigned:    assignedNames(f),
		evalCtx:     map[lower.ID]types.Context{},
		classTypes:  map[lower.ID]string{},
	}
	a.walk(f)
	return a.annotations, a.diags, a.st
}

// loweredAnalysis is pass 2's state: what the legacy walkers thread through
// every call as arguments.
type loweredAnalysis struct {
	pos         *lower.Positions
	idx         *ProjectIndex
	opts        Options
	st          *SymbolTable
	annotations map[lower.ID]types.Type
	diags       []Diagnostic
	// subs is each file-level named sub's body, which a user-defined
	// guard reads.
	subs map[string]*lower.AnonSub
	// assigned is the name of every variable an assignment writes, which
	// a sub body reads at its runtime value.
	assigned map[string]bool
	// evalCtx is the context the expression around a node evaluates it
	// in, keyed by the node's id and set before the node is walked: an
	// assignment's right side, an operator's operands, a ternary's arms.
	// An absent node's is UnknownCtx.
	evalCtx map[lower.ID]types.Context
	// classTypes is the class of the object a constructor call makes,
	// keyed by the call's id, for the assignment that takes it.
	classTypes map[lower.ID]string
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
	case *lower.Interp:
		a.walkChildren(n)
		t = types.Str
	case *lower.Var:
		t = a.variable(n)
	case *lower.Deref:
		a.walkChildren(n)
		t = derefTypes[n.Sigil]
	case *lower.Index:
		t = a.index(n)
	case *lower.Key:
		t = a.key(n)
	case *lower.Slice:
		a.slice(n)
		t = types.Unknown
	case *lower.ArrayLiteral:
		a.walkChildren(n)
		t = types.ArrayRef
	case *lower.HashLiteral:
		a.walkChildren(n)
		t = types.HashRef
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
	case *lower.Binary:
		t = a.binary(n)
	case *lower.Unary:
		t = a.unary(n)
	case *lower.Ternary:
		t = a.ternary(n)
	case *lower.Repeat:
		t = a.repeat(n)
	case *lower.Range:
		t = a.rangeOp(n)
	case *lower.Assign:
		t = a.assign(n)
	case *lower.Call:
		t = a.call(n)
	case *lower.MethodCall:
		t = a.methodCall(n)
	case *lower.MapGrepSort:
		t = a.mapGrepSort(n)
	case *lower.Ref:
		t = a.ref(n)
	case *lower.Match:
		t = a.match(n)
	case *lower.Glob:
		t = a.glob(n)
	case *lower.Wantarray:
		t = builtinReturns(parse.CoreBuiltin("wantarray"), a.evalCtx[n.ID])
	case *lower.EvalStr:
		// eval STRING is the soundness boundary (§4.14.2): its code is
		// not read, so its value may be anything.
		a.walkChildren(n)
		t = types.Any
	case *lower.EvalBlock, *lower.DoBlock, *lower.DoFile, *lower.Readline, *lower.FileTest:
		// None of these is typed on the legacy path.
		a.walkChildren(n)
		t = types.Unknown
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
// first declared variable is narrowed by its initialiser, as narrow says.
func (a *loweredAnalysis) decl(d *lower.Decl) types.Type {
	switch d.Kind {
	case lower.DeclMy, lower.DeclOur, lower.DeclState, lower.DeclLocal:
		for _, v := range d.Vars {
			a.walk(v)
		}
		if d.Init == nil {
			return types.Unknown
		}
		a.evalCtx[d.Init.Meta().ID] = declContext(d)
		t := a.walk(d.Init)
		// local neither declares nor narrows: the legacy path reads no
		// name through it.
		vars := declaredVars(d)
		if len(vars) == 0 || d.Kind == lower.DeclLocal {
			return t
		}
		if a.bindsCaptures(d) {
			return types.List
		}
		return a.narrow(vars[0], d.Init, t)
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
			widened := a.widenWritten(d.Init)
			a.walk(d.Init)
			ret := types.Unknown
			if body, ok := d.Init.(*lower.AnonSub); ok && body.Body != nil {
				b, _ := body.Body.(*lower.Block)
				ret = a.explicitReturns(body.Body) | a.implicitReturn(b)
			}
			written := map[string]types.Type{}
			for name, typ := range widened {
				if sym, ok := a.st.Lookup(name); ok && sym.Type != typ {
					written[name] = sym.Type
				}
			}
			a.st.ExitScope()
			for name, typ := range written {
				a.st.UpdateType(name, typ)
			}
			if name := subName(d); name != "" && ret != types.Unknown {
				a.st.UpdateReturnType(name, ret)
			}
			a.recordParamTypes(d)
		}
		return types.Code
	}
	// field, class and method: no fixture has captured what the legacy
	// path does with them yet.
	a.walkChildren(d)
	return types.Any
}

// widenWritten shadows, in the sub scope just entered, each variable body
// reads from outside it that an assignment anywhere in the file writes,
// with the variable's sigil type, and returns those it shadowed. A sub
// reads such a variable at the value it has when the sub is called, which
// that assignment may have set: perl 5.42 -w is silent on `my $p = undef;
// sub r { $p eq 'n' } sub w { $p = 'n' } w(); r();`. A variable nothing
// assigns but its declaration keeps its type, so `my $p = undef; sub r {
// $p eq 'n' } r();` is still reported, as perl warns.
//
// What the sub assigns to a shadowed variable reaches the code after it
// as an assignment in its body always has, by the caller's UpdateType.
//
// ponytail: assignments are matched by name, not by the variable they
// resolve to, so a same-named variable elsewhere widens this one too;
// that only ever silences.
func (a *loweredAnalysis) widenWritten(body lower.Node) map[string]types.Type {
	widened := map[string]types.Type{}
	var find func(n lower.Node)
	find = func(n lower.Node) {
		if v, ok := n.(*lower.Var); ok {
			name := v.Sigil + v.Name
			typ, named := sigilTypes[v.Sigil]
			if _, done := widened[name]; named && !done && a.assigned[name] {
				if sym, ok := a.st.Lookup(name); ok && sym.Type != typ {
					sym.Type = typ
					a.st.Define(sym)
					widened[name] = typ
				}
			}
		}
		for _, c := range lower.Children(n) {
			find(c)
		}
	}
	find(body)
	return widened
}

// assignedNames is the name of every variable an assignment writes, by
// name and not by the scope it resolves to: `$x = ...`, `$x .= ...`, and
// each variable of `($x, $y) = ...`. A declaration's initialiser is not
// an assignment.
func assignedNames(f *lower.File) map[string]bool {
	names := map[string]bool{}
	var find func(n lower.Node)
	find = func(n lower.Node) {
		if as, ok := n.(*lower.Assign); ok {
			targets := []lower.Node{as.L}
			if l, ok := as.L.(*lower.List); ok {
				targets = l.Elems
			}
			for _, t := range targets {
				if v, ok := t.(*lower.Var); ok {
					names[v.Sigil+v.Name] = true
				}
			}
		}
		for _, c := range lower.Children(n) {
			find(c)
		}
	}
	find(f)
	return names
}

// declContext is the context a declaration evaluates its initialiser in:
// list context for the parenthesised form or an array or hash, scalar
// context for a scalar.
func declContext(d *lower.Decl) types.Context {
	if d.ListAssign {
		return types.ListCtx
	}
	if len(d.Vars) == 1 {
		if v, ok := d.Vars[0].(*lower.Var); ok {
			return sigilContext(v.Sigil)
		}
	}
	return types.UnknownCtx
}

// sigilContext is the context assigning to a variable of sigil imposes.
func sigilContext(sigil string) types.Context {
	switch sigil {
	case "$":
		return types.ScalarCtx
	case "@", "%":
		return types.ListCtx
	}
	return types.UnknownCtx
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
