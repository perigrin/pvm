// ABOUTME: A `)` whose `{` follows a heredoc body opens a block: the body is not
// ABOUTME: between them to perl, which has already cut it out of the line.

package lexer

import "testing"

// TestBlockAfterHeredocBody: measured on 5.42.0,
//
//	$ cat hd.pl
//	foreach my $l (split /\n/, <<'EOF')
//	1 { a }
//	EOF
//	{
//	    f($l);
//	}
//	{ g(2); h(3); }
//	$ perl -MO=Deparse hd.pl
//	foreach my $l (split(/\n/, "1 { a }\n", 0)) {
//	    f($l);
//	}
//	{
//	    g(2);
//	    h(3);
//	}
//
// perl.git t/comp/parser.t:169.
func TestBlockAfterHeredocBody(t *testing.T) {
	src := "foreach my $l (split /\\n/, <<'EOF')\n1 { a }\nEOF\n{\n    f($l);\n}\n{ g(2); h(3); }\n"
	var braces int
	for _, tok := range Tokenize([]byte(src)) {
		if tok.Kind == Operator && src[tok.Start:tok.End] == "{" {
			braces++
			if !tok.OpensBlock {
				t.Errorf("the `{` at %d opens a block", tok.Start)
			}
		}
	}
	if braces != 2 {
		t.Errorf("got %d `{` tokens, want 2 -- the body's is inside the heredoc", braces)
	}
}
