// ABOUTME: A format's argument line may open a `{` block spanning lines, and a
// ABOUTME: `.` inside that block -- a nested format's own end -- is not the body's.

package lexer

import "testing"

// TestFormatArgumentBlock: perlform, "the expressions may be spread out to
// more than one line if enclosed in braces. If so, the opening brace must be
// the first token on the first line". The block is Perl, and in perl.git
// t/op/write.t:1938 it declares a second format whose `.` ends only that one.
// Measured on 5.42.0, the file writes "birds" then "nest" and runs the
// statement after the outer `.`.
func TestFormatArgumentBlock(t *testing.T) {
	src := "format NEST =\n@<<<\n{\n    my $birds = \"birds\";\n    local *NEST = *BIRDS{FORMAT};\n    write NEST;\n    format BIRDS =\n@<<<<<\n$birds;\n.\n    \"nest\"\n}\n.\nprint \"after\\n\";\n"
	tok, ok := firstOfKind(src, FormatBody)
	if !ok {
		t.Fatalf("no FormatBody")
	}
	want := "@<<<\n{\n    my $birds = \"birds\";\n    local *NEST = *BIRDS{FORMAT};\n    write NEST;\n    format BIRDS =\n@<<<<<\n$birds;\n.\n    \"nest\"\n}\n.\n"
	if got := src[tok.Start:tok.End]; got != want {
		t.Errorf("FormatBody %q, want %q", got, want)
	}
	// A picture line with no field has no argument line: a `{` there is text.
	plain := "format P =\n{ just text\n.\nprint 1;\n"
	if tok, _ := firstOfKind(plain, FormatBody); plain[tok.Start:tok.End] != "{ just text\n.\n" {
		t.Errorf("%q: FormatBody %q", plain, plain[tok.Start:tok.End])
	}
}
