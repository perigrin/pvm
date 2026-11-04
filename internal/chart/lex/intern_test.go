// ABOUTME: Holds every corpus token's terminal-set id to a reference that matches each token class directly.
// ABOUTME: Benchmarks Intern against the committed token table per significant token of std5k.pl, beside lexer.Tokenize.

package lex

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
	"unsafe"

	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/lexer"
)

// committed is the token table the generator builds from perly.go, and the
// classes it was compiled from.
func committed(tb testing.TB) (*grammar.TokenTable, []grammar.TokenClass) {
	tb.Helper()
	g, ref, err := grammar.Committed()
	if err != nil {
		tb.Fatal(err)
	}
	tab, err := grammar.Generate(g, ref)
	if err != nil {
		tb.Fatal(err)
	}
	classes, err := g.ResolveClasses(tab)
	if err != nil {
		tb.Fatal(err)
	}
	return tab.Tokens, classes
}

// refClassOf is ClassOf as it stood before Intern was made cheaper, kept
// here so the reference shares no code with what it checks.
func refClassOf(t grammar.TokenView, pred grammar.Neighbour) grammar.Neighbour {
	c := grammar.Neighbour{Kind: t.Kind, OpensBlock: t.OpensBlock, TermFollows: t.TermFollows}
	switch t.Kind {
	case lexer.Word, lexer.Operator, lexer.CloseBracket:
		c.Text = t.Text
	}
	if t.Kind == lexer.Quote && (t.Prefix >= 'a' && t.Prefix <= 'z' || t.Prefix == 'Q') {
		c.Text = grammar.IdentQuote
	}
	if t.Kind == lexer.Operator && t.Text == "{" {
		switch {
		case t.OpensBlock, pred.Kind == lexer.DerefSigil:
			c.Brace = grammar.BraceBlock
		case pred.Kind == lexer.Variable, pred.Kind == lexer.CloseBracket && !pred.OpensBlock,
			pred.Kind == lexer.Operator && pred.Text == "->":
			c.Brace = grammar.BraceSubscript
		default:
			c.Brace = grammar.BraceHash
		}
	}
	return c
}

