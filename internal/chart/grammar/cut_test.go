// ABOUTME: Tests the token cut on small grammars for each rewrite form, and on the vendored Guacamole grammar.
// ABOUTME: Holds the committed rewrite list equal to the cut's; -grammar.update-rewrites regenerates it.

package grammar

import (
	"bytes"
	"flag"
	"os"
	"slices"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/lexer"
)

var updateRewrites = flag.Bool("grammar.update-rewrites", false,
	"rewrite testdata/token-cut-rewrites.txt from the cut of the vendored Guacamole grammar")

const rewritesPath = "testdata/token-cut-rewrites.txt"

func cut(t *testing.T, src string, spec CutSpec) *CutGrammar {
	t.Helper()
	c, err := ApplyCut(read(t, src), spec)
	if err != nil {
		t.Fatalf("ApplyCut: %v", err)
	}
	return c
}

func cutErr(t *testing.T, src string, spec CutSpec) string {
	t.Helper()
	c, err := ApplyCut(read(t, src), spec)
	if err == nil {
		t.Fatalf("ApplyCut = %+v, want an error", c)
	}
	return err.Error()
}

func hasRewrite(t *testing.T, c *CutGrammar, want string) {
	t.Helper()
	if !slices.Contains(c.Rewrites, want) {
		t.Errorf("rewrites lack %q; have:\n%s", want, strings.Join(c.Rewrites, "\n"))
	}
}

// matches reports whether any predicate of terminal sym accepts the token.
func matches(c *CutGrammar, sym string, k lexer.Kind, text string) bool {
	for _, p := range c.Terminals[sym] {
		if p.Match(k, text) {
			return true
		}
	}
	return false
}

func TestCutCollapsesG0Rule(t *testing.T) {
	c := cut(t, "S ::= Num Arrow Times\nNum ~ [0-9]+\nArrow ~ '->'\nTimes ~ '*' | 'x'\n",
		CutSpec{Tokens: []TokenDef{{Name: "Num", Kind: "Number"}}})
	if !matches(c, "Num", lexer.Number, "12") || matches(c, "Num", lexer.Number, "1a") || matches(c, "Num", lexer.Number, "1 2") || matches(c, "Num", lexer.Word, "12") {
		t.Error("Num should accept exactly Number tokens of digits")
	}
	if !matches(c, "Times", lexer.Operator, "*") || !matches(c, "Times", lexer.Word, "x") || matches(c, "Times", lexer.Operator, "x") {
		t.Error("Times should be Operator '*' or Word 'x'")
	}
	hasRewrite(t, c, "collapse Num ~ [0-9]+ => Number ~ Num")
	hasRewrite(t, c, "collapse Arrow ~ '->' => Operator '->'")
	hasRewrite(t, c, "collapse Times ~ '*' | 'x' => Operator '*' | Word 'x'")
	if want := []Rule{{LHS: "S", RHS: []Symbol{name("Num"), name("Arrow"), name("Times")}, Line: 1, Decl: 1}}; !slices.EqualFunc(c.G1, want, ruleEqual) {
		t.Errorf("G1 = %+v, want %+v", c.G1, want)
	}
}

func TestCutCollapsesG1Rule(t *testing.T) {
	c := cut(t, "S ::= Str\nStr ::= Q Body Q | Q Q\nQ ~ [']\nBody ~ [a-z]+\n",
		CutSpec{Tokens: []TokenDef{{Name: "Str", Kind: "Quote"}}})
	if !matches(c, "Str", lexer.Quote, "'ab'") || !matches(c, "Str", lexer.Quote, "''") || matches(c, "Str", lexer.Quote, "'a") {
		t.Error("Str should accept the quoted strings its rules derive")
	}
	hasRewrite(t, c, "collapse Str ::= Q Body Q | Q Q => Quote ~ Str")
}

const straddle = `Program ::= VarScalar
VarScalar ::= SigilScalar VarIdentExpr ElemSeq0
ElemSeq0 ::= Element*
Element ::= LBracket RBracket
VarIdentExpr ::= VarIdent | VarIdentExpr PackageSep VarIdent
SigilScalar ~ '$'
VarIdent ~ [a-z]+
PackageSep ~ '::'
LBracket ~ '['
RBracket ~ ']'
`

func TestCutSplitsStraddlingRule(t *testing.T) {
	c := cut(t, straddle, CutSpec{
		Tokens: []TokenDef{{Name: "ScalarVar", Kind: "Variable", Def: "SigilScalar VarIdentExpr"}},
		Splits: []Split{{From: "SigilScalar VarIdentExpr", To: "ScalarVar"}},
	})
	hasRewrite(t, c, "split VarScalar ::= SigilScalar VarIdentExpr ElemSeq0 => VarScalar ::= ScalarVar ElemSeq0")
	hasRewrite(t, c, "token ScalarVar => Variable ~ SigilScalar VarIdentExpr")
	hasRewrite(t, c, "absorb VarIdentExpr ::= VarIdent | VarIdentExpr PackageSep VarIdent")
	// A G1 sequence admits the whitespace pvm folds into one token ("$ x");
	// a G0 rule, such as VarIdent, does not ("$x y").
	for text, want := range map[string]bool{"$x": true, "$x::y": true, "$ x": true, "$x::": false, "@x": false, "$": false, "$x y": false} {
		if got := matches(c, "ScalarVar", lexer.Variable, text); got != want {
			t.Errorf("ScalarVar on %q = %v, want %v", text, got, want)
		}
	}
	for _, r := range c.G1 {
		if r.LHS == "VarIdentExpr" {
			t.Errorf("absorbed rule kept: %+v", r)
		}
	}
}

