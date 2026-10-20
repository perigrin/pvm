// ABOUTME: The call rules of Analyze: one Call arm for every builtin and user sub, MethodCall and the class it records, and parameter types.
// ABOUTME: Each reads the lowered kinds: the filehandle is Call.Handle, never one of the Args.

package infer

import (
	"strconv"
	"strings"

	"tamarou.com/pvm/internal/lower"
	"tamarou.com/pvm/internal/parse"
	"tamarou.com/pvm/internal/types"
)

// call types a call to a builtin or a user sub: a builtin's CORE.tpm
// candidates check the arity and each argument's type and give the call's
// type; a sub of this file gives its inferred return type, and a
// fully-qualified one the project index's. A call through a reference is
// not typed. `&name` never calls CORE, so it is a user sub's call whatever
// its name, and so is a builtin the parse resolved to a sub, as overridden
// says.
//
// The filehandle of `print $fh LIST` is Handle, not an argument, so it is
// neither counted nor checked. An argument perl rejects is a Hole, which is
// counted as one argument and typed Any, so it satisfies every slot.
func (a *analysis) call(n *lower.Call) types.Type {
	if n.Handle != nil {
		a.walk(n.Handle)
	}
	if n.Fn != nil {
		a.walk(n.Fn)
	}
	name := coreName(n.Name)
	if name == "undef" {
		// undef denotes the undefined value whatever it is given: CORE.tpm
		// declares no return type for it, so it is typed here.
		a.args(n.Args, nil)
		return types.Undef
	}
	sigs := parse.CoreBuiltin(name)
	if n.Fn != nil || n.Ampersand || overridden(n) || len(sigs) == 0 {
		if n.Fn == nil && !n.Ampersand && len(n.Signatures) > 0 {
			return a.declaredCall(n)
		}
		a.args(n.Args, nil)
		return a.userCall(n)
	}
	a.args(n.Args, sigs)

	if least := builtinMinArity(sigs); len(n.Args) < least {
		start, end, _ := a.pos.Span(n.ID)
		a.diags = append(a.diags, Diagnostic{
			StartByte: uint32(start),
			EndByte:   uint32(end),
			Severity:  Error,
			Code:      CodeArityMismatch,
			Message:   arityMessage(name, least, len(n.Args)),
		})
		return a.callReturns(n, sigs)
	}
	opened := openedVar(name, n.Args)
	for i, arg := range n.Args {
		if i == 0 && opened != nil && types.IsSubtype(a.typeOf(arg), types.Str) {
			continue
		}
		a.checkArg(name, i, arg, builtinArgType(sigs, i), slotContext(sigs, i) == types.ListCtx, CodeTypeMismatch)
	}
	if opened != nil {
		a.st.UpdateType("$"+opened.Name, types.Scalar)
	}

	if t, ok := a.contextualReturn(n); ok {
		return t
	}
	return a.callReturns(n, sigs)
}

// declaredCall types a call to a sub a declaration file types, from the
// candidates the parse resolved it to, n.Signatures: RFC 0001 "Call
// sites". The candidate is selected by the call's arity, its arguments'
// types and its context, and the call is its return type, or the join of
// the candidates the call site cannot decide between. The parse leaves a
// call to a sub with no prototype as perl compiles it, so one whose arity
// no candidate takes is reported here, and has no type. An argument of
// the selected candidate's that does not fit its parameter is coerced to
// it, so is reported as a coercion mismatch.
//
// An array or a hash argument flattens into the call's list, as many
// arguments as it holds, so a call passing one has at least its other
// arguments, and no arity to select by: see flattenedSelect. One in a
// position the prototype perl applies to the call gives a slot, as the
// parse resolved it, is one argument: `\@` passes the one container it is.
func (a *analysis) declaredCall(n *lower.Call) types.Type {
	sigs := n.Signatures
	a.args(n.Args, sigs)
	args := a.argTypes(n.Args)
	fixed := 0
	for i, t := range args {
		if t&(types.Array|types.Hash) == 0 || parse.SlotAt(n.Prototype, i) {
			fixed++
		}
	}
	flattens := fixed < len(args)
	var sel types.Selection
	if flattens {
		sel = flattenedSelect(sigs, fixed, a.evalCtx[n.ID])
	} else {
		sel = types.Select(sigs, args, a.evalCtx[n.ID])
	}
	switch {
	case sel.Outcome == types.Failed && (flattens || !types.Accepts(sigs, len(n.Args))):
		count := strconv.Itoa(len(n.Args))
		if flattens {
			count = "at least " + strconv.Itoa(fixed)
		}
		start, end, _ := a.pos.Span(n.ID)
		a.diags = append(a.diags, Diagnostic{
			StartByte: uint32(start),
			EndByte:   uint32(end),
			Severity:  Error,
			Code:      CodeArityMismatch,
			Message:   "call to " + n.Name + ": no declared candidate takes " + count + " argument(s)",
		})
	case sel.Outcome == types.Selected:
		chosen := sigs[sel.Candidate : sel.Candidate+1]
		for i, arg := range n.Args {
			a.checkArg(n.Name, i, arg, builtinArgType(chosen, i), slotContext(chosen, i) == types.ListCtx, CodeCoercionMismatch)
		}
	}
	a.markPossibleUndef(n, sel)
	return sel.Returns
}

