// ABOUTME: A Test::Builder::Module subclass imports its @EXPORT whatever plan it is
// ABOUTME: given, or the names under `import => [...]`: `use Test::More tests => 9`.

package parse_test

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestBuilderModuleImport: Test::Builder::Module's import (perl 5.42.0's
// Test/Builder/Module.pm:75-122) takes the names under `import => [...]` as
// the import list, hands every other argument to plan(), and calls
// Exporter::import with what it took -- @EXPORT when that is nothing.
// Test::More's import_extra only records no_diag, merges `import` lists and
// exports $TODO. So `use Test::More tests => 9;` imports is, ok and the rest,
// which the import reader dropped as a computed list. 81 of the 83 T1 files
// still dirty with a loader refuse on that.
func TestBuilderModuleImport(t *testing.T) {
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
	write("Test/Builder/Module.pm", "package Test::Builder::Module;\nuse Exporter;\nour @ISA = qw(Exporter);\nsub import { }\n1;\n")
	write("My/T.pm", "package My::T;\nuse Test::Builder::Module;\nour @ISA = qw(Test::Builder::Module);\n"+
		"our @EXPORT = qw(ok is);\nsub ok { } sub is { }\n1;\n")
	load := parse.DirLoader(lib)

	for src, want := range map[string]string{
		`use My::T tests => 3;`:                     "is ok",
		`use My::T;`:                                "is ok",
		`use My::T 'no_plan';`:                      "is ok",
		`use My::T import => ['ok'];`:               "ok",
		`use My::T tests => 2, import => [qw(is)];`: "is",
	} {
		imps := parse.Imports(parse.ParseWithLoader([]byte(src), load))
		var got []string
		for name := range imps {
			got = append(got, name)
		}
		sort.Strings(got)
		if g := join(got); g != want {
			t.Errorf("%q imports %q, want %q", src, g, want)
		}
	}

	src := "use My::T tests => 1;\nmy @seen;\nis scalar(@seen), 0, 'empty';\n"
	if n := parse.ParseWithLoader([]byte(src), load); countUnknown(n) != 0 {
		t.Errorf("%q: %d Unknown; got %s", src, countUnknown(n), shape(n))
	}
}

func join(s []string) string {
	out := ""
	for i, v := range s {
		if i > 0 {
			out += " "
		}
		out += v
	}
	return out
}
