// ABOUTME: PerlLines reads a `#line` directive by perl's own rules, the cases comp/parser.t holds perl to.
// ABOUTME: Each expectation is perl 5.42's nextstate line and file for the same source.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestPerlLinesReadsDirectivesAsPerlDoes is toke.c's S_incline, case by
// case: blanks around `line` and the number, a number glued to a word, a
// bare name with a space in it, quoted names with and without a closing
// quote, a CRLF line end, an empty name, line 0, and a name that returns
// to this file. The want column is what `perl -MO=Concise,-exec rules.pl`
// printed for each statement's nextstate.
func TestPerlLinesReadsDirectivesAsPerlDoes(t *testing.T) {
	src := "$a=1;\n" +
		"#\tline 13\n$a=2;\n" +
		"#line\t17\n$a=3;\n" +
		"#line 77sevenseven\n$a=4;\n" +
		"#line 47 bang eth\n$a=5;\n" +
		"#line 29 \"KAHEEEE\"   \n$a=6;\n" +
		"#line 31 \"CLINK CLOINK\"\n$a=7;\n" +
		"#line 43 \"BBF\n$a=8;\n" +
		"#line 50 x.pl\r\n$a=9;\n" +
		"#line 60 \"\"\n$a=10;\n" +
		"#line 0\n$a=11;\n" +
		"#line 5 \"rules.pl\"\n$a=12;\n"
	line, own := parse.PerlLines([]byte(src), "rules.pl")
	for _, c := range []struct {
		physical, want int
		own            bool
	}{
		{1, 1, true},
		{3, 13, true},   // a tab between # and line
		{5, 17, true},   // a tab between line and the number
		{7, 19, true},   // 77sevenseven: no blank after the number
		{9, 21, true},   // bang eth: a bare name ends at a blank
		{11, 29, false}, // KAHEEEE, trailing blanks allowed
		{13, 31, false}, // CLINK CLOINK, quoted
		{15, 43, false}, // "BBF, the unclosed quote is the name
		{17, 50, false}, // x.pl, before a CRLF
		{19, 60, false}, // "" names no file: still x.pl
		{21, 0, false},  // line 0
		{23, 5, true},   // back to this file
	} {
		if line[c.physical] != c.want || own[c.physical] != c.own {
			t.Errorf("physical line %d: got line %d own %v, perl says %d own %v",
				c.physical, line[c.physical], own[c.physical], c.want, c.own)
		}
	}
}
