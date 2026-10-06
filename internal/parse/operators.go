// ABOUTME: Operator declarations in .pmt files, RFC 0001 "Operator declarations".
// ABOUTME: `sub + :infix(ADD) (Num $x, Num $y) Num;` declares a symbol's fixity, class and types.

package parse

import (
	"errors"
	"fmt"
	"strings"

	"tamarou.com/pvm/internal/types"
)

// operatorDecl is one operator a `.pmt` declares: `sub + :infix(ADD) (Num $x,
// Num $y) Num;` is "+", "infix", "ADD" and its signature. A prefix or postfix
// operator has no class.
type operatorDecl struct {
	name, fixity, class string
	sig                 types.Signature
}

// operatorClasses are the classes `:infix(CLASS)` may name, each to the
// perly.y level in precedence.go that it stands for. RFC 0001: XS::Parse::
// Infix's classification (XSParseInfix.h, XPI_CLS_*) without the `XPI_CLS_`
// prefix and `_MISC` suffix. LOW and HIGH are the plugin hooks either side of
// the core operators, levels 3 and 28.
//
// BITOR, BITAND and SHIFT are not XS::Parse::Infix's: it classes no operator
// at those levels, so it cannot register one there, and they exist only for
// CORE.pmt to declare perl's own.
var operatorClasses = map[string]int{
	"LOW":             3,
	"LOGICAL_OR_LOW":  4,
	"LOGICAL_AND_LOW": 5,
	"BITOR":           14,
	"BITAND":          15,
	"SHIFT":           21,
	"ASSIGN":          9,
	"LOGICAL_OR":      12,
	"LOGICAL_AND":     13,
	"EQUALITY":        16,
	"ORDERING":        16,
	"RELATION":        17,
	"ISA":             17,
	"ADD":             22,
	"MUL":             23,
	"MATCHRE":         24,
	"POW":             26,
	"HIGH":            28,
}

// operatorArity is how many operands each fixity takes.
var operatorArity = map[string]int{"infix": 2, "prefix": 1, "postfix": 1}

// fixityAttrs returns the attributes of a sub declaration that state a
// fixity: `:infix(ADD)`, `:prefix`, `:postfix`. perl has none of them --
// measured on 5.42.0, `sub f :prefix;` is "Invalid CODE attribute: prefix".
func fixityAttrs(n *Node) []string {
	var out []string
	for _, c := range n.Children {
		if c.Kind != Attribute {
			continue
		}
		name, _, _ := strings.Cut(strings.TrimPrefix(c.Text, ":"), "(")
		if _, ok := operatorArity[name]; ok {
			out = append(out, c.Text)
		}
	}
	return out
}

// declareOperator records n as an operator declaration when it states a
// fixity, taking the signature parseTypedSignature set aside for it. One it
// cannot record is an error on p.typedErrs.
func (p *parser) declareOperator(n *Node) {
	attrs := fixityAttrs(n)
	if len(attrs) == 0 {
		return
	}
	name, _ := declaredSub(n)
	sig := p.operatorSig
	p.operatorSig = types.Signature{}
	op, err := operatorOf(name, attrs, sig)
	if err != nil {
		p.typedErrs = append(p.typedErrs, fmt.Errorf("sub %s: %w", name, err))
		return
	}
	p.operators = append(p.operators, op)
}

// operatorOf reads one fixity attribute and checks the signature against it.
func operatorOf(name string, attrs []string, sig types.Signature) (operatorDecl, error) {
	if len(attrs) > 1 {
		return operatorDecl{}, fmt.Errorf("two fixities, %s and %s", attrs[0], attrs[1])
	}
	fixity, class, hasClass := strings.Cut(strings.TrimPrefix(attrs[0], ":"), "(")
	class = strings.TrimSuffix(class, ")")
	switch {
	case fixity == "infix" && !hasClass:
		return operatorDecl{}, errors.New(":infix needs a class")
	case fixity != "infix" && hasClass:
		return operatorDecl{}, fmt.Errorf(":%s takes no class", fixity)
	}
	if _, ok := operatorClasses[class]; hasClass && !ok {
		return operatorDecl{}, fmt.Errorf("unknown operator class %q", class)
	}
	if want := operatorArity[fixity]; len(sig.Params) != want {
		return operatorDecl{}, fmt.Errorf(":%s takes %s, the signature has %d", fixity, []string{"", "one operand", "two operands"}[want], len(sig.Params))
	}
	return operatorDecl{name: name, fixity: fixity, class: class, sig: sig}, nil
}

// precedenceMismatch reports an infix operator whose class names a level
// other than the one precedence.go gives it. The table is authoritative
// (RFC 0001, "Operator declarations"); the class says the same thing again.
func precedenceMismatch(op operatorDecl) error {
	if op.fixity != "infix" {
		return nil
	}
	if level, want := operatorClasses[op.class], infix[op.name].Level; level != want {
		return fmt.Errorf("sub %s: class %s is level %d, and precedence.go puts %s at level %d", op.name, op.class, level, op.name, want)
	}
	return nil
}
