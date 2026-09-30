// ABOUTME: Calls: named unaries take one argument, list operators take the whole list.
// ABOUTME: An undeclared callee is Call{Resolved:false}, never Unknown — §4.8.3.

package parse

import (
	"strings"

	"tamarou.com/pvm/internal/lexer"
)

// Binding powers for the two call shapes, spec §4.2.
//
// A named unary is level 19: tighter than comparison at 17, looser than
// arithmetic at 22. That is why `length $x + 1` is `length($x + 1)` and
// `length $x < 5` is `length($x) < 5` -- both measured on the optree.
//
// A list operator is level 7, below the comma at 8, which is exactly how it
// swallows the whole list.
const (
	bpNamedUnary = 190
	bpListOp     = 70

	// A dereference binds tighter than every infix and postfix operator --
	// `->` at level 29 and the subscripts at 32 included -- so its operand
	// is the braced expression or the single variable and nothing more. At
	// 320 the operand parse stops before a subscript, whose own power is
	// 320: below that, `@$r[1,2]` read as a deref of `$r[1,2]`.
	//
	// `$$x[0]` is `${$x}[0]` -- the subscript applies to the DEREFERENCE,
	// not to `$x` -- so the sigil must take its operand before any postfix
	// gets a chance. Parsing at 300 leaves `[0]` to the caller's led loop,
	// which then wraps the whole Unary in an Index. Measured on perl 5.42.0:
	//
	//	$ perl -MO=Deparse -e 'my $r = [7]; print $$r[0];'
	//	print $r->[0];
	//
	// Deparse prints the arrow form, which is the same operation spelled the
	// other way -- and is why §4.14 gives both one node with an `Arrow` flag.
	bpDeref = 320
)

// parseWordTerm turns a bareword in term position into a call, a bareword
// term, or a declaration-like keyword the caller handles.
// keywordName is the keyword a word names: `CORE::X` is X, the builtin
// whatever else is in scope -- measured on 5.42.0, Deparse writes
// `CORE::state $x`, `CORE::say` and `CORE::length $s` as the bare keywords.
// `CORE::GLOBAL::` is a namespace for overriding builtins, not a keyword
// prefix, and is left alone.
func keywordName(word string) string {
	rest, ok := strings.CutPrefix(word, "CORE::")
	if !ok || strings.HasPrefix(rest, "GLOBAL::") {
		return word
	}
	return rest
}

// gatedUnary are the named unaries that exist only under a feature: `fc` and
// `evalbytes`. Unfeatured, perl reads `fc $a` as the METHOD call `$a->fc`;
// featured, as `fc($a)` with a following comma outside it. Measured on
// 5.42.0 with -MO=Deparse,-p:
//
//	use feature "fc"; my $z = fc $a, $b;   ((my($z) = fc($a)), $b)
//	use v5.14; my $z = fc $a;              (my($z) = $a->fc)
//
// This is issue 01a0dfbd-f8d9's option (b): the parser tracks the feature
// rather than tabling the word unconditionally. The unfeatured method-call
// reading is not modelled and stays as it was.
var gatedUnary = map[string]bool{"fc": true, "evalbytes": true}

// namedUnaryHere reports whether a word parses as a named unary at this
// point in the file: always for namedUnary, and for a gatedUnary when its
// feature is on or the word is spelled with `CORE::`, which names the builtin
// whatever is enabled.
func (p *parser) namedUnaryHere(spelled, text string) bool {
	if namedUnary[text] {
		return true
	}
	return gatedUnary[text] && (spelled != text || p.features[text])
}

