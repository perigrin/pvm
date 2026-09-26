// ABOUTME: Canon emits Perl from a tree, parenthesised the way the TREE groups.
// ABOUTME: The parens it writes are what makes the comparison against source mean anything.

package parse_test

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/lexer"
	"tamarou.com/pvm/internal/parse"
)

// TestCanonEmitsTreeParens: the emission's grouping comes from the tree.
//
// An emitter that prints each leaf's original text in order is SourceText
// with extra steps -- it round-trips by construction and proves nothing. The
// parens below are the whole point: Canon must write one wherever a child
// binds looser than its parent, whether or not the source had it, and must
// NOT write one where the tree's own shape already says the grouping.
func TestCanonEmitsTreeParens(t *testing.T) {
	for _, tc := range []struct{ src, want string }{
		// Precedence alone: * binds tighter, so the tree needs no paren
		// and Canon must not invent one.
		{"1 + 2 * 3;", "1 + 2 * 3;"},

		// The source's paren changed the grouping, so the tree holds it and
		// Canon must write it back.
		{"(1 + 2) * 3;", "(1 + 2) * 3;"},

		// A paren the source had but the tree did not need. Canon prints
		// the TREE, so the redundant paren is gone.
		{"1 + (2 * 3);", "1 + 2 * 3;"},

		// Left associative: the right operand at equal power needs a paren,
		// the left one does not. `1 - 2 - 3` is (1-2)-3.
		{"1 - 2 - 3;", "1 - 2 - 3;"},
		{"1 - (2 - 3);", "1 - (2 - 3);"},

		// Right associative: the mirror image. `2 ** 3 ** 2` is 2**(3**2),
		// so the right operand needs no paren and the left one does.
		{"2 ** 3 ** 2;", "2 ** 3 ** 2;"},
		{"(2 ** 3) ** 2;", "(2 ** 3) ** 2;"},
	} {
		got := parse.Canon(parse.Parse([]byte(tc.src)), []byte(tc.src))
		if strings.TrimSpace(got) != tc.want {
			t.Errorf("Canon(parse(%q)):\n  got  %q\n  want %q", tc.src, strings.TrimSpace(got), tc.want)
		}
	}
}

// TestCanonCatchesMisgrouping: a tree that groups differently from its source
// re-emits differently, and round-trip does not notice.
//
// The tree is built by hand rather than parsed, because the parser gets this
// case right -- and a check whose failure mode cannot be produced is not a
// check. `print (1+2)*3` is the case chapter 7 names: perl reads it as
// `print(1+2) * 3`, and a parser that read it as `print((1+2)*3)` has the
// same leaves in the same order.
//
// That is why this test asserts BOTH halves. SourceText agrees with the
// wrong tree; Canon does not. The two checks are disjoint, not weaker and
// stronger versions of one thing.
func TestCanonCatchesMisgrouping(t *testing.T) {
	src := []byte("print (1+2)*3;")

	// The wrong tree: print swallowing the whole product.
	//
	//	statement
	//	  call print          span: the whole expression
	//	    binary *
	//	      binary +
	//	        term 1
	//	        term 2
	//	      term 3
	mul := &parse.Node{Kind: parse.Binary, Text: "*", Start: 6, End: 13, Children: []*parse.Node{
		{Kind: parse.Binary, Text: "+", Start: 7, End: 10, Children: []*parse.Node{
			{Kind: parse.Term, Text: "1", Start: 7, End: 8},
			{Kind: parse.Term, Text: "2", Start: 9, End: 10},
		}},
		{Kind: parse.Term, Text: "3", Start: 12, End: 13},
	}}
	wrong := &parse.Node{Kind: parse.SourceFile, Start: 0, End: len(src), Children: []*parse.Node{
		{Kind: parse.Statement, Start: 0, End: len(src), Children: []*parse.Node{
			{Kind: parse.Call, Text: "print", Resolved: true, Start: 0, End: 13, Children: []*parse.Node{mul}},
		}},
	}}

	// Round-trip is blind to it: same leaves, same order, every byte back.
	if got := wrong.SourceText(src); got != string(src) {
		t.Fatalf("the wrong tree must still round-trip, or it is not the case "+
			"this test is about:\n  got  %q\n  want %q", got, string(src))
	}

	// Canon is not blind to it, because it writes its OWN grouping.
	got := strings.TrimSpace(parse.Canon(wrong, src))
	right := strings.TrimSpace(parse.Canon(parse.Parse(src), src))
	if got == right {
		t.Errorf("Canon does not distinguish the mis-grouped tree from the correct one; "+
			"both emitted %q", got)
	}
}

