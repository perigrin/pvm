// ABOUTME: The operator rules of AnalyzeLowered: Binary, Unary, Ternary, Repeat, Range and Assign, the operand checks, and assignment narrowing.
// ABOUTME: Each is the legacy rule rewritten against the lowered kinds: the operator is a field, never re-read from source text.

package infer

import (
	"tamarou.com/pvm/internal/lower"
	"tamarou.com/pvm/internal/parse"
	"tamarou.com/pvm/internal/types"
)

// binary is the legacy inferBinaryExprType: the operator's CORE.pmt
// candidates, chosen by the operands' shapes and the context an assignment
// puts the expression in, check both operands and give its type. A logical
// operator yields one of its operands, so it is their join. An operator
// CORE.pmt does not declare is not typed.
func (a *loweredAnalysis) binary(n *lower.Binary) types.Type {
	return a.infix(n, n.Op, operandShape(n.L)+operandShape(n.R), n.L, n.R)
}

// infix types an infix operator op over l and r, whose shapes are one "$"
// or "@" each, and checks both operands against what op expects.
//
// An operand is evaluated in scalar context unless the operator takes a
// list there. A logical operator tests its left operand, in scalar context,
// and yields its right in the context the operator is in: perl 5.42, with
// @a = (7,8,9), `(@a || 0) + 1` is 4 and `my @l = 0 || @a` is (7, 8, 9).
func (a *loweredAnalysis) infix(n lower.Node, op, shapes string, l, r lower.Node) types.Type {
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
	a.operand(r)
	if len(sigs) == 0 {
		return types.Unknown
	}
	a.checkOperand(l, builtinArgType(sigs, 0), op, "left")
	a.checkOperand(r, countType(op, builtinArgType(sigs, 1)), op, "right")
	result := builtinReturns(sigs, types.UnknownCtx)
	if result == types.Any && logicalOps[op] {
		return types.Join(a.valueOf(l), a.valueOf(r))
	}
	return result
}

// repeat is the legacy rule for `x` on Repeat. A list repeat is
// `@`-shaped, so CORE.pmt's list candidate takes it: perl, `my @z = (0) x 3`
// is (0, 0, 0) and `my @w = qw(a b) x 2` is (a, b, a, b). ListRepeat says
// which, from the parens or the qw list on the left.
func (a *loweredAnalysis) repeat(n *lower.Repeat) types.Type {
	left := "$"
	if n.ListRepeat {
		left = "@"
	}
	return a.infix(n, "x", left+operandShape(n.Count), n.Left, n.Count)
}

// countType is what a repeat's count takes. CORE.pmt declares it Int, and
// perl truncates any finite number to one silently: perl 5.42 -w, `"ab" x
// 2.9` is "abab". So the count takes a Num, while a reference or an undef
// count is still reported, as perl warns or repeats by an address.
// Elsewhere expected is returned as it is.
func countType(op string, expected types.Type) types.Type {
	if op == "x" && expected == types.Int {
		return types.Num
	}
	return expected
}

// rangeOp types `..` and `...` by CORE.pmt's candidates, as a binary
// operator: a list in list context and the flip-flop's Str in scalar.
func (a *loweredAnalysis) rangeOp(n *lower.Range) types.Type {
	op := ".."
	if n.Exclusive {
		op = "..."
	}
	return a.infix(n, op, operandShape(n.Lo)+operandShape(n.Hi), n.Lo, n.Hi)
}

// assign is the plain-assignment half of the legacy
// inferAssignmentNarrowing: the right side is evaluated in the context the
// left side imposes, and a variable on the left is narrowed by it, as
// narrow says. Every assignment operator narrows so, `+=` included. An
// element, a slice or a list on the left is not narrowed.
//
// The assignment's own type is its right side's, an aggregate counted when
// a scalar takes it.
func (a *loweredAnalysis) assign(n *lower.Assign) types.Type {
	if n.R != nil {
		a.evalCtx[n.R.Meta().ID] = assignContext(n.L)
	}
	a.operand(n.L)
	t := a.operand(n.R)
	v, ok := n.L.(*lower.Var)
	if !ok || n.R == nil {
		return t
	}
	if _, named := sigilTypes[v.Sigil]; !named {
		return t
	}
	return a.narrow(v, n.R, t)
}

// assignContext is the legacy assignedContext: assigning to a scalar or an
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

// narrow is the legacy inferAssignmentNarrowing for v assigned init, whose
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
func (a *loweredAnalysis) narrow(v *lower.Var, init lower.Node, t types.Type) types.Type {
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

// checkOperand is the legacy checkBinaryOperand: an operand whose type does
// not satisfy what the operator expects is a coercion mismatch. A missing
// operand is not checked, and a Hole is Any, which satisfies every
// operator, Strict or not. The operand is judged as finite and numified
// say.
func (a *loweredAnalysis) checkOperand(n lower.Node, expected types.Type, op, side string) {
	if n == nil {
		return
	}
	if a.opts.satisfies(numified(finite(a.typeOf(n)), expected), expected) {
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

// numified is actual as a slot expecting expected judges it. perl numifies
// a Boolean to 1 or 0 without a warning: perl 5.42 -w, `"-" x (3 > 2)` is
// "-" and `(3 > 2) / 1` is 1. So where an Int or a Num is expected, a
// Boolean is the Int it becomes. Elsewhere actual is returned as it is.
func numified(actual, expected types.Type) types.Type {
	if actual&types.Boolean == 0 || expected == types.Unknown || !types.IsSubtype(expected, types.Num) {
		return actual
	}
	return actual&^types.Boolean | types.Int
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

// unary is the legacy inferUnaryExprType: a prefix operator's CORE.pmt
// return type, except that minus and plus keep a numeric operand's own
// type -- -5 is an Int and -5.5 a Num.
//
// Increment and decrement are not typed: the legacy path had no rule for
// them.
func (a *loweredAnalysis) unary(n *lower.Unary) types.Type {
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

// ternary is the legacy inferConditionalType: the value comes from one arm
// or the other, so it is their join. The condition says nothing about it.
// A ternary missing an arm is not typed.
//
// Each arm is evaluated in the ternary's own context, so in scalar context
// an array arm is its count: perl 5.42, with @a = (7,8,9),
// `(1 ? @a : 42) + 1` is 4 and `my $n = 1 ? @a : 42` is 3, while
// `my @t = 1 ? @a : 42` is (7, 8, 9).
func (a *loweredAnalysis) ternary(n *lower.Ternary) types.Type {
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
func (a *loweredAnalysis) valueOf(n lower.Node) types.Type {
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
func (a *loweredAnalysis) operand(n lower.Node) types.Type {
	if n == nil {
		return types.Unknown
	}
	return a.walk(n)
}

// typeOf is the type already recorded for n: Unknown when n is missing or
// was not typed.
func (a *loweredAnalysis) typeOf(n lower.Node) types.Type {
	if n == nil {
		return types.Unknown
	}
	return a.annotations[n.Meta().ID]
}