func (p *parser) parseWordTerm(word lexer.Token) *Node {
	// spelled is the word as written and is what the node records; text is
	// what it names, which keywordName gives with any `CORE::` prefix off, so
	// `CORE::length $s` classifies as `length` and still canons as written.
	spelled := p.text(word)
	text := keywordName(spelled)

	if n := p.parseIndirect(word, spelled, text); n != nil {
		return n
	}

	if text == "return" {
		return p.parseReturnTerm(word)
	}

	// `require MODULE` inside an expression: a named unary whose operand is
	// a bareword module name -- `(require Foo, 2)`, `... or require mro,
	// diag "x"`, measured on 5.42.0. At a statement's start the Use path
	// reads it; here the name is its one operand and the expression goes on.
	if text == "require" {
		if name, ok := p.peekAfter(word); ok && name.Kind == lexer.Word &&
			!isPerlKeyword(keywordName(p.text(name))) {
			if after, ok := p.peekAfter(name); !ok || p.text(after) != "=>" {
				p.advanceTo(name)
				p.notePackage(p.text(name))
				return &Node{
					Kind: Call, Text: spelled, Resolved: true,
					Start: word.Start, End: name.End,
					Children: []*Node{{Kind: Term, Text: p.text(name), Start: name.Start, End: name.End}},
				}
			}
		}
	}

	// `Foo::` is a class-name string: toke.c:8072 strips the `::` and makes
	// a constant without looking for a sub, so there is no callee.
	if len(spelled) > 2 && strings.HasSuffix(spelled, "::") {
		p.advanceTo(word)
		return &Node{Kind: Term, Text: spelled, Start: word.Start, End: word.End}
	}

	// A loop control in an expression, with its label: `... and last BIN`.
	switch text {
	case "last", "next", "redo":
		return p.parseLoopControlTerm(word)
	}

	// A niladic builtin takes nothing: `time`, `wantarray`.
	if niladicParse[text] {
		p.advanceTo(word)
		return &Node{
			Kind: Call, Text: spelled, Resolved: true,
			Start: word.Start, End: word.End,
		}
	}

	p.advanceTo(word)
	n := &Node{Kind: Call, Text: spelled, Start: word.Start}

	// The paren cliff, §4.8.1. A `(` immediately after the name makes this a
	// FUNC1 -- the parens delimit the arguments and nothing beyond them
	// belongs to the call. toke.c's UNI3 returns FUNC1 here and UNIOP
	// otherwise, and the difference is visible:
	//
	//	length $x + 1      length(add($x, 1))    the + is inside
	//	length ($x) + 1    add(length($x), 1)    the + is outside
	if next, ok := p.peekSignificant(); ok && p.text(next) == "(" {
		p.advanceTo(next)
		// The BLOCK slot is inside the parens for `map({...} @a)`, which is
		// the same call as `map {...} @a` -- perl's Deparse emits the
		// parenthesised spelling for both -- so it is read here for the
		// reason the parenless path reads it below. Canon parenthesises
		// every call, so without this its own emission of a block argument
		// did not re-parse: the `;` after the block's last statement ended
		// the statement and the `}` was orphaned.
		//
		// It asks the token, as the parenless path does. The lexer's
		// listOpParen carry ran intuit_curly at the `(`, so a HASHREF first
		// argument -- `map({a => 1}, @a)` -- leaves OpensBlock unset and
		// falls through to parseCallArgs as the ordinary argument it is.
		//
		// A FILEHANDLE belongs here too, in all three spellings the parenless
		// path reads, and by the same no-comma rule. Measured on 5.42.0 with
		// -MO=Deparse: `print($fh "x")` and `print(STDERR "y")` are handles,
		// `print($a, "z")` a list. Canon writes a handle as `print($fh "x")`
		// or `print({$fh;} "x")`, so without this its own output did not
		// re-parse. It carries Handle, which infer reads -- a handle counted
		// as argument 1 makes every typed-handle print a false Str mismatch.
		if blk := p.parseListOpBlock(text); blk != nil {
			n.Children = append(n.Children, blk)
		} else if cmp := p.parseSortComparator(text); cmp != nil {
			n.Children = append(n.Children, cmp)
		} else if fh := p.parseFilehandleSlot(text); fh != nil {
			fh.Handle = true
			n.Children = append(n.Children, fh)
		}
		if arg := p.parseCallArgs(); arg != nil {
			n.Children = append(n.Children, arg)
		}
		if close, ok := p.peekSignificant(); ok && p.text(close) == ")" {
			p.advanceTo(close)
		}
		n.End = p.prevEnd()
		n.Resolved = p.namedUnaryHere(spelled, text) || listOperator[text] || niladicParse[text]
		return n
	}

	// An operator with NO argument: `shift;`, `die;`, `$x or die;`.
	//
	// Most named unaries default their argument -- `shift` takes @_ or @ARGV,
	// `die` re-raises $@ -- so the argument is optional in practice even
	// where the grammar allows one. Calling parseExpr here returns nil, and
	// the caller then saw an unconsumed `;` and declared the statement
	// Unknown.
	//
	// Measured: `shift;` and `die;` between them are the first failure in
	// dozens of T1 files, and `... or die;` appears in almost every file
	// that opens a filehandle.
	if next, ok := p.peekSignificant(); !ok || endsArgumentList(next, p.src) {
		n.End = p.prevEnd()
		// An imported sub called with no arguments is as resolved as a
		// builtin one: `done_testing;` and `maybe;` are calls whose callee
		// this parser has seen declared.
		n.Resolved = p.namedUnaryHere(spelled, text) || listOperator[text] ||
			niladicParse[text] || p.knowsShape(text)
		return n
	}

	blk, blockFollows := p.wordBlockFollows()

	switch {
	case p.namedUnaryHere(spelled, text):
		// One argument, parsed at level 19 so arithmetic binds into it and
		// comparison does not.
		if arg := p.parseExpr(bpNamedUnary); arg != nil {
			n.Children = append(n.Children, arg)
		}
		n.Resolved = true

	case listOperator[text]:
		// A BLOCK slot comes before the list, and what makes it a slot
		// rather than a first argument is the ABSENCE of a comma:
		//
		//	map { $_ } @a          a block, then the list
		//	map +{ x => 1 }, @a    a hashref, and the comma says so
		//
		// Whether the brace is a block at all was settled by the lexer,
		// which ran perl's intuit_curly over it, so this reads OpensBlock
		// rather than deciding a second time. The hashref form leaves it
		// unset and falls through to the ordinary list below, which is
		// right: it IS an ordinary first argument.
		//
		// Without this the list was orphaned. `map { $_ => 1 } @a` became a
		// call holding an AnonHash, and `@a` a statement of its own -- every
		// byte in the tree, round-tripping, and not a parse of its source.
		if blk := p.parseListOpBlock(text); blk != nil {
			n.Children = append(n.Children, blk)
		} else if cmp := p.parseSortComparator(text); cmp != nil {
			n.Children = append(n.Children, cmp)
		}
		// A filehandle slot works the same way and for the same reason:
		//
		//	print STDERR "a";      bareword handle
		//	print $fh "a";         scalar handle
		//	print {$fh} "a";       block handle
		//
		// Without this the list parser reads `STDERR`, then finds a string
		// with no operator between them and the statement falls to Unknown.
		// Measured: 5,319 of T1's 13,558 Unknown nodes started at `print`,
		// 39% of the whole gap from this one omission.
		if fh := p.parseFilehandleSlot(text); fh != nil {
			// Marked here rather than in parseFilehandleSlot's three
			// branches: one place that cannot be forgotten when a fourth
			// handle form arrives, and the fact is about the SLOT rather
			// than about the node's own shape.
			//
			// Without it the only evidence is child shape -- two children
			// means a handle, one means a comma list -- which
			// `internal/infer/infer.go:1155-1165` records as having made
			// every typed-handle print a false Str mismatch.
			fh.Handle = true
			n.Children = append(n.Children, fh)
		}
		// The whole comma list, parsed below the comma at level 7.
		if arg := p.parseExpr(bpListOp); arg != nil {
			n.Children = append(n.Children, arg)
		}
		n.Resolved = true

	case p.knowsShape(text):
		// A sub this file imported or declared, whose prototype says how a
		// call to it parses. §4.8.3's "let a later pass decide" is satisfied
		// HERE, because the shape came from a declaration rather than a guess.
		//
		// THREE ROUTES REACH THIS, and only the first needs a module:
		//
		//   - An IMPORT whose module's source was readable. Measured: in T1,
		//     889 of 986 files use Test::More and `subtest` appears 691
		//     times, always without parens.
		//   - A `sub NAME` EARLIER IN THIS FILE, recorded by parseSubDecl as
		//     it is read (decl.go, declareSub). Measured: 316 of T2's 354
		//     parenless-call refusals had a callee declared this way and 0
		//     had one reachable by import -- no T2 file uses Test::More, they
		//     `require './test.pl'`.
		//   - A `.pl` FILE this one `require`d by a literal path, read by
		//     use.go's resolveRequiredFile. That is what closes the T2
		//     sentence above: `t/test.pl` declares 78 subs, `ok`, `is` and
		//     `like` among them, and 394 of the 435 dirty files in perl.git's
		//     `t/` hold 7,211 of its 7,490 Unknown nodes behind exactly this.
		//     Measured: resolving it moves 93 files clean and 3,267 nodes.
		//
		// Which route supplied the shape is not a distinction this site can
		// or should make: `Import.Local` records whether the file declared it
		// ITSELF, and a prototype means the same thing however it was read.
		p.parseByShape(n, text)

	case blockFollows:
		// `WORD BLOCK` and then, optionally, a list: `defer { ... } print "b"`
		// and an arbitrary `zzz { 1 } print "b"`. The lexer settled that the
		// brace opens a block rather than a hash (intuitCurly, the same
		// lookahead map/grep/sort get); what is left is the statement
		// BOUNDARY, which is that the block and what follows are one
		// statement.
		//
		// Perl reads this three ways depending on the symbol table -- an
		// indirect method call on the block's value when undeclared, a call
		// with an anon-hash argument when declared, a call with a code ref
		// when prototyped `(&)`. Resolving WHICH is M4's, once prototypes
		// resolve. All three agree on the boundary, so all three are served by
		// taking the block and then the list, and none of them is the Index
		// this fell to before: `zzz(){1}`, a call subscripted by a hash, is a
		// shape perl never produces.
		//
		// Resolved stays false for the same reason the default branch's does:
		// the callee is a bareword whose declaration this parser has not seen.
		// The SHAPE is known, the callee is not.
		//
		// The list is taken only when something is actually there.
		// `zzz { 1 };` has a `;` next, and parsing a list from it read the
		// terminator INTO the argument list -- canon `zzz({1;};);` -- which
		// turned a form that already parsed into an Unknown. endsArgumentList
		// is the same guard the no-argument path above makes, for the same
		// reason.
		//
		// A `WORD BLOCK` with no `;` may be followed by a whole STATEMENT, and
		// that statement is not its argument. Measured on 5.42.0, where the two
		// come back as siblings:
		//
		//	$ perl -MO=Deparse -e 'use feature "defer"; while(1){
		//	      defer { print "d" } if ($i == 3) { last; } }'
		//	    defer {
		//	        print 'd';
		//	    }
		//	    if ($i == 3) {
		//	        last;
		//	    }
		//
		// `defer.t` is where this bites: `defer { ... }` on one line and `if
		// (...) { ... }` on the next, with no terminator between them. Taking
		// the `if` as an argument emitted `if (defer {...}) ($i == 3){last}`,
		// two paren groups and not a parse of anything. A modifier is the same
		// token in the same place -- `zzz { 1 } if $x` is a modifier, never an
		// argument -- so one check covers both, and `applyModifier` at
		// statement level already reads it.
		n.Children = append(n.Children, p.parseBlock(blk))
		if next, ok := p.peekSignificant(); ok && !endsArgumentList(next, p.src) &&
			!(next.Kind == lexer.Word && modifiers[p.text(next)]) {
			if arg := p.parseExpr(bpListOp); arg != nil {
				n.Children = append(n.Children, arg)
			}
		}
		n.Resolved = false

	default:
		// A bareword this parser does not know. It might be a user sub taking
		// a list, a class name, or a hash key -- perl decides with the symbol
		// table and this parser cannot.
		//
		// §4.8.3: produce a Call with Resolved false rather than an Unknown.
		// A call has a known SHAPE and an unknown callee, and the harness
		// scores that as an Unresolved site -- `wider`, not WRONG.
		//
		// No arguments are consumed. Guessing that a bareword takes a list
		// would swallow the rest of the statement on every hash key and
		// class name in the corpus.
		n.Resolved = false
	}

	n.End = p.prevEnd()
	return n
}

