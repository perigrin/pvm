// ABOUTME: The chart as a parseoracle subject: SubjectFacts read from the forest forest.Parse leaves, in-process.
// ABOUTME: A site every reading holds is decided; one only some readings hold is Unresolved, never a guess.

package forest

import (
	"slices"
	"strings"
	"unicode/utf8"

	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/chart/lex"
	"tamarou.com/pvm/internal/chart/recognize"
	"tamarou.com/pvm/internal/lexer"
	"tamarou.com/pvm/internal/parse"
	"tamarou.com/pvm/internal/parseoracle"
)

// The site kinds, as indexes: a call, then parseoracle's five marker kinds.
const (
	kindCall = iota
	kindReference
	kindHash
	kindMatch
	kindReadline
	kindAnonhash
	nkinds
)

var kindName = [nkinds]string{"", parseoracle.SiteKindReference, parseoracle.SiteKindHash,
	parseoracle.SiteKindMatch, parseoracle.SiteKindReadline, parseoracle.SiteKindAnonhash}

// factsGrammar is what SubjectFacts reads from the tables, built once per tables value.
type factsGrammar struct {
	// kind is each rule's site kind, -1 for none; a hash site is at the node's last token (a subscript's }, a cast),
	// so nested subscripts sharing a start stay apart. hash is which hash read a kindHash rule is (hashSite).
	kind     []int
	hash     []uint8
	level    []levelShape // the element levels of a subscript chain, {...} and [...]
	deref    []uint8      // where a dereference's operand stands, for derefSites
	sigil    []byte       // a dereference's sigil: @, %, $, # for $#, * for a glob, A for ->@{...}, ? for ->%{...}
	lhs      []string     // each rule's left-hand side and right-hand side, for context
	rhs      [][]string
	matchop  int    // the rule term MATCHOP term, whose site is at the operator
	fullstmt int    // fullstmt's symbol id
	pmfunc   int    // PMFUNC's
	notCore  []bool // the terminals a word takes as a sub or method rather than a builtin
	// refs is each rule's operands op.c's ref_array_or_hash takes a reference to where they are an array or a hash,
	// each as how many symbols before the rule's last it stands: ~~'s two, and given's and when's condition.
	refs       [][]int
	smartmatch int    // the rule term NCEQOP term, whose operator may be <=> or cmp rather than ~~
	paren      int    // the rule ( expr ), which leaves an array or hash inside it one
	aggregate  []bool // the rules whose term is an array or a hash: @a, %h, @{...}, %$r, ->@* and ->%*
	glob       []bool // the rules whose term is a glob: *FOO, *{...} and ->**
	// calls is each call rule's shape: where its arguments stand, as refs counts; one where they are a single term,
	// else a list; block where a block before them fills the prototype's first slot.
	calls []callShape
	// the rules a call's argument list is read through, down to its terms
	listComma, listTrail, listOne, exprList, optList, optEmpty int
}

type callShape struct {
	ok, one, block bool
	back           int
}

// The hash reads, as perl builds them (hashSite).
const (
	hashNamed = 1 + iota // %h
	hashRef              // %$r or %{...}, ->%* and ->%{...}
	hashSlice            // @h{...}, @$r{...} and ->@{...}
	hashLevel            // a {...} element level
)

// The operands derefSites reads: a cast's indirob, the term before a postfix ->, and local's or \'s.
const (
	derefCast = 1 + iota
	derefPost
	derefLocal
)

// levelShape is an element level's rule: its base as a child count (child), whether it subscripts a hash, and
// whether it reads through a reference, by an arrow written or implied between two subscripts.
type levelShape struct {
	ok, brace, arrow bool
	base             int
}

func buildFactsGrammar(tab *grammar.Tables) *factsGrammar {
	id := func(name string) int {
		return slices.IndexFunc(tab.Symbols, func(s grammar.TableSymbol) bool { return s.Name == name })
	}
	g := &factsGrammar{kind: make([]int, len(tab.Rules)), hash: make([]uint8, len(tab.Rules)),
		level: make([]levelShape, len(tab.Rules)), deref: make([]uint8, len(tab.Rules)), sigil: make([]byte, len(tab.Rules)),
		lhs: make([]string, len(tab.Rules)), rhs: make([][]string, len(tab.Rules)), matchop: -1,
		fullstmt: id("fullstmt"), pmfunc: id("PMFUNC"), notCore: make([]bool, len(tab.Symbols)),
		refs: make([][]int, len(tab.Rules)), smartmatch: -1, paren: -1, aggregate: make([]bool, len(tab.Rules)),
		glob: make([]bool, len(tab.Rules)), calls: make([]callShape, len(tab.Rules)),
		listComma: -1, listTrail: -1, listOne: -1, exprList: -1, optList: -1, optEmpty: -1}
	for _, w := range []string{"NOAMP", "FUNC0SUB", "UNIOPSUB", "LSTOPSUB", "METHCALL0", "METHCALL", "BAREWORD"} {
		if x := id(w); x >= 0 {
			g.notCore[x] = true
		}
	}
	name := func(x int) string { return tab.Symbols[x].Name }
	for r, rule := range tab.Rules {
		g.kind[r] = -1
		rhs := make([]string, len(rule.RHS))
		for i, x := range rule.RHS {
			rhs[i] = name(x)
		}
		lhs, joined := name(rule.LHS), strings.Join(rhs, " ")
		g.lhs[r], g.rhs[r] = lhs, rhs
		switch {
		case lhs == "termeqop" && joined == "term NCEQOP term":
			g.refs[r], g.smartmatch = []int{2, 0}, r
		case joined == "KW_GIVEN PERLY_PAREN_OPEN remember mexpr PERLY_PAREN_CLOSE mblock",
			joined == "KW_WHEN PERLY_PAREN_OPEN remember mexpr PERLY_PAREN_CLOSE mblock":
			g.refs[r] = []int{2}
		case lhs == "sideff" && joined == "expr KW_WHEN condition":
			g.refs[r] = []int{0}
		case lhs == "term" && joined == "PERLY_PAREN_OPEN expr PERLY_PAREN_CLOSE":
			g.paren = r
		case lhs == "term" && slices.Contains([]string{"ary", "hsh", "term ARROW CAST_SNAIL", "term ARROW CAST_PERCENT_SIGN"}, joined):
			g.aggregate[r] = true
		case lhs == "term" && (joined == "star" || joined == "term ARROW CAST_STAR"):
			g.glob[r] = true
		case lhs == "term" && joined == "NOAMP optlistexpr":
			g.calls[r] = callShape{ok: true}
		case lhs == "term" && joined == "UNIOPSUB term":
			g.calls[r] = callShape{ok: true, one: true}
		case lhs == "term" && joined == "amper PERLY_PAREN_OPEN expr PERLY_PAREN_CLOSE":
			g.calls[r] = callShape{ok: true, back: 1}
		case lhs == "listop" && joined == "LSTOPSUB startanonsub block optlistexpr":
			g.calls[r] = callShape{ok: true, block: true}
		case lhs == "listexpr" && joined == "listexpr PERLY_COMMA term":
			g.listComma = r
		case lhs == "listexpr" && joined == "listexpr PERLY_COMMA":
			g.listTrail = r
		case lhs == "listexpr" && joined == "term":
			g.listOne = r
		case lhs == "expr" && joined == "listexpr":
			g.exprList = r
		case lhs == "optlistexpr" && joined == "listexpr":
			g.optList = r
		case lhs == "optlistexpr" && joined == "empty":
			g.optEmpty = r
		}
		switch {
		case len(rhs) > 0 && slices.Contains([]string{"NOAMP", "UNIOPSUB", "FUNC0SUB", "LSTOPSUB"}, rhs[0]),
			lhs == "term" && len(rhs) > 0 && rhs[0] == "amper": // f(...), which toke.c lexes as &f(...); altSite drops a written &
			g.kind[r] = kindCall
		case lhs == "term" && joined == "REFGEN term":
			g.kind[r] = kindReference
		case lhs == "anonymous" && len(rhs) > 0 && rhs[0] == "HASHBRACK":
			g.kind[r] = kindAnonhash
		case lhs == "termbinop" && joined == "term MATCHOP term":
			g.kind[r], g.matchop = kindMatch, r
		case lhs == "hsh" && joined == "PERLY_PERCENT_SIGN":
			g.kind[r], g.hash[r] = kindHash, hashNamed
		case lhs == "hsh", joined == "term ARROW CAST_PERCENT_SIGN", lhs == "kvslice" && joined == "term ARROW PERLY_PERCENT_SIGN":
			g.kind[r], g.hash[r] = kindHash, hashRef
		case lhs == "term" && joined == "sliceme PERLY_BRACE_OPEN expr PERLY_BRACE_CLOSE":
			g.kind[r], g.hash[r] = kindHash, hashSlice
		case lhs == "subscripted" && len(rhs) >= 4 && slices.Contains([]string{"scalar", "ARROW", "subscripted"}, rhs[len(rhs)-4]) &&
			(rhs[len(rhs)-3] == "PERLY_BRACE_OPEN" || rhs[len(rhs)-3] == "PERLY_BRACKET_OPEN"):
			sh := levelShape{ok: true, brace: rhs[len(rhs)-3] == "PERLY_BRACE_OPEN", arrow: rhs[0] != "scalar", base: 3}
			if rhs[1] == "ARROW" {
				sh.base = 4
			}
			g.level[r] = sh
			if sh.brace {
				g.kind[r], g.hash[r] = kindHash, hashLevel
			}
		}
		switch {
		case len(rhs) == 2 && strings.HasPrefix(rhs[0], "CAST_") && rhs[1] == "indirob":
			g.deref[r] = derefCast // %{...}, @{...}, ${...} and $#{...}; *{...} and &{...} fold as a read does
			g.sigil[r] = map[string]byte{"hsh": '%', "ary": '@', "scalar": '$', "arylen": '#'}[lhs]
		case len(rhs) == 3 && rhs[0] == "term" && rhs[1] == "ARROW" &&
			slices.Contains([]string{"CAST_PERCENT_SIGN", "CAST_SNAIL", "CAST_STAR", "DOLSHARP", "PERLY_SNAIL", "PERLY_PERCENT_SIGN"}, rhs[2]):
			g.deref[r] = derefPost // ->%*, ->@*, ->$* (and ->**, one rule with it), ->$#*, ->@{...} and ->%{...}
			g.sigil[r] = map[string]byte{"CAST_PERCENT_SIGN": '%', "CAST_SNAIL": '@', "CAST_STAR": '$', "DOLSHARP": '#',
				"PERLY_SNAIL": 'A', "PERLY_PERCENT_SIGN": '?'}[rhs[2]]
		case lhs == "term" && (joined == "KW_LOCAL term" || joined == "REFGEN term"):
			g.deref[r] = derefLocal // local, and \ (whose refassign keeps an rv2hv: op/lvref.t's \$h{a} = ...)
		}
	}
	return g
}

