// ABOUTME: Holds a .tpm's prototypes and export marks to perl itself: each declaration batch calls
// ABOUTME: holdPrototypes and holdExports with its modules. Fixture modules test both helpers.

package parse

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tamarou.com/pvm/internal/perltest"
	"tamarou.com/pvm/internal/types"
)

// holdPrototypes holds each module's shipped .tpm to perl: every sub it
// declares has the prototype perl's prototype() reports, or both have
// none. A batch of declarations calls it once with its modules.
func holdPrototypes(t *testing.T, modules ...string) {
	t.Helper()
	hold(t, prototypeDisagreements, modules)
}

// holdExports holds each module's shipped .tpm to perl: its :export and
// :export_ok marks are what `use` installs in a caller (exportDisagreements).
// A batch of declarations calls it once with its modules.
func holdExports(t *testing.T, modules ...string) {
	t.Helper()
	hold(t, exportDisagreements, modules)
}

// hold reports, as test errors, how each module's shipped .tpm disagrees
// with perl 5.42 by check.
func hold(t *testing.T, check func(*testing.T, perltest.Perl, string, string, []byte) []string, modules []string) {
	t.Helper()
	holdFiltered(t, check, modules, nil)
}

// heldExempt holds, per runtime.GOOS, prototypeDisagreements or
// exportDisagreements lines read as a substring (to survive a volatile
// suffix perl's own message appends, such as an eval sequence number)
// that this platform's perl genuinely produces and Linux's does not --
// each cites the CI job that measured it. holdFiltered drops any
// disagreement containing one of these, so every other disagreement
// in the same modules still fails the test if it reappears, and
// TestHeldExemptionsAreChecked fails on an entry that no longer matches.
func heldExempt() []string {
	// is_8bit, is_16bit, utf8_upgrade and utf8_downgrade are each in
	// Encode's @EXPORT_OK and ':all' tag, and `use Encode
	// qw(is_8bit)` does not die, but Encode.pm never defines a body
	// for any of the four (Encode.tpm's own comment): measured, every
	// call dies "Undefined subroutine", so this platform-independent
	// fact holds on every OS, not only the one heldExempt's own
	// Windows-only entries below are measured on.
	exempt := []string{
		"is_8bit is marked :export_ok; use Encode qw(is_8bit) does not install it: installs",
		"is_16bit is marked :export_ok; use Encode qw(is_16bit) does not install it: installs",
		"utf8_upgrade is marked :export_ok; use Encode qw(utf8_upgrade) does not install it: installs",
		"utf8_downgrade is marked :export_ok; use Encode qw(utf8_downgrade) does not install it: installs",
		"is_8bit is marked all tag; use Encode ':all' does not install it",
		"is_16bit is marked all tag; use Encode ':all' does not install it",
		"utf8_upgrade is marked all tag; use Encode ':all' does not install it",
		"utf8_downgrade is marked all tag; use Encode ':all' does not install it",
	}
	if runtime.GOOS != "windows" {
		return exempt
	}
	return append(exempt,
		// measured in CI job 39489951290
		"Time::HiRes::getitimer declares prototype ($); perl's is none",
		"Time::HiRes::setitimer declares prototype ($$;$); perl's is none",
		"Time::HiRes::utime declares prototype ($$@); perl's is ($$;@)",
		"clock is marked :export_ok; use Time::HiRes qw(clock) does not install it: dies Time::HiRes::clock(): unimplemented in this platform",
		"clock_nanosleep is marked :export_ok; use Time::HiRes qw(clock_nanosleep) does not install it: dies Time::HiRes::clock_nanosleep(): unimplemented in this platform",
		"nanosleep is marked :export_ok; use Time::HiRes qw(nanosleep) does not install it: dies Time::HiRes::nanosleep(): unimplemented in this platform",
		"ualarm is marked :export_ok; use Time::HiRes qw(ualarm) does not install it: dies Time::HiRes::ualarm(): unimplemented in this platform",
		"utime is marked :export_ok; use Time::HiRes qw(utime) does not install it: dies Time::HiRes::utime(): unimplemented in this platform",
		"File::Spec::Functions::case_tolerant declares prototype (); perl's is none",
		"File::Spec::Functions::devnull declares prototype (); perl's is none",
		"File::Spec::Functions::rootdir declares prototype (); perl's is none",
	)
}

// heldExemptHolds are the holds that filter by heldExempt: the checks
// and modules TestXSLibrariesPrototypes, TestXSLibrariesExports and
// TestIOExportersPrototypes run. TestHeldExemptionsAreChecked holds
// every heldExempt entry to the disagreements these produce.
var heldExemptHolds = []struct {
	check   func(*testing.T, perltest.Perl, string, string, []byte) []string
	modules []string
}{
	{prototypeDisagreements, perltest.XSLibraries},
	{exportDisagreements, perltest.XSLibraries},
	{prototypeDisagreements, perltest.IOExporters},
}

// TestHeldExemptionsAreChecked: every heldExempt entry this OS keeps
// still matches a disagreement its holds produce on this perl -- one
// that matches none is stale, and would hide that disagreement if it
// came back somewhere it is not expected.
func TestHeldExemptionsAreChecked(t *testing.T) {
	p := perltest.Perl542(t)
	var disagreements []string
	for _, h := range heldExemptHolds {
		for _, module := range h.modules {
			src, ok := declaration(module)
			require.True(t, ok, "%s: no declaration", module)
			disagreements = append(disagreements, h.check(t, p, "", module, src)...)
		}
	}
	for _, entry := range unmatchedFilters(heldExempt(), disagreements) {
		t.Errorf("heldExempt entry %q matches no disagreement on %s: stale", entry, p.Path)
	}
}

// unmatchedFilters is the entries of filter no disagreement contains.
func unmatchedFilters(filter, disagreements []string) []string {
	var unmatched []string
	for _, entry := range filter {
		if !slices.ContainsFunc(disagreements, func(d string) bool { return strings.Contains(d, entry) }) {
			unmatched = append(unmatched, entry)
		}
	}
	return unmatched
}

// TestUnmatchedFiltersReportsAStaleEntry: a filter entry no
// disagreement contains is reported; one some disagreement contains is
// not.
func TestUnmatchedFiltersReportsAStaleEntry(t *testing.T) {
	got := unmatchedFilters([]string{"used", "stale"}, []string{"a disagreement that is used here"})
	assert.Equal(t, []string{"stale"}, got)
}

// holdFiltered is hold, with any disagreement containing one of
// filter's entries dropped before it is reported.
func holdFiltered(t *testing.T, check func(*testing.T, perltest.Perl, string, string, []byte) []string, modules []string, filter []string) {
	t.Helper()
	p := perltest.Perl542(t)
	for _, module := range modules {
		src, ok := declaration(module)
		if !ok {
			t.Errorf("%s: no declaration", module)
			continue
		}
		for _, d := range check(t, p, "", module, src) {
			if containsAny(d, filter) {
				continue
			}
			if runtime.GOOS != "linux" && isExtraInstall(d) {
				continue
			}
			t.Error(d)
		}
	}
}

