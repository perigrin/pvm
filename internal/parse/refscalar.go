// ABOUTME: The \$ prototype slot: a call that fills one with something perl refuses there becomes an Unknown.
// ABOUTME: RFC 0001 "Refusing what perl refuses": a \$ slot takes a scalar lvalue, and nothing else compiles.
package parse

import (
	"slices"
	"strings"
)

// refuseRefScalarSlot returns n, or an Unknown spanning it when n calls a sub
// whose prototype has a `\$` slot and the call fills that slot with an
// argument perl refuses at compile time. Measured on 5.42.0 with `sub f {}
// sub sref (\$) {}`:
//
//	sref(1)      Type of arg 1 to main::sref must be scalar (not constant item)
//	sref(f())    Type of arg 1 to main::sref must be scalar (not subroutine entry)
//	sref($x)     syntax OK
//
// The test perl applies is lvalue-ness (op.c's ck_entersub_args_proto calls
// op_lvalue on the argument), so only arguments that are certainly not
// scalar lvalues are refused here; anything this parser cannot classify
// keeps its parse. A builtin of the same name is the builtin, not the sub.
func (p *parser) refuseRefScalarSlot(n *Node) *Node {
	if n == nil || n.Kind != Call || p.isPerlKeyword(keywordName(n.Text)) {
		return n
	}
	imp, ok := p.lookupSub(n.Text)
	if !ok || !imp.PrototypeKnown {
		return n
	}
	slots, _ := refScalarSlots(imp.Prototype)
	if !slices.ContainsFunc(slots, func(s protoSlot) bool { return s.refScalar }) {
		return n
	}
	var args []*Node
	for _, c := range n.Children {
		args = append(args, commaItems(c)...)
	}
	// A slot left empty, or an argument past the last, is the arity
	// refusal's (arity.go): `sref()` is "Not enough arguments for
	// main::sref".
	refused := false
	for i, slot := range slots {
		refused = refused || slot.refScalar && i < len(args) && p.notScalarLvalue(args[i])
	}
	if refused {
		return refusedScalarLvalue(n)
	}
	return n
}

// refuseAliasedOperand returns n, the node of the operator op, or an
// Unknown spanning it when CORE.pmt declares op's first operand an aliased
// scalar, `\$`, and n's is certainly not a scalar lvalue: a `\$` operand
// refuses what a `\$` slot does. Measured on 5.42.0, `++1` is "Can't
// modify constant item in preincrement (++)" and `1 += 2` "Can't modify
// constant item in addition (+)" as `sref(1)` is refused. CORE.pmt's own
// read, which CoreOperator's table comes from, is not checked.
func (p *parser) refuseAliasedOperand(n *Node, op, fixity string) *Node {
	if p.buildingCore || len(n.Children) == 0 || !p.notScalarLvalue(n.Children[0]) {
		return n
	}
	for _, sig := range CoreOperator(op, fixity) {
		if len(sig.Params) == len(n.Children) && sig.Params[0].Alias && sig.Params[0].Sigil == '$' {
			return refusedScalarLvalue(n)
		}
	}
	return n
}

// refusedScalarLvalue is the Unknown spanning n that a `\$` slot or
// operand refuses it as.
func refusedScalarLvalue(n *Node) *Node {
	return &Node{Kind: Unknown, Refusal: RefScalarSlot, Start: n.Start, End: n.End}
}

// protoSlot is one argument slot of a prototype: whether it is `\$`, and
// whether it follows the `;` that makes the rest optional.
type protoSlot struct{ refScalar, optional bool }

// refScalarSlots reads a prototype, parens included, into its argument
// slots, and reports whether a final `@` or `%` takes the rest of the
// list. A `\[...]` group is one slot; `_` is optional of itself.
func refScalarSlots(proto string) (slots []protoSlot, slurpy bool) {
	inner := strings.TrimSuffix(strings.TrimPrefix(proto, "("), ")")
	optional := false
	for i := 0; i < len(inner); i++ {
		switch c := inner[i]; c {
		case ';':
			optional = true
		case '\\':
			i++
			slots = append(slots, protoSlot{refScalar: i < len(inner) && inner[i] == '$', optional: optional})
			if i < len(inner) && inner[i] == '[' {
				for i < len(inner) && inner[i] != ']' {
					i++
				}
			}
		case '@', '%':
			return slots, true
		case '$', '&', '*', '+', '_':
			slots = append(slots, protoSlot{optional: optional || c == '_'})
		}
	}
	return slots, false
}

// commaItems splits an argument list at its commas. A parenthesised list is
// a List, and one argument: measured on 5.42.0, `sref(($x, $y))` compiles.
func commaItems(n *Node) []*Node {
	if n.Kind == Binary && (n.Text == "," || n.Text == "=>") {
		var items []*Node
		for _, c := range n.Children {
			items = append(items, commaItems(c)...)
		}
		return items
	}
	return []*Node{n}
}

// notScalarLvalue reports whether an argument is certainly not a scalar
// lvalue: a literal constant, a whole array or hash, or the result of a sub
// this file declared earlier without `:lvalue`. Measured on 5.42.0, a sub perl has not seen
// declared may yet be an lvalue sub, so `sref(g()); sub g {}` compiles; so
// does a lexical sub's result, `my sub g {} sref(g())`.
//
// ponytail: a sub imported from a module is not checked, since Import does
// not carry its `:lvalue`; carry it through moduleFacts if such a call shows
// up in a corpus.
func (p *parser) notScalarLvalue(a *Node) bool {
	switch a.Kind {
	case Term:
		if len(a.Children) > 0 || a.Text == "" {
			return false
		}
		text := a.Text
		switch {
		case text[0] >= '0' && text[0] <= '9',
			text[0] == '.' && len(text) > 1 && text[1] >= '0' && text[1] <= '9',
			text[0] == '\'' || text[0] == '"',
			quoteOpIs(text, "q"), quoteOpIs(text, "qq"), quoteOpIs(text, "qw"):
			return true
		case text[0] == '&':
			return p.declaredNonLvalue(text[1:])
		case text[0] == '@' || text[0] == '%':
			// A whole array or hash; a slice, `@a[0]`, is an Index.
			return len(text) > 1
		}
	case Unary:
		// `@$r`, `%{$r}`: a whole aggregate through a reference.
		return a.Text == "@" || a.Text == "%"
	case Declaration:
		// `my @q`, `our %q`: a declared aggregate.
		return len(a.Children) == 1 && a.Children[0].Kind == Term && p.notScalarLvalue(a.Children[0])
	case Call:
		return p.declaredNonLvalue(a.Text)
	case Index:
		// `&f()`: a call through the ampersand.
		if a.Text == "(" && len(a.Children) > 0 && a.Children[0].Kind == Term &&
			strings.HasPrefix(a.Children[0].Text, "&") {
			return p.declaredNonLvalue(a.Children[0].Text[1:])
		}
	}
	return false
}

// declaredNonLvalue reports whether name is a sub this file declared, before
// this point and not as a builtin's name, without `:lvalue`.
func (p *parser) declaredNonLvalue(name string) bool {
	if p.isPerlKeyword(keywordName(name)) {
		return false
	}
	imp, ok := p.lookupSub(name)
	return ok && imp.Local && !imp.Lvalue
}
