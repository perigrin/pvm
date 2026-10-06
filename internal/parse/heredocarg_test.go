// ABOUTME: A heredoc body can fall INSIDE its statement, when the statement continues past the
// ABOUTME: opener's line: `f(<<EOF,\nbody\nEOF\n 1);` still has an argument after the body.
package parse_test

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestHeredocBodyInsideItsStatement holds the shape perl's own suite calls
// its fresh-perl helpers with:
//
//	fresh_perl_is(<<~'EOF',
//	    ...program...
//	    EOF
//	    "expected", "name");
//
// perl reads a body from the line AFTER the opener, whatever the statement is
// doing, and then resumes the statement where the line broke off. Measured on
// 5.42.0:
//
//	$ printf 'sub f { print join("|", @_), "\n" }\nf(<<EOF,\nbody\nEOF\n 1);\n' | perl
//	body
//	|1
//
// The parser took bodies only AFTER a statement's terminator, which is right
// for `print <<A;` and wrong here: the body arrived in the middle of the
// argument list, and the call stopped at it. run/todo.t alone carried a dozen
// of these.
//
// Canon moves the body to after the terminator, which is a valid spelling of
// the same program -- the body still begins on the line after its opener --
// and every case checks the canon re-parses to itself.
func TestHeredocBodyInsideItsStatement(t *testing.T) {
	cases := []struct {
		name  string
		src   string
		canon string
	}{
		{"argument after the body",
			"f(<<EOF,\nbody\nEOF\n 1);\n",
			"f(<<EOF , 1);\nbody\nEOF\n"},
		{"indented, quoted",
			"f(<<~'EOF',\n  body\n  EOF\n  \"n\");\n",
			"f(<<~'EOF' , \"n\");\n  body\n  EOF\n"},
		{"opener on its own line, trailing comma",
			"f(\n  <<~EOF,\n    body\n    EOF\n  2,\n);\n",
			"f(<<~EOF , 2 ,);\n    body\n    EOF\n"},
		// A body inside a nested statement belongs to THAT statement, and the
		// outer one must not claim it a second time.
		{"nested in a block",
			"f(sub { g(<<EOF,\nbody\nEOF\n 1) }, 2);\n",
			"f(sub {g(<<EOF , 1);\nbody\nEOF\n} , 2);"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			src := []byte(c.src)
			n := parse.Parse(src)
			if got := countUnknown(n); got != 0 {
				t.Errorf("Parse(%q) has %d Unknown, want 0", c.src, got)
			}
			if got := n.SourceText(src); got != c.src {
				t.Errorf("round-trip:\n  got  %q\n  want %q", got, c.src)
			}
			got := strings.TrimSpace(parse.Canon(n, src))
			if got != strings.TrimSpace(c.canon) {
				t.Errorf("Canon(%q):\n  got  %q\n  want %q", c.src, got, strings.TrimSpace(c.canon))
			}
			again := strings.TrimSpace(parse.Canon(parse.Parse([]byte(got+"\n")), []byte(got+"\n")))
			if again != got {
				t.Errorf("canon is not a fixpoint:\n  once  %q\n  twice %q", got, again)
			}
		})
	}
}
