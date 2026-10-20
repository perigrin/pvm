// ABOUTME: Holds each .tpm's declared return types to perl: perl calls each sub with its oracle call cases.
// ABOUTME: A value perl returns that the declared return type excludes fails the test.

package infer_test

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tamarou.com/pvm/internal/parse"
	"tamarou.com/pvm/internal/perltest"
	"tamarou.com/pvm/internal/types"
)

// oracleExempt are the declared subs the return-type oracle never calls,
// "Module::sub" to why: one that is unsafe in a scratch directory --
// network, fork/exec, signals, _exit, file mutation outside scratch,
// blocking on a terminal. Each is held to its prototype and exports only.
var oracleExempt = map[string]string{}

// holdReturnTypes holds each module's shipped .tpm to perl: every sub it
// declares, called with the cases in testdata/oracle/<Module>.pl (File::Spec
// at File-Spec.pl), returns
// only values its declared return type admits. A batch of declarations
// calls it once with its modules.
func holdReturnTypes(t *testing.T, modules ...string) {
	t.Helper()
	p := perltest.Perl542(t)
	for _, module := range modules {
		rel := strings.ReplaceAll(module, "::", "/")
		tpm, err := os.ReadFile(filepath.Join("..", "parse", "declarations", filepath.FromSlash(rel)+".tpm"))
		if err != nil {
			t.Errorf("%s: no declaration: %v", module, err)
			continue
		}
		cases, err := filepath.Abs(filepath.Join("testdata", "oracle", strings.ReplaceAll(module, "::", "-")+".pl"))
		require.NoError(t, err)
		for _, f := range returnTypeFailures(t, p, oracleRun{Module: module, TPM: tpm, Cases: cases, Exempt: oracleExempt}) {
			t.Error(f)
		}
	}
}

// oracleRun is one module the return-type oracle holds to perl.
type oracleRun struct {
	// Module is the package perl requires, and TPM its declaration source.
	Module string
	TPM    []byte
	// Cases is the call-case file: sub name => [ [args], ... ].
	Cases string
	// Inc is a directory perl searches first, or "".
	Inc string
	// Exempt are the subs never called, "Module::sub" to why.
	Exempt map[string]string
}

