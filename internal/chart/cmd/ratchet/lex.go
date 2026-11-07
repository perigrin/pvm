// ABOUTME: The ratchet's chart/lex keys: lex-unmapped over every corpus file, and lex-overhead timed by the §9.3 protocol.
// ABOUTME: lex-overhead is the larger of std5k.pl's and edit5k.pl's long-lived figures (spec §9.3, amended at fdb2ece2).

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/chart/lex"
	"tamarou.com/pvm/internal/lexer"
	"tamarou.com/pvm/internal/parse"
)

const defaultT1 = "/opt/corpora/PerlOnJava/src/test/resources/unit"

var (
	std5k  = filepath.Join(testdata, "std5k.pl")
	edit5k = filepath.Join("internal", "psc", "testdata", "latency", "edit5k.pl")
)

// corpus is every waypoint-1 corpus file, by the path the lexer lists use
// (T1 files as T1/<name>.t), with its bytes. A corpus with no file is an error.
func corpus() ([]File, map[string][]byte, error) {
	t1 := os.Getenv("T1_CORPUS")
	if t1 == "" {
		t1 = defaultT1
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
			return nil, nil, fmt.Errorf("corpus %s: no file matches %s (set T1_CORPUS for T1)", c.name, c.glob)
		}
		for _, p := range paths {
			src, err := os.ReadFile(p)
			if err != nil {
				return nil, nil, err
			}
			name := p
			if c.name == "T1" { // outside the tree: named as the lexer lists name it
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
// one set that lex-overhead and the speed ratios all time.
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

// lexOverheadKeys adds lex-overhead, the larger of the two LSP buffers'
// long-lived figures (spec §9.3), lex-overhead-std5k and lex-overhead-aggregate.
// The aggregate's cold-start figure is reported, never gated.
func lexOverheadKeys(p Protocol, agg []string, measured map[string]string) {
	s, e := p.Ratio("lex-overhead-std5k", []string{std5k}, false), p.Ratio("lex-overhead-edit5k", []string{edit5k}, false)
	measured["lex-overhead"] = fmt.Sprintf("%.3f", max(s, e))
	measured["lex-overhead-std5k"] = fmt.Sprintf("%.3f", s)
	measured["lex-overhead-aggregate"] = fmt.Sprintf("%.3f", p.Ratio("lex-overhead-aggregate", agg, false))
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

	start := time.Now()
	chart, ref := lexSides(srcs, tt, lexer.Tokenize)
	p := Protocol{Chart: chart, Ref: ref, Now: func() time.Duration { return time.Since(start) }, GC: runtime.GC, Stderr: os.Stderr}
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
