// ABOUTME: A heredoc terminator line ended by CRLF still ends the body, as perl reads CRLF as LF.
// ABOUTME: Each case is perl 5.42's `perl -c` verdict on a file, and the near misses perl refuses stay unterminated.

package lexer

import "testing"

// terminated reports whether src's first heredoc body found its terminator:
// a HeredocBody token rather than the UnknownRest of a body run to EOF.
func terminated(src string) bool {
	for _, tok := range Tokenize([]byte(src)) {
		switch tok.Kind {
		case HeredocBody:
			return true
		case UnknownRest:
			return false
		}
	}
	return false
}

// TestCRLFHeredocTerminator: perl reads a CRLF line as an LF one, so
// `EOT\r\n` ends the body. Measured on 5.42.0 with `perl -c`, each of the
// first three is "syntax OK", and running the first prints the length of
// "abc\n", 4. A `\r` that does not end a line is the line's own: `EOT \r\n`
// and `EOT\r` at end of file are "Can't find string terminator".
func TestCRLFHeredocTerminator(t *testing.T) {
	for _, src := range []string{
		"my $s = <<EOT;\r\nabc\r\nEOT\r\nmy $x = undef;\r\n",
		"my $s = <<~EOT;\r\n  abc\r\n  EOT\r\n1;\r\n",
		"my $s = <<EOT;\nabc\nEOT\r\n",
	} {
		if !terminated(src) {
			t.Errorf("%q: perl ends the body at EOT, the lexer ran it to EOF", src)
		}
	}
	for _, src := range []string{
		"my $s = <<EOT;\r\nabc\r\nEOT \r\n",
		"my $s = <<EOT;\r\nabc\r\nEOT\r",
	} {
		if terminated(src) {
			t.Errorf("%q: perl finds no terminator, the lexer did", src)
		}
	}
}
