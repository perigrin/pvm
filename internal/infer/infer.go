// ABOUTME: The judgements PSC type inference shares across its rules: options, operand and argument checks,
// ABOUTME: builtin signature reads, capture groups, and the guard narrowing a conditional applies.

package infer

import (
	"fmt"
	"strconv"
	"strings"

	"tamarou.com/pvm/internal/parse"
	"tamarou.com/pvm/internal/types"
)

// Options controls how analysis reports what it could not determine.
type Options struct {
	// Strict makes an un-inferred value satisfy nothing rather than
	// everything. By default Unknown behaves as TypeScript's `any` — it
	// passes every check, so the checker goes quiet exactly where it knows
	// least. Under Strict it behaves as TypeScript's `unknown`: a value whose
	// type could not be determined must be narrowed before it is used.
	//
	// There is ONE Unknown, and it is a status rather than a diagnosis. A
	// value reaches it by two quite different routes:
	//
	//   the question was never asked — `unlink` has no entry in the builtin
	//   signature table, so nothing was consulted and nothing came back. That
	//   is a gap in PSC's knowledge and it is fixable by adding the entry.
	//
	//   the question was asked and has no answer — `my $line = <$T>` was
	//   analysed and Perl does not determine the type statically. No table
	//   will fix that; it is a property of the language.
	//
	// Inference can distinguish these AT THE POINT IT RUNS, because it knows
	// whether it consulted anything. The stored Unknown cannot: once written
	// it carries no record of the route, and every consumer sees the same
	// value. Strict mode operates on the stored status, so it reports both
	// alike — correctly, since for the purpose of "must this be narrowed
	// before use?" the answer is yes either way.
	//
	// Telling the two apart in REPORTING would mean recording the route at
	// inference time rather than splitting the type. That is worth doing —
	// it separates a PSC to-do list from an irreducible property of Perl —
	// and it is not what this flag does.
	Strict bool

	// WarnLossyCoercion reports a coercion perl makes silently but that
	// loses the value: a reference or an object where a Str or a number is
	// expected, whose string HASH(0x...) or address does not give the
	// referent back, and a Str of unknown content where a number is
	// expected. By default they are accepted, as perl 5.42 -w accepts
	// `print $hashref`, `$hashref + 1` and `"$]" < 5.016`. A string
	// literal known not to be numeric is reported either way, as perl warns
	// on `"abc" + 1`.
	WarnLossyCoercion bool
}

// satisfies reports whether actual satisfies required under opts.
//
// This is the single place strictness is applied, so the difference between
// the two modes stays one function rather than a flag threaded through every
// walk site.
func (o Options) satisfies(actual, required types.Type) bool {
	if o.Strict {
		return types.TypeSatisfiesStrict(actual, required)
	}
	return types.TypeSatisfies(actual, required)
}

// skipUnknownOperand reports whether an operand of the given type should be
// passed over without checking. Permissively an un-inferred operand is skipped
// to avoid false positives; strictly it is exactly what we want to report.
func (o Options) skipUnknownOperand(actual types.Type) bool {
	return actual == types.Unknown && !o.Strict
}

