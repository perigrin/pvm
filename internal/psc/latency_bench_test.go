// ABOUTME: M2's latency measurement (spec chapter 6 §6.11 step 4): parse, lower, analyze and a 1-char edit on edit5k.pl.
// ABOUTME: Each benchmark times every iteration and reports p50 and p99 in ms, the percentiles §6.10's budget is set in.

package psc

import (
	"bytes"
	"os"
	"slices"
	"testing"
	"time"

	"tamarou.com/pvm/internal/infer"
	"tamarou.com/pvm/internal/lower"
	"tamarou.com/pvm/internal/parse"
)

// latencySource is the fixture's text; a benchmark that cannot read it fails.
func latencySource(b *testing.B) []byte {
	b.Helper()
	src, err := os.ReadFile(latencyFixturePath)
	if err != nil {
		b.Fatal(err)
	}
	return src
}

// timeEach runs f b.N times, timing each run, and reports the p50 and p99
// of those times as ms/p50 and ms/p99 alongside the benchmark's mean.
func timeEach(b *testing.B, f func(i int)) {
	b.Helper()
	times := make([]time.Duration, b.N)
	b.ResetTimer()
	for i := range b.N {
		start := time.Now()
		f(i)
		times[i] = time.Since(start)
	}
	b.StopTimer()
	p50, p99 := percentiles(times)
	b.ReportMetric(float64(p50)/float64(time.Millisecond), "ms/p50")
	b.ReportMetric(float64(p99)/float64(time.Millisecond), "ms/p99")
}

// percentiles is the nearest-rank p50 and p99 of times.
func percentiles(times []time.Duration) (p50, p99 time.Duration) {
	if len(times) == 0 {
		return 0, 0
	}
	sorted := slices.Clone(times)
	slices.Sort(sorted)
	rank := func(p int) time.Duration {
		i := (p*len(sorted)+99)/100 - 1
		return sorted[max(i, 0)]
	}
	return rank(50), rank(99)
}

func TestPercentilesNearestRank(t *testing.T) {
	var times []time.Duration
	for i := 1; i <= 100; i++ {
		times = append(times, time.Duration(i))
	}
	if p50, p99 := percentiles(times); p50 != 50 || p99 != 99 {
		t.Errorf("percentiles(1..100) = %d, %d; want 50, 99", p50, p99)
	}
	if p50, p99 := percentiles([]time.Duration{7}); p50 != 7 || p99 != 7 {
		t.Errorf("percentiles([7]) = %d, %d; want 7, 7", p50, p99)
	}
}

// BenchmarkLatencyParse: parse.Parse over the fixture.
func BenchmarkLatencyParse(b *testing.B) {
	src := latencySource(b)
	timeEach(b, func(int) { parse.Parse(src) })
}

// BenchmarkLatencyLower: lower.Lower and its position map over the
// fixture's parse tree, the half of AnalyzeParsed before inference.
func BenchmarkLatencyLower(b *testing.B) {
	root := parse.Parse(latencySource(b))
	timeEach(b, func(int) {
		f := lower.Lower(root)
		lower.NewPositions(root, f)
	})
}

// BenchmarkLatencyAnalyze: infer.Analyze over the fixture's lowered tree.
func BenchmarkLatencyAnalyze(b *testing.B) {
	root := parse.Parse(latencySource(b))
	f := lower.Lower(root)
	pos := lower.NewPositions(root, f)
	timeEach(b, func(int) { infer.Analyze(f, pos, nil, infer.Options{}) })
}

// BenchmarkLatencyOpen: the LSP's whole path for one buffer, OpenDocument,
// which parses, lowers and analyzes. The uri is not a file:, so no `use`
// reads a module from disk and the time is the fixture's own.
func BenchmarkLatencyOpen(b *testing.B) {
	src := latencySource(b)
	s := NewLSPServer()
	timeEach(b, func(int) {
		if err := s.OpenDocument("untitled:edit5k.pl", src); err != nil {
			b.Fatal(err)
		}
	})
}

// BenchmarkLatencyEdit: a 1-char edit to the fixture, as didChange
// delivers it today (full sync, §6.11 step 3): each iteration inserts or
// removes one space at the start of the middle block's `sub scale` line,
// alternately, and the LSP re-opens the whole buffer.
func BenchmarkLatencyEdit(b *testing.B) {
	src := latencySource(b)
	at := editOffset(b, src)
	edited := slices.Concat(src[:at], []byte(" "), src[at:])
	s := NewLSPServer()
	if err := s.OpenDocument("untitled:edit5k.pl", src); err != nil {
		b.Fatal(err)
	}
	timeEach(b, func(i int) {
		buf := edited
		if i%2 == 1 {
			buf = src
		}
		if err := s.OpenDocument("untitled:edit5k.pl", buf); err != nil {
			b.Fatal(err)
		}
	})
}

// editOffset is the start of the middle `sub scale` line in src.
func editOffset(tb testing.TB, src []byte) int {
	tb.Helper()
	marker := []byte("\nsub scale {")
	var at []int
	for i := 0; ; {
		j := bytes.Index(src[i:], marker)
		if j < 0 {
			break
		}
		at = append(at, i+j+1)
		i += j + 1
	}
	if len(at) == 0 {
		tb.Fatal("the latency fixture has no `sub scale` line to edit")
	}
	return at[len(at)/2]
}

// TestLatencyEditParsesClean: the edited fixture, the buffer
// BenchmarkLatencyEdit re-opens on odd iterations, parses as cleanly as
// the fixture, so the edit benchmark times the same work.
func TestLatencyEditParsesClean(t *testing.T) {
	src, err := os.ReadFile(latencyFixturePath)
	if err != nil {
		t.Fatal(err)
	}
	at := editOffset(t, src)
	edited := slices.Concat(src[:at], []byte(" "), src[at:])
	if n := unknowns(parse.Parse(edited)); n != 0 {
		t.Errorf("the edited fixture parses with %d Unknown node(s), want 0", n)
	}
}
