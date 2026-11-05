// ABOUTME: Tests the token-class compiler on toy grammars: each key part, unions, edges and determinism.
// ABOUTME: Terminal ids are bare integers; no committed token table exists yet.

package grammar

import (
	"bytes"
	"fmt"
	"slices"
	"testing"

	"tamarou.com/pvm/internal/lexer"
)

func word(s string) TokenView { return TokenView{Kind: lexer.Word, Text: s} }
func op(s string) TokenView   { return TokenView{Kind: lexer.Operator, Text: s} }

func mustCompile(t *testing.T, classes ...TokenClass) *TokenTable {
	t.Helper()
	tt, err := CompileTokens(classes)
	if err != nil {
		t.Fatal(err)
	}
	return tt
}

// interned is t interned after pred.
func (tt *TokenTable) interned(t TokenView, pred Neighbour) Token {
	var it Token
	tt.Intern(&t, &pred, &it)
	return it
}

// lookup interns the token and its three neighbours' classes, as a caller does.
func (tt *TokenTable) lookup(t TokenView, p2, p1, n Neighbour) int {
	return tt.Lookup(tt.interned(t, none), tt.ClassID(p2), tt.ClassID(p1), tt.ClassID(n))
}

func (tt *TokenTable) set(t TokenView, p2, p1, n Neighbour) []int {
	return tt.Sets[tt.lookup(t, p2, p1, n)]
}

var none = Neighbour{Kind: lexer.Whitespace}

func TestTokenKeyParts(t *testing.T) {
	semi, wordC, openSub := ClassOf(op(";"), none), ClassOf(word("w"), none), ClassOf(op("{"), ClassOf(TokenView{Kind: lexer.Variable, Text: "$h"}, none))
	openHash := ClassOf(op("{"), ClassOf(op("+"), none))
	base := word("foo")
	cases := []struct {
		name      string
		class     TokenClass
		hit, miss func() (TokenView, Neighbour, Neighbour, Neighbour)
	}{
		{"kind", TokenClass{Kind: ptr(lexer.Word)},
			func() (TokenView, Neighbour, Neighbour, Neighbour) { return base, none, none, none },
			func() (TokenView, Neighbour, Neighbour, Neighbour) { return op("foo"), none, none, none }},
		{"text", TokenClass{Text: ptr("foo")},
			func() (TokenView, Neighbour, Neighbour, Neighbour) { return base, none, none, none },
			func() (TokenView, Neighbour, Neighbour, Neighbour) { return word("bar"), none, none, none }},
		{"opensblock", TokenClass{OpensBlock: ptr(true)},
			func() (TokenView, Neighbour, Neighbour, Neighbour) {
				b := op("{")
				b.OpensBlock = true
				return b, none, none, none
			},
			func() (TokenView, Neighbour, Neighbour, Neighbour) { return op("{"), none, none, none }},
		{"termfollows", TokenClass{TermFollows: ptr(true)},
			func() (TokenView, Neighbour, Neighbour, Neighbour) {
				v := base
				v.TermFollows = true
				return v, none, none, none
			},
			func() (TokenView, Neighbour, Neighbour, Neighbour) { return base, none, none, none }},
		{"prefix", TokenClass{Prefix: ptr(byte('$'))},
			func() (TokenView, Neighbour, Neighbour, Neighbour) {
				v := base
				v.Prefix = '$'
				return v, none, none, none
			},
			func() (TokenView, Neighbour, Neighbour, Neighbour) {
				v := base
				v.Prefix = '@'
				return v, none, none, none
			}},
		{"previous", TokenClass{Prev1: &semi},
			func() (TokenView, Neighbour, Neighbour, Neighbour) { return base, none, semi, none },
			func() (TokenView, Neighbour, Neighbour, Neighbour) { return base, none, wordC, none }},
		{"second previous", TokenClass{Prev2: &semi},
			func() (TokenView, Neighbour, Neighbour, Neighbour) { return base, semi, none, none },
			func() (TokenView, Neighbour, Neighbour, Neighbour) { return base, none, semi, none }},
		{"next", TokenClass{Next: &semi},
			func() (TokenView, Neighbour, Neighbour, Neighbour) { return base, none, none, semi },
			func() (TokenView, Neighbour, Neighbour, Neighbour) { return base, none, none, wordC }},
		{"subscript or hash brace", TokenClass{Prev2: &openSub},
			func() (TokenView, Neighbour, Neighbour, Neighbour) { return base, openSub, none, none },
			func() (TokenView, Neighbour, Neighbour, Neighbour) { return base, openHash, none, none }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			c.class.Terminals = []int{1}
			tt := mustCompile(t, c.class)
			v, a, b, d := c.hit()
			if got := tt.set(v, a, b, d); !slices.Equal(got, []int{1}) {
				t.Errorf("hit = %v, want [1]", got)
			}
			v, a, b, d = c.miss()
			if got := tt.set(v, a, b, d); len(got) != 0 {
				t.Errorf("miss = %v, want empty", got)
			}
		})
	}
}

