// ABOUTME: A bare `use` imports @EXPORT only, `:DEFAULT` in a list names @EXPORT,
// ABOUTME: and an export list assigned without `our` is read as with it.

package parse_test

import (
	"path/filepath"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestExporterDefaults: measured on 5.42.0 against a module whose @EXPORT is
// (alpha) and @EXPORT_OK (beta), assigned without `our` as Devel::Peek does:
//
//	$ perl -I. -e 'use My::Mod; print alpha(), " ",
//	      defined(&beta) ? "beta imported" : "beta not imported", "\n";
//	      package Q; use My::Mod qw(:DEFAULT beta); print alpha(), beta(), "\n";'
//	a beta not imported
//	ab
//
// perl.git t/op/chdir.t:28 imports `qw(:DEFAULT splitdir rel2abs splitpath)`
// from File::Spec::Functions; t/op/hash-clear-placeholders.t uses
// Devel::Peek, whose `@EXPORT = qw(Dump ...)` has no `our`.
func TestExporterDefaults(t *testing.T) {
	dir := writeModules(t, map[string]string{
		"My/Mod.pm": "package My::Mod;\n@EXPORT = qw(alpha);\n@EXPORT_OK = qw(beta);\n" +
			"sub alpha { 1 }\nsub beta { 2 }\n1;\n",
		"bare.pl":    "use My::Mod;\n",
		"default.pl": "use My::Mod qw(:DEFAULT beta);\n",
	})
	for _, tc := range []struct {
		file string
		want []string
		not  []string
	}{
		{"bare.pl", []string{"alpha"}, []string{"beta"}},
		{"default.pl", []string{"alpha", "beta"}, nil},
	} {
		root, err := parse.ParseFile(filepath.Join(dir, tc.file))
		if err != nil {
			t.Fatal(err)
		}
		imports := parse.Imports(root)
		for _, name := range tc.want {
			if _, ok := imports[name]; !ok {
				t.Errorf("%s: %s was not imported; have %v", tc.file, name, keys(imports))
			}
		}
		for _, name := range tc.not {
			if _, ok := imports[name]; ok {
				t.Errorf("%s: %s is @EXPORT_OK and a bare use does not import it", tc.file, name)
			}
		}
	}
}
