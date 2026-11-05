// ABOUTME: Tests quote-likes, interpolation and heredocs, each one token: strings, qw, patterns, substitutions, commands, readlines.
// ABOUTME: Snippets under testdata/port/quote-likes are recognized through internal/lexer, chart/lex and the recognizer; quote-likes-refused ones refuse.

package chart

import (
	"os"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/chart/lex"
)

// quoteLikesLines are the perly.y alternatives this family ports or fits. internal/lexer emits a quote-like and
// its interior as one token where toke.c sublexes it, so 1328 and 1699 are fitted to that token and 1797 and
// 1798, the substitution's replacement inside it, are excluded folded-into-token. None is left for the completing issue.
var quoteLikesLines = []int{1328, 1699, 1797, 1798}

func TestPortQuoteLikes(t *testing.T) {
	if err := grammar.CheckFamilyMarked(quoteLikesLines, nil); err != nil {
		t.Error(err)
	}
	// Bare statements, which hold no prefix the reference refuses, so they sit here rather than in a snippet.
	for _, src := range []string{"'abc';", "s/a/b/;", "m/a/;", "`ls`;", "<STDIN>;", "print <<EOF;\nhi\nEOF\n"} {
		if v := recognizeSource(t, []byte(src)); !v.accepted {
			t.Errorf("%q is refused at %d, expecting %v", src, v.at, v.expected)
		}
	}
	for _, f := range snippets(t, "quote-likes") {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if v := recognizeSource(t, src); !v.accepted {
			t.Errorf("%s is refused at byte %d (expected %v); perl accepts it", f, v.at, v.expected)
		}
	}
}

func TestPortQuoteLikesRefusals(t *testing.T) {
	for _, f := range snippets(t, "quote-likes-refused") {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if v := recognizeSource(t, src); v.accepted {
			t.Errorf("%s is accepted; perl refuses it", f)
		}
	}
	// The accepted form beside each refusal.
	for _, src := range []string{`"abc";`, "qw(a b);", "$x =~ s/a/b/;", "$x =~ m/a/;", "tr/a/b/;", "print <<EOF;\nhi\nEOF\n"} {
		if v := recognizeSource(t, []byte(src)); !v.accepted {
			t.Errorf("%q is refused at %d, expecting %v", src, v.at, v.expected)
		}
	}
	// Spec §2's parity limit: interpolated code is inside its one token, so perl's refusal of it is not seen. It shows
	// in chart-accepted-perl-refused, never as a refusal; this fails when the limit goes.
	if src := `$y = "@{[ 1 + ]}";`; !recognizeSource(t, []byte(src)).accepted {
		t.Errorf("%q is refused; the parity limit is gone, so move it to quote-likes-refused and drop the README line", src)
	}
}

// A heredoc's body is trivia on its opener: chart/lex keeps it out of the token stream, so a body that is not Perl
// is accepted and the statement after the terminator is read as a statement.
func TestPortQuoteLikesHeredocBodyIsTrivia(t *testing.T) {
	body := "print <<EOF;\n} ) ;;\nEOF\n"
	if v := recognizeSource(t, []byte(body+"$x = 1;\n")); !v.accepted {
		t.Errorf("a heredoc whose body is not Perl is refused at %d, expecting %v", v.at, v.expected)
	}
	tab, err := committed()
	if err != nil {
		t.Fatal(err)
	}
	src := []byte(body + "$x = ;\n")
	s, _ := lex.Lex(src, tab.Tokens)
	if len(s.Bodies) != 1 || len(s.Tokens) != 6 {
		t.Errorf("the stream holds %d tokens and %d bodies, want 6 (print, <<EOF, ;, $x, =, ;) and 1", len(s.Tokens), len(s.Bodies))
	}
	if v := recognizeSource(t, src); v.accepted || v.at != strings.Index(string(src), "= ;")+2 {
		t.Errorf("the statement after the terminator, `$x = ;`, is not refused at its ; (accepted %v, at %d)", v.accepted, v.at)
	}
	// Known limit (the shortcut in chart/lex's Map): a body whose opener is inside a quote-like goes to the token before
	// it, here the ;, not to the string. This fails when the limit goes; drop the README line and the shortcut comment then.
	s, _ = lex.Lex([]byte("$y = \"@{[ <<EOF ]}\";\nhi\nEOF\n"), tab.Tokens)
	if len(s.Bodies) != 1 || s.Bodies[0].Opener != 3 {
		t.Errorf("the bodies are %v, want one under token 3, the ; before it", s.Bodies)
	}
	// A body goes to the opener its terminator names, not to the first waiting one: B's is the string's, A's is <<A's.
	s, _ = lex.Lex([]byte("print \"@{[ <<B ]}\" . <<A;\nb\nB\na\nA\n"), tab.Tokens)
	if len(s.Bodies) != 2 || s.Bodies[0].Opener != 4 || s.Bodies[1].Opener != 3 {
		t.Errorf("the bodies are %v, want B's under token 4 (the ; before it) and A's under token 3 (<<A)", s.Bodies)
	}
}