func TestTokenUnions(t *testing.T) {
	tt := mustCompile(t,
		TokenClass{Terminals: []int{1}, Kind: ptr(lexer.Word)},
		TokenClass{Terminals: []int{2}, Text: ptr("foo")},
		TokenClass{Terminals: []int{3}, Text: ptr("bar")})
	both, one := tt.lookup(word("foo"), none, none, none), tt.lookup(word("bar"), none, none, none)
	if !slices.Equal(tt.Sets[both], []int{1, 2}) || !slices.Equal(tt.Sets[one], []int{1, 3}) || both == one {
		t.Errorf("foo = %v (id %d), bar = %v (id %d), want [1 2] and [1 3] apart", tt.Sets[both], both, tt.Sets[one], one)
	}
	if tt.lookup(word("foo"), none, none, none) != both {
		t.Error("the same key yields a different id")
	}
	for i := range tt.Sets {
		for j := range i {
			if slices.Equal(tt.Sets[i], tt.Sets[j]) {
				t.Errorf("set %v is stored twice", tt.Sets[i])
			}
		}
	}
}

func TestTokenMultiTerminal(t *testing.T) {
	tt := mustCompile(t, TokenClass{Terminals: []int{3, 1, 2, 1}, Text: ptr("-")})
	if got := tt.set(op("-"), none, none, none); !slices.Equal(got, []int{1, 2, 3}) {
		t.Errorf("set = %v, want [1 2 3]", got)
	}
}

func TestTokenLookupAllocationFree(t *testing.T) {
	semi := ClassOf(op(";"), none)
	tt := mustCompile(t,
		TokenClass{Terminals: []int{1}, Kind: ptr(lexer.Word), Text: ptr("foo"), Prev1: &semi, Next: &semi},
		TokenClass{Terminals: []int{2}, Prefix: ptr(byte('$')), Prev2: &semi})
	tok := TokenView{Kind: lexer.Word, Text: "foo", Prefix: '$'}
	c := tt.ClassID(semi)
	if n := testing.AllocsPerRun(1000, func() { tt.Lookup(tt.interned(tok, semi), c, c, c) }); n != 0 {
		t.Errorf("Intern and Lookup allocate %v times", n)
	}
	if got := tt.set(tok, semi, semi, semi); !slices.Equal(got, []int{1, 2}) {
		t.Errorf("the measured lookup yields %v, want [1 2]", got)
	}
}

func TestTokenUnmatchedKeyEmptySet(t *testing.T) {
	tt := mustCompile(t, TokenClass{Terminals: []int{1}, Text: ptr("foo")}, TokenClass{Terminals: []int{2}, Text: ptr("bar")})
	if id := tt.lookup(word("baz"), none, none, none); id != 0 {
		t.Errorf("unmatched key id = %d, want exactly 0", id)
	}
	if len(tt.Sets[0]) != 0 {
		t.Errorf("set 0 = %v, want empty", tt.Sets[0])
	}
	foo, bar := tt.lookup(word("foo"), none, none, none), tt.lookup(word("bar"), none, none, none)
	if foo == 0 || bar == 0 || foo == bar || !slices.Equal(tt.Sets[foo], []int{1}) {
		t.Errorf("matching ids foo=%d bar=%d, sets %v", foo, bar, tt.Sets)
	}
}

