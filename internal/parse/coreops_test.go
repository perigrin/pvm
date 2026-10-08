// ABOUTME: Tests for operator declarations in .pmt files: `sub + :infix(ADD) (Num $x, Num $y) Num;`.
// ABOUTME: Each declares its symbol, fixity, precedence class and typed signature.
package parse

import (
	"maps"
	"reflect"
	"slices"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/types"
)

// TestPmtOperatorDeclarationParses: RFC 0001 "Operator declarations". An
// operator is declared under its symbol with `:infix(CLASS)`, `:prefix` or
// `:postfix`, and its typed signature. A word operator is declared the same
// way. An unknown class is an error that names it.
func TestPmtOperatorDeclarationParses(t *testing.T) {
	src := "sub + :infix(ADD) (Num $x, Num $y) Num;\n" +
		"sub ! :prefix (Scalar $x) Boolean;\n" +
		"sub eq :infix(EQUALITY) (Str $x, Str $y) Boolean;\n"
	facts := readDeclaration([]byte(src), nil)
	if len(facts.errs) > 0 {
		t.Fatalf("errors: %v", facts.errs)
	}
	num := func(n string) types.Param { return types.Param{Name: n, Sigil: '$', Type: types.Num, Required: true} }
	str := func(n string) types.Param { return types.Param{Name: n, Sigil: '$', Type: types.Str, Required: true} }
	want := []operatorDecl{
		{name: "+", fixity: "infix", class: "ADD", sig: types.Signature{Params: []types.Param{num("x"), num("y")}, Returns: types.Num}},
		{name: "!", fixity: "prefix", sig: types.Signature{Params: []types.Param{{Name: "x", Sigil: '$', Type: types.Scalar, Required: true}}, Returns: types.Boolean}},
		{name: "eq", fixity: "infix", class: "EQUALITY", sig: types.Signature{Params: []types.Param{str("x"), str("y")}, Returns: types.Boolean}},
	}
	if !reflect.DeepEqual(facts.operators, want) {
		t.Errorf("operators:\n got %+v\nwant %+v", facts.operators, want)
	}
	// An operator is not a sub: it has no prototype, and no signature
	// under its name beside the subs'.
	for _, name := range []string{"+", "!", "eq"} {
		if p, ok := facts.protos[name]; ok {
			t.Errorf("%s recorded as a sub with prototype %q", name, p)
		}
		if s, ok := facts.signatures[name]; ok {
			t.Errorf("%s recorded as a sub signature %+v", name, s)
		}
	}

	// A keyword that is also an operator keeps both: `not` is a named
	// unary with a prototype and the prefix operator.
	both := readDeclaration([]byte("sub not (Scalar $x) Boolean;\nsub not :prefix (Scalar $x) Boolean;\n"), nil)
	if len(both.errs) > 0 || len(both.signatures["not"]) != 1 || len(both.operators) != 1 {
		t.Errorf("not: errors %v, sub signatures %+v, operators %+v", both.errs, both.signatures["not"], both.operators)
	}

	operatorRefuses(t, map[string]string{
		"sub + :infix(PLUS) (Num $x, Num $y) Num;\n": `sub +: unknown operator class "PLUS"`,
	})
}

// operatorRefuses reads each source as a `.pmt` and wants exactly the one
// error given for it, no operator recorded, and every byte still in the tree.
func operatorRefuses(t *testing.T, cases map[string]string) {
	t.Helper()
	for src, want := range cases {
		facts := readDeclaration([]byte(src), nil)
		if len(facts.errs) != 1 || facts.errs[0].Error() != want {
			t.Errorf("%q: got errors %v, want %q", src, facts.errs, want)
		}
		if len(facts.operators) > 0 {
			t.Errorf("%q: recorded %+v", src, facts.operators)
		}
		if root, _ := parseSource([]byte(src), nil, true); root.SourceText([]byte(src)) != src {
			t.Errorf("%q: round trip lost bytes", src)
		}
	}
}