// containsAny reports whether s contains any of substrs.
func containsAny(s string, substrs []string) bool {
	for _, sub := range substrs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

// isExtraInstall reports whether d is sameNames' or exportDisagreements'
// "perl installs X; the .tpm does not mark it" direction -- real
// information only when checking a shipped .tpm, which describes Linux
// perl 5.42.0, against that same Linux perl: off Linux, "installs more
// than the .tpm says" just means that platform's perl exports more than
// Linux's did (Windows's Errno and POSIX each add hundreds of Winsock-
// and SDK-only names, measured), which the .tpm was never trying to
// settle. holdFiltered and holdPurePerlExporters drop it there, for a
// real module's .tpm; a fixture module that tests exportDisagreements
// itself (TestExportHelperRejectsAMissingName and its siblings) calls it
// directly, bypassing this, so it still sees both directions on every
// platform.
func isExtraInstall(d string) bool {
	return strings.Contains(d, "the .tpm does not mark it")
}

// holdPurePerlExporters is hold for perltest.PureExporters, with Errno's
// names this perl does not define at all filtered out first --
// perltest.MissingSubs measures them directly, since which constants
// Errno exports depends on the C library (and, on Windows, the SDK and
// Winsock headers) perl was built against, not something this repository
// can hand-list once for every platform CI runs on.
func holdPurePerlExporters(t *testing.T, check func(*testing.T, perltest.Perl, string, string, []byte) []string) {
	t.Helper()
	p := perltest.Perl542(t)
	for _, module := range perltest.PureExporters {
		src, ok := declaration(module)
		if !ok {
			t.Errorf("%s: no declaration", module)
			continue
		}
		var missing map[string]string
		if module == "Errno" {
			if facts, err := readHeld(module, src); err == nil {
				missing = perltest.MissingSubs(t, p, "", module, slices.Sorted(maps.Keys(facts.protos)))
			}
		}
		for _, d := range check(t, p, "", module, src) {
			if missingName(d, module, missing) {
				continue
			}
			if runtime.GOOS != "linux" && isExtraInstall(d) {
				continue
			}
			t.Error(d)
		}
	}
}

// missingName reports whether disagreement d is about a name in
// missing: exportDisagreements leads with the bare name ("EADV is
// marked..."), prototypeDisagreements with "module::name" ("Errno::EADV
// is declared..."), so both leading shapes are checked.
func missingName(d, module string, missing map[string]string) bool {
	for name := range missing {
		if strings.HasPrefix(d, name+" ") || strings.HasPrefix(d, module+"::"+name+" ") {
			return true
		}
	}
	return false
}

// TestMissingNameMatchesBothDisagreementShapes: exportDisagreements'
// bare-name line and prototypeDisagreements' module::name line both
// match an entry in missing, and an unrelated line matches neither.
func TestMissingNameMatchesBothDisagreementShapes(t *testing.T) {
	missing := map[string]string{"EADV": "perl defines no sub Errno::EADV"}
	assert.True(t, missingName("EADV is marked :export_ok; use Errno qw(EADV) does not install it: dies ...", "Errno", missing))
	assert.True(t, missingName("Errno::EADV is declared; perl defines no such sub", "Errno", missing))
	assert.False(t, missingName("Errno::E2BIG declares prototype (); perl's is none", "Errno", missing))
}

// heldToPerl runs testdata/held_to_perl.pl's mode for module, perl
// searching inc first unless it is "", and returns its output lines
// split at tabs. Perl finds the module through @INC and %INC, never a
// path joined to privlib: measured on 5.42.0, 23 of the 50 most-used core
// modules load from archlib.
func heldToPerl(t *testing.T, p perltest.Perl, inc, mode, module string, args ...string) ([][]string, error) {
	t.Helper()
	script, err := filepath.Abs(filepath.Join("testdata", "held_to_perl.pl"))
	require.NoError(t, err)
	argv := append([]string{script, mode, module}, args...)
	if inc != "" {
		argv = append([]string{"-I", inc}, argv...)
	}
	cmd := p.Command(argv...)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("%s: %v: %s", module, err, strings.TrimSpace(stderr.String()))
	}
	var lines [][]string
	for _, line := range strings.Split(string(out), "\n") {
		if line == "" {
			continue
		}
		lines = append(lines, strings.Split(strings.TrimRight(line, "\r"), "\t"))
	}
	return lines, nil
}

// modulePath is the file perl loads module from: $INC{'M.pm'} after
// `require M`.
func modulePath(t *testing.T, p perltest.Perl, inc, module string) (string, error) {
	t.Helper()
	lines, err := heldToPerl(t, p, inc, "inc", module)
	if err != nil {
		return "", err
	}
	return lines[0][0], nil
}

// tpmSkeleton runs held_to_perl.pl's skeleton mode for module and returns
// the .tpm skeleton text it prints, the names it reports installed, with
// a TODO return type (sub) or TODO type (our) a worker must replace.
func tpmSkeleton(t *testing.T, module string) string {
	t.Helper()
	p := perltest.Perl542(t)
	lines, err := heldToPerl(t, p, "", "skeleton", module)
	require.NoError(t, err)
	rows := make([]string, len(lines))
	for i, line := range lines {
		rows[i] = strings.Join(line, "\t")
	}
	return strings.Join(rows, "\n") + "\n"
}

// readHeld reads a .tpm's source as the resolver does, refusing one in
// error: a declaration perl is held to must first be one psc reads.
func readHeld(module string, src []byte) (moduleFacts, error) {
	facts := readLibraryDeclaration(src, nil)
	if len(facts.errs) > 0 {
		return facts, fmt.Errorf("%s: declaration errors: %v", module, facts.errs)
	}
	return facts, nil
}

// prototypeDisagreements is how the prototypes of the subs src declares
// for module differ from perl's prototype(): one line each.
func prototypeDisagreements(t *testing.T, p perltest.Perl, inc, module string, src []byte) []string {
	t.Helper()
	facts, err := readHeld(module, src)
	if err != nil {
		return []string{err.Error()}
	}
	lines, err := heldToPerl(t, p, inc, "prototypes", module, slices.Sorted(maps.Keys(facts.protos))...)
	if err != nil {
		return []string{err.Error()}
	}
	var out []string
	for _, f := range lines {
		name, declared := f[0], facts.protos[f[0]]
		perls := ""
		switch f[1] {
		case "missing":
			out = append(out, fmt.Sprintf("%s::%s is declared; perl defines no such sub", module, name))
			continue
		case "proto":
			perls = "(" + f[2] + ")"
		}
		if declared != perls {
			out = append(out, fmt.Sprintf("%s::%s declares prototype %s; perl's is %s", module, name, orNone(declared), orNone(perls)))
		}
	}
	return out
}

// orNone is a prototype as a message spells it.
func orNone(proto string) string {
	if proto == "" {
		return "none"
	}
	return proto
}

