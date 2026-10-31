// ABOUTME: The ratchet's chart/lex keys: lex-unmapped over every corpus file, and lex-overhead timed by the §9.3 protocol.
// ABOUTME: Until the superset population exists lex-overhead reads every corpus file (perigrin, "All files until POP").

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
)

const defaultT1 = "/opt/corpora/PerlOnJava/src/test/resources/unit"

var std5k = filepath.Join(testdata, "std5k.pl")

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
		{"edit5k", filepath.Join("internal", "psc", "testdata", "latency", "edit5k.pl")},
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

// measureLex adds lex-unmapped, lex-overhead and lex-overhead-std5k to
// measured, and returns lex-unmapped's files as "<path> <byte>" pairs.
func measureLex(tt *grammar.TokenTable, measured map[string]string) ([]string, error) {
	files, srcs, err := corpus()
	if err != nil {
		return nil, err
	}
	var unmapped, all []string
	for _, f := range files {
		all = append(all, f.Path)
		if _, at := lex.Lex(srcs[f.Path], tt); at >= 0 {
			unmapped = append(unmapped, fmt.Sprintf("%s %d", f.Path, at))
		}
	}
	measured["lex-unmapped"] = strconv.Itoa(len(unmapped))

	start := time.Now()
	chart, ref := lexSides(srcs, tt, lexer.Tokenize)
	p := Protocol{Chart: chart, Ref: ref, Now: func() time.Duration { return time.Since(start) }, GC: runtime.GC, Stderr: os.Stderr}
	measured["lex-overhead"] = fmt.Sprintf("%.3f", p.Ratio("lex-overhead", all, false))
	measured["lex-overhead-std5k"] = fmt.Sprintf("%.3f", p.Ratio("lex-overhead-std5k", []string{std5k}, false))
	return unmapped, nil
}