// wordBlockFollows reports the `{` of a block immediately following the word
// just read, or false when the next token is not one.
//
// Whether the brace is a block at all was settled by the lexer, which ran
// perl's intuit_curly over it, so this reads OpensBlock rather than deciding a
// second time -- the same thing parseListOpBlock does for map, grep and sort.
func (p *parser) wordBlockFollows() (lexer.Token, bool) {
	tok, ok := p.peekSignificant()
	if !ok || !tok.OpensBlock || p.text(tok) != "{" {
		return lexer.Token{}, false
	}
	return tok, true
}

// endsArgumentList reports whether a token closes the context an operator's
// arguments would live in, so there is nothing left for it to take.
//
// A terminator, a closing bracket, or an infix operator: `$x or die;` reaches
// `die` with a `;` next, and `f(shift)` reaches `shift` with a `)` next.
func endsArgumentList(tok lexer.Token, src []byte) bool {
	switch tok.Kind {
	case lexer.Semicolon, lexer.CloseBracket:
		return true
	case lexer.Operator:
		// An infix operator cannot start an argument, so the call takes
		// none: `shift || 1`. A PREFIX operator can -- `die -1` -- so only
		// the unambiguously-infix ones count.
		//
		// The concatenation, comparison and binding operators belong here for
		// exactly that reason, and perl draws the line in the same place.
		// Measured on 5.42.0, `sub f { 1 }` in scope:
		//
		//	my $v = f . 2;      f() . '2'     the `.` is infix: no argument
		//	my $v = f == 2;     f() == 2      likewise
		//	my $v = f eq 2;     f() eq 2      likewise, word-spelled
		//	my $v = f =~ 2;     f() =~ /2/    likewise
		//	my $v = f + 2;      f(2)          `+` is PREFIX-capable: argument
		//	my $v = f - 2;      f(-2)         so is `-`
		//
		// `+` and `-` are deliberately absent: perl reads them as the sign of
		// the argument, and listing them would cut a list perl does not cut.
		// `*` is absent for a third reason -- measured, `f * 2` is `f(*2)`,
		// a GLOB, so it starts a term as well.
		//
		// `<=>` is absent for that same third reason and it is the surprise
		// of the set. Measured, `my $v = f <=> 2;` does not deparse as a
		// comparison at all -- it pulls in `use File::Glob ()`, because after
		// a list operator's name perl reads `<...>` as a GLOB. So a `<`
		// starts a term here and must not cut the list.
		//
		// `->` is here for the same reason: it cannot start a term, so the
		// call before it takes none and the arrow applies to its result.
		// Measured, `sub t (;$) { 1 } my $x = t->m(1);` deparses as
		// `t()->m(1)`, and `shift->m` is every accessor's first line.
		//
		// The range operators likewise: measured, `(undef..2)` is
		// `((undef) .. 2)` and `(g ... 2)` is `(g() ... 2)`.
		switch string(src[tok.Start:tok.End]) {
		case ",", "=>", "||", "&&", "//", "=", "?", ":",
			".", "==", "!=", "=~", "!~", "->", "..", "...":
			return true
		}
	case lexer.Word:
		// The word-spelled logical operators, which the lexer emits as Word
		// because they are identifiers by shape.
		//
		// The word-spelled COMPARISONS are here for the infix reason above
		// rather than the logical one: measured, `f eq 2` is `f() eq 2`.
		switch string(src[tok.Start:tok.End]) {
		case "or", "and", "xor", "if", "unless", "while", "until", "for", "foreach",
			"eq", "ne", "cmp":
			return true
		}
	}
	return false
}

