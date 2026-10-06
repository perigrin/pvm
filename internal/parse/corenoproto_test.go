// ABOUTME: Tests CORE.pmt's builtins with no prototype: their declarations, :unary, and each's and select's multis.
// ABOUTME: Names and measured rows are held here, each measured on perl 5.42.
package parse_test

import (
	"slices"
	"testing"
)

// TestCoreDerivedPrototypeCheckReadsUndefAsAt: the CORE.pmt check treats
// perl's undef prototype as a derived `@` (RFC 0001, "Builtins that keep
// their own parse": a sub with no prototype and one with `(@)` parse
// alike). A line for a builtin perl gives no prototype passes when its
// types derive `@`, and is still reported when they derive anything else.
func TestCoreDerivedPrototypeCheckReadsUndefAsAt(t *testing.T) {
	perl := perlPrototypes(t)
	if _, ok := perl["return"]; ok {
		t.Fatal("perl reports a prototype for return; measured none")
	}
	for line, want := range map[string][]string{
		`sub return (List[Any] @values) Any;`: nil,
		`sub return (Scalar $value) Any;`:     {`return: CORE.pmt gives ($); perl reports no prototype`},
	} {
		if _, bad := corePrototypeDisagreements([]byte("package CORE;\n"+line+"\n"), perl); !slices.Equal(bad, want) {
			t.Errorf("%s: reported %q; want %q", line, bad, want)
		}
	}
}
