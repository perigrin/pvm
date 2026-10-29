// ABOUTME: Tests the CORE.tpm cross-check: G1's builtin shapes against internal/parse's exported table.
// ABOUTME: Holds the committed difference list equal to the cross-check's; -grammar.update-core-tpm regenerates it.

package grammar

import (
	"bytes"
	"flag"
	"maps"
	"os"
	"slices"
	"strings"
	"testing"
)

var updateCoreTpm = flag.Bool("grammar.update-core-tpm", false,
	"rewrite testdata/core-tpm-differences.txt from the cross-check of the vendored Guacamole grammar")

const coreTpmPath = "testdata/core-tpm-differences.txt"

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

// TestCoreTpmGuacamole holds the committed difference list equal to the
// cross-check of the vendored grammar against CORE.tpm, byte for byte.
func TestCoreTpmGuacamole(t *testing.T) {
	diffs, err := GuacamoleCoreTpmDifferences(guacamolePath)
	if err != nil {
		t.Fatal(err)
	}
	got := []byte(strings.Join(diffs, "\n") + "\n")
	if *updateCoreTpm {
		if err := os.WriteFile(coreTpmPath, got, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(coreTpmPath)
	if err != nil {
		t.Fatalf("%v (regenerate with go test ./internal/chart/grammar/ -run TestCoreTpmGuacamole -grammar.update-core-tpm)", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatal(firstDifference(coreTpmPath, "the cross-check", lines(got), lines(want)))
	}
}
