// ABOUTME: The ratchet's verdict keys: what the chart, recursive descent and perl say about each corpus file (spec §9.2).
// ABOUTME: A file with no perl label is an error; Q11 leaves lexer-list files out of refusal counts.

package main

import (
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"tamarou.com/pvm/internal/chart"
	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/chart/lex"
	"tamarou.com/pvm/internal/chart/recognize"
	"tamarou.com/pvm/internal/parse"
)

// perlVerdict is perl's verdict on one file: whether it refuses, and in which
// category (spec §9.1).
type perlVerdict struct {
	Refused  bool
	Category string
}

// chartResult is one file's recognition: whether the chart accepts it, its
// mapped tokens (those whose terminal set is not empty) and the terminals
// those sets hold, and over the nodes after the first that have an Earley
// set (Positions): their items and how many sets were already interned.
type chartResult struct {
	Accepted               bool
	Tokens, Terminals      int
	Positions, Items, Hits int
}

// verdictInputs are the sides of a verdict, by corpus path. Perl returns nil
// for a file with no label.
// Excluded is Q11's set: the files on either lexer list.
type verdictInputs struct {
	Chart    func(path string) (chartResult, error)
	RD       func(path string) (bool, error)
	Perl     func(path string) *perlVerdict
	Excluded map[string]bool
}

// measureVerdicts adds the verdict keys over files to measured. It returns
// the files behind each gated count, for listRises. A recognizer error is an
// error naming the file, never a refusal, and so is a file with no perl
// label (supersetPopulation refuses one first in a real run).
func measureVerdicts(files []File, in verdictInputs, measured map[string]string) (map[string][]string, error) {
	var tokens, terminals, positions, items, hits int
	var verdicts []*perlVerdict
	var res []chartResult
	var rd []bool
	for _, f := range files {
		r, err := in.Chart(f.Path)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", f.Path, err)
		}
		a, err := in.RD(f.Path)
		if err != nil {
			return nil, err
		}
		v := in.Perl(f.Path)
		if v == nil {
			return nil, fmt.Errorf("%s has no perl label", f.Path)
		}
		res, rd, verdicts = append(res, r), append(rd, a), append(verdicts, v)
		tokens, terminals = tokens+r.Tokens, terminals+r.Terminals
		if r.Accepted { // §7's cost is the cost on recognized input
			positions, items, hits = positions+r.Positions, items+r.Items, hits+r.Hits
		}
	}
	ratio := func(key string, n, d int) {
		if d > 0 {
			measured[key] = fmt.Sprintf("%.3f", float64(n)/float64(d))
		}
	}
	ratio("terminals-per-token", terminals, tokens)
	ratio("items-per-token", items, positions)
	ratio("set-reuse", hits, positions)

	var pop, recognized, perlRefused, rdPerlRefused int
	refused := map[string][]string{"rd-accepted-chart-refused": nil, "chart-accepted-both-refused": nil}
	for i, f := range files {
		v, accepted := verdicts[i], res[i].Accepted
		if v.Refused && rd[i] {
			rdPerlRefused++
		}
		if v.Refused && accepted {
			perlRefused++
			if v.Category == "grammar" && !rd[i] && !in.Excluded[f.Path] {
				refused["chart-accepted-both-refused"] = append(refused["chart-accepted-both-refused"], f.Path)
			}
		}
		if inPopulation(v, rd[i], in.Excluded[f.Path]) { // §9.1: less both lexer lists
			pop++
			if accepted {
				recognized++
			} else {
				refused["rd-accepted-chart-refused"] = append(refused["rd-accepted-chart-refused"], f.Path)
			}
		}
	}
	for k, fs := range refused {
		measured[k] = strconv.Itoa(len(fs))
	}
	measured["chart-accepted-perl-refused"] = strconv.Itoa(perlRefused)
	measured["superset-population"] = strconv.Itoa(pop)
	measured["rd-accepted-perl-refused"] = strconv.Itoa(rdPerlRefused)
	ratio("superset-recognized", recognized, pop)
	return refused, nil
}

