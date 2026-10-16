// ABOUTME: Holds each .tpm's declared return types to perl: perl calls each sub with its oracle call cases.
// ABOUTME: A value perl returns that the declared return type excludes fails the test.

package infer_test

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sort"
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
		args = append(args, name+":"+callContext(sigs[name]))
	}

	scratch := t.TempDir()
	cmd := exec.Command(p.Path, args...)
	cmd.Dir = scratch
	cmd.Env = append(os.Environ(), "HOME="+scratch, "TMPDIR="+scratch)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return []string{fmt.Sprintf("%s: the oracle did not run: %v: %s", r.Module, err, stderr.String())}
	}
	var obs []observation
	if err := json.Unmarshal(out, &obs); err != nil {
		return []string{fmt.Sprintf("%s: the oracle emitted %q: %v", r.Module, out, err)}
	}

	var failures []string
	for _, o := range obs {
		got, known := observedToLattice[o.Type]
		admitted := known && slices.ContainsFunc(sigs[o.Name], func(s types.Signature) bool {
			return types.IsSubtype(got, s.Returns)
		})
		if !admitted {
			failures = append(failures, fmt.Sprintf("%s::%s call case %d returned %s, which its declared return type excludes",
				r.Module, o.Name, o.Line, o.Type))
		}
	}
	sort.Strings(failures)
	return failures
}

// callContext is the context a sub's call cases are called in: a sub
// declared never to return is called inside eval, one returning a scalar
// in scalar context, and any other in list context.
//
// ponytail: a declared List[T]'s T is not in types.Signature, so a list
// return is held only to List, which admits every value; the element
// type is checked once the signature carries it.
func callContext(sigs []types.Signature) string {
	switch {
	case slices.ContainsFunc(sigs, func(s types.Signature) bool { return s.Returns == types.None }):
		return "none"
	case slices.ContainsFunc(sigs, func(s types.Signature) bool { return !types.IsSubtype(s.Returns, types.Scalar) }):
		return "list"
	}
	return "scalar"
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
