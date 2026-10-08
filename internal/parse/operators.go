// ABOUTME: Operator declarations in .pmt files, RFC 0001 "Operator declarations".
// ABOUTME: `sub + :infix(ADD) (Num $x, Num $y) Num;` declares a symbol's fixity, class and types.

package parse

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"tamarou.com/pvm/internal/types"
)

// operatorDecl is one operator a `.pmt` declares: `sub + :infix(ADD) (Num $x,
// Num $y) Num;` is "+", "infix", "ADD" and its signature. A prefix or postfix
// operator has no class. multi is set for a `multi sub`, one of several
// candidates for its symbol (RFC 0001, "Operators that fork").
//
// tighter, looser and equiv are the operators its `:tighter(OP)`,
// `:looser(OP)` and `:equiv(OP)` name, and assoc its `:assoc(...)` (RFC 0001,
// "Precedence is a relation between operators").
type operatorDecl struct {
	name, fixity, class    string
	sig                    types.Signature
	multi                  bool
	tighter, looser, equiv []string
	assoc                  string
}

// equivs is the operators whose level op is in: those its `:equiv(OP)`
// names, and its class's anchor, `:infix(MUL)` being `:equiv(*)`.
func (op operatorDecl) equivs() []string {
	if anchor, ok := classAnchors[op.class]; ok {
		return append(slices.Clone(op.equiv), anchor)
	}
	return op.equiv
}

// classAnchors is the operator each class of operatorClasses stands for the
// level of, so `:infix(CLASS)` reads as `:equiv` of it. LOW and HIGH have
// none: they are XS::Parse::Infix's plugin levels, where perl has no
// operator.
var classAnchors = map[string]string{
	"LOGICAL_OR_LOW":  "or",
	"LOGICAL_AND_LOW": "and",
	"BITOR":           "|",
	"BITAND":          "&",
	"SHIFT":           "<<",
	"ASSIGN":          "=",
	"RANGE":           "..",
	"LOGICAL_OR":      "||",
	"LOGICAL_AND":     "&&",
	"EQUALITY":        "==",
	"ORDERING":        "<=>",
	"RELATION":        "<",
	"ISA":             "isa",
	"ADD":             "+",
	"MUL":             "*",
	"MATCHRE":         "=~",
	"POW":             "**",
}

// assocs are the associativities `:assoc(...)` may state, perlop's:
// `chained` for `<` and its row, `chain_na` for `==` and its.
var assocs = map[string]bool{"left": true, "right": true, "nonassoc": true, "chained": true, "chain_na": true}

// operatorClasses are the classes `:infix(CLASS)` may name, each to the
// perly.y level in precedence.go that it stands for. RFC 0001: XS::Parse::
// Infix's classification (XSParseInfix.h, XPI_CLS_*) without the `XPI_CLS_`
// prefix and `_MISC` suffix. LOW and HIGH are the plugin hooks either side of
// the core operators, levels 3 and 28.
//
// BITOR, BITAND, SHIFT and RANGE are not XS::Parse::Infix's: it classes no
// operator at those levels, so it cannot register one there, and they exist
// only for CORE.pmt to declare perl's own.
var operatorClasses = map[string]int{
	"LOW":             3,
	"LOGICAL_OR_LOW":  4,
	"LOGICAL_AND_LOW": 5,
	"BITOR":           14,
	"BITAND":          15,
	"SHIFT":           21,
	"ASSIGN":          9,
	"RANGE":           11,
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

// coreOnlyClasses are the classes operatorClasses coins for CORE.pmt, which
// a library's declaration may not name: XS::Parse::Infix cannot register an
// operator at their levels.
var coreOnlyClasses = map[string]bool{"BITOR": true, "BITAND": true, "SHIFT": true, "RANGE": true}

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
	op.multi = p.multi
	if err := readRelations(n, &op); err != nil {
		p.typedErrs = append(p.typedErrs, fmt.Errorf("sub %s: %w", name, err))
		return
	}
	p.operators = append(p.operators, op)
}

