// ABOUTME: Maps GLOSSARY.md category names onto this lexer's token kinds.
// ABOUTME: The single point of coupling; another lexer replaces this file alone.
package conformance

import (
	"tamarou.com/pvm/internal/lexer"
	"tamarou.com/pvm/internal/parse"
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

	// perl spells some operators with letters -- `x`, `cmp`, `and`,
	// `not` -- and our lexer gives them Kind Word, so the text decides,
	// exactly as it does for the quote pair below.
	//
	// A word-shaped operator is ALSO a `word`, and both categories
	// answer yes. That is the point: `word` is a claim about spelling at
	// the token layer, and this is the narrower claim a file makes when
	// it means perl's operator rather than any identifier.
	//
	// The predicate is `internal/parse`'s, not a list here. That package
	// owns the precedence tables, which ARE the set of perl's operators,
	// so a copy would be a second list free to drift -- the failure the
	// quote-op comment below records.
	"word-shaped operator": func(k lexer.Kind, text string) bool {
		return k == lexer.Word && parse.IsWordShapedOperator(text)
	},

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

	// Tier 13's categories. Each names a construct the OPTREE cannot see
	// -- pod and a data section compile to nothing at all, a format body
	// to nothing until something calls `write` -- so a token fact is the
	// only assertion available about them.
	//
	// `readline operator` covers `<FH>` and `<*.c>` alike, because the
	// lexer cannot tell a handle from a glob pattern without knowing what
	// the name inside means. perl distinguishes them at the optree; the
	// category is named for the syntax the lexer sees.
	"readline operator": kind(lexer.Readline),
	"pod block":         kind(lexer.Pod),
	"data section":      kind(lexer.DataSection),
	"format body":       kind(lexer.FormatBody),
}

// kind matches on the token kind alone, which is the common case.
func kind(want lexer.Kind) func(lexer.Kind, string) bool {
	return func(got lexer.Kind, _ string) bool { return got == want }
}