// exportDisagreements is how the :export and :export_ok marks src
// declares for module differ from what perl's `use` installs in a fresh
// caller, one line each (ruling 1, 2026-10-16: the marks are held to
// installs, never to the module's lists). A bare `use` must install the
// :export names and no others; `use M qw(name)` must install each
// :export_ok name, and a name it installs that the .tpm does not mark is
// missing; `use M ':tag'` must install exactly the names marked with tag.
// The names asked for are the marked ones and those in the module's
// @EXPORT and @EXPORT_OK, an `&name` there being the sub name, and the
// tags asked for are the declared ones and those in its %EXPORT_TAGS, so a
// tag the .tpm omits fails by the names it installs.
func exportDisagreements(t *testing.T, p perltest.Perl, inc, module string, src []byte) []string {
	t.Helper()
	facts, err := readHeld(module, src)
	if err != nil {
		return []string{err.Error()}
	}
	lists := []string{""}
	for _, name := range facts.exports {
		if !slices.Contains(facts.defaults, name) {
			lists = append(lists, name)
		}
	}
	for _, tag := range slices.Sorted(maps.Keys(facts.tags)) {
		lists = append(lists, ":"+tag)
	}
	lines, err := heldToPerl(t, p, inc, "installs", module, lists...)
	if err != nil {
		return []string{err.Error()}
	}
	var out []string
	for _, f := range lines {
		list, installs := f[0], f[2:]
		if f[1] == "dies" {
			installs = nil
		}
		switch {
		case list == "":
			out = append(out, sameNames(fmt.Sprintf("use %s", module), installs, facts.defaults, ":export")...)
		case strings.HasPrefix(list, ":"):
			out = append(out, sameNames(fmt.Sprintf("use %s '%s'", module, list), installs, facts.tags[list[1:]], list[1:]+" tag")...)
		// An :export name is held to the bare use alone: Test::More's import
		// reads `use Test::More qw(ok)` as a plan and dies.
		case slices.Contains(facts.exports, list) && !slices.Contains(facts.defaults, list) && !slices.Contains(installs, list):
			out = append(out, fmt.Sprintf("%s is marked :export_ok; use %s qw(%s) does not install it: %s", list, module, list, strings.Join(f[1:], " ")))
		case !slices.Contains(facts.exports, list) && slices.Contains(installs, list):
			out = append(out, fmt.Sprintf("use %s qw(%s) installs %s; the .tpm does not mark it", module, list, list))
		}
	}
	return out
}

// sameNames is how the names use installs differ from those the .tpm
// marks with mark.
func sameNames(use string, installs, marked []string, mark string) []string {
	var out []string
	for _, name := range installs {
		if !slices.Contains(marked, name) {
			out = append(out, fmt.Sprintf("%s installs %s; the .tpm does not mark it %s", use, name, mark))
		}
	}
	for _, name := range marked {
		if !slices.Contains(installs, name) {
			out = append(out, fmt.Sprintf("%s is marked %s; %s does not install it", name, mark, use))
		}
	}
	return out
}

// perlFixture writes module's perl source under a lib directory in a temp
// dir and returns that directory.
func perlFixture(t *testing.T, module, pm string) string {
	t.Helper()
	lib := t.TempDir()
	perltest.WriteModule(t, lib, module, pm)
	return lib
}

// TestPrototypeHelperRejectsADisagreement: a declared prototype perl's
// prototype() does not report fails, and the one it reports passes.
func TestPrototypeHelperRejectsADisagreement(t *testing.T) {
	p := perltest.Perl542(t)
	lib := perlFixture(t, "My::Proto", "package My::Proto;\nsub f ($$) { }\n1;\n")

	got := prototypeDisagreements(t, p, lib, "My::Proto", []byte("package My::Proto;\nsub f :prototype($);\n"))
	require.Len(t, got, 1)
	assert.Contains(t, got[0], "My::Proto::f")
	assert.Empty(t, prototypeDisagreements(t, p, lib, "My::Proto", []byte("package My::Proto;\nsub f :prototype($$);\n")))
}

// TestPrototypeHelperDistinguishesUndefFromEmpty: a sub with no
// prototype is not one with the empty prototype (), either way round.
func TestPrototypeHelperDistinguishesUndefFromEmpty(t *testing.T) {
	p := perltest.Perl542(t)
	lib := perlFixture(t, "My::Empty", "package My::Empty;\nsub none { }\nsub empty () { 1 }\n1;\n")

	assert.Len(t, prototypeDisagreements(t, p, lib, "My::Empty", []byte("package My::Empty;\nsub none () Int;\nsub empty (List @x) Int;\n")), 2,
		"none declared (), empty declared with no prototype")
	assert.Empty(t, prototypeDisagreements(t, p, lib, "My::Empty", []byte("package My::Empty;\nsub none (List @x) Int;\nsub empty () Int;\n")))
}

// TestPrototypeHelperHoldsADeclarationWithNoSubs: a .tpm declaring only
// a variable, as Config's does, has no prototype to disagree with.
func TestPrototypeHelperHoldsADeclarationWithNoSubs(t *testing.T) {
	p := perltest.Perl542(t)
	lib := perlFixture(t, "My::Vars", "package My::Vars;\nuse Exporter 'import';\nour @EXPORT = qw(%V);\nour %V;\n1;\n")
	tpm := []byte("package My::Vars;\nour Hash[Str] %V :export;\n")

	assert.Empty(t, prototypeDisagreements(t, p, lib, "My::Vars", tpm))
	assert.Empty(t, exportDisagreements(t, p, lib, "My::Vars", tpm))
}

// exporter is a fixture module exporting by Exporter's lists.
const exporter = "package My::Exp;\nuse Exporter 'import';\nour @EXPORT = qw(f);\nour @EXPORT_OK = qw(g h);\nour %EXPORT_TAGS = (gh => [qw(g h)]);\nsub f { } sub g { } sub h { }\n1;\n"

// TestExportHelperRejectsAMissingName: a name `use M qw(name)` installs
// that the .tpm does not mark fails.
func TestExportHelperRejectsAMissingName(t *testing.T) {
	p := perltest.Perl542(t)
	lib := perlFixture(t, "My::Exp", exporter)

	got := exportDisagreements(t, p, lib, "My::Exp", []byte("package My::Exp;\nsub f :export;\nsub g :export_ok(gh);\n"))
	assert.Contains(t, strings.Join(got, "\n"), "qw(h) installs h")
	assert.Empty(t, exportDisagreements(t, p, lib, "My::Exp",
		[]byte("package My::Exp;\nsub f :export;\nsub g :export_ok(gh);\nsub h :export_ok(gh);\n")))
}

// TestExportHelperRejectsAnExtraName: a name marked :export that a bare
// `use` does not install fails.
func TestExportHelperRejectsAnExtraName(t *testing.T) {
	p := perltest.Perl542(t)
	lib := perlFixture(t, "My::Exp", exporter)

	got := exportDisagreements(t, p, lib, "My::Exp",
		[]byte("package My::Exp;\nsub f :export;\nsub g :export(gh);\nsub h :export_ok(gh);\n"))
	require.Len(t, got, 1)
	assert.Contains(t, got[0], "g is marked :export")
}

// TestExportHelperMeasuresInstallsNotLists: a module whose own import
// installs a name its (empty) lists lack is held to the install.
func TestExportHelperMeasuresInstallsNotLists(t *testing.T) {
	p := perltest.Perl542(t)
	lib := perlFixture(t, "My::Own", "package My::Own;\nour @EXPORT;\nsub import { no strict 'refs'; *{caller() . '::loc'} = sub { 1 } }\n1;\n")

	got := exportDisagreements(t, p, lib, "My::Own", []byte("package My::Own;\n"))
	require.Len(t, got, 1)
	assert.Contains(t, got[0], "installs loc")
	assert.Empty(t, exportDisagreements(t, p, lib, "My::Own", []byte("package My::Own;\nsub loc :export;\n")))
}

