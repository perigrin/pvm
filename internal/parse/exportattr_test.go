// ABOUTME: Tests that :export and :export_ok on a .tpm's subs decide what a `use` imports.
// ABOUTME: The fixture is testdata/declarations/My/Exports.tpm, read through the resolver.
package parse

import (
	"slices"
	"testing"
)

// importedFrom parses src with the fixture declarations in place and
// returns the names its `use` imported, sorted.
func importedFrom(t *testing.T, src string) []string {
	t.Helper()
	useLibraryFixture(t)
	root := ParseWithLoader([]byte(src), func(string) ([]byte, bool) { return nil, false })
	if errs := root.declErrs; len(errs) > 0 {
		t.Fatalf("%q: declaration errors: %v", src, errs)
	}
	var names []string
	for name, imp := range Imports(root) {
		if !imp.Local {
			names = append(names, name)
		}
	}
	slices.Sort(names)
	return names
}

// wantImports wants each src to import exactly the names given for it.
func wantImports(t *testing.T, cases map[string][]string) {
	t.Helper()
	for src, want := range cases {
		if got := importedFrom(t, src); !slices.Equal(got, want) {
			t.Errorf("%q imports %v, want %v", src, got, want)
		}
	}
}

// TestSubExportAttribute: RFC 0001 "Every exported name". `:export` is
// what a bare `use` brings in, `:export_ok` what must be asked for.
func TestSubExportAttribute(t *testing.T) {
	wantImports(t, map[string][]string{
		"use My::Exports;\n":         {"both", "f"},
		"use My::Exports qw(g);\n":   {"g"},
		"use My::Exports qw(f g);\n": {"f", "g"},
	})
}

// TestSubExportTag: an attribute's argument is a tag. `:export_ok(flock)`
// puts LOCK_EX in `:flock`, which a bare `use` leaves alone, and
// `:export(a, b)` puts `both` in each of its tags.
func TestSubExportTag(t *testing.T) {
	wantImports(t, map[string][]string{
		"use My::Exports ':flock';\n":    {"LOCK_EX"},
		"use My::Exports qw(LOCK_EX);\n": {"LOCK_EX"},
		"use My::Exports ':a';\n":        {"both"},
		"use My::Exports ':b';\n":        {"both"},
	})
	if got := importedFrom(t, "use My::Exports;\n"); slices.Contains(got, "LOCK_EX") {
		t.Errorf("a bare use imports %v; LOCK_EX is :export_ok", got)
	}
}

// TestSubExportUnknownTagImportsNothing: a tag no declared name carries
// names nothing.
func TestSubExportUnknownTagImportsNothing(t *testing.T) {
	wantImports(t, map[string][]string{"use My::Exports ':nosuch';\n": nil})
}

// TestSubWithoutExportNotImported: a declaration states every name its
// module exports, so `h`, which carries no mark, is imported neither by a
// bare `use` nor by name, as Exporter refuses an unexported name.
func TestSubWithoutExportNotImported(t *testing.T) {
	wantImports(t, map[string][]string{"use My::Exports qw(h);\n": nil})
	if got := importedFrom(t, "use My::Exports;\n"); slices.Contains(got, "h") {
		t.Errorf("a bare use imports %v; h carries no export mark", got)
	}
}
