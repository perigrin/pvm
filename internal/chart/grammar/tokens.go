// ABOUTME: Compiles the grammar's token classes into a sparse table: a token's key yields a terminal-set id.
// ABOUTME: Spec §7: lookup is constant-time, allocation-free and regexp-free; unions are precomputed.

package grammar

import (
	"cmp"
	"errors"
	"fmt"
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

// partVal is one part of a key; only the field its part uses is set.
type partVal struct {
	n int
	s string
	c Neighbour
}

const nParts = 8

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

// parts is the key of a token in part order: kind, text, the two flags, the
// prefix, then the second previous, previous and next neighbours.
func parts(t TokenView, p2, p1, n Neighbour) [nParts]partVal {
	return [nParts]partVal{{n: int(t.Kind)}, {s: t.Text}, {n: b2i(t.OpensBlock)}, {n: b2i(t.TermFollows)},
		{n: int(t.Prefix)}, {c: p2}, {c: p1}, {c: n}}
}

func (c TokenClass) pattern() (p [nParts]*partVal) {
	if c.Kind != nil {
		p[0] = &partVal{n: int(*c.Kind)}
	}
	if c.Text != nil {
		p[1] = &partVal{s: *c.Text}
	}
	if c.OpensBlock != nil {
		p[2] = &partVal{n: b2i(*c.OpensBlock)}
	}
	if c.TermFollows != nil {
		p[3] = &partVal{n: b2i(*c.TermFollows)}
	}
	if c.Prefix != nil {
		p[4] = &partVal{n: int(*c.Prefix)}
	}
	for i, nb := range []*Neighbour{c.Prev2, c.Prev1, c.Next} {
		if nb != nil {
			p[5+i] = &partVal{c: *nb}
		}
	}
	return p
}

// node tests one part: kids by value, def when the value is not among them.
// A node with no kids is a leaf.
type node struct {
	part int
	kids map[partVal]*node
	def  *node
	set  []int
	id   int
}

// TokenTable is the compiled classes. Sets[0] is the empty set, what a key
// no class matches yields; Sets is what Tables.TerminalSets takes.
type TokenTable struct {
	Sets [][]int
	root *node
}

// CompileTokens builds the table. A part is tested only below the classes
// that name it, so the table is refined by a neighbour only where a class
// needs it. Set ids are numbered by the sets' contents, so the table does
// not depend on the order of classes or on map iteration.
func CompileTokens(classes []TokenClass) (*TokenTable, error) {
	pats := make([][nParts]*partVal, len(classes))
	for i, c := range classes {
		if len(c.Terminals) == 0 {
			return nil, fmt.Errorf("token class %d admits no terminals; an unmatched key already yields the empty set", i)
		}
		pats[i] = c.pattern()
	}
	var build func(in []int, part int) *node
	build = func(in []int, part int) *node {
		for ; part < nParts; part++ {
			if slices.ContainsFunc(in, func(c int) bool { return pats[c][part] != nil }) {
				break
			}
		}
		if part == nParts {
			var set []int
			for _, c := range in {
				set = append(set, classes[c].Terminals...)
			}
			slices.Sort(set)
			return &node{set: slices.Compact(set)}
		}
		n := &node{part: part, kids: map[partVal]*node{}}
		var wild []int
		for _, c := range in {
			if pats[c][part] == nil {
				wild = append(wild, c)
			}
		}
		n.def = build(wild, part+1)
		for _, c := range in {
			if p := pats[c][part]; p != nil && n.kids[*p] == nil {
				same := slices.DeleteFunc(slices.Clone(in), func(o int) bool { q := pats[o][part]; return q != nil && *q != *p })
				n.kids[*p] = build(same, part+1)
			}
		}
		return n
	}
	all := make([]int, len(classes))
	for i := range all {
		all[i] = i
	}
	tt := &TokenTable{root: build(all, 0), Sets: [][]int{{}}}
	var leaves []*node
	var walk func(*node)
	walk = func(n *node) {
		if n.kids == nil {
			leaves = append(leaves, n)
			return
		}
		walk(n.def)
		for _, k := range n.kids {
			walk(k)
		}
	}
	walk(tt.root)
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

// Lookup is the terminal-set id of a token given the classes of the two
// significant tokens before it and the one after (Boundary at a file's edge).
func (tt *TokenTable) Lookup(t TokenView, prev2, prev1, next Neighbour) int {
	k := parts(t, prev2, prev1, next)
	n := tt.root
	for n.kids != nil {
		if c, ok := n.kids[k[n.part]]; ok {
			n = c
		} else {
			n = n.def
		}
	}
	return n.id
}

// Text is the table as deterministic text: the sets, then the tests.
func (tt *TokenTable) Text() []byte {
	var b strings.Builder
	for i, s := range tt.Sets {
		fmt.Fprintf(&b, "set %d%s\n", i, ints(s))
	}
	var dump func(n *node, depth int)
	dump = func(n *node, depth int) {
		pad := strings.Repeat(" ", depth)
		if n.kids == nil {
			fmt.Fprintf(&b, "%s-> %d\n", pad, n.id)
			return
		}
		keys := make([]partVal, 0, len(n.kids))
		for k := range n.kids {
			keys = append(keys, k)
		}
		slices.SortFunc(keys, func(a, b partVal) int { return cmp.Compare(fmt.Sprintf("%+v", a), fmt.Sprintf("%+v", b)) })
		for _, k := range keys {
			fmt.Fprintf(&b, "%spart %d %+v\n", pad, n.part, k)
			dump(n.kids[k], depth+1)
		}
		fmt.Fprintf(&b, "%spart %d otherwise\n", pad, n.part)
		dump(n.def, depth+1)
	}
	dump(tt.root, 0)
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
