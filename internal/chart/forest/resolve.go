// ABOUTME: Resolve, after the forest: a scope-decided word keeps only the readings toke.c gives it under the subs in scope at its byte.
// ABOUTME: Declared subs and packages come from the file's own forest; imported subs from the caller's Scope, parse.Imports' file-wide answer.

package forest

import (
	"slices"
	"strings"

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
}

// scopeGrammar is the perly.y symbols resolve reads, by id, built once per
// tables value; ok is false for a grammar without them, which resolve leaves alone.
type scopeGrammar struct {
	ok bool
	// the readings scope decides
	noamp, func0sub, uniopsub, lstopsub, methcall0, methcall, bareword int
	lstop, fn, prototype                                               int
	subStmt, sigStmt, pkgStmt, pkgBlock                                int
	// decided is the readings above, and term the terminals that begin a term, by symbol id
	decided, term []bool
}

func buildScopeGrammar(tab *grammar.Tables) *scopeGrammar {
	id := func(name string) int {
		return slices.IndexFunc(tab.Symbols, func(s grammar.TableSymbol) bool { return s.Name == name })
	}
	g := &scopeGrammar{noamp: id("NOAMP"), func0sub: id("FUNC0SUB"), uniopsub: id("UNIOPSUB"), lstopsub: id("LSTOPSUB"),
		methcall0: id("METHCALL0"), methcall: id("METHCALL"), bareword: id("BAREWORD"), lstop: id("LSTOP"), fn: id("FUNC"),
		prototype: id("PROTOTYPE"), subStmt: id("bare_statement_sub_traditional"), sigStmt: id("bare_statement_sub_signature"),
		pkgStmt: id("bare_statement_package_declaration"), pkgBlock: id("bare_statement_package_definition")}
	t := id("term")
	g.ok = !slices.Contains([]int{g.noamp, g.func0sub, g.uniopsub, g.lstopsub, g.methcall0, g.methcall, g.bareword, g.lstop,
		g.fn, g.prototype, g.subStmt, g.sigStmt, g.pkgStmt, g.pkgBlock, t}, -1)
	if g.ok {
		g.decided, g.term = make([]bool, len(tab.Symbols)), make([]bool, len(tab.Symbols))
		for _, x := range []int{g.noamp, g.func0sub, g.uniopsub, g.lstopsub, g.methcall0, g.methcall, g.bareword} {
			g.decided[x] = true
		}
		for _, x := range tab.First[t] {
			g.term[x] = true
		}
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
	dead        map[[2]int]bool  // (token, terminal) readings resolve drops
	// named is the tokens toke.c forces to a name, a method's class or a glob's, which are no scope-decided words
	named []bool
}

// resolve drops from f each reading toke.c would not give a scope-decided
// word (one whose token admits NOAMP) under what is declared or imported at
// its byte, and from the token after such a word the readings toke.c's
// PL_expect rules out there, then every alternative left without a
// derivation. It returns the byte of the first dropped reading on the way to
// the root where f has no derivation left, else -1.
func resolve(tab *grammar.Tables, f *Forest, s *lex.Stream, src []byte, scope Scope) int {
	g := grammar.Derived(tab, buildScopeGrammar)
	if !g.ok {
		return -1
	}
	r := &resolver{tab: tab, g: g, s: s, src: src, scope: scope, n: len(s.Tokens), subs: map[string][]sub{},
		lexical: map[string][]sub{}, own: map[string]bool{}, packages: map[string]int{}, dead: map[[2]int]bool{}, named: make([]bool, len(s.Tokens)+1)}
	r.braces()
	r.declarations(f)
	for k, t := range s.Tokens {
		if !r.named[k] && slices.Contains(tab.TerminalSets[t.TerminalSet], g.noamp) {
			r.word(k)
		}
	}
	if len(r.dead) == 0 {
		return -1
	}
	return r.prune(f)
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
// file's end, or its block's for a my or state sub.
func (r *resolver) declarations(f *Forest) {
	type region struct {
		from, to int
		name     string
	}
	var pkgs []region
	for _, nd := range f.Nodes[1:] {
		from, to := int(nd.From), int(nd.To)
		if s := int(nd.Sym); s == r.g.pkgStmt || s == r.g.pkgBlock {
			if at, ok := r.packages[r.text(from+1)]; !ok || from < at {
				r.packages[r.text(from+1)] = from
			}
		}
		switch int(nd.Sym) {
		case r.g.pkgStmt:
			pkgs = append(pkgs, region{to, r.blockEnd(from), r.text(from + 1)})
		case r.g.pkgBlock:
			pkgs = append(pkgs, region{from, to, r.text(from + 1)})
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
	lexical := r.is(k, "my") || r.is(k, "state")
	if lexical || r.is(k, "our") {
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
	if lexical {
		d.to = r.blockEnd(int(nd.From))
		r.lexical[name] = append(r.lexical[name], d)
		return
	}
	q := qualify(r.pkg[nd.From], name)
	r.subs[q] = append(r.subs[q], d)
}

// qualify is word's sub name in package pkg, as gv_fetchpvn reads it.
func qualify(pkg, word string) string {
	switch {
	case strings.HasPrefix(word, "::"):
		word = "main" + word
	case !strings.Contains(word, "::"):
		word = pkg + "::" + word
	}
	for strings.HasPrefix(word, "main::") && strings.Count(word, "::") > 1 {
		word = word[len("main::"):]
	}
	return word
}

// lookup is what word names at token k, and its sub where it is declared.
func (r *resolver) lookup(word string, k int) (sub, int) {
	if d, ok := latest(r.lexical[word], k); ok && !strings.Contains(word, "::") {
		return d, declared
	}
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
	// shortcut: a sub use subs predeclares, or one use builtin imports, is undeclared here, since parse.Imports
	// reports neither; upgrade when it does.
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
		r.named[k+1] = true
		r.drop(k+1, []int{g.bareword})
		r.xterm(k + 2)
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
	if p <= 0 || p >= r.n {
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
