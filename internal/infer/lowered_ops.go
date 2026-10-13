// ABOUTME: The operator rules of AnalyzeLowered: Binary, Unary, Ternary, Repeat and Range typing, and the operand checks.
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
func (a *loweredAnalysis) infix(n lower.Node, op, shapes string, l, r lower.Node) types.Type {
	a.operand(l)
	a.operand(r)
	sigs := shapedOperatorCandidates(op, "infix", shapes, a.rhsCtx[n.Meta().ID])
	if len(sigs) == 0 {
		return types.Unknown
	}
	a.checkOperand(l, builtinArgType(sigs, 0), op, "left")
	a.checkOperand(r, builtinArgType(sigs, 1), op, "right")
	result := builtinReturns(sigs, types.UnknownCtx)
	if result == types.Any && logicalOps[op] {
		return types.Join(a.typeOf(l), a.typeOf(r))
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

// rangeOp types `..` and `...` by CORE.pmt's candidates, as a binary
// operator: a list in list context and the flip-flop's Str in scalar.
func (a *loweredAnalysis) rangeOp(n *lower.Range) types.Type {
	op := ".."
	if n.Exclusive {
		op = "..."
	}
	return a.infix(n, op, operandShape(n.Lo)+operandShape(n.Hi), n.Lo, n.Hi)
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
// operand and a Hole are not inspected.
func (a *loweredAnalysis) checkOperand(n lower.Node, expected types.Type, op, side string) {
	if n == nil {
		return
	}
	if _, ok := n.(*lower.Hole); ok {
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
func (a *loweredAnalysis) ternary(n *lower.Ternary) types.Type {
	a.operand(n.Cond)
	then, els := a.operand(n.Then), a.operand(n.Else)
	if n.Then == nil || n.Else == nil {
		return types.Unknown
	}
	return types.Join(then, els)
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
