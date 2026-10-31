// ABOUTME: Tests the thin chart/lex: trivia to the next significant token, heredoc bodies to their openers, one lookup per token.
// ABOUTME: Every token is mapped (an unmatched key to the empty set), the first unmapped byte is kept, and the lookup allocates nothing.

package lex

import (
	"slices"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/lexer"
)

func ptr[T any](v T) *T { return &v }

// toyTable maps a `$` variable to terminal 1, a semicolon to 2, `=` to 3 and
// a number to 4; anything else (a word, a string) matches no class.
func toyTable(t testing.TB) *grammar.TokenTable {
	t.Helper()
	tt, err := grammar.CompileTokens([]grammar.TokenClass{
		{Terminals: []int{1}, Kind: ptr(lexer.Variable), Prefix: ptr(byte('$'))},
		{Terminals: []int{2}, Kind: ptr(lexer.Semicolon)},
		{Terminals: []int{3}, Kind: ptr(lexer.Operator), Text: ptr("=")},
		{Terminals: []int{4}, Kind: ptr(lexer.Number)},
	})
	if err != nil {
		t.Fatal(err)
	}
	return tt
}

// texts is each token's text, and its lead's, as "lead|token".
func texts(src string, s *Stream) []string {
	var out []string
	for _, tok := range s.Tokens {
		var lead strings.Builder
		for _, l := range tok.Lead {
			lead.WriteString(src[l.Start:l.End])
		}
		out = append(out, lead.String()+"|"+src[tok.Start:tok.End])
	}
	return out
}

func span(src string, ts []lexer.Token) string {
	var b strings.Builder
	for _, t := range ts {
		b.WriteString(src[t.Start:t.End])
	}
	return b.String()
}

func TestLexTriviaToNextToken(t *testing.T) {
	src := "  # c\n$x =\n=pod\n\n=cut\n 1;"
	s, _ := Lex([]byte(src), toyTable(t))
	want := []string{"  # c\n|$x", " |=", "\n=pod\n\n=cut\n |1", "|;"}
	if got := texts(src, s); !slices.Equal(got, want) {
		t.Errorf("tokens %q, want %q", got, want)
	}
	if got := string(s.Text([]byte(src))); got != src {
		t.Errorf("Text = %q, want the source", got)
	}
}

func TestLexHeredocBodyToOpener(t *testing.T) {
	src := "$x = <<A;\nbody\nA\n$y;"
	s, _ := Lex([]byte(src), toyTable(t))
	var opener *Token
	for i := range s.Tokens {
		if s.Tokens[i].Kind == lexer.HeredocOpen {
			opener = &s.Tokens[i]
		}
		if s.Tokens[i].Kind == lexer.HeredocBody {
			t.Fatalf("the body is a token of the stream: %q", texts(src, s))
		}
	}
	if opener == nil {
		t.Fatalf("no heredoc opener in %q", texts(src, s))
	}
	if got := span(src, opener.Body); !strings.HasSuffix(got, "body\nA\n") {
		t.Errorf("the opener's body is %q, want it to end with the heredoc's body", got)
	}
	if got := string(s.Text([]byte(src))); got != src {
		t.Errorf("Text = %q, want the source", got)
	}
}

func TestLexMapsEveryToken(t *testing.T) {
	src := "$x = foo; $y = 1;"
	s, first := Lex([]byte(src), toyTable(t))
	want := []int{1, 3, 0, 2, 1, 3, 4, 2} // foo matches no class: the empty set, and the stream goes on
	var got []int
	for _, tok := range s.Tokens {
		got = append(got, tok.TerminalSet)
	}
	sets := toyTable(t).Sets
	for i, id := range got {
		if id == 0 {
			continue
		}
		if !slices.Equal(sets[id], []int{want[i]}) {
			t.Errorf("token %d (%q) has set %v, want [%d]", i, src[s.Tokens[i].Start:s.Tokens[i].End], sets[id], want[i])
		}
	}
	if len(got) != len(want) || got[2] != 0 {
		t.Fatalf("sets %v: want %d tokens, the third (foo) the empty set 0", got, len(want))
	}
	if first != strings.Index(src, "foo") {
		t.Errorf("first unmapped byte %d, want %d", first, strings.Index(src, "foo"))
	}
}