// returnTypeFailures has perl call each sub r.TPM declares, except those
// exempt, with its call cases, and returns one line per returned value
// the declared return type excludes, or why the oracle could not run.
// Perl runs with its working directory, HOME and TMPDIR in a temp dir.
func returnTypeFailures(t *testing.T, p perltest.Perl, r oracleRun) []string {
	t.Helper()
	sigs, errs := parse.DeclaredSignatures(r.TPM)
	if len(errs) > 0 {
		return []string{fmt.Sprintf("%s: declaration errors: %v", r.Module, errs)}
	}
	observer, err := filepath.Abs(filepath.Join("testdata", "observe_types.pl"))
	require.NoError(t, err)
	driver, err := filepath.Abs(filepath.Join("testdata", "return_oracle.pl"))
	require.NoError(t, err)

	args := []string{driver, observer, r.Cases, r.Module}
	if r.Inc != "" {
		args = append([]string{"-I", r.Inc}, args...)
	}
	for _, name := range slices.Sorted(maps.Keys(sigs)) {
		if _, exempt := r.Exempt[r.Module+"::"+name]; exempt {
			continue
		}
		for _, call := range calls(sigs[name]) {
			args = append(args, name+":"+call)
		}
	}

	scratch := t.TempDir()
	cmd := p.Command(args...)
	cmd.Dir = scratch
	cmd.Env = append(cmd.Env, "HOME="+scratch, "TMPDIR="+scratch)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return []string{fmt.Sprintf("%s: the oracle did not run: %v: %s", r.Module, err, stderr.String())}
	}
	// The driver says it is complete once every sub is called, so a sub
	// that ends perl early fails here rather than passing the rest.
	report, complete := strings.CutPrefix(string(out), "complete\n")
	if !complete {
		return []string{fmt.Sprintf("%s: the oracle ended before calling every sub: %s", r.Module, stderr.String())}
	}
	var obs []observation
	if err := json.Unmarshal([]byte(report), &obs); err != nil {
		return []string{fmt.Sprintf("%s: the oracle emitted %q: %v", r.Module, out, err)}
	}

	var failures []string
	// failed are the calls a returned value already failed, so the call's
	// Returned marker, which follows its values, adds no second failure.
	failed := map[string]bool{}
	for _, o := range obs {
		f := strings.Split(o.Name, ":")
		name, ctx, form := f[0], f[1], f[2]
		n, _ := strconv.Atoi(f[3])
		call := fmt.Sprintf("%s:%d", o.Name, o.Line)
		if o.Type == "Returned" {
			// A call returning the empty list records no value; it still
			// returned, which only a candidate that returns may take.
			returns := slices.ContainsFunc(sigs[name], func(s types.Signature) bool {
				return s.Returns != types.None && takes(s, form, n) && callContext(s) == ctx
			})
			if !returns && !failed[call] {
				failures = append(failures, fmt.Sprintf("%s::%s call case %d in %s context returned, which no candidate taking %d arguments in that context admits",
					r.Module, name, o.Line, ctx, n))
			}
			continue
		}
		if o.Type == "Died" {
			if !dies(sigs[name], form, n) {
				failures = append(failures, fmt.Sprintf("%s::%s call case %d in %s context died, which no candidate declared None taking %d arguments admits",
					r.Module, name, o.Line, ctx, n))
			}
			continue
		}
		got, known := observedToLattice[o.Type]
		admitted := known && slices.ContainsFunc(sigs[name], func(s types.Signature) bool {
			if s.Returns == types.None || !takes(s, form, n) || callContext(s) != ctx {
				return false
			}
			// In list context each value is one element of the list.
			if ctx == "list" && s.ReturnElement != types.Unknown {
				return types.IsSubtype(got, s.ReturnElement)
			}
			return types.IsSubtype(got, s.Returns)
		})
		if !admitted {
			failed[call] = true
			failures = append(failures, fmt.Sprintf("%s::%s call case %d in %s context returned %s, which its declared return type excludes",
				r.Module, name, o.Line, ctx, o.Type))
		}
	}
	sort.Strings(failures)
	return failures
}

// calls is how the oracle calls a sub's cases, "CONTEXT:FORM", once for
// each context and form one of its candidates answers. A candidate
// declared never to return answers no context: its cases die in whatever
// context the others are called in, or in scalar context when no other
// candidate takes its form.
func calls(cands []types.Signature) []string {
	var out []string
	for _, s := range cands {
		if call := callContext(s) + ":" + callForm(s); s.Returns != types.None && !slices.Contains(out, call) {
			out = append(out, call)
		}
	}
	for _, s := range cands {
		form := callForm(s)
		if s.Returns == types.None && !slices.ContainsFunc(out, func(c string) bool { return strings.HasSuffix(c, ":"+form) }) {
			out = append(out, "scalar:"+form)
		}
	}
	return out
}

// callContext is the context the oracle calls a candidate's cases in,
// and the one its returns are held in: list for one answering list
// context, RFC 0001 "Context selects by return type", or declared Void,
// where a bare return is the empty list; scalar for any other.
func callContext(s types.Signature) string {
	if a := s.Answers(); a == types.ListCtx || a == types.VoidCtx {
		return "list"
	}
	return "scalar"
}

// callForm is how the oracle calls a candidate: as a method, with the
// module as its invocant, when it declares one; as a sub otherwise.
func callForm(s types.Signature) string {
	if s.Invocant != nil {
		return "method"
	}
	return "sub"
}

// dies is whether a call of form with n arguments may die: some
// candidate declared None takes it.
func dies(cands []types.Signature, form string, n int) bool {
	return slices.ContainsFunc(cands, func(s types.Signature) bool {
		return s.Returns == types.None && takes(s, form, n)
	})
}

// takes is whether candidate s takes a call of form with n arguments
// besides any invocant, so a value the call returns, or its death, is
// held to s.
func takes(s types.Signature, form string, n int) bool {
	if form == "method" {
		n++
	}
	return callForm(s) == form && types.Accepts([]types.Signature{s}, n)
}

