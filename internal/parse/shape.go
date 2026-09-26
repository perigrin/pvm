// ABOUTME: Parse shape derived from a prototype string — arity, and the leading & that licenses a block.
// ABOUTME: Derived on demand and never stored, because the prototype already is the fact.

package parse

import "strings"

// Shape is how a call to a sub parses: what it may take, and in what form.
//
// DERIVED from the prototype, never stored beside it. A stored shape would be
// a second copy of one fact, and it does not fit either -- `(;$)` is
// zero-or-one, and a leading `&` changes what is legal at the call site
// without changing arity. The string carries both; an enum has a slot for
// neither, so this is computed on demand from the string.
type Shape int

const (
	// ShapeList is a list operator: it takes everything up to the next thing
	// that closes the list. A sub with no prototype is one.
	//
	//	$ perl -MO=Deparse,-p -e 'sub np { } np 1, 2, 3;'
	//	np(1, 2, 3);
	ShapeList Shape = iota

	// ShapeUnary takes a single term, and a following comma belongs to the
	// enclosing list rather than to the call.
	//
	//	$ perl -MO=Deparse,-p -e 'sub one ($) { } one 1, 2;'
	//	(one(1), '???');
	ShapeUnary

	// ShapeNiladic takes nothing, so what follows is an operator rather than
	// an argument: `nil + 1` is `nil() + 1`.
	ShapeNiladic

	// ShapeBlock licenses the `f { ... } LIST` form, which a LEADING `&`
	// grants and nothing else does.
	ShapeBlock
)

func (s Shape) String() string {
	switch s {
	case ShapeList:
		return "list"
	case ShapeUnary:
		return "unary"
	case ShapeNiladic:
		return "niladic"
	case ShapeBlock:
		return "block"
	}
	return "?"
}

// ShapeOf derives the parse shape of a call from a prototype string, parens
// included. An empty string is "no prototype", which is a list operator.
//
// Every rule here was measured against perl 5.42.0 rather than read from a
// manual; the cases are in shape_test.go with their deparse output.
func ShapeOf(proto string) Shape {
	inner := proto
	if len(inner) >= 2 && inner[0] == '(' && inner[len(inner)-1] == ')' {
		inner = inner[1 : len(inner)-1]
	}

	// No prototype at all: a list operator, which is what makes
	// `subtest 'x' => sub {}` parse once the import is known.
	if proto == "" {
		return ShapeList
	}

	// A LEADING `&` licenses the block form, and the position is the whole
	// rule. Measured:
	//
	//	sub bf (&@) { }   bf { $_[0]*2 } (1,2,3)   ->  2,4,6
	//	sub ni ($&) { }   ni { 1 } (2)             ->  syntax error
	//	sub rf (\&) { }   rf { 1 }                 ->  Type of arg 1 ...
	//
	// A backslash makes it a reference slot rather than a block, so the `&`
	// must be the first character of the prototype body.
	if len(inner) > 0 && inner[0] == '&' {
		return ShapeBlock
	}

	// Count the MANDATORY argument slots, and notice a slurpy one.
	//
	// `;` opens the optional group, so slots after it do not raise the count.
	// A `\` escapes the slot that follows -- `\@` is ONE argument and takes a
	// single term, measured:
	//
	//	$ perl -MO=Deparse,-p -e 'sub rf (\&) { } rf \&foo, 2;'
	//	(&rf((\&foo)), 2);      <- the comma is the enclosing list's
	//
	// while a BARE `@` or `%` is slurpy and swallows everything:
	//
	//	$ perl -MO=Deparse,-p -e 'sub sl (@) { } sl 1, 2, 3;'
	//	sl(1, 2, 3);
	slots, optional, slurpy := 0, false, false
	for i := 0; i < len(inner); i++ {
		switch c := inner[i]; c {
		case ';':
			optional = true
		case '\\':
			// A reference slot: one argument whatever follows the backslash,
			// including a `[...]` group of alternatives.
			i++
			if i < len(inner) && inner[i] == '[' {
				for i < len(inner) && inner[i] != ']' {
					i++
				}
			}
			if !optional {
				slots++
			}
		case '@', '%':
			// Slurpy, and only when not escaped -- the `\` case above
			// consumed those already.
			slurpy = true
			if !optional {
				slots++
			}
		case '$', '&', '*', '+':
			if !optional {
				slots++
			}
		}
	}

	switch {
	case slurpy:
		// Anything slurpy takes the whole list, whatever precedes it.
		return ShapeList
	case slots == 0 && !optional:
		// `()` -- takes nothing, so what follows is an operator.
		return ShapeNiladic
	case slots <= 1:
		// One mandatory slot, or none with an optional group: unary at the
		// call site. It takes at most one term and never swallows a
		// following comma.
		return ShapeUnary
	}
	return ShapeList
}