func TestTokenKeyAtFileEdges(t *testing.T) {
	semi := ClassOf(op(";"), none)
	tt := mustCompile(t,
		TokenClass{Terminals: []int{1}, Text: ptr("a"), Prev1: &semi},
		TokenClass{Terminals: []int{2}, Text: ptr("a"), Prev1: ptr(Boundary), Prev2: ptr(Boundary)},
		TokenClass{Terminals: []int{3}, Text: ptr("z"), Next: ptr(Boundary)})
	if got := tt.set(word("a"), Boundary, Boundary, semi); !slices.Equal(got, []int{2}) {
		t.Errorf("first token = %v, want [2]: a class refined by a previous token must not match at the start", got)
	}
	if got := tt.set(word("z"), semi, semi, Boundary); !slices.Equal(got, []int{3}) {
		t.Errorf("last token = %v, want [3]", got)
	}
	if got := tt.set(word("a"), Boundary, semi, semi); !slices.Equal(got, []int{1}) {
		t.Errorf("second token = %v, want [1]", got)
	}
}

func TestTokenBraceKeepsLexerDecision(t *testing.T) {
	plus, h := ClassOf(op("+"), none), ClassOf(TokenView{Kind: lexer.Variable, Text: "$h"}, none)
	block := op("{")
	block.OpensBlock = true
	if got := ClassOf(block, plus); got.Brace != BraceBlock {
		t.Errorf("decided { after + = %v, want block", got.Brace)
	}
	if got := ClassOf(block, h); got.Brace != BraceBlock {
		t.Errorf("decided { after a variable = %v, want block", got.Brace)
	}
	if got := ClassOf(op("{"), plus); got.Brace != BraceHash {
		t.Errorf("undecided { after + = %v, want hash", got.Brace)
	}
	if got := ClassOf(op("{"), h); got.Brace != BraceSubscript {
		t.Errorf("undecided { after a variable = %v, want subscript", got.Brace)
	}
	deref := ClassOf(TokenView{Kind: lexer.DerefSigil, Text: "@"}, none)
	if got := ClassOf(op("{"), deref); got.Brace != BraceBlock {
		t.Errorf("undecided { after a deref sigil = %v, want block", got.Brace)
	}
	arrow := ClassOf(op("->"), none)
	if got := ClassOf(op("{"), arrow); got.Brace != BraceSubscript {
		t.Errorf("undecided { after -> = %v, want subscript", got.Brace)
	}
	dash := ClassOf(op("-"), none)
	hashed, blocked := ClassOf(block, plus), ClassOf(op("{"), plus)
	tt := mustCompile(t, TokenClass{Terminals: []int{1}, Text: ptr("foo"), Prev2: &blocked})
	if got := tt.set(word("foo"), blocked, dash, none); !slices.Equal(got, []int{1}) {
		t.Errorf("undecided brace key = %v", got)
	}
	if got := tt.set(word("foo"), hashed, dash, none); len(got) != 0 {
		t.Errorf("decided block after + keyed as hash: %v", got)
	}
}

func TestTokenKeyDoesNotRecurse(t *testing.T) {
	semi := ClassOf(op(";"), none)
	tt := mustCompile(t,
		TokenClass{Terminals: []int{1}, Text: ptr("foo"), Prev1: &semi},
		TokenClass{Terminals: []int{2}, Text: ptr("foo")},
		TokenClass{Terminals: []int{3}, Text: ptr("x"), Next: ptr(ClassOf(word("foo"), none))})
	inCtxA, inCtxB := tt.set(word("foo"), none, semi, none), tt.set(word("foo"), none, none, none)
	if slices.Equal(inCtxA, inCtxB) {
		t.Fatalf("foo's two contexts must differ, both %v", inCtxA)
	}
	a := tt.lookup(word("x"), none, none, ClassOf(word("foo"), semi))
	b := tt.lookup(word("x"), none, none, ClassOf(word("foo"), none))
	if a != b || !slices.Equal(tt.Sets[a], []int{3}) {
		t.Errorf("same-class neighbours keyed apart: %d vs %d", a, b)
	}
}

