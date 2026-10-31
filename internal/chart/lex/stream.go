// ABOUTME: The token and stream types chart/lex hands the lattice: tokens with their leading trivia, and a lossless stream of them.
// ABOUTME: The lattice reads them; no grammar-specific mapping produces them yet.

package lex

import (
	"bytes"

	"tamarou.com/pvm/internal/lexer"
)

// Token is one significant token: the production lexer's token, the trivia
// before it, and the terminals that admit it. A token no terminal admits
// stays in the stream with none.
type Token struct {
	lexer.Token
	Lead      []lexer.Token
	Terminals []string
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

// all is every token in source order, trivia included.
func (s *Stream) all() []lexer.Token {
	var out []lexer.Token
	for _, t := range s.Tokens {
		out = append(out, t.Lead...)
		out = append(out, t.Token)
	}
	return append(out, s.Tail...)
}
