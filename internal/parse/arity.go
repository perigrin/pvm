// ABOUTME: Call arity: a call passing a number of arguments no declaration of its sub accepts becomes an Unknown.
// ABOUTME: RFC 0001 "Refusing what perl refuses": the arity is the declarations', save for builtins whose own parse is looser.
package parse

import (
	"strings"

	"tamarou.com/pvm/internal/types"
)

// refuseArity returns n, or an Unknown spanning it when n calls a sub with
// a number of arguments its prototype refuses, or a builtin with a number
// none of CORE.pmt's candidates for it accepts. Measured on 5.42.0,
// `each()` is "Not enough arguments for each" and `localtime(1, 2)` "Too
// many arguments for localtime".
//
// A builtin the file overrides is not the builtin: measured on 5.42.0,
// `use Time::HiRes qw(sleep); sleep(1, 2)` and `BEGIN {
// *CORE::GLOBAL::localtime = sub {1} } localtime(1, 2)` compile.
//
// ponytail: a module's DEFAULT export, `use M;` with no list and M unread,
// may override a builtin unseen. The ones measured on 5.42.0 keep an arity
// within the builtin's: `use autodie; chdir(1, 2)` and `use bigint; hex(1,
// 2)` are "Too many arguments", and `unlink()` and `oct()` compile with or
// without them. Treat an unread module's default import as overriding if
// one turns up that accepts what its builtin refuses.
//
// A `.pmt` is not checked: CORE.pmt is one, and reading it builds the
// candidates this consults.
func (p *parser) refuseArity(n *Node) *Node {
	if n == nil || n.Kind != Call || p.typed || p.mayBeAutoquoted(n) || p.mayBeIndirect(n) {
		return n
	}
	count, exact := argCount(n)
	if exact && !p.arityAccepted(n.Text, count) {
		return &Node{Kind: Unknown, Refusal: CallArity, Start: n.Start, End: n.End}
	}
	return n
}

// arityAccepted reports whether a call to name may pass count arguments:
// whether its prototype takes that many, or, for a builtin, whether one of
// its candidates does. A sub with no prototype takes any number. Measured
// on 5.42.0, `sub f ($$) {} f(1)` is "Not enough arguments for main::f".
func (p *parser) arityAccepted(name string, count int) bool {
	builtin := keywordName(name)
	if !p.isPerlKeyword(builtin) {
		if imp, ok := p.lookupSub(name); ok {
			return !imp.PrototypeKnown || imp.Prototype == "" || prototypeAccepts(imp.Prototype, count)
		}
	}
	cands := coreSignatures()[builtin]
	if len(cands) == 0 || name == builtin && (p.overridden(builtin) || p.gatedOff(builtin)) ||
		types.Accepts(cands, count) {
		return true
	}
	switch looseArity[builtin] {
	case looseAny:
		return true
	case looseEmpty:
		return count == 0
	case looseExtra:
		// Too many is the loose side: some smaller count is accepted.
		for k := range count {
			if types.Accepts(cands, k) {
				return true
			}
		}
	}
	return false
}

