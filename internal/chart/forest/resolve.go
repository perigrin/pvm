// ABOUTME: Resolve, after the forest: a scope-decided word keeps only the readings toke.c gives it under the subs and features in scope at its byte.
// ABOUTME: Declared subs, packages, classes and features come from the file itself; imported subs from the caller's Scope, parse.Imports' file-wide answer.

package forest

import (
	"slices"
	"strings"
	"unicode/utf8"

	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/chart/lex"
	"tamarou.com/pvm/internal/lexer"
	"tamarou.com/pvm/internal/parse"
)

// Scope is what resolve reads from outside the file: the subs its use
// statements import, as parse.Imports gives them over a tree from
// parse.ParseFileFrom or parse.ParseWithLoader. The caller decides how they
// were obtained; the forest never loads a module. A nil map imports nothing.
// An entry parse marks Local is the file's own sub, which resolve reads from
// the forest instead, where it is declared.
// shortcut: imports are file-wide, so a call before its use reads as
// imported (perigrin, 2026-11-10, "File-wide for now"): parse.Imports does
// not say which use brought a name in; upgrade when it carries each import's position.
type Scope struct {
	Imports map[string]parse.Import
	// LibraryOperators are the infix operators the file's use statements bring into scope, as
	// parse.LibraryOperators gives them over the same tree: each from the end of its use, where BEGIN imports it.
	// shortcut: file-wide past that end, as parse's scope is, so a block's } or a no does not end it, where perl's
	// %^H is lexical (`{ use My::Ops; } 1 plus 2;` is accepted, `TestLibraryOperatorsLimits`); upgrade when
	// parse.LibraryOperators carries where each stops applying.
	LibraryOperators []parse.LibraryOperator
}

// scopeGrammar is the perly.y symbols resolve reads, by id, built once per
// tables value; ok is false for a grammar without them, which resolve leaves alone.
type scopeGrammar struct {
	ok bool
	// the readings scope decides
	noamp, func0sub, uniopsub, lstopsub, methcall0, methcall, bareword int
	lstop, fn, prototype, privateref, attrlist                         int
	subStmt, sigStmt, pkgStmt, pkgBlock, classStmt, classBlock         int
	// decided is the readings above, term the terminals that begin a term, keyword the readings a feature keyword's
	// own classes give it (grammar.FeatureReading), and sig and plainSub a sub's with signatures on and off, by symbol id
	decided, term, keyword, sig, plainSub []bool
}

func buildScopeGrammar(tab *grammar.Tables) *scopeGrammar {
	id := func(name string) int {
		return slices.IndexFunc(tab.Symbols, func(s grammar.TableSymbol) bool { return s.Name == name })
	}
	g := &scopeGrammar{noamp: id("NOAMP"), func0sub: id("FUNC0SUB"), uniopsub: id("UNIOPSUB"), lstopsub: id("LSTOPSUB"),
		methcall0: id("METHCALL0"), methcall: id("METHCALL"), bareword: id("BAREWORD"), lstop: id("LSTOP"), fn: id("FUNC"),
		prototype: id("PROTOTYPE"), privateref: id("PRIVATEREF"), attrlist: id("ATTRLIST"), subStmt: id("bare_statement_sub_traditional"),
		sigStmt: id("bare_statement_sub_signature"), pkgStmt: id("bare_statement_package_declaration"),
		pkgBlock: id("bare_statement_package_definition"), classStmt: id("bare_statement_class_declaration"),
		classBlock: id("bare_statement_class_definition")}
	t, subs := id("term"), []int{id("KW_SUB_named"), id("KW_SUB_anon"), id("KW_SUB_named_sig"), id("KW_SUB_anon_sig")}
	g.ok = !slices.Contains(append([]int{g.noamp, g.func0sub, g.uniopsub, g.lstopsub, g.methcall0, g.methcall, g.bareword,
		g.lstop, g.fn, g.prototype, g.privateref, g.attrlist, g.subStmt, g.sigStmt, g.pkgStmt, g.pkgBlock, g.classStmt, g.classBlock, t}, subs...), -1)
	if g.ok {
		n := len(tab.Symbols)
		g.decided, g.term, g.keyword, g.sig, g.plainSub = make([]bool, n), make([]bool, n), make([]bool, n), make([]bool, n), make([]bool, n)
		for _, x := range []int{g.noamp, g.func0sub, g.uniopsub, g.lstopsub, g.methcall0, g.methcall, g.bareword} {
			g.decided[x] = true
		}
		for _, x := range tab.First[t] {
			g.term[x] = true
		}
		for x, sym := range tab.Symbols {
			g.keyword[x] = grammar.FeatureReading(sym.Name)
		}
		g.plainSub[subs[0]], g.plainSub[subs[1]], g.sig[subs[2]], g.sig[subs[3]] = true, true, true, true
	}
	return g
}

