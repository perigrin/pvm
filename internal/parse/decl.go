// ABOUTME: Declarations: my/our/local/state, sub in both body forms, package in both forms.
// ABOUTME: A declaration's target list is not a call — `my ($a, $b)` declares two variables.

package parse

import (
	"errors"
	"fmt"
	"strings"

	"tamarou.com/pvm/internal/lexer"
	"tamarou.com/pvm/internal/types"
)

// declarators are the variable-introducing keywords. `local` is not one
// strictly -- it saves and restores a global rather than creating a lexical --
// but it takes the same target forms and parses identically.
var declarators = map[string]bool{
	"my": true, "our": true, "local": true, "state": true,

	// 5.38's `field $x :param = $default;`. Structurally a declarator with
	// attributes, which `my` also accepts (`my $x :shared`), so it takes the
	// same path rather than one of its own.
	"field": true,
}

// parseDeclaration parses one declaration, or returns nil if this word does
// not start one.
func (p *parser) parseDeclaration(word lexer.Token) *Node {
	text := keywordName(p.text(word))
	if d := p.parsePrefixedSubDecl(word); d != nil {
		return d
	}
	switch {
	case declarators[text]:
		if lex := p.parseLexicalSub(word); lex != nil {
			return lex
		}
		return p.parseVarDecl(word)
	// `method` declares only under the class feature; elsewhere it is an
	// ordinary name -- op/method.t calls a sub named method as
	// `method Pack ("a")`, which is `'Pack'->method('a')`, measured on 5.42.0.
	case text == "sub" || text == "method" && p.features["class"]:
		// `sub {` with no name is an anonymous sub, which is an EXPRESSION
		// and not a declaration. Declining here sends it to parseTerm, whose
		// anon-sub branch (`term.go:126-132`) has always read it correctly --
		// `my $f = sub { 1 }` was right all along because the expression
		// path is reached there.
		//
		// In statement position parseDeclaration ran first and produced a
		// Declaration spanning only `sub`, then a SECOND statement holding
		// an AnonHash for the body. A tree that is not a parse of its
		// source, and it round-trips.
		//
		// perl makes the same test: a name after `sub` is a declaration, a
		// `{` is an anonymous sub.
		if next, ok := p.peekAfter(word); ok && (p.text(next) == "{" ||
			next.Kind == lexer.Prototype || p.text(next) == "(") {
			return nil
		}
		return p.parseSubDecl(word)
	case text == "package":
		return p.parsePackageDecl(word)
	case text == "format":
		return p.parseFormatDecl(word)
	}
	return nil
}

// parseSignature reads a sub's signature as the parenthesised list it is
// shaped like.
func (p *parser) parseSignature(n *Node) {
	if p.typed && p.parseTypedSignature(n) {
		return
	}
	p.inSignature = true
	sig := p.parseTerm()
	p.inSignature = false
	if sig != nil {
		n.Children = append(n.Children, sig)
	}
}

// parseTypedSignature reads a `.pmt` declaration's typed signature, RFC 0001
// "Typed Perl, in `.pmt` only": `(Ref $ref, Str $class = __PACKAGE__)`, each
// parameter a lattice type name, a variable and an optional default, then
// the return type, if one is stated. It is recorded on p.signatures under the
// sub's name.
//
// A signature it cannot read is an error on p.typedErrs, and nothing is
// recorded -- a partial signature would claim parameters the declaration
// does not state. It then reports false with nothing consumed, and the
// parens are read as parseSignature reads any signature.
func (p *parser) parseTypedSignature(n *Node) bool {
	save := p.pos
	open, _ := p.peekSignificant()
	p.advanceTo(open)
	sig := &Node{Kind: List, Paren: true, Start: open.Start}
	var s types.Signature
	var ret *Node
	err := func() error {
		for {
			tok, ok := p.peekSignificant()
			if ok && p.text(tok) == ")" {
				p.advanceTo(tok)
				sig.End = p.prevEnd()
				var err error
				s.Returns, ret, err = p.returnType()
				return err
			}
			param, node, err := p.typedParam()
			if err != nil {
				return err
			}
			s.Params = append(s.Params, param)
			sig.Children = append(sig.Children, node)
			sep, ok := p.peekSignificant()
			switch {
			case !ok:
				return errors.New("signature not terminated")
			case p.text(sep) == ",":
				p.advanceTo(sep)
			case p.text(sep) != ")":
				return fmt.Errorf("parameter %c%s is not followed by `,` or `)`", param.Sigil, param.Name)
			}
		}
	}()
	name, _ := declaredSub(n)
	if err != nil {
		p.typedErrs = append(p.typedErrs, fmt.Errorf("sub %s: %w", name, err))
		p.pos = save
		return false
	}
	n.Children = append(n.Children, sig)
	if ret != nil {
		n.Children = append(n.Children, ret)
	}
	if p.signatures == nil {
		p.signatures = map[string]types.Signature{}
	}
	p.signatures[name] = s
	return true
}

