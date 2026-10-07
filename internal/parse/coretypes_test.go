// ABOUTME: Tests that CORE.pmt's typed lines derive perl's own prototypes.
// ABOUTME: Each line's prototype, derived from its types where it has them, is asked of perl 5.42.
package parse_test

import (
	"fmt"
	"maps"
	"os"
	"os/exec"
	"regexp"
	"slices"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/conformance"
	"tamarou.com/pvm/internal/parse"
	"tamarou.com/pvm/internal/types"
)

// corePrototypeDisagreements holds the prototype each line of a CORE.pmt
// gives -- derived from its types where the line is typed, as written where
// it is not -- to perl's, and reports each line that disagrees, naming its
// builtin. A file in error derives nothing, so its errors are the report.
// checked is how many lines it compared. Perl's undef is a derived `@`
// (RFC 0001, "Builtins that keep their own parse"): a sub with no
// prototype and one with `(@)` parse alike.
func corePrototypeDisagreements(src []byte, perl map[string]string) (checked int, bad []string) {
	derived, err := parse.DerivedPrototypes(src)
	if err != nil {
		return 0, []string{err.Error()}
	}
	for _, name := range slices.Sorted(maps.Keys(derived)) {
		checked++
		want, ok := perl[name]
		switch {
		case !ok && derived[name] == "@":
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
	// The 188 builtins perl prototypes, and return, which perl gives none
	// and whose types derive `@`. The other builtins with no prototype
	// derive none, by `:unary`, `:listop`, an invocant colon or a multi's
	// candidates.
	if checked != 189 {
		t.Errorf("checked %d CORE.pmt lines; perl prototypes 188 builtins, and return derives @", checked)
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

// measuredRow is a builtin's typing as internal/types/signatures.go wrote
// it: the fewest arguments a call passes, each argument's type in order,
// a slurpy's element standing for every argument it takes, and the return
// type. Its fields are exported so that fmt prints their types by name.
type measuredRow struct {
	MinArity int
	Args     []types.Type
	Returns  types.Type
}

// rowOf reads a declaration in measuredRow's terms. A required parameter
// counts toward the arity; a slurpy `List[Str] @args` is its element, Str,
// and a plain `List @list` is List.
func rowOf(sig types.Signature) measuredRow {
	row := measuredRow{Returns: sig.Returns}
	for _, p := range sig.Params {
		if p.Required {
			row.MinArity++
		}
		typ := p.Type
		if p.Slurpy() && p.Element != types.Unknown {
			typ = p.Element
		}
		row.Args = append(row.Args, typ)
	}
	return row
}

// TestCoreTypesMatchMeasuredSignatures: CORE.pmt carries the measured
// types of the 25 prototyped builtins internal/types/signatures.go types,
// each row copied here so the check outlives that file. Where a row
// disagreed with perl 5.42 it is perl's here, and says what was measured.
func TestCoreTypesMatchMeasuredSignatures(t *testing.T) {
	golden := map[string]measuredRow{
		// MinArity was 2 for push, unshift and join: `push(@a)`,
		// `unshift(@a)` and `join(":")` compile, while `push()` and
		// `join()` are "Not enough arguments".
		"push":    {1, []types.Type{types.Array, types.List}, types.Int},
		"pop":     {0, []types.Type{types.Array}, types.Scalar},
		"shift":   {0, []types.Type{types.Array}, types.Scalar},
		"unshift": {1, []types.Type{types.Array, types.List}, types.Int},
		"splice":  {1, []types.Type{types.Array, types.Int, types.Int, types.List}, types.List},
		"length":  {0, []types.Type{types.Str}, types.Int},
		"chr":     {0, []types.Type{types.Int}, types.Str},
		"ord":     {0, []types.Type{types.Str}, types.Int},
		"join":    {1, []types.Type{types.Str, types.List}, types.Str},
		"sprintf": {1, []types.Type{types.Str, types.List}, types.Str},
		// The fourth argument is the replacement string, not another
		// number: `substr($s, 0, 1, 5)` makes "abc" "5bc".
		"substr": {2, []types.Type{types.Str, types.Num, types.Num, types.Str}, types.Str},
		"ref":    {0, []types.Type{types.Scalar}, types.Str},
		// The argument is coerced to a Scalar, as a `$` prototype's is:
		// with `@a = (5, 6, 7)`, `scalar(@a)` is 3. A `$` parameter typed
		// List would derive `+`, which perl's `$` is not.
		"scalar": {1, []types.Type{types.Scalar}, types.Scalar},
		"die":    {0, []types.Type{types.Str}, types.None},
		"warn":   {0, []types.Type{types.Str}, types.Boolean},
		"bless":  {1, []types.Type{types.Ref, types.Str}, types.Object},
		// MinArity was 1 for abs, int, uc, lc, ucfirst, lcfirst and
		// reverse: each compiles with no argument, `abs()` and `uc()`
		// taking `$_`, and `reverse()` an empty list in list context but
		// `$_` reversed in scalar context ("xyz" gives "zyx").
		"abs":     {0, []types.Type{types.Num}, types.Num},
		"int":     {0, []types.Type{types.Num}, types.Int},
		"uc":      {0, []types.Type{types.Str}, types.Str},
		"lc":      {0, []types.Type{types.Str}, types.Str},
		"ucfirst": {0, []types.Type{types.Str}, types.Str},
		"lcfirst": {0, []types.Type{types.Str}, types.Str},
		"index":   {2, []types.Type{types.Str, types.Str, types.Int}, types.Int},
		"rindex":  {2, []types.Type{types.Str, types.Str, types.Int}, types.Int},
		"reverse": {0, []types.Type{types.List}, types.List},
	}
	core := parse.CoreSignatures()
	for _, name := range slices.Sorted(maps.Keys(golden)) {
		sigs := core[name]
		// A row measured one value per call; a builtin perl answers by
		// context (reverse) is held to the candidate list context selects.
		sig := types.Signature{}
		switch {
		case len(sigs) == 1:
			sig = sigs[0]
		case len(sigs) > 1:
			sel := types.Select(sigs, golden[name].Args, types.ListCtx)
			if sel.Outcome != types.Selected {
				t.Errorf("%s: no single list-context candidate among %d: %+v", name, len(sigs), sel)
				continue
			}
			sig = sigs[sel.Candidate]
		default:
			t.Errorf("%s: CORE.pmt declares no signature", name)
			continue
		}
		if got, want := rowOf(sig), golden[name]; fmt.Sprint(got) != fmt.Sprint(want) {
			t.Errorf("%s: CORE.pmt has %v; measured %v", name, got, want)
		}
	}
}

// perlFunctionKinds asks perl 5.42's Pod::Functions for the builtins
// perlfunc files under each kind ("Perl Functions by Category" in perlfunc),
// keeping those perl prototypes: a batch of CORE.pmt's prototyped lines.
func perlFunctionKinds(t *testing.T, kinds ...string) []string {
	t.Helper()
	protos := perlPrototypes(t)
	perl, err := conformance.PerlPath()
	if err != nil {
		t.Skipf("no perl 5.42 to ask: %v", err)
	}
	args := append([]string{"-MPod::Functions", "-e", `print "@{$Kinds{$_}}\n" for @ARGV`, "--"}, kinds...)
	out, err := exec.Command(perl, args...).Output()
	if err != nil {
		t.Fatalf("asking perl: %v", err)
	}
	var names []string
	for _, name := range strings.Fields(string(out)) {
		if _, ok := protos[name]; ok && !slices.Contains(names, name) {
			names = append(names, name)
		}
	}
	slices.Sort(names)
	return names
}

// untypedLines reports each of names that a CORE.pmt declares with no
// types, a prototype-only line. Such a line has the signature its prototype
// gives, but no prototype states a return type, which a typed line does.
func untypedLines(src []byte, names []string) []string {
	sigs, err := parse.CoreSignaturesOf(src)
	if err != nil {
		return []string{err.Error()}
	}
	var bad []string
	for _, name := range names {
		if len(sigs[name]) == 0 || sigs[name][0].Returns == types.Unknown {
			bad = append(bad, name)
		}
	}
	return bad
}

// stringNumericRegexBuiltins is perlfunc's "Functions for SCALARs or
// strings", "Regular expressions and pattern matching" and "Numeric
// functions" among CORE.pmt's prototyped lines, without pos: its glob slot
// has no spelling yet (RFC 0001, open question 11), so its line stays
// prototype-only.
func stringNumericRegexBuiltins(t *testing.T) []string {
	t.Helper()
	names := slices.DeleteFunc(perlFunctionKinds(t, "String", "Regexp", "Math"),
		func(name string) bool { return name == "pos" })
	// Measured on 5.42.0: an empty or short answer is perl not being asked.
	if len(names) != 29 {
		t.Fatalf("Pod::Functions gives %d prototyped string, regex and numeric builtins besides pos, %v; measured 29", len(names), names)
	}
	return names
}

// TestCoreStringNumericRegexBuiltinsTyped: every string, regex and numeric
// builtin has a typed CORE.pmt line.
func TestCoreStringNumericRegexBuiltinsTyped(t *testing.T) {
	src, err := os.ReadFile("declarations/CORE.pmt")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range untypedLines(src, stringNumericRegexBuiltins(t)) {
		t.Errorf("%s: CORE.pmt declares it untyped", name)
	}
}

// TestCoreBatchCheckCatchesUntypedLine: the batch check is not
// tautological. Over a copy of CORE.pmt whose crypt line is prototype-only
// again, it reports crypt, and nothing else.
func TestCoreBatchCheckCatchesUntypedLine(t *testing.T) {
	src, err := os.ReadFile("declarations/CORE.pmt")
	if err != nil {
		t.Fatal(err)
	}
	crypt := regexp.MustCompile(`(?m)^sub crypt .*;$`)
	if len(crypt.FindAll(src, -1)) != 1 {
		t.Fatal("CORE.pmt has no one crypt line to untype")
	}
	untyped := crypt.ReplaceAll(src, []byte(`sub crypt :prototype($$);`))
	if bad := untypedLines(untyped, stringNumericRegexBuiltins(t)); !slices.Equal(bad, []string{"crypt"}) {
		t.Errorf("reported %q; want crypt alone", bad)
	}
}

// TestCoreFeatureGatedBuiltinTypedParseUnchanged: fc, a builtin only under
// the fc feature or as CORE::fc, is typed like the rest of its batch, and its
// parse is what it was before: without the feature perl reads `fc $a` as the
// method call `$a->fc`, which this parser leaves as it was (featuregated_test.go).
func TestCoreFeatureGatedBuiltinTypedParseUnchanged(t *testing.T) {
	src, err := os.ReadFile("declarations/CORE.pmt")
	if err != nil {
		t.Fatal(err)
	}
	if bad := untypedLines(src, []string{"fc"}); len(bad) != 0 {
		t.Errorf("CORE.pmt declares fc untyped")
	}
	for src, canon := range map[string]string{
		`my $z = fc $a, $b;`:                   `my $z = fc $a, $b;`,
		`use feature 'fc'; my $z = fc $a, $b;`: `use feature 'fc'; my $z = fc($a) , $b;`,
		`my $m = ord CORE::fc $c;`:             `my $m = ord(CORE::fc($c));`,
	} {
		n := parse.Parse([]byte(src))
		if got := strings.TrimSpace(parse.Canon(n, []byte(src))); got != canon {
			t.Errorf("Canon(%q):\n  got  %s\n  want %s", src, got, canon)
		}
	}
}
