// ABOUTME: Tests for precedence stated as relations between operators, RFC 0001 "Precedence is a relation between operators".
// ABOUTME: `:tighter`, `:looser`, `:equiv` and `:assoc` are read from a .pmt, and a total order is derived from them.
package parse

import (
	"reflect"
	"slices"
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
		if _, err := precedenceOrder(coreOperatorDecls([]byte(src))); err == nil || err.Error() != want {
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
// `chain_na`). An infix operator is its symbol and a prefix one `prefix
// SYMBOL`. Rows perlop describes rather than lists hold no operator:
// "terms and list operators (leftward)", "named unary operators" and "list
// operators (rightward)"; the operators perlop puts in the `=` row besides
// the assignments, goto last next redo dump, are named unaries' shape, not
// operators, and are not listed.
var perlopTable = []precLevel{
	{assoc: "left"}, // terms and list operators (leftward)
	{assoc: "left", ops: []string{"->"}},
	{assoc: "nonassoc", ops: []string{"++", "--", "prefix ++", "prefix --"}},
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
	{assoc: "right", ops: []string{"=", "+=", "-=", "*=", "/=", ".=", "%=", "**=", "x=", "&=", "|=", "^=", "<<=", ">>=", "&&=", "||=", "//=", "&.=", "|.=", "^.=", "^^="}},
	{assoc: "left", ops: []string{",", "=>"}},
	{assoc: "nonassoc"}, // list operators (rightward)
	{assoc: "right", ops: []string{"prefix not"}},
	{assoc: "left", ops: []string{"and"}},
	{assoc: "left", ops: []string{"or", "xor"}},
}

// TestCoreDerivedPrecedenceIsPerlops: the order and associativity CORE.pmt's
// relations derive are perlop's table, row for row. The comparison is over
// the operators CORE.pmt declares: perlop's rows are cut to those, and a
// row left empty drops out. perlop names operators CORE.pmt does not yet
// declare -- `->`, `++` and `--`, `~.`, `~~`, `&.`, `|.`, `^.`, `^^`, `?:`,
// the compound assignments, `,` and `=>` -- and its named unary and list
// operator rows are no operator's, so those have no place in the
// comparison until CORE.pmt declares them. Every operator CORE.pmt
// declares must be in perlop's table.
func TestCoreDerivedPrecedenceIsPerlops(t *testing.T) {
	got, err := precedenceOrder(coreOperators(t))
	if err != nil {
		t.Fatal(err)
	}
	declared := map[string]bool{}
	for _, l := range got {
		for _, op := range l.ops {
			declared[op] = true
		}
	}
	listed := map[string]bool{}
	var want []precLevel
	for _, row := range perlopTable {
		var ops []string
		for _, op := range row.ops {
			listed[op] = true
			if declared[op] {
				ops = append(ops, op)
			}
		}
		if len(ops) > 0 {
			want = append(want, precLevel{ops: ops, assoc: row.assoc})
		}
	}
	for op := range declared {
		if !listed[op] {
			t.Errorf("CORE.pmt declares %s, which perlop's table does not list", op)
		}
	}
	// Within a row the order is the declaration's, so compare rows as sets.
	norm := func(ls []precLevel) []precLevel {
		var out []precLevel
		for _, l := range ls {
			out = append(out, precLevel{ops: slices.Sorted(slices.Values(l.ops)), assoc: l.assoc})
		}
		return out
	}
	if !reflect.DeepEqual(norm(got), norm(want)) {
		t.Errorf("derived order:\n got %+v\nwant %+v", norm(got), norm(want))
	}
}