// returnType reads the type a `.pmt` declaration states after its signature,
// `) Object;`, where a definition would have its body. A declaration with no
// type name there states none. The `;` is left for finishBodyOrSemicolon.
func (p *parser) returnType() (types.Type, *Node, error) {
	if tok, ok := p.peekSignificant(); !ok || tok.Kind != lexer.Word {
		return types.Unknown, nil, nil
	}
	typ, _, node, err := p.typeExpr()
	if err != nil {
		return types.Unknown, nil, err
	}
	end, ok := p.peekSignificant()
	switch {
	case !ok:
		return types.Unknown, nil, fmt.Errorf("return type %s not terminated", node.Text)
	case end.Kind != lexer.Semicolon:
		return types.Unknown, nil, fmt.Errorf("return type %s is not followed by `;`", node.Text)
	}
	return typ, node, nil
}

// typeExpr reads a type expression into its Type, its element type, and a
// TypeName node covering it. RFC 0001 "Type names": a lattice type name, a
// union `A|B` with spaces around `|` allowed, or a container type
// `Name[Type]` such as `List[Str]`, whose element is returned beside it.
//
// ponytail: a union of containers ORs their elements and a nested
// container's element is its outer container, until the paper says what
// containers mean.
func (p *parser) typeExpr() (typ, elem types.Type, node *Node, err error) {
	start := -1
	for {
		tok, ok := p.peekSignificant()
		if !ok || tok.Kind != lexer.Word {
			if start < 0 {
				return types.Unknown, types.Unknown, nil, errors.New("want a type name")
			}
			return types.Unknown, types.Unknown, nil, fmt.Errorf("union %s has no type after `|`", p.src[start:p.prevEnd()])
		}
		if start < 0 {
			start = tok.Start
		}
		member, err := types.FromName(p.text(tok))
		if err != nil {
			return types.Unknown, types.Unknown, nil, fmt.Errorf("unknown type name %q", p.text(tok))
		}
		p.advanceTo(tok)
		typ |= member
		if open, ok := p.peekSignificant(); ok && p.text(open) == "[" {
			p.advanceTo(open)
			if end, ok := p.peekSignificant(); ok && p.text(end) == "]" {
				return types.Unknown, types.Unknown, nil, fmt.Errorf("container type %s has no element type", p.src[tok.Start:end.End])
			}
			inner, _, _, err := p.typeExpr()
			if err != nil {
				return types.Unknown, types.Unknown, nil, err
			}
			end, ok := p.peekSignificant()
			if !ok || p.text(end) != "]" {
				return types.Unknown, types.Unknown, nil, fmt.Errorf("container type %s is not closed by `]`", p.src[tok.Start:p.prevEnd()])
			}
			p.advanceTo(end)
			elem |= inner
		}
		bar, ok := p.peekSignificant()
		if !ok || p.text(bar) != "|" {
			break
		}
		p.advanceTo(bar)
	}
	text := string(p.src[start:p.prevEnd()])
	return typ, elem, &Node{Kind: TypeName, Text: text, Start: start, End: p.prevEnd()}, nil
}

