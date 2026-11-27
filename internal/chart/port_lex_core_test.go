// ABOUTME: Tests the core rules the first token classes need: program start, statement and block spine, expression core, variable heads.
// ABOUTME: Snippets under testdata/port/lex-core are recognized through internal/lexer, chart/lex and the recognizer; lex-core-refused ones refuse.

package chart

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"

	"tamarou.com/pvm/internal/chart/forest"
	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/chart/lex"
	"tamarou.com/pvm/internal/chart/recognize"
)

// lexCoreLines are the perly.y alternatives the core ports or fits.
var lexCoreLines = []int{187, 270, 333, 532, 775, 783, 794, 825, 826, 846, 879, 884, 890, 935, 973,
	1262, 1266, 1268, 1273, 1415, 1564, 1582, 1584, 1586, 1588, 1590, 1634, 1636, 1842, 1846, 1850, 1856, 1862, 1868}

var committed = sync.OnceValues(func() (*grammar.Tables, error) {
	g, ref, err := grammar.Committed()
	if err != nil {
		return nil, err
	}
	return grammar.Generate(g, ref)
})

// verdict is a recognition: accepted, or refused at a byte with what was expected there.
type verdict struct {
	accepted bool
	at       int
	expected []string
	by       string // a refusal's cause: "table", chart/lex's lex-unmapped, or "grammar", the recognizer
}

// recognizeSource runs src through chart/lex and the recognizer over the committed tables, un-narrowed. A refusal's
// cause and byte are forest.ChartRefusal's, Parse's rule; Expected is the judged chart's where it died.
func recognizeSource(t *testing.T, src []byte) verdict {
	t.Helper()
	tab, err := committed()
	if err != nil {
		t.Fatalf("the committed grammar generates no tables: %v", err)
	}
	s, unmapped := lex.Lex(src, tab.Tokens)
	l := lex.Linear(s)
	r := recognize.New(tab, recognize.Options{})
	c, err := recognize.Recognize(r, l)
	if err != nil {
		t.Fatal(err)
	}
	if c.Accepted() {
		return verdict{accepted: true}
	}
	// The chart died at the first node with no set, past the token that leads to it, or at end of input.
	pos := len(c.Sets) - 1
	if d := slices.Index(c.Sets, nil); d > 0 {
		pos = d - 1
	}
	exp, err := c.Expected(pos)
	if err != nil {
		t.Fatal(err)
	}
	ref := forest.ChartRefusal(tab, r, c, s, l, len(src), unmapped, false)
	// A token no class admits has the empty set, so the recognizer refuses at it.
	by := ref.Reason
	if by == "lex-unmapped" {
		by = "table"
	}
	return verdict{at: ref.Byte, expected: exp, by: by}
}

// parseSource is src's verdict through forest.Parse, nothing imported: recognition, then precedence and resolve, which
// refuses a word read as a library operator none of which is in scope. A refusal's by is forest.Refusal's reason.
func parseSource(t *testing.T, src []byte) verdict {
	t.Helper()
	tab, err := committed()
	if err != nil {
		t.Fatalf("the committed grammar generates no tables: %v", err)
	}
	if f, ref := forest.Parse(tab, src, forest.Scope{}); f == nil {
		return verdict{at: ref.Byte, by: ref.Reason}
	}
	return verdict{accepted: true}
}

func snippets(t *testing.T, dir string) []string {
	t.Helper()
	fs, err := filepath.Glob(filepath.Join("testdata", "port", dir, "*.pl"))
	if err != nil || len(fs) == 0 {
		t.Fatalf("no snippets under testdata/port/%s: %v", dir, err)
	}
	return fs
}

func TestPortLexCore(t *testing.T) {
	if err := grammar.CheckFamilyMarked(lexCoreLines, nil); err != nil {
		t.Error(err)
	}
	for _, f := range snippets(t, "lex-core") {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if v := recognizeSource(t, src); !v.accepted {
			t.Errorf("%s is refused at byte %d (expected %v); perl accepts it", f, v.at, v.expected)
		}
	}
}

func TestPortLexCoreRefusals(t *testing.T) {
	for name, c := range map[string]struct {
		at        string // the refused byte is the first occurrence of this text; "" is end of file
		expect    string // a terminal Expected must name there, if any
		neighbour string // an accepted snippet beside it
	}{
		"unclosed-block.pl": {"", "PERLY_BRACE_CLOSE", "block.pl"},
		"stray-close.pl":    {"}", "", "assign-thing.pl"},
		"missing-term.pl":   {";", "", "assign-thing.pl"},
		"two-statements.pl": {"$y", "", "eof-no-semicolon.pl"},
		"lone-equals.pl":    {"=", "", "empty.pl"},
	} {
		src, err := os.ReadFile(filepath.Join("testdata", "port", "lex-core-refused", name))
		if err != nil {
			t.Fatal(err)
		}
		want := len(src)
		if c.at != "" {
			want = strings.Index(string(src), c.at)
		}
		v := recognizeSource(t, src)
		switch {
		case v.accepted:
			t.Errorf("%s is accepted; perl refuses it", name)
		case v.at != want:
			t.Errorf("%s is refused at byte %d, want %d", name, v.at, want)
		case c.expect != "" && !slices.Contains(v.expected, c.expect):
			t.Errorf("%s: Expected %v lacks %s", name, v.expected, c.expect)
		}
		nb, err := os.ReadFile(filepath.Join("testdata", "port", "lex-core", c.neighbour))
		if err != nil {
			t.Fatal(err)
		}
		if !recognizeSource(t, nb).accepted {
			t.Errorf("%s's accepted neighbour %s is refused", name, c.neighbour)
		}
	}
	if len(snippets(t, "lex-core-refused")) != 5 {
		t.Error("a refused snippet has no row here")
	}
}

// The first attempt's Program did not derive the empty program.
func TestPortLexCoreEmptyProgram(t *testing.T) {
	for _, src := range []string{"", "# only a comment\n", "\n\n"} {
		if v := recognizeSource(t, []byte(src)); !v.accepted {
			t.Errorf("%q is refused at %d", src, v.at)
		}
	}
	if v := recognizeSource(t, []byte("=")); v.accepted {
		t.Error("a lone = is accepted")
	}
}

// toke.c inserts a ; before } and at end of file; pvm emits none, so the recognizer inserts it.
func TestPortLexCoreInsertedSemicolon(t *testing.T) {
	for _, src := range []string{"{ $x = 1 }", "$x = 1; $y = 2", "{ $x = 1; $y = 2 }\n$z = 3\n"} {
		if v := recognizeSource(t, []byte(src)); !v.accepted {
			t.Errorf("%q is refused at %d", src, v.at)
		}
	}
	for _, src := range []string{"$x = 1 $y = 2;", "{ $x = 1 $y = 2 }", "$x = 1 $y = 2"} {
		if v := recognizeSource(t, []byte(src)); v.accepted {
			t.Errorf("%q, two statements with no ; between them, is accepted", src)
		}
	}
}
