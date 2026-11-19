// ABOUTME: The chart as a parseoracle subject: SubjectFacts read from the forest forest.Parse leaves, in-process.
// ABOUTME: A site every reading holds is decided; one only some readings hold is Unresolved, never a guess.

package forest

import (
	"regexp"
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
	deref    []uint8      // where a dereference's operand stands, for derefHedge
	matchop  int          // the rule term MATCHOP term, whose site is at the operator
	fullstmt int          // fullstmt's symbol id
	pmfunc   int          // PMFUNC's
	notCore  []bool       // the terminals a word takes as a sub or method rather than a builtin
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

// The operands derefHedge reads: a cast's indirob, the term before a postfix ->, and local's or \'s.
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
		level: make([]levelShape, len(tab.Rules)), deref: make([]uint8, len(tab.Rules)), matchop: -1,
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
		case len(rhs) == 3 && rhs[0] == "term" && rhs[1] == "ARROW" &&
			slices.Contains([]string{"CAST_PERCENT_SIGN", "CAST_SNAIL", "CAST_STAR", "DOLSHARP", "PERLY_SNAIL", "PERLY_PERCENT_SIGN"}, rhs[2]):
			g.deref[r] = derefPost // ->%*, ->@*, ->$* (and ->**, one rule with it), ->$#*, ->@{...} and ->%{...}
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
			if k, ok := sf.derefHedge(n, a); ok {
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
		if k, has := sf.derefHedge(n, a); has {
			ay = union(ay, []int{k})
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
	switch sf.g.hash[a.Rule] {
	case hashNamed:
		return sf.named(int(nd.From))
	case hashSlice:
		if base := sf.f.Nodes[sf.child(a, 3)]; base.To == base.From+1 && len(sf.text(int(base.From))) > 1 {
			return sf.named(int(base.From)) // @h{...}; @$r{...} and ->@{...} read through a reference
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
// constant, a call to a constant sub say, is unsure. derefHedge covers a chain read as an aggregate.
// shortcut: after a range whose constant fold perl abandons (`"*x".."az"`, not `"*x"..3`), perl folds nothing more of
// the main program's statements into multideref, so a read there keeps its rv2hv where the chart reports none; perl.git
// op/range.t:301 and :446 grade WRONG (TestOracleSubjectHashSites); upgrade when the cause in peep.c is known.
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

// derefHedge is a hedge at a dereference (its sigil, its ->, local or \) whose operand holds an element chain that
// reads a hash through a reference or a package hash, the chain itself or one a ?: passes on (op/ref.t:928):
// S_maybe_multideref leaves the last level out where op_lvalue marked it OPpDEREF_AV or _HV, which turns on what is
// done with the dereference (Concise: `foreach (@{$r->{a}})` keeps an rv2hv under helem sKM/DREFAV, `my @x =
// @{$r->{a}}` is multideref($r->{"a"})), and on local and refassign (`delete local $r->{a}` and op/lvref.t's
// `\$h{a} = ...` keep one); the chart does not read that context.
func (sf *subjectFacts) derefHedge(n int32, a Alt) (int, bool) {
	nd := sf.f.Nodes[n]
	if nd.Sym < 0 || sf.g.deref[a.Rule] == 0 {
		return 0, false
	}
	at, op := int(nd.From), sf.f.Nodes[a.Right]
	switch sf.g.deref[a.Rule] {
	case derefCast:
		if sf.text(int(op.From)) == "{" && op.To > op.From+1 {
			op.From, op.To = op.From+1, op.To-1
		}
	case derefPost:
		op = sf.f.Nodes[sf.child(a, 2)]
		at = int(op.To)
	}
	i, _ := slices.BinarySearchFunc(sf.levels, op.From, func(l int32, from int32) int { return int(sf.f.Nodes[l].From - from) })
	for ; i < len(sf.levels) && sf.f.Nodes[sf.levels[i]].From < op.To; i++ {
		if sf.f.Nodes[sf.levels[i]].To <= op.To && sf.reads(sf.levels[i]) {
			return site(kindHash, at), true
		}
	}
	return 0, false
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
			if t := sf.text(k); interpolatesVariable(t) || strings.HasPrefix(t, "qw") && len(strings.Fields(t[3:len(t)-1])) > 1 {
				return keyComputed // qw/a b/ is a list, joined with $; as a , is
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

// interpolatesVariable is whether quote t interpolates a variable: an unescaped $ or @ before a name, a { or a :.
func interpolatesVariable(t string) bool {
	if strings.HasPrefix(t, "'") || len(t) > 1 && t[0] == 'q' && !strings.ContainsRune("qwrx", rune(t[1])) {
		return false
	}
	for i := 0; i+1 < len(t); i++ {
		switch t[i] {
		case '\\':
			i++
		case '$', '@':
			if c := t[i+1]; isWordByte(c) || c == '{' || c == ':' {
				return true
			}
		}
	}
	return false
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
}

type hashDecl struct{ from, to, kind int }

// at is what the hash a token's text names is at token k: the declaration its own token makes, else the innermost
// in scope, else a package hash. A qualified name is always a package's.
// shortcut: a lexical %h a module's import installs (builtin::export_lexically) is not seen, so its uses read as the
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
	return kind
}

// hashScopes reads the file's hash declarations: my, state and field hashes, lexical from the end of their statement
// (perly.y's intro_my) to their block's end, or to the end of the if, unless, while, until or for whose condition
// declares them; our's, the package hash, over the same span; and a signature's, over its sub's body.
func (r *resolver) hashScopes() *hashScopes {
	h := &hashScopes{decls: map[string][]hashDecl{}, declared: map[int]int{}}
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
		vars := []int{k + 1}
		if k+2 < r.n && r.isWord(k+1) && !r.is(k+2, "(") { // my Class %h
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
		from, to := r.statementEnd(r.s.Raw[r.s.Tokens[k].Index].Start), r.scopeEnd(k)
		at, _ := slices.BinarySearchFunc(r.s.Tokens, from, func(t lex.Token, b int) int { return r.s.Raw[t.Index].Start - b })
		for _, j := range vars {
			if j >= r.n {
				continue
			}
			if t := r.text(j); strings.HasPrefix(t, "%") && hashName(t) != "" && !strings.Contains(hashName(t), "::") {
				h.declared[j] = kind
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
	text := r.text(j)
	if r.is(j, "(") {
		for d := 0; j < r.n; j++ {
			if r.is(j, "(") {
				d++
			} else if r.is(j, ")") {
				if d--; d == 0 {
					break
				}
			}
			text += " " + r.text(j)
		}
	} else if r.s.Raw[r.s.Tokens[j].Index].Kind != lexer.Prototype {
		return
	}
	body := j
	for body < r.n && !r.is(body, "{") {
		body++
	}
	if body >= r.n || r.match[body] < 0 {
		return
	}
	for _, m := range sigHash.FindAllStringSubmatch(text, -1) {
		h.decls[m[1]] = append(h.decls[m[1]], hashDecl{j, r.match[body], kind})
	}
}

// sigHash is a signature's hash parameter.
var sigHash = regexp.MustCompile(`%\s*(\w+)`)

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
