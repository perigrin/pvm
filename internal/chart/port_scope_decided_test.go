// ABOUTME: Tests the scope-decided terminals: NOAMP, the *SUB tokens and PRIVATEREF, as superpositions on a bareword.
// ABOUTME: Snippets under testdata/port/scope-decided are recognized; scope-decided-refused ones refuse.

package chart

import (
	"os"
	"slices"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/chart/forest"
	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/parse"
)

// scopeDecidedLines are the perly.y alternatives this family ports or fits. None is left for a completing issue.
var scopeDecidedLines = []int{1046, 1330, 1646, 1678, 1680, 1691}

func TestPortScopeDecided(t *testing.T) {
	if err := grammar.CheckFamilyMarked(scopeDecidedLines, nil); err != nil {
		t.Error(err)
	}
	for _, f := range snippets(t, "scope-decided") {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if v := recognizeSource(t, src); !v.accepted {
			t.Errorf("%s is refused at byte %d (expected %v); perl accepts it", f, v.at, v.expected)
		}
	}
}

// Whether a bareword is a declared sub, and of which prototype, is scope's, so the chart accepts what any
// declaration would let perl accept and refuses what none would. `print FH x 3;`, `new Foo x 3;`,
// `foo bar x 3;` and `foo { a => 1 } x 3;` stay accepted, as the indirect objects left them, and `x 1;` as `foo 1;`
// is: perl accepts each with some declaration (`sub x; x 1;`).
func TestPortScopeDecidedRefusals(t *testing.T) {
	for _, f := range snippets(t, "scope-decided-refused") {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if v := recognizeSource(t, src); v.accepted {
			t.Errorf("%s is accepted; perl refuses it", f)
		}
	}
	// After a package sub y toke.c reads y as tr (only the pad comes before a quote-like), and internal/lexer reads it
	// so too, as an unterminated quote: perl refuses it ("Transliteration pattern not terminated"), at lex here.
	for _, src := range []string{"sub y { 1 } my @a = (y, 1);", "foo(1;", "foo 1 2;", "sub foo; foo 1 2;", "a::();", "Foo:: 1;", "Foo:: $x;", "next CORE(1);"} {
		if v := recognizeSource(t, []byte(src)); v.accepted {
			t.Errorf("%q is accepted; perl refuses it", src)
		}
	}
	for _, src := range []string{
		"foo;", "foo(1);", "sub foo; foo 1, 2;", "my sub bar { 1 } bar();", "&foo;", "f()->{foo};",
		"print FH x 3;", "new Foo x 3;", "foo bar x 3;", "foo { a => 1 } x 3;", "x 1;", // scope's, above
		"new Foo::(1);", "Foo::->new;", "foo => 1;", "x(1);", "$y = do foo;", "$y = do foo 1;",
		"do x(1);", "next x(1);", "goto x(1);", "require x(1);", "sub CORE {1} do CORE(1);", // keywords force_word and yyl_do leave alone
	} {
		if v := recognizeSource(t, []byte(src)); !v.accepted {
			t.Errorf("%q is refused at %d, expecting %v", src, v.at, v.expected)
		}
	}
}

