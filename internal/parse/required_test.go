// ABOUTME: A literal `require './helper.pl'` contributes the helper's `sub NAME` declarations.
// ABOUTME: Its boundary: computed, missing and nested requires contribute nothing and never fail the parse.

package parse_test

import (
	"os"
	"path/filepath"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestRequiredFileDeclarations: a `require` of a LITERAL relative `.pl` path
// brings that file's `sub NAME` set into scope, so a parenless call to one of
// them parses.
//
// This is not what perl does at compile time, and the difference is worth
// recording because it is the reason this resolution is enrichment rather than
// emulation. `require` is a RUNTIME load, so at the point perl compiles the
// call site the subs do not exist yet, and perl reads the undeclared form as
// INDIRECT OBJECT NOTATION. Measured on 5.42.0:
//
//	$ perl -MO=Deparse -e 'like $@, qr/x/, "msg";'
//	$@->like, qr/x/, '???';
//	$ perl -MO=Deparse -e 'sub like {1} like $@, qr/x/, "msg";'
//	like $@, qr/x/, 'msg';
//
// So `perl -c t/base/lex.t` is `syntax OK` and the failure arrives at RUNTIME:
// "Can't call method \"like\" without a package or object reference". perl
// compiles a method call on `$@` where the source means a function call, and
// discovers the mistake only when it runs. A parser whose job is to say what
// the source MEANS reads the helper instead, which is the tree the source is
// written against -- 394 of the 435 dirty files in perl.git's `t/` are this
// shape, holding 7,211 of its 7,490 Unknown nodes.
func TestRequiredFileDeclarations(t *testing.T) {
	// The boundary rules each get a case, because the safety of this whole
	// resolution is that it declines rather than guesses.
	cases := []struct {
		name string
		// helper is written beside the main file, or not written at all
		// when empty.
		helper string
		main   string
		// want is the Unknown node count the main file must parse to.
		want int
		// loaded is what the helper's path must be recorded as, empty for
		// "nothing was loaded".
		loaded string
	}{
		{
			// RULE 1 and 2: a literal path, resolved beside the requiring
			// file. Measured, this is the commonest spelling in perl.git's
			// `t/`: `require './test.pl'` appears 381 times.
			name:   "a literal ./helper.pl is read",
			helper: "sub ok { 1 }\nsub is ($$;$) { 1 }\n1;\n",
			main:   "require './helper.pl';\nok 1, 'x';\nis 1, 1, 'y';\n",
			want:   0,
			loaded: "./helper.pl",
		},
		{
			// The double-quoted spelling, measured at 78. Same file, and the
			// quote character is not part of the name.
			name:   "the double-quoted spelling is the same file",
			helper: "sub ok { 1 }\n1;\n",
			main:   "require \"./helper.pl\";\nok 1, 'x';\n",
			want:   0,
			loaded: "./helper.pl",
		},
		{
			// No `./` at all, measured at 4. Still relative to a directory
			// rather than to the file, so it takes the same route.
			name:   "a bare helper.pl is relative too",
			helper: "sub ok { 1 }\n1;\n",
			main:   "require 'helper.pl';\nok 1, 'x';\n",
			want:   0,
			loaded: "helper.pl",
		},
		{
			// The prototype travels with the name, because it decides the
			// EXTENT of the call. `($)` is unary, so the comma belongs to
			// the enclosing list and the statement still parses -- the same
			// rule the in-file route follows.
			name:   "the prototype travels with the name",
			helper: "sub one ($) { 1 }\n1;\n",
			main:   "require './helper.pl';\nmy @a = (one 1, 2);\n",
			want:   0,
			loaded: "./helper.pl",
		},
		{
			// RULE 1: a COMPUTED require is opaque. perl cannot know the
			// path at compile time and neither can this, so the call stays
			// unresolved and the statement refuses -- which is the answer it
			// gives today.
			name:   "a computed require resolves nothing",
			helper: "sub ok { 1 }\n1;\n",
			main:   "my $f = './helper.pl';\nrequire $f;\nok 1, 'x';\n",
			want:   1,
			loaded: "",
		},
		{
			// RULE 1: an INTERPOLATED path is computed even though it is
			// spelled as a string. `"./$name.pl"` names a file only at
			// runtime.
			name:   "an interpolated path resolves nothing",
			helper: "sub ok { 1 }\n1;\n",
			main:   "my $n = 'helper';\nrequire \"./$n.pl\";\nok 1, 'x';\n",
			want:   1,
			loaded: "",
		},
		{
			// RULE 5, and it is the one that must not be got wrong: a
			// MISSING helper leaves the call sites Unknown exactly as they
			// are today, and does NOT fail the parse. `t/` files are parsed
			// from many working directories, including by a suite with no
			// perl5 checkout at all.
			name:   "a missing helper does not fail the parse",
			helper: "",
			main:   "require './nosuch.pl';\nok 1, 'x';\n",
			want:   1,
			loaded: "",
		},
		{
			// RULE 3: only `sub NAME`. A variable the helper declares is not
			// a call shape, so `$Level` brings nothing and `level 1, 2`
			// stays unresolved.
			name:   "only sub NAME, not variables",
			helper: "our $Level = 1;\nour @EXPORT = qw(level);\n1;\n",
			main:   "require './helper.pl';\nlevel 1, 2;\n",
			want:   1,
			loaded: "./helper.pl",
		},
		{
			// RULE 4: ONE LEVEL. The helper's own `require` is not followed,
			// so a sub the GRANDCHILD declares is not in scope.
			name:   "one level only",
			helper: "require './deeper.pl';\nsub ok { 1 }\n1;\n",
			main:   "require './helper.pl';\nok 1, 'x';\ndeep 1, 'y';\n",
			want:   1,
			loaded: "./helper.pl",
		},
		{
			// A module name is NOT a path and keeps its own route. The `.pl`
			// SUFFIX is what separates them, and a bareword can never carry
			// one because `.` is not legal in a bareword -- so `Foo::Bar`
			// takes the module spelling, `Foo/Bar.pm`, and resolves nothing
			// here.
			name:   "a module name is not a path",
			helper: "",
			main:   "require Foo::Bar;\n",
			want:   0,
			loaded: "",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dir := t.TempDir()
			if c.helper != "" {
				err := os.WriteFile(filepath.Join(dir, "helper.pl"),
					[]byte(c.helper), 0o644)
				if err != nil {
					t.Fatal(err)
				}
			}
			// The grandchild of the "one level only" case: present on disk,
			// so that what the test proves is the RULE and not the file's
			// absence.
			err := os.WriteFile(filepath.Join(dir, "deeper.pl"),
				[]byte("sub deep { 1 }\n1;\n"), 0o644)
			if err != nil {
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
				t.Errorf("Unknown nodes = %d, want %d\n%s",
					got, c.want, c.main)
			}

			// What was READ is asserted separately from what parsed, because
			// a case can reach the right node count for the wrong reason:
			// the "only sub NAME" and "one level only" cases both READ the
			// helper and still refuse, and a resolution that silently read
			// nothing would pass their counts.
			var want []string
			if c.loaded != "" {
				want = []string{c.loaded}
			}
			got := parse.LoadedModules(root)
			if len(got) != len(want) {
				t.Fatalf("loaded = %v, want %v", got, want)
			}
			for i := range want {
				if got[i] != want[i] {
					t.Errorf("loaded = %v, want %v", got, want)
				}
			}
		})
	}
}

