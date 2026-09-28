// ABOUTME: The rest of T2: use/no, phasers, the 5.38 class syntax, and sort/map/grep.
// ABOUTME: Parse only — what `use feature` turns on is M2's, and BEGIN's compile-time effect is §4.8.3's.

package parse_test

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/lexer"
	"tamarou.com/pvm/internal/parse"
)

// TestUseAndNoForms: every shape that appears in T2, surveyed rather than
// imagined:
//
//	12  use v5.36
//	12  use feature 'class'
//	12  no warnings 'experimental::class'
//	 2  use strict
//	 2  no warnings qw(syntax deprecated)
//	 1  use test_use { () }
func TestUseAndNoForms(t *testing.T) {
	for _, src := range []string{
		"use strict;",
		"use warnings;",
		"no warnings;",
		"use v5.36;",
		"use 5.036;",
		"use feature 'class';",
		"no warnings 'experimental::class';",
		"no warnings qw(syntax deprecated);",
		"use POSIX qw(floor ceil);",
		"use parent -norequire, 'Foo';",
		"require Foo::Bar;",
		// A block argument, which t/comp/use.t actually contains.
		"use test_use { () };",
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q must parse: %v", src, kinds(root))
			continue
		}
		if firstOfKind(root, parse.Use) == nil {
			t.Errorf("%q must produce a Use node: %v", src, kinds(root))
		}
	}
}

// TestPhaserBlocks: BEGIN and friends are named blocks.
//
// That BEGIN runs at compile time, and can declare a sub which changes how a
// LATER line parses, is the undecidability §4.8.3 describes. It is not this
// issue's problem: Call{Resolved:false} already absorbs it.
func TestPhaserBlocks(t *testing.T) {
	for _, src := range []string{
		"BEGIN { 1 }",
		"END { 1 }",
		"CHECK { 1 }",
		"INIT { 1 }",
		"UNITCHECK { 1 }",
		"BEGIN { my $x = 1; }",
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q must parse: %v", src, kinds(root))
			continue
		}
		ph := firstOfKind(root, parse.Phaser)
		if ph == nil {
			t.Errorf("%q must produce a Phaser: %v", src, kinds(root))
			continue
		}
		if firstOfKind(ph, parse.Block) == nil {
			t.Errorf("%q: a phaser holds a block: %v", src, kinds(ph))
		}
	}
}

// TestAdjustThenMethod: `ADJUST` is a phaser, so what follows its block is a
// new statement in either order.
//
// The failure this pins was order-dependent, which is what made it an
// adjacency defect rather than a coverage gap. Measured at the commit before
// the fix, counting Unknown nodes in the class body:
//
//	field $x = 1; ADJUST { $x = 2 } method m { $x }    1
//	field $x = 1; method m { $x } ADJUST { $x = 2 }    0
//	field $x = 1; method m { $x } method n { $x }      0
//	ADJUST { 1 } method m { 1 }                        1
//
// Both orders are valid perl. Measured on 5.42.0, both print 2:
//
//	class Foo { field $x = 1; ADJUST { $x = 2 } method m { $x } }
//	class Foo { field $x = 1; method m { $x } ADJUST { $x = 2 } }
//
// The cause was that `ADJUST` was in neither phaser table, so its `{` was
// classified from XTerm as an anonymous hash. Its `}` then reported a closed
// subscript, the block never closed as a statement, and the declaration that
// followed was swallowed into a term where a statement belonged. Perl agrees
// the brace is a block: two ADJUSTs in a row with no `;` between them run
// both, printing 3 for `ADJUST { $x = 2 } ADJUST { $x++ }` over `field $x =
// 1`.
//
// The reversed order is the regression risk -- it passed before the fix, so
// only the failing rows here are new.
func TestAdjustThenMethod(t *testing.T) {
	for _, src := range []string{
		// The order that refused.
		"class C { field $x = 1; ADJUST { $x = 2 } method m { $x } }",
		"class C { ADJUST { 1 } method m { 1 } }",

		// Not a rule about `method`: any declaration after ADJUST's block.
		"class C { ADJUST { 1 } sub s { 1 } }",
		"class C { ADJUST { 1 } field $y = 1; }",
		"class C { ADJUST { 1 } class D { } }",

		// Two in a row, which perl runs both of.
		"class C { field $x = 1; ADJUST { $x = 2 } ADJUST { $x++ } }",

		// The reversed order, which parsed before the fix and must keep
		// parsing.
		"class C { field $x = 1; method m { $x } ADJUST { $x = 2 } }",

		// The `;` spelling, which parsed before the fix by a different route.
		"class C { ADJUST { 1 }; method m { 1 } }",
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q must parse: %v", src, kinds(root))
		}
	}
}