// prototypeAccepts reports whether a prototype, parens included, takes
// count arguments: at least its slots before the `;`, `_` aside, and no
// more than it has unless one is slurpy.
func prototypeAccepts(proto string, count int) bool {
	slots, slurpy := refScalarSlots(proto)
	least := 0
	for _, s := range slots {
		if !s.optional {
			least++
		}
	}
	return count >= least && (slurpy || count <= len(slots))
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

// mayBeIndirect reports whether n's first argument is a bare class name,
// `new IO::File`, which perl may read as an indirect method call, where no
// prototype applies: measured on 5.42.0 with -MO=Deparse,-p, under
// `package P; sub new ($$;$) {}`, `new IO::File` is `'IO::File'->new`.
func (p *parser) mayBeIndirect(n *Node) bool {
	if len(n.Children) == 0 {
		return false
	}
	a := n.Children[0]
	if a.Kind != Call && a.Kind != Term || len(a.Children) > 0 || a.Text == "" ||
		a.End-a.Start != len(a.Text) || strings.IndexByte("$@%&*", a.Text[0]) >= 0 {
		return false
	}
	return strings.Contains(a.Text, "::") || p.packages[a.Text] || interpreterPackages[a.Text]
}

// overridden reports whether a call to the builtin name, spelled without
// `CORE::`, may be an override: a sub this parse has seen, a CORE::GLOBAL::
// override, or a word a `use` list names, whose module may export it
// unread. Measured on 5.42.0, `use POSIX qw(localtime); localtime(1, 2)`
// and `use subs "localtime"; localtime(1, 2)` compile, while `use POSIX
// (); localtime(1, 2)` and `use Time::HiRes qw(sleep); CORE::sleep(1, 2)`
// are "Too many arguments".
func (p *parser) overridden(name string) bool {
	_, sub := p.lookupSub(name)
	_, global := p.lookupSub("CORE::GLOBAL::" + name)
	return sub || global || p.listed[name]
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
// die` (perl.git t/op/getppid.t) parses with the `or` inside the `my`. An
// anonymous sub, `sub {1}`, is one argument.
func parenthesisedList(a *Node) bool {
	if a.Kind == Declaration && a.Text != "sub" {
		return len(a.Children) != 1 || a.Children[0].Kind != Term
	}
	return a.Kind == List
}

// looseness is which side of its candidates' arity a builtin's own parse
// accepts beyond them.
type looseness int

const (
	// looseNone: the candidates say what perl refuses.
	looseNone looseness = iota

	// looseEmpty: an empty call compiles, `system()`, `exec`, `do()`,
	// while one with too many arguments, `do($x, $x)`, does not.
	looseEmpty

	// looseExtra: a named unary takes a parenthesised list, so too many
	// arguments compile, `close($x, $x)`, while too few, `scalar()`, do not.
	looseExtra

	// looseAny: any count compiles. `not()` takes no operand to its own
	// parse, and `CORE::dump($x)` a label.
	looseAny
)

// looseArity are the builtins whose own parse accepts calls their CORE.pmt
// candidates refuse, so the candidates cannot say what perl refuses there.
// Measured on 5.42.0 by calling every CORE.pmt builtin with none to six
// arguments, plain and as `CORE::NAME`, and running `perl -c` over each
// call refused here; each name below compiled at least once, and every
// refused call on the other side of its looseness died. A word gated on a
// feature is gatedWords', not this table's.
var looseArity = map[string]looseness{
	"system": looseEmpty, "exec": looseEmpty, "do": looseEmpty,

	"eval": looseExtra, "scalar": looseExtra, "prototype": looseExtra,
	"write": looseExtra, "stat": looseExtra, "lstat": looseExtra,
	"readline": looseExtra, "tell": looseExtra, "telldir": looseExtra,
	"close": looseExtra, "closedir": looseExtra, "eof": looseExtra,
	"fileno": looseExtra, "getc": looseExtra, "getpeername": looseExtra,
	"getsockname": looseExtra, "readdir": looseExtra, "rewinddir": looseExtra,

	// evalbytes takes a parenthesised list as eval does: measured,
	// `CORE::evalbytes($x, $x)` compiles.
	"evalbytes": looseExtra,

	// method is a declarator under the class feature: `method { 1 }` is
	// an anonymous method, measured to compile (perl.git t/class/field.t),
	// and its candidates take no argument.
	"not": looseAny, "dump": looseAny, "method": looseAny,
}

// gatedWords are the builtins that exist only under a feature, each to the
// feature's name in p.features. Spelled plainly with the feature off, the
// word is a user's sub, which perl calls with any arguments; spelled
// `CORE::NAME`, or with the feature on, it is the builtin. Measured on
// 5.42.0 with `perl -c`: `fc($x, $x)` and `any()` compile, while `use
// feature "fc"`, `use v5.16` and `CORE::fc` make `fc($x, $x)` "Too many
// arguments for fc", and `use feature "keyword_any"` or `CORE::any()` make
// `any()` "Not enough arguments for any". With the feature on or as
// `CORE::NAME`, perl refuses `isa($x)`, `catch($x)`, `break($x)`,
// `__SUB__($x)` and `__CLASS__($x)` outright. method, a declarator under
// its feature, is looseArity's.
//
// try, switch and current_sub are not among the features p.features
// tracks, so catch, break and __SUB__ map to a name it never sets: their
// plain spelling is always taken for a user's sub, which refuses less than
// perl does and never more.
var gatedWords = map[string]string{
	"fc": "fc", "evalbytes": "evalbytes",
	"any": "keyword_any", "all": "keyword_all",
	"isa": "isa", "__CLASS__": "class",
	"catch": "try", "break": "switch", "__SUB__": "current_sub",
}

// gatedOff reports whether name is a builtin gated on a feature this file
// has not turned on, so that its plain spelling names a user's sub.
func (p *parser) gatedOff(name string) bool {
	feature, gated := gatedWords[name]
	return gated && !p.features[feature]
}