// typedParam reads one parameter of a typed signature, `Str $class =
// __PACKAGE__`, into its Param and the node that covers it: a Declaration
// whose text is the type name, as `my`'s is its declarator, holding the
// variable and the default.
func (p *parser) typedParam() (types.Param, *Node, error) {
	tok, ok := p.peekSignificant()
	if !ok {
		return types.Param{}, nil, errors.New("signature not terminated")
	}
	if tok.Kind != lexer.Word {
		return types.Param{}, nil, fmt.Errorf("want a type name, got %q", p.text(tok))
	}
	typ, elem, tn, err := p.typeExpr()
	if err != nil {
		return types.Param{}, nil, err
	}
	v, ok := p.peekSignificant()
	if !ok || v.Kind != lexer.Variable || v.End-v.Start < 2 || !strings.ContainsRune("$@%", rune(p.src[v.Start])) {
		return types.Param{}, nil, fmt.Errorf("type %s names no variable", tn.Text)
	}
	p.advanceTo(v)
	param := types.Param{Name: p.text(v)[1:], Sigil: p.src[v.Start], Type: typ, Element: elem, Required: p.src[v.Start] == '$'}
	node := &Node{Kind: Declaration, Text: tn.Text, Start: tok.Start, Children: []*Node{
		{Kind: Term, Text: p.text(v), Start: v.Start, End: v.End},
	}}
	if eq, ok := p.peekSignificant(); ok && p.text(eq) == "=" {
		p.advanceTo(eq)
		var def *Node
		if next, ok := p.peekSignificant(); ok && p.text(next) != "," && p.text(next) != ")" {
			def = p.parseExpr(bpBelowComma)
		}
		if def == nil {
			return types.Param{}, nil, fmt.Errorf("default for %s has no expression", p.text(v))
		}
		node.Children = append(node.Children, def)
		// A `die` default runs only when the argument is omitted, so it
		// makes the parameter required rather than optional (RFC 0001, "A
		// required argument defaults to `die`").
		param.Required = def.Kind == Call && keywordName(def.Text) == "die"
		if !param.Required {
			param.Default = string(p.src[def.Start:def.End])
		}
	}
	node.End = p.prevEnd()
	return param, node, nil
}

// hasHead reports whether a sub declaration already holds a prototype or a
// signature -- a parenthesised child.
func hasHead(n *Node) bool {
	for _, c := range n.Children {
		if c.Kind == PrototypeNode || c.Paren {
			return true
		}
	}
	return false
}

// parseLexicalSub: `my sub NAME`, `our sub NAME`, `state sub NAME`, or nil
// when the declarator is not followed by a named sub.
//
// The declarator path handed `sub` to the anonymous-sub term, which takes no
// name, so the name fell out as a call -- `our sub h;` canon'd as
// `our sub();h();` at Unknown=0. Each form is valid on 5.42.0; lexical subs
// need no feature since 5.26.
//
// The sub is read by parseSubDecl, so a lexical sub gets the name, prototype
// or signature, attributes and body a package sub does, and is declared for
// the calls below it the same way. The declarator wraps it as a Declaration
// whose one child is the sub, which canon emits as `my sub NAME ...`.
//
// A `{` after `sub` is an anonymous sub -- `my $f = sub {...}` never reaches
// here, and `my sub {` is not perl -- so a name is required.
func (p *parser) parseLexicalSub(word lexer.Token) *Node {
	if p.text(word) == "local" || p.text(word) == "field" {
		return nil
	}
	sub, ok := p.peekAfter(word)
	if !ok || sub.Kind != lexer.Word {
		return nil
	}
	// `my method NAME` under class syntax is the lexical method, 5.42's, as
	// `my sub NAME` is the lexical sub.
	if kw := p.text(sub); kw != "sub" && (kw != "method" || !p.features["class"]) {
		return nil
	}
	if name, ok := p.peekAfter(sub); !ok || name.Kind != lexer.Word {
		return nil
	}
	p.advanceTo(word)
	inner := p.parseSubDecl(sub)
	return &Node{
		Kind: Declaration, Text: p.text(word),
		Start: word.Start, End: inner.End, Children: []*Node{inner},
	}
}

