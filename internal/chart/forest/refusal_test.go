// ABOUTME: Tests the refusal reason by rerun: where narrowing changed the lattice, the un-narrowed rerun decides grammar or scope.
// ABOUTME: lex-unmapped follows V1's rule on the run V1 made; the rerun's recognizer is its own, never the caller's.

package forest

import (
	"slices"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/chart/lex"
	"tamarou.com/pvm/internal/chart/prectab"
	"tamarou.com/pvm/internal/chart/recognize"
	"tamarou.com/pvm/internal/lexer"
)

// narrowedDeath is the byte where src's narrowed chart dies under an empty scope, -1 where it accepts, and whether
// narrowing changed its lattice.
func narrowedDeath(t *testing.T, src string) (int, bool) {
	t.Helper()
	tab := committedTables(t)
	s, _ := lex.Lex([]byte(src), tab.Tokens)
	l, narrowed := narrow(tab, s, []byte(src), Scope{})
	c, err := recognize.Recognize(recognize.New(tab, recognize.Options{}), l)
	if err != nil {
		t.Fatal(err)
	}
	if c.Accepted() {
		return -1, narrowed
	}
	return diedAt(c, s, len(src)), narrowed
}

// refusalRow is a source, its refusal under an empty scope, and where its narrowed chart dies.
type refusalRow struct {
	src  string
	want Refusal
	died int
}

// checkRefusals parses each row under an empty scope and checks its refusal, and that its narrowed chart dies where
// the row says, on a lattice narrowing changed: a row whose fixture does not hold proves nothing.
func checkRefusals(t *testing.T, rows []refusalRow) {
	t.Helper()
	tab := committedTables(t)
	for _, row := range rows {
		died, narrowed := narrowedDeath(t, row.src)
		if !narrowed || died != row.died {
			t.Errorf("%q: the narrowed chart dies at %d (narrowed %v); the fixture wants %d on a narrowed lattice", row.src, died, narrowed, row.died)
		}
		if f, got := Parse(tab, []byte(row.src), Scope{}); f != nil || got != row.want {
			t.Errorf("%q: Parse refuses %+v, want %+v", row.src, got, row.want)
		}
	}
}

// TestRefusalRerunScopeVersusGrammar: a file the narrowed chart refuses is scope at that chart's byte where the
// un-narrowed rerun accepts it, and grammar at the rerun's byte where the rerun dies. `foo plus 2 plus 3;` is the
// scope byte perigrin's ruling lets move: V1's resolve located it at 4, the narrowed chart dies at 11.
func TestRefusalRerunScopeVersusGrammar(t *testing.T) {
	checkRefusals(t, []refusalRow{
		{"foo plus 2 plus 3;", Refusal{Byte: 11, Reason: "scope"}, 11},
		{"1 plus 2;\n", Refusal{Byte: 2, Reason: "scope"}, 2},
		{"1 plus 2; 1 + ;", Refusal{Byte: 14, Reason: "grammar"}, 2},
		{"foo plus 2; )", Refusal{Byte: 12, Reason: "grammar"}, 12},
	})
	if v1, _ := V1Forest(committedTables(t), []byte("foo plus 2 plus 3;"), Scope{}); v1 != nil {
		t.Error("V1 accepts `foo plus 2 plus 3;`, which it refuses for scope")
	}
}

// TestRefusalRerunTieIsGrammar: where the un-narrowed rerun dies at the narrowed chart's own byte, the reason is
// grammar: the rerun died, so narrowing did not refuse the file.
func TestRefusalRerunTieIsGrammar(t *testing.T) {
	checkRefusals(t, []refusalRow{{"foo plus 2; )", Refusal{Byte: 12, Reason: "grammar"}, 12}})
}

// TestRefusalRerunLaterGrammarFaultIsGrammar: a scope fault at byte 2 and a grammar fault at byte 14 are grammar at
// 14, as V1 gives it: the rerun dying anywhere makes the reason grammar.
func TestRefusalRerunLaterGrammarFaultIsGrammar(t *testing.T) {
	checkRefusals(t, []refusalRow{{"1 plus 2; 1 + ;", Refusal{Byte: 14, Reason: "grammar"}, 2}})
}

// TestRefusalRerunScopeThenUnmappedIsLexUnmapped: a scope fault at byte 2, then an unterminated heredoc, the first
// unmapped token, at 16: the rerun dies at 16, so the reason is lex-unmapped there, as V1 gives it.
func TestRefusalRerunScopeThenUnmappedIsLexUnmapped(t *testing.T) {
	const src = "1 plus 2;\nprint <<EOT;\nno terminator\n"
	if _, unmapped := lex.Lex([]byte(src), committedTables(t).Tokens); unmapped != 16 {
		t.Fatalf("%q: first unmapped byte %d, want 16", src, unmapped)
	}
	checkRefusals(t, []refusalRow{{src, Refusal{Byte: 16, Reason: "lex-unmapped"}, 2}})
}

