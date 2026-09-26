// ABOUTME: use/no/require, the phasers, and the 5.38 class syntax — the rest of what T2 contains.
// ABOUTME: Syntax only: what `use feature` enables is M2's, and BEGIN's compile-time effect is M4's.

package parse

import "tamarou.com/pvm/internal/lexer"

// phasers are the compile-and-run-phase blocks.
//
// `ADJUST` is 5.38's class phaser. It takes the same shape -- a keyword, a
// block, and no `;` required -- so it takes the same path. That it is only
// legal inside a `class` body is a semantic check and not a syntactic one:
// perl parses it anywhere and then rejects it at compile time with
// "Cannot 'ADJUST' outside of a 'class'", which is M2's to say, not ours.
var phasers = map[string]bool{
	"BEGIN": true, "END": true, "CHECK": true,
	"INIT": true, "UNITCHECK": true, "ADJUST": true,
}

// parseTheRest handles the statement forms this issue owns, or returns nil.
func (p *parser) parseTheRest(word lexer.Token) *Node {
	text := p.text(word)
	switch {
	case text == "use" || text == "no" || text == "require":
		return p.parseUse(word)
	case phasers[text]:
		return p.parsePhaser(word)
	case text == "class":
		return p.parseClass(word)
	}
	return nil
}

// parseUse: `use MODULE LIST;`, `use VERSION;`, and the `no` and `require`
// spellings.
//
// The import list is an ordinary expression, which covers every shape the T2
// corpus actually contains -- surveyed rather than guessed:
//
//	12  use v5.36
//	12  use feature 'class'
//	12  no warnings 'experimental::class'
//	 2  no warnings qw(syntax deprecated)
//	 1  use test_use { () }
//
// The last one is why the list is parsed as an expression rather than as a
// comma-separated list of literals: `{ () }` is a term, and a narrower parser
// would decline it.
func (p *parser) parseUse(word lexer.Token) *Node {
	p.advanceTo(word)
	n := &Node{Kind: Use, Text: p.text(word), Start: word.Start}
	var module string

	// The module name is a BAREWORD, not an expression.
	//
	// Parsing it as one makes `use feature 'class';` leave the string
	// orphaned: `feature` becomes an unresolved Call, which by design
	// consumes no arguments, so `'class'` is never read and the statement
	// falls to Unknown. Measured on five t/class files before this split.
	//
	// So the name is taken directly and the import list is parsed after it.
	if name, ok := p.peekSignificant(); ok && name.Kind == lexer.Word {
		p.advanceTo(name)
		term := &Node{
			Kind: Term, Text: p.text(name),
			Start: name.Start, End: name.End,
		}

		// A version bundle USED TO BE three tokens: `use v5.36;` lexed as
		// Word("v5") Operator(".") Number(36), because `v5` is a valid
		// identifier and the lexer had no reason to know better. Taking only
		// the Word left `.36` behind, which is not an expression, so the
		// statement fell to Unknown -- measured on seven t/class files. The
		// whole run was absorbed into one term instead.
		//
		// `scanVString` now emits ONE Quote for any `v`-prefixed run with a
		// dot, so a version never reaches this branch: the guard below is
		// `name.Kind == lexer.Word` and a Quote misses it. This code and
		// `isVersionPrefix` have no reachable caller -- tracked by 01a0dc74,
		// which also covers the rest of the stranded reassembly surface.
		if isVersionPrefix(term.Text) {
			for {
				dot, ok := p.peekSignificant()
				if !ok || p.text(dot) != "." {
					break
				}
				num, ok := p.peekAfter(dot)
				if !ok || num.Kind != lexer.Number {
					break
				}
				p.advanceTo(num)
				term.End = num.End
			}
		}
		module = term.Text
		n.Children = append(n.Children, term)
	}

	// The import list, or a version. A version is now ONE Quote token
	// (`scanVString`), so it misses the Word-shaped name path above and
	// reads as an expression here -- which round-trips even though the shape
	// is not what perl's grammar calls a version.
	//
	// It formerly lexed as Word("v5") Operator(".") Number, with the Word
	// half taken as the name above and the remainder reassembled. The Term's
	// text was the truncated "v5" then and is the whole "v5.36" now; the
	// span was already correct either way.
	//
	// A `}` ends the statement as surely as a `;` does -- the final semicolon
	// of a block's last statement is optional in perl, and `eval { require
	// Errno }` as a feature probe is how perl's own suite writes it. Asking
	// only `!= Semicolon` read the closer as the start of an import list and
	// the operand hunt then CONSUMED it, taking the `}` out of the enclosing
	// block: 10 files and 77 nodes of perl.git t/, and a wrong tree rather
	// than only a count, because canon then emitted a spurious `};`.
	//
	// endsStatement is the predicate that already answers this, and answers it
	// for the same reason -- "a `}` belongs to an enclosing construct".
	var list *Node
	if next, ok := p.peekSignificant(); ok && !endsStatement(next, p.src) {
		if arg := p.parseExpr(0); arg != nil {
			list = arg
			n.Children = append(n.Children, arg)
		}
	}

	// Resolve the module, if this parse has a loader and the name is a module
	// rather than a version bundle.
	//
	// Enrichment only: whether the source is found changes what a later call
	// KNOWS, never whether this statement parses. `use` is keyword, bareword,
	// optional list, semicolon, and that is true whether or not the file
	// exists. perl dies here; this parser must not, because parsing text is
	// not running it.
	//
	// `no` is deliberately not resolved, because it UNIMPORTS.
	//
	// `require MODULE` is not resolved either, and for the reason that used to
	// cover `require` entirely: it is a runtime load with NO IMPORT AT ALL, so
	// there is no export list to apply and nothing enters the caller's scope
	// by name.
	//
	// `require './FILE.pl'` is the exception, and what makes it one is that
	// perl's own suite writes its helpers that way: a required `.pl` file
	// declares subs into the CALLER'S package -- it has no `package`
	// statement of its own -- so every `sub NAME` in it is a name the
	// requiring file can call. There is no import list to consult because
	// there is no import; the declarations themselves are the interface.
	// See resolveRequiredFile for the boundary that keeps this safe.
	switch {
	case module != "" && !isVersionPrefix(module) && n.Text == "use":
		p.resolveImports(module, list)
	case module == "" && n.Text == "require":
		p.resolveRequiredFile(list)
	}

	if tok, ok := p.peekSignificant(); ok && tok.Kind == lexer.Semicolon {
		p.advanceTo(tok)
	}
	n.End = p.prevEnd()
	return n
}