// TestClassSyntax: 5.38's class, field and method, in 12 T2 files.
//
// All three are feature-gated, which is why the keyword table's
// "identifier where the feature is off" rule matters: t/class/ uses them as
// keywords and older files use `method` as a plain identifier.
func TestClassSyntax(t *testing.T) {
	for _, src := range []string{
		"class Point { }",
		"class Point;",
		"class Point 1.0 { }",
		"class Point { field $x; }",
		"class Point { field $x :param; }",
		"class Point { method zero { 1 } }",
		"class Point { field $x :param; method get { $x } }",
		"class Point :isa(Shape) { }",
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q must parse: %v", src, kinds(root))
		}
	}

	// And the words still work as identifiers where the feature is off.
	for _, src := range []string{
		"my $x = $h{class};",
		"my $x = $h{field};",
		"my $x = $h{method};",
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q uses a class keyword as an identifier: %v", src, kinds(root))
		}
	}
}

// TestSortMapGrepShapes: the three shapes of §4.9.1.
//
//	sort BLOCK LIST
//	sort SUBNAME LIST
//	sort LIST
func TestSortMapGrepShapes(t *testing.T) {
	for _, src := range []string{
		"my @s = sort @list;",
		"my @s = sort { $a <=> $b } @list;",
		"my @s = sort byname @list;",
		"my @m = map { $_ * 2 } @list;",
		"my @m = map $_ * 2, @list;",
		"my @g = grep { $_ > 1 } @list;",
		"my @g = grep $_ > 1, @list;",
		"my @r = reverse sort @list;",
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q must parse: %v", src, kinds(root))
		}
	}
}

// TestClassCorpusHasNoUseGaps: the forms THIS issue owns parse in the 12
// t/class files.
//
// The issue's acceptance criterion originally said "no statement-level
// Unknown" across those files. That was too strong, and saying so is the
// point: 162 Unknown nodes remain, and none of them is a form this issue
// owns. What they are, measured:
//
//	method calls   `Testcase1->new`, `$o->can("h")`
//	ADJUST blocks  a class phaser this issue did not scope -- since closed,
//	               which is the twelve nodes the pin below dropped
//	heredocs       `eval <<'CLASS';`
//	filetests      the `-d 't'` the named-unary issue left
//	multiple attrs `field $s :reader :writer = "..."`
//
// Asserting zero here would make this issue's gate depend on four others,
// which is how a gate stops meaning anything.
//
// So the assertion is narrowed to the forms in scope, and the total is
// ratcheted separately below.
func TestClassCorpusHasNoUseGaps(t *testing.T) {
	files := corpusFilesIn(t, "class")
	if len(files) == 0 {
		t.Skip("no t/class files in the corpus")
	}
	for _, path := range files {
		src := readFile(t, path)
		root := parse.Parse(src)

		// An Unknown must START a statement to count, and its opening TOKEN
		// must be the keyword -- not merely its opening bytes.
		//
		// Matching on a text prefix reads heredoc bodies as code. Three of
		// these files embed Perl in a heredoc:
		//
		//	eval <<'CLASS';          gh23511.t -- `class MyTest {` inside
		//	fresh_perl_is(<<'CODE'   method.t  -- `use feature` inside
		//	ok(eval <<'EOS', ...)    inherit.t -- `use A::B;` inside
		//
		// That text is DATA. The statement holding the heredoc is Unknown for
		// its own reasons and its span covers the body, so any byte-based
		// scan finds keywords that were never statements.
		//
		// The first token is what separates them: a heredoc body is one
		// HeredocBody token and its contents are never tokens of their own.
		var check func(*parse.Node)
		check = func(parent *parse.Node) {
			for _, n := range parent.Children {
				if n.Kind == parse.Unknown {
					if startsWithHeredocBody(src, n) {
						continue
					}
					switch firstWordOf(src, n) {
					case "use", "no", "require", "class", "BEGIN", "field":
						t.Errorf("%s: a form this issue owns is Unknown: %q",
							path, truncate(strings.TrimSpace(
								string(src[n.Start:min(n.End, n.Start+40)])), 40))
					}
					continue
				}
				if n.Kind == parse.Block || n.Kind == parse.SourceFile {
					check(n)
				}
			}
		}
		check(root)
	}
}