// oracleFixture writes module's perl source under a lib directory in a
// temp dir, and its call cases beside it, returning the lib directory and
// the case file.
func oracleFixture(t *testing.T, module, pm, cases string) (lib, caseFile string) {
	t.Helper()
	dir := t.TempDir()
	lib = filepath.Join(dir, "lib")
	pmPath := filepath.Join(lib, filepath.FromSlash(strings.ReplaceAll(module, "::", "/"))+".pm")
	require.NoError(t, os.MkdirAll(filepath.Dir(pmPath), 0o755))
	require.NoError(t, os.WriteFile(pmPath, []byte(pm), 0o644))
	caseFile = filepath.Join(dir, "cases.pl")
	require.NoError(t, os.WriteFile(caseFile, []byte(cases), 0o644))
	return lib, caseFile
}

// TestReturnOracleRejectsAnExcludedValue: a sub declared Int whose perl
// returns a string fails the oracle, and one returning an integer passes.
func TestReturnOracleRejectsAnExcludedValue(t *testing.T) {
	p := perltest.Perl542(t)
	tpm := []byte("package My::Oracle;\nsub count () Int;\n")
	cases := "(count => [ [] ]);\n"

	lib, caseFile := oracleFixture(t, "My::Oracle", "package My::Oracle;\nsub count { 'many' }\n1;\n", cases)
	failures := returnTypeFailures(t, p, oracleRun{Module: "My::Oracle", TPM: tpm, Cases: caseFile, Inc: lib})
	require.Len(t, failures, 1, "a Str where Int is declared")
	assert.Contains(t, failures[0], "count")

	lib, caseFile = oracleFixture(t, "My::Oracle", "package My::Oracle;\nsub count { 3 }\n1;\n", cases)
	assert.Empty(t, returnTypeFailures(t, p, oracleRun{Module: "My::Oracle", TPM: tpm, Cases: caseFile, Inc: lib}))
}

// TestReturnOracleCallsInTheDeclaredContext: a sub declared to return a
// scalar is called in scalar context, where perl's wantarray is false.
func TestReturnOracleCallsInTheDeclaredContext(t *testing.T) {
	p := perltest.Perl542(t)
	lib, caseFile := oracleFixture(t, "My::Wants", "package My::Wants;\nsub n { wantarray ? 'list' : 1 }\n1;\n", "(n => [ [] ]);\n")
	tpm := []byte("package My::Wants;\nsub n () Int;\n")

	assert.Empty(t, returnTypeFailures(t, p, oracleRun{Module: "My::Wants", TPM: tpm, Cases: caseFile, Inc: lib}))
}

// TestReturnOracleNeverCallsAnExemptSub: an exempt sub that would write a
// marker file is not called, so no marker is left.
func TestReturnOracleNeverCallsAnExemptSub(t *testing.T) {
	p := perltest.Perl542(t)
	marker := filepath.Join(t.TempDir(), "called")
	pm := fmt.Sprintf("package My::Exempt;\nsub touch { open my $fh, '>', %q or die; 1 }\nsub safe { 1 }\n1;\n", marker)
	lib, caseFile := oracleFixture(t, "My::Exempt", pm, "(touch => [ [] ], safe => [ [] ]);\n")
	tpm := []byte("package My::Exempt;\nsub touch () Int;\nsub safe () Int;\n")
	exempt := map[string]string{"My::Exempt::touch": "writes a file outside scratch"}

	assert.Empty(t, returnTypeFailures(t, p, oracleRun{Module: "My::Exempt", TPM: tpm, Cases: caseFile, Inc: lib, Exempt: exempt}))
	assert.NoFileExists(t, marker, "the exempt sub was called")
}

// TestReturnOracleExemptionsNameAReason: every exemption says why.
func TestReturnOracleExemptionsNameAReason(t *testing.T) {
	for sub, reason := range oracleExempt {
		assert.NotEmpty(t, strings.TrimSpace(reason), "%s is exempt from the return-type oracle without a reason", sub)
	}
}