// fileOps are the list operators after which toke.c reads an undeclared
// bareword as a filehandle and leaves PL_expect at XTERM (yyl_just_a_word,
// toke.c:8122-8167): those it lexes with XREF, and those whose first
// operand is a filehandle (OA_FILEREF in regen/opcodes) that it lexes with LOP.
var fileOps = map[string]bool{"print": true, "printf": true, "say": true, "exec": true, "system": true,
	"open": true, "pipe": true, "binmode": true, "select": true, "read": true, "sysopen": true, "sysseek": true,
	"sysread": true, "syswrite": true, "seek": true, "fcntl": true, "ioctl": true, "flock": true, "send": true,
	"recv": true, "socket": true, "socketpair": true, "bind": true, "connect": true, "listen": true, "accept": true,
	"shutdown": true, "getsockopt": true, "setsockopt": true, "opendir": true}

// featureOf is the feature keywords.c reads each feature-gated keyword under (FEATURE_*_IS_ENABLED).
var featureOf = map[string]lex.Feature{"try": lex.FeatureTry, "catch": lex.FeatureTry, "finally": lex.FeatureTry,
	"defer": lex.FeatureDefer, "given": lex.FeatureSwitch, "when": lex.FeatureSwitch, "default": lex.FeatureSwitch,
	"break": lex.FeatureSwitch, "state": lex.FeatureState, "__SUB__": lex.FeatureCurrentSub, "field": lex.FeatureClass,
	"method": lex.FeatureClass, "class": lex.FeatureClass, "ADJUST": lex.FeatureClass, "__CLASS__": lex.FeatureClass,
	"any": lex.FeatureKeywordAny, "all": lex.FeatureKeywordAll, "fc": lex.FeatureFc, "evalbytes": lex.FeatureEvalbytes}

// shortcut: the indirect feature is not read, so a word before a variable stays a method where a bundle turns it off
// (`use v5.36; no feature "state"; state $x = 1;`, refused by perl); upgrade when lex.Features carries indirect.

// classOnly are the keywords toke.c refuses outside a class (croak_kw_unless_class).
// shortcut: __CLASS__ outside a method or a field's initializer, which perl also refuses, is accepted; upgrade when
// resolve reads the method and field scopes.
var classOnly = map[string]bool{"field": true, "method": true, "ADJUST": true}

// sub is a sub the file declares: in scope over tokens from to to, with its
// prototype as written, parens included, or "" for none.
type sub struct {
	from, to int
	proto    string
}

// What a name is at a token: no sub, a sub whose prototype is known, or
// something resolve cannot see (an XS sub's shape, a qualified name the
// file does not declare), whose readings all stand.
const (
	undeclared = iota
	declared
	unknown
)

// resolver is one file's tokens and what it declares.
type resolver struct {
	tab   *grammar.Tables
	g     *scopeGrammar
	s     *lex.Stream
	src   []byte
	scope Scope
	n     int
	// match is the token closing each { (-1 for any other token), encl the
	// innermost { open at each token (-1 for none).
	match, encl []int
	pkg         []string         // the package at each token
	subs        map[string][]sub // package subs, by qualified name
	lexical     map[string][]sub // my and state subs, by name
	own         map[string]bool  // every sub name the file declares, as written
	packages    map[string]int   // the packages the file declares, by the token their first declaration starts at
	classes     map[string]int   // the packages of those that a class statement declares, by the token it starts at
	dead        map[[2]int]bool  // (token, terminal) readings resolve drops
	// named is the tokens toke.c forces to a name, a method's class or a glob's, which are no scope-decided words
	named []bool
	// local is set while word decides a feature-gated keyword whose feature may be on, which leaves its neighbours alone
	local bool
}