// keywordGrammar has a keyword lexeme with priority, a call name and a
// method name, in the shape Guacamole gives them.
const keywordGrammar = `S ::= Call | Method | OpKeywordExitExpr
Call ::= SubNameCallExpr LParen RParen
Method ::= Arrow SubNameExpr
OpKeywordExitExpr ::= OpKeywordExit
SubNameCallExpr ::= SubNameNonQLike | SubNameCallExpr PackageSep SubNameNonQLike
SubNameExpr ::= SubName | SubNameExpr PackageSep SubName
SubNameNonQLike ~ [a-zA-Z_]+
SubName ~ [a-zA-Z_]+
PackageSep ~ '::'
LParen ~ '('
RParen ~ ')'
Arrow ~ '->'
OpKeywordExit ~ 'exit'
:lexeme ~ OpKeywordExit priority => 1
`

func TestCutKeywordPriority(t *testing.T) {
	c := cut(t, keywordGrammar, CutSpec{Tokens: []TokenDef{
		{Name: "SubNameCallExpr", Kind: "Word", Except: ":priority"},
		{Name: "SubNameExpr", Kind: "Word"},
	}})
	hasRewrite(t, c, "priority OpKeywordExit 1 => except :priority")
	hasRewrite(t, c, "collapse SubNameCallExpr ::= SubNameNonQLike | SubNameCallExpr PackageSep SubNameNonQLike => Word ~ SubNameCallExpr except :priority")
	checkKeywordPriority(t, c, "exit", "POSIX::exit")

	// The same holds of Guacamole's own cut.
	g := guacamoleCut(t)
	checkKeywordPriority(t, g, "print", "POSIX::exit")
	checkKeywordPriority(t, g, "exit", "Time::HiRes::time")
}

func checkKeywordPriority(t *testing.T, c *CutGrammar, keyword, qualified string) {
	t.Helper()
	if !matches(c, "SubNameExpr", lexer.Word, keyword) {
		t.Errorf("SubNameExpr refuses keyword %q, which may name a method or a definition", keyword)
	}
	if matches(c, "SubNameCallExpr", lexer.Word, keyword) {
		t.Errorf("SubNameCallExpr admits unqualified keyword %q: a call to it would have two derivations", keyword)
	}
	if !matches(c, "SubNameCallExpr", lexer.Word, qualified) {
		t.Errorf("SubNameCallExpr refuses qualified %q", qualified)
	}
	if !matches(c, "SubNameCallExpr", lexer.Word, "foo") {
		t.Error("SubNameCallExpr refuses foo")
	}
}

func TestCutDropsTrivia(t *testing.T) {
	c := cut(t, "S ~ 'a'\nT ::= S\n:discard ~ ws\nws ~ [\\s]+\n", CutSpec{Trivia: []TriviaDef{{Kind: "Whitespace", Discard: "ws"}, {Kind: "Pod"}}})
	if want := []lexer.Kind{lexer.Whitespace, lexer.Pod}; !slices.Equal(c.Trivia, want) {
		t.Errorf("Trivia = %v, want %v", c.Trivia, want)
	}
	hasRewrite(t, c, "trivia Whitespace => :discard ~ ws")
	hasRewrite(t, c, "trivia Pod")
	if e := cutErr(t, "S ~ 'a'\nT ::= S\n:discard ~ ws\nws ~ [\\s]+\n", CutSpec{}); !strings.Contains(e, ":discard ~ ws") {
		t.Errorf("an unaccounted :discard: error %q does not name it", e)
	}
}

func TestCutRejectsLeftoverG0(t *testing.T) {
	e := cutErr(t, "S ::= A B\nA ~ 'a'\nB ~ [b]+\n", CutSpec{})
	if !strings.Contains(e, "S ::= A B") || !strings.Contains(e, "B") {
		t.Errorf("error %q does not name the rule and its G0 symbol", e)
	}
	e = cutErr(t, "S ::= 'a' B\nB ~ 'b'\n", CutSpec{})
	if !strings.Contains(e, "S ::= 'a' B") {
		t.Errorf("a literal left in G1: error %q does not name the rule", e)
	}
}