// operandMismatch is the judgement checkOperand makes of one operand
// of op, whose type is actual where the operator expects expected: the
// severity and message of the coercion mismatch to report, and whether
// there is one. side is "left" or "right".
func operandMismatch(opts Options, actual, expected types.Type, op, side string) (Severity, string, bool) {
	// A scalar-context operator imposes that context on its operand, and an
	// aggregate in scalar context is its element count: `@fields == 3` and
	// `while (@_ > 1)` are the ordinary idiom, not type errors. Apply the
	// context before judging the operand.
	//
	// This narrows ONLY the aggregate bits — a reference is one value in any
	// context and still cannot be a number, so `$hashref + 1` is unaffected.
	//
	// The test is whether the operator wants an AGGREGATE, not whether its
	// expected type intersects List: under the arity ordering Scalar <: List,
	// so List contains Num's bits and `expected & List` is non-zero for every
	// scalar operator. Only an operator whose expected type is List itself
	// takes an aggregate without imposing scalar context.
	// Narrow only a genuine aggregate. Applying this to a broad type such as
	// Any would strip its Array and Hash bits and hand back the remaining
	// union, which then satisfies nothing — turning a type that accepted
	// everything into one that accepts nothing.
	if expected != types.List && (actual == types.Array || actual == types.Hash || actual == types.List) {
		if narrowed, ok := types.NarrowByContext(actual, types.ScalarCtx); ok {
			actual = narrowed
		}
	}

	if opts.skipUnknownOperand(actual) {
		return 0, "", false
	}

	// Skip logical and assignment operators — they accept any type.
	if expected == types.Any {
		return 0, "", false
	}

	if opts.satisfies(actual, expected) {
		return 0, "", false
	}

	// Severity follows the COERCION, not the operator.
	//
	// If perl will convert the value and the conversion is defined, the code
	// runs and does something specific — `0 .. 3.7` is (0 1 2 3), `"ab" x 2.9`
	// is "abab", `$level / 8` really is 2.5 and `"\t" x 2.5` is two tabs. The
	// paper calls Num -> Int a coercion that truncates toward zero, not a
	// failure, and perl emits no warning of its own. Calling that an error
	// claims the program is broken when the language defines the behaviour.
	//
	// A REFERENCE is the other case, and IsCoercible alone does not separate
	// it: perl numifies a reference too, but to its ADDRESS. `$hashref + 1`
	// measured 94453646558489 — not a numeric interpretation of the value, the
	// way 3.7 -> 3 is, but a fact about where it happens to live. This is the
	// paper's own argument for references not being in Str, applied to Num.
	// So the conversion has to preserve a value, which means both sides must
	// be in the non-reference scalar family.
	//
	// Undef is excluded from that family here, and perl's own diagnostics are
	// the reason: it warns "Use of uninitialized value" for `undef + 1` and
	// says NOTHING for `0 .. 3.7` or `"ab" x 2.9`. Undef is not a value being
	// converted, it is the absence of one, so it keeps Error severity while
	// truncation drops to Warning.
	//
	// An OBJECT is included, and it is the one case where the answer depends
	// on code PSC cannot see. A class declaring `use overload '0+'` has a
	// real conversion — measured, `$money + 1` is 6 — while a plain blessed
	// hashref numifies to its address like any other reference. Resolving
	// which requires finding the class, so the honest severity is the one
	// that does not assert the conversion is impossible. The paper is equally
	// careful here: an overloaded object is still NOT a Num, because the
	// conversion runs out of the type and nothing converts back, so this
	// stays a diagnostic either way.
	//
	// Comparison operators stay a warning regardless, since they execute via
	// coercion in every case.
	severity := Error
	valueLike := types.Boolean | types.Str | types.DualVar | types.Object
	if types.IsCoercible(actual, expected) &&
		types.IsSubtype(actual, valueLike) && types.IsSubtype(expected, valueLike) {
		severity = Warning
	}
	switch op {
	case "==", "!=", "<", ">", "<=", ">=", "<=>":
		severity = Warning
	}

	return severity, fmt.Sprintf("%s operand of %q: expected %s, got %s", side, op, expected, actual), true
}

// argDirectedReturn is the type of a call to int, abs or scalar with nargs
// arguments, the first of type argType, which follows the argument rather
// than the signature; false for any other call.
func argDirectedReturn(name string, nargs int, argType types.Type) (types.Type, bool) {
	// int() truncates toward zero, so the result is an Int whatever went in —
	// measured, int(3.9) and int(-3.9) are both Int. This one does not depend
	// on the argument and could be a signature, except that Int is currently
	// spelled Num there.
	if name == "int" {
		return types.Int, true
	}

	// abs() follows its ARGUMENT: abs(-5) is Int and abs(-5.5) is Num.
	if name == "abs" && nargs == 1 {
		if argType == types.Unknown {
			return types.Unknown, false
		}
		if types.IsSubtype(argType, types.Num) {
			return argType, true
		}
		return types.Num, true
	}

	if name != "scalar" || nargs != 1 {
		return types.Unknown, false
	}
	if argType == types.Unknown {
		return types.Unknown, false
	}
	if narrowed, ok := types.NarrowByContext(argType, types.ScalarCtx); ok {
		return narrowed, true
	}
	return argType, true
}

