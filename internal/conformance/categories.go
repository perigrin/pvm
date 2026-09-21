// ABOUTME: Maps GLOSSARY.md category names onto this lexer's token kinds.
// ABOUTME: The single point of coupling; another lexer replaces this file alone.
package conformance

import "tamarou.com/pvm/internal/lexer"

// categories translates a glossary category into a predicate over our
// token kinds.
//
// THIS FILE IS THE COUPLING, and it is deliberately the only one. Corpus
// files assert in the glossary's vocabulary ("numeric literal"), never in
// ours (`lexer.Number`), so a project whose lexer has different kinds
// adopts the corpus by rewriting this map and nothing else.
//
// A category may only be added here once GLOSSARY.md defines it, or the
// corpus asserts in a vocabulary with no definition behind it.
var categories = map[string]func(lexer.Kind) bool{
	"numeric literal": is(lexer.Number),
	"string literal":  is(lexer.Quote),
	"word":            is(lexer.Word),
	"variable":        is(lexer.Variable),
	"operator":        is(lexer.Operator),
}

func is(want lexer.Kind) func(lexer.Kind) bool {
	return func(got lexer.Kind) bool { return got == want }
}
