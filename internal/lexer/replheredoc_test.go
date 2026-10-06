// ABOUTME: A heredoc opened in a multi-line /e replacement takes its body from the
// ABOUTME: replacement's own lines; only a single-line one reads past the statement.

package lexer

import "testing"

// TestReplacementHeredocBodyInside: toke.c's scan_heredoc, inside a
// quote-like operator, looks for a newline after the opener in the
// construct's own buffer first and climbs to the parent's lines only when
// there is none (toke.c:11735-11738). Measured on 5.42.0:
//
//	$ cat he.pl
//	$_ = "";
//	s|(?:)|"${\<<END}"
//	ok inner
//	END
//	|e;
//	print $_;
//	print "after\n";
//	$ perl he.pl
//	ok inner
//	after
//
// perl.git t/base/lex.t:323.
func TestReplacementHeredocBodyInside(t *testing.T) {
	src := "$_ = \"\";\ns|(?:)|\"${\\<<END}\"\nok inner\nEND\n|e;\nprint $_;\nprint \"after\\n\";\n"
	for _, tok := range Tokenize([]byte(src)) {
		if tok.Kind == HeredocBody || tok.Kind == UnknownRest {
			t.Errorf("%v %q: the body is inside the replacement", tok.Kind, src[tok.Start:tok.End])
		}
	}
	// A single-line replacement still reads its body after the line.
	one := "s/^not /substr(<<EOF, 0, 0)/e;\n  Ignored\nEOF\nprint 1;\n"
	if _, ok := firstOfKind(one, HeredocBody); !ok {
		t.Errorf("%q: the body follows the line", one)
	}
}
