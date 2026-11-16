// ABOUTME: The oracle grading measure (spec §9.4): the chart and recursive descent graded per site against perl's committed facts.
// ABOUTME: In-process, one recursive descent parse per file feeding both sides; not yet a ratchet key (KEYS makes it one).

package main

import (
	"errors"
	"io"

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

func defaultSubjects(tab *grammar.Tables) []oracleSubject { return nil }

func gradeOracle(corpora []oracleCorpus, subjects []oracleSubject, stderr io.Writer) (oracleGrading, error) {
	return oracleGrading{}, errors.New("not yet")
}