// takesFilehandle is the set of list operators whose first slot may be a
// filehandle with no comma after it.
//
// Not every list operator has one -- `push @a, 1` has no handle slot and
// treating `@a` as one would be wrong -- so the set is explicit.
//
// `system` and `exec` share the slot: it holds the program to run rather than
// a handle, in the same three spellings and by the same no-comma rule.
// Measured on 5.42.0 with -MO=Deparse, `system { "ls" } "ls", "-l"` is
// `system({'ls';} 'ls', '-l')` and `system $shell "-sh"` keeps `$shell` in
// the slot.
var takesFilehandle = map[string]bool{
	"print": true, "printf": true, "say": true,
	"system": true, "exec": true,
}

// takesBlock is the set of list operators whose first slot may be a BLOCK
// with no comma after it.
//
// The three operators where the brace is genuinely ambiguous and the slot it
// fills is the BLOCK argument. The lexer's blockTaking table is wider -- it
// also names print, printf and say, whose ambiguous brace fills a FILEHANDLE
// slot rather than this one -- which is exactly why the two tables are
// separate: the lexer's decides how to LEX the brace, this one decides which
// slot reads it. A single shared table would answer one question with the
// other's list, and put a filehandle block in the comparator's slot.
var takesBlock = map[string]bool{
	"map": true, "grep": true, "sort": true,
}

