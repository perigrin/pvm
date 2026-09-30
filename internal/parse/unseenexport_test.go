// ABOUTME: A sub a module exports but builds out of sight -- a glob assigned in a
// ABOUTME: loop -- has no prototype; an XS module's unread subs stay unknown.

package parse_test

import (
	"os"
	"path/filepath"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestUnseenExportHasNoPrototype: File::Spec::Functions exports catdir but
// never writes `sub catdir`; it assigns `*{$meth}` in a loop over @EXPORT.
// Measured on 5.42.0, `prototype "main::catdir"` is undef after `use
// File::Spec::Functions`, so `catdir $Cwd, 'op'` is `catdir($Cwd, 'op')`.
// A sub built in Perl this way is taken to have no prototype, by
// perigrin's decision (2026-09-30).
//
// An XS module is different: its subs are C, and one whose .xs was not read
// may well have a prototype -- List::Util's `first` is `&@`, and reading it
// as a list operator would make `first { ... } @a` a syntax error. Those
// stay unknown. perl.git t/op/chdir.t:188.
func TestUnseenExportHasNoPrototype(t *testing.T) {
	lib := t.TempDir()
	write := func(rel, body string) {
		path := filepath.Join(lib, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("Foo/Gen.pm", "package Foo::Gen;\nuse Exporter 'import';\nour @EXPORT = qw(catdir);\n"+
		"foreach my $meth (@EXPORT) { no strict 'refs'; *{$meth} = sub { join '/', @_ }; }\n1;\n")
	write("Foo/C.pm", "package Foo::C;\nuse Exporter 'import';\nour @EXPORT = qw(first);\n"+
		"require XSLoader;\nXSLoader::load();\n1;\n")
	load := parse.DirLoader(lib)

	src := `use Foo::Gen; my $Cwd; my $d = catdir $Cwd, 'op';`
	root := parse.ParseWithLoader([]byte(src), load)
	if got := countUnknown(root); got != 0 {
		t.Errorf("%q: %d Unknown, want 0; got %s", src, got, shape(root))
	}
	if imp := parse.Imports(root)["catdir"]; !imp.PrototypeKnown || imp.Prototype != "" {
		t.Errorf("catdir imported as %+v, want no prototype, known", imp)
	}
	// A listed name the module does not export is no sub of its: `use
	// feature 'defer'` must leave `defer { ... }` the block it is.
	write("Foo/Flag.pm", "package Foo::Flag;\nsub import { }\n1;\n")
	if imp, ok := parse.Imports(parse.ParseWithLoader([]byte(`use Foo::Flag 'defer';`), load))["defer"]; ok && imp.PrototypeKnown {
		t.Errorf("defer imported as %+v; Foo::Flag does not export it", imp)
	}
	if imp := parse.Imports(parse.ParseWithLoader([]byte(`use Foo::C;`), load))["first"]; imp.PrototypeKnown {
		t.Errorf("first imported as %+v; an unread XS sub's prototype is not known", imp)
	}
}