// markPossibleUndef marks n's own id, for possibleUndef (below) to read,
// whenever sel's own resolved return (whether types.Select Selected one
// candidate or joined several, Undecided) carries the Undef bit beside
// some other bit (perigrin, 2026-10-20): "An Undef from a declared
// return is possible-undef unless the call's arguments prove that the
// Undef-returning candidate is selected." A union return proves nothing
// of the kind on its own -- Scalar::Util::blessed's one declared
// candidate, `(Any $ref) Str|Undef`, answers this way for every call
// regardless of argument, and Scalar::Util::reftype's and refaddr's
// widest, catch-all candidate the same -- so the call's own arguments,
// not just its declaration, would have to rule out every other outcome
// before this counts as proven.
//
// sel.Returns equal to Undef outright, with no other bit at all, is the
// one shape that does: List::Util::sum's own `()` candidate, the only
// one admitting zero arguments, answers nothing but Undef, so a call
// the arguments force there -- sum() -- has no other possibility to be
// silent about, and is not marked.
func (a *analysis) markPossibleUndef(n *lower.Call, sel types.Selection) {
	if sel.Returns&types.Undef != 0 && sel.Returns != types.Undef {
		a.undecidedUndef[n.ID] = true
	}
}

// possibleUndef reports whether n's Undef, in a slot expecting expected,
// is accepted by default under WarnPossibleUndef (perigrin, 2026-10-20):
// n is a declared call markPossibleUndef marked, or a variable a marked
// call's result was assigned to (Symbol.PossibleUndef, carried there by
// narrow in ops.go), its Undef one of several outcomes the call's own
// declaration states rather than the call's arguments proving its
// return could be nothing else, and removing that one bit would
// satisfy expected on its own -- a mismatch WarnPossibleUndef has no
// bearing on, a wrong type entirely, still reports.
func (a *analysis) possibleUndef(n lower.Node, actual, expected types.Type) bool {
	if a.opts.WarnPossibleUndef || actual&types.Undef == 0 || !a.markedPossibleUndef(n) {
		return false
	}
	rest := numified(finite(actual&^types.Undef), expected)
	return a.opts.satisfies(rest, expected) || a.opts.handleSatisfies(rest, expected)
}

// markedPossibleUndef reports whether n is itself a declared call
// markPossibleUndef marked, or a variable read whose symbol carries the
// same mark forward from the call its value was last assigned from.
func (a *analysis) markedPossibleUndef(n lower.Node) bool {
	if a.undecidedUndef[n.Meta().ID] {
		return true
	}
	v, ok := n.(*lower.Var)
	if !ok {
		return false
	}
	sym, ok := a.st.Lookup(v.Sigil + v.Name)
	return ok && sym.PossibleUndef
}