// TestCanonVisitsInteriorNodes: no interior node emits its own source text.
//
// A kind that falls through to `default` writes SourceText -- which is the
// thing this emitter exists NOT to be. Its children are never visited, so a
// misgrouped subtree beneath it is invisible to every check in this file:
//
//	my $x = 1 + 2 * 3;   with ((1+2)*3) under the Declaration
//	emits                my $x = 1 + 2 * 3;      byte-identical
//
// Measured before this test existed: ten kinds fell through, covering 81.4%
// of T1 bytes. The assertion is structural rather than a byte count, because
// a percentage moves with the corpus and the property does not.
func TestCanonVisitsInteriorNodes(t *testing.T) {
	// A misgrouped subtree under each interior kind must change the emission.
	// If the kind emits raw source, the wrong tree and the right tree produce
	// the same bytes and this fails.
	misgrouped := func() *parse.Node {
		return &parse.Node{Kind: parse.Binary, Text: "*", Start: 8, End: 17, Children: []*parse.Node{
			{Kind: parse.Binary, Text: "+", Start: 8, End: 13, Children: []*parse.Node{
				{Kind: parse.Term, Text: "1", Start: 8, End: 9},
				{Kind: parse.Term, Text: "2", Start: 12, End: 13},
			}},
			{Kind: parse.Term, Text: "3", Start: 16, End: 17},
		}}
	}
	correct := func() *parse.Node {
		return &parse.Node{Kind: parse.Binary, Text: "+", Start: 8, End: 17, Children: []*parse.Node{
			{Kind: parse.Term, Text: "1", Start: 8, End: 9},
			{Kind: parse.Binary, Text: "*", Start: 12, End: 17, Children: []*parse.Node{
				{Kind: parse.Term, Text: "2", Start: 12, End: 13},
				{Kind: parse.Term, Text: "3", Start: 16, End: 17},
			}},
		}}
	}

	src := []byte("my $x = 1 + 2 * 3;")
	build := func(expr *parse.Node) *parse.Node {
		return &parse.Node{Kind: parse.SourceFile, Start: 0, End: len(src), Children: []*parse.Node{
			{Kind: parse.Statement, Start: 0, End: len(src), Children: []*parse.Node{
				{Kind: parse.Declaration, Text: "my", Start: 0, End: 17, Children: []*parse.Node{
					{Kind: parse.Term, Text: "$x", Start: 3, End: 5},
					expr,
				}},
			}},
		}}
	}

	wrong := parse.Canon(build(misgrouped()), src)
	right := parse.Canon(build(correct()), src)
	if wrong == right {
		t.Errorf("a misgrouped subtree under a Declaration emits identically to "+
			"the correct one (%q) -- the kind is emitting its own source text "+
			"and never visiting its children", wrong)
	}
}

// TestCanonEmitsValidPerl: the emission is Perl that perl itself accepts.
//
// The fixpoint cannot see this class of defect, which is why it is asserted
// separately. `print($fh"x")` re-parses to itself and is stable -- and is not
// valid Perl. A filehandle slot takes no comma and a space is the only thing
// separating it from the list, so dropping the space produces text that
// satisfies every other check in this file while being unparseable.
//
// Checked by structure rather than by running perl: the ratchet reads .t
// files as text and must not acquire an interpreter dependency for one case.
func TestCanonEmitsValidPerl(t *testing.T) {
	for _, tc := range []struct{ src, want string }{
		{`print $fh "x";`, `print($fh "x");`},
		{`print STDERR "x";`, `print(STDERR "x");`},
		{`print "x";`, `print("x");`},

		// A nonassoc operator needs a paren on BOTH sides at equal power.
		// `1 .. 2 .. 3` is a syntax error, so dropping it emits one.
		{`(1 .. 2) .. 3;`, `(1 .. 2) .. 3;`},

		// A statement ending in `}` still takes its semicolon unless it is a
		// BLOCK form. `$x = sub {1;}$y = 2;` does not parse.
		{`$x = sub { 1 }; $y = 2;`, `$x = sub {1;};$y = 2;`},
		{`$x ||= {}; print 1;`, `$x ||= {};print(1);`},

		// A word operator needs a separator: `not foo` calls foo and
		// `notfoo` is the bareword "notfoo".
		{`not $a and $b;`, `not $a and $b;`},

		// A block form takes none.
		{`if (1) { 2; }`, `if (1) {2;}`},
	} {
		got := strings.TrimSpace(parse.Canon(parse.Parse([]byte(tc.src)), []byte(tc.src)))
		if got != tc.want {
			t.Errorf("Canon(parse(%q)):\n  got  %q\n  want %q", tc.src, got, tc.want)
		}
	}
}

