// ABOUTME: Classifies a source paren by position, and holds psc's reading of each context-changing one to perl's ops.
// ABOUTME: A paren before = or x, after my/our/local/state or after \ makes a list, so dropping it is a tree defect.
package parse_test

import (
	"fmt"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/lexer"
	"tamarou.com/pvm/internal/lower"
	"tamarou.com/pvm/internal/parse"
)

// The positions where a paren changes the ops perl builds. Measured on perl
// 5.44.0 with -MO=Concise:
//
//	my ($x) = @_      aassign, padsv sPRM*/LVINTRO inside ex-list lK
//	my $x = @_        padsv_store, rv2av sK/1
//	($x) x 3          repeat vK/DOLIST,2 over ex-list sK
//	$x x 3            repeat vK/2 over padsv s
//	($x) = f()        aassign, entersub lKS
//	() = f()          aassign, stub lPRM*
//	local ($x) = 1    aassign, ex-rv2sv sKPRM*/LVINTRO
//	our ($x) = @_     aassign;  our $x = @_     sassign
//	state ($x)        padsv vPM/LVINTRO,STATE;  state $x  padsv vM/LVINTRO,STATE
//	\(@a)             refgen vK/1 over padav lPM
//	\@a               srefgen vK/1 over padav lRM
//	\($x) = \$y       aassign;  \my $lex = \$y  refassign
//
// A paren anywhere else only groups, and dropping one that groups is canon's
// business (01a1664d-75d4-76eb-aad2-10b322f366a9), not the tree's.
const (
	beforeAssign = "before ="
	beforeRepeat = "before x"
	afterMy      = "after my"
	afterOur     = "after our"
	afterLocal   = "after local"
	afterState   = "after state"
	afterRef     = `after \`
)

var contextPositions = []string{beforeAssign, beforeRepeat, afterMy, afterOur, afterLocal, afterState, afterRef}

// parenPosition is the context-changing position of the `(` at byte offset
// at in src, or "" when that paren only groups. The token before the paren
// is asked first, so `\(my $lex) = ...` is after `\`.
func parenPosition(src []byte, at int) string {
	var toks []string
	open := -1
	for _, tok := range lexer.Tokenize(src) {
		switch tok.Kind {
		case lexer.Whitespace, lexer.Comment, lexer.Pod, lexer.DataSection:
			continue
		}
		if tok.Start == at {
			open = len(toks)
		}
		toks = append(toks, string(src[tok.Start:tok.End]))
	}
	if open < 0 || toks[open] != "(" {
		return ""
	}
	if open > 0 {
		switch prev := toks[open-1]; prev {
		case `\`:
			return afterRef
		case "my", "our", "local", "state":
			return "after " + prev
		}
	}
	depth := 0
	for i := open; i < len(toks); i++ {
		switch toks[i] {
		case "(":
			depth++
		case ")":
			depth--
		}
		if depth == 0 {
			if i+1 < len(toks) {
				switch toks[i+1] {
				case "=":
					return beforeAssign
				case "x":
					return beforeRepeat
				}
			}
			return ""
		}
	}
	return ""
}

// parenFixture is one context-changing paren: at locates it in src, class is
// its position, and want is the list perl builds there as psc lowers it, a
// fragment of lower.Sexpr.
type parenFixture struct {
	class, src, at, want string
}

// parenFixtures holds the context-changing dropped parens of perl.git t/,
// reduced, and one per position the corpus does not reach.
var parenFixtures = []parenFixture{
	// op/taint.t:66
	{beforeAssign, `($old{$x}) = $ENV{$x} =~ /^(.*)$/;`, `($old`, `Assign Op="=" ListAssign`},
	// op/try.t:195
	{beforeAssign, `(flist) = (4, 5, 6);`, `(flist`, `Assign Op="=" ListAssign`},
	{beforeAssign, `my $x; ($x) = f();`, `($x`, `Assign Op="=" ListAssign`},
	{beforeAssign, `() = f();`, `()`, `Assign Op="=" ListAssign`},
	// op/inccode.t:298
	{beforeRepeat, `@INC = (\&fake_module)x2;`, `(\&`, `Repeat ListRepeat`},
	{beforeRepeat, `my $x; my @r = ($x) x 3;`, `($x)`, `Repeat ListRepeat`},
	{afterMy, `sub f { my ($x) = @_; $x }`, `($x`, `Decl ListAssign`},
	// perl: repeat lKP/DOLIST over padsv sPM/LVINTRO -- the declaration
	// repeated, not a declaration of a repeat.
	{afterMy, `my @r = (my ($x) x 2);`, `($x`, `ListRepeat Left:(Decl Vars:(Var Paren`},
	{afterMy, `my @r = (my ($x, $y) x 2);`, `($x`, `ListRepeat Left:(Decl Vars:(List Paren`},
	// op/for-many.t:474
	{afterMy, `foreach my (\@array) (["A"]) { 1 }`, `(\@`, `Decl Vars:(Ref Paren`},
	{afterOur, `our ($x) = @_;`, `($x`, `Decl Kind=1 ListAssign`},
	{afterOur, `our @r = (our ($x) x 2);`, `(our ($x`, `ListRepeat Left:(Decl Kind=1 Vars:(Var Paren`},
	{afterLocal, `our $x; local ($x) = 1;`, `($x)`, `Decl Kind=3 ListAssign`},
	{afterState, `use feature 'state'; my @r = (state ($x) x 2);`, `($x`, `ListRepeat Left:(Decl Kind=2 Vars:(Var Paren`},
	{afterRef, `my @a; my @r = \(@a);`, `(@a`, `Ref Operand:(Var Paren Sigil="@"`},
	// op/aassign.t:163: perl builds aassign, where `\my $lex = \$y` is refassign.
	{afterRef, `our $pkg10; \(my $lex) = \$pkg10;`, `(my`, `Assign Op="=" ListAssign L:(Ref Operand:(Decl Paren`},
	{afterRef, `our $y; \($x) = \$y;`, `($x`, `Assign Op="=" ListAssign L:(Ref Operand:(Var Paren`},
	// op/sub_lval.t:10
	{afterRef, `${\(bless \$a)};`, `(bless`, `Ref Operand:(Call Paren Name="bless"`},
}

// TestDroppedParenPositionsAreClassified: each fixture's paren is in its
// position, and psc reads a list there where perl builds one.
func TestDroppedParenPositionsAreClassified(t *testing.T) {
	for _, f := range parenFixtures {
		src := []byte(f.src)
		// The fixture's own paren, whichever match its locator's first
		// byte is: `(my ($x)` locates the inner one.
		at := strings.Index(f.src, f.at)
		if at < 0 {
			t.Fatalf("%q: locator %q is not in the source", f.src, f.at)
		}
		at += strings.LastIndex(f.at, "(")
		if got := parenPosition(src, at); got != f.class {
			t.Errorf("%q: the paren at %q is %q, want %q", f.src, f.at, got, f.class)
		}
		if got := lower.Sexpr(lower.Lower(parse.Parse(src))); !strings.Contains(got, f.want) {
			t.Errorf("%q (%s): psc lowers\n\t%s\nwhich lacks perl's list %q", f.src, f.class, got, f.want)
		}
	}
}

// TestDroppedParenAfterBackslashIsContextChanging: `\(@a)` is refgen over
// padav lPM and `\@a` is srefgen, so the paren after `\` is classified as
// context-changing, psc's trees for the two differ, and Faithful does not
// forgive an emission that drops it.
func TestDroppedParenAfterBackslashIsContextChanging(t *testing.T) {
	withParen := []byte(`my @a; my @r = \(@a);`)
	without := []byte(`my @a; my @r = \@a;`)
	if got := parenPosition(withParen, strings.Index(string(withParen), "(@a")); got != afterRef {
		t.Errorf(`the paren of \(@a) is %q, want %q`, got, afterRef)
	}
	lw := lower.Sexpr(lower.Lower(parse.Parse(withParen)))
	lo := lower.Sexpr(lower.Lower(parse.Parse(without)))
	if lw == lo {
		t.Errorf("\\(@a) and \\@a lower alike, but perl builds refgen for one and srefgen for the other:\n\t%s", lw)
	}
	if ok, _ := parse.Faithful(parse.Parse(without), withParen); ok {
		t.Error(`Faithful forgave \@a as a re-emission of \(@a)`)
	}
	if got := parenPosition(without, strings.Index(string(without), "@a;")); got != "" {
		t.Errorf(`\@a has no paren, but its position classified as %q`, got)
	}
}

// fixtureCoverage reports each context-changing position no fixture holds.
func fixtureCoverage(fixtures []parenFixture) error {
	held := map[string]int{}
	for _, f := range fixtures {
		held[f.class]++
	}
	var missing []string
	for _, c := range contextPositions {
		if held[c] == 0 {
			missing = append(missing, c)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("no fixture for %q", missing)
	}
	return nil
}

// TestDroppedParenFixturesCoverEveryContextPosition: the fixtures cannot pass
// empty -- each position holds at least one, and a list missing one fails.
func TestDroppedParenFixturesCoverEveryContextPosition(t *testing.T) {
	if err := fixtureCoverage(parenFixtures); err != nil {
		t.Error(err)
	}
	for _, drop := range contextPositions {
		var rest []parenFixture
		for _, f := range parenFixtures {
			if f.class != drop {
				rest = append(rest, f)
			}
		}
		if err := fixtureCoverage(rest); err == nil || !strings.Contains(err.Error(), drop) {
			t.Errorf("fixtures without %q passed the coverage check: %v", drop, err)
		}
	}
	if err := fixtureCoverage(nil); err == nil {
		t.Error("no fixtures at all passed the coverage check")
	}
}
