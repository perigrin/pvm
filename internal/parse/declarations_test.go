// ABOUTME: Tests for module declaration files: what a module defines, stated where its source cannot show it.
// ABOUTME: A declaration stands in for the module's source, installed or not.
package parse_test

import (
	"os"
	"path/filepath"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestDeclarationStandsInForSource: a module whose own `import` builds its
// exports -- Moose::Exporter's setup_import_methods, Mojolicious::Lite's
// monkey_patch loop -- defines subs no reading of its source can recover. Its
// declaration file states them, and is read in place of that source whether
// or not the module is installed: here a Moose/Role.pm with its own `import`
// sits on the search path and the declaration still answers. Measured on
// 5.42.0 with Moose 2.4000 and Mojolicious 9.49: every sub below has no
// prototype except Mojolicious::Lite's `group`, which is `&`. PerlOnJava
// unit/moose_role_constant_method.t:21 and
// unit/mojolicious_skipped_closure_lifecycle_regression.t:22.
func TestDeclarationStandsInForSource(t *testing.T) {
	lib := t.TempDir()
	if err := os.MkdirAll(filepath.Join(lib, "Moose"), 0o755); err != nil {
		t.Fatal(err)
	}
	role := "package Moose::Role;\nsub import { my $to = caller; *{\"${to}::requires\"} = sub { 1 } }\n1;\n"
	if err := os.WriteFile(filepath.Join(lib, "Moose", "Role.pm"), []byte(role), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, src := range []string{
		"use Moose::Role; requires 'PROVIDED'; my $x;",
		"package P; use Moose; with qw(A B);",
		"use Mojolicious::Lite; get '/hello' => sub { shift->render(text => 'hello') };",
		"use Mojolicious::Lite; group { get '/x' => sub { 1 } };",
	} {
		n := parse.ParseWithLoader([]byte(src), parse.DirLoader(lib))
		if bad := collectUnknownNodes(n); len(bad) > 0 {
			t.Errorf("%q: refused; got %s", src, shape(n))
		}
	}
}

// TestBeginImportIsUse: `BEGIN { require M; M->import(LIST) }` is what `use M
// LIST` means, and is how a test loads a module it may skip without --
// wrapped in `eval { }` with `plan skip_all` when it fails, which exits
// inside the BEGIN, so the code after it is compiled only when the import
// happened. A literal class name's import is applied as `use` applies it,
// and leaves the sub table complete when the module's exports are all known:
// measured on 5.42.0, after `BEGIN { require Test::More; Test::More->import() }`
// `foo $f;` deparses as `$f->foo`. An invocant that is not a literal name may
// import anything, and the table is no longer complete. PerlOnJava
// unit/mojolicious_skipped_closure_lifecycle_regression.t:12.
func TestBeginImportIsUse(t *testing.T) {
	lib := t.TempDir()
	foo := "package Foo;\nour @EXPORT = qw(two);\nsub two ($$) { }\n1;\n"
	if err := os.WriteFile(filepath.Join(lib, "Foo.pm"), []byte(foo), 0o644); err != nil {
		t.Fatal(err)
	}
	load := parse.DirLoader(lib)
	for _, src := range []string{
		"BEGIN { require Foo; Foo->import; } two 1, 2;",
		"BEGIN { eval { require Foo; Foo->import(qw(two)); 1 } or die; } two 1, 2;",
	} {
		n := parse.ParseWithLoader([]byte(src), load)
		if call := findCall(n, "two"); call == nil || len(collectUnknownNodes(n)) > 0 {
			t.Errorf("%q: want two(1, 2); got %s", src, shape(n))
		}
	}
	if indirectCall(parse.ParseWithLoader([]byte("BEGIN { require Foo; Foo->import(); } my $f; s2 $f;"), load), "s2") == nil {
		t.Errorf("a literal import left the table incomplete")
	}
	if indirectCall(parse.ParseWithLoader([]byte("BEGIN { my $c = 'Foo'; $c->import; } my $f; s2 $f;"), load), "s2") != nil {
		t.Errorf("a computed import left the table complete")
	}
}

// TestDeclaredKeywords: a keyword plugin is a grammar extension, not a sub,
// and Future::AsyncAwait's .xs states its grammar as data that XS::Parse::
// Sublike and XS::Parse::Keyword interpret: `async` is a PREFIX on `sub`,
// `await` takes one XPK_TERMEXPR, and CANCEL takes an XPK_ANONSUB and builds
// a statement (KEYWORD_PLUGIN_STMT). Its declaration file says the same, and
// the keywords exist once the module is imported. Measured on 5.42.0 with
// Future::AsyncAwait 0.71:
//
//   - `await $f + 1` is await($f + 1), and `(await $f, 3)` is a two-element
//     list: a term expression runs down to assignment and stops at a comma
//     (toke.c:14265, parse_termexpr; B::Concise shows the add under await);
//   - `CANCEL { 1 } print "x";` compiles: the statement ends at its block;
//   - `my $c = async sub { 1 };` is an anonymous async sub.
//
// PerlOnJava unit/custom_warning_async_suspend_state.t:27.
func TestDeclaredKeywords(t *testing.T) {
	load := parse.DirLoader(t.TempDir())
	use := "use Future::AsyncAwait; "
	for _, src := range []string{
		"async sub f { my ($g) = @_; await $g; } f(1);",
		"async sub g { my $f; CANCEL { 1 } print \"x\"; }",
		"my $x; my $c = async sub { await $x };",
	} {
		n := parse.ParseWithLoader([]byte(use+src), load)
		if bad := collectUnknownNodes(n); len(bad) > 0 {
			t.Errorf("%q: refused; got %s", src, shape(n))
		}
	}

	n := parse.ParseWithLoader([]byte(use+"async sub f { my $f; my $r = await $f + 1, 2; }"), load)
	decl := firstOfKind(n, parse.Declaration)
	if decl == nil || decl.Text != "async" || len(decl.Children) != 1 || decl.Children[0].Text != "sub" {
		t.Fatalf("want async wrapping sub f; got %s", shape(n))
	}
	if await := findCall(n, "await"); await == nil || len(await.Children) != 1 || await.Children[0].Text != "+" {
		t.Errorf("want await($f + 1); got %s", shape(n))
	}

	n = parse.ParseWithLoader([]byte(use+"async sub g { CANCEL { 1 } print \"x\"; }"), load)
	if c := firstOfKind(n, parse.Conditional); c == nil || c.Text != "CANCEL" {
		t.Errorf("want a CANCEL statement; got %s", shape(n))
	}

	// Without the import there are no keywords: `use Future::AsyncAwait ()`
	// loads it and never calls import.
	for _, src := range []string{"async sub f { 1 }", "use Future::AsyncAwait (); async sub f { 1 }"} {
		n := parse.ParseWithLoader([]byte(src), load)
		if d := firstOfKind(n, parse.Declaration); d != nil && d.Text == "async" {
			t.Errorf("%q: async declared a sub with no import; got %s", src, shape(n))
		}
	}
}