// listRises names on w each file behind a gated count that rose against the
// baseline text, so a regression says which files to look at.
func listRises(w io.Writer, baselineText string, ms []Measure, lists map[string][]string) {
	baseline, err := Parse(baselineText)
	if err != nil {
		return // check reports a malformed baseline
	}
	for _, b := range baseline {
		for _, m := range ms {
			if m.Key == b.Key && m.Value != na && b.Value != na && number(m.Value) > number(b.Value) {
				for _, f := range lists[m.Key] {
					fmt.Fprintf(w, "%s%s: %s\n", linePrefix, m.Key, f)
				}
			}
		}
	}
}

// listedPaths is the set of paths on a lexer list, whose lines are
// `<path> <byte> <issue ref>`.
func listedPaths(text string) map[string]bool {
	out := map[string]bool{}
	for _, line := range strings.Split(text, "\n") {
		if f := strings.Fields(line); len(f) > 0 {
			out[f[0]] = true
		}
	}
	return out
}

// lexerListed is Q11's exclusion: the paths on lex-unmapped-known.txt and
// lexer-defects.txt, and no others.
func lexerListed() (map[string]bool, error) {
	out := map[string]bool{}
	for _, f := range []string{knownFile, defectsFile} {
		text, err := os.ReadFile(f)
		if err != nil {
			return nil, err
		}
		maps.Copy(out, listedPaths(string(text)))
	}
	return out, nil
}

// labelVerdicts reads perl's verdicts from the committed label files in dir
// (offline/label.sh's in-tree.txt and t1.txt), by corpus path.
func labelVerdicts(dir string) (func(path string) *perlVerdict, error) {
	out := map[string]*perlVerdict{}
	for name, prefix := range map[string]string{"in-tree.txt": "", "t1.txt": "T1/"} {
		lf, err := chart.ReadLabels(filepath.Join(dir, name))
		if err != nil {
			return nil, err
		}
		for _, l := range lf.Labels {
			out[prefix+l.Path] = &perlVerdict{Refused: l.Refused(), Category: l.Category}
		}
	}
	return func(path string) *perlVerdict { return out[path] }, nil
}

// realInputs are chartInputs with the committed lexer lists and perl labels.
func realInputs(tab *grammar.Tables, srcs map[string][]byte) (verdictInputs, error) {
	excluded, err := lexerListed()
	if err != nil {
		return verdictInputs{}, err
	}
	perl, err := labelVerdicts(filepath.Join(testdata, "labels"))
	if err != nil {
		return verdictInputs{}, err
	}
	return chartInputs(tab, srcs, excluded, perl), nil
}

// chartInputs are the real sides over srcs: the chart is chart/lex then a
// fresh recognizer per file, recursive descent is parse.Parse, perl is the
// committed labels.
func chartInputs(tab *grammar.Tables, srcs map[string][]byte, excluded map[string]bool, perl func(string) *perlVerdict) verdictInputs {
	return verdictInputs{
		Chart: func(path string) (chartResult, error) {
			s, _ := lex.Lex(srcs[path], tab.Tokens) // an unmapped token has the empty set, which refuses
			var r chartResult
			for _, t := range s.Tokens {
				if t.TerminalSet != 0 { // the empty set is a token the table cannot map
					r.Tokens++
					r.Terminals += len(tab.TerminalSets[t.TerminalSet])
				}
			}
			c, err := recognize.Recognize(recognize.New(tab, recognize.Options{}), lex.Linear(s))
			if err != nil {
				return r, err
			}
			r.Accepted = c.Accepted()
			for _, set := range c.Sets[1:] {
				r.Items += len(set.Items())
			}
			// Node 0 is a set of its own in a fresh recognizer, never a hit.
			r.Hits, r.Positions = c.Reuse()
			r.Positions--
			return r, nil
		},
		RD: func(path string) (bool, error) {
			return rdAccepts(path, func() *parse.Node { return parse.Parse(srcs[path]) })
		},
		Perl:     perl,
		Excluded: excluded,
	}
}
