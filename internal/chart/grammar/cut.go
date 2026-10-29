// ABOUTME: The token cut: rewrites a read SLIF grammar so its terminals are pvm token predicates.
// ABOUTME: Collapses sub-token rules, splits straddling ones, carries lexeme priorities, and lists every rewrite.

package grammar

import (
	"fmt"
	"regexp"
	"slices"
	"sort"
	"strings"

	"tamarou.com/pvm/internal/lexer"
)

// Pred is a token predicate: a lexer kind, and the text a token of that
// kind must have. The text is the language of Def, a SLIF right-hand side,
// less any text Except matches.
type Pred struct {
	Kind   lexer.Kind
	Def    string
	Except string

	text, except *regexp.Regexp
}

// Match reports whether a token of kind k and text text satisfies p.
func (p Pred) Match(k lexer.Kind, text string) bool {
	return k == p.Kind && p.text.MatchString(text) && (p.except == nil || !p.except.MatchString(text))
}

// String is the predicate as the rewrite list writes it: `Operator '->'`
// for a single literal, `Variable ~ SigilScalar VarIdentExpr` otherwise.
func (p Pred) String() string {
	s := p.Kind.String() + " ~ " + p.Def
	if _, ok := literalText(p.Def); ok {
		s = p.Kind.String() + " " + p.Def
	}
	if p.Except != "" {
		s += " except " + p.Except
	}
	return s
}

// TokenDef is one entry of a cut: the symbol Name becomes a terminal, or
// gains an alternative, whose predicate has lexer kind Kind (spelled as
// lexer.Kind.String spells it) and text Def, a SLIF right-hand side. An empty
// Def is the symbol's own rules. A TokenDef with no Kind defines text only,
// for other Defs to name. Except, if set, is a right-hand side whose text the
// predicate refuses; ":priority" there is every word a priority lexeme
// matches, which is how keyword priority is carried.
type TokenDef struct {
	Name, Kind, Def, Except string
}

// Split replaces the symbol sequence From, wherever it appears in a G1
// right-hand side, by the sequence To.
type Split struct {
	From, To string
}

// PriorityNote says how a lexeme priority that no "except :priority"
// carries is accounted for.
type PriorityNote struct {
	Symbol, How string
}

// TriviaDef names a lexer kind the recognizer skips, and the :discard it
// stands for, if any.
type TriviaDef struct {
	Kind, Discard string
}

// CutSpec is the reviewed table a cut applies.
type CutSpec struct {
	Tokens     []TokenDef
	Splits     []Split
	Priorities []PriorityNote
	Trivia     []TriviaDef
}

