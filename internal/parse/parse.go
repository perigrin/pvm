// ABOUTME: The parse tree: nodes over byte spans, and the Unknown that declines rather than guesses.
// ABOUTME: Package parse, not ast — internal/parser is the tree-sitter wrapper and will be for a while.

package parse

import "tamarou.com/pvm/internal/lexer"

// Kind names what a node is.
type Kind int

const (
	// SourceFile is the root. It always exists, including for empty input:
	// an empty program is a valid program, and a nil root would make every
	// consumer check for one.
	SourceFile Kind = iota

	// Statement is one statement, whatever its form.
	Statement

	// Unknown is a construct this parser does not understand, spanning
	// exactly the bytes it could not parse.
	//
	// Unknown is the milestone's whole discipline in one node kind. A parser
	// that guesses at an unfamiliar construct produces a tree that is not a
	// parse of its source, and the fidelity harness scores that WRONG -- the
	// only bucket that fails a build. Emitting Unknown instead is a refusal,
	// and a refusal is never a wrong answer.
	//
	// But a refusal only counts as one if it is REPORTED. See Sites: an
	// Unknown that says nothing to the harness is scored as claiming there
	// is nothing there, which is a silent wrong answer rather than a
	// decline. The node and the hedge are one mechanism in two places.
	Unknown

	// Trivia is whitespace, a comment, or POD: bytes that belong to no
	// statement but must still be in the tree, because every byte is.
	Trivia

	// Term is a leaf value: a literal, a variable, a bareword.
	Term

	// Binary is an infix operator and its two operands.
	Binary

	// Unary is a prefix operator and its one operand. The `\` of §4.4.5 is
	// one of these, and it is the srefgen the fidelity harness measures.
	Unary

	// Postfix is `++` or `--` after its operand.
	Postfix

	// Ternary is `?:`, with condition, then-branch and else-branch.
	Ternary

	// CmpChain is a chained comparison, 5.32+. One n-ary node rather than
	// nested binaries, because each operand is evaluated once:
	//
	//	$ perl -MO=Deparse,-p -e 'my $q = $a < $b < $c'
	//	(my($q) = ($a < $b < $c));
	CmpChain

	// List is a parenthesised or comma-separated list.
	List

	// AnonArray is `[...]` and AnonHash is `{...}` in term position. Anonhash
	// is one of the five oracle markers.
	AnonArray
	AnonHash

	// Index is a subscript or dereference: `$r->[0]`, `$h{k}`, `@a[0,1]`.
	Index

	// Declaration is `my`/`our`/`local`/`state`, `sub` or `package`.
	//
	// One kind with the declarator in Text rather than six kinds: every
	// consumer that cares which one it is reads Text anyway, and a kind per
	// keyword makes the switch in every walker longer without making any of
	// them more precise.
	Declaration

	// PrototypeNode carries a prototype the lexer recognised, so M4 has
	// something to resolve against. Named for the node because Prototype is
	// the lexer's token kind.
	PrototypeNode

	// Conditional is `if`/`unless` with its branches. Text is the keyword,
	// so `unless` is not rewritten into a negated `if` -- the CST records
	// what was written, and an LSP renaming or formatting it needs the
	// original spelling.
	Conditional

	// Loop is `while`, `until`, `for` or `foreach`. One kind with the
	// keyword in Text, for the same reason: the C-style and list forms of
	// `for` differ in their head, which is in the children, not in the kind.
	Loop

	// Label is `NAME:` before a statement.
	//
	// Labels stack -- `A: B: for (...)` is two of them -- and only the
	// innermost attaches to the loop, so this is a list of siblings rather
	// than a field on Loop. Every one is kept because hover and
	// go-to-definition need them; resolving last/next/redo against the
	// innermost is a later concern.
	Label

	// LoopControl is `last`, `next` or `redo` with an optional label.
	// perly.y's LOOPEX, level 2: a statement form, not an expression
	// operator.
	LoopControl

	// Use is `use`, `no` or `require` with a module or version and an
	// optional import list.
	//
	// Parsed, not executed. What `use feature` turns ON is a different
	// problem -- threading feature state through the rest of the parse
	// belongs to M2, and this milestone only needs the syntax.
	Use

	// Phaser is `BEGIN`, `END`, `CHECK`, `INIT` or `UNITCHECK` with a block.
	//
	// That BEGIN runs at compile time, and can declare a sub which changes
	// how a LATER line parses, is the undecidability §4.8.3 describes.
	// Call{Resolved:false} absorbs it; this kind is only the syntax.
	Phaser

	// Call is a named unary, a list operator, or a call to a user sub.
	//
	// Resolved says whether this parser knows what it is calling. A builtin
	// is resolved; an undeclared bareword is not, and §4.8.3 is explicit
	// that an unresolved call is a Call rather than an Unknown:
	//
	//	The expression parser's job is to produce a `Call` node with
	//	`Resolved: false` and let a later pass decide.
	//
	// The difference is what the harness scores. Unknown is for constructs
	// with no known shape; a call has a known shape and an unknown callee,
	// and the adapter reports it as an Unresolved site -- `wider`, not
	// WRONG.
	Call

	// Block is `{ ... }` holding statements rather than a value.
	//
	// Distinct from AnonHash, which is the same two bytes holding a list.
	// The lexer's brace stack already made the call -- a `{` read in XState
	// opens a block -- so this kind records a decision rather than repeating
	// one.
	Block
)

