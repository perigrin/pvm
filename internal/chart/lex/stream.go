// ABOUTME: The token and stream types chart/lex hands the lattice: tokens with their leading trivia, and a lossless stream of them.
// ABOUTME: Each token carries one terminal-set id, and a heredoc opener its body.

package lex

import (
	"bytes"
	"slices"

	"tamarou.com/pvm/internal/lexer"
)

// Token is one significant token: the production lexer's token, the trivia
// before it, and the terminals that admit it as a terminal-set id, an index
// into grammar.Tables.TerminalSets. Id 0 is the empty set, so a token no
// terminal admits, or never given an id, stays in the stream and refuses.
type Token struct {
	lexer.Token
	Lead []lexer.Token
	// Body is a heredoc opener's body, which the source holds after the
	// opener's line rather than after the opener.
	Body        []lexer.Token
	TerminalSet int
}

// Stream is a source's tokens in order, with the trivia after the last one
// in Tail.
type Stream struct {
	Tokens []Token
	Tail   []lexer.Token
}

// Text is every byte the stream's tokens and trivia span, in order. For a
// lossless stream it is the source.
func (s *Stream) Text(src []byte) []byte {
	var b bytes.Buffer
	for _, t := range s.all() {
		b.Write(src[t.Start:t.End])
	}
	return b.Bytes()
}

// all is every token in source order, trivia and heredoc bodies included.
func (s *Stream) all() []lexer.Token {
	var out []lexer.Token
	for _, t := range s.Tokens {
		out = append(out, t.Lead...)
		out = append(out, t.Token)
		out = append(out, t.Body...)
	}
	out = append(out, s.Tail...)
	slices.SortStableFunc(out, func(a, b lexer.Token) int { return a.Start - b.Start })
	return out
}
