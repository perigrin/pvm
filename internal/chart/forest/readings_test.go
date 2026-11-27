// ABOUTME: Tests the readings count: per statement after precedence and resolve, collapse above 2^10, saturation.
// ABOUTME: Toy counts are checked against brute force; perly.y rows show precedence and scope cutting readings.

package forest

import (
	"cmp"
	"flag"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"tamarou.com/pvm/internal/chart/lex"
	"tamarou.com/pvm/internal/chart/recognize"
	"tamarou.com/pvm/internal/parse"
)

// builtReadings is src's readings over the forest Build makes, before precedence and resolve.
func builtReadings(t *testing.T, src string) Readings {
	t.Helper()
	tab := committedTables(t)
	s, _ := lex.Lex([]byte(src), tab.Tokens)
	l := lex.Linear(s)
	c, err := recognize.Recognize(recognize.NewNoPrecedence(tab, recognize.Options{}), l)
	if err != nil || !c.Accepted() {
		t.Fatalf("%q: recognition refuses it (%v)", src, err)
	}
	return Build(tab, c, l).countReadings(symbol(tab, "fullstmt"))
}

// parsedReadings is src's readings as Parse leaves them under scope, failing the test where Parse refuses it.
func parsedReadings(t *testing.T, src string, scope Scope) Readings {
	t.Helper()
	f, ref := Parse(committedTables(t), []byte(src), scope)
	if f == nil {
		t.Fatalf("%q: Parse refuses it (%+v)", src, ref)
	}
	return f.Readings
}

// ambiguous is a statement perl leaves ambiguous to the grammar: Foo::bar's { opens a block (a block argument or an
// indirect object) or an anonymous hash (an argument to a unary or a list call); its inner statement is a block's.
const ambiguous = "Foo::bar { a => 1 };"

// TestReadingsCollapse: a statement's readings are its derivations after precedence and resolve, collapse is above
// 2^10 and not at it, and the toy counts agree with brute force.
func TestReadingsCollapse(t *testing.T) {
	// L is a run of X, each one A read two ways, so n As make one statement of 2^n readings.
	tab := toyTables(t, "file ::= fullstmt", "fullstmt ::= L", "L ::= L X | X", "X ::= A | Y", "Y ::= A")
	stmt := symbol(tab, "fullstmt")
	for _, c := range []struct {
		n         int
		collapsed int
		total     uint64
	}{{1, 0, 2}, {10, 0, 1 << 10}, {11, 1, 1}} {
		words := slices.Repeat([]string{"A"}, c.n)
		f, _, _ := toyForest(t, tab, words, recognize.Options{})
		if f == nil {
			t.Fatalf("%d As: no forest", c.n)
		}
		if b := brute(tab, words, stmt, 0, c.n, map[[3]int]bool{}); c.collapsed == 0 && uint64(b) != c.total {
			t.Fatalf("%d As: brute force counts %d, the row says %d", c.n, b, c.total)
		}
		want := Readings{Statements: 1, Collapsed: c.collapsed, Total: c.total}
		if got := f.countReadings(stmt); got != want {
			t.Errorf("%d As: %+v, want %+v", c.n, got, want)
		}
	}

	// shortcut, pinned: on a toy's two-node cycle fullstmt -> T -> fullstmt, counting fullstmt first memoizes T as 0
	// (its one derivation, T -> fullstmt -> A, holds fullstmt, met on the path); T counted first would be 1.
	cyc := toyTables(t, "fullstmt ::= T | A", "T ::= fullstmt")
	cf, _, _ := toyForest(t, cyc, []string{"A"}, recognize.Options{})
	if cf == nil {
		t.Fatal("cyclic toy: no forest")
	}
	tn := int32(slices.IndexFunc(cf.Nodes, func(n Node) bool { return int(n.Sym) == symbol(cyc, "T") }))
	if tn <= 0 {
		t.Fatal("cyclic toy: no T node; the forest dropped the cycle")
	}
	if count := cf.counter(); count(cf.Root) != 1 || count(tn) != 0 {
		t.Errorf("cyclic toy: fullstmt %d, then T %d; want 1, then 0 (the shortcut's undercount)", count(cf.Root), count(tn))
	}

	imported := Scope{Imports: map[string]parse.Import{"foo": {Name: "foo", Prototype: "($)", PrototypeKnown: true}}}
	for _, c := range []struct {
		what, src string
		scope     Scope
		want      Readings
	}{
		// 14 groupings of four binary operators before precedence; bison builds one.
		{"an operator chain precedence reduces", "$x = 1 + 2 * 3 - 4;", Scope{}, Readings{Statements: 1, Total: 1}},
		// foo with a ($) prototype is a unary: foo(1), 2, not a list call or a bareword.
		{"a scope-decided word resolve reduces", "foo 1, 2;", imported, Readings{Statements: 1, Total: 1}},
		// four readings of the outer statement, and the inner a => 1 as the block's own statement, once.
		{"an ambiguity perl leaves", ambiguous, Scope{}, Readings{Statements: 2, Total: 5}},
	} {
		got := parsedReadings(t, c.src, c.scope)
		if got != c.want {
			t.Errorf("%s: %q reads %+v, want %+v", c.what, c.src, got, c.want)
		}
		if before := builtReadings(t, c.src); c.want.Total == 1 && before.Total <= 1 {
			t.Errorf("%s: %q has %d readings before precedence and resolve; the row shows nothing cut", c.what, c.src, before.Total)
		}
	}

	// five ambiguous statements in a block are 4^5 = 2^10 readings of it, kept; six collapse it. Each keeps its own
	// four, and each inner statement its one.
	for _, c := range []struct {
		n    int
		want Readings
	}{{5, Readings{Statements: 11, Total: 1<<10 + 5*4 + 5}}, {6, Readings{Statements: 13, Collapsed: 1, Total: 1 + 6*4 + 6}}} {
		src := "{ " + strings.Repeat(ambiguous+" ", c.n) + "}"
		if got := parsedReadings(t, src, Scope{}); got != c.want {
			t.Errorf("a block of %d: %+v, want %+v", c.n, got, c.want)
		}
	}
}

