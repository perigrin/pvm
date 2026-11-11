// ABOUTME: Tests that lex-overhead runs chart/lex and lexer.Tokenize through the §9.3 protocol with Tokenize inside the numerator.
// ABOUTME: Also lex-overhead's definition (25 buffers a pass), its production clock, its unsoftened 1.3 gate and its stability under load.

package main

import (
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand/v2"
	"os"
	"path/filepath"
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

// TestLexOverheadStableUnderLoad models LEXLOAD's established cause on a fake
// clock: the measuring thread waits for a CPU in delays that land per unit of
// work at random, so a short pass's wall time is decided by the scheduler
// rather than by the work, and 25 copies a pass average it out. It shows the
// 25-copy pass's rationale, not that real load averages out: under heavy
// real load it did not, so the ratchet measures only on a quiet machine
// (TestLexOverheadRefusesBusyMachine). The fake chart does 1.15
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

// lexWorld is a fake clock for lexOverheadKeys: GC starts pass k+1, the
// reference costs 10 a buffer, the chart 10*chart(k, path). It logs each
// pass as its file and run count.
type lexWorld struct {
	now    time.Duration
	pass   int
	passes [][]string
	chart  func(pass int, path string) float64
	stderr strings.Builder
}

func (w *lexWorld) protocol() Protocol {
	side := func(chart bool) Side {
		return func() Runner {
			return func(path string) {
				w.passes[len(w.passes)-1] = append(w.passes[len(w.passes)-1], path)
				c := 10.0
				if chart {
					c *= w.chart(w.pass, path)
				}
				w.now += time.Duration(c * 1000)
			}
		}
	}
	return Protocol{
		Chart: side(true), Ref: side(false),
		Now:    func() time.Duration { return w.now },
		GC:     func() { w.pass++; w.passes = append(w.passes, nil) },
		Stderr: &w.stderr,
	}
}

// TestLexOverheadDefinitionUnchanged: lex-overhead is 2 warm-up and 9 timed
// pairs, the median of the per-pair ratios, the median of three
// measurements, over std5k.pl and edit5k.pl, the key the larger of the two;
// each pass lexes its one buffer 25 times (the fourth amendment, about 200 ms
// of work a pass).
func TestLexOverheadDefinitionUnchanged(t *testing.T) {
	rows := [][]float64{
		{1.21, 1.29, 1.22, 1.28, 1.25, 9.9, 1.23, 1.27, 1.24},
		{1.01, 1.09, 3.5, 1.02, 1.05, 1.07, 1.03, 1.06, 1.04},
		{1.11, 1.19, 1.12, 1.15, 1.17, 1.13, 1.16, 1.14, 7.0},
	}
	const perFile = 3 * 11 * 2
	w := &lexWorld{chart: func(pass int, path string) float64 {
		k := pass - 1 // GC has already counted this pass
		if path != std5k {
			k -= perFile
		}
		m, i := k/2/11, k/2%11
		switch {
		case path == "a.pl":
			return 1
		case i < 2:
			return 0.01 // a warm-up counted among the timed pairs would move every median
		case path == edit5k:
			return 1.2
		}
		return rows[m][i-2]
	}}
	m := map[string]string{}
	lexOverheadKeys(w.protocol(), []string{"a.pl"}, m)
	if m["lex-overhead"] != "1.200" || m["lex-overhead-std5k"] != "1.150" {
		t.Errorf("lex-overhead %s, std5k %s; want 1.200 (edit5k's, the larger) and 1.150 (the median of medians 1.25, 1.05, 1.15)",
			m["lex-overhead"], m["lex-overhead-std5k"])
	}
	for i, f := range []string{std5k, edit5k} {
		for j, pass := range w.passes[i*perFile : (i+1)*perFile] {
			if want := slices.Repeat([]string{f}, 25); !slices.Equal(pass, want) {
				t.Fatalf("%s pass %d ran %v, want %s 25 times", f, j, pass, f)
			}
		}
	}
	for _, s := range []string{"lex-overhead-std5k measurement 1: median 1.25", "lex-overhead-std5k measurement 2: median 1.05",
		"lex-overhead-std5k measurement 3: median 1.15", "lex-overhead-edit5k measurement 3: median 1.2"} {
		if !strings.Contains(w.stderr.String(), s) {
			t.Errorf("stderr %q does not report %q", w.stderr.String(), s)
		}
	}
}

// TestLexOverheadLimitNotSoftened: 1.301 fails the ratchet's check and 1.3
// passes; a missing or na lex-overhead fails against a numeric baseline; and
// a first reading above 1.3 stands, taken once and never re-measured.
func TestLexOverheadLimitNotSoftened(t *testing.T) {
	baseline := Format(report(map[string]string{"lex-overhead": "1.200"}))
	for _, tc := range []struct {
		name     string
		measured map[string]string
		want     int
	}{
		{"1.301", map[string]string{"lex-overhead": "1.301"}, 1},
		{"exactly 1.3", map[string]string{"lex-overhead": "1.3"}, 0},
		{"na", map[string]string{"lex-overhead": na}, 1},
		{"missing", map[string]string{}, 1},
	} {
		if got := check(report(tc.measured), nil, baseline, "", io.Discard); got != tc.want {
			t.Errorf("lex-overhead %s: check exits %d, want %d", tc.name, got, tc.want)
		}
	}

	// std5k reads 1.31 on its first figure and 1.0 on any pass after it.
	const perFile = 3 * 11 * 2
	w := &lexWorld{chart: func(pass int, path string) float64 {
		if path == std5k && pass <= perFile {
			return 1.31
		}
		return 1
	}}
	m := map[string]string{}
	lexOverheadKeys(w.protocol(), []string{"a.pl"}, m)
	n := 0
	for _, p := range w.passes {
		if p[0] == std5k {
			n++
		}
	}
	if m["lex-overhead"] != "1.310" || n != perFile {
		t.Errorf("lex-overhead %s over %d std5k passes; want 1.310 over %d: a reading above 1.3 is never re-measured", m["lex-overhead"], n, perFile)
	}
	if got := check(report(m), nil, baseline, "", io.Discard); got != 1 {
		t.Errorf("a first lex-overhead of 1.310: check exits %d, want 1", got)
	}
}

// TestLexOverheadRefusesBusyMachine: above the load limit lex-overhead is
// refused with "machine too busy to measure" and no reading, which fails the
// ratchet; at the limit it is measured as before; an unreadable load is an
// error, never a pass. The load is injected: the test reads no real load.
func TestLexOverheadRefusesBusyMachine(t *testing.T) {
	baseline := Format(report(map[string]string{"lex-overhead": "1.200", "lex-overhead-std5k": "1.200", "lex-overhead-aggregate": "1.200"}))
	for _, tc := range []struct {
		name    string
		load    float64
		loadErr error
		refuse  string
	}{
		{"above the limit", 4.01, nil, "machine too busy to measure"},
		{"not a number", math.NaN(), nil, "machine too busy to measure"},
		{"unreadable", 0, errors.New("no /proc/loadavg"), "no /proc/loadavg"},
		{"at the limit", 4, nil, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := &lexWorld{chart: func(int, string) float64 { return 1.1 }}
			m := map[string]string{}
			err := quietLexOverheadKeys(w.protocol(), []string{"a.pl"}, m, func() (float64, error) { return tc.load, tc.loadErr }, 4)
			if tc.refuse == "" {
				if err != nil || m["lex-overhead"] != "1.100" || m["lex-overhead-std5k"] != "1.100" || m["lex-overhead-aggregate"] != "1.100" {
					t.Errorf("load %v at limit 4: error %v, measured %v; want lex-overhead measured as before", tc.load, err, m)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.refuse) {
				t.Errorf("error %v, want one containing %q", err, tc.refuse)
			}
			if len(m) != 0 || len(w.passes) != 0 {
				t.Errorf("measured %v over %d passes; a refusal times nothing and prints no reading", m, len(w.passes))
			}
			if got := check(report(m), nil, baseline, "", io.Discard); got == 0 {
				t.Error("a refused lex-overhead passes the ratchet's check")
			}
		})
	}
}

// TestLoadAvg reads the 1-minute figure; a missing or empty file is an
// error, which is how the ratchet behaves off Linux (a shortcut, README).
func TestLoadAvg(t *testing.T) {
	dir := t.TempDir()
	write := func(name, text string) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(text), 0o600); err != nil {
			t.Fatal(err)
		}
		return p
	}
	if got, err := loadAvg(write("loadavg", "0.31 0.07 0.02 1/200 4242\n")); err != nil || got != 0.31 {
		t.Errorf("loadAvg = %v, %v; want 0.31", got, err)
	}
	for _, p := range []string{write("empty", ""), filepath.Join(dir, "missing")} {
		if _, err := loadAvg(p); err == nil {
			t.Errorf("loadAvg(%s) has no error; an unreadable load must not pass", filepath.Base(p))
		}
	}
}
