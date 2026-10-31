// ABOUTME: The token and stream types chart/lex hands the lattice: the lexer's own token slice, and indexes into it with ids.
// ABOUTME: Each significant token carries one terminal-set id; trivia attaches by index; heredoc bodies sit in a side table.

package lex

import (
	"bytes"
	"slices"

	"tamarou.com/pvm/internal/lexer"
)

// Token is one significant token: its index into the stream's Raw, and the
// terminals that admit it as a terminal-set id, an index into
// grammar.Tables.TerminalSets. Id 0 is the empty set, so a token no
// terminal admits, or never given an id, stays in the stream and refuses.
// It holds no pointer, so a stream's Tokens is never scanned by the GC.
type Token struct {
	Index       int
	TerminalSet int
}

// Stream is a source's tokens: Raw is the production lexer's slice, shared
// rather than copied, and Tokens indexes its significant tokens in order.
// Heredoc bodies, which the source holds after the opener's line rather
// than after the opener, are a side table.
type Stream struct {
	Raw    []lexer.Token
	Tokens []Token
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

// Lead is the trivia before Tokens[i], since the significant token before
// it, less any heredoc body, which is its opener's.
func (s *Stream) Lead(i int) []lexer.Token {
	return trivia(s.Raw[s.after(i):s.Tokens[i].Index])
}

// Tail is the trivia after the last token, less any heredoc body.
func (s *Stream) Tail() []lexer.Token {
	return trivia(s.Raw[s.after(len(s.Tokens)):])
}

// after is the index in Raw just past Tokens[i-1], or 0 for the first token.
func (s *Stream) after(i int) int {
	if i == 0 {
		return 0
	}
	return s.Tokens[i-1].Index + 1
}

// trivia is ts without its heredoc bodies; ts itself when it has none.
func trivia(ts []lexer.Token) []lexer.Token {
	if !slices.ContainsFunc(ts, isBody) {
		return ts
	}
	return slices.DeleteFunc(slices.Clone(ts), isBody)
}

func isBody(t lexer.Token) bool { return t.Kind == lexer.HeredocBody }

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
	for i, t := range s.Tokens {
		out = append(out, s.Lead(i)...)
		out = append(out, s.Raw[t.Index])
	}
	for _, b := range s.Bodies {
		out = append(out, b.Token)
	}
	out = append(out, s.Tail()...)
	slices.SortStableFunc(out, func(a, b lexer.Token) int { return a.Start - b.Start })
	return out
}