// TestRefusalLexUnmappedNotRerun: a narrowed chart that dies at the first unmapped token is lex-unmapped there with
// no rerun, since the un-narrowed run dies there too.
func TestRefusalLexUnmappedNotRerun(t *testing.T) {
	const src = "foo plus 2;\nprint <<EOT;\nno terminator\n"
	before := reruns.Load()
	checkRefusals(t, []refusalRow{{src, Refusal{Byte: 18, Reason: "lex-unmapped"}, 18}})
	if n := reruns.Load() - before; n != 0 {
		t.Errorf("%q: %d reruns, want none", src, n)
	}
}

// TestRefusalEverySetAliveAtEnd: a file refused with every set alive, narrowed or not, is refused at its end (F2-11).
func TestRefusalEverySetAliveAtEnd(t *testing.T) {
	tab := committedTables(t)
	for _, src := range []string{"1 +", "foo plus 2 +"} {
		s, _ := lex.Lex([]byte(src), tab.Tokens)
		l, _ := narrow(tab, s, []byte(src), Scope{})
		c, err := recognize.Recognize(recognize.New(tab, recognize.Options{}), l)
		if err != nil {
			t.Fatal(err)
		}
		if c.Accepted() || slices.Contains(c.Sets, nil) {
			t.Fatalf("%q: the chart accepts it, or a set died, so the test proves nothing", src)
		}
		if f, got := Parse(tab, []byte(src), Scope{}); f != nil || got != (Refusal{Byte: len(src), Reason: "grammar"}) {
			t.Errorf("%q: Parse refuses %+v, want grammar at %d, its end", src, got, len(src))
		}
	}
}

// TestRefusalRerunLeavesCallerRecognizerUntouched: a refused file's rerun recognizes with a recognizer of its own.
// Parsing it through the caller's long-lived recognizer, which has already recognized its narrowed lattice and so
// holds every set and transition of the narrowed run, adds no entry, though the rerun's un-narrowed lattice would
// add some there.
func TestRefusalRerunLeavesCallerRecognizerUntouched(t *testing.T) {
	tab := committedTables(t)
	const src = "foo plus 2 plus 3;"
	r := recognize.New(tab, recognize.Options{})
	s, _ := lex.Lex([]byte(src), tab.Tokens)
	l, _ := narrow(tab, s, []byte(src), Scope{})
	if _, err := recognize.Recognize(r, l); err != nil {
		t.Fatal(err)
	}
	sets, cached := r.Entries()
	before := reruns.Load()
	if _, ref := ParseWith(tab, r, []byte(src), Scope{}); ref.Reason != "scope" || reruns.Load() == before {
		t.Fatalf("%q: refusal %+v with %d reruns; want a scope refusal by rerun", src, ref, reruns.Load()-before)
	}
	if s, c := r.Entries(); s != sets || c != cached {
		t.Errorf("the caller's recognizer holds %d sets and %d transitions after the rerun, %d and %d before", s, c, sets, cached)
	}
	if _, err := recognize.Recognize(r, lex.Linear(s)); err != nil {
		t.Fatal(err)
	}
	if s, c := r.Entries(); s == sets && c == cached {
		t.Error("the un-narrowed lattice adds no entry to the recognizer, so the check above proves nothing")
	}
}

// judgedDeath is where the judged chart of src's narrowed lattice dies, under an empty scope; -1 where it accepts.
func judgedDeath(t *testing.T, src string) int {
	t.Helper()
	tab := committedTables(t)
	s, _ := lex.Lex([]byte(src), tab.Tokens)
	l, _ := narrow(tab, s, []byte(src), Scope{})
	c, err := recognize.Recognize(recognize.New(tab, recognize.Options{}), l)
	if err != nil {
		t.Fatal(err)
	}
	if c.Accepted() {
		return -1
	}
	return diedAt(c, s, len(src))
}

// TestPrecedenceRefusalByteWhereChartDied: each constructed precedence fixture, which only precedence refuses, is
// refused for precedence at the byte where the judged chart died, never byte 0.
func TestPrecedenceRefusalByteWhereChartDied(t *testing.T) {
	tab := committedTables(t)
	for src, at := range map[string]int{"print $fh [];": 11, "$a <=> $b <=> $c;": 10, "1 .. 2 .. 3;": 7} {
		if died := judgedDeath(t, src); died != at {
			t.Errorf("%q: the judged chart dies at %d, the fixture wants %d", src, died, at)
		}
		if f, got := Parse(tab, []byte(src), Scope{}); f != nil || got != (Refusal{Byte: at, Reason: "precedence"}) {
			t.Errorf("%q: Parse refuses %+v, want precedence at %d", src, got, at)
		}
	}
}

