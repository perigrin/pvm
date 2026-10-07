// ABOUTME: Tests for operator declarations in .pmt files: `sub + :infix(ADD) (Num $x, Num $y) Num;`.
// ABOUTME: Each declares its symbol, fixity, precedence class and typed signature.
package parse

import (
	"reflect"
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
// names its class, a prefix one names none, and a declaration has one
// fixity.
func TestPmtOperatorFixityErrors(t *testing.T) {
	operatorRefuses(t, map[string]string{
		"sub + :infix (Num $x, Num $y) Num;\n":              "sub +: :infix needs a class",
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

// TestCoreOperatorClassCheckCatchesMismatch: the check that holds a declared
// class to precedence.go is not tautological. `*` is a MULOP, so a `*`
// declared ADD is reported, and one declared MUL is not.
func TestCoreOperatorClassCheckCatchesMismatch(t *testing.T) {
	facts := readDeclaration([]byte("sub * :infix(ADD) (Num $x, Num $y) Num;\nsub * :infix(MUL) (Num $x, Num $y) Num;\n"), nil)
	if len(facts.errs) > 0 || len(facts.operators) != 2 {
		t.Fatalf("errors %v, operators %+v", facts.errs, facts.operators)
	}
	want := "sub *: class ADD is level 22, and precedence.go puts * at level 23"
	if err := precedenceMismatch(facts.operators[0]); err == nil || err.Error() != want {
		t.Errorf("ADD: got %v, want %q", err, want)
	}
	if err := precedenceMismatch(facts.operators[1]); err != nil {
		t.Errorf("MUL: got %v, want none", err)
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
// operand. An operator that forks has a row for each candidate, keyed by
// the `:context` it states.
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

		"infix x :context(@)": bin(types.List, I, types.List), "infix x": bin(S, I, S),

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

		"infix =~ :context(@)": bin(S, types.Regex, types.List), "infix =~ :context($)": bin(S, types.Regex, B|S),
		"infix !~": bin(S, types.Regex, B),

		"infix =": bin(A, A, A),

		"infix .. :context(@)": bin(S, S, types.List), "infix .. :context($)": bin(A, A, S),
		"infix ... :context(@)": bin(S, S, types.List), "infix ... :context($)": bin(A, A, S),

		"prefix -": un(N, N), "prefix +": un(N, N), "prefix !": un(A, B),
		"prefix not": un(A, B), "prefix ~": un(I, I), `prefix \`: un(A, types.Ref),
		`prefix \ :context(@)`: un(types.List, types.List),
	}
	got := map[string]row{}
	for _, op := range coreOperators(t) {
		r := row{result: op.sig.Returns}
		for _, p := range op.sig.Params {
			r.operands = append(r.operands, p.Type)
		}
		key := op.fixity + " " + op.name
		switch op.sig.Context {
		case types.ContextSet(types.ListCtx):
			key += " :context(@)"
		case types.ContextSet(types.ScalarCtx):
			key += " :context($)"
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
// declares names, by its class, the level precedence.go gives it. The table
// is authoritative (RFC 0001, "Operator declarations").
func TestCoreOperatorClassesMatchPrecedence(t *testing.T) {
	n := 0
	for _, op := range coreOperators(t) {
		if op.fixity != "infix" {
			continue
		}
		n++
		if err := precedenceMismatch(op); err != nil {
			t.Error(err)
		}
	}
	if n == 0 {
		t.Error("CORE.pmt declares no infix operator")
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
// declares it as a `:context` multi, and `...` likewise. Measured on
// 5.42.0: `my @l = ("a".."c")` is `a b c` and `(1.7..3.2)` is `1 2 3`; in
// scalar context `/x/ .. /y/` over `a x b c y d` gives "", "1", "2", "3",
// "4E0", "", each a string. `...` gives the same in both.
func TestCoreRangeIsContextMulti(t *testing.T) {
	param := func(n string, ty types.Type) types.Param {
		return types.Param{Name: n, Sigil: '$', Type: ty, Required: true}
	}
	for _, sym := range []string{"..", "..."} {
		want := []operatorDecl{
			{name: sym, fixity: "infix", class: "RANGE", multi: true, sig: types.Signature{
				Params: []types.Param{param("x", types.Str), param("y", types.Str)}, Returns: types.List,
				Context: types.ContextSet(types.ListCtx)}},
			{name: sym, fixity: "infix", class: "RANGE", multi: true, sig: types.Signature{
				Params: []types.Param{param("x", types.Any), param("y", types.Any)}, Returns: types.Str,
				Context: types.ContextSet(types.ScalarCtx)}},
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
			Returns: types.List, Context: types.ContextSet(types.ListCtx)}},
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
		{name: "=~", fixity: "infix", class: "MATCHRE", multi: true, sig: types.Signature{
			Params: params, Returns: types.List, Context: types.ContextSet(types.ListCtx)}},
		{name: "=~", fixity: "infix", class: "MATCHRE", multi: true, sig: types.Signature{
			Params: params, Returns: types.Boolean | types.Str, Context: types.ContextSet(types.ScalarCtx)}},
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
		{name: `\`, fixity: "prefix", multi: true, sig: types.Signature{
			Params:  []types.Param{{Name: "l", Sigil: '@', Type: types.List}},
			Returns: types.List, Context: types.ContextSet(types.ListCtx)}},
		{name: `\`, fixity: "prefix", multi: true, sig: types.Signature{
			Params:  []types.Param{{Name: "x", Sigil: '$', Type: types.Any, Required: true}},
			Returns: types.Ref}},
	}
	if got := coreCandidates(t, "prefix", `\`); !reflect.DeepEqual(got, want) {
		t.Errorf(`\:`+"\n got %+v\nwant %+v", got, want)
	}
}