// GuacamoleCut is the reviewed cut of Guacamole's grammar. Every G0 lexeme
// G1 names whose alternatives are single literals collapses without an entry
// here; testdata/token-cut-rewrites.txt lists what this table does.
var GuacamoleCut = CutSpec{
	Tokens: []TokenDef{
		// A sigil and its name are one Variable token.
		{Name: "ScalarVar", Kind: "Variable", Def: "SigilScalar VarIdentExpr"},
		{Name: "ArrayVar", Kind: "Variable", Def: "SigilArray VarIdentExpr"},
		{Name: "HashVar", Kind: "Variable", Def: "SigilHash VarIdentExpr"},
		{Name: "ArrayTopVar", Kind: "Variable", Def: "SigilArrayTop VarIdentExpr"},
		{Name: "SpecialScalarVar", Kind: "Variable", Def: "SigilScalar GlobalVariables"},
		{Name: "SpecialArrayVar", Kind: "Variable", Def: "SigilArray GlobalVariables"},
		{Name: "SpecialHashVar", Kind: "Variable", Def: "SigilHash GlobalVariables"},
		{Name: "DollarHashVar", Kind: "Variable", Def: "'$#'"},
		{Name: "ArgsTopVar", Kind: "Variable", Def: "'$#_'"},
		// & and * are tokens of their own before a name: FuncSigil, and the
		// Operator * that pvm gives a glob.
		{Name: "CodeSigil", Kind: "FuncSigil", Def: "SigilCode"},
		{Name: "CodeName", Kind: "Word", Def: "VarIdentExpr"},
		{Name: "GlobSigil", Kind: "Operator", Def: "SigilGlob"},
		{Name: "GlobName", Kind: "Word", Def: "VarIdentExpr"},
		{Name: "SpecialGlobName", Kind: "Word", Def: "GlobalVariables"},
		{Name: "SpecialGlobName", Kind: "Number", Def: "GlobalVariables"},
		// A sigil before a block is a DerefSigil token. pvm lexes $#{...} as
		// one Variable whatever the block holds, so only blocks the token's
		// text can show Guacamole derives are admitted: a lone scalar.
		{Name: "DerefScalarSigil", Kind: "DerefSigil", Def: "SigilScalar"},
		{Name: "DerefArraySigil", Kind: "DerefSigil", Def: "SigilArray"},
		{Name: "DerefHashSigil", Kind: "DerefSigil", Def: "SigilHash"},
		{Name: "ArrayTopBlockVar", Kind: "Variable", Def: "SigilArrayTop '{' ScalarVar '}'"},
		{Name: "ArrayTopBlockVar", Kind: "Variable", Def: "SigilArrayTop '{' SpecialScalarVar '}'"},
		// After ->, a slice's sigil is a Variable token and its bracket another.
		{Name: "SliceArraySigil", Kind: "Variable", Def: "SigilArray"},
		{Name: "SliceHashSigil", Kind: "Variable", Def: "SigilHash"},
		// A name, qualified or not, is one Word. Keyword priority: an
		// unqualified keyword is never a call's name, and a phase name never a
		// defined sub's.
		{Name: "SubNameExpr", Kind: "Word"},
		{Name: "SubNameCallExpr", Kind: "Word", Except: ":priority"},
		{Name: "SubDefName", Kind: "Word", Def: "SubNameExpr", Except: "PhaseName"},
		{Name: "IdentComp", Kind: "Word"},
		// A string or quote-like is one Quote token, and <FH> one Readline.
		{Name: "LitString", Kind: "Quote"},
		{Name: "InterpolString", Kind: "Quote"},
		{Name: "QLikeValueExpr", Kind: "Quote"},
		{Name: "QLikeValueExprWithMods", Kind: "Quote"},
		{Name: "Diamond", Kind: "Readline", Def: "'<' ScalarVar '>'"},
		{Name: "Diamond", Kind: "Readline", Def: "'<' BuiltinFilehandle '>'"},
		{Name: "Diamond", Kind: "Readline", Def: "'<>'"},
		{Name: "DoubleDiamond", Kind: "Readline", Def: "'<<' ScalarVar '>>'"},
		{Name: "DoubleDiamond", Kind: "Readline", Def: "'<<' BuiltinFilehandle '>>'"},
		{Name: "DoubleDiamond", Kind: "Readline", Def: "'<<>>'"},
		// ->@* and the like are one Variable token, as is $_.
		{Name: "DerefVariableArgsAll", Kind: "Variable"},
		{Name: "VarDefaultArg", Kind: "Variable"},
		// Numbers. A version is a Number (5.010), a v-string Quote (v5.36,
		// 1.2.3) or, with one segment, a Word (v5).
		{Name: "NumberDec", Kind: "Number"},
		{Name: "NumberOct", Kind: "Number"},
		{Name: "NumberHex", Kind: "Number"},
		{Name: "NumberBin", Kind: "Number"},
		{Name: "VersionNumber", Kind: "Number"},
		{Name: "VersionNumber", Kind: "Quote"},
		{Name: "VersionNumber", Kind: "Word"},
		// pvm lexes the (...) of :prototype(...) as one Prototype token; other
		// attributes' arguments it lexes as code, which no predicate admits.
		{Name: "SubAttrArgs", Kind: "Prototype"},
	},
	Splits: []Split{
		{From: "SigilScalar VarIdentExpr", To: "ScalarVar"},
		{From: "SigilArray VarIdentExpr", To: "ArrayVar"},
		{From: "SigilHash VarIdentExpr", To: "HashVar"},
		{From: "SigilArrayTop VarIdentExpr", To: "ArrayTopVar"},
		{From: "SigilCode VarIdentExpr", To: "CodeSigil CodeName"},
		{From: "SigilGlob VarIdentExpr", To: "GlobSigil GlobName"},
		{From: "SigilScalar GlobalVariables", To: "SpecialScalarVar"},
		{From: "SigilArray GlobalVariables", To: "SpecialArrayVar"},
		{From: "SigilHash GlobalVariables", To: "SpecialHashVar"},
		{From: "SigilGlob GlobalVariables", To: "GlobSigil SpecialGlobName"},
		{From: "SigilScalar BlockNonEmpty", To: "DerefScalarSigil BlockNonEmpty"},
		{From: "SigilArray BlockNonEmpty", To: "DerefArraySigil BlockNonEmpty"},
		{From: "SigilHash BlockNonEmpty", To: "DerefHashSigil BlockNonEmpty"},
		{From: "SigilGlob BlockNonEmpty", To: "GlobSigil BlockNonEmpty"},
		{From: "SigilArrayTop BlockNonEmpty", To: "ArrayTopBlockVar"},
		{From: "'$#'", To: "DollarHashVar"},
		{From: "'$#_'", To: "ArgsTopVar"},
		{From: "'@['", To: "SliceArraySigil LBracket"},
		{From: "'@{'", To: "SliceArraySigil LBrace"},
		{From: "'%['", To: "SliceHashSigil LBracket"},
		{From: "'%{'", To: "SliceHashSigil LBrace"},
		{From: "']'", To: "RBracket"},
		{From: "'}'", To: "RBrace"},
		{From: "OpKeywordSub SubNameExpr", To: "OpKeywordSub SubDefName"},
	},
	Priorities: []PriorityNote{
		{Symbol: "QLikeValueExpr", How: "settled by the lexer: a quote-like is one Quote token, never a Word"},
		{Symbol: "QLikeValueExprWithMods", How: "settled by the lexer: a quote-like is one Quote token, never a Word"},
		{Symbol: "VersionNumber", How: "not encoded: after use or no and a class name, and after require, a number that is a version also reads as an expression"},
		{Symbol: "OpAdd", How: "not encoded: where OpUnary is also acceptable, as after sort $x, + and - read both ways"},
	},
	Trivia: []TriviaDef{
		{Kind: "Whitespace", Discard: "whitespace"},
		{Kind: "Comment", Discard: "hash comment"},
		{Kind: "Pod"},
		{Kind: "DataSection"},
		{Kind: "FormatBody"},
	},
}