// resolve drops from f each reading toke.c would not give a scope-decided
// word (one whose token admits NOAMP) under what is declared or imported at
// its byte, and from the token after such a word the readings toke.c's
// PL_expect rules out there, then every alternative left without a
// derivation. It returns the byte of the first dropped reading on the way to
// the root where f has no derivation left, else -1.
func resolve(tab *grammar.Tables, f *Forest, s *lex.Stream, src []byte, scope Scope) int {
	r := newResolver(tab, f, s, src, scope)
	if r == nil {
		return -1
	}
	for k, t := range s.Tokens {
		set := tab.TerminalSets[t.TerminalSet]
		r.attrArg(k, set)
		if !r.named[k] && !r.s.Raw[t.Index].AttrName && !r.feature(k, set) && slices.Contains(set, r.g.noamp) {
			r.word(k)
		}
	}
	if len(r.dead) == 0 {
		return -1
	}
	return r.prune(f)
}

// newResolver is f's resolver, its braces matched and its declarations read; nil for a grammar resolve leaves alone.
func newResolver(tab *grammar.Tables, f *Forest, s *lex.Stream, src []byte, scope Scope) *resolver {
	g := grammar.Derived(tab, buildScopeGrammar)
	if !g.ok {
		return nil
	}
	r := &resolver{tab: tab, g: g, s: s, src: src, scope: scope, n: len(s.Tokens), subs: map[string][]sub{},
		lexical: map[string][]sub{}, own: map[string]bool{}, packages: map[string]int{}, classes: map[string]int{},
		dead: map[[2]int]bool{}, named: make([]bool, len(s.Tokens)+1)}
	r.braces()
	r.declarations(f)
	return r
}

// braces matches the file's braces.
func (r *resolver) braces() {
	r.match, r.encl = make([]int, r.n), make([]int, r.n)
	var open []int
	for k := range r.n {
		r.match[k], r.encl[k] = -1, -1
		if len(open) > 0 {
			r.encl[k] = open[len(open)-1]
		}
		switch t := r.s.Raw[r.s.Tokens[k].Index]; {
		case t.Kind == lexer.Operator && r.is(k, "{"):
			open = append(open, k)
		case t.Kind == lexer.CloseBracket && r.is(k, "}") && len(open) > 0:
			r.match[open[len(open)-1]] = k
			open = open[:len(open)-1]
		}
	}
}

// blockEnd is the token closing the innermost block open at k, or n.
func (r *resolver) blockEnd(k int) int {
	if o := r.encl[k]; o >= 0 && r.match[o] >= 0 {
		return r.match[o]
	}
	return r.n
}

// declarations reads the packages and subs the file declares from its
// forest's statements: a package from its statement's end to its block's, or
// over its own block; a sub from where its declaration completes, to the
// file's end, or its block's for a my or state sub and an our sub's alias.
func (r *resolver) declarations(f *Forest) {
	type region struct {
		from, to int
		name     string
	}
	var pkgs []region
	for _, nd := range f.Nodes[1:] {
		from, to, s := int(nd.From), int(nd.To), int(nd.Sym)
		class := false
		if s == r.g.classStmt || s == r.g.classBlock {
			fs := r.s.FeaturesAt(from)
			class = (fs.On|fs.Maybe)&lex.FeatureClass != 0
		}
		if !class && s != r.g.pkgStmt && s != r.g.pkgBlock {
			continue
		}
		name := r.text(from + 1)
		if at, ok := r.packages[name]; !ok || from < at {
			r.packages[name] = from
		}
		if at, ok := r.classes[name]; class && (!ok || from < at) {
			r.classes[name] = from
		}
		if s == r.g.pkgStmt || s == r.g.classStmt {
			pkgs = append(pkgs, region{to, r.blockEnd(from), name})
		} else {
			pkgs = append(pkgs, region{from, to, name})
		}
	}
	slices.SortStableFunc(pkgs, func(x, y region) int { return x.from - y.from })
	r.pkg = make([]string, r.n)
	for k := range r.pkg {
		r.pkg[k] = "main"
	}
	for _, p := range pkgs { // a later start is the inner or the later package
		for k := p.from; k < p.to; k++ {
			r.pkg[k] = p.name
		}
	}
	for _, nd := range f.Nodes[1:] {
		if s := int(nd.Sym); (s == r.g.subStmt || s == r.g.sigStmt) && nd.Alt != 0 {
			r.declare(f, nd)
		}
	}
}