// TestPmtOperatorFixityErrors: fixity is spelled one way. An infix operator
// may name its class, a prefix one names none, and a declaration has one
// fixity.
func TestPmtOperatorFixityErrors(t *testing.T) {
	operatorRefuses(t, map[string]string{
		"sub ! :prefix(ADD) (Scalar $x) Boolean;\n":         "sub !: :prefix takes no class",
		"sub + :infix(ADD) :prefix (Num $x, Num $y) Num;\n": "sub +: two fixities, :infix(ADD) and :prefix",
	})
}

// TestPmtOperatorArityMatchesFixity: an infix operator takes two operands
// and a prefix one takes one, so its signature has that many parameters.
func TestPmtOperatorArityMatchesFixity(t *testing.T) {
	operatorRefuses(t, map[string]string{
		"sub + :infix(ADD) (Num $x) Num;\n":       "sub +: :infix takes two operands, the signature has 1",
		"sub - :prefix (Num $x, Num $y) Num;\n":   "sub -: :prefix takes one operand, the signature has 2",
		"sub ++ :postfix (Num $x, Num $y) Num;\n": "sub ++: :postfix takes one operand, the signature has 2",
	})
}

// TestPmtOperatorAttributeAfterSignatureRefused: declaration order holds for
// operators as for subs (RFC 0001, "Declaration order: Perl's"): the
// attributes come before the signature.
func TestPmtOperatorAttributeAfterSignatureRefused(t *testing.T) {
	operatorRefuses(t, map[string]string{
		"sub + (Num $x, Num $y) :infix(ADD) Num;\n": "sub +: subroutine attributes must come before the signature",
	})
}

// TestOperatorSymbolOnlyInPmt: a symbol names an operator only in a `.pmt`.
// In perl, measured on 5.42.0, `sub + { 1 }` is "Illegal declaration of
// anonymous subroutine", and ordinary source declares nothing by it.
func TestOperatorSymbolOnlyInPmt(t *testing.T) {
	for _, src := range []string{"sub + { 1 }\n", "sub + :infix(ADD) (Num $x, Num $y) Num;\n"} {
		facts := readModule(Parse([]byte(src)))
		if len(facts.protos) > 0 || len(facts.operators) > 0 {
			t.Errorf("%q: ordinary source declared subs %v, operators %v", src, facts.protos, facts.operators)
		}
	}
}

// TestCoreOperatorClassCheckCatchesMismatch: a class is a relation, so one
// its operator's other relations contradict is a declaration error. `*`
// declared ADD while tighter than `+` would be tighter than its own level,
// and declared MUL it is not.
func TestCoreOperatorClassCheckCatchesMismatch(t *testing.T) {
	precedenceRefuses(t, map[string]string{
		"sub + :infix(ADD) :assoc(left) (Num $x, Num $y) Num;\n" +
			"sub * :infix(ADD) :tighter(+) :assoc(left) (Num $x, Num $y) Num;\n": "precedence relations form a cycle among +, *",
	})
	ok := "sub + :infix(ADD) :assoc(left) (Num $x, Num $y) Num;\n" +
		"sub * :infix(MUL) :tighter(+) :assoc(left) (Num $x, Num $y) Num;\n"
	if _, err := precedenceOrder(corePrecedenceDecls([]byte(ok), nil)); err != nil {
		t.Errorf("MUL: got %v, want none", err)
	}
}

// xpiClasses are XS::Parse::Infix's classes (XSParseInfix.h, XPI_CLS_*),
// spelled as RFC 0001 "Fixity and precedence" spells them.
var xpiClasses = []string{"LOW", "LOGICAL_OR_LOW", "LOGICAL_AND_LOW", "ASSIGN", "LOGICAL_OR", "LOGICAL_AND",
	"EQUALITY", "ORDERING", "RELATION", "ISA", "ADD", "MUL", "MATCHRE", "POW", "HIGH"}