// firstWordOf returns the first Word token inside a node, or "" if the node
// does not begin with one.
//
// Re-lexes the node's span rather than reading its bytes, so a keyword that
// appears inside a heredoc body or a string is not mistaken for one that
// opens a statement.
func firstWordOf(src []byte, n *parse.Node) string {
	for _, tok := range lexer.Tokenize(src[n.Start:n.End]) {
		switch tok.Kind {
		// DataSection joins the other three: `isTrivia` in parse.go is the
		// authority and this is the fourth copy of its set. A span holding
		// a data section would otherwise fall to `case lexer.Word` and
		// return `__DATA__` as the statement's first word. Not reachable
		// today -- a section is its own trailing Trivia node -- but the
		// PAAD gate on 01a0dc84 found this copy diverged when the other
		// three were updated.
		case lexer.Whitespace, lexer.Comment, lexer.Pod, lexer.DataSection:
			continue
		case lexer.Word:
			return string(src[n.Start+tok.Start : n.Start+tok.End])
		}
		return ""
	}
	return ""
}

// startsWithHeredocBody reports whether a node's span opens inside a heredoc
// body rather than at a statement.
//
// The parser's Unknown for `eval <<'CLASS';` spans the body too, because the
// body follows the statement in the byte stream. Re-lexing that span finds the
// Perl inside the heredoc, which was never a statement -- so a node whose
// span begins mid-body is skipped.
func startsWithHeredocBody(src []byte, n *parse.Node) bool {
	for _, tok := range lexer.Tokenize(src) {
		if tok.Kind == lexer.HeredocBody && tok.Start <= n.Start && n.Start < tok.End {
			return true
		}
	}
	return false
}

