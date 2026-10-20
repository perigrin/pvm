// ABOUTME: Holds each .tpm's declared parameter types to perl: the candidate a call case selects.
// ABOUTME: An argument that candidate's declared parameter excludes is a lossy coercion (RFC 0001).

package infer_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
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

// paramTypeAt is the type candidate s takes its i'th explicit argument
// as (never counting an invocant, which is Select's own concern, not
// this one): the slurpy's past the last parameter, and Any where no
// type is stated. It is Signature's own unexported paramType, which
// only internal/types needs for Select; the oracle needs the identical
// rule to hold a case's arguments to the very candidate Select would
// pick for it.
func paramTypeAt(s types.Signature, i int) types.Type {
	if len(s.Params) == 0 {
		return types.Any
	}
	p := s.Params[min(i, len(s.Params)-1)]
	if p.Type == types.Unknown {
		return types.Any
	}
	return p.Type
}

// parameterOracleFailures has perl build each of r.TPM's non-exempt
// call cases' own arguments (testdata/parameter_oracle.pl, never
// calling the sub itself), and holds every argument to the parameter
// type of the candidate types.Select or types.SelectInvoked picks for
// that case -- the same selection inference itself makes. RFC 0001
// "What a parameter type means": an argument fits when it satisfies
// the selected candidate's parameter (types.TypeSatisfies, the same
// membership the checker itself asks of a call site, which also takes
// a CodeRef for a Code slot -- sort's comparator and a `&`-sigil
// parameter both do, by the same rule); anything else is a coercion
// the oracle counts as lossy, unless the call case is a known,
// deliberate one (parameterExempt) or the sub's module is already
// pending its own audit fix (pendingAudit).
func parameterOracleFailures(t *testing.T, p perltest.Perl, r oracleRun) []string {
	t.Helper()
	sigs, errs := parse.DeclaredSignatures(r.TPM)
	if len(errs) > 0 {
		return []string{fmt.Sprintf("%s: declaration errors: %v", r.Module, errs)}
	}
	observer, err := filepath.Abs(filepath.Join("testdata", "observe_types.pl"))
	require.NoError(t, err)
	driver, err := filepath.Abs(filepath.Join("testdata", "parameter_oracle.pl"))
	require.NoError(t, err)

	args := []string{driver, observer, r.Cases, r.Module}
	if r.Inc != "" {
		args = append([]string{"-I", r.Inc}, args...)
	}
	args = append(args, oracleCallArgs(sigs, r.Module, r.Exempt)...)

	scratch, err := filepath.EvalSymlinks(t.TempDir())
	require.NoError(t, err)
	cmd := p.Command(args...)
	cmd.Dir = scratch
	cmd.Env = append(cmd.Env, "HOME="+scratch, "TMPDIR="+scratch)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return []string{fmt.Sprintf("%s: the parameter oracle did not run: %v: %s", r.Module, err, stderr.String())}
	}
	report, complete := strings.CutPrefix(string(out), "complete\n")
	if !complete {
		return []string{fmt.Sprintf("%s: the parameter oracle ended before observing every case: %s", r.Module, stderr.String())}
	}
	var obs []observation
	if err := json.Unmarshal([]byte(report), &obs); err != nil {
		return []string{fmt.Sprintf("%s: the parameter oracle emitted %q: %v", r.Module, out, err)}
	}

	return paramMisfits(r.Module, sigs, obs, r.IgnorePending)
}

// paramCall is one oracle call case the parameter oracle observed
// arguments for: the sub name, its context and form, and the case's
// own index (so two cases of the same sub and context are held apart).
type paramCall struct {
	name, ctx, form string
	index           int
}

