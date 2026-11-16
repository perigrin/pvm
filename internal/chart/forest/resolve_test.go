// ABOUTME: Tests resolve: the scope-decided words collapsed by the subs declared in the file and imported by its use statements.
// ABOUTME: Each row states the pinned perl's verdict (/opt/perl-94e50866-debug/bin/perl -c), which forest.Parse must give.

package forest

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/chart/lex"
	"tamarou.com/pvm/internal/chart/recognize"
	"tamarou.com/pvm/internal/parse"
)

// testModules are the modules a row's use loads. perl's verdicts for the rows that use M were taken with this M.pm on @INC.
var testModules = map[string]string{
	"M": "package M;\nuse Exporter 'import';\nour @EXPORT_OK = qw(foo one nil first true false);\nsub foo {}\nsub one ($) {}\nsub nil () { 1 }\nsub first (&@) {}\nsub true {1}\nsub false () { 0 }\n1;\n",
}

// scopeOf is src's Scope as a caller builds it: parse.Imports over a tree from parse.ParseWithLoader.
func scopeOf(src string) Scope {
	load := func(name string) ([]byte, bool) {
		m, ok := testModules[name]
		return []byte(m), ok
	}
	return Scope{Imports: parse.Imports(parse.ParseWithLoader([]byte(src), load))}
}

// checkVerdicts runs each row through forest.Parse with its own scope and holds it to perl's verdict.
func checkVerdicts(t *testing.T, rows []verdict) {
	t.Helper()
	tab := committedTables(t)
	for _, r := range rows {
		f, ref := Parse(tab, []byte(r.src), scopeOf(r.src))
		if r.accepted && f == nil {
			t.Errorf("%q is refused (%+v); perl accepts it", r.src, ref)
		}
		if !r.accepted && f != nil {
			t.Errorf("%q is accepted; perl refuses it", r.src)
		}
		if !r.accepted && f == nil && ref.Reason != "scope" && ref.Reason != "grammar" {
			t.Errorf("%q is refused for %q, want scope or grammar", r.src, ref.Reason)
		}
	}
}

// verdict is a source and whether the pinned perl -c accepts it.
type verdict struct {
	src      string
	accepted bool
}

