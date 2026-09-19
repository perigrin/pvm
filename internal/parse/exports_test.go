// ABOUTME: Reading a module's exports and prototypes out of its own text — no symbol table, no perl.
// ABOUTME: Three tiers: prototype known, name known but shape unknown (XS), name unknown.

package parse_test

import (
	"os"
	"path/filepath"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// writeModules lays out a temp directory and returns it.
func writeModules(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, src := range files {
		p := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// TestReadsLiteralExportList: `our @EXPORT = qw(...)` is read from the
// module's own text.
//
// Measured on Test/More.pm, 5.42.0: the list is a literal qw, and there are
// 39 static `sub NAME` declarations with 0 glob-assignment generations. The
// text has everything.
func TestReadsLiteralExportList(t *testing.T) {
	dir := writeModules(t, map[string]string{
		"My/Mod.pm": `package My::Mod;
our @EXPORT = qw(alpha beta);
sub alpha { 1 }
sub beta { 2 }
1;
`,
		"main.pl": "use My::Mod;\n",
	})

	root, err := parse.ParseFile(filepath.Join(dir, "main.pl"))
	if err != nil {
		t.Fatal(err)
	}

	imports := parse.Imports(root)
	for _, name := range []string{"alpha", "beta"} {
		if _, ok := imports[name]; !ok {
			t.Errorf("%s was not imported; have %v", name, keys(imports))
		}
	}
	if len(imports) != 2 {
		t.Errorf("imported %v, want exactly alpha and beta", keys(imports))
	}
}

// TestNonLiteralExportListDeclines: an export list that is not a literal qw
// or list of strings leaves the module UNRESOLVED rather than half-read.
//
// This is the quarantine boundary. A computed list is opaque by construction,
// and half an answer is worse than none: a name missing from a partial list
// reads as "not exported" when the truth is "not known".
func TestNonLiteralExportListDeclines(t *testing.T) {
	dir := writeModules(t, map[string]string{
		"Computed.pm": `package Computed;
our @EXPORT = map { "gen_$_" } qw(a b);
sub gen_a { 1 }
1;
`,
		"main.pl": "use Computed;\n",
	})

	root, err := parse.ParseFile(filepath.Join(dir, "main.pl"))
	if err != nil {
		t.Fatal(err)
	}

	if got := parse.Imports(root); len(got) != 0 {
		t.Errorf("imports = %v, want none: the export list is computed", keys(got))
	}
}

// TestReadsPrototypeFromSource: a sub's prototype comes from its declaration,
// with an explicit null for none -- distinct from not-known.
//
// That three-way distinction is the whole point:
//
//	prototype known    Test::More   resolve fully
//	no prototype       sub subtest  known to have none, which IS a shape
//	not known          List::Util   XS, refuse rather than guess
func TestReadsPrototypeFromSource(t *testing.T) {
	dir := writeModules(t, map[string]string{
		"Protos.pm": `package Protos;
our @EXPORT = qw(one none opt);
sub one ($) { 1 }
sub none { 2 }
sub opt (;$) { 3 }
1;
`,
		"main.pl": "use Protos;\n",
	})

	root, err := parse.ParseFile(filepath.Join(dir, "main.pl"))
	if err != nil {
		t.Fatal(err)
	}
	imports := parse.Imports(root)

	for _, c := range []struct {
		name  string
		proto string
		known bool
	}{
		{"one", "($)", true},
		{"opt", "(;$)", true},
		{"none", "", true}, // declared with no prototype: known, and empty
	} {
		got, ok := imports[c.name]
		if !ok {
			t.Errorf("%s not imported", c.name)
			continue
		}
		if got.PrototypeKnown != c.known {
			t.Errorf("%s: PrototypeKnown = %v, want %v",
				c.name, got.PrototypeKnown, c.known)
		}
		if got.Prototype != c.proto {
			t.Errorf("%s: prototype = %q, want %q",
				c.name, got.Prototype, c.proto)
		}
	}
}

// TestImportListRestricts: `use Test::More qw(ok)` imports a subset, and the
// list is in the `use` statement, so this is resolvable rather than assumed.
func TestImportListRestricts(t *testing.T) {
	mod := `package My::Mod;
our @EXPORT = qw(alpha beta gamma);
sub alpha { 1 }
sub beta { 2 }
sub gamma { 3 }
1;
`
	t.Run("a list restricts", func(t *testing.T) {
		dir := writeModules(t, map[string]string{
			"My/Mod.pm": mod,
			"main.pl":   "use My::Mod qw(alpha gamma);\n",
		})
		root, err := parse.ParseFile(filepath.Join(dir, "main.pl"))
		if err != nil {
			t.Fatal(err)
		}
		imports := parse.Imports(root)
		if _, ok := imports["beta"]; ok {
			t.Error("beta was not in the import list and must not be imported")
		}
		for _, name := range []string{"alpha", "gamma"} {
			if _, ok := imports[name]; !ok {
				t.Errorf("%s was in the import list and must be imported", name)
			}
		}
	})

	t.Run("no list means @EXPORT", func(t *testing.T) {
		dir := writeModules(t, map[string]string{
			"My/Mod.pm": mod,
			"main.pl":   "use My::Mod;\n",
		})
		root, err := parse.ParseFile(filepath.Join(dir, "main.pl"))
		if err != nil {
			t.Fatal(err)
		}
		if got := parse.Imports(root); len(got) != 3 {
			t.Errorf("imports = %v, want all three of @EXPORT", keys(got))
		}
	})

	t.Run("an empty list imports nothing", func(t *testing.T) {
		// `use My::Mod ();` is the explicit "load but import nothing" form,
		// and is NOT the same as omitting the list.
		dir := writeModules(t, map[string]string{
			"My/Mod.pm": mod,
			"main.pl":   "use My::Mod ();\n",
		})
		root, err := parse.ParseFile(filepath.Join(dir, "main.pl"))
		if err != nil {
			t.Fatal(err)
		}
		if got := parse.Imports(root); len(got) != 0 {
			t.Errorf("imports = %v, want none: the list is empty", keys(got))
		}
	})
}

// TestXSModuleNamesWithoutShapes: a module whose subs are implemented in C has
// names in the text and bodies in a shared library.
//
// Measured on List/Util.pm, 5.42.0: 4 static `sub NAME` declarations, 2
// XSLoader/bootstrap calls, and a literal @EXPORT_OK naming all of first,
// reduce, sum, max. At runtime perl reports `first &@` -- the leading
// ampersand that licenses the block form -- and none of that is in the file.
//
// So the name is known and the shape is NOT. Refusing is not conservatism:
// guessing "list operator" for `first` makes the block form a syntax error,
//
//	sub fake (@) { }   fake { $_ > 1 } (1,2)   ->  syntax error near "} ("
//
// while `first { $_ > 1 } @a` is valid and returns 2.
func TestXSModuleNamesWithoutShapes(t *testing.T) {
	dir := writeModules(t, map[string]string{
		"Fast/Util.pm": `package Fast::Util;
require XSLoader;
XSLoader::load('Fast::Util');
our @EXPORT_OK = qw(quick slow);
1;
`,
		"main.pl": "use Fast::Util qw(quick);\n",
	})

	root, err := parse.ParseFile(filepath.Join(dir, "main.pl"))
	if err != nil {
		t.Fatal(err)
	}

	imports := parse.Imports(root)
	got, ok := imports["quick"]
	if !ok {
		t.Fatalf("quick must be imported by NAME; have %v", keys(imports))
	}
	if got.PrototypeKnown {
		t.Error("quick has no visible declaration: its shape must be unknown")
	}
}

// TestLocalSubShadowsImport: a local sub shadowing an import is visible in the
// file being parsed and wins.
func TestLocalSubShadowsImport(t *testing.T) {
	dir := writeModules(t, map[string]string{
		"My/Mod.pm": `package My::Mod;
our @EXPORT = qw(shared);
sub shared ($) { 1 }
1;
`,
		"main.pl": `use My::Mod;
sub shared { 42 }
`,
	})

	root, err := parse.ParseFile(filepath.Join(dir, "main.pl"))
	if err != nil {
		t.Fatal(err)
	}

	got, ok := parse.Imports(root)["shared"]
	if !ok {
		t.Fatal("shared must still be known")
	}
	if got.Local != true {
		t.Error("a local sub shadows the import and must be marked local")
	}
	// The local declaration has no prototype, and that is the answer -- not
	// the module's ($).
	if got.Prototype != "" {
		t.Errorf("prototype = %q, want the LOCAL sub's (none)", got.Prototype)
	}
}

func keys(m map[string]parse.Import) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