// readRelations reads n's `:tighter(OP)`, `:looser(OP)`, `:equiv(OP)` and
// `:assoc(...)` into op.
func readRelations(n *Node, op *operatorDecl) error {
	for _, c := range n.Children {
		if c.Kind != Attribute {
			continue
		}
		name, arg, _ := strings.Cut(strings.TrimPrefix(c.Text, ":"), "(")
		arg = strings.TrimSuffix(arg, ")")
		var into *[]string
		switch name {
		case "tighter":
			into = &op.tighter
		case "looser":
			into = &op.looser
		case "equiv":
			into = &op.equiv
		case "assoc":
			if !assocs[arg] {
				return fmt.Errorf("unknown associativity %q", arg)
			}
			op.assoc = arg
			continue
		default:
			continue
		}
		if arg == "" {
			return fmt.Errorf(":%s names no operator", name)
		}
		*into = append(*into, arg)
	}
	return nil
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

// precLevel is one level of a derived precedence order: its operators, an
// infix one by its symbol and any other as `prefix -`, and its
// associativity.
type precLevel struct {
	ops   []string
	assoc string
}

// precedenceOrder derives the total order of ops's levels, tightest first,
// from their relations (RFC 0001, "Precedence is a relation between
// operators"): `:equiv` joins two operators' levels, and `:tighter` and
// `:looser` order two levels, which are sorted topologically. A relation
// names the infix operator of its spelling where there is one, so `:tighter(+)`
// is binary `+`, and otherwise the prefix or postfix one. Relations that
// form a cycle, leave a level related to nothing, or leave two levels
// unordered derive no order and are an error naming the operators, as are
// two associativities stated in one level.
func precedenceOrder(ops []operatorDecl) ([]precLevel, error) {
	type key = [2]string
	var keys []key
	index := map[key]int{}
	for _, op := range ops {
		k := key{op.name, op.fixity}
		if _, seen := index[k]; !seen {
			index[k] = len(keys)
			keys = append(keys, k)
		}
	}
	label := func(i int) string {
		if keys[i][1] == "infix" {
			return keys[i][0]
		}
		return keys[i][1] + " " + keys[i][0]
	}
	target := func(op operatorDecl, attr, name string) (int, error) {
		for _, fixity := range []string{"infix", "prefix", "postfix"} {
			if j, ok := index[key{name, fixity}]; ok {
				return j, nil
			}
		}
		return 0, fmt.Errorf("sub %s: :%s(%s) names no declared operator", op.name, attr, name)
	}

	// Each operator's level is the root of its :equiv joins.
	parent := make([]int, len(keys))
	for i := range parent {
		parent[i] = i
	}
	find := func(i int) int {
		for parent[i] != i {
			i = parent[i]
		}
		return i
	}
	type edge struct{ tight, loose int }
	var edges []edge
	for _, op := range ops {
		i := index[key{op.name, op.fixity}]
		for _, name := range op.equivs() {
			j, err := target(op, "equiv", name)
			if err != nil {
				return nil, err
			}
			parent[find(j)] = find(i)
		}
		for _, name := range op.tighter {
			j, err := target(op, "tighter", name)
			if err != nil {
				return nil, err
			}
			edges = append(edges, edge{i, j})
		}
		for _, name := range op.looser {
			j, err := target(op, "looser", name)
			if err != nil {
				return nil, err
			}
			edges = append(edges, edge{j, i})
		}
	}

	// The levels, in the order their first operator is declared, each
	// with the associativity its operators state.
	var roots []int
	levels := map[int]*precLevel{}
	stated := map[int]string{}
	for i := range keys {
		r := find(i)
		if levels[r] == nil {
			levels[r] = &precLevel{}
			roots = append(roots, r)
		}
		levels[r].ops = append(levels[r].ops, label(i))
	}
	for _, op := range ops {
		if op.assoc == "" {
			continue
		}
		i := index[key{op.name, op.fixity}]
		l := levels[find(i)]
		if l.assoc != "" && l.assoc != op.assoc {
			return nil, fmt.Errorf("%s states :assoc(%s) and %s :assoc(%s), in one level", stated[find(i)], l.assoc, label(i), op.assoc)
		}
		l.assoc, stated[find(i)] = op.assoc, label(i)
	}

	// Sort the levels: each step takes the one level no remaining level
	// is tighter than.
	looser := map[int][]int{}
	tighterCount := map[int]int{}
	touched := map[int]bool{}
	for _, e := range edges {
		t, l := find(e.tight), find(e.loose)
		looser[t] = append(looser[t], l)
		tighterCount[l]++
		touched[t], touched[l] = true, true
	}
	if len(roots) > 1 {
		for _, r := range roots {
			if !touched[r] {
				return nil, fmt.Errorf("operator %s is related to no other operator", strings.Join(levels[r].ops, ", "))
			}
		}
	}
	names := func(rs []int) string {
		var out []string
		for _, r := range rs {
			out = append(out, levels[r].ops...)
		}
		return strings.Join(out, ", ")
	}
	var order []precLevel
	remaining := slices.Clone(roots)
	for len(remaining) > 0 {
		var ready []int
		for _, r := range remaining {
			if tighterCount[r] == 0 {
				ready = append(ready, r)
			}
		}
		switch {
		case len(ready) == 0:
			return nil, fmt.Errorf("precedence relations form a cycle among %s", names(remaining))
		case len(ready) > 1:
			return nil, fmt.Errorf("precedence relations leave %s unordered", names(ready))
		}
		r := ready[0]
		order = append(order, *levels[r])
		remaining = slices.DeleteFunc(remaining, func(x int) bool { return x == r })
		for _, l := range looser[r] {
			tighterCount[l]--
		}
	}
	return order, nil
}
