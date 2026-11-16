// ABOUTME: The chart as a parseoracle subject: SubjectFacts read from the forest forest.Parse leaves, in-process.
// ABOUTME: A site every reading holds is decided; one only some readings hold is Unresolved, never a guess.

package forest

import (
	"slices"
	"strings"
	"unicode"

	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/chart/lex"
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
	// kind is each rule's site kind, -1 for none; last is whether its site is at the node's last token
	// (a subscript's }, a cast) rather than its first, so nested subscripts sharing a start stay apart.
	kind     []int
	last     []bool
	matchop  int    // the rule term MATCHOP term, whose site is at the operator
	fullstmt int    // fullstmt's symbol id
	pmfunc   int    // PMFUNC's
	notCore  []bool // the terminals a word takes as a sub or method rather than a builtin
}

func buildFactsGrammar(tab *grammar.Tables) *factsGrammar {
	id := func(name string) int {
		return slices.IndexFunc(tab.Symbols, func(s grammar.TableSymbol) bool { return s.Name == name })
	}
	g := &factsGrammar{kind: make([]int, len(tab.Rules)), last: make([]bool, len(tab.Rules)), matchop: -1,
		fullstmt: id("fullstmt"), pmfunc: id("PMFUNC"), notCore: make([]bool, len(tab.Symbols))}
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
		case len(rhs) > 0 && slices.Contains([]string{"NOAMP", "UNIOPSUB", "FUNC0SUB", "LSTOPSUB"}, rhs[0]),
			lhs == "term" && len(rhs) > 0 && rhs[0] == "amper": // f(...), which toke.c lexes as &f(...); altSite drops a written &
			g.kind[r] = kindCall
		case lhs == "term" && joined == "REFGEN term":
			g.kind[r] = kindReference
		case lhs == "anonymous" && len(rhs) > 0 && rhs[0] == "HASHBRACK":
			g.kind[r] = kindAnonhash
		case lhs == "termbinop" && joined == "term MATCHOP term":
			g.kind[r], g.matchop = kindMatch, r
		case lhs == "hsh", joined == "term ARROW CAST_PERCENT_SIGN", lhs == "kvslice" && joined == "term ARROW PERLY_PERCENT_SIGN":
			g.kind[r], g.last[r] = kindHash, true
		case lhs == "subscripted" || lhs == "term":
			// a brace subscript reads a hash: an element, an arrow's, a chained one, or a slice of one
			if i := slices.Index(rhs, "PERLY_BRACE_OPEN"); i > 0 && slices.Contains([]string{"scalar", "ARROW", "subscripted", "sliceme"}, rhs[i-1]) {
				g.kind[r], g.last[r] = kindHash, true
			}
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
	uncertain map[int]bool // the sites some readings hold and others do not
	all       map[int]bool // every site any reading holds
	stmts     [][2]int     // the statements' token spans, top level only, sorted
}