// Known limits, pinned so they are noticed when they end. perl refuses the first four by a toke.c croak while it
// scans the word, which internal/lexer does not make (v0.1 01a1aa1e); a declared sub named x is BAREWORD only past a
// statement start (the shortcut on the scope-decided classes in grammar/perly.go).
func TestPortScopeDecidedKnownLimits(t *testing.T) {
	for _, src := range []string{"my sub Foo::bar {1}", "our sub Foo::bar {1}", "state sub Foo::bar {1}", "CORE::foo(1);", "$y = CORE::;"} {
		if !recognizeSource(t, []byte(src)).accepted {
			t.Errorf("%q is refused; the known limit is gone, so drop the README line", src)
		}
	}
	// perl -c (94e50866) accepts it: a lexical sub named after a quote-like, called bare before a subscript's }, which
	// the table gives no word's reading there (the shortcut in grammar/perly.go's addStatementWords).
	if src := "my sub q {1} my %h; my $y = $h{q\n};"; recognizeSource(t, []byte(src)).accepted {
		t.Errorf("%q is accepted; the known limit is gone, so drop the shortcut and the README line", src)
	}
	// perl -c (94e50866) accepts both: past a statement start x is BAREWORD only where no operator is expected (the
	// shortcut on the scope-decided classes in grammar/perly.go).
	for _, src := range []string{"$y = x { 1 };", "sub x {1} $y = x 1, 2;"} {
		if recognizeSource(t, []byte(src)).accepted {
			t.Errorf("%q is accepted; the known limit is gone, so drop the shortcut and the README line", src)
		}
	}
	// perl refuses these: with its feature on the word is the keyword only. Recognition cannot see a feature, so the
	// word reading stands until resolve collapses it by feature (SCOPEFEAT). The snippets are the families' refusals
	// that a feature-gated word's word reading accepts.
	for _, src := range []string{`use feature "try"; foo(try, 1);`, `use feature "class"; no warnings; my $x = method;`} {
		if !recognizeSource(t, []byte(src)).accepted {
			t.Errorf("%q is refused; the known limit is gone, so drop the README line", src)
		}
	}
	for _, f := range snippets(t, "feature-gated") {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if !recognizeSource(t, src).accepted {
			t.Errorf("%s is refused; the known limit is gone, so move it back to its family's refusals", f)
		}
	}
	// perl accepts it with the try feature off (a call to catch); the PERLY_DOLLAR class that guards a `catch (`
	// variable refuses it (the shortcut in grammar/perly.go).
	if recognizeSource(t, []byte("catch($_);")).accepted {
		t.Errorf("catch($_); is accepted; the known limit is gone, so drop the shortcut and the README line")
	}
	// perl accepts each: catch is a hash key and a method name there. Only `catch (` guards its variable (grammar/perly.go).
	for _, src := range []string{"catch => $_;", "(catch => $_);", "foo(catch => $_);", "x(catch => $Foo::x);", "$p->catch($Foo::x);"} {
		if v := recognizeSource(t, []byte(src)); !v.accepted {
			t.Errorf("%q is refused at %d; perl accepts it", src, v.at)
		}
	}
	// perl accepts both: CORE::qw and CORE::q are quote-likes, which internal/lexer emits as a word and loose tokens
	// (v0.1 01a1ae56), so the stripped word is qw or q with no delimiters.
	for _, src := range []string{"@y = CORE::qw(a b);", "$y = CORE::q{x};"} {
		if recognizeSource(t, []byte(src)).accepted {
			t.Errorf("%q is accepted; the known limit is gone, so drop the README line", src)
		}
	}
}

