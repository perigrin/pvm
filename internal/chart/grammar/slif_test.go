// ABOUTME: Tests the SLIF-subset reader on hand-written snippets and on the vendored Guacamole grammar.
// ABOUTME: Covers each construct, the errors for unsupported and truncated input, and full-grammar rule counts.

package grammar

import (
	"errors"
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

func read(t *testing.T, src string) *Grammar {
	t.Helper()
	g, err := ReadSLIF("test.slif", []byte(src))
	if err != nil {
		t.Fatalf("ReadSLIF(%q): %v", src, err)
	}
	return g
}

func readErr(t *testing.T, src string) *Error {
	t.Helper()
	g, err := ReadSLIF("test.slif", []byte(src))
	if err == nil {
		t.Fatalf("ReadSLIF(%q) = %+v, want an error", src, g)
	}
	var e *Error
	if !errors.As(err, &e) {
		t.Fatalf("ReadSLIF(%q) error %v is not an *Error", src, err)
	}
	return e
}

func name(s string) Symbol { return Symbol{Name, s} }

func TestReaderAlternatives(t *testing.T) {
	g := read(t, "A ::= B C\n  | 'x' [a-z]\n")
	want := []Rule{
		{LHS: "A", RHS: []Symbol{name("B"), name("C")}, Line: 1, Decl: 1},
		{LHS: "A", RHS: []Symbol{{Literal, "x"}, {Class, "[a-z]"}}, Line: 2, Decl: 1},
	}
	if !reflect.DeepEqual(g.G1, want) {
		t.Errorf("G1 = %+v, want %+v", g.G1, want)
	}
	if len(g.G0) != 0 {
		t.Errorf("G0 = %+v, want none", g.G0)
	}
}

func TestReaderMultilineStatements(t *testing.T) {
	// The right-hand side may start on the line after the operator, and the
	// operator on the line after the left-hand side.
	g := read(t, "A ::=\n    B\n  | C\nD\n  ~ 'd'\n")
	if len(g.G1) != 2 || g.G1[0].Line != 2 || g.G1[1].Line != 3 || g.G1[1].Decl != 1 {
		t.Errorf("G1 = %+v", g.G1)
	}
	want := []Rule{{LHS: "D", RHS: []Symbol{{Literal, "d"}}, Line: 5, Decl: 4}}
	if !reflect.DeepEqual(g.G0, want) {
		t.Errorf("G0 = %+v, want %+v", g.G0, want)
	}
}

func TestReaderSequences(t *testing.T) {
	g := read(t, "L ::= E*\nM ::= E+\nN ~ [a-z]*\nO ~ <x y>+\n")
	for i, want := range []Rule{
		{LHS: "L", RHS: []Symbol{name("E")}, Seq: Star, Line: 1, Decl: 1},
		{LHS: "M", RHS: []Symbol{name("E")}, Seq: Plus, Line: 2, Decl: 2},
	} {
		if !reflect.DeepEqual(g.G1[i], want) {
			t.Errorf("G1[%d] = %+v, want %+v", i, g.G1[i], want)
		}
	}
	for i, want := range []Rule{
		{LHS: "N", RHS: []Symbol{{Class, "[a-z]"}}, Seq: Star, Line: 3, Decl: 3},
		{LHS: "O", RHS: []Symbol{name("x y")}, Seq: Plus, Line: 4, Decl: 4},
	} {
		if !reflect.DeepEqual(g.G0[i], want) {
			t.Errorf("G0[%d] = %+v, want %+v", i, g.G0[i], want)
		}
	}
}

func TestReaderBracketedNames(t *testing.T) {
	g := read(t, "<hash   comment> ~ '#' <comment body>\n<Foo> ~ 'f'\n")
	if g.G0[0].LHS != "hash comment" || g.G0[0].RHS[1] != name("comment body") || g.G0[1].LHS != "Foo" {
		t.Errorf("G0 = %+v", g.G0)
	}
}

func TestReaderPriority(t *testing.T) {
	g := read(t, "K ~ 'k'\n:lexeme ~ K priority => 1\n:lexeme ~ <L m> priority => -2\n")
	want := map[string]int{"K": 1, "L m": -2}
	if !reflect.DeepEqual(g.Priority, want) {
		t.Errorf("Priority = %v, want %v", g.Priority, want)
	}
	if len(g.G0) != 1 {
		t.Errorf("G0 = %+v, want only K", g.G0)
	}
}

func TestReaderDiscard(t *testing.T) {
	g := read(t, ":discard ~ whitespace\nwhitespace ~ [\\s]+\n:discard ~ <hash comment>\n")
	if want := []string{"whitespace", "hash comment"}; !reflect.DeepEqual(g.Discard, want) {
		t.Errorf("Discard = %v, want %v", g.Discard, want)
	}
}

func TestReaderLATMDefault(t *testing.T) {
	if g := read(t, "lexeme default = latm => 1\nA ~ 'a'\n"); !g.LATM {
		t.Error("LATM = false after latm => 1")
	}
	if g := read(t, "A ~ 'a'\n"); g.LATM {
		t.Error("LATM = true with no lexeme default")
	}
}

func TestReaderIgnoresAdverbs(t *testing.T) {
	plain := read(t, "A ::= B C | D\nE ::= F\n")
	adorned := read(t, "A ::= B C action => ::first | D action => ::first\n"+
		"E ::= F name => Block\n"+
		":default ::= action => [ name, start, length, values ]\n")
	if !reflect.DeepEqual(plain.G1, adorned.G1) {
		t.Errorf("adverbs changed the rules:\n got %+v\nwant %+v", adorned.G1, plain.G1)
	}
}

func TestReaderKeepsG1G0Apart(t *testing.T) {
	g := read(t, "A ::= B\nA ~ 'a'\n")
	if want := []Rule{{LHS: "A", RHS: []Symbol{name("B")}, Line: 1, Decl: 1}}; !reflect.DeepEqual(g.G1, want) {
		t.Errorf("G1 = %+v, want %+v", g.G1, want)
	}
	if want := []Rule{{LHS: "A", RHS: []Symbol{{Literal, "a"}}, Line: 2, Decl: 2}}; !reflect.DeepEqual(g.G0, want) {
		t.Errorf("G0 = %+v, want %+v", g.G0, want)
	}
}

func TestReaderRejectsUnsupported(t *testing.T) {
	for _, tc := range []struct {
		src  string
		line int
	}{
		{"A ::= B\nevent sqe = completed A\n", 2},
		{"A ::= B\n:start ::= A\n", 2},
		{"A ::= B\n  || C\n", 2},
		{"A ::= B\n  | C rank => 1\n", 2},
		{"A ~ 'a'\n:lexeme ~ A pause => before\n", 2},
		{"A ~ 'a'\n:lexeme ~ A priority => 1 event => x\n", 2},
		{"A ::= B* C\n", 1},
		{"A ::= B C*\n", 1},
		{"A ::= B*\n  | C\n", 2},
		{"A ::= B+ separator => C\n", 1},
		{"A ::= (B C)\n", 1},
		{"A ~ 'a':i\n", 1},
		{"A ~ \"a\"\n", 1},
		{"A ~ 'a' action => ::first\n", 1},
		{"lexeme default = action => ::first\n", 1},
		{"A ~ 'a'\n:lexeme ~ A priority => 1\n:lexeme ~ A priority => 2\n", 3},
	} {
		if e := readErr(t, tc.src); e.Line != tc.line {
			t.Errorf("ReadSLIF(%q) error %q names line %d, want %d", tc.src, e, e.Line, tc.line)
		}
	}
}

func TestReaderRejectsTruncated(t *testing.T) {
	for _, tc := range []struct {
		src  string
		line int
	}{
		{":lexeme ~ X priority =>", 1},
		{":lexeme ~ X priority =>\nB ::= C\n", 1},
		{"A ::= B\n:lexeme ~ X priority =>\n", 2},
		{"A ::= B |", 1},
		{"A ::= B |\nC ::= D\n", 1},
		{"A ::= B\nC ::= D\n  |\n", 3},
		{"A ::=\nB ::= C\n", 1},
		{"A ::= B action =>\n", 1},
		{"", 1},
		{"# only a comment\n\n", 1},
	} {
		if e := readErr(t, tc.src); e.Line != tc.line {
			t.Errorf("ReadSLIF(%q) error %q names line %d, want %d", tc.src, e, e.Line, tc.line)
		}
	}
}

func TestReaderErrorNamesFile(t *testing.T) {
	_, err := ReadSLIF("guacamole.slif", []byte("A ::= B\nevent x = completed A\n"))
	if err == nil || !strings.HasPrefix(err.Error(), "guacamole.slif:2: ") {
		t.Errorf("error = %v, want it to start guacamole.slif:2: ", err)
	}
}

const guacamolePath = "testdata/guacamole/guacamole.slif"

// TestReaderReadsGuacamole reads the vendored grammar and checks that every
// statement in it became rules, against counts taken from the text by line:
//
//	G1 statements: lines with ::= that start with neither # nor :      436
//	               grep -c '^[^#:]*::=' guacamole.slif
//	G0 statements: lines starting "Name ~" or "<name> ~"               395
//	               plus lines starting "~" (LHS alone on the line above) 2
//	                                                                   = 397
//	OpKeyword*Expr: lines starting "OpKeyword<name>Expr ::=", digits included (Atan2) 198
//	:lexeme lines                                                      214
//	alternatives:  statements plus each "|" outside quotes, classes and comments
func TestReaderReadsGuacamole(t *testing.T) {
	src, err := os.ReadFile(guacamolePath)
	if err != nil {
		t.Fatal(err)
	}
	g, err := ReadSLIF("guacamole.slif", src)
	if err != nil {
		t.Fatalf("ReadSLIF: %v", err)
	}

	count := func(re string) int {
		return len(regexp.MustCompile("(?m)"+re).FindAll(src, -1))
	}
	g1Lines := count(`^[^#:\n]*::=`)
	g0Lines := count(`^[ \t]*(<[^>\n]+>|[A-Za-z_]\w*)[ \t]*~`) + count(`^[ \t]*~`)
	kwLines := count(`^OpKeyword\w+Expr[ \t]*::=`)
	lexemeLines := count(`^:lexeme`)
	for _, c := range []struct {
		what      string
		got, want int
		pinned    int
	}{
		{"G1 statements", statements(g.G1, ""), g1Lines, 436},
		{"G0 statements", statements(g.G0, ""), g0Lines, 397},
		{"OpKeyword*Expr statements", statements(g.G1, "OpKeyword"), kwLines, 198},
		{"priorities", len(g.Priority), lexemeLines, 214},
	} {
		if c.got != c.want || c.want != c.pinned {
			t.Errorf("%s: reader %d, line count %d, pinned %d", c.what, c.got, c.want, c.pinned)
		}
	}

	pipes := 0
	strip := regexp.MustCompile(`'[^'\n]*'|\[(?:[^\]\\\n]|\\.)*\]|#.*`)
	for _, line := range strings.Split(string(src), "\n") {
		pipes += strings.Count(strip.ReplaceAllString(line, ""), "|")
	}
	if got, want := len(g.G1)+len(g.G0), g1Lines+g0Lines+pipes; got != want {
		t.Errorf("alternatives: reader %d, statements plus pipes %d", got, want)
	}

	if !g.LATM {
		t.Error("LATM = false, want lexeme default = latm => 1")
	}
	if want := []string{"whitespace", "hash comment"}; !reflect.DeepEqual(g.Discard, want) {
		t.Errorf("Discard = %v, want %v", g.Discard, want)
	}
}

// statements counts distinct statements among rules whose LHS starts with
// prefix and, for a non-empty prefix, ends in Expr.
func statements(rules []Rule, prefix string) int {
	type stmt struct {
		lhs  string
		decl int
	}
	seen := map[stmt]bool{}
	for _, r := range rules {
		if prefix != "" && !(strings.HasPrefix(r.LHS, prefix) && strings.HasSuffix(r.LHS, "Expr")) {
			continue
		}
		seen[stmt{r.LHS, r.Decl}] = true
	}
	return len(seen)
}
