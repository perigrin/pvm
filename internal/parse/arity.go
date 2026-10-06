// ABOUTME: Call arity: a call passing a number of arguments no declaration of its sub accepts becomes an Unknown.
// ABOUTME: RFC 0001 "Refusing what perl refuses": the arity is the declarations', save for builtins whose own parse is looser.
package parse

import "tamarou.com/pvm/internal/types"

// refuseArity returns n, or an Unknown spanning it when n calls a builtin
// with a number of arguments none of CORE.pmt's candidates for it accepts.
// Measured on 5.42.0, `each()` is "Not enough arguments for each" and
// `localtime(1, 2)` "Too many arguments for localtime".
//
// A builtin the file overrides is not the builtin: measured on 5.42.0,
// `use Time::HiRes qw(sleep); sleep(1, 2)` and `BEGIN {
// *CORE::GLOBAL::localtime = sub {1} } localtime(1, 2)` compile.
//
// ponytail: an override installed by a module this parse does not read, or
// by `use subs`, is not seen; record those as imports if a corpus file
// calls an overridden builtin with an arity its own declaration refuses.
//
// A `.pmt` is not checked: CORE.pmt is one, and reading it builds the
// candidates this consults.
func (p *parser) refuseArity(n *Node) *Node {
	if n == nil || n.Kind != Call || p.typed {
		return n
	}
	name := keywordName(n.Text)
	cands := coreSignatures()[name]
	if len(cands) == 0 || looseArity[name] || p.overridden(name) || p.mayBeAutoquoted(n) {
		return n
	}
	count, exact := argCount(n)
	if exact && !types.Accepts(cands, count) {
		return &Node{Kind: Unknown, Refusal: CallArity, Start: n.Start, End: n.End}
	}
	return n
}

// mayBeAutoquoted reports whether n is a bare word that a following `=>`
// or a closing hash subscript may make a string: `values => 1`, `$h{sort}`.
// The autoquote is decided after the word is read (see autoquote), so it is
// read here from the token that follows. A bare call at the end of a block,
// `do { each }`, is left alone with it.
func (p *parser) mayBeAutoquoted(n *Node) bool {
	if len(n.Children) > 0 || n.End-n.Start != len(n.Text) {
		return false
	}
	next, ok := p.peekSignificant()
	return ok && (p.text(next) == "=>" || p.text(next) == "}")
}

// overridden reports whether this parse has seen a sub named name, or a
// CORE::GLOBAL:: override of it, that a call to the builtin may be.
func (p *parser) overridden(name string) bool {
	_, sub := p.lookupSub(name)
	_, global := p.lookupSub("CORE::GLOBAL::" + name)
	return sub || global
}

// argCount is how many arguments a call writes: each child, a block or a
// filehandle among them, split at its commas. It is not exact when one is
// a parenthesised list, which a builtin may count as one argument or as
// its elements: measured on 5.42.0, `link((1, 2))` and `pipe(my ($r, $w))`
// compile, while `localtime((1, 2))` is "Too many arguments".
func argCount(n *Node) (count int, exact bool) {
	exact = true
	for _, c := range n.Children {
		for _, a := range commaItems(c) {
			count++
			exact = exact && !parenthesisedList(a)
		}
	}
	return count, exact
}

// parenthesisedList reports whether a is a parenthesised list, `(1, 2)`,
// or a declaration of anything but one variable, `my ($r, $w)`. A
// declaration's child is not always the list alone: `pipe my ($r, $w) or
// die` (perl.git t/op/getppid.t) parses with the `or` inside the `my`.
func parenthesisedList(a *Node) bool {
	if a.Kind == Declaration {
		return len(a.Children) != 1 || a.Children[0].Kind != Term
	}
	return a.Kind == List
}

// looseArity are the builtins whose own parse accepts calls their CORE.pmt
// candidates refuse, so the candidates cannot say what perl refuses.
// Measured on 5.42.0 by calling every CORE.pmt builtin with none to six
// arguments, plain and as `CORE::NAME`, and running `perl -c` over each
// call refused here; each name below compiled at least once:
//
//	not()  CORE::not()          the operand is optional to the parse
//	system()  exec  do()        an empty list
//	CORE::dump($x)              its operand is a label
//	close($x, $x)  eval($x, $x) a named unary takes a parenthesised list
//	fc($x, $x)  any()           without its feature, a user's sub
//
// ponytail: a whole name is exempt, so `scalar()`, which perl refuses, is
// not refused either; split the exemption by direction (none or too many)
// if a corpus file needs the other half.
var looseArity = map[string]bool{
	"not": true, "system": true, "exec": true, "do": true, "dump": true,
	"eval": true, "scalar": true, "prototype": true, "write": true,
	"stat": true, "lstat": true, "readline": true, "tell": true,
	"telldir": true, "close": true, "closedir": true, "eof": true,
	"fileno": true, "getc": true, "getpeername": true, "getsockname": true,
	"readdir": true, "rewinddir": true,
	"fc": true, "evalbytes": true, "any": true, "all": true, "catch": true,
	"method": true, "isa": true, "break": true, "__CLASS__": true, "__SUB__": true,
}