// flattenedSelect is types.Select for a call in context ctx with at least
// least arguments, an array or a hash among them holding the rest: the
// join of what each count the call may have selects, its arguments' types
// unknown, since a flattened list moves those after it. It fails only when
// no count does. A slurpy takes any count past its parameters, so one past
// the longest, or least if it is more, stands for them all.
func flattenedSelect(sigs []types.Signature, least int, ctx types.Context) types.Selection {
	most := 0
	for _, s := range sigs {
		most = max(most, len(s.Params))
	}
	joined, taken := types.Unknown, false
	for count := least; count <= max(least, most+1); count++ {
		if sel := types.Select(sigs, make([]types.Type, count), ctx); sel.Outcome != types.Failed {
			joined, taken = types.Join(joined, sel.Returns), true
		}
	}
	if !taken {
		return types.Selection{Outcome: types.Failed, Candidate: -1}
	}
	return types.Selection{Outcome: types.Undecided, Candidate: -1, Returns: joined}
}

// openedVar is the scalar variable an open opens a handle in, its first
// argument, or nil. open puts the handle in it, `open(my $fh, ...)`, or
// opens the handle a string in it names, which perl takes without strict
// refs: perl 5.42 -w is silent on Devel::SelfStubber's `$fh =
// "${module}::DATA"; open($fh, '<', $f); fileno($fh); close($fh)`. So
// its string is not checked as a handle, and it holds a Scalar after.
func openedVar(name string, args []lower.Node) *lower.Var {
	if name != "open" || len(args) == 0 {
		return nil
	}
	if v, ok := args[0].(*lower.Var); ok && v.Sigil == "$" {
		return v
	}
	return nil
}

// coreName is the builtin a call names: `CORE::push` names push, as
// perl reads it whatever a `use` list overrides. `CORE::GLOBAL::die` is
// a sub of its own, and keeps its name.
func coreName(name string) string {
	if rest, ok := strings.CutPrefix(name, "CORE::"); ok && !strings.HasPrefix(rest, "GLOBAL::") {
		return rest
	}
	return name
}

// strongKeywords are the builtins no sub overrides: measured on 5.42.0 by
// compiling `use subs 'NAME'; sub NAME { 1 } sub t { NAME(...) }` for every
// CORE.tpm builtin, and finding no entersub in t. say is one under the
// say feature, and a plain word without it.
var strongKeywords = map[string]bool{
	"defined": true, "delete": true, "eval": true, "exists": true,
	"goto": true, "grep": true, "map": true, "pos": true, "print": true,
	"printf": true, "prototype": true, "return": true, "say": true,
	"scalar": true, "shift": true, "sort": true, "split": true,
	"study": true, "undef": true,
}

// overridden reports whether c calls a builtin the parse resolved to a
// sub, as parse.Node.Override says; a strong keyword is never overridden.
func overridden(c *lower.Call) bool {
	return c.Override && !strongKeywords[c.Name]
}

// args walks a call's arguments, each in the context its parameter slot in
// sigs imposes. A user sub's arguments, sigs nil, are walked as they are.
func (a *analysis) args(args []lower.Node, sigs []types.Signature) {
	for i, arg := range args {
		if ctx := slotContext(sigs, i); ctx != types.UnknownCtx {
			a.evalCtx[arg.Meta().ID] = ctx
		} else if _, ok := arg.(*lower.Match); ok && sigs != nil {
			// A `$` slot takes a match in scalar context, where it is a
			// Boolean: perl 5.42 -w is silent on `chr $u =~ /re/`, which
			// is chr($u =~ /re/).
			a.evalCtx[arg.Meta().ID] = types.ScalarCtx
		}
		a.walk(arg)
	}
}

// slotContext is the context the i-th argument of a builtin whose CORE.tpm
// candidates are sigs is evaluated in: list context when every candidate
// takes it into a slurpy, UnknownCtx otherwise.
//
// A `$` slot is left UnknownCtx, not scalar context, so an aggregate in
// one reaches checkArg as it is, which counts it as counted says or
// reports it where perl does, a ternary's arm included: perl 5.42 -w,
// `length(1 ? @a : "x")` warns "length() used on @a (did you mean
// "scalar(@a)"?)".
func slotContext(sigs []types.Signature, i int) types.Context {
	ctx := types.UnknownCtx
	for _, s := range sigs {
		if len(s.Params) == 0 {
			continue
		}
		p := s.Params[min(i, len(s.Params)-1)]
		if i >= len(s.Params) && !p.Slurpy() {
			continue
		}
		if !p.Slurpy() {
			return types.UnknownCtx
		}
		ctx = types.ListCtx
	}
	return ctx
}