// significant returns the tokens that carry meaning: whitespace, comments and
// POD are dropped.
//
// Canon drops trivia on purpose -- it belongs to no statement and carries no
// grouping -- so a byte comparison would fail on every file for a reason that
// says nothing about the tree. Comparing tokens is what makes the check about
// structure.
func significant(src []byte) []string {
	var out []string
	for _, tok := range lexer.Tokenize(src) {
		switch tok.Kind {
		case lexer.Whitespace, lexer.Comment, lexer.Pod, lexer.DataSection:
			continue
		}
		out = append(out, tok.Kind.String()+" "+string(src[tok.Start:tok.End]))
	}
	return out
}

// canonMatches reports whether canon(parse(src)) is a FIXPOINT: re-parsing
// the emission and emitting it again gives the same text back.
//
// THIS IS A STABILITY CHECK, NOT A FIDELITY CHECK, and the distinction is
// load-bearing. The source appears on neither side of the comparison after
// the first emission, so a tree that is not a parse of its source can still
// be a fixpoint. Measured, with the wrong tree TestCanonCatchesMisgrouping
// builds by hand:
//
//	WRONG tree: once="print((1 + 2) * 3);"
//	            twice="print((1 + 2) * 3);"   fixpoint
//
// An earlier version of this comment claimed the fixpoint caught that. It
// does not. Fidelity is asserted by TestCanonCatchesMisgrouping and
// TestCanonVisitsInteriorNodes, which compare against a correct tree; this
// sweeps the corpus for emissions perl could not read back.
//
// The literal criterion -- `tokens(canon(parse(S))) == tokens(S)` -- is what
// would carry fidelity corpus-wide, and no correct emitter can satisfy it:
// 519 of 658 disagreeing files differ only by a paren around a parenless
// call, and `print "x"` and `print("x")` parse to the SAME TREE. Quotienting
// by exactly that equivalence is the open work; see the issue.
func canonMatches(src []byte) (bool, string) {
	once := []byte(parse.Canon(parse.Parse(src), src))
	twice := []byte(parse.Canon(parse.Parse(once), once))

	want, got := significant(once), significant(twice)
	for i := 0; i < len(want) && i < len(got); i++ {
		if want[i] != got[i] {
			return false, fmt.Sprintf("token %d: first emission has %q, second has %q", i, want[i], got[i])
		}
	}
	if len(want) != len(got) {
		return false, fmt.Sprintf("first emission has %d significant tokens, second has %d", len(want), len(got))
	}
	return true, ""
}

// TestCanonTokenIdentity: every file that re-emits to the same significant
// tokens keeps doing so.
//
// The issue asked for 100% over the corpus, and that is the right target --
// a file that PARSES and re-emits differently is a defect with a location,
// not a coverage gap. It is not where the corpus is: 171 of 986 files
// disagree at the commit that introduced this check, and the ones inspected
// are PARSER defects the emitter merely made visible (01a0b983 is the
// largest class, a declarator losing its subscript to a second statement).
//
// So the gate holds the files that pass rather than demanding the ones that
// do not. TestCanonRatchet carries the count, and this asserts the property
// that must never break: a file in the passing set stays in it.
func TestCanonTokenIdentity(t *testing.T) {
	dir, files := t1Files(t)

	baseline, err := parseRatchetFile(filepath.Join("testdata", "t1canon.ratchet"))
	if err != nil {
		t.Fatalf("%v\nrun with -parse.update-ratchet to create it", err)
	}

	var broke []string
	for _, rel := range files {
		if baseline[rel] != 0 {
			continue // known-disagreeing; TestCanonRatchet holds the count
		}
		src, readErr := os.ReadFile(filepath.Join(dir, rel))
		if readErr != nil {
			continue
		}
		if ok, why := canonMatches(src); !ok {
			broke = append(broke, rel+": "+why)
		}
	}
	if len(broke) > 0 {
		sort.Strings(broke)
		t.Errorf("%d file(s) re-emitted identically at the baseline and no longer do:\n  %s",
			len(broke), strings.Join(broke, "\n  "))
	}
}

