// ABOUTME: A `require './helper.pl'` inside BEGIN declares the helper's subs to the rest of the parse.
// ABOUTME: A run-time require declares nothing, because perl has compiled the calls before it runs.

package parse_test

import (
	"os"
	"path/filepath"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestBeginRequireDeclaresSubs: a file a BEGIN block requires declares its
// subs, with their prototypes, before the code after the block compiles; a
// file required at run time declares nothing at parse time.
//
// Measured on 5.42.0, with helper.pl declaring `sub is ($$@)` and `sub plan`:
//
//	BEGIN { require "./helper.pl" }  is 1, 2, 3;   Deparse: &is(1, 2, 3)
//	require "./helper.pl";           is 1, 2, 3;   syntax error, "Do you need
//	                                               to predeclare "is"?"
//	require "./helper.pl";           is $x, 2, 3;  Deparse: $x->is, ...
//
// The require is a statement like any other; only BEGIN runs it before the
// rest of the file is compiled. perl's own suite writes
// `BEGIN { chdir 't' if -d 't'; require './test.pl'; ... }` for exactly that
// reason, and resolves the relative path against the working directory, `t/`.
// A BEGIN-time require of a file that is not there dies in perl ("Can't locate
// ./nope.pl ... BEGIN failed"); here it declares nothing and the parse goes on.
func TestBeginRequireDeclaresSubs(t *testing.T) {
	helper := "sub is ($$@) { 1 }\nsub plan { 1 }\nsub like ($$@) { 1 }\n1;\n"

	cases := []struct {
		name   string
		main   string
		want   int
		loaded string
	}{
		{
			name:   "a BEGIN-time require declares the helper's subs",
			main:   "BEGIN { require './helper.pl' }\nplan tests => 4;\nis 1, 2, 'x';\nlike $@, qr/x/, 'y';\n",
			want:   0,
			loaded: "./helper.pl",
		},
		{
			// perl's own suite: the require is one statement among several,
			// and nested blocks inside BEGIN run at BEGIN time too.
			name: "inside nested blocks in BEGIN, as perl's suite writes it",
			main: "BEGIN {\n    chdir 't' if -d 't';\n    if (1) { require './helper.pl'; }\n}\n" +
				"is 1, 2, 'x';\n",
			want:   0,
			loaded: "./helper.pl",
		},
		{
			// The negative: perl refuses `is 1, 2` after a run-time require,
			// because `is` is not yet a sub when the call compiles.
			name:   "a run-time require declares nothing",
			main:   "require './helper.pl';\nis 1, 2, 'x';\n",
			want:   1,
			loaded: "",
		},
		{
			name:   "a missing file in BEGIN declares nothing and fails nothing",
			main:   "BEGIN { require './nosuch.pl' }\nis 1, 2, 'x';\n",
			want:   1,
			loaded: "",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, "helper.pl"), []byte(helper), 0o644); err != nil {
				t.Fatal(err)
			}
			main := filepath.Join(dir, "main.t")
			if err := os.WriteFile(main, []byte(c.main), 0o644); err != nil {
				t.Fatal(err)
			}

			root, err := parse.ParseFile(main)
			if err != nil {
				t.Fatalf("ParseFile: %v", err)
			}
			if got := countUnknown(root); got != c.want {
				t.Errorf("Unknown nodes = %d, want %d\n%s", got, c.want, c.main)
			}
			var want []string
			if c.loaded != "" {
				want = []string{c.loaded}
			}
			if got := parse.LoadedModules(root); len(got) != len(want) || (len(want) == 1 && got[0] != want[0]) {
				t.Errorf("loaded = %v, want %v", got, want)
			}
		})
	}
}
