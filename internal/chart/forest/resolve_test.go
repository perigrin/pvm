// ABOUTME: Tests resolve: the scope-decided words collapsed by the subs declared in the file and imported by its use statements.
// ABOUTME: Each row states the pinned perl's verdict (/opt/perl-94e50866-debug/bin/perl -c), which forest.Parse must give.

package forest

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/chart/lex"
	"tamarou.com/pvm/internal/chart/recognize"
	"tamarou.com/pvm/internal/parse"
)

// testModules are the modules a row's use loads. perl's verdicts for the rows that use M were taken with this M.pm on @INC.
var testModules = map[string]string{
	"M": "package M;\nuse Exporter 'import';\nour @EXPORT_OK = qw(foo one nil first);\nsub foo {}\nsub one ($) {}\nsub nil () { 1 }\nsub first (&@) {}\n1;\n",
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
	})
}

// TestResolveKnownLimits pins resolve's limits, each with its root outside resolve.
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
	// perl -c refuses it; x right after print carries no scope-decided reading in the token table, so resolve
	// cannot read it as a filehandle.
	if f, _ := Parse(tab, []byte("print x x 3;"), Scope{}); f == nil {
		t.Error("print x x 3; is refused; the token-table limit is gone, so move it to TestResolveScopeDecidedWords")
	}
	// perl -c accepts it as 'Foo'->new(*2): after the class toke.c reads *2 as a glob, where internal/lexer emits
	// the 2 as a Number, which no glob's name is.
	if f, _ := Parse(tab, []byte("new Foo * 2;"), Scope{}); f != nil {
		t.Error("new Foo * 2; is accepted; the lexer limit is gone, so move it to TestResolveScopeDecidedWords")
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

// TestResolveKeepsCorpusAccepted: over the Guacamole cases and std5k.pl, resolve refuses no file the pinned perl
// accepts (labels/in-tree.txt) and recognition accepts, with the scope parse.ParseFileFrom gives each, the Guacamole
// stubs on its path as the labels' PERL5LIB.
func TestResolveKeepsCorpusAccepted(t *testing.T) {
	tab := committedTables(t)
	labels, err := os.Open(filepath.Join("..", "testdata", "labels", "in-tree.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer labels.Close()
	stubs := filepath.Join("..", "testdata", "guacamole", "stubs")
	checked := 0
	sc := bufio.NewScanner(labels)
	for sc.Scan() {
		path, label, _ := strings.Cut(sc.Text(), "\t")
		rel, ok := strings.CutPrefix(path, "internal/chart/")
		if label != "accept" || !ok || !strings.HasPrefix(rel, "testdata/guacamole/") && rel != "testdata/std5k.pl" {
			continue
		}
		p := filepath.Join("..", filepath.FromSlash(rel))
		src, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		if built(t, string(src)) == nil {
			continue // recognition refuses it: not resolve's
		}
		root, err := parse.ParseFileFrom(p, stubs)
		if err != nil {
			t.Fatalf("%s: %v", p, err)
		}
		checked++
		if f, ref := Parse(tab, src, Scope{Imports: parse.Imports(root)}); f == nil {
			t.Errorf("%s is refused at byte %d (%s); perl accepts it", p, ref.Byte, ref.Reason)
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	t.Logf("checked %d files", checked)
	if checked < 700 {
		t.Errorf("checked %d files; the labels name over 800 the chart accepts", checked)
	}
}
