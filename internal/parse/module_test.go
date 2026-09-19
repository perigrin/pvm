// ABOUTME: ParseFile and the module loader — a `use` resolved against the file's own directory.
// ABOUTME: The error is the read and only the read; an unresolvable import declines rather than failing.

package parse_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestResolvesModuleSource: a module named in `use` is located on the search
// path and PARSED -- with this parser, not by running perl.
//
// The loader is the seam. Production searches directories; a test hands over
// bytes it wrote itself, so the suite needs no installed Test::More and no
// filesystem layout to be true.
func TestResolvesModuleSource(t *testing.T) {
	dir := t.TempDir()
	mod := filepath.Join(dir, "My", "Mod.pm")
	if err := os.MkdirAll(filepath.Dir(mod), 0o755); err != nil {
		t.Fatal(err)
	}
	const modSrc = `package My::Mod;
our @EXPORT = qw(hello);
sub hello ($) { 1 }
1;
`
	if err := os.WriteFile(mod, []byte(modSrc), 0o644); err != nil {
		t.Fatal(err)
	}

	main := filepath.Join(dir, "main.pl")
	if err := os.WriteFile(main, []byte("use My::Mod;\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	root, err := parse.ParseFile(main)
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	if root == nil {
		t.Fatal("ParseFile returned no tree")
	}

	// The module was reached and parsed. Asserted on the loader's record
	// rather than on any enrichment of the tree: reading exports is
	// 01a0b84d-caa8 and deriving shape is 01a0b84d-ec59, so this issue
	// proves only that the source was located and handed to the parser.
	loaded := parse.LoadedModules(root)
	if len(loaded) != 1 || loaded[0] != "My::Mod" {
		t.Errorf("loaded modules = %v, want [My::Mod]", loaded)
	}
}

// TestParseFileErrorIsTheReadOnly: ParseFile's error is os.ReadFile failing on
// the TOP-LEVEL file. Once bytes exist a tree exists, so nothing found during
// resolution can produce an error.
//
// The parser is total -- Unknown carries its text and an unidentified bareword
// is Call{Resolved:false} -- and three things depend on that: the ratchets
// count Unknown NODES, round-trip is unconditional, and the fuzzer asserts a
// tree comes back for arbitrary bytes.
func TestParseFileErrorIsTheReadOnly(t *testing.T) {
	dir := t.TempDir()

	t.Run("missing file errors", func(t *testing.T) {
		if _, err := parse.ParseFile(filepath.Join(dir, "nope.pl")); err == nil {
			t.Error("want an error for a file that does not exist")
		}
	})

	t.Run("a directory errors", func(t *testing.T) {
		if _, err := parse.ParseFile(dir); err == nil {
			t.Error("want an error when the path is a directory")
		}
	})

	t.Run("an unreadable module does not error", func(t *testing.T) {
		// The module exists and cannot be read. Practically identical to not
		// being found, so it takes the same path: silently unresolved.
		mod := filepath.Join(dir, "Locked.pm")
		if err := os.WriteFile(mod, []byte("1;\n"), 0o000); err != nil {
			t.Fatal(err)
		}
		main := filepath.Join(dir, "reader.pl")
		if err := os.WriteFile(main, []byte("use Locked;\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		root, err := parse.ParseFile(main)
		if err != nil {
			t.Errorf("resolution must not error: %v", err)
		}
		if root == nil {
			t.Fatal("want a tree even when a module cannot be read")
		}
	})
}

// TestUnresolvableImportDeclines: a `use` naming a module that cannot be found
// still PARSES, and calls to its imports stay Call{Resolved:false}.
//
// perl dies here. This parser must not, because parsing text is not running
// it: the `use` statement is keyword, bareword, optional list, semicolon, and
// that is true whether or not the file exists.
//
// Were it a failure, the 986-file T1 ratchet would lose every file naming an
// uninstalled module -- which is most of T1.
func TestUnresolvableImportDeclines(t *testing.T) {
	dir := t.TempDir()
	main := filepath.Join(dir, "main.pl")
	const src = `use Nowhere::To::Be::Found;
subtest 'x' => sub { 1 };
`
	if err := os.WriteFile(main, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	root, err := parse.ParseFile(main)
	if err != nil {
		t.Fatalf("an unfindable module is not an error: %v", err)
	}

	// The use statement itself is complete and correct.
	if !containsKind(root, parse.Use) {
		t.Error("the use statement must parse regardless of the module")
	}

	// Nothing was resolved, and the parser says so rather than guessing.
	if got := parse.LoadedModules(root); len(got) != 0 {
		t.Errorf("loaded = %v, want none: the module does not exist", got)
	}

	// Whatever the import would have provided stays unresolved. Today the
	// parenless `subtest 'x' => sub {}` is an Unknown span -- a bareword
	// followed by a string does not reach the Call path at all -- and
	// turning it into Call{Resolved:false}, then into a resolved list
	// operator, is 01a0b84d-ec59's work.
	//
	// What this issue owns is the NEGATIVE: an unfindable module must not
	// cause a guess. Asserted as "no resolved call to subtest", which holds
	// now and keeps holding once the shape work lands.
	if call := findCall(root, "subtest"); call != nil && call.Resolved {
		t.Error("subtest must not resolve when its module is unfindable")
	}

	// And the file still round-trips, Unknown included.
	if got := leafText(root, []byte(src)); got != src {
		t.Errorf("round-trip broken:\n got %q\nwant %q", got, src)
	}
}

// TestModuleCycleTerminates: `use` is recursive by nature and a cycle is
// ordinary. A visited set keyed by resolved path makes it terminate, and the
// cache means a module used twice is parsed once.
func TestModuleCycleTerminates(t *testing.T) {
	dir := t.TempDir()
	write := func(name, src string) {
		p := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("A.pm", "package A;\nuse B;\n1;\n")
	write("B.pm", "package B;\nuse A;\n1;\n")
	write("main.pl", "use A;\nuse B;\n")

	// If the cycle does not terminate this does not fail, it hangs -- so the
	// assertion that matters is that control returns at all.
	root, err := parse.ParseFile(filepath.Join(dir, "main.pl"))
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}

	loaded := parse.LoadedModules(root)
	seen := map[string]int{}
	for _, m := range loaded {
		seen[m]++
	}
	for _, m := range []string{"A", "B"} {
		if seen[m] != 1 {
			t.Errorf("module %s parsed %d times, want exactly 1", m, seen[m])
		}
	}
}

// TestNoPerlDependency: no perl subprocess is involved, at parse time or build
// time.
//
// Reading the symbol table means libperl has parsed the text; moving that from
// parse time to build time hides the dependency rather than removing it. A
// Perl parser that must run Perl to parse Perl is circular.
//
// Asserted structurally -- the package must not import os/exec -- because an
// assertion on behaviour would pass a build-time generator that had already
// run.
func TestNoPerlDependency(t *testing.T) {
	for _, path := range parseSources(t) {
		src, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if bytes.Contains(src, []byte(`"os/exec"`)) {
			t.Errorf("%s imports os/exec: the parser must never run perl", path)
		}
	}
}

// parseSources lists the package's non-test Go files.
func parseSources(t *testing.T) []string {
	t.Helper()
	all, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, p := range all {
		if strings.HasSuffix(p, "_test.go") {
			continue
		}
		out = append(out, p)
	}
	if len(out) == 0 {
		t.Fatal("no sources found")
	}
	return out
}

// findCall returns the first Call node with the given text.
func findCall(n *parse.Node, text string) *parse.Node {
	if n.Kind == parse.Call && n.Text == text {
		return n
	}
	for _, c := range n.Children {
		if got := findCall(c, text); got != nil {
			return got
		}
	}
	return nil
}
