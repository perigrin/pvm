// ABOUTME: The token and stream types chart/lex hands the lattice: tokens with their leading trivia, and a lossless stream of them.
// ABOUTME: Each token carries one terminal-set id; heredoc bodies sit in a side table by opener.

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
	Lead        []lexer.Token
	TerminalSet int
}

// Stream is a source's tokens in order, with the trivia after the last one
// in Tail. Heredoc bodies, which the source holds after the opener's line
// rather than after the opener, are a side table so a Token stays small.
type Stream struct {
	Tokens []Token
	Tail   []lexer.Token
	Bodies []Body
}

// Body is a heredoc body and the index in Tokens of its opener.
type Body struct {
	Opener int
	lexer.Token
}

// Body is the heredoc body of Tokens[i], if it opens one.
func (s *Stream) Body(i int) []lexer.Token {
	for _, b := range s.Bodies {
		if b.Opener == i {
			return []lexer.Token{b.Token}
		}
	}
	return nil
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
	}
	for _, b := range s.Bodies {
		out = append(out, b.Token)
	}
	out = append(out, s.Tail...)
	slices.SortStableFunc(out, func(a, b lexer.Token) int { return a.Start - b.Start })
	return out
}