// TestRequiredFileIsResolvedFromTheParseRoot: `require './t/test.pl'` names the
// same file as `require './test.pl'` does, reached from the repository root
// instead of from `t/`.
//
// 17 files in perl.git's `t/` write it that way, because they are run from the
// top of the tree rather than from `t/`. Rule 2's second half: the file's own
// directory first, then the directory the parse was rooted at.
func TestRequiredFileIsResolvedFromTheParseRoot(t *testing.T) {
	dir := t.TempDir()
	tDir := filepath.Join(dir, "t")
	if err := os.MkdirAll(tDir, 0o755); err != nil {
		t.Fatal(err)
	}
	helper := filepath.Join(tDir, "test.pl")
	if err := os.WriteFile(helper, []byte("sub ok { 1 }\n1;\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// The requiring file sits IN `t/`, and names the helper by a path
	// relative to the tree root. Neither `t/t/test.pl` nor a bare
	// `./t/test.pl` beside the file exists, so only the root search finds it.
	main := filepath.Join(tDir, "main.t")
	src := "require './t/test.pl';\nok 1, 'x';\n"
	if err := os.WriteFile(main, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	root, err := parse.ParseFileFrom(main, dir)
	if err != nil {
		t.Fatalf("ParseFileFrom: %v", err)
	}
	if got := countUnknown(root); got != 0 {
		t.Errorf("Unknown nodes = %d, want 0", got)
	}
}
