// ABOUTME: The token and stream types chart/lex hands the lattice: tokens with their leading trivia, and a lossless stream of them.
// ABOUTME: Independent of the predicate mapping in lex.go, so the mapping can go without them.

package lex

import (
	"bytes"
	"fmt"

	"tamarou.com/pvm/internal/lexer"
)

// Token is one significant token: the production lexer's token, the trivia
// before it, and the cut's terminals whose predicate admits it. A token no
// predicate admits stays in the stream with no terminals.
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

// covers checks that the stream tiles bytes 0 to n with no gap or overlap.
func (s *Stream) covers(n int) error {
	at := 0
	for _, t := range s.all() {
		if t.Start != at {
			return fmt.Errorf("byte %d: the stream resumes at byte %d", at, t.Start)
		}
		at = t.End
	}
	if at != n {
		return fmt.Errorf("byte %d: the stream ends before the source's %d bytes", at, n)
	}
	return nil
}
