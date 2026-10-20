// ABOUTME: Holds a .tpm's prototypes and export marks to perl itself: each declaration batch calls
// ABOUTME: holdPrototypes and holdExports with its modules. Fixture modules test both helpers.

package parse

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tamarou.com/pvm/internal/perltest"
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
	p := perltest.Perl542(t)
	for _, module := range modules {
		src, ok := declaration(module)
		if !ok {
			t.Errorf("%s: no declaration", module)
			continue
		}
		for _, d := range check(t, p, "", module, src) {
			t.Error(d)
		}
	}
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
	path := filepath.Join(lib, filepath.FromSlash(strings.ReplaceAll(module, "::", "/"))+".pm")
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(pm), 0o644))
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
	assert.True(t, strings.HasPrefix(path, p.Archlib), "%s is not in archlib %s", path, p.Archlib)
	assert.FileExists(t, path)
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

// TestHeldToPerlIgnoresTheCallersPerlLib: perl finds a module as an
// installed perl does, whatever PERL5LIB, PERLLIB or PERL5OPT the test
// runs with: a Carp there is not the one measured.
func TestHeldToPerlIgnoresTheCallersPerlLib(t *testing.T) {
	p := perltest.Perl542(t)
	fake := perlFixture(t, "Carp", "package Carp;\nour @EXPORT = qw(fake);\nsub import { }\n1;\n")
	src, ok := declaration("Carp")
	require.True(t, ok)
	for _, env := range [][2]string{{"PERL5LIB", fake}, {"PERLLIB", fake}, {"PERL5OPT", "-I" + fake}} {
		t.Run(env[0], func(t *testing.T) {
			t.Setenv(env[0], env[1])
			path, err := modulePath(t, p, "", "Carp")
			require.NoError(t, err)
			assert.False(t, strings.HasPrefix(path, fake), "perl measured %s", path)
			assert.Empty(t, exportDisagreements(t, p, "", "Carp", src))
		})
	}
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
	if n.Kind == Unknown {
		return true
	}
	for _, c := range n.Children {
		if containsUnknown(c) {
			return true
		}
	}
	return false
}

// pureExporters is the pure-Perl exporters batch (01a1464e-db4c): every
// module whose .tpm declares the names a `use` of it installs. Symbol is
// held back (see docs/rfcs/0001-type-libraries.md and the issue's own
// report): its qualify, declared accurately as Str|GlobRef, makes
// IO/Handle.pm's `tell qualify($_[0], caller) if ref($_[0])` a new
// TestPSCCheckLibBaseline diagnostic that perl -w does not confirm, since
// $_[0] there is always a ref; perigrin must settle how to admit it.
var pureExporters = []string{
	"File::Basename", "File::Find", "File::Path", "Errno",
	"Text::ParseWords", "Class::Struct", "File::Copy", "Time::Local", "Getopt::Std",
}

// TestPurePerlExportersPrototypes holds each pure-Perl exporter's .tpm
// prototypes to perl.
func TestPurePerlExportersPrototypes(t *testing.T) {
	holdPrototypes(t, pureExporters...)
}

// TestPurePerlExportersExports holds each pure-Perl exporter's .tpm
// :export and :export_ok marks to what `use` installs.
func TestPurePerlExportersExports(t *testing.T) {
	holdExports(t, pureExporters...)
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

// xsLibraries is the XS-library batch (01a1464e-dbd3): Encode,
// Time::HiRes, Compress::Raw::Zlib and SDBM_File.
var xsLibraries = []string{"Encode", "Time::HiRes", "Compress::Raw::Zlib", "SDBM_File"}

// TestXSLibrariesPrototypes holds each XS library's .tpm prototypes to
// perl.
func TestXSLibrariesPrototypes(t *testing.T) {
	holdPrototypes(t, xsLibraries...)
}

// TestXSLibrariesExports holds each XS library's .tpm :export and
// :export_ok marks to what `use` installs.
func TestXSLibrariesExports(t *testing.T) {
	holdExports(t, xsLibraries...)
}

// TestTimeHiResBareUseKeepsCoreTime: measured on 5.42.0, Time::HiRes's
// @EXPORT is empty, so a bare `use Time::HiRes;` installs nothing and a
// later bare `time()` still resolves to CORE's.
func TestTimeHiResBareUseKeepsCoreTime(t *testing.T) {
	got := importedFrom(t, "use Time::HiRes;\n")
	assert.Empty(t, got)
}

// ooModules is the OO-only batch (01a1464e-dd27): each declares methods,
// not exports, since measured on 5.42.0 a bare `use` of every one but
// IO::Socket::IP and Locale::Maketext::Simple installs nothing.
var ooModules = []string{
	"File::Spec", "File::Spec::Unix", "IO::Socket::IP", "IO::Select",
	"Term::ReadLine", "Tie::Hash", "HTTP::Tiny", "Locale::Maketext::Simple",
}

// TestOOModulesPrototypes holds each OO module's .tpm prototypes to
// perl, methods included: held_to_perl.pl's prototypes mode falls back
// to $module->can($name), the way perl itself finds an inherited
// method.
func TestOOModulesPrototypes(t *testing.T) {
	holdPrototypes(t, ooModules...)
}

// TestOOModulesExports holds each OO module's .tpm :export and
// :export_ok marks to what `use` installs.
func TestOOModulesExports(t *testing.T) {
	holdExports(t, ooModules...)
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

// variableExporters is the variable-exporters batch (01a1464e-dca3):
// every module below exports a plain Perl variable, measured on
// 2026-10-16.
var variableExporters = []string{
	"Config", "Getopt::Long", "Test::More", "Text::Wrap", "ExtUtils::MakeMaker", "File::stat",
}

// TestVariableExportersPrototypes holds each variable exporter's .tpm
// prototypes to perl.
func TestVariableExportersPrototypes(t *testing.T) {
	holdPrototypes(t, variableExporters...)
}

// TestVariableExportersExports holds each variable exporter's .tpm
// :export and :export_ok marks to what `use` installs.
func TestVariableExportersExports(t *testing.T) {
	holdExports(t, variableExporters...)
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

// TestFileStatEmptyImportKeepsCoreStat: `use File::stat ();` is the
// explicit "load but import nothing" form (RFC 0001), so a later bare
// `stat($f)` still resolves to CORE's.
func TestFileStatEmptyImportKeepsCoreStat(t *testing.T) {
	got := importedFrom(t, "use File::stat ();\n")
	assert.Empty(t, got)
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

// ioExporters is the IO and file-handling batch (01a1464e-db76):
// FileHandle, IO::Handle, IO::File, IO::Socket and the five plain
// exporters File::Spec::Functions, File::Temp, Archive::Tar,
// ExtUtils::Manifest and Net::Ping.
var ioExporters = []string{
	"FileHandle", "IO::Handle", "IO::File", "IO::Socket",
	"File::Spec::Functions", "File::Temp", "Archive::Tar", "ExtUtils::Manifest", "Net::Ping",
}

// TestIOExportersPrototypes holds each module's .tpm prototypes to perl.
func TestIOExportersPrototypes(t *testing.T) {
	holdPrototypes(t, ioExporters...)
}

// TestIOExportersExports holds each module's .tpm :export and
// :export_ok marks to what `use` installs.
func TestIOExportersExports(t *testing.T) {
	holdExports(t, ioExporters...)
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