// TestReadingsCollapsedFileStillAccepted: collapse is no refusal. A file with a collapsed statement is accepted, and
// the statements beside it and inside it keep their own counts.
func TestReadingsCollapsedFileStillAccepted(t *testing.T) {
	src := "my $x = 1;\n{ " + strings.Repeat(ambiguous+" ", 6) + "}\n$x = 1 + 2 * 3;\n" + ambiguous + "\n"
	f, ref := Parse(committedTables(t), []byte(src), Scope{})
	if f == nil || ref != (Refusal{}) {
		t.Fatalf("a file with a collapsed statement is refused: %+v", ref)
	}
	// my, the block (collapsed, once), its six statements and their six inner ones, the chain, the last statement
	// (4) and its inner one.
	want := Readings{Statements: 1 + 1 + 6 + 6 + 1 + 2, Collapsed: 1, Total: 1 + 1 + 6*4 + 6 + 1 + 4 + 1}
	if f.Readings != want {
		t.Errorf("%+v, want %+v", f.Readings, want)
	}
}

// TestReadingsCountBoundedOnPathologicalInput: a statement with more readings than a uint64 holds saturates, and
// counting finishes promptly, since it memoizes each node rather than enumerating.
func TestReadingsCountBoundedOnPathologicalInput(t *testing.T) {
	// S S over 200 As has Catalan(199) groupings, about 10^116; the cyclic, nullable alternatives add none.
	tab := toyTables(t, "fullstmt ::= fullstmt fullstmt | fullstmt | A |")
	f, _, _ := toyForest(t, tab, slices.Repeat([]string{"A"}, 200), recognize.Options{})
	if f == nil {
		t.Fatal("no forest")
	}
	start := time.Now()
	got := f.countReadings(symbol(tab, "fullstmt"))
	top := f.counter()(f.Root)
	if d := time.Since(start); d > 10*time.Second {
		t.Errorf("counting took %v", d)
	}
	if top != math.MaxUint64 || got.Collapsed == 0 {
		t.Errorf("Catalan(199) readings: the top statement's %d, want MaxUint64; %+v, want statements collapsed", top, got)
	}
	// 4^40 readings of a perly.y block: collapsed, counted once, its 40 statements and their inner ones kept.
	if r := parsedReadings(t, "{ "+strings.Repeat(ambiguous+" ", 40)+"}", Scope{}); r != (Readings{Statements: 81, Collapsed: 1, Total: 1 + 40*4 + 40}) {
		t.Errorf("a block of 40: %+v", r)
	}
	if s := mulSat(math.MaxUint64/3, 4); s != math.MaxUint64 {
		t.Errorf("mulSat overflowed to %d", s)
	}
	if s := addSat(math.MaxUint64-1, 2); s != math.MaxUint64 {
		t.Errorf("addSat overflowed to %d", s)
	}
}

