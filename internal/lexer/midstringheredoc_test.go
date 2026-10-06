// ABOUTME: A heredoc queued on a line whose string runs on takes its body from the
// ABOUTME: next line, out of the middle of that string: perl removes it from the input.

package lexer

import "testing"

// TestHeredocInsideRunningString: perl's scan_heredoc reads a body as soon as
// it lexes the opener, removing those lines from the input, so a string
// that continues past the line resumes after them. Measured on 5.42.0,
//
//	my $v13 = <<E13 . "@{[ <<E14 ]}text     prints "[outer 1\ninner\ntext\nafter]"
//	outer 1
//	E13
//	inner
//	E14
//	after";
//	my $v17 = <<E21 . 'single                prints "[E21 content\nsingle\nquoted]"
//	E21 content
//	E21
//	quoted';
//
// PerlOnJava unit/nested_heredoc.t:94-100 and :148.
func TestHeredocInsideRunningString(t *testing.T) {
	for _, src := range []string{
		"my $v13 = <<E13 . \"@{[ <<E14 ]}text\nouter 1\nE13\ninner\nE14\nafter\";\nprint 1;\n",
		"my $v17 = <<E21 . 'single\nE21 content\nE21\nquoted';\nprint 1;\n",
	} {
		for _, tok := range Tokenize([]byte(src)) {
			if tok.Kind == UnknownRest || tok.Kind == HeredocBody {
				t.Errorf("%q: %v %q; the body is taken out of the string", src, tok.Kind, src[tok.Start:tok.End])
			}
		}
		if tok, ok := firstOfKind(src, Word); !ok || src[tok.Start:tok.End] != "my" {
			t.Errorf("%q: first word %q", src, src[tok.Start:tok.End])
		}
		// `print` after the statement is still code.
		found := false
		for _, tok := range Tokenize([]byte(src)) {
			if tok.Kind == Word && src[tok.Start:tok.End] == "print" {
				found = true
			}
		}
		if !found {
			t.Errorf("%q: `print` after the statement is not a Word", src)
		}
	}
}