func (k Kind) String() string {
	switch k {
	case SourceFile:
		return "source_file"
	case Statement:
		return "statement"
	case Unknown:
		return "unknown"
	case Trivia:
		return "trivia"
	case Term:
		return "term"
	case Binary:
		return "binary"
	case Unary:
		return "unary"
	case Postfix:
		return "postfix"
	case Ternary:
		return "ternary"
	case CmpChain:
		return "cmp_chain"
	case List:
		return "list"
	case AnonArray:
		return "anon_array"
	case AnonHash:
		return "anon_hash"
	case Index:
		return "index"
	case Block:
		return "block"
	case Declaration:
		return "declaration"
	case PrototypeNode:
		return "prototype"
	case Conditional:
		return "conditional"
	case Loop:
		return "loop"
	case Label:
		return "label"
	case LoopControl:
		return "loop_control"
	case Call:
		return "call"
	case Use:
		return "use"
	case Phaser:
		return "phaser"
	}
	return "?"
}

// Node is one node of the parse tree.
//
// Byte offsets rather than line/column, for the reason the lexer's Token
// gives: line and column are derived at the LSP layer, and carrying them
// doubles the state that must stay consistent under edit.
//
// A node's span always covers its children's spans exactly, with no gaps.
// That is what makes the round-trip property hold: concatenating every
// leaf's text reproduces the input, Unknown included.
type Node struct {
	Kind       Kind
	Start, End int
	Children   []*Node

	// loaded names the modules reached while parsing, and is set on the
	// SourceFile root alone. Read through LoadedModules.
	//
	// Unexported because it is not part of the tree: it says what the PARSE
	// did, not what the source contains, and a consumer walking children
	// must not find it hanging off an interior node. It plays no part in the
	// span arithmetic above and never affects round-trip.
	loaded []string

	// imports is what the file's `use` statements brought into scope, plus
	// the subs it declares itself. Root only, same reasoning as loaded.
	// Read through Imports.
	imports map[string]Import

	// Text is the operator or literal that names this node -- "+", "?:",
	// "$x". Empty for nodes whose meaning is entirely their kind and
	// children, such as a List.
	//
	// It duplicates bytes the span already covers, which is deliberate: an
	// interior node's span covers its operands too, so `$a + $b` and
	// `$a - $b` are indistinguishable by span alone. A consumer that had to
	// re-lex the span to learn which operator it got would be re-deriving
	// what the parser already knew.
	Text string

	// Refusal names WHICH construction site produced this Unknown, and is
	// set on Unknown alone. See refusal.go for the ten codes and what each
	// one means.
	//
	// Not folded into Text, which is documented above as the operator or
	// literal NAMING the node and is what canon.go reads for a Binary's
	// `+`. An Unknown has no such text -- canon writes its source bytes
	// back instead -- so the field is free, and that is exactly the
	// argument against reusing it: a consumer reading Text to learn an
	// operator would start seeing refusal codes.
	//
	// Empty is a refusal with no code, which no site produces and
	// TestEveryUnknownSiteHasACode forbids.
	Refusal RefusalCode

	// Resolved is set on a Call whose callee this parser knows -- a builtin,
	// or a sub declared in this file. False means "a call to something I
	// have not seen", which §4.8.3 says is a Call and not an Unknown.
	//
	// Only meaningful for Call; false everywhere else and not read there.
	Resolved bool

	// The five flags below follow Resolved's rule: each is meaningful on one
	// kind of node, false everywhere else, and not read there. A flag that
	// leaks onto nodes it does not describe is worse than no flag, because a
	// consumer cannot tell a real answer from a stray one.
	//
	// Separate bools rather than a bitfield, and the choice is measured:
	// Node is 72 bytes with one bool and 72 bytes with five, because they
	// land in padding the struct already had. A uint8 of bits would also be
	// 72 and read worse.

	// Arrow is set on an Index reached through `->`. §4.14 gives `$h{k}` and
	// `$h->{k}` ONE node with this flag rather than two shapes, because they
	// are one operation -- and they read DIFFERENT VARIABLES, which is why
	// the distinction has to survive. Measured on perl 5.42.0:
	//
	//	%h = (k => 'hash'); $h = {k => 'ref'};
	//	print $h{k};     hash     the hash %h
	//	print $h->{k};   ref      the hashref $h
	//
	// 25.3% of T1's Index nodes are the arrow form.
	Arrow bool

	// Paren is set on a node the source wrapped in grouping parentheses.
	//
	// Two of Perl's rules turn on it, so it is not decoration. §4.10, list
	// repeat versus string repeat -- `("a") x 3` gives three elements and
	// `"a" x 3` gives one. §4.12.2, list versus scalar assignment -- with
	// `sub f {(1,2,3)}`, `my ($x) = f()` is 1 and `my $y = f()` is 3.
	//
	// Not set for a call's parens: `f($y)` has no grouping in it, and
	// marking it would make every call look like a list assignment target.
	Paren bool

	// Fat is set on a list element that was followed by `=>`.
	//
	// §4.5.4: a fat comma quotes the word to its left, so `(a => 1)` is the
	// string "a" where `(a, 1)` is a call to `a`. parseList consumed `,` and
	// `=>` alike, which destroyed the fact inside a List, AnonArray or
	// AnonHash while it survived in call arguments -- inconsistent as well
	// as lossy.
	Fat bool

	// Handle is set on the filehandle slot of `print`, `printf` or `say`.
	//
	// parseFilehandleSlot decides the slot at construction, from the ABSENCE
	// of a comma, and emitted a plain Term -- so the only evidence left was
	// child shape. `internal/infer/infer.go:1155-1165` records what that
	// cost: counting the handle as argument 1 made every typed-handle print
	// a false Str mismatch.
	Handle bool

	// HeredocBody is set on a Term holding a heredoc's body and terminator
	// line, which is the one child whose bytes lie AFTER its statement's
	// `;` rather than inside it.
	//
	// That inversion is the construct: `my $h = <<"EOT";` continues on the
	// same line while the body begins on the next, so the lexer emits the
	// opener in term position and defers the body to the end of the line
	// (`lexer/heredoc.go`). The opener's Term carries the VALUE's spelling
	// and this one carries the value, and a consumer that wants the string
	// needs both.
	//
	// The flag exists because canon cannot infer the ordering from the tree:
	// children are emitted in order and this one has to follow the
	// terminator, on a line of its own. Reading it off the span -- "a child
	// starting after the statement's semicolon" -- would require canon to
	// find the semicolon, which is exactly the byte canon SUPPLIES rather
	// than reads.
	HeredocBody bool
}