// knowsShape reports whether this parser can say how a call to name parses.
//
// True only for a name whose declaration was actually read -- from an imported
// module's source, from a `sub NAME` earlier in this same file, or from a
// `.pl` file this one `require`d by a literal path. A name known by import but
// with no visible declaration -- the XS tier -- is false, because guessing
// "list operator" for `first` would make `first { $_ > 1 } @a` a syntax error,
// and a wrong shape turns working code into an error.
//
// EARLIER is load-bearing for the in-file route and is perl's own rule: the
// table is filled as declarations are parsed, so a call above a declaration
// does not see it. See decl.go's declareSub.
func (p *parser) knowsShape(name string) bool {
	imp, ok := p.imports[name]
	return ok && imp.PrototypeKnown
}

// blockShapeTakesList reports whether a ShapeBlock prototype has argument slots
// after its leading `&`, so a LIST may follow the block.
//
// `(&)` has none and `(&@)` does. See parseByShape's ShapeBlock for the deparse
// output that draws the line here.
func blockShapeTakesList(proto string) bool {
	inner := proto
	if len(inner) >= 2 && inner[0] == '(' && inner[len(inner)-1] == ')' {
		inner = inner[1 : len(inner)-1]
	}
	// Past the `&` itself; a `;` only opens the optional group, so a slot
	// after it still means a list may be written.
	return len(strings.TrimLeft(inner[1:], ";")) > 0
}

// parseByShape consumes a call's arguments according to its prototype, and
// marks it resolved.
func (p *parser) parseByShape(n *Node, name string) {
	n.Resolved = true

	switch ShapeOf(p.imports[name].Prototype) {
	case ShapeNiladic:
		// `()` takes nothing, so what follows is an operator rather than an
		// argument: `nil + 1` is `nil() + 1`.

	case ShapeUnary:
		// One term. Parsed at the named-unary level so arithmetic binds into
		// the argument and a comma does not: `one 1, 2` is `(one(1), 2)`.
		if arg := p.parseExpr(bpNamedUnary); arg != nil {
			n.Children = append(n.Children, arg)
		}

	case ShapeBlock:
		// A leading `&` licenses `f { ... } LIST`. The block is optional at
		// the call -- `first(sub {...}, @a)` is the same sub called the
		// other way -- so a missing one falls through to the list.
		blk := p.parseBlockArgument()
		if blk != nil {
			n.Children = append(n.Children, blk)
		}
		// A LIST follows only when the prototype has a slot for one. `(&)` is
		// the block and nothing else, so a comma after it belongs to the
		// ENCLOSING list -- the same rule ShapeUnary follows, and perl draws
		// the line in the same place. Measured on 5.42.0:
		//
		//	sub es (&)  {1}  is(es { 2 }, 'x', 'y');
		//	  is &es(sub { 2; }), 'x', 'y';      the commas are is()'s
		//	sub es (&@) {1}  es { 2 } "x", "y";
		//	  &es(sub { 2; }, 'x', 'y');         the commas are es()'s
		//
		// Taking the list unconditionally swallowed the arguments of the call
		// AROUND a `(&)` block-call: `caller_anonymous_callback.t`'s
		// `is(exception_style { ... }, 'main::__ANON__', '...')` lost its two
		// remaining arguments, and the whole statement fell to Unknown.
		//
		// Only when the block was actually taken: without one this is an
		// ordinary parenless call whose arguments are still ahead of it.
		if blk == nil || blockShapeTakesList(p.imports[name].Prototype) {
			if arg := p.parseExpr(bpListOp); arg != nil {
				n.Children = append(n.Children, arg)
			}
		}

	default:
		// A list operator takes everything up to what closes the list.
		if arg := p.parseExpr(bpListOp); arg != nil {
			n.Children = append(n.Children, arg)
		}
	}
}

// parseBlockArgument reads the `{ ... }` of a block-call, or returns nil when
// the next token is not one.
//
// The lexer already decided whether a brace opens a block, so this reads that
// decision rather than repeating it.
func (p *parser) parseBlockArgument() *Node {
	tok, ok := p.peekSignificant()
	if !ok || p.src[tok.Start] != '{' || !tok.OpensBlock {
		return nil
	}
	return p.parseBlockOrDecline()
}
