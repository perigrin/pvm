// ABOUTME: Tests the token-class compiler on toy grammars: each key part, unions, edges and determinism.
// ABOUTME: Terminal ids are bare integers; no committed token table exists yet.

package grammar

import (
	"bytes"
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

// set is the terminals the lookup yields.
func (tt *TokenTable) set(t TokenView, p2, p1, n Neighbour) []int {
	return tt.Sets[tt.Lookup(t, p2, p1, n)]
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
	both, one := tt.Lookup(word("foo"), none, none, none), tt.Lookup(word("bar"), none, none, none)
	if !slices.Equal(tt.Sets[both], []int{1, 2}) || !slices.Equal(tt.Sets[one], []int{1, 3}) || both == one {
		t.Errorf("foo = %v (id %d), bar = %v (id %d), want [1 2] and [1 3] apart", tt.Sets[both], both, tt.Sets[one], one)
	}
	if tt.Lookup(word("foo"), none, none, none) != both {
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
	if n := testing.AllocsPerRun(1000, func() { tt.Lookup(tok, semi, semi, semi) }); n != 0 {
		t.Errorf("Lookup allocates %v times", n)
	}
	if got := tt.set(tok, semi, semi, semi); !slices.Equal(got, []int{1, 2}) {
		t.Errorf("the measured lookup yields %v, want [1 2]", got)
	}
}

func TestTokenUnmatchedKeyEmptySet(t *testing.T) {
	tt := mustCompile(t, TokenClass{Terminals: []int{1}, Text: ptr("foo")}, TokenClass{Terminals: []int{2}, Text: ptr("bar")})
	if id := tt.Lookup(word("baz"), none, none, none); id != 0 {
		t.Errorf("unmatched key id = %d, want exactly 0", id)
	}
	if len(tt.Sets[0]) != 0 {
		t.Errorf("set 0 = %v, want empty", tt.Sets[0])
	}
	foo, bar := tt.Lookup(word("foo"), none, none, none), tt.Lookup(word("bar"), none, none, none)
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
	a := tt.Lookup(word("x"), none, none, ClassOf(word("foo"), semi))
	b := tt.Lookup(word("x"), none, none, ClassOf(word("foo"), none))
	if a != b || !slices.Equal(tt.Sets[a], []int{3}) {
		t.Errorf("same-class neighbours keyed apart: %d vs %d", a, b)
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
