// ABOUTME: The call rules of AnalyzeLowered: one Call arm for every builtin and user sub, MethodCall and the class it records, and parameter types.
// ABOUTME: Each is the legacy rule rewritten against the lowered kinds: the filehandle is Call.Handle, never one of the Args.

package infer

import (
	"strings"
	"unicode"

	"tamarou.com/pvm/internal/lower"
	"tamarou.com/pvm/internal/parse"
	"tamarou.com/pvm/internal/types"
)

// call is the legacy inferFunctionCallType, inferFunc1opCallType and
// inferFunc0opCallType in one: a builtin's CORE.pmt candidates check the
// arity and each argument's type and give the call's type; a sub of this
// file gives its inferred return type, and a fully-qualified one the
// project index's. A call through a reference is not typed. `&name` never
// calls CORE, so it is a user sub's call whatever its name, and so is a
// builtin a `use` list overrides.
//
// The filehandle of `print $fh LIST` is Handle, not an argument, so it is
// neither counted nor checked. An argument perl rejects is a Hole, which is
// counted as one argument and typed Any, so it satisfies every slot.
func (a *loweredAnalysis) call(n *lower.Call) types.Type {
	if n.Handle != nil {
		a.walk(n.Handle)
	}
	if n.Fn != nil {
		a.walk(n.Fn)
	}
	if n.Name == "undef" {
		// undef denotes the undefined value whatever it is given: CORE.pmt
		// declares no return type for it, and the legacy path typed it so.
		a.args(n.Args, nil)
		return types.Undef
	}
	sigs := parse.CoreBuiltin(n.Name)
	if n.Fn != nil || n.Ampersand || a.overrides[n.Name] || len(sigs) == 0 {
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
			Message:   arityMessage(n.Name, least, len(n.Args)),
		})
		return a.callReturns(n, sigs)
	}
	for i, arg := range n.Args {
		a.checkArg(n.Name, i, arg, builtinArgType(sigs, i), slotContext(sigs, i) == types.ListCtx)
	}

	if t, ok := a.contextualReturn(n); ok {
		return t
	}
	return a.callReturns(n, sigs)
}

// strongKeywords are the builtins no sub overrides: measured on 5.42.0 by
// compiling `use subs 'NAME'; sub NAME { 1 } sub t { NAME(...) }` for every
// CORE.pmt builtin, and finding no entersub in t. say is one under the
// say feature, and a plain word without it.
var strongKeywords = map[string]bool{
	"defined": true, "delete": true, "eval": true, "exists": true,
	"goto": true, "grep": true, "map": true, "pos": true, "print": true,
	"printf": true, "prototype": true, "return": true, "say": true,
	"scalar": true, "shift": true, "sort": true, "split": true,
	"study": true, "undef": true,
}

// overrides is each builtin f's `use` lists override: the names `use subs`
// predeclares, and those a module's import list names, which it exports
// as subs. perl 5.42 -w, `use subs "rename"; sub rename { "fake" } print
// rename()` prints "fake", and `use POSIX qw(localtime); localtime(1, 2)`
// compiles. A pragma's list is its arguments, `use feature "say"`, and
// names no sub; a strong keyword is never overridden.
//
// The file is read as one scope: a `use` anywhere overrides the name in
// every package and before the `use`, where perl would not.
func overrides(f *lower.File) map[string]bool {
	names := map[string]bool{}
	var add func(n lower.Node)
	add = func(n lower.Node) {
		switch n := n.(type) {
		case *lower.StrLit:
			if !n.Backtick {
				names[n.Value] = true
			}
		case *lower.QwList:
			for _, w := range n.Words {
				names[w] = true
			}
		case *lower.List:
			for _, e := range n.Elems {
				add(e)
			}
		}
	}
	var walk func(n lower.Node)
	walk = func(n lower.Node) {
		if u, ok := n.(*lower.UseStmt); ok && u.Op == "use" && u.Args != nil &&
			(u.Module == "subs" || u.Module != "" && unicode.IsUpper(rune(u.Module[0]))) {
			add(u.Args)
		}
		for _, c := range lower.Children(n) {
			walk(c)
		}
	}
	walk(f)
	for name := range names {
		if strongKeywords[name] || len(parse.CoreBuiltin(name)) == 0 {
			delete(names, name)
		}
	}
	return names
}

