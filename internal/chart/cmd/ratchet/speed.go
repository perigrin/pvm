// ABOUTME: The §9.3 speed protocol: alternating chart/reference passes, GC outside the timed span, medians of pairs.
// ABOUTME: Also the aggregate's file set: std5k.pl, edit5k.pl and any file above 10% of RD's population time left out.

package main

import (
	"fmt"
	"io"
	"maps"
	"path/filepath"
	"slices"
	"time"
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
type Protocol struct {
	Chart, Ref Side
	Now        func() time.Duration
	GC         func()
	Stderr     io.Writer
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
// lex-overhead and the speed ratios get one aggregate. A timed file outside
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