// paramMisfits groups obs by the call case each argument observation
// belongs to, selects the candidate types.Select or types.SelectInvoked
// would pick for it, and reports each argument that candidate's
// declared parameter excludes. ignorePending turns off pendingAudit's
// and parameterExempt's own filtering, for TestPendingAuditEntriesStillMisfit.
func paramMisfits(module string, sigs map[string][]types.Signature, obs []observation, ignorePending bool) []string {
	groups := map[paramCall][]types.Type{}
	var order []paramCall
	for _, o := range obs {
		f := strings.Split(o.Name, ":")
		if len(f) != 5 {
			return []string{fmt.Sprintf("%s: the parameter oracle emitted an unparseable observation %q", module, o.Name)}
		}
		argIdx, err := strconv.Atoi(f[4])
		if err != nil {
			return []string{fmt.Sprintf("%s: the parameter oracle emitted a non-numeric argument index %q", module, o.Name)}
		}
		key := paramCall{name: f[0], ctx: f[1], form: f[2], index: o.Line}
		if _, ok := groups[key]; !ok {
			order = append(order, key)
		}
		lst := groups[key]
		for len(lst) <= argIdx {
			lst = append(lst, types.Unknown)
		}
		got, known := observedToLattice[o.Type]
		if !known {
			return []string{fmt.Sprintf("%s: observed argument type %q has no lattice mapping", module, o.Type)}
		}
		lst[argIdx] = got
		groups[key] = lst
	}

	var failures []string
	for _, key := range order {
		if !ignorePending && (pendingAuditMatch(module, key.name, "parameter") || parameterExemptMatch(module, key.name)) {
			continue
		}
		argTypes := groups[key]
		cands := sigs[key.name]
		ctx := types.ScalarCtx
		if key.ctx == "list" {
			ctx = types.ListCtx
		}
		var sel types.Selection
		if key.form == "method" {
			sel = types.SelectInvoked(cands, append([]types.Type{types.Unknown}, argTypes...), "", ctx)
		} else {
			sel = types.SelectShaped(cands, argTypes, "", ctx)
		}
		switch sel.Outcome {
		case types.Undecided:
			// Every candidate Taking() kept already passed fits():
			// an ambiguous call is one none of its fitting
			// candidates is ruled out by, so there is no single
			// candidate's parameter left for the oracle to hold
			// this case to.
			continue
		case types.Failed:
			failures = append(failures, fmt.Sprintf(
				"%s::%s call case %d in %s context: no candidate's arity takes the %d arguments its own call case built",
				module, key.name, key.index, key.ctx, len(argTypes)))
			continue
		}
		cand := cands[sel.Candidate]
		for i, a := range argTypes {
			if a == types.Unknown {
				continue
			}
			if pt := paramTypeAt(cand, i); !types.TypeSatisfies(a, pt) {
				failures = append(failures, fmt.Sprintf(
					"%s::%s call case %d in %s context passed argument %d as %s, which the selected candidate's declared parameter (%s) excludes as a lossy coercion",
					module, key.name, key.index, key.ctx, i, a, pt))
			}
		}
	}
	sort.Strings(failures)
	return failures
}

// parameterExempt are the oracle call cases that deliberately pass an
// argument a declared Boolean parameter admits only by Perl's universal
// boolean coercion -- any scalar is true or false in an if -- rather
// than by type membership, "Module::sub" to why. A literal 0 or 1 is
// the idiomatic Perl spelling of such a flag (unlike builtin::true,
// builtin::false or a comparison's own result, each of which observes
// as Boolean already), and every entry's own oracle case is the
// module's documented flag argument, not a declaration gap.
var parameterExempt = map[string]string{
	"File::Spec::splitpath":       "oracle/File-Spec.pl passes 1, the idiomatic Perl literal, for $no_file's documented boolean flag (its source does `if ($nofile)`, a truth test)",
	"File::Spec::Unix::splitpath": "oracle/File-Spec-Unix.pl passes 1, the idiomatic Perl literal, for $no_file's documented boolean flag (its source does `if ($nofile)`, a truth test)",
	"File::stat::stat_cando":      "oracle/File-stat.pl passes 1, the idiomatic Perl literal, for $effective's documented boolean flag",
}