// args walks a call's arguments, each in the context its parameter slot in
// sigs imposes. A user sub's arguments, sigs nil, are walked as they are.
func (a *loweredAnalysis) args(args []lower.Node, sigs []types.Signature) {
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

// slotContext is the context the i-th argument of a builtin whose CORE.pmt
// candidates are sigs is evaluated in: list context when every candidate
// takes it into a slurpy, UnknownCtx otherwise.
//
// A `$` slot is left UnknownCtx, not scalar context, so an aggregate in
// one stays the mistake checkArg reports, a ternary's arm included: perl
// 5.42 -w, `length(1 ? @a : "x")` warns "length() used on @a (did you
// mean "scalar(@a)"?)".
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
func (a *loweredAnalysis) userCall(n *lower.Call) types.Type {
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

// checkArg is the legacy argument check: an argument whose type does not
// satisfy what the builtin's slot expects is a type mismatch. scalar's
// argument is counted first, as builtinArgActual says. One that does not
// satisfy the slot as it is may still as finite and numified judge it,
// one in a slurpy slot as flattened says, and one in a `$` slot as
// scalarValue says, and last as lossyCoercion says. Flattening is a second
// chance, so a Hole's Any, which satisfies the slot as it is, is never
// rewritten; one that fails both is reported as flattened.
func (a *loweredAnalysis) checkArg(name string, i int, arg lower.Node, expected types.Type, slurpy bool) {
	actual := builtinArgActual(name, a.typeOf(arg))
	if a.argSatisfies(actual, expected) || a.argSatisfies(a.succeeded(arg, actual), expected) {
		return
	}
	if slurpy {
		actual = a.flattened(arg, actual)
		if a.argSatisfies(actual, expected) {
			return
		}
	} else if a.argSatisfies(scalarValue(arg, actual), expected) {
		return
	}
	if a.lossyCoercion(arg, actual, expected) {
		return
	}
	start, end, _ := a.pos.Span(arg.Meta().ID)
	a.diags = append(a.diags, Diagnostic{
		StartByte:  uint32(start),
		EndByte:    uint32(end),
		Severity:   argumentSeverity(actual, expected),
		Code:       CodeTypeMismatch,
		Message:    typeMismatchMessage(name, i, expected, actual),
		Suggestion: SuggestGuard(argVarName(arg), actual, expected),
	})
}

// argSatisfies reports whether a builtin argument of type actual may fill
// a slot expecting expected: as it is, as a handle, or as finite and
// numified judge it. An Unknown one is not checked unless Strict.
func (a *loweredAnalysis) argSatisfies(actual, expected types.Type) bool {
	return a.opts.skipUnknownOperand(actual) || a.opts.satisfies(actual, expected) || a.opts.handleSatisfies(actual, expected) ||
		a.opts.satisfies(numified(finite(actual), expected), expected)
}

// succeeded is actual, the type of n, as a slot judges it when n is a call
// to a builtin: without the Undef the builtin returns only when it fails.
// fileno is Int|Undef, undef for a closed handle, and perl 5.42 -w is
// silent on `fileno(STDOUT) . ">"`; read, getsockopt and prototype are the
// same. `scalar` hands on its argument's result. A type that is only
// Undef, `undef()`'s, is kept, and so is anything that is no builtin's
// result. It is a second chance, never a rewrite: a Scalar or a Hole's
// Any satisfies a slot as it is.
//
// A call to a sub of this file whose every returned value is a builtin's
// result is judged so too: perl 5.42 -w is silent on `sub xfileno {
// return fileno $_[0] } xfileno(\*STDOUT) == 1`.
//
// Only a call's own result is judged so; a variable that holds one,
// `my $fd = fileno(STDOUT); $fd . ">"`, is still judged with its Undef.
func (a *loweredAnalysis) succeeded(n lower.Node, actual types.Type) types.Type {
	c, ok := n.(*lower.Call)
	if !ok || c.Fn != nil {
		return actual
	}
	user := c.Ampersand || len(parse.CoreBuiltin(c.Name)) == 0
	if user && !returnsBuiltinResult(a.subs[c.Name]) {
		return actual
	}
	if c.Name == "scalar" && len(c.Args) == 1 {
		return a.succeeded(c.Args[0], actual)
	}
	if actual&^types.Undef == 0 {
		return actual
	}
	return actual &^ types.Undef
}

// returnsBuiltinResult reports whether every value sub returns, by an
// explicit return or as its body's last statement, is a call to a
// builtin. A bare return gives the sub's own Undef, and so does a body
// whose last statement is not such a call, so neither is one. A sub
// nested in sub returns from itself, not from sub.
func returnsBuiltinResult(sub *lower.AnonSub) bool {
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
		return ok && c.Fn == nil && !c.Ampersand && len(parse.CoreBuiltin(c.Name)) > 0
	}
	all := true
	var returns func(n lower.Node)
	returns = func(n lower.Node) {
		if _, ok := n.(*lower.AnonSub); ok {
			return
		}
		if r, ok := n.(*lower.LoopEx); ok && r.Op == "return" {
			all = all && r.Label != nil && builtin(r.Label)
			return
		}
		for _, c := range lower.Children(n) {
			returns(c)
		}
	}
	returns(body)
	last, ok := body.Stmts[len(body.Stmts)-1].(*lower.ExprStmt)
	if !ok || last.X == nil {
		return false
	}
	if r, ok := last.X.(*lower.LoopEx); ok && r.Op == "return" {
		return all
	}
	return all && builtin(last.X)
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

// flattened is what an argument of type actual hands a slurpy slot: an
// array or a hash flattens into its elements, so the slot checks those --
// the recorded element type, else any Scalar -- and a hash's keys, which
// are strings. `print @parts` prints each element.
func (a *loweredAnalysis) flattened(arg lower.Node, actual types.Type) types.Type {
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
	return actual&^(types.Array|types.Hash) | elem
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
func (a *loweredAnalysis) callReturns(n *lower.Call, sigs []types.Signature) types.Type {
	shapes := ""
	for _, arg := range n.Args {
		if _, ok := arg.(*lower.Slice); ok {
			shapes += "@"
		} else {
			shapes += "$"
		}
	}
	return shapedCallReturns(sigs, shapes, a.evalCtx[n.ID])
}

// contextualReturn is the legacy contextualReturnType: the builtins whose
// type follows what they are given. sprintf's follows its format, pop's
// and shift's the array's elements, and int's, abs's and scalar's their
// argument, as argDirectedReturn says.
func (a *loweredAnalysis) contextualReturn(n *lower.Call) (types.Type, bool) {
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

// callElemType is the legacy callElementType for a Call: the element type
// of the list a builtin returns. A hash's keys are strings; values, splice
// and reverse hand back elements of the aggregate they read. Any other
// call's elements are not known.
func (a *loweredAnalysis) callElemType(n *lower.Call) types.Type {
	switch n.Name {
	case "keys":
		return types.Str
	case "values", "splice", "reverse":
		for i := len(n.Args) - 1; i >= 0; i-- {
			if t := a.aggElemType(n.Args[i]); t != types.Unknown {
				return t
			}
		}
	}
	return types.Unknown
}

// aggElemType is the legacy aggregateElemType: the recorded element type
// of an array or hash, or the join of a list's elements.
func (a *loweredAnalysis) aggElemType(n lower.Node) types.Type {
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

// methodCall is the legacy inferMethodCallType. `Class->new` is an Object
// of Class, recorded under the call's id for the assignment that takes it.
// Another class method, and a method on a variable whose class is
// recorded, is resolved through the project index. Anything else -- a
// dynamic method name, an invocant of no known class -- is not typed.
//
// `new Foo(...)`, the indirect form, has Foo as its Invocant, so it is the
// same constructor.
func (a *loweredAnalysis) methodCall(n *lower.MethodCall) types.Type {
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

// subParams is the legacy ResolveSubParams: a signature's mandatory
// scalars, else the scalars `my (...) = @_` unpacks as the body's first
// statement, else the variables of the `my $v = shift` statements the body
// opens with. nil when the sub has none of them.
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

// paramTypes is the legacy inferParamTypesFromUsage: each parameter's type
// is what the operators and builtins it is passed to expect of it, every
// constraint intersected. A parameter nothing constrains is Unknown.
func (a *loweredAnalysis) paramTypes(body lower.Node, params []string) []types.Type {
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

// recordParamTypes is the parameter half of the legacy
// walkSubroutineDeclaration: a named sub's parameters are typed by how its
// body uses them, and recorded only when one of them is known.
func (a *loweredAnalysis) recordParamTypes(d *lower.Decl) {
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
