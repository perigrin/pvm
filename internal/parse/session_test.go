// ABOUTME: Tests for Session: module facts read once and shared across the parses of one run.
// ABOUTME: Files searching different directories never share what a module said.
package parse_test

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestSessionSharesModules: a corpus run parses hundreds of files that use
// the same modules, and re-reading Test::More and everything it uses for each
// one made the T1 ratchet outlast go test's ten-minute timeout. A Session
// reads a module once per set of search directories: the second file's parse
// reads nothing and resolves the same. A file in another directory may see a
// different module by the same name -- the file's own directory is searched
// first -- so it shares nothing with the first.
func TestSessionSharesModules(t *testing.T) {
	root := t.TempDir()
	write := func(rel, body string) string {
		p := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		return p
	}
	write("a/Foo.pm", "package Foo;\nour @EXPORT = qw(two);\nsub two ($$) { }\n1;\n")
	write("b/Foo.pm", "package Foo;\nsub import { }\n1;\n")
	first := write("a/one.t", "use Foo; two 1, 2;")
	second := write("a/two.t", "use Foo; two 1, 2;")
	other := write("b/three.t", "use Foo; my $f; s2 $f;")

	s := parse.NewSession()
	for i, path := range []string{first, second} {
		n, err := s.ParseFileFrom(path)
		if err != nil {
			t.Fatal(err)
		}
		if findCall(n, "two") == nil || len(collectUnknownNodes(n)) > 0 {
			t.Errorf("%s: want two(1, 2); got %s", path, shape(n))
		}
		if read := parse.LoadedModules(n); i == 1 && slices.Contains(read, "Foo") {
			t.Errorf("%s: read Foo again; loaded %v", path, read)
		}
	}
	n, err := s.ParseFileFrom(other)
	if err != nil {
		t.Fatal(err)
	}
	if indirectCall(n, "s2") != nil {
		t.Errorf("%s: shared a/Foo.pm's facts; b/Foo.pm's import may define s2", other)
	}
}
