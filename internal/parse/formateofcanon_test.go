// ABOUTME: A format whose `.` line ends the file with no newline: canon closes
// ABOUTME: that line, so the separator after it does not join the terminator.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestFormatAtEOFCanonSettles: `format=\n0\n.` with no newline after the `.`
// is a canon fixpoint.
//
// perl's terminator is the whole `.` line: a `.`, then any spaces or tabs,
// then a newline OR THE END OF THE FILE. Measured on 5.42.0 with
// `perl -MO=Deparse`, each of these is syntax OK and deparses to the same
// one-line format:
//
//	format=\n0\n.        format=\n0\n.\n
//	format=\n0\n.<sp>    format=\n0\n.<sp>\n
//
// while `.<sp>x` and `.x` are picture lines and the format runs on. The
// lexer's FormatBody already ends where perl's does, trailing blanks included.
// What broke was canon: a format is a block form, so the statement after it
// is separated by a space, and when the source ended on the `.` that space
// landed on the terminator's own line. The re-lex rightly took it into the
// FormatBody, the next emission added another, and canon never settled.
func TestFormatAtEOFCanonSettles(t *testing.T) {
	for _, src := range []string{
		"format=\n0\n.",
		"format=\n0\n. ",
		"format=\n0\n.\t",
		"format =\n0\n.\nformat X =\n1\n.",
	} {
		once := parse.Canon(parse.Parse([]byte(src)), []byte(src))
		twice := parse.Canon(parse.Parse([]byte(once)), []byte(once))
		if once != twice {
			t.Errorf("%q: canon is not a fixpoint:\n once  %q\n twice %q", src, once, twice)
		}
	}

	// A format whose `.` line already ends in a newline emits as it always
	// has: its source text, then the block form's space.
	for src, want := range map[string]string{
		"format=\n0\n.\n":     "format=\n0\n.\n ",
		"format=\n0\n.\n1;\n": "format=\n0\n.\n 1;",
	} {
		if got := parse.Canon(parse.Parse([]byte(src)), []byte(src)); got != want {
			t.Errorf("%q: canon is %q, want %q", src, got, want)
		}
	}
}

// TestFormatWithNoBodyRefuses: a format declaration the file ends inside
// is not a format. perl 5.42 says "Format not terminated" for each of these
// under `perl -c`, so the parser refuses them rather than building a
// declaration with no body -- a tree canon wrote back as `format=\n `, whose
// re-lex took the space as an unterminated body.
func TestFormatWithNoBodyRefuses(t *testing.T) {
	for _, src := range []string{"format=", "format =\n", "format STDOUT =", "format =\nabc\n"} {
		found := collectUnknown(parse.Parse([]byte(src)))
		if len(found) == 0 || found[0].Refusal != parse.FormatNotTerminated {
			t.Errorf("%q: want a format_not_terminated refusal, got %v", src, refusalCodes(found))
		}
	}
	if found := collectUnknown(parse.Parse([]byte("format =\n.\n"))); len(found) != 0 {
		t.Errorf("format =\\n.\\n: an empty format is terminated, but it refuses: %v", found)
	}
}

// refusalCodes names the codes of found, for a failure message.
func refusalCodes(found []*parse.Node) []parse.RefusalCode {
	var codes []parse.RefusalCode
	for _, n := range found {
		codes = append(codes, n.Refusal)
	}
	return codes
}
