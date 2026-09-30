// ABOUTME: `WORD $var ARGS` is `$var->WORD(ARGS)` when WORD is no sub perl could
// ABOUTME: know -- read only while every import this parse saw is fully known.

package parse_test

import (
	"os"
	"path/filepath"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// indirectCall finds the indirect Call named word, or nil.
func indirectCall(n *parse.Node, word string) *parse.Node {
	if n.Kind == parse.Call && n.Indirect && n.Text == word {
		return n
	}
	for _, c := range n.Children {
		if got := indirectCall(c, word); got != nil {
			return got
		}
	}
	return nil
}

// TestIndirectOnScalar: toke.c:8216 makes a word followed by `$` a method
// call on that scalar when the word has no CV, under the indirect feature.
// Measured on 5.42.0 with -MO=Deparse:
//
//	s2 $f;                       $f->s2;
//	Foo $x;                      $x->Foo;
//	$foo = doit $object "FOO";   $foo = $object->doit('FOO');
//	method $obj ("a","b");       $obj->method('a', 'b');
//	is(method $obj, "m");        is($obj->method, 'm');
//	my $x = s2 $f + 1;           my $x = $f->s2 + 1;
//	new threads(sub {1});        'threads'->new(sub {1});
//	sub s5; s5 $f;               s5($f);
//	use v5.36; s2 $f;            syntax error
//	print s2 $f;                 print s2 $f;     (a filehandle)
//
// "Has no CV" is perl's whole symbol table; this parser has it only while
// every import is one it read in full, so the reading waits on a loader and
// stops at the first module it cannot see into.
//
// perl.git t/op/gv.t:680, t/op/ref.t:335, t/op/method.t:71, t/op/getpid.t:33.
func TestIndirectOnScalar(t *testing.T) {
	lib := t.TempDir()
	load := parse.DirLoader(lib)

	for _, c := range []struct {
		src, word, invocant string
		args                bool
	}{
		{`my $f; s2 $f;`, "s2", "$f", false},
		{`my $x; Foo $x;`, "Foo", "$x", false},
		{`my $object; $foo = doit $object "FOO";`, "doit", "$object", true},
		{`my $obj; method $obj ("a","b");`, "method", "$obj", true},
		{`my $obj; is(method $obj, "m");`, "method", "$obj", false},
		{`my $f; my $x = s2 $f + 1;`, "s2", "$f", false},
		{`new threads(sub {1});`, "new", "threads", true},
		{`BEGIN { require strict } my $f; s2 $f;`, "s2", "$f", false},
		// In parens the call keeps its Indirect flag: t/op/method.t:79.
		{`my $obj; is((method $obj ()), "method");`, "method", "$obj", false},
	} {
		root := parse.ParseWithLoader([]byte(c.src), load)
		if got := countUnknown(root); got != 0 {
			t.Errorf("%q: %d Unknown, want 0", c.src, got)
		}
		call := indirectCall(root, c.word)
		if call == nil {
			t.Errorf("%q: no indirect call to %s", c.src, c.word)
			continue
		}
		if got := call.Children[0].Text; got != c.invocant {
			t.Errorf("%q: invocant %q, want %q", c.src, got, c.invocant)
		}
		if got := len(call.Children) > 1; got != c.args {
			t.Errorf("%q: has arguments %v, want %v", c.src, got, c.args)
		}
		canon := parse.Canon(root, []byte(c.src))
		if again := parse.Canon(parse.ParseWithLoader([]byte(canon), load), []byte(canon)); again != canon {
			t.Errorf("%q: canon %q is not a fixpoint, reparses to %q", c.src, canon, again)
		}
	}

	// A sub perl would know, an import this parse cannot see into, or the
	// feature off: never the method reading.
	if err := os.WriteFile(filepath.Join(lib, "Custom.pm"),
		[]byte("package Custom; sub import { 1 } 1;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, src := range []string{
		`sub s5; my $f; s5 $f;`,
		`use NotThere; my $f; s2 $f;`,
		`use Custom; my $f; s2 $f;`,
		`BEGIN { eval "sub s7 {}" } my $f; s7 $f;`,
		`BEGIN { *s6 = sub {} } my $f; s6 $f;`,
		`use v5.36; my $f; s2 $f;`,
		`my $f; print s2 $f;`,
	} {
		if call := indirectCall(parse.ParseWithLoader([]byte(src), load), "s"); call != nil {
			t.Errorf("%q: unexpected indirect call", src)
		}
		root := parse.ParseWithLoader([]byte(src), load)
		for _, w := range []string{"s2", "s5", "s6", "s7"} {
			if indirectCall(root, w) != nil {
				t.Errorf("%q: %s read as a method; perl knows it or cannot say", src, w)
			}
		}
	}

	// A keyword is never the method: `last L` leaves the loop and `require
	// mro` loads it (t/op/loopctl.t:394, t/op/universal.t:311).
	for src, word := range map[string]string{
		`for (;; last L) { }`:      "last",
		`ok(1) or require mro, 1;`: "require",
	} {
		if indirectCall(parse.ParseWithLoader([]byte(src), load), word) != nil {
			t.Errorf("%q: the keyword %s read as a method", src, word)
		}
	}

	// Without a loader nothing is known about imports, so nothing is read.
	if indirectCall(parse.Parse([]byte(`my $f; s2 $f;`)), "s2") != nil {
		t.Errorf("Parse with no loader read `s2 $f` as a method")
	}
}

// TestIndirectUnknownClass: `WORD Bareword` where WORD has no CV is a method
// call on the bareword whatever package it names -- intuit_method's `!cv ||
// gv_stashpvn(...)`. Measured on 5.42.0: `my $x = new Foo "a";` is `my $x =
// 'Foo'->new('a');` with no package Foo anywhere. Read while every import is
// known, as TestIndirectOnScalar's half is.
func TestIndirectUnknownClass(t *testing.T) {
	src := `my $x = new Foo "a";`
	root := parse.ParseWithLoader([]byte(src), parse.DirLoader(t.TempDir()))
	call := indirectCall(root, "new")
	if call == nil {
		t.Fatalf("%q: no indirect call to new; got %s", src, shape(root))
	}
	if got := call.Children[0].Text; got != "Foo" || len(call.Children) != 2 {
		t.Errorf("%q: invocant %q with %d children, want Foo and one argument",
			src, got, len(call.Children))
	}
	if indirectCall(parse.Parse([]byte(src)), "new") != nil {
		t.Errorf("%q: read as a method with no loader", src)
	}
}

// TestUseIfImportFails: `use if COND, MODULE, LIST` imports only when COND
// is true at compile time, which this parser cannot evaluate. By perigrin's
// decision (2026-09-30) the import is assumed to fail: it defines no sub and
// leaves the sub table complete, so a name it would have imported is an
// unknown word -- `catfile $dir, 'aaa'` reads as perl reads an unknown word,
// `$dir->catfile, 'aaa'`. perl.git t/op/coreamp.t:842.
func TestUseIfImportFails(t *testing.T) {
	src := `use if !is_miniperl(), File::Spec::Functions, qw "catfile"; my $dir; my $f = catfile $dir, 'aaa';`
	root := parse.ParseWithLoader([]byte(src), parse.DirLoader(t.TempDir()))
	if got := countUnknown(root); got != 0 {
		t.Errorf("%q: %d Unknown, want 0; got %s", src, got, shape(root))
	}
	call := indirectCall(root, "catfile")
	if call == nil || call.Children[0].Text != "$dir" || len(call.Children) != 1 {
		t.Errorf("%q: want `$dir->catfile` with no arguments; got %s", src, shape(root))
	}
}