// A word before => is BAREWORD and one before ( is the zero-width & and the name, PRIVATEREF here (toke.c:8176-8211),
// whatever is declared; only a word they leave undecided carries the scope-decided readings.
func TestPortScopeDecidedSuperpositionBounded(t *testing.T) {
	undecided := []string{"BAREWORD", "FUNC0SUB", "NOAMP", "UNIOPSUB"}
	for _, tc := range []struct {
		src  string
		at   int
		want []string
	}{
		{"$y = foo => 1;", 5, []string{"BAREWORD"}},
		{"$y = foo(1);", 5, []string{"PRIVATEREF"}},
		{"$y = a::(1);", 5, []string{"BAREWORD"}},
		{"$y = Foo:: 1;", 5, []string{"BAREWORD"}},
		// the words toke.c forces to a name (forcedWord), and do before a ( (callForced)
		{"$x->foo;", 4, []string{"BAREWORD", "METHCALL0"}},
		{"&foo;", 1, []string{"BAREWORD", "PRIVATEREF"}},
		{"$y = *foo;", 6, []string{"BAREWORD"}},
		{"sub foo { 1 }", 4, []string{"BAREWORD", "METHCALL0", "PRIVATEREF"}},
		{"use foo;", 4, []string{"BAREWORD"}},
		{"no foo;", 3, []string{"BAREWORD"}},
		{"package foo;", 8, []string{"BAREWORD"}},
		{"format foo =\n.\n", 7, []string{"BAREWORD"}},
		{"sort foo @x;", 5, []string{"BAREWORD"}},
		{"require foo;", 8, []string{"BAREWORD"}},
		{"require foo();", 8, []string{"BAREWORD"}},
		{"sort(foo @x);", 5, []string{"ATTRLIST", "BAREWORD"}},
		{"next foo;", 5, []string{"BAREWORD"}},
		{"next foo();", 5, []string{"BAREWORD"}},
		{"last foo();", 5, []string{"BAREWORD"}},
		{"redo foo();", 5, []string{"BAREWORD"}},
		{"dump foo();", 5, []string{"BAREWORD"}},
		{"goto foo();", 5, []string{"BAREWORD"}},
		{"next CORE(1);", 5, []string{"BAREWORD"}},
		{"last foo;", 5, []string{"BAREWORD"}},
		{"redo foo;", 5, []string{"BAREWORD"}},
		{"dump foo;", 5, []string{"BAREWORD"}},
		{"goto foo;", 5, []string{"BAREWORD"}},
		{"do foo(1);", 3, []string{"BAREWORD"}},
		// a keyword force_word and yyl_do leave alone, and CORE after do, stay calls
		{"do x(1);", 3, []string{"BAREWORD", "PRIVATEREF"}},
		{"next x(1);", 5, []string{"BAREWORD", "PRIVATEREF"}},
		{"require x(1);", 8, []string{"BAREWORD", "PRIVATEREF"}},
		{"sub CORE {1} do CORE(1);", 16, []string{"PRIVATEREF"}},
		// undecided: every reading scope may give it
		{"$y = foo 1;", 5, undecided},
		{"$y = foo;", 5, undecided},
		{"$y = foo + 1;", 5, undecided},
		{"$y = do foo;", 8, undecided},
	} {
		if got := terminalsAt(t, tc.src, tc.at); !slices.Equal(got, tc.want) {
			t.Errorf("%q at %d admits %v, want only %v", tc.src, tc.at, got, tc.want)
		}
	}
	// LSTOPSUB only before a {, where the indirect objects' METHCALL0 is a reading too.
	if got, want := terminalsAt(t, "$y = first { 1 } 2;", 5), []string{"BAREWORD", "FUNC0SUB", "LSTOPSUB", "METHCALL0", "NOAMP", "UNIOPSUB"}; !slices.Equal(got, want) {
		t.Errorf("first before { admits %v, want %v", got, want)
	}
}

// toke.c reads a statement word as a sub's name where a statement starts and as any word elsewhere (toke.c:8313-8322).
// perl -c (94e50866) refuses each statement-start row ("Illegal declaration of subroutine") and accepts its nearby form.
func TestPortScopeDecidedStatementWords(t *testing.T) {
	for _, tc := range []struct{ refused, nearby string }{
		{"BEGIN + 1;", "$y = BEGIN + 1;"}, {"1; END + 1;", "1; $y = END + 1;"},
		{"{ INIT + 1; }", "{ f(INIT + 1); }"}, {"L: CHECK + 1;", "L: f(CHECK) + 1;"},
	} {
		if v := recognizeSource(t, []byte(tc.refused)); v.accepted {
			t.Errorf("%q is accepted; perl refuses it", tc.refused)
		}
		if v := recognizeSource(t, []byte(tc.nearby)); !v.accepted {
			t.Errorf("%q, near %q, is refused at %d", tc.nearby, tc.refused, v.at)
		}
	}
}