// CutGrammar is a grammar after the cut: G1 rules whose every right-hand
// symbol is a G1 nonterminal or a terminal, the terminals' predicates, the
// trivia kinds, and every rewrite the cut made, one line each.
type CutGrammar struct {
	G1        []Rule
	Terminals map[string][]Pred
	Trivia    []lexer.Kind
	Rewrites  []string
}

// ApplyCut applies spec to g. It fails, naming the rule, on anything the
// spec leaves unaccounted for: a G0 symbol or literal still in a G1 rule, an
// unknown lexer kind, a priority or :discard nobody carries.
func ApplyCut(g *Grammar, spec CutSpec) (*CutGrammar, error) {
	c := &cutter{
		g: g, out: &CutGrammar{Terminals: map[string][]Pred{}},
		g0: byLHS(g.G0), g1: byLHS(g.G1), defs: map[string][]TokenDef{},
		re: map[string]string{}, busy: map[string]bool{},
	}
	for _, d := range spec.Tokens {
		if c.defs[d.Name] == nil {
			c.order = append(c.order, d.Name)
		}
		c.defs[d.Name] = append(c.defs[d.Name], d)
	}
	steps := []func(CutSpec) error{c.split, c.collapse, c.prune, c.literals, c.checkLeftovers, c.priorities, c.excepts, c.trivia}
	for _, step := range steps {
		if err := step(spec); err != nil {
			return nil, err
		}
	}
	return c.out, nil
}

type cutter struct {
	g      *Grammar
	out    *CutGrammar
	g0, g1 map[string][]Rule
	defs   map[string][]TokenDef
	order  []string // TokenDef names, in spec order
	rules  []Rule   // G1 rules being cut

	re   map[string]string // text regexp per symbol, unanchored
	busy map[string]bool

	priorityWords []string // regexps of the priority lexemes "except :priority" refuses
}

// kindNamed maps each lexer kind's name, as lexer.Kind.String spells it, to
// the kind.
var kindNamed = func() map[string]lexer.Kind {
	m := map[string]lexer.Kind{}
	for k := lexer.Kind(0); k.String() != "Kind(?)"; k++ {
		m[k.String()] = k
	}
	return m
}()