// userCall types a call to a sub that is not a builtin: a fully-qualified
// name the project index resolves, else a sub of this file whose return
// type is already inferred. A sub declared after the call has none yet.
func (a *analysis) userCall(n *lower.Call) types.Type {
	if n.Fn != nil || n.Name == "" {
		return types.Unknown
	}
	if a.idx != nil {
		if sep := strings.LastIndex(n.Name, "::"); sep >= 0 {
			if sym, ok := a.idx.LookupSymbol(n.Name[:sep], n.Name[sep+2:]); ok && sym.ReturnType != types.Unknown {
				return sym.ReturnType
			}
		}
	}
	if sym, ok := a.st.Lookup(n.Name); ok && sym.ReturnType != types.Unknown {
		return sym.ReturnType
	}
	return types.Unknown
}

// checkArg is the argument check: an argument whose type does not
// satisfy what the builtin's slot expects is a type mismatch. scalar's
// argument is counted first, as builtinArgActual says. One that does not
// satisfy the slot as it is may still as finite and numified judge it,
// one in a slurpy slot as flattened says, and one in a `$` slot as
// scalarValue or counted says, and last as lossyCoercion says. Flattening is a second
// chance, so a Hole's Any, which satisfies the slot as it is, is never
// rewritten; one that fails both is reported as flattened.
func (a *analysis) checkArg(name string, i int, arg lower.Node, expected types.Type, slurpy bool, code string) {
	actual := builtinArgActual(name, a.typeOf(arg))
	if a.argSatisfies(actual, expected) || a.argSatisfies(a.succeeded(arg, actual), expected) {
		return
	}
	if slurpy {
		actual = a.flattened(arg, actual)
		if a.argSatisfies(actual, expected) {
			return
		}
	} else if a.argSatisfies(scalarValue(arg, actual), expected) || a.argSatisfies(counted(name, actual), expected) {
		return
	}
	if a.lossyCoercion(arg, actual, expected) || a.silenced(actual, expected) || a.possibleUndef(arg, actual, expected) {
		return
	}
	start, end, _ := a.pos.Span(arg.Meta().ID)
	a.diags = append(a.diags, Diagnostic{
		StartByte:  uint32(start),
		EndByte:    uint32(end),
		Severity:   argumentSeverity(actual, expected),
		Code:       code,
		Message:    typeMismatchMessage(name, i, expected, actual),
		Suggestion: SuggestGuard(argVarName(arg), actual, expected),
	})
}

// argSatisfies reports whether a builtin argument of type actual may fill
// a slot expecting expected: as it is, as a handle, or as finite and
// numified judge it. An Unknown one is not checked unless Strict.
func (a *analysis) argSatisfies(actual, expected types.Type) bool {
	return a.opts.skipUnknownOperand(actual) || a.opts.satisfies(actual, expected) || a.opts.handleSatisfies(actual, expected) ||
		a.opts.satisfies(numified(finite(actual), expected), expected)
}

// succeeded is actual, the type of n, as a slot judges it when n is a call
// to a builtin: without the Undef the builtin returns only when it fails,
// as failureUndef says. fileno is Int|Undef, undef for a closed handle,
// and perl 5.42 -w is silent on `fileno(STDOUT) . ">"`; read, getsockopt
// and prototype are the same. A builtin the parse resolved to a sub is a
// user sub, as overridden says. `scalar` hands on its argument's result.
// A type that is only Undef, `undef()`'s, is kept, and so is anything
// that is no builtin's result. It is a second chance, never a rewrite: a Scalar or a Hole's
// Any satisfies a slot as it is.
//
// A call to a sub of this file whose every returned value is a builtin's
// result is judged so too: perl 5.42 -w is silent on `sub xfileno {
// return fileno $_[0] } xfileno(\*STDOUT) == 1`.
//
// Only a call's own result is judged so; a variable that holds one,
// `my $fd = fileno(STDOUT); $fd . ">"`, is still judged with its Undef.
func (a *analysis) succeeded(n lower.Node, actual types.Type) types.Type {
	c, ok := n.(*lower.Call)
	if !ok || c.Fn != nil {
		return actual
	}
	user := c.Ampersand || overridden(c) || len(parse.CoreBuiltin(c.Name)) == 0
	if user && !a.returnsBuiltinResult(a.subs[c.Name]) {
		return actual
	}
	if c.Name == "scalar" && len(c.Args) == 1 {
		return a.succeeded(c.Args[0], actual)
	}
	if !user && !failureUndef[c.Name] {
		return actual
	}
	if actual&^types.Undef == 0 {
		return actual
	}
	return actual &^ types.Undef
}