// argumentSeverity mirrors the operator rule for builtin arguments: a
// conversion that is defined and preserves a value is a warning, and one that
// cannot happen at all is an error.
//
// An OBJECT is the case this exists for. A class declaring `use overload '0+'`
// has a real numeric conversion — `int($money)` is a defined operation — and
// PSC cannot see whether a given class declares one without resolving it
// across files. Reporting an error asserts the conversion is impossible.
func argumentSeverity(actual, expected types.Type) Severity {
	valueLike := types.Boolean | types.Str | types.DualVar | types.Object
	if types.IsCoercible(actual, expected) &&
		types.IsSubtype(actual, valueLike) && types.IsSubtype(expected, valueLike) {
		return Warning
	}
	return Error
}

// numericFormatType reports the type of a format string that consists of
// exactly one numeric conversion and nothing else.
func numericFormatType(format string) (types.Type, bool) {
	if len(format) < 2 || format[0] != '%' {
		return types.Unknown, false
	}
	body := format[1:]

	// Flags that do not introduce non-numeric characters. "-" is excluded:
	// left-justification pads with TRAILING SPACES, so the result is not a
	// bare number even though it still numifies.
	for len(body) > 0 && (body[0] == '0' || body[0] == '+' || body[0] == ' ') {
		body = body[1:]
	}
	// Width and precision.
	for len(body) > 0 && (body[0] >= '0' && body[0] <= '9' || body[0] == '.') {
		body = body[1:]
	}
	if len(body) != 1 {
		return types.Unknown, false
	}

	switch body[0] {
	case 'd', 'i', 'u':
		return types.Int, true
	case 'f', 'e', 'g':
		return types.Num, true
	}
	return types.Unknown, false
}

// captureGroups returns the body of each top-level capture group in a pattern,
// in order. Non-capturing groups (?:...) and lookarounds are skipped, since
// they do not consume a $N.
func captureGroups(pattern string) []string {
	var groups []string
	depth := 0
	var current []rune
	capturing := []bool{}

	runes := []rune(pattern)
	for i := 0; i < len(runes); i++ {
		c := runes[i]

		if c == '\\' && i+1 < len(runes) {
			if depth > 0 {
				current = append(current, c, runes[i+1])
			}
			i++
			continue
		}

		switch c {
		case '(':
			isCap := !(i+1 < len(runes) && runes[i+1] == '?')
			capturing = append(capturing, isCap)
			depth++
			if depth == 1 && isCap {
				current = nil
			} else if depth > 1 {
				current = append(current, c)
			}
		case ')':
			if depth > 0 {
				depth--
				wasCap := capturing[len(capturing)-1]
				capturing = capturing[:len(capturing)-1]
				if depth == 0 && wasCap {
					groups = append(groups, string(current))
					current = nil
				} else if depth > 0 {
					current = append(current, c)
				}
			}
		default:
			if depth > 0 {
				current = append(current, c)
			}
		}
	}
	return groups
}

// isNumericOnlyGroup reports whether a capture group's body can match nothing
// but digits.
//
// Deliberately narrow: only \d and [0-9] with the usual quantifiers. Anything
// it does not recognise stays Str, because a wrong Int here would claim a
// capture is numeric when it can hold letters — and unlike the reverse, that
// is a claim perl contradicts.
func isNumericOnlyGroup(body string) bool {
	if body == "" {
		return false
	}
	rest := body
	saw := false
	for len(rest) > 0 {
		switch {
		case strings.HasPrefix(rest, `\d`):
			rest = rest[2:]
			saw = true
		case strings.HasPrefix(rest, "[0-9]"):
			rest = rest[5:]
			saw = true
		case strings.HasPrefix(rest, "+"), strings.HasPrefix(rest, "*"),
			strings.HasPrefix(rest, "?"):
			rest = rest[1:]
		case strings.HasPrefix(rest, "{"):
			end := strings.Index(rest, "}")
			if end < 0 {
				return false
			}
			rest = rest[end+1:]
		default:
			return false
		}
	}
	return saw
}