// TestReturnOracleRejectsAReturningNoneSub: a sub declared None that
// returns fails the oracle, and one that dies passes it.
func TestReturnOracleRejectsAReturningNoneSub(t *testing.T) {
	p := perltest.Perl542(t)
	tpm := []byte("package My::Never;\nsub stop () None;\n")
	cases := "(stop => [ [] ]);\n"

	lib, caseFile := oracleFixture(t, "My::Never", "package My::Never;\nsub stop { 1 }\n1;\n", cases)
	failures := returnTypeFailures(t, p, oracleRun{Module: "My::Never", TPM: tpm, Cases: caseFile, Inc: lib})
	require.Len(t, failures, 1, "a sub declared None returned")
	assert.Contains(t, failures[0], "stop")

	lib, caseFile = oracleFixture(t, "My::Never", "package My::Never;\nsub stop { die 'stopped' }\n1;\n", cases)
	assert.Empty(t, returnTypeFailures(t, p, oracleRun{Module: "My::Never", TPM: tpm, Cases: caseFile, Inc: lib}))
}

// TestReturnOracleRejectsAnUncoveredSub: a declared sub with neither a
// call case nor an exemption fails the oracle by name.
func TestReturnOracleRejectsAnUncoveredSub(t *testing.T) {
	p := perltest.Perl542(t)
	lib, caseFile := oracleFixture(t, "My::Uncovered",
		"package My::Uncovered;\nsub one { 1 }\nsub two { 2 }\n1;\n", "(one => [ [] ]);\n")
	tpm := []byte("package My::Uncovered;\nsub one () Int;\nsub two () Int;\n")

	failures := returnTypeFailures(t, p, oracleRun{Module: "My::Uncovered", TPM: tpm, Cases: caseFile, Inc: lib})
	require.Len(t, failures, 1, "two has no call case")
	assert.Contains(t, failures[0], "two has no call case")
}

// TestReturnOracleRunsInScratchDir: perl's working directory, HOME and
// TMPDIR are the oracle's temp dir, so a sub writing under $ENV{HOME}
// writes there, as one writing to a relative path or under TMPDIR does.
func TestReturnOracleRunsInScratchDir(t *testing.T) {
	p := perltest.Perl542(t)
	pm := `package My::Scratch;
use Cwd ();
sub where {
    open my $fh, '>', "$ENV{HOME}/home-marker" or die;
    open $fh, '>', "$ENV{TMPDIR}/tmp-marker" or die;
    open $fh, '>', 'cwd-marker' or die;
    return Cwd::getcwd() eq $ENV{HOME} && $ENV{HOME} eq $ENV{TMPDIR} ? 1 : 'elsewhere';
}
1;
`
	lib, caseFile := oracleFixture(t, "My::Scratch", pm, "(where => [ [] ]);\n")
	tpm := []byte("package My::Scratch;\nsub where () Int;\n")
	home, _ := os.UserHomeDir()

	assert.Empty(t, returnTypeFailures(t, p, oracleRun{Module: "My::Scratch", TPM: tpm, Cases: caseFile, Inc: lib}),
		"perl's cwd, HOME and TMPDIR are one scratch directory")
	assert.NoFileExists(t, filepath.Join(home, "home-marker"), "the sub wrote under the real HOME")
}

// TestCarpReturnTypes holds Carp.tpm's return types to perl.
func TestCarpReturnTypes(t *testing.T) {
	holdReturnTypes(t, "Carp")
}

// TestReturnOracleChecksEachListElement: a sub declared List[Int] is held
// element by element, so a list holding a Str, a reference or undef fails.
func TestReturnOracleChecksEachListElement(t *testing.T) {
	p := perltest.Perl542(t)
	tpm := []byte("package My::Nums;\nsub nums () List[Int];\n")
	cases := "(nums => [ [] ]);\n"

	lib, caseFile := oracleFixture(t, "My::Nums", "package My::Nums;\nsub nums { ('a', [1], undef) }\n1;\n", cases)
	failures := returnTypeFailures(t, p, oracleRun{Module: "My::Nums", TPM: tpm, Cases: caseFile, Inc: lib})
	require.Len(t, failures, 3, "a Str, an ArrayRef and an Undef where List[Int] is declared")
	assert.Contains(t, strings.Join(failures, "\n"), "returned Str")

	lib, caseFile = oracleFixture(t, "My::Nums", "package My::Nums;\nsub nums { (1, 2) }\n1;\n", cases)
	assert.Empty(t, returnTypeFailures(t, p, oracleRun{Module: "My::Nums", TPM: tpm, Cases: caseFile, Inc: lib}))
}

