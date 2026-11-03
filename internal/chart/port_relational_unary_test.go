// ABOUTME: Tests the relational, equality and unary operator rules: termrelop, termeqop, their chains, and termunop.
// ABOUTME: Snippets under testdata/port/relational-unary are recognized through internal/lexer, chart/lex and the recognizer; the -refused ones refuse.

package chart

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
)

// relationalUnaryLines are the perly.y alternatives this family ports or fits. 1470 (PLUGIN_REL_OP)
// is ported but has no snippet: the terminal needs an infix plugin loaded. 1512 (POSTJOIN) is
// fitted to a class no token matches, so postjoin.pl is recognized as a string, not through the rule.
var relationalUnaryLines = []int{1436, 1438, 1462, 1464, 1470, 1474, 1476, 1480, 1482, 1490, 1492,
	1497, 1499, 1502, 1504, 1506, 1509, 1512, 1523, 1526, 1565}

// relationalUnaryLeft are the four comparison chains, left for the issue that completes the port:
// each is refused by precedence and takes its ledger mark and refusal test there.
var relationalUnaryLeft = []int{1466, 1468, 1484, 1486}

func TestPortRelationalAndUnary(t *testing.T) {
	all := append(append([]int{}, relationalUnaryLines...), relationalUnaryLeft...)
	if err := grammar.CheckFamilyMarked(all, relationalUnaryLeft); err != nil {
		t.Error(err)
	}
	for _, f := range snippets(t, "relational-unary") {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if v := recognizeSource(t, src); !v.accepted {
			t.Errorf("%s is refused at byte %d (expected %v); perl accepts it", f, v.at, v.expected)
		}
	}
}

func TestPortRelationalAndUnaryRefusals(t *testing.T) {
	rows := map[string]struct {
		at        string // the refused byte is the first occurrence of this text
		neighbour string // an accepted form beside it
	}{
		"dangling-relational.pl": {";", "$a < $b;"},
		"dangling-equality.pl":   {";", "$a == $b;"},
		"bare-not.pl":            {";", "! $a;"},
		"binary-minus.pl":        {";", "$a - $b;"},
		"incdec-then-term.pl":    {"$b", "$a++; $b;"},
	}
	for name, c := range rows {
		src, err := os.ReadFile(filepath.Join("testdata", "port", "relational-unary-refused", name))
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
	if len(snippets(t, "relational-unary-refused")) != len(rows) {
		t.Error("a refused snippet has no row here")
	}
	// A relational chain is accepted by perl; a binary minus is told from a unary one by position.
	for _, src := range []string{"- $a;", "$a - - $b;", "$a < $b < $c;"} {
		if v := recognizeSource(t, []byte(src)); !v.accepted {
			t.Errorf("%q is refused at %d", src, v.at)
		}
	}
}