// parseFormatDecl: `format NAME =` followed by an opaque body.
//
// A declaration rather than a statement kind of its own, for the reason
// Declaration's doc gives: the declarator is in Text, and every consumer that
// cares reads it there. It sits beside `package` because it shares the shape
// -- a keyword, an optional bareword name, and a terminator that is not a
// `;`.
//
// The terminator is what makes this form unlike any other. A format
// declaration ends at a line holding a lone `.`, and the lexer has already
// found it: `scanFormatBody` (`lexer/pod.go:106`) emits ONE FormatBody token
// running from the newline after the `=` through that `.`, with no token
// inside it. So there is nothing to parse here beyond consuming it -- and
// nothing to parse INSIDE it, which is the point. `@<<<<<` in a picture line
// is a left-justified column, not an array sigil followed by two left shifts,
// and tier 13's whole thesis is that the region is delimited without being
// lexed.
//
// No `;` is consumed, because perl does not want one -- the `.` line ends the
// declaration and the next line is the next statement. Measured on perl
// 5.42.0, the declaration compiles to nothing at all:
//
//	$ perl -MO=Concise -e 'format STDOUT =
//	> a fixed report line
//	> .
//	> write;'
//	3  <0> enterwrite v ->4
//
// One op, and it belongs to `write`. The picture lines are compiled into a
// format no op mentions, which is why the only assertion available about the
// body is a token fact.
//
// The name is optional: `format =` is the STDOUT default, and perl accepts it.
func (p *parser) parseFormatDecl(word lexer.Token) *Node {
	p.advanceTo(word)
	n := &Node{Kind: Declaration, Text: p.text(word), Start: word.Start}

	if name, ok := p.peekSignificant(); ok && name.Kind == lexer.Word {
		p.advanceTo(name)
		n.Children = append(n.Children, &Node{
			Kind: Term, Text: p.text(name),
			Start: name.Start, End: name.End,
		})
	}

	// The `=`. An ordinary Operator token, as `conformance/README.md`
	// records: the introducer is two words and an operator, and only the body
	// is opaque.
	if eq, ok := p.peekSignificant(); ok && p.text(eq) == "=" {
		p.advanceTo(eq)
	}

	// The body, consumed whole and never descended into.
	if body, ok := p.peekSignificant(); ok && body.Kind == lexer.FormatBody {
		p.advanceTo(body)
		n.Children = append(n.Children, &Node{
			Kind: Term, Start: body.Start, End: body.End,
		})
	}

	n.End = p.prevEnd()
	return n
}

// parseVarDecl: `my $x`, `my ($a, $b) = @_`, and the rest.
//
// The target is parsed as an expression above the comma, so `my ($a, $b)`
// yields the parenthesised list and `= @_` is picked up by the caller as an
// ordinary assignment. That keeps the declaration from re-implementing
// assignment, which is already at level 9.
func (p *parser) parseVarDecl(word lexer.Token) *Node {
	n := p.parseVarDeclNoSemi(word)
	// The terminating `;` belongs to the declaration, like any other
	// statement's. Without it the statement ends before the semicolon and
	// the leftover becomes an Unknown sitting beside a perfectly good tree.
	if tok, ok := p.peekSignificant(); ok && tok.Kind == lexer.Semicolon {
		p.advanceTo(tok)
		n.End = p.prevEnd()
	}
	return n
}

