// ABOUTME: The statement and flow rules of Analyze: guard narrowing in conditionals and loops, branch joins, returns and use.
// ABOUTME: Each reads the lowered kinds, and keeps the approximation its goldens hold.

package infer

import (
	"strings"

	"tamarou.com/pvm/internal/lower"
	"tamarou.com/pvm/internal/parse"
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
// A modifier form (`X if C`) is walked as modifierStmt says. Its body, when C
// is a match it runs only after, reads that match's captures, as capture
// says. When its body exits, what follows knows the guard went the other
// way, as modifierExit says.
func (a *analysis) ifStmt(n *lower.IfStmt) {
	if !isBranch(n.Then) || !isBranch(n.Else) {
		m, success := successMatch(n)
		if m == nil || success == nil {
			a.modifierStmt(n)
		} else {
			a.walk(m)
			outer := a.modifier
			a.modifier = n
			a.walk(success)
			a.modifier = outer
		}
		a.modifierExit(n)
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
		primaryExits = a.blockExits(b)
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

	// The body leaves the loop, or starts its next pass, and nothing else
	// follows it: the rest of the loop body knows the guard failed.
	if b, ok := primary.(*lower.Block); ok && secondary == nil && a.blockLeavesLoop(b) {
		a.leaveLoop(a.flatten(guard, !negate))
	}
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
	if a.blockExits(b) {
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

// modifierStmt walks a modifier form, `X if C` or `X unless C`, as `C && X`
// and `C || X`: X runs only where C held, or failed, and is walked under
// what C guards there, as logicalRight says. What X assigns a variable C
// does not guard joins what it held before, as X may not have run: perl
// 5.42 -w is silent on `$a = int($a) if defined $a`, and warns "Use of
// uninitialized value" on `my $a = undef; $a = 5 if @ARGV; $a + 1` with
// no arguments.
func (a *analysis) modifierStmt(n *lower.IfStmt) {
	if n.Cond != nil {
		a.walk(n.Cond)
	}
	body, op := n.Then, "&&"
	if body == nil {
		body, op = n.Else, "||"
	}
	if body == nil {
		return
	}
	pre := a.preBranchTypes(body)
	for _, g := range flattenGuards(a.guard(n.Cond, true), false) {
		delete(pre, g.VarName)
	}
	for _, g := range flattenGuards(a.guard(n.Cond, true), true) {
		delete(pre, g.VarName)
	}
	a.logicalRight(op, n.Cond, body)
	for name, before := range pre {
		if sym, ok := a.st.Lookup(name); ok && sym.Type != before {
			a.st.UpdateType(name, types.Join(before, sym.Type))
		}
	}
}

// modifierExit narrows what follows a modifier form whose body exits:
// past `die unless defined $x` the guard held, and past `exit if defined
// $x` it failed. `unless` puts the body in Else.
func (a *analysis) modifierExit(n *lower.IfStmt) {
	body, negate := n.Then, true
	if body == nil {
		body, negate = n.Else, false
	}
	if a.exits(body) {
		narrowContinuation(a.flatten(a.guard(n.Cond, true), negate), a.st)
	} else if leavesLoop(body) {
		a.leaveLoop(a.flatten(a.guard(n.Cond, true), negate))
	}
}

// exprStmt walks an expression statement. A logical operator whose right
// operand exits narrows what follows it: past `defined $x or die` the left
// operand held, and past `!defined $x and die` it failed.
func (a *analysis) exprStmt(n *lower.ExprStmt) {
	a.walkChildren(n)
	b, ok := n.X.(*lower.Binary)
	if !ok || (!a.exits(b.R) && !leavesLoop(b.R)) {
		return
	}
	narrow := func(guards []*guardResult) { narrowContinuation(guards, a.st) }
	if !a.exits(b.R) {
		narrow = a.leaveLoop
	}
	switch b.Op {
	case "||", "or":
		narrow(a.flatten(a.guard(b.L, true), false))
	case "&&", "and":
		narrow(a.flatten(a.guard(b.L, true), true))
	}
}

// scopedName is a variable as the scope that declares it holds it.
type scopedName struct {
	scope *Scope
	name  string
}

// loop walks a loop by walk, and after it joins each variable a loop
// exit in it narrowed with the type the variable had where that exit
// left, as leaveLoop recorded: the code after the loop runs on that path
// too. A variable declared in the loop is gone by then, and is skipped.
// A bare block is a loop that runs once.
func (a *analysis) loop(walk func()) {
	a.loopExits = append(a.loopExits, map[scopedName]types.Type{})
	walk()
	exits := a.loopExits[len(a.loopExits)-1]
	a.loopExits = a.loopExits[:len(a.loopExits)-1]
	for v, typ := range exits {
		if a.st.inScope(v.scope) {
			if sym, ok := v.scope.symbols[v.name]; ok {
				sym.Type = types.Join(sym.Type, typ)
				v.scope.symbols[v.name] = sym
			}
		}
	}
}

// leaveLoop narrows, as narrowContinuation does, what follows a last,
// next or redo that runs only when guards fail, which the rest of the
// loop body runs past: perl 5.42 -w is silent on `my $len =
// sysread(...); last unless defined $len; $len == 0`. Where the exit
// leaves, each variable still holds its type before the narrowing, which
// the loop joins in after it, as loop says. Outside a loop nothing is
// narrowed.
func (a *analysis) leaveLoop(guards []*guardResult) {
	if len(a.loopExits) == 0 {
		return
	}
	exits := a.loopExits[len(a.loopExits)-1]
	for _, g := range guards {
		if s := a.st.scopeOf(g.VarName); s != nil {
			v := scopedName{s, g.VarName}
			exits[v] = types.Join(exits[v], s.symbols[g.VarName].Type)
		}
	}
	narrowContinuation(guards, a.st)
}

// blockLeavesLoop reports whether a statement directly in b is a last,
// next or redo, as leavesLoop says.
func (a *analysis) blockLeavesLoop(b *lower.Block) bool {
	for _, s := range b.Stmts {
		if es, ok := s.(*lower.ExprStmt); ok && leavesLoop(es.X) {
			return true
		}
	}
	return false
}

// leavesLoop reports whether n is a last, next or redo: the rest of its
// loop's body does not run past it.
func leavesLoop(n lower.Node) bool {
	x, ok := n.(*lower.LoopEx)
	return ok && (x.Op == "last" || x.Op == "next" || x.Op == "redo")
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
	if n.Module == "warnings" && n.Op != "require" {
		a.warningsOff.set(n.Op == "no", n.Args)
	}
	if a.idx != nil && n.Module != "" {
		if path, err := a.idx.ResolveModule(n.Module); err == nil {
			_, _ = a.idx.AnalyzeFile(path)
		}
	}
	a.walkChildren(n)
}

// warningsOff is the warning categories PSC maps a diagnostic to that a
// `no warnings` has turned off. perl 5.42 -w, measured: `no warnings
// "uninitialized"` silences "Use of uninitialized value" in `"x" . $u`,
// `$u + 1`, `lc($u)` and `waitpid($u, 0)`, and `no warnings "numeric"`
// silences "Argument isn't numeric" for `"abc" + 1`, `"abc" < 5` and
// `*STDOUT + 1`. Neither silences "Not a GLOB reference" for close({}),
// which perl dies with.
type warningsOff struct {
	uninitialized, numeric bool
}

// set turns the categories args names off, for `no warnings`, or back on,
// for `use warnings`. No list, or "all", names both. A category PSC maps
// no diagnostic to changes nothing.
func (w *warningsOff) set(off bool, args lower.Node) {
	words := pragmaWords(args)
	if len(words) == 0 {
		words = []string{"all"}
	}
	for _, c := range words {
		switch c {
		case "all":
			w.uninitialized, w.numeric = off, off
		case "uninitialized":
			w.uninitialized = off
		case "numeric":
			w.numeric = off
		}
	}
}

// pragmaWords is the words an import list spells: qw/a b/, "a", or a
// list of them.
func pragmaWords(n lower.Node) []string {
	switch n := n.(type) {
	case *lower.QwList:
		return n.Words
	case *lower.StrLit:
		return []string{n.Value}
	case *lower.List:
		var words []string
		for _, e := range n.Elems {
			words = append(words, pragmaWords(e)...)
		}
		return words
	}
	return nil
}

// silenced reports whether a `no warnings` in force turns off the warning
// perl gives for a value of type actual where expected is wanted: once
// the parts of actual its categories cover are taken away, Undef for
// uninitialized and a string or a glob in a numeric slot for numeric,
// what is left satisfies expected. A mismatch with any other part is
// still reported.
func (a *analysis) silenced(actual, expected types.Type) bool {
	off := a.warningsOff
	if !off.uninitialized && !off.numeric {
		return false
	}
	rest := numified(finite(actual), expected)
	if off.uninitialized {
		rest &^= types.Undef
	}
	if off.numeric && types.IsSubtype(expected, types.Num) {
		rest &^= types.Str&^types.Num | types.Glob
	}
	return rest == types.Unknown || a.opts.satisfies(rest, expected) || a.opts.handleSatisfies(rest, expected)
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
		if (g.Guard.Kind == types.GuardDefined || g.Guard.Kind == types.GuardTruthy) && !g.Negated {
			if sym, ok := a.st.Lookup(g.VarName); ok && sym.Type == types.Any {
				continue
			}
		}
		kept = append(kept, g)
	}
	return kept
}

// blockExits reports whether a block always exits: a statement directly in the
// block is an exit, as exits says. Exits nested deeper are not looked for.
func (a *analysis) blockExits(b *lower.Block) bool {
	for _, s := range b.Stmts {
		if es, ok := s.(*lower.ExprStmt); ok && a.exits(es.X) {
			return true
		}
	}
	return false
}

// exits reports whether n never returns: a return, or a call whose
// signature returns None, as CORE.pmt declares die and exit and Carp.pmt
// croak and confess. A call to a sub has the signatures the parser
// resolved it to; a call to a builtin, CORE.pmt's. last, next and redo
// are not exits here.
func (a *analysis) exits(n lower.Node) bool {
	switch x := n.(type) {
	case *lower.LoopEx:
		return x.Op == "return"
	case *lower.Call:
		if x.Fn != nil || x.Ampersand {
			return false
		}
		sigs := x.Signatures
		if len(sigs) == 0 && !overridden(x) {
			sigs = parse.CoreBuiltin(coreName(x.Name))
		}
		for _, s := range sigs {
			if s.Returns != types.None {
				return false
			}
		}
		return len(sigs) > 0
	}
	return false
}

// guard is the guard a
// condition states, or nil. user allows a call to a sub of this file whose
// result is itself a guard on its argument, one level deep.
//
// A scalar is a guard on its own truth, which a true value passes only
// when defined: perl 5.42 -w is silent on `if ($pid) { waitpid($pid, 0) }`
// with $pid from fork. A condition that assigns a scalar tests the value
// it assigns, as assignedVar says.
func (a *analysis) guard(n lower.Node, user bool) *guardResult {
	switch n := assignedVar(n).(type) {
	case *lower.Var:
		if name := guardVarName(n); name != "" {
			return &guardResult{VarName: name, Guard: types.GuardPattern{Kind: types.GuardTruthy}}
		}
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
			for _, g := range []*guardResult{referentGuard(n.L, n.R), referentGuard(n.R, n.L), refEqGuard(n.L, n.R), refEqGuard(n.R, n.L)} {
				if g != nil {
					return g
				}
			}
			return nil
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

// assignedVar is the scalar variable n assigns when n is a scalar
// assignment, `$x = EXPR`, or a declaration of one scalar with an
// initialiser, `my $x = EXPR`: their value is what the variable holds
// after, so a guard on them is a guard on it. perl 5.42 -w is silent on
// `while (defined($r = read($fh, $buf, 4)) && $r > 0)` and on `if (my
// $pid = fork) { waitpid($pid, 0) }`. A list assignment's value is a
// count, and is not one. Anything else is n itself.
func assignedVar(n lower.Node) lower.Node {
	switch x := n.(type) {
	case *lower.Assign:
		if v, ok := x.L.(*lower.Var); ok && x.Op == "=" && !x.ListAssign && v.Sigil == "$" {
			return v
		}
	case *lower.Decl:
		if x.Kind == lower.DeclLocal || x.ListAssign || x.Init == nil {
			return n
		}
		if vars := declaredVars(x); len(vars) == 1 && len(x.Vars) == 1 && vars[0].Sigil == "$" {
			return vars[0]
		}
	}
	return n
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

// refEqGuard is the guard `ref($v) eq 'TYPE'` states on $v, with call
// the `ref($v)` and lit the literal TYPE, or nil when they are not that:
// $v is a reference of that type. perl 5.42 -w is silent on Encode::Alias's
// `ref($alias) eq 'Regexp' && $k =~ $alias`.
func refEqGuard(call, lit lower.Node) *guardResult {
	c, ok := call.(*lower.Call)
	if !ok || c.Name != "ref" || c.Fn != nil || len(c.Args) != 1 {
		return nil
	}
	name := guardVarName(c.Args[0])
	content, known := literalContent(lit)
	if name == "" || !known || content == "" {
		return nil
	}
	return &guardResult{VarName: name, Guard: types.GuardPattern{Kind: types.GuardRef, RefType: content}}
}

// callGuard is the guard a call states: defined or ref on a scalar or an
// element, a guard function of the builtin table on a scalar, or a user
// guard.
func (a *analysis) callGuard(c *lower.Call, user bool) *guardResult {
	if len(c.Args) != 1 {
		return nil
	}
	switch coreName(c.Name) {
	case "defined":
		if name := guardVarName(assignedVar(c.Args[0])); name != "" {
			return &guardResult{VarName: name, Guard: types.GuardPattern{Kind: types.GuardDefined}}
		}
		return nil
	case "ref":
		if name := guardVarName(assignedVar(c.Args[0])); name != "" {
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
	// A guard on any other variable, one the body declares, says
	// nothing of the caller's.
	g := a.guard(ret, false)
	if !guardsOnly(g, param) {
		return nil
	}
	return remapGuardVar(g, param, arg)
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

// subParam is a sub's one parameter, as subParams finds them: "" when
// it has none, or more than one.
func subParam(s *lower.AnonSub) string {
	if params := subParams(s); len(params) == 1 {
		return params[0]
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

// subReturnExpr is a sub's returned expression: the value of its one
// return when that return is the body's last statement, or the body's
// last statement when that is an expression and the sub has no return. A
// return a modifier makes conditional, or that is not last, is one of two
// values the sub may hand back, and more than one return is too much to
// read, and so is a bare one.
func subReturnExpr(body lower.Node) lower.Node {
	b, ok := body.(*lower.Block)
	if !ok || len(b.Stmts) == 0 {
		return nil
	}
	var rets []*lower.LoopEx
	returnsOf(b, func(r *lower.LoopEx, _ *lower.IfStmt) {
		rets = append(rets, r)
	})
	last, ok := b.Stmts[len(b.Stmts)-1].(*lower.ExprStmt)
	if !ok {
		return nil
	}
	switch {
	case len(rets) == 0:
		return last.X
	case len(rets) == 1 && last.X == lower.Node(rets[0]):
		return rets[0].Label
	}
	return nil
}

// guardsOnly reports whether every leaf of g guards name. A side of a
// compound guard that is no guard says nothing of any variable.
func guardsOnly(g *guardResult, name string) bool {
	if g == nil {
		return true
	}
	if g.Compound != nil {
		return guardsOnly(g.Compound.Left, name) && guardsOnly(g.Compound.Right, name)
	}
	return g.VarName == name
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

// compound is a compound guard, or nil when neither side is a guard. A
// side that is not a guard is a nil side, which says nothing, and the
// operator is kept: `defined $x && $y` failing may be $y failing.
func compound(op string, l, r *guardResult) *guardResult {
	if l == nil && r == nil {
		return nil
	}
	return &guardResult{Compound: &compoundGuard{Op: op, Left: l, Right: r}}
}

// negateGuard is a guard's negation: a simple guard
// flips, and a compound one goes through De Morgan at every level.
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
	return &guardResult{Compound: &compoundGuard{Op: op, Left: negateGuard(g.Compound.Left), Right: negateGuard(g.Compound.Right)}}
}

// explicitReturns is the join of the value of every return of the sub
// whose body is n, a bare return giving Undef, and whether each of them
// is of a known type.
func (a *analysis) explicitReturns(n lower.Node) (types.Type, bool) {
	var t types.Type
	known := true
	returnsOf(n, func(r *lower.LoopEx, _ *lower.IfStmt) {
		if r.Label == nil {
			t |= types.Undef
			return
		}
		v := a.returnedValue(r.Label)
		t |= v
		known = known && v != types.Unknown
	})
	return t, known
}

// returnsOf calls visit with each return of the sub whose body is n, in
// source order, and the `if` statement whose successful match guards it,
// or nil. A return in a sub nested in n, named or anonymous, returns from
// that sub, and is not visited.
func returnsOf(n lower.Node, visit func(r *lower.LoopEx, match *lower.IfStmt)) {
	var walk func(n lower.Node, match *lower.IfStmt)
	walk = func(n lower.Node, match *lower.IfStmt) {
		switch n := n.(type) {
		case nil, *lower.AnonSub:
			return
		case *lower.Decl:
			if n.Kind == lower.DeclSub {
				return
			}
		case *lower.LoopEx:
			if n.Op == "return" {
				visit(n, match)
				return
			}
		case *lower.IfStmt:
			if m, success := successMatch(n); m != nil {
				for _, c := range lower.Children(n) {
					if c == success {
						walk(c, n)
					} else {
						walk(c, nil)
					}
				}
				return
			}
		}
		for _, c := range lower.Children(n) {
			walk(c, match)
		}
	}
	walk(n, nil)
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

// implicitReturn is the type of the value a block hands back when
// control falls off its end, and whether it is known: its last
// statement's when that is an expression, or the join of both bodies of a
// closing conditional. A block that always returns, dies or exits falls
// off nothing, and adds no type that is not known. Any other last
// statement's value, and an empty block's, is not known.
func (a *analysis) implicitReturn(b *lower.Block) (types.Type, bool) {
	if b == nil || len(b.Stmts) == 0 {
		return types.Unknown, false
	}
	if a.blockExits(b) {
		return types.Unknown, true
	}
	switch s := b.Stmts[len(b.Stmts)-1].(type) {
	case *lower.ExprStmt:
		if s.X != nil {
			t := a.returnedValue(s.X)
			return t, t != types.Unknown
		}
	case *lower.IfStmt:
		if !isBranch(s.Then) || !isBranch(s.Else) {
			return types.Unknown, false
		}
		then, _ := s.Then.(*lower.Block)
		els, _ := s.Else.(*lower.Block)
		tt, tk := a.implicitReturn(then)
		et, ek := a.implicitReturn(els)
		return tt | et, tk && ek
	}
	return types.Unknown, false
}