// checkParses holds each row's forest.Parse verdict, nothing imported, to the pinned perl -c's (94e50866).
func checkParses(t *testing.T, rows map[string]bool) {
	t.Helper()
	tab, err := committed()
	if err != nil {
		t.Fatal(err)
	}
	for src, accepted := range rows {
		if f, _ := forest.Parse(tab, []byte(src), forest.Scope{}); (f != nil) != accepted {
			t.Errorf("%q: forest.Parse accepts=%v, perl -c accepts=%v", src, f != nil, accepted)
		}
	}
}

// Forms whose verdict waypoint 1 left to scope, each as perl -c (94e50866) gives it. x at a statement start is a word,
// as toke.c's KEY_x makes it wherever no operator is expected (ruling 01a18fa6-ec70), with every reading scope gives
// a word there. After print or say a word is a filehandle, BAREWORD only (toke.c lexes it with PL_expect XREF), and the
// variable after it is lexed at XTERM, leaving XOPERATOR, where bar is "Bareword found where operator expected".
// Recognition accepts the three refused rows, reading the handle as METHCALL0 with $foo its indirect object
// ($foo->STDOUT(bar()), as perl reads `STDOUT $foo bar();`); resolve drops that reading from a word in a filehandle's slot.
func TestScopeDecidedParity(t *testing.T) {
	checkParses(t, map[string]bool{
		"print STDOUT $foo bar();":                  false,
		"use feature 'say'; say STDOUT $foo bar();": false,
		"print STDOUT $foo bar;":                    false,
		"print $foo bar();":                         true, // $foo is the handle: XREF reads a $ before a word as one
	})
	// use builtin imports a word with builtin.c's prototype, as parse.Imports reports it: indexed is "@", so a call; a
	// name it did not import stays undeclared, a word before @a, which perl refuses (perl.git t/perf/opcount.t).
	tab, err := committed()
	if err != nil {
		t.Fatal(err)
	}
	// An import inside a block, sub body or package block ends with it, so a call after the block is a word
	// again; and one there does not shadow M's true after it.
	load := func(name string) ([]byte, bool) {
		return []byte("package M;\nuse Exporter 'import';\nour @EXPORT_OK = qw(true);\nsub true {1}\n1;\n"), name == "M"
	}
	for src, accepted := range map[string]bool{
		"use builtin qw(indexed); my @a; foreach my ($i, $v) (indexed @a) {}":             true,
		"use builtin qw(true); my @a; foreach my ($i, $v) (indexed @a) {}":                false,
		"{ use builtin qw(indexed); } my @a; foreach my ($i, $v) (indexed @a) {}":         false,
		"sub f { use builtin 'indexed'; } my @a; indexed @a;":                             false,
		"if (1) { use builtin qw(indexed); } my @a; for my ($i, $v) (indexed @a) {}":      false,
		"{ use v5.40; } my @a; for my ($i, $v) (indexed @a) {}":                           false,
		"{ use builtin ':5.40'; } my @a; for my ($i, $v) (indexed @a) {}":                 false,
		"package Foo { use builtin qw(indexed); } my @a; for my ($i, $v) (indexed @a) {}": false,
		"use M qw(true); { use builtin qw(true); } my @x = (true 1, 2);":                  true,
	} {
		scope := forest.Scope{Imports: parse.Imports(parse.ParseWithLoader([]byte(src), load))}
		if f, _ := forest.Parse(tab, []byte(src), scope); (f != nil) != accepted {
			t.Errorf("%q: forest.Parse accepts=%v, perl -c accepts=%v", src, f != nil, accepted)
		}
	}
	for _, src := range []string{"x {1};", "sub x {} x {1};", "sub x {1} x 1, 2;", "x { a => 1 } 1;"} {
		if v := recognizeSource(t, []byte(src)); !v.accepted {
			t.Errorf("%q is refused at %d, expecting %v; perl accepts it", src, v.at, v.expected)
		}
	}
	for _, tail := range []string{" {1};", " 1, 2;", " $y;", ";"} {
		if got, want := terminalsAt(t, "x"+tail, 0), terminalsAt(t, "foo"+tail, 0); !slices.Equal(got, want) {
			t.Errorf("x%s at a statement start admits %v; foo there admits %v", tail, got, want)
		}
		at := len(tail) - len(strings.TrimLeft(tail, " "))
		if got, want := terminalsAt(t, "x"+tail, 1+at), terminalsAt(t, "foo"+tail, 3+at); !slices.Equal(got, want) {
			t.Errorf("the token after x in x%s admits %v; after foo it admits %v", tail, got, want)
		}
	}
}

