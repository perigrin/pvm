// ABOUTME: The ratchet's chart/lex keys: lex-unmapped over every corpus file, and lex-overhead timed by the §9.3 protocol.
// ABOUTME: lex-overhead is the larger of std5k.pl's and edit5k.pl's long-lived figures (spec §9.3, amended at fdb2ece2 and by perigrin 2026-11-10: 25 buffers a pass).

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"time"

	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/chart/lex"
	"tamarou.com/pvm/internal/lexer"
	"tamarou.com/pvm/internal/parse"
)

// defaultT1 is the in-tree copy of T1, relative to the repository root the ratchet runs from.
const defaultT1 = "testdata/perlonjava/unit"

var (
	std5k  = filepath.Join(testdata, "std5k.pl")
	edit5k = filepath.Join("internal", "psc", "testdata", "latency", "edit5k.pl")
)

// corpus is every waypoint-1 corpus file, by the path the lexer lists use
// (T1 files as T1/<name>.t), with its bytes. A corpus with no file is an error.
func corpus() ([]File, map[string][]byte, error) {
	t1, t1From := os.Getenv("T1_CORPUS"), "T1_CORPUS"
	if t1 == "" {
		t1, t1From = defaultT1, "the in-tree default"
	}
	var files []File
	srcs := map[string][]byte{}
	for _, c := range []struct{ name, glob string }{
		{"guacamole", filepath.Join(testdata, "guacamole", "parses", "*.pl")},
		{"guacamole", filepath.Join(testdata, "guacamole", "parsent", "*.pl")},
		{"std5k", std5k},
		{"edit5k", edit5k},
		{"T1", filepath.Join(t1, "*.t")},
	} {
		paths, err := filepath.Glob(c.glob)
		if err != nil || len(paths) == 0 {
			from := ""
			if c.name == "T1" {
				from = " (T1 from " + t1From + ")"
			}
			return nil, nil, fmt.Errorf("corpus %s: no file matches %s%s", c.name, c.glob, from)
		}
		for _, p := range paths {
			src, err := os.ReadFile(p)
			if err != nil {
				return nil, nil, err
			}
			name := p
			if c.name == "T1" { // named as the lexer lists name it
				name = "T1/" + filepath.Base(p)
			}
			files = append(files, File{name, c.name})
			srcs[name] = src
		}
	}
	return files, srcs, nil
}

// lexSides are chart/lex and lexer.Tokenize as protocol sides over srcs.
// The chart side calls tokenize inside its span, so lex-overhead's numerator
// is the whole of chart/lex, the lexer included.
func lexSides(srcs map[string][]byte, tt *grammar.TokenTable, tokenize func([]byte) []lexer.Token) (chart, ref Side) {
	chart = func() Runner { return func(p string) { lex.Map(srcs[p], tokenize(srcs[p]), tt) } }
	ref = func() Runner { return func(p string) { tokenize(srcs[p]) } }
	return chart, ref
}

// aggregateFiles is the aggregate of the superset population over files, the
// one set that lex-overhead-aggregate and the speed ratios all time.
func aggregateFiles(p Protocol, rd Side, files []File, in verdictInputs) ([]string, error) {
	population, err := supersetPopulation(files, in)
	if err != nil {
		return nil, err
	}
	timed := make([]string, len(population))
	for i, f := range population {
		timed[i] = f.Path
	}
	return p.Aggregate(rd, population, timed), nil
}

// lexRepeat is how many times a lex-overhead pass lexes its buffer, so a
// pass is about 200 ms of work rather than 8 to 10 ms: long enough that the
// scheduler's delays under load average out (perigrin, 2026-11-10, the
// definition's fourth amendment).
const lexRepeat = 25

// lexOverheadKeys adds lex-overhead, the larger of the two LSP buffers'
// long-lived figures, each pass lexing its buffer lexRepeat times (spec
// §9.3, amended), lex-overhead-std5k and lex-overhead-aggregate.
// The aggregate's cold-start figure is reported, never gated.
func lexOverheadKeys(p Protocol, agg []string, measured map[string]string) {
	s := p.Ratio("lex-overhead-std5k", slices.Repeat([]string{std5k}, lexRepeat), false)
	e := p.Ratio("lex-overhead-edit5k", slices.Repeat([]string{edit5k}, lexRepeat), false)
	measured["lex-overhead"] = fmt.Sprintf("%.3f", max(s, e))
	measured["lex-overhead-std5k"] = fmt.Sprintf("%.3f", s)
	measured["lex-overhead-aggregate"] = fmt.Sprintf("%.3f", p.Ratio("lex-overhead-aggregate", agg, false))
}

// productionProtocol is the Protocol the ratchet times with: the monotonic
// wall clock, and runtime.GC, which pass calls before its span starts.
func productionProtocol(chart, ref Side) Protocol {
	start := time.Now()
	return Protocol{Chart: chart, Ref: ref, Now: func() time.Duration { return time.Since(start) }, GC: runtime.GC, Stderr: os.Stderr}
}

// measureLex adds lex-unmapped over files, lex-overhead over the LSP buffers,
// lex-overhead-std5k, lex-overhead-aggregate and the speed keys (measureSpeed) to measured.
// It returns lex-unmapped's files as "<path> <byte>" pairs.
func measureLex(tab *grammar.Tables, files []File, in verdictInputs, srcs map[string][]byte, measured map[string]string) ([]string, error) {
	tt := tab.Tokens
	var unmapped []string
	for _, f := range files {
		if _, at := lex.Lex(srcs[f.Path], tt); at >= 0 {
			unmapped = append(unmapped, fmt.Sprintf("%s %d", f.Path, at))
		}
	}
	measured["lex-unmapped"] = strconv.Itoa(len(unmapped))

	p := productionProtocol(lexSides(srcs, tt, lexer.Tokenize))
	rd := func() Runner { return func(path string) { parse.Parse(srcs[path]) } }
	agg, err := aggregateFiles(p, rd, files, in)
	if err != nil {
		return nil, err
	}
	lexOverheadKeys(p, agg, measured)
	p.Chart, p.Ref = chartSide(srcs, recognizerFor(tab)), rd
	measureSpeed(p, agg, srcs, measured)
	return unmapped, nil
}