// resolveImports reads a module's source and records what it brought in.
//
// A module that cannot be found contributes nothing, and calls to what it
// would have exported stay unresolved -- correct, not a gap.
func (p *parser) resolveImports(module string, list *Node) {
	if p.res == nil {
		return
	}
	facts, ok := p.res.resolve(module)
	if !ok {
		return
	}

	// An import list restricts what is imported, and `use M ()` -- an empty
	// list -- is the explicit "load but import nothing" form, which is NOT
	// the same as omitting the list.
	var names []string
	listGiven := list != nil
	if listGiven {
		if got, ok := literalNameList(list); ok {
			names = got
		} else if !list.Paren || len(list.Children) != 0 {
			// A computed import list is opaque: it is not an empty import.
			return
		}
	}

	if p.imports == nil {
		p.imports = map[string]Import{}
	}
	for _, imp := range importsFrom(facts, names, listGiven) {
		p.imports[imp.Name] = imp
	}
}

// resolveRequiredFile reads a `require './helper.pl'` and records every
// `sub NAME` the helper declares, so a parenless call to one of them parses.
//
// list is the require's argument as parsed -- a single Term for the literal
// spellings, anything at all otherwise.
//
// THE BOUNDARY IS THE SAFETY PROPERTY, and each rule is enforced at a named
// place rather than assumed:
//
//  1. A LITERAL string only, here. `require $file` and `require "./$n.pl"`
//     name a file only at runtime, so they contribute nothing. Measured over
//     perl.git's `t/`: 503 files hold a literal `require`, against 6 with a
//     computed one and 3 with an interpolated path, and BOTH kinds occur in
//     files that also hold a literal one. So refusing them costs nothing and
//     accepting them would be a guess at a name.
//  2. Resolved against the file's own directory and then the parse root, in
//     DirLoader. THE PARSE ROOT IS THE LOAD-BEARING HALF, and it was measured
//     to be so: 464 of those 503 files run `chdir 't' if -d 't'` before the
//     require, so `./test.pl` in `t/op/select.t` names `t/test.pl` and NOT
//     `t/op/test.pl`. Resolving against the requiring file's own directory
//     alone moves 0 files and 0 nodes; with `t/` as the root it moves 93 files
//     and 3,267 nodes. The caller supplies the root because only the caller
//     knows the working directory the suite is run from.
//  3. Only `sub NAME`, here: facts.protos and not facts.exports. A `.pl`
//     helper has no `package` and no `@EXPORT`, so its declarations ARE its
//     interface -- but a variable it declares is not a call shape and brings
//     nothing.
//  4. ONE LEVEL, in resolver.resolveFile. A helper's own `require` is not
//     followed, and it COSTS NOTHING -- measured rather than assumed, because
//     one helper does nest: `thread_it.pl` requires `./test.pl` at its line
//     10, so the 13 `*_thr.t` files that require `thread_it.pl` see its subs
//     and not `test.pl`'s. All 13 were ALREADY CLEAN before this change and
//     stayed clean after it: they are wrappers that set up a thread and
//     `require` the real test, with no parenless call site of their own. A
//     second level would buy zero files and zero nodes.
//  5. An unreadable or unparseable helper contributes nothing and NEVER fails
//     the parse -- resolveFile returns "not found", which is the answer a
//     missing module already gets. `t/` files are parsed from many working
//     directories, including by a suite with no perl5 checkout at all.
func (p *parser) resolveRequiredFile(list *Node) {
	if p.res == nil || list == nil {
		return
	}

	// Rule 1, and INTERPOLATION is the half that is easy to miss.
	// literalNameList reads a single quoted string, which is what a literal
	// `require` argument is -- but `"./$name.pl"` is ALSO a single quoted
	// string to it, and names a file only at runtime. A sigil inside a
	// double-quoted body is what separates the two, so it is checked on the
	// text AS WRITTEN, before the quotes come off.
	if interpolates(list.Text) {
		return
	}
	names, ok := literalNameList(list)
	if !ok || len(names) != 1 || !isRequiredPath(names[0]) {
		return
	}

	facts, ok := p.res.resolveFile(names[0])
	if !ok {
		return
	}

	// Rule 3. Every declared sub, with its prototype, because a `.pl` helper
	// declares into the CALLER'S package and has no export list to narrow
	// this. Not Local: the name came from another file, and `Local` marks a
	// sub this file declares itself.
	if p.imports == nil {
		p.imports = map[string]Import{}
	}
	for name, proto := range facts.protos {
		p.imports[name] = Import{
			Name:           name,
			Prototype:      proto,
			PrototypeKnown: true,
		}
	}
}

