// ABOUTME: Tests for precedence stated as relations between operators, RFC 0001 "Precedence is a relation between operators".
// ABOUTME: `:tighter`, `:looser`, `:equiv` and `:assoc` are read from a .pmt, and a total order is derived from them.
package parse

import (
	"reflect"
	"slices"
	"strings"
	"testing"
)

// TestCorePrecedenceRelationsAreParsed: an operator line relates itself to
// an operator it can see with `:tighter(OP)`, `:looser(OP)` and
// `:equiv(OP)`, and states its level's associativity with `:assoc(...)`.
// `:infix(CLASS)` is `:equiv` of the class's anchor operator, so `:infix(MUL)`
// puts an operator in `*`'s level.
func TestCorePrecedenceRelationsAreParsed(t *testing.T) {
	src := "sub * :infix(MUL) :tighter(+) :looser(**) :assoc(left) (Num $x, Num $y) Num;\n" +
		"sub <=> :infix(ORDERING) :equiv(==) (Num $x, Num $y) Int;\n" +
		"sub ! :prefix :tighter(=~) :assoc(right) (Any $x) Boolean;\n" +
		"sub - :prefix :equiv(!) (Num $x) Num;\n"
	facts := readDeclaration([]byte(src), nil)
	if len(facts.errs) > 0 {
		t.Fatalf("errors: %v", facts.errs)
	}
	type rel struct {
		tighter, looser, equiv []string
		assoc                  string
	}
	want := []rel{
		{tighter: []string{"+"}, looser: []string{"**"}, equiv: []string{"*"}, assoc: "left"},
		{equiv: []string{"==", "<=>"}},
		{tighter: []string{"=~"}, assoc: "right"},
		{equiv: []string{"!"}},
	}
	var got []rel
	for _, op := range facts.operators {
		got = append(got, rel{op.tighter, op.looser, op.equivs(), op.assoc})
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("relations:\n got %+v\nwant %+v", got, want)
	}

	// A named operator perl places apart from its shape states its level
	// on its own line: goto's operand is read at `=`'s level.
	named := readDeclaration([]byte("sub goto :unary :equiv(=) (Str $label) None;\n"), nil)
	if want := []operatorDecl{{name: "goto", fixity: "named", equiv: []string{"="}}}; len(named.errs) > 0 || !reflect.DeepEqual(named.relations, want) {
		t.Errorf("goto: errors %v, relations %+v, want %+v", named.errs, named.relations, want)
	}

	operatorRefuses(t, map[string]string{
		"sub + :infix(ADD) :assoc(sideways) (Num $x, Num $y) Num;\n": `sub +: unknown associativity "sideways"`,
		"sub + :infix(ADD) :tighter() (Num $x, Num $y) Num;\n":       "sub +: :tighter names no operator",
	})
}

// precedenceRefuses reads each source as a CORE.pmt and wants the one error
// given for it, the declaration error deriving its precedence order raises.
func precedenceRefuses(t *testing.T, cases map[string]string) {
	t.Helper()
	for src, want := range cases {
		if _, err := precedenceOrder(corePrecedenceDecls([]byte(src), nil)); err == nil || err.Error() != want {
			t.Errorf("%q: got error %v, want %q", src, err, want)
		}
	}
}

// TestPrecedenceRelationErrors: the relations derive one total order or
// the file is in error, naming the operators. A cycle has no order; an
// operator related to nothing has no place in it; two levels nothing
// orders against each other leave a parse without an answer; and a
// relation names an operator the file declares.
func TestPrecedenceRelationErrors(t *testing.T) {
	precedenceRefuses(t, map[string]string{
		"sub + :infix(ADD) :tighter(*) :assoc(left) (Num $x, Num $y) Num;\n" +
			"sub * :infix(MUL) :tighter(+) :assoc(left) (Num $x, Num $y) Num;\n": "precedence relations form a cycle among +, *",
		"sub * :infix(MUL) :tighter(+) :assoc(left) (Num $x, Num $y) Num;\n" +
			"sub + :infix(ADD) :assoc(left) (Num $x, Num $y) Num;\n" +
			"sub ~ :prefix :assoc(right) (Int $x) Int;\n": "operator prefix ~ is related to no other operator",
		"sub * :infix(MUL) :tighter(+) :assoc(left) (Num $x, Num $y) Num;\n" +
			"sub ** :infix(POW) :tighter(+) :assoc(right) (Num $x, Num $y) Num;\n" +
			"sub + :infix(ADD) :assoc(left) (Num $x, Num $y) Num;\n": "precedence relations leave *, ** unordered",
		"sub * :infix(MUL) :tighter(+) :assoc(left) (Num $x, Num $y) Num;\n" +
			"sub + :infix(ADD) :tighter(@@) :assoc(left) (Num $x, Num $y) Num;\n": "sub +: :tighter(@@) names no declared operator",
	})
}