// A variable's name is not part of its class, so $x and $y key alike.
func TestTokenNeighbourTextOnlyForWordsOperatorsBrackets(t *testing.T) {
	x, y := TokenView{Kind: lexer.Variable, Text: "$x"}, TokenView{Kind: lexer.Variable, Text: "$y"}
	if a, b := ClassOf(x, none), ClassOf(y, none); a != b {
		t.Errorf("$x and $y classes differ: %+v vs %+v", a, b)
	}
	if a, b := ClassOf(word("if"), none), ClassOf(word("my"), none); a == b {
		t.Error("two words share a class")
	}
	cx := ClassOf(x, none)
	tt := mustCompile(t, TokenClass{Terminals: []int{1}, Text: ptr("foo"), Prev1: &cx})
	if got := tt.set(word("foo"), none, ClassOf(y, none), none); !slices.Equal(got, []int{1}) {
		t.Errorf("foo after $y = %v, want the class naming $x's class", got)
	}
}

// A class naming no part matches every key; set 0 stays the reserved empty set.
func TestTokenClassWithNoPartsMatchesEverything(t *testing.T) {
	tt := mustCompile(t, TokenClass{Terminals: []int{1}})
	if id := tt.lookup(word("anything"), none, none, none); !slices.Equal(tt.Sets[id], []int{1}) || id == 0 {
		t.Errorf("lookup = id %d, sets %v", id, tt.Sets)
	}
	if err := (&Tables{}).FillTerminalSets(tt.Sets); err != nil {
		t.Errorf("sets refused: %v", err)
	}
}

// The refined subtree is shared by every keyword, so the table's size does
// not grow with the number of text-keyed classes.
func TestTokenNodesSharedAcrossKeywords(t *testing.T) {
	semi, brace := ClassOf(op(";"), none), ClassOf(op("{"), none)
	size := func(keywords int) int {
		classes := []TokenClass{
			{Terminals: []int{1}, Kind: ptr(lexer.Word), Prev1: &semi},
			{Terminals: []int{2}, Kind: ptr(lexer.Word), Prev2: &brace, Next: &semi},
			{Terminals: []int{3}, Kind: ptr(lexer.Word), Next: &brace},
		}
		for i := range keywords {
			classes = append(classes, TokenClass{Terminals: []int{4 + i%7}, Text: ptr(fmt.Sprintf("kw%d", i))})
		}
		return mustCompile(t, classes...).Nodes()
	}
	small, large := size(20), size(400)
	if large > small {
		t.Errorf("nodes grew from %d to %d for 20 -> 400 keywords: refined subtrees are copied", small, large)
	}
}

// The class id Intern gives a token is the id ClassID gives its class, for
// kinds that carry no text too, so a caller can pass Intern(pred).Class.
func TestTokenInternedClassMatchesClassID(t *testing.T) {
	for name, pred := range map[string]TokenView{
		"variable": {Kind: lexer.Variable, Text: "$x"},
		"number":   {Kind: lexer.Number, Text: "42"},
		"semi":     {Kind: lexer.Semicolon, Text: ";"},
		"word":     word("my"),
	} {
		t.Run(name, func(t *testing.T) {
			c := ClassOf(pred, none)
			tt := mustCompile(t, TokenClass{Terminals: []int{1}, Text: ptr("foo"), Prev1: &c})
			p := tt.interned(pred, none)
			if p.Class == 0 || p.Class != tt.ClassID(c) {
				t.Fatalf("Intern class %d, ClassID %d", p.Class, tt.ClassID(c))
			}
			if got := tt.Sets[tt.Lookup(tt.interned(word("foo"), p.N), tt.Edge, p.Class, tt.Edge)]; !slices.Equal(got, []int{1}) {
				t.Errorf("foo after %s via Intern = %v, want [1]", name, got)
			}
		})
	}
}