// failureUndef are the builtins CORE.tpm types with an Undef they return
// only when they fail: a closed handle's fileno, a failed read or open.
// Those that return undef in normal use are not among them: perl 5.42 -w
// warns "Use of uninitialized value in concatenation" on `getc(STDIN) .
// "x"` at end of file and on `each(%h) . "y"` for an empty %h, and
// getlogin, wantarray and sort give undef without any failure.
var failureUndef = map[string]bool{
	"accept": true, "alarm": true, "bind": true, "binmode": true, "closedir": true,
	"connect": true, "dbmopen": true, "fcntl": true, "fileno": true, "fork": true,
	"getpeername": true, "getsockname": true, "getsockopt": true, "gmtime": true,
	"ioctl": true, "listen": true, "localtime": true, "msgctl": true, "msgget": true,
	"open": true, "opendir": true, "pipe": true, "print": true, "printf": true,
	"prototype": true, "read": true, "readlink": true, "readpipe": true, "recv": true,
	"rewinddir": true, "say": true, "seekdir": true, "semctl": true, "semget": true,
	"setsockopt": true, "shmctl": true, "shmget": true, "shutdown": true,
	"socket": true, "socketpair": true, "sysopen": true, "sysread": true,
	"sysseek": true, "syswrite": true, "telldir": true, "truncate": true,
}

// returnsBuiltinResult reports whether every value sub returns, by an
// explicit return or as its body's last statement, is a call to a
// builtin failureUndef names, or a capture its own successful match
// guards, as ownCapture says. A bare return gives the sub's own Undef, and so does a body
// whose last statement is not such a call, so neither is one. A sub
// nested in sub returns from itself, not from sub.
func (a *analysis) returnsBuiltinResult(sub *lower.AnonSub) bool {
	if sub == nil {
		return false
	}
	body, ok := sub.Body.(*lower.Block)
	if !ok || len(body.Stmts) == 0 {
		return false
	}
	builtin := func(n lower.Node) bool {
		if l, ok := n.(*lower.List); ok && len(l.Elems) > 0 {
			n = l.Elems[len(l.Elems)-1]
		}
		c, ok := n.(*lower.Call)
		return ok && c.Fn == nil && !c.Ampersand && failureUndef[c.Name]
	}
	all := true
	returnsOf(body, func(r *lower.LoopEx, guard *lower.IfStmt) {
		all = all && r.Label != nil && (builtin(r.Label) || a.ownCapture(r.Label, guard))
	})
	last, ok := body.Stmts[len(body.Stmts)-1].(*lower.ExprStmt)
	if !ok || last.X == nil {
		return false
	}
	if r, ok := last.X.(*lower.LoopEx); ok && r.Op == "return" {
		return all
	}
	return all && builtin(last.X)
}

// ownCapture reports whether n is a capture, $1 or the like, that the
// match guarding stmt defines: n is in stmt's successful branch, the
// match has its group, and it is the match n reads, as ownMatch says.
// perl 5.42 -w is silent on `return $1 if $_[0] =~ /(\d+)/` handed to
// `==`, and warns on `$2` there, or under `!~`.
func (a *analysis) ownCapture(n lower.Node, stmt *lower.IfStmt) bool {
	if l, ok := n.(*lower.List); ok && len(l.Elems) > 0 {
		n = l.Elems[len(l.Elems)-1]
	}
	v, ok := n.(*lower.Var)
	if !ok || v.Sigil != "$" {
		return false
	}
	group, err := strconv.Atoi(v.Name)
	start, _, _ := a.pos.Span(v.ID)
	g := a.ownMatch(start, stmt)
	return err == nil && group >= 1 && g != nil &&
		group <= len(captureGroups(quotePattern(a.pos.Quote(g.Meta().ID))))
}