// declare records the sub a declaration statement names, if it names one:
// BEGIN and its kin, and a method, do not. Its prototype is the last
// PROTOTYPE token before its body, as written or as a :prototype attribute's.
func (r *resolver) declare(f *Forest, nd Node) {
	k := int(nd.From)
	lexical, our := r.is(k, "my") || r.is(k, "state"), r.is(k, "our")
	if lexical || our {
		k++
	}
	if !r.is(k, "sub") || !r.isWord(k+1) {
		return
	}
	name, body := r.text(k+1), int(f.Nodes[f.Alts[nd.Alt].Right].From)
	d := sub{from: int(nd.To), to: r.n}
	r.own[name] = true
	for j := k + 2; j < body; j++ {
		if slices.Contains(r.tab.TerminalSets[r.s.Tokens[j].TerminalSet], r.g.prototype) {
			d.proto = r.text(j)
		}
	}
	if lexical || our {
		l := d
		l.to = r.blockEnd(int(nd.From))
		r.lexical[name] = append(r.lexical[name], l)
	}
	if lexical {
		return
	}
	q := qualify(r.pkg[nd.From], name)
	if our { // perl's our sub is its package's, an always-main name included (&Foo::ENV)
		q = r.pkg[nd.From] + "::" + name
	}
	r.subs[q] = append(r.subs[q], d)
}

// qualify is word's sub name in package pkg, as gv_fetchpvn reads it: a bare name inMain holds is main's.
func qualify(pkg, word string) string {
	switch {
	case strings.HasPrefix(word, "::"):
		word = "main" + word
	case !strings.Contains(word, "::") && inMain(word):
		word = "main::" + word
	case !strings.Contains(word, "::"):
		word = pkg + "::" + word
	}
	for strings.HasPrefix(word, "main::") && strings.Count(word, "::") > 1 {
		word = word[len("main::"):]
	}
	return word
}

// inMain is gv.c's S_gv_is_in_main: the names that are main's in every
// package, and every name that does not open as an identifier.
func inMain(name string) bool {
	switch name {
	case "_", "INC", "ENV", "SIG", "ARGV", "STDIN", "STDOUT", "STDERR", "ARGVOUT":
		return true
	}
	return name == "" || name[0] < utf8.RuneSelf && !isAsciiLetter(name[0]) && name[0] != '_'
}

// isAsciiLetter is internal/lexer's, which inMain reads.
func isAsciiLetter(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
}

// lookup is what word names at token k, and its sub where it is declared.
func (r *resolver) lookup(word string, k int) (sub, int) {
	if d, ok := latest(r.lexical[word], k); ok && !strings.Contains(word, "::") {
		return d, declared
	}
	// shortcut: a sub the file declares is read before an import, so an import of its name, by use builtin, by a
	// module or by a BEGIN glob alias, does not shadow it, where perl takes whichever came last; upgrade when
	// parse.Imports says where each import was made.
	if d, ok := latest(r.subs[qualify(r.pkg[k], word)], k); ok {
		return d, declared
	}
	// parse keys an import into another package by its qualified name, and marks Local the subs the file and the
	// files it requires declare; the file's own are read above, where they are declared.
	im, ok := r.scope.Imports[word]
	if !ok {
		im, ok = r.scope.Imports[qualify(r.pkg[k], word)]
	}
	switch {
	case ok && im.Local && r.own[word]:
	case ok && !im.PrototypeKnown:
		return sub{}, unknown
	case ok:
		return sub{proto: im.Prototype}, declared
	case strings.Contains(word, "::"):
		// shortcut: a module's subs are seen only through the names it imports, so a qualified name neither the file
		// nor its imports declare keeps every reading; upgrade when parse exports the subs each loaded package defines.
		return sub{}, unknown
	}
	// shortcut: a sub use subs predeclares, or one a use builtin inside a block imports, is undeclared here (the
	// latter read with a file-scope module import's prototype if there is one, above), since parse.Imports, which is
	// file-wide, reports neither; upgrade when it does.
	return sub{}, undeclared
}

