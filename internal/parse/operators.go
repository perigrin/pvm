// ABOUTME: Operator declarations in .pmt files, RFC 0001 "Operator declarations".
// ABOUTME: `sub + :infix(ADD) (Num $x, Num $y) Num;` declares a symbol's fixity, class and types.

package parse

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"tamarou.com/pvm/internal/types"
)

// operatorDecl is one operator a `.pmt` declares: `sub + :infix(ADD) (Num $x,
// Num $y) Num;` is "+", "infix", "ADD" and its signature. A prefix or postfix
// operator has no class, nor an infix one at a level XS::Parse::Infix does
// not class, `sub & :infix`. multi is set for a `multi sub`, one of several
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
	if anchor := classAnchors[op.class]; anchor != "" {
		return append(slices.Clone(op.equiv), anchor)
	}
	return op.equiv
}

// classAnchors are the classes `:infix(CLASS)` may name, each to the
// operator it stands for the level of, so `:infix(CLASS)` reads as `:equiv`
// of it. RFC 0001: XS::Parse::Infix's classification (XSParseInfix.h,
// XPI_CLS_*) without the `XPI_CLS_` prefix and `_MISC` suffix. LOW and HIGH
// have no anchor: they are XS::Parse::Infix's plugin levels either side of
// the core operators, where perl has no operator.
//
// XS::Parse::Infix classes no operator at the levels of `&`, of `|` and
// `^`, of `<<` and `>>`, or of `..` and `...`, so no class names them:
// those levels are named by their operators.
var classAnchors = map[string]string{
	"LOW":             "",
	"HIGH":            "",
	"LOGICAL_OR_LOW":  "or",
	"LOGICAL_AND_LOW": "and",
	"ASSIGN":          "=",
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
	name, _ := declaredSub(n)
	// A named operator's level is its shape's, derived as its prototype
	// is; one perl places elsewhere, goto at `=`'s level, states it.
	if len(attrs) == 0 {
		rel := operatorDecl{name: name, fixity: "named"}
		if err := readRelations(n, &rel); err != nil {
			p.typedErrs = append(p.typedErrs, fmt.Errorf("sub %s: %w", name, err))
		} else if rel.tighter != nil || rel.looser != nil || rel.equiv != nil || rel.assoc != "" {
			p.relations = append(p.relations, rel)
		}
		return
	}
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
	if fixity != "infix" && hasClass {
		return operatorDecl{}, fmt.Errorf(":%s takes no class", fixity)
	}
	if _, ok := classAnchors[class]; hasClass && !ok {
		return operatorDecl{}, fmt.Errorf("unknown operator class %q", class)
	}
	if want := operatorArity[fixity]; len(sig.Params) != want {
		return operatorDecl{}, fmt.Errorf(":%s takes %s, the signature has %d", fixity, []string{"", "one operand", "two operands"}[want], len(sig.Params))
	}
	return operatorDecl{name: name, fixity: fixity, class: class, sig: sig}, nil
}

