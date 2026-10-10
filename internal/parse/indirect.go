// ABOUTME: Indirect object notation on a class -- `new Foo ARGS` -- read by toke.c's
// ABOUTME: intuit_method rule and built as a Call flagged Indirect, its invocant first.

package parse

import (
	"strings"

	"tamarou.com/pvm/internal/lexer"
)

// notePackage records a package or class name, so a bareword naming it reads
// as a class. See parseIndirect.
func (p *parser) notePackage(name string) {
	if p.packages == nil {
		p.packages = map[string]bool{}
	}
	p.packages[name] = true
}

// parseIndirect reads `WORD INVOCANT ARGS` as a method call, or returns nil
// when this word does not start one. The word is the method; it has not been
// consumed.
//
// It is the bareword half of toke.c's S_intuit_method, for a word that is not
// a keyword: `WORD Bareword ...` is `'Bareword'->WORD(...)` when Bareword names
// a package -- toke.c's `gv_stashpvn` test -- and is not a known sub and has
// no `=>` after it.
//
// perl also reads it when WORD is simply an unknown sub, whether or not the
// bareword names a package. That half needs a symbol table this parser has
// only part of: a function imported from a module it could not read looks
// unknown, so `is Pkg::fill(), 'ok'` -- Test::More's `is`, unresolved -- read
// as `'Pkg::fill'->is(...)`, and 13 T1 files went wrong in canon. So the class
// must be a package this parse KNOWS: declared by `package` or `class`, or
// loaded by `use` or `require`. A class it does not know stays a refusal,
// which the issue this implements names as the honest outcome.
//
// Measured on 5.42.0 with -MO=Deparse: `new Foo "a"` is `'Foo'->new('a')`,
// `doit $object "FOO"` is `$object->doit('FOO')`, `sub new {1} new Foo` is
// `new('Foo')`, and under `use v5.36` indirect notation is a syntax error.
//
// The OTHER half -- `WORD $var ...` as `$var->WORD(...)` when WORD has no CV,
// toke.c:8216 -- and the unknown-sub reading of `WORD Bareword` are read
// only while the sub table is complete (symbolsOpen). perl knows every sub
// from its symbol table; this parser knows only the ones it has read declared
// or imported, and a sub imported from a module it could not read looks
// exactly like an unknown word. Measured: reading that half unconditionally
// turned `ok $x, 'n'` -- Test::More's `ok`, unresolved -- into `$x->ok(...)`
// and made 35 T1 files wrong in canon. With no loader, or once an import it
// cannot see into has run, a refusal is still the honest answer.
func (p *parser) parseIndirect(word lexer.Token, spelled, text string) *Node {
	if p.noIndirect || p.isPerlKeyword(text) || p.keywordHere(text) ||
		p.namedUnaryHere(spelled, text) || p.operatorHere(text) || modifiers[text] {
		return nil
	}
	next, ok := p.peekAfter(word)
	if !ok {
		return nil
	}
	// unseen: the word is one perl has no CV for. Knowable only while the
	// sub table is complete (symbolsOpen), and only for an unqualified word,
	// since a loaded module's XS subs are not in the table by their
	// qualified names.
	_, wordKnown := p.lookupSub(spelled)
	unseen := !p.symbolsOpen && !wordKnown && !strings.ContainsAny(spelled, ":'")

	// `WORD $var`: toke.c:8216, a method call on the scalar when the word has
	// no CV. Its arguments follow only if a term does -- in `s2 $f + 1` the
	// `+` is binary and perl reads `$f->s2 + 1`.
	if next.Kind == lexer.Variable && strings.HasPrefix(p.text(next), "$") {
		if !unseen {
			return nil
		}
		return p.finishIndirect(word, spelled, next, false)
	}

	if next.Kind != lexer.Word {
		return nil
	}
	class := p.text(next)
	if p.isPerlKeyword(keywordName(class)) || p.keywordHere(keywordName(class)) ||
		p.operatorHere(class) || modifiers[class] {
		return nil
	}
	if _, known := p.lookupSub(class); known {
		return nil
	}
	// A known package, or any bareword after a word perl has no CV for:
	// intuit_method's `!cv || gv_stashpvn(...)`.
	if !p.packages[class] && !interpreterPackages[class] && !unseen {
		return nil
	}
	if after, ok := p.peekAfter(next); ok && p.text(after) == "=>" {
		return nil
	}
	return p.finishIndirect(word, spelled, next, true)
}

// finishIndirect builds the indirect call whose invocant is the token after
// word. A bareword invocant takes a list after it, as toke.c's METHCALL0
// does; a scalar one only a list that begins with a term.
func (p *parser) finishIndirect(word lexer.Token, spelled string, next lexer.Token, bareword bool) *Node {
	invocant := &Node{Kind: Term, Text: p.text(next), Start: next.Start, End: next.End}

	p.advanceTo(next)
	n := &Node{
		Kind: Call, Text: spelled, Indirect: true, Resolved: true,
		Start: word.Start, Children: []*Node{invocant},
	}
	// The arguments: parenthesised when a `(` follows the invocant, a list
	// otherwise -- toke.c's METHCALL and METHCALL0.
	if open, ok := p.peekSignificant(); ok && p.text(open) == "(" {
		p.advanceTo(open)
		if arg := p.parseCallArgs(); arg != nil {
			n.Children = append(n.Children, arg)
		}
		if close, ok := p.peekSignificant(); ok && p.text(close) == ")" {
			p.advanceTo(close)
		}
	} else if tok, ok := p.peekSignificant(); ok && !endsArgumentList(tok, p.src) &&
		(bareword || startsTerm(tok, p.src)) {
		if arg := p.parseExpr(bpListOp); arg != nil {
			n.Children = append(n.Children, arg)
		}
	}
	n.End = p.prevEnd()
	return n
}

// gatedKeywords are keywords a feature gates, by the feature. Without it each
// is an ordinary word -- `method $obj` is `$obj->method` and `isa
// Local::Child('X')` is `'Local::Child'->isa('X')` on 5.42.0 -- because
// keywords.c's keyword() returns 0 for them unless the feature is on
// (keywords.c:1145, 1217, 1549, 1715, 2927 for class; 353 for isa).
var gatedKeywords = map[string]string{
	"class": "class", "method": "class", "field": "class", "ADJUST": "class", "__CLASS__": "class",
	"isa": "isa",
}

// keywordHere reports whether word is a keyword at this point of the file:
// toke.c reaches intuit_method only for a word that is not. `last TEST16` and
// `require mro` are keywords, never a method on TEST16 or mro.
func (p *parser) keywordHere(word string) bool {
	if feature, gated := gatedKeywords[word]; gated {
		return p.features[feature]
	}
	return lexer.IsKeyword(word)
}

// operatorHere reports whether word is an infix operator at this point:
// `isa` is one only under its feature.
func (p *parser) operatorHere(word string) bool {
	if feature, gated := gatedKeywords[word]; gated {
		return infix[word].BP > 0 && p.features[feature]
	}
	return infix[word].BP > 0
}
