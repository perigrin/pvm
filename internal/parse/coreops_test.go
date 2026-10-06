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
// the operand and result types below, measured on 5.42.0: under the declared
// operand types each result is within its bound (`<=>` and `cmp` give -1, 0
// or 1; the predicates give perl's booleans; `&&`, `||`, `//`, `and`, `or`
// and `=` give an operand). An operator that forks has a row for each
// candidate, keyed by the `:context` it states.
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
		"infix +": bin(N, N, N), "infix -": bin(N, N, N), "infix *": bin(N, N, N),
		"infix /": bin(N, N, N), "infix %": bin(N, N, N), "infix **": bin(N, N, N),

		"infix .": bin(S, S, S),

		"infix ==": bin(N, N, B), "infix !=": bin(N, N, B), "infix <": bin(N, N, B),
		"infix >": bin(N, N, B), "infix <=": bin(N, N, B), "infix >=": bin(N, N, B),
		"infix <=>": bin(N, N, I),

		"infix eq": bin(S, S, B), "infix ne": bin(S, S, B), "infix lt": bin(S, S, B),
		"infix gt": bin(S, S, B), "infix le": bin(S, S, B), "infix ge": bin(S, S, B),
		"infix cmp": bin(S, S, I),

		"infix &&": bin(A, A, A), "infix ||": bin(A, A, A), "infix //": bin(A, A, A),
		"infix and": bin(A, A, A), "infix or": bin(A, A, A), "infix xor": bin(A, A, B),

		"infix &": bin(I, I, I), "infix |": bin(I, I, I), "infix ^": bin(I, I, I),
		"infix <<": bin(I, I, I), "infix >>": bin(I, I, I),

		"infix isa": bin(types.Scalar, S, B),

		"infix =~": bin(S, types.Regex, B), "infix !~": bin(S, types.Regex, B),

		"infix =": bin(A, A, A),

		"infix .. :context(@)": bin(S, S, types.List), "infix .. :context($)": bin(A, A, S),
		"infix ... :context(@)": bin(S, S, types.List), "infix ... :context($)": bin(A, A, S),

		"prefix -": un(N, N), "prefix +": un(N, N), "prefix !": un(A, B),
		"prefix not": un(A, B), "prefix ~": un(I, I), `prefix \`: un(A, types.Ref),
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
