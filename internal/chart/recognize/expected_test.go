// ABOUTME: Tests that Expected on a judged chart reports only the continuations precedence allows (V2 design §4.4),
// ABOUTME: with Leo on and off agreeing, over perly.y examples, a toy where every candidate is refused, and the corpus.

package recognize

import (
	"cmp"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/chart/lex"
)

// expectedBefore is Expected, on the chart r recognizes from src over perly.y, at the node before src's last token.
func expectedBefore(t *testing.T, r *Recognizer, src string) []string {
	t.Helper()
	s, _ := lex.Lex([]byte(src), r.tab.Tokens)
	l := lex.Linear(s)
	return expectedOf(t, mustRecognize(t, r, l), l.Edges[len(l.Edges)-1].From)
}

// judgedOptions are the options Expected must answer alike under.
var judgedOptions = []Options{{}, {NoLeo: true}, {NoLookahead: true}, {NoCache: true}}

// TestExpectedJudgesPrecedence: DOTDOT is nonassoc and NCEQOP may not chain, so after `1 .. 2` no DOTDOT and after
// `$a <=> $b` no NCEQOP is a continuation, whatever follows; the no-precedence recognizer offers both there.
func TestExpectedJudgesPrecedence(t *testing.T) {
	tab := committedTables(t)
	for _, tc := range []struct{ src, never string }{
		{"1 .. 2;", "DOTDOT"}, {"1 .. 2 + 3;", "DOTDOT"}, {"1 .. 2 }", "DOTDOT"},
		{"$a <=> $b;", "NCEQOP"}, {"$a <=> $b }", "NCEQOP"},
	} {
		if plain := expectedBefore(t, NewNoPrecedence(tab, Options{}), tc.src); !slices.Contains(plain, tc.never) {
			t.Fatalf("%q: the no-precedence Expected %v lacks %s: the fixture forbids nothing", tc.src, plain, tc.never)
		}
		for _, opts := range judgedOptions {
			got := expectedBefore(t, mustJudged(t, tab, nil, opts), tc.src)
			if slices.Contains(got, tc.never) {
				t.Errorf("%+v %q: Expected before the last token %v offers %s, which precedence forbids", opts, tc.src, got, tc.never)
			}
		}
	}
}

// TestExpectedKeepsAllowedContinuations: what precedence allows stays: `;` and a higher operator after `1 .. 2`,
// a term after `1 ..`, and `;` after `$a <=> $b`.
func TestExpectedKeepsAllowedContinuations(t *testing.T) {
	tab := committedTables(t)
	for _, tc := range []struct {
		src  string
		want []string
	}{
		{"1 .. 2;", []string{"PERLY_SEMICOLON", "ADDOP", "MULOP", "OROP"}},
		{"1 .. ;", []string{"THING", "PERLY_DOLLAR", "PERLY_MINUS"}},
		{"$a <=> $b;", []string{"PERLY_SEMICOLON", "ADDOP", "ANDAND"}},
	} {
		for _, opts := range judgedOptions {
			got := expectedBefore(t, mustJudged(t, tab, nil, opts), tc.src)
			for _, w := range tc.want {
				if !slices.Contains(got, w) {
					t.Errorf("%+v %q: Expected before the last token %v lacks %s", opts, tc.src, got, w)
				}
			}
		}
	}
}

