// ABOUTME: Tests the scope-decided terminals: NOAMP, the *SUB tokens and PRIVATEREF, as superpositions on a bareword.
// ABOUTME: Snippets under testdata/port/scope-decided are recognized; scope-decided-refused ones refuse.

package chart

import (
	"os"
	"slices"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
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
	for _, src := range []string{"foo(1;", "foo 1 2;", "sub foo; foo 1 2;", "a::();", "Foo:: 1;", "Foo:: $x;", "next CORE(1);"} {
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
	// perl -c (94e50866) accepts both: past a statement start x is BAREWORD only where no operator is expected (the
	// shortcut on the scope-decided classes in grammar/perly.go).
	for _, src := range []string{"$y = x {1};", "sub x {1} $y = x 1, 2;"} {
		if recognizeSource(t, []byte(src)).accepted {
			t.Errorf("%q is accepted; the known limit is gone, so drop the shortcut and the README line", src)
		}
	}
	// perl refuses it (syntax error); a word before a Variable may read as METHCALL0 after print.
	if !recognizeSource(t, []byte("print STDOUT $foo bar();")).accepted {
		t.Errorf("print STDOUT $foo bar(); is refused; the known limit is gone, so drop the README line")
	}
	// perl accepts both: catch is a hash key and a method name there. The PERLY_DOLLAR class two tokens after catch
	// refuses the first, and the arrow before the word makes it a name (Neighbour.Name), which spares the second (grammar/perly.go).
	for _, src := range []string{"catch => $_;", "(catch => $_);", "foo(catch => $_);", "x(catch => $Foo::x);"} {
		if recognizeSource(t, []byte(src)).accepted {
			t.Errorf("%q is accepted; the known limit is gone, so drop the shortcut and the README line", src)
		}
	}
	if !recognizeSource(t, []byte("$p->catch($Foo::x);")).accepted {
		t.Errorf("$p->catch($Foo::x); is refused; perl accepts it")
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

// Forms perl -c (94e50866) accepts whose refusal waypoint 1 left to scope. x at a statement start is a word, as toke.c's
// KEY_x makes it wherever no operator is expected (ruling 01a18fa6-ec70), with every reading scope gives a word there.
func TestScopeDecidedParity(t *testing.T) {
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
