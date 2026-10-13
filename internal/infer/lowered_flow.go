// ABOUTME: The statement and flow rules of AnalyzeLowered: guard narrowing in conditionals and loops, branch joins, returns and use.
// ABOUTME: Each is the legacy rule of the same name rewritten against the lowered kinds, keeping the legacy approximation.

package infer

import (
	"strings"

	"tamarou.com/pvm/internal/lower"
	"tamarou.com/pvm/internal/types"
)

// ifStmt is the legacy walkConditionalStatement on the lowered IfStmt.
//
// `unless` swaps the branches rather than negating Cond (§4.14.5), so the
// branch the guard holds in is Then and the branch it fails in is Else,
// whichever keyword spelled them. The legacy path still distinguishes the
// first body in the source -- only it ends the statement early or is left
// out of the join when it exits -- so primary is that body, and the other
// slot holds the else body or the elsif chain.
//
// A modifier form (`X if C`) is not narrowed: the legacy path never saw it
// as a conditional statement.
func (a *loweredAnalysis) ifStmt(n *lower.IfStmt) {
	if !isBranch(n.Then) || !isBranch(n.Else) {
		a.walkChildren(n)
		return
	}
	if n.Cond != nil {
		a.walk(n.Cond)
	}
	guard := a.guard(n.Cond, true)

	primary, secondary, negate := n.Then, n.Else, false
	if primary == nil || (secondary != nil && a.before(secondary, primary)) {
		primary, secondary, negate = n.Else, n.Then, true
	}
	elseBlock, hasElse := secondary.(*lower.Block)
	elsif, _ := secondary.(*lower.IfStmt)

	preIfTypes := collectPreIfTypes(guard, a.st)

	var ifBranchTypes map[string]types.Type
	primaryExits := false
	if b, ok := primary.(*lower.Block); ok {
		primaryExits = blockExits(b)
		if !primaryExits {
			ifBranchTypes = map[string]types.Type{}
		}
		a.blockWithGuard(b, flattenGuards(guard, negate), ifBranchTypes)
	} else if primary != nil {
		a.walk(primary)
	}

	// The body exits and nothing else follows it: what runs after the
	// statement knows the guard failed.
	if primaryExits && guard != nil && guard.VarName != "" && secondary == nil {
		narrowContinuation(flattenGuards(guard, !negate), a.st)
		return
	}

	var elsifJoinTypes map[string]types.Type
	if elsif != nil {
		elsifJoinTypes = a.elsif(elsif)
	}

	var elseBranchTypes map[string]types.Type
	if hasElse {
		elseBranchTypes = map[string]types.Type{}
		a.blockWithGuard(elseBlock, flattenGuards(guard, !negate), elseBranchTypes)
	} else if secondary != nil && elsif == nil {
		a.walk(secondary)
	}

	applyBranchJoin(guard, preIfTypes, ifBranchTypes, elsifJoinTypes, elseBranchTypes, hasElse, a.st)
}

// isBranch reports whether n can be a block-form conditional's branch: a
// Block, an elsif's IfStmt, or absent. Anything else is a modifier's body.
func isBranch(n lower.Node) bool {
	switch n.(type) {
	case nil, *lower.Block, *lower.IfStmt:
		return true
	}
	return false
}

// before reports whether x starts before y in the source.
func (a *loweredAnalysis) before(x, y lower.Node) bool {
	xs, _, xok := a.pos.Span(x.Meta().ID)
	ys, _, yok := a.pos.Span(y.Meta().ID)
	return xok && yok && xs < ys
}

