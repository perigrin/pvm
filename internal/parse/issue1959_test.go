// ABOUTME: Tests for round-2 review findings I9, S14 and S16: a parenthesised
// ABOUTME: literal list, a VERSION before a list, and nested/q() export forms.
package parse

import (
	"slices"
	"testing"
	"testing/fstest"
)

// TestParenLiteralListIsNames: RFC 0001, round-2 finding I9. `use M
// ('x')` is a one-name list, not the explicit empty list `()` --
// finishList (term.go) copies that one item BY VALUE, so the copy keeps
// the item's own Kind -- a Term, not a List -- and `len(Children) == 0`
// alone cannot tell that Term apart from an empty list. The bug was at
// use.go:378 (the declared-import branch) and the same test, at :324 and
// :342, also withheld a library's declared syntax and infix operators
// from a use-list shaped this way.
//
// Measured on 5.42.0: `use Test::More ('no_plan')` installs the same
// set `use Test::More;` does -- not the nothing `use Test::More ();`
// installs.
func TestParenLiteralListIsNames(t *testing.T) {
	bare := importedFrom(t, "use Test::More;\n")
	if len(bare) == 0 {
		t.Fatalf("use Test::More: imports nothing")
	}
	wantImports(t, map[string][]string{
		"use Test::More ('no_plan');\n": bare,
	})
	if got := importedFrom(t, "use Test::More ();\n"); len(got) != 0 {
		t.Errorf("use Test::More (): imports %v, want none", got)
	}
}

// TestParenLiteralListBringsDeclaredSyntax: the same finding, for the
// syntax and infix gates at use.go:324 and :342, which share
// isEmptyParenList with the declared-import branch. My::Lib declares
// the "await" keyword and the "plus" infix operator; TestLibraryTpmEmpty
// ImportBringsNothing already covers `use My::Lib ()` withholding both.
// A single-name paren list is not that -- it brings both in, the same
// as a bare `use My::Lib;` does.
func TestParenLiteralListBringsDeclaredSyntax(t *testing.T) {
	useLibraryFixture(t)
	body := " my ($f, $r); $r = (await $f + 1, 2);\n"
	src := []byte("use My::Lib ('need');" + body)
	root := ParseWithLoader(src, noModules)
	if errs := DeclarationErrors(root); len(errs) > 0 {
		t.Fatalf("declaration errors: %v", errs)
	}
	if imp, ok := Imports(root)["need"]; !ok {
		t.Errorf("'need': not imported: %+v", imp)
	}
	bare := []byte("use My::Lib;" + body)
	got, want := Canon(root, src), Canon(ParseWithLoader(bare, noModules), bare)
	want = want[len("use My::Lib;"):]
	if suffix := "use My::Lib ('need');"; got != suffix+want {
		t.Errorf("use My::Lib ('need'): await parsed as a call, not the declared keyword: %q", got)
	}
	if got, want := lastStatementGrouping(t, "use My::Lib ('need');\n1 plus 2 * 3;"), "(1 plus (2 * 3))"; got != want {
		t.Errorf("use My::Lib ('need'): plus is no operator: got %s, want %s", got, want)
	}
}

