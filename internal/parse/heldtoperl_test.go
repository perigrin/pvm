// ABOUTME: Holds a .tpm's prototypes and export marks to perl itself: each declaration batch calls
// ABOUTME: holdPrototypes and holdExports with its modules. Fixture modules test both helpers.

package parse

import (
	"fmt"
	"maps"
	"os"
	"os/exec"
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
	cmd := exec.Command(p.Path, argv...)
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
// @EXPORT and @EXPORT_OK, an `&name` there being the sub name.
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
		case slices.Contains(facts.exports, list) && !slices.Contains(installs, list):
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