// parseVarDeclNoSemi is parseVarDecl without consuming a terminator.
//
// The init clause of a C-style for head is a declaration whose `;` is the
// head's SEPARATOR, not the declaration's terminator: `for (my $i = 0; ...)`.
// Eating it there loses the head's structure.
func (p *parser) parseVarDeclNoSemi(word lexer.Token) *Node {
	p.advanceTo(word)
	n := &Node{Kind: Declaration, Text: p.text(word), Start: word.Start}

	// A typed lexical: `my Foo $f`, `our Foo::Bar ($a, $b)`. toke.c's KEY_my
	// branch takes a package name after `my`, `our` or `state` as the
	// variable's class (find_in_my_stash) -- `local` takes none. Read as a
	// call, the class split the statement: `my Foo();$f;`.
	switch keywordName(p.text(word)) {
	case "my", "our", "state":
		if class, ok := p.peekSignificant(); ok && class.Kind == lexer.Word {
			if next, ok := p.peekAfter(class); ok &&
				(next.Kind == lexer.Variable || p.text(next) == "(") {
				p.advanceTo(class)
				n.Children = append(n.Children, &Node{
					Kind: TypeName, Text: p.text(class),
					Start: class.Start, End: class.End,
				})
			}
		}
	}

	// The variable, then any attributes, then an optional initialiser.
	// Attributes come between: `field $x :param = 1` and `my $x :shared = 1`
	// both put them after the name, where an expression parser would read
	// the `:` as a ternary's colon with nothing to match.
	if v, ok := p.peekSignificant(); ok && v.Kind == lexer.Variable {
		p.advanceTo(v)
		target := &Node{
			Kind: Term, Text: p.text(v), Start: v.Start, End: v.End,
		}

		// A subscript belongs to the TARGET, not to a statement of its own.
		// `local $SIG{__WARN__} = sub {...}` stopped at the variable and left
		// `{__WARN__} = ...` to be re-read in statement position, where a
		// brace means an anonymous hash -- so the tree said a hash was being
		// constructed where one was being indexed.
		//
		// This is the failure mode the `sub {` branch above already
		// describes: a declarator that stops mid-expression, producing a tree
		// that is not a parse of its source AND round-trips. Found by
		// canonical re-emission (01a0a8d8), which re-parses the emission and
		// sees a different tree.
		for {
			open, arrow, ok := p.peekSubscript()
			if !ok {
				break
			}
			target = p.parseSubscript(target, open)
			target.Arrow = arrow
		}
		n.Children = append(n.Children, target)
		p.parseAttributes(n)

		// The initialiser, if there is one. ANY assignment operator, not just
		// `=`: `local $ENV{PATH} .= $x` is valid perl, and accepting only `=`
		// left the `.= ...` as an Unknown beside a complete target.
		//
		// Level 9 is the whole assignment class in one token group
		// (toke.c:250), so the test is the level rather than a list of
		// nineteen spellings.
		hadInit := false
		if eq, ok := p.peekSignificant(); ok && infix[p.text(eq)].Level == assignLevel {
			hadInit = true
			p.advanceTo(eq)
			// Parsed above the three word operators rather than at 0, so
			// `and`, `or` and `xor` stop here and belong ABOVE the
			// declaration instead of inside it. That gap is a different
			// program, not a different grouping. Measured on perl 5.42.0:
			//
			//	$ perl -MO=Deparse -e 'my $a=1; my $b=0; my $y = $a and $b;'
			//	$b if my $y = $a;
			//
			// At 0 the `and` was swallowed and the tree said `my $y = ($a
			// and $b)`, which is what the PARENTHESISED source means. The
			// two forms had one tree between them.
			//
			// The floor is bpBelowComma and NOT the assignment's own right
			// power, which would also have excluded the COMMA at 80 --
			// measured, `my $x = 1, my $y = 2` is legal perl and deparses
			// unchanged, so the comma must still be reached from here.
			if init := p.parseExpr(bpBelowComma); init != nil {
				n.Children = append(n.Children, init)
			}
		}
		n.End = p.prevEnd()

		// With NO initialiser, any other infix operator takes the whole
		// declaration as its left operand. Measured on 5.42.0, each
		// deparses unchanged: `my $s =~ /x/;`, `my $t . 'a';`, `our @a ==
		// 1;`. Left unconsumed, each refused as trailing_tokens. Read up to
		// the comma and no further, so the word operators below it still
		// reach the test that follows, and `my $x, my $y` keeps the reading
		// it already had.
		left := n
		if !hadInit {
			if op, ok := p.peekSignificant(); ok {
				text := p.text(op)
				if info, isInfix := infix[text]; isInfix && info.Level != assignLevel &&
					text != "," && text != "=>" && info.BP > bpBelowComma {
					left = p.parseInfix(n, bpBelowComma)
				}
			}
		}

		// A word operator below the comma now stands unconsumed, with the
		// whole declaration as its left operand. Resuming the loop here is
		// what puts it above rather than inside.
		if p.atOperatorBelowComma() {
			return p.parseInfix(left, 0)
		}
		return left
	}

	// A list target with ATTRIBUTES: `my ($c, @g, %b) : teapots = qw[...]`.
	// They sit between the list and the initialiser, as a scalar's do, so
	// the shape is the scalar path's -- target, attributes, initialiser --
	// which canon already writes. Measured on 5.42.0, Deparse keeps
	// `my($c, @g, %b) :teapots = ('a', 'b', 'c')`.
	if d := p.listDeclWithAttributes(n); d != nil {
		return d
	}

	// A list target -- `my ($a, $b) = @_` -- or anything else, parsed as one
	// expression so the assignment at level 9 is already inside it.
	if target := p.parseExpr(0); target != nil {
		n.Children = append(n.Children, target)
	}
	n.End = p.prevEnd()
	return n
}