// SourceText reconstructs the bytes this node covers, walking the tree.
//
// It is the round-trip invariant made checkable by a consumer rather than
// only by a test. An LSP renders from the tree, so "the tree still holds
// every byte" has to be something the tree itself can answer.
//
// Children do not tile their parent's span: `$a + $b` has two children and
// the `+` and the spaces belong to neither. The gaps between children are
// the parent's own bytes and are emitted in place, which is why an operator
// is NOT made a child node -- an operator is not an operand, and a tree that
// said so would force every consumer to filter it back out.
func (n *Node) SourceText(src []byte) string {
	var b []byte
	var walk func(*Node)
	walk = func(n *Node) {
		at := n.Start
		for _, c := range n.Children {
			if c.Start > at {
				b = append(b, src[at:c.Start]...)
			}
			walk(c)
			at = c.End
		}
		if n.End > at {
			b = append(b, src[at:n.End]...)
		}
	}
	walk(n)
	return string(b)
}

// Parse builds a tree covering every byte of src exactly once.
//
// Like Tokenize, it has no error return and never panics. An LSP sees
// half-typed buffers constantly, and a parser that gives up on the first
// unfamiliar construct is useless to it.
//
// At this stage every statement is Unknown: the skeleton establishes the
// invariants before there is any grammar to test them against. Later issues
// replace Unknown with real nodes one construct at a time, and the round-trip
// test is what keeps each replacement honest.
func Parse(src []byte) *Node {
	return parseRoot(src, nil)
}

// parseRoot is Parse with a resolver threaded through it. A nil resolver
// resolves nothing, which is what Parse promises.
func parseRoot(src []byte, res *resolver) *Node {
	root := &Node{Kind: SourceFile, Start: 0, End: len(src)}
	toks := lexer.Tokenize(src)

	p := &parser{src: src, toks: toks, res: res}
	for p.pos < len(p.toks) {
		before := p.pos
		if n := p.statement(); n != nil {
			root.Children = append(root.Children, n)
		}
		// Forward progress, asserted rather than assumed. The lexer learned
		// this the hard way: scanFormatBody was wrapped in a step() guard and
		// the fuzzer still found a `format =\n` input that spun. A parser
		// loop that can return without consuming is the same bug one layer
		// up, so the guard is here rather than in a comment.
		if p.pos == before {
			p.pos++
		}
	}
	if res != nil {
		root.loaded = res.loaded
		root.imports = p.imports

		// A local sub shadows an import, and its OWN prototype is the one
		// that applies. Done after the whole file is parsed so that
		// `Imports()` reports EVERY sub the file declares, wherever it sits
		// -- a consumer asking "what does this file define" wants the one
		// below the call as much as the one above it.
		//
		// This is NOT what a call site reads, and the distinction is perl's.
		// A call sees only declarations already parsed, which is why
		// parseSubDecl records each one as it is read (decl.go, declareSub):
		// `f 1, 2; sub f { }` is a syntax error in perl and this pass, run
		// after the fact, would wrongly make it a call. The two populations
		// differ deliberately, and only this one is complete.
		for name, proto := range readModule(root).protos {
			if root.imports == nil {
				root.imports = map[string]Import{}
			}
			root.imports[name] = Import{
				Name:           name,
				Prototype:      proto,
				PrototypeKnown: true,
				Local:          true,
			}
		}
	}
	return root
}

