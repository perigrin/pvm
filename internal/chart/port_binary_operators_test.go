// ABOUTME: Tests the low-precedence logical operators and the binary operators: expr's and/or/xor and termbinop's infix forms.
// ABOUTME: Snippets under testdata/port/binary-operators are recognized through internal/lexer, chart/lex and the recognizer; the -refused ones refuse.

package chart

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
)

// binaryOperatorsLines are the perly.y alternatives this family ports. The PLUGIN_* ones (1256, 1260,
// 1413, 1417, 1421, 1428, 1432, 1448, 1452, 1458) are ported as grammar and have no snippet: their
// terminals have no token class until waypoint 2 decides how a library operator reaches them.
var binaryOperatorsLines = []int{1254, 1256, 1258, 1260, 1413, 1417, 1419, 1421, 1423, 1428, 1430,
	1432, 1434, 1440, 1442, 1444, 1446, 1448, 1450, 1452, 1454, 1456, 1458}

func TestPortBinaryOperators(t *testing.T) {
	if err := grammar.CheckFamilyMarked(binaryOperatorsLines, nil); err != nil {
		t.Error(err)
	}
	for _, f := range snippets(t, "binary-operators") {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if v := recognizeSource(t, src); !v.accepted {
			t.Errorf("%s is refused at byte %d (expected %v); perl accepts it", f, v.at, v.expected)
		}
	}
}

func TestPortBinaryOperatorsRefusals(t *testing.T) {
	rows := map[string]struct {
		at        string // the refused byte is the first occurrence of this text
		neighbour string // an accepted form beside it
	}{
		"dangling-additive.pl":       {";", "$a + 1;"},
		"dangling-low-and.pl":        {";", "$a and $b;"},
		"dangling-low-or.pl":         {";", "$a or $b;"},
		"dangling-or-or.pl":          {";", "$a || $b;"},
		"dangling-caret-caret.pl":    {";", "$a ^^ $b;"},
		"dangling-multiplicative.pl": {";", "$a * $b;"},
		"dangling-match.pl":          {";", "$a =~ $b;"},
		"dangling-pow.pl":            {";", "$a ** $b;"},
		"dangling-range.pl":          {";", "$a .. $b;"},
		"fat-comma-and.pl":           {"and", "$a and $b;"},
		"leading-infix.pl":           {"==", "$a == $b;"},
		"leading-low-and.pl":         {"and", "$a and $b;"},
		"leading-or-or.pl":           {"||", "$a || $b;"},
	}
	for name, c := range rows {
		src, err := os.ReadFile(filepath.Join("testdata", "port", "binary-operators-refused", name))
		if err != nil {
			t.Fatal(err)
		}
		if v := recognizeSource(t, src); v.accepted {
			t.Errorf("%s is accepted; perl refuses it", name)
		} else if want := strings.Index(string(src), c.at); v.at != want {
			t.Errorf("%s is refused at byte %d, want %d", name, v.at, want)
		}
		if v := recognizeSource(t, []byte(c.neighbour)); !v.accepted {
			t.Errorf("%s's accepted neighbour %q is refused at %d", name, c.neighbour, v.at)
		}
	}
	if len(snippets(t, "binary-operators-refused")) != len(rows) {
		t.Error("a refused snippet has no row here")
	}
}
