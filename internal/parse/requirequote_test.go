// ABOUTME: `require q(./helper.pl)` reads the helper the way `require './helper.pl'` does.
// ABOUTME: 61 perl.git t/ files spell their test.pl require with `q`, and lost every parenless call.
package parse_test

import (
	"os"
	"path/filepath"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestRequireQuoteForm holds the q-quoted spelling of a literal require.
// mro/package_aliases.t opens with `require q(./test.pl);`, and its first
// PARENLESS call to a test.pl sub -- `fresh_perl_is q~...~, ...` -- refused
// because nothing had declared it. Calls with parens parsed either way, which
// is why the file looked mostly clean.
func TestRequireQuoteForm(t *testing.T) {
	for _, main := range []string{
		"require q(./helper.pl);\nok 1, 'x';\n",
		"require q{./helper.pl};\nok 1, 'x';\n",
		"require qq(./helper.pl);\nok 1, 'x';\n",
	} {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "helper.pl"),
			[]byte("sub ok { 1 }\n1;\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(dir, "main.t")
		if err := os.WriteFile(path, []byte(main), 0o644); err != nil {
			t.Fatal(err)
		}
		root, err := parse.ParseFile(path)
		if err != nil {
			t.Fatalf("ParseFile: %v", err)
		}
		if got := countUnknown(root); got != 0 {
			t.Errorf("%q: %d Unknown, want 0 -- the helper's `ok` was not declared", main, got)
		}
	}
}