// elsif is the legacy walkElsifNode: an elsif narrows its own body by its
// own guard and its else by the negation, and returns the union of the
// branch-end types of the variables it guarded, its trailing chain's
// included. The conditions before it do not narrow it.
func (a *loweredAnalysis) elsif(n *lower.IfStmt) map[string]types.Type {
	if n.Cond != nil {
		a.walk(n.Cond)
	}
	guard := a.guard(n.Cond, true)

	var joinTypes map[string]types.Type
	if b, ok := n.Then.(*lower.Block); ok {
		if !blockExits(b) {
			joinTypes = map[string]types.Type{}
		}
		a.blockWithGuard(b, flattenGuards(guard, false), joinTypes)
	} else if n.Then != nil {
		a.walk(n.Then)
	}

	switch e := n.Else.(type) {
	case *lower.Block:
		elseTypes := map[string]types.Type{}
		a.blockWithGuard(e, flattenGuards(guard, true), elseTypes)
		joinTypes = mergeBranchTypes(joinTypes, elseTypes)
	case *lower.IfStmt:
		joinTypes = mergeBranchTypes(joinTypes, a.elsif(e))
	case nil:
	default:
		a.walk(e)
	}
	return joinTypes
}

// whileStmt is the legacy walkLoopStatement: the body runs with the guard
// held. An `until` is a Negated while, so its body runs with the guard
// failed. A modifier form, `do BLOCK while` included, is not narrowed: the
// legacy path never saw it as a loop statement.
func (a *loweredAnalysis) whileStmt(n *lower.WhileStmt) {
	body, ok := n.Body.(*lower.Block)
	if !ok || n.PostCond {
		a.walkChildren(n)
		return
	}
	for _, l := range n.Labels {
		a.walk(l)
	}
	if n.Cond != nil {
		a.walk(n.Cond)
	}
	a.blockWithGuard(body, flattenGuards(a.guard(n.Cond, true), n.Negated), nil)
	if n.Continue != nil {
		a.walk(n.Continue)
	}
}

// foreachStmt is the legacy walkForStatement: the loop variable is a Scalar
// defined in a scope of its own, and what the body wrote to it is written
// back through the alias to what the loop iterates.
//
// `for my $i (LIST)` and `for $i (LIST)` define $i. `foreach ($n)` has no
// loop variable: $n is the aliased slot, so $_ is defined instead and $n
// keeps its outer binding. Any other shape -- `for (@a)` -- has no scalar
// to scope, and is walked as it stands.
func (a *loweredAnalysis) foreachStmt(n *lower.ForeachStmt) {
	body, isBlock := n.Body.(*lower.Block)
	loopVar := foreachVar(n.Var)
	iter := n.List
	if loopVar == nil {
		if v, ok := n.List.(*lower.Var); ok && v.Sigil == "$" && n.Var == nil {
			loopVar, iter = v, nil
		}
	} else if v, ok := n.List.(*lower.Var); ok && v.Sigil == "$" {
		// The legacy path took the first scalar as the loop variable and
		// skipped every later one, so `for my $i ($n)` iterated nothing
		// and scoped $_ as `foreach ($n)` does.
		iter = nil
	}
	if loopVar == nil || !isBlock {
		a.walkChildren(n)
		return
	}

	if iter != nil {
		a.walk(iter)
	}
	name := loopVar.Sigil + loopVar.Name
	a.st.EnterScope("for")
	defer a.st.ExitScope()
	if iter != nil {
		a.st.Define(Symbol{Name: name, Type: types.Scalar, Kind: SymVariable})
	} else {
		a.st.Define(Symbol{Name: "$_", Type: types.Scalar, Kind: SymVariable})
	}
	a.annotations[loopVar.ID] = types.Scalar
	for _, s := range body.Stmts {
		a.walk(s)
	}

	written := types.Unknown
	if final, ok := a.st.Lookup(name); ok && final.Type != types.Scalar {
		written = final.Type
	}
	if underscore, ok := a.st.Lookup("$_"); ok && underscore.Type != types.Scalar {
		written = types.Join(written, underscore.Type)
	}
	if written == types.Unknown {
		return
	}
	// The legacy aliasWriteBack: `foreach ($n)` aliases the scalar itself,
	// `for my $x (@a)` the aggregate's elements.
	if iter == nil {
		if sym, ok := a.st.Lookup(name); ok {
			a.st.UpdateType(name, types.Join(sym.Type, written))
		}
		return
	}
	if v, ok := iter.(*lower.Var); ok && (v.Sigil == "@" || v.Sigil == "%") {
		a.st.UpdateElemType(v.Sigil+v.Name, written)
	}
}

