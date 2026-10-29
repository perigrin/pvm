// ABOUTME: Tests the CORE.tpm cross-check: G1's builtin shapes against internal/parse's exported table.
// ABOUTME: Toy grammars check the shape extraction, and that a one-sided builtin is listed, never skipped.

package grammar

import (
	"maps"
	"slices"
	"strings"
	"testing"
)

// keywords is a toy grammar in Guacamole's shape: foo takes no argument,
// bar takes one term, baz a list or nothing, and qux only in G1.
const keywords = `S ::= OpKeywordFooExpr | OpKeywordBarExpr | OpKeywordBazExpr | OpKeywordQuxExpr
OpKeywordFooExpr ::= OpKeywordFoo | OpKeywordFoo LParen RParen
OpKeywordBarExpr ::= OpKeywordBar OpUnaryKeywordArg
OpKeywordBazExpr ::= OpKeywordBaz OpListKeywordArg | OpKeywordBaz
OpKeywordQuxExpr ::= OpKeywordQux OpAssignKeywordArg
OpUnaryKeywordArg ::= Num
OpListKeywordArg ::= Num | OpListKeywordArg Comma Num
OpAssignKeywordArg ::= Num
OpKeywordFoo ~ 'foo'
OpKeywordBar ~ 'bar'
OpKeywordBaz ~ 'baz'
OpKeywordQux ~ 'qux'
LParen ~ '('
RParen ~ ')'
Comma ~ ','
Num ~ [0-9]+
`

func toyShapes(t *testing.T) map[string]string {
	t.Helper()
	shapes, err := G1Shapes(cut(t, keywords, CutSpec{Tokens: []TokenDef{{Name: "Num", Kind: "Number"}}}))
	if err != nil {
		t.Fatal(err)
	}
	return shapes
}

func TestG1Shapes(t *testing.T) {
	want := map[string]string{"foo": "niladic", "bar": "unary", "baz": "list", "qux": "assign"}
	if got := toyShapes(t); !maps.Equal(got, want) {
		t.Errorf("G1Shapes = %v, want %v", got, want)
	}
}

// TestCoreTpmDifferences: a G1 rule taking no argument against a table entry
// taking one is listed; a builtin whose shapes agree is not.
func TestCoreTpmDifferences(t *testing.T) {
	got := CoreTpmDifferences(toyShapes(t), map[string]string{"foo": "unary", "bar": "unary", "baz": "list", "qux": "assign"})
	if want := []string{"foo niladic unary"}; !slices.Equal(got, want) {
		t.Errorf("CoreTpmDifferences = %q, want %q", got, want)
	}
}

// TestCoreTpmOneSidedListed: a builtin present in only one of G1 and the
// table is listed, with none on the side that lacks it, never skipped.
func TestCoreTpmOneSidedListed(t *testing.T) {
	got := CoreTpmDifferences(toyShapes(t), map[string]string{"foo": "niladic", "bar": "unary", "baz": "list", "zap": "unary"})
	if want := []string{"qux assign none", "zap none unary"}; !slices.Equal(got, want) {
		t.Errorf("CoreTpmDifferences = %q, want %q", got, want)
	}
}

func TestG1ShapesRejectsUnknownArgument(t *testing.T) {
	src := "S ::= OpKeywordFooExpr\nOpKeywordFooExpr ::= OpKeywordFoo Mystery\nMystery ::= OpKeywordFoo\nOpKeywordFoo ~ 'foo'\n"
	_, err := G1Shapes(cut(t, src, CutSpec{}))
	if err == nil || !strings.Contains(err.Error(), "Mystery") || !strings.Contains(err.Error(), "OpKeywordFooExpr") {
		t.Errorf("G1Shapes on an unclassified argument: error %v, want one naming the rule and the symbol", err)
	}
}