// parameterExemptMatch reports whether module::sub is in parameterExempt.
func parameterExemptMatch(module, sub string) bool {
	_, ok := parameterExempt[module+"::"+sub]
	return ok
}

// TestParameterExemptionsNameAReason: every parameterExempt entry says why.
func TestParameterExemptionsNameAReason(t *testing.T) {
	for sub, reason := range parameterExempt {
		assert.NotEmpty(t, strings.TrimSpace(reason), "%s is exempt from the parameter oracle without a reason", sub)
	}
}

// parameterOracleSkip is the reason to skip module's parameter oracle
// on this platform, "" when there is none -- the same platform gaps
// the matching return-type oracle test already skips (see each named
// test for the measurement), since a constant or method undefined here
// could as easily unbalance a call case's own arguments as its return.
func parameterOracleSkip(module string) string {
	switch runtime.GOOS {
	case "darwin":
		switch {
		case module == "Fcntl" || module == "Socket":
			return "Fcntl's Linux-only dnotify, seal and pipe-size constants are not defined on macOS's perl -- see TestConstantExportersReturnTypes"
		case module == "POSIX":
			return "POSIX's limit and locale constants differ on macOS's perl -- see TestPOSIXReturnTypes"
		case slices.Contains(perltest.IOExporters, module):
			return "FileHandle, IO::File and IO::Socket's Linux-only fcntl and socket-family constants are absent or differ on macOS's perl -- see TestIOExportersReturnTypes"
		case slices.Contains(perltest.XSLibraries, module):
			return "Time::HiRes's Linux-only clock ids are undefined or differ on macOS's perl -- see TestXSLibrariesReturnTypes"
		}
	case "windows":
		switch {
		case module == "POSIX":
			return `oracle/POSIX.pl calls Fcntl::F_GETFD, undefined on Windows's perl -- see TestPOSIXReturnTypes`
		case slices.Contains(perltest.XSLibraries, module):
			return "Time::HiRes::clock is unimplemented on Windows's perl -- see TestXSLibrariesReturnTypes"
		}
	}
	if runtime.GOOS != "linux" && slices.Contains(perltest.PureExporters, module) {
		return "Errno's Linux-only error codes are not defined on this platform's perl -- see TestPurePerlExportersReturnTypes"
	}
	return ""
}

// TestOracleCasesFitDeclaredParameters holds every module the return-type
// oracle holds (perltest.AllHeld) to the parameter oracle too: each
// call case's own arguments, built the same way return_oracle.pl builds
// them, must fit the candidate selection picks for it.
func TestOracleCasesFitDeclaredParameters(t *testing.T) {
	p := perltest.Perl542(t)
	for _, module := range perltest.AllHeld() {
		t.Run(module, func(t *testing.T) {
			if reason := parameterOracleSkip(module); reason != "" {
				t.Skip(reason)
			}
			rel := strings.ReplaceAll(module, "::", "/")
			tpm, err := os.ReadFile(filepath.Join("..", "parse", "declarations", filepath.FromSlash(rel)+".tpm"))
			require.NoError(t, err, "%s: no declaration", module)
			cases, err := filepath.Abs(filepath.Join("testdata", "oracle", strings.ReplaceAll(module, "::", "-")+".pl"))
			require.NoError(t, err)
			for _, f := range parameterOracleFailures(t, p, oracleRun{Module: module, TPM: tpm, Cases: cases, Exempt: exemptMapFor(module)}) {
				t.Error(f)
			}
		})
	}
}

// splitModuleSub splits a parameterExempt key, "Module::sub", on its
// last "::": the module name may itself hold "::" (File::Spec::Unix),
// but the sub name never does.
func splitModuleSub(key string) (module, sub string) {
	i := strings.LastIndex(key, "::")
	return key[:i], key[i+2:]
}