// logicalOps are the operators that return one of their operands rather than
// a value of their own.
var logicalOps = map[string]bool{
	"&&": true, "||": true, "//": true, "and": true, "or": true,
}

// builtinArgType returns the expected type for the i-th argument of a builtin
// whose CORE.pmt candidates are sigs: the join of what each candidate taking
// an i-th argument takes there. Choosing one candidate is RFC 0001 "Call
// sites"', so an argument some candidate takes is not a mismatch. A slurpy
// takes every trailing argument, and one stating an element type, `List[Str]
// @args`, takes each as that element. A slot that may be absent takes a
// written argument as its type less Void. Any where no candidate says.
func builtinArgType(sigs []types.Signature, i int) types.Type {
	t := types.Unknown
	for _, s := range sigs {
		if len(s.Params) == 0 {
			continue
		}
		p := s.Params[min(i, len(s.Params)-1)]
		if i >= len(s.Params) && !p.Slurpy() {
			continue
		}
		pt := p.Type
		if p.Element != types.Unknown {
			pt = p.Element
		}
		// An argument that is written is never absent, so an
		// `Optional[Int]` slot holds it to Int. Any is the permissive top,
		// not a union with Void, and stays Any.
		if !p.Slurpy() && pt != types.Any {
			pt &^= types.Void
		}
		if pt == types.Unknown {
			return types.Any
		}
		t = types.Join(t, pt)
	}
	if t == types.Unknown {
		return types.Any
	}
	return t
}

// builtinArgActual is an argument's type as the builtin name takes it.
// scalar's argument is in the scalar context the builtin exists to impose,
// so an aggregate there is its count: measured, `scalar(@a)` on (1,2,3) is
// 3. Another `$` slot takes an aggregate as the mistake perl calls it:
// `length(@a)` warns "did you mean scalar(@a)?" and `defined(@a)` does not
// compile. As in operandMismatch, only a genuine aggregate narrows.
func builtinArgActual(name string, actual types.Type) types.Type {
	if name != "scalar" || (actual != types.Array && actual != types.Hash && actual != types.List) {
		return actual
	}
	if narrowed, ok := types.NarrowByContext(actual, types.ScalarCtx); ok {
		return narrowed
	}
	return actual
}

// shapedOperatorCandidates is the CORE.pmt candidates of op with the given
// fixity that take operands of the given shapes, one "$" or "@" each, in
// context ctx, or in any context when ctx is UnknownCtx.
func shapedOperatorCandidates(op, fixity, shapes string, ctx types.Context) []types.Signature {
	ctxs := []types.Context{ctx}
	if ctx == types.UnknownCtx {
		ctxs = []types.Context{types.ScalarCtx, types.ListCtx, types.VoidCtx}
	}
	cands := parse.CoreOperator(op, fixity)
	taken := map[int]bool{}
	for _, x := range ctxs {
		for _, i := range types.Taking(cands, make([]types.Type, len(shapes)), shapes, x) {
			taken[i] = true
		}
	}
	var takes []types.Signature
	for i, c := range cands {
		if taken[i] {
			takes = append(takes, c)
		}
	}
	return takes
}

// builtinMinArity is the fewest arguments any of a builtin's candidates
// takes: its required parameters, and its invocant when that is required.
func builtinMinArity(sigs []types.Signature) int {
	least := -1
	for _, s := range sigs {
		n := 0
		if s.Invocant != nil && s.Invocant.Required {
			n++
		}
		for _, p := range s.Params {
			if p.Required {
				n++
			}
		}
		if least < 0 || n < least {
			least = n
		}
	}
	return least
}

// shapedCallReturns is a call's type in context ctx from the candidates sigs
// that take arguments of the given shapes, one "$" or "@" each, falling back
// to builtinReturns when none does.
func shapedCallReturns(sigs []types.Signature, shapes string, ctx types.Context) types.Type {
	ctxs := []types.Context{ctx}
	if ctx == types.UnknownCtx {
		ctxs = []types.Context{types.ScalarCtx, types.ListCtx, types.VoidCtx}
	}
	unknown := make([]types.Type, len(shapes))
	for _, sh := range []string{shapes, ""} {
		t, taken := types.Unknown, false
		for _, x := range ctxs {
			if sel := types.SelectShaped(sigs, unknown, sh, x); sel.Outcome != types.Failed {
				t, taken = types.Join(t, sel.Returns), true
			}
		}
		if taken {
			return t
		}
	}
	return builtinReturns(sigs, ctx)
}

