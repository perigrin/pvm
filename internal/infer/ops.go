// ABOUTME: The operator rules of Analyze: Binary, Unary, Ternary, Repeat, Range and Assign, the operand checks, and assignment narrowing.
// ABOUTME: Each reads the lowered kinds: the operator is a field, never re-read from source text.

package infer

import (
	"fmt"
	"regexp"
	"strings"

	"tamarou.com/pvm/internal/lower"
	"tamarou.com/pvm/internal/parse"
	"tamarou.com/pvm/internal/types"
)

// binary types a binary operator: the operator's CORE.pmt
// candidates, chosen by the operands' shapes and the context an assignment
// puts the expression in, check both operands and give its type. A logical
// operator yields one of its operands, so it is their join, as leftValue
// says of the left. An operator
// CORE.pmt does not declare is not typed.
func (a *analysis) binary(n *lower.Binary) types.Type {
	return a.infix(n, n.Op, operandShape(n.L)+operandShape(n.R), n.L, n.R)
}

// infix types an infix operator op over l and r, whose shapes are one "$"
// or "@" each, and checks both operands against what op expects.
//
// An operand is evaluated in scalar context unless the operator takes a
// list there. A logical operator tests its left operand, in scalar context,
// and yields its right in the context the operator is in: perl 5.42, with
// @a = (7,8,9), `(@a || 0) + 1` is 4 and `my @l = 0 || @a` is (7, 8, 9).
func (a *analysis) infix(n lower.Node, op, shapes string, l, r lower.Node) types.Type {
	ctx := a.evalCtx[n.Meta().ID]
	sigs := shapedOperatorCandidates(op, "infix", shapes, ctx)
	for i, o := range []lower.Node{l, r} {
		if o == nil || len(sigs) == 0 {
			continue
		}
		switch {
		case logicalOps[op] && i == 1:
			a.evalCtx[o.Meta().ID] = ctx
		case logicalOps[op] || builtinArgType(sigs, i) != types.List:
			a.evalCtx[o.Meta().ID] = types.ScalarCtx
		}
	}
	a.operand(l)
	a.logicalRight(op, l, r)
	if len(sigs) == 0 {
		return types.Unknown
	}
	a.checkOperand(l, builtinArgType(sigs, 0), op, "left")
	a.checkOperand(r, patternType(op, builtinArgType(sigs, 1)), op, "right")
	result := builtinReturns(sigs, types.UnknownCtx)
	if op == "<=>" {
		return spaceshipResult(a.typeOf(l), a.typeOf(r), result)
	}
	if result == types.Any && logicalOps[op] {
		return types.Join(a.leftValue(op, l), a.valueOf(r))
	}
	return result
}

// leftValue is the value a logical operator op yields from its left
// operand l. `//`, `||` and `or` yield the right operand in place of an
// undefined left, so the left gives no Undef: perl 5.42 -w is silent on
// `my $v = $x // 0; $v + 1` with $x undef. `&&` and `and` yield a false
// left as it is, Undef included.
func (a *analysis) leftValue(op string, l lower.Node) types.Type {
	if op == "//" || op == "||" || op == "or" {
		return a.valueOf(l) &^ types.Undef
	}
	return a.valueOf(l)
}

// repeat is the rule for `x` on Repeat. A list repeat is
// `@`-shaped, so CORE.pmt's list candidate takes it: perl, `my @z = (0) x 3`
// is (0, 0, 0) and `my @w = qw(a b) x 2` is (a, b, a, b). ListRepeat says
// which, from the parens or the qw list on the left.
func (a *analysis) repeat(n *lower.Repeat) types.Type {
	left := "$"
	if n.ListRepeat {
		left = "@"
	}
	return a.infix(n, "x", left+operandShape(n.Count), n.Left, n.Count)
}

// spaceshipResult is `<=>`'s type, result, over operands typed l and r.
// CORE.pmt declares it Int|Undef, as perl gives undef when an operand is
// NaN, but a value that only might be NaN is compared as finite says:
// perl 5.42 -w is silent on `$sgn * ($re1 <=> $re2)`. Only an operand
// known to be NaN, or Inf|NaN, with no other numeric bit, keeps the Undef.
func spaceshipResult(l, r, result types.Type) types.Type {
	for _, t := range []types.Type{l, r} {
		if t&types.Num == 0 && t&types.NaN != 0 {
			return result
		}
	}
	return result &^ types.Undef
}