// TestReturnOracleCallsEachCandidateInItsContext: a multi sub whose
// candidates answer scalar and list context is called in both, each
// return held to the candidate answering that context.
func TestReturnOracleCallsEachCandidateInItsContext(t *testing.T) {
	p := perltest.Perl542(t)
	tpm := []byte("package My::Both;\nmulti sub n () Int;\nmulti sub n () List[Str];\n")
	cases := "(n => [ [] ]);\n"

	lib, caseFile := oracleFixture(t, "My::Both", "package My::Both;\nsub n { wantarray ? ('a', 'b') : 'one' }\n1;\n", cases)
	failures := returnTypeFailures(t, p, oracleRun{Module: "My::Both", TPM: tpm, Cases: caseFile, Inc: lib})
	require.Len(t, failures, 1, "a Str in scalar context, where the scalar candidate returns Int")
	assert.Contains(t, failures[0], "in scalar context returned Str")

	lib, caseFile = oracleFixture(t, "My::Both", "package My::Both;\nsub n { wantarray ? ('a', 'b') : 2 }\n1;\n", cases)
	assert.Empty(t, returnTypeFailures(t, p, oracleRun{Module: "My::Both", TPM: tpm, Cases: caseFile, Inc: lib}))
}

// TestReturnOracleFailsOnASubThatExits: a sub calling exit ends perl
// before the subs after it are called, which fails the oracle rather than
// passing it with those subs unchecked.
func TestReturnOracleFailsOnASubThatExits(t *testing.T) {
	p := perltest.Perl542(t)
	// exit dies, which a sub declared Int may not; CORE::exit ends perl
	// before the driver says it is complete.
	want := map[string]string{
		"exit 0":       "My::Exits::a_exits call case 0 in scalar context died",
		"CORE::exit 0": "My::Exits: the oracle ended before calling every sub",
	}
	for _, exit := range []string{"exit 0", "CORE::exit 0"} {
		t.Run(exit, func(t *testing.T) {
			pm := fmt.Sprintf("package My::Exits;\nsub a_exits { %s }\nsub b_str { 'str' }\n1;\n", exit)
			lib, caseFile := oracleFixture(t, "My::Exits", pm, "(a_exits => [ [] ], b_str => [ [] ]);\n")
			tpm := []byte("package My::Exits;\nsub a_exits () Int;\nsub b_str () Int;\n")

			failures := returnTypeFailures(t, p, oracleRun{Module: "My::Exits", TPM: tpm, Cases: caseFile, Inc: lib})
			assert.Contains(t, strings.Join(failures, "\n"), want[exit])
		})
	}
}

// TestReturnOracleTurnsExitIntoADeath: exit dies under the oracle, so a
// sub declared None that calls it is held as one that dies.
func TestReturnOracleTurnsExitIntoADeath(t *testing.T) {
	p := perltest.Perl542(t)
	lib, caseFile := oracleFixture(t, "My::Quits", "package My::Quits;\nsub quit { exit 0 }\n1;\n", "(quit => [ [] ]);\n")
	tpm := []byte("package My::Quits;\nsub quit () None;\n")

	assert.Empty(t, returnTypeFailures(t, p, oracleRun{Module: "My::Quits", TPM: tpm, Cases: caseFile, Inc: lib}))
}

