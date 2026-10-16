// ABOUTME: The statement and flow rules of Analyze: guard narrowing in conditionals and loops, branch joins, returns and use.
// ABOUTME: Each reads the lowered kinds, and keeps the approximation its goldens hold.

package infer

import (
	"strings"

	"tamarou.com/pvm/internal/lower"
	"tamarou.com/pvm/internal/types"
)

// ifStmt narrows the branches of an IfStmt by its guard.
//
// `unless` swaps the branches rather than negating Cond (§4.14.5), so the
// branch the guard holds in is Then and the branch it fails in is Else,
// whichever keyword spelled them. The first body in the source is still
// distinguished -- only it ends the statement early or is left
// out of the join when it exits -- so primary is that body, and the other
// slot holds the else body or the elsif chain.
//
// A modifier form (`X if C`) is not narrowed as a conditional statement
// is. Its body, when C is a match it runs only
// after, reads that match's captures, as capture says.
func (a *analysis) ifStmt(n *lower.IfStmt) {
	if !isBranch(n.Then) || !isBranch(n.Else) {
		m, success := successMatch(n)
		if m == nil || success == nil {
			a.walkChildren(n)
			return
		}
		a.walk(m)
		outer := a.modifier
		a.modifier = n
		a.walk(success)
		a.modifier = outer
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

	pre := a.preBranchTypes(n)

	var ends []map[string]types.Type
	primaryExits := false
	if b, ok := primary.(*lower.Block); ok {
		primaryExits = blockExits(b)
		ends = append(ends, a.branch(b, a.flatten(guard, negate), pre))
	} else if primary != nil {
		a.walk(primary)
	}

	// The body exits and nothing else follows it: what runs after the
	// statement knows the guard failed.
	if primaryExits && guard != nil && guard.VarName != "" && secondary == nil {
		a.restore(pre)
		narrowContinuation(a.flatten(guard, !negate), a.st)
		return
	}

	exhaustive := hasElse
	if elsif != nil {
		var more []map[string]types.Type
		more, exhaustive = a.elsif(elsif, pre)
		ends = append(ends, more...)
	}

	if hasElse {
		ends = append(ends, a.branch(elseBlock, a.flatten(guard, !negate), pre))
	} else if secondary != nil && elsif == nil {
		a.walk(secondary)
	}

	a.joinBranches(pre, ends, exhaustive)
}

// preBranchTypes is the type each variable named in n has before n's
// bodies run: each body starts from it, as perl runs at most one of them,
// and the join after n takes it for the path on which none ran.
func (a *analysis) preBranchTypes(n lower.Node) map[string]types.Type {
	pre := map[string]types.Type{}
	var find func(n lower.Node)
	find = func(n lower.Node) {
		if v, ok := n.(*lower.Var); ok {
			if _, named := sigilTypes[v.Sigil]; named {
				if sym, ok := a.st.Lookup(v.Sigil + v.Name); ok {
					pre[v.Sigil+v.Name] = sym.Type
				}
			}
		}
		for _, c := range lower.Children(n) {
			find(c)
		}
	}
	find(n)
	return pre
}

// restore sets each of pre's variables back to the type pre holds.
func (a *analysis) restore(pre map[string]types.Type) {
	for name, typ := range pre {
		a.st.UpdateType(name, typ)
	}
}

// branch walks b, one body of a conditional, under guards and from the
// types pre holds, and returns the type each of pre's variables has at its
// end: nil when b always exits, as its end never reaches the join.
func (a *analysis) branch(b *lower.Block, guards []*guardResult, pre map[string]types.Type) map[string]types.Type {
	a.restore(pre)
	if blockExits(b) {
		a.blockWithGuard(b, guards, nil)
		return nil
	}
	end := make(map[string]types.Type, len(pre))
	for name := range pre {
		end[name] = types.None
	}
	a.blockWithGuard(b, guards, end)
	return end
}

// joinBranches sets each of pre's variables to the join of the types the
// bodies of a conditional left it, ends, and of its type before the
// conditional when no body need have run, as when there is no else. A
// body a guard proves unreachable leaves its variable None, the join
// identity, so it adds nothing.
func (a *analysis) joinBranches(pre map[string]types.Type, ends []map[string]types.Type, exhaustive bool) {
	for name, before := range pre {
		joined, reached := types.None, false
		if !exhaustive {
			joined, reached = before, true
		}
		for _, end := range ends {
			if end != nil {
				joined, reached = types.Join(joined, end[name]), true
			}
		}
		if !reached {
			joined = before
		}
		a.st.UpdateType(name, joined)
	}
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
func (a *analysis) before(x, y lower.Node) bool {
	xs, _, xok := a.pos.Span(x.Meta().ID)
	ys, _, yok := a.pos.Span(y.Meta().ID)
	return xok && yok && xs < ys
}

// elsif walks an elsif, which narrows its own body by its own guard and
// its else by the negation, each body from the types pre holds, as branch
// says. It returns each body's end types, its trailing chain's included,
// and whether the chain ends in an else. The conditions before it do not
// narrow it.
func (a *analysis) elsif(n *lower.IfStmt, pre map[string]types.Type) ([]map[string]types.Type, bool) {
	a.restore(pre)
	if n.Cond != nil {
		a.walk(n.Cond)
	}
	guard := a.guard(n.Cond, true)

	var ends []map[string]types.Type
	if b, ok := n.Then.(*lower.Block); ok {
		ends = append(ends, a.branch(b, a.flatten(guard, false), pre))
	} else if n.Then != nil {
		a.walk(n.Then)
	}

	switch e := n.Else.(type) {
	case *lower.Block:
		return append(ends, a.branch(e, a.flatten(guard, true), pre)), true
	case *lower.IfStmt:
		more, exhaustive := a.elsif(e, pre)
		return append(ends, more...), exhaustive
	case nil:
	default:
		a.walk(e)
	}
	return ends, false
}

// whileStmt walks a while loop, whose body runs with the guard
// held. An `until` is a Negated while, so its body runs with the guard
// failed. A modifier form, `do BLOCK while` included, is not narrowed as a
// loop statement is.
func (a *analysis) whileStmt(n *lower.WhileStmt) {
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
	a.blockWithGuard(body, a.flatten(a.guard(n.Cond, true), n.Negated), nil)
	if n.Continue != nil {
		a.walk(n.Continue)
	}
}

// foreachStmt walks a foreach loop, whose loop variable is a Scalar
// defined in a scope of its own, and what the body wrote to it is written
// back through the alias to what the loop iterates.
//
// `for my $i (LIST)` and `for $i (LIST)` define $i, and `foreach ($n)`,
// which has no loop variable, defines $_. A loop over one scalar,
// `for my $i ($n)` or `foreach ($n)`, runs once with its variable aliased
// to $n. Any other shape -- `for (@a)` -- has no scalar to scope, and is
// walked as it stands.
func (a *analysis) foreachStmt(n *lower.ForeachStmt) {
	body, isBlock := n.Body.(*lower.Block)
	loopVar := foreachVar(n.Var)
	iter, _ := n.List.(*lower.Var)
	overScalar := iter != nil && iter.Sigil == "$"
	if !isBlock || (loopVar == nil && (n.Var != nil || !overScalar)) {
		a.walkChildren(n)
		return
	}

	if n.List != nil {
		a.walk(n.List)
	}
	name := "$_"
	if loopVar != nil {
		name = loopVar.Sigil + loopVar.Name
		a.annotations[loopVar.ID] = types.Scalar
	}
	a.st.EnterScope("for")
	a.st.Define(Symbol{Name: name, Type: types.Scalar, Kind: SymVariable})
	a.block(body)
	final, _ := a.st.Lookup(name)
	a.st.ExitScope()

	if final.Type == types.Scalar || final.Type == types.Unknown || iter == nil {
		return
	}
	// The alias write-back: a loop over one scalar aliases the scalar
	// itself, `for my $x (@a)` the aggregate's elements.
	switch iter.Sigil {
	case "$":
		if sym, ok := a.st.Lookup("$" + iter.Name); ok {
			a.st.UpdateType("$"+iter.Name, types.Join(sym.Type, final.Type))
		}
	case "@", "%":
		a.st.UpdateElemType(iter.Sigil+iter.Name, final.Type)
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

// useStmt walks a use statement: a module the project index can
// resolve is analysed, so its symbols are there for a cross-file lookup.
// Without an index, or for a module outside the project, it does nothing.
// `no Foo` and `require Foo` name a module as `use Foo` does, and perl
// loads it for each.
func (a *analysis) useStmt(n *lower.UseStmt) {
	if a.idx != nil && n.Module != "" {
		if path, err := a.idx.ResolveModule(n.Module); err == nil {
			_, _ = a.idx.AnalyzeFile(path)
		}
	}
	a.walkChildren(n)
}

// blockWithGuard walks a block under guards: the guarded variables
// are shadowed, in one "guard" scope, by what the guards narrow them to,
// and the block's statements are walked under it. When branchTypes is not
// nil it receives the type at the end of the block of each variable it
// holds a key for, a shadowed one's being its shadow's.
func (a *analysis) blockWithGuard(b *lower.Block, guards []*guardResult, branchTypes map[string]types.Type) {
	shadows, classNames := guardShadows(guards, a.st)
	if len(shadows) > 0 {
		a.st.EnterScope("guard")
		for name, typ := range shadows {
			a.st.Define(Symbol{Name: name, Type: typ, ClassType: classNames[name], Kind: SymVariable})
		}
	}
	a.block(b)
	for name := range branchTypes {
		if sym, ok := a.st.Lookup(name); ok {
			branchTypes[name] = sym.Type
		}
	}
	if len(shadows) > 0 {
		a.st.ExitScope()
	}
}

// logicalRight walks a logical operator's right operand under what its left
// operand guards: perl runs the right operand of `&&` and `and` only where
// the left is true, and of `||` and `or` only where it is false. perl 5.42
// -w is silent on `defined $ch and $ch eq '}'` and `!defined $ch or $ch ne
// ')'` with $ch undef. Any other operator's right operand is walked as it
// stands.
//
// What the right operand assigns to a guarded variable joins the type the
// variable has where the right operand is skipped, on which the left
// operand's test went the other way: `defined $x or $x = 'b'` leaves $x
// defined either way.
func (a *analysis) logicalRight(op string, l, r lower.Node) {
	var guard *guardResult
	runsWhen := false
	switch op {
	case "&&", "and":
		guard, runsWhen = a.guard(l, true), true
	case "||", "or":
		guard = a.guard(l, true)
	}
	shadows, classNames := guardShadows(a.flatten(guard, !runsWhen), a.st)
	if len(shadows) == 0 {
		a.operand(r)
		return
	}
	skipped, _ := guardShadows(a.flatten(guard, runsWhen), a.st)
	a.st.EnterScope("guard")
	for name, typ := range shadows {
		a.st.Define(Symbol{Name: name, Type: typ, ClassType: classNames[name], Kind: SymVariable})
	}
	a.operand(r)
	written := map[string]types.Type{}
	for name, typ := range shadows {
		if sym, ok := a.st.Lookup(name); ok && sym.Type != typ {
			written[name] = sym.Type
		}
	}
	a.st.ExitScope()
	for name, typ := range written {
		other, ok := skipped[name]
		if !ok {
			sym, _ := a.st.Lookup(name)
			other = sym.Type
		}
		a.st.UpdateType(name, types.Join(other, typ))
	}
}

// flatten is flattenGuards less each defined guard on a variable typed
// Any. Any is the permissive top, which a defined test leaves as it is, as
// TypeScript's `any` stays `any` past `x !== undefined`: narrowed, it is
// a union of every defined type, Void and Array among them, that fails a
// Scalar slot. perl 5.42 -w is silent on Math::Complex's
// `my $i = eval "9**9**9"; if (defined $i && $i > 1e308) { $Inf = $i }
// $Inf = 1.79e308 unless defined $Inf`.
func (a *analysis) flatten(guard *guardResult, negate bool) []*guardResult {
	var kept []*guardResult
	for _, g := range flattenGuards(guard, negate) {
		if g.Guard.Kind == types.GuardDefined && !g.Negated {
			if sym, ok := a.st.Lookup(g.VarName); ok && sym.Type == types.Any {
				continue
			}
		}
		kept = append(kept, g)
	}
	return kept
}

// blockExits reports whether a block always exits: a statement directly in the
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

// guard is the guard a
// condition states, or nil. user allows a call to a sub of this file whose
// result is itself a guard on its argument, one level deep.
func (a *analysis) guard(n lower.Node, user bool) *guardResult {
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
		case "eq":
			if g := referentGuard(n.L, n.R); g != nil {
				return g
			}
			return referentGuard(n.R, n.L)
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

// referentGuard is the guard `ref \$v eq 'TYPE'` states on $v, with
// call the `ref \$v` and lit the literal TYPE, or nil when they are not
// that. It tests what $v itself holds, not what $v refers to.
func referentGuard(call, lit lower.Node) *guardResult {
	c, ok := call.(*lower.Call)
	if !ok || c.Name != "ref" || c.Fn != nil || len(c.Args) != 1 {
		return nil
	}
	r, ok := c.Args[0].(*lower.Ref)
	if !ok {
		return nil
	}
	v, ok := r.Operand.(*lower.Var)
	content, known := literalContent(lit)
	if !ok || v.Sigil != "$" || !known {
		return nil
	}
	return &guardResult{VarName: "$" + v.Name, Guard: types.GuardPattern{Kind: types.GuardReferent, RefType: content}}
}

// callGuard is the guard a call states: defined or ref on a scalar or an
// element, a guard function of the builtin table on a scalar, or a user
// guard.
func (a *analysis) callGuard(c *lower.Call, user bool) *guardResult {
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

// userGuard is a user-defined guard: a sub of this file,
// declared at file level and taking one parameter, whose returned
// expression is a guard on that parameter, is a guard on its argument.
func (a *analysis) userGuard(name, arg string) *guardResult {
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

// subParam is a sub's one parameter: a signature's one mandatory
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

// subReturnExpr is a sub's returned expression: the value of the
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

// compound is a compound guard: both sides when
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

// negateGuard is a guard's negation: a simple guard
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

// explicitReturns is the join of every
// return's value under n, a bare return giving Undef.
func (a *analysis) explicitReturns(n lower.Node) types.Type {
	if r, ok := n.(*lower.LoopEx); ok && r.Op == "return" {
		if r.Label == nil {
			return types.Undef
		}
		return a.returnedValue(r.Label)
	}
	var t types.Type
	for _, c := range lower.Children(n) {
		t |= a.explicitReturns(c)
	}
	return t
}

// returnedValue is the type of what a sub hands back in scalar context. A
// List is its last element there, by the comma operator; anything else is
// its own type.
//
// perl: `sub r { return ("a", 2) }` makes `my $v = r()` 2, and
// `sub t { return 1, "b" }` makes `my $u = t()` "b".
func (a *analysis) returnedValue(n lower.Node) types.Type {
	if l, ok := n.(*lower.List); ok && len(l.Elems) > 0 && l.Elems[len(l.Elems)-1] != nil {
		n = l.Elems[len(l.Elems)-1]
	}
	return a.annotations[n.Meta().ID]
}

// implicitReturn is the type of a
// block's last statement when it is an expression, or the join of both
// bodies of a closing conditional.
func (a *analysis) implicitReturn(b *lower.Block) types.Type {
	if b == nil || len(b.Stmts) == 0 {
		return types.Unknown
	}
	switch s := b.Stmts[len(b.Stmts)-1].(type) {
	case *lower.ExprStmt:
		if s.X != nil {
			return a.returnedValue(s.X)
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