// parseListOpBlock reads the leading BLOCK of `map`, `grep` or `sort`, or
// returns nil when there is none.
//
// It asks the token, not the source: the lexer already ran intuit_curly and
// recorded the answer as OpensBlock. A parser that re-derived it could
// disagree with the token stream that produced it -- the same argument
// TestBraceDecisionUsesExpectState makes for every other brace.
func (p *parser) parseListOpBlock(op string) *Node {
	if !takesBlock[op] && !p.featureTakesBlock(op) {
		return nil
	}
	tok, ok := p.peekSignificant()
	if !ok || !tok.OpensBlock || p.text(tok) != "{" {
		return nil
	}
	return p.parseBlock(tok)
}

// featureTakesBlock reports whether op is `any` or `all` with its feature
// on: `use feature 'keyword_any'` makes `any BLOCK LIST` a builtin that
// takes a block as grep does, parenthesised or not. Measured on 5.42.0, `use
// feature qw(keyword_any); any( { $_ > 10 } 1 .. 20)` is true, and without
// the feature the same line is a syntax error.
func (p *parser) featureTakesBlock(op string) bool {
	return (op == "any" || op == "all") && p.features["keyword_"+op]
}

// parseSortComparator reads the SUBNAME or `$subref` of `sort SUBNAME LIST`,
// or returns nil when there is none.
//
// perl's rule is toke.c's KEY_sort:
//
//	s = force_word(s, BAREWORD, CHECK_KEYWORD | ALLOW_PACKAGE);
//
// A word that is NOT a perl keyword is forced to a bareword, the comparator's
// name, so `sort keys %h` sorts the keys and `sort foo @h` sorts by foo. A
// declared sub is no exception -- that is perl's documented gotcha, and the
// reason `sort uniq @x` does not do what it looks like. perly.y's
// `LSTOP indirob listexpr` admits a scalar in the same slot.
//
// Taken only when a term follows with no comma, the test parseFilehandleSlot
// makes: `sort $x, $y` sorts two scalars. Measured on 5.42.0, with and without
// parens:
//
//	sort foo 3,1,2    3,2,1      sort $s 3,1,2     1,2,3
//	sort(foo 3,1,2)   3,2,1      sort($s 3,1,2)    1,2,3
func (p *parser) parseSortComparator(op string) *Node {
	if op != "sort" {
		return nil
	}
	tok, ok := p.peekSignificant()
	if !ok {
		return nil
	}
	switch {
	case tok.Kind == lexer.Variable && p.src[tok.Start] == '$':
	case tok.Kind == lexer.Word && !isPerlKeyword(p.text(tok)):
	default:
		return nil
	}
	if next, ok := p.peekAfter(tok); !ok || !startsTerm(next, p.src) {
		return nil
	}
	p.advanceTo(tok)
	return &Node{
		Kind: Term, Text: p.text(tok), Comparator: true,
		Start: tok.Start, End: tok.End,
	}
}