// builtinReturns is a builtin call's type in context ctx from its
// candidates' declared return types: the join of those answering ctx, RFC
// 0001 "Context selects by return type", or of all of them when ctx is
// UnknownCtx. reverse is why: measured, `my @r = reverse "ab", "cd"` is
// ("cd","ab") and `my $s = reverse "ab", "cd"` is "dcba", where a list in
// scalar context would be a count. Selecting by argument types is RFC 0001
// "Call sites"', so a multi's call is otherwise the join of its
// candidates', as for a call site that cannot decide.
func builtinReturns(sigs []types.Signature, ctx types.Context) types.Type {
	t := types.Unknown
	if ctx == types.UnknownCtx {
		for _, s := range sigs {
			t = types.Join(t, s.Returns)
		}
		return t
	}
	for _, i := range types.Answering(sigs, ctx) {
		t = types.Join(t, sigs[i].Returns)
	}
	return t
}

// arityMessage produces a human-readable arity mismatch message.
func arityMessage(name string, min, got int) string {
	return "call to " + name + ": expected at least " +
		strconv.Itoa(min) + " argument(s), got " + strconv.Itoa(got)
}

// typeMismatchMessage produces a human-readable type mismatch message.
func typeMismatchMessage(name string, argIdx int, expected, actual types.Type) string {
	return "call to " + name + ": argument " + strconv.Itoa(argIdx+1) +
		" expects " + expected.String() + ", got " + actual.String()
}

// compoundGuard holds the structure of a compound boolean guard expression.
// Op is "&&" for AND conditions and "||" for OR conditions.
// Left and Right are the sub-guards, each of which may itself be compound.
type compoundGuard struct {
	Op    string // "&&" or "||"
	Left  *guardResult
	Right *guardResult
}

// guardResult holds the result of extracting a guard pattern from a condition node.
// When Compound is non-nil, this is a compound guard and VarName/Guard/Negated are unused.
// ClassName is set for GuardIsa guards and holds the class name from the "isa" RHS bareword.
type guardResult struct {
	VarName   string
	Guard     types.GuardPattern
	Negated   bool
	ClassName string // Set for GuardIsa guards: the class name on the RHS of "isa"
	Compound  *compoundGuard
}

// guardFunctionTable maps known guard function base names to their guard patterns.
// Both bare names and fully-qualified names are recognized — the package prefix
// is stripped before lookup.
var guardFunctionTable = map[string]types.GuardPattern{
	"blessed": {Kind: types.GuardIsa},
	"reftype": {Kind: types.GuardRef},
	"is_bool": {Kind: types.GuardBool},
}

// narrowContinuation narrows each guarded variable in place, for the code
// after a conditional whose body always exits: there the guard is known to
// have failed, so the guards given are the else branch's.
func narrowContinuation(guards []*guardResult, st *SymbolTable) {
	for _, eg := range guards {
		sym, found := st.Lookup(eg.VarName)
		if !found {
			continue
		}
		var elseType types.Type
		var narrowed bool
		if eg.Negated {
			elseType, narrowed = types.NegateGuard(sym.Type, eg.Guard)
		} else {
			elseType, narrowed = types.NarrowByGuard(sym.Type, eg.Guard)
		}
		if narrowed {
			st.UpdateType(eg.VarName, elseType)
		}
	}
}

// collectPreIfTypes returns the current types of all variables referenced by
// the guard (leaf guards only). This records the pre-if types for the implicit
// else and for the join computation.
func collectPreIfTypes(guard *guardResult, st *SymbolTable) map[string]types.Type {
	if guard == nil {
		return nil
	}
	result := make(map[string]types.Type)
	collectLeafVarTypes(guard, st, result)
	if len(result) == 0 {
		return nil
	}
	return result
}

