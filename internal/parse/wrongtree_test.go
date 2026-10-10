// ABOUTME: Four trees that were not a parse of their source, and all four round-tripped.
// ABOUTME: Round-trip proves no byte was lost, never that the parse was right (chapter 7 §7.2(c)).

package parse_test

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// topStatements counts the statements directly under the root.
func topStatements(n *parse.Node) int {
	count := 0
	for _, c := range n.Children {
		if c.Kind == parse.Statement {
			count++
		}
	}
	return count
}

// TestAnonSubInStatementPosition: `sub { 1 };` is ONE statement holding one
// anonymous sub.
//
// It produced two: a Declaration spanning only `sub`, then a separate
// statement holding an AnonHash. That is a tree which is not a parse of its
// source, and it round-trips -- the exact failure `Unknown` exists to
// prevent (`parse.go:20-33`), arriving through nodes that look ordinary.
//
// In EXPRESSION position it was always right: `my $f = sub { 1 }` gives a
// Declaration with a Block, because parseTerm's anon-sub check
// (`term.go:126-132`) is reached there. The statement path ran
// parseDeclaration first and never got to it.
func TestAnonSubInStatementPosition(t *testing.T) {
	for _, src := range []string{
		"sub { 1 };",
		"sub { my $x = shift; $x };",
		"sub { 1 }->();",
	} {
		root := parse.Parse([]byte(src))
		if n := topStatements(root); n != 1 {
			t.Errorf("%q: %d top-level statements, want 1: %v",
				src, n, kinds(root))
		}
		if containsKind(root, parse.AnonHash) {
			t.Errorf("%q: the body is a BLOCK, not an anonymous hash: %v",
				src, kinds(root))
		}
		// The Declaration must own a Block, not stop at the keyword.
		decl := find(root, parse.Declaration)
		if decl == nil {
			t.Errorf("%q: no Declaration: %v", src, kinds(root))
			continue
		}
		if find(decl, parse.Block) == nil {
			t.Errorf("%q: the anonymous sub has no Block child: %v",
				src, kinds(root))
		}
	}
}

// TestNamedSubStillParses is the anon-sub negative: a NAMED sub must keep
// its existing shape, which is the path the fix runs through.
func TestNamedSubStillParses(t *testing.T) {
	for _, src := range []string{
		"sub f { 1 }",
		"sub f;",
		"my $f = sub { 1 };",
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q must parse: %v", src, kinds(root))
		}
		if topStatements(root) != 1 {
			t.Errorf("%q: %d top-level statements, want 1: %v",
				src, topStatements(root), kinds(root))
		}
	}
}

// TestStatementStartBrace: `{a=>1};` is an anonymous hash, and perl says so.
//
// Measured on perl 5.42.0:
//
//	$ perl -MO=Deparse -e '{a=>1};'
//	+{'a', 1};
//
// The leading `+` is Deparse disambiguating for a reader; the node is an
// anonhash. toke.c:6698-6842's heuristic fires at statement start too, which
// is what this parser was missing: it read a Block holding a `=>` statement
// and then an Unknown for the `;`.
func TestStatementStartBrace(t *testing.T) {
	for _, src := range []string{
		"{a=>1};",
		"{a=>1, b=>2};",
		`{"a", 1};`,
		`{"a" => 1};`,
		// `{};` is in TestEmptyBracesAtStatementAreAnonHash: an empty brace
		// is a hash too, by its own rule. Deparse writes it `{};` with no
		// `+`, so read it with -MO=Concise, which shows emptyavhv.
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q must parse: %v", src, kinds(root))
		}
		if !containsKind(root, parse.AnonHash) {
			t.Errorf("%q is an anonymous hash, and perl deparses it as one: %v",
				src, kinds(root))
		}
	}
}

// TestBareBlockStillParses is the statement-start brace's negative, and the
// harder half: a bare block IS a block, and the two are the same two bytes.
//
// Measured:
//
//	$ perl -MO=Deparse -e '{ my $x = 1; }'
//	{
//	    my $x = 1;
//	}
//
// A `{` at statement start holding STATEMENTS is a block; holding a comma
// list it is a constructor. That is the distinction, and a fix that made
// every statement-start brace a hash would break every bare block, every
// `if` body reached this way, and every loop.
func TestBareBlockStillParses(t *testing.T) {
	for _, src := range []string{
		"{ my $x = 1; }",
		"{ f(); g(); }",
		"{ ; }",
		"if ($x) { 1 }",
		"while ($x) { f(); }",
		// `{ }` and `{};` are NOT here: an empty brace at statement position
		// is an anonymous hash -- measured with -MO=Concise, emptyavhv --
		// and TestEmptyBracesAtStatementAreAnonHash owns them.
		// `LOOP: { last LOOP; }` is NOT here. A labelled bare block is valid
		// Perl -- `perl -MO=Deparse` gives it back unchanged -- and this
		// parser declines it today, before and after this change. Its own
		// defect, not this one's: a label before a BLOCK rather than before
		// a loop or statement. Untracked.
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q must parse: %v", src, kinds(root))
		}
		if !containsKind(root, parse.Block) {
			t.Errorf("%q is a block, not a constructor: %v", src, kinds(root))
		}
	}
}