// isVersionPrefix reports whether a word is the `v5` of a `v5.36`.
func isVersionPrefix(text string) bool {
	if len(text) < 2 || (text[0] != 'v' && text[0] != 'V') {
		return false
	}
	for i := 1; i < len(text); i++ {
		if text[i] < '0' || text[i] > '9' {
			return false
		}
	}
	return true
}

// parsePhaser: `BEGIN { ... }` and the other four.
func (p *parser) parsePhaser(word lexer.Token) *Node {
	p.advanceTo(word)
	n := &Node{Kind: Phaser, Text: p.text(word), Start: word.Start}

	if blk := p.parseBlockOrDecline(); blk != nil {
		n.Children = append(n.Children, blk)
	}
	if tok, ok := p.peekSignificant(); ok && tok.Kind == lexer.Semicolon {
		p.advanceTo(tok)
	}
	n.End = p.prevEnd()
	return n
}

// parseClass: 5.38's `class NAME { ... }` and `class NAME;`.
//
// Structurally a package with a different keyword, so it shares
// finishBodyOrSemicolon. `field` and `method` inside the body are handled by
// parseDeclaration, which already knows `method` as a sub spelling.
func (p *parser) parseClass(word lexer.Token) *Node {
	p.advanceTo(word)
	n := &Node{Kind: Declaration, Text: p.text(word), Start: word.Start}

	if name, ok := p.peekSignificant(); ok && name.Kind == lexer.Word {
		p.advanceTo(name)
		n.Children = append(n.Children, &Node{
			Kind: Term, Text: p.text(name),
			Start: name.Start, End: name.End,
		})
	}

	// An optional version: `class Point 1.0 { }`.
	if v, ok := p.peekSignificant(); ok && v.Kind == lexer.Number {
		p.advanceTo(v)
		n.Children = append(n.Children, &Node{
			Kind: Term, Text: p.text(v), Start: v.Start, End: v.End,
		})
	}

	// Attributes: `class Point :isa(Shape) { }`.
	p.parseAttributes(n)

	p.finishBodyOrSemicolon(n)
	return n
}