// scalarValue is actual, the type of arg, as a `$` slot judges it when arg
// is a call or an assignment: perl evaluates either in scalar context, and
// it yields one scalar. So a List there -- caller's and readline's, whose
// one candidate covers both contexts, or a scalar assignment's -- is a
// Scalar. An array or a hash is neither and keeps its type:
// `defined(@a)` is perl's compile error, and `length(@a)` warns.
func scalarValue(arg lower.Node, actual types.Type) types.Type {
	switch arg.(type) {
	case *lower.Call, *lower.Assign:
		if actual&(types.Array|types.Hash) != 0 {
			return actual&^(types.Array|types.Hash|types.Void) | types.Scalar
		}
	}
	return actual
}

// aggregateRefused are the builtins perl refuses or warns on an array or
// a hash in their `$` slot, measured on 5.42.0 with -c over every builtin
// CORE.tpm gives a `$` slot: `defined(@a)` is "Can't use
// 'defined(@array)'", `exists(@a)` and `delete(@a)` are "not a HASH or
// ARRAY element", and `length(@a)` warns "length() used on @a".
var aggregateRefused = map[string]bool{"defined": true, "exists": true, "delete": true, "length": true}

// counted is actual, the type of an argument in a `$` slot of the builtin
// name, as perl evaluates it there: in scalar context, where an array or a
// hash is its count. perl 5.42 -w is silent on `int(rand(@m))` and `my
// @e; srand(@e)`, "0 but true". An aggregate perl refuses there, as
// aggregateRefused says, keeps its type. Elsewhere actual is returned as
// it is.
func counted(name string, actual types.Type) types.Type {
	if (actual == types.Array || actual == types.Hash) && !aggregateRefused[name] {
		if narrowed, ok := types.NarrowByContext(actual, types.ScalarCtx); ok {
			return narrowed
		}
	}
	return actual
}

// flattened is what an argument of type actual hands a slurpy slot: an
// array or a hash flattens into its elements, so the slot checks those --
// the recorded element type, else any Scalar -- and a hash's keys, which
// are strings. `print @parts` prints each element. A List's Void is the
// empty list, which hands the slot no element at all: perl 5.42 -w is
// silent on `print readline(*DATA)` at end of file.
func (a *analysis) flattened(arg lower.Node, actual types.Type) types.Type {
	agg := actual & (types.Array | types.Hash)
	if agg == 0 {
		return actual
	}
	elem := a.aggElemType(arg)
	if elem == types.Unknown {
		elem = types.Scalar
	}
	if agg&types.Hash != 0 {
		elem |= types.Str
	}
	return actual&^(types.Array|types.Hash|types.Void) | elem
}

// handleSatisfies reports whether actual may be the handle a filehandle
// slot -- one whose expected type holds Glob -- takes. A handle is a Glob,
// a GlobRef (`open my $fh` gives one), or an IO handle object such as an
// IO::File, and a Scalar may hold either of the last two. perl 5.42 -w is
// silent on a lexical handle and an IO::File in seek, read, stat and
// close. A value that cannot be a handle, a HashRef or an Int, is not one.
func (o Options) handleSatisfies(actual, expected types.Type) bool {
	if expected&types.Glob == 0 {
		return false
	}
	handle := expected | types.FileHandle | types.Object
	return o.satisfies(actual, handle) || o.satisfies(actual, handle&types.Scalar)
}

// argVarName is the sigil and name of an argument that is a scalar, array
// or hash variable, "" for anything else.
func argVarName(n lower.Node) string {
	if v, ok := n.(*lower.Var); ok {
		if _, named := sigilTypes[v.Sigil]; named {
			return v.Sigil + v.Name
		}
	}
	return ""
}

// callReturns is a builtin call's type in the context around it, from the
// candidates its arguments' shapes take: a slice is `@`-shaped, anything
// else `$`-shaped.
func (a *analysis) callReturns(n *lower.Call, sigs []types.Signature) types.Type {
	shapes := ""
	for _, arg := range n.Args {
		if _, ok := arg.(*lower.Slice); ok {
			shapes += "@"
		} else {
			shapes += "$"
		}
	}
	return shapedCallReturns(sigs, a.argTypes(n.Args), shapes, a.evalCtx[n.ID])
}

// argTypes is the type of each of a call's arguments.
func (a *analysis) argTypes(args []lower.Node) []types.Type {
	ts := make([]types.Type, len(args))
	for i, arg := range args {
		ts[i] = a.typeOf(arg)
	}
	return ts
}