// TestCoreClassesAreXPIs: `:infix(CLASS)` names only XS::Parse::Infix's
// classes (RFC 0001, "Fixity and precedence"), so every class CORE.pmt
// names is one, and the levels it classes no operator at are named by
// their operators: no BITAND, BITOR, SHIFT or RANGE remains.
func TestCoreClassesAreXPIs(t *testing.T) {
	if got := slices.Sorted(maps.Keys(classAnchors)); !reflect.DeepEqual(got, slices.Sorted(slices.Values(xpiClasses))) {
		t.Errorf("classAnchors are %v, want XS::Parse::Infix's %v", got, slices.Sorted(slices.Values(xpiClasses)))
	}
	for _, op := range coreOperators(t) {
		if op.class != "" && !slices.Contains(xpiClasses, op.class) {
			t.Errorf("sub %s: class %s is not XS::Parse::Infix's", op.name, op.class)
		}
	}
	src, _ := declaration("CORE")
	for _, coined := range []string{"BITAND", "BITOR", "SHIFT", "RANGE"} {
		if strings.Contains(string(src), ":infix("+coined+")") {
			t.Errorf("CORE.pmt names the class %s", coined)
		}
	}
}

// coreOperators reads CORE.pmt's operator declarations, failing the test on
// any error in the file.
func coreOperators(t *testing.T) []operatorDecl {
	t.Helper()
	src, ok := declaration("CORE")
	if !ok {
		t.Fatal("declarations/CORE.pmt is not embedded")
	}
	facts := readDeclaration(src, nil)
	if len(facts.errs) > 0 {
		t.Fatalf("CORE.pmt: %v", facts.errs)
	}
	return facts.operators
}