// parser is the cursor over the token stream.
type parser struct {
	src  []byte
	toks []lexer.Token
	res  *resolver

	// imports is what this file's `use` statements brought into scope,
	// accumulated as they are parsed and lifted onto the root at the end.
	imports map[string]Import
	pos     int
}

// peekSignificant returns the next non-trivia token without consuming it.
//
// Trivia is skipped for LOOKUP but never for CONSUMPTION: advanceTo moves the
// cursor past the trivia as well as the token, so every byte still lands in
// exactly one node and round-trip holds. A parser that dropped trivia here
// would have to put it back somewhere.
func (p *parser) peekSignificant() (lexer.Token, bool) {
	for i := p.pos; i < len(p.toks); i++ {
		if !isTrivia(p.toks[i].Kind) {
			return p.toks[i], true
		}
	}
	return lexer.Token{}, false
}

// advanceTo consumes through the given token, trivia included.
func (p *parser) advanceTo(tok lexer.Token) {
	for p.pos < len(p.toks) && p.toks[p.pos].Start <= tok.Start {
		p.pos++
	}
}

// text is the token's source bytes.
func (p *parser) text(tok lexer.Token) string {
	return string(p.src[tok.Start:tok.End])
}

// prevEnd is where the last consumed token ended, which is where a node that
// closed on the cursor should end.
func (p *parser) prevEnd() int {
	if p.pos == 0 {
		return 0
	}
	return p.toks[p.pos-1].End
}

// skipToStatementEnd consumes to the next `;` or unmatched `}`, which is
// where recovery resumes.
//
// It stops ON the boundary token, not after trailing trivia: the newline
// after a statement belongs to no statement, and a recovery that swallowed it
// would make the Unknown's span disagree with every other node's.
func (p *parser) skipToStatementEnd() {
	depth := 0
	sawOpen := false
	for p.pos < len(p.toks) {
		tok := p.toks[p.pos]
		p.pos++
		switch tok.Kind {
		case lexer.Operator:
			if tok.End-tok.Start == 1 {
				switch p.src[tok.Start] {
				case '{', '(', '[':
					depth++
					sawOpen = true
				}
			}
		case lexer.CloseBracket:
			// A closer with nothing open belongs to an enclosing construct;
			// consuming past it would swallow the rest of the file.
			if depth == 0 {
				return
			}
			depth--
			// Back to depth zero on a `}`: a block-bearing statement such as
			// `if (..) { .. }` ends at its closing brace with no semicolon.
			// The two openers of that form are why "depth is already zero" is
			// not the test -- by the time the `}` is reached, the `(` has
			// already been balanced.
			if depth == 0 && sawOpen && p.src[tok.Start] == '}' {
				return
			}
		case lexer.Semicolon:
			if depth == 0 {
				return
			}
		}
	}
}

