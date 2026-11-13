// ABOUTME: The §9.3 speed protocol: alternating chart/reference passes, GC outside the timed span, medians of pairs.
// ABOUTME: Also the aggregate's file set: std5k.pl, edit5k.pl and any file above 10% of RD's population time left out.

package main

import (
	"fmt"
	"io"
	"maps"
	"path/filepath"
	"runtime"
	"slices"
	"testing"
	"time"

	"tamarou.com/pvm/internal/chart"
	"tamarou.com/pvm/internal/chart/forest"
	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/chart/lex"
	"tamarou.com/pvm/internal/chart/recognize"
	"tamarou.com/pvm/internal/lexer"
)

// Runner parses one file. Side makes a fresh one: once per pass long-lived,
// once per file cold.
type (
	Runner func(path string)
	Side   func() Runner
)

// File is one file of the population and the corpus it came from.
type File struct{ Path, Corpus string }

// Protocol times Chart against Ref (recursive descent, or lexer.Tokenize for
// lex-overhead). Now and GC are injected so tests can run it on a fake clock.
// lex-overhead is measured only when the 1-minute load read from LoadPath is
// at most LoadLimit (lexOverheadKeys); tests point LoadPath at a file.
type Protocol struct {
	Chart, Ref Side
	Now        func() time.Duration
	GC         func()
	Stderr     io.Writer
	LoadPath   string
	LoadLimit  float64
}

const (
	warmupPairs  = 2
	timedPairs   = 9
	measurements = 3
)

// pass times one side over every file once, after a GC outside the span.
func (p Protocol) pass(side Side, files []string, cold bool) time.Duration {
	p.GC()
	start := p.Now()
	var run Runner
	for _, f := range files {
		if run == nil || cold {
			run = side()
		}
		run(f)
	}
	return p.Now() - start
}

// Ratio is the median of three measurements of chart over Ref on files.
func (p Protocol) Ratio(name string, files []string, cold bool) float64 {
	ms := make([]float64, measurements)
	for m := range ms {
		ratios := make([]float64, 0, timedPairs)
		for i := range warmupPairs + timedPairs {
			c := p.pass(p.Chart, files, cold)
			r := p.pass(p.Ref, files, cold)
			if i >= warmupPairs {
				ratios = append(ratios, float64(c)/float64(r))
			}
		}
		slices.Sort(ratios)
		ms[m] = ratios[timedPairs/2]
		fmt.Fprintf(p.Stderr, "chart-ratchet: %s measurement %d: median %.4g min %.4g max %.4g\n",
			name, m+1, ms[m], ratios[0], ratios[timedPairs-1])
	}
	slices.Sort(ms)
	return ms[measurements/2]
}

// Aggregate returns timed less std5k.pl, edit5k.pl and every file whose
// recursive-descent time alone is more than 10% of recursive descent's over
// the whole population. rd is passed explicitly, not taken from p.Ref, so
// lex-overhead-aggregate and the speed ratios get one aggregate. A timed file outside
// the population is a caller bug, so it panics.
func (p Protocol) Aggregate(rd Side, population []File, timed []string) []string {
	rdTime := make(map[string]time.Duration, len(population))
	corpus := make(map[string]string, len(population))
	var total time.Duration
	for _, f := range population {
		rdTime[f.Path] = p.pass(rd, []string{f.Path}, true)
		total += rdTime[f.Path]
		corpus[f.Path] = f.Corpus
	}
	var kept []string
	share := map[string]time.Duration{}
	var keptTotal time.Duration
	for _, f := range timed {
		if _, ok := rdTime[f]; !ok {
			panic("chart-ratchet: aggregate: timed file " + f + " is not in the population")
		}
		if b := filepath.Base(f); b == "std5k.pl" || b == "edit5k.pl" || rdTime[f]*10 > total {
			fmt.Fprintf(p.Stderr, "chart-ratchet: aggregate excludes %s (%.1f%% of RD's population time)\n",
				f, 100*float64(rdTime[f])/float64(total))
			continue
		}
		kept = append(kept, f)
		share[corpus[f]] += rdTime[f]
		keptTotal += rdTime[f]
	}
	for _, c := range slices.Sorted(maps.Keys(share)) {
		fmt.Fprintf(p.Stderr, "chart-ratchet: aggregate share %s %.1f%%\n", c, 100*float64(share[c])/float64(keptTotal))
	}
	return kept
}

// chartSide is the chart as a protocol side: construct makes one recognizer,
// which the side calls on a file's bytes. Side makes a Runner per pass
// (long-lived) or per file (cold), so the lifetimes live in the protocol.
func chartSide(srcs map[string][]byte, construct func() func(src []byte)) Side {
	return func() Runner {
		recognizeSrc := construct()
		return func(path string) { recognizeSrc(srcs[path]) }
	}
}

// recognizerFor is chart/lex then recognition on a fresh recognizer. A
// recognizer error panics: the verdict keys have already recognized every
// corpus file, so one here is a bug, not a slow file.
func recognizerFor(tab *grammar.Tables) func() func([]byte) {
	return func() func([]byte) {
		r := recognize.New(tab, recognize.Options{})
		return func(src []byte) {
			s, _ := lex.Lex(src, tab.Tokens)
			if _, err := recognize.Recognize(r, lex.Linear(s)); err != nil {
				panic("chart-ratchet: speed: " + err.Error())
			}
		}
	}
}