// TestCoreOperatorTypesMatchMeasured: CORE.pmt types perl's operators with
// the operand and result types below, measured on 5.42.0. `cmp` gives -1, 0
// or 1, and so does `<=>`, or undef when an operand is NaN; the predicates
// give perl's booleans; `&&`, `||`, `//`, `and`, `or` and `=` give an
// operand, and `^^`, `~~` and `^^=` perl's boolean; the string bitwise
// operators give a string; a compound assignment gives its left operand,
// holding what its operator gives. An operator that forks has a row for each candidate, its list
// candidate's keyed "(list)": the one whose return type answers list
// context (RFC 0001, "Context selects by return type").
//
// Arithmetic on Nums can leave Num, which excludes Inf and NaN, and its
// result says so: measured, `1e308*10` is Inf and `(-1)**0.5` is NaN
// (TestCoreArithmeticReturnsWhatPerlReturns).
//
// The rows are golden values, held here rather than read from anywhere
// else, so the declarations answer to them alone.
func TestCoreOperatorTypesMatchMeasured(t *testing.T) {
	type row struct {
		operands []types.Type
		result   types.Type
	}
	bin := func(l, r, res types.Type) row { return row{[]types.Type{l, r}, res} }
	un := func(o, res types.Type) row { return row{[]types.Type{o}, res} }
	N, S, I, B, A := types.Num, types.Str, types.Int, types.Boolean, types.Any
	want := map[string]row{
		"infix +": bin(N, N, N|types.Inf), "infix -": bin(N, N, N|types.Inf), "infix *": bin(N, N, N|types.Inf),
		"infix /": bin(N, N, N|types.Inf), "infix %": bin(N, N, N), "infix **": bin(N, N, N|types.NaN|types.Inf),

		"infix .": bin(S, S, S),

		"infix x (list)": bin(types.List, I, types.List), "infix x": bin(S, I, S),

		"infix ==": bin(N, N, B), "infix !=": bin(N, N, B), "infix <": bin(N, N, B),
		"infix >": bin(N, N, B), "infix <=": bin(N, N, B), "infix >=": bin(N, N, B),
		"infix <=>": bin(N, N, I|types.Undef),

		"infix eq": bin(S, S, B), "infix ne": bin(S, S, B), "infix lt": bin(S, S, B),
		"infix gt": bin(S, S, B), "infix le": bin(S, S, B), "infix ge": bin(S, S, B),
		"infix cmp": bin(S, S, I),

		"infix &&": bin(A, A, A), "infix ||": bin(A, A, A), "infix //": bin(A, A, A),
		"infix and": bin(A, A, A), "infix or": bin(A, A, A), "infix xor": bin(A, A, B),

		"infix &": bin(I, I, I), "infix |": bin(I, I, I), "infix ^": bin(I, I, I),
		"infix <<": bin(I, I, I), "infix >>": bin(I, I, I),

		"infix isa": bin(types.Scalar, S, B),

		"infix =~ (list)": bin(S, types.Regex, types.List), "infix =~": bin(S, types.Regex, B|S),
		"infix !~": bin(S, types.Regex, B),

		"infix =": bin(A, A, A),

		"infix .. (list)": bin(S, S, types.List), "infix ..": bin(A, A, S),
		"infix ... (list)": bin(S, S, types.List), "infix ...": bin(A, A, S),

		"prefix -": un(N, N), "prefix +": un(N, N), "prefix !": un(A, B),
		"prefix not": un(A, B), "prefix ~": un(I, I), `prefix \`: un(A, types.Ref),
		`prefix \ (list)`: un(types.List, types.List),

		"prefix ++": un(types.Scalar, S), "prefix --": un(types.Scalar, N|types.NaN|types.Inf),
		"postfix ++": un(types.Scalar, B|S|types.DualVar|types.Ref), "postfix --": un(types.Scalar, types.Scalar),

		"infix ^^": bin(A, A, B), "infix ~~": bin(A, A, B),
		"infix &.": bin(S, S, S), "infix |.": bin(S, S, S), "infix ^.": bin(S, S, S), "prefix ~.": un(S, S),

		"infix +=": bin(types.Scalar, N, N|types.Inf), "infix -=": bin(types.Scalar, N, N|types.Inf),
		"infix *=": bin(types.Scalar, N, N|types.Inf), "infix /=": bin(types.Scalar, N, N|types.Inf),
		"infix **=": bin(types.Scalar, N, N|types.NaN|types.Inf), "infix %=": bin(types.Scalar, N, N),
		"infix .=": bin(types.Scalar, S, S), "infix x=": bin(types.Scalar, I, S),
		"infix &=": bin(types.Scalar, I, I), "infix |=": bin(types.Scalar, I, I), "infix ^=": bin(types.Scalar, I, I),
		"infix <<=": bin(types.Scalar, I, I), "infix >>=": bin(types.Scalar, I, I),
		"infix &.=": bin(types.Scalar, S, S), "infix |.=": bin(types.Scalar, S, S), "infix ^.=": bin(types.Scalar, S, S),
		"infix &&=": bin(types.Scalar, types.Scalar, types.Scalar), "infix ||=": bin(types.Scalar, types.Scalar, types.Scalar),
		"infix //=": bin(types.Scalar, types.Scalar, types.Scalar), "infix ^^=": bin(types.Scalar, types.Scalar, B),
	}
	got := map[string]row{}
	for _, op := range coreOperators(t) {
		r := row{result: op.sig.Returns}
		for _, p := range op.sig.Params {
			r.operands = append(r.operands, p.Type)
		}
		key := op.fixity + " " + op.name
		if op.multi && op.sig.Returns&(types.Array|types.Hash) != 0 {
			key += " (list)"
		}
		if _, dup := got[key]; dup {
			t.Errorf("%s: declared twice", key)
		}
		got[key] = r
	}
	for key, w := range want {
		g, ok := got[key]
		if !ok {
			t.Errorf("%s: not declared in CORE.pmt", key)
			continue
		}
		if !reflect.DeepEqual(g, w) {
			t.Errorf("%s: CORE.pmt declares %v -> %v, measured %v -> %v", key, g.operands, g.result, w.operands, w.result)
		}
	}
	for key := range got {
		if _, ok := want[key]; !ok {
			t.Errorf("%s: declared in CORE.pmt with no measured row", key)
		}
	}
}

// TestCoreArithmeticReturnsWhatPerlReturns: a numeric operator or builtin's
// operands are Num, which excludes Inf and NaN, but its result is what perl
// returns for them (perigrin, 2026-10-07; the paper's [Plus] rule types the
// operands through Num and the result only as a number). Measured on 5.42.0
// with finite operands, Inf and -Inf alike being the lattice's unsigned Inf.
// `+`, `-`, `*` and `/` overflow but never give NaN: `1e308 + 1e308`,
// `-1e308 - 1e308`, `-1e308 * 10` and `1e308 / 1e-308` are Inf or -Inf, and
// x/0 dies "Illegal division by zero". `**` gives both: `10 ** 400` and
// `0 ** -1` are Inf, `(-10) ** 401` -Inf, `(-1) ** 0.5` and `(-8) ** (1/3)`
// NaN. `%` is finite, `1e308 % 7` being 3 and `7.5 % -1e308` -1e308, and a
// modulus truncating to 0 dies "Illegal modulus zero". Prefix `-` of 1e308 is
// -1e308, and prefix `+` is its operand. exp(1000) is Inf. hex("f" x 256)
// and oct("7" x 400) are Inf, with an overflow warning, and returned. The
// rest stay finite: abs(-1e308), int(-1e308), sqrt(1e308), log(5e-324),
// sin(1e308), cos(-1e308), atan2(1e308, 1e-308), atan2(0, 0), rand(1e308)
// and rand(-1e308); sqrt(-1), log(0) and log(-1) die, which is no value.
func TestCoreArithmeticReturnsWhatPerlReturns(t *testing.T) {
	returns := func(sigs []types.Signature) []types.Type {
		var out []types.Type
		for _, s := range sigs {
			out = append(out, s.Returns)
		}
		return out
	}
	N, I := types.Num, types.Int
	operators := map[string]types.Type{
		"+": N | types.Inf, "-": N | types.Inf, "*": N | types.Inf, "/": N | types.Inf,
		"**": N | types.NaN | types.Inf, "%": N,
	}
	for op, want := range operators {
		sigs := CoreOperator(op, "infix")
		if len(sigs) != 1 || sigs[0].Returns != want {
			t.Errorf("infix %s: CORE.pmt declares results %v, measured %v", op, returns(sigs), want)
		}
	}
	for _, op := range []string{"-", "+"} {
		if sigs := CoreOperator(op, "prefix"); len(sigs) != 1 || sigs[0].Returns != N {
			t.Errorf("prefix %s: CORE.pmt declares results %v, measured Num", op, returns(sigs))
		}
	}
	builtins := map[string]types.Type{
		"exp": N | types.Inf, "hex": I | types.Inf, "oct": I | types.Inf,
		"abs": N, "int": I, "sqrt": N, "log": N, "sin": N, "cos": N, "atan2": N, "rand": N,
	}
	for name, want := range builtins {
		sigs := CoreBuiltin(name)
		if len(sigs) != 1 || sigs[0].Returns != want {
			t.Errorf("%s: CORE.pmt declares results %v, measured %v", name, returns(sigs), want)
		}
	}
}

// TestCoreOperatorClassesMatchPrecedence: each infix operator CORE.pmt
// declares with a class binds, in the parser, as the operator its class
// stands for: `/`, declared MUL, as `*`.
func TestCoreOperatorClassesMatchPrecedence(t *testing.T) {
	coreShapes()
	n := 0
	for _, op := range coreOperators(t) {
		anchor := classAnchors[op.class]
		if op.fixity != "infix" || anchor == "" {
			continue
		}
		n++
		if infix[op.name] != infix[anchor] {
			t.Errorf("sub %s: class %s, and it binds as %+v where %s binds as %+v", op.name, op.class, infix[op.name], anchor, infix[anchor])
		}
	}
	if n == 0 {
		t.Error("CORE.pmt declares no classed infix operator")
	}
}

// coreCandidates are CORE.pmt's declarations of the operator sym with the
// given fixity, in the order the file states them.
func coreCandidates(t *testing.T, fixity, sym string) []operatorDecl {
	t.Helper()
	var out []operatorDecl
	for _, op := range coreOperators(t) {
		if op.fixity == fixity && op.name == sym {
			out = append(out, op)
		}
	}
	return out
}

// TestCoreRangeIsContextMulti: RFC 0001 "Operators that fork". `..` is a
// range in list context and a flip-flop in scalar context, so CORE.pmt
// declares it as a multi whose return types fork on context, and `...`
// likewise. Measured on
// 5.42.0: `my @l = ("a".."c")` is `a b c` and `(1.7..3.2)` is `1 2 3`; in
// scalar context `/x/ .. /y/` over `a x b c y d` gives "", "1", "2", "3",
// "4E0", "", each a string. `...` gives the same in both.
func TestCoreRangeIsContextMulti(t *testing.T) {
	param := func(n string, ty types.Type) types.Param {
		return types.Param{Name: n, Sigil: '$', Type: ty, Required: true}
	}
	for _, sym := range []string{"..", "..."} {
		want := []operatorDecl{
			{name: sym, fixity: "infix", multi: true, sig: types.Signature{
				Params: []types.Param{param("x", types.Str), param("y", types.Str)}, Returns: types.List}},
			{name: sym, fixity: "infix", multi: true, sig: types.Signature{
				Params: []types.Param{param("x", types.Any), param("y", types.Any)}, Returns: types.Str}},
		}
		// XS::Parse::Infix classes no operator at their level, so it is
		// named by `..`, whose first candidate states the level's relation
		// and associativity, and `...`'s first joins it.
		if sym == ".." {
			want[0].tighter, want[0].assoc = []string{"="}, "nonassoc"
		} else {
			want[0].equiv = []string{".."}
		}
		if got := coreCandidates(t, "infix", sym); !reflect.DeepEqual(got, want) {
			t.Errorf("%s:\n got %+v\nwant %+v", sym, got, want)
		}
	}
}

// TestCoreRepeatIsShapeMulti: RFC 0001 "Operators that fork". `x` repeats
// a list only when its left operand is parenthesised and the call is in
// list context, so CORE.pmt declares two MUL candidates, `(LIST) x N` for
// list context and `EXPR x N` for any. Measured on 5.42.0: `my @l = (1,2)
// x 2` is `1 2 1 2`; `my $x = (1,2) x 2`, `my @l = @a x 2` and `"ab" x 2`
// are strings, `22`, `22` and `abab`. The List operand is one operand, a
// parenthesised list, so an operator's `@` parameter may come first.
func TestCoreRepeatIsShapeMulti(t *testing.T) {
	want := []operatorDecl{
		{name: "x", fixity: "infix", class: "MUL", multi: true, sig: types.Signature{
			Params: []types.Param{
				{Name: "l", Sigil: '@', Type: types.List},
				{Name: "n", Sigil: '$', Type: types.Int, Required: true},
			},
			Returns: types.List}},
		{name: "x", fixity: "infix", class: "MUL", multi: true, sig: types.Signature{
			Params: []types.Param{
				{Name: "s", Sigil: '$', Type: types.Str, Required: true},
				{Name: "n", Sigil: '$', Type: types.Int, Required: true},
			},
			Returns: types.Str}},
	}
	if got := coreCandidates(t, "infix", "x"); !reflect.DeepEqual(got, want) {
		t.Errorf("x:\n got %+v\nwant %+v", got, want)
	}
}

// TestPmtListParameterLastOnlyForSubs: a sub's List parameter takes every
// remaining argument, so nothing may follow it; an operator's is one
// operand, and may.
func TestPmtListParameterLastOnlyForSubs(t *testing.T) {
	facts := readDeclaration([]byte("sub x :infix(MUL) (List @l, Int $n) List;\n"), nil)
	if len(facts.errs) > 0 || len(facts.operators) != 1 {
		t.Errorf("operator: errors %v, operators %+v", facts.errs, facts.operators)
	}
	facts = readDeclaration([]byte("sub f (List @l, Int $n) List;\n"), nil)
	if len(facts.errs) != 1 {
		t.Errorf("sub: errors %v, want the List parameter refused", facts.errs)
	}
}

// TestCoreMatchIsContextMulti: `=~` forks on context. Measured on 5.42.0,
// in list context `("ab" =~ qr/(a)(b)/)` is the captures `a b`, a match
// with none is `(1)` or `()`, and `s///g` is its count. In scalar context
// a match is perl's boolean, `s///g` and `tr///` a count (3), and `s///r`
// and `tr///r` the new string ("baa"): a Boolean or a Str.
func TestCoreMatchIsContextMulti(t *testing.T) {
	params := []types.Param{
		{Name: "x", Sigil: '$', Type: types.Str, Required: true},
		{Name: "y", Sigil: '$', Type: types.Regex, Required: true},
	}
	want := []operatorDecl{
		{name: "=~", fixity: "infix", class: "MATCHRE", multi: true, tighter: []string{"*"}, assoc: "left", sig: types.Signature{
			Params: params, Returns: types.List}},
		{name: "=~", fixity: "infix", class: "MATCHRE", multi: true, sig: types.Signature{
			Params: params, Returns: types.Boolean | types.Str}},
	}
	if got := coreCandidates(t, "infix", "=~"); !reflect.DeepEqual(got, want) {
		t.Errorf("=~:\n got %+v\nwant %+v", got, want)
	}
}

// TestCoreRefgenIsShapeMulti: `\` forks on a parenthesis as `x` does.
// Measured on 5.42.0, `my @r = \(1,2,3)` and `my @r = \(@a)` are three
// SCALAR references, while `my @r = \@a` is one ARRAY reference, and
// `my $r = \(1,2,3)` is a reference to the 3.
func TestCoreRefgenIsShapeMulti(t *testing.T) {
	want := []operatorDecl{
		{name: `\`, fixity: "prefix", multi: true, equiv: []string{"!"}, sig: types.Signature{
			Params:  []types.Param{{Name: "l", Sigil: '@', Type: types.List}},
			Returns: types.List}},
		{name: `\`, fixity: "prefix", multi: true, sig: types.Signature{
			Params:  []types.Param{{Name: "x", Sigil: '$', Type: types.Any, Required: true}},
			Returns: types.Ref}},
	}
	if got := coreCandidates(t, "prefix", `\`); !reflect.DeepEqual(got, want) {
		t.Errorf(`\:`+"\n got %+v\nwant %+v", got, want)
	}
}

// TestCoreDeclaresIncDec: `++` and `--` are each prefix and postfix, in one
// level perlop's table makes `nonassoc` and puts above `**`, on an aliased
// scalar: perl refuses `++1`, "Can't modify constant item in preincrement".
// Measured on 5.42.0, the operand is any scalar: undef, a number, a string,
// a reference. Prefix `++` gives the new value: a number, or a string's
// magic increment ("aa" is "ab", "Az" "Ba", "zz" "aaa", "a9" "b0"), Str
// covering both. Prefix `--` has no string magic: "aa" is -1, and "inf"
// Inf and "nan" NaN. Postfix gives the old value, but postfix `++` gives 0
// for undef where postfix `--` gives undef.
func TestCoreDeclaresIncDec(t *testing.T) {
	operand := []types.Param{{Name: "x", Sigil: '$', Type: types.Scalar, Required: true, Alias: true}}
	want := map[[2]string][]operatorDecl{
		{"prefix", "++"}: {{name: "++", fixity: "prefix", tighter: []string{"**"}, assoc: "nonassoc",
			sig: types.Signature{Params: operand, Returns: types.Str}}},
		{"prefix", "--"}: {{name: "--", fixity: "prefix", equiv: []string{"++"},
			sig: types.Signature{Params: operand, Returns: types.Num | types.NaN | types.Inf}}},
		{"postfix", "++"}: {{name: "++", fixity: "postfix", equiv: []string{"++"},
			sig: types.Signature{Params: operand, Returns: types.Boolean | types.Str | types.DualVar | types.Ref}}},
		{"postfix", "--"}: {{name: "--", fixity: "postfix", equiv: []string{"++"},
			sig: types.Signature{Params: operand, Returns: types.Scalar}}},
	}
	for key, w := range want {
		if got := coreCandidates(t, key[0], key[1]); !reflect.DeepEqual(got, w) {
			t.Errorf("%s %s:\n got %+v\nwant %+v", key[0], key[1], got, w)
		}
	}
	for _, op := range undeclaredOperators {
		if op.name == "++" || op.name == "--" {
			t.Errorf("undeclaredOperators still places %s %s", op.fixity, op.name)
		}
	}
}

// TestCoreDeclaresRemainingInfix: perlop's operators past the plain ones.
// Measured on 5.42.0: `^^` is 5.40's logical xor, in `||`'s level, and
// gives perl's boolean (builtin::is_bool) for any operands. The string
// bitwise operators, under `use feature 'bitwise'`, take their operands as
// strings and give a string -- `3 &. 5` is "1" -- in their numeric forms'
// levels, and `~.` is `~`'s. Smartmatch `~~` is in `==`'s level and gives
// perl's boolean; it warns nothing on 5.42 without `use v5.42`, whose bundle
// turns the smartmatch feature off and makes `1 ~~ 1` a syntax error. Each
// compound assignment is in `=`'s level, its left operand an aliased scalar
// -- `1 += 2` is "Can't modify constant item in addition (+)", `@a x= 3`
// "Can't modify private array in repeat (x)" -- and its result is that
// scalar, `\($x += 1)` aliasing $x, holding what its operator gives. The
// logical ones take their right operand in scalar context: `$x ||= (4, 5)`
// is 5.
func TestCoreDeclaresRemainingInfix(t *testing.T) {
	param := func(n string, ty types.Type) types.Param {
		return types.Param{Name: n, Sigil: '$', Type: ty, Required: true}
	}
	pair := func(l, r types.Type) []types.Param { return []types.Param{param("x", l), param("y", r)} }
	want := map[[2]string][]operatorDecl{
		{"infix", "^^"}: {{name: "^^", fixity: "infix", class: "LOGICAL_OR",
			sig: types.Signature{Params: pair(types.Any, types.Any), Returns: types.Boolean}}},
		{"infix", "&."}: {{name: "&.", fixity: "infix", equiv: []string{"&"},
			sig: types.Signature{Params: pair(types.Str, types.Str), Returns: types.Str}}},
		{"infix", "|."}: {{name: "|.", fixity: "infix", equiv: []string{"|"},
			sig: types.Signature{Params: pair(types.Str, types.Str), Returns: types.Str}}},
		{"infix", "^."}: {{name: "^.", fixity: "infix", equiv: []string{"|"},
			sig: types.Signature{Params: pair(types.Str, types.Str), Returns: types.Str}}},
		{"prefix", "~."}: {{name: "~.", fixity: "prefix", equiv: []string{"!"},
			sig: types.Signature{Params: []types.Param{param("x", types.Str)}, Returns: types.Str}}},
		{"infix", "~~"}: {{name: "~~", fixity: "infix", class: "EQUALITY",
			sig: types.Signature{Params: pair(types.Any, types.Any), Returns: types.Boolean}}},
	}
	aliased := types.Param{Name: "x", Sigil: '$', Type: types.Scalar, Required: true, Alias: true}
	for _, c := range []struct {
		ops           []string
		right, result types.Type
	}{
		{[]string{"+=", "-=", "*=", "/="}, types.Num, types.Num | types.Inf},
		{[]string{"**="}, types.Num, types.Num | types.NaN | types.Inf},
		{[]string{"%="}, types.Num, types.Num},
		{[]string{".=", "&.=", "|.=", "^.="}, types.Str, types.Str},
		{[]string{"x="}, types.Int, types.Str},
		{[]string{"&=", "|=", "^=", "<<=", ">>="}, types.Int, types.Int},
		{[]string{"&&=", "||=", "//="}, types.Scalar, types.Scalar},
		{[]string{"^^="}, types.Scalar, types.Boolean},
	} {
		for _, op := range c.ops {
			want[[2]string{"infix", op}] = []operatorDecl{{name: op, fixity: "infix", class: "ASSIGN",
				sig: types.Signature{Params: []types.Param{aliased, param("y", c.right)}, Returns: c.result}}}
		}
	}
	for key, w := range want {
		if got := coreCandidates(t, key[0], key[1]); !reflect.DeepEqual(got, w) {
			t.Errorf("%s %s:\n got %+v\nwant %+v", key[0], key[1], got, w)
		}
	}
}