// site is one site's kind and the token it stands at, as one key.
func site(kind, at int) int { return at*nkinds + kind }

// subjectFacts reads one forest's facts.
type subjectFacts struct {
	f   *Forest
	g   *factsGrammar
	s   *lex.Stream
	src []byte
	// must and may are each node's sites as every derivation of it holds them and as any does, once computed;
	// busy marks a node on the current path, failed one with no derivation off it.
	must, may map[int32][]int
	busy      map[int32]bool
	failed    map[int32]bool
	proto     func(at int) (string, bool) // the prototype of the declared sub the word at token at calls, if it has one
	uncertain map[int]bool                // the sites some readings hold and others do not
	all       map[int]bool                // every site any reading holds
	stmts     [][2]int                    // the statements' token spans, top level only, sorted
	hashes    *hashScopes                 // what each hash name is at a token
	levels    []int32                     // the element level nodes, by their first token
	up        [][]upRef                   // where each node reachable from the root stands
}

// SubjectFacts is the chart's answer to parseoracle's subject contract for src, a file at path: whether
// forest.Parse accepts it under scope, the prototypes resolve saw, and a site for each decision at an oracle
// marker in the readings forest.Parse leaves, on its statement's lines as perl numbers them. A site some
// readings hold and others do not is Unresolved, in a statement with more than CollapseAbove readings as in any
// other: the packed forest holds every reading, so a collapsed one hedges only what its readings disagree on
// (perigrin, 2026-11-16). Neither list is ever nil.
func SubjectFacts(tab *grammar.Tables, src []byte, path string, scope Scope) parseoracle.SubjectFacts {
	facts := parseoracle.SubjectFacts{Prototypes: map[string]string{}, CallSites: []parseoracle.SubjectCallSite{}}
	f, s, _ := parseStream(tab, recognize.New(tab, recognize.Options{}), src, scope)
	if f == nil {
		return facts
	}
	facts.OK = true
	r := newResolver(tab, f, s, src, scope)
	if r == nil {
		return facts
	}
	for name, im := range scope.Imports {
		if im.PrototypeKnown && im.Prototype != "" && !r.own[name] {
			facts.Prototypes[name] = bareProto(im.Prototype)
		}
	}
	for name, ds := range r.subs {
		if d := slices.MaxFunc(ds, func(x, y sub) int { return x.from - y.from }); d.proto != "" {
			facts.Prototypes[strings.TrimPrefix(name, "main::")] = bareProto(d.proto)
		}
	}

	sf := &subjectFacts{f: f, g: grammar.Derived(tab, buildFactsGrammar), s: s, src: src,
		must: map[int32][]int{}, may: map[int32][]int{}, busy: map[int32]bool{}, failed: map[int32]bool{},
		uncertain: map[int]bool{}, all: map[int]bool{}, hashes: r.hashScopes()}
	for i, nd := range f.Nodes {
		if nd.Alt != 0 && sf.g.level[f.Alts[nd.Alt].Rule].ok {
			sf.levels = append(sf.levels, int32(i))
		}
	}
	slices.SortFunc(sf.levels, func(x, y int32) int { return int(f.Nodes[x].From - f.Nodes[y].From) })
	sf.index()
	sf.proto = func(at int) (string, bool) {
		d, st := r.lookup(r.text(at), at)
		return d.proto, st == declared && d.proto != ""
	}
	sf.walk()

	keys := make([]int, 0, len(sf.all))
	for k := range sf.all {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	line, own := parse.PerlLines(src, path)
	starts := lineStarts(src)
	// lines is the first and last lines of the statement holding token at, as perl numbers them, the last at least
	// the source line end; ok false outside a statement or where a #line directive moved it into another file.
	lines := func(at, end int) (int, int, bool) {
		st, ok := sf.statement(at)
		if !ok {
			return 0, 0, false
		}
		first := lineAt(starts, s.Raw[s.Tokens[st[0]].Index].Start)
		last := max(end, lineAt(starts, s.Raw[s.Tokens[st[1]-1].Index].End-1))
		return line[first], line[last], own[first] && own[last]
	}
	for _, k := range keys {
		kind, at := k%nkinds, k/nkinds
		first, last, ok := lines(at, 0)
		if !ok {
			continue
		}
		cs := parseoracle.SubjectCallSite{Kind: kindName[kind], Line: first, EndLine: last,
			Unresolved: sf.uncertain[k]}
		if kind == kindCall {
			cs.Name = r.text(at)
			if _, st := r.lookup(cs.Name, at); st != declared {
				// a word nothing declares may still be a sub a module resolve cannot see gave a \ prototype; a
				// declared sub's prototype is fitted slot by slot (protoRefs)
				cs.Unresolved = true
			}
			// shortcut: lookup reads the file's own sub before an import of its name, so a later use importing one
			// with a \ prototype is decided on the file's declaration, where perl applies the import's and grades it
			// WRONG; parse.Imports keeps one entry per name with no position (resolve.go's lookup); upgrade when it
			// says where each import was made.
		}
		cs.TookReference = kind == kindReference && !cs.Unresolved
		facts.CallSites = append(facts.CallSites, cs)
	}
	// Code interpolated in a literal or a heredoc body is one token the forest does not read, and perl compiles it
	// into ordinary ops, so its statement hedges every marker, over the body's lines too.
	// shortcut: parse.InterpolatesCode is asked of every quote and body, so '${\ x}' or a <<'E' body holding code
	// hedges where perl compiles nothing (TestOracleSubject's singlequote.pl and quotedheredoc.pl); a hedge explains
	// only a deficit, so it costs nothing there; upgrade if parse exports which literals interpolate.
	hedge := func(at, end int) {
		if first, last, ok := lines(at, end); ok {
			for kind := kindReference; kind < nkinds; kind++ {
				facts.CallSites = append(facts.CallSites, parseoracle.SubjectCallSite{Kind: kindName[kind], Line: first, EndLine: last, Unresolved: true})
			}
		}
	}
	for i, t := range s.Tokens {
		if raw := s.Raw[t.Index]; raw.Kind == lexer.Quote && parse.InterpolatesCode(string(src[raw.Start:raw.End])) {
			hedge(i, 0)
		}
	}
	for _, b := range s.Bodies {
		if parse.InterpolatesCode(string(src[b.Start:b.End])) {
			hedge(b.Opener, lineAt(starts, b.End-1))
		}
	}
	return facts
}

// bareProto is a prototype as prototype() returns it: without its parens or blanks.
func bareProto(p string) string {
	return strings.ReplaceAll(strings.TrimSuffix(strings.TrimPrefix(p, "("), ")"), " ", "")
}

// walk visits every node a derivation of the root reaches: it records the statements, every site in all, and, at each node with more than one alternative, the sites its alternatives disagree on in
// uncertain.
func (sf *subjectFacts) walk() {
	f := sf.f
	seen := make([]bool, len(f.Nodes))
	seen[f.Root] = true
	var spans [][2]int
	stack := []int32{f.Root}
	for len(stack) > 0 {
		n := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		nd := f.Nodes[n]
		if int(nd.Sym) == sf.g.fullstmt && nd.To > nd.From && int(nd.From) < len(sf.s.Tokens) {
			sp := [2]int{int(nd.From), min(int(nd.To), len(sf.s.Tokens))}
			spans = append(spans, sp)
		}
		if k, ok := sf.leafSite(n); ok {
			sf.all[k] = true
		}
		alts := 0
		for x := nd.Alt; x != 0; x = f.Alts[x].Next {
			alts++
			a := f.Alts[x]
			if k, unsure, ok := sf.altSite(n, a); ok {
				sf.all[k] = true
				sf.uncertain[k] = sf.uncertain[k] || unsure
			}
			decided, hedged := sf.derefSites(n, a)
			for _, k := range decided {
				sf.all[k] = true
			}
			for _, k := range hedged {
				sf.all[k], sf.uncertain[k] = true, true
			}
			ks, unsure := sf.refSites(n, a)
			for _, k := range ks {
				sf.all[k] = true
				sf.uncertain[k] = sf.uncertain[k] || unsure
			}
			for _, c := range [2]int32{a.Left, a.Right} {
				if c != 0 && !seen[c] {
					seen[c] = true
					stack = append(stack, c)
				}
			}
		}
		if alts > 1 {
			must, may, _ := sf.sets(n)
			for _, k := range may {
				if _, ok := slices.BinarySearch(must, k); !ok {
					sf.uncertain[k] = true
				}
			}
		}
	}
	slices.SortFunc(spans, func(x, y [2]int) int { return x[0] - y[0] })
	for _, sp := range spans { // nested and overlapping statements merge into the top-level one
		if l := len(sf.stmts) - 1; l >= 0 && sp[0] < sf.stmts[l][1] {
			sf.stmts[l][1] = max(sf.stmts[l][1], sp[1])
			continue
		}
		sf.stmts = append(sf.stmts, sp)
	}
}

// sets is the sites every derivation of node n holds and those any does, sorted, and ok false where n has no
// derivation off the current path.
func (sf *subjectFacts) sets(n int32) (must, may []int, ok bool) {
	if m, done := sf.may[n]; done {
		return sf.must[n], m, true
	}
	if sf.busy[n] || sf.failed[n] {
		return nil, nil, false
	}
	f := sf.f
	nd := f.Nodes[n]
	if nd.Alt == 0 {
		if k, has := sf.leafSite(n); has {
			must = []int{k}
		}
		sf.must[n], sf.may[n] = must, must
		return must, must, true
	}
	sf.busy[n] = true
	first := true
	for x := nd.Alt; x != 0; x = f.Alts[x].Next {
		a := f.Alts[x]
		var am, ay []int
		live := true
		for _, c := range [2]int32{a.Left, a.Right} {
			if c == 0 {
				continue
			}
			cm, cy, cok := sf.sets(c)
			if !cok {
				live = false
				break
			}
			am, ay = union(am, cm), union(ay, cy)
		}
		if !live {
			continue
		}
		if k, unsure, has := sf.altSite(n, a); has {
			if !unsure {
				am = union(am, []int{k})
			}
			ay = union(ay, []int{k})
		}
		if decided, hedged := sf.derefSites(n, a); len(decided)+len(hedged) > 0 {
			slices.Sort(decided)
			am, ay = union(am, decided), union(union(ay, decided), hedged)
		}
		if ks, unsure := sf.refSites(n, a); len(ks) > 0 {
			if !unsure {
				am = union(am, ks)
			}
			ay = union(ay, ks)
		}
		if first {
			must, may, first = am, ay, false
			continue
		}
		must, may = intersect(must, am), union(may, ay)
	}
	delete(sf.busy, n)
	if first {
		sf.failed[n] = true
		return nil, nil, false
	}
	sf.must[n], sf.may[n] = must, may
	return must, may, true
}

// altSite is the site alternative a of node n stands for, if any: its rule's, at the token the rule anchors it to,
// and unsure where a hash read's count is not in the text.
func (sf *subjectFacts) altSite(n int32, a Alt) (k int, unsure, ok bool) {
	nd := sf.f.Nodes[n]
	if nd.Sym < 0 || sf.g.kind[a.Rule] < 0 {
		return 0, false, false // a rule position is no site; its rule's node is
	}
	at := int(nd.From)
	switch {
	case sf.g.kind[a.Rule] == kindCall && sf.s.Raw[sf.s.Tokens[at].Index].Kind != lexer.Word:
		return 0, false, false // &f(...) bypasses the prototype, so it decides nothing
	case int(a.Rule) == sf.g.matchop:
		right := sf.f.Nodes[a.Right]
		if right.To == right.From+1 && quoteOp(sf.text(int(right.From))) != "" {
			return 0, false, false // a pattern operand is its own site, or none for s/// and tr///
		}
		at = int(right.From) - 1
	case sf.g.kind[a.Rule] == kindHash:
		if unsure, ok = sf.hashSite(n, a); !ok {
			return 0, false, false
		}
		at = int(nd.To) - 1
	}
	return site(sf.g.kind[a.Rule], at), unsure, true
}

// hashSite is whether hash read a of node n builds an rv2hv, and unsure where that is not in the text. A name toke.c's
// S_pending_ident finds in the pad is a PADANY that op.c's Perl_newHVREF makes padhv; any other is rv2hv over a gv
// (perl -MO=Concise: `my %h; keys %h` is padhv, `our %h; keys %h` rv2hv). A dereference is rv2hv, `%h[...]` and
// `->%[...]` slice an array, and an element level is peep.c's S_maybe_multideref's (elementSite).
func (sf *subjectFacts) hashSite(n int32, a Alt) (unsure, ok bool) {
	nd := sf.f.Nodes[n]
	if h := sf.g.hash[a.Rule]; (h == hashNamed || h == hashRef) && int(nd.To) < len(sf.s.Tokens) && sf.text(int(nd.To)) == "[" {
		return false, false // toke.c's yyl_percent reads %name[ as the array's key/value slice
	}
	if sf.g.hash[a.Rule] == hashRef && sf.g.deref[a.Rule] == derefPost && sf.g.sigil[a.Rule] == '?' && sf.absorbs(sf.child(a, 2)) {
		return false, false
	}
	switch sf.g.hash[a.Rule] {
	case hashNamed:
		if sf.refaliased(n) {
			return false, false
		}
		return sf.named(int(nd.From))
	case hashSlice:
		if base := sf.f.Nodes[sf.child(a, 3)]; base.To == base.From+1 && len(sf.text(int(base.From))) > 1 {
			return sf.named(int(base.From)) // @h{...}; @$r{...} and ->@{...} read through a reference
		} else if b := sf.f.Alts[base.Alt]; base.Alt != 0 && sf.g.deref[b.Rule] == derefPost && sf.absorbs(sf.child(b, 2)) {
			return false, false
		}
	case hashLevel:
		return sf.elementSite(a)
	}
	return false, true
}

// named is whether the named hash at token k builds an rv2hv: none for a lexical, one for a package hash (our's
// alias included, whose declaration is rv2hv vK/OURINTR), unsure where a feature that may be off declares it.
func (sf *subjectFacts) named(k int) (unsure, ok bool) {
	switch sf.hashes.at(sf.text(k), k) {
	case hashLexical:
		return false, false
	case hashUnknown:
		return true, true
	}
	return false, true
}

// elementSite is whether element level a keeps its rv2hv where the chain it tops is read, after peep.c's
// S_maybe_multideref: a level through a reference or a package hash folds into multideref (Concise:
// `$r->{a}{b}{c}` is multideref($r->{"a"}{"b"}{"c"})) while its key is a constant or a plain scalar; a computed key
// ends the run and keeps its level's rv2hv unless the run before it folded, which then takes the rv2hv as its last
// action (`$r->{"x$k"}{a}` keeps one, `our %h; $h{a}{"x$k"}` none: multideref($h{"a"}->)). A key that may fold to a
// constant, a call to a constant sub say, is unsure. derefSites adds what a dereference's context keeps.
// shortcut: after a range whose constant fold perl abandons (`"*x".."az"`, not `"*x"..3`), perl folds nothing more of
// the main program's statements into multideref, so a read there keeps its rv2hv where the chart reports none; perl.git
// op/range.t:301 and :446 are missed, graded no-answer only because nothing else in the file is a hash site
// (TestOracleSubjectHashSites); upgrade when the cause in peep.c is known.
func (sf *subjectFacts) elementSite(a Alt) (unsure, ok bool) {
	through, maybe := sf.through(a)
	if !through && !maybe {
		return false, false // a lexical's element: padhv, folded or not
	}
	switch sf.key(sf.child(a, 1)) {
	case keySimple:
		return false, false
	case keyUnknown:
		return true, true
	}
	if _, b, ok := sf.levelOf(sf.child(a, sf.g.level[a.Rule].base)); ok {
		switch sf.key(sf.child(b, 1)) {
		case keySimple:
			return false, false // the run below folded and takes this rv2hv
		case keyUnknown:
			return true, true
		}
	}
	return maybe, true
}

// through is whether element level a reads a hash through a reference or a package hash, and maybe where a feature
// that may be off declares its name.
func (sf *subjectFacts) through(a Alt) (yes, maybe bool) {
	sh := sf.g.level[a.Rule]
	if !sh.brace {
		return false, false
	}
	if base := sf.f.Nodes[sf.child(a, sh.base)]; !sh.arrow && base.To == base.From+1 {
		maybe, yes = sf.named(int(base.From)) // $h{...}; ${...}{...} and $$r{...} read through a reference
		return yes && !maybe, maybe
	}
	return true, false
}

// What op_lvalue and the ops taking an aggregate (op.c's Perl_doref, called by ck_fun for an OA_AVREF or OA_HVREF
// operand and by op_lvalue for a modified rv2av or rv2hv) leave on the element a dereference reads: nothing, where
// it is read; OPpDEREF_SV, under a modified ${...}; OPpDEREF_AV or _HV, under a modified, sliced or aggregate-taking
// @{...} or %{...}; or a context the chart does not read.
const (
	ctxRead = iota
	ctxScalar
	ctxAggregate
	ctxUnknown
)

// derefSites is the sites the element chains a dereference, local or \ reads carry beyond elementSite's: decided
// ones for the levels peep.c's S_maybe_multideref leaves out of multideref under the context perl gives it, and a
// hedge at the dereference (its sigil, its ->, local or \) where that context is not read. Under OPpDEREF_AV or _HV
// the run holding the last level does not fold, so each of its levels through a reference or a package hash keeps
// its rv2hv (Concise: `foreach (@{$r->{a}})` keeps one under helem sKM/DREFAV, `my @x = @{$r->{a}}` is
// multideref($r->{"a"})); under OPpDEREF_SV the last level keeps it only where it starts its run (`${$r->{a}} = 1`
// keeps one, `${$r->{a}{b}} = 1` none). `delete local` keeps the last level as OPpDEREF_SV does, and a refassign's
// `\$h{a} = ...` (op/lvref.t:148) is a context the chart does not read.
func (sf *subjectFacts) derefSites(n int32, a Alt) (decided, hedged []int) {
	nd := sf.f.Nodes[n]
	if nd.Sym < 0 || sf.g.deref[a.Rule] == 0 {
		return nil, nil
	}
	at, op, ctx := int(nd.From), sf.f.Nodes[a.Right], ctxUnknown
	mod := func(sigil byte) int {
		if sigil == '$' {
			return ctxScalar
		}
		return ctxAggregate
	}
	switch sigil := sf.g.sigil[a.Rule]; sf.g.deref[a.Rule] {
	case derefCast, derefPost:
		if sf.g.deref[a.Rule] == derefPost {
			op = sf.f.Nodes[sf.child(a, 2)]
			at = int(op.To)
			if sigil == '$' && sf.text(at+1) == "**" {
				sigil = '*'
			}
		} else if sf.text(int(op.From)) == "{" && op.To > op.From+1 {
			op.From, op.To = op.From+1, op.To-1
		}
		switch sigil {
		case '*':
			return nil, nil // *{...} and ->** fold as a read does
		case 'A', '?':
			return nil, nil // ->@{...}, ->@[...], ->%{...} and ->%[...] fold the chain: multideref($r->{"a"}->)
		case '@', '%', '$':
			if ctx = sf.context(n, sigil); ctx == ctxAggregate || ctx == ctxScalar {
				ctx = mod(sigil)
			}
		}
	case derefLocal:
		ctx = ctxRead // local $r->{a} = 1 and \$r->{a} fold
		switch w := sf.text(at); {
		case w == "local" && sf.owner(n) == "delete":
			ctx = ctxScalar
		case w != "local" && sf.assigned(n):
			ctx = ctxUnknown
		}
	}
	if ctx == ctxRead {
		return nil, nil
	}
	tops, other := sf.chains(int(op.From), int(op.To))
	if other || ctx == ctxUnknown && len(tops) > 0 {
		return nil, []int{site(kindHash, at)}
	}
	for _, t := range tops {
		d, unsure := sf.kept(t, ctx)
		decided = append(decided, d...)
		if unsure {
			hedged = []int{site(kindHash, at)}
		}
	}
	return decided, hedged
}

// chains is the element chains reading through a reference or a package hash that tokens from to to are, through ?:
// branches, and other where they hold such a chain otherwise, or under a ?: whose condition may be constant, which
// perl folds to one branch.
func (sf *subjectFacts) chains(from, to int) (tops []int32, other bool) {
	i, _ := slices.BinarySearchFunc(sf.levels, int32(from), func(l, from int32) int { return int(sf.f.Nodes[l].From - from) })
	for j := i; j < len(sf.levels) && int(sf.f.Nodes[sf.levels[j]].From) == from; j++ {
		if int(sf.f.Nodes[sf.levels[j]].To) == to {
			if sf.reads(sf.levels[j]) {
				return []int32{sf.levels[j]}, false
			}
			return nil, false
		}
	}
	q, c, depth, constant := -1, -1, 0, true
	for k := from; k < to; k++ {
		switch t := sf.text(k); {
		case t == "(" || t == "[" || t == "{":
			depth++
		case t == ")" || t == "]" || t == "}":
			depth--
		case depth == 0 && t == "?" && q < 0:
			q = k
		case depth == 0 && t == ":" && q >= 0 && c < 0:
			c = k
		}
		if kind := sf.s.Raw[sf.s.Tokens[k].Index].Kind; q < 0 && (kind == lexer.Variable || kind == lexer.DerefSigil) {
			constant = false
		}
	}
	if q > from && c > q && !constant {
		t1, o1 := sf.chains(q+1, c)
		t2, o2 := sf.chains(c+1, to)
		return append(t1, t2...), o1 || o2
	}
	for ; i < len(sf.levels) && int(sf.f.Nodes[sf.levels[i]].From) < to; i++ {
		if int(sf.f.Nodes[sf.levels[i]].To) <= to && sf.reads(sf.levels[i]) {
			return nil, true
		}
	}
	return nil, false
}

// kept is the rv2hv sites chain top keeps under ctx beyond elementSite's, and unsure where a level's answer is not in
// the text: none where its last key is computed (elementSite counts that level); under ctxScalar the last level where
// it starts its run; under ctxAggregate each level of the run of constant or plain-scalar keys ending at the last.
func (sf *subjectFacts) kept(top int32, ctx int) (sites []int, unsure bool) {
	for l := top; ; {
		_, a, ok := sf.levelOf(l)
		if !ok || sf.key(sf.child(a, 1)) != keySimple {
			return sites, unsure || ok && sf.key(sf.child(a, 1)) == keyUnknown
		}
		base := sf.child(a, sf.g.level[a.Rule].base)
		if _, b, ok := sf.levelOf(base); ctx == ctxScalar && ok && sf.key(sf.child(b, 1)) == keySimple {
			return nil, false // the run below folds and takes the last level
		}
		switch yes, maybe := sf.through(a); {
		case maybe:
			unsure = true
		case yes:
			sites = append(sites, site(kindHash, int(sf.f.Nodes[l].To)-1))
		}
		if ctx == ctxScalar {
			return sites, unsure
		}
		l = base
	}
}

// absorbs is whether node n, the term a postfix slice ->@{...} or ->%{...} subscripts, is an element level whose key
// S_maybe_multideref holds, which then takes the slice's rv2hv as its last action (Concise: `$r->{a}->@{qw(x y)}` and
// `my %h; $h{a}->@{qw(x y)}` are multideref($r->{"a"}->) and multideref($h{"a"}->); `$r->{"x$k"}->@{qw(x y)}` keeps
// both rv2hv).
func (sf *subjectFacts) absorbs(n int32) bool {
	_, a, ok := sf.levelOf(n)
	return ok && sf.key(sf.child(a, 1)) == keySimple
}

// reads is whether some level of the chain node n tops reads a hash through a reference or a package hash.
func (sf *subjectFacts) reads(n int32) bool {
	for {
		_, a, ok := sf.levelOf(n)
		if !ok {
			return false
		}
		if yes, maybe := sf.through(a); yes || maybe {
			return true
		}
		n = sf.child(a, sf.g.level[a.Rule].base)
	}
}

// levelOf is the element level node n is, through unit rules, and its first such alternative.
func (sf *subjectFacts) levelOf(n int32) (int32, Alt, bool) {
	for {
		nd := sf.f.Nodes[n]
		if nd.Alt == 0 {
			return 0, Alt{}, false
		}
		a := sf.f.Alts[nd.Alt]
		switch {
		case sf.g.level[a.Rule].ok:
			return n, a, true
		case a.Left == 0 && a.Right != 0:
			n = a.Right
		default:
			return 0, Alt{}, false
		}
	}
}

// What S_maybe_multideref makes of a subscript's key: a constant (after folding) or a plain scalar, which it holds;
// a computed value, which it cannot; or one that may fold to a constant, which the text does not settle.
const (
	keySimple = iota
	keyComputed
	keyUnknown
)

// key is what subscript key n is: one number, word, uninterpolated string or plain $name; a -word or -number; else
// computed where a variable is read, constant where only literals and operators are, and unknown otherwise (a word
// beside others may be a call to a constant sub). Concise: `$r->{-a}`, `$r->{1+1}`, `$r->{$_}` and `$r->{'a$b'}`
// are multideref, `$r->{@a}` and `$r->{"x$k"}` keep their rv2hv.
func (sf *subjectFacts) key(n int32) int {
	from, to := int(sf.f.Nodes[n].From), int(sf.f.Nodes[n].To)
	kind := func(k int) lexer.Kind { return sf.s.Raw[sf.s.Tokens[k].Index].Kind }
	if to > from+1 && sf.text(from) == "+" && kind(from+1) == lexer.Variable {
		from++ // unary plus builds no op (Concise: $h{+$x} is multideref($h{$x})); +word is a call, not a string
	}
	if to == from+1 && sf.text(from) == "$$" {
		return keySimple // a punctuation scalar is a gvsv (Concise: $h{$$} is multideref($h{$$}))
	}
	if to == from+1 {
		switch t := sf.text(from); kind(from) {
		case lexer.Number, lexer.Word:
			return keySimple
		case lexer.Variable:
			if t[0] == '$' && !strings.HasPrefix(t, "$#") && hashName(t) != "" {
				return keySimple
			}
		}
	}
	if to == from+2 && sf.text(from) == "-" && (kind(from+1) == lexer.Word || kind(from+1) == lexer.Number) {
		return keySimple
	}
	class := keySimple
	for k := from; k < to; k++ {
		switch kind(k) {
		case lexer.Variable, lexer.DerefSigil:
			return keyComputed
		case lexer.Quote:
			switch quoteKey(sf.text(k)) {
			case keyComputed:
				return keyComputed
			case keyUnknown:
				class = keyUnknown
			}
		case lexer.Operator:
			if t := sf.text(k); t == "," || t == "=>" {
				return keyComputed // $h{'a', 'b'} is join($;, ...), reading $; (Concise: op/magic.t:237)
			}
		case lexer.Number, lexer.CloseBracket:
		default:
			class = keyUnknown
		}
	}
	return class
}

// perlSpace is the bytes perl's isSPACE takes (handy.h isSPACE_A): toke.c skips them before a quote's delimiter and
// qw splits on them.
const perlSpace = " \t\n\r\v\f"

// quoteKey is quote t's key class. A string is a constant unless it interpolates (toke.c scan_const): a $ before
// anything, an @ before a name or one of :'{$+-, unsure before a non-ASCII byte, which use utf8 may make a name; \c
// takes the next character. A qw or qq whose delimiter follows whitespace and is # is unsure, since that # starts a
// comment. A qq delimited by $ or @ is unsure: scan_str drops the \ before such a delimiter, so an
// escaped one interpolates. qw/a b/ is a list, joined with $; as a , is, and an empty qw is computed too (Concise:
// rv2hv). A v-string is a constant; qr, m, s, tr, y, a match's / and a backtick or qx are ops. Anything else is unsure.
func quoteKey(t string) int {
	var body string
	switch {
	case t[0] == '\'' || t[0] == 'q' && !strings.ContainsRune("qwxr", rune(t[1])):
		return keySimple
	case (strings.HasPrefix(t, "qw") || strings.HasPrefix(t, "qq")) && t[2] != '#' && strings.TrimLeft(t[2:], perlSpace)[0] == '#':
		return keyUnknown // after whitespace a # starts a comment (toke.c), so the body is past it
	case strings.HasPrefix(t, "qw"):
		if w := strings.TrimLeft(t[2:], perlSpace); len(strings.FieldsFunc(w[1:len(w)-1], func(r rune) bool { return strings.ContainsRune(perlSpace, r) })) != 1 {
			return keyComputed
		}
		return keySimple
	case t[0] == '"':
		body = t[1 : len(t)-1]
	case strings.HasPrefix(t, "qq"):
		body = strings.TrimLeft(t[2:], perlSpace)
		if body[0] == '$' || body[0] == '@' {
			return keyUnknown
		}
		body = body[1 : len(body)-1]
	case t[0] == 'v' || t[0] >= '0' && t[0] <= '9':
		return keySimple // a v-string, v1.2 or 1.2.3
	case t[0] == '`' || t[0] == '/' || slices.ContainsFunc([]string{"qr", "qx", "m", "s", "tr", "y"}, func(op string) bool {
		return strings.HasPrefix(t, op) && len(t) > len(op) && !isWordByte(t[len(op)])
	}):
		return keyComputed
	default:
		return keyUnknown
	}
	class := keySimple
	for i := 0; i < len(body); i++ {
		switch body[i] {
		case '\\':
			if i+1 < len(body) && body[i+1] == 'c' {
				i++
			}
			i++
		case '$':
			return keyComputed
		case '@':
			if i+1 < len(body) && (isWordByte(body[i+1]) || strings.IndexByte(":'{$+-", body[i+1]) >= 0) {
				return keyComputed
			}
			if i+1 < len(body) && body[i+1] >= 0x80 {
				class = keyUnknown
			}
		}
	}
	return class
}

// refSites is the references alternative a of node n takes, sorted, each at its operand's first token, and unsure
// where some readings of an operand take one and others do not: op.c's ref_array_or_hash takes one for each operand
// of ~~, given's condition or when's that is an array or a hash (not a slice or an anonymous array), and
// ck_entersub_args_proto one for each argument a declared sub's prototype wraps (protoRefs).
func (sf *subjectFacts) refSites(n int32, a Alt) (sites []int, unsure bool) {
	if sf.g.calls[a.Rule].ok {
		return sf.protoRefs(n, a)
	}
	backs := sf.g.refs[a.Rule]
	if backs == nil || int(a.Rule) == sf.g.smartmatch && sf.text(int(sf.f.Nodes[sf.child(a, 1)].From)) != "~~" {
		return nil, false
	}
	for _, back := range backs {
		o := sf.child(a, back)
		if yes, no := sf.is(o, sf.g.aggregate); yes {
			sites, unsure = append(sites, site(kindReference, int(sf.f.Nodes[o].From))), unsure || no
		}
	}
	return sites, unsure
}

// protoRefs is the references ck_entersub_args_proto takes in call alternative a of node n, sorted: where the word
// calls a declared sub with a prototype, one for each argument at a \ slot, \[...] included, at a + slot that is an
// array or a hash, and at a * slot that is a glob; none at a $, _, & or ; slot, nor from a @ or % slot on. Where the
// arguments are not a list of terms every reading divides the same way, the call's own token is an unsure site.
func (sf *subjectFacts) protoRefs(n int32, a Alt) (sites []int, unsure bool) {
	if sf.f.Nodes[n].Sym < 0 {
		return nil, false // a rule position is no call; its rule's node is
	}
	at, shape := int(sf.f.Nodes[n].From), sf.g.calls[a.Rule]
	proto, ok := sf.proto(at)
	if !ok || sf.s.Raw[sf.s.Tokens[at].Index].Kind != lexer.Word || !strings.ContainsAny(proto, `\*+`) {
		return nil, false // &f(...) bypasses the prototype
	}
	args, ok := []int32{sf.child(a, shape.back)}, true
	if !shape.one {
		args, ok = sf.terms(args[0])
	}
	if !ok {
		return []int{site(kindReference, at)}, true
	}
	p := bareProto(proto)
	if shape.block {
		p = strings.TrimPrefix(strings.TrimLeft(p, ";"), "&")
	}
	for _, arg := range args {
		if p = strings.TrimLeft(p, ";"); p == "" || p[0] == '@' || p[0] == '%' {
			break
		}
		var yes, no bool
		switch p[0] {
		case '\\':
			yes, p = true, p[1:]
			if strings.HasPrefix(p, "[") {
				p = p[max(strings.IndexByte(p, ']'), 0):]
			}
		case '+':
			yes, no = sf.is(arg, sf.g.aggregate)
		case '*':
			yes, no = sf.is(arg, sf.g.glob)
		}
		if p != "" {
			p = p[1:]
		}
		if yes {
			sites, unsure = append(sites, site(kindReference, int(sf.f.Nodes[arg].From))), unsure || no
		}
	}
	return sites, unsure
}

// terms is the terms of argument list n, an optlistexpr or an expr, in order, and ok false where n is no list of
// terms, or one some node of divides more than one way.
func (sf *subjectFacts) terms(n int32) (ts []int32, ok bool) {
	f := sf.f
	for {
		nd := f.Nodes[n]
		if nd.Alt == 0 || f.Alts[nd.Alt].Next != 0 {
			return nil, false
		}
		switch a := f.Alts[nd.Alt]; int(a.Rule) {
		case sf.g.optEmpty:
			return ts, true
		case sf.g.exprList, sf.g.optList:
			n = a.Right
		case sf.g.listOne:
			return append([]int32{a.Right}, ts...), true
		case sf.g.listComma:
			ts, n = append([]int32{a.Right}, ts...), sf.child(a, 2)
		case sf.g.listTrail:
			n = sf.child(a, 1)
		default:
			return nil, false
		}
	}
}

// is is whether some readings of node n are a term of one of rules, through unit rules and parentheses, and whether
// some are not.
func (sf *subjectFacts) is(n int32, rules []bool) (yes, no bool) {
	f := sf.f
	if f.Nodes[n].Alt == 0 {
		return false, true
	}
	for x := f.Nodes[n].Alt; x != 0; x = f.Alts[x].Next {
		a := f.Alts[x]
		var y, o bool
		switch {
		case rules[a.Rule]:
			y = true
		case int(a.Rule) == sf.g.paren:
			y, o = sf.is(sf.child(a, 1), rules)
		case a.Left == 0 && a.Right != 0:
			y, o = sf.is(a.Right, rules)
		default:
			o = true
		}
		yes, no = yes || y, no || o
	}
	return yes, no
}

// child is the node of the symbol back places before the last of a's rule, through each rule position's first
// alternative: a rule position's span ends where the next symbol's starts, and the rules refSites reads put a token
// or a closed group there, so it divides one way.
func (sf *subjectFacts) child(a Alt, back int) int32 {
	for ; back > 0; back-- {
		if sf.f.Nodes[a.Left].Sym >= 0 {
			return a.Left // the rule's only symbol before its last
		}
		a = sf.f.Alts[sf.f.Nodes[a.Left].Alt]
	}
	return a.Right
}

// leafSite is the site leaf n stands for, if any: a pattern read as PMFUNC that matches, a <...> that reads a
// line, or the word readline read as the builtin.
func (sf *subjectFacts) leafSite(n int32) (int, bool) {
	nd := sf.f.Nodes[n]
	if nd.Alt != 0 || nd.Sym < 0 || nd.To != nd.From+1 {
		return 0, false
	}
	k := int(nd.From)
	t, text := sf.s.Raw[sf.s.Tokens[k].Index], sf.text(k)
	switch {
	case int(nd.Sym) == sf.g.pmfunc && (quoteOp(text) == "m"):
		return site(kindMatch, k), true
	case t.Kind == lexer.Readline && (text == "<<>>" || parse.IsReadlineTerm(text)) &&
		(len(text) == utf8.RuneCountInString(text) || sf.utf8At(sf.s.Tokens[k].Index)):
		// shortcut: parse.IsReadlineTerm calls <<>> a glob where toke.c reads readline; drop the <<>> clause when
		// parse's predicate learns it (recursive descent's answer moves then, so its own issue).
		return site(kindReadline, k), true
	case t.Kind == lexer.Word && (text == "readline" || text == "CORE::readline") && !sf.g.notCore[nd.Sym]:
		return site(kindReadline, k), true
	}
	return 0, false
}

// quoteOp is the quote-like operator a pattern token spells: m for a match, whether m// or //, s, tr or y, or ""
// for anything else.
func quoteOp(text string) string {
	if strings.HasPrefix(text, "/") || strings.HasPrefix(text, "?") {
		return "m"
	}
	for _, op := range []string{"tr", "m", "s", "y"} {
		if rest, ok := strings.CutPrefix(text, op); ok && rest != "" && !isWordByte(rest[0]) {
			return op
		}
	}
	return ""
}

func isWordByte(c byte) bool {
	return c == '_' || c >= '0' && c <= '9' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
}

// utf8At is whether use utf8 is in force at raw token at: a use or no utf8 statement sets it for the rest of its
// block, as lex reads the isa feature.
// shortcut: a utf8 a BEGIN block, a string eval or use if turns on is not seen; upgrade if a corpus file needs one.
func (sf *subjectFacts) utf8At(at int) bool {
	raw := sf.s.Raw
	on, scopes := false, []bool{}
	for i, t := range raw[:at] {
		switch w := sf.src[t.Start:t.End]; {
		case t.Kind == lexer.Operator && string(w) == "{":
			scopes = append(scopes, on)
		case t.Kind == lexer.CloseBracket && string(w) == "}" && len(scopes) > 0:
			on, scopes = scopes[len(scopes)-1], scopes[:len(scopes)-1]
		case t.Kind == lexer.Word && (string(w) == "use" || string(w) == "no"):
			next := func(j int) (int, string) {
				for j++; j < at && lexer.IsTrivia(raw[j].Kind); j++ {
				}
				if j < at {
					return j, string(sf.src[raw[j].Start:raw[j].End])
				}
				return j, ""
			}
			// an empty list, () or qw(), calls neither import nor unimport, so it leaves utf8 as it was
			if j, word := next(i); word == "utf8" {
				j, arg := next(j)
				_, after := next(j)
				body := strings.TrimSpace(strings.TrimPrefix(arg, "qw"))
				if !(arg == "(" && after == ")" || arg != body && len(body) >= 2 && strings.TrimSpace(body[1:len(body)-1]) == "") {
					on = w[0] == 'u'
				}
			}
		}
	}
	return on
}

// statement is the top-level statement holding token at.
func (sf *subjectFacts) statement(at int) ([2]int, bool) {
	i, _ := slices.BinarySearchFunc(sf.stmts, at+1, func(st [2]int, k int) int { return st[0] - k })
	if i == 0 || at >= sf.stmts[i-1][1] {
		return [2]int{}, false
	}
	return sf.stmts[i-1], true
}

func (sf *subjectFacts) text(k int) string {
	t := sf.s.Raw[sf.s.Tokens[k].Index]
	return string(sf.src[t.Start:t.End])
}

// lineStarts is the byte each line of src starts at.
func lineStarts(src []byte) []int {
	starts := []int{0}
	for i, c := range src {
		if c == '\n' {
			starts = append(starts, i+1)
		}
	}
	return starts
}

// lineAt is the 1-based line holding byte off.
func lineAt(starts []int, off int) int {
	i, found := slices.BinarySearch(starts, off)
	if found {
		return i + 1
	}
	return i
}

// union is the sorted union of two sorted sets, reusing either where the other adds nothing.
func union(x, y []int) []int {
	if len(y) == 0 {
		return x
	}
	if len(x) == 0 {
		return y
	}
	out := make([]int, 0, len(x)+len(y))
	i, j := 0, 0
	for i < len(x) || j < len(y) {
		switch {
		case j == len(y) || i < len(x) && x[i] < y[j]:
			out = append(out, x[i])
			i++
		case i == len(x) || y[j] < x[i]:
			out = append(out, y[j])
			j++
		default:
			out = append(out, x[i])
			i, j = i+1, j+1
		}
	}
	return out
}

// intersect is the sorted intersection of two sorted sets.
func intersect(x, y []int) []int {
	var out []int
	for _, k := range x {
		if _, ok := slices.BinarySearch(y, k); ok {
			out = append(out, k)
		}
	}
	return out
}

// What a hash name is at a token: a package hash (the default, and our's alias of one), a lexical, or one a feature
// that may be off declares.
const (
	hashPackage = iota
	hashLexical
	hashUnknown
)

// hashScopes is the hash declarations a file makes: each declared name's scopes, and what each declaring token is.
type hashScopes struct {
	decls    map[string][]hashDecl
	declared map[int]int
	exports  bool // the file names export_lexically
}

type hashDecl struct{ from, to, kind int }

// at is what the hash a token's text names is at token k: the declaration its own token makes, else the innermost
// in scope, else a package hash. A qualified name is always a package's.
// In a file that names export_lexically, a package hash is unsure: the call may have made it lexical (Concise: `BEGIN {
// builtin::export_lexically("%h" => {}) } keys %h` is padhv).
// shortcut: a lexical %h a module's import installs with export_lexically is not seen, so its uses read as the
// package hash; upgrade if parse.Imports reports lexical exports.
func (h *hashScopes) at(text string, k int) int {
	if kind, ok := h.declared[k]; ok {
		return kind
	}
	kind, best := hashPackage, -1
	for _, d := range h.decls[hashName(text)] {
		if d.from <= k && k < d.to && d.from > best {
			kind, best = d.kind, d.from
		}
	}
	if kind == hashPackage && h.exports && !strings.Contains(hashName(text), "::") {
		return hashUnknown
	}
	return kind
}

// hashScopes reads the file's hash declarations: my, state and field hashes, lexical from the end of their statement
// (perly.y's intro_my) to their block's end, or to the end of the if, unless, while, until or for whose condition
// declares them; our's, the package hash, over the same span; and a signature's, over its sub's body.
func (r *resolver) hashScopes() *hashScopes {
	h := &hashScopes{decls: map[string][]hashDecl{}, declared: map[int]int{}, exports: strings.Contains(string(r.src), "export_lexically")}
	feature := func(k int, f lex.Feature) (int, bool) {
		switch fs := r.s.FeaturesAt(k); {
		case fs.On&f != 0:
			return hashLexical, true
		case fs.Maybe&f != 0:
			return hashUnknown, true
		}
		return 0, false
	}
	for k := range r.n {
		var kind int
		var ok bool
		switch w := r.text(k); {
		case w == "my":
			kind, ok = hashLexical, true
		case w == "our":
			kind, ok = hashPackage, true
		case w == "state":
			kind, ok = feature(k, lex.FeatureState)
		case w == "field":
			kind, ok = feature(k, lex.FeatureClass)
		case (w == "sub" || w == "method") && r.isWord(k):
			r.signature(h, k, feature)
			continue
		}
		if !ok || !r.isWord(k) {
			continue
		}
		vars, refalias := []int{k + 1}, r.is(k+1, "\\")
		if k+2 < r.n && (r.isWord(k+1) && !r.is(k+2, "(") || r.is(k+1, "\\")) { // my Class %h, my \%h
			vars[0] = k + 2
		}
		if j := vars[0]; j < r.n && r.is(j, "(") {
			vars = nil
			for d := 0; j < r.n; j++ {
				if r.is(j, "(") {
					d++
				} else if r.is(j, ")") {
					if d--; d == 0 {
						break
					}
				}
				vars = append(vars, j)
			}
		}
		from, to := r.statementEnd(r.s.Raw[r.s.Tokens[r.introAt(k)].Index].Start), r.scopeEnd(k)
		at, _ := slices.BinarySearchFunc(r.s.Tokens, from, func(t lex.Token, b int) int { return r.s.Raw[t.Index].Start - b })
		if open, end := r.loopBody(k); end > 0 { // for my \%a (...) {...} and for \my %a: the alias is the body's
			at, to = open, end
		}
		for _, j := range vars {
			if j >= r.n {
				continue
			}
			if t := r.text(j); strings.HasPrefix(t, "%") && hashName(t) != "" && !strings.Contains(hashName(t), "::") {
				h.declared[j] = kind
				if refalias || r.is(j-1, "\\") { // our \%h, our (\%h, \%g)
					h.declared[j] = hashLexical // a refalias builds no rv2hv, our's included (Concise: our \%h = {})
				}
				h.decls[hashName(t)] = append(h.decls[hashName(t)], hashDecl{at, to, kind})
			}
		}
	}
	return h
}

// signature records the hash parameters of the signature after sub or method at token k, as one PROTOTYPE token or
// as code, over the sub's body: toke.c reads a signature where the feature is on, a prototype where it is off.
func (r *resolver) signature(h *hashScopes, k int, feature func(int, lex.Feature) (int, bool)) {
	j := k + 1
	if j < r.n && r.isWord(j) {
		j++
	}
	kind, ok := feature(k, lex.FeatureSignatures)
	if !ok || j >= r.n {
		return
	}
	start := r.s.Raw[r.s.Tokens[j].Index].Start
	if r.is(j, "(") {
		for d := 0; j < r.n; j++ {
			if r.is(j, "(") {
				d++
			} else if r.is(j, ")") {
				if d--; d == 0 {
					break
				}
			}
		}
	} else if r.s.Raw[r.s.Tokens[j].Index].Kind != lexer.Prototype {
		return
	}
	if j >= r.n {
		return
	}
	text := r.src[start:r.s.Raw[r.s.Tokens[j].Index].End]
	body := j
	for body < r.n && !r.is(body, "{") {
		body++
	}
	if body >= r.n || r.match[body] < 0 {
		return
	}
	// a parameter is a %name directly after ( or , at the signature's own depth; one in a default is an expression
	// (Concise: our %h; sub f ($x = \%h) {} reads rv2hv), and a % after a term is a modulus, which the lexer tells apart
	depth, prev, sigil := 0, "", false
	for _, t := range lexer.Tokenize(text) {
		s := string(text[t.Start:t.End])
		param := depth == 1 && (prev == "(" || prev == ",")
		if p := strings.TrimSpace(s[1:]); param && t.Kind == lexer.Variable && (p == "," || p == "=" || p == ")") {
			s = p // a placeholder: a bare sigil at a parameter's position, which the lexer reads as $, or $ = or @)
		}
		name := ""
		switch {
		case t.Kind == lexer.Whitespace || t.Kind == lexer.Comment:
			continue
		case s == ")" || s == "]" || s == "}":
			depth--
		case s == "(" || s == "[" || s == "{":
			depth++
		case t.Kind == lexer.Variable && param && s[0] == '%':
			name = hashName(s)
		case sigil && t.Kind == lexer.Word:
			name = s // ($, %h): after the placeholder $, the lexer reads % h as a modulus
		}
		if name != "" {
			h.decls[name] = append(h.decls[name], hashDecl{j, r.match[body], kind})
		}
		sigil = param && s == "%"
		prev = s
	}
}

// loopBody is the tokens opening the body of the foreach whose iterator a declaration at token k declares and closing
// its last block, a continue's included, or 0, 0: the alias is not in scope over the loop's own list (Concise: `our
// %a; for my \%a (\%a) {}` reads the package %a) and is over a continue block.
func (r *resolver) loopBody(k int) (open, end int) {
	before := k - 1
	if before > 0 && r.is(before, "\\") {
		before--
	}
	if before < 0 || !r.is(before, "for") && !r.is(before, "foreach") {
		return 0, 0
	}
	j := k
	for j < r.n && !r.is(j, "(") {
		j++
	}
	for depth := 0; j < r.n; j++ {
		if r.is(j, "(") {
			depth++
		} else if r.is(j, ")") {
			if depth--; depth == 0 {
				break
			}
		}
	}
	if j+1 >= r.n || !r.is(j+1, "{") || r.match[j+1] < 0 {
		return 0, 0
	}
	open, end = j+1, r.match[j+1]
	if end+2 < r.n && r.is(end+1, "continue") && r.is(end+2, "{") && r.match[end+2] >= 0 {
		end = r.match[end+2]
	}
	return open, end
}

// introAt is the token whose statement's end introduces a declaration at token k: k itself where it stands in a
// condition, whose ) introduces it, else the outermost parenthesis it stands in, so that `(my %h = ()), keys %h`
// reads the package hash (perly.y's intro_my at the statement's end).
func (r *resolver) introAt(k int) int {
	at, depth := k, 0
	for i := k - 1; i >= 0; i-- {
		switch {
		case r.is(i, ")") || r.is(i, "]"):
			depth++
		case (r.is(i, "(") || r.is(i, "[")) && depth > 0:
			depth--
		case r.is(i, "(") || r.is(i, "["):
			if r.is(i, "(") && i > 0 && slices.Contains([]string{"if", "unless", "elsif", "while", "until", "for", "foreach"}, r.text(i-1)) {
				return k
			}
			at = i
		case depth == 0 && (r.is(i, ";") || r.is(i, "{") || r.is(i, "}")):
			return at
		}
	}
	return at
}

// scopeEnd is the token after the scope a declaration at token k ends with: its block's, or, where it stands in the
// condition of an if, unless, elsif, while, until, for or foreach, the end of that statement's last block.
func (r *resolver) scopeEnd(k int) int {
	open, depth := -1, 0
	for i := k - 1; i > 0 && open < 0; i-- {
		switch {
		case r.is(i, ")"):
			depth++
		case r.is(i, "(") && depth > 0:
			depth--
		case r.is(i, "("):
			if slices.Contains([]string{"if", "unless", "elsif", "while", "until", "for", "foreach"}, r.text(i-1)) {
				open = i
			}
		case depth == 0 && (r.is(i, ";") || r.is(i, "{") || r.is(i, "}")):
			return r.blockEnd(k)
		}
	}
	end := open
	for open >= 0 {
		for depth = 0; end < r.n; end++ { // the condition's )
			if r.is(end, "(") {
				depth++
			} else if r.is(end, ")") {
				if depth--; depth == 0 {
					break
				}
			}
		}
		if end+1 >= r.n || !r.is(end+1, "{") || r.match[end+1] < 0 {
			return r.blockEnd(k)
		}
		end = r.match[end+1]
		switch {
		case end+2 < r.n && (r.is(end+1, "else") || r.is(end+1, "continue")) && r.is(end+2, "{") && r.match[end+2] >= 0:
			return r.match[end+2] + 1
		case end+2 < r.n && r.is(end+1, "elsif") && r.is(end+2, "("):
			end += 2
			continue
		}
		return end + 1
	}
	return r.blockEnd(k)
}

// hashName is the name a sigilled variable's text names, `h` for %h, % h, %{h}, $h or @h, and "" for a dereference.
func hashName(t string) string {
	if len(t) < 2 {
		return ""
	}
	name := strings.TrimSpace(t[1:])
	if inner, ok := strings.CutPrefix(name, "{"); ok {
		name = strings.TrimSpace(strings.TrimSuffix(inner, "}"))
	}
	if name == "" || name[0] == '$' || name[0] == '{' {
		return ""
	}
	return name
}

// upRef is one place a node stands: in an alternative of rule of node p, as its symbol idx, or as the rule position
// to the left of p's last symbol (idx -1).
type upRef struct {
	p, rule int32
	idx     int
}

// index records where each node reachable from the root stands, for context.
func (sf *subjectFacts) index() {
	f := sf.f
	sf.up = make([][]upRef, len(f.Nodes))
	seen := make([]bool, len(f.Nodes))
	seen[f.Root] = true
	for stack := []int32{f.Root}; len(stack) > 0; {
		n := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		nd := f.Nodes[n]
		for x := nd.Alt; x != 0; x = f.Alts[x].Next {
			a := f.Alts[x]
			if a.Right != 0 {
				idx := int(nd.Dot) - 1
				if nd.Sym >= 0 {
					idx = len(sf.g.rhs[a.Rule]) - 1
				}
				sf.up[a.Right] = append(sf.up[a.Right], upRef{n, a.Rule, idx})
			}
			if a.Left != 0 {
				idx := 0
				if f.Nodes[a.Left].Sym < 0 {
					idx = -1
				}
				sf.up[a.Left] = append(sf.up[a.Left], upRef{n, a.Rule, idx})
			}
			for _, c := range [2]int32{a.Left, a.Right} {
				if c != 0 && !seen[c] {
					seen[c] = true
					stack = append(stack, c)
				}
			}
		}
	}
}

// owners is the rule nodes node n is a symbol of, through rule positions, with its index there.
func (sf *subjectFacts) owners(n int32) []upRef {
	var out []upRef
	var climb func(p int32, rule int32, idx int)
	climb = func(p, rule int32, idx int) {
		if sf.f.Nodes[p].Sym >= 0 {
			out = append(out, upRef{p, rule, idx})
			return
		}
		for _, u := range sf.up[p] {
			if u.idx == -1 {
				climb(u.p, u.rule, idx)
			}
		}
	}
	for _, u := range sf.up[n] {
		if u.idx >= 0 {
			climb(u.p, u.rule, u.idx)
		}
	}
	return out
}

// context is what op.c makes of the dereference node n of sigil, read through the lists, parentheses, ?: branches
// and unit rules around it to the operator that takes it, ctxUnknown where readings disagree or perl was not measured:
//   - a subscript after it (@{...}[...], ${...}{...}) or \, ++ and --, local, an assignment's left side, a sub's or a
//     method's arguments (entersub's op_lvalue), foreach's list, keys, values, each, chomp and chop, and push, unshift,
//     splice, pop and shift's first argument, grep's, map's and sort's list, and sprintf's after its format: modified
//     (Concise: push @{$r->{a}}, 1 keeps an rv2hv under helem sKM/DREFAV);
//   - !, an assignment's right side, + - * / ** . x, a comparison, || && //, an if, unless, elsif or while condition,
//     return, print, join and reverse: read (my @x = @{$r->{a}} is multideref($r->{"a"}));
//   - scalar and a ?: branch: whatever their own context is.
func (sf *subjectFacts) context(n int32, sigil byte) int {
	ctx := -1
	merge := func(c int) {
		if ctx != -1 && ctx != c {
			c = ctxUnknown
		}
		ctx = c
	}
	var visit func(n int32, first bool)
	visit = func(n int32, first bool) {
		us := sf.owners(n)
		if len(us) == 0 {
			merge(ctxUnknown)
		}
		for _, u := range us {
			lhs, rhs, i := sf.g.lhs[u.rule], sf.g.rhs[u.rule], u.idx
			joined := strings.Join(rhs, " ")
			switch {
			case i == 0 && len(rhs) > 1 && (lhs == "subscripted" || lhs == "term") &&
				(rhs[1] == "PERLY_BRACE_OPEN" || rhs[1] == "PERLY_BRACKET_OPEN"):
				if sigil == '%' {
					merge(ctxUnknown) // %{...}{...} and %{...}[...], unmeasured
				} else {
					merge(ctxAggregate)
				}
			case len(rhs) == 1 && slices.Contains([]string{"term", "listexpr", "expr", "optlistexpr", "optexpr", "mexpr",
				"texpr", "condition", "sliceme", "kvslice", "termbinop", "termunop", "listop"}, lhs):
				visit(u.p, first)
			case joined == "listexpr PERLY_COMMA term":
				visit(u.p, first && i == 0)
			case joined == "listexpr PERLY_COMMA", joined == "PERLY_PAREN_OPEN expr PERLY_PAREN_CLOSE":
				visit(u.p, first)
			case joined == "term PERLY_QUESTION_MARK term PERLY_COLON term" && i > 0:
				visit(u.p, first)
			case sf.text(int(sf.f.Nodes[u.p].From)) == "scalar" && slices.Contains([]string{"UNIOP", "FUNC1"}, rhs[0]):
				// op_lvalue passes its context into ?:'s branches and, as entersub's, into scalar's operand (op.c's
				// Perl_op_lvalue_flags, case OP_SCALAR; Concise: is(scalar(@{$c->{a}}), 1) keeps an rv2hv,
				// my $n = scalar(@{$c->{a}}) is multideref($c->{"a"}))
				visit(u.p, true)
			default:
				merge(sf.ownerContext(u, n, first))
			}
		}
	}
	visit(n, true)
	if ctx == -1 {
		return ctxUnknown
	}
	return ctx
}

// ownerContext is what the rule u that takes node n makes of it, as context reads it; first is whether n is the
// first item of the list u takes.
func (sf *subjectFacts) ownerContext(u upRef, n int32, first bool) int {
	lhs, rhs, i := sf.g.lhs[u.rule], sf.g.rhs[u.rule], u.idx
	word := sf.text(int(sf.f.Nodes[u.p].From))
	sym := func(j int) string {
		if j < len(rhs) {
			return rhs[j]
		}
		return ""
	}
	switch {
	case lhs == "termbinop" && sym(1) == "ASSIGNOP":
		if i == 0 {
			return ctxAggregate
		}
		if op := sf.text(int(sf.f.Nodes[n].From) - 1); op == "=" || op == "+=" {
			return ctxRead
		}
	case lhs == "termbinop" && slices.Contains([]string{"ADDOP", "MULOP", "POWOP", "OROR", "ANDAND", "DORDOR"}, sym(1)),
		lhs == "termeqop", lhs == "termrelop", lhs == "termunop" && sym(0) == "PERLY_EXCLAMATION_MARK":
		return ctxRead
	case lhs == "termunop" && sym(0) != "PERLY_MINUS" && sym(0) != "PERLY_PLUS" && sym(0) != "PERLY_TILDE",
		lhs == "term" && (sym(0) == "REFGEN" || sym(0) == "KW_LOCAL"):
		return ctxAggregate // ++, --, \ and local
	case slices.Contains([]string{"NOAMP", "UNIOPSUB", "LSTOPSUB", "amper", "METHCALL", "METHCALL0"}, sym(0)),
		len(rhs) > 2 && sym(1) == "ARROW" && sym(2) == "methodname":
		return ctxAggregate // a sub's arguments: entersub's op_lvalue
	case strings.HasPrefix(lhs, "bare_statement_for") && !slices.Contains(rhs, "PERLY_SEMICOLON"),
		lhs == "sideff" && sym(1) == "KW_FOR" && i == 2:
		return ctxAggregate // foreach aliases its list
	case lhs == "bare_statement_if", lhs == "bare_statement_unless", lhs == "bare_statement_while",
		lhs == "else" && sym(0) == "KW_ELSIF",
		lhs == "sideff" && i == 2 && slices.Contains([]string{"KW_IF", "KW_UNLESS", "KW_WHILE"}, sym(1)):
		return ctxRead
	case slices.Contains([]string{"LSTOP", "FUNC", "UNIOP", "FUNC1", "BLKLSTOP", "LOOPEX"}, sym(0)):
		block := len(rhs) > 1 && (sym(1) == "indirob" || sym(1) == "block")
		switch word {
		case "keys", "values", "each", "chomp", "chop":
			return ctxAggregate
		case "push", "unshift", "splice", "pop", "shift":
			if first {
				return ctxAggregate
			}
		case "grep", "map", "sort":
			if word == "sort" && block {
				if sf.sortHedge(int(sf.f.Nodes[u.p].From) + 1) {
					return ctxUnknown
				}
			}
			if block || !first {
				return ctxAggregate
			}
		case "sprintf":
			if !first {
				return ctxAggregate
			}
		case "print", "join", "reverse", "return":
			return ctxRead
		}
	}
	return ctxUnknown
}

// sortHedge is whether the sort block opening at token k holds a <=> or cmp, at any depth, so the chart hedges it:
// op.c's S_simplify_sort drops a block whose one op, after constant folding, is <=> or cmp over the current stash's
// $a and $b (Concise: sort { $a <=> $b }, { +($a <=> $b) } and { 1 ? $a <=> $b : 0 } over @{$r->{a}} are each
// multideref($r->{"a"})), which neither spelling nor folding nor lexicals let the chart settle; a block holding no
// comparison is kept, as perl keeps it.
func (sf *subjectFacts) sortHedge(k int) bool {
	if k >= len(sf.s.Tokens) || sf.text(k) != "{" {
		return false
	}
	for j, depth := k+1, 1; j < len(sf.s.Tokens) && depth > 0; j++ {
		switch sf.text(j) {
		case "{":
			depth++
		case "}":
			depth--
		case "<=>", "cmp":
			return true
		}
	}
	return false
}

// owner is the word of the operator that takes node n, through lists and parentheses, or "".
func (sf *subjectFacts) owner(n int32) string {
	for _, u := range sf.owners(n) {
		switch rhs := sf.g.rhs[u.rule]; {
		case len(rhs) == 1, strings.Join(rhs, " ") == "listexpr PERLY_COMMA term":
			if w := sf.owner(u.p); w != "" {
				return w
			}
		case slices.Contains([]string{"LSTOP", "FUNC", "UNIOP", "FUNC1"}, rhs[0]):
			return sf.text(int(sf.f.Nodes[u.p].From))
		}
	}
	return ""
}

// refaliased is whether hash node n is what a refalias assigns, under a \ (through parentheses, lists and local) that
// is an assignment's left side or foreach's alias: perl builds no rv2hv for it (Concise: `\%h = {}`, `(\%h) = ...`,
// `\local %h = ...` and `for \%h (...)` are refassign or an iterator over gv(*h); perl.git op/lvref.t).
func (sf *subjectFacts) refaliased(n int32) bool {
	for _, u := range sf.owners(n) {
		switch lhs, joined := sf.g.lhs[u.rule], strings.Join(sf.g.rhs[u.rule], " "); {
		case joined == "REFGEN term":
			if sf.assigned(u.p) {
				return true
			}
		case strings.HasPrefix(lhs, "bare_statement_for") && (strings.Contains(joined, "REFGEN") || strings.Contains(joined, "my_refgen")):
			return true
		case len(sf.g.rhs[u.rule]) == 1, joined == "PERLY_PAREN_OPEN expr PERLY_PAREN_CLOSE",
			joined == "listexpr PERLY_COMMA term", joined == "KW_LOCAL term", lhs == "myattrterm":
			if sf.refaliased(u.p) {
				return true
			}
		}
	}
	return false
}

// assigned is whether node n is an assignment's left side, or an item of one.
func (sf *subjectFacts) assigned(n int32) bool {
	for _, u := range sf.owners(n) {
		switch rhs := sf.g.rhs[u.rule]; {
		case len(rhs) == 1, slices.Contains([]string{"PERLY_PAREN_OPEN expr PERLY_PAREN_CLOSE", "listexpr PERLY_COMMA term",
			"KW_LOCAL term"}, strings.Join(rhs, " ")):
			if sf.assigned(u.p) { // through parentheses, a list's items and local: (\%h, $x) = ... and local \%h = ...
				return true
			}
		case sf.g.lhs[u.rule] == "termbinop" && rhs[1] == "ASSIGNOP" && u.idx == 0:
			return true
		}
	}
	return false
}
