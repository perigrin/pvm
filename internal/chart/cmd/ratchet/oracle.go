// ABOUTME: The oracle grading measure (spec §9.4): the chart and recursive descent graded per site against perl's committed facts.
// ABOUTME: In-process, one recursive descent parse per file feeding both sides; not yet a ratchet key (KEYS makes it one).

package main

import (
	"cmp"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"tamarou.com/pvm/internal/chart"
	"tamarou.com/pvm/internal/chart/forest"
	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/parse"
	"tamarou.com/pvm/internal/parseoracle"
)

// oracleCorpus is one committed facts file: its name in the split, the file, where each key's source sits, and the
// directories recursive descent searches after the file's own.
type oracleCorpus struct {
	name  string
	facts string
	path  func(key string) string
	roots []string
}

// oracleSubject answers parseoracle's subject contract for one file, given recursive descent's resolved tree of it.
type oracleSubject struct {
	name  string
	facts func(path string, src []byte, rd *parse.Node) parseoracle.SubjectFacts
}

// oracleGrading is each subject's sites per corpus, by bucket, and the records whose file could not be read.
type oracleGrading struct {
	Sites      map[string]map[string]parseoracle.Totals
	Unreadable int
}

// oracleCorpora are the committed facts (testdata/oracle) over the files they were taken from, each searched as
// oracle-facts.sh ran perl: in-tree with Guacamole's stubs, T1 from its own directory, perl.git t/ as
// internal/parse/cmd/subject resolves it, from t/, its parent and lib/.
func oracleCorpora() []oracleCorpus {
	dir := filepath.Join(testdata, "oracle")
	perl5 := cmp.Or(os.Getenv("PERL5_CORPUS"), "/opt/corpora/perl5")
	under := func(root, prefix string) func(string) string {
		return func(k string) string { return filepath.Join(root, strings.TrimPrefix(k, prefix)) }
	}
	return []oracleCorpus{
		{"in-tree", filepath.Join(dir, "in-tree.txt"), under("", ""), []string{filepath.Join(testdata, "guacamole", "stubs")}},
		{"T1", filepath.Join(dir, "t1.txt"), under(t1Dir(), "T1/"), nil},
		{"perl.git", filepath.Join(dir, "perlgit.txt"), under(filepath.Join(perl5, "t"), "perlgit/"),
			[]string{filepath.Join(perl5, "t"), perl5, filepath.Join(perl5, "lib")}},
	}
}

// defaultSubjects are the chart, under the scope recursive descent's tree resolves, and recursive descent itself,
// answering as internal/parse/cmd/subject does.
func defaultSubjects(tab *grammar.Tables) []oracleSubject {
	return []oracleSubject{
		{"chart", func(path string, src []byte, rd *parse.Node) parseoracle.SubjectFacts {
			return forest.SubjectFacts(tab, src, path, forest.Scope{Imports: parse.Imports(rd)})
		}},
		{"rd", rdFacts},
	}
}

// rdFacts is internal/parse/cmd/subject's answer, in-process: OK, every import's known prototype, and parse.Sites
// renumbered by parse.PerlLines, dropping a site a #line directive moved into another file.
func rdFacts(path string, src []byte, rd *parse.Node) parseoracle.SubjectFacts {
	facts := parseoracle.SubjectFacts{OK: true, Prototypes: map[string]string{}, CallSites: []parseoracle.SubjectCallSite{}}
	for name, imp := range parse.Imports(rd) {
		if imp.PrototypeKnown && imp.Prototype != "" {
			facts.Prototypes[name] = strings.TrimSuffix(strings.TrimPrefix(imp.Prototype, "("), ")")
		}
	}
	line, own := parse.PerlLines(src, path)
	for _, s := range parse.Sites(rd, src) {
		end := max(s.EndLine, s.Line)
		if own[s.Line] && own[end] {
			s.Line, s.EndLine = line[s.Line], line[end]
			facts.CallSites = append(facts.CallSites, s)
		}
	}
	return facts
}