// TestReturnOracleKeepsAModulesLoadTimeOutputOutOfItsReport: a module that
// dups STDOUT as it loads, as Test::Builder does, writes to STDERR through
// its dup, and its END setting perl's exit status, as Test::Builder's does
// when no plan was declared, leaves a complete run successful.
func TestReturnOracleKeepsAModulesLoadTimeOutputOutOfItsReport(t *testing.T) {
	p := perltest.Perl542(t)
	pm := "package My::Builder;\nuse Test::More ();\nsub pass_one { Test::More::ok(1, 'name') }\n1;\n"
	lib, caseFile := oracleFixture(t, "My::Builder", pm, "(pass_one => [ [] ]);\n")
	tpm := []byte("package My::Builder;\nsub pass_one () Int;\n")

	assert.Empty(t, returnTypeFailures(t, p, oracleRun{Module: "My::Builder", TPM: tpm, Cases: caseFile, Inc: lib}))
}

// TestReturnOracleCallsAVoidSubInListContext: a sub declared Void is
// called in list context, where a bare return is the empty list and any
// value it returns fails.
func TestReturnOracleCallsAVoidSubInListContext(t *testing.T) {
	p := perltest.Perl542(t)
	tpm := []byte("package My::Void;\nsub v () Void;\n")
	cases := "(v => [ [] ]);\n"

	lib, caseFile := oracleFixture(t, "My::Void", "package My::Void;\nsub v { return }\n1;\n", cases)
	assert.Empty(t, returnTypeFailures(t, p, oracleRun{Module: "My::Void", TPM: tpm, Cases: caseFile, Inc: lib}))

	lib, caseFile = oracleFixture(t, "My::Void", "package My::Void;\nsub v { 1 }\n1;\n", cases)
	failures := returnTypeFailures(t, p, oracleRun{Module: "My::Void", TPM: tpm, Cases: caseFile, Inc: lib})
	require.Len(t, failures, 1, "a value where Void is declared")
	assert.Contains(t, failures[0], "in list context returned Int")
}

// TestReturnOracleAdmitsADeathOnlyThroughTheNoneCandidate: a multi sub
// with an Int candidate taking one argument and a None candidate taking
// none may die when called with none, and must return an Int when called
// with one.
func TestReturnOracleAdmitsADeathOnlyThroughTheNoneCandidate(t *testing.T) {
	p := perltest.Perl542(t)
	tpm := []byte("package My::Maybe;\nmulti sub f (Int $x) Int;\nmulti sub f () None;\n")
	cases := "(f => [ [1], [] ]);\n"

	lib, caseFile := oracleFixture(t, "My::Maybe", "package My::Maybe;\nsub f { @_ ? $_[0] : die 'none' }\n1;\n", cases)
	assert.Empty(t, returnTypeFailures(t, p, oracleRun{Module: "My::Maybe", TPM: tpm, Cases: caseFile, Inc: lib}))

	lib, caseFile = oracleFixture(t, "My::Maybe", "package My::Maybe;\nsub f { die 'always' }\n1;\n", cases)
	failures := returnTypeFailures(t, p, oracleRun{Module: "My::Maybe", TPM: tpm, Cases: caseFile, Inc: lib})
	require.Len(t, failures, 1, "a death with one argument, which only the Int candidate takes")
	assert.Contains(t, failures[0], "My::Maybe::f call case 0 in scalar context died")

	lib, caseFile = oracleFixture(t, "My::Maybe", "package My::Maybe;\nsub f { @_ ? $_[0] : 0 }\n1;\n", cases)
	failures = returnTypeFailures(t, p, oracleRun{Module: "My::Maybe", TPM: tpm, Cases: caseFile, Inc: lib})
	require.Len(t, failures, 1, "a return with no argument, which only the None candidate takes")
	assert.Contains(t, failures[0], "My::Maybe::f call case 1 in scalar context returned Int")
}