// statement consumes one statement, or one run of trivia.
//
// Trivia comes out as its own node rather than being attached to a
// statement, because "which statement does a blank line between two subs
// belong to" has no good answer and every answer complicates round-trip.
func (p *parser) statement() *Node {
	if isTrivia(p.toks[p.pos].Kind) {
		start := p.toks[p.pos].Start
		end := start
		for p.pos < len(p.toks) && isTrivia(p.toks[p.pos].Kind) {
			end = p.toks[p.pos].End
			p.pos++
		}
		return &Node{Kind: Trivia, Start: start, End: end}
	}

	start := p.toks[p.pos].Start

	// The EMPTY STATEMENT: a lone `;`. Valid Perl and common enough to
	// matter -- `for (;;)`, a stray `;` after a block, and the leading `;`
	// that forces `map { ; $_ }` to be read as a block.
	//
	// Without this the expression parser is handed a `;`, returns nil, and
	// skipToStatementEnd runs PAST the enclosing `}` looking for a
	// terminator it has already gone by. Measured: `map { ; $_ } @a` lost
	// its closing brace into an Unknown, and `{};` did the same.
	if tok, ok := p.peekSignificant(); ok && tok.Kind == lexer.Semicolon {
		p.advanceTo(tok)
		return &Node{Kind: Statement, Start: start, End: p.prevEnd()}
	}

	// Labels come first and they stack: `A: B: for (...)` is two of them.
	// Read here rather than inside the loop forms, because `LOOP: { ... }`
	// labels a bare block too, and because this is the only place that KNOWS
	// a statement may start -- which is what separates a label from the same
	// shape in `$h{LOOP}`.
	labels := p.parseLabels()

	// Everything after the labels starts HERE, not at `start`. An Unknown
	// spanning from `start` would cover the label bytes that are already in
	// the tree as Label children, and SourceText would emit them twice --
	// measured as "SKIP:SKIP: {" across four corpus files.
	if len(labels) > 0 {
		if tok, ok := p.peekSignificant(); ok {
			start = tok.Start
		}
	}

	// An EMPTY statement, now that the labels are read. The same check above
	// runs BEFORE them and so cannot see this one: `HERE: ;` is a label on an
	// empty statement, which perl accepts and Deparse emits back verbatim --
	//
	//	$ perl -MO=Deparse -e 'my $x; HERE: ; goto HERE if !$x++;'
	//	my $x;
	//	HERE: ;
	//	goto HERE unless $x++;
	//
	// and without this it fell to the expression path, which was handed a `;`,
	// refused, and produced an Unknown starting AT the `;`. The label bytes
	// were then in no node at all -- canon of `HERE: ;` was `;`, six bytes
	// short, and the canon of that differed again. It is in perl's own suite
	// at t/class/field.t:289.
	if tok, ok := p.peekSignificant(); ok && tok.Kind == lexer.Semicolon {
		p.advanceTo(tok)
		n := &Node{Kind: Statement, Start: start, End: p.prevEnd()}
		if len(labels) > 0 {
			n.Start = labels[0].Start
			n.Children = labels
		}
		return n
	}

	// A bare block. The lexer's brace stack already decided this `{` opens a
	// block rather than a subscript or an anonymous hash, and says so on the
	// token, so the decision is READ here rather than made a second time.
	if tok, ok := p.peekSignificant(); ok && tok.OpensBlock && p.text(tok) == "{" {
		// A `{` at statement start is USUALLY a block, and sometimes an
		// anonymous hash. perl decides with intuit_curly
		// (toke.c:6698-6842), which peeks past the brace at the first token
		// inside; the lexer's brace stack cannot, because it classifies from
		// the state the `{` was READ in.
		//
		// Measured on perl 5.42.0, and the rule is narrower than it looks:
		//
		//	{a=>1};      +{'a', 1}    a hash: BAREWORD then `=>`
		//	{"a", 1};    +{'a', 1}    a hash: STRING then `,`
		//	{a, 1};      a block      a bareword and a comma is NOT enough
		//	{$k=>1};     a block      a variable is not a key here
		//	{};          a block      empty is a block, not an empty hash
		//
		// So the first token must be a word or a string, and only those two
		// separators qualify. Everything else stays a block, which is what
		// keeps every bare block, `if` body and loop body working.
		if !p.braceOpensAnonHash(tok) {
			return p.withLabels(labels, p.parseBlock(tok), start)
		}
		// A hash falls through to the expression path below, which reads the
		// `{` as the AnonHash parseTerm already builds.
	}

	if tok, ok := p.peekSignificant(); ok && tok.Kind == lexer.Word {
		// NO applyModifier here, and that is issue 01a0dee8: `goto HERE if $x`
		// comes back as three statements with ZERO Unknown nodes -- `goto
		// HERE`, an `if` Conditional with no children, and `$x;`. parseGoto
		// declines the modifier on purpose so this site can apply it (see its
		// comment at control.go:203) and this site never has. Measured at
		// 9325864f and unchanged by 01a0de8b, so it is not that issue's to fix;
		// left alone rather than widened here because `last if $x` also takes
		// `if` as its LABEL, and the two have to move together.
		if c := p.parseControlFlow(tok); c != nil {
			return p.withLabels(labels, c, start)
		}
		if r := p.parseTheRest(tok); r != nil {
			return p.withLabels(labels, r, start)
		}
		if d := p.parseDeclaration(tok); d != nil {
			// A declaration may be the left side of an assignment:
			// `my ($a, $b) = @_`. parseVarDecl parses its target as a full
			// expression, so the `=` is already inside it.
			//
			// It may also carry a modifier: `my $x = 1 if $c`. parseVarDecl
			// stops at the `if`, leaving it here.
			if mod := p.applyModifier(d, start); mod != nil {
				return p.withLabels(labels, mod, start)
			}
			return p.withLabels(labels, d, start)
		}

		// A statement FORM this milestone has not reached is not an
		// expression, and the expression parser must not be handed one.
		//
		// It would not fail if it were. `if ($x) { 1 }` read as an
		// expression gives index(index(if, ($x)), {1}): every node
		// well-formed, the whole thing a fiction. That is the guess the plan
		// forbids -- a tree that is not a parse of its source, which the
		// harness scores WRONG.
		//
		// So the remaining forms are named and declined. The list shrinking
		// as each issue lands is the milestone's progress.
		if statementKeywords[p.text(tok)] {
			p.skipToStatementEnd()
			return &Node{Kind: Unknown, Refusal: UnimplementedStatement, Start: start, End: p.prevEnd()}
		}
	}

	expr := p.parseExpr(0)
	if expr == nil {
		p.skipToStatementEnd()
		return &Node{Kind: Unknown, Refusal: NotAnExpression, Start: start, End: p.prevEnd()}
	}

	// A statement modifier, which inverts the tree: `print if $x` is a
	// conditional whose body is the print. Applied here because the
	// expression parser has already taken everything it can, and a modifier
	// binds looser than anything it could have taken -- looser than `or`.
	if mod := p.applyModifier(expr, start); mod != nil {
		return p.withLabels(labels, mod, start)
	}

	// Anything left BEFORE the terminator was not consumed by the expression
	// parser -- a statement form it does not know. The whole statement
	// becomes Unknown rather than a half-parsed expression next to a
	// mystery: a partial tree claims to be a parse of bytes it did not read,
	// and that is what the harness scores WRONG.
	//
	// This check must come BEFORE the semicolon is consumed. Running it
	// after looks at the NEXT statement's first token, which is never a
	// terminator, so every statement followed by another became Unknown --
	// `$x;\n$y;\n` collapsed into a single Unknown spanning the file.
	//
	// TestSourceFileIsStatements did not catch it: the test asserted only
	// that children TILE the file, which one Unknown satisfies perfectly.
	// It now asserts a Statement exists as well.
	//
	// A `WORD BLOCK` statement is COMPLETE at its `}` and takes no `;`, so what
	// follows it is the next statement rather than a trailing token. perl agrees
	// -- measured on 5.42.0, `defer { print "d" }` and the `if (...) { ... }` on
	// the line after it come back as siblings with no terminator between them.
	// Without this exception `defer.t`'s test 11 refused the pair as one
	// statement, and the `}` of the `if` was left for canon to emit after an
	// `if` that had already taken `($i == 3)` as a modifier condition.
	if tok, ok := p.peekSignificant(); ok && !endsStatement(tok, p.src) &&
		!endsInBlock(expr) {
		p.skipToStatementEnd()
		return &Node{Kind: Unknown, Refusal: TrailingTokens, Start: start, End: p.prevEnd()}
	}

	// Through the terminating `;` if there is one, so the statement owns its
	// punctuation and the next statement starts clean.
	if tok, ok := p.peekSignificant(); ok && tok.Kind == lexer.Semicolon {
		p.advanceTo(tok)
	}

	if expr.Kind == Unknown {
		// The inner refusal's CODE travels out with it. This site widens
		// an Unknown from the expression to the whole statement, which is
		// a change of span and not a change of cause -- stamping a code of
		// its own here would erase the one distinction the codes exist to
		// make, and `my $x = ${};` and `my $x = .5;` would arrive at the
		// corpus indistinguishable.
		//
		// The bodies are consumed even so. They belong to an opener inside
		// the Unknown, and leaving them for the statement loop makes a
		// SECOND Unknown out of bytes the first already accounts for.
		p.takeHeredocBodies()
		return &Node{Kind: Unknown, Refusal: expr.Refusal, Start: start, End: p.prevEnd()}
	}
	// The labels belong on the statement, and this site was dropping them.
	//
	// Every other statement form leaves through withLabels; the plain
	// EXPRESSION statement was built by hand here and its labels went on the
	// floor. `SKIP: print 1;` came back Unknown=0 -- which is why nothing
	// caught it -- with canon emitting `print(1);`, a canon of different
	// source. perl accepts the form (`perl -e 'FOO: print "hi\n";'` prints
	// hi) and Deparse keeps the label.
	//
	// Not routed through withLabels, because this node's span must reach
	// p.prevEnd() to cover the `;` just consumed, and withLabels ends at its
	// child. The label children are the whole of what was missing.
	children := append(append([]*Node{}, labels...), expr)
	n := &Node{Kind: Statement, Start: start, Children: children}
	if len(labels) > 0 {
		n.Start = labels[0].Start
	}
	n.Children = append(n.Children, p.takeHeredocBodies()...)
	n.End = p.prevEnd()
	return n
}