// TestCanonBlockArgumentRoundTrips: canon parenthesises a call, so a BLOCK
// argument lands after a `(` -- and the emission must still re-lex as a
// block.
//
// This is the fixpoint at its smallest. Canon writes `map({$_ + 1;}@a)`,
// which perl reads: `perl -MO=Deparse` emits that exact spelling for BOTH
// `map { $_+1 } @a` and the parenthesised form, so the text is not the
// defect. The defect was that our own lexer classified the `{` from XTerm --
// it required the brace to be ADJACENT to the word -- so the `;` ended a
// statement, the `}` was orphaned, and canon's output parsed to Unknown=1
// and re-emitted as `map({$_ + 1});}@a);`.
//
// The hashref reading must survive the same change, which is why the last
// two cases are here: perl runs intuit_curly inside the parens too, and
// `map({a => 1}, @a)` is a HASHREF where `map({; a => 1} @a)` is a block.
// A fix that forced a block after `WORD (` would break the first.
func TestCanonBlockArgumentRoundTrips(t *testing.T) {
	for _, src := range []string{
		"my @a = (1,2);\nmy @b = map { $_ + 1 } @a;\n",
		"my @b = grep { $_ > 1 } @a;\n",
		"my @b = sort { $a <=> $b } @a;\n",

		// Already parenthesised in the source, so the emission is the input's
		// own shape rather than one canon introduced.
		"my @b = map({$_ + 1;} @a);\n",

		// The hashref reading, which the fix must not swallow.
		"my @b = map({a => 1}, @a);\n",
		"my @b = map({; a => 1} @a);\n",

		// `print {$fh} "x"` is the other Block that reaches an argument list
		// -- 42 of the 46 files this defect touched hold one -- but it does
		// not parse at all yet, for a reason that is not this one: the
		// filehandle slot needs `WORD BLOCK ARG`, which is 01a0d087. It is
		// left out rather than marked skipped because a case that fails for
		// an unrelated reason measures that reason, not this fix.
	} {
		b := []byte(src)
		once := parse.Canon(parse.Parse(b), b)
		if ok, why := canonMatches(b); !ok {
			t.Errorf("canon of %q is not a fixpoint: %s\n  emission %q", src, why, strings.TrimSpace(once))
		}
		// A fixpoint is not enough on its own: an emission that fails to
		// parse can still be stable. The emission must also PARSE, which is
		// what the orphaned `}` broke.
		ob := []byte(once)
		if n := countUnknown(parse.Parse(ob)); n != 0 {
			t.Errorf("canon of %q does not re-parse: Unknown=%d\n  emission %q", src, n, strings.TrimSpace(once))
		}
	}
}

// TestCanonRatchet: the per-file re-emission result, failing in either
// direction.
//
// One per file rather than a total, because the total hides a swap -- a
// change that fixes five files and breaks five reads as no movement. The
// count is 1 for a file whose emission is not a fixpoint and 0 for one whose
// is, so "clean" in the header is the number of files that re-emit
// identically.
func TestCanonRatchet(t *testing.T) {
	dir, files := t1Files(t)

	now := make(map[string]int, len(files))
	for _, rel := range files {
		src, err := os.ReadFile(filepath.Join(dir, rel))
		if err != nil {
			continue
		}
		now[rel] = 0
		if ok, _ := canonMatches(src); !ok {
			now[rel] = 1
		}
	}

	checkRatchet(t, filepath.Join("testdata", "t1canon.ratchet"),
		"Files whose canon emission is NOT a fixpoint (1 = disagrees).", now)
}