// gradeOracle grades each subject on every record of every corpus with parseoracle.CompareFacts, counting perl's
// sites per bucket, and names each WRONG site on stderr as "oracle-grade: <subject> WRONG <file>:<line> <marker>".
// A corpus with no record is an error, never WRONG 0. A record whose file cannot be read is named on stderr and
// counted, each subject's verdict on it a whole-file no-answer.
func gradeOracle(corpora []oracleCorpus, subjects []oracleSubject, stderr io.Writer) (oracleGrading, error) {
	if len(corpora) == 0 {
		return oracleGrading{}, errors.New("oracle grading: no corpus")
	}
	g := oracleGrading{Sites: map[string]map[string]parseoracle.Totals{}}
	for _, s := range subjects {
		g.Sites[s.name] = map[string]parseoracle.Totals{}
	}
	session := parse.NewSession()
	for _, c := range corpora {
		of, err := chart.ReadOracleFacts(c.facts)
		if err != nil {
			return oracleGrading{}, err
		}
		if len(of.Records) == 0 {
			return oracleGrading{}, fmt.Errorf("oracle grading: %s holds no committed facts", c.facts)
		}
		for _, rec := range of.Records {
			path := c.path(rec.Key)
			src, err := os.ReadFile(path)
			var rd *parse.Node
			if err != nil {
				g.Unreadable++
				fmt.Fprintf(stderr, "oracle-grade: unreadable %s: %v\n", rec.Key, err)
			} else {
				rd = session.ParseSourceFrom(path, src, c.roots...)
			}
			for _, s := range subjects {
				v := parseoracle.Verdict{Bucket: parseoracle.BucketNoAnswer, Marker: parseoracle.MarkerNone}
				if rd != nil {
					sf := s.facts(path, src, rd)
					// what parseoracle's decoder marks: a nil list is a question not answered
					sf.KnowsCallSites, sf.KnowsPrototypes = sf.CallSites != nil, sf.Prototypes != nil
					v = parseoracle.CompareFacts(rec.Facts, sf)
				}
				buckets, err := parseoracle.Result{Path: path, Facts: rec.Facts, Verdict: v}.SiteBuckets()
				if err != nil {
					return oracleGrading{}, err
				}
				t := g.Sites[s.name][c.name]
				for _, b := range buckets {
					for i, n := range b {
						t[i] += n
					}
				}
				g.Sites[s.name][c.name] = t
				for _, f := range v.Findings {
					if f.Bucket == parseoracle.BucketWrong {
						fmt.Fprintf(stderr, "oracle-grade: %s WRONG %s:%d %s\n", s.name, path, f.Line, f.Marker)
					}
				}
			}
		}
	}
	return g, nil
}

// printOracle runs gradeOracle over oracleCorpora and prints, per subject and corpus, "oracle-grade: <subject>
// <corpus> exact <n> wider <n> wrong <n> no-answer <n>", then the unreadable records and the run time; 2 on an error.
func printOracle(g *grammar.Grammar, ref []grammar.Entry) int {
	start, corpora := time.Now(), oracleCorpora()
	tab, err := grammar.Generate(g, ref)
	var got oracleGrading
	if err == nil {
		got, err = gradeOracle(corpora, defaultSubjects(tab), os.Stderr)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "oracle-grade: %v\n", err)
		return 2
	}
	for _, s := range []string{"chart", "rd"} {
		for _, c := range corpora {
			t := got.Sites[s][c.name]
			fmt.Printf("oracle-grade: %s %s exact %d wider %d wrong %d no-answer %d\n", s, c.name, t[parseoracle.BucketExact],
				t[parseoracle.BucketWider], t[parseoracle.BucketWrong], t[parseoracle.BucketNoAnswer])
		}
	}
	fmt.Printf("oracle-grade: unreadable %d\noracle-grade: seconds %.1f\n", got.Unreadable, time.Since(start).Seconds())
	return 0
}