// TestRefusalRerunOrder: on files built to have two faults the reason follows V1's order, grammar before scope before
// precedence, wherever in the file each fault is, and a precedence refusal is never at byte 0.
func TestRefusalRerunOrder(t *testing.T) {
	tab := committedTables(t)
	for name, rows := range map[string][]refusalRow{
		"grammar-before-precedence": {
			{"1 .. 2 .. 3; 1 + ;", Refusal{Byte: 17, Reason: "grammar"}, 7},
			{"1 + ; 1 .. 2 .. 3;", Refusal{Byte: 4, Reason: "grammar"}, 4},
		},
		"scope-before-precedence": {
			{"1 .. 2 .. 3; 1 plus 2;", Refusal{Byte: 15, Reason: "scope"}, 7},
			{"1 plus 2; 1 .. 2 .. 3;", Refusal{Byte: 2, Reason: "scope"}, 2},
		},
		"precedence-byte-not-zero": {
			{"$x = 1;\n1 .. 2 .. 3;", Refusal{Byte: 15, Reason: "precedence"}, 15},
		},
	} {
		t.Run(name, func(t *testing.T) {
			for _, row := range rows {
				if died := judgedDeath(t, row.src); died != row.died {
					t.Errorf("%q: the judged chart dies at %d, the fixture wants %d", row.src, died, row.died)
				}
				if f, got := Parse(tab, []byte(row.src), Scope{}); f != nil || got != row.want || got.Byte == 0 {
					t.Errorf("%q: Parse refuses %+v, want %+v", row.src, got, row.want)
				}
			}
		})
	}
}

// TestRefusalRerunPrecedenceThenUnmappedIsLexUnmapped: a precedence fault at byte 7, then an unterminated heredoc,
// the first unmapped token, at 19: the un-narrowed rerun dies at 19, so the reason is lex-unmapped there, as V1's recognizer gives it.
func TestRefusalRerunPrecedenceThenUnmappedIsLexUnmapped(t *testing.T) {
	const src = "1 .. 2 .. 3;\nprint <<EOT;\nno terminator\n"
	tab := committedTables(t)
	if _, unmapped := lex.Lex([]byte(src), tab.Tokens); unmapped != 19 {
		t.Fatalf("%q: first unmapped byte %d, want 19", src, unmapped)
	}
	if died := judgedDeath(t, src); died != 7 {
		t.Fatalf("%q: the judged chart dies at %d, want 7, the precedence fault", src, died)
	}
	if f, got := Parse(tab, []byte(src), Scope{}); f != nil || got != (Refusal{Byte: 19, Reason: "lex-unmapped"}) {
		t.Errorf("%q: Parse refuses %+v, want lex-unmapped at 19", src, got)
	}
}

// TestRefusalByteWhereExpectedEmpty: recognize's TestExpectedEmptyWhenPrecedenceForbidsAll toy, T nonassoc, on the
// source "A T A T A", one byte a token two apart: precedence forbids every candidate after A T A, so the judged chart
// dies at the second T with Expected empty there, and the refusal byte diedAt produces is that T's, byte 6.
func TestRefusalByteWhereExpectedEmpty(t *testing.T) {
	tab := toyTables(t, "S ::= E", "E ::= E T E | A")
	p, err := prectab.NewPrecedence(tab, []grammar.Level{{N: 1, Assoc: "nonassoc", Tokens: []string{"T"}}}, make([]string, len(tab.Rules)), nil)
	if err != nil {
		t.Fatal(err)
	}
	r, err := recognize.NewJudged(tab, p, recognize.Options{})
	if err != nil {
		t.Fatal(err)
	}
	words := []string{"A", "T", "A", "T", "A"}
	c, err := recognize.Recognize(r, toyLattice(tab, words))
	if err != nil {
		t.Fatal(err)
	}
	if got, err := c.Expected(3); err != nil || len(got) != 0 || c.Sets[3] == nil {
		t.Fatalf("Expected after A T A is %v, %v (set %v); want empty and no error at a node with a set", got, err, c.Sets[3])
	}
	s := &lex.Stream{}
	for i := range words {
		s.Raw = append(s.Raw, lexer.Token{Start: 2 * i, End: 2*i + 1})
		s.Tokens = append(s.Tokens, lex.Token{Index: i})
	}
	if at := diedAt(c, s, 2*len(words)-1); at != 6 {
		t.Errorf("the refusal byte is %d; want 6, the second T's", at)
	}
}