// TestPrecedenceAssocConflict: a level states its associativity once, so
// two of its operators stating different ones is an error naming both.
func TestPrecedenceAssocConflict(t *testing.T) {
	precedenceRefuses(t, map[string]string{
		"sub * :infix(MUL) :tighter(+) :assoc(left) (Num $x, Num $y) Num;\n" +
			"sub + :infix(ADD) :assoc(left) (Num $x, Num $y) Num;\n" +
			"sub - :infix(ADD) :assoc(right) (Num $x, Num $y) Num;\n": "+ states :assoc(left) and - :assoc(right), in one level",
	})
}

// perlopTable is perlop's "Operator Precedence and Associativity" at the
// corpus pin (perl5 94e5086608, pod/perlop.pod), highest precedence first,
// each row's associativity in `:assoc`'s spelling (perlop's `chain/na` is
// `chain_na`). An infix operator is its symbol, a prefix or postfix one
// `prefix SYMBOL` or `postfix SYMBOL`, and a named one placed by its line
// its name. The rows perlop
// describes rather than lists, "terms and list operators (leftward)",
// "named unary operators" and "list operators (rightward)", hold a named
// operator by its shape, `term`, `unary` and `listop` (perlopShapeRows).
var perlopTable = []precLevel{
	{assoc: "left"}, // terms and list operators (leftward)
	{assoc: "left", ops: []string{"->"}},
	{assoc: "nonassoc", ops: []string{"postfix ++", "postfix --", "prefix ++", "prefix --"}},
	{assoc: "right", ops: []string{"**"}},
	{assoc: "right", ops: []string{"prefix !", "prefix ~", "prefix ~.", `prefix \`, "prefix +", "prefix -"}},
	{assoc: "left", ops: []string{"=~", "!~"}},
	{assoc: "left", ops: []string{"*", "/", "%", "x"}},
	{assoc: "left", ops: []string{"+", "-", "."}},
	{assoc: "left", ops: []string{"<<", ">>"}},
	{assoc: "nonassoc"}, // named unary operators
	{assoc: "nonassoc", ops: []string{"isa"}},
	{assoc: "chained", ops: []string{"<", ">", "<=", ">=", "lt", "gt", "le", "ge"}},
	{assoc: "chain_na", ops: []string{"==", "!=", "eq", "ne", "<=>", "cmp", "~~"}},
	{assoc: "left", ops: []string{"&", "&."}},
	{assoc: "left", ops: []string{"|", "|.", "^", "^."}},
	{assoc: "left", ops: []string{"&&"}},
	{assoc: "left", ops: []string{"||", "^^", "//"}},
	{assoc: "nonassoc", ops: []string{"..", "..."}},
	{assoc: "right", ops: []string{"?"}},
	{assoc: "right", ops: []string{"=", "+=", "-=", "*=", "/=", ".=", "%=", "**=", "x=", "&=", "|=", "^=", "<<=", ">>=", "&&=", "||=", "//=", "&.=", "|.=", "^.=", "^^=", "goto", "last", "next", "redo", "dump"}},
	{assoc: "left", ops: []string{",", "=>"}},
	{assoc: "nonassoc"}, // list operators (rightward)
	{assoc: "right", ops: []string{"prefix not"}},
	{assoc: "left", ops: []string{"and"}},
	{assoc: "left", ops: []string{"or", "xor"}},
}

// perlopShapeRows is the row of perlopTable each shape's named operators
// are in: a term (and a list operator with its parentheses) in the first,
// a named unary in its own, and a list operator without them below `,`.
var perlopShapeRows = map[string]int{"term": 0, "unary": 9, "listop": 21}

// TestCoreDerivedPrecedenceIsPerlops: the order and associativity CORE.pmt
// derives, from its operators' relations and its builtins' shapes, are
// perlop's table, row for row: each derived level's operators are all in
// one row, with that row's associativity, and the levels run down the
// table, each in a lower row than the last. The comparison is over what
// CORE.pmt declares. perlop names operators CORE.pmt does not yet declare
// -- `->`, `~.`, `~~`, `&.`, `|.`, `^.`, `^^`, `?:`, the compound
// assignments, `,` and `=>`, and last, next and redo, which are statement
// forms here -- and those have no place in the comparison until it does. Every operator CORE.pmt declares must be in perlop's table.
func TestCoreDerivedPrecedenceIsPerlops(t *testing.T) {
	rowOf := map[string]int{}
	for i, row := range perlopTable {
		for _, op := range row.ops {
			rowOf[op] = i
		}
	}
	row := func(op string) (int, bool) {
		if i, ok := rowOf[op]; ok {
			return i, true
		}
		shape, _, _ := strings.Cut(op, " ")
		i, ok := perlopShapeRows[shape]
		return i, ok
	}
	last := -1
	for _, l := range coreOrder(t) {
		want := -1
		for _, op := range l.ops {
			i, ok := row(op)
			switch {
			case !ok:
				t.Errorf("CORE.pmt declares %s, which perlop's table does not list", op)
			case want < 0:
				want = i
			case i != want:
				t.Errorf("%s is in perlop's row %d, and its level's first operator %s in row %d", op, i, l.ops[0], want)
			}
		}
		if want < 0 {
			continue
		}
		if want <= last {
			t.Errorf("level %v is perlop's row %d, after row %d", l.ops, want, last)
		}
		if l.assoc != perlopTable[want].assoc {
			t.Errorf("level %v: :assoc(%s), perlop's row %d says %s", l.ops, l.assoc, want, perlopTable[want].assoc)
		}
		last = want
	}
}

// coreOrder is the precedence order CORE.pmt derives, failing the test
// when it derives none.
func coreOrder(t *testing.T) []precLevel {
	t.Helper()
	src, ok := declaration("CORE")
	if !ok {
		t.Fatal("declarations/CORE.pmt is not embedded")
	}
	order, err := precedenceOrder(corePrecedenceDecls(src, coreShapes()))
	if err != nil {
		t.Fatal(err)
	}
	return order
}

// levelOf is the index in order of the level holding op, or -1.
func levelOf(order []precLevel, op string) int {
	for i, l := range order {
		if slices.Contains(l.ops, op) {
			return i
		}
	}
	return -1
}

// TestCoreShapeLevels: a named operator's level is derived from its shape,
// as its prototype is, and not stated (RFC 0001, "Precedence is a relation
// between operators"). Measured on 5.42.0 with -MO=Deparse,-p: `defined $x
// < 2`, `ref $x < 2`, `chdir $x < 2` and `sleep $x < 2` are `(op($x) <
// 2)`, a named unary between `<<`/`>>` and `isa`; `print STDOUT $x < 2,
// 3` is `print(STDOUT ($x < 2), 3)`, a list operator, which perlop puts
// below `,` -- CORE.pmt declares no `,` yet, so it is held below `=` and
// above `not`, its declared neighbours; `time ** 2` is `(time ** 2)`, a
// term. `not $x < 2` is `not(($x < 2))`: `not` is a prefix operator
// between `..` and `and`, though CORE.pmt also has it as a builtin. `goto
// $x = 1` is `goto ($x = 1)` and `CORE::dump $x = 1` `CORE::dump ($x =
// 1)`: the control words have no prototype that says so, and their lines
// state `:equiv(=)`.
func TestCoreShapeLevels(t *testing.T) {
	order := coreOrder(t)
	at := func(op string) int {
		i := levelOf(order, op)
		if i < 0 {
			t.Fatalf("%s: no level in %+v", op, order)
		}
		return i
	}
	between := func(op, tighter, looser string) {
		if !(at(tighter) < at(op) && at(op) < at(looser)) {
			t.Errorf("%s at level %d, want between %s (%d) and %s (%d)", op, at(op), tighter, at(tighter), looser, at(looser))
		}
	}
	for _, op := range []string{"unary defined", "unary ref", "unary chdir", "unary sleep", "unary chomp"} {
		between(op, "<<", "isa")
		if at(op) != at("unary defined") {
			t.Errorf("%s is not in defined's level", op)
		}
	}
	between("prefix not", "..", "and")
	if levelOf(order, "unary not") >= 0 {
		t.Errorf("not has a named unary's level beside its prefix operator's")
	}
	for _, op := range []string{"listop print", "listop join", "listop split", "listop grep"} {
		between(op, "=", "prefix not")
		if at(op) != at("listop print") {
			t.Errorf("%s is not in print's level", op)
		}
	}
	if at("term time") >= at("**") {
		t.Errorf("time at level %d, want tighter than ** (%d)", at("term time"), at("**"))
	}
	for _, op := range []string{"goto", "dump"} {
		if at(op) != at("=") {
			t.Errorf("%s at level %d, want ='s (%d)", op, at(op), at("="))
		}
	}
}

// parserAssoc is the associativity the Pratt loop gives each `:assoc`. A
// chaining level groups leftward, its chaining being the comparison path's
// (cmpClasses).
var parserAssoc = map[string]Assoc{
	"left": AssocLeft, "right": AssocRight, "nonassoc": AssocNone,
	"chained": AssocLeft, "chain_na": AssocLeft,
}

// TestParserPrecedenceFromCore: the binding power and associativity the
// parser uses for each operator are the order CORE.pmt's relations derive
// (RFC 0001, "Precedence is a relation between operators"). Each level's
// operators bind with one power, each level's tighter than the next, with
// its associativity: an infix or postfix operator by the infix table, a
// prefix one by the prefix table, a named unary and a list operator by the
// powers their operands are parsed at. A term takes no operand, and goto
// and dump keep their own parse, so neither has a power to compare. A
// chaining level's operators are the comparisons that chain: `chained` is
// `<` and its row, and `chain_na` `==`'s, where `<=>` and `cmp` do not chain.
func TestParserPrecedenceFromCore(t *testing.T) {
	type power struct {
		bp    int
		assoc Assoc
		infix bool
	}
	powerOf := func(op string) (power, bool) {
		shape, name, spaced := strings.Cut(op, " ")
		switch {
		case !spaced, shape == "postfix":
			if !spaced {
				name = op
			}
			info, ok := infix[name]
			return power{info.BP, info.Assoc, true}, ok
		case shape == "prefix":
			bp, ok := prefix[name]
			return power{bp: bp}, ok
		case shape == "unary":
			return power{bp: bpNamedUnary}, true
		case shape == "listop":
			return power{bp: bpListOp}, true
		}
		return power{}, false
	}
	last := 0
	for _, l := range parserOrder(t) {
		bp := 0
		for _, op := range l.ops {
			if op == "goto" || op == "dump" || strings.HasPrefix(op, "term ") {
				continue
			}
			p, ok := powerOf(op)
			if !ok {
				t.Errorf("%s: the parser has no power for it", op)
				continue
			}
			if bp == 0 {
				bp = p.bp
			} else if p.bp != bp {
				t.Errorf("%s binds at %d, and its level %v at %d", op, p.bp, l.ops, bp)
			}
			if p.infix && p.assoc != parserAssoc[l.assoc] {
				t.Errorf("%s: the parser's associativity is %v, CORE.pmt's level :assoc(%s)", op, p.assoc, l.assoc)
			}
			if cls := cmpClasses[op]; l.assoc == "chained" && cls != chRelop || l.assoc == "chain_na" && cls != chEqop && cls != ncEqop {
				t.Errorf("%s: comparison class %d in a level of :assoc(%s)", op, cls, l.assoc)
			}
			if alias, ok := infix["CORE::"+op]; ok && alias != infix[op] {
				t.Errorf("CORE::%s binds as %+v, %s as %+v", op, alias, op, infix[op])
			}
		}
		if bp == 0 {
			continue
		}
		if last != 0 && bp >= last {
			t.Errorf("level %v binds at %d, not looser than the level before it at %d", l.ops, bp, last)
		}
		last = bp
	}
}

// parserOrder is the precedence order the parser's powers are held to:
// CORE.pmt's, with the operators it has no line for, read as any parse
// reads, with the powers derived.
func parserOrder(t *testing.T) []precLevel {
	t.Helper()
	src, ok := declaration("CORE")
	if !ok {
		t.Fatal("declarations/CORE.pmt is not embedded")
	}
	order, err := precedenceOrder(slices.Concat(corePrecedenceDecls(src, coreShapes()), undeclaredOperators))
	if err != nil {
		t.Fatal(err)
	}
	return order
}