// refIDs is the terminal-set id of each significant token of src as the
// table encodes it: the union of the terminals of every class whose given
// parts all match the token's key, found in tt.Sets. It walks no table and
// hashes nothing, so it checks Intern and Lookup rather than restating them.
func refIDs(tb testing.TB, src []byte, tt *grammar.TokenTable, classes []grammar.TokenClass) []int {
	type tok struct {
		v grammar.TokenView
		n grammar.Neighbour
	}
	var toks []tok
	pred, opens := grammar.Boundary, 0
	for _, t := range lexer.Tokenize(src) {
		switch {
		case lexer.IsTrivia(t.Kind):
			continue
		case t.Kind == lexer.HeredocBody && opens > 0:
			opens--
			continue
		case t.Kind == lexer.HeredocOpen:
			opens++
		}
		tx := string(src[t.Start:t.End])
		v := grammar.TokenView{Kind: t.Kind, Text: tx, OpensBlock: t.OpensBlock, TermFollows: t.TermFollows, Prefix: prefix(t.Kind, tx)}
		pred = refClassOf(v, pred)
		toks = append(toks, tok{v, pred})
	}
	nb := func(i int) grammar.Neighbour {
		if i < 0 || i >= len(toks) {
			return grammar.Boundary
		}
		return toks[i].n
	}
	// A neighbour of grammar.OtherWord is a word, flags as given, that no class names.
	named := map[string]bool{}
	for _, c := range classes {
		if c.Text != nil {
			named[*c.Text] = true
		}
		for _, n := range append(append([]*grammar.Neighbour{c.Prev2, c.Prev1, c.Next, c.NotNext}, ptrs(c.NotPrev1)...), ptrs(c.NotPrev2)...) {
			if n != nil {
				named[n.Text] = true
			}
		}
		for _, x := range c.NotText {
			named[x] = true
		}
	}
	is := func(want *grammar.Neighbour, got grammar.Neighbour) bool {
		if want != nil && want.Text == grammar.OtherWord {
			w := *want
			w.Text = got.Text
			return w == got && !named[got.Text] && (got.Kind != lexer.Word || got.Text[0] == '_' || got.Text[0]|0x20 >= 'a' && got.Text[0]|0x20 <= 'z' || got.Text[0] >= 0x80)
		}
		return want == nil || *want == got
	}
	// ident is whether toks[i] starts with an identifier character: a word opening with a letter or _, or a quote-like opening with a letter.
	ident := func(i int) bool {
		return i >= 0 && i < len(toks) && (toks[i].v.Kind == lexer.Word && (toks[i].v.Text[0] == '_' || toks[i].v.Text[0]|0x20 >= 'a' && toks[i].v.Text[0]|0x20 <= 'z' || toks[i].v.Text[0] >= 0x80) ||
			toks[i].v.Kind == lexer.Quote && (toks[i].v.Prefix >= 'a' && toks[i].v.Prefix <= 'z' || toks[i].v.Prefix == 'Q'))
	}
	ids := make([]int, len(toks))
	for i, t := range toks {
		var set []int
		for _, c := range classes {
			if (c.Kind == nil || *c.Kind == t.v.Kind) && (c.Text == nil || *c.Text == t.v.Text) &&
				(c.OpensBlock == nil || *c.OpensBlock == t.v.OpensBlock) && (c.TermFollows == nil || *c.TermFollows == t.v.TermFollows) &&
				(c.Prefix == nil || *c.Prefix == t.v.Prefix) && is(c.Prev2, nb(i-2)) && is(c.Prev1, nb(i-1)) && is(c.Next, nb(i+1)) &&
				!slices.Contains(c.NotText, t.v.Text) && !slices.Contains(c.NotPrev1, nb(i-1)) && !slices.Contains(c.NotPrev2, nb(i-2)) && !(c.NotNextIdent && ident(i+1)) && (c.NotNext == nil || *c.NotNext != nb(i+1)) {
				set = append(set, c.Terminals...)
			}
		}
		slices.Sort(set)
		set = slices.Compact(set)
		id, ok := slices.BinarySearchFunc(tt.Sets, set, slices.Compare)
		if !ok {
			tb.Fatalf("token %d (%q): the union %v is not one of the table's sets", i, t.v.Text, set)
		}
		ids[i] = id
	}
	return ids
}

// corpusFiles is every waypoint-1 corpus file the ratchet reads: the
// in-tree ones always, T1's when T1_CORPUS names it.
func corpusFiles(t *testing.T) []string {
	var files []string
	globs := []string{"../testdata/guacamole/parses/*.pl", "../testdata/guacamole/parsent/*.pl", "../testdata/std5k.pl", "../../psc/testdata/latency/edit5k.pl"}
	if t1 := os.Getenv("T1_CORPUS"); t1 != "" {
		globs = append(globs, filepath.Join(t1, "*.t"))
	} else {
		t.Log("T1_CORPUS is unset: the in-tree corpus only")
	}
	for _, g := range globs {
		m, err := filepath.Glob(g)
		if err != nil || len(m) == 0 {
			t.Fatalf("no corpus file matches %s", g)
		}
		files = append(files, m...)
	}
	return files
}

// Every significant token of every corpus file gets the terminal-set id the
// reference gives it, so a cheaper Intern changes no id.
func TestInternIdentityOverCorpus(t *testing.T) {
	tt, classes := committed(t)
	files := corpusFiles(t)
	tokens := 0
	for _, f := range files {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		s, _ := Lex(src, tt)
		want := refIDs(t, src, tt, classes)
		got := make([]int, len(s.Tokens))
		for i, tok := range s.Tokens {
			got[i] = tok.TerminalSet
		}
		if i := slices.Compare(got, want); i != 0 {
			for j := range min(len(got), len(want)) {
				if got[j] != want[j] {
					r := s.Raw[s.Tokens[j].Index]
					t.Errorf("%s: token %d (%q at byte %d) has set %d, the reference %d", f, j, src[r.Start:r.End], r.Start, got[j], want[j])
					break
				}
			}
			if len(got) != len(want) {
				t.Errorf("%s: %d tokens, the reference %d", f, len(got), len(want))
			}
		}
		tokens += len(want)
	}
	t.Logf("%d files, %d tokens", len(files), tokens)
}