// listDeclWithAttributes reads `(LIST) :ATTR... [= INIT]` into n, or returns
// nil with nothing consumed when the list has no attribute after it.
func (p *parser) listDeclWithAttributes(n *Node) *Node {
	open, ok := p.peekSignificant()
	if !ok || p.text(open) != "(" {
		return nil
	}
	save := p.pos
	list := p.parseTerm()
	colon, ok := p.peekSignificant()
	if list == nil || !ok || p.text(colon) != ":" {
		p.pos = save
		return nil
	}
	if name, ok := p.peekAfter(colon); !ok || name.Kind != lexer.Word {
		p.pos = save
		return nil
	}
	n.Children = append(n.Children, list)
	p.parseAttributes(n)
	if eq, ok := p.peekSignificant(); ok && infix[p.text(eq)].Level == assignLevel {
		p.advanceTo(eq)
		if init := p.parseExpr(bpBelowComma); init != nil {
			n.Children = append(n.Children, init)
		}
	}
	n.End = p.prevEnd()
	return n
}

// parseSubDecl: `sub NAME [PROTO] ( BLOCK | ";" )`, spec §5.5.2.
//
// The bodiless form is §0.13 rank 4, 7 corpus files. For a fresh parser it is
// just the `;` alternative of the grammar -- the "statement-level leak" the
// findings describe was a tree-sitter recovery artifact, not a subtlety of
// the language.
func (p *parser) parseSubDecl(word lexer.Token) *Node {
	p.advanceTo(word)
	n := &Node{Kind: Declaration, Text: p.text(word), Start: word.Start}

	// The name. An anonymous sub has none, and `sub { ... }` is an
	// expression rather than a declaration -- but it reaches here only as a
	// statement, where perl also treats it as a declaration of nothing.
	if name, ok := p.peekSignificant(); ok && name.Kind == lexer.Word {
		p.advanceTo(name)
		n.Children = append(n.Children, &Node{
			Kind: Term, Text: p.text(name),
			Start: name.Start, End: name.End,
		})
	}

	// The prototype, if the lexer found one. Recognition is its job; this
	// only carries the result so M4 has something to resolve.
	sigFirst := false
	if proto, ok := p.peekSignificant(); ok && proto.Kind == lexer.Prototype {
		p.advanceTo(proto)
		n.Children = append(n.Children, &Node{
			Kind: PrototypeNode, Text: p.text(proto),
			Start: proto.Start, End: proto.End,
		})
	} else if open, ok := p.peekSignificant(); ok && p.text(open) == "(" {
		// A SIGNATURE. With `use v5.36` the lexer declines to scan a
		// prototype (`proto.go:32-35`) because the two are different things
		// -- measured, `prototype(\&g)` is undef for a signatured sub -- and
		// nothing downstream picked the parens up. The body was then read as
		// a hash slice OF the signature and the sub body was LOST, so
		// turning the feature ON made the parse worse.
		//
		// Read as an ordinary parenthesised list, which is what a signature
		// is shaped like: `($x, $y)`, `($x = 1)`, `(@rest)`. Whether each
		// element is a parameter, and what its default means, is M2's --
		// this milestone owns the syntax.
		p.parseSignature(n)
		sigFirst = true
	}

	// Attributes: `sub f :lvalue { 1 }`, `sub f :prototype($$) { 1 }`. They
	// sit between the prototype and the body, which is the order perl's own
	// grammar has (perly.y's subrout: `SUB subname startsub proto subattrlist
	// subbody`).
	attrsAt := len(n.Children)
	p.parseAttributes(n)

	// Typed Perl is read with signatures on, so it keeps their order and
	// refuses a signature before the attributes as perl does (RFC 0001,
	// "Declaration order: Perl's"). The signature is not recorded.
	if p.typed && sigFirst && len(n.Children) > attrsAt {
		name, _ := declaredSub(n)
		delete(p.signatures, name)
		p.typedErrs = append(p.typedErrs, fmt.Errorf("sub %s: subroutine attributes must come before the signature", name))
	}

	// Under the signatures feature the order is the other way round: the
	// attributes come FIRST and the signature after them, and perl rejects
	// the prototype order -- "Subroutine attributes must come before the
	// signature". Measured on 5.42.0, `sub t106 :prototype(@) ($a) { $a }`
	// deparses unchanged. Read here only when no head came before the
	// attributes.
	if !hasHead(n) {
		if open, ok := p.peekSignificant(); ok && p.text(open) == "(" {
			p.parseSignature(n)
		}
	}

	// The declaration enters scope HERE, before its own body and before
	// anything below it is read. That ordering is perl's rule, not a
	// convenience: perl parses top to bottom, so a call ABOVE the
	// declaration is a syntax error and a call below it is a call.
	// Measured on 5.42.0:
	//
	//	$ perl -e 'sub f { } f 1, 2;'      syntax OK
	//	$ perl -e 'f 1, 2; sub f { }'      Do you need to predeclare "f"?
	//
	// `conformance/mdtest/argument-extent.md`'s third case pins the second
	// form, and a pre-pass over the whole file would parse it as a call and
	// be WRONG in the oracle's sense. The declaration a call can see is the
	// one already read.
	//
	// Recorded before finishBodyOrSemicolon, which is one shape MORE
	// permissive than perl and deliberately so. perl does not put the name in
	// scope for a parenless call inside its OWN body -- measured,
	// `sub f { f 1 if 0 } f 2;` is `Do you need to predeclare "f"?`, and it
	// takes a `sub f;` above to make that body legal. Recording after the
	// body would need the name held aside and replayed, and the only source
	// it would change is a self-recursive parenless call, which perl rejects:
	// declining it would be CORRECT and accepting it is `wider`, not WRONG in
	// a way the gate counts. Named here rather than fixed because the shape
	// does not occur: measured over all 1,042 T1 and T2 files, there are ZERO
	// parenless calls to the enclosing sub's own name. The shape to look for,
	// should one arrive, is exactly that.
	//
	// Unconditional on the resolver. parseRoot lifts the same facts onto the
	// root under `res != nil` for the `Imports()` accessor, and that path
	// runs after the whole file is parsed -- too late for any call site, and
	// absent entirely from `Parse`, which is what the T1 and T2 ratchets
	// measure with. Measured: 316 of T2's 354 parenless-call refusals have a
	// callee declared by a `sub NAME` in the same file, and 0 have one
	// reachable by import, because no T2 file uses Test::More.
	p.declareSub(n)

	p.finishBodyOrSemicolon(n)
	return n
}

