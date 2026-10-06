// ABOUTME: The \$ prototype slot: a call that fills one with something perl refuses there becomes an Unknown.
// ABOUTME: RFC 0001 "Refusing what perl refuses": a \$ slot takes a scalar lvalue, and nothing else compiles.
package parse

import "strings"

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
	if n == nil || n.Kind != Call || isPerlKeyword(keywordName(n.Text)) {
		return n
	}
	imp, ok := p.lookupSub(n.Text)
	if !ok || !imp.PrototypeKnown {
		return n
	}
	var args []*Node
	for _, c := range n.Children {
		args = append(args, commaItems(c)...)
	}
	for i, slot := range refScalarSlots(imp.Prototype) {
		if slot && i < len(args) && p.notScalarLvalue(args[i]) {
			return &Node{Kind: Unknown, Refusal: RefScalarSlot, Start: n.Start, End: n.End}
		}
	}
	return n
}

// refScalarSlots reads a prototype, parens included, into one entry per
// argument slot, true where the slot is `\$`. A `\[...]` group is one slot.
func refScalarSlots(proto string) []bool {
	inner := strings.TrimSuffix(strings.TrimPrefix(proto, "("), ")")
	var slots []bool
	for i := 0; i < len(inner); i++ {
		switch inner[i] {
		case '\\':
			i++
			slots = append(slots, i < len(inner) && inner[i] == '$')
			if i < len(inner) && inner[i] == '[' {
				for i < len(inner) && inner[i] != ']' {
					i++
				}
			}
		case '$', '&', '*', '+', '_', '@', '%':
			slots = append(slots, false)
		}
	}
	return slots
}

// commaItems splits an argument list at its unparenthesised commas. A
// parenthesised list is one argument: measured on 5.42.0, `sref(($x, $y))`
// compiles.
func commaItems(n *Node) []*Node {
	if n.Kind == Binary && (n.Text == "," || n.Text == "=>") && !n.Paren {
		var items []*Node
		for _, c := range n.Children {
			items = append(items, commaItems(c)...)
		}
		return items
	}
	return []*Node{n}
}

// notScalarLvalue reports whether an argument is certainly not a scalar
// lvalue: a literal constant, or the result of a sub this file declared
// earlier without `:lvalue`. Measured on 5.42.0, a sub perl has not seen
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
		}
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
	if isPerlKeyword(keywordName(name)) {
		return false
	}
	imp, ok := p.lookupSub(name)
	return ok && imp.Local && !imp.Lvalue
}