// colonForms are the : after a declarator and a variable with each kind of token after it, and the : after other things.
const colonForms = "$c ? my $x : shared; $c ? my $x :shared; $c ? our @x : if; $c ? state $x : qw(a); $c ? my $x : q(a); " +
	"$c ? my $x : my $z; $c ? my $x : 1; $c ? my $x : $z; $c ? my $x : -1; $c ? my $x : \"s\"; $c ? my $x : \\$z; " +
	"$c ? my $x : 'a'; $c ? my $x : s/a/b/; $c ? my $x : tr/a/b/; $c ? my $x : y/a/b/; $c ? my $x : m/a/; $c ? my $x : qq(a); " +
	"$c ? my $x : ::foo; use utf8; $c ? my $x : é; my $x :Foo é bar; my $x : Foo ::bar baz; $c ? (my $x) : shared; $c ? local $x : shared; $c ? $x : shared; $c ? my ($x) : shared; $c ? my $x :\n"

// The negated parts (NotText, NotPrev1, NotPrev2, NotNext, NotNextIdent) need keys the corpus may not hold: x after a unary
// operator, a keyword as a word, a unary operator before a parenthesis.
func TestInternIdentityNegatedParts(t *testing.T) {
	tt, classes := committed(t)
	src := []byte("shift x 3; $a x 3; ref($a) x 3; ref x 3; last x; not (x); if; foo; time x 3; " + colonForms)
	s, _ := Lex(src, tt)
	want := refIDs(t, src, tt, classes)
	for i, tok := range s.Tokens {
		if tok.TerminalSet != want[i] {
			t.Errorf("token %d has set %d, the reference %d", i, tok.TerminalSet, want[i])
		}
	}
	if len(s.Tokens) != len(want) {
		t.Errorf("%d tokens, the reference %d", len(s.Tokens), len(want))
	}
}

// std5k is std5k.pl and its significant tokens.
func std5k(b *testing.B) ([]byte, []lexer.Token) {
	src, err := os.ReadFile("../testdata/std5k.pl")
	if err != nil {
		b.Fatal(err)
	}
	var sig []lexer.Token
	for _, t := range lexer.Tokenize(src) {
		if !lexer.IsTrivia(t.Kind) && t.Kind != lexer.HeredocBody {
			sig = append(sig, t)
		}
	}
	return src, sig
}

func perToken(b *testing.B, n int) {
	b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N)/float64(n), "ns/token")
}

// BenchmarkIntern is Intern alone, each token given its predecessor's class
// as Map gives it, reported per significant token.
func BenchmarkIntern(b *testing.B) {
	tt, _ := committed(b)
	src, sig := std5k(b)
	text := unsafe.String(unsafe.SliceData(src), len(src))
	b.ResetTimer()
	for range b.N {
		var slots [2]grammar.Token
		pred := &grammar.Boundary
		for i, t := range sig {
			intern(tt, text, t, pred, &slots[i&1])
			pred = &slots[i&1].N
		}
	}
	perToken(b, len(sig))
}

// BenchmarkTokenize is the reference for Intern's share: lexer.Tokenize per
// significant token of the same file.
func BenchmarkTokenize(b *testing.B) {
	src, sig := std5k(b)
	b.ResetTimer()
	for range b.N {
		lexer.Tokenize(src)
	}
	perToken(b, len(sig))
}

func ptrs(ns []grammar.Neighbour) []*grammar.Neighbour {
	ps := make([]*grammar.Neighbour, len(ns))
	for i := range ns {
		ps[i] = &ns[i]
	}
	return ps
}