// TestExportHelperCountsAutoloadedSubs: a sub declared for AUTOLOAD,
// existing and not yet defined, is installed, as POSIX's are.
func TestExportHelperCountsAutoloadedSubs(t *testing.T) {
	p := perltest.Perl542(t)
	lib := perlFixture(t, "My::Auto",
		"package My::Auto;\nuse Exporter 'import';\nour @EXPORT = qw(later);\nsub later;\nsub AUTOLOAD { 1 }\n1;\n")

	assert.Empty(t, exportDisagreements(t, p, lib, "My::Auto", []byte("package My::Auto;\nsub later :export;\n")))
}

// TestExportHelperTreatsAmpersandAsSubName: `&f` in an export list is the
// sub f, matched by `sub f :export` and not reported a second time.
func TestExportHelperTreatsAmpersandAsSubName(t *testing.T) {
	p := perltest.Perl542(t)
	lib := perlFixture(t, "My::Amp",
		"package My::Amp;\nuse Exporter 'import';\nour @EXPORT = qw(&f);\nour @EXPORT_OK = qw(&g);\nsub f { } sub g { }\n1;\n")

	assert.Empty(t, exportDisagreements(t, p, lib, "My::Amp", []byte("package My::Amp;\nsub f :export;\nsub g :export_ok;\n")))
	got := exportDisagreements(t, p, lib, "My::Amp", []byte("package My::Amp;\nsub f :export;\n"))
	require.Len(t, got, 1, "&g is the sub g, unmarked")
	assert.Contains(t, got[0], "qw(g) installs g")
}

// TestHeldToPerlFailsOnAModulePerlCannotLoad: a declared module perl
// cannot require fails each helper by name.
func TestHeldToPerlFailsOnAModulePerlCannotLoad(t *testing.T) {
	p := perltest.Perl542(t)
	tpm := []byte("package No::Such::Module;\nsub f :export;\n")
	for _, got := range [][]string{
		prototypeDisagreements(t, p, "", "No::Such::Module", tpm),
		exportDisagreements(t, p, "", "No::Such::Module", tpm),
	} {
		require.Len(t, got, 1)
		assert.Contains(t, got[0], "cannot require No::Such::Module")
	}
}

// TestHeldToPerlFindsArchlibModules: POSIX loads from archlib, where
// joining privlib with its path finds nothing; perl's %INC finds it.
func TestHeldToPerlFindsArchlibModules(t *testing.T) {
	p := perltest.Perl542(t)
	path, err := modulePath(t, p, "", "POSIX")
	require.NoError(t, err)
	assert.FileExists(t, path)
	if runtime.GOOS == "windows" {
		// Strawberry perl ships POSIX.pm as pure Perl under lib, not
		// archlib -- measured in CI job 39489951290: C:/Strawberry/perl/lib/POSIX.pm
		// is not in archlib C:\STRAWB~1\perl\lib. perl's %INC and
		// $Config{privlibexp} can each spell that directory differently
		// (an 8.3 short name here, measured), so PathUnder canonicalizes
		// both through perl before comparing, rather than a bare prefix
		// check.
		assert.True(t, perltest.PathUnder(t, p, path, p.Privlib), "%s is not in privlib %s", path, p.Privlib)
		return
	}
	assert.True(t, strings.HasPrefix(path, p.Archlib), "%s is not in archlib %s", path, p.Archlib)
	assert.NoFileExists(t, filepath.Join(p.Privlib, "POSIX.pm"))
}

// TestCarpPrototypes holds Carp.tpm's prototypes to perl.
func TestCarpPrototypes(t *testing.T) {
	holdPrototypes(t, "Carp")
}

// TestCarpExports holds Carp.tpm's export marks to what `use Carp` installs.
func TestCarpExports(t *testing.T) {
	holdExports(t, "Carp")
}

// TestPrototypeHelperRejectsASubPerlLacks: a declared sub perl does not
// define fails by name, rather than passing as a sub with no prototype.
func TestPrototypeHelperRejectsASubPerlLacks(t *testing.T) {
	p := perltest.Perl542(t)
	lib := perlFixture(t, "My::Lacks", "package My::Lacks;\nsub f { }\n1;\n")

	got := prototypeDisagreements(t, p, lib, "My::Lacks", []byte("package My::Lacks;\nsub f (List @x) Int;\nsub g (List @x) Int;\n"))
	require.Len(t, got, 1)
	assert.Contains(t, got[0], "My::Lacks::g is declared; perl defines no such sub")
}

// TestExportHelperHoldsAnImportThatDiesOnNames: an :export name is held
// to the bare `use`, so a module whose import dies when given names, as
// Test::More's reads them as a plan, does not fail it.
func TestExportHelperHoldsAnImportThatDiesOnNames(t *testing.T) {
	p := perltest.Perl542(t)
	lib := perlFixture(t, "My::Plan", `package My::Plan;
our @EXPORT = qw(ok);
sub import { shift; die "plan() doesn't understand @_" if @_; no strict 'refs'; *{caller() . '::ok'} = \&ok }
sub ok { 1 }
1;
`)

	assert.Empty(t, exportDisagreements(t, p, lib, "My::Plan", []byte("package My::Plan;\nsub ok :export;\n")))
}

// TestExportHelperRejectsAnUndeclaredTag: a tag in the module's
// %EXPORT_TAGS that installs names fails a .tpm declaring no such tag.
func TestExportHelperRejectsAnUndeclaredTag(t *testing.T) {
	p := perltest.Perl542(t)
	lib := perlFixture(t, "My::Exp", exporter)

	got := exportDisagreements(t, p, lib, "My::Exp", []byte("package My::Exp;\nsub f :export;\nsub g :export_ok;\nsub h :export_ok;\n"))
	require.NotEmpty(t, got)
	assert.Contains(t, strings.Join(got, "\n"), "use My::Exp ':gh' installs g")
}

// TestXSUtilitiesPrototypes holds the XS utility modules' prototypes to perl.
func TestXSUtilitiesPrototypes(t *testing.T) {
	holdPrototypes(t, perltest.XSUtilities...)
}

// TestXSUtilitiesExports holds the XS utility modules' export marks to
// what `use` installs, bare, by name and by tag.
func TestXSUtilitiesExports(t *testing.T) {
	holdExports(t, perltest.XSUtilities...)
}

// TestTpmSkeletonMatchesCarp: Carp's generated skeleton, with each TODO
// replaced by Carp.tpm's own return type for that sub, agrees with
// Carp.tpm on every name, mark, tag and prototype -- holdPrototypes and
// holdExports pass on it, as they do on the shipped Carp.tpm.
func TestTpmSkeletonMatchesCarp(t *testing.T) {
	skeleton := tpmSkeleton(t, "Carp")
	src, ok := declaration("Carp")
	require.True(t, ok)
	sigs, errs := DeclaredSignatures(src)
	require.Empty(t, errs)

	lines := strings.Split(strings.TrimRight(skeleton, "\n"), "\n")
	for i, line := range lines {
		if !strings.HasPrefix(line, "sub ") {
			continue
		}
		name, _, ok := strings.Cut(line[len("sub "):], " ")
		require.True(t, ok)
		cands, ok := sigs[name]
		require.True(t, ok, "Carp.tpm declares no %s", name)
		lines[i] = strings.TrimSuffix(line, "TODO;") + cands[0].Returns.String() + ";"
	}
	filled := strings.Join(lines, "\n") + "\n"

	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "declarations"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "declarations", "Carp.tpm"), []byte(filled), 0o644))
	restore := OverlayDeclarations(os.DirFS(dir))
	defer restore()

	holdPrototypes(t, "Carp")
	holdExports(t, "Carp")
}

