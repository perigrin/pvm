// ABOUTME: Tests that CORE.pmt's typed lines derive perl's own prototypes.
// ABOUTME: Each line's prototype, derived from its types where it has them, is asked of perl 5.42.
package parse_test

import (
	"fmt"
	"maps"
	"os"
	"slices"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// corePrototypeDisagreements holds the prototype each line of a CORE.pmt
// gives -- derived from its types where the line is typed, as written where
// it is not -- to perl's, and reports each line that disagrees, naming its
// builtin. A file in error derives nothing, so its errors are the report.
// checked is how many lines it compared.
func corePrototypeDisagreements(src []byte, perl map[string]string) (checked int, bad []string) {
	derived, err := parse.DerivedPrototypes(src)
	if err != nil {
		return 0, []string{err.Error()}
	}
	for _, name := range slices.Sorted(maps.Keys(derived)) {
		checked++
		want, ok := perl[name]
		switch {
		case !ok:
			bad = append(bad, fmt.Sprintf("%s: CORE.pmt gives (%s); perl reports no prototype", name, derived[name]))
		case derived[name] != want:
			bad = append(bad, fmt.Sprintf("%s: CORE.pmt gives (%s); perl says (%s)", name, derived[name], want))
		}
	}
	return checked, bad
}

// TestCoreDerivedPrototypesArePerls: every CORE.pmt line's prototype,
// derived from its types where the line is typed, is the one
// `prototype("CORE::name")` reports on perl 5.42.0 (RFC 0001, "A typed
// signature and a prototype say the same thing").
func TestCoreDerivedPrototypesArePerls(t *testing.T) {
	perl := perlPrototypes(t)
	src, err := os.ReadFile("declarations/CORE.pmt")
	if err != nil {
		t.Fatal(err)
	}
	checked, bad := corePrototypeDisagreements(src, perl)
	for _, b := range bad {
		t.Error(b)
	}
	if checked != 188 {
		t.Errorf("checked %d CORE.pmt lines; perl prototypes 188 builtins", checked)
	}
}

// TestCoreDerivedPrototypeCheckCatchesMismatch: the check is not
// tautological. A wrong push line run through it is reported, naming push:
// `(List @a, List @l)` is a declaration error, a List parameter that is not
// last, so it derives nothing; `(Array \@array, Scalar $x)` reads, and
// derives `\@$` where perl says `\@@`.
func TestCoreDerivedPrototypeCheckCatchesMismatch(t *testing.T) {
	perl := perlPrototypes(t)
	for line, want := range map[string]string{
		`sub push (List @a, List @l) Int;`:         "sub push: List parameter @a is not last; a single array followed by more parameters is Array \\@a",
		`sub push (Array \@array, Scalar $x) Int;`: `push: CORE.pmt gives (\@$); perl says (\@@)`,
	} {
		_, bad := corePrototypeDisagreements([]byte("package CORE;\n"+line+"\n"), perl)
		if len(bad) != 1 || bad[0] != want {
			t.Errorf("%s: reported %q; want %q", line, bad, want)
		}
	}
}

// TestCoreTypedLineDisagreeingWithItsPrototype: a typed CORE.pmt line whose
// types derive another prototype than the `:prototype(...)` it carries is a
// declaration error. Neither side wins: the file builds no table at all.
func TestCoreTypedLineDisagreeingWithItsPrototype(t *testing.T) {
	derived, err := parse.DerivedPrototypes([]byte("package CORE;\nsub push :prototype(\\@@) (List @a) Int;\n"))
	want := `sub push: :prototype(\@@) disagrees with its types, which give (@)`
	if err == nil || err.Error() != want {
		t.Errorf("error %v; want %q", err, want)
	}
	if derived != nil {
		t.Errorf("a file in error built %v", derived)
	}
}
