// ABOUTME: The statement skeleton: a file is a sequence of statements with no gaps, and recovery resyncs.
// ABOUTME: The cascade this prevents is measured — op/filetest.t:96 derailed a reference 35 lines later.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestSourceFileIsStatements: the root's children tile the file. No gaps,
// no overlaps, in order.
//
// Gap-freedom is what makes round-trip hold, but it is a stronger property
// and worth asserting on its own: a tree can round-trip while a node's span
// disagrees with its neighbours, because SourceText emits the gaps as the
// parent's own bytes. Here the gaps must not exist at all.
func TestSourceFileIsStatements(t *testing.T) {
	for _, src := range []string{
		"$x;\n$y;\n",
		"# comment\n$x;\n",
		"{ }\n$x;\n",
		"if ($c) { 1 }\n$x;\n",
		"\n\n$x;\n\n",
		"$x;",
	} {
		root := parse.Parse([]byte(src))
		at := 0
		for i, n := range root.Children {
			if n.Start != at {
				t.Errorf("%q: child %d starts at %d, want %d (gap or overlap)",
					src, i, n.Start, at)
			}
			at = n.End
		}
		if at != len(src) {
			t.Errorf("%q: children end at %d, want %d", src, at, len(src))
		}
	}
}

// TestExpressionStatementAndBlock: the two forms this issue owns.
//
// A bare block is a statement, not a hash constructor. The lexer's brace
// stack already made that call -- a `{` in XState opens a block -- so this
// reads the decision rather than re-deriving it.
func TestExpressionStatementAndBlock(t *testing.T) {
	root := parse.Parse([]byte("$x + 1;\n"))
	if n := firstOfKind(root, parse.Statement); n == nil {
		t.Errorf("an expression statement must parse: %v", kinds(root))
	}

	root = parse.Parse([]byte("{ $x; }\n"))
	blk := firstOfKind(root, parse.Block)
	if blk == nil {
		t.Fatalf("a bare block must parse as a Block: %v", kinds(root))
	}
	if firstOfKind(blk, parse.Statement) == nil {
		t.Errorf("the block's body must hold its statements: %v", kinds(blk))
	}

	// An empty block is still a block.
	root = parse.Parse([]byte("{ }\n"))
	if firstOfKind(root, parse.Block) == nil {
		t.Errorf("an empty block must parse as a Block: %v", kinds(root))
	}

	// And a block nests.
	root = parse.Parse([]byte("{ { $x; } }\n"))
	outer := firstOfKind(root, parse.Block)
	if outer == nil || firstOfKind(outer, parse.Block) == nil {
		t.Errorf("blocks must nest: %v", kinds(root))
	}
}

// TestStatementRecovery: an unknown statement becomes Unknown, and parsing
// resumes at the next boundary rather than giving up.
func TestStatementRecovery(t *testing.T) {
	src := []byte(unimplementedStatement + "$x;\n")
	root := parse.Parse(src)

	if firstOfKind(root, parse.Unknown) == nil {
		t.Fatalf("an unimplemented statement form must be Unknown: %v", kinds(root))
	}
	// The statement AFTER it still parses. That is the whole point of
	// resynchronising rather than returning.
	if firstOfKind(root, parse.Statement) == nil {
		t.Errorf("parsing must resume after an Unknown: %v", kinds(root))
	}
}

// TestRecoveryDoesNotCascade is the property the cascade cost us before.
//
// `op/filetest.t:96` derailed a reference at line 131 under the tree-sitter
// grammar: one unparseable line, 35 lines of wreckage. So the test is not
// "recovery happens" but "the statements after an unparseable one are
// IDENTICAL to what they would be without it".
func TestRecoveryDoesNotCascade(t *testing.T) {
	good := "$a;\n$b;\n$c;\n"
	bad := unimplementedStatement + good

	goodRoot := parse.Parse([]byte(good))
	badRoot := parse.Parse([]byte(bad))

	offset := len(unimplementedStatement)
	goodStmts := statementsOf(goodRoot)
	badStmts := statementsOf(badRoot)

	if len(goodStmts) != len(badStmts) {
		t.Fatalf("got %d statements after the bad line, want %d: %v",
			len(badStmts), len(goodStmts), kinds(badRoot))
	}
	for i := range goodStmts {
		wantStart := goodStmts[i].Start + offset
		wantEnd := goodStmts[i].End + offset
		if badStmts[i].Start != wantStart || badStmts[i].End != wantEnd {
			t.Errorf("statement %d spans [%d,%d), want [%d,%d): the bad line shifted it",
				i, badStmts[i].Start, badStmts[i].End, wantStart, wantEnd)
		}
	}
}

func statementsOf(root *parse.Node) []*parse.Node {
	var out []*parse.Node
	for _, n := range root.Children {
		if n.Kind == parse.Statement {
			out = append(out, n)
		}
	}
	return out
}
