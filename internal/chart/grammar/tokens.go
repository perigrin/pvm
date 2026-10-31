// ABOUTME: Compiles the grammar's token classes into a sparse table: a token's key yields a terminal-set id.
// ABOUTME: Spec §7: lookup is constant-time, allocation-free and regexp-free; unions are precomputed.

package grammar

import (
	"cmp"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"

	"tamarou.com/pvm/internal/lexer"
)

// TokenView is what the lookup reads of the token itself.
type TokenView struct {
	Kind                    lexer.Kind
	Text                    string
	OpensBlock, TermFollows bool
	// Prefix is the sigil or quote-prefix class, 0 for none.
	Prefix byte
}

// Brace says what a `{` opens.
type Brace int

const (
	BraceNone Brace = iota
	BraceBlock
	BraceSubscript
	BraceHash
)

// Neighbour is a neighbouring token as the key sees it: its lexer kind, its
// flags and, for words, operators and closing brackets, its text, never its
// own terminal set, so the key does not recurse.
type Neighbour struct {
	Kind                    lexer.Kind
	Text                    string
	OpensBlock, TermFollows bool
	Brace                   Brace // only on `{`
}

// Boundary is the class of the neighbour that a file's edge lacks.
var Boundary = Neighbour{Kind: -1}

// ClassOf is t's class given its predecessor's. A `{` the lexer decided
// (OpensBlock) keeps that decision; an undecided one is a subscript after a
// variable, `->` or a subscript's closing bracket, a block after a deref
// sigil (`@{ ... }`), and an anonymous hash otherwise, as in toke.c.
// shortcut: a `{` after a do or eval block's `}` is read as a hash where toke.c
// reads a subscript; telling that `}` apart needs a lexer bit, outside spec §11.
func ClassOf(t TokenView, pred Neighbour) Neighbour {
	c := Neighbour{Kind: t.Kind, OpensBlock: t.OpensBlock, TermFollows: t.TermFollows}
	switch t.Kind {
	case lexer.Word, lexer.Operator, lexer.CloseBracket:
		c.Text = t.Text
	}
	if t.Kind == lexer.Operator && t.Text == "{" {
		switch {
		case t.OpensBlock, pred.Kind == lexer.DerefSigil:
			c.Brace = BraceBlock
		case pred.Kind == lexer.Variable, pred.Kind == lexer.CloseBracket && !pred.OpensBlock,
			pred.Kind == lexer.Operator && pred.Text == "->":
			c.Brace = BraceSubscript
		default:
			c.Brace = BraceHash
		}
	}
	return c
}

// TokenClass declares that tokens whose key matches every part given (a nil
// part matches anything) admit Terminals. A key several classes match admits
// the union of theirs.
type TokenClass struct {
	Terminals               []int
	Kind                    *lexer.Kind
	Text                    *string
	OpensBlock, TermFollows *bool
	Prefix                  *byte
	Prev2, Prev1, Next      *Neighbour
}

func ptr[T any](v T) *T { return &v }

const nParts = 8

// free marks a part a class leaves unconstrained. Every value, a lexer kind
// included, is a non-negative integer: text and neighbour values are the
// dictionary ids CompileTokens gives them, 0 meaning one no class names.
const free = -1

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

// iclass is a Neighbour with its text replaced by its dictionary id.
type iclass struct {
	kind, text  int
	opens, term bool
	brace       Brace
}

// node tests one part: kids is indexed by the part's value, def covers the
// rest. A node with no kids is a leaf.
type node struct {
	part int
	kids []*node
	def  *node
	set  []int
	id   int
	uid  int // set when the node is interned while the table is built
}

// TokenTable is the compiled classes. Sets[0] is the empty set, what a key
// no class matches yields; Sets is what Tables.TerminalSets takes.
type TokenTable struct {
	Sets [][]int
	// Edge is the class id of Boundary, for the neighbours a file's edge lacks.
	Edge    int
	root    *node
	text    map[string]int
	classes map[iclass]int
	texts   []string    // by text id - 1, for Text()
	nbs     []Neighbour // by class id - 1, for Text()
}

// Token is a token with its text and class looked up in the dictionaries.
// Intern it once; the class id is what the three tokens around it reuse.
type Token struct {
	N     Neighbour // the token's class, the next token's predecessor
	Class int       // N's dictionary id
	key   [nParts]int
}

