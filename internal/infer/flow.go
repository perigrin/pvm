// ABOUTME: The statement and flow rules of Analyze: guard narrowing in conditionals and loops, branch joins, returns and use.
// ABOUTME: Each reads the lowered kinds, and keeps the approximation its goldens hold.

package infer

import (
	"maps"
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
	a.condition(n.Cond)
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
	// dropped collects, across every body of this conditional, the
	// element guard shadows (I5) at least one of them wrote to: branch
	// undoes each body's own touches before the next body runs from the
	// same pre, so this is applied only once every body has run, past
	// the join below.
	dropped := map[string]bool{}

	var ends []map[string]flowVar
	primaryExits := false
	if b, ok := primary.(*lower.Block); ok {
		primaryExits = a.blockExits(b)
		ends = append(ends, a.branch(b, a.flatten(guard, negate), pre, dropped))
	} else if primary != nil {
		a.walk(primary)
	}

	// The body exits and nothing else follows it: what runs after the
	// statement knows the guard failed.
	if primaryExits && guard != nil && guard.VarName != "" && secondary == nil {
		a.restore(pre)
		a.dropElementShadows(dropped)
		narrowContinuation(a.flatten(guard, !negate), a.st)
		return
	}

	exhaustive := hasElse
	if elsif != nil {
		var more []map[string]flowVar
		more, exhaustive = a.elsif(elsif, a.elsifPre(pre, a.flatten(guard, !negate)), dropped)
		ends = append(ends, more...)
	}

	if hasElse {
		ends = append(ends, a.branch(elseBlock, a.flatten(guard, !negate), pre, dropped))
	} else if secondary != nil && elsif == nil {
		a.walk(secondary)
	}

	a.joinBranches(pre, ends, exhaustive)
	a.dropElementShadows(dropped)

	// The body leaves the loop, or starts its next pass, and nothing else
	// follows it: the rest of the loop body knows the guard failed.
	if b, ok := primary.(*lower.Block); ok && secondary == nil && a.blockLeavesLoop(b) {
		a.leaveLoop(a.flatten(guard, !negate))
	}
}

// dropElementShadows deletes, for good, every element guard shadow
// dropped (branch's own undoTouched) collected: now that every body of
// the conditional that collected it has run and been joined, at least
// one of them proved it stale, so nothing past the whole conditional may
// still trust it (I5, R6/S4).
func (a *analysis) dropElementShadows(dropped map[string]bool) {
	for name := range dropped {
		a.st.Delete(name)
	}
}

// flowVar is a variable's type and possible-undef mark together, the pair
// pre/restore/branch/joinBranches carry through a conditional: only an
// actual Undef bit in the type makes the mark matter at all, and the two
// are read and written off the symbol table as one unit.
type flowVar struct {
	Type  types.Type
	Undef bool
}