// patternType is what the right of a bind takes. CORE.pmt declares it
// Regex, and perl compiles a string there as a pattern: perl 5.42 -w,
// `"Foo" =~ q#\A\w+\z#` matches silently. So it takes a Str too, while an
// undef is still reported, as perl warns "Use of uninitialized value in
// regexp compilation". Elsewhere expected is returned as it is.
func patternType(op string, expected types.Type) types.Type {
	if (op == "=~" || op == "!~") && expected == types.Regex {
		return types.Regex | types.Str
	}
	return expected
}

// rangeOp types `..` and `...` by CORE.pmt's candidates, as a binary
// operator: a list in list context and the flip-flop's Str in scalar.
func (a *analysis) rangeOp(n *lower.Range) types.Type {
	op := ".."
	if n.Exclusive {
		op = "..."
	}
	return a.infix(n, op, operandShape(n.Lo)+operandShape(n.Hi), n.Lo, n.Hi)
}

// assign is the plain-assignment half of assignment narrowing: the right
// side is evaluated in the context the
// left side imposes, and a variable on the left is narrowed by it, as
// narrow says. A compound assignment narrows it to what its operator
// gives, as compoundResult says. An element, a slice or a list on the
// left is not narrowed.
//
// The assignment's own type is its right side's, an aggregate counted when
// a scalar takes it.
func (a *analysis) assign(n *lower.Assign) types.Type {
	if n.R != nil {
		a.evalCtx[n.R.Meta().ID] = assignContext(n.L)
	}
	a.operand(n.L)
	t := a.operand(n.R)
	v, ok := n.L.(*lower.Var)
	if !ok || n.R == nil {
		return t
	}
	sigilType, named := sigilTypes[v.Sigil]
	if !named {
		return t
	}
	// An untyped value still replaces what the variable held: perl 5.42
	// -w is silent on `my $v = undef; $v = <STDIN>; $v =~ /x/`.
	if t == types.Unknown && n.Op == "=" {
		a.st.UpdateType(v.Sigil+v.Name, sigilType)
	}
	if n.Op != "=" {
		t = compoundResult(n.Op, a.typeOf(n.L), t)
	}
	return a.narrow(v, n.R, t)
}

// compoundResult is what a compound assignment `op=` leaves in a target
// that held old, given a right side typed r: what CORE.pmt says op gives,
// not r. perl 5.42, `my $s = "a"; $s .= 1` is "a1", a Str. The logical
// ones assign r only sometimes: `//=` only to an undefined target, so a
// defined one keeps its type, and `||=` only to a false one, so an
// undefined target never survives either; `&&=` only to a true one, so
// an undefined target stays undef. perl 5.42 -w is silent on `my $k = 4;
// $k //= undef; $k + 1`, and warns "Use of uninitialized value" on `my
// $u; $u &&= 5; $u + 1`.
func compoundResult(op string, old, r types.Type) types.Type {
	switch op {
	case "//=":
		if old != types.Unknown && old&types.Undef == 0 {
			return old
		}
		return types.Join(old&^types.Undef, r)
	case "||=":
		return types.Join(old&^types.Undef, r)
	case "&&=":
		return types.Join(old, r)
	}
	if sigs := parse.CoreOperator(op, "infix"); len(sigs) > 0 {
		return builtinReturns(sigs, types.UnknownCtx)
	}
	return r
}

// assignContext is an assignment's context: assigning to a scalar or an
// element evaluates the right side in scalar context, and assigning to an
// array, a hash or a list in list context.
func assignContext(l lower.Node) types.Context {
	switch l := l.(type) {
	case *lower.Var:
		return sigilContext(l.Sigil)
	case *lower.Index, *lower.Key:
		return types.ScalarCtx
	case *lower.List:
		return types.ListCtx
	}
	return types.UnknownCtx
}