// TestResolveScopeDecidedWords: a word whose readings scope decides keeps the ones toke.c gives it under what is
// declared or imported at its byte (yyl_just_a_word, intuit_method), and the token after it is read as toke.c's
// PL_expect leaves it. The first fourteen rows are the forms waypoint 1 deferred here (README "Indirect objects").
func TestResolveScopeDecidedWords(t *testing.T) {
	checkVerdicts(t, []verdict{
		// A bareword after a list operator that takes a filehandle is one unless a sub of that name is declared, and
		// leaves PL_expect at XTERM, so x after it is a word: "Number found where operator expected".
		{"print FH x 3;", false},
		{"print(FH x 3);", false},
		{"print STDOUT x 3;", false},
		{"sub FH () {1} print FH x 3;", true}, // a constant completes a term, so x is the operator
		{"sub FH () {1} print(FH x 3);", true},
		{"sub STDOUT () {1} print STDOUT x 3;", true},
		{"sub FH {} print FH x 3;", false}, // NOAMP leaves PL_expect at XTERM
		{"sub x {} print FH x 3;", true},   // x is then a call (see TestResolveKnownLimits)
		// intuit_method: an undeclared word before a word that is no keyword and no declared sub is a method, its
		// class forced, and PL_expect XTERM after the class.
		{"new Foo x 3;", false},
		{"sub new {} new Foo x 3;", true},
		{"foo bar x 3;", false},
		{"sub foo {} foo bar x 3;", true},
		// a block after an undeclared word is a method's, and PL_expect is XTERM after it; a hash after a declared sub
		{"foo { a => 1 } x 3;", false},
		{"sub foo {} foo { a => 1 } x 3;", true},
		// declared before, after, forward, by prototype, lexically, and not declared
		{"foo 1;", false},
		{"sub foo {} foo 1;", true},
		{"foo 1; sub foo {}", false},
		{"sub foo { foo 1; }", false}, // in scope only once its declaration completes
		{"sub foo; foo 1;", true},
		{"sub foo; sub foo { foo 1; }", true},
		{"sub foo ($) {} foo 1, 2;", true},
		{"sub foo () {} foo 1;", false},
		{"sub foo () {} foo + 1;", true},
		{"sub foo {} foo { 1 } 2;", false}, // NOAMP leaves XTERM, so the { is a hash
		{"my sub foo {} foo 1;", true},
		{"{ my sub foo {} } foo 1;", false},
		{"our sub foo {} foo 1;", true},
		// imported, with and without a prototype
		{"use M qw(foo); foo 1;", true},
		{"use M qw(one); one 1, 2;", true},
		{"use M qw(nil); nil 1;", false},
		{"use M qw(false); { use builtin qw(false); my @x = (false 1, 2); }", false},
		{"use M qw(nil); nil + 1;", true},
		{"use M qw(first); first { 1 } 2;", true},
		{"use constant FH => 1; print FH x 3;", true},
		// method calls and filehandles perl accepts
		{"foo bar;", true},
		{"sub bar {} foo bar;", false}, // intuit_method finds no class before a declared sub
		{"new Foo, 1;", true},
		{"new Foo { a => 1 };", true}, // the class intuit_method forces is no scope-decided word
		{"print FH 1;", true},
		{"print FH foo;", true},
		{"select(FH);", true},
		{"open(FH x 3);", false},
		// a declared sub before a package the file declares is a method (intuit_method's gv_stashpvn)
		{"sub new {} package Foo; package main; new Foo x 3;", false},
		{"sub new {} package Foo; package main; new Foo 1, 2;", true},
		// subs parse.Imports reports: a BEGIN's glob assignment, and an import into another package
		{"BEGIN { *foo = sub {} } my $x = foo 1;", true},
		{"package Foo; BEGIN { *plan = sub {} } plan 9;", true},
		{"use builtin qw(indexed); my @a; for my ($i, $v) (indexed @a) {}", true}, // perl.git t/perf/opcount.t
		// a glob after a call: its name is no word of scope's, and the { after its * a block or a subscript
		{"sub is {} is *_{ARRAY}, undef;", true},
		// & after a term is the bitwise and (yyl_ampersand, PL_expect XOPERATOR), and a sub's sigil where a term is
		// expected; T1/fcntl_accmode.t's form is the first
		{"use Fcntl qw(O_RDONLY O_ACCMODE); my $x = O_RDONLY & O_ACCMODE;", true},
		{"my $x = FOO & BAR;", true},
		{"my $x = FOO & 2;", true},
		{"sub FOO () {1} my $x = FOO & 2;", true},
		{"my $x = FOO & ;", false},
		{"sub FOO () {1} my $x = FOO & ;", false},
		{"foo &bar;", true},
		{"&$code;", true},
		{"sub foo {} foo &bar;", true},
		{"sub foo {} my $x = foo & 2;", true}, // foo(&2)
		{"sub foo ($) {} my $x = foo & 2;", true},
		{"print FH &bar;", true},
		{"new Foo &bar;", true},
		{"my $x = foo &$c;", true},
		{"return &bar;", true},
	})
}