// collectLeafVarTypes recursively walks a guardResult tree and collects the
// current symbol-table type for each leaf variable referenced.
func collectLeafVarTypes(guard *guardResult, st *SymbolTable, out map[string]types.Type) {
	if guard == nil {
		return
	}
	if guard.Compound != nil {
		collectLeafVarTypes(guard.Compound.Left, st, out)
		collectLeafVarTypes(guard.Compound.Right, st, out)
		return
	}
	if guard.VarName != "" {
		if _, already := out[guard.VarName]; !already {
			if sym, found := st.Lookup(guard.VarName); found {
				out[guard.VarName] = sym.Type
			}
		}
	}
}

// mergeBranchTypes joins the types from src into dst, creating dst if nil.
// Variables present only in one map are taken as-is. An unreachable
// branch's None is the join identity, so its sentinel bit never reaches a
// variable's type.
func mergeBranchTypes(dst, src map[string]types.Type) map[string]types.Type {
	if src == nil {
		return dst
	}
	if dst == nil {
		dst = make(map[string]types.Type)
	}
	for name, typ := range src {
		if existing, ok := dst[name]; ok {
			dst[name] = types.Join(existing, typ)
		} else {
			dst[name] = typ
		}
	}
	return dst
}

// applyBranchJoin computes the post-if join type for each guarded variable and
// applies it to the outer scope via UpdateType.
//
// The join is types.Join of all non-early-exit branch-end types. When there
// is no else branch, the implicit else contributes the pre-if type. A branch
// a guard proves unreachable ends with its variable None, the join identity:
// it adds nothing, and its sentinel bit is not a type a value can have.
func applyBranchJoin(
	guard *guardResult,
	preIfTypes map[string]types.Type,
	ifBranchTypes map[string]types.Type,
	elsifJoinTypes map[string]types.Type,
	elseBranchTypes map[string]types.Type,
	hasElse bool,
	st *SymbolTable,
) {
	if preIfTypes == nil {
		return
	}

	for name, preTyp := range preIfTypes {
		// Start with the if-branch contribution.
		joinType := types.None
		hasContribution := false

		if ifBranchTypes != nil {
			if t, ok := ifBranchTypes[name]; ok {
				joinType = types.Join(joinType, t)
				hasContribution = true
			}
		}

		// Add elsif contributions.
		if elsifJoinTypes != nil {
			if t, ok := elsifJoinTypes[name]; ok {
				joinType = types.Join(joinType, t)
				hasContribution = true
			}
		}

		// Add else contribution or implicit-else pre-if type.
		if hasElse {
			if elseBranchTypes != nil {
				if t, ok := elseBranchTypes[name]; ok {
					joinType = types.Join(joinType, t)
					hasContribution = true
				}
			}
		} else {
			// No else: implicit else contributes pre-if type.
			joinType = types.Join(joinType, preTyp)
			hasContribution = true
		}

		if hasContribution && joinType != preTyp {
			st.UpdateType(name, joinType)
		}
	}
}

// remapGuardVar replaces occurrences of paramName in the guard result's
// VarName with argName. For compound guards, it recursively remaps both
// sides. Returns nil if the guard is nil.
func remapGuardVar(guard *guardResult, paramName, argName string) *guardResult {
	if guard == nil {
		return nil
	}

	result := *guard // shallow copy

	if result.Compound != nil {
		result.Compound = &compoundGuard{
			Op:    guard.Compound.Op,
			Left:  remapGuardVar(guard.Compound.Left, paramName, argName),
			Right: remapGuardVar(guard.Compound.Right, paramName, argName),
		}
		return &result
	}

	if result.VarName == paramName {
		result.VarName = argName
	}
	return &result
}