// TestTpmSkeletonMatchesListUtilPrototypes: List::Util's generated
// skeleton, with every TODO return type replaced by a placeholder type,
// agrees with perl's prototype() for every export -- holdPrototypes
// passes on it. Carp declares no prototypes, so
// TestTpmSkeletonMatchesCarp cannot exercise this: List::Util's max is
// (@) and first is (&@).
func TestTpmSkeletonMatchesListUtilPrototypes(t *testing.T) {
	skeleton := tpmSkeleton(t, "List::Util")
	filled := strings.ReplaceAll(skeleton, "TODO;", "Any;")

	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "declarations", "List"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "declarations", "List", "Util.tpm"), []byte(filled), 0o644))
	restore := OverlayDeclarations(os.DirFS(dir))
	defer restore()

	holdPrototypes(t, "List::Util")
}

// TestTpmSkeletonMarksTags: Fcntl's SEEK_SET is export_ok only through
// its "seek" tag in %EXPORT_TAGS, so the skeleton marks it
// :export_ok(seek), the tag name the installs mechanism found it under.
func TestTpmSkeletonMarksTags(t *testing.T) {
	skeleton := tpmSkeleton(t, "Fcntl")
	assert.Contains(t, skeleton, "sub SEEK_SET :export_ok(seek) () TODO;")
}

// TestTpmSkeletonUnfilledRefused: a skeleton's TODO is not a type, so
// readLibraryDeclaration refuses it until a worker replaces every one.
func TestTpmSkeletonUnfilledRefused(t *testing.T) {
	skeleton := tpmSkeleton(t, "Carp")
	facts := readLibraryDeclaration([]byte(skeleton), nil)
	require.NotEmpty(t, facts.errs)
	var msgs []string
	for _, e := range facts.errs {
		msgs = append(msgs, e.Error())
	}
	assert.Contains(t, strings.Join(msgs, "\n"), "TODO")
}

// TestPOSIXPrototypes holds POSIX.tpm's prototypes to perl: every sub it
// declares has the prototype perl's prototype() reports, or both have
// none.
func TestPOSIXPrototypes(t *testing.T) {
	holdPrototypes(t, "POSIX")
}

// TestPOSIXExports holds POSIX.tpm's :export and :export_ok marks to
// what `use POSIX` installs in a caller, bare, by name and by tag.
func TestPOSIXExports(t *testing.T) {
	holdExports(t, "POSIX")
}

// TestPOSIXBareUseImportsSIGRT: measured on 5.42.0, a bare `use POSIX`
// installs the tied variable %SIGRT along with its 586 subs.
func TestPOSIXBareUseImportsSIGRT(t *testing.T) {
	got := importedFrom(t, "use POSIX;\n")
	assert.Contains(t, got, "%SIGRT")
}

// TestPOSIXExportsAreNotVacuous: the comparison holdExports makes is
// against a real measurement, not an empty one -- POSIX builds its
// export lists inside its own import, so a driver that reads them before
// any use has run would see 0/0/0 and never catch a missing mark.
func TestPOSIXExportsAreNotVacuous(t *testing.T) {
	got := importedFrom(t, "use POSIX;\n")
	assert.GreaterOrEqual(t, len(got), 586, "a bare use POSIX installs 586 subs and %%SIGRT, counting AUTOLOAD names exists finds that defined would miss")
}

// TestPOSIXEmptyImportListImportsNothing: `use POSIX ();` calls no
// import, so a bare call to floor in the same file does not resolve to
// POSIX's declaration.
func TestPOSIXEmptyImportListImportsNothing(t *testing.T) {
	got := importedFrom(t, "use POSIX ();\n")
	assert.Empty(t, got)
}

// TestPOSIXImportKeepsPrintfFilehandleParse: CORE::printf has no
// prototype and keeps its own filehandle parse; POSIX marks its own
// printf :export_ok, so it is not in scope after a bare `use POSIX`,
// and `printf STDOUT "x"` still reads STDOUT as the handle rather than
// an Unknown construct.
func TestPOSIXImportKeepsPrintfFilehandleParse(t *testing.T) {
	root := ParseWithLoader([]byte("use POSIX;\nprintf STDOUT \"x\";\n"), func(string) ([]byte, bool) { return nil, false })
	if errs := DeclarationErrors(root); len(errs) > 0 {
		t.Fatalf("declaration errors: %v", errs)
	}
	if containsUnknown(root) {
		t.Errorf("printf STDOUT \"x\" after use POSIX did not parse cleanly: %v", root)
	}
	if imports := Imports(root); func() bool {
		_, ok := imports["printf"]
		return ok
	}() {
		t.Errorf("use POSIX imports printf; it is marked :export_ok and must be asked for")
	}
}

// TestPOSIXImportMatchesPerl: POSIX declares no `import` of its own.
// Its own import (POSIX.pm:174) rewrites a legacy bareword tag ("foo_h")
// to ":foo_h" and dies on an unimplemented name before it ever reaches
// Exporter, but every other case -- a plain named list, a `:tag`, and
// the explicit empty list -- forwards unchanged to Exporter::import, so
// the plain :export/:export_ok marks this parser already reads get them
// right without a declared import (measured on 5.42.0; audit E, S2).
// The declared-import mechanism (RFC 0001's one altering key) has
// nothing to say about a per-item bareword rewrite, so the legacy
// "qw(math_h)" form (without the colon) is not expressed here: it
// installs nothing under this parser where perl installs the whole
// tag, a known gap for perigrin, not fixed by this issue.
func TestPOSIXImportMatchesPerl(t *testing.T) {
	assert.Equal(t, []string{"floor"}, importedFrom(t, "use POSIX qw(floor);\n"))

	got := importedFrom(t, "use POSIX qw(:math_h);\n")
	assert.Contains(t, got, "floor")
	assert.Contains(t, got, "tan")
	assert.NotContains(t, got, "times")

	assert.Empty(t, importedFrom(t, "use POSIX ();\n"))
}

// TestConstantExportersPrototypes holds Fcntl.tpm's and Socket.tpm's
// prototypes to perl.
func TestConstantExportersPrototypes(t *testing.T) {
	holdPrototypes(t, "Fcntl", "Socket")
}

// TestConstantExportersExports holds Fcntl.tpm's and Socket.tpm's
// :export and :export_ok marks to what `use` installs in a caller.
func TestConstantExportersExports(t *testing.T) {
	holdExports(t, "Fcntl", "Socket")
}

// TestSocketBareUseOmitsCRLFVariable: $CRLF is in Socket's @EXPORT_OK
// only (measured), so a bare `use Socket` does not import it.
func TestSocketBareUseOmitsCRLFVariable(t *testing.T) {
	got := importedFrom(t, "use Socket;\n")
	assert.NotContains(t, got, "$CRLF")
}

// TestFcntlBareUseOmitsFlockConstants: LOCK_EX is in Fcntl's
// @EXPORT_OK only (measured), so a bare `use Fcntl` does not import it.
func TestFcntlBareUseOmitsFlockConstants(t *testing.T) {
	got := importedFrom(t, "use Fcntl;\n")
	assert.NotContains(t, got, "LOCK_EX")
}

// containsUnknown is whether n or any descendant is an Unknown node.
func containsUnknown(n *Node) bool {
	if n.Refused() {
		return true
	}
	for _, c := range n.Children {
		if containsUnknown(c) {
			return true
		}
	}
	return false
}