// SubjectFacts is the chart's answer to parseoracle's subject contract for src, a file at path: whether
// forest.Parse accepts it under scope, the prototypes resolve saw, and a site for each decision at an oracle
// marker in the readings forest.Parse leaves, on its statement's lines as perl numbers them. A site some
// readings hold and others do not is Unresolved. Neither list is ever nil.
func SubjectFacts(tab *grammar.Tables, src []byte, path string, scope Scope) parseoracle.SubjectFacts {
	facts := parseoracle.SubjectFacts{Prototypes: map[string]string{}, CallSites: []parseoracle.SubjectCallSite{}}
	f, s, _ := parseStream(tab, src, scope)
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
		uncertain: map[int]bool{}, all: map[int]bool{}}
	sf.walk()

	keys := make([]int, 0, len(sf.all))
	for k := range sf.all {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	line, own := parse.PerlLines(src, path)
	starts := lineStarts(src)
	for _, k := range keys {
		kind, at := k%nkinds, k/nkinds
		st, ok := sf.statement(at)
		if !ok {
			continue
		}
		first := lineAt(starts, s.Raw[s.Tokens[st[0]].Index].Start)
		last := lineAt(starts, s.Raw[s.Tokens[st[1]-1].Index].End-1)
		if !own[first] || !own[last] {
			continue // a #line directive moved it into another file
		}
		cs := parseoracle.SubjectCallSite{Kind: kindName[kind], Line: line[first], EndLine: line[last], Unresolved: sf.uncertain[k]}
		if kind == kindCall {
			cs.Name = r.text(at)
			if d, st := r.lookup(cs.Name, at); st != declared || strings.ContainsAny(d.proto, `\*+`) {
				// a word nothing declares may still be a sub a module resolve cannot see gave a \ prototype.
				// shortcut: a call whose prototype may take an argument by reference is Unresolved rather than
				// fitted slot by slot, which grades WIDER where perl took one; upgrade when ORACLEGRADE shows the
				// WIDER sites cost exact ones recursive descent has.
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
	return facts
}

// bareProto is a prototype as prototype() returns it: without its parens or blanks.
func bareProto(p string) string {
	return strings.ReplaceAll(strings.TrimSuffix(strings.TrimPrefix(p, "("), ")"), " ", "")
}

// walk visits every node a derivation of the root reaches: it records the statements, every site in all, and,
// at each node with more than one alternative, the sites its alternatives disagree on in uncertain.
// shortcut: code interpolated in a string or heredoc is one token the forest does not read, so a marker inside it
// reports nothing, which grades WRONG where perl built one; hedging it needs parse's unexported interpolatesCode,
// pending perigrin's export ruling, and WRONGZERO (01a1c324-1fd7) owns it.
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
			spans = append(spans, [2]int{int(nd.From), min(int(nd.To), len(sf.s.Tokens))})
		}
		if k, ok := sf.leafSite(n); ok {
			sf.all[k] = true
		}
		alts := 0
		for x := nd.Alt; x != 0; x = f.Alts[x].Next {
			alts++
			a := f.Alts[x]
			if k, ok := sf.altSite(n, a); ok {
				sf.all[k] = true
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
		if k, has := sf.altSite(n, a); has {
			am, ay = union(am, []int{k}), union(ay, []int{k})
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

// altSite is the site alternative a of node n stands for, if any: its rule's, at the token the rule anchors it to.
func (sf *subjectFacts) altSite(n int32, a Alt) (int, bool) {
	nd := sf.f.Nodes[n]
	if nd.Sym < 0 || sf.g.kind[a.Rule] < 0 {
		return 0, false // a rule position is no site; its rule's node is
	}
	at := int(nd.From)
	switch {
	case sf.g.kind[a.Rule] == kindCall && sf.s.Raw[sf.s.Tokens[at].Index].Kind != lexer.Word:
		return 0, false // &f(...) bypasses the prototype, so it decides nothing
	case int(a.Rule) == sf.g.matchop:
		right := sf.f.Nodes[a.Right]
		if right.To == right.From+1 && quoteOp(sf.text(int(right.From))) != "" {
			return 0, false // a pattern operand is its own site, or none for s/// and tr///
		}
		at = int(right.From) - 1
	case sf.g.last[a.Rule]:
		at = int(nd.To) - 1
	}
	return site(sf.g.kind[a.Rule], at), true
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
	case t.Kind == lexer.Readline && readsLine(text):
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

// readsLine is whether a <...> token is readline rather than glob, as toke.c's scan_inputsymbol decides: empty,
// <<>>, or a bare handle or simple scalar inside, its name in runes, since use utf8 widens it (<FÉ>).
// shortcut: this mirrors parse's unexported isReadlineTerm; replace it with that predicate if perigrin rules it
// exported.
func readsLine(text string) bool {
	in := strings.TrimSuffix(strings.TrimPrefix(text, "<"), ">")
	if in == "" || in == "<>" {
		return true
	}
	in = strings.TrimPrefix(in, "$")
	if in == "" || in[0] >= '0' && in[0] <= '9' {
		return false
	}
	return strings.IndexFunc(in, func(c rune) bool {
		return c != '_' && c != ':' && !unicode.IsLetter(c) && !unicode.IsDigit(c)
	}) < 0
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
