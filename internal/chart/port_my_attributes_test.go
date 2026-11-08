// ABOUTME: Tests the my-declaration and attribute-list rules: my/our/state, their list and reference forms, and the attributes that follow.
// ABOUTME: Snippets under testdata/port/my-attributes are recognized through internal/lexer, chart/lex and the recognizer; my-attributes-refused ones refuse.

package chart

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
)

// myAttributesLines are the perly.y alternatives this family ports or fits. Every one is ported:
// none refuses unconditionally, so none is left for a completing issue.
// An attribute's argument is one token (`:Foo(bar)`): the lexer scans it as Perl, where toke.c takes
// it as one opaque string, so `:Foo(a, b)`, `:Foo(1+2)` and `:Foo(())` are refused until 01a19fb5-79d7-76c5-a8b1-25fd04ddfb4a.
// The same issue covers attribute names: pvm emits `:Foo::Bar` and `:Foo'Bar` as one qualified word, which
// the chart accepts and perl refuses as an invalid separator.
var myAttributesLines = []int{1070, 1072, 1572, 1719, 1721, 1723, 1725, 1730, 1732, 1735, 1737, 1739}

func TestPortMyAndAttributes(t *testing.T) {
	if err := grammar.CheckFamilyMarked(myAttributesLines, nil); err != nil {
		t.Error(err)
	}
	for _, f := range snippets(t, "my-attributes") {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if v := recognizeSource(t, src); !v.accepted {
			t.Errorf("%s is refused at byte %d (expected %v); perl accepts it", f, v.at, v.expected)
		}
	}
}

// TestPortMyAndAttributesRefusals: each refused snippet fails at the byte named, with an accepted neighbour.
func TestPortMyAndAttributesRefusals(t *testing.T) {
	for name, c := range map[string]struct {
		at        string // the refused byte is the first occurrence of this text; "" is end of file
		neighbour string
	}{
		"bare-my.pl":          {";", "scalar.pl"},
		"unclosed-list.pl":    {";", "list.pl"},
		"open-paren.pl":       {";", "list.pl"},
		"no-comma.pl":         {"$y", "scalar.pl"},
		"arg-without-name.pl": {"(", "attr-arg.pl"},
		"two-colons.pl":       {":Bar", "trailing-colon.pl"},
		"paren-close-arg.pl":  {");", "attr-arg.pl"},
	} {
		src, err := os.ReadFile(filepath.Join("testdata", "port", "my-attributes-refused", name))
		if err != nil {
			t.Fatal(err)
		}
		if v := recognizeSource(t, src); v.accepted {
			t.Errorf("%s is accepted; perl refuses it", name)
		} else if want := atOrEOF(string(src), c.at); v.at != want {
			t.Errorf("%s is refused at byte %d, want %d", name, v.at, want)
		}
		nb, err := os.ReadFile(filepath.Join("testdata", "port", "my-attributes", c.neighbour))
		if err != nil {
			t.Fatal(err)
		}
		if v := recognizeSource(t, nb); !v.accepted {
			t.Errorf("%s, near %s, is refused at %d", c.neighbour, name, v.at)
		}
	}
}

func atOrEOF(s, sub string) int {
	if sub == "" {
		return len(s)
	}
	return strings.Index(s, sub)
}

// `my $x :Foo((;` is not a snippet: perl refuses it as an unterminated attribute parameter, not as a
// syntax error, so it cannot sit beside the others. The chart refuses the second ( as an argument.
func TestPortMyAndAttributesUnterminatedArgument(t *testing.T) {
	src := "my $x :Foo((;\n"
	if v := recognizeSource(t, []byte(src)); v.accepted || v.at != strings.Index(src, "(;") {
		t.Errorf("%q: accepted=%v, refused at %d, want refused at %d", src, v.accepted, v.at, strings.Index(src, "(;"))
	}
}