// narrow is assignment narrowing for v assigned init, whose
// type is t, and returns the type v takes: t, unless that is Unknown.
//
// A scalar imposes scalar context, and an aggregate there is its count:
// `my $count = @arr` is 3, not the array. Only a genuine aggregate is
// narrowed, so a broad type is not stripped to one that satisfies nothing.
// `my $n = () = EXPR` is the count-of idiom: perl, `my $n = () = (1, 2, 3)`
// is 3. An array, a hash, or a scalar holding an anonymous one records its
// element type.
//
// A constructor's result records its class on v. A declaration binding a
// match's captures is bindsCaptures', and does not reach here.
func (a *analysis) narrow(v *lower.Var, init lower.Node, t types.Type) types.Type {
	name := v.Sigil + v.Name
	if v.Sigil == "$" && (t == types.Array || t == types.Hash || t == types.List) {
		if narrowed, ok := types.NarrowByContext(t, types.ScalarCtx); ok {
			t = narrowed
		}
	}
	if v.Sigil == "$" && isCountOf(init) {
		a.st.UpdateType(name, types.Int)
		return types.Int
	}
	a.recordElemType(v, init)
	if t != types.Unknown {
		a.st.UpdateType(name, t)
	}
	if class := a.classTypes[init.Meta().ID]; class != "" {
		a.st.UpdateClassType(name, class)
	}
	return t
}

// firstElement is the type of the first element of init, a list
// initialiser typed t, which a parenthesised declaration's first scalar
// takes: perl 5.42, `my @a = (1, 2); my ($x) = @a` makes $x 1, not the
// count. A hash's first element is a key, a Str. An initialiser that is
// no aggregate is t as it is.
func (a *analysis) firstElement(init lower.Node, t types.Type) types.Type {
	if t != types.Array && t != types.Hash && t != types.List {
		return t
	}
	if v, ok := init.(*lower.Var); ok {
		if v.Sigil == "%" {
			return types.Str
		}
		return a.element(v.Sigil + v.Name)
	}
	elem := a.elemType(init)
	if elem&(types.Array|types.Hash|types.List) != 0 {
		return types.Scalar
	}
	return oneElement(elem)
}

// isCountOf reports whether n is the inner `() = EXPR` of the count-of
// idiom: an assignment to the empty list.
func isCountOf(n lower.Node) bool {
	as, ok := n.(*lower.Assign)
	if !ok {
		return false
	}
	l, ok := as.L.(*lower.List)
	return ok && len(l.Elems) == 0
}

// operandShape is an operand's shape as CORE.pmt's candidates take it: "@"
// for a parenthesised list, "$" for anything else.
func operandShape(n lower.Node) string {
	if _, ok := n.(*lower.List); ok {
		return "@"
	}
	return "$"
}

// checkOperand is the operand check: an operand whose type does
// not satisfy what the operator expects is a coercion mismatch. A missing
// operand is not checked, and a Hole is Any, which satisfies every
// operator, Strict or not. The operand is judged as finite and numified
// say, and a builtin's result also as succeeded says.
func (a *analysis) checkOperand(n lower.Node, expected types.Type, op, side string) {
	if n == nil {
		return
	}
	if a.opts.satisfies(numified(finite(a.typeOf(n)), expected), expected) ||
		a.opts.satisfies(numified(finite(a.succeeded(n, a.typeOf(n))), expected), expected) ||
		a.lossyCoercion(n, a.typeOf(n), expected) {
		return
	}
	severity, message, ok := operandMismatch(a.opts, a.typeOf(n), expected, op, side)
	if !ok {
		return
	}
	start, end, _ := a.pos.Span(n.Meta().ID)
	a.diags = append(a.diags, Diagnostic{
		StartByte: uint32(start),
		EndByte:   uint32(end),
		Severity:  severity,
		Code:      CodeCoercionMismatch,
		Message:   message,
	})
}