// TestReturnOracleRejectsAnEmptyReturnOnlyTheNoneCandidateTakes: a call
// that returns the empty list records no value, yet it returned, so a
// call only the None candidate takes fails even when it returns nothing.
func TestReturnOracleRejectsAnEmptyReturnOnlyTheNoneCandidateTakes(t *testing.T) {
	p := perltest.Perl542(t)
	tpm := []byte("package My::Empty;\nmulti sub f (Int $x) List[Int];\nmulti sub f () None;\n")
	cases := "(f => [ [1], [] ]);\n"

	lib, caseFile := oracleFixture(t, "My::Empty", "package My::Empty;\nsub f { @_ ? ($_[0]) : die 'none' }\n1;\n", cases)
	assert.Empty(t, returnTypeFailures(t, p, oracleRun{Module: "My::Empty", TPM: tpm, Cases: caseFile, Inc: lib}))

	lib, caseFile = oracleFixture(t, "My::Empty", "package My::Empty;\nsub f { @_ ? ($_[0]) : () }\n1;\n", cases)
	failures := returnTypeFailures(t, p, oracleRun{Module: "My::Empty", TPM: tpm, Cases: caseFile, Inc: lib})
	require.Len(t, failures, 1, "an empty return with no argument, which only the None candidate takes")
	assert.Contains(t, failures[0], "My::Empty::f call case 1 in list context returned")
}

// TestReturnOracleCallsAnInheritedClassMethod: a declaration with an
// invocant is a method, found through can as perl's method call finds an
// inherited one, and called with the class as its invocant.
func TestReturnOracleCallsAnInheritedClassMethod(t *testing.T) {
	p := perltest.Perl542(t)
	pm := `package My::Kid;
our @ISA = ('My::Base');
package My::Base;
sub cat { my ($class, @parts) = @_; $class eq 'My::Kid' ? join('/', @parts) : [] }
1;
`
	lib, caseFile := oracleFixture(t, "My::Kid", pm, "(cat => [ [ 'a', 'b' ] ]);\n")

	tpm := []byte("package My::Kid;\nsub cat (Str $class: List[Str] @parts) Str;\n")
	assert.Empty(t, returnTypeFailures(t, p, oracleRun{Module: "My::Kid", TPM: tpm, Cases: caseFile, Inc: lib}))

	tpm = []byte("package My::Kid;\nsub cat (Str $class: List[Str] @parts) Int;\n")
	failures := returnTypeFailures(t, p, oracleRun{Module: "My::Kid", TPM: tpm, Cases: caseFile, Inc: lib})
	require.Len(t, failures, 1, "a Str where Int is declared")
	assert.Contains(t, failures[0], "returned Str")
}

// TestReturnOracleRejectsASubWithNoCases: a call-case entry listing no
// argument lists calls nothing, so it fails as a missing entry does.
func TestReturnOracleRejectsASubWithNoCases(t *testing.T) {
	p := perltest.Perl542(t)
	lib, caseFile := oracleFixture(t, "My::Empty", "package My::Empty;\nsub one { 'one' }\n1;\n", "(one => []);\n")
	tpm := []byte("package My::Empty;\nsub one () Int;\n")

	failures := returnTypeFailures(t, p, oracleRun{Module: "My::Empty", TPM: tpm, Cases: caseFile, Inc: lib})
	require.Len(t, failures, 1)
	assert.Contains(t, failures[0], "one has no call case")
}

// TestReturnOracleRejectsAnUndefinedNoneSub: a sub declared None that
// perl does not define dies when called, "Undefined subroutine", which is
// no evidence it never returns; it fails the oracle.
func TestReturnOracleRejectsAnUndefinedNoneSub(t *testing.T) {
	p := perltest.Perl542(t)
	lib, caseFile := oracleFixture(t, "My::Gone", "package My::Gone;\n1;\n", "(stop => [ [] ]);\n")
	tpm := []byte("package My::Gone;\nsub stop () None;\n")

	failures := returnTypeFailures(t, p, oracleRun{Module: "My::Gone", TPM: tpm, Cases: caseFile, Inc: lib})
	require.Len(t, failures, 1)
	assert.Contains(t, failures[0], "perl defines no sub stop")
}

// TestReturnOracleKeepsASubsOutputOutOfItsReport: a sub printing to
// STDOUT, as Test::More's ok does, leaves the oracle's report readable.
func TestReturnOracleKeepsASubsOutputOutOfItsReport(t *testing.T) {
	p := perltest.Perl542(t)
	lib, caseFile := oracleFixture(t, "My::Prints", "package My::Prints;\nsub ok { print \"ok 1\\n\"; 1 }\n1;\n", "(ok => [ [] ]);\n")
	tpm := []byte("package My::Prints;\nsub ok () Int;\n")

	assert.Empty(t, returnTypeFailures(t, p, oracleRun{Module: "My::Prints", TPM: tpm, Cases: caseFile, Inc: lib}))
}