// pendingOracleRun is module's real .tpm and oracle cases, run with
// IgnorePending set: pendingAudit's and parameterExempt's own filtering
// turned off, so the oracle's raw failures tell whether an entry's
// misfit is still real.
func pendingOracleRun(t *testing.T, module string) oracleRun {
	t.Helper()
	rel := strings.ReplaceAll(module, "::", "/")
	tpm, err := os.ReadFile(filepath.Join("..", "parse", "declarations", filepath.FromSlash(rel)+".tpm"))
	require.NoError(t, err, "%s: no declaration", module)
	cases, err := filepath.Abs(filepath.Join("testdata", "oracle", strings.ReplaceAll(module, "::", "-")+".pl"))
	require.NoError(t, err)
	return oracleRun{Module: module, TPM: tpm, Cases: cases, Exempt: exemptMapFor(module), IgnorePending: true}
}

// TestPendingAuditEntriesStillMisfit holds every pendingAudit and
// parameterExempt entry to the oracle run it defers or excuses: run
// with the entry's own filtering off, its module and sub must still
// produce a misfit ("Module::Sub call case" -- both oracles' failure
// messages start this way). A fix that lands without deleting its own
// entry here stops reproducing and fails loudly, rather than the entry
// silently going stale.
func TestPendingAuditEntriesStillMisfit(t *testing.T) {
	p := perltest.Perl542(t)
	type target struct{ module, sub, kind string }
	var targets []target
	for _, entries := range pendingAudit {
		for _, m := range entries {
			targets = append(targets, target{m.Module, m.Sub, m.Kind})
		}
	}
	for key := range parameterExempt {
		module, sub := splitModuleSub(key)
		targets = append(targets, target{module, sub, "parameter"})
	}

	for _, tgt := range targets {
		t.Run(tgt.module+"::"+tgt.sub+"/"+tgt.kind, func(t *testing.T) {
			if reason := parameterOracleSkip(tgt.module); reason != "" {
				t.Skip(reason)
			}
			run := pendingOracleRun(t, tgt.module)
			var failures []string
			switch tgt.kind {
			case "return":
				failures = returnTypeFailures(t, p, run)
			case "parameter":
				failures = parameterOracleFailures(t, p, run)
			default:
				t.Fatalf("unknown pendingAudit kind %q", tgt.kind)
			}
			want := fmt.Sprintf("%s::%s call case", tgt.module, tgt.sub)
			assert.True(t, slices.ContainsFunc(failures, func(f string) bool { return strings.Contains(f, want) }),
				"%s::%s no longer misfits; delete its entry", tgt.module, tgt.sub)
		})
	}
}

// TestParameterOracleRejectsANarrowParam: a declaration whose parameter
// is narrower than perl accepts fails the oracle on the call case that
// passes the excluded argument, and widening the parameter passes it.
func TestParameterOracleRejectsANarrowParam(t *testing.T) {
	p := perltest.Perl542(t)
	tpm := []byte("package My::Oracle;\nsub greet (Int $n) Str;\n")
	cases := "(greet => [ ['many'] ]);\n"

	lib, caseFile := oracleFixture(t, "My::Oracle", "package My::Oracle;\nsub greet { \"hi $_[0]\" }\n1;\n", cases)
	failures := parameterOracleFailures(t, p, oracleRun{Module: "My::Oracle", TPM: tpm, Cases: caseFile, Inc: lib})
	require.Len(t, failures, 1, "a Str argument where Int is declared")
	assert.Contains(t, failures[0], "greet")
	assert.Contains(t, failures[0], "lossy coercion")

	tpm = []byte("package My::Oracle;\nsub greet (Str $n) Str;\n")
	assert.Empty(t, parameterOracleFailures(t, p, oracleRun{Module: "My::Oracle", TPM: tpm, Cases: caseFile, Inc: lib}))
}
