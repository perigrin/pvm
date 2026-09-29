// ABOUTME: Indirect object notation on a class -- `new Foo ARGS` -- read by toke.c's
// ABOUTME: intuit_method rule and built as a Call flagged Indirect, its invocant first.

package parse

import "tamarou.com/pvm/internal/lexer"

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
// The OTHER half -- `WORD $var ...` as `$var->WORD(...)`, unless WORD is a
// known sub or starts uppercase -- is deliberately not read. perl knows every
// sub from its symbol table; this parser knows only the ones it has read
// declared or imported, and a sub imported from a module it could not read
// looks exactly like an unknown word. Measured: reading that half turned
// `ok $x, 'n'` -- Test::More's `ok`, unresolved -- into `$x->ok(...)` and
// made 35 T1 files wrong in canon. A refusal there is honest; a method call
// is a confident wrong tree.
func (p *parser) parseIndirect(word lexer.Token, spelled, text string) *Node {
	if p.noIndirect || isPerlKeyword(text) || p.namedUnaryHere(spelled, text) ||
		infix[text].BP > 0 || modifiers[text] {
		return nil
	}
	next, ok := p.peekAfter(word)
	if !ok {
		return nil
	}
	if next.Kind != lexer.Word {
		return nil
	}
	class := p.text(next)
	if isPerlKeyword(keywordName(class)) || infix[class].BP > 0 || modifiers[class] {
		return nil
	}
	if _, known := p.lookupSub(class); known {
		return nil
	}
	if !p.packages[class] && !interpreterPackages[class] {
		return nil
	}
	if after, ok := p.peekAfter(next); ok && p.text(after) == "=>" {
		return nil
	}
	invocant := &Node{Kind: Term, Text: class, Start: next.Start, End: next.End}

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
	} else if tok, ok := p.peekSignificant(); ok && !endsArgumentList(tok, p.src) {
		if arg := p.parseExpr(bpListOp); arg != nil {
			n.Children = append(n.Children, arg)
		}
	}
	n.End = p.prevEnd()
	return n
}