// contextualReturn types the builtins whose
// type follows what they are given. sprintf's follows its format, pop's
// and shift's the array's elements, and int's, abs's and scalar's their
// argument, as argDirectedReturn says.
func (a *analysis) contextualReturn(n *lower.Call) (types.Type, bool) {
	switch n.Name {
	case "sprintf":
		if len(n.Args) >= 1 {
			if s, ok := n.Args[0].(*lower.StrLit); ok && !s.Backtick {
				return numericFormatType(s.Value)
			}
			return types.Unknown, false
		}
	case "pop", "shift":
		if len(n.Args) == 1 {
			if v, ok := n.Args[0].(*lower.Var); ok && v.Sigil == "@" {
				if a.elemOf("@"+v.Name) != types.Unknown {
					return a.element("@" + v.Name), true
				}
			}
			return types.Unknown, false
		}
	}
	arg0 := types.Unknown
	if len(n.Args) == 1 {
		arg0 = a.typeOf(n.Args[0])
	}
	return argDirectedReturn(n.Name, len(n.Args), arg0)
}

// callElemType is, for a Call, the element type
// of the list a builtin returns. A hash's keys are strings; undef is one
// undefined value; values, splice and reverse hand back elements of the
// aggregate they read. Any other
// call's elements are not known.
func (a *analysis) callElemType(n *lower.Call) types.Type {
	switch n.Name {
	case "keys":
		return types.Str
	case "undef":
		return types.Undef
	case "values", "splice", "reverse":
		for i := len(n.Args) - 1; i >= 0; i-- {
			if t := a.aggElemType(n.Args[i]); t != types.Unknown {
				return t
			}
		}
	}
	return types.Unknown
}

// aggElemType is the recorded element type
// of an array or hash, or the join of a list's elements.
func (a *analysis) aggElemType(n lower.Node) types.Type {
	switch n := n.(type) {
	case *lower.Var:
		if n.Sigil == "@" || n.Sigil == "%" {
			return a.elemOf(n.Sigil + n.Name)
		}
	case *lower.List:
		return a.joinTypes(n.Elems)
	}
	return types.Unknown
}

// methodCall types a method call. `Class->new` is an Object
// of Class, recorded under the call's id for the assignment that takes it.
// Another class method, and a method on a variable whose class is
// recorded, is resolved through the project index. Anything else -- a
// dynamic method name, an invocant of no known class -- is not typed.
//
// `new Foo(...)`, the indirect form, has Foo as its Invocant, so it is the
// same constructor.
func (a *analysis) methodCall(n *lower.MethodCall) types.Type {
	a.walkChildren(n)
	if n.Dynamic != nil || n.Name == "" {
		return types.Unknown
	}
	class := ""
	switch inv := n.Invocant.(type) {
	case *lower.Bareword:
		if n.Name == "new" {
			a.classTypes[n.ID] = inv.Name
			return types.Object
		}
		class = inv.Name
	case *lower.Var:
		if inv.Sigil != "$" {
			return types.Unknown
		}
		if sym, ok := a.st.Lookup("$" + inv.Name); ok {
			class = sym.ClassType
		}
	}
	if class == "" || a.idx == nil {
		return types.Unknown
	}
	if sym, ok := a.idx.LookupSymbol(class, n.Name); ok && sym.ReturnType != types.Unknown {
		return sym.ReturnType
	}
	return types.Unknown
}

// subParams is a sub's parameters: a signature's mandatory
// scalars, else the scalars `my (...) = @_` unpacks as the body's first
// statement, else the variables of the `my $v = shift` statements the body
// opens with, else $_[0] when the body reads it. nil when the sub has none
// of them.
func subParams(s *lower.AnonSub) []string {
	if len(s.Sig) > 0 {
		var params []string
		for _, p := range s.Sig {
			if v, ok := p.(*lower.Var); ok && v.Sigil == "$" {
				params = append(params, "$"+v.Name)
			}
		}
		return params
	}
	body, ok := s.Body.(*lower.Block)
	if !ok || len(body.Stmts) == 0 {
		return nil
	}
	if params := unpacksArgs(body.Stmts[0]); len(params) > 0 {
		return params
	}
	var params []string
	for _, stmt := range body.Stmts {
		name := shiftParam(stmt)
		if name == "" {
			break
		}
		params = append(params, name)
	}
	if len(params) == 0 && readsArg0(body) {
		return []string{"$_[0]"}
	}
	return params
}