// The refusal of print FILEHANDLE $var word() does not hang on how the handle is spelled, and the forms around it
// perl accepts stay accepted. {$fh} is refused by recognition; the rest by resolve.
func TestPrintFilehandleVarWordForms(t *testing.T) {
	checkParses(t, map[string]bool{
		"print STDERR $foo bar();":            false,
		"print {$fh} $foo bar();":             false,
		"print(STDOUT $foo bar());":           false,
		"printf STDOUT $foo bar();":           false,
		"sub FH {} print FH $foo bar();":      false,
		"sub bar {} print STDOUT $foo bar();": false,
		"print STDOUT $foo $bar;":             false,
		"print main::STDOUT $foo bar();":      false,
		"print main::STDOUT $foo;":            true,
		"print STDOUT $foo;":                  true,
		"print $foo bar();":                   true,
		"print STDOUT $foo, bar();":           true,
		"print STDOUT $foo->bar;":             true,
		"print STDOUT $foo x 3;":              true,
		"STDOUT $foo bar();":                  true, // the method call recognition reads inside the refused rows
		// a feature-gated keyword in the handle's slot is no method either, its feature on or off
		"print try $x bar();":                                  false,
		"sub try {} print try $x bar();":                       false,
		"use feature 'try'; no warnings; print try $x bar();":  false,
		"print try $x;":                                        true,
		"print STDOUT $foo try();":                             false,
		"use feature 'say'; sub try {} say STDOUT $foo try();": false,
	})
}

// Where perl expects an operator x is repetition (KEY_x), and widening x where a statement starts leaves it so.
func TestXOperatorAfterTermUnchanged(t *testing.T) {
	for _, tc := range []struct {
		src  string
		at   int
		want []string
	}{
		{"$s x 3;", 3, []string{"BAREWORD", "MULOP"}},
		{"my @a = (1) x 3;", 12, []string{"ATTRLIST", "BAREWORD", "MULOP"}},
		{"print $a x 3;", 9, []string{"BAREWORD", "MULOP"}},
		{"x x 3;", 2, []string{"BAREWORD", "MULOP"}},
	} {
		if v := recognizeSource(t, []byte(tc.src)); !v.accepted {
			t.Errorf("%q is refused at %d; perl accepts it", tc.src, v.at)
		}
		if got := terminalsAt(t, tc.src, tc.at); !slices.Equal(got, tc.want) {
			t.Errorf("x in %q admits %v, want only %v: no scope-decided word reading after a term", tc.src, got, tc.want)
		}
	}
}

// Known limit: stateStarts reads a do, eval or anonymous-sub block's } as a statement start (the 01a1a422 shortcut on
// stateStarts in grammar/perly.go), so x after one carries the statement-start word readings though perl expects an
// operator there. No verdict changes: the grammar never sets two terms side by side.
func TestXAfterBlockBraceKnownLimit(t *testing.T) {
	src := "$y = do { 1 } x 3;"
	if v := recognizeSource(t, []byte(src)); !v.accepted {
		t.Errorf("%q is refused at %d; perl accepts it", src, v.at)
	}
	if got, want := terminalsAt(t, src, 14), []string{"BAREWORD", "FUNC0SUB", "MULOP", "NOAMP", "UNIOPSUB"}; !slices.Equal(got, want) {
		t.Errorf("x in %q admits %v, not %v; the known limit changed, so update the README line", src, got, want)
	}
}