// latest is the last-declared of ds in scope at token k.
func latest(ds []sub, k int) (sub, bool) {
	var best sub
	found := false
	for _, d := range ds {
		if d.from <= k && k < d.to && (!found || d.from >= best.from) {
			best, found = d, true
		}
	}
	return best, found
}

// feature drops the readings the features in force at token k, whose readings are set, rule out, and reports
// whether it is a feature-gated keyword with its feature on: a keyword then, which no scope decides. Such a keyword
// keeps its own readings only with its feature on and a plain word's only with it off (keywords.c, yyl_just_a_word);
// field, method and ADJUST with it on keep none outside a class (croak_kw_unless_class); and sub keeps its signature
// readings only with the signatures feature on, its prototype's only with it off (yyl_sub's is_sigsub). Where an
// import or a BEGIN block may have turned the feature on (lex.Features.Maybe), a keyword keeps what either gives it,
// and sub both its readings.
func (r *resolver) feature(k int, set []int) bool {
	g := r.g
	has := func(of []bool) bool { return slices.ContainsFunc(set, func(x int) bool { return of[x] }) }
	sig := has(g.sig) && has(g.plainSub)
	if !sig && !has(g.keyword) {
		return false // most tokens, before their text is read
	}
	var drop func(x int) bool
	keyword, w := false, r.text(k)
	fs := r.s.FeaturesAt(k)
	switch f, ok := featureOf[w]; {
	case sig && fs.Maybe&lex.FeatureSignatures != 0:
		return false
	case sig:
		on := fs.On&lex.FeatureSignatures != 0
		drop = func(x int) bool { return on && g.plainSub[x] || !on && g.sig[x] }
	case !ok:
		return false
	case fs.Maybe&f != 0:
		// both: its keyword readings stand, and a plain word's are decided as any word's, at its own token only,
		// since what toke.c would expect after the word is no constraint on the keyword's reading
		if slices.Contains(set, g.noamp) {
			r.local = true
			r.word(k)
			r.local = false
		}
		return true
	case fs.On&f == 0:
		drop = func(x int) bool { return g.keyword[x] }
	case !slices.Contains(set, g.noamp) && (!slices.Contains(set, g.privateref) || slices.Contains(set, g.bareword)):
		// a name toke.c forces (after ->, sub, & or *, where a BAREWORD or a lone METHCALL0 reads it), which it never
		// looks up; where it does, the word admits NOAMP, or before ( a call's PRIVATEREF alone
		return false
	case classOnly[w] && !r.inClass(k):
		drop, keyword = func(int) bool { return true }, true
	default:
		drop, keyword = func(x int) bool { return g.decided[x] || x == g.privateref }, true
	}
	for _, x := range set {
		if drop(x) {
			r.dead[[2]int{k, x}] = true
		}
	}
	return keyword
}

// attrArg drops token k's reading as an attribute's argument where a gap parts it from the attribute's name: toke.c
// reads an argument only where ( follows the name directly (yyl_colon's scan_str), so `sub f :lvalue ($x)` and
// `sub f :prototype ($)` have none, which the token table, reading no gaps, cannot tell. The argument is the ATTRLIST
// after a (, or a Prototype token the lexer scans whole.
func (r *resolver) attrArg(k int, set []int) {
	arg, name, x := k-1, k-2, r.g.attrlist
	if r.s.Raw[r.s.Tokens[k].Index].Kind == lexer.Prototype {
		arg, name, x = k, k-1, r.g.prototype
	} else if !r.is(k-1, "(") {
		return
	}
	if name < 0 || !slices.Contains(set, x) || !slices.Contains(r.tab.TerminalSets[r.s.Tokens[name].TerminalSet], r.g.attrlist) {
		return
	}
	if r.s.Raw[r.s.Tokens[arg].Index].Start > r.s.Raw[r.s.Tokens[name].Index].End {
		r.dead[[2]int{k, x}] = true
	}
}

