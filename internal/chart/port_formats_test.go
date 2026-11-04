// ABOUTME: Tests formats: `format NAME =` and its body, which internal/lexer emits as one FormatBody token.
// ABOUTME: Snippets under testdata/port/formats are recognized through internal/lexer, chart/lex and the recognizer; formats-refused ones refuse.

package chart

import (
	"os"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
)

// formatsLines are the perly.y alternatives this family ports or fits. The lexer emits a format's picture
// lines, argument lines and closing `.` as one FormatBody token (01a1a5b7-1bcb-7c4c-bbdc-1abedccd7e6b), so
// 787 is fitted to that token and 836, 837, 906, 924 and 925, which that token folds, are excluded
// folded-into-token. No alternative here refuses in its action: nothing is left for the completing issue.
var formatsLines = []int{488, 787, 836, 837, 887, 906, 924, 925, 1019, 1020, 1039}

func TestPortFormats(t *testing.T) {
	if err := grammar.CheckFamilyMarked(formatsLines, nil); err != nil {
		t.Error(err)
	}
	// `format` is the keyword only where a statement can start with it. As a key before => or in braces it is
	// refused like every keyword (`if => 1`, `$h{my}`): that is no format limit, so no row holds it.
	for _, src := range []string{"$x->format;", "sub format { 1 }"} {
		if v := recognizeSource(t, []byte(src)); !v.accepted {
			t.Errorf("%q is refused at %d", src, v.at)
		}
	}
	for _, f := range snippets(t, "formats") {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if v := recognizeSource(t, src); !v.accepted {
			t.Errorf("%s is refused at byte %d (expected %v); perl accepts it", f, v.at, v.expected)
		}
	}
}

func TestPortFormatsRefusals(t *testing.T) {
	for _, f := range snippets(t, "formats-refused") {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if v := recognizeSource(t, src); v.accepted {
			t.Errorf("%s is accepted; perl refuses it", f)
		}
	}
	for _, src := range []string{"format STDOUT =\n@<<<\n$x\n.\n", "format STDOUT =\n.\n"} {
		if v := recognizeSource(t, []byte(src)); !v.accepted {
			t.Errorf("%q is refused at %d", src, v.at)
		}
	}
	// Known limit (01a1a5b7-1bcb-7c4c-bbdc-1abedccd7e6b): the lexer emits the argument lines inside one
	// FormatBody token, so perl's refusal of a dangling operator there is not seen. This fails when the
	// limit goes; move it to the refused snippets then and drop the README line and the shortcut comment.
	if src := "format STDOUT =\n@<<<\n$x +\n.\n"; !recognizeSource(t, []byte(src)).accepted {
		t.Errorf("%q is refused; the known limit is gone, so move it to formats-refused and drop the README line", src)
	}
}