func TestTokenTableDeterministic(t *testing.T) {
	semi := ClassOf(op(";"), none)
	classes := []TokenClass{
		{Terminals: []int{1}, Kind: ptr(lexer.Word)},
		{Terminals: []int{2}, Text: ptr("foo"), Prev1: &semi},
		{Terminals: []int{3, 4}, Text: ptr("bar")},
		{Terminals: []int{5}, Text: ptr("baz"), Next: &semi},
		{Terminals: []int{6}, Prefix: ptr(byte('$'))},
	}
	want := mustCompile(t, classes...).Text()
	if len(want) == 0 {
		t.Fatal("empty dump")
	}
	for range 20 {
		rev := slices.Clone(classes)
		slices.Reverse(rev)
		if got := mustCompile(t, rev...).Text(); !bytes.Equal(got, want) {
			t.Fatalf("declaration order changed the table:\n%s\nvs\n%s", got, want)
		}
	}
}

func TestTokenSetZeroRefused(t *testing.T) {
	tab := &Tables{}
	for name, sets := range map[string][][]int{"none": nil, "non-empty": {{1}, {}}} {
		if err := tab.FillTerminalSets(sets); err == nil {
			t.Errorf("%s: set 0 accepted", name)
		}
	}
	tt := mustCompile(t, TokenClass{Terminals: []int{1}, Text: ptr("a")})
	if err := tab.FillTerminalSets(tt.Sets); err != nil || len(tab.TerminalSets) != len(tt.Sets) {
		t.Errorf("compiled sets refused or not stored: %v", err)
	}
	if _, err := CompileTokens([]TokenClass{{Text: ptr("a")}}); err == nil {
		t.Error("a class with no terminals accepted")
	}
}

func TestTokenOtherWordNeighbour(t *testing.T) {
	tt := mustCompile(t,
		TokenClass{Terminals: []int{1}, Kind: ptr(lexer.Word), Prev1: &Neighbour{Kind: lexer.Word, Text: OtherWord}},
		TokenClass{Terminals: []int{2}, Text: ptr("use")})
	if got := tt.set(word("b"), none, ClassOf(word("a"), none), none); !slices.Equal(got, []int{1}) {
		t.Errorf("a word after an unnamed word admits %v, want [1]", got)
	}
	if got := tt.set(word("b"), none, ClassOf(word("use"), none), none); len(got) != 0 {
		t.Errorf("a word after a word some class names admits %v, want none", got)
	}
	if got := tt.set(word("b"), none, ClassOf(op("+"), none), none); len(got) != 0 {
		t.Errorf("a word after an operator admits %v, want none", got)
	}
}

// NotNext keeps a class off the keys whose next token is that class, so two
// classes on one word are told apart by the next token, not unioned.
func TestTokenNotNext(t *testing.T) {
	paren, semi := ClassOf(op("("), none), ClassOf(op(";"), none)
	tt := mustCompile(t,
		TokenClass{Terminals: []int{1}, Text: ptr("ref"), Next: &paren},
		TokenClass{Terminals: []int{2}, Text: ptr("ref"), NotNext: []Neighbour{paren}})
	for _, c := range []struct {
		next Neighbour
		want []int
	}{{paren, []int{1}}, {semi, []int{2}}, {Boundary, []int{2}}} {
		if got := tt.set(word("ref"), none, none, c.next); !slices.Equal(got, c.want) {
			t.Errorf("ref before %+v = %v, want %v", c.next, got, c.want)
		}
	}
	if got := tt.set(word("other"), none, none, semi); len(got) != 0 {
		t.Errorf("other = %v, want the empty set", got)
	}
}

