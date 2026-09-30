// ABOUTME: A heredoc opened inside a string's `@{[ ... ]}` or `${\ ...}` takes its
// ABOUTME: body from the lines after the statement, as one in a code block does.

package lexer

import "testing"

// TestInterpolationHeredoc: measured on 5.42.0,
//
//	my $text = "x @{[ <<'EOT' ]} x";      prints "[x HERE\n x]"
//	HERE
//	EOT
//	my $t2 = qq{a ${\ <<E2} b};           prints "[a body\n b]"
//	body
//	E2
//
// An escaped `\${` is text. PerlOnJava unit/string_interpolation.t:350.
func TestInterpolationHeredoc(t *testing.T) {
	src := "my $text = \"x @{[ <<'EOT' ]} x\";\nHERE\nEOT\nprint 1;\n" +
		"my $t2 = qq{a ${\\ <<E2} b};\nbody\nE2\nprint 2;\n"
	var bodies []string
	for _, tok := range Tokenize([]byte(src)) {
		if tok.Kind == HeredocBody {
			bodies = append(bodies, src[tok.Start:tok.End])
		}
	}
	if len(bodies) != 2 || bodies[0] != "HERE\nEOT\n" || bodies[1] != "body\nE2\n" {
		t.Errorf("heredoc bodies %q, want [\"HERE\\nEOT\\n\" \"body\\nE2\\n\"]", bodies)
	}
	if _, ok := firstOfKind("my $s = \"a \\${ <<E } b\";\nprint 1;\n", HeredocBody); ok {
		t.Errorf("an escaped `\\${` opened a heredoc")
	}
}
