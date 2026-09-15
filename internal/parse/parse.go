// ABOUTME: The parse tree: nodes over byte spans, and the Unknown that declines rather than guesses.
// ABOUTME: Package parse, not ast — internal/parser is the tree-sitter wrapper and will be for a while.

package parse

import "tamarou.com/pvm/internal/lexer"

// Kind names what a node is.
type Kind int

const (
	// SourceFile is the root. It always exists, including for empty input:
	// an empty program is a valid program, and a nil root would make every
	// consumer check for one.
	SourceFile Kind = iota

	// Statement is one statement, whatever its form.
	Statement

	// Unknown is a construct this parser does not understand, spanning
	// exactly the bytes it could not parse.
	//
	// Unknown is the milestone's whole discipline in one node kind. A parser
	// that guesses at an unfamiliar construct produces a tree that is not a
	// parse of its source, and the fidelity harness scores that WRONG -- the
	// only bucket that fails a build. Emitting Unknown instead is a refusal,
	// and a refusal is never a wrong answer.
	//
	// But a refusal only counts as one if it is REPORTED. See Sites: an
	// Unknown that says nothing to the harness is scored as claiming there
	// is nothing there, which is a silent wrong answer rather than a
	// decline. The node and the hedge are one mechanism in two places.
	Unknown

	// Trivia is whitespace, a comment, or POD: bytes that belong to no
	// statement but must still be in the tree, because every byte is.
	Trivia
)

func (k Kind) String() string {
	switch k {
	case SourceFile:
		return "source_file"
	case Statement:
		return "statement"
	case Unknown:
		return "unknown"
	case Trivia:
		return "trivia"
	}
	return "?"
}

// Node is one node of the parse tree.
//
// Byte offsets rather than line/column, for the reason the lexer's Token
// gives: line and column are derived at the LSP layer, and carrying them
// doubles the state that must stay consistent under edit.
//
// A node's span always covers its children's spans exactly, with no gaps.
// That is what makes the round-trip property hold: concatenating every
// leaf's text reproduces the input, Unknown included.
type Node struct {
	Kind       Kind
	Start, End int
	Children   []*Node
}

// Parse builds a tree covering every byte of src exactly once.
//
// Like Tokenize, it has no error return and never panics. An LSP sees
// half-typed buffers constantly, and a parser that gives up on the first
// unfamiliar construct is useless to it.
//
// At this stage every statement is Unknown: the skeleton establishes the
// invariants before there is any grammar to test them against. Later issues
// replace Unknown with real nodes one construct at a time, and the round-trip
// test is what keeps each replacement honest.
func Parse(src []byte) *Node {
	root := &Node{Kind: SourceFile, Start: 0, End: len(src)}
	toks := lexer.Tokenize(src)

	p := &parser{src: src, toks: toks}
	for p.pos < len(p.toks) {
		before := p.pos
		if n := p.statement(); n != nil {
			root.Children = append(root.Children, n)
		}
		// Forward progress, asserted rather than assumed. The lexer learned
		// this the hard way: scanFormatBody was wrapped in a step() guard and
		// the fuzzer still found a `format =\n` input that spun. A parser
		// loop that can return without consuming is the same bug one layer
		// up, so the guard is here rather than in a comment.
		if p.pos == before {
			p.pos++
		}
	}
	return root
}

// parser is the cursor over the token stream.
type parser struct {
	src  []byte
	toks []lexer.Token
	pos  int
}

// statement consumes one statement, or one run of trivia.
//
// Trivia comes out as its own node rather than being attached to a
// statement, because "which statement does a blank line between two subs
// belong to" has no good answer and every answer complicates round-trip.
func (p *parser) statement() *Node {
	if isTrivia(p.toks[p.pos].Kind) {
		start := p.toks[p.pos].Start
		end := start
		for p.pos < len(p.toks) && isTrivia(p.toks[p.pos].Kind) {
			end = p.toks[p.pos].End
			p.pos++
		}
		return &Node{Kind: Trivia, Start: start, End: end}
	}

	// Everything else is Unknown for now, spanning to the end of the
	// statement. Resynchronising on a statement boundary is what stops one
	// unfamiliar construct from cascading: op/filetest.t:96 derailed a
	// reference 35 lines later under the tree-sitter grammar.
	start := p.toks[p.pos].Start
	end := start
	depth := 0
	for p.pos < len(p.toks) {
		tok := p.toks[p.pos]
		end = tok.End
		p.pos++

		switch tok.Kind {
		case lexer.Operator:
			// The lexer has no OpenBracket kind: an opener is an Operator,
			// because it moves the expect state the same way every other
			// operator does. Only a CLOSER is distinct, since it moves the
			// state back. So depth is tracked on the byte here, and the
			// asymmetry is the lexer's rather than an oversight.
			if tok.End-tok.Start == 1 {
				switch p.src[tok.Start] {
				case '{', '(', '[':
					depth++
				}
			}
		case lexer.CloseBracket:
			// A closer with nothing open ends the statement: it belongs to
			// an enclosing construct, and consuming past it would swallow
			// the rest of the file.
			if depth == 0 {
				return &Node{Kind: Unknown, Start: start, End: end}
			}
			depth--
			if depth == 0 {
				return &Node{Kind: Unknown, Start: start, End: end}
			}
		case lexer.Semicolon:
			if depth == 0 {
				return &Node{Kind: Unknown, Start: start, End: end}
			}
		}
	}
	return &Node{Kind: Unknown, Start: start, End: end}
}

func isTrivia(k lexer.Kind) bool {
	switch k {
	case lexer.Whitespace, lexer.Comment, lexer.Pod:
		return true
	}
	return false
}