// TestReadingsNestedStatementsCountedOnce: a statement inside another's block is one statement of the file's total,
// counted once as its own fullstmt, however many of the outer one's readings hold it.
func TestReadingsNestedStatementsCountedOnce(t *testing.T) {
	for _, c := range []struct {
		src  string
		want int
	}{
		{"if ($x) { foo(); bar(); } baz();", 4},
		{"while (1) { if ($x) { last; } }", 3},
		{ambiguous, 2}, // its inner statement is held by two of its four readings
	} {
		if got := parsedReadings(t, c.src, Scope{}); got.Statements != c.want {
			t.Errorf("%q: %d statements, want %d", c.src, got.Statements, c.want)
		}
	}
}

// TestParseAppliesPrecedence: Parse refuses a file precedence alone refuses, with reason precedence at the byte where
// the judged chart died, the second <=>, not Byte 0.
func TestParseAppliesPrecedence(t *testing.T) {
	if f, ref := Parse(committedTables(t), []byte("$a <=> $b <=> $c;"), Scope{}); f != nil || ref != (Refusal{Byte: 10, Reason: "precedence"}) {
		t.Errorf("$a <=> $b <=> $c;: forest %v, refusal %+v; want refused by precedence at byte 10", f != nil, ref)
	}
}

var readingsRate = flag.Bool("forest.readings-rate", false, "log each corpus's collapsed statements (TestReadingsCorpusRate)")

// TestReadingsCorpusRate logs, with -forest.readings-rate, each corpus's statements and collapsed statements as
// forest.Parse leaves them, with each file's imports, and whether the corpus is over spec §10's limit (more than 1%,
// and at least 5, collapsing). It measures; the ratchet's collapse-over-limit is the gate.
func TestReadingsCorpusRate(t *testing.T) {
	if !*readingsRate {
		t.Skip("set -forest.readings-rate to measure")
	}
	tab := committedTables(t)
	stubs := filepath.Join("..", "testdata", "guacamole", "stubs")
	perl5 := cmp.Or(os.Getenv("PERL5_CORPUS"), "/opt/corpora/perl5")
	var perlgit []string
	_ = filepath.WalkDir(filepath.Join(perl5, "t"), func(p string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() && strings.HasSuffix(p, ".t") {
			perlgit = append(perlgit, p)
		}
		return nil
	})
	glob := func(patterns ...string) []string {
		var out []string
		for _, g := range patterns {
			ps, _ := filepath.Glob(g)
			out = append(out, ps...)
		}
		return out
	}
	for _, c := range []struct {
		name  string
		files []string
	}{
		{"guacamole", glob(filepath.Join("..", "testdata", "guacamole", "parses", "*.pl"), filepath.Join("..", "testdata", "guacamole", "parsent", "*.pl"))},
		{"std5k", glob(filepath.Join("..", "testdata", "std5k.pl"))},
		{"edit5k", glob(filepath.Join("..", "..", "psc", "testdata", "latency", "edit5k.pl"))},
		{"T1", glob(filepath.Join(cmp.Or(os.Getenv("T1_CORPUS"), filepath.Join("..", "..", "..", "testdata", "perlonjava", "unit")), "*.t"))},
		{"perlgit", perlgit},
	} {
		var accepted, statements, collapsed int
		for _, p := range c.files {
			src, err := os.ReadFile(p)
			if err != nil {
				t.Fatal(err)
			}
			root := parse.ParseWithLoader(src, parse.DirLoader(filepath.Dir(p), stubs))
			if f, _ := Parse(tab, src, Scope{Imports: parse.Imports(root)}); f != nil {
				accepted++
				statements += f.Readings.Statements
				collapsed += f.Readings.Collapsed
			}
		}
		over := collapsed >= 5 && collapsed*100 > statements
		t.Logf("%s: %d files, %d accepted, %d statements, %d collapsed (%.2f%%), over the limit: %v",
			c.name, len(c.files), accepted, statements, collapsed, 100*float64(collapsed)/float64(max(statements, 1)), over)
	}
}