func (tt *TokenTable) classID(n Neighbour, text int) int {
	return tt.classes[iclass{int(n.Kind), text, n.OpensBlock, n.TermFollows, n.Brace}]
}

// ClassID is the dictionary id of n, 0 if no class names it.
func (tt *TokenTable) ClassID(n Neighbour) int { return tt.classID(n, tt.text[n.Text]) }

// Intern looks t up in the dictionaries, with pred the class of the
// significant token before it. It hashes t's text once.
func (tt *TokenTable) Intern(t TokenView, pred Neighbour) Token {
	n, text := ClassOf(t, pred), tt.text[t.Text]
	ct := text
	if n.Text == "" {
		ct = tt.text[""]
	}
	return Token{N: n, Class: tt.classID(n, ct),
		key: [nParts]int{int(t.Kind), text, b2i(t.OpensBlock), b2i(t.TermFollows), int(t.Prefix)}}
}

// CompileTokens builds the table. A part is tested only below the classes
// that name it, so the table is refined by a neighbour only where a class
// needs it. Dictionary ids are numbered by value, so the table does not
// depend on the order of classes or on map iteration.
func CompileTokens(classes []TokenClass) (*TokenTable, error) {
	for i, c := range classes {
		if len(c.Terminals) == 0 {
			return nil, fmt.Errorf("token class %d admits no terminals; an unmatched key already yields the empty set", i)
		}
	}
	tt := &TokenTable{Sets: [][]int{{}}, text: map[string]int{}, classes: map[iclass]int{}}
	// The dictionaries: every text and every neighbour class a class names.
	texts, nbs := map[string]bool{}, map[Neighbour]bool{}
	for _, c := range classes {
		if c.Text != nil {
			texts[*c.Text] = true
		}
		for _, nb := range []*Neighbour{c.Prev2, c.Prev1, c.Next} {
			if nb != nil {
				texts[nb.Text], nbs[*nb] = true, true
			}
		}
	}
	tt.texts = slices.Sorted(maps.Keys(texts))
	for i, s := range tt.texts {
		tt.text[s] = i + 1
	}
	tt.nbs = slices.SortedFunc(maps.Keys(nbs), func(a, b Neighbour) int { return cmp.Compare(fmt.Sprintf("%+v", a), fmt.Sprintf("%+v", b)) })
	for i, nb := range tt.nbs {
		tt.classes[iclass{int(nb.Kind), tt.text[nb.Text], nb.OpensBlock, nb.TermFollows, nb.Brace}] = i + 1
	}
	tt.Edge = tt.ClassID(Boundary)

	pats := make([][nParts]int, len(classes))
	for i, c := range classes {
		p := [nParts]int{free, free, free, free, free, free, free, free}
		if c.Kind != nil {
			p[0] = int(*c.Kind)
		}
		if c.Text != nil {
			p[1] = tt.text[*c.Text]
		}
		if c.OpensBlock != nil {
			p[2] = b2i(*c.OpensBlock)
		}
		if c.TermFollows != nil {
			p[3] = b2i(*c.TermFollows)
		}
		if c.Prefix != nil {
			p[4] = int(*c.Prefix)
		}
		for j, nb := range []*Neighbour{c.Prev2, c.Prev1, c.Next} {
			if nb != nil {
				p[5+j] = tt.ClassID(*nb)
			}
		}
		pats[i] = p
	}

	// Identical subtrees are one node, so the table does not grow with the
	// number of classes that differ only in a part tested above them.
	shared, memo := map[string]*node{}, map[string]*node{}
	var leaves []*node
	intern := func(key string, n *node) *node {
		if old, ok := shared[key]; ok {
			return old
		}
		n.uid = len(shared)
		shared[key] = n
		if n.kids == nil {
			leaves = append(leaves, n)
		}
		return n
	}
	var build func(in []int, part int) *node
	build = func(in []int, part int) *node {
		for ; part < nParts; part++ {
			if slices.ContainsFunc(in, func(c int) bool { return pats[c][part] != free }) {
				break
			}
		}
		memoKey := fmt.Sprint(part, in)
		if n, ok := memo[memoKey]; ok {
			return n
		}
		var n *node
		if part == nParts {
			var set []int
			for _, c := range in {
				set = append(set, classes[c].Terminals...)
			}
			slices.Sort(set)
			set = slices.Compact(set)
			n = intern("leaf"+fmt.Sprint(set), &node{set: set})
		} else {
			var wild, values []int
			for _, c := range in {
				if v := pats[c][part]; v == free {
					wild = append(wild, c)
				} else {
					values = append(values, v)
				}
			}
			slices.Sort(values)
			values = slices.Compact(values)
			cand := &node{part: part, def: build(wild, part+1), kids: make([]*node, values[len(values)-1]+1)}
			sig := []string{fmt.Sprint(part, cand.def.uid)}
			for i := range cand.kids {
				cand.kids[i] = cand.def
			}
			for _, v := range values {
				same := slices.DeleteFunc(slices.Clone(in), func(o int) bool { q := pats[o][part]; return q != free && q != v })
				cand.kids[v] = build(same, part+1)
			}
			for _, k := range cand.kids {
				sig = append(sig, fmt.Sprint(k.uid))
			}
			n = intern(strings.Join(sig, ";"), cand)
		}
		memo[memoKey] = n
		return n
	}
	all := make([]int, len(classes))
	for i := range all {
		all[i] = i
	}
	tt.root = build(all, 0)
	for _, l := range leaves {
		if len(l.set) > 0 {
			tt.Sets = append(tt.Sets, l.set)
		}
	}
	slices.SortFunc(tt.Sets, slices.Compare)
	tt.Sets = slices.CompactFunc(tt.Sets, slices.Equal)
	for _, l := range leaves {
		l.id, _ = slices.BinarySearchFunc(tt.Sets, l.set, slices.Compare)
	}
	return tt, nil
}

