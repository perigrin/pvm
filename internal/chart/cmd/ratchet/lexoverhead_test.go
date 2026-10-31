// ABOUTME: Tests that lex-overhead runs chart/lex and lexer.Tokenize through the §9.3 protocol with Tokenize inside the numerator.
// ABOUTME: On a fake clock that only an injected Tokenize moves: alternation, a GC before each pass, the median of three.

package main

import (
	"math"
	"strings"
	"testing"
	"time"

	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/lexer"
)

func TestLexOverheadProtocol(t *testing.T) {
	srcs := map[string][]byte{"a.pl": []byte("$x = 1;"), "b.pl": []byte("$y = foo;")}
	files := []string{"a.pl", "b.pl"}
	tt := toyLexTable(t)

	// run times lex-overhead where Tokenize costs delay(pass) on the fake
	// clock and nothing else moves it. It returns the figure and the log.
	run := func(delay func(pass int) time.Duration) (float64, []string, string) {
		var now time.Duration
		var log []string
		pass := -1
		tokenize := func(src []byte) []lexer.Token {
			log = append(log, "tokenize")
			now += delay(pass)
			return lexer.Tokenize(src)
		}
		chart, ref := lexSides(srcs, tt, tokenize)
		var stderr strings.Builder
		p := Protocol{
			Chart: chart, Ref: ref,
			Now:    func() time.Duration { return now },
			GC:     func() { pass++; log = append(log, "gc") },
			Stderr: &stderr,
		}
		return p.Ratio("lex-overhead", files, false), log, stderr.String()
	}

	// Tokenize's delay raises the numerator as well as the denominator:
	// with only Tokenize costing anything the figure is 1, whatever the
	// delay. A numerator that timed only the lookup would read 0.
	for _, d := range []time.Duration{10, 1000} {
		if got, _, _ := run(func(int) time.Duration { return d }); math.Abs(got-1) > 1e-9 {
			t.Errorf("Tokenize at %d per call: lex-overhead %v, want 1 (the numerator must include Tokenize)", d, got)
		}
	}

	// Pass k is chart/lex when k is even (chart first), Tokenize when odd;
	// pair k/2 is measurement pair/11, warm-ups its first two. Chart passes
	// cost a row's ratio times the Tokenize pass, warm-ups absurdly little.
	rows := [][]float64{
		{1.21, 1.29, 1.22, 1.28, 1.25, 9.9, 1.23, 1.27, 1.24},
		{1.01, 1.09, 3.5, 1.02, 1.05, 1.07, 1.03, 1.06, 1.04},
		{1.11, 1.19, 1.12, 1.15, 1.17, 1.13, 1.16, 1.14, 7.0},
	}
	got, log, stderr := run(func(pass int) time.Duration {
		pair := pass / 2
		m, i := pair/11, pair%11
		switch {
		case pass%2 == 1:
			return 1000
		case m >= len(rows):
			return 1000 // beyond the protocol's 33 pairs; the log reports it
		case i < 2:
			return 1
		}
		return time.Duration(1000 * rows[m][i-2])
	})
	if math.Abs(got-1.15) > 1e-9 {
		t.Errorf("lex-overhead %v, want the median of the three measurement medians, 1.15", got)
	}
	// Alternating passes, each one GC then Tokenize once per file.
	want := strings.Repeat("gc tokenize tokenize ", 3*11*2)
	if strings.Join(log, " ")+" " != want {
		t.Errorf("event log %q, want %d passes of a GC then one Tokenize per file", log, 3*11*2)
	}
	for _, s := range []string{"lex-overhead measurement 1: median 1.25", "measurement 2: median 1.05", "measurement 3: median 1.15"} {
		if !strings.Contains(stderr, s) {
			t.Errorf("stderr %q does not report %q", stderr, s)
		}
	}
}

// toyLexTable maps a variable to terminal 1; every other token is unmapped.
func toyLexTable(t *testing.T) *grammar.TokenTable {
	t.Helper()
	k := lexer.Variable
	tt, err := grammar.CompileTokens([]grammar.TokenClass{{Terminals: []int{1}, Kind: &k}})
	if err != nil {
		t.Fatal(err)
	}
	return tt
}
