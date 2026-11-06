// ABOUTME: Tests the remaining term rules: anonymous arrays and hashes, do, the ternary, refgen, local, parentheses, qw lists and &-calls.
// ABOUTME: Snippets under testdata/port/terms are recognized through internal/lexer, chart/lex and the recognizer; the -refused ones refuse.

package chart

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/chart/lex"
)

// termsLines are the perly.y alternatives this family ports or fits; 1536 is fitted without the ;
// before the } that toke.c inserts. None is left for the completing issue.
var termsLines = []int{1534, 1536, 1558, 1560, 1566, 1567, 1568, 1570, 1574, 1576, 1578, 1580, 1638, 1641}

func TestPortTerms(t *testing.T) {
	if err := grammar.CheckFamilyMarked(termsLines, nil); err != nil {
		t.Error(err)
	}
	for _, f := range snippets(t, "terms") {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if v := recognizeSource(t, src); !v.accepted {
			t.Errorf("%s is refused at byte %d (expected %v); perl accepts it", f, v.at, v.expected)
		}
	}
	// The bare-term statements, which have no refused prefix for the truncation test to cut at.
	for _, src := range []string{"[1, 2];", "+{ foo => 1 };", "do { 1 };", "$c ? 1 : 2;", "\\$x;", "local $x;", "(1);", "();", "qw(a b);", "&foo();"} {
		if v := recognizeSource(t, []byte(src)); !v.accepted {
			t.Errorf("%q is refused at %d (expected %v)", src, v.at, v.expected)
		}
	}
}

// A v-string is a THING wherever a term stands (toke.c:9879-9887, and scan_num's two dots, toke.c:13135-13137), and
// the version, never a term, after use MODULE.
func TestPortTermsVString(t *testing.T) {
	for _, tc := range []struct {
		src  string
		at   int
		want bool
	}{
		{"$y = v1.2.3;", 5, true}, {"$y = 1.2.3;", 5, true}, {"print v65.66;", 6, true}, {"use Foo v1.2;", 8, false},
	} {
		if got := slices.Contains(terminalsAt(t, tc.src, tc.at), "THING"); got != tc.want {
			t.Errorf("%q at %d: THING %v, want %v", tc.src, tc.at, got, tc.want)
		}
	}
	for _, src := range []string{"my @a = (v1.2.3, 4.5.6);", "print v65.66;", "$y = v1.2 x 3;"} {
		if v := recognizeSource(t, []byte(src)); !v.accepted {
			t.Errorf("%q is refused at %d (expected %v)", src, v.at, v.expected)
		}
	}
}

func TestPortTermsRefusals(t *testing.T) {
	rows := map[string]struct {
		at        string // the refused byte is the first occurrence of this text, the last after a ^; "" is end of file
		neighbour string // an accepted form beside it
	}{
		"unclosed-array.pl":         {";", "[1, 2];"},
		"ternary-no-colon.pl":       {";", "$c ? 1 : 2;"},
		"ternary-no-middle.pl":      {":", "$c ? 1 : 2;"},
		"refgen-no-operand.pl":      {";", "\\$x;"},
		"hash-written-semicolon.pl": {";", "+{ foo => 1 };"},
		"array-two-terms.pl":        {"2", "$y = [1, 2];"},
		"do-no-operand.pl":          {";", "$y = do { 1 };"},
		"local-no-operand.pl":       {";", "$y = local $x;"},
		"unclosed-paren.pl":         {";", "$y = (1, 2);"},
		"unclosed-call.pl":          {";", "$y = &foo(1);"},
		"qwlist-two-terms.pl":       {"qw(b)", "$y = qw(a);"},
		// toke.c's yyl_colon (toke.c:6456 at 94e50866) makes a : after my, our or state (yyl_my, toke.c:7572)
		// and a variable an attribute colon when an identifier character follows, so the ternary lacks its colon.
		"colon-my-word.pl":       {";", "$y = $c ? my $x : 1;"},
		"colon-my-word-tight.pl": {";", "$y = $c ? my $x : 1;"},
		"colon-my-keyword.pl":    {";", "$y = $c ? my $x : 1;"},
		"colon-my-qw.pl":         {"qw", "$y = $c ? my $x : 1;"},
		"colon-state-word.pl":    {"^;", "use feature 'state'; $y = $c ? state $x : 1;"},
		"colon-our-word.pl":      {";", "$y = $c ? our @x : 1;"},
	}
	for name, c := range rows {
		src, err := os.ReadFile(filepath.Join("testdata", "port", "terms-refused", name))
		if err != nil {
			t.Fatal(err)
		}
		want := len(src)
		if last, ok := strings.CutPrefix(c.at, "^"); ok {
			want = strings.LastIndex(string(src), last)
		} else if c.at != "" {
			want = strings.Index(string(src), c.at)
		}
		if v := recognizeSource(t, src); v.accepted {
			t.Errorf("%s is accepted; perl refuses it", name)
		} else if v.at != want {
			t.Errorf("%s is refused at byte %d, want %d", name, v.at, want)
		}
		if v := recognizeSource(t, []byte(c.neighbour)); !v.accepted {
			t.Errorf("%s's accepted neighbour %q is refused at %d", name, c.neighbour, v.at)
		}
	}
	if len(snippets(t, "terms-refused")) != len(rows) {
		t.Error("a refused snippet has no row here")
	}
}