// inClass is whether token k is in a package a class statement declared before it, as HvSTASH_IS_CLASS reads PL_curstash.
func (r *resolver) inClass(k int) bool {
	at, ok := r.classes[r.pkg[k]]
	return ok && at < k
}

// What the word after a scope-decided one is to intuit_method: not a word;
// a word it takes for no class (a keyword, a declared sub, a word before =>,
// or any word after a sub whose prototype opens with *); one it would take
// for the class; or one resolve cannot tell.
const (
	notWord = iota
	noClass
	isClass
	maybeClass
)

// word decides the scope-decided word at token k as yyl_just_a_word and the
// intuit_method it calls do: it drops the readings toke.c would not give it,
// and marks the token toke.c then lexes with PL_expect at XTERM.
func (r *resolver) word(k int) {
	g := r.g
	d, st := r.lookup(r.text(k), k)
	if st == unknown {
		if r.fileSlot(k) && !r.isWord(k+1) {
			// toke.c reads a filehandle slot's word before a non-word as a bareword or, were it a sub, a call: never a method
			r.dead[[2]int{k, g.methcall0}], r.dead[[2]int{k, g.methcall}] = true, true
		}
		return
	}
	cls := notWord
	if r.isWord(k + 1) {
		next := r.text(k + 1)
		_, nst := r.lookup(next, k+1)
		switch {
		case lexer.IsKeyword(next) || nst == declared || r.is(k+2, "=>") || strings.HasPrefix(strings.TrimLeft(d.proto, "( ;"), "*"):
			cls = noClass
		case nst == unknown:
			cls = maybeClass
		default:
			cls = isClass
		}
	}
	var keep []int
	switch {
	case st == declared && cls != notWord && cls != noClass && r.declaredPackage(r.text(k+1), k):
		fallthrough // intuit_method takes a declared sub for a method where the word after it names a package
	case st == undeclared && cls == isClass:
		keep = []int{g.methcall0, g.methcall} // a method: its class is forced, and the token after that is a term's
		if !r.local {
			r.named[k+1] = true
			r.drop(k+1, []int{g.bareword})
			r.xterm(k + 2)
		}
	case st == declared:
		keep = r.called(k, d.proto)
		if cls == isClass || cls == maybeClass {
			// shortcut: a package or filehandle created outside the file (by a module it loads, or by naming
			// Foo::bar) also makes intuit_method take a declared sub before that word for a method, which resolve
			// cannot see, so both readings stand there; upgrade when parse exports the packages a file's uses create.
			keep = append(keep, g.methcall0, g.methcall)
		}
	case r.is(k+1, "{") && !r.fileSlot(k):
		keep = []int{g.methcall0} // a method whose indirect object is the block, after which a term is expected
		r.xterm(r.match[k+1] + 1)
	default:
		keep = []int{g.bareword}
		fh := r.fileSlot(k)
		if fh {
			r.xterm(k + 1) // a filehandle
		}
		if cls == maybeClass || cls == notWord && !fh {
			keep = append(keep, g.methcall0, g.methcall) // a word resolve cannot tell, or a method before a variable
		}
	}
	r.drop(k, keep)
}

// drop drops token k's scope-decided readings other than keep.
func (r *resolver) drop(k int, keep []int) {
	for _, x := range r.tab.TerminalSets[r.s.Tokens[k].TerminalSet] {
		if r.g.decided[x] && !slices.Contains(keep, x) {
			r.dead[[2]int{k, x}] = true
		}
	}
}

// declaredPackage is whether the file declares package name before token k, as gv_stashpvn would find it there.
func (r *resolver) declaredPackage(name string, k int) bool {
	at, ok := r.packages[name]
	return ok && at < k
}