// Lookup is the terminal-set id of an interned token given the class ids of
// the two significant tokens before it and the one after (Edge at a file's
// edge). It allocates nothing and hashes nothing.
func (tt *TokenTable) Lookup(t Token, prev2, prev1, next int) int {
	k := t.key
	k[5], k[6], k[7] = prev2, prev1, next
	n := tt.root
	for n.kids != nil {
		if v := k[n.part]; uint(v) < uint(len(n.kids)) {
			n = n.kids[v]
		} else {
			n = n.def
		}
	}
	return n.id
}

// Text is the table as deterministic text: the sets, then the nodes, each
// shared node once.
func (tt *TokenTable) Text() []byte {
	var b strings.Builder
	for i, s := range tt.Sets {
		fmt.Fprintf(&b, "set %d%s\n", i, ints(s))
	}
	// value names a part's value in the dump: ids are an implementation detail.
	value := func(part, v int) string {
		switch {
		case part == 1:
			return fmt.Sprintf("%q", tt.texts[v-1])
		case part >= 5:
			return fmt.Sprintf("%+v", tt.nbs[v-1])
		}
		return fmt.Sprint(v)
	}
	num := map[*node]int{}
	var dump func(n *node) int
	dump = func(n *node) int {
		if i, ok := num[n]; ok {
			return i
		}
		if n.kids == nil {
			num[n] = len(num)
			fmt.Fprintf(&b, "node %d leaf set %d\n", num[n], n.id)
			return num[n]
		}
		kids := make([]int, len(n.kids))
		for v, k := range n.kids {
			if k != n.def {
				kids[v] = dump(k)
			}
		}
		def := dump(n.def)
		num[n] = len(num)
		fmt.Fprintf(&b, "node %d part %d otherwise %d\n", num[n], n.part, def)
		for v, k := range n.kids {
			if k != n.def {
				fmt.Fprintf(&b, "  %s -> %d\n", value(n.part, v), kids[v])
			}
		}
		return num[n]
	}
	fmt.Fprintf(&b, "root %d\n", dump(tt.root))
	return []byte(b.String())
}

// FillTerminalSets stores the terminal-set table the recognizer indexes. It
// refuses one whose set 0 is missing or non-empty, since id 0 is the empty
// set that an unmapped token carries.
func (t *Tables) FillTerminalSets(sets [][]int) error {
	if len(sets) == 0 || len(sets[0]) != 0 {
		return errors.New("terminal sets: set 0 must exist and be empty")
	}
	t.TerminalSets = sets
	return nil
}

// Nodes is the number of distinct nodes in the table.
func (tt *TokenTable) Nodes() int {
	seen := map[*node]bool{}
	var walk func(*node)
	walk = func(n *node) {
		if seen[n] {
			return
		}
		seen[n] = true
		if n.kids != nil {
			walk(n.def)
			for _, k := range n.kids {
				walk(k)
			}
		}
	}
	walk(tt.root)
	return len(seen)
}