// An attribute colon is told from the ternary's by what precedes and follows it, as toke.c does.
func TestPortTermsAttributeColon(t *testing.T) {
	for _, src := range []string{
		"$y = $c ? my $x : my $z;", // perl: Invalid separator character, not a syntax error
		"$y = $c ? my $x : x;", "$y = $c ? my $x : Foo::bar;", "$y = $c ? my $x : q(a);", "$y = $c ? my $x : s/a/b/;",
		"use utf8; $y = $c ? my $x : é;", "my $x : Foo ::bar baz;", // perl: Invalid separator character
		"$y = $c ? our $x : shared;", "$y = $c ? my @x : shared;", "$y = $c ? my %x : shared;",
	} {
		if v := recognizeSource(t, []byte(src)); v.accepted {
			t.Errorf("%q is accepted; perl refuses it", src)
		}
	}
	for _, src := range []string{
		"$y = $c ? my $x : 1;", "$y = $c ? my $x : $z;", "$y = $c ? my $x : -1;", `$y = $c ? my $x : "s";`,
		"$y = $c ? my $x : \\$z;", "$y = $c ? my $x : (my $z);", "$y = $c ? my $x : ::foo;", "use utf8; my $x :Foo é bar;", "$y = $c ? 1 : ::foo;", "$y = $c ? my $x : [1];", "$y = $c ? my $x : @a;",
		"$y = $c ? (my $x) : shared;", "$y = $c ? local $g : shared;", "$y = $c ? $x : shared;", "$y = $c ? my $x\n: 1;",
		"my $x : shared;", "my $x :shared;", "$y = $c ? shared : my $x;", "$y = $c ? my $x : 'a';",
	} {
		if v := recognizeSource(t, []byte(src)); !v.accepted {
			t.Errorf("%q is refused at %d (expected %v); perl accepts it", src, v.at, v.expected)
		}
	}
	// Known limit: a colon that separates attributes is eaten by toke.c's attribute loop, so it is never the
	// ternary's, but the key sees a word and a colon before it here and in $a ? $b ? 1 : c : d, which perl
	// accepts. These are accepted where perl refuses them ("Invalid separator character") until the lexer
	// marks the separator (v0.1 lexer issue 01a1a314-e3f8-765f-abf2-159425ed764e).
	for _, src := range []string{
		"$y = $c ? my $x : shared : 1;", "$y = $c ? my $x :shared :1;", "$y = $c ? my $x : a : b : 1;",
		"$y = $c ? my $x : foo(1) : 2;", "$y = $c ? our @x : shared : ();", "$c ? my $x : L: 1;",
	} {
		if v := recognizeSource(t, []byte(src)); !v.accepted {
			t.Errorf("%q is refused at %d: the separator-colon limit is gone, so remove the shortcut and make this a refusal", src, v.at)
		}
	}
	// Known limit: the key sees two tokens back, so my ($x) : shared looks like (my $x) : shared, which
	// perl accepts. Perl refuses it; this pins the over-acceptance until the lexer marks the ) that closes
	// a my list (v0.1 lexer issue 01a1a314-e3f8-765f-abf2-159425ed764e), when the chart can follow toke.c.
	for _, src := range []string{"$y = $c ? my ($x) : shared;", "$y = $c ? my ($x, $z) : shared;", "$y = $c ? my () : shared;"} {
		if v := recognizeSource(t, []byte(src)); !v.accepted {
			t.Errorf("%q is refused at %d: the my (...) limit is gone, so remove the shortcut and make this a refusal", src, v.at)
		}
	}
}

// terminalsOf is the names of the terminals admitting the nth (from 0) token of src whose text is text.
func terminalsOf(t *testing.T, src, text string, n int) []string {
	t.Helper()
	tab, err := committed()
	if err != nil || tab == nil {
		t.Fatalf("the committed grammar generates no tables: %v", err)
	}
	s, _ := lex.Lex([]byte(src), tab.Tokens)
	for _, tok := range s.Tokens {
		if raw := s.Raw[tok.Index]; src[raw.Start:raw.End] == text {
			if n--; n >= 0 {
				continue
			}
			var names []string
			for _, id := range tab.TerminalSets[tok.TerminalSet] {
				names = append(names, tab.Symbols[id].Name)
			}
			return names
		}
	}
	t.Fatalf("no such token %q in %q", text, src)
	return nil
}

// A word inside +{ ... } keeps the call reading its terminals give it, and the brace itself is the
// anonymous hash's, which a subscript's is not; so the two braces key the word after them differently.
func TestPortTermsHashWordIsCall(t *testing.T) {
	for _, src := range []string{"$y = +{ time };", "$y = +{ time => 1 };"} {
		if v := recognizeSource(t, []byte(src)); !v.accepted {
			t.Errorf("%q is refused at %d (expected %v)", src, v.at, v.expected)
		}
	}
	if got := terminalsOf(t, "$y = +{ time };", "time", 0); !slices.Contains(got, "FUNC0") {
		t.Errorf("time inside +{ } admits %v, want the FUNC0 call", got)
	}
	if got := terminalsOf(t, "$y = +{ time };", "{", 0); !slices.Equal(got, []string{"HASHBRACK"}) {
		t.Errorf("the { of +{ } admits %v, want HASHBRACK", got)
	}
	// Each row is a { that ClassOf reads as a subscript's or a block's, and which one of the source's.
	for _, c := range []struct {
		src string
		n   int
	}{
		{"$h{time};", 0}, {"$h->{time};", 0}, {"$h[0]{time};", 0}, {"$h{a}{time};", 1},
		{"${ $r }{time};", 0}, {"${ $r }{time};", 1}, {"@{ $r }{time};", 0}, {"@{ $r }{time};", 1},
		{"do { time };", 0}, {"if ($c) { time }", 0},
	} {
		if got := terminalsOf(t, c.src, "{", c.n); slices.Contains(got, "HASHBRACK") {
			t.Errorf("the { number %d of %q admits %v, only an anonymous hash's brace is HASHBRACK", c.n, c.src, got)
		}
	}
}
