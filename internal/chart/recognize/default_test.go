// ABOUTME: Tests the switch-over: recognize.New is the judged recognizer, so the precedence fixtures die in its chart.
// ABOUTME: The no-precedence constructor still accepts each of them, as V1's recognizer did.

package recognize

import (
	"testing"

	"tamarou.com/pvm/internal/chart/lex"
)

// TestDefaultRecognizerJudgesPrecedence: over the committed tables recognize.New refuses each precedence fixture in
// the chart (indirob before `[` on the reduce side, the inner NCEQOP and DOTDOT shifts on the shift side), and
// NewNoPrecedence accepts each.
func TestDefaultRecognizerJudgesPrecedence(t *testing.T) {
	tab := committedTables(t)
	for _, src := range []string{"print $fh [];", "$a <=> $b <=> $c;", "1 .. 2 .. 3;"} {
		s, _ := lex.Lex([]byte(src), tab.Tokens)
		l := lex.Linear(s)
		if mustRecognize(t, New(tab, Options{}), l).Accepted() {
			t.Errorf("%q: recognize.New accepts it; the default must judge precedence", src)
		}
		if !mustRecognize(t, NewNoPrecedence(tab, Options{}), l).Accepted() {
			t.Errorf("%q: the no-precedence constructor refuses it", src)
		}
	}
}
