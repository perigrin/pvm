// ABOUTME: Tests CORE.pmt's builtins with no prototype: their declarations, :unary, and each's and select's multis.
// ABOUTME: Names and measured rows are held here, each measured on perl 5.42.
package parse_test

import (
	"fmt"
	"os"
	"slices"
	"testing"

	"tamarou.com/pvm/internal/parse"
	"tamarou.com/pvm/internal/types"
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

// noPrototypeBuiltins are the builtins perl 5.42 gives no prototype that
// CORE.pmt declares: the keywords whose `prototype("CORE::k")` is undef
// and which Pod::Functions files as functions, less the declarators
// (`my`, `sub`, `package`, `use` and kin) and the statement forms the
// parser owns, `last`, `next`, `redo` and `require` (keyword.go). grep,
// map and sort are declared with their own parse, by the issue after this
// one; split waits on a ruling (TestCoreSplitDeclared).
var noPrototypeBuiltins = []string{
	"chomp", "chop", "defined", "delete", "do", "eval", "exec", "exists",
	"goto", "print", "printf", "return", "say", "select", "system",
}

// namedUnaryBuiltins are those of noPrototypeBuiltins perl reads as named
// unaries, RFC 0001's eight. Measured with B::Deparse's -p, which places
// the comma: `my $z = KW $a, $b` deparses as `((my($z) = defined($a)),
// $b)`, the comma outside, for each of these (`delete $h{a}` and
// `exists $h{a}` for the two that take an element), and inside for the
// rest: `split(/$a/, $b, 0)`, `(return $a, $b)`; `select $a, $b` is "Not
// enough arguments for select system call", both taken.
var namedUnaryBuiltins = []string{"chomp", "chop", "defined", "delete", "do", "eval", "exists", "goto"}

// TestCoreUnaryBuiltins: CORE.pmt declares each builtin with no
// prototype, and `:unary` marks exactly the named unaries among
// everything it declares (RFC 0001, "Builtins with no prototype").
func TestCoreUnaryBuiltins(t *testing.T) {
	perl := perlPrototypes(t)
	core := parse.CoreSignatures()
	for _, name := range noPrototypeBuiltins {
		if proto, ok := perl[name]; ok {
			t.Errorf("%s: perl reports (%s); measured none", name, proto)
		}
		if len(core[name]) == 0 {
			t.Errorf("%s: CORE.pmt does not declare it", name)
		}
	}
	var unary []string
	for name, sigs := range core {
		if slices.ContainsFunc(sigs, func(s types.Signature) bool { return s.Unary }) {
			unary = append(unary, name)
		}
	}
	slices.Sort(unary)
	if !slices.Equal(unary, namedUnaryBuiltins) {
		t.Errorf(":unary marks %v; perl's named unaries with no prototype are %v", unary, namedUnaryBuiltins)
	}
}

// TestCoreUnaryDerivesNoPrototype: a `:unary` line's types derive no
// prototype, as perl reports none for the named unaries it marks: defined's
// `(Scalar $thing = $_)` would otherwise give `_`. So it stays out of the
// prototype table, and a `:prototype` stated beside it is an error.
func TestCoreUnaryDerivesNoPrototype(t *testing.T) {
	derived, err := parse.DerivedPrototypes([]byte("package CORE;\nsub defined :unary (Scalar $thing = $_) Boolean;\n"))
	if err != nil {
		t.Fatal(err)
	}
	if proto, ok := derived["defined"]; ok {
		t.Errorf("defined derives (%s); want none", proto)
	}
	_, err = parse.DerivedPrototypes([]byte("package CORE;\nsub defined :prototype(_) :unary (Scalar $thing = $_) Boolean;\n"))
	want := "sub defined: :prototype(_) disagrees with :unary, which derives no prototype"
	if err == nil || err.Error() != want {
		t.Errorf("error %v; want %q", err, want)
	}
}

// rowsOf reads each of a name's candidates as a measuredRow.
func rowsOf(sigs []types.Signature) []measuredRow {
	rows := make([]measuredRow, len(sigs))
	for i, s := range sigs {
		rows[i] = rowOf(s)
	}
	return rows
}

// TestCoreSelectAndEachAreMultis: select and each are the RFC's multis
// ("Multi declarations"). select forks on arity: `() Str` for the
// selected handle, `(FileHandle $fh) Str` for the one selected before, and
// `($r, $w, $e, Num $timeout) Int` for the system call's count. each forks
// on the container its operand is, and its two candidates derive perl's
// `\[%@]`.
func TestCoreSelectAndEachAreMultis(t *testing.T) {
	core := parse.CoreSignatures()
	for name, want := range map[string][]measuredRow{
		"select": {
			{0, nil, types.Str},
			{1, []types.Type{types.FileHandle}, types.Str},
			{4, []types.Type{types.Unknown, types.Unknown, types.Unknown, types.Num}, types.Int},
		},
		"each": {
			{1, []types.Type{types.Hash}, types.List},
			{1, []types.Type{types.Array}, types.List},
		},
	} {
		if got := rowsOf(core[name]); fmt.Sprint(got) != fmt.Sprint(want) {
			t.Errorf("%s: CORE.pmt has %v; want %v", name, got, want)
		}
	}
	src, err := os.ReadFile("declarations/CORE.pmt")
	if err != nil {
		t.Fatal(err)
	}
	derived, err := parse.DerivedPrototypes(src)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := derived["each"], perlPrototypes(t)["each"]; got != want {
		t.Errorf("each derives (%s); perl says (%s)", got, want)
	}
}