// TestPurePerlExportersPrototypes holds each pure-Perl exporter's .tpm
// prototypes to perl (perltest.PureExporters: 01a1464e-db4c).
func TestPurePerlExportersPrototypes(t *testing.T) {
	holdPurePerlExporters(t, prototypeDisagreements)
}

// TestPurePerlExportersExports holds each pure-Perl exporter's .tpm
// :export and :export_ok marks to what `use` installs.
func TestPurePerlExportersExports(t *testing.T) {
	holdPurePerlExporters(t, exportDisagreements)
}

// TestErrnoDeclaresNoExportedVariable: Errno.tpm marks no `our`
// variable -- measured on 5.42.0, %! is in none of Errno's export
// lists, so every name it exports is a constant sub.
func TestErrnoDeclaresNoExportedVariable(t *testing.T) {
	src, ok := declaration("Errno")
	require.True(t, ok)
	facts, err := readHeld("Errno", src)
	require.NoError(t, err)
	assert.Empty(t, facts.variables, "Errno.tpm declares an our variable; perl's use Errno installs none")
}

// TestXSLibrariesPrototypes holds each XS library's .tpm prototypes to
// perl (perltest.XSLibraries: 01a1464e-dbd3).
func TestXSLibrariesPrototypes(t *testing.T) {
	holdFiltered(t, prototypeDisagreements, perltest.XSLibraries, heldExempt())
}

// TestXSLibrariesExports holds each XS library's .tpm :export and
// :export_ok marks to what `use` installs.
func TestXSLibrariesExports(t *testing.T) {
	holdFiltered(t, exportDisagreements, perltest.XSLibraries, heldExempt())
}

// TestTimeHiResBareUseKeepsCoreTime: measured on 5.42.0, Time::HiRes's
// @EXPORT is empty, so a bare `use Time::HiRes;` installs nothing and a
// later bare `time()` still resolves to CORE's.
func TestTimeHiResBareUseKeepsCoreTime(t *testing.T) {
	got := importedFrom(t, "use Time::HiRes;\n")
	assert.Empty(t, got)
}

// TestOOModulesPrototypes holds each OO module's .tpm prototypes to
// perl (perltest.OOModules: 01a1464e-dd27), methods included:
// held_to_perl.pl's prototypes mode falls back to $module->can($name),
// the way perl itself finds an inherited method.
func TestOOModulesPrototypes(t *testing.T) {
	holdPrototypes(t, perltest.OOModules...)
}

// TestOOModulesExports holds each OO module's .tpm :export and
// :export_ok marks to what `use` installs.
func TestOOModulesExports(t *testing.T) {
	holdExports(t, perltest.OOModules...)
}

// TestFileSpecBareUseImportsNothing: measured on 5.42.0, File::Spec's
// @EXPORT and @EXPORT_OK are both empty, so a bare `use File::Spec;`
// installs nothing and a bare `catfile(...)` does not resolve to its
// method.
func TestFileSpecBareUseImportsNothing(t *testing.T) {
	assert.Empty(t, importedFrom(t, "use File::Spec;\n"))
}

// TestIOSocketIPBareUseInstallsSocketNames: measured on 5.42.0, a bare
// `use IO::Socket::IP;` installs exactly the 170 names Socket's own
// bare `use Socket;` does (IO::Socket::IP inherits IO::Socket's import,
// which forwards to Socket's Exporter), so a bare `inet_aton(...)`
// resolves.
func TestIOSocketIPBareUseInstallsSocketNames(t *testing.T) {
	got := importedFrom(t, "use IO::Socket::IP;\n")
	assert.Len(t, got, 170)
	assert.Contains(t, got, "inet_aton")
	assert.Contains(t, got, "AF_INET")
}

// TestIOSocketIPRegisterInstallsSocketNames: IO::Socket::IP's own import
// (IP.pm:122-139) reads only its "-register" option (registering
// IO::Socket::IP::_ForINET as the AF_INET/AF_INET6 domain handler) and
// forwards the rest, with -register filtered out, to IO::Socket's own
// import -- measured on 5.42.0, `use IO::Socket::IP -register;`
// installs the same 170 names a bare `use IO::Socket::IP;` does, since
// -register leaves no symbols behind to forward.
func TestIOSocketIPRegisterInstallsSocketNames(t *testing.T) {
	got := importedFrom(t, "use IO::Socket::IP -register;\n")
	assert.Len(t, got, 170)
	assert.Contains(t, got, "inet_aton")
	assert.Contains(t, got, "AF_INET")
}

// TestLocaleMaketextSimpleBareUseInstallsLoc: measured on 5.42.0,
// Locale::Maketext::Simple's own import defaults Export to 'loc' and
// installs "${Export}_lang" beside it, so a bare `use
// Locale::Maketext::Simple;` installs both loc and loc_lang.
func TestLocaleMaketextSimpleBareUseInstallsLoc(t *testing.T) {
	got := importedFrom(t, "use Locale::Maketext::Simple;\n")
	assert.ElementsMatch(t, []string{"loc", "loc_lang"}, got)
}

// TestLocaleMaketextSimpleEmptyImportInstallsNothing: measured on
// 5.42.0, `use Locale::Maketext::Simple ();` installs neither loc nor
// loc_lang.
func TestLocaleMaketextSimpleEmptyImportInstallsNothing(t *testing.T) {
	assert.Empty(t, importedFrom(t, "use Locale::Maketext::Simple ();\n"))
}

// TestLocaleMaketextSimpleArgumentFormInstallsLoc: Locale::Maketext::Simple
// declares its own `import` (perigrin's ruling, 2026-10-20: "the parameters
// are the signature to import()"), so an option list is import's
// arguments, not names to resolve against :export -- measured on 5.42.0,
// `use Locale::Maketext::Simple (Style => 'gettext');` still installs loc
// and loc_lang, exactly as a bare use does.
func TestLocaleMaketextSimpleArgumentFormInstallsLoc(t *testing.T) {
	got := importedFrom(t, "use Locale::Maketext::Simple (Style => 'gettext');\n")
	assert.ElementsMatch(t, []string{"loc", "loc_lang"}, got)
}

// TestLocaleMaketextSimpleExportKeyClaimsNoNames: perigrin's ruling,
// 2026-10-20 ("the signature names the keys"): Locale::Maketext::Simple's
// import signature names Export as the key whose presence changes what
// import installs -- measured on 5.42.0, `use Locale::Maketext::Simple
// (Export => 'zz');` installs zz and zz_lang, not loc and loc_lang, so
// this parser claims no names rather than a wrong one.
func TestLocaleMaketextSimpleExportKeyClaimsNoNames(t *testing.T) {
	assert.Empty(t, importedFrom(t, "use Locale::Maketext::Simple (Export => 'zz');\n"))
}

// TestVariableExportersPrototypes holds each variable exporter's .tpm
// prototypes to perl (perltest.VariableExporters: 01a1464e-dca3).
func TestVariableExportersPrototypes(t *testing.T) {
	holdPrototypes(t, perltest.VariableExporters...)
}

// TestVariableExportersExports holds each variable exporter's .tpm
// :export and :export_ok marks to what `use` installs.
func TestVariableExportersExports(t *testing.T) {
	holdExports(t, perltest.VariableExporters...)
}

