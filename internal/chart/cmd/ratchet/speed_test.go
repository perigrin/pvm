// ABOUTME: Tests the §9.3 speed protocol on a fake clock: alternation, GC outside the timed span, pairs, medians.
// ABOUTME: Also the aggregate's exclusions: std5k.pl, edit5k.pl and the strict 10%-of-RD rule over the whole population.

package main

import (
	"fmt"
	"math"
	"slices"
	"strings"
	"testing"
	"time"
)

// fakeWorld is a deterministic clock that only moves when a side runs a
// file: rd costs cost[path], the chart cost[path] times chartMult(pair),
// where pair counts pairs from pass starts (a pass reads Now twice, a pair
// four times), so a stray GC shows in the event log rather than in pair.
type fakeWorld struct {
	t         time.Duration
	nows      int
	log       []string
	cost      map[string]time.Duration
	chartMult func(pair int) float64
	stderr    strings.Builder
}

func (w *fakeWorld) protocol() Protocol {
	side := func(name string) Side {
		return func() Runner {
			w.log = append(w.log, "new "+name)
			return func(path string) {
				w.log = append(w.log, name+" "+path)
				c := w.cost[path]
				if name == "chart" {
					c = time.Duration(float64(c) * w.chartMult(w.nows/4))
				}
				w.t += c
			}
		}
	}
	return Protocol{
		Chart:  side("chart"),
		Ref:    side("rd"),
		Now:    func() time.Duration { w.log = append(w.log, "now"); w.nows++; return w.t },
		GC:     func() { w.log = append(w.log, "gc") },
		Stderr: &w.stderr,
	}
}

func TestSpeedProtocol(t *testing.T) {
	files := []string{"a.pl", "b.pl"}
	// 3 measurements of 2 warm-up + 9 timed pairs. Warm-ups are absurd so
	// a protocol that times them shows. Each row has one outlier, so its
	// mean is not its median; the measurement medians 9.5, 0.5, 1.5 have a
	// mean (3.83) and a first value (9.5) that both differ from their median.
	timed := [][]float64{
		{9.9, 9.1, 9.8, 9.2, 9.5, 99, 9.3, 9.7, 9.4},
		{0.9, 0.1, 5, 0.2, 0.5, 0.7, 0.3, 0.6, 0.4},
		{1.1, 1.9, 1.2, 1.5, 1.7, 1.3, 1.6, 1.4, 20},
	}
	mult := func(pair int) float64 {
		m, i := pair/11, pair%11
		if m >= len(timed) {
			return 1 // a pair beyond the protocol's 33; the event log reports it
		}
		if i < 2 {
			return 0.01
		}
		return timed[m][i-2]
	}

	for _, cold := range []bool{false, true} {
		w := &fakeWorld{cost: map[string]time.Duration{"a.pl": 1000, "b.pl": 3000}, chartMult: mult}
		got := w.protocol().Ratio("test", files, cold)
		if math.Abs(got-1.5) > 1e-9 {
			t.Errorf("cold=%v: ratio %v, want the median of the three measurement medians, 1.5", cold, got)
		}

		// Alternation, chart first; a GC before each pass, outside the
		// timed span; one parser per pass long-lived, one per file cold.
		pass := func(name string) []string {
			l := []string{"gc", "now"}
			for i, f := range files {
				if cold || i == 0 {
					l = append(l, "new "+name)
				}
				l = append(l, name+" "+f)
			}
			return append(l, "now")
		}
		var want []string
		for range 3 * 11 {
			want = append(want, pass("chart")...)
			want = append(want, pass("rd")...)
		}
		if !slices.Equal(w.log, want) {
			t.Errorf("cold=%v: event log\n%q\nwant\n%q", cold, w.log, want)
		}

		for _, s := range []string{"median 9.5 min 9.1 max 99", "median 0.5 min 0.1 max 5", "median 1.5 min 1.1 max 20"} {
			if !strings.Contains(w.stderr.String(), s) {
				t.Errorf("cold=%v: stderr %q does not report %q", cold, w.stderr.String(), s)
			}
		}
	}

	// The aggregate: std5k.pl and edit5k.pl always out, a file above 10%
	// of RD's time over the population out, the rest in, shares printed.
	w := &fakeWorld{cost: map[string]time.Duration{
		"t/std5k.pl": 10, "x/edit5k.pl": 10, "t1/big.t": 300,
		"t1/a.t": 50, "t1/b.t": 50, "g/c.pl": 50, "g/d.pl": 50, "g/e.pl": 40,
	}}
	pop := []File{
		{"t/std5k.pl", "std5k"}, {"x/edit5k.pl", "edit5k"}, {"t1/big.t", "T1"},
		{"t1/a.t", "T1"}, {"t1/b.t", "T1"}, {"g/c.pl", "guacamole"}, {"g/d.pl", "guacamole"}, {"g/e.pl", "guacamole"},
	}
	var all []string
	for _, f := range pop {
		all = append(all, f.Path)
	}
	got := w.protocol().Aggregate(pop, all)
	want := []string{"t1/a.t", "t1/b.t", "g/c.pl", "g/d.pl", "g/e.pl"}
	if !slices.Equal(got, want) {
		t.Errorf("aggregate %q, want %q", got, want)
	}
	for _, s := range []string{"excludes t/std5k.pl", "excludes x/edit5k.pl", "excludes t1/big.t", "share T1 41.7%", "share guacamole 58.3%"} {
		if !strings.Contains(w.stderr.String(), s) {
			t.Errorf("stderr %q does not report %q", w.stderr.String(), s)
		}
	}
	// RD's per-file times are taken one file at a time, a GC before each.
	for i := 0; i < len(w.log); i++ {
		if strings.HasPrefix(w.log[i], "rd ") && (i < 3 || w.log[i-3] != "gc") {
			t.Fatalf("RD timing of %s not preceded by a GC outside its span: %q", w.log[i], w.log)
		}
	}
}

func TestSpeedProtocolExclusionBoundary(t *testing.T) {
	// Population total 1000: "edge" at exactly 10% stays, "over" just above is out.
	for _, tc := range []struct {
		cost    time.Duration
		wantOut bool
	}{{100, false}, {101, true}} {
		w := &fakeWorld{cost: map[string]time.Duration{"edge": tc.cost, "rest": 1000 - tc.cost}}
		pop := []File{{"edge", "T1"}, {"rest", "T1"}}
		got := w.protocol().Aggregate(pop, []string{"edge", "rest"})
		if out := !slices.Contains(got, "edge"); out != tc.wantOut {
			t.Errorf("%d of 1000: excluded=%v, want %v", tc.cost, out, tc.wantOut)
		}
	}

	// Over the whole population, not the timed subset: "sub" is half the
	// subset's RD time but 5% of the population's, so it stays.
	w := &fakeWorld{cost: map[string]time.Duration{"sub": 50, "mate": 50}}
	pop := []File{{"sub", "T1"}, {"mate", "T1"}}
	for i := range 9 {
		p := fmt.Sprintf("other%d", i)
		w.cost[p] = 100
		pop = append(pop, File{p, "T1"})
	}
	if got := w.protocol().Aggregate(pop, []string{"sub", "mate"}); !slices.Equal(got, []string{"sub", "mate"}) {
		t.Errorf("subset aggregate %q, want both files: the 10%% rule reads the whole population", got)
	}
}