// parseAttributes reads a run of `:name` or `:name(args)` and appends them.
//
// Attributes hang off classes, fields, methods and subs alike, and each one
// is an Attribute node: canon has to tell a declaration's HEAD from its
// INITIALISER, and an attribute is head. As a Term it was indistinguishable
// and `my $x :shared = 1;` re-emitted as `my $x = :shared = 1;`.
func (p *parser) parseAttributes(n *Node) {
	for {
		colon, ok := p.peekSignificant()
		if !ok || p.text(colon) != ":" {
			return
		}
		name, ok := p.peekAfter(colon)
		if !ok || name.Kind != lexer.Word {
			return
		}
		p.advanceTo(name)
		attr := &Node{
			Kind: Attribute, Text: ":" + p.text(name),
			Start: colon.Start, End: name.End,
		}

		// An optional parenthesised argument, taken as an opaque span: an
		// attribute's argument is not Perl in general (`:isa(Shape)` is, but
		// `:lvalue` takes none and XS attributes take arbitrary text).
		//
		// `:prototype(...)` arrives as ONE Prototype token, because its
		// contents are a prototype and the lexer scans them the way it scans
		// `sub f ($$)`. Balancing brackets cannot read it: `$)` is a real
		// perl variable and would swallow the closing paren.
		if proto, ok := p.peekSignificant(); ok && proto.Kind == lexer.Prototype {
			p.advanceTo(proto)
			attr.Text += p.text(proto)
			attr.End = proto.End
			n.Children = append(n.Children, attr)
			continue
		}
		if open, ok := p.peekSignificant(); ok && p.text(open) == "(" {
			depth := 0
			for p.pos < len(p.toks) {
				tok := p.toks[p.pos]
				p.pos++
				if p.text(tok) == "(" {
					depth++
				}
				if tok.Kind == lexer.CloseBracket && p.src[tok.Start] == ')' {
					depth--
					if depth == 0 {
						break
					}
				}
			}
			attr.End = p.prevEnd()
			// Text, not just the span. Canon writes Text, so extending only
			// End DROPPED the argument: measured, `class Point :isa(Shape)`
			// re-emitted as `class Point :isa` and lost its parent class.
			// Unknown could not see it -- the tree was right and only the
			// emission was wrong -- so the count stayed at zero.
			attr.Text = string(p.src[attr.Start:attr.End])
		}
		n.Children = append(n.Children, attr)
	}
}

// peekAfter returns the next significant token after the given one.
func (p *parser) peekAfter(tok lexer.Token) (lexer.Token, bool) {
	save := p.pos
	p.advanceTo(tok)
	next, ok := p.peekSignificant()
	p.pos = save
	return next, ok
}
