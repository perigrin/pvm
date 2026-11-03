// ABOUTME: The superset population (spec §9.1): files perl and recursive descent both accept, less both lexer lists.
// ABOUTME: Reads the committed perl labels and recursive descent's live verdict; a missing label or panic is an error.

package main

import (
	"fmt"

	"tamarou.com/pvm/internal/parse"
)

// refusedAnywhere reports whether n or any node under it is a refusal.
func refusedAnywhere(n *parse.Node) bool {
	if n.Refused() {
		return true
	}
	for _, c := range n.Children {
		if refusedAnywhere(c) {
			return true
		}
	}
	return false
}

// rdAccepts is recursive descent's verdict on path: it accepts when no node of
// the tree parseTree builds is a refusal. A panic is an error naming the file,
// never an accept and never a file dropped from the corpus.
func rdAccepts(path string, parseTree func() *parse.Node) (accepted bool, err error) {
	defer func() {
		if r := recover(); r != nil {
			accepted, err = false, fmt.Errorf("%s: recursive descent panicked: %v", path, r)
		}
	}()
	return !refusedAnywhere(parseTree()), nil
}

// inPopulation is the membership rule, shared by the population and the
// verdict keys so superset-population and superset-recognized cannot disagree.
func inPopulation(v *perlVerdict, rd, listed bool) bool { return !v.Refused && rd && !listed }

// supersetPopulation is the files perl and recursive descent both accept, less
// the lexer-listed. A file with no perl label is an error, never a smaller population.
func supersetPopulation(files []File, in verdictInputs) ([]File, error) {
	var pop []File
	for _, f := range files {
		v := in.Perl(f.Path)
		if v == nil {
			return nil, fmt.Errorf("%s has no perl label", f.Path)
		}
		rd, err := in.RD(f.Path)
		if err != nil {
			return nil, err
		}
		if inPopulation(v, rd, in.Excluded[f.Path]) {
			pop = append(pop, f)
		}
	}
	return pop, nil
}
