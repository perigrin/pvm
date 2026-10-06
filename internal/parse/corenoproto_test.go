// ABOUTME: Tests CORE.pmt's builtins with no prototype: their declarations, :unary, and each's and select's multis.
// ABOUTME: Names and measured rows are held here, each measured on perl 5.42.
package parse_test

import (
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"
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
// one; split waits on a ruling, filed as 01a1113c (TestCoreSplitDeclared).
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
// `($r, $w, $e, Num $timeout) List` for the system call's (nfound,
// timeleft), measured as TestCoreDeleteSliceAndSelectListRows says. each forks
// on the container its operand is, and its two candidates derive perl's
// `\[%@]`.
func TestCoreSelectAndEachAreMultis(t *testing.T) {
	core := parse.CoreSignatures()
	for name, want := range map[string][]measuredRow{
		"select": {
			{0, nil, types.Str},
			{1, []types.Type{types.FileHandle}, types.Str},
			{4, []types.Type{types.Unknown, types.Unknown, types.Unknown, types.Num}, types.List},
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

// TestCoreSelectNoArgumentsSelectsHandle: a call to select with no
// argument selects the `() Str` candidate. Measured on 5.42, `select()`
// compiles and is "main::STDOUT", the selected handle.
func TestCoreSelectNoArgumentsSelectsHandle(t *testing.T) {
	sel := types.Select(parse.CoreSignatures()["select"], nil, types.ScalarCtx)
	if sel.Outcome != types.Selected || sel.Candidate != 0 || sel.Returns != types.Str {
		t.Errorf("select() selects %+v; want candidate 0, () Str", sel)
	}
}

// TestCoreEachRefusesScalar: each on a scalar, `each $x`, matches neither
// candidate. perl 5.42 refuses it at compile time: "Experimental each on
// scalar is now forbidden". A hash and an array each select their own.
func TestCoreEachRefusesScalar(t *testing.T) {
	each := parse.CoreSignatures()["each"]
	for arg, want := range map[types.Type]int{types.Scalar: -1, types.Hash: 0, types.Array: 1} {
		if sel := types.Select(each, []types.Type{arg}, types.ListCtx); sel.Candidate != want || (want < 0) != (sel.Outcome == types.Failed) {
			t.Errorf("each %v selects %+v; want candidate %d", arg, sel, want)
		}
	}
}

// TestCoreNoPrototypeBuiltinsOutOfCoreTable: no builtin with no prototype
// enters the prototype table, where aliasTarget would read presence as a
// known prototype.
func TestCoreNoPrototypeBuiltinsOutOfCoreTable(t *testing.T) {
	table := parse.CoreTable()
	for _, name := range noPrototypeBuiltins {
		if proto, ok := table[name]; ok {
			t.Errorf("%s: in the prototype table as (%s); perl reports none", name, proto)
		}
	}
}

// TestCoreUnaryAndListOperatorParseDiffer: `:unary` is not inferred. A
// marked builtin and an unmarked one with no prototype read `NAME $a, $b`
// differently, as perl does (B::Deparse -p): `defined $a, $b` is
// `defined($a), $b`, while `print $a, $b` is `print($a, $b)`, one call
// with both.
func TestCoreUnaryAndListOperatorParseDiffer(t *testing.T) {
	core := parse.CoreSignatures()
	if !core["defined"][0].Unary || core["print"][0].Unary {
		t.Errorf(":unary on defined %v, print %v; want defined alone", core["defined"][0].Unary, core["print"][0].Unary)
	}
	for src, canon := range map[string]string{
		`my $z = defined $a, $b;`: `my $z = defined($a) , $b;`,
		`my $z = print $a, $b;`:   `my $z = print($a , $b);`,
	} {
		n := parse.Parse([]byte(src))
		if got := strings.TrimSpace(parse.Canon(n, []byte(src))); got != canon {
			t.Errorf("Canon(%q):\n  got  %s\n  want %s", src, got, canon)
		}
	}
}

// TestCoreNoPrototypeBuiltinsMatchMeasuredSignatures: CORE.pmt carries the
// measured types internal/types/signatures.go gives the builtins with no
// prototype, and each, each row copied here so the check outlives that
// file. each's row is the union of its candidates', position by position.
// split's row is TestCoreSplitDeclared's.
func TestCoreNoPrototypeBuiltinsMatchMeasuredSignatures(t *testing.T) {
	golden := map[string]measuredRow{
		// An element's operand, which perl does not read (open question
		// 10), is a Scalar; a slice's yields a list, `delete @h{qw(a b)}`
		// (1, 2), so the union of the two candidates is List, where
		// signatures.go's Scalar row covered only the element.
		"delete":  {1, []types.Type{types.List}, types.List},
		"exists":  {1, []types.Type{types.Scalar}, types.Boolean},
		"each":    {1, []types.Type{types.Hash | types.Array}, types.List},
		"chomp":   {0, []types.Type{types.Str}, types.Int},
		"chop":    {0, []types.Type{types.Str}, types.Str},
		"defined": {0, []types.Type{types.Scalar}, types.Boolean},
		"return":  {0, []types.Type{types.Any}, types.Any},
	}
	core := parse.CoreSignatures()
	for _, name := range slices.Sorted(maps.Keys(golden)) {
		sigs := core[name]
		if len(sigs) == 0 {
			t.Errorf("%s: CORE.pmt does not declare it", name)
			continue
		}
		got := rowOf(sigs[0])
		for _, s := range sigs[1:] {
			row := rowOf(s)
			got.MinArity = min(got.MinArity, row.MinArity)
			for i, typ := range row.Args {
				if i < len(got.Args) {
					got.Args[i] |= typ
				}
			}
			got.Returns |= row.Returns
		}
		if want := golden[name]; fmt.Sprint(got) != fmt.Sprint(want) {
			t.Errorf("%s: CORE.pmt has %v; measured %v", name, got, want)
		}
	}
}

// TestCoreSplitDeclared: split has no prototype and is a list operator,
// `split $a, $b` being `split(/$a/, $b, 0)` (B::Deparse -p), so its line
// must derive `@` or nothing. Its measured types derive more: the string
// is taken in scalar context (`split /,/, @a` over two elements is "2")
// and a fourth argument is "Too many arguments for split", which
// `(Regex|Str $pattern = ' ', Str $string = $_, Int $limit = 0) List`
// says and the `;$_$` it derives claims as a prototype. `(List @args)`
// derives `@` but says the string flattens, which perl measures false.
// No rule of RFC 0001's spells both, so the line waits on a ruling, filed
// as 01a1113c.
func TestCoreSplitDeclared(t *testing.T) {
	t.Skip("split's declaration waits on a ruling (01a1113c): its measured types derive a prototype perl does not report")
	// The row internal/types/signatures.go measured, which the line is to carry.
	want := measuredRow{0, []types.Type{types.Regex | types.Str, types.Str, types.Int}, types.List}
	sigs := parse.CoreSignatures()["split"]
	if len(sigs) != 1 || fmt.Sprint(rowOf(sigs[0])) != fmt.Sprint(want) {
		t.Errorf("split: CORE.pmt has %v; measured %v", sigs, want)
	}
}