// isPerlKeyword reports whether perl's keyword() would claim a word, which is
// what CHECK_KEYWORD tests: every builtin, the declarators, and the handful of
// statement-forming words that are not builtins. A `CORE::` spelling names a
// builtin by definition.
func isPerlKeyword(word string) bool {
	if namedUnary[word] || listOperator[word] || niladicParse[word] ||
		declarators[word] || strings.HasPrefix(word, "CORE::") {
		return true
	}
	switch word {
	case "sub", "do", "eval", "return":
		return true
	}
	return false
}

// parseFilehandleSlot reads `STDERR`, `$fh` or `{$fh}` before a list, or
// returns nil when there is none.
//
// The distinguishing feature is the ABSENCE of a comma: `print $fh "a"` has a
// handle, `print $x, "a"` does not. So the slot is taken only when the token
// after the candidate begins a new term rather than continuing the list.
func (p *parser) parseFilehandleSlot(op string) *Node {
	if !takesFilehandle[op] {
		return nil
	}
	tok, ok := p.peekSignificant()
	if !ok {
		return nil
	}

	switch {
	case tok.OpensBlock && p.text(tok) == "{":
		// `print {$fh} "a"` -- the block form, which exists precisely to
		// disambiguate an expression in the slot.
		return p.parseBlock(tok)

	case tok.Kind == lexer.Word && p.isBarewordHandle(p.text(tok)):
		// A bareword handle: any word that is neither a builtin nor a sub
		// this parse knows, which is perl's rule -- see isBarewordHandle.
		// A declared `foo` keeps `print foo 1` a call.
		//
		// AND ONLY WHEN NO COMMA FOLLOWS, which is perl's actual rule and
		// the same test the `$fh` branch below makes. Measured on 5.42.0:
		//
		//	$ perl -e 'print FOO, "x";'           No comma allowed after filehandle
		//	$ perl -e 'print __CLASS__, "x";'     No comma allowed after filehandle
		//	$ perl -e 'print __PACKAGE__, "x";'   syntax OK
		//
		// A comma after a handle is an ERROR, so a bareword followed by
		// one was never a handle. Without this test `print __PACKAGE__,
		// "\n"` put the token in the slot and stranded the comma, which
		// surfaced as `trailing_tokens` over the whole statement.
		//
		// The test is on the FOLLOWING TOKEN rather than on the word,
		// because the word cannot answer it. `__CLASS__` and `__SUB__`
		// are values under their features and bareword filehandles
		// without them, and this site has no feature state; the comma is
		// right in both readings. A list of exempt names was tried first
		// and was wrong on `__CLASS__` and missing `__SUB__` -- one fault
		// each way, from encoding a judgement perl does not make.
		if next, ok := p.peekAfter(tok); !ok || !startsTerm(next, p.src) {
			return nil
		}
		p.advanceTo(tok)
		return &Node{
			Kind: Term, Text: p.text(tok),
			Start: tok.Start, End: tok.End,
		}

	case tok.Kind == lexer.Variable && p.src[tok.Start] == '$':
		// `print $fh "a"`. Only when what FOLLOWS starts a new term with no
		// comma -- otherwise `print $x, "a"` would lose its first argument.
		next, ok := p.peekAfter(tok)
		if !ok || !startsTerm(next, p.src) {
			return nil
		}
		p.advanceTo(tok)
		return &Node{
			Kind: Term, Text: p.text(tok),
			Start: tok.Start, End: tok.End,
		}
	}
	return nil
}

