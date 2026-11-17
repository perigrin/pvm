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
	"tamarou.com/pvm/internal/chart/forest"
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

// chartResult is one file's chart verdict and recognition: whether the
// whole pipeline (forest.Parse) accepts it, whether recognition alone does,
// its mapped tokens (those whose terminal set is not empty) and the
// terminals those sets hold, and over the nodes after the first that have an
// Earley set (Positions): their items and how many sets were already interned.
// An accepted file's forest has Statements, Collapsed of them (forest.Readings).
type chartResult struct {
	Accepted, Recognized   bool
	Tokens, Terminals      int
	Positions, Items, Hits int
	Statements, Collapsed  int
}

// verdictInputs are the sides of a verdict, by corpus path. Perl returns nil
// for a file with no label.
// Excluded is Q11's set: the files on either lexer list. Scope is a file's
// scope facts, from the verdict RD's tree.
type verdictInputs struct {
	Chart    func(path string) (chartResult, error)
	RD       func(path string) (bool, error)
	Perl     func(path string) *perlVerdict
	Scope    func(path string) forest.Scope
	Excluded map[string]bool
}

// gatedCategory is perl's refusal categories chart-accepted-both-refused
// counts: what perly.y refuses, and, since the verdict is the whole
// pipeline, what the precedence constraint and resolve refuse (spec §9.2).
var gatedCategory = map[string]bool{"grammar": true, "precedence": true, "scope": true}

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
		if f.Corpus == "perlgit" { // the width and cost keys are not result keys (perigrin, 2026-11-15, results first)
			continue
		}
		tokens, terminals = tokens+r.Tokens, terminals+r.Terminals
		if r.Recognized { // §7's cost is the cost on recognized input, whatever the pipeline does after
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

	var pop, popPerlGit, recognized, perlRefused, rdPerlRefused, beats, collapsed, over int
	statements, collapses := map[string]int{}, map[string]int{} // by corpus
	refused := map[string][]string{"rd-accepted-chart-refused": nil, "chart-accepted-both-refused": nil}
	for i, f := range files {
		v, accepted := verdicts[i], res[i].Accepted
		statements[f.Corpus] += res[i].Statements
		collapses[f.Corpus] += res[i].Collapsed
		collapsed += res[i].Collapsed
		if v.Refused && rd[i] {
			rdPerlRefused++
		}
		if !v.Refused && accepted && !rd[i] { // 41b5: recursive descent resolves imports here
			beats++
		}
		if v.Refused && accepted {
			perlRefused++
			if gatedCategory[v.Category] && !rd[i] && !in.Excluded[f.Path] {
				refused["chart-accepted-both-refused"] = append(refused["chart-accepted-both-refused"], f.Path)
			}
		}
		if inPopulation(v, rd[i], in.Excluded[f.Path]) { // §9.1: less both lexer lists
			pop++
			if f.Corpus == "perlgit" {
				popPerlGit++
			}
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
	measured["superset-population-perlgit"] = strconv.Itoa(popPerlGit)
	measured["chart-beats-rd"] = strconv.Itoa(beats)
	for c, n := range collapses { // spec §10's stop: more than 1%, and at least 5, of a corpus's statements
		if n >= 5 && n*100 > statements[c] {
			over++
		}
	}
	measured["readings-collapsed"] = strconv.Itoa(collapsed)
	measured["collapse-over-limit"] = strconv.Itoa(over)
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
	for _, line := range textLines(text) {
		out[strings.Fields(line)[0]] = true
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
// (offline/label.sh's in-tree.txt, t1.txt and perlgit.txt), by corpus path.
func labelVerdicts(dir string) (func(path string) *perlVerdict, error) {
	out := map[string]*perlVerdict{}
	for name, prefix := range map[string]string{"in-tree.txt": "", "t1.txt": "T1/", "perlgit.txt": "perlgit/"} {
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
	return chartInputs(tab, srcs, excluded, perl, realDirs), nil
}

// realDirs are the directories recursive descent searches for a corpus
// path's imports, the file's real directory first: T1 files are named
// T1/<name> and live in the T1 directory; perl.git's are named
// perlgit/<path under t/> and search as the oracle's RD subject does from
// t/ (".", "..", "../lib"); every other path is real.
func realDirs(path string) []string {
	if strings.HasPrefix(path, "T1/") {
		return []string{t1Dir()}
	}
	if rel, ok := strings.CutPrefix(path, "perlgit/"); ok {
		t := filepath.Join(perl5Root(), "t")
		return []string{filepath.Join(t, filepath.Dir(rel)), t, perl5Root(), filepath.Join(perl5Root(), "lib")}
	}
	return []string{filepath.Dir(path)}
}

// ownDir is realDirs for a path that is real: its own directory alone.
func ownDir(path string) []string { return []string{filepath.Dir(path)} }

// chartInputs are the real sides over srcs. The chart's verdict is the whole
// pipeline, forest.Parse under the file's scope facts; its counts are
// recognition's, chart/lex then a fresh recognizer. Recursive descent follows
// `use` as psc does, and its tree gives the scope facts (parse.Imports), each
// file parsed once; perl is the committed labels. The speed side's recursive
// descent stays bare (speedParse).
func chartInputs(tab *grammar.Tables, srcs map[string][]byte, excluded map[string]bool, perl func(string) *perlVerdict, dirs func(string) []string) verdictInputs {
	type rdResult struct {
		ok    bool
		scope forest.Scope
		err   error
	}
	rds := map[string]rdResult{}
	rd := func(path string) rdResult {
		if r, ok := rds[path]; ok {
			return r
		}
		var r rdResult
		r.ok, r.err = rdAccepts(path, func() *parse.Node {
			// ParseFileFrom's search: the file's own directory, then its roots, then the declarations.
			n := parse.ParseWithLoader(srcs[path], parse.DirLoader(dirs(path)...))
			r.scope.Imports = parse.Imports(n)
			return n
		})
		rds[path] = r
		return r
	}
	scope := func(path string) forest.Scope { return rd(path).scope }
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
			r.Recognized = c.Accepted()
			for _, set := range c.Sets[1:] {
				r.Items += len(set.Items())
			}
			// Node 0 is a set of its own in a fresh recognizer, never a hit.
			r.Hits, r.Positions = c.Reuse()
			r.Positions--
			if r.Recognized {
				f, ref := forest.Parse(tab, srcs[path], scope(path))
				if !pipelineReason[ref.Reason] {
					return r, fmt.Errorf("the pipeline: %s", ref.Reason)
				}
				r.Accepted = ref.Reason == ""
				if r.Accepted {
					r.Statements, r.Collapsed = f.Readings.Statements, f.Readings.Collapsed
				}
			}
			return r, nil
		},
		RD:       func(path string) (bool, error) { r := rd(path); return r.ok, r.err },
		Perl:     perl,
		Scope:    scope,
		Excluded: excluded,
	}
}

// pipelineReason is forest.Parse's verdicts: accepted (""), or refused for
// one of its reasons. Any other reason is an error inside it.
var pipelineReason = map[string]bool{"": true, "lex-unmapped": true, "grammar": true, "precedence": true, "scope": true}