func TestTokenNotText(t *testing.T) {
	tt := mustCompile(t,
		TokenClass{Terminals: []int{1}, Kind: ptr(lexer.Word), NotText: []string{"if", "not"}},
		TokenClass{Terminals: []int{2}, Text: ptr("not")})
	for text, want := range map[string][]int{"foo": {1}, "if": nil, "not": {2}} {
		if got := tt.set(word(text), none, none, none); !slices.Equal(got, want) && len(got)+len(want) > 0 {
			t.Errorf("%s = %v, want %v", text, got, want)
		}
	}
}

func TestTokenNotPrev1(t *testing.T) {
	shift, semi := ClassOf(word("shift"), none), ClassOf(op(";"), none)
	tt := mustCompile(t,
		TokenClass{Terminals: []int{1}, Text: ptr("x"), NotPrev1: []Neighbour{shift}},
		TokenClass{Terminals: []int{2}, Text: ptr("x")})
	for _, c := range []struct {
		prev Neighbour
		want []int
	}{{shift, []int{2}}, {semi, []int{1, 2}}, {Boundary, []int{1, 2}}} {
		if got := tt.set(word("x"), none, c.prev, none); !slices.Equal(got, c.want) {
			t.Errorf("x after %+v = %v, want %v", c.prev, got, c.want)
		}
	}
}

func TestTokenNotPrev2(t *testing.T) {
	my, semi := ClassOf(word("my"), none), ClassOf(op(";"), none)
	tt := mustCompile(t,
		TokenClass{Terminals: []int{1}, Text: ptr("x"), NotPrev2: []Neighbour{my}},
		TokenClass{Terminals: []int{2}, Text: ptr("x")})
	for _, c := range []struct {
		prev2 Neighbour
		want  []int
	}{{my, []int{2}}, {semi, []int{1, 2}}, {Boundary, []int{1, 2}}} {
		if got := tt.set(word("x"), c.prev2, none, none); !slices.Equal(got, c.want) {
			t.Errorf("x after %+v is %v, want %v", c.prev2, got, c.want)
		}
	}
}

// NotNextIdent keeps a class off the keys whose next token starts with an identifier character: any
// word, a word some class names or none does, or a quote-like opening with a letter.
func TestTokenNotNextIdent(t *testing.T) {
	if_ := ClassOf(word("if"), none)
	tt := mustCompile(t,
		TokenClass{Terminals: []int{1}, Text: ptr(":"), NotNextIdent: true},
		TokenClass{Terminals: []int{2}, Text: ptr(":"), Prev1: &if_}) // names the word if as a neighbour
	for _, c := range []struct {
		name string
		next TokenView
		want []int
	}{
		{"unnamed word", word("shared"), nil},
		{"named word", word("if"), nil},
		{"word opening with ::", word("::foo"), []int{1}},
		{"word opening with _", word("_a"), nil},
		{"qw", TokenView{Kind: lexer.Quote, Text: "qw(a)", Prefix: 'w'}, nil},
		{"qq", TokenView{Kind: lexer.Quote, Text: "qq(a)", Prefix: 'Q'}, nil},
		{"s", TokenView{Kind: lexer.Quote, Text: "s/a/b/", Prefix: 's'}, nil},
		{"double quote", TokenView{Kind: lexer.Quote, Text: `"a"`, Prefix: '"'}, []int{1}},
		{"single quote", TokenView{Kind: lexer.Quote, Text: `'a'`, Prefix: '\''}, []int{1}},
		{"number", TokenView{Kind: lexer.Number, Text: "1"}, []int{1}},
		{"variable", TokenView{Kind: lexer.Variable, Text: "$z", Prefix: '$'}, []int{1}},
		{"operator", op("-"), []int{1}},
	} {
		next := tt.interned(c.next, none).Class
		if got := tt.Sets[tt.Lookup(tt.interned(op(":"), none), tt.Edge, tt.Edge, next)]; !slices.Equal(got, c.want) && len(got)+len(c.want) > 0 {
			t.Errorf("%s: : before it is %v, want %v", c.name, got, c.want)
		}
	}
	if got := tt.Sets[tt.Lookup(tt.interned(op(":"), none), tt.Edge, tt.Edge, tt.Edge)]; !slices.Equal(got, []int{1}) {
		t.Errorf("a : at the end of the file is %v, want [1]", got)
	}
}