// allocsPer is side's heap allocations per token over files, one cold pass
// after a GC (§9.3: a run too short to trigger a collection would hide it).
func allocsPer(side Side, files []string, tokens int) float64 {
	var before, after runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)
	for _, f := range files {
		side()(f)
	}
	runtime.ReadMemStats(&after)
	return float64(after.Mallocs-before.Mallocs) / float64(tokens)
}

// measureSpeed adds the six speed and allocation keys to measured: p's Chart
// against its Ref (recursive descent) on std5k.pl alone and on the aggregate
// agg, long-lived and cold; allocations are per lexer token over agg.
func measureSpeed(p Protocol, agg []string, srcs map[string][]byte, measured map[string]string) {
	one := []string{std5k}
	for _, r := range []struct {
		key   string
		files []string
		cold  bool
	}{
		{"speed-ratio-std5k-longlived", one, false}, {"speed-ratio-std5k-cold", one, true},
		{"speed-ratio-aggregate-longlived", agg, false}, {"speed-ratio-aggregate-cold", agg, true},
	} {
		measured[r.key] = fmt.Sprintf("%.4f", p.Ratio(r.key, r.files, r.cold))
	}
	tokens := 0
	for _, f := range agg {
		tokens += len(lexer.Tokenize(srcs[f]))
	}
	measured["alloc-per-token-chart"] = fmt.Sprintf("%.3f", allocsPer(p.Chart, agg, tokens))
	measured["alloc-per-token-rd"] = fmt.Sprintf("%.3f", allocsPer(p.Ref, agg, tokens))
}

// measureStages adds the three speed-stage-*-aggregate-longlived keys: chart
// lex, recognition and the forest each timed alone against p's Ref (recursive
// descent) on agg, long-lived, each stage's inputs prepared outside its span
// (lattices for recognition, charts for the forest). They are in RD units, so
// their sum, printed on stderr and not a key, is the whole chart side's.
func measureStages[L, C any](p Protocol, agg []string, lexOf func(path string) L, recognizer func() func(L) C,
	forestOf func(C, L), measured map[string]string) {
	ls, cs := make(map[string]L, len(agg)), make(map[string]C, len(agg))
	rec := recognizer()
	for _, f := range agg {
		ls[f] = lexOf(f)
		cs[f] = rec(ls[f])
	}
	sum := 0.0
	for _, s := range []struct {
		key  string
		side Side
	}{
		{"speed-stage-lex-aggregate-longlived", func() Runner { return func(f string) { lexOf(f) } }},
		{"speed-stage-recognize-aggregate-longlived", func() Runner { rec := recognizer(); return func(f string) { rec(ls[f]) } }},
		{"speed-stage-forest-aggregate-longlived", func() Runner { return func(f string) { forestOf(cs[f], ls[f]) } }},
	} {
		p.Chart = s.side
		r := p.Ratio(s.key, agg, false)
		sum += r
		measured[s.key] = fmt.Sprintf("%.4f", r)
	}
	fmt.Fprintf(p.Stderr, "chart-ratchet: speed-stage sum lex+recognize+forest %.4g\n", sum)
}

// chartStages times the chart's stages (measureStages) over tab. Like
// recognizerFor, a recognizer error panics; a refused file has no forest
// (forest.Parse's rule), so its forest stage does nothing.
func chartStages(p Protocol, tab *grammar.Tables, agg []string, srcs map[string][]byte, measured map[string]string) {
	measureStages(p, agg,
		func(path string) *lex.Lattice[lex.None] {
			s, _ := lex.Lex(srcs[path], tab.Tokens)
			return lex.Linear(s)
		},
		func() func(*lex.Lattice[lex.None]) *recognize.Chart {
			r := recognize.New(tab, recognize.Options{})
			return func(l *lex.Lattice[lex.None]) *recognize.Chart {
				c, err := recognize.Recognize(r, l)
				if err != nil {
					panic("chart-ratchet: speed: " + err.Error())
				}
				return c
			}
		},
		func(c *recognize.Chart, l *lex.Lattice[lex.None]) {
			if c.Accepted() {
				forest.Build(tab, c, l)
			}
		}, measured)
}

// measureBench runs chart's BenchLex, BenchRecognize and BenchExpected, and forest's Bench as BenchForest, on std5k.pl
// (spec §9.3: benchmarks run in the ratchet, not in make test) and writes one measurement line each to stderr. They are
// not keys.
//
// A b.Fatal inside testing.Benchmark panics here, outside a test binary, so the one step that can fail,
// recognizing src, is checked first, and a failure, or a source the chart refuses, is returned naming the benchmarks
// that run it.
func measureBench(tab *grammar.Tables, src []byte, stderr io.Writer) error {
	s, _ := lex.Lex(src, tab.Tokens)
	c, err := recognize.Recognize(recognize.New(tab, recognize.Options{}), lex.Linear(s))
	if err != nil {
		return fmt.Errorf("BenchRecognize, BenchExpected and BenchForest: %w", err)
	}
	if !c.Accepted() { // BenchForest would time nothing and report it as a figure
		return fmt.Errorf("BenchForest: the chart refuses the source, so it has no forest to time")
	}
	for _, b := range []struct {
		name string
		f    func(*testing.B, *grammar.Tables, []byte)
	}{{"BenchLex", chart.BenchLex}, {"BenchRecognize", chart.BenchRecognize}, {"BenchExpected", chart.BenchExpected}, {"BenchForest", forest.Bench}} {
		r := testing.Benchmark(func(tb *testing.B) { b.f(tb, tab, src) })
		fmt.Fprintf(stderr, "%s%s %d ns/op %d allocs/op\n", linePrefix, b.name, r.NsPerOp(), r.AllocsPerOp())
	}
	return nil
}
