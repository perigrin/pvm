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
	raw := lexer.TokenizeBarewords(src, BarewordOperators)
	// The facts come from the lexer's own tokens: the split's `}` would
	// close a `use constant {` early.
	s := &Stream{Facts: facts(src, raw)}
	var errs []error
	refused := map[int]bool{}
	// A `$#{` is refused after a trigger, when its own interior holds one or
	// a nested `$#{`, and after any refused `$#{`, whose interior went unread.
	from := fileStateFrom(src, raw)
	for _, t := range raw {
		if !isArrayTop(src, t) {
			continue
		}
		if t.Start >= from || interiorUnsafe(src, t) {
			from = min(from, t.Start)
			refused[t.Start] = true
			errs = append(errs, fmt.Errorf("byte %d: `$#{` interior needs the lexer's file state (01a1882d-d8a4)", t.Start))
		}
	}
	var lead []lexer.Token
	for _, t := range splitArrayTop(src, raw, refused) {
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
		if tok.Terminals == nil && !refused[t.Start] {
			errs = append(errs, fmt.Errorf("byte %d: %s token %q maps to no predicate", t.Start, t.Kind, text))
		}
		s.Tokens = append(s.Tokens, tok)
	}
	s.Tail = lead
	if err := s.covers(len(src)); err != nil {
		errs = append(errs, err)
	}
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
// no closer. One whose start is refused stays whole, and unmapped. A nested
// `$#{` is always refused with its outer one, so the interior never holds one.
func splitArrayTop(src []byte, toks []lexer.Token, refused map[int]bool) []lexer.Token {
	var out []lexer.Token
	for _, t := range toks {
		if !isArrayTop(src, t) || refused[t.Start] {
			out = append(out, t)
			continue
		}
		open, end, inner := interior(src, t)
		out = append(out, lexer.Token{Kind: lexer.Variable, Start: t.Start, End: open})
		for _, it := range inner {
			it.Start += open
			it.End += open
			out = append(out, it)
		}
		if end < t.End {
			out = append(out, lexer.Token{Kind: lexer.CloseBracket, Start: end, End: t.End})
		}
	}
	return out
}

func isArrayTop(src []byte, t lexer.Token) bool {
	return t.Kind == lexer.Variable && bytes.HasPrefix(src[t.Start:t.End], []byte("$#{"))
}

// interior is a `$#{` token's interior: its bytes open..end, past the `$#{`
// and before any closing `}`, and their tokens from a blank lexer state,
// with offsets relative to open.
func interior(src []byte, t lexer.Token) (open, end int, toks []lexer.Token) {
	open, end = t.Start+3, t.End
	if end > open && src[end-1] == '}' {
		end--
	}
	return open, end, lexer.TokenizeBarewords(src[open:end], BarewordOperators)
}

// interiorUnsafe reports whether a `$#{` interior holds what the rule
// refuses before a `$#{`, or a nested `$#{`.
func interiorUnsafe(src []byte, t lexer.Token) bool {
	open, end, toks := interior(src, t)
	in := src[open:end]
	return fileStateFrom(in, toks) < len(in) || slices.ContainsFunc(toks, func(it lexer.Token) bool { return isArrayTop(in, it) })
}

// fileStateFrom is the first byte from which the lexer's file state might
// differ from the blank state a `$#{` interior is re-lexed in, or len(src).
// It fails closed: only what is known to leave the state alone passes. That
// is `use` or `no` of strict or warnings, and anything that is not one of
// these, which are what move the production lexer's file state:
//
//   - any other `use` or `no`: utf8, feature and experimental (signatures,
//     class, the apostrophe separator), versions, constant, and any module;
//   - `sub`, which declares a constant or a lexical sub;
//   - `class`, which turns class syntax on with no `use`;
//   - `BEGIN`;
//   - a heredoc opener, whose body is pending until the next newline.
//
// shortcut: refuses rather than re-lexing in the file's state, until
// 01a1882d-d8a4 lets the production lexer lex the interior in place.
func fileStateFrom(src []byte, raw []lexer.Token) int {
	pragma := false
	for _, t := range raw {
		if lexer.IsTrivia(t.Kind) {
			continue
		}
		text := string(src[t.Start:t.End])
		word := t.Kind == lexer.Word
		switch {
		case pragma && !(word && (text == "strict" || text == "warnings")),
			word && (text == "sub" || text == "class" || text == "BEGIN"),
			t.Kind == lexer.HeredocOpen:
			return t.Start
		}
		pragma = word && (text == "use" || text == "no")
	}
	return len(src)
}