// preBranchTypes is the type and possible-undef mark each variable named
// in n has before n's bodies run: each body starts from them, as perl
// runs at most one of them, and the join after n takes them for the path
// on which none ran.
func (a *analysis) preBranchTypes(n lower.Node) map[string]flowVar {
	pre := map[string]flowVar{}
	var find func(n lower.Node)
	find = func(n lower.Node) {
		if n == nil {
			return
		}
		if v, ok := n.(*lower.Var); ok {
			if _, named := sigilTypes[v.Sigil]; named {
				if sym, ok := a.st.Lookup(v.Sigil + v.Name); ok {
					pre[v.Sigil+v.Name] = flowVar{sym.Type, sym.PossibleUndef}
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

// restore sets each of pre's variables back to the type and mark pre holds.
func (a *analysis) restore(pre map[string]flowVar) {
	for name, fv := range pre {
		a.setFlow(name, fv)
	}
}

// elsifPre is pre, narrowed by guards for the elsif chain that runs only
// where guards holds: an elsif's condition is reached only once every
// earlier condition in its chain has failed, so each variable that failure
// narrows is narrowed again before the elsif's own guard is read. perl 5.42
// -w is silent on Test2::Mock's `my $type = reftype($val); if (!$type) {
// ... } elsif ($type eq 'HASH') { ... }`: the first branch's failure rules
// out undef before the elsif's `eq` runs.
//
// Only a guard flattened to hold (Negated false) is kept: NarrowByGuard,
// which such a leaf applies, only ever removes Undef or intersects with
// Ref -- safe to carry across branches, as the primary's own guard var
// being unconditionally falsy or unconditionally a ref stays true in the
// elsif too. NegateGuard, a Negated leaf's, is the failed-guard case
// (`defined($x)` failing makes $x exactly Undef) and is precise about the
// primary's own condition, not safe to feed into the elsif's unrelated
// one: TestFlowNarrowingElsifChainThreeBranches's `if (defined($x)) {}
// elsif (ref($x)) {}` must still read $x as Ref in the elsif, not as the
// None a Negated defined-guard narrowing would leave it.
func (a *analysis) elsifPre(pre map[string]flowVar, guards []*guardResult) map[string]flowVar {
	var held []*guardResult
	for _, g := range guards {
		if !g.Negated {
			held = append(held, g)
		}
	}
	if len(held) == 0 {
		return pre
	}
	a.restore(pre)
	shadows, _ := guardShadows(held, a.st)
	if len(shadows) == 0 {
		return pre
	}
	next := maps.Clone(pre)
	for name, typ := range shadows {
		if fv, ok := next[name]; ok {
			next[name] = flowVar{typ, fv.Undef}
		}
	}
	return next
}

// branch walks b, one body of a conditional, under guards and from the
// types pre holds, and returns the type each of pre's variables has at its
// end: nil when b always exits, as its end never reaches the join.
//
// Element guard shadows (flow.go's guardVarName) are not among pre's
// names, so nothing above restores them between sibling bodies the way
// restore(pre) does for an ordinary variable; without the PushFrame span
// here, a write in one body that drops one (ops.go's
// invalidateElementShadow, which deletes from whatever ancestor scope
// actually holds it) would delete it right out of a scope this body
// shares with a sibling that never wrote anything, and the sibling would
// read the stale shadow as gone too (I5, R6/S4's overreach). Undoing
// every touch this body made puts it back for whichever sibling body
// (ifStmt, elsif) runs next from the same pre; dropped collects the
// element shadows among them, for the caller to delete for good only once
// every body has had its turn (dropElementShadows).
func (a *analysis) branch(b *lower.Block, guards []*guardResult, pre map[string]flowVar, dropped map[string]bool) map[string]flowVar {
	a.restore(pre)
	a.st.PushFrame()
	if a.blockExits(b) {
		a.blockWithGuard(b, guards, nil)
		a.undoTouched(a.st.TouchedElems(), dropped)
		a.st.PopFrame()
		return nil
	}
	end := make(map[string]flowVar, len(pre))
	for name := range pre {
		end[name] = flowVar{types.None, false}
	}
	a.blockWithGuard(b, guards, end)
	a.undoTouched(a.st.TouchedElems(), dropped)
	a.st.PopFrame()
	return end
}

// undoTouched puts back each element guard shadow (a "[" in the name,
// guardVarName's own shape) a body touched, as TouchedElems holds them,
// then merges them into dropped, for the caller to delete for good once
// every body of the same conditional has run (I5, R6/S4). Any other
// touched name is left exactly as the body left it: an ordinary
// variable's own narrowing is already handled by pre/restore/joinBranches
// (flowVar), and a fresh local the body declared is scoped by its own
// EnterScope/ExitScope regardless -- undoing either here would discard
// precision the body legitimately established, not protect a sibling
// from anything.
func (a *analysis) undoTouched(touched map[scopedName]Symbol, dropped map[string]bool) {
	for n, before := range touched {
		a.st.put(n, before)
		dropped[n.name] = true
	}
}

// joinBranches sets each of pre's variables to the join of the types the
// bodies of a conditional left it, ends, and of its type before the
// conditional when no body need have run, as when there is no else. A
// body a guard proves unreachable leaves its variable None, the join
// identity, so it adds nothing.
//
// The possible-undef mark joins alongside the type, but not by the same
// rule: true only when every body whose own end still carries an Undef
// bit left the mark set, since a body that leaves an unmarked, plain
// Undef reachable past the conditional is not forgiven just because a
// sibling body's Undef came from a declared call.
func (a *analysis) joinBranches(pre map[string]flowVar, ends []map[string]flowVar, exhaustive bool) {
	for name, before := range pre {
		var reached []flowVar
		for _, end := range ends {
			if end != nil {
				reached = append(reached, end[name])
			}
		}
		a.setFlow(name, joinFlow(before, reached, exhaustive))
	}
}

// joinFlow is the join joinBranches sets one variable to: of ends, the
// values bodies that reached the join left it, and of before, its value
// before the conditional, unless exhaustive says some body always runs.
// With nothing to join, it is before.
func joinFlow(before flowVar, ends []flowVar, exhaustive bool) flowVar {
	if !exhaustive {
		ends = append(ends, before)
	}
	if len(ends) == 0 {
		return flowVar{before.Type, before.Type&types.Undef != 0 && before.Undef}
	}
	joined, marked, any := types.None, true, false
	for _, fv := range ends {
		joined = types.Join(joined, fv.Type)
		if fv.Type&types.Undef != 0 {
			any, marked = true, marked && fv.Undef
		}
	}
	return flowVar{joined, any && marked}
}

// setFlow sets name's type and possible-undef mark to fv's, in one
// write. A variable that already holds them is not touched at all, so restoring or joining the many variables a long elsif chain
// names records nothing for the ones its bodies left alone (V5).
func (a *analysis) setFlow(name string, fv flowVar) {
	s := a.st.scopeOf(name)
	if s == nil {
		return
	}
	sym := s.symbols[name]
	if sym.Type == fv.Type && sym.PossibleUndef == fv.Undef {
		return
	}
	sym.Type, sym.PossibleUndef = fv.Type, fv.Undef
	a.st.set(scopedName{s, name}, sym)
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
		if sym, ok := a.st.Lookup(name); ok && sym.Type != before.Type {
			a.st.UpdateType(name, types.Join(before.Type, sym.Type))
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
// operand held, and past `!defined $x and die` it failed. A statement
// that always exits leaves the rest of its scope unreached, as Unreached
// says.
func (a *analysis) exprStmt(n *lower.ExprStmt) {
	a.walkChildren(n)
	if a.exits(n.X) {
		a.st.Unreached()
	}
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

// loop walks n, a loop, by walk, and after it joins each variable a loop
// exit in it narrowed with the type the variable had where that exit
// left, as leaveLoop recorded: the code after the loop runs on that path
// too. A variable declared in the loop is gone by then, and is skipped.
// A bare block is a loop that runs once.
//
// A deref in a part of the loop that may run zero times -- a body, a
// C-style for's step, a map, grep or sort block -- forces its base to a
// ref only on a pass that runs at all (I4), so past the loop the base
// also holds the type it had where the first such part began, past the
// condition, list or initialiser, which runs at least once: `my $c=0; my
// $y=undef; for my $i (1..$c) { $y->{k} } bless $y, 'Foo'` dies "Can't
// bless non-reference value" in perl 5.42, and `for ($y = {}; $c; ) {
// $y->{k} } bless $y, 'Foo'` does not. Inside the pass, what follows the
// deref sees the ref, as a straight-line deref's does.
func (a *analysis) loop(n lower.Node, walk func()) {
	a.loopExits = append(a.loopExits, map[scopedName]types.Type{})
	spans, parts := a.mayNotRun(n)
	a.loops = append(a.loops, openLoop{a.st.current.depth, spans, parts, map[scopedName]bool{}, map[scopedName]Symbol{}})
	walk()
	l := a.loops[len(a.loops)-1]
	a.loops = a.loops[:len(a.loops)-1]
	for v := range l.derefs {
		if before, ok := l.before[v]; ok && a.st.inScope(v.scope) {
			if sym, ok := v.scope.symbols[v.name]; ok {
				sym.Type = types.Join(sym.Type, before.Type)
				a.st.set(v, sym)
			}
		}
	}
	exits := a.loopExits[len(a.loopExits)-1]
	a.loopExits = a.loopExits[:len(a.loopExits)-1]
	for v, typ := range exits {
		if a.st.inScope(v.scope) {
			if sym, ok := v.scope.symbols[v.name]; ok {
				sym.Type = types.Join(sym.Type, typ)
				a.st.put(v, sym)
			}
		}
	}
}

// openLoop is a loop loop is walking: the depth of the scope it opened
// in; the source spans of its parts that may run zero times, and the ids
// of those not yet walked; each variable a deref in one of them forced to
// a ref; and each variable's value from where the first of them began,
// the value it keeps past the loop when no pass runs.
type openLoop struct {
	depth     int
	mayNotRun [][2]int
	parts     []lower.ID
	derefs    map[scopedName]bool
	before    map[scopedName]Symbol
}

// mayNotRun is the source spans and ids of n's parts that may run zero
// times: a while's or foreach's body and continue block, a C-style for's
// body and step, a map's, grep's or sort's block or expression. A
// do-while's body and a bare block run at least once.
func (a *analysis) mayNotRun(n lower.Node) ([][2]int, []lower.ID) {
	var parts []lower.Node
	switch n := n.(type) {
	case *lower.WhileStmt:
		if !n.PostCond {
			parts = []lower.Node{n.Body, n.Continue}
		}
	case *lower.ForeachStmt:
		parts = []lower.Node{n.Body, n.Continue}
	case *lower.CForStmt:
		parts = []lower.Node{n.Body, n.Post}
	case *lower.MapGrepSort:
		parts = []lower.Node{n.Block, n.FirstExpr}
	}
	var spans [][2]int
	var ids []lower.ID
	for _, p := range parts {
		if p == nil {
			continue
		}
		ids = append(ids, p.Meta().ID)
		if start, end, ok := a.pos.Span(p.Meta().ID); ok {
			spans = append(spans, [2]int{start, end})
		}
	}
	return spans, ids
}

// enterLoopPart reports whether n is a part of the innermost open loop
// that may run zero times, not walked yet, and if so opens a frame for
// leaveLoopPart to read the values the part's derefs began from.
func (a *analysis) enterLoopPart(n lower.Node) bool {
	if len(a.loops) == 0 {
		return false
	}
	l := &a.loops[len(a.loops)-1]
	id := n.Meta().ID
	for i, p := range l.parts {
		if p == id {
			l.parts = append(l.parts[:i:i], l.parts[i+1:]...)
			a.st.PushFrame()
			return true
		}
	}
	return false
}

// leaveLoopPart closes the frame enterLoopPart opened, keeping each
// variable's value from before its first touch in any part -- what it
// held where the first part began -- whether or not a deref has marked
// it yet: a deref in a later part, a while's continue block, may mark a
// variable the body wrote, and loop reads only the marked ones.
func (a *analysis) leaveLoopPart() {
	l := a.loops[len(a.loops)-1]
	for v, before := range a.st.Touched() {
		if _, seen := l.before[v]; !seen {
			l.before[v] = before
		}
	}
	a.st.PopFrame()
}

// markLoopDeref records, in each open loop with a part that may run zero
// times around base, that a deref there forced name to a ref, for loop
// to join its type from before the part back in past it.
//
// The variable marked is the one the deref narrows that is still in
// scope past the loop: a guard's shadow -- in an if or else body, a
// modifier, a ternary arm or a map block -- opened inside the loop is
// gone by then, and the guard's join wrote what the shadow became to the
// variable it narrowed, the holder beneath it. perl 5.42 dies "Can't
// bless non-reference value" on `my $y=undef; for my $i (1..$n) {
// $y->{k}=1 unless defined $y } bless $y, 'Foo'` with $n 0. A `my $y`
// inside the loop is another variable, gone past the loop, and its mark
// is skipped there. Only a guard scope is stepped past: any other scope
// the loop opened that holds name declares a variable of its own, gone
// with the loop, whose deref says nothing of the one past it.
func (a *analysis) markLoopDeref(base lower.Node, name string) {
	if len(a.loops) == 0 {
		return
	}
	start, _, ok := a.pos.Span(base.Meta().ID)
	if !ok {
		return
	}
	holders := a.st.holders[name]
	for _, l := range a.loops {
		i := len(holders) - 1
		for i >= 0 && holders[i].Name == "guard" && holders[i].depth > l.depth {
			i--
		}
		if i < 0 {
			continue
		}
		for _, s := range l.mayNotRun {
			if start >= s[0] && start < s[1] {
				l.derefs[scopedName{holders[i], name}] = true
				break
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
// and whether the chain ends in an else. pre already holds every earlier
// condition's failure narrowed in, as elsifPre says; a further elsif in the
// chain gets this one's failure narrowed in too.
func (a *analysis) elsif(n *lower.IfStmt, pre map[string]flowVar, dropped map[string]bool) ([]map[string]flowVar, bool) {
	a.restore(pre)
	a.condition(n.Cond)
	if n.Cond != nil {
		a.walk(n.Cond)
	}
	guard := a.guard(n.Cond, true)

	var ends []map[string]flowVar
	if b, ok := n.Then.(*lower.Block); ok {
		ends = append(ends, a.branch(b, a.flatten(guard, false), pre, dropped))
	} else if n.Then != nil {
		a.walk(n.Then)
	}

	switch e := n.Else.(type) {
	case *lower.Block:
		return append(ends, a.branch(e, a.flatten(guard, true), pre, dropped)), true
	case *lower.IfStmt:
		more, exhaustive := a.elsif(e, a.elsifPre(pre, a.flatten(guard, true)), dropped)
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
	a.condition(n.Cond)
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
		// Neither shape scopes a loop variable of its own -- `for (@r)
		// {...}` and the modifier form both alias each element to $_
		// directly -- but a write to $_ anywhere in the body still
		// writes through the alias the same way the scoped form's
		// write-back below does, so a guard's per-index shadow of the
		// array is just as stale after it (I5).
		if bodyWritesAlias(n.Body, "$_") {
			a.clearArraysIn(n.List)
		}
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

	// A write through the alias reaches some element of the array
	// whether or not it changes the alias's own type -- s/// and tr///
	// never do -- so this cannot wait for final.Type's own change below
	// to decide whether the array's element shadows are stale (I5). The
	// list may name the array any way that aliases its elements: whole,
	// by element or slice, or through reverse or grep (V6).
	if bodyWritesAlias(body, name) {
		a.clearArraysIn(n.List)
	}

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

// clearArraysIn drops the element guard shadows of each array n names
// whole, by element or by slice: a write through an alias to one of
// n's values may land on any of their elements.
func (a *analysis) clearArraysIn(n lower.Node) {
	if n == nil {
		return
	}
	switch n := n.(type) {
	case *lower.Var:
		if n.Sigil == "@" {
			a.st.ClearElementShadows(n.Name)
		}
	case *lower.Index:
		if v, ok := n.Base.(*lower.Var); ok && !n.Arrow && v.Sigil == "$" {
			a.st.ClearElementShadows(v.Name)
		}
	case *lower.Slice:
		if v, ok := n.Base.(*lower.Var); ok {
			a.st.ClearElementShadows(v.Name)
		}
	}
	for _, c := range lower.Children(n) {
		a.clearArraysIn(c)
	}
}

// bodyWritesAlias reports whether a plain assignment, a substitution or
// transliteration (not suffixed /r, which returns a modified copy
// instead of writing through), an explicit `=~` binding, a call that
// writes it as writtenArgs (ops.go) says or a method call given it, an
// increment or a decrement, or a reference taken to it, writes
// through name (with its sigil, as guardVarName spells it) anywhere under
// n, or may: each mutates its target in place, the way a foreach's alias
// write reaches whatever element of the array it is bound to (I5, V6) --
// a modifier foreach's body, which is the write itself, or a scoped or
// unscoped block form's.
func bodyWritesAlias(n lower.Node, name string) bool {
	if n == nil {
		return false
	}
	writes := func(target lower.Node) bool {
		if target == nil {
			return name == "$_"
		}
		v, ok := target.(*lower.Var)
		return ok && v.Sigil+v.Name == name
	}
	switch n := n.(type) {
	case *lower.Assign:
		if writes(n.L) {
			return true
		}
	case *lower.Match:
		if (n.Op == "s" || n.Op == "tr" || n.Op == "y") && !strings.Contains(n.Flags, "r") && writes(n.Target) {
			return true
		}
	case *lower.Unary:
		if (n.Op == "++" || n.Op == "--") && n.Operand != nil && writes(n.Operand) {
			return true
		}
	case *lower.Ref:
		if n.Operand != nil && writes(n.Operand) {
			return true
		}
	case *lower.Call:
		if base := coreName(n.Name); (base == "chop" || base == "chomp") && len(n.Args) == 0 && name == "$_" {
			return true
		}
		for _, arg := range writtenArgs(n) {
			if writes(arg) {
				return true
			}
		}
	case *lower.MethodCall:
		for _, arg := range n.Args {
			if writes(arg) {
				return true
			}
		}
	}
	for _, c := range lower.Children(n) {
		if bodyWritesAlias(c, name) {
			return true
		}
	}
	return false
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
// what is left satisfies expected. A dualvar in what is left is read as
// lossyCoercion reads it, as readLeg says: perl 5.42 -w is silent on
// `no warnings 'uninitialized'; $x + 1` for $x a dualvar or undef. A
// mismatch with any other part is still reported.
func (a *analysis) silenced(actual, expected types.Type) bool {
	off := a.warningsOff
	if !off.uninitialized && !off.numeric {
		return false
	}
	rest := numified(readLeg(finite(actual), expected, types.DualVar, types.Num), expected)
	if off.uninitialized {
		rest &^= types.Undef
	}
	if off.numeric && types.IsSubtype(expected, types.Num) {
		rest &^= types.Str&^types.Num | types.Glob
	}
	return rest == types.Unknown || a.opts.satisfies(rest, expected) || a.opts.handleSatisfies(rest, expected)
}

// stillPossibleUndef is name's own possible-undef mark, read before a
// guard's narrowed shadow replaces it: a guard only ever removes Undef or
// intersects with Ref (elsifPre's own comment), so an Undef bit the
// shadow's narrowed type still carries is the exact same bit, from the
// same declared call or ternary, that name's own mark already answered
// for (I3) -- `my $b = $c ? blessed($x) : []; if (!ref $b) { $b eq "C" }`
// is silent on pu and perl 5.42 because blessed's Undef is still
// forgiven inside the guard, not a fresh, unmarked one.
func (a *analysis) stillPossibleUndef(name string) bool {
	sym, ok := a.st.Lookup(name)
	return ok && sym.PossibleUndef
}

// blockWithGuard walks a block under guards: the guarded variables
// are shadowed, in one "guard" scope, by what the guards narrow them to,
// and the block's statements are walked under it. When branchTypes is not
// nil it receives the type at the end of the block of each variable it
// holds a key for, a shadowed one's being its shadow's.
func (a *analysis) blockWithGuard(b *lower.Block, guards []*guardResult, branchTypes map[string]flowVar) {
	shadows, classNames := guardShadows(guards, a.st)
	if len(shadows) > 0 {
		a.st.EnterScope("guard")
		for name, typ := range shadows {
			a.st.Define(Symbol{Name: name, Type: typ, ClassType: classNames[name], Kind: SymVariable, PossibleUndef: a.stillPossibleUndef(name)})
		}
	}
	a.block(b)
	for name := range branchTypes {
		if sym, ok := a.st.Lookup(name); ok {
			branchTypes[name] = flowVar{sym.Type, sym.PossibleUndef}
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
// defined either way. Any other variable the right operand narrows --
// narrowDerefBase's arrow-dereference included -- joins the same way with
// what it held before, as walkJoined says: the operand runs on only one of
// the expression's two paths, so whatever it establishes there must still
// include the value the variable keeps on the path where it does not run.
func (a *analysis) logicalRight(op string, l, r lower.Node) {
	var guard *guardResult
	runsWhen := false
	switch op {
	case "&&", "and":
		guard, runsWhen = a.guard(l, true), true
	case "||", "or":
		guard = a.guard(l, true)
	}
	held := a.flatten(guard, !runsWhen)
	shadows, classNames := guardShadows(held, a.st)
	skip := make(map[string]bool, len(held))
	for _, g := range held {
		skip[g.VarName] = true
	}
	if len(shadows) == 0 {
		a.walkJoined(r, skip)
		return
	}
	skipped, _ := guardShadows(a.flatten(guard, runsWhen), a.st)
	a.st.EnterScope("guard")
	for name, typ := range shadows {
		a.st.Define(Symbol{Name: name, Type: typ, ClassType: classNames[name], Kind: SymVariable, PossibleUndef: a.stillPossibleUndef(name)})
	}
	a.walkJoined(r, skip)
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

// walkJoined walks r, then joins each name a mutation inside it actually
// touched with the type and mark it had afterward: r runs on only one of
// two paths, so anything it narrows -- a dereference's base (vars.go's
// narrowDerefBase) chief among them, since nothing else shadows it the
// way a guard's own tested variable already is -- must still include
// what the variable holds on the path where r does not run at all. skip
// names a guard's own tested variable, held separately by the EnterScope
// shadow logicalRight manages around this call, not by this join.
//
// Looking only at names a PushFrame span actually recorded, rather than
// every name preBranchTypes could find anywhere in r's subtree, is what
// keeps a long right-nested ternary or logical chain linear in its depth
// instead of cubic (I10): the old pre, built by walking r's whole
// remaining subtree, and the old post-walk re-Lookup of every one of its
// names, both cost a scope-chain Lookup per name, and that chain grows by
// one guard scope at every level the chain descends.
func (a *analysis) walkJoined(r lower.Node, skip map[string]bool) {
	a.st.PushFrame()
	a.operand(r)
	touched := a.st.Touched()
	pre := make(map[string]flowVar, len(touched))
	end := make(map[string]flowVar, len(touched))
	for n, before := range touched {
		if skip[n.name] || !a.st.inScope(n.scope) {
			continue
		}
		pre[n.name] = flowVar{before.Type, before.PossibleUndef}
		sym := n.scope.symbols[n.name]
		end[n.name] = flowVar{sym.Type, sym.PossibleUndef}
	}
	a.st.PopFrame()
	a.joinBranches(pre, []map[string]flowVar{end}, false)
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
// signature returns None, as CORE.tpm declares die and exit and Carp.tpm
// croak and confess. A call to a sub has the signatures the parser
// resolved it to; a call to a builtin, CORE.tpm's. last, next and redo
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
// of an element read with a literal index, `$arr[0]`. $_[0] is this
// shape too, which is how a user guard's parameter matches it; any other
// named array's literal-index element is read back the same way by index
// (vars.go), so a guard on it narrows that read.
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