// TestGlobDerefIsNotAConstructor: `*{"foo"}` is a glob dereference.
//
// It read as `term "*"` over an AnonHash -- a tree saying a hash is being
// BUILT where a symbol-table slot is being named. The braces group an
// expression; they do not construct.
func TestGlobDerefIsNotAConstructor(t *testing.T) {
	for _, src := range []string{
		`*{"foo"};`,
		`*{"foo"} = sub { 1 };`,
		`*{$name} = \&other;`,
	} {
		root := parse.Parse([]byte(src))
		glob := find(root, parse.Term)
		if glob == nil {
			t.Errorf("%q: no Term: %v", src, kinds(root))
			continue
		}
		// The glob's operand must not be an AnonHash: `{"foo"}` is a group.
		for _, n := range findAll(root, parse.AnonHash) {
			_ = n
			t.Errorf("%q: `{...}` after a glob sigil groups an expression, "+
				"it does not construct a hash: %v", src, kinds(root))
			break
		}
	}
}

// TestAnonHashStillConstructs is the glob's negative. A `{...}` that is NOT
// after a glob sigil is still a constructor.
func TestAnonHashStillConstructs(t *testing.T) {
	for _, src := range []string{
		`my $r = {a => 1};`,
		`f({a => 1});`,
		`my $r = {};`,
	} {
		root := parse.Parse([]byte(src))
		if !containsKind(root, parse.AnonHash) {
			t.Errorf("%q: a constructor must stay a constructor: %v",
				src, kinds(root))
		}
	}
}

// TestSignaturesKeepTheirBody is the worst of the four, because turning the
// feature ON made the parse WORSE.
//
// Without the pragma, `sub f ($x, $y) { 1 }` reads `($x, $y)` as a prototype
// and `{ 1 }` as a block -- which is what perl does when signatures are off.
// With `use v5.36`, `internal/lexer/proto.go` declines to scan the
// prototype and nothing downstream picks the signature up, so the body was
// read as a hash slice OF the signature and the sub body was lost.
//
// Measured on perl 5.42.0:
//
//	$ perl -MO=Deparse -e 'use v5.36; sub f ($x, $y) { $x + $y }'
//	sub f ($x, $y) {
//	    $x + $y;
//	}
//
// Body intact, signature intact.
func TestSignaturesKeepTheirBody(t *testing.T) {
	for _, src := range []string{
		"use v5.36;\nsub f ($x, $y) { 1 }\n",
		"use feature 'signatures';\nsub f ($x) { $x }\n",
		"use v5.36;\nsub f ($x = 1) { $x }\n",
		"use v5.36;\nsub f (@rest) { 1 }\n",
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q must parse: %v", src, kinds(root))
		}
		decl := find(root, parse.Declaration)
		if decl == nil {
			t.Errorf("%q: no Declaration: %v", src, kinds(root))
			continue
		}
		if find(decl, parse.Block) == nil {
			t.Errorf("%q: the sub KEEPS ITS BODY -- perl deparses it back "+
				"with the block intact: %v", src, kinds(root))
		}
	}
}

// TestPrototypeWithoutSignatures is the signature's negative. With no
// pragma, `($x)` after a sub name is a PROTOTYPE, which is what perl reads
// it as, and that path must not change.
func TestPrototypeWithoutSignatures(t *testing.T) {
	for _, src := range []string{
		"sub f ($) { 1 }",
		"sub f ($$) { 1 }",
		"sub f (\\@) { 1 }",
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q must parse: %v", src, kinds(root))
		}
		if !containsKind(root, parse.PrototypeNode) {
			t.Errorf("%q: with no signature pragma this is a PROTOTYPE: %v",
				src, kinds(root))
		}
	}
}