// TestConfigBareUseImportsOnlyConfigHash: a bare `use Config;` installs
// %Config and none of Config's @EXPORT_OK subs (ruling 1).
func TestConfigBareUseImportsOnlyConfigHash(t *testing.T) {
	got := importedFrom(t, "use Config;\n")
	assert.Equal(t, []string{"%Config"}, got)
}

// TestGetoptLongBareUseOmitsConfigure: Configure is in Getopt::Long's
// @EXPORT_OK alone (measured), so a bare `use Getopt::Long;` does not
// import it.
func TestGetoptLongBareUseOmitsConfigure(t *testing.T) {
	got := importedFrom(t, "use Getopt::Long;\n")
	assert.NotContains(t, got, "Configure")
	assert.Contains(t, got, "GetOptions")
}

// TestGetoptLongConfigInstallsGetOptions: Getopt::Long's own import
// moves everything from ':config' onward to Configure, so `@syms`, what
// is left, is empty, and it falls through to SUPER::import's default
// @EXPORT (perigrin's ruling, 2026-10-20: "the parameters are the
// signature to import()") -- measured on 5.42.0, `use Getopt::Long
// qw(:config bundling);` installs GetOptions (audit E, finding 9).
func TestGetoptLongConfigInstallsGetOptions(t *testing.T) {
	got := importedFrom(t, "use Getopt::Long qw(:config bundling);\n")
	assert.Contains(t, got, "GetOptions")
}

// TestGetoptLongFromArrayInstallsGetOptions: measured on 5.42.0, `use
// Getopt::Long qw(GetOptionsFromArray);` also installs GetOptions --
// Getopt::Long's own import always adds &GetOptions once @syms is
// non-empty (audit E, finding 9). The declared-import mechanism (RFC
// 0001) has no key for this: its only two answers to a named list are
// "install the full :export set" or "the names are unknown", so this
// also installs the three order variables, which perl's own import
// does not for this list. That over-claim is tracked in issue
// 01a15baf; this guards only that GetOptions still resolves.
func TestGetoptLongFromArrayInstallsGetOptions(t *testing.T) {
	got := importedFrom(t, "use Getopt::Long qw(GetOptionsFromArray);\n")
	assert.Contains(t, got, "GetOptions")
	assert.Contains(t, got, "GetOptionsFromArray")
}

// TestFileStatEmptyImportKeepsCoreStat: `use File::stat ();` is the
// explicit "load but import nothing" form (RFC 0001), so a later bare
// `stat($f)` still resolves to CORE's.
func TestFileStatEmptyImportKeepsCoreStat(t *testing.T) {
	got := importedFrom(t, "use File::stat ();\n")
	assert.Empty(t, got)
}

// Exporter, XSLoader and DynaLoader are the infrastructure-module batch
// (01a1464e-dd6c, -ddb0, -dde3): the only three of the 50 core modules
// RFC 0001 permits to ship unshipped (perigrin's ruling 4, 2026-10-16) if
// measurement shows a declaration cannot keep every module that uses it
// resolving as before. Each is held to perl, and its exports measured,
// separately: a shared batch name would not match the issues' own
// criteria commands.

// TestExporterPrototypes holds Exporter.tpm's prototypes to perl.
func TestExporterPrototypes(t *testing.T) {
	holdPrototypes(t, "Exporter")
}

// TestExporterExports holds Exporter.tpm's :export and :export_ok marks
// to what `use Exporter` installs in a caller.
func TestExporterExports(t *testing.T) {
	holdExports(t, "Exporter")
}

// TestExporterBareUseDoesNotImportImport: measured on 5.42.0, a bare
// `use Exporter;` installs nothing -- Exporter carries no @EXPORT of its
// own -- while `use Exporter 'import';` installs import, the one name
// its own import() special-cases.
func TestExporterBareUseDoesNotImportImport(t *testing.T) {
	assert.Empty(t, importedFrom(t, "use Exporter;\n"))
	got := importedFrom(t, "use Exporter 'import';\n")
	assert.Equal(t, []string{"import"}, got)
}

// TestXSLoaderPrototypes holds XSLoader.tpm's prototypes to perl.
func TestXSLoaderPrototypes(t *testing.T) {
	holdPrototypes(t, "XSLoader")
}

// TestXSLoaderExports holds XSLoader.tpm's :export and :export_ok marks
// to what `use XSLoader` installs in a caller.
func TestXSLoaderExports(t *testing.T) {
	holdExports(t, "XSLoader")
}

// TestXSLoaderBareUseImportsNothing: measured on 5.42.0, XSLoader never
// calls Exporter, so neither a bare `use XSLoader;` nor one naming load
// installs anything -- a later bare `load()` still fails to resolve to
// XSLoader::load.
func TestXSLoaderBareUseImportsNothing(t *testing.T) {
	assert.Empty(t, importedFrom(t, "use XSLoader;\n"))
	assert.Empty(t, importedFrom(t, "use XSLoader qw(load);\n"))
}

// TestDynaLoaderPrototypes holds DynaLoader.tpm's prototypes to perl.
func TestDynaLoaderPrototypes(t *testing.T) {
	holdPrototypes(t, "DynaLoader")
}

// TestDynaLoaderExports holds DynaLoader.tpm's :export and :export_ok
// marks to what `use DynaLoader` installs in a caller.
func TestDynaLoaderExports(t *testing.T) {
	holdExports(t, "DynaLoader")
}

// TestDynaLoaderBareUseImportsNothing: measured on 5.42.0, DynaLoader
// never calls Exporter either, so no `use DynaLoader` form installs a name;
// every caller reaches it through @ISA or a qualified call.
func TestDynaLoaderBareUseImportsNothing(t *testing.T) {
	assert.Empty(t, importedFrom(t, "use DynaLoader;\n"))
	assert.Empty(t, importedFrom(t, "use DynaLoader qw(bootstrap);\n"))
}

// TestTestMoreArgumentFormImportsSameNames: Test::More's import is
// Test::Builder::Module's (measured), which reads an import list with
// no `import => [...]` pair as plan() arguments, not names --
// `use Test::More tests => 3;` installs the same names a bare use
// does (measured on 5.42.0: both leave every one of its 26 subs and
// $TODO defined in the caller). The forward risk this issue settles
// (dadb's gate): without facts.builder, this list reads as computed
// and opaque instead.
func TestTestMoreArgumentFormImportsSameNames(t *testing.T) {
	bare := importedFrom(t, "use Test::More;\n")
	withArgs := importedFrom(t, "use Test::More tests => 3;\n")
	assert.Equal(t, bare, withArgs)
	assert.Contains(t, withArgs, "ok")
	assert.Contains(t, withArgs, "plan")
}

// TestTestMoreImportKeyClaimsNoNames: perigrin's ruling, 2026-10-20
// ("the signature names the keys"): Test::More's import signature names
// import as the key whose presence changes what import installs --
// measured on 5.42.0, `use Test::More import => ['ok'], tests => 1;`
// installs only ok, not the full bare-use set, so this parser claims no
// names rather than a wrong one.
func TestTestMoreImportKeyClaimsNoNames(t *testing.T) {
	assert.Empty(t, importedFrom(t, "use Test::More import => ['ok'], tests => 1;\n"))
}