// A neighbour says what it was, as a { says subscript or hash: a * is a glob sigil unless the token before
// it completes a term, and a word after ->, & or a glob's * is a name.
func TestInternGlobAndName(t *testing.T) {
	tt := mustCompile(t,
		TokenClass{Terminals: []int{1}, Kind: ptr(lexer.Variable), Completes: true},
		TokenClass{Terminals: []int{2}, Kind: ptr(lexer.Word), Text: ptr("time"), Completes: true},
		TokenClass{Terminals: []int{3}, Kind: ptr(lexer.Word), NotText: []string{"time", "print"}, Completes: true},
		TokenClass{Terminals: []int{4}, Kind: ptr(lexer.CloseBracket), Text: ptr("}"), OpensBlock: ptr(true), Completes: true},
		TokenClass{Terminals: []int{5}, Kind: ptr(lexer.Word), Text: ptr("print")})
	star, amp, arrow := op("*"), TokenView{Kind: lexer.FuncSigil, Text: "&"}, op("->")
	for name, c := range map[string]struct {
		src         []TokenView
		glob, names []bool // by token: Glob, Name
	}{
		"statement start":     {[]TokenView{star, word("print")}, []bool{true, false}, []bool{false, true}},
		"after a variable":    {[]TokenView{{Kind: lexer.Variable, Text: "$a"}, star, word("print")}, []bool{false, false, false}, []bool{false, false, false}},
		"after a bareword":    {[]TokenView{word("foo"), star}, []bool{false, false}, []bool{false, false}},
		"after time":          {[]TokenView{word("time"), star}, []bool{false, false}, []bool{false, false}},
		"after a keyword":     {[]TokenView{word("print"), star}, []bool{false, true}, []bool{false, false}},
		"after a block's }":   {[]TokenView{{Kind: lexer.CloseBracket, Text: "}", OpensBlock: true}, star}, []bool{false, true}, []bool{false, false}},
		"after a glob's name": {[]TokenView{star, word("print"), star}, []bool{true, false, false}, []bool{false, true, false}},
		"after a glob":        {[]TokenView{star, star, word("print")}, []bool{true, true, false}, []bool{false, false, true}},
		"after &":             {[]TokenView{amp, word("print"), star}, []bool{false, false, false}, []bool{false, true, false}},
		"after ->":            {[]TokenView{{Kind: lexer.Variable, Text: "$o"}, arrow, word("print"), star}, []bool{false, false, false, false}, []bool{false, false, true, false}},
	} {
		pred := Boundary
		for i, v := range c.src {
			var it Token
			tt.Intern(&v, &pred, &it)
			if it.N.Glob != c.glob[i] || it.N.Name != c.names[i] {
				t.Errorf("%s: token %d (%q) has Glob %v Name %v, want %v %v", name, i, v.Text, it.N.Glob, it.N.Name, c.glob[i], c.names[i])
			}
			pred = it.N
		}
	}
}