// structure renders a subtree as dumpTree does, without its trivia, so a
// shape reads as the nodes a walker sees.
func structure(n *parse.Node) string {
	var b strings.Builder
	var walk func(*parse.Node, int)
	walk = func(n *parse.Node, d int) {
		if n.Kind == parse.Trivia {
			return
		}
		b.WriteString(strings.Repeat(" ", d) + n.Kind.String())
		if n.Text != "" {
			b.WriteString(" " + n.Text)
		}
		b.WriteString("\n")
		for _, c := range n.Children {
			walk(c, d+2)
		}
	}
	walk(n, 0)
	return b.String()
}

// TestSignatureParametersAreNodes: each parameter of a signature is its own
// node under the Declaration, not one span the walker cannot see into.
// perl compiles one argelem per parameter, and the default as its own op:
//
//	$ perl -MO=Concise,f -e 'use feature "signatures"; sub f ($x, $y = 1, @r) { }'
//	4  <+> argelem(0)[$x:3,6] /SV ->5
//	8  <+> argelem(1)[$y:4,6] KS/SV ->9
//	6     <|> argdefelem(other->7)[1] sK/1 ->8
//	7        <$> const[IV 1] s ->8
//	a  <+> argelem(2)[@r:5,6] /AV ->b
func TestSignatureParametersAreNodes(t *testing.T) {
	src := "use feature 'signatures'; sub f ($x, $y = 1, @r) { }"
	root := parse.Parse([]byte(src))
	if containsKind(root, parse.Unknown) {
		t.Fatalf("%q must parse:\n%s", src, dumpTree(root, 0))
	}
	var decl *parse.Node
	for _, d := range findAll(root, parse.Declaration) {
		if d.Text == "sub" {
			decl = d
		}
	}
	if decl == nil {
		t.Fatalf("%q: no sub Declaration:\n%s", src, dumpTree(root, 0))
	}
	want := `declaration sub
  term f
  list
    term $x
    binary =
      term $y
      term 1
    term @r
  block
`
	if got := structure(decl); got != want {
		t.Errorf("%q: want three parameter nodes, the default its own operand:\ngot:\n%swant:\n%s",
			src, got, want)
	}
}

// TestWrongTreesAreGone asserts the SHAPE of each of the four trees that
// round-tripped while being wrong, because a round trip cannot tell them
// apart from the right ones. Measured on perl 5.42.0 with -MO=Deparse:
//
//	sub { 1 };                    sub { 1; };         one anoncode
//	{a=>1};                       +{'a', 1};          an anonhash
//	*{"foo"};                     *{'foo';};          rv2gv over the string
//	use v5.36; sub f ($x, $y) {1} sub f ($x, $y) { 1; }
func TestWrongTreesAreGone(t *testing.T) {
	stmt := func(src string, i int) *parse.Node {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Fatalf("%q must parse:\n%s", src, dumpTree(root, 0))
		}
		var stmts []*parse.Node
		for _, c := range root.Children {
			if c.Kind == parse.Statement {
				stmts = append(stmts, c)
			}
		}
		if len(stmts) <= i {
			t.Fatalf("%q: want statement %d:\n%s", src, i, dumpTree(root, 0))
		}
		return stmts[i]
	}
	for _, tc := range []struct {
		src  string
		stmt int
		want string
	}{
		// One statement, one anonymous sub owning its block.
		{"sub { 1 };", 0, `statement
  declaration sub
    block
      statement
        term 1
`},
		// An anonymous hash, the same node `my $r = {a=>1}` builds.
		{"{a=>1};", 0, `statement
  anon_hash
    call a
    term 1
`},
		// The signature is the Declaration's list and the body its block.
		{"use v5.36;\nsub f ($x, $y) { 1 }\n", 1, `statement
  declaration sub
    term f
    list
      term $x
      term $y
    block
      statement
        term 1
`},
	} {
		if got := structure(stmt(tc.src, tc.stmt)); got != tc.want {
			t.Errorf("%q:\ngot:\n%swant:\n%s", tc.src, got, tc.want)
		}
	}

	// `*{"foo"}`: the glob sigil applied to the string, which is its one
	// operand. The node's kind is left unasserted: `${"foo"}` is a Unary and
	// this is a Term, and which one a glob deref should be is not this
	// test's question.
	src := `*{"foo"};`
	s := stmt(src, 0)
	if len(s.Children) != 1 {
		t.Fatalf("%q: want one expression:\n%s", src, dumpTree(s, 0))
	}
	glob := s.Children[0]
	if glob.Text != "*" || len(glob.Children) != 1 ||
		glob.Children[0].Kind != parse.Term || glob.Children[0].Text != `"foo"` {
		t.Errorf("%q: want `*` over the term \"foo\":\n%s", src, dumpTree(s, 0))
	}
}