// TestReturnOracleIgnoresTheCallersPerlLib: perl finds the module under
// test as an installed perl does, whatever PERL5LIB, PERLLIB or PERL5OPT
// the test runs with: a Carp there whose croak returns is not measured.
func TestReturnOracleIgnoresTheCallersPerlLib(t *testing.T) {
	p := perltest.Perl542(t)
	fake, _ := oracleFixture(t, "Carp", "package Carp;\nsub croak { 1 }\nsub confess { 1 }\nsub carp { 1 }\nsub cluck { 1 }\nsub longmess { 'l' }\nsub shortmess { 's' }\n1;\n", "")
	tpm, err := os.ReadFile(filepath.Join("..", "parse", "declarations", "Carp.tpm"))
	require.NoError(t, err)
	cases, err := filepath.Abs(filepath.Join("testdata", "oracle", "Carp.pl"))
	require.NoError(t, err)
	for _, env := range [][2]string{{"PERL5LIB", fake}, {"PERLLIB", fake}, {"PERL5OPT", "-I" + fake}} {
		t.Run(env[0], func(t *testing.T) {
			t.Setenv(env[0], env[1])
			assert.Empty(t, returnTypeFailures(t, p, oracleRun{Module: "Carp", TPM: tpm, Cases: cases}))
		})
	}
}

// xsUtilities are the XS utility modules declared as one batch (parse's
// heldtoperl_test.go holds the same list to perl's prototypes and exports).
var xsUtilities = []string{"Scalar::Util", "List::Util", "Cwd", "Data::Dumper", "MIME::Base64", "Storable"}

// TestXSUtilitiesReturnTypes holds the XS utility modules' declared return
// types to perl, with the call cases in testdata/oracle/.
func TestXSUtilitiesReturnTypes(t *testing.T) {
	holdReturnTypes(t, xsUtilities...)
}

// TestScalarUtilBlessedUnblessedReturnIsDeclared: blessed on an unblessed
// reference returns undef, which Scalar/Util.tpm's declared Str|Undef
// admits; testdata/oracle/Scalar-Util.pl's blessed case includes one,
// {}. Measured on 5.42.0, `Scalar::Util::blessed({})` is undef.
func TestScalarUtilBlessedUnblessedReturnIsDeclared(t *testing.T) {
	holdReturnTypes(t, "Scalar::Util")
}

// TestMIMEBase64InvalidInputReturnIsDeclared: decode_base64 given invalid
// base64 returns the garbage bytes its decoder produces, a Str,
// MIME/Base64.tpm's declared return type for it; testdata/oracle/
// MIME-Base64.pl's decode_base64 case includes one, 'not valid base64
// !!!'. Measured on 5.42.0, `MIME::Base64::decode_base64('not valid
// base64 !!!')` is defined, not undef, and not a death.
func TestMIMEBase64InvalidInputReturnIsDeclared(t *testing.T) {
	holdReturnTypes(t, "MIME::Base64")
}

// TestXSUtilitiesOracleStaysInScratch: Storable's oracle cases (fixture.store,
// store.out, and the rest) write under the oracle's temp directory, never
// the package directory `go test` runs from -- running Storable's return-type
// oracle leaves that directory's entries unchanged.
func TestXSUtilitiesOracleStaysInScratch(t *testing.T) {
	before := dirEntries(t, ".")
	holdReturnTypes(t, "Storable")
	after := dirEntries(t, ".")
	assert.Equal(t, before, after, "Storable's oracle cases wrote outside the test's scratch directory")
}

// dirEntries is the sorted names dir holds, for before/after comparisons
// that must see no new file appear.
func dirEntries(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	names := make([]string, len(entries))
	for i, e := range entries {
		names[i] = e.Name()
	}
	sort.Strings(names)
	return names
}
