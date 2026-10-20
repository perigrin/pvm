// ABOUTME: Exporter, XSLoader and DynaLoader are read by most core modules, so each must keep
// ABOUTME: every privlib file that uses it resolving identically with its .tpm in place or masked out.
package parse

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tamarou.com/pvm/internal/perltest"
)

// blockingFS reads every file from base except blocked, which it reports
// missing -- the resolver's state before an infrastructure module's .tpm
// existed, so a file naming that module falls back to reading its real
// source instead.
type blockingFS struct {
	base    fs.FS
	blocked string
}

func (b blockingFS) Open(name string) (fs.File, error) {
	if name == b.blocked {
		return nil, fs.ErrNotExist
	}
	return b.base.Open(name)
}

// privlibFilesNaming returns every .pm and .pl file under privlib whose
// source names module as a whole word, sorted: the candidates a `use`,
// `require`, `@ISA`, `use base` or `use parent` of module could appear in.
func privlibFilesNaming(t *testing.T, privlib, module string) []string {
	t.Helper()
	pattern := regexp.MustCompile(`\b` + regexp.QuoteMeta(module) + `\b`)
	var files []string
	err := filepath.WalkDir(privlib, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		ext := filepath.Ext(p)
		if ext != ".pm" && ext != ".pl" {
			return nil
		}
		content, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		if pattern.Match(content) {
			files = append(files, p)
		}
		return nil
	})
	require.NoError(t, err)
	sort.Strings(files)
	return files
}

// importsDiffer reports whether after changed what before already knew:
// its name, prototype, prototype knowledge or local/lvalue marks, or a
// signature before already carried. A name gaining a signature it had
// none of before -- declaring the infrastructure module's own sub adds
// exactly this -- is the point of the .tpm, not a regression, so that
// alone is not a difference.
func importsDiffer(before, after Import) bool {
	if before.Name != after.Name || before.Prototype != after.Prototype ||
		before.PrototypeKnown != after.PrototypeKnown ||
		before.Local != after.Local || before.Lvalue != after.Lvalue {
		return true
	}
	return len(before.Signatures) > 0 && !reflect.DeepEqual(before.Signatures, after.Signatures)
}

// keepsResolution parses every privlib file naming module twice: once
// with its .tpm masked out of declarations (the resolver's state before
// this issue) and once with it in place, sharing module reads within each
// pass through a Session. compared is how many files parsed cleanly both
// ways; diffs names each whose resolved imports -- names, prototypes,
// prototype knowledge and any signature already known -- differ between
// the two; a name's signature going from unknown to known is not a
// difference (see importsDiffer).
func keepsResolution(t *testing.T, module, tpmPath string) (compared int, diffs []string) {
	t.Helper()
	p := perltest.Perl542(t)
	files := privlibFilesNaming(t, p.Privlib, module)
	roots := []string{p.Privlib, p.Archlib}

	saved := declarations
	declarations = blockingFS{base: saved, blocked: path.Join("declarations", tpmPath)}
	before := map[string]map[string]Import{}
	sBefore := NewSession()
	for _, f := range files {
		if n, err := sBefore.ParseFileFrom(f, roots...); err == nil {
			before[f] = Imports(n)
		}
	}
	declarations = saved

	sAfter := NewSession()
	for _, f := range files {
		n, err := sAfter.ParseFileFrom(f, roots...)
		if err != nil {
			continue
		}
		beforeImports, ok := before[f]
		if !ok {
			continue
		}
		compared++
		afterImports := Imports(n)
		if len(beforeImports) != len(afterImports) {
			diffs = append(diffs, fmt.Sprintf("%s: %d imports before %s's .tpm, %d after: before %#v, after %#v",
				relTo(p.Privlib, f), len(beforeImports), module, len(afterImports), beforeImports, afterImports))
			continue
		}
		for name, b := range beforeImports {
			a, ok := afterImports[name]
			if !ok || importsDiffer(b, a) {
				diffs = append(diffs, fmt.Sprintf("%s: %s differs with %s's .tpm in place: before %#v, after %#v",
					relTo(p.Privlib, f), name, module, b, a))
			}
		}
	}
	sort.Strings(diffs)
	return compared, diffs
}

// relTo is file relative to root, or file itself if that fails.
func relTo(root, file string) string {
	rel, err := filepath.Rel(root, file)
	if err != nil {
		return file
	}
	return rel
}

// namesIn reports whether a privlib file's source matches pattern, for the
// not-vacuous checks: the comparison must exercise each form a module
// reaches an infrastructure module by, not only the commonest one.
func namesIn(t *testing.T, file string, pattern *regexp.Regexp) bool {
	t.Helper()
	content, err := os.ReadFile(file)
	require.NoError(t, err)
	return pattern.Match(content)
}

var (
	useBaseExporter   = regexp.MustCompile(`use\s+base\b[^;]*Exporter`)
	useParentExporter = regexp.MustCompile(`use\s+parent\b[^;]*Exporter`)
	isaExporter       = regexp.MustCompile(`@ISA\s*=[^;]*\bExporter\b`)
	requireExporter   = regexp.MustCompile(`require\s+Exporter\b`)
)

// TestExporterKeepsResolution: every privlib module using Exporter resolves
// as it did without the declaration.
func TestExporterKeepsResolution(t *testing.T) {
	_, diffs := keepsResolution(t, "Exporter", "Exporter.tpm")
	for _, d := range diffs {
		t.Error(d)
	}
}

// TestExporterKeepsResolutionIsNotVacuous: the comparison is not vacuous,
// and it exercises every form a module reaches Exporter by -- use base,
// use parent, @ISA and require alike -- not only the commonest.
func TestExporterKeepsResolutionIsNotVacuous(t *testing.T) {
	p := perltest.Perl542(t)
	files := privlibFilesNaming(t, p.Privlib, "Exporter")
	compared, _ := keepsResolution(t, "Exporter", "Exporter.tpm")
	require.Greater(t, compared, 0, "compared no file where privlib is present")

	var sawBase, sawParent, sawISA, sawRequire bool
	for _, f := range files {
		sawBase = sawBase || namesIn(t, f, useBaseExporter)
		sawParent = sawParent || namesIn(t, f, useParentExporter)
		sawISA = sawISA || namesIn(t, f, isaExporter)
		sawRequire = sawRequire || namesIn(t, f, requireExporter)
	}
	assert.True(t, sawBase, "no privlib file reaches Exporter by use base")
	assert.True(t, sawParent, "no privlib file reaches Exporter by use parent")
	assert.True(t, sawISA, "no privlib file reaches Exporter by @ISA")
	assert.True(t, sawRequire, "no privlib file reaches Exporter by require")
}

// TestXSLoaderKeepsResolution: every privlib module using XSLoader
// resolves as it did without the declaration.
func TestXSLoaderKeepsResolution(t *testing.T) {
	_, diffs := keepsResolution(t, "XSLoader", "XSLoader.tpm")
	for _, d := range diffs {
		t.Error(d)
	}
}

// TestXSLoaderKeepsResolutionIsNotVacuous: the comparison is not vacuous.
func TestXSLoaderKeepsResolutionIsNotVacuous(t *testing.T) {
	compared, _ := keepsResolution(t, "XSLoader", "XSLoader.tpm")
	require.Greater(t, compared, 0, "compared no file where privlib is present")
}