// TestTestMoreImportParamAdmitsArrayRef: pass 3's gate (perigrin,
// 2026-10-20): the parameter's name carries the altering key, but its
// type must still be what perl passes there -- measured on 5.42.0,
// Test::Builder::Module::_strip_imports dereferences the value paired
// with the import key as an array (`@{ $list->[$idx+1] }`), so it is an
// ArrayRef, not a Str.
func TestTestMoreImportParamAdmitsArrayRef(t *testing.T) {
	facts := resolveLibrary(t, "Test::More")
	sigs := facts.signatures["import"]
	if len(sigs) != 1 || len(sigs[0].Params) == 0 {
		t.Fatalf("Test::More::import signature: %+v, want one signature with a parameter named import", sigs)
	}
	got := sigs[0].Params[0]
	if got.Name != "import" || got.Type != types.ArrayRef {
		t.Errorf("Test::More::import's import parameter: %+v, want ArrayRef", got)
	}
}

// TestIOExportersPrototypes holds each module's .tpm prototypes to perl
// (perltest.IOExporters: 01a1464e-db76).
func TestIOExportersPrototypes(t *testing.T) {
	holdFiltered(t, prototypeDisagreements, perltest.IOExporters, heldExempt())
}

// TestIOExportersExports holds each module's .tpm :export and
// :export_ok marks to what `use` installs.
func TestIOExportersExports(t *testing.T) {
	holdExports(t, perltest.IOExporters...)
}

// ioExportersMethodsOnly are the subs declared in the batch's four
// class .tpm files that are never exported: callers reach each only
// through `->`, inherited from the class that defines it, regardless
// of import. Measured on 5.42.0: none is in its module's @EXPORT or
// @EXPORT_OK.
var ioExportersMethodsOnly = map[string][]string{
	"IO::Handle": {"new", "new_from_fd", "opened", "fileno", "eof", "getc", "close"},
	"IO::File":   {"new", "new_tmpfile", "open"},
	"IO::Socket": {"new", "socket", "connect", "bind", "listen", "accept", "send", "recv", "close"},
	"FileHandle": {"new_from_fd", "close", "fileno", "getc", "eof", "flush", "new", "new_tmpfile", "open", "seek", "tell", "getpos", "setpos"},
}

// TestIOExportersMethodsAreNotExported: a method declared beside a
// class .tpm's exports -- IO::Handle's close, IO::File's new, and the
// rest ioExportersMethodsOnly names -- carries no :export or
// :export_ok mark, since perl's `use` of that module never installs
// it; it is reachable only through `->`.
func TestIOExportersMethodsAreNotExported(t *testing.T) {
	for module, methods := range ioExportersMethodsOnly {
		src, ok := declaration(module)
		require.True(t, ok, "%s: no declaration", module)
		facts, err := readHeld(module, src)
		require.NoError(t, err)
		for _, name := range methods {
			_, declared := facts.signatures[name]
			require.True(t, declared, "%s: %s is not declared", module, name)
			if slices.Contains(facts.exports, name) {
				t.Errorf("%s::%s carries an export mark, but perl's use %s installs no such name", module, name, module)
			}
		}
	}
}

// documentedOOInterfaces is, per module, the documented OO methods
// audit-C.md's findings named as entirely omitted before this fix:
// File::Temp's OO interface (finding 14) and three not-exported
// functions (finding 15); Archive::Tar's whole method interface
// (finding 18); Term::ReadLine's new() IN/OUT parameters and four
// omitted methods (findings 31, 32); HTTP::Tiny's generated attribute
// accessors (finding 29). IO::Select's class-method select and
// Net::Ping's broader OO interface (findings 13, 22) are deferred --
// see IO/Select.tpm's and Net/Ping.tpm's own comments -- so neither
// is named here.
var documentedOOInterfaces = map[string][]string{
	"File::Temp": {
		"new", "newdir", "filename", "unlink_on_destroy", "safe_level", "top_system_uid",
		"tempnam", "cmpstat", "unlink1",
	},
	"Archive::Tar": {
		"new", "create_archive", "iter", "list_archive", "extract_archive",
		"has_io_string", "has_perlio", "has_zlib_support", "has_bzip2_support",
		"has_xz_support", "can_handle_compressed_files",
		"read", "contains_file", "extract", "extract_file", "list_files",
		"get_files", "get_content", "replace_content", "rename", "chmod",
		"chown", "remove", "clear", "write", "add_files", "add_data", "error",
		"setcwd",
	},
	"Term::ReadLine": {"IN", "OUT", "ornaments", "newTTY", "tkRunning", "event_loop"},
	"HTTP::Tiny": {
		"cookie_jar", "default_headers", "http_proxy", "https_proxy", "keep_alive",
		"local_address", "max_redirect", "max_size", "proxy", "no_proxy",
		"SSL_options", "verify_SSL",
	},
}

// TestDocumentedOOInterfacesDeclared: every method documentedOOInterfaces
// names is declared, with a typed return, in its module's .tpm --
// audit-C.md's findings for the OO interfaces this fix adds (14, 15,
// 18, 29, 31, 32).
func TestDocumentedOOInterfacesDeclared(t *testing.T) {
	for module, methods := range documentedOOInterfaces {
		src, ok := declaration(module)
		if !assert.True(t, ok, "%s: no declaration", module) {
			continue
		}
		facts, err := readHeld(module, src)
		if !assert.NoError(t, err, "%s: %v", module, err) {
			continue
		}
		for _, name := range methods {
			sigs, declared := facts.signatures[name]
			if !assert.True(t, declared, "%s::%s is not declared", module, name) {
				continue
			}
			for _, s := range sigs {
				assert.NotEqual(t, types.Unknown, s.Returns, "%s::%s should be typed, not Unknown", module, name)
			}
		}
	}
}

// TestIOSocketDocumentedMethodsDeclared: IO/Socket.pm documents (or, for
// configure and register_domain, defines without a POD =head2) these
// methods beside the exports (IO/Socket.pm:56-395), each measured on
// 5.42.0 (01a15ac0, IO::Socket finding 7); a prior pass of IO/Socket.tpm
// left every one of them out.
func TestIOSocketDocumentedMethodsDeclared(t *testing.T) {
	methods := []string{
		"configure", "socketpair", "blocking", "sockname", "peername",
		"connected", "shutdown", "setsockopt", "getsockopt", "sockopt",
		"atmark", "timeout", "sockdomain", "socktype", "protocol",
		"register_domain",
	}
	src, ok := declaration("IO::Socket")
	require.True(t, ok, "IO::Socket: no declaration")
	facts, err := readHeld("IO::Socket", src)
	require.NoError(t, err)
	for _, name := range methods {
		_, declared := facts.signatures[name]
		assert.True(t, declared, "IO::Socket::%s is not declared", name)
	}
}

// TestFileHandleFilenoIsDeclared: FileHandle.pm aliases IO::Handle's
// fileno straight into FileHandle:: at require time, so FileHandle.tpm
// must declare it with a concrete return type -- a name the parser
// leaves out of a declaration reads as Unknown to inference, the same
// as a method no .tpm ever mentions (01a1464e-db76 review finding 1).
func TestFileHandleFilenoIsDeclared(t *testing.T) {
	src, ok := declaration("FileHandle")
	require.True(t, ok, "FileHandle: no declaration")
	facts, err := readHeld("FileHandle", src)
	require.NoError(t, err)
	sigs, declared := facts.signatures["fileno"]
	require.True(t, declared, "FileHandle::fileno is not declared")
	require.NotEmpty(t, sigs)
	for _, s := range sigs {
		assert.NotEqual(t, types.Unknown, s.Returns, "FileHandle::fileno should be typed, not Unknown")
	}
}
