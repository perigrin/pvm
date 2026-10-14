// ABOUTME: `*` where a term is expected, touching a punctuation byte, is the glob of that
// ABOUTME: punctuation variable -- `*|` is $|'s glob, as `$|` is its scalar -- never an operator.
package parse_test

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// findTerm returns the first Term in n whose Text is text, or nil.
func findTerm(n *parse.Node, text string) *parse.Node {
	if n.Kind == parse.Term && n.Text == text {
		return n
	}
	for _, c := range n.Children {
		if f := findTerm(c, text); f != nil {
			return f
		}
	}
	return nil
}

// findBinary returns the first Binary in n whose Text is op, or nil.
func findBinary(n *parse.Node, op string) *parse.Node {
	if n.Kind == parse.Binary && n.Text == op {
		return n
	}
	for _, c := range n.Children {
		if f := findBinary(c, op); f != nil {
			return f
		}
	}
	return nil
}

// TestGlobOfPunctuationVariable holds perl's reading of a glob whose name is
// one punctuation byte. Measured on 5.42.0 with -MO=Deparse, every one of
// these deparses as written:
//
//	my $x = *|;   *OUTPUT_AUTOFLUSH = *|;   *OFS = *,;   f(*|, *,);
//	my $x = *;;   [*,, *;];                 *|{SCALAR};
//
// English.pm:169 is `*OUTPUT_AUTOFLUSH = *|;`; reading the `|` as bitwise-or
// put a wrong tree under the lowered path. Where an operator is expected the
// star is multiplication: `$a *0` and `$a *- $b` deparse as `$a * 0` and
// `$a * -$b`, while `$a *| $b` and `$a *, $b` are syntax errors.
func TestGlobOfPunctuationVariable(t *testing.T) {
	// Every single punctuation byte perl 5.42.0 accepts after a term-position
	// `*`, each measured to deparse as `my $x = *P;`. `{` is not one: `*{`
	// opens `*{EXPR}`.
	for _, punct := range strings.Split(`| / , ; ! @ & " . \ - + ? $ < > ( ) [ ] ^ : ~ = % # ' } *`, " ") {
		src := "my $x = *" + punct + ";"
		n := parse.Parse([]byte(src))
		if got := countUnknown(n); got != 0 {
			t.Errorf("Parse(%q) has %d Unknown, want 0", src, got)
		}
		if findTerm(n, "*"+punct) == nil {
			t.Errorf("Parse(%q) has no glob term *%s", src, punct)
		}
		got := strings.TrimSpace(parse.Canon(n, []byte(src)))
		if got != src {
			t.Errorf("Canon(%q):\n  got  %s\n  want %s", src, got, src)
		}
	}

	globs := []struct {
		src, canon string
		globs      []string
	}{
		{"*OUTPUT_AUTOFLUSH = *|;", "*OUTPUT_AUTOFLUSH = *|;", []string{"*|"}},
		{"*OFS = *,;", "*OFS = *,;", []string{"*,"}},
		{"f(*|, *,);", "f(*| , *,);", []string{"*|", "*,"}},
		{"my $x = [*,, *;];", "my $x = [*,, *;];", []string{"*,", "*;"}},
		{"local *, = *;;", "local *, = *;;", []string{"*,", "*;"}},
		{"my $x = *0;", "my $x = *0;", []string{"*0"}},
		{"my $x = *^W;", "my $x = *^W;", []string{"*^W"}},
		{"my $x = *{^GLOBAL_PHASE};", "my $x = *{^GLOBAL_PHASE};", nil},
	}
	for _, c := range globs {
		n := parse.Parse([]byte(c.src))
		if got := countUnknown(n); got != 0 {
			t.Errorf("Parse(%q) has %d Unknown, want 0", c.src, got)
		}
		for _, g := range c.globs {
			if findTerm(n, g) == nil {
				t.Errorf("Parse(%q) has no glob term %s", c.src, g)
			}
		}
		if b := findBinary(n, "|"); b != nil {
			t.Errorf("Parse(%q) reads a bitwise-or", c.src)
		}
		got := strings.TrimSpace(parse.Canon(n, []byte(c.src)))
		if got != c.canon {
			t.Errorf("Canon(%q):\n  got  %s\n  want %s", c.src, got, c.canon)
		}
		again := strings.TrimSpace(parse.Canon(parse.Parse([]byte(got)), []byte(got)))
		if again != got {
			t.Errorf("canon is not a fixpoint:\n  once  %s\n  twice %s", got, again)
		}
	}

	// Where an operator is expected the star stays multiplication.
	for _, src := range []string{"$x = $a *0;", "$x = $a *$b;", "$x = $a *- $b;", "$x = $a *(2);", "$y = $x-- *2;"} {
		n := parse.Parse([]byte(src))
		if got := countUnknown(n); got != 0 {
			t.Errorf("Parse(%q) has %d Unknown, want 0", src, got)
		}
		if findBinary(n, "*") == nil {
			t.Errorf("Parse(%q) has no multiplication", src)
		}
	}
	// And `**` and `*=` stay one operator each there, after a postfix `++`
	// as after any term.
	for _, src := range []string{"$x = $a ** 2;", "$x *= 2;", "$x **= 2;", "$y = $x++ ** 2;"} {
		n := parse.Parse([]byte(src))
		if got := countUnknown(n); got != 0 {
			t.Errorf("Parse(%q) has %d Unknown, want 0", src, got)
		}
		if got := strings.TrimSpace(parse.Canon(n, []byte(src))); got != src {
			t.Errorf("Canon(%q):\n  got  %s\n  want %s", src, got, src)
		}
	}
}