// TestResolveKnownLimits pins resolve's limits, each beside the pinned perl's verdict, which it does not give.
func TestResolveKnownLimits(t *testing.T) {
	tab := committedTables(t)
	// shortcut (xterm): after a filehandle toke.c reads a declared x as a call, which the token table gives no call
	// reading after a word (README "Scope-decided words", addStatementX), so the x operator stands and the chart
	// accepts it as print(FH x 3): perl's verdict, not perl's grouping.
	src := "sub x {} print FH x 3;"
	if f, _ := Parse(tab, []byte(src), Scope{}); f == nil {
		t.Errorf("%q is refused; perl accepts it", src)
	} else if countSym(t, tab, f, "MULOP") == 0 {
		t.Errorf("%q has no x operator left; the shortcut is gone, so drop this row", src)
	}
	for _, c := range []verdict{
		// x right after print carries no scope-decided reading in the token table, so resolve cannot read it as a
		// filehandle; perl refuses it.
		{"print x x 3;", true},
		// shortcut (lookup): a qualified name neither the file nor its imports declare keeps every reading; perl
		// refuses both.
		{"Foo::bar 1;", true},
		{"package A; sub foo {} package main; A::bar 1;", true},
		// shortcut (word): a package created outside the file's package statements, here by naming Foo::bar, makes
		// new a method to perl, which refuses it.
		{"sub new {} Foo::bar(); new Foo x 3;", true},
		// toke.c leaves XTERM after a call, where < opens a <> and ** a glob's * twice and ? an old pattern match,
		// which the token table gives no reading after a word; perl refuses all three, the last for its arguments.
		{"sub foo {} foo < 3;", true},
		{"sub foo {} my $x = foo ** 2;", true},
		{"sub foo ($) {} foo ? 1 : 2;", true},
		// toke.c reads * before a Number as a glob wherever it expects a term (*2), where internal/lexer emits the
		// 2 a Number, which no glob's name is; perl accepts all three.
		{"new Foo * 2;", false},
		{"sub foo {} foo * 2;", false},
		{"print STDOUT * 2;", false},
		// shortcut (lookup): parse.Imports reports nothing for use subs, so the names it declares are undeclared to
		// resolve; perl accepts it.
		{"use subs qw(foo); foo 1;", false},
		// shortcut (lookup): parse.Imports reports only a file-scope use builtin, so a name one imports inside a block
		// is undeclared there too; perl accepts it.
		{"{ use builtin 'indexed'; my @a; for my ($i, $v) (indexed @a) {} }", false},
		// and where a module imports the name at file scope, the chart reads it there with the module's prototype,
		// where perl reads builtin's; perl refuses it ("Number found where operator expected").
		{"use M qw(true); { use builtin qw(true); my @x = (true 1, 2); }", true},
		// shortcut (lookup): a sub the file declares is read before an import, so an import of its name, by use builtin
		// or by a module, does not shadow it; perl refuses each ("Number found where operator expected").
		{"sub true {1} use builtin qw(true); my @x = (true 1, 2);", true},
		{"sub true {1} { use builtin qw(true); my @x = (true 1, 2); }", true},
		{"sub nil {} use M qw(nil); my @x = (nil 1);", true},
		// shortcut (word): resolve does not know the feature state, so a feature-gated keyword's readings all stand
		// and any one before a term is accepted with its feature on or off; perl refuses each, the word there
		// followed by 1, no operator.
		{"given 1;", true},
		{"state 1;", true},
		{"method 1;", true},
		{"try 1;", true},
	} {
		if f, _ := Parse(tab, []byte(c.src), scopeOf(c.src)); (f != nil) != c.accepted {
			t.Errorf("%q: accepted %v, pinned %v; the limit has moved, so move the row to the rows perl's verdict holds",
				c.src, f != nil, c.accepted)
		}
	}
}

// countSym is how many leaves named sym the derivations of f hold.
func countSym(t *testing.T, tab *grammar.Tables, f *Forest, sym string) int {
	t.Helper()
	trees, _ := f.Derivations(f.Root, 64)
	n := 0
	for _, tr := range trees {
		for _, l := range leaves(tr) {
			if tab.Symbols[l.Sym].Name == sym {
				n++
			}
		}
	}
	return n
}

// TestResolveImportsAreFileWide pins the recorded limit: parse.Imports does not say which use brought a name in, so
// a call written before its use reads as imported (perigrin, 2026-11-10, "File-wide for now"). perl -c refuses
// the first row ("Number found where operator expected") and accepts the second.
func TestResolveImportsAreFileWide(t *testing.T) {
	tab := committedTables(t)
	for _, src := range []string{"foo 1;\nuse M qw(foo);\n", "use M qw(foo);\nfoo 1;\n"} {
		if f, ref := Parse(tab, []byte(src), scopeOf(src)); f == nil {
			t.Errorf("%q is refused (%+v): imports are file-wide, so the call before the use reads as imported", src, ref)
		}
	}
	if f, _ := Parse(tab, []byte("foo 1;\n"), Scope{}); f != nil {
		t.Error("foo 1; is accepted with nothing imported; the row above would pin nothing")
	}
}

// TestResolveKeepsKeyDecidedWords: a word whose readings its key already decided (after ->, in a hash subscript,
// before =>) is not collapsed, a declaration of that name or not: Parse's forest is Build's.
func TestResolveKeepsKeyDecidedWords(t *testing.T) {
	tab := committedTables(t)
	for _, src := range []string{
		"$x->foo;", "$h{foo};", "%h = (foo => 1);", "$x->foo x 3;",
		"sub foo {} $x->foo;", "sub foo {} $h{foo} x 3;", "sub foo {} %h = (foo => 1);", "sub foo {} $x->foo x 3;",
	} {
		f, ref := Parse(tab, []byte(src), Scope{})
		if f == nil {
			t.Errorf("%q is refused (%+v); perl accepts it", src, ref)
			continue
		}
		if got, want := dump(f), dump(built(t, src)); got != want {
			t.Errorf("resolve changed the forest of %q:\n%s\nwant:\n%s", src, got, want)
		}
	}
}