// lossyCoercion reports whether n, typed actual, is a coercion perl makes
// silently but that loses its value, and so takes no ordinary mismatch: a
// reference or an object where a Str or a number is expected, and a Str
// of unknown content where a number is. perl 5.42 -w is silent on `print
// $hashref`, which prints HASH(0x...), on `$h + 1`, which adds the
// address, on `version->new("1.2") <= version->new("1.3")`, which
// overloading compares, and on `"$]" < 5.016`. PSC cannot see most
// overloads -- version's come from XS -- so an object is judged as any
// reference is. Under WarnLossyCoercion it reports one as a
// lossy-coercion warning.
//
// A string literal's content is known: one that looks like a number is
// no coercion at all, as perl 5.42 -w is silent on `"3" + 1`, and one that
// does not is reported as it is, as perl warns `Argument "abc" isn't
// numeric` on `"abc" + 1`.
func (a *analysis) lossyCoercion(n lower.Node, actual, expected types.Type) bool {
	actual = numified(finite(actual), expected)
	if actual == types.Unknown || expected == types.Unknown {
		return false
	}
	refToValue := types.IsSubtype(expected, types.Str) &&
		actual&types.Ref != 0 && types.IsSubtype(actual, types.Str|types.Ref)
	strToNum := types.IsSubtype(expected, types.Num) && types.IsSubtype(actual, types.Str) &&
		actual&^(types.Num|types.NaN|types.Inf) != 0
	if !refToValue && !strToNum {
		return false
	}
	if content, known := literalContent(n); known && strToNum {
		return numericString.MatchString(content)
	}
	if a.opts.WarnLossyCoercion {
		start, end, _ := a.pos.Span(n.Meta().ID)
		a.diags = append(a.diags, Diagnostic{
			StartByte: uint32(start),
			EndByte:   uint32(end),
			Severity:  Warning,
			Code:      CodeLossyCoercion,
			Message:   fmt.Sprintf("lossy coercion of %s to %s", actual, expected),
		})
	}
	return true
}

// numericString is perl's looks_like_number for a string literal, as
// perl 5.42 -w measures it: silent on "3", " 3.5e2 ", "1.", "+.5", "inf",
// "-nan" and "0 but true", and warning on "abc", "", " ", "1e", "1abc",
// "0x10", "0b1" and "1_000".
var numericString = regexp.MustCompile(`^\s*(?:[+-]?(?:(?:\d+\.?\d*|\.\d+)(?:[eE][+-]?\d+)?|(?i:inf(?:inity)?|nan))|0 but true)\s*$`)

// literalContent is the content of n when it is a string literal that
// interpolates nothing: 'abc' and "abc", but not "$]" or a command.
func literalContent(n lower.Node) (string, bool) {
	lit, ok := n.(*lower.StrLit)
	if !ok || lit.Backtick || lit.Interpolated && strings.ContainsAny(lit.Value, "$@") {
		return "", false
	}
	return lit.Value, true
}

// numified is actual as a slot expecting expected judges it. perl numifies
// a Boolean to 1 or 0 without a warning: perl 5.42 -w, `"-" x (3 > 2)` is
// "-" and `(3 > 2) / 1` is 1. So where an Int or a Num is expected, a
// Boolean is the Int it becomes. It stringifies to "1" or "" as silently:
// `"" . (1 == 2)` is "", so where a Str is expected it is a Str. Where an
// Int is expected, perl truncates any number to one silently: perl 5.42
// -w is silent on `"ab" x 2.9`, `splice @h, 0, 2.5` and `1 << (20 / 8)`,
// so a Num there is the Int it becomes. A glob stringifies to its name
// as silently: perl 5.42 -w, `"" . *STDIN` is "*main::STDIN", so where a
// Str is expected a Glob is a Str. It is no number: `*STDIN + 0` warns,
// so where a Num is expected, which is a Str too, it is kept. A type that satisfies expected as it is, such as Scalar or Any,
// is kept whole. Elsewhere actual is returned as it is.
func numified(actual, expected types.Type) types.Type {
	if expected == types.Unknown || types.TypeSatisfies(actual, expected) {
		return actual
	}
	if actual&types.Boolean != 0 {
		switch {
		case types.IsSubtype(expected, types.Num):
			actual = actual&^types.Boolean | types.Int
		case types.IsSubtype(expected, types.Str):
			actual = actual&^types.Boolean | types.Str
		}
	}
	if actual&types.Glob != 0 && types.IsSubtype(expected, types.Str) && !types.IsSubtype(expected, types.Num) {
		actual = actual&^types.Glob | types.Str
	}
	if expected&types.Num == types.Int && actual&types.Num != 0 {
		actual = actual&^types.Num | types.Int
	}
	return actual
}

