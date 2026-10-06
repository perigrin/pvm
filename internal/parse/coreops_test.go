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
