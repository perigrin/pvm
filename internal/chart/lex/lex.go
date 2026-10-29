// ABOUTME: Up-front chart/lex: the production lexer's token stream, mapped onto the cut grammar's token predicates.
// ABOUTME: Trivia is attached to the token it precedes so the stream is lossless; $#{BLOCK} is split as the cut splits it.

// Package lex is the chart's lexer (spec §5, "chart/lex"). In up-front mode
// the production lexer, with its own expect machine, produces the whole
// token stream before recognition, as internal/parse uses it.
package lex

import (
	"bytes"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"

	"tamarou.com/pvm/internal/chart/grammar"
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
// in Tail, and the context facts derived from them.
type Stream struct {
	Tokens []Token
	Tail   []lexer.Token
	Facts  *Context
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

// Lexer maps tokens onto one cut grammar's terminals.
type Lexer struct {
	byKind map[lexer.Kind][]terminal
	trivia map[lexer.Kind]bool
}

type terminal struct {
	name string
	pred grammar.Pred
}

// New is a Lexer over cut's terminals and trivia kinds.
func New(cut *grammar.CutGrammar) *Lexer {
	x := &Lexer{byKind: map[lexer.Kind][]terminal{}, trivia: map[lexer.Kind]bool{}}
	for _, name := range slices.Sorted(maps.Keys(cut.Terminals)) {
		for _, p := range cut.Terminals[name] {
			x.byKind[p.Kind] = append(x.byKind[p.Kind], terminal{name, p})
		}
	}
	for _, k := range cut.Trivia {
		x.trivia[k] = true
	}
	return x
}

// Guacamole is a Lexer over the reviewed cut of the Guacamole grammar read
// from slif.
func Guacamole(slif string) (*Lexer, error) {
	src, err := os.ReadFile(slif)
	if err != nil {
		return nil, err
	}
	g, err := grammar.ReadSLIF(filepath.Base(slif), src)
	if err != nil {
		return nil, err
	}
	cut, err := grammar.ApplyCut(g, grammar.GuacamoleCut)
	if err != nil {
		return nil, err
	}
	return New(cut), nil
}

// Lex lexes src up front and maps each token. The error names the byte
// offset of every token no predicate admits, and of any byte the stream
// fails to cover; the stream is returned whole either way.
func (x *Lexer) Lex(src []byte) (*Stream, error) {
	s := &Stream{}
	var errs []error
	var lead []lexer.Token
	for _, t := range splitArrayTop(src, lexer.TokenizeBarewords(src, BarewordOperators)) {
		if x.trivia[t.Kind] {
			lead = append(lead, t)
			continue
		}
		tok := Token{Token: t, Lead: lead}
		lead = nil
		text := string(src[t.Start:t.End])
		for _, term := range x.byKind[t.Kind] {
			if term.pred.Match(t.Kind, text) {
				tok.Terminals = append(tok.Terminals, term.name)
			}
		}
		if tok.Terminals == nil {
			errs = append(errs, fmt.Errorf("byte %d: %s token %q maps to no predicate", t.Start, t.Kind, text))
		}
		s.Tokens = append(s.Tokens, tok)
	}
	s.Tail = lead
	if err := s.covers(len(src)); err != nil {
		errs = append(errs, err)
	}
	s.Facts = facts(src, s.all())
	return s, errors.Join(errs...)
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

// splitArrayTop splits each `$#{BLOCK}` Variable token as the cut splits it
// (Guacamole's SigilArrayTop BlockNonEmpty): a Variable `$#{`, the interior
// re-lexed by the production lexer with its offsets in the file, and the
// closing `}` as the production lexer's CloseBracket. An unterminated one has
// no closer.
func splitArrayTop(src []byte, toks []lexer.Token) []lexer.Token {
	var out []lexer.Token
	for _, t := range toks {
		if t.Kind != lexer.Variable || !bytes.HasPrefix(src[t.Start:t.End], []byte("$#{")) {
			out = append(out, t)
			continue
		}
		open, end := t.Start+3, t.End
		closed := end > open && src[end-1] == '}'
		if closed {
			end--
		}
		out = append(out, lexer.Token{Kind: lexer.Variable, Start: t.Start, End: open})
		inner := lexer.TokenizeBarewords(src[open:end], BarewordOperators)
		for i := range inner {
			inner[i].Start += open
			inner[i].End += open
		}
		out = append(out, splitArrayTop(src, inner)...)
		if closed {
			out = append(out, lexer.Token{Kind: lexer.CloseBracket, Start: end, End: t.End})
		}
	}
	return out
}