func byLHS(rules []Rule) map[string][]Rule {
	m := map[string][]Rule{}
	for _, r := range rules {
		m[r.LHS] = append(m[r.LHS], r)
	}
	return m
}

func (c *cutter) rewrite(format string, args ...any) {
	c.out.Rewrites = append(c.out.Rewrites, fmt.Sprintf(format, args...))
}

func cutErrorf(format string, args ...any) error {
	return fmt.Errorf("token cut: "+format, args...)
}

// split applies the splits to every G1 right-hand side.
func (c *cutter) split(spec CutSpec) error {
	type seqs struct{ from, to []Symbol }
	var splits []seqs
	for _, s := range spec.Splits {
		from, err := parseSymbols(s.From)
		if err != nil {
			return cutErrorf("split %q: %v", s.From, err)
		}
		to, err := parseSymbols(s.To)
		if err != nil {
			return cutErrorf("split %q: %v", s.To, err)
		}
		splits = append(splits, seqs{from, to})
	}
	for _, r := range c.g.G1 {
		var rhs []Symbol
	scan:
		for i := 0; i < len(r.RHS); {
			for _, s := range splits {
				if i+len(s.from) <= len(r.RHS) && slices.Equal(r.RHS[i:i+len(s.from)], s.from) {
					rhs = append(rhs, s.to...)
					i += len(s.from)
					continue scan
				}
			}
			rhs = append(rhs, r.RHS[i])
			i++
		}
		if !slices.Equal(rhs, r.RHS) {
			after := r
			after.RHS = rhs
			c.rewrite("split %s => %s", ruleText(r, "::="), ruleText(after, "::="))
			r = after
		}
		c.rules = append(c.rules, r)
	}
	return nil
}

func parseSymbols(s string) ([]Symbol, error) {
	g, err := ReadSLIF("split", []byte("X ::= "+s))
	if err != nil {
		return nil, err
	}
	if len(g.G1) != 1 || g.G1[0].Seq != NotSeq {
		return nil, fmt.Errorf("want one sequence of symbols")
	}
	return g.G1[0].RHS, nil
}

// collapse makes every TokenDef with a kind a terminal, and drops the G1
// rules of the symbols it collapses.
func (c *cutter) collapse(CutSpec) error {
	for _, name := range c.order {
		for _, d := range c.defs[name] {
			if d.Kind == "" {
				continue
			}
			k, ok := kindNamed[d.Kind]
			if !ok {
				return cutErrorf("token %s: no lexer kind %q", name, d.Kind)
			}
			def := d.Def
			if def == "" {
				def = name
				if lit, ok := c.literalOnly(name); ok && len(lit) == 1 {
					def = "'" + lit[0] + "'"
				}
			}
			text, err := c.defRegexp(name, d.Def)
			if err != nil {
				return err
			}
			c.out.Terminals[name] = append(c.out.Terminals[name], Pred{Kind: k, Def: def, Except: d.Except, text: text})
		}
	}
	c.rules = slices.DeleteFunc(c.rules, func(r Rule) bool { return c.out.Terminals[r.LHS] != nil })
	return nil
}

// literals collapses, without a TokenDef, each literal-only G0 lexeme a kept
// G1 rule names, each literal to the kind the lexer gives it. Then it lists
// every collapse: split-introduced tokens, then collapsed G1 symbols in G1
// order, then G0 lexemes in G0 order.
func (c *cutter) literals(CutSpec) error {
	for _, r := range c.rules {
		for _, s := range r.RHS {
			if s.Kind != Name || c.out.Terminals[s.Text] != nil || c.defs[s.Text] != nil || c.g1[s.Text] != nil {
				continue
			}
			lits, ok := c.literalOnly(s.Text)
			if !ok {
				continue
			}
			for _, l := range lits {
				c.out.Terminals[s.Text] = append(c.out.Terminals[s.Text], Pred{Kind: literalKind(l), Def: "'" + l + "'", text: anchored(regexp.QuoteMeta(l))})
			}
		}
	}

	for _, name := range c.order {
		if c.out.Terminals[name] != nil && c.g1[name] == nil && c.g0[name] == nil {
			c.rewrite("token %s => %s", name, predsText(c.out.Terminals[name]))
		}
	}
	listed := map[string]bool{}
	for _, rules := range [][]Rule{c.g.G1, c.g.G0} {
		for _, r := range rules {
			if listed[r.LHS] || c.out.Terminals[r.LHS] == nil {
				continue
			}
			listed[r.LHS] = true
			op := "::="
			if c.g1[r.LHS] == nil {
				op = "~"
			}
			c.rewrite("collapse %s => %s", alternativesText(r.LHS, op, c.rulesOf(r.LHS)), predsText(c.out.Terminals[r.LHS]))
		}
	}
	return nil
}