// isBarewordHandle reports whether a bareword can fill a handle slot.
//
// perl's rule is not a naming convention. Measured on 5.42.0 with -MO=Deparse:
//
//	print foo "x";                 print foo 'x';      a handle, lower case
//	sub foo {1} print foo "x";     print foo('x');     a declared sub: a call
//	print length "x";              print 1;            a builtin
//
// So a word is a handle unless it names a builtin or a sub this parse has
// already seen declared or imported. The ALL-CAPS test that stood here took
// `STDERR` and missed `print tmp "..."` (op/while.t) and `print foo "ok 6\n"`
// (io/print.t).
//
// THE SHAPE IS ONLY HALF THE TEST. What follows the word decides the
// rest, and the caller applies it -- see parseFilehandleSlot.
func (p *parser) isBarewordHandle(word string) bool {
	if word == "" || isPerlKeyword(keywordName(word)) {
		return false
	}
	_, known := p.lookupSub(word)
	return !known
}

// startsTerm reports whether a token begins a new term rather than continuing
// an expression -- the test that separates `print $fh "a"` from `print $x, 1`.
func startsTerm(tok lexer.Token, src []byte) bool {
	switch tok.Kind {
	// A DerefSigil is the `$` of `$$code` or the `@` of `@$lines`: a sigil
	// applied to an expression, which begins a term as a variable does.
	case lexer.Variable, lexer.Number, lexer.Quote, lexer.HeredocOpen, lexer.DerefSigil:
		return true
	case lexer.Word:
		// A word that is an INFIX OPERATOR continues the expression
		// rather than starting a term. `eq`, `ne`, `cmp`, `lt`, `x`,
		// `and` and the rest are Words to the lexer and operators to
		// perl, and both filehandle branches ask this question:
		//
		//	$ perl -MO=Deparse -e 'print FOO eq "x" ? "a" : "b";'
		//	print 'b';
		//
		// perl compares and prints the result -- FOO is the comparison's
		// left operand, not a destination. Answering true here made both
		// branches take the slot and strand the rest of the expression.
		//
		// Asked of `infix` rather than of a list written here, so a word
		// operator arrives in ONE place and every caller follows.
		//
		// A statement MODIFIER does not start a term either: `print $warn if
		// length $warn` prints $warn, measured on 5.42.0. Answering yes made
		// the scalar a handle and left the modifier nothing to attach to.
		word := string(src[tok.Start:tok.End])
		_, isOperator := infix[word]
		return !isOperator && !modifiers[word]
	}
	return false
}

// parseCallArgs reads the inside of a parenthesised argument list.
func (p *parser) parseCallArgs() *Node {
	if tok, ok := p.peekSignificant(); ok && p.text(tok) == ")" {
		return nil
	}
	return p.parseExpr(0)
}