// takeHeredocBodies consumes the heredoc bodies queued by openers on the
// statement just read, in the order their openers appeared.
//
// Called AFTER the `;`, because that is where the bytes are. The lexer reads
// `<<TERM` in term position and defers the body to the end of the line, so
// `print <<A, <<B;` puts the `;` before either body and both bodies after it
// (`lexer/heredoc.go`'s pending queue). A statement's span therefore runs
// past its own terminator whenever it opened one, and that is not a defect to
// be tidied: the bytes have to land in some node, and the statement that
// names them is the only node with a claim.
//
// Left to the statement loop instead, each body reached `parseTerm` as the
// first token of a would-be statement and became an Unknown -- `not_a_term`
// alone, `trailing_tokens` when a real statement followed on the next line
// and got dragged in with it. Measured at 2963c47d, that was three of
// `heredocs.md`'s cases plus the sole survivor in `adjacency-13_opaque.md`.
//
// A loop rather than one token, because openers stack on a line and the
// lexer emits one body per opener.
func (p *parser) takeHeredocBodies() []*Node {
	var bodies []*Node
	for p.pos < len(p.toks) {
		tok, ok := p.peekSignificant()
		if !ok || tok.Kind != lexer.HeredocBody {
			break
		}
		p.advanceTo(tok)
		bodies = append(bodies, &Node{
			Kind: Term, HeredocBody: true,
			Start: tok.Start, End: tok.End,
		})
	}
	return bodies
}