// TestClassCorpusRatchet pins how much of t/class is still Unknown, so the
// number cannot regress quietly while other issues land.
func TestClassCorpusRatchet(t *testing.T) {
	files := corpusFilesIn(t, "class")
	if len(files) == 0 {
		t.Skip("no t/class files in the corpus")
	}
	var unknown int
	for _, path := range files {
		root := parse.Parse(readFile(t, path))
		unknown += len(collect(root, parse.Unknown))
	}

	// Measured after use/no/require, the phasers, class, field and return
	// landed. Update it in the same commit as the change that moves it, in
	// either direction -- a drop is a win worth recording, and a rise is a
	// regression worth seeing.
	//
	// 74 -> 48 when the lexer stopped reading a quote-op keyword as a quote
	// operator where perl reads it as a name. `method y { ... }` was a
	// transliteration delimited by `{`, which swallowed the class body.
	//
	// 45 -> 43 when a declarator kept its subscript: `field $x{k}` no longer
	// splits, and a declaration's initialiser accepts any assignment
	// operator rather than only `=`.
	//
	// 43 -> 40 when a heredoc body became a child of the statement its
	// opener sits in. Three Unknowns, one each in `method.t`, `inherit.t`
	// and `gh22169.t` -- measured, and the same three the T2 shortfall map
	// dropped in this commit.
	//
	// 40 -> 28 when `ADJUST` joined both phaser tables. Twelve nodes across
	// four files, and only files holding an `ADJUST` moved: phasers.t 5 -> 0
	// (13 blocks), destruct.t 7 -> 4 (2), gh22169.t 4 -> 2 (5) and
	// inherit.t 6 -> 4 (4). The other four class files hold none and are
	// unchanged.
	//
	// 28 -> 26 when `undef` joined `namedUnary` (issue 01a0dd43). Both
	// nodes are in `destruct.t`, which writes `undef $notifier;` and
	// `undef $obj;` and nothing else either change reaches. The two fixes
	// are independent and compose on that one file: ADJUST took it 7 -> 4
	// and `undef` takes it 4 -> 2.
	//
	// 26 -> 24 when `no feature "signatures"` began turning the feature OFF.
	// Both Unknowns are in `method.t` and both are on LINE 33, inside the
	// `no feature 'signatures'` block that opens at line 28:
	//
	//	method retnamed ( :$named = 456 ) { return $named; }
	//
	// Before the false path existed the feature could not go off, so that
	// `(...)` was lexed as a signature and `:$named` produced a bare
	// Unknown plus a `trailing_tokens` over the rest of the parens. With the
	// feature off the whole group is one opaque Prototype token and neither
	// refusal happens. Measured by walking the file's Unknowns either side
	// of the change; no other file in t/class moved.
	//
	// Worth recording WHY that is an improvement rather than a coincidence,
	// because the direction is the surprising one: turning signatures off
	// reclassified a signature as a prototype, and a prototype is not lexed
	// at all. A `method` is signatured in perl whether or not the feature is
	// on -- that is exactly what method.t:28 is testing -- so the tree is
	// still not what perl builds here. It simply has two fewer refusals,
	// because an unlexed group cannot refuse.
	//
	// Three fixes, three disjoint file sets: ADJUST moved four files,
	// `undef` moved destruct.t, and this moved method.t. None overlaps
	// another, which is why the deltas add rather than interact.
	//
	// 24 -> 21 when a brace after ANY word got intuit_curly's lookahead
	// instead of only map, grep and sort (issue 01a0d087). All three are in
	// method.t, and they are `method priv { ... }` declarations whose NAME is
	// followed by a brace: gated on the three-word table, that body lexed as
	// a subscript and every declaration after it fell to trailing_tokens.
	//
	// 21 -> 19 with issue 01a0de8b, both nodes in field.t and no other file
	// in t/class moving. Neither is a labelled BLOCK -- they are the two
	// labels inside a `do { ... }`:
	//
	//	field $forwards  = do { goto HERE; HERE: 1 };
	//	field $backwards = do { my $x; HERE: ; goto HERE if !$x++; 2 };
	//
	// One is a label on a plain expression statement, the other a label on an
	// EMPTY statement, and the parser was dropping the label in both cases and
	// then refusing what followed. perl keeps both; Deparse emits `HERE: ;`
	// back verbatim.
	// 19 -> 14 with the leading `::` (issue 01a0de97-3c5d), and the five add
	// up across exactly two files: gh22169.t 2 -> 0, field.t 5 -> 2. Both
	// call `test.pl`'s functions from inside a `class` block, which is the
	// reason perl's suite spells them `::fail(...)` and `::is(...)` -- the
	// leading separator reaches past the lexically-scoped package to
	// `main::`. No other file in t/class moved.
	//
	// 14 -> 13 when a trailing comma stopped being read as an operator
	// missing its operand (issue 01a0dfbd). field.t 2 -> 1: it spells its
	// `ok(eq_array([...], [...]),` assertions across lines with the comma
	// last, which is perl's own suite style everywhere a call takes a list.
	//
	// 13 -> 11 when a block's opening brace began leaving a statement
	// boundary (issue 01a0ea25-4fe7). class/inherit.t 4 -> 2: a bare block
	// as the first member of a class body read as an anonymous hash and lost
	// the declaration after it, which is issue 01a0ddac-bd24's shape.
	const want = 11
	if unknown != want {
		t.Errorf("t/class holds %d Unknown nodes, want %d: update this pin in "+
			"the same commit as the change that moved it", unknown, want)
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
