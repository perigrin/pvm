// ABOUTME: `use if` is parsed both ways: its import assumed to fail (the default)
// ABOUTME: or to pass, when the condition holds -- `use if C, M, LIST` is `use M LIST`.

package parse_test

import (
	"os"
	"path/filepath"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestUseIfBothWays: the condition of `use if COND, MODULE, LIST` is known
// only at compile time, so a file using it is read both ways (perigrin,
// 2026-09-30). Failing, the import defines nothing and `catdir $C, 'op'` is
// a method on $C. Passing, it is `use MODULE LIST`, catdir is imported, and
// the call is `catdir($C, 'op')` -- what a full perl does in perl.git
// t/op/coreamp.t:842, where !is_miniperl holds.
func TestUseIfBothWays(t *testing.T) {
	dir := t.TempDir()
	lib := filepath.Join(dir, "lib")
	if err := os.MkdirAll(filepath.Join(lib, "Foo"), 0o755); err != nil {
		t.Fatal(err)
	}
	gen := "package Foo::Gen;\nuse Exporter 'import';\nour @EXPORT_OK = qw(catdir);\n" +
		"foreach my $meth (@EXPORT_OK) { no strict 'refs'; *{$meth} = sub { join '/', @_ }; }\n1;\n"
	if err := os.WriteFile(filepath.Join(lib, "Foo", "Gen.pm"), []byte(gen), 0o644); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "x.t")
	src := "use if !$ENV{MINI}, Foo::Gen, qw(catdir);\nuse if 1, 'Foo::Gen';\nmy $C; my $d = catdir $C, 'op';\n"
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	fails, err := parse.ParseFileAssumingUseIf(path, false, lib)
	if err != nil {
		t.Fatal(err)
	}
	if call := indirectCall(fails, "catdir"); call == nil || countUnknown(fails) != 0 {
		t.Errorf("import failing: want `$C->catdir`, no Unknown; got %s", shape(fails))
	}

	passes, err := parse.ParseFileAssumingUseIf(path, true, lib)
	if err != nil {
		t.Fatal(err)
	}
	if indirectCall(passes, "catdir") != nil || countUnknown(passes) != 0 {
		t.Errorf("import passing: want `catdir($C, 'op')`, no Unknown; got %s", shape(passes))
	}
	if imp, ok := parse.Imports(passes)["catdir"]; !ok || !imp.PrototypeKnown {
		t.Errorf("import passing: catdir imported as %+v (found %v)", imp, ok)
	}
}

// TestSpacedQwImportList: `qw "catfile"` with a space before its delimiter is
// the list `catfile` -- measured on 5.42.0, `qw "a b"` is ("a", "b"). The
// import list read the delimiter itself as part of the first name.
// perl.git t/op/coreamp.t:842.
func TestSpacedQwImportList(t *testing.T) {
	lib := t.TempDir()
	if err := os.MkdirAll(filepath.Join(lib, "Foo"), 0o755); err != nil {
		t.Fatal(err)
	}
	pm := "package Foo::Q;\nuse Exporter 'import';\nour @EXPORT_OK = qw(a b);\nsub a {1} sub b {2}\n1;\n"
	if err := os.WriteFile(filepath.Join(lib, "Foo", "Q.pm"), []byte(pm), 0o644); err != nil {
		t.Fatal(err)
	}
	imps := parse.Imports(parse.ParseWithLoader([]byte(`use Foo::Q qw "a b";`), parse.DirLoader(lib)))
	for _, name := range []string{"a", "b"} {
		if _, ok := imps[name]; !ok {
			t.Errorf("`qw \"a b\"` did not import %s; imported %v", name, imps)
		}
	}
}