// declareSub records a `sub NAME` into the shape table, so a call to it
// further down the file knows how far its arguments run.
//
// A sub with no name declares nothing: `sub { ... }` in statement position
// reaches parseSubDecl and names no callee.
func (p *parser) declareSub(n *Node) {
	name, proto := declaredSub(n)
	if name == "" {
		return
	}
	if p.imports == nil {
		p.imports = map[string]Import{}
	}
	p.imports[subKey(name)] = Import{
		Name:           name,
		Prototype:      proto,
		PrototypeKnown: true,
		Local:          true,
	}
	// `sub Pack::method` creates the package Pack as surely as `package
	// Pack` does, so `method Pack (...)` can name it -- op/method.t:54.
	if i := strings.LastIndex(name, "::"); i > 0 {
		p.notePackage(name[:i])
	}
}

// parsePackageDecl: `package NAME;` and `package NAME { ... }`.
//
// The two differ in scope -- to the end of the enclosing block, or to its own
// braces -- which is a later concern. Here they differ only in which
// alternative finishes them.
func (p *parser) parsePackageDecl(word lexer.Token) *Node {
	p.advanceTo(word)
	n := &Node{Kind: Declaration, Text: p.text(word), Start: word.Start}

	if name, ok := p.peekSignificant(); ok && name.Kind == lexer.Word {
		p.advanceTo(name)
		n.Children = append(n.Children, &Node{
			Kind: Term, Text: p.text(name),
			Start: name.Start, End: name.End,
		})
		p.notePackage(p.text(name))
	}

	// An optional version: `package Foo 1.0;`.
	if v, ok := p.peekSignificant(); ok && v.Kind == lexer.Number {
		p.advanceTo(v)
		n.Children = append(n.Children, &Node{
			Kind: Term, Text: p.text(v),
			Start: v.Start, End: v.End,
		})
	}

	p.finishBodyOrSemicolon(n)
	return n
}

// finishBodyOrSemicolon consumes the `( Block | ";" )` that ends a sub or
// package declaration, appending the block when there is one.
func (p *parser) finishBodyOrSemicolon(n *Node) {
	if tok, ok := p.peekSignificant(); ok {
		switch {
		case tok.OpensBlock && p.text(tok) == "{":
			n.Children = append(n.Children, p.parseBlock(tok))
		case tok.Kind == lexer.Semicolon:
			p.advanceTo(tok)
		}
	}
	n.End = p.prevEnd()
}
