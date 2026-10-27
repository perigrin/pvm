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

// TestPercentilesNearestRankEdgeCases: no samples, one, all equal and
// unsorted input each give an in-range answer, and the input is not sorted
// in place.
func TestPercentilesNearestRankEdgeCases(t *testing.T) {
	cases := []struct {
		name     string
		times    []time.Duration
		p50, p99 time.Duration
	}{
		{"empty", nil, 0, 0},
		{"one", []time.Duration{3}, 3, 3},
		{"all equal", []time.Duration{5, 5, 5, 5}, 5, 5},
		{"unsorted", []time.Duration{9, 1, 7, 3, 5, 2, 8, 4, 6, 10}, 5, 10},
	}
	for _, c := range cases {
		in := slices.Clone(c.times)
		if p50, p99 := percentiles(in); p50 != c.p50 || p99 != c.p99 {
			t.Errorf("%s: percentiles(%v) = %d, %d; want %d, %d", c.name, c.times, p50, p99, c.p50, c.p99)
		}
		if !slices.Equal(in, c.times) {
			t.Errorf("%s: percentiles reordered its input to %v", c.name, in)
		}
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
// delivers it today (full sync, §6.11 step 3): iteration i re-opens the
// whole buffer after latencyEdits' i-th edit, a typing sequence and not a
// toggle, so no two iterations see the same document.
func BenchmarkLatencyEdit(b *testing.B) {
	src := latencySource(b)
	bufs := latencyEdits(src, b.N)
	s := NewLSPServer()
	if err := s.OpenDocument("untitled:edit5k.pl", src); err != nil {
		b.Fatal(err)
	}
	timeEach(b, func(i int) {
		if err := s.OpenDocument("untitled:edit5k.pl", bufs[i]); err != nil {
			b.Fatal(err)
		}
	})
}

// scaleSubs is the offset of every `sub scale {` line in src, in order.
func scaleSubs(src []byte) []int {
	marker := []byte("\nsub scale {")
	var at []int
	for i := 0; ; {
		j := bytes.Index(src[i:], marker)
		if j < 0 {
			return at
		}
		at = append(at, i+j+1)
		i += j + 1
	}
}

// scaleZero is the offset of the `0` in `$_ > 0` in the `sub scale` body
// that starts at sub; a fixture without one is a broken generator.
func scaleZero(src []byte, sub int) int {
	j := bytes.Index(src[sub:], []byte("$_ > 0"))
	if j < 0 {
		panic("a `sub scale` in the latency fixture has no `$_ > 0`")
	}
	return sub + j + len("$_ > ")
}

// editOffset is the `0` of `$_ > 0` on the second line of the middle
// `sub scale` body in src, the one edit tests that ask "is it local" read.
func editOffset(tb testing.TB, src []byte) int {
	tb.Helper()
	subs := scaleSubs(src)
	if len(subs) == 0 {
		tb.Fatal("the latency fixture has no `sub scale` line to edit")
	}
	return scaleZero(src, subs[len(subs)/2])
}

// latencyEdits is n buffers, each its predecessor (src for the first) plus
// one digit: edit k inserts 1 + (s mod 9) before the `0` of `$_ > 0` in the
// s-th `sub scale` body, s = k mod (number of such subs). The edits
// accumulate (site 3: `$_ > 0` → `$_ > 40` → `$_ > 440`), so every buffer
// is distinct and a cache keyed by content cannot answer them.
func latencyEdits(src []byte, n int) [][]byte {
	subs := scaleSubs(src)
	zeros := make([]int, len(subs))
	for s, sub := range subs {
		zeros[s] = scaleZero(src, sub)
	}
	bufs := make([][]byte, n)
	buf := src
	for k := range n {
		s := k % len(subs)
		at := zeros[s]
		buf = slices.Concat(buf[:at], []byte{byte('1' + s%9)}, buf[at:])
		// Every site at or after s moves right by the inserted byte; s
		// itself too, since its next digit goes before the same `0`.
		for r := s; r < len(zeros); r++ {
			zeros[r]++
		}
		bufs[k] = buf
	}
	return bufs
}

// TestLatencyEditParsesClean: the first, middle and last buffers of the
// timed edit sequence parse as cleanly as the fixture, so the edit
// benchmark times the same work.
func TestLatencyEditParsesClean(t *testing.T) {
	src, err := os.ReadFile(latencyFixturePath)
	if err != nil {
		t.Fatal(err)
	}
	bufs := latencyEdits(src, 200)
	for _, k := range []int{0, len(bufs) / 2, len(bufs) - 1} {
		if n := unknowns(parse.Parse(bufs[k])); n != 0 {
			t.Errorf("edit buffer %d parses with %d Unknown node(s), want 0", k, n)
		}
	}
}

// TestLatencyEditIsOneByteInsideASub: editOffset is the `0` of `$_ > 0` on
// the second line of the middle `sub scale` body, a one-byte literal and
// not trivia, so a 1-char insert there is local to that sub.
func TestLatencyEditIsOneByteInsideASub(t *testing.T) {
	src, err := os.ReadFile(latencyFixturePath)
	if err != nil {
		t.Fatal(err)
	}
	at := editOffset(t, src)
	if at <= 0 || at >= len(src)-1 {
		t.Fatalf("editOffset = %d, want strictly inside (0, %d)", at, len(src)-1)
	}
	subs := scaleSubs(src)
	mid := subs[len(subs)/2]
	line2 := mid + bytes.IndexByte(src[mid:], '\n') + 1
	end := mid + bytes.Index(src[mid:], []byte("\n}\n"))
	if at <= line2 || at >= end {
		t.Fatalf("editOffset = %d, want inside the middle `sub scale` body after its first line (%d, %d)", at, line2, end)
	}
	if at < 5 || string(src[at-5:at+1]) != "$_ > 0" {
		t.Errorf("editOffset %d is not the 0 of `$_ > 0`", at)
	}
	if leaf := leafAt(parse.Parse(src), at); leaf.Kind == parse.Trivia || leaf.End-leaf.Start != 1 {
		t.Errorf("editOffset %d is on %v [%d,%d), want the one-byte literal 0", at, leaf.Kind, leaf.Start, leaf.End)
	}
	edited := slices.Concat(src[:at], []byte("1"), src[at:])
	if n := unknowns(parse.Parse(edited)); n != 0 {
		t.Errorf("the edited fixture parses with %d Unknown node(s), want 0", n)
	}
}

// TestLatencyEditsPairwiseDistinct: the timed edits are ≥ 200 one-byte
// inserts across ≥ 20 `sub scale` bodies, and no two buffers are equal,
// so the p99 is not taken over a two-state toggle.
func TestLatencyEditsPairwiseDistinct(t *testing.T) {
	src, err := os.ReadFile(latencyFixturePath)
	if err != nil {
		t.Fatal(err)
	}
	const n = 200
	bufs := latencyEdits(src, n)
	if len(bufs) != n {
		t.Fatalf("latencyEdits(src, %d) gave %d buffers", n, len(bufs))
	}
	seen := map[string]int{}
	sites := map[int]bool{}
	prev := src
	for k, buf := range bufs {
		if j, dup := seen[string(buf)]; dup {
			t.Fatalf("buffers %d and %d are equal", j, k)
		}
		seen[string(buf)] = k
		if len(buf) != len(prev)+1 {
			t.Fatalf("edit %d changes the length by %d, want 1", k, len(buf)-len(prev))
		}
		i := 0
		for i < len(prev) && prev[i] == buf[i] {
			i++
		}
		if !bytes.Equal(prev[i:], buf[i+1:]) {
			t.Fatalf("edit %d is not a one-byte insert", k)
		}
		subs := scaleSubs(buf)
		s := len(subs) - 1
		for s >= 0 && subs[s] > i {
			s--
		}
		if s < 0 || i >= subs[s]+bytes.Index(buf[subs[s]:], []byte("\n}\n")) {
			t.Fatalf("edit %d at %d is not inside a `sub scale` body", k, i)
		}
		sites[s] = true
		prev = buf
	}
	if len(sites) < 20 {
		t.Errorf("the edits touch %d `sub scale` bodies, want at least 20", len(sites))
	}
}

// leafAt is the deepest node under n whose span holds offset.
func leafAt(n *parse.Node, offset int) *parse.Node {
	for _, c := range n.Children {
		if c.Start <= offset && offset < c.End {
			return leafAt(c, offset)
		}
	}
	return n
}