func TestCutRejectsUnknownKind(t *testing.T) {
	e := cutErr(t, "S ::= B\nB ~ [b]+\n", CutSpec{Tokens: []TokenDef{{Name: "B", Kind: "Wurd"}}})
	if !strings.Contains(e, "B") || !strings.Contains(e, `"Wurd"`) {
		t.Errorf("error %q does not name the token and the kind", e)
	}
	e = cutErr(t, "S ~ 'a'\nT ::= S\n", CutSpec{Trivia: []TriviaDef{{Kind: "Pud"}}})
	if !strings.Contains(e, `"Pud"`) {
		t.Errorf("error %q does not name the trivia kind", e)
	}
}

func TestCutRejectsBadPriority(t *testing.T) {
	// The reader accepts a priority on a symbol no G0 rule defines; Marpa
	// does not, and neither does the cut.
	e := cutErr(t, "S ::= A\nA ~ 'a'\n:lexeme ~ Nope priority => 1\n", CutSpec{})
	if !strings.Contains(e, "Nope") {
		t.Errorf("error %q does not name the undefined :lexeme", e)
	}
	// A priority on a lexeme that is not all words needs a note.
	e = cutErr(t, "S ::= A\nA ~ '+'\n:lexeme ~ A priority => 1\n", CutSpec{})
	if !strings.Contains(e, "A") {
		t.Errorf("error %q does not name the unaccounted priority", e)
	}
	c := cut(t, "S ::= A\nA ~ '+'\n:lexeme ~ A priority => 1\n", CutSpec{Priorities: []PriorityNote{{"A", "not encoded"}}})
	hasRewrite(t, c, "priority A 1 => not encoded")
}

func guacamoleCut(t *testing.T) *CutGrammar {
	t.Helper()
	src, err := os.ReadFile(guacamolePath)
	if err != nil {
		t.Fatal(err)
	}
	g, err := ReadSLIF("guacamole.slif", src)
	if err != nil {
		t.Fatal(err)
	}
	c, err := ApplyCut(g, GuacamoleCut)
	if err != nil {
		t.Fatalf("ApplyCut(Guacamole): %v", err)
	}
	return c
}

// TestCutGuacamole applies the reviewed cut to the vendored grammar. No G1
// rule may name a G0 symbol afterwards, and the rewrites must equal the
// committed list.
func TestCutGuacamole(t *testing.T) {
	c := guacamoleCut(t)
	nonterminal := map[string]bool{}
	for _, r := range c.G1 {
		nonterminal[r.LHS] = true
	}
	for _, r := range c.G1 {
		for _, s := range r.RHS {
			if s.Kind != Name || !nonterminal[s.Text] && c.Terminals[s.Text] == nil {
				t.Errorf("%s (line %d) names %q, neither a G1 nonterminal nor a terminal", r.LHS, r.Line, s.Text)
			}
		}
	}

	got := []byte(strings.Join(c.Rewrites, "\n") + "\n")
	if *updateRewrites {
		if err := os.WriteFile(rewritesPath, got, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(rewritesPath)
	if err != nil {
		t.Fatalf("%v (regenerate with go test ./internal/chart/grammar/ -run TestCutGuacamole -grammar.update-rewrites)", err)
	}
	if !bytes.Equal(got, want) {
		gl, wl := lines(got), lines(want)
		for i := 0; ; i++ {
			switch {
			case i >= len(gl):
				t.Fatalf("%s: line %d lists %q, which the cut no longer makes", rewritesPath, i+1, wl[i])
			case i >= len(wl):
				t.Fatalf("%s: the cut makes %q, which is not listed (after line %d)", rewritesPath, gl[i], i)
			case gl[i] != wl[i]:
				t.Fatalf("%s: line %d lists %q, but the cut makes %q", rewritesPath, i+1, wl[i], gl[i])
			}
		}
	}
}

func lines(b []byte) []string { return strings.Split(strings.TrimSuffix(string(b), "\n"), "\n") }

// TestCutLiteralKinds checks the kind the cut gives each literal against the
// production lexer: each must lex as one token of that kind and text, after a
// term or at the start of a statement.
func TestCutLiteralKinds(t *testing.T) {
	c := guacamoleCut(t)
	seen := map[string]bool{}
	for _, preds := range c.Terminals {
		for _, p := range preds {
			text, ok := literalText(p.Def)
			if !ok || p.Except != "" || seen[p.Kind.String()+text] {
				continue
			}
			seen[p.Kind.String()+text] = true
			if !lexesAs(p.Kind, text) {
				t.Errorf("%s %q: the lexer gives neither $a %[2]s $b nor %[2]s $b a single %[1]s token %[2]q", p.Kind, text)
			}
		}
	}
}

func lexesAs(k lexer.Kind, text string) bool {
	for _, src := range []string{"$a " + text + " $b", text + " $b", "f(" + text + ")"} {
		for _, tok := range lexer.Tokenize([]byte(src)) {
			if tok.Kind == k && src[tok.Start:tok.End] == text {
				return true
			}
		}
	}
	return false
}

func ruleEqual(a, b Rule) bool {
	return a.LHS == b.LHS && slices.Equal(a.RHS, b.RHS) && a.Seq == b.Seq
}
