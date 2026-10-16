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

// TestShippedDeclarationsKeepTheirImports: marking exports with :export
// and :export_ok imports what each shipped .tpm's export lists did, but
// for Carp's verbose -- see TestCarpVerboseImportsNothing.
func TestShippedDeclarationsKeepTheirImports(t *testing.T) {
	wantImports(t, map[string][]string{
		"use Carp;\n":                        {"carp", "confess", "croak"},
		"use Carp qw(croak);\n":              {"croak"},
		"use Carp qw(cluck);\n":              {"cluck"},
		"use Carp qw(longmess shortmess);\n": {"longmess", "shortmess"},
		// Measured on perl 5.42.0: `use Carp qw(verbose)` installs nothing
		// in the caller; Carp's @EXPORT_FAIL takes verbose as a flag. The
		// export list put it in @EXPORT_OK, which imported "verbose".
		"use Carp qw(verbose);\n": nil,
		"use Moose;\n": {"after", "around", "augment", "before", "blessed", "confess",
			"extends", "has", "inner", "meta", "override", "super", "with"},
		"use Moose qw(meta);\n": {"meta"},
		"use Moose::Role;\n": {"after", "around", "augment", "before", "blessed", "confess",
			"excludes", "extends", "has", "inner", "meta", "override", "requires", "super", "with"},
		"use Mojolicious::Lite;\n": {"any", "app", "del", "get", "group", "helper", "hook", "new",
			"options", "patch", "plugin", "post", "put", "query", "under", "websocket"},
		"use Mojolicious::Lite qw(get);\n": {"get"},
		"use Future::AsyncAwait;\n":        nil,
		"use My::Lib;\n":                   {"alias_all", "emit", "fields", "join_all", "need", "pick", "size", "when_"},
		"use My::Lib qw(join_all);\n":      {"join_all"},
		"use My::Lib qw(hidden);\n":        nil,
		"use My::Lib qw(plus);\n":          nil,
	})
}

// TestCarpVerboseImportsNothing: measured on perl 5.42.0, `use Carp
// qw(verbose)` defines no verbose in the caller, so Carp.tpm marks none.
func TestCarpVerboseImportsNothing(t *testing.T) {
	wantImports(t, map[string][]string{"use Carp qw(verbose);\n": nil})
}

// TestModuleSourceExportListStillRead: the refusal of export lists is for
// a .tpm alone; a module's own source is read for them as ever.
func TestModuleSourceExportListStillRead(t *testing.T) {
	src := "package Plain;\nuse Exporter 'import';\nour @EXPORT = qw(f);\nour @EXPORT_OK = qw(g);\nsub f { }\nsub g { }\n"
	root := ParseWithLoader([]byte("use Plain;\nuse Plain qw(g);\n"), func(m string) ([]byte, bool) {
		return []byte(src), m == "Plain"
	})
	if errs := DeclarationErrors(root); len(errs) > 0 {
		t.Fatalf("declaration errors: %v", errs)
	}
	for _, name := range []string{"f", "g"} {
		if _, ok := Imports(root)[name]; !ok {
			t.Errorf("%s not imported from a module's own export list: %v", name, Imports(root))
		}
	}
}
