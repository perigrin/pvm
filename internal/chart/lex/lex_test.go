// ABOUTME: Tests chart/lex's up-front stream: lossless over the source, each token mapped to the cut's predicates.
// ABOUTME: Covers the $#{BLOCK} split, the edge cases of losslessness, an unmapped kind, and the context facts.

package lex

import (
	"bytes"
	"slices"
	"strconv"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/lexer"
)

func guacamole(t *testing.T) *Lexer {
	t.Helper()
	x, err := Guacamole("../grammar/testdata/guacamole/guacamole.slif")
	if err != nil {
		t.Fatal(err)
	}
	return x
}

// text is a token's bytes.
func text(src []byte, t lexer.Token) string { return string(src[t.Start:t.End]) }

// TestStreamLosslessAndMapped: the stream concatenates back to the source
// bytes, trivia included, and each token carries the terminals whose
// predicate the cut names for it.
func TestStreamLosslessAndMapped(t *testing.T) {
	x := guacamole(t)
	src := []byte("# a comment\nmy $x = 1;\nprint $x;\n")
	s, err := x.Lex(src)
	if err != nil {
		t.Fatal(err)
	}
	if got := s.Text(src); !bytes.Equal(got, src) {
		t.Errorf("stream is %q, want the source %q", got, src)
	}
	want := map[string]string{"my": "OpKeywordMy", "$x": "ScalarVar", "=": "OpAssign", "1": "NumberDec", ";": "Semicolon", "print": "OpKeywordPrint"}
	for _, tok := range s.Tokens {
		if len(tok.Terminals) == 0 {
			t.Errorf("%q maps to no terminal", text(src, tok.Token))
		}
		if w, ok := want[text(src, tok.Token)]; ok && !slices.Contains(tok.Terminals, w) {
			t.Errorf("%q maps to %v, want %s among them", text(src, tok.Token), tok.Terminals, w)
		}
	}
	if len(s.Tokens) != 8 {
		t.Errorf("%d significant tokens, want 8: trivia is attached, not a token", len(s.Tokens))
	}
}

// TestStreamLosslessEdgeCases: an empty file, a file of only POD and a file
// with no final newline each concatenate back to their bytes.
func TestStreamLosslessEdgeCases(t *testing.T) {
	x := guacamole(t)
	for name, src := range map[string]string{
		"empty":      "",
		"only POD":   "=pod\n\nNothing but documentation.\n\n=cut\n",
		"no newline": "my $x = 1;",
	} {
		s, err := x.Lex([]byte(src))
		if err != nil {
			t.Errorf("%s: %v", name, err)
			continue
		}
		if got := string(s.Text([]byte(src))); got != src {
			t.Errorf("%s: stream is %q, want %q", name, got, src)
		}
	}
}

// TestUnmappedKindNamesOffset: a token whose kind has no predicate is an
// error naming its byte offset, and it stays in the stream. An unterminated
// string is one UnknownRest token, which no predicate admits.
func TestUnmappedKindNamesOffset(t *testing.T) {
	x := guacamole(t)
	src := []byte(`my $x = "abc`)
	s, err := x.Lex(src)
	if err == nil {
		t.Fatal("an UnknownRest token mapped to a predicate")
	}
	if !strings.Contains(err.Error(), "byte 8") || !strings.Contains(err.Error(), "UnknownRest") {
		t.Errorf("error %q does not name UnknownRest at byte 8", err)
	}
	if got := s.Text(src); !bytes.Equal(got, src) {
		t.Errorf("stream is %q: the unmapped token was dropped", got)
	}
}

// TestArrayTopSplit: the lexer's one `$#{...}` Variable token is split as
// the cut splits it, an ArrayTopOpen `$#{`, the interior re-lexed with its
// offsets in the file, and a `}` that RBrace admits.
func TestArrayTopSplit(t *testing.T) {
	x := guacamole(t)
	src := []byte(`my $n = $#{ $foo->{'bar'} };`)
	s, err := x.Lex(src)
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, tok := range s.Tokens {
		got = append(got, text(src, tok.Token))
	}
	want := []string{"my", "$n", "=", "$#{", "$foo", "->", "{", "'bar'", "}", "}", ";"}
	if !slices.Equal(got, want) {
		t.Fatalf("tokens %q, want %q", got, want)
	}
	open, closer := s.Tokens[3], s.Tokens[9]
	if open.Kind != lexer.Variable || !slices.Contains(open.Terminals, "ArrayTopOpen") {
		t.Errorf("$#{ is %v %v, want a Variable ArrayTopOpen admits", open.Kind, open.Terminals)
	}
	if !slices.Contains(closer.Terminals, "RBrace") || closer.Start != len(src)-2 {
		t.Errorf("closer is %v at %d, want RBrace at %d", closer.Terminals, closer.Start, len(src)-2)
	}
	// The interior's spans are the file's bytes: `$foo` starts after `$#{ `.
	if foo := s.Tokens[4]; foo.Start != 12 || foo.End != 16 || len(foo.Lead) != 1 {
		t.Errorf("$foo spans %d..%d with %d trivia, want 12..16 after one space", foo.Start, foo.End, len(foo.Lead))
	}
	if got := s.Text(src); !bytes.Equal(got, src) {
		t.Errorf("stream is %q, want the source", got)
	}
}

// TestArrayTopNeedsFileState: a `$#{` interior is re-lexed from a blank
// state, so where the file's state could differ from blank before it the
// file is refused, naming the `$#{`, rather than lexed. perl reads the probe's
// `/2/` as two divisions, because N is a constant; a blank re-lex reads a
// pattern (gate review of #541, F1).
func TestArrayTopNeedsFileState(t *testing.T) {
	x := guacamole(t)
	for _, src := range []string{
		"use constant N => 2; my @n = $#{ [ N /2/ 1 ] };",
		"sub N () { 2 } $#{$r};",
		"use utf8; $#{$r};",
		"use feature 'say'; $#{$r};",
		"no feature 'say'; $#{$r};",
		"use v5.36; $#{$r};",
		"use 5.036; $#{$r};",
		"no 5.036; $#{$r};",
	} {
		_, err := x.Lex([]byte(src))
		at := "byte " + strconv.Itoa(strings.Index(src, "$#{"))
		if err == nil || !strings.Contains(err.Error(), at) || !strings.Contains(err.Error(), "needs the lexer's file state (01a1882d-d8a4)") {
			t.Errorf("%q: got %v, want a refusal at %s naming 01a1882d-d8a4", src, err, at)
		}
	}
	// A `use` that changes no lexer state does not refuse.
	if _, err := x.Lex([]byte("use strict; $#{$r};")); err != nil {
		t.Errorf("use strict before $#{: %v", err)
	}
}

// TestArrayTopSplitBadInterior: `$#{1+}` splits too, and its interior is
// `1 +`, a statement the grammar refuses: nothing here judges the block.
func TestArrayTopSplitBadInterior(t *testing.T) {
	x := guacamole(t)
	src := []byte(`$#{1+};`)
	s, err := x.Lex(src)
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, tok := range s.Tokens {
		got = append(got, text(src, tok.Token))
	}
	if want := []string{"$#{", "1", "+", "}", ";"}; !slices.Equal(got, want) {
		t.Errorf("tokens %q, want %q", got, want)
	}
}