// TestExpectedEmptyWhenPrecedenceForbidsAll: with T nonassoc, after A T A the only candidate is T: completing
// E T E before it is refused, and the inner E . T E's shift of it is refused under the right operand's top. So the
// chart dies at the second T, and Expected at the node before it is empty, without error, where the no-precedence
// recognizer offers T.
func TestExpectedEmptyWhenPrecedenceForbidsAll(t *testing.T) {
	tab, p := precToy(t, []grammar.Level{{N: 1, Assoc: "nonassoc", Tokens: []string{"T"}}}, nil, nil,
		"S ::= E", "E ::= E T E | A")
	l := tokens(tab, "A", "T", "A", "T", "A")
	if got := expectedOf(t, mustRecognize(t, NewNoPrecedence(tab, Options{}), l), 3); !slices.Contains(got, "T") {
		t.Fatalf("the no-precedence Expected at node 3 is %v; want T, or the fixture forbids nothing", got)
	}
	for _, opts := range judgedOptions {
		c := mustRecognize(t, mustJudged(t, tab, p, opts), l)
		if c.Accepted() || c.Sets[3] == nil || c.Sets[4] != nil {
			t.Fatalf("%+v: accepted %v, set at 3 %v, at 4 %v; want refused at the second T", opts, c.Accepted(), c.Sets[3], c.Sets[4])
		}
		got, err := c.Expected(3)
		if err != nil || len(got) != 0 {
			t.Errorf("%+v: Expected at node 3 is %v, %v; want empty and no error", opts, got, err)
		}
	}
}

// expectedProbes are perly.y inputs with right recursions Leo takes and precedence judging their continuations.
var expectedProbes = []string{
	"1 .. 2;", "$a <=> $b;", "$a = $b = $c = 1;", "$x = 2 ** 3 ** 4;", "$a ? $b : $c ? $d : $e;",
	"$a = $b ? 1 .. 2 : $c <=> $d;", "print 1, 2, 3 if $a || $b && $c;", "$h{1}{2}[3] = -$x ** 2;",
	"$a = $b .. $c = $d;", "my $x = $y // $z || 1 .. 2;",
}

// TestExpectedLeoAgreesUnderPrecedence: Expected on the judged chart is the same with Leo on and off at every node
// of the probes, and some of those nodes' answers were reached through a Leo memo.
func TestExpectedLeoAgreesUnderPrecedence(t *testing.T) {
	tab := committedTables(t)
	positions, leo := 0, 0
	for _, src := range expectedProbes {
		s, _ := lex.Lex([]byte(src), tab.Tokens)
		l := lex.Linear(s)
		r := mustJudged(t, tab, nil, Options{NoCache: true})
		on, off := mustRecognize(t, r, l), mustRecognize(t, mustJudged(t, tab, nil, Options{NoLeo: true}), l)
		for pos := range on.Sets {
			steps := r.leoSteps
			a, b := expectedOf(t, on, pos), expectedOf(t, off, pos)
			if !slices.Equal(a, b) {
				t.Errorf("%q node %d: Expected with Leo %v, without %v", src, pos, a, b)
			}
			positions++
			if r.leoSteps > steps {
				leo++
			}
		}
	}
	t.Logf("positions: %d", positions)
	t.Logf("leo positions: %d", leo)
	if positions == 0 || leo == 0 {
		t.Errorf("compared %d positions, %d through a Leo memo; want both non-zero", positions, leo)
	}
}

// TestLeoDisagreeZeroUnderPrecedence: over every T1 file (T1_CORPUS, else the in-tree copy), LeoDisagree, which
// compares Expected on the judged charts at every node, finds no file where Leo on and off differ.
func TestLeoDisagreeZeroUnderPrecedence(t *testing.T) {
	tab := committedTables(t)
	root := cmp.Or(os.Getenv("T1_CORPUS"), filepath.Join("..", "..", "..", "testdata", "perlonjava", "unit"))
	files, err := filepath.Glob(filepath.Join(root, "*.t"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no .t files under %s: %v", root, err)
	}
	n := 0
	for _, f := range files {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		s, _ := lex.Lex(src, tab.Tokens)
		d, err := LeoDisagree(tab, lex.Linear(s))
		if err != nil {
			t.Fatalf("%s: %v", f, err)
		}
		if d {
			t.Errorf("%s: Leo on and off disagree", f)
			n++
		}
	}
	t.Logf("files: %d", len(files))
	t.Logf("leo-disagree: %d", n)
}
