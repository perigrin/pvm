// ABOUTME: `print $fh <<"EOF"`: a scalar right after a filehandle-taking builtin is a handle
// ABOUTME: when what follows it starts a term, so the next token lexes in term position.

package lexer

import "testing"

// TestFilehandleSlotLeavesTermPosition holds toke.c's yyl_dollar heuristic for
// a scalar in a list operator's first slot, followed by whitespace. perl peeks
// at the next character and, when it can only start a term, sets XTERM --
// `print $fh <<"EOF"` is a heredoc, not a left shift. Measured on 5.42.0 with
// -MO=Deparse:
//
//	print $f <<"EOT"; ...   print $f "x\n";      heredoc
//	print $f -1;            print $f -1;         a negative number
//	print $f /2/;           print $f /2/;        a pattern
//	print $x - 1;           print $x - 1;        subtraction: space after `-`
//	print $x == 1;          print $x == 1;       comparison
//
// Seven perl.git t/ files write `print $fh <<...`, and the heredoc lexed as a
// shift operator swallowed its body into the statement after it.
func TestFilehandleSlotLeavesTermPosition(t *testing.T) {
	term := []struct{ src, want string }{
		{"print $f <<\"EOT\";\nx\nEOT\n", "<<\"EOT\""},
		{"print $f <<'EOT';\nx\nEOT\n", "<<'EOT'"},
		{"print $f /2/;", "/2/"},
		{"printf $f <<EOT;\nx\nEOT\n", "<<EOT"},
		// Parenthesised, which is canon's spelling: perl applies the same
		// heuristic, measured `print($f <<"EOT")` is a heredoc.
		{"print($f <<EOT);\nx\nEOT\n", "<<EOT"},
	}
	for _, c := range term {
		toks := significant(Tokenize([]byte(c.src)))
		// The token after the handle variable.
		at := -1
		for i, tok := range toks {
			if tok.Kind == Variable && i+1 < len(toks) {
				at = i + 1
				break
			}
		}
		if at < 0 {
			t.Errorf("%q: no token after the handle", c.src)
			continue
		}
		got := c.src[toks[at].Start:toks[at].End]
		if got != c.want || toks[at].Kind == Operator {
			t.Errorf("%q: after the handle %v %q, want the term %q", c.src, toks[at].Kind, got, c.want)
		}
	}

	// An operator position still reads as one: a space after the operator,
	// or an operator no term can start with.
	for _, src := range []string{"print $x - 1;", "print $x == 1;", "my $y = $x <<2;"} {
		toks := significant(Tokenize([]byte(src)))
		op := toks[2]
		if src == "my $y = $x <<2;" {
			op = toks[4]
		}
		if op.Kind != Operator {
			t.Errorf("%q: %v %q is an operator here", src, op.Kind, src[op.Start:op.End])
		}
	}
}
