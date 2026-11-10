// ABOUTME: Tests that lex-overhead runs chart/lex and lexer.Tokenize through the §9.3 protocol with Tokenize inside the numerator.
// ABOUTME: On a fake clock that only an injected Tokenize moves: alternation, a GC before each pass, the median of three.

package main

import (
	"fmt"
	"io"
	"math"
	"math/rand/v2"
	"reflect"
	"runtime"
	"slices"
	"strconv"
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

// TestLexOverheadKeys: lex-overhead is the larger of std5k.pl's and
// edit5k.pl's long-lived figures and times only those two files;
// lex-overhead-std5k is std5k.pl's; lex-overhead-aggregate comes from the
// aggregate's files. The fake clock charges 10 per file to the reference and
// 10*factor/10 to the chart.
func TestLexOverheadKeys(t *testing.T) {
	for _, tc := range []struct {
		name                      string
		stdF, editF, aggF         time.Duration
		wantMax, wantStd, wantAgg string
	}{
		{"edit5k slower", 11, 12, 13, "1.200", "1.100", "1.300"},
		{"std5k slower", 12, 11, 13, "1.200", "1.200", "1.300"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var now time.Duration
			factor := map[string]time.Duration{std5k: tc.stdF, edit5k: tc.editF, "a.pl": tc.aggF}
			var order []string // each file as it first runs after another file ran
			side := func(chart bool) Side {
				return func() Runner {
					return func(p string) {
						if len(order) == 0 || order[len(order)-1] != p {
							order = append(order, p)
						}
						if chart {
							now += factor[p]
						} else {
							now += 10
						}
					}
				}
			}
			p := Protocol{Chart: side(true), Ref: side(false), Now: func() time.Duration { return now }, GC: func() {}, Stderr: &strings.Builder{}}
			m := map[string]string{}
			lexOverheadKeys(p, []string{"a.pl"}, m)
			if m["lex-overhead"] != tc.wantMax || m["lex-overhead-std5k"] != tc.wantStd || m["lex-overhead-aggregate"] != tc.wantAgg {
				t.Errorf("lex-overhead %s, std5k %s, aggregate %s; want %s, %s, %s",
					m["lex-overhead"], m["lex-overhead-std5k"], m["lex-overhead-aggregate"], tc.wantMax, tc.wantStd, tc.wantAgg)
			}
			if want := []string{std5k, edit5k, "a.pl"}; !slices.Equal(order, want) {
				t.Errorf("files timed in order %v, want %v (the LSP buffers, then the aggregate)", order, want)
			}
		})
	}
}

// TestLexOverheadStableUnderLoad reproduces LEXLOAD's established cause: under
// load the measuring thread waits for a CPU for most of a pass, in delays
// that land per unit of work at random, so a short pass's wall time is
// decided by the scheduler rather than by the work. The fake chart does 1.15
// times the reference's work, 11.5 ms against 10 ms a buffer; one 100 µs
// tick of work in 50 waits an exponential delay averaging 10 ms, so a pass's
// wall time is about three times its work, in a few large delays. Over 20 seeds lex-overhead must stay within
// 1.3.
func TestLexOverheadStableUnderLoad(t *testing.T) {
	const tick = 100 * time.Microsecond
	var bad []string
	for seed := range uint64(20) {
		rng := rand.New(rand.NewPCG(seed, 1))
		var now time.Duration
		side := func(ticks int) Side {
			return func() Runner {
				return func(string) {
					for range ticks {
						now += tick
						if rng.IntN(50) == 0 {
							now += time.Duration(rng.ExpFloat64() * float64(100*tick))
						}
					}
				}
			}
		}
		p := Protocol{Chart: side(115), Ref: side(100), Now: func() time.Duration { return now }, GC: func() {}, Stderr: io.Discard}
		m := map[string]string{}
		lexOverheadKeys(p, []string{"a.pl"}, m)
		if v, err := strconv.ParseFloat(m["lex-overhead"], 64); err != nil || v > lexOverheadMax {
			bad = append(bad, fmt.Sprintf("seed %d: %s", seed, m["lex-overhead"]))
		}
	}
	if len(bad) > 0 {
		t.Errorf("work ratio 1.15 under scheduling delays read above %g on %d of 20 seeds: %v", lexOverheadMax, len(bad), bad)
	}
}

// TestLexOverheadProductionClock: the ratchet's Protocol reads the monotonic
// wall clock, which advances across a sleep where a CPU-time or thread-time
// clock would not, and collects with runtime.GC.
func TestLexOverheadProductionClock(t *testing.T) {
	p := productionProtocol(nil, nil)
	before := p.Now()
	time.Sleep(50 * time.Millisecond)
	if d := p.Now() - before; d < 50*time.Millisecond {
		t.Errorf("Now advanced %v across a 50ms sleep; the production clock must be wall time", d)
	}
	if reflect.ValueOf(p.GC).Pointer() != reflect.ValueOf(runtime.GC).Pointer() {
		t.Error("the production protocol's GC is not runtime.GC")
	}
}