// foreachVar is the scalar a foreach's Var slot names: `my $i` or `$i`.
func foreachVar(n lower.Node) *lower.Var {
	switch n := n.(type) {
	case *lower.Var:
		if n.Sigil == "$" {
			return n
		}
	case *lower.Decl:
		if vars := declaredVars(n); len(vars) == 1 && len(n.Vars) == 1 && vars[0].Sigil == "$" {
			return vars[0]
		}
	}
	return nil
}

// useStmt is the legacy walkUseStatement: a module the project index can
// resolve is analysed, so its symbols are there for a cross-file lookup.
// Without an index, or for a module outside the project, it does nothing.
// `no Foo` and `require Foo` name a module as `use Foo` does, and perl
// loads it for each.
func (a *loweredAnalysis) useStmt(n *lower.UseStmt) {
	if a.idx != nil && n.Module != "" {
		if path, err := a.idx.ResolveModule(n.Module); err == nil {
			_, _ = a.idx.AnalyzeFile(path)
		}
	}
	a.walkChildren(n)
}

// blockWithGuard is the legacy walkBlockWithGuard: the guarded variables
// are shadowed, in one "guard" scope, by what the guards narrow them to,
// and the block's statements are walked under it. When branchTypes is not
// nil it receives each shadowed variable's type at the end of the block.
func (a *loweredAnalysis) blockWithGuard(b *lower.Block, guards []*guardResult, branchTypes map[string]types.Type) {
	shadows, classNames := guardShadows(guards, a.st)
	if len(shadows) > 0 {
		a.st.EnterScope("guard")
		for name, typ := range shadows {
			a.st.Define(Symbol{Name: name, Type: typ, ClassType: classNames[name], Kind: SymVariable})
		}
	}
	for _, s := range b.Stmts {
		a.walk(s)
	}
	if len(shadows) > 0 {
		if branchTypes != nil {
			for name := range shadows {
				if sym, ok := a.st.Lookup(name); ok {
					branchTypes[name] = sym.Type
				}
			}
		}
		a.st.ExitScope()
	}
}

// blockExits is the legacy blockAlwaysExits: a statement directly in the
// block is a return, a die or an exit. Exits nested deeper are not
// looked for, and last, next and redo are not exits here.
func blockExits(b *lower.Block) bool {
	for _, s := range b.Stmts {
		es, ok := s.(*lower.ExprStmt)
		if !ok {
			continue
		}
		switch x := es.X.(type) {
		case *lower.LoopEx:
			if x.Op == "return" {
				return true
			}
		case *lower.Call:
			if x.Name == "die" || x.Name == "exit" {
				return true
			}
		}
	}
	return false
}

// guard is the legacy extractGuardPattern on the lowered tree: the guard a
// condition states, or nil. user allows a call to a sub of this file whose
// result is itself a guard on its argument; the legacy path looks one level
// deep, and so does this.
func (a *loweredAnalysis) guard(n lower.Node, user bool) *guardResult {
	switch n := n.(type) {
	case *lower.Call:
		return a.callGuard(n, user)
	case *lower.Binary:
		switch n.Op {
		case "isa":
			name := guardVarName(n.L)
			if name == "" {
				return nil
			}
			return &guardResult{VarName: name, Guard: types.GuardPattern{Kind: types.GuardIsa}, ClassName: className(n.R)}
		case "&&", "and":
			return compound("&&", a.guard(n.L, user), a.guard(n.R, user))
		case "||", "or":
			return compound("||", a.guard(n.L, user), a.guard(n.R, user))
		}
	case *lower.Unary:
		if n.Op == "!" || n.Op == "not" {
			return negateGuard(a.guard(n.Operand, user))
		}
	}
	return nil
}