// finite is actual checked as a finite value. Arithmetic is typed Num|Inf,
// and `**` Num|NaN|Inf, because it can overflow, but a value that only
// might be Inf or NaN is checked as the number it almost always is: perl
// 5.42 -w is silent on `my $q = 20 / 8; $q + 1`, on `"\t" x (20 / 8)` and
// on `2 ** 0.5 + 1`. Only a value known to be Inf or NaN, with no other
// numeric bit, is kept as it is, and so is still reported where a number
// is expected.
func finite(actual types.Type) types.Type {
	if actual&types.Num == 0 {
		return actual
	}
	return actual &^ (types.Inf | types.NaN)
}

// prefixNames maps a Unary's Op to the name CORE.pmt declares its prefix
// operator under: the parser spells unary minus and plus neg and pos.
var prefixNames = map[string]string{"neg": "-", "pos": "+"}

// unary types a prefix operator: its CORE.pmt
// return type, except that minus and plus keep a numeric operand's own
// type -- -5 is an Int and -5.5 a Num.
//
// Plus is the identity: `+X` is X's value, an aggregate counted outside
// list context. perl 5.42 -w is silent on `splice @t, 0, +@d`, and warns
// "Use of uninitialized value in splice" on `splice @t, 0, +undef`, as it
// does without the plus.
//
// Increment and decrement are not typed.
func (a *analysis) unary(n *lower.Unary) types.Type {
	if n.Op == "pos" && n.Operand != nil {
		ctx := a.evalCtx[n.ID]
		if ctx != types.ListCtx {
			ctx = types.ScalarCtx
		}
		a.evalCtx[n.Operand.Meta().ID] = ctx
		a.operand(n.Operand)
		return a.valueOf(n.Operand)
	}
	operand := a.operand(n.Operand)
	if n.Op == "++" || n.Op == "--" {
		return types.Unknown
	}
	op := n.Op
	if name, ok := prefixNames[op]; ok {
		op = name
	}
	sigs := parse.CoreOperator(op, "prefix")
	if len(sigs) == 0 {
		return types.Unknown
	}
	if (op == "-" || op == "+") && types.IsSubtype(operand, types.Num) {
		return operand
	}
	return builtinReturns(sigs, types.UnknownCtx)
}

// ternary types a ternary: the value comes from one arm
// or the other, so it is their join. The condition says nothing about it.
// A ternary missing an arm is not typed.
//
// Each arm is evaluated in the ternary's own context, so in scalar context
// an array arm is its count: perl 5.42, with @a = (7,8,9),
// `(1 ? @a : 42) + 1` is 4 and `my $n = 1 ? @a : 42` is 3, while
// `my @t = 1 ? @a : 42` is (7, 8, 9).
func (a *analysis) ternary(n *lower.Ternary) types.Type {
	ctx := a.evalCtx[n.ID]
	for _, arm := range []lower.Node{n.Then, n.Else} {
		if arm != nil {
			a.evalCtx[arm.Meta().ID] = ctx
		}
	}
	a.operand(n.Cond)
	a.operand(n.Then)
	a.operand(n.Else)
	if n.Then == nil || n.Else == nil {
		return types.Unknown
	}
	return types.Join(a.valueOf(n.Then), a.valueOf(n.Else))
}

// valueOf is n's type as the value its context takes: an array, a hash or a
// list evaluated in scalar context is its count. Only a genuine aggregate is
// counted, so a broad type is not stripped to one that satisfies nothing.
func (a *analysis) valueOf(n lower.Node) types.Type {
	t := a.typeOf(n)
	if n == nil || a.evalCtx[n.Meta().ID] != types.ScalarCtx {
		return t
	}
	if t == types.Array || t == types.Hash || t == types.List {
		if narrowed, ok := types.NarrowByContext(t, types.ScalarCtx); ok {
			return narrowed
		}
	}
	return t
}

// operand walks n, which may be missing, and returns its type.
func (a *analysis) operand(n lower.Node) types.Type {
	if n == nil {
		return types.Unknown
	}
	return a.walk(n)
}

// typeOf is the type already recorded for n: Unknown when n is missing or
// was not typed.
func (a *analysis) typeOf(n lower.Node) types.Type {
	if n == nil {
		return types.Unknown
	}
	return a.annotations[n.Meta().ID]
}