// literalOnly returns a G0 lexeme's literal texts if each alternative is one
// literal.
func (c *cutter) literalOnly(name string) ([]string, bool) {
	rules := c.g0[name]
	if len(rules) == 0 || c.g1[name] != nil {
		return nil, false
	}
	var lits []string
	for _, r := range rules {
		if len(r.RHS) != 1 || r.RHS[0].Kind != Literal || r.Seq != NotSeq {
			return nil, false
		}
		lits = append(lits, r.RHS[0].Text)
	}
	return lits, true
}

// literalKind is the kind pvm's lexer gives a literal's text, as
// TestCutLiteralKinds checks against the lexer: word-shaped operators
// (and x=) are words, closing brackets and semicolons have their own kinds,
// and the rest are operators.
func literalKind(text string) lexer.Kind {
	switch {
	case text == ")" || text == "]" || text == "}":
		return lexer.CloseBracket
	case text == ";":
		return lexer.Semicolon
	case wordLiteral.MatchString(text):
		return lexer.Word
	}
	return lexer.Operator
}

var wordLiteral = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*=?$`)

// literalText unquotes a Def that is exactly one SLIF literal.
func literalText(def string) (string, bool) {
	if len(def) >= 2 && def[0] == '\'' && def[len(def)-1] == '\'' && !strings.Contains(def[1:len(def)-1], "'") {
		return def[1 : len(def)-1], true
	}
	return "", false
}

func (c *cutter) rulesOf(name string) []Rule {
	if r := c.g1[name]; r != nil {
		return r
	}
	return c.g0[name]
}

// prune drops G1 rules no longer reachable from the start symbol: their text
// now lives inside a token. A rule the original grammar could not reach is
// listed apart, since the cut did not absorb it.
func (c *cutter) prune(CutSpec) error {
	if len(c.g.G1) == 0 {
		return nil
	}
	start := c.g.G1[0].LHS
	before := reachable(start, c.g.G1)
	after := reachable(start, c.rules)
	listed := map[string]bool{}
	for _, r := range c.rules {
		if after[r.LHS] || listed[r.LHS] {
			continue
		}
		listed[r.LHS] = true
		form := "absorb"
		if !before[r.LHS] {
			form = "unreachable"
		}
		c.rewrite("%s %s", form, alternativesText(r.LHS, "::=", c.g1[r.LHS]))
	}
	c.rules = slices.DeleteFunc(c.rules, func(r Rule) bool { return !after[r.LHS] })
	return nil
}

func reachable(start string, rules []Rule) map[string]bool {
	byL := byLHS(rules)
	seen := map[string]bool{start: true}
	work := []string{start}
	for len(work) > 0 {
		n := work[len(work)-1]
		work = work[:len(work)-1]
		for _, r := range byL[n] {
			for _, s := range r.RHS {
				if s.Kind == Name && !seen[s.Text] && byL[s.Text] != nil {
					seen[s.Text] = true
					work = append(work, s.Text)
				}
			}
		}
	}
	return seen
}

// checkLeftovers fails on a G1 right-hand symbol that is neither a kept G1
// nonterminal nor a terminal.
func (c *cutter) checkLeftovers(CutSpec) error {
	kept := byLHS(c.rules)
	for _, r := range c.rules {
		for _, s := range r.RHS {
			switch {
			case s.Kind != Name:
				return cutErrorf("%s (line %d) still has the literal %s below pvm's token level", ruleText(r, "::="), r.Line, symbolText(s))
			case kept[s.Text] == nil && c.out.Terminals[s.Text] == nil:
				what := "names G0 symbol"
				if c.g0[s.Text] == nil {
					what = "names undefined symbol"
				}
				return cutErrorf("%s (line %d) %s %s, which no token predicate replaces", ruleText(r, "::="), r.Line, what, s.Text)
			}
		}
	}
	c.out.G1 = c.rules
	return nil
}

// priorities accounts for every :lexeme priority.
func (c *cutter) priorities(spec CutSpec) error {
	notes := map[string]string{}
	for _, n := range spec.Priorities {
		if _, ok := c.g.Priority[n.Symbol]; !ok {
			return cutErrorf("priority note for %s, which has no :lexeme priority", n.Symbol)
		}
		notes[n.Symbol] = n.How
	}
	var syms []string
	for s := range c.g.Priority {
		syms = append(syms, s)
	}
	sort.Strings(syms)
	for _, s := range syms {
		n := c.g.Priority[s]
		if c.g0[s] == nil {
			return cutErrorf(":lexeme ~ %s priority => %d names no G0 rule", s, n)
		}
		if how, ok := notes[s]; ok {
			c.rewrite("priority %s %d => %s", s, n, how)
			continue
		}
		preds := c.out.Terminals[s]
		if len(preds) == 0 || slices.ContainsFunc(preds, func(p Pred) bool { return p.Kind != lexer.Word }) {
			return cutErrorf(":lexeme ~ %s priority => %d is carried by no rewrite: it is not all words, and no priority note covers it", s, n)
		}
		for _, p := range preds {
			c.priorityWords = append(c.priorityWords, p.text.String())
		}
		c.rewrite("priority %s %d => except :priority", s, n)
	}
	return nil
}

// excepts compiles each predicate's Except, now the priority words are known.
func (c *cutter) excepts(CutSpec) error {
	for name, preds := range c.out.Terminals {
		for i := range preds {
			p := &preds[i]
			switch p.Except {
			case "":
			case ":priority":
				p.except = anchored(strings.Join(c.priorityWords, "|"))
			default:
				re, err := c.defRegexp(name+" except", p.Except)
				if err != nil {
					return err
				}
				p.except = re
			}
		}
	}
	return nil
}

func (c *cutter) trivia(spec CutSpec) error {
	covered := map[string]bool{}
	for _, t := range spec.Trivia {
		k, ok := kindNamed[t.Kind]
		if !ok {
			return cutErrorf("trivia: no lexer kind %q", t.Kind)
		}
		c.out.Trivia = append(c.out.Trivia, k)
		if t.Discard == "" {
			c.rewrite("trivia %s", t.Kind)
			continue
		}
		if !slices.Contains(c.g.Discard, t.Discard) {
			return cutErrorf("trivia %s: no :discard ~ %s", t.Kind, nameText(t.Discard))
		}
		covered[t.Discard] = true
		c.rewrite("trivia %s => :discard ~ %s", t.Kind, nameText(t.Discard))
	}
	for _, d := range c.g.Discard {
		if !covered[d] {
			return cutErrorf(":discard ~ %s names no trivia kind", nameText(d))
		}
	}
	return nil
}

// defRegexp is the anchored regexp of def, or of name's own rules if def is
// empty.
func (c *cutter) defRegexp(name, def string) (*regexp.Regexp, error) {
	var s string
	var err error
	if def == "" {
		s, err = c.symbolRegexp(name)
	} else {
		s, err = c.rhsRegexp(name, def)
	}
	if err != nil {
		return nil, err
	}
	re, err := regexp.Compile(`^(?:` + s + `)$`)
	if err != nil {
		return nil, cutErrorf("token %s: %v", name, err)
	}
	return re, nil
}

func anchored(s string) *regexp.Regexp { return regexp.MustCompile(`^(?:` + s + `)$`) }

func (c *cutter) rhsRegexp(name, def string) (string, error) {
	g, err := ReadSLIF(name, []byte("X ~ "+def))
	if err != nil {
		return "", cutErrorf("token %s: %v", name, err)
	}
	return c.rulesRegexp(name, g.G0, g1Gap)
}

// symbolRegexp is the text a symbol derives: the union of its TokenDefs'
// Defs if it has any, else its own rules.
func (c *cutter) symbolRegexp(name string) (string, error) {
	if s, ok := c.re[name]; ok {
		return s, nil
	}
	if c.busy[name] {
		return "", cutErrorf("%s derives itself other than by left recursion, so it is no single token", name)
	}
	c.busy[name] = true
	defer delete(c.busy, name)
	var alts []string
	own := len(c.defs[name]) == 0
	for _, d := range c.defs[name] {
		if d.Def == "" {
			own = true
			continue
		}
		s, err := c.rhsRegexp(name, d.Def)
		if err != nil {
			return "", err
		}
		alts = append(alts, s)
	}
	if own {
		rules := c.rulesOf(name)
		if len(rules) == 0 {
			return "", cutErrorf("%s has no rules and no TokenDef, so it has no text", name)
		}
		if c.g0[name] != nil && c.g1[name] != nil {
			return "", cutErrorf("%s has both G1 and G0 rules, so its text is ambiguous", name)
		}
		gap := ""
		if c.g1[name] != nil {
			gap = g1Gap
		}
		s, err := c.rulesRegexp(name, rules, gap)
		if err != nil {
			return "", err
		}
		alts = append(alts, s)
	}
	s := strings.Join(alts, "|")
	c.re[name] = s
	return s, nil
}

// g1Gap is what may sit between two symbols of a G1 sequence: Guacamole
// discards whitespace there, and pvm's lexer folds it into the one token, as
// in "$ foo" and "< $fh >". Inside a G0 rule nothing may.
const g1Gap = `\s*`

// rulesRegexp is the regexp of a symbol's alternatives, gap between the
// symbols of each. A left-recursive alternative X ::= X tail becomes a
// repeated tail after the others.
func (c *cutter) rulesRegexp(name string, rules []Rule, gap string) (string, error) {
	var base, tails []string
	for _, r := range rules {
		rhs := r.RHS
		recursive := len(rhs) > 1 && rhs[0] == (Symbol{Name, name}) && r.Seq == NotSeq
		if recursive {
			rhs = rhs[1:]
		}
		var parts []string
		for _, s := range rhs {
			p, err := c.symbolText(s)
			if err != nil {
				return "", err
			}
			parts = append(parts, p)
		}
		seq := strings.Join(parts, gap)
		switch r.Seq {
		case Star:
			seq = "(?:" + gap + seq + ")*"
		case Plus:
			seq = "(?:" + gap + seq + ")+"
		}
		if recursive {
			tails = append(tails, gap+seq)
		} else {
			base = append(base, seq)
		}
	}
	if len(base) == 0 {
		return "", cutErrorf("%s has only left-recursive alternatives", name)
	}
	s := "(?:" + strings.Join(base, "|") + ")"
	if len(tails) > 0 {
		s += "(?:" + strings.Join(tails, "|") + ")*"
	}
	return s, nil
}

func (c *cutter) symbolText(s Symbol) (string, error) {
	switch s.Kind {
	case Literal:
		return regexp.QuoteMeta(s.Text), nil
	case Class:
		if _, err := regexp.Compile(s.Text); err != nil {
			return "", cutErrorf("character class %s: %v", s.Text, err)
		}
		return s.Text, nil
	}
	r, err := c.symbolRegexp(s.Text)
	return "(?:" + r + ")", err
}

func predsText(preds []Pred) string {
	var parts []string
	for _, p := range preds {
		parts = append(parts, p.String())
	}
	return strings.Join(parts, " | ")
}

func alternativesText(lhs, op string, rules []Rule) string {
	var alts []string
	for _, r := range rules {
		alts = append(alts, rhsText(r))
	}
	return nameText(lhs) + " " + op + " " + strings.Join(alts, " | ")
}

func ruleText(r Rule, op string) string { return nameText(r.LHS) + " " + op + " " + rhsText(r) }

func rhsText(r Rule) string {
	var parts []string
	for _, s := range r.RHS {
		parts = append(parts, symbolText(s))
	}
	s := strings.Join(parts, " ")
	switch r.Seq {
	case Star:
		s += "*"
	case Plus:
		s += "+"
	}
	return s
}

func symbolText(s Symbol) string {
	switch s.Kind {
	case Literal:
		return "'" + s.Text + "'"
	case Class:
		return s.Text
	}
	return nameText(s.Text)
}

// nameText writes a name the way SLIF does: bracketed if it has a space.
func nameText(n string) string {
	if strings.Contains(n, " ") {
		return "<" + n + ">"
	}
	return n
}
