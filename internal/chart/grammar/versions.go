// ABOUTME: The version tokens toke.c forces into BAREWORDs after package, use and no, as classes over pvm's tokens.
// ABOUTME: Generated in init beside perly.go's literal because one version token has a class per spelling and context.

package grammar

import "tamarou.com/pvm/internal/lexer"

// init declares the BAREWORD a version is. toke.c forces it from the text after `package NAME` (8858,
// force_strict_version) and after `use`/`no` (5435, tokenize_use: force_version, then an empty BAREWORD
// if `;` or `}` follows). pvm emits it as a Number, whose key chart/lex gives a VersionStrict or
// VersionLax Prefix from its text and the byte after it, or as a Quote opening with `v` or a digit for a
// v-string (`v5.36`, `5.10.1`).
//
// The first token of `use VERSION` is a version only where perl accepts what follows: `;`, `}` or the end
// of the file, else `use 5.036 1;` would read as a module list. A version after `use MODULE` is a version
// wherever a number or v-string stands there, except before a comma, which no list opens with. After
// `package NAME` only a strict Number is one. Its v-string, `package Foo v1.2.3`, lexes as Word(v1)
// Operator(.) Number(2.3): a lexer defect (01a1a481-ebef), listed in lexer-defects.txt.
//
// shortcut: `use VERSION MODULE` (`use 5.036 strict`) is refused: toke.c lexes that module as an
// ordinary term, so taking it would also take `use 5.036 foo bar;`, which perl refuses as an
// indirect-object call. And the version BAREWORD is also the term BAREWORD, which no terminal tells
// apart: `use lib 1.2 => 3;`, `use lib 1.2 . 1;` and `use lib 1.2 x 3;` are taken where perl refuses
// them, since only a comma is excluded after it. Upgrade both with a version terminal no list opens with.
// afterUse is the neighbour `use` or `no` makes of the word before it.
var afterUse = []Neighbour{{Kind: lexer.Word, Text: "use"}, {Kind: lexer.Word, Text: "no"}}

func init() {
	versions := []TokenClass{{Kind: ptr(lexer.Number), Prefix: ptr(byte(VersionStrict))}, {Kind: ptr(lexer.Number), Prefix: ptr(byte(VersionLax))}}
	for _, p := range "v0123456789" {
		versions = append(versions, TokenClass{Kind: ptr(lexer.Quote), Prefix: ptr(byte(p))})
	}
	notVersion := []Neighbour{{Kind: lexer.Number}, {Kind: lexer.Quote}}
	endings := []*Neighbour{{Kind: lexer.Semicolon}, {Kind: lexer.CloseBracket, Text: "}", OpensBlock: true}, &Boundary}
	add := func(c TokenClass) { perlyClasses = append(perlyClasses, ClassDecl{"BAREWORD", c}) }

	pkg := versions[0] // only a strict decimal: its v-string is a lexer defect, above
	pkg.Prev2, pkg.NotPrev1 = &Neighbour{Kind: lexer.Word, Text: "package"}, notVersion
	add(pkg)
	for _, kw := range afterUse {
		for _, v := range versions {
			after := v
			after.Prev2, after.NotPrev1, after.NotNext = &kw, notVersion, &Neighbour{Kind: lexer.Operator, Text: ","}
			add(after)
			for _, next := range endings {
				lone := v
				lone.Prev1, lone.Next = &kw, next
				add(lone)
			}
		}
	}
}