// The per-token work, interning a token's key and looking it up, allocates nothing.
func TestLexLookupAllocationFree(t *testing.T) {
	src := "$x = 1; $y = foo; # c\n$z = 2;"
	tt := toyTable(t)
	text, raw := src, lexer.Tokenize([]byte(src))
	if a := testing.AllocsPerRun(100, func() {
		pred, p2, p1 := grammar.Boundary, tt.Edge, tt.Edge
		for _, r := range raw {
			if lexer.IsTrivia(r.Kind) {
				continue
			}
			it := intern(tt, text, r, pred)
			tt.Lookup(it, p2, p1, tt.Edge)
			pred, p2, p1 = it.N, p1, it.Class
		}
	}); a != 0 {
		t.Errorf("the per-token lookup allocates %v times per run, want 0", a)
	}
}

func TestLexFirstUnmappedByteKept(t *testing.T) {
	tt := toyTable(t)
	for _, c := range []struct {
		src  string
		want int
	}{
		{"$x = foo; $y = bar;", 5}, // two unmapped: the first's byte, not the second's
		{"foo; $x = 1;", 0},        // unmapped at byte 0 is not "nothing unmapped"
		{"$x = 1;", -1},            // fully mapped: none
		{"", -1},
	} {
		if _, got := Lex([]byte(c.src), tt); got != c.want {
			t.Errorf("%q: first unmapped byte %d, want %d", c.src, got, c.want)
		}
	}
}

func TestLexTrailingTrivia(t *testing.T) {
	tt := toyTable(t)
	for _, src := range []string{"", "  \n\t", "$x = 1; # trailing\n\n"} {
		s, _ := Lex([]byte(src), tt)
		if got := string(s.Text([]byte(src))); got != src {
			t.Errorf("%q: Text = %q, trivia lost", src, got)
		}
	}
	src := "$x = 1; # trailing\n"
	s, _ := Lex([]byte(src), tt)
	if got := span(src, s.Tail); got != " # trailing\n" {
		t.Errorf("tail %q, want the trivia after the last token", got)
	}
	src = "# lead\n$x;"
	s, _ = Lex([]byte(src), tt)
	if got := texts(src, s); !slices.Equal(got, []string{"# lead\n|$x", "|;"}) {
		t.Errorf("tokens %q: trivia before a token must attach to it", got)
	}
}

func TestLexTwoHeredocsOneLine(t *testing.T) {
	src := "$x = <<A, <<B; # c\nfirst\nA\nsecond\nB\n$y;"
	s, _ := Lex([]byte(src), toyTable(t))
	var bodies []string
	for _, tok := range s.Tokens {
		if tok.Kind == lexer.HeredocOpen {
			bodies = append(bodies, span(src, tok.Body))
		}
		if tok.Kind == lexer.Variable && src[tok.Start:tok.End] == "$y" {
			if lead := span(src, tok.Lead); strings.Contains(lead, "first") || strings.Contains(lead, "second") {
				t.Errorf("a heredoc body is attached to the token after the opener's line: lead %q", lead)
			}
			if lead := span(src, tok.Lead); lead != " # c\n" {
				t.Errorf("$y's lead is %q, want the trivia before it, %q", lead, " # c\n")
			}
		}
	}
	if len(bodies) != 2 || bodies[0] != "first\nA\n" || bodies[1] != "second\nB\n" {
		t.Errorf("bodies %q, want each opener its own, in order", bodies)
	}
	if got := string(s.Text([]byte(src))); got != src {
		t.Errorf("Text = %q, want the source", got)
	}
}
