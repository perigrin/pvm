// ABOUTME: Lexical sub declarations -- my sub, our sub, state sub -- keep their name,
// ABOUTME: prototype or signature, and body; the declarator does not make them anonymous.
package parse_test

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestLexicalSubDeclaration holds `my sub NAME`, `our sub NAME` and
// `state sub NAME`. The declarator path handed `sub` to the anonymous-sub
// term, which takes no name, so the name fell out as a call:
//
//	our sub h;       canon  our sub();h();      at Unknown=0
//	my sub e ($);    canon  my sub();e ($);     Unknown=1
//
// Each is valid on 5.42.0. Lexical subs need no feature since 5.26; `state
// sub` needs the `state` feature, and `($x)` is a signature only under the
// `signatures` feature and a prototype otherwise. `use v5.36` turns on both:
//
//	$ perl -e 'use v5.36; my sub sq ($x) { $x * $x } print sq(3), "\n"'
//	9
func TestLexicalSubDeclaration(t *testing.T) {
	cases := []struct {
		name  string
		src   string
		canon string
	}{
		{"our forward declaration", "our sub h;", "our sub h;"},
		{"my with a prototype", "my sub e ($);", "my sub e ($);"},
		{"my with a body", "my sub f { 1 }", "my sub f {1;}"},
		{"state with a body", "state sub g { 2 }", "state sub g {2;}"},
		{"our with a body", "our sub d { 3 }", "our sub d {3;}"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			src := []byte(c.src)
			n := parse.Parse(src)
			if got := countUnknown(n); got != 0 {
				t.Errorf("Parse(%q) has %d Unknown, want 0", c.src, got)
			}
			got := strings.TrimSpace(parse.Canon(n, src))
			if got != c.canon {
				t.Errorf("Canon(%q):\n  got  %s\n  want %s", c.src, got, c.canon)
			}
			again := strings.TrimSpace(parse.Canon(parse.Parse([]byte(got)), []byte(got)))
			if again != got {
				t.Errorf("canon is not a fixpoint:\n  once  %s\n  twice %s", got, again)
			}
		})
	}

	// A lexical sub is declared for the calls below it, as `sub NAME` is: a
	// parenless call with a list takes the whole list.
	src := []byte("my sub f { 1 }\nf 1, 2;\n")
	if got := countUnknown(parse.Parse(src)); got != 0 {
		t.Errorf("Parse(%q) has %d Unknown: the lexical sub was not declared", src, got)
	}

	// REGRESSION GUARD: an anonymous sub stays anonymous.
	anon := []byte("my $f = sub { 1 };")
	if got := strings.TrimSpace(parse.Canon(parse.Parse(anon), anon)); got != "my $f = sub {1;};" {
		t.Errorf("Canon(%q) = %s, want the anonymous sub unchanged", anon, got)
	}
}