// NextName and NotNextName split the keys a sub's name follows from the rest: an identifier start, a
// quote-like opening with a letter, or a word opening with :: or ' is a name; anything else, and the end
// of the file, is not.
func TestTokenNextName(t *testing.T) {
	tt := mustCompile(t,
		TokenClass{Terminals: []int{1}, Text: ptr(":"), NextName: true},
		TokenClass{Terminals: []int{2}, Text: ptr(":"), NotNextName: true})
	for _, c := range []struct {
		name string
		next TokenView
		want []int
	}{
		{"word", word("shared"), []int{1}},
		{"keyword", word("if"), []int{1}},
		{"word opening with ::", word("::foo"), []int{1}},
		{"word opening with '", word("'foo"), []int{1}},
		{"qw", TokenView{Kind: lexer.Quote, Text: "qw(a)", Prefix: 'w'}, []int{1}},
		{"double quote", TokenView{Kind: lexer.Quote, Text: `"a"`, Prefix: '"'}, []int{2}},
		{"number", TokenView{Kind: lexer.Number, Text: "1"}, []int{2}},
		{"variable", TokenView{Kind: lexer.Variable, Text: "$z", Prefix: '$'}, []int{2}},
		{"operator", op("-"), []int{2}},
	} {
		next := tt.interned(c.next, none).Class
		if got := tt.Sets[tt.Lookup(tt.interned(op(":"), none), tt.Edge, tt.Edge, next)]; !slices.Equal(got, c.want) {
			t.Errorf("%s: : before it is %v, want %v", c.name, got, c.want)
		}
	}
	if got := tt.Sets[tt.Lookup(tt.interned(op(":"), none), tt.Edge, tt.Edge, tt.Edge)]; !slices.Equal(got, []int{2}) {
		t.Errorf("a : at the end of the file is %v, want [2]", got)
	}
}

// Lone cuts every other class of a word where the Lone class holds, all of what it names at once: a word between
// the braces with its Blanks bit set is the Lone class's alone, after a minus too, and a word with a neighbour or the
// bit changed keeps its classes.
func TestTokenLone(t *testing.T) {
	open := Neighbour{Kind: lexer.Operator, Text: "{", Brace: BraceSubscript}
	hash := Neighbour{Kind: lexer.Operator, Text: "{", Brace: BraceHash}
	closer := Neighbour{Kind: lexer.CloseBracket, Text: "}"}
	minus := Neighbour{Kind: lexer.Operator, Text: "-"}
	tt := mustCompile(t,
		TokenClass{Terminals: []int{1}, Kind: ptr(lexer.Word), Prev1: &open, Next: &closer, Blanks: ptr(true), Lone: true},
		TokenClass{Terminals: []int{1}, Kind: ptr(lexer.Word), Prev2: &open, Prev1: &minus, Next: &closer, Blanks: ptr(true), Lone: true},
		TokenClass{Terminals: []int{2}, Kind: ptr(lexer.Word), Text: ptr("time")},
		TokenClass{Terminals: []int{3}, Kind: ptr(lexer.Word), Text: ptr("time"), NotNext: []Neighbour{closer}},
		TokenClass{Terminals: []int{4}, Kind: ptr(lexer.Word), NotText: []string{"time"}},
		TokenClass{Terminals: []int{5}, Kind: ptr(lexer.Operator), Text: ptr("-")})
	for _, c := range []struct {
		name       string
		w          string
		blanks     bool
		p2, p1, nx Neighbour
		want       []int
	}{
		{"alone", "time", true, none, open, closer, []int{1}},
		{"another word alone", "foo", true, none, open, closer, []int{1}},
		{"after a minus", "time", true, open, minus, closer, []int{1}},
		{"a newline or a package part", "time", false, none, open, closer, []int{2}},
		{"a newline after a minus", "time", false, open, minus, closer, []int{2}},
		{"another word, no blanks", "foo", false, none, open, closer, []int{4}},
		{"a word with more after it", "time", true, none, open, none, []int{2, 3}},
		{"in an anonymous hash", "time", true, none, hash, closer, []int{2}},
		{"not after a brace", "time", true, none, none, closer, []int{2}},
		{"a minus without the brace", "time", true, none, minus, closer, []int{2}},
		{"another word, more after it", "foo", true, none, open, none, []int{4}},
	} {
		v := word(c.w)
		v.Blanks = c.blanks
		it := tt.interned(v, none)
		if got := tt.Sets[tt.Lookup(it, tt.ClassID(c.p2), tt.ClassID(c.p1), tt.ClassID(c.nx))]; !slices.Equal(got, c.want) {
			t.Errorf("%s: %v, want %v", c.name, got, c.want)
		}
	}
	if got := tt.set(op("-"), none, open, none); !slices.Equal(got, []int{5}) {
		t.Errorf("a minus after the brace is %v, want it kept: [5]", got)
	}
}
