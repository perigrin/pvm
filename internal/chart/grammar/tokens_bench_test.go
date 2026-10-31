// ABOUTME: Benchmarks the token-table lookup against lexer.Tokenize on std5k.pl, per significant token.
// ABOUTME: The ratio is what chart/lex inherits as its share of lex-overhead (spec §9.3).

package grammar

import (
	"os"
	"testing"

	"tamarou.com/pvm/internal/lexer"
)

// benchTable is shaped like a real one: a text-keyed class per word of the
// file, and a few generic classes refined by their neighbours.
func benchTable(tb testing.TB, src []byte, toks []lexer.Token) *TokenTable {
	semi, lbrace, arrow := ClassOf(op(";"), none), ClassOf(op("{"), none), ClassOf(op("->"), none)
	classes := []TokenClass{
		{Terminals: []int{1}, Kind: ptr(lexer.Word), Prev1: &semi},
		{Terminals: []int{2}, Kind: ptr(lexer.Word), Prev2: &lbrace, Next: &semi},
		{Terminals: []int{3}, Kind: ptr(lexer.Word), Next: &lbrace},
		{Terminals: []int{4}, Kind: ptr(lexer.Word), Prev1: &arrow},
		{Terminals: []int{5}, Kind: ptr(lexer.Variable), Prev1: &lbrace, Next: &arrow},
		{Terminals: []int{6}, Kind: ptr(lexer.Operator), Prev1: &semi},
	}
	seen := map[string]bool{}
	for _, t := range toks {
		if w := string(src[t.Start:t.End]); t.Kind == lexer.Word && !seen[w] && len(seen) < 300 {
			seen[w] = true
			classes = append(classes, TokenClass{Terminals: []int{7 + len(seen)%9}, Text: ptr(w)})
		}
	}
	tt, err := CompileTokens(classes)
	if err != nil {
		tb.Fatal(err)
	}
	return tt
}

func benchInput(tb testing.TB) ([]byte, []lexer.Token, []lexer.Token) {
	src, err := os.ReadFile("../testdata/std5k.pl")
	if err != nil {
		tb.Fatal(err)
	}
	all := lexer.Tokenize(src)
	var sig []lexer.Token
	for _, t := range all {
		if t.Kind != lexer.Whitespace && t.Kind != lexer.Comment {
			sig = append(sig, t)
		}
	}
	return src, all, sig
}

func BenchmarkTokenLookup(b *testing.B) {
	src, all, sig := benchInput(b)
	tt := benchTable(b, src, all)
	view := func(t lexer.Token) TokenView {
		return TokenView{Kind: t.Kind, Text: string(src[t.Start:t.End]), OpensBlock: t.OpensBlock, TermFollows: t.TermFollows}
	}
	views := make([]TokenView, len(sig))
	for i, t := range sig {
		views[i] = view(t)
	}
	toks := make([]Token, len(views))
	b.ResetTimer()
	for range b.N {
		// Per token: intern it (one text hash), then look it up.
		pred := Boundary
		for i, v := range views {
			toks[i] = tt.Intern(v, pred)
			pred = toks[i].N
		}
		p2, p1 := tt.Edge, tt.Edge
		for i, t := range toks {
			next := tt.Edge
			if i+1 < len(toks) {
				next = toks[i+1].Class
			}
			tt.Lookup(t, p2, p1, next)
			p2, p1 = p1, t.Class
		}
	}
	b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N)/float64(len(views)), "ns/token")
	b.ReportMetric(float64(tt.Nodes()), "nodes")
}

// BenchmarkTokenizePerToken is the reference: lexer.Tokenize's cost per significant token.
func BenchmarkTokenizePerToken(b *testing.B) {
	src, _, sig := benchInput(b)
	for range b.N {
		lexer.Tokenize(src)
	}
	b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N)/float64(len(sig)), "ns/token")
}