// TestUseVersionListIsOneStatement: RFC 0001, round-2 finding S14. `use
// MODULE VERSION LIST` is ONE statement: VERSION is a bare,
// unparenthesised number between the module name and the list, and
// reading "VERSION LIST" as a single expression (no operator joins two
// bare terms) used to take only the number, leaving the list to start a
// second, bare statement of its own.
//
// Measured on 5.42.0: `use POSIX 1.0 qw(floor); floor(1,2)` dies "Usage:
// POSIX::floor(x)" (floor takes one argument); `use Test::More 1.0 ();
// ok();` dies "Undefined subroutine &main::ok" (the explicit empty list
// imports nothing). The split let `floor` escape to the opaque-list
// branch, silencing its arity, and let `use Test::More 1.0` install
// Test::More's defaults as a list of one unmatched number does, giving
// `ok()` a false arity error instead.
func TestUseVersionListIsOneStatement(t *testing.T) {
	for _, src := range []string{
		"use POSIX 1.0 qw(floor);",
		"use Test::More 1.0 ();",
		"use POSIX 1.0;",
		"use POSIX v1.0 qw(floor);",
		"use POSIX 1.2.3 qw(floor);",
		"use Test::More v0.1 ();",
		"use POSIX v1.0;",
	} {
		root := ParseWithLoader([]byte(src), noModules)
		if errs := DeclarationErrors(root); len(errs) > 0 {
			t.Errorf("%q: declaration errors: %v", src, errs)
		}
		if got := len(root.Children); got != 1 {
			t.Errorf("%q: %d statements, want 1", src, got)
		}
		once := Canon(root, []byte(src))
		twice := Canon(Parse([]byte(once)), []byte(once))
		if once != twice {
			t.Errorf("%q: canon is not a fixpoint: %q then %q", src, once, twice)
		}
	}

	if got := importedFrom(t, "use POSIX 1.0 qw(floor);\n"); !slices.Equal(got, []string{"floor"}) {
		t.Errorf("use POSIX 1.0 qw(floor): imports %v, want [floor]", got)
	}
	if got := importedFrom(t, "use Test::More 1.0 ();\n"); len(got) != 0 {
		t.Errorf("use Test::More 1.0 (): imports %v, want none", got)
	}
	bare := importedFrom(t, "use POSIX;\n")
	if got := importedFrom(t, "use POSIX 1.0;\n"); !slices.Equal(got, bare) {
		t.Errorf("use POSIX 1.0: imports %v, want the bare set %v", got, bare)
	}

	// A v-string VERSION, `v1.0` or a bare `1.2.3`, is a VERSION too.
	// Measured on 5.42.0: `use POSIX v1.0 qw(floor); floor(1,2)` and
	// `use POSIX 1.2.3 qw(floor); floor(1,2)` die "Usage: POSIX::floor(x)",
	// `use Test::More v0.1 (); ok();` dies "Undefined subroutine
	// &main::ok", and `use POSIX v1.0;` installs floor, as a bare use does.
	for _, src := range []string{"use POSIX v1.0 qw(floor);\n", "use POSIX 1.2.3 qw(floor);\n"} {
		if got := importedFrom(t, src); !slices.Equal(got, []string{"floor"}) {
			t.Errorf("%q: imports %v, want [floor]", src, got)
		}
	}
	if got := importedFrom(t, "use Test::More v0.1 ();\n"); len(got) != 0 {
		t.Errorf("use Test::More v0.1 (): imports %v, want none", got)
	}
	if got := importedFrom(t, "use POSIX v1.0;\n"); !slices.Equal(got, bare) {
		t.Errorf("use POSIX v1.0: imports %v, want the bare set %v", got, bare)
	}

	// `use if COND, MODULE, LIST` is not `use MODULE VERSION LIST`: a
	// numeric COND stays the list's first element, comma-joined, as
	// useIfTarget already assumes. Measured on 5.42.0, `use if 1,
	// 'strict';` deparses as `use if (1, 'strict')`.
	root := ParseWithLoader([]byte("use if 1, 'strict';"), noModules)
	if errs := DeclarationErrors(root); len(errs) > 0 {
		t.Errorf("use if 1, 'strict': declaration errors: %v", errs)
	}
	if got := len(root.Children); got != 1 {
		t.Errorf("use if 1, 'strict': %d statements, want 1", got)
	}
	if use := root.Children[0].Children[0]; use.Kind != Use || len(use.Children) != 2 {
		t.Errorf("use if 1, 'strict': %+v, want a Use node with module and one list child", use)
	}
}