// parseBlock consumes `{ ... }` as a sequence of statements.
//
// Recursive rather than iterative over a depth counter: a block's body is a
// statement sequence with exactly the properties the file level has, and
// writing it twice is how the two drift apart.
func (p *parser) parseBlock(open lexer.Token) *Node {
	p.advanceTo(open)
	n := &Node{Kind: Block, Start: open.Start}

	for p.pos < len(p.toks) {
		tok := p.toks[p.pos]
		if tok.Kind == lexer.CloseBracket && p.src[tok.Start] == '}' {
			p.pos++
			n.End = tok.End
			return n
		}
		before := p.pos
		if child := p.statement(); child != nil {
			n.Children = append(n.Children, child)
		}
		// The same forward-progress guard the file level has, for the same
		// reason: a statement that consumes nothing would spin here rather
		// than there.
		if p.pos == before {
			p.pos++
		}
	}

	// Unterminated. The block still spans what it consumed, so round-trip
	// holds; an LSP sees this constantly while someone is typing.
	n.End = p.prevEnd()
	return n
}

// statementKeywords open a statement form this milestone has not implemented
// yet. Each is owned by a later issue in M1:
//
//	my our local state sub package  declarations
//	if elsif else unless            control flow
//	while until for foreach do      control flow
//	use no require BEGIN END ...    the rest of T2
//	class field method              the rest of T2
//
// They are listed rather than detected because the expression parser cannot
// fail on them -- it reads `if ($x) { 1 }` as a call-then-subscript and
// returns a confident wrong answer. A list that must be shortened as each
// issue lands is the honest form of "not yet": the entry disappears when the
// parser can really read the form.
var statementKeywords = map[string]bool{
	// my, our, local, state, sub, method and package are GONE from this list:
	// the declarations issue landed and parseDeclaration reads them. That is
	// what shortening this list means, and it is the measure of progress
	// through the milestone.

	// if, elsif, else, unless, while, until, for and foreach are GONE: the
	// control-flow issue landed and parseControlFlow reads them.
	// `do` is GONE: parseBlockOperator reads `do BLOCK` as a term, and
	// `do EXPR` is an ordinary named unary. It reached statementKeywords
	// only because it was unimplemented.
	"continue": true,

	// use, no, require, the phasers and class are GONE: parseTheRest reads
	// them. `field` and `method` are read by parseDeclaration.
	//
	// `defer` is GONE: parseWordTerm reads `WORD BLOCK` as a block followed by
	// a list, which is the shape `defer { ... }` has and the one perl gives it
	// under all three of its symbol-table readings. What that entry declined
	// was the UNGATED reading, where the brace group is an anonymous hash whose
	// block runs EAGERLY -- verified on 5.42.0, `sub f { defer { print "D\n" }
	// print "body\n" } f();` prints `D` then `body` and only then fails to find
	// a `defer` method. Gated, it prints `body` then `D`. Same tokens, same
	// statement boundary, opposite execution ORDER, and nothing at compile time
	// tells them apart -- so declining either one declined a shape this parser
	// can read. Running the block at scope exit is M2's.
	//
	// try, catch and finally STAY. They are read as `WORD BLOCK` now too, but
	// `try BLOCK catch (VAR) BLOCK` is a chain whose second clause takes a
	// PARENTHESISED VARIABLE before its block, and that shape is not this one.
	"try": true, "catch": true, "finally": true,

	// Loop controls and `return` take an optional term and are statement
	// forms in perly.y (levels 2 and 7), not expression operators.
	// last, next and redo are GONE: parseLoopControl reads them.
	// return is GONE: parseReturn reads it.
	// goto is GONE: parseGoto reads it, with a bareword as a LABEL rather
	// than a call -- measured, `goto FOO` says "Can't find label FOO".

	// `format` is GONE: parseFormatDecl reads it as a Declaration whose body
	// is the one FormatBody token the lexer already delimited. `write` was
	// never here -- it is an ordinary named unary and always parsed.
}