// callGuard is the guard a call states: defined or ref on a scalar or an
// element, a guard function of the builtin table on a scalar, or a user
// guard.
func (a *loweredAnalysis) callGuard(c *lower.Call, user bool) *guardResult {
	if len(c.Args) != 1 {
		return nil
	}
	switch c.Name {
	case "defined":
		if name := guardVarName(c.Args[0]); name != "" {
			return &guardResult{VarName: name, Guard: types.GuardPattern{Kind: types.GuardDefined}}
		}
		return nil
	case "ref":
		if name := guardVarName(c.Args[0]); name != "" {
			return &guardResult{VarName: name, Guard: types.GuardPattern{Kind: types.GuardRef}}
		}
		return nil
	}
	arg, ok := c.Args[0].(*lower.Var)
	if !ok || arg.Sigil != "$" || c.Name == "" {
		return nil
	}
	base := c.Name
	if i := strings.LastIndex(base, "::"); i >= 0 {
		base = base[i+2:]
	}
	if g, ok := guardFunctionTable[base]; ok {
		return &guardResult{VarName: "$" + arg.Name, Guard: g}
	}
	if !user {
		return nil
	}
	return a.userGuard(c.Name, "$"+arg.Name)
}

// userGuard is the legacy extractUserDefinedGuard: a sub of this file,
// declared at file level and taking one parameter, whose returned
// expression is a guard on that parameter, is a guard on its argument.
func (a *loweredAnalysis) userGuard(name, arg string) *guardResult {
	if sym, ok := a.st.Lookup(name); !ok || sym.Kind != SymSubroutine {
		return nil
	}
	body := a.subs[name]
	if body == nil {
		return nil
	}
	param := subParam(body)
	if param == "" {
		return nil
	}
	ret := subReturnExpr(body.Body)
	if ret == nil {
		return nil
	}
	return remapGuardVar(a.guard(ret, false), param, arg)
}

// fileSubs maps the name of each named sub declared at file level to its
// body, the last declaration winning as it does in the symbol table.
func fileSubs(f *lower.File) map[string]*lower.AnonSub {
	subs := map[string]*lower.AnonSub{}
	for _, s := range f.Stmts {
		es, ok := s.(*lower.ExprStmt)
		if !ok {
			continue
		}
		if d, ok := es.X.(*lower.Decl); ok && d.Kind == lower.DeclSub {
			if body, ok := d.Init.(*lower.AnonSub); ok {
				subs[subName(d)] = body
			}
		}
	}
	return subs
}

// subParam is the legacy ResolveSubParam: a signature's one mandatory
// scalar, else `my $v = shift` as the body's first statement, else $_[0]
// read anywhere in the body. "" if the sub has none of them.
func subParam(s *lower.AnonSub) string {
	if len(s.Sig) > 0 {
		var params []string
		for _, p := range s.Sig {
			if v, ok := p.(*lower.Var); ok && v.Sigil == "$" {
				params = append(params, "$"+v.Name)
			}
		}
		if len(params) == 1 {
			return params[0]
		}
		return ""
	}
	body, ok := s.Body.(*lower.Block)
	if !ok {
		return ""
	}
	if len(body.Stmts) > 0 {
		if name := shiftParam(body.Stmts[0]); name != "" {
			return name
		}
	}
	if readsArg0(body) {
		return "$_[0]"
	}
	return ""
}

// readsArg0 reports whether $_[0] is read anywhere under n.
func readsArg0(n lower.Node) bool {
	if guardVarName(n) == "$_[0]" {
		if _, ok := n.(*lower.Index); ok {
			return true
		}
	}
	for _, c := range lower.Children(n) {
		if readsArg0(c) {
			return true
		}
	}
	return false
}

