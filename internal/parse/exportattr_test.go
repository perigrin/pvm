// ABOUTME: Tests that :export and :export_ok on a .tpm's subs decide what a `use` imports.
// ABOUTME: The fixture is testdata/declarations/My/Exports.tpm, read through the resolver.
package parse

import (
	"slices"
	"testing"
	"testing/fstest"
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

// TestDeclaredImportMarkedExportKeepsNameSemantics: perigrin's
// refinement, 2026-10-20 -- only a module's own `import` that carries no
// :export or :export_ok mark turns a use-list into import's arguments.
// My::MarkedImport marks its import :export_ok, as Exporter's does, so
// `use My::MarkedImport 'import';` installs the name import, not plan
// arguments, and `use My::MarkedImport 'helper';` installs helper alone.
func TestDeclaredImportMarkedExportKeepsNameSemantics(t *testing.T) {
	wantImports(t, map[string][]string{
		"use My::MarkedImport 'import';\n": {"import"},
		"use My::MarkedImport 'helper';\n": {"helper"},
		"use My::MarkedImport;\n":          nil,
	})
}

// TestEmptyUseListImportsNothing: measured on perl 5.42.0, `use Test::More
// qw();` installs no names at all, the same as `use Test::More ();`. A
// module whose own `import` is declared (Test::More, Locale::Maketext::Simple)
// must read `qw()` as the explicit empty list, not fall through to its
// :export defaults, as `use Cwd qw();` already does for the generic path.
func TestEmptyUseListImportsNothing(t *testing.T) {
	wantImports(t, map[string][]string{
		"use Test::More ();\n":   nil,
		"use Test::More qw();\n": nil,
	})
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

// refusedList wants a .tpm stating list to be in error, naming the module
// and line 3, and to import nothing from the list: neither a name it holds
// nor a tag it states.
func refusedList(t *testing.T, list, want string) {
	t.Helper()
	saved := declarations
	t.Cleanup(func() { declarations = saved })
	src := "package My::Listed;\nsub f :export () Int;\n" + list + "\nsub g () Int;\n"
	declarations = layeredFS{top: fstest.MapFS{"declarations/My/Listed.tpm": {Data: []byte(src)}}, base: saved}
	root := ParseWithLoader([]byte("use My::Listed;\nuse My::Listed qw(g);\n"), noModules)
	if errs := DeclarationErrors(root); len(errs) != 1 || errs[0].Error() != want {
		t.Errorf("errors %v, want %q", errs, want)
	}
	if _, ok := Imports(root)["g"]; ok {
		t.Errorf("g imported from a refused list: %v", Imports(root))
	}
	if _, ok := Imports(root)["f"]; !ok {
		t.Errorf("f, marked :export, not imported: %v", Imports(root))
	}
	if imports := Imports(ParseWithLoader([]byte("use My::Listed ':all';\n"), noModules)); len(imports) > 0 {
		t.Errorf("use My::Listed ':all' imported from a refused list: %v", imports)
	}
}

// TestDeclarationExportListRefused: ruling 2 (2026-10-16): a .tpm marks
// its exports with :export and :export_ok, and states no @EXPORT.
func TestDeclarationExportListRefused(t *testing.T) {
	refusedList(t, "our @EXPORT = qw(g);",
		"My::Listed: line 3: our @EXPORT: a .tpm marks each export :export or :export_ok, not in a list")
}

// TestDeclarationExportOkListRefused: nor an @EXPORT_OK.
func TestDeclarationExportOkListRefused(t *testing.T) {
	refusedList(t, "our @EXPORT_OK = qw(g);",
		"My::Listed: line 3: our @EXPORT_OK: a .tpm marks each export :export or :export_ok, not in a list")
}

// TestDeclarationExportTagsRefused: nor a %EXPORT_TAGS; a tag is the
// attribute's argument, :export_ok(tag). Its tag holds the exported f, so
// only the refusal keeps `:all` from importing it.
func TestDeclarationExportTagsRefused(t *testing.T) {
	refusedList(t, "our %EXPORT_TAGS = (all => [qw(f)]);",
		"My::Listed: line 3: our %EXPORT_TAGS: a .tpm marks each export :export or :export_ok, not in a list")
}

// TestModuleSourceExportListAmpersand: a module's own list may spell a
// sub `&f`. Measured on perl 5.42.0 with `@EXPORT = qw(&f)`, `@EXPORT_OK
// = qw(&g &h)` and `%EXPORT_TAGS = (t => [qw(&h)])`: a bare use defines
// f, `qw(g)` and `qw(&g)` define g, and `':t'` defines h.
func TestModuleSourceExportListAmpersand(t *testing.T) {
	src := "package Amp;\nuse Exporter 'import';\nour @EXPORT = qw(&f);\nour @EXPORT_OK = qw(&g &h);\nour %EXPORT_TAGS = (t => [qw(&h)]);\nsub f { 1 }\nsub g { 2 }\nsub h { 3 }\n"
	loader := func(m string) ([]byte, bool) { return []byte(src), m == "Amp" }
	for use, want := range map[string]string{
		"use Amp;\n": "f", "use Amp qw(g);\n": "g", "use Amp qw(&g);\n": "g", "use Amp ':t';\n": "h",
	} {
		var got []string
		for name := range Imports(ParseWithLoader([]byte(use), loader)) {
			got = append(got, name)
		}
		if !slices.Equal(got, []string{want}) {
			t.Errorf("%q imports %v, want [%s]", use, got, want)
		}
	}
}

// TestDeclarationUnknownAttributeRefused: a .tpm marks its exports, so a
// misspelled mark is an error, not a name silently left unexported. A sub
// takes perl's own lvalue, method and prototype and the declaration
// language's; an our, :export and :export_ok.
func TestDeclarationUnknownAttributeRefused(t *testing.T) {
	saved := declarations
	t.Cleanup(func() { declarations = saved })
	for src, want := range map[string]string{
		"sub bad :exported () Int;\n":                "My::Odd: sub bad: unknown attribute :exported",
		"multi sub bad :export_okay () Int;\n":       "My::Odd: sub bad: unknown attribute :export_okay",
		"our $x :exported;\n":                        "My::Odd: our $x: unknown attribute :exported; an our takes :export or :export_ok",
		"sub fine :lvalue :method :export () Int;\n": "",
	} {
		declarations = layeredFS{top: fstest.MapFS{"declarations/My/Odd.tpm": {Data: []byte("package My::Odd;\n" + src)}}, base: saved}
		var got []string
		for _, err := range DeclarationErrors(ParseWithLoader([]byte("use My::Odd;\n"), noModules)) {
			got = append(got, err.Error())
		}
		if want == "" && len(got) > 0 || want != "" && !slices.Equal(got, []string{want}) {
			t.Errorf("%q: errors %q, want %q", src, got, want)
		}
	}
}

// TestMultiCandidatesMarkedDifferentlyRefused: perl exports a name, not a
// candidate, so a multi's candidates marking it differently contradict
// each other: :export against :export_ok, different tags, or a mark on
// only some candidates.
func TestMultiCandidatesMarkedDifferentlyRefused(t *testing.T) {
	saved := declarations
	t.Cleanup(func() { declarations = saved })
	const differ = "My::Multi: multi sub f: candidates mark f differently"
	for src, want := range map[string]string{
		"multi sub f :export (Int $i) Str;\nmulti sub f :export_ok (Str $s) Int;\n":               differ,
		"multi sub f :export_ok(a) (Int $i) Str;\nmulti sub f :export_ok(b) (Str $s) Int;\n":      differ,
		"multi sub f :export (Int $i) Str;\nmulti sub f (Str $s) Int;\n":                          differ,
		"multi sub f (Int $i) Str;\nmulti sub f :export (Str $s) Int;\n":                          differ,
		"multi sub f :export_ok(a,b) (Int $i) Str;\nmulti sub f :export_ok(b, a) (Str $s) Int;\n": "",
	} {
		declarations = layeredFS{top: fstest.MapFS{"declarations/My/Multi.tpm": {Data: []byte("package My::Multi;\n" + src)}}, base: saved}
		var got []string
		for _, err := range DeclarationErrors(ParseWithLoader([]byte("use My::Multi;\n"), noModules)) {
			got = append(got, err.Error())
		}
		if want == "" && len(got) > 0 || want != "" && !slices.Equal(got, []string{want}) {
			t.Errorf("%q: errors %q, want %q", src, got, want)
		}
	}
}

// TestDeclarationExportListFormsRefused: an export list is refused in
// every form a .tpm could state one, not only as a top-level `our` or
// assignment: in a declared list, pushed onto, package-qualified, inside
// a block, declared by `use vars`, or assigned an element. A parameter
// named like one is local and exports nothing, and %EXPORT is no list.
func TestDeclarationExportListFormsRefused(t *testing.T) {
	saved := declarations
	t.Cleanup(func() { declarations = saved })
	const tail = ": a .tpm marks each export :export or :export_ok, not in a list"
	for src, want := range map[string][]string{
		"our (@EXPORT) = qw(f);\n":                {"My::Odd: line 2: our @EXPORT" + tail},
		"our (@EXPORT, @EXPORT_OK) = (qw(f));\n":  {"My::Odd: line 2: our @EXPORT" + tail, "My::Odd: line 2: our @EXPORT_OK" + tail},
		"push @EXPORT, 'f';\n":                    {"My::Odd: line 2: @EXPORT" + tail},
		"@My::Odd::EXPORT = qw(f);\n":             {"My::Odd: line 2: @My::Odd::EXPORT" + tail},
		"BEGIN { our @EXPORT = qw(f) }\n":         {"My::Odd: line 2: our @EXPORT" + tail},
		"use vars qw(@EXPORT_OK);\n":              {"My::Odd: line 2: use vars @EXPORT_OK" + tail},
		"$EXPORT_TAGS{all} = [qw(f)];\n":          {"My::Odd: line 2: $EXPORT_TAGS" + tail},
		"our $EXPORT :export;\n":                  nil,
		"$EXPORT[0] = 'f';\n":                     {"My::Odd: line 2: $EXPORT" + tail},
		"$EXPORT{f} = 1;\n":                       nil,
		"our %EXPORT :export;\n":                  nil,
		"sub g (Array \\@EXPORT) Int;\n":          nil,
		"sub g (@EXPORT) Int;\n":                  nil,
		"sub g (Hash \\%EXPORT_TAGS) Int;\n":      nil,
		"multi sub g (Array \\@EXPORT_OK) Int;\n": nil,
	} {
		declarations = layeredFS{top: fstest.MapFS{"declarations/My/Odd.tpm": {Data: []byte("package My::Odd;\n" + src + "sub f () Int;\n")}}, base: saved}
		var got []string
		for _, err := range DeclarationErrors(ParseWithLoader([]byte("use My::Odd;\n"), noModules)) {
			got = append(got, err.Error())
		}
		if !slices.Equal(got, want) {
			t.Errorf("%q: errors %q, want %q", src, got, want)
		}
	}
}
