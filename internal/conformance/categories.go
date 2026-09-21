// ABOUTME: Maps GLOSSARY.md category names onto this lexer's token kinds.
// ABOUTME: The single point of coupling; another lexer replaces this file alone.
package conformance

import "tamarou.com/pvm/internal/lexer"

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
	// q, qq, qw, qr, qx, m, s, tr, y -- or with backticks, where a
	// string literal is spelled with delimiters alone. Both are Quote to
	// us, so the text decides.
	//
	// The predicate is the LEXER's, not a second copy of its table here:
	// a hand-written copy lost `qx` immediately, which reclassified
	// `qx/ls/` as a string. Asking the lexer is what keeps the two from
	// drifting. An adopting project replaces this call along with the
	// rest of the file.
	"string literal": func(k lexer.Kind, text string) bool {
		return k == lexer.Quote && !lexer.HasQuoteOperator(text)
	},
	"quote-like operator": func(k lexer.Kind, text string) bool {
		return k == lexer.Quote && lexer.HasQuoteOperator(text)
	},

	// A heredoc is two tokens: the `<<EOT` that appears in the statement
	// and the body that follows the line. They are separate categories
	// because a corpus file asserting where a heredoc STARTS is making a
	// different claim from one asserting what it CONTAINS.
	"heredoc opener": kind(lexer.HeredocOpen),
	"heredoc body":   kind(lexer.HeredocBody),
}

// kind matches on the token kind alone, which is the common case.
func kind(want lexer.Kind) func(lexer.Kind, string) bool {
	return func(got lexer.Kind, _ string) bool { return got == want }
}
