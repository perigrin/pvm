// ABOUTME: The milestone's closing gate: every top-50 core module has a declaration or a recorded,
// ABOUTME: ruling-4-sanctioned reason why not, read from core_module_not_shipped.txt.
package parse_test

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// notShippedEntry is one parsed line of core_module_not_shipped.txt.
type notShippedEntry struct {
	Module string
	Issue  string
	Reason string
}

// fullUUID matches a complete 36-character id, never the 8-character
// prefix a short citation uses: a prefix is ambiguous (reIssue's comment
// in internal/conformance/file.go measures the collision), and this file's
// format requires the whole id.
var fullUUID = regexp.MustCompile(`^[0-9a-f]{8}(?:-[0-9a-f]{4}){3}-[0-9a-f]{12}$`)

// infrastructureModules are the only modules ruling 4 (perigrin,
// 2026-10-16) permits listing not shipped: Exporter, XSLoader and
// DynaLoader, each inherited from or called by many other modules' own
// imports, so a declaration for one could change how those resolve.
var infrastructureModules = map[string]bool{"Exporter": true, "XSLoader": true, "DynaLoader": true}

// parseNotShipped reads core_module_not_shipped.txt's format: one
// Module<TAB>full-uuid<TAB>reason line per entry; blank lines and lines
// starting with # (after leading space) are comments, not entries. An
// entry missing a field, an empty reason, an id that is not a full uuid,
// or a module other than Exporter, XSLoader or DynaLoader is an error, not
// a silent skip, so a malformed or unauthorized entry never passes the
// coverage check vacuously.
func parseNotShipped(data []byte) (map[string]notShippedEntry, []error) {
	entries := map[string]notShippedEntry{}
	var errs []error
	for i, line := range strings.Split(string(data), "\n") {
		line = strings.TrimRight(line, "\r")
		if line == "" || strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) != 3 || fields[0] == "" || fields[1] == "" || fields[2] == "" {
			errs = append(errs, fmt.Errorf("line %d: not Module<TAB>full-uuid<TAB>reason: %q", i+1, line))
			continue
		}
		module, issue, reason := fields[0], fields[1], fields[2]
		if !fullUUID.MatchString(issue) {
			errs = append(errs, fmt.Errorf("line %d: %s: %q is not a full uuid", i+1, module, issue))
			continue
		}
		if !infrastructureModules[module] {
			errs = append(errs, fmt.Errorf("line %d: %s: not shipped is allowed only for Exporter, XSLoader and DynaLoader (ruling 4)", i+1, module))
			continue
		}
		entries[module] = notShippedEntry{module, issue, reason}
	}
	return entries, errs
}

// modulePath is the declaration path for module, relative to
// declarations/: `::` becomes `/`, and an underscore stays, so
// `File::Spec::Functions` maps to `File/Spec/Functions.tpm` and
// `SDBM_File` to `SDBM_File.tpm`.
func modulePath(module string) string {
	return strings.ReplaceAll(module, "::", "/") + ".tpm"
}

// checkCoverage requires every ranked module to have exactly one of a
// declaration (declared reports whether modulePath's file exists) or a
// not-shipped entry -- never both, never neither.
func checkCoverage(ranked []string, declared func(path string) bool, notShipped map[string]notShippedEntry) []error {
	var errs []error
	for _, module := range ranked {
		path := modulePath(module)
		has := declared(path)
		_, skipped := notShipped[module]
		switch {
		case has && skipped:
			errs = append(errs, fmt.Errorf("%s: ships declarations/%s and is also listed not shipped", module, path))
		case !has && !skipped:
			errs = append(errs, fmt.Errorf("%s: no declarations/%s and no not-shipped entry", module, path))
		}
	}
	return errs
}

// readRankedModules reads the module names (ignoring their counts) from
// core_module_ranking.txt, the pinned top-50 ranking.
func readRankedModules(t *testing.T) []string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "core_module_ranking.txt"))
	require.NoError(t, err)
	var names []string
	for _, line := range strings.Split(strings.TrimRight(string(data), "\n"), "\n") {
		fields := strings.Fields(line)
		require.Len(t, fields, 2, "core_module_ranking.txt line: %q", line)
		names = append(names, fields[0])
	}
	return names
}

func TestCoreModuleDeclarationsCoverTheTop50(t *testing.T) {
	ranked := readRankedModules(t)
	require.Len(t, ranked, 50)

	data, err := os.ReadFile(filepath.Join("testdata", "core_module_not_shipped.txt"))
	require.NoError(t, err)
	notShipped, errs := parseNotShipped(data)
	require.Empty(t, errs, "core_module_not_shipped.txt")

	declared := func(path string) bool {
		_, err := os.Stat(filepath.Join("declarations", filepath.FromSlash(path)))
		return err == nil
	}
	errs = checkCoverage(ranked, declared, notShipped)
	assert.Empty(t, errs, "every top-50 module needs a declaration or a recorded not-shipped reason")
}

func TestCoverageRejectsAnUndeclaredModule(t *testing.T) {
	errs := checkCoverage([]string{"Foo::Bar"}, func(string) bool { return false }, nil)
	require.Len(t, errs, 1)
	assert.Contains(t, errs[0].Error(), "Foo::Bar")
}

func TestCoverageRejectsANotShippedEntryWithoutReason(t *testing.T) {
	cases := map[string]string{
		"two fields, no reason":                   "Exporter\t01a1464e-dd6c-7403-95d2-25f4855404b2\n",
		"empty reason field":                      "Exporter\t01a1464e-dd6c-7403-95d2-25f4855404b2\t\n",
		"id is an 8-char prefix, not a full uuid": "Exporter\t01a1464e\tmeasured resolution change\n",
	}
	for name, data := range cases {
		t.Run(name, func(t *testing.T) {
			entries, errs := parseNotShipped([]byte(data))
			assert.NotEmpty(t, errs, name)
			assert.Empty(t, entries, name)
		})
	}
}

func TestCoverageNotShippedSkipsComments(t *testing.T) {
	data := "# Module\tfull-uuid\treason (header, not an entry)\n" +
		"Exporter\t01a1464e-dd6c-7403-95d2-25f4855404b2\tmeasured clean against the corpus\n"
	entries, errs := parseNotShipped([]byte(data))
	require.Empty(t, errs)
	require.Len(t, entries, 1)
	assert.Contains(t, entries, "Exporter")
}

func TestCoverageRejectsShippedAndNotShipped(t *testing.T) {
	notShipped := map[string]notShippedEntry{
		"Exporter": {"Exporter", "01a1464e-dd6c-7403-95d2-25f4855404b2", "measured resolution change"},
	}
	errs := checkCoverage([]string{"Exporter"}, func(string) bool { return true }, notShipped)
	require.Len(t, errs, 1)
	assert.Contains(t, errs[0].Error(), "Exporter")
}

func TestCoverageNotShippedOnlyInfrastructure(t *testing.T) {
	for _, module := range []string{"POSIX", "JSON::PP"} {
		data := fmt.Sprintf("%s\t01a1464e-dd6c-7403-95d2-25f4855404b2\tnot actually excludable\n", module)
		entries, errs := parseNotShipped([]byte(data))
		assert.NotEmpty(t, errs, module)
		assert.Empty(t, entries, module)
	}
}

func TestCoverageMapsModuleNamesToPaths(t *testing.T) {
	assert.Equal(t, "File/Spec/Functions.tpm", modulePath("File::Spec::Functions"))
	assert.Equal(t, "SDBM_File.tpm", modulePath("SDBM_File"))
}