// braceOpensAnonHash reports whether a `{` at statement start constructs a
// hash rather than opening a block.
//
// This is intuit_curly's lookahead (toke.c:6698-6842) reduced to the two
// shapes perl actually accepts. Measured on perl 5.42.0 with -MO=Deparse:
//
//	{a=>1};      +{'a', 1}    BAREWORD then `=>`
//	{"a", 1};    +{'a', 1}    STRING then `,`
//	{a, 1};      a block      a bareword and a comma is NOT enough
//	{$k=>1};     a block      a variable is not a key here
//	{};          a block      empty is a block, not an empty hash
//
// Narrow on purpose. The default for a statement-start brace is a block, and
// it has to be: every bare block, `if` body and loop body in the corpus is
// one, so a wrong answer here is far more expensive than a missed hash.
func (p *parser) braceOpensAnonHash(brace lexer.Token) bool {
	first, ok := p.peekAfter(brace)
	if !ok {
		return false
	}

	// The key must be a word or a string. A variable, a number or anything
	// else leaves this a block, which is what perl decided above.
	switch first.Kind {
	case lexer.Word:
		// A bareword key needs a FAT comma. `{a, 1}` is a block.
		next, ok := p.peekAfter(first)
		return ok && p.text(next) == "=>"
	case lexer.Quote:
		// A string key takes either separator: `{"a", 1}` and `{"a" => 1}`
		// are both hashes.
		next, ok := p.peekAfter(first)
		if !ok {
			return false
		}
		return p.text(next) == "=>" || p.text(next) == ","
	}
	return false
}

// withLabels wraps a statement's node in a Statement carrying any labels that
// preceded it.
//
// Labels are SIBLINGS of what they label rather than a field on it, because
// they stack and because `A: B: for (...)` puts both in the tree while only
// the innermost resolves. A field would have to be a slice anyway, and a
// slice of nodes beside the statement is the shape perly.y's labfullstmt
// already describes.
func (p *parser) withLabels(labels []*Node, n *Node, start int) *Node {
	if n == nil {
		return nil
	}
	// Heredoc bodies belong to whatever statement form opened them, and
	// EVERY form arrives here -- a declaration, a loop, an `if`, a bare
	// block. Taking them at this one funnel is why the statement forms do
	// not each need a rule: the bodies sit after the form's own terminator,
	// whether that is a `;` or a `}`, so the only place that knows they are
	// next is the place the form has just finished.
	bodies := p.takeHeredocBodies()
	end := n.End
	if len(bodies) > 0 {
		end = bodies[len(bodies)-1].End
	}

	if len(labels) == 0 {
		return &Node{
			Kind: Statement, Start: start, End: end,
			Children: append([]*Node{n}, bodies...),
		}
	}
	// The wrapper starts at the FIRST LABEL, not at the caller's `start`.
	// Both would usually be the same byte, but the labelled node now starts
	// after its labels, and a wrapper that began earlier than its own first
	// child would emit the gap -- the label text -- and then the Label child
	// would emit it again.
	children := append(append([]*Node{}, labels...), n)
	children = append(children, bodies...)
	return &Node{
		Kind: Statement, Start: labels[0].Start, End: end,
		Children: children,
	}
}

// endsStatement reports whether this token closes the statement rather than
// continuing it. A `}` belongs to an enclosing construct, so it stops the
// statement without being consumed by it.
func endsStatement(tok lexer.Token, src []byte) bool {
	return tok.Kind == lexer.Semicolon ||
		(tok.Kind == lexer.CloseBracket && src[tok.Start] == '}')
}

// isTrivia reports whether a token carries no program structure.
//
// A DataSection -- `__END__` or `__DATA__` and everything after it -- is
// trivia for the same reason POD is: the lexer ate the whole region on
// purpose, and its contents are data rather than Perl. It is not a
// statement form. Perl compiles nothing for the marker; what an op stream
// shows is the READ, `gv` and `readline` on the DATA handle, which is a
// statement of its own well before the marker.
//
// Trivia here still means every byte lands in a Trivia node, so the
// section's body survives in the tree -- which matters more for a data
// section than for the rest, because those bytes are what `DATA` yields.
//
// WHAT GUARDS THAT IS NOT CANON. An earlier revision of this comment
// credited the round-trip machinery, and the PAAD gate measured otherwise:
// excluding the section's bytes from the Trivia span leaves
// `TestCanon*` AND `TestCorpus` green. Canon cannot see it twice over --
// `canon.go`'s `case Trivia:` drops trivia unemitted, and
// `canon_test.go`'s `significant()` skips `lexer.DataSection`, so both
// sides of the comparison have already discarded the section.
//
// The guard is `TestDataSectionIsTrivia`'s `last.End != len(src)`
// assertion in `stmt_test.go`, plus a `TestFuzzSeeds` case that catches an
// empty trivia node. Both fail on that mutation. Naming the wrong guard is
// how a guarantee quietly stops being one, so it is named here.
func isTrivia(k lexer.Kind) bool {
	switch k {
	case lexer.Whitespace, lexer.Comment, lexer.Pod, lexer.DataSection:
		return true
	}
	return false
}