// flattenGuards converts a guardResult (possibly compound) into a flat list
// of leaf guards with the Negated flag resolved per compound semantics.
//
// The negate parameter represents an outer negation to XOR into each leaf's
// own Negated flag:
//
//   - For && (negate=false): both sub-guards apply → recurse into each.
//   - For || (negate=false): neither can be guaranteed → return nil.
//   - For && (negate=true, De Morgan → ||): no narrowing → return nil.
//   - For || (negate=true, De Morgan → &&): both negated sub-guards apply → recurse.
//
// For simple (non-compound) guards: return a one-element slice with Negated
// set to guard.Negated XOR negate.
func flattenGuards(guard *guardResult, negate bool) []*guardResult {
	if guard == nil {
		return nil
	}

	if guard.Compound == nil {
		// Leaf guard: resolve the final Negated flag and return it.
		// ClassName is preserved so that GuardIsa guards retain the class name
		// when the flattened leaf is used to shadow variables in guardShadows.
		leaf := &guardResult{
			VarName:   guard.VarName,
			Guard:     guard.Guard,
			Negated:   guard.Negated != negate, // XOR: outer negate flips the leaf
			ClassName: guard.ClassName,
		}
		return []*guardResult{leaf}
	}

	// Compound guard: determine whether this compound's semantics yield all-leaf
	// or no-leaf based on the operator and the outer negate.
	//
	// Effective operator after applying De Morgan for the outer negation:
	//   negate=false: keep Op as-is.
	//   negate=true:  flip Op (De Morgan: !(A&&B) = !A||!B, !(A||B) = !A&&!B).
	effectiveOp := guard.Compound.Op
	if negate {
		if effectiveOp == "&&" {
			effectiveOp = "||"
		} else {
			effectiveOp = "&&"
		}
	}

	switch effectiveOp {
	case "&&":
		// Both sub-guards apply — flatten each with the (possibly flipped) negate.
		left := flattenGuards(guard.Compound.Left, negate)
		right := flattenGuards(guard.Compound.Right, negate)
		return append(left, right...)
	case "||":
		// Either sub-guard could be the true one — no narrowing is certain.
		return nil
	}

	return nil
}

// guardShadows applies guards to the symbol table's types and returns what
// each guarded variable narrows to, and the class an isa guard names. Only
// a variable at least one guard narrowed is in shadows.
//
// Guards that share the same variable are applied sequentially (the output
// type of one guard becomes the input type for the next), so that compound
// conditions on the same variable (e.g. defined($x) && ref($x)) correctly
// intersect their effects. A variable never declared is skipped, so no
// phantom shadow is made for it, except under a referent guard: what
// `ref \$v eq 'GLOB'` narrows $v to does not depend on how $v was typed,
// and a variable a sub body declares is not found from the lowered walk.
func guardShadows(guards []*guardResult, st *SymbolTable) (map[string]types.Type, map[string]string) {
	// Build accumulated narrowed types by applying guards sequentially.
	// working tracks the current type for each variable as guards are applied,
	// whether or not each individual guard produced a change.
	// shadows collects only variables where at least one guard narrowed.
	// guardClassNames records the class name for GuardIsa guards by variable name.
	working := make(map[string]types.Type)
	shadows := make(map[string]types.Type)
	guardClassNames := make(map[string]string)

	for _, guard := range guards {
		// Determine the starting type for this variable: use the working type
		// from prior guards, or look up the declared type for the first guard.
		currentType, seen := working[guard.VarName]
		if !seen {
			sym, found := st.Lookup(guard.VarName)
			if !found && guard.Guard.Kind != types.GuardReferent {
				// Variable was never declared — skip to avoid phantom shadows.
				continue
			}
			currentType = sym.Type
		}

		var narrowedType types.Type
		var didNarrow bool
		if guard.Negated {
			narrowedType, didNarrow = types.NegateGuard(currentType, guard.Guard)
		} else {
			narrowedType, didNarrow = types.NarrowByGuard(currentType, guard.Guard)
		}

		if didNarrow {
			working[guard.VarName] = narrowedType
			shadows[guard.VarName] = narrowedType
		} else {
			// Guard did not change the type, but record the working type so
			// subsequent guards for the same variable chain from it.
			working[guard.VarName] = currentType
		}

		// Record the class name for GuardIsa guards so the shadow symbol can
		// carry ClassType, enabling method resolution inside the block.
		if guard.Guard.Kind == types.GuardIsa && guard.ClassName != "" && !guard.Negated {
			guardClassNames[guard.VarName] = guard.ClassName
		}
	}
	return shadows, guardClassNames
}
