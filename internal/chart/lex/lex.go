// ABOUTME: The thin chart/lex: lexer.Tokenize up front, trivia to the next significant token, one table lookup per token.
// ABOUTME: Heredoc bodies go to their openers; a token no class admits gets the empty set and the first such byte is kept.

// Package lex is the chart's lexer (spec §5.2, "A thin chart/lex"): the
// production lexer runs up front with its own expect machine, and each
// significant token is mapped to a terminal-set id through the compiled
// token table. Nothing grammar-specific is decided here.
package lex

import (
	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/lexer"
)

// Lex is Map over lexer.Tokenize's tokens of src.
func Lex(src []byte, tt *grammar.TokenTable) (*Stream, int) {
	return Map(src, lexer.Tokenize(src), tt)
}

// Map maps raw, src's tokens, through tt. It returns the stream and the byte
// of the first token no class admits, or -1. Trivia attaches to the next
// significant token, and each heredoc body to its opener, in order.
func Map(src []byte, raw []lexer.Token, tt *grammar.TokenTable) (*Stream, int) {
	text := string(src) // once per file, so each token's text is a slice of it
	s := &Stream{Tokens: make([]Token, 0, len(raw))}
	first := -1
	pred, p2, p1 := grammar.Boundary, tt.Edge, tt.Edge
	var cur grammar.Token
	// done looks up the last token, now that its next neighbour's class is known.
	done := func(next int) {
		last := &s.Tokens[len(s.Tokens)-1]
		last.TerminalSet = tt.Lookup(cur, p2, p1, next)
		if last.TerminalSet == 0 && first < 0 {
			first = last.Start
		}
		p2, p1 = p1, cur.Class
	}
	var opens []int        // heredoc openers whose bodies have not come yet, in order
	var held []lexer.Token // trivia a heredoc body interrupted, for the next token
	a := 0                 // raw[a:i] is the trivia since the last token or body
	for i, t := range raw {
		switch {
		case lexer.IsTrivia(t.Kind):
			continue
		case t.Kind == lexer.HeredocBody && len(opens) > 0:
			held = append(held, raw[a:i]...)
			s.Tokens[opens[0]].Body = raw[i : i+1 : i+1]
			opens, a = opens[1:], i+1
			continue
		}
		it := intern(tt, text, t, pred)
		if len(s.Tokens) > 0 {
			done(it.Class)
		}
		lead := raw[a:i:i]
		if held != nil {
			lead, held = append(held, lead...), nil
		}
		if t.Kind == lexer.HeredocOpen {
			opens = append(opens, len(s.Tokens))
		}
		s.Tokens = append(s.Tokens, Token{Token: t, Lead: lead})
		pred, cur, a = it.N, it, i+1
	}
	if len(s.Tokens) > 0 {
		done(tt.Edge)
	}
	s.Tail = raw[a:]
	if held != nil {
		s.Tail = append(held, s.Tail...)
	}
	return s, first
}

// intern is t's key interned in tt, given its predecessor's class.
func intern(tt *grammar.TokenTable, text string, t lexer.Token, pred grammar.Neighbour) grammar.Token {
	tx := text[t.Start:t.End]
	return tt.Intern(grammar.TokenView{Kind: t.Kind, Text: tx, OpensBlock: t.OpensBlock, TermFollows: t.TermFollows, Prefix: prefix(t.Kind, tx)}, pred)
}

// prefix is a variable's sigil (`#` for `$#`) or a quote's class: its opening
// quote, or its operator (`Q` for qq, `y` for tr); 0 for anything else.
func prefix(k lexer.Kind, tx string) byte {
	switch {
	case k == lexer.Variable && len(tx) > 2 && tx[:2] == "$#":
		return '#'
	case k == lexer.Variable && tx != "":
		return tx[0]
	case k != lexer.Quote || tx == "":
		return 0
	}
	n := 0
	for n < len(tx) && 'a' <= tx[n] && tx[n] <= 'z' {
		n++
	}
	switch tx[:n] {
	case "":
		return tx[0]
	case "qq":
		return 'Q'
	case "qw":
		return 'w'
	case "qr":
		return 'r'
	case "qx":
		return 'x'
	case "tr":
		return 'y'
	}
	return tx[0] // q, m, s, y
}