// precLevel is one level of a derived precedence order: its operators, an
// infix one and a named one stating its level by its name, and any other by
// its fixity or shape and name, `prefix -` or `unary defined`, and its
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
		if keys[i][1] == "infix" || keys[i][1] == "named" {
			return keys[i][0]
		}
		return keys[i][1] + " " + keys[i][0]
	}
	target := func(op operatorDecl, attr, name string) (int, error) {
		for _, fixity := range []string{"infix", "prefix", "postfix", "named", "unary", "listop", "term"} {
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

// classedLevels is each operator of order, by its label, to whether
// XS::Parse::Infix classes its level: whether the level holds a class's
// anchor.
func classedLevels(order []precLevel) map[string]bool {
	anchors := slices.Collect(maps.Values(classAnchors))
	out := map[string]bool{}
	for _, l := range order {
		classed := slices.ContainsFunc(l.ops, func(op string) bool { return slices.Contains(anchors, op) })
		for _, op := range l.ops {
			out[op] = classed
		}
	}
	return out
}

// shapeRelations places each shape's level, perly.y's own rows for the
// named operators: a named unary (UNIOP) between `<<` and `isa`, a list
// operator (LSTOP) below `=` -- perlop's `,` lies between, and is not yet
// declared -- and above `not`, and a term tighter than `++`.
var shapeRelations = map[Shape]operatorDecl{
	ShapeUnary:   {fixity: "unary", tighter: []string{"isa"}, looser: []string{"<<"}, assoc: "nonassoc"},
	ShapeList:    {fixity: "listop", tighter: []string{"not"}, looser: []string{"="}, assoc: "nonassoc"},
	ShapeBlock:   {fixity: "listop", tighter: []string{"not"}, looser: []string{"="}, assoc: "nonassoc"},
	ShapeNiladic: {fixity: "term", tighter: []string{"++"}, assoc: "left"},
}

// corePrecedenceDecls is what a CORE.pmt's precedence order is derived
// from: its operators' relations, its named operators' stated ones, and a
// level for each other builtin, its shape's (RFC 0001, "Precedence is a
// relation between operators"). A builtin's level is derived from its
// shape as the shape is from its prototype; a builtin that is also an
// operator, `not`, has the operator's level.
func corePrecedenceDecls(src []byte, shapes map[string]Shape) []operatorDecl {
	p := newParser(src, nil, true)
	p.buildingCore = true
	return precedenceDeclsOf(readDeclarationWith(p), shapes)
}

// precedenceDeclsOf is corePrecedenceDecls of a CORE.pmt already read.
func precedenceDeclsOf(facts moduleFacts, shapes map[string]Shape) []operatorDecl {
	out := slices.Concat(facts.operators, facts.relations)
	placed := map[string]bool{}
	for _, op := range out {
		placed[op.name] = true
	}
	first := map[string]string{}
	for _, name := range slices.Sorted(maps.Keys(shapes)) {
		if placed[name] {
			continue
		}
		rel := shapeRelations[shapes[name]]
		rel.name = name
		if anchor, ok := first[rel.fixity]; ok {
			rel = operatorDecl{name: name, fixity: rel.fixity, equiv: []string{anchor}}
		} else {
			first[rel.fixity] = name
		}
		out = append(out, rel)
	}
	return out
}

// undeclaredOperators places, by relations as their CORE.pmt lines will,
// the operators the parser reads that CORE.pmt has no line for yet. Each
// group names the issue that declares it, and leaves this list when it does.
var undeclaredOperators = slices.Concat([]operatorDecl{
	// 01a118fe-8e5f, perlop's remaining infix operators. `^^` is 5.40's
	// logical xor, sharing `||`'s level (toke.c:6441); the string-bitwise
	// operators share their numeric forms' levels; `~~` is a non-chaining
	// equality operator (cmpClasses).
	{name: "^^", fixity: "infix", equiv: []string{"||"}},
	{name: "&.", fixity: "infix", equiv: []string{"&"}},
	{name: "|.", fixity: "infix", equiv: []string{"|"}},
	{name: "^.", fixity: "infix", equiv: []string{"|"}},
	{name: "~.", fixity: "prefix", equiv: []string{"~"}},
	{name: "~~", fixity: "infix", equiv: []string{"=="}},
	// `,` lies between a rightward list operator, which swallows it, and
	// `=`; `=>` is a comma that autoquotes its left bareword.
	{name: ",", fixity: "infix", tighter: []string{"print"}, looser: []string{"="}, assoc: "left"},
	{name: "=>", fixity: "infix", equiv: []string{","}},

	// 01a11923-baeb, `?:` and `->`. The ternary is right associative,
	// measured:
	//   perl -MO=Deparse -e 'my $x = $a ? $b : $c ? $d : $e;'
	//   my $x = $a ? $b : ($c ? $d : $e);
	{name: "?", fixity: "infix", tighter: []string{"="}, looser: []string{".."}, assoc: "right"},
	{name: "->", fixity: "infix", tighter: []string{"++"}, assoc: "left"},

	// The postfix call and subscripts, perly.y's PERLY_PAREN_OPEN and its
	// brackets, the tightest operators: the terms, which `time` names the
	// level of, are tighter still. perlop's table has no row for them, and
	// no issue gives them a CORE.pmt line.
	{name: "(", fixity: "postfix", tighter: []string{"->"}, assoc: "left"},
	{name: "[", fixity: "postfix", tighter: []string{"("}, looser: []string{"time"}, assoc: "left"},
	{name: "{", fixity: "postfix", equiv: []string{"["}},
}, compoundAssignments())

// compoundAssignments are the operators `=` and an infix operator spell
// together, each in `=`'s level: they and `=` are one token class,
// toke.c:250. Declaring them is 01a118fe-8e5f's too.
func compoundAssignments() []operatorDecl {
	var out []operatorDecl
	for _, op := range []string{"+=", "-=", "*=", "/=", ".=", "%=", "**=", "x=", "||=", "&&=", "//=",
		"|=", "&=", "^=", "<<=", ">>=", "|.=", "&.=", "^.=", "^^="} {
		out = append(out, operatorDecl{name: op, fixity: "infix", equiv: []string{"="}})
	}
	return out
}

// bindingPowers are the powers the Pratt parser binds with: the infix and
// postfix operators', the prefix operators', and those a named unary's and
// a list operator's operands are parsed at.
type bindingPowers struct {
	infix              map[string]OpInfo
	prefix             map[string]int
	namedUnary, listOp int
}

// opAssoc is the parser's associativity for each `:assoc`. A chaining
// level groups leftward; which of its operators chain is cmpClasses'.
var opAssoc = map[string]Assoc{
	"left": AssocLeft, "right": AssocRight, "nonassoc": AssocNone,
	"chained": AssocLeft, "chain_na": AssocLeft,
}

// deriveBindingPowers derives the parser's binding powers from the
// precedence order decls derive: each level binds ten above the next
// looser one, the loosest at ten, so a right-associative operator's right
// operand (OpInfo.rightBP) stops only at a looser level.
func deriveBindingPowers(decls []operatorDecl) (bindingPowers, error) {
	order, err := precedenceOrder(decls)
	if err != nil {
		return bindingPowers{}, err
	}
	infixNames := map[string]bool{}
	for _, op := range decls {
		if op.fixity == "infix" {
			infixNames[op.name] = true
		}
	}
	bp := bindingPowers{infix: map[string]OpInfo{}, prefix: map[string]int{}}
	for i, l := range order {
		power := (len(order) - i) * 10
		for _, op := range l.ops {
			fixity, name, spaced := strings.Cut(op, " ")
			switch {
			case !spaced && infixNames[op]:
				bp.infix[op] = OpInfo{BP: power, Assoc: opAssoc[l.assoc]}
			case fixity == "postfix":
				bp.infix[name] = OpInfo{BP: power, Assoc: opAssoc[l.assoc]}
			case fixity == "prefix":
				bp.prefix[name] = power
			case fixity == "unary":
				bp.namedUnary = power
			case fixity == "listop":
				bp.listOp = power
			}
		}
	}
	return bp, nil
}
