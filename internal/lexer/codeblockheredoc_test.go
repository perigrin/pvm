// ABOUTME: A heredoc opened in a pattern's code block, `(?{<<END})`, takes its
// ABOUTME: body from the lines after the statement when the pattern is one line.

package lexer

import "testing"

// TestCodeBlockHeredoc: the code of `(?{ })` and `(??{ })` is lexed as
// perl, and scan_heredoc climbs to the file's lines when the pattern has no
// newline after the opener (toke.c:11735). Measured on 5.42.0:
//
//	$ cat qh.pl
//	print "not " unless qr/(?{<<END})/ eq q{(?^:(?{<<END}))};
//	foo
//	END
//	print "ok\n";
//	m/x(??{<<E2})/;
//	bar
//	E2
//	print "ok2\n";
//	$ perl qh.pl
//	ok
//	ok2
//
// perl.git t/base/lex.t:346.
func TestCodeBlockHeredoc(t *testing.T) {
	src := "print \"not \" unless qr/(?{<<END})/ eq q{(?^:(?{<<END}))};\nfoo\nEND\nprint \"ok\\n\";\n" +
		"m/x(??{<<E2})/;\nbar\nE2\nprint \"ok2\\n\";\n"
	var bodies []string
	for _, tok := range Tokenize([]byte(src)) {
		if tok.Kind == HeredocBody {
			bodies = append(bodies, src[tok.Start:tok.End])
		}
		if tok.Kind == UnknownRest {
			t.Errorf("UnknownRest %q", src[tok.Start:tok.End])
		}
	}
	if len(bodies) != 2 || bodies[0] != "foo\nEND\n" || bodies[1] != "bar\nE2\n" {
		t.Errorf("heredoc bodies %q, want [\"foo\\nEND\\n\" \"bar\\nE2\\n\"]", bodies)
	}
	// A `<<` in the pattern outside a code block is regex text.
	if _, ok := firstOfKind("m/a<<END/;\nprint 1;\n", HeredocBody); ok {
		t.Errorf("`<<END` outside a code block opened a heredoc")
	}
}