// built is src's forest as Build makes it from the chart, before resolve.
func built(t *testing.T, src string) *Forest {
	t.Helper()
	tab := committedTables(t)
	s, _ := lex.Lex([]byte(src), tab.Tokens)
	l := lex.Linear(s)
	c, err := recognize.Recognize(recognize.New(tab, recognize.Options{}), l)
	if err != nil {
		t.Fatal(err)
	}
	return Build(tab, c, l)
}

// TestResolvePackageScopedSubs: a sub is declared in the package its declaration is in, so a bare name finds only
// the current package's, and a qualified one its own. Each verdict is perl -c's.
func TestResolvePackageScopedSubs(t *testing.T) {
	checkVerdicts(t, []verdict{
		{"package A; sub foo {} package B; foo 1;", false},
		{"package A { sub foo {} } foo 1;", false},
		{"package A; sub foo {} package B; A::foo 1;", true},
		{"package A { sub foo {} foo 1; }", true},
		{"sub foo {} { package A; } foo 1;", true},
		{"package A; sub foo {} package main; package A; foo 1;", true},
		{"sub main::foo {} foo 1;", true},
		{"sub foo {} ::foo 1;", true},
		{"sub A::foo {} package A; foo 1;", true},
	})
}

// TestResolveEmptyScope: with no imports, nil or empty, resolve runs without panic and the file's own declarations
// still act.
func TestResolveEmptyScope(t *testing.T) {
	tab := committedTables(t)
	for _, scope := range []Scope{{}, {Imports: map[string]parse.Import{}}} {
		for _, c := range []verdict{{"", true}, {"print 1;", true}, {"sub foo {} foo 1;", true}, {"foo 1;", false}} {
			if f, _ := Parse(tab, []byte(c.src), scope); (f != nil) != c.accepted {
				t.Errorf("%q with %+v: accepted %v, want %v", c.src, scope, f != nil, c.accepted)
			}
		}
	}
}

// TestResolveKeepsCorpusAccepted: resolve refuses no labelled file that the pinned perl (labels/in-tree.txt,
// labels/t1.txt), recursive descent following use statements as the ratchet's verdict side does, and recognition all
// accept: the Guacamole cases, std5k.pl, edit5k.pl and T1 (T1_CORPUS, else the in-tree copy). Each file's scope is
// parse.Imports over the same use-resolving tree, the labels' PERL5LIB on its path.
func TestResolveKeepsCorpusAccepted(t *testing.T) {
	tab := committedTables(t)
	repo := filepath.Join("..", "..", "..")
	t1 := os.Getenv("T1_CORPUS")
	if t1 == "" {
		t1 = filepath.Join(repo, "testdata", "perlonjava", "unit")
	}
	stubs := filepath.Join(repo, "internal", "chart", "testdata", "guacamole", "stubs")
	checked := 0
	for _, c := range []struct{ labels, root, lib string }{{"in-tree.txt", repo, stubs}, {"t1.txt", t1, ""}} {
		for _, rel := range acceptedLabels(t, filepath.Join("..", "testdata", "labels", c.labels)) {
			p := filepath.Join(c.root, filepath.FromSlash(rel))
			src, err := os.ReadFile(p)
			if err != nil {
				t.Fatalf("%s (%s names it): %v", p, c.labels, err)
			}
			if built(t, string(src)) == nil {
				continue // recognition refuses it: not resolve's
			}
			dirs := []string{filepath.Dir(p)}
			if c.lib != "" {
				dirs = append(dirs, c.lib)
			}
			root := parse.ParseWithLoader(src, parse.DirLoader(dirs...))
			if refusedAnywhere(root) {
				continue // recursive descent refuses it
			}
			checked++
			if f, ref := Parse(tab, src, Scope{Imports: parse.Imports(root)}); f == nil {
				t.Errorf("%s is refused at byte %d (%s); perl and recursive descent accept it", p, ref.Byte, ref.Reason)
			}
		}
	}
	t.Logf("checked %d files", checked)
	if checked < 1400 {
		t.Errorf("checked %d files; the in-tree and T1 labels name over 1,400 that perl, recursive descent and recognition accept", checked)
	}
}

// acceptedLabels is the paths a label file says the pinned perl accepts.
func acceptedLabels(t *testing.T, path string) []string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, line := range strings.Split(string(data), "\n") {
		if name, label, _ := strings.Cut(line, "\t"); label == "accept" {
			out = append(out, name)
		}
	}
	return out
}

// refusedAnywhere is whether any node of n's tree is a refusal, recursive descent's verdict.
func refusedAnywhere(n *parse.Node) bool {
	return n.Refused() || slices.ContainsFunc(n.Children, refusedAnywhere)
}