// called is the reading a declared sub's name at token k takes by its
// prototype, as yyl_constant_op and yyl_subproto give it, marking the token
// after it when that reading leaves PL_expect at XTERM.
//
// shortcut: wherever a branch below leaves a term expected (no prototype, or $, ;$, _, +, *, @, %, $$ and the like),
// toke.c's XTERM reads a { after the call as an anonymous hash, which internal/lexer opens as a block unless its first
// item is followed by , or => (01a1e9a2), so the forest has no reading of sub foo; foo {1}; and it is refused; upgrade
// when the lexer reads that { as perl does (TestResolveKnownLimits).
func (r *resolver) called(k int, proto string) []int {
	g := r.g
	if proto == "" {
		r.xterm(k + 1)
		return []int{g.noamp}
	}
	p := strings.ReplaceAll(strings.TrimSuffix(strings.TrimPrefix(proto, "("), ")"), " ", "")
	if p == "" {
		return []int{g.bareword, g.func0sub} // a constant, or a sub of no arguments: either completes a term
	}
	p = strings.TrimLeft(p, ";")
	switch {
	case len(p) == 1 && strings.ContainsAny(p, "$_*+"), len(p) == 2 && p[0] == '\\',
		strings.HasPrefix(p, `\[`) && strings.IndexByte(p, ']') == len(p)-1:
		r.xterm(k + 1)
		return []int{g.uniopsub}
	case strings.HasPrefix(p, "&") && r.is(k+1, "{"):
		r.xterm(r.match[k+1] + 1)
		return []int{g.lstopsub}
	}
	r.xterm(k + 1)
	return []int{g.noamp}
}

// fileSlot is whether token k is where a list operator taking a filehandle
// reads one: right after it, or after its (.
func (r *resolver) fileSlot(k int) bool {
	p := k - 1
	if r.is(p, "(") {
		p--
	}
	if p < 0 || !fileOps[strings.TrimPrefix(r.text(p), "CORE::")] {
		return false
	}
	set := r.tab.TerminalSets[r.s.Tokens[p].TerminalSet]
	return slices.Contains(set, r.g.lstop) || slices.Contains(set, r.g.fn)
}

// xterm marks token p as lexed with PL_expect at XTERM: where it can begin a
// term, toke.c reads it as one, so its other readings are dropped.
func (r *resolver) xterm(p int) {
	if p <= 0 || p >= r.n || r.local {
		return
	}
	set := r.tab.TerminalSets[r.s.Tokens[p].TerminalSet]
	if !slices.ContainsFunc(set, func(x int) bool { return r.g.term[x] }) {
		return
	}
	if r.is(p, "x") {
		if _, st := r.lookup("x", p); st == declared {
			// shortcut: toke.c calls a declared x here, which the token table gives no call reading after a word
			// (addStatementX), so the operator stands: perl's verdict, not its grouping; upgrade when x carries the
			// scope-decided readings wherever no operator is expected.
			return
		}
	}
	for _, x := range set {
		if !r.g.term[x] {
			r.dead[[2]int{p, x}] = true
		}
	}
	r.named[p+1] = r.is(p, "*") // yyl_star's glob sigil: scan_ident takes the word after it as the glob's name
}

// pluginGrammar is the PLUGIN_* terminals by symbol id and by name, and by terminal set id whether a set holds one.
type pluginGrammar struct {
	is, sets []bool
	id       map[string]int
}

func buildPluginGrammar(tab *grammar.Tables) *pluginGrammar {
	g := &pluginGrammar{is: make([]bool, len(tab.Symbols)), sets: make([]bool, len(tab.TerminalSets)), id: map[string]int{}}
	for x, sym := range tab.Symbols {
		if sym.Terminal && strings.HasPrefix(sym.Name, "PLUGIN_") {
			g.is[x], g.id[sym.Name] = true, x
		}
	}
	for i, set := range tab.TerminalSets {
		g.sets[i] = slices.ContainsFunc(set, func(x int) bool { return g.is[x] })
	}
	return g
}

// libraryOperators drops from f each PLUGIN_* reading but the one toke.c's tokentype_for_plugop gives the library
// operator of the token's spelling whose use ends before it (the last such), then every alternative left without a
// derivation. It returns what prune does.
func libraryOperators(tab *grammar.Tables, f *Forest, s *lex.Stream, src []byte, scope Scope) int {
	g := grammar.Derived(tab, buildPluginGrammar)
	r := &resolver{tab: tab, s: s, src: src, n: len(s.Tokens), dead: map[[2]int]bool{}}
	var from []int
	for k, t := range s.Tokens {
		if !g.sets[t.TerminalSet] {
			continue
		}
		if from == nil {
			for _, op := range scope.LibraryOperators {
				from = append(from, r.statementEnd(op.From))
			}
		}
		keep := -1
		for i, op := range scope.LibraryOperators {
			if from[i] <= s.Raw[t.Index].Start && op.Name == r.text(k) {
				keep = g.id[grammar.LibraryTerminal(op)]
			}
		}
		for _, x := range tab.TerminalSets[t.TerminalSet] {
			if g.is[x] && x != keep {
				r.dead[[2]int{k, x}] = true
			}
		}
	}
	if len(r.dead) == 0 {
		return -1
	}
	return r.prune(f)
}