// subReturnExpr is the legacy ExtractSubReturnExpr: the value of the
// body's one return, or the last expression statement when it has none.
// More than one return is too much to read, and so is a bare one.
func subReturnExpr(body lower.Node) lower.Node {
	var rets []*lower.LoopEx
	var find func(n lower.Node)
	find = func(n lower.Node) {
		if r, ok := n.(*lower.LoopEx); ok && r.Op == "return" {
			rets = append(rets, r)
			return
		}
		for _, c := range lower.Children(n) {
			if d, ok := c.(*lower.Decl); ok && d.Kind == lower.DeclSub {
				continue
			}
			find(c)
		}
	}
	if body == nil {
		return nil
	}
	find(body)
	switch len(rets) {
	case 0:
	case 1:
		return rets[0].Label
	default:
		return nil
	}
	b, ok := body.(*lower.Block)
	if !ok {
		return nil
	}
	var last lower.Node
	for _, s := range b.Stmts {
		if es, ok := s.(*lower.ExprStmt); ok {
			last = es.X
		}
	}
	return last
}

// guardVarName is the name a guard narrows: $x for a scalar, and the text
// of an element read with a literal index, $_[0], which only a user guard's
// parameter ever matches.
func guardVarName(n lower.Node) string {
	switch n := n.(type) {
	case *lower.Var:
		if n.Sigil == "$" {
			return "$" + n.Name
		}
	case *lower.Index:
		v, vok := n.Base.(*lower.Var)
		i, iok := n.Idx.(*lower.NumLit)
		if vok && iok && !n.Arrow && v.Sigil == "$" {
			return "$" + v.Name + "[" + i.Text + "]"
		}
	}
	return ""
}

// className is the class an isa tests against: the bareword on its right.
func className(n lower.Node) string {
	switch n := n.(type) {
	case *lower.Bareword:
		return n.Name
	case *lower.Call:
		return n.Name
	}
	return ""
}

// compound is the legacy extractCompoundGuard's result: both sides when
// both are guards, the one that is when only one is.
func compound(op string, l, r *guardResult) *guardResult {
	switch {
	case l != nil && r != nil:
		return &guardResult{Compound: &compoundGuard{Op: op, Left: l, Right: r}}
	case l != nil:
		return l
	}
	return r
}

// negateGuard is the legacy extractNegatedGuard's result: a simple guard
// flips, and a compound one goes through De Morgan.
func negateGuard(g *guardResult) *guardResult {
	if g == nil {
		return nil
	}
	if g.Compound == nil {
		n := *g
		n.Negated = !n.Negated
		return &n
	}
	op := "||"
	if g.Compound.Op == "||" {
		op = "&&"
	}
	l, r := *g.Compound.Left, *g.Compound.Right
	l.Negated, r.Negated = !l.Negated, !r.Negated
	return &guardResult{Compound: &compoundGuard{Op: op, Left: &l, Right: &r}}
}

// explicitReturns is the legacy collectExplicitReturns: the join of every
// return's value under n, a bare return giving Undef.
func (a *loweredAnalysis) explicitReturns(n lower.Node) types.Type {
	if r, ok := n.(*lower.LoopEx); ok && r.Op == "return" {
		if r.Label == nil {
			return types.Undef
		}
		return a.annotations[r.Label.Meta().ID]
	}
	var t types.Type
	for _, c := range lower.Children(n) {
		t |= a.explicitReturns(c)
	}
	return t
}

// implicitReturn is the legacy collectImplicitReturn: the type of a
// block's last statement when it is an expression, or the join of both
// bodies of a closing conditional.
func (a *loweredAnalysis) implicitReturn(b *lower.Block) types.Type {
	if b == nil || len(b.Stmts) == 0 {
		return types.Unknown
	}
	switch s := b.Stmts[len(b.Stmts)-1].(type) {
	case *lower.ExprStmt:
		if s.X != nil {
			return a.annotations[s.X.Meta().ID]
		}
	case *lower.IfStmt:
		if !isBranch(s.Then) || !isBranch(s.Else) {
			return types.Unknown
		}
		then, _ := s.Then.(*lower.Block)
		els, _ := s.Else.(*lower.Block)
		return a.implicitReturn(then) | a.implicitReturn(els)
	}
	return types.Unknown
}
