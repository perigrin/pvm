// ABOUTME: The version tokens toke.c forces into BAREWORDs after package, use and no, as classes over pvm's tokens.
// ABOUTME: Generated in init beside perly.go's literal because one version token has a class per spelling and context.

package grammar

import (
	"strings"

	"tamarou.com/pvm/internal/lexer"
)

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
// indirect-object call. The version BAREWORD is also the term BAREWORD, which no terminal tells apart,
// so after a version that follows a module only the tokens in notAfterVersion are refused: `-`, `+`,
// `*`, `%`, `&`, `<`, `/` and `<<` can open a term, so `use lib 1.2 - 3;` is taken where perl
// reads it as a term, and `use lib 1.2 x;`, where x is a word, is refused. Upgrade with a version
// terminal no list opens with. And a word after `use`/`no` loses its version reading only: afterUse also
// removes the term reading of a delimited number there, so `(no => 1 , 2)` will be refused once keywords
// auto-quote before `=>` (the fat-comma family's to settle).
// afterUse is the neighbour `use` or `no` makes of the word before it.
var afterUse = []Neighbour{{Kind: lexer.Word, Text: "use"}, {Kind: lexer.Word, Text: "no"}}

// afterWordSlot is the neighbours whose second token after is never the BAREWORD toke.c forces for a
// version: force_strict_version refuses a word after a package name, and a word after use MODULE is a term.
var afterWordSlot = append([]Neighbour{{Kind: lexer.Word, Text: "package"}}, afterUse...)

// notAfterVersion are the next tokens that cannot follow a forced version, because after it toke.c expects a
// term: the comma and the operators that only join two terms. - + * % & < / << can open a term too, so
// the lexer's reading of them decides and they are accepted (see the shortcut above). isa is a word
// toke.c reads as an operator only under its feature, and <= is a readline's < to toke.c here.
var notAfterVersion = func() (ns []Neighbour) {
	for _, o := range strings.Fields(", => . ? == != || // && .. ... =~ !~ ** <=> = += -= *= /= .= %= **= x= ||= &&= //= |= &= ^= <<= >>= -> | ^ >> > >=") {
		ns = append(ns, Neighbour{Kind: lexer.Operator, Text: o})
	}
	for _, w := range strings.Fields("x eq ne lt gt cmp") {
		ns = append(ns, Neighbour{Kind: lexer.Word, Text: w})
	}
	return ns
}()

// notAfterWord are the next tokens a bare word after use MODULE cannot be followed by, as a term: toke.c reads
// `foo 'a'` and `foo 1.2` as a call it refuses, where `foo bar`, `foo $x` and `foo (3)` are method calls and calls.
var notAfterWord = []Neighbour{{Kind: lexer.Number}, {Kind: lexer.Quote}, {Kind: lexer.Quote, Text: IdentQuote},
	{Kind: lexer.Operator, Text: "["}, {Kind: lexer.Operator, Text: `\`}}

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
		// The generic BAREWORD classes leave a word after use MODULE (NotPrev2), and these give it back as the
		// term it is, unless what follows it is not one: it is no version either.
		add(TokenClass{Kind: ptr(lexer.Word), NotText: lexer.Keywords(), Prev2: &kw, NotNext: notAfterWord, Completes: true})
		add(TokenClass{Kind: ptr(lexer.Word), Text: ptr("x"), Prev2: &kw, NotNext: notAfterWord, Completes: true})
		for _, v := range versions {
			after := v
			after.Prev2, after.NotPrev1, after.NotNext = &kw, notVersion, notAfterVersion
			add(after)
			for _, next := range endings {
				lone := v
				lone.Prev1, lone.Next = &kw, next
				add(lone)
			}
		}
	}
}