// unpacksArgs is the scalars stmt assigns from @_, `my ($x, $y) = @_` or
// `($x, $y) = @_`, in order; nil when stmt is no such assignment.
func unpacksArgs(stmt lower.Node) []string {
	es, ok := stmt.(*lower.ExprStmt)
	if !ok {
		return nil
	}
	var lhs []lower.Node
	var rhs lower.Node
	switch x := es.X.(type) {
	case *lower.Decl:
		lhs, rhs = x.Vars, x.Init
	case *lower.Assign:
		lhs, rhs = []lower.Node{x.L}, x.R
	default:
		return nil
	}
	if v, ok := rhs.(*lower.Var); !ok || v.Sigil != "@" || v.Name != "_" {
		return nil
	}
	var params []string
	var collect func(n lower.Node)
	collect = func(n lower.Node) {
		if v, ok := n.(*lower.Var); ok {
			if v.Sigil == "$" {
				params = append(params, "$"+v.Name)
			}
			return
		}
		for _, c := range lower.Children(n) {
			collect(c)
		}
	}
	for _, n := range lhs {
		collect(n)
	}
	return params
}

// shiftParam is the scalar stmt declares from a shift, `my $v = shift` or
// `my $v = shift @_`, "" when stmt is not one.
func shiftParam(stmt lower.Node) string {
	es, ok := stmt.(*lower.ExprStmt)
	if !ok {
		return ""
	}
	d, ok := es.X.(*lower.Decl)
	if !ok || d.Kind != lower.DeclMy {
		return ""
	}
	c, ok := d.Init.(*lower.Call)
	if !ok || c.Name != "shift" {
		return ""
	}
	vars := declaredVars(d)
	if len(vars) == 0 || vars[len(vars)-1].Sigil != "$" {
		return ""
	}
	return "$" + vars[len(vars)-1].Name
}

// paramTypes infers parameter types from usage: each parameter's type
// is what the operators and builtins it is passed to expect of it, every
// constraint intersected. A parameter nothing constrains is Unknown.
func (a *analysis) paramTypes(body lower.Node, params []string) []types.Type {
	slot := make(map[string]int, len(params))
	constraints := make([]types.Type, len(params))
	for i, p := range params {
		slot[p] = i
		constraints[i] = types.Any
	}
	constrain := func(n lower.Node, expected types.Type) {
		v, ok := n.(*lower.Var)
		if !ok || expected == types.Any {
			return
		}
		if i, ok := slot[v.Sigil+v.Name]; ok {
			constraints[i] &= expected
		}
	}
	var walk func(n lower.Node)
	walk = func(n lower.Node) {
		if n == nil {
			return
		}
		switch n := n.(type) {
		case *lower.Binary:
			sigs := shapedOperatorCandidates(n.Op, "infix", operandShape(n.L)+operandShape(n.R), a.evalCtx[n.ID])
			if len(sigs) > 0 {
				constrain(n.L, builtinArgType(sigs, 0))
				constrain(n.R, builtinArgType(sigs, 1))
			}
		case *lower.Call:
			if sigs := parse.CoreBuiltin(n.Name); n.Fn == nil && len(sigs) > 0 {
				for i, arg := range n.Args {
					constrain(arg, builtinArgType(sigs, i))
				}
			}
		}
		for _, c := range lower.Children(n) {
			walk(c)
		}
	}
	walk(body)
	for i := range constraints {
		if constraints[i] == types.Any {
			constraints[i] = types.Unknown
		}
	}
	return constraints
}

// recordParamTypes is the parameter half of a sub declaration's walk: a
// named sub's parameters are typed by how its
// body uses them, and recorded only when one of them is known.
func (a *analysis) recordParamTypes(d *lower.Decl) {
	name := subName(d)
	body, ok := d.Init.(*lower.AnonSub)
	if name == "" || !ok {
		return
	}
	params := subParams(body)
	if len(params) == 0 {
		return
	}
	pts := a.paramTypes(body.Body, params)
	for _, pt := range pts {
		if pt != types.Unknown {
			a.st.UpdateParamTypes(name, params, pts)
			return
		}
	}
}