// TestNestedAndQExportFormsRefused: RFC 0001, round-2 finding S16.
// declarationStatements (the walk behind the "an export mark is read
// only where a declaration ends its own statement" rule) walked only a
// bare block's or a phaser's body, so a sub's own `:export` inside a
// PACKAGE's body, an `if`'s or a `do` block's was never visited as
// nested at all -- not refused, simply never looked at, so it was
// silently unexported with no error naming the problem.
//
// Separately, `use vars` read only `qw(...)` and a quoted string for its
// own argument, missing `q(...)`/`qq(...)`'s generic delimiters, and a
// direct assignment's refusal checked only the `@`/`%` sigils, missing a
// glob assignment, `*EXPORT = [...]`, which reaches the same list a
// `@EXPORT` assignment does.
func TestNestedAndQExportFormsRefused(t *testing.T) {
	saved := declarations
	t.Cleanup(func() { declarations = saved })
	refused := func(t *testing.T, module, src string, want bool) {
		t.Helper()
		declarations = layeredFS{top: fstest.MapFS{"declarations/My/" + module + ".tpm": {Data: []byte(src)}}, base: saved}
		errs := DeclarationErrors(ParseWithLoader([]byte("use My::"+module+";\n"), noModules))
		if got := len(errs) > 0; got != want {
			t.Errorf("%s (%q): errors %v, want refused=%v", module, src, errs, want)
		}
	}

	// A declaration ending its own statement still exports cleanly.
	refused(t, "Nested1", "package My::Nested1;\nsub f :export () Int;\n", false)
	// ... nested in a package's own body.
	refused(t, "Nested2", "package My::Nested2 { sub f :export () Int; }\n", true)
	// ... nested in an if block.
	refused(t, "Nested3", "package My::Nested3;\nif (1) { sub f :export () Int; }\n", true)
	// ... nested in a do block.
	refused(t, "Nested4", "package My::Nested4;\ndo { sub f :export () Int; }\n", true)
	// `use vars` naming an export list through q()/qq(), not only qw()
	// or a plain quoted string.
	refused(t, "Nested5", "package My::Nested5;\nuse vars q(@EXPORT);\nsub f () Int;\n", true)
	refused(t, "Nested6", "package My::Nested6;\nuse vars qq(@EXPORT_OK);\nsub f () Int;\n", true)
	// A glob assigned an export list.
	refused(t, "Nested7", "package My::Nested7;\n*EXPORT = ['f'];\nsub f () Int;\n", true)
	refused(t, "Nested8", "package My::Nested8;\n*EXPORT_OK = ['f'];\nsub f () Int;\n", true)
	// A glob to an unrelated name is not an export list.
	refused(t, "Nested9", "package My::Nested9;\n*Other = ['f'];\nsub f () Int;\n", false)
	// ... nested in an else block, an elsif block, or an expression-level
	// do block: blocks that sit deeper than a statement's direct child.
	refused(t, "Nested10", "package My::Nested10;\nif (1) { 1 } else { sub f :export () Int; }\n", true)
	refused(t, "Nested11", "package My::Nested11;\nif (1) { 1 } elsif (2) { sub f :export () Int; }\n", true)
	refused(t, "Nested12", "package My::Nested12;\nmy $x = do { sub f :export () Int; };\n", true)
}

// TestNamedListInstallsWhatPerlDoes: RFC 0001, round-2 finding I9's
// repro set, run through every module it names. `use M ('x')` -- the
// paren the bug read as the empty list -- installs across Test::More,
// Getopt::Long and Mojolicious::Lite exactly what a bare `use` already
// does, measured on 5.42.0: each module's own `import` reads only its
// own flags or keys and otherwise ignores a use-list's content, so any
// non-empty list, named or not, installs the same defaults.
func TestNamedListInstallsWhatPerlDoes(t *testing.T) {
	cases := map[string]string{
		"use Test::More ('no_plan');\n":               "use Test::More;\n",
		"use Getopt::Long ('GetOptionsFromArray');\n": "",
		"use Mojolicious::Lite ('-signatures');\n":    "use Mojolicious::Lite;\n",
	}
	for src, bareUse := range cases {
		got := importedFrom(t, src)
		if len(got) == 0 {
			t.Errorf("%q: imports nothing", src)
			continue
		}
		if bareUse == "" {
			continue
		}
		if want := importedFrom(t, bareUse); !slices.Equal(got, want) {
			t.Errorf("%q: imports %v, want the bare set %v", src, got, want)
		}
	}
	// Getopt::Long's own import always adds GetOptions once its list is
	// non-empty (measured; the three order variables it also carries
	// over are a known, accepted over-claim, 01a15baf): a named list
	// still resolves GetOptions and the literal name together.
	got := importedFrom(t, "use Getopt::Long ('GetOptionsFromArray');\n")
	for _, want := range []string{"GetOptions", "GetOptionsFromArray"} {
		if !slices.Contains(got, want) {
			t.Errorf("use Getopt::Long ('GetOptionsFromArray'): imports %v, want %s", got, want)
		}
	}
}