// statementEnd is the byte after the statement starting at byte at: its ; outside brackets, the } closing a BEGIN
// block, or the } closing the block around it.
func (r *resolver) statementEnd(at int) int {
	k, _ := slices.BinarySearchFunc(r.s.Tokens, at, func(t lex.Token, at int) int { return r.s.Raw[t.Index].Start - at })
	begin, depth := r.is(k, "BEGIN"), 0
	for ; k < r.n; k++ {
		switch t := r.s.Raw[r.s.Tokens[k].Index]; {
		case t.Kind == lexer.Semicolon && depth == 0:
			return t.End
		case t.Kind == lexer.CloseBracket:
			if depth--; depth < 0 || depth == 0 && begin {
				return t.End
			}
		case r.is(k, "(") || r.is(k, "[") || r.is(k, "{"):
			depth++
		}
	}
	return len(r.src)
}

// prune unlinks every alternative a dropped reading leaves without a
// derivation. It returns -1 where the root keeps one, else the byte of the
// first dropped reading reached from the root through nodes left without one.
func (r *resolver) prune(f *Forest) int {
	const (
		onPath = 1 + iota
		live
		gone
	)
	state := make([]uint8, len(f.Nodes))
	var keep func(n int32) bool
	keep = func(n int32) bool {
		switch state[n] {
		case live:
			return true
		case gone, onPath: // a node met again on its own path has no derivation there
			return false
		}
		nd := f.Nodes[n]
		if nd.Alt == 0 {
			if r.dead[[2]int{int(nd.From), int(nd.Sym)}] && nd.To == nd.From+1 {
				state[n] = gone
				return false
			}
			state[n] = live
			return true
		}
		state[n] = onPath
		var first, last int32
		for x := nd.Alt; x != 0; {
			a, next := f.Alts[x], f.Alts[x].Next
			if (a.Left == 0 || keep(a.Left)) && (a.Right == 0 || keep(a.Right)) {
				if first == 0 {
					first = x
				} else {
					f.Alts[last].Next = x
				}
				last = x
			}
			x = next
		}
		if first == 0 {
			state[n] = gone
			return false
		}
		f.Alts[last].Next = 0
		f.Nodes[n].Alt, state[n] = first, live
		return true
	}
	if keep(f.Root) {
		return -1
	}
	at, seen := r.n, make([]bool, len(f.Nodes))
	var walk func(n int32)
	walk = func(n int32) {
		if n == 0 || seen[n] || keep(n) {
			return
		}
		seen[n] = true
		if nd := f.Nodes[n]; nd.Alt == 0 {
			at = min(at, int(nd.From))
			return
		}
		for x := f.Nodes[n].Alt; x != 0; x = f.Alts[x].Next {
			walk(f.Alts[x].Left)
			walk(f.Alts[x].Right)
		}
	}
	walk(f.Root)
	if at == r.n {
		return len(r.src)
	}
	return r.s.Raw[r.s.Tokens[at].Index].Start
}

// text is token k's text.
func (r *resolver) text(k int) string {
	t := r.s.Raw[r.s.Tokens[k].Index]
	return string(r.src[t.Start:t.End])
}

// is is whether token k exists and is s.
func (r *resolver) is(k int, s string) bool {
	if k < 0 || k >= r.n {
		return false
	}
	t := r.s.Raw[r.s.Tokens[k].Index]
	return string(r.src[t.Start:t.End]) == s
}

// isWord is whether token k exists and is a word.
func (r *resolver) isWord(k int) bool {
	return k >= 0 && k < r.n && r.s.Raw[r.s.Tokens[k].Index].Kind == lexer.Word
}
