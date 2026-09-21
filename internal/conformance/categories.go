// ABOUTME: Maps GLOSSARY.md category names onto this lexer's token kinds.
// ABOUTME: The single point of coupling; another lexer replaces this file alone.
package conformance

import (
	"strings"

	"tamarou.com/pvm/internal/lexer"
)

// categories translates a glossary category into a predicate over one of
// our tokens.
//
// THIS FILE IS THE COUPLING, and it is deliberately the only one. Corpus
// files assert in the glossary's vocabulary ("numeric literal"), never in
// ours (`lexer.Number`), so a project whose lexer has different kinds
// adopts the corpus by rewriting this map and nothing else.
//
// A category may only be added here once GLOSSARY.md defines it, or the
// corpus asserts in a vocabulary with no definition behind it.
//
// The predicate takes the token TEXT as well as its kind because a kind
// is not always enough: our lexer gives `"hi"`, `qw(a b)` and `tr/a/b/`
// the same Quote kind, while the glossary distinguishes a string literal
// from a quote-like operator. Another lexer may split them by kind and
// ignore the text.
var categories = map[string]func(lexer.Kind, string) bool{
	"numeric literal": kind(lexer.Number),
	"word":            kind(lexer.Word),
	"variable":        kind(lexer.Variable),
	"operator":        kind(lexer.Operator),

	// A quote-like operator is spelled with a name and a delimiter --
	// q, qq, qw, qr, m, s, tr, y -- where a string literal is spelled
	// with delimiters alone. Both are Quote to us, so the text decides.
	"string literal": func(k lexer.Kind, text string) bool {
		return k == lexer.Quote && !hasQuoteOperator(text)
	},
	"quote-like operator": func(k lexer.Kind, text string) bool {
		return k == lexer.Quote && hasQuoteOperator(text)
	},

	// A heredoc is two tokens: the `<<EOT` that appears in the statement
	// and the body that follows the line. They are separate categories
	// because a corpus file asserting where a heredoc STARTS is making a
	// different claim from one asserting what it CONTAINS.
	"heredoc opener": kind(lexer.HeredocOpen),
	"heredoc body":   kind(lexer.HeredocBody),
}

// quoteOperators are the names that introduce a quote-like operator, per
// perlop "Quote and Quote-like Operators". Longest first, so `qw` is not
// mistaken for `q` with a `w` delimiter.
var quoteOperators = []string{"qq", "qw", "qr", "tr", "q", "m", "s", "y"}

// hasQuoteOperator reports whether a Quote token is spelled with a leading
// operator name rather than with bare delimiters.
func hasQuoteOperator(text string) bool {
	for _, op := range quoteOperators {
		rest, ok := strings.CutPrefix(text, op)
		if !ok || rest == "" {
			continue
		}
		// The character after the name must be the delimiter, not more
		// word characters: `sort` starts with `s` but is not a
		// substitution, and our lexer would not call it a Quote anyway.
		if c := rest[0]; !isWordByte(c) {
			return true
		}
	}
	return false
}

func isWordByte(c byte) bool {
	return c == '_' ||
		('a' <= c && c <= 'z') ||
		('A' <= c && c <= 'Z') ||
		('0' <= c && c <= '9')
}

// kind matches on the token kind alone, which is the common case.
func kind(want lexer.Kind) func(lexer.Kind, string) bool {
	return func(got lexer.Kind, _ string) bool { return got == want }
}
